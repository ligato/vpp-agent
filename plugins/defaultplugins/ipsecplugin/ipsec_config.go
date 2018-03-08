// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate protoc --proto_path=../common/model/ipsec --gogo_out=../common/model/ipsec ../common/model/ipsec/ipsec.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/ipsec.api.json --output-dir=../common/bin_api

// Package ipsecplugin implements the IPSec plugin that handles management of IPSec for VPP.
package ipsecplugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/ipsec"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ipsecplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// SPDIfCacheEntry contains info about cached assignment of interface to SPD
type SPDIfCacheEntry struct {
	ifaceName string
	spdID     uint32
}

// IPSecConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/ipsec/ipsec.proto"
// and stored in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/ipsec".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type IPSecConfigurator struct {
	Log       logging.Logger
	Stopwatch *measure.Stopwatch // timer used to measure and store time

	GoVppmux govppmux.API
	vppCh    *govppapi.Channel

	SwIfIndexes ifaceidx.SwIfIndexRW

	SaIndexSeq  uint32
	SaIndexes   idxvpp.NameToIdxRW
	SpdIndexSeq uint32
	SpdIndexes  idxvpp.NameToIdxRW

	SPDIfCache []SPDIfCacheEntry
}

// Init members (channels...) and start go routines
func (plugin *IPSecConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing IPSec configurator")

	plugin.vppCh, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}
	if err := vppcalls.CheckMsgCompatibilityForIPSec(plugin.vppCh); err != nil {
		return err
	}

	return nil
}

// Close GOVPP channel
func (plugin *IPSecConfigurator) Close() error {
	return safeclose.Close(plugin.vppCh)
}

// ConfigureSPD configures SPD
func (plugin *IPSecConfigurator) ConfigureSPD(spd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.Log.Debugf("Configuring SPD %v", spd.Name)

	spdID := plugin.SpdIndexSeq
	plugin.SpdIndexSeq++

	if err := vppcalls.AddSPD(spdID, plugin.vppCh, plugin.Stopwatch); err != nil {
		return err
	}

	plugin.SpdIndexes.RegisterName(spd.Name, spdID, nil)
	plugin.Log.Infof("Registered SPD %v (%d)", spd.Name, spdID)

	for _, iface := range spd.Interfaces {
		plugin.Log.Infof("Assigning SPD to interface %v", iface)

		swIfIdx, _, exists := plugin.SwIfIndexes.LookupIdx(iface.Name)
		if !exists {
			plugin.SPDIfCache = append(plugin.SPDIfCache, SPDIfCacheEntry{
				ifaceName: iface.Name,
				spdID:     spdID,
			})
			plugin.Log.Warnf("Interface %q for SPD %q not found, caching assignment of interface to SPD", iface.Name, spd.Name)
			continue
		}

		if err := vppcalls.InterfaceAddSPD(spdID, swIfIdx, plugin.vppCh, plugin.Stopwatch); err != nil {
			plugin.Log.Errorf("assigning interface to SPD failed: %v", err)
			continue
		}

		plugin.Log.Infof("Assigned SPD %q to interface %q", spd.Name, iface.Name)
	}

	for _, entry := range spd.PolicyEntries {
		plugin.Log.Infof("Adding SPD policy entry %v", entry)

		var saID uint32
		if entry.Sa != "" {
			var exists bool
			if saID, _, exists = plugin.SaIndexes.LookupIdx(entry.Sa); !exists {
				plugin.Log.Warnf("SA %q for SPD %q not found, skipping SPD policy entry configuration", entry.Sa, spd.Name)
				continue
			}
		}

		if err := vppcalls.AddSPDEntry(spdID, saID, entry, plugin.vppCh, plugin.Stopwatch); err != nil {
			plugin.Log.Errorf("adding SPD policy entry failed: %v", err)
			continue
		}

		plugin.Log.Infof("Added SPD policy entry")
	}

	plugin.Log.Infof("Configured SPD %v", spd.Name)

	return nil
}

// ModifySPD
func (plugin *IPSecConfigurator) ModifySPD(oldSpd *ipsec.SecurityPolicyDatabases_SPD, spd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.Log.Debugf("Modifying SPD %v", spd.Name)

	return nil
}

// DeleteSPD
func (plugin *IPSecConfigurator) DeleteSPD(oldSpd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.Log.Debugf("Deleting SPD %v", oldSpd.Name)

	spdID, _, exists := plugin.SpdIndexes.LookupIdx(oldSpd.Name)
	if !exists {
		plugin.Log.Warnf("SPD %q not found", oldSpd.Name)
		return nil
	}
	if err := vppcalls.DelSPD(spdID, plugin.vppCh, plugin.Stopwatch); err != nil {
		return err
	}

	// remove cache entries related to the SPD
	for i, entry := range plugin.SPDIfCache {
		if entry.spdID == spdID {
			plugin.Log.Debugf("Removing cache entry for assignment of SPD %q to interface %q", entry.spdID, entry.ifaceName)
			plugin.SPDIfCache = append(plugin.SPDIfCache[:i], plugin.SPDIfCache[i+1:]...)
		}
	}

	plugin.SpdIndexes.UnregisterName(oldSpd.Name)
	plugin.Log.Infof("Deleted SPD %v", oldSpd.Name)

	return nil
}

// ConfigureSA
func (plugin *IPSecConfigurator) ConfigureSA(sa *ipsec.SecurityAssociations_SA) error {
	plugin.Log.Debugf("Configuring SA %v", sa.Name)

	saID := plugin.SaIndexSeq
	plugin.SaIndexSeq++

	if err := vppcalls.AddSAEntry(saID, sa, plugin.vppCh, plugin.Stopwatch); err != nil {
		return err
	}

	plugin.SaIndexes.RegisterName(sa.Name, saID, nil)
	plugin.Log.Infof("Registered SA %v (%d)", sa.Name, saID)

	return nil
}

// ModifySA
func (plugin *IPSecConfigurator) ModifySA(oldSa *ipsec.SecurityAssociations_SA, sa *ipsec.SecurityAssociations_SA) error {
	plugin.Log.Debugf("Modifying SA %v", sa.Name)

	return nil
}

// DeleteSA
func (plugin *IPSecConfigurator) DeleteSA(oldSa *ipsec.SecurityAssociations_SA) error {
	plugin.Log.Debugf("Deleting SA %v", oldSa.Name)

	saID, _, exists := plugin.SaIndexes.LookupIdx(oldSa.Name)
	if !exists {
		plugin.Log.Warnf("SA %q not found", oldSa.Name)
		return nil
	}
	if err := vppcalls.DelSAEntry(saID, oldSa, plugin.vppCh, plugin.Stopwatch); err != nil {
		return err
	}

	plugin.SaIndexes.UnregisterName(oldSa.Name)
	plugin.Log.Infof("Deleted SA %v", oldSa.Name)

	return nil
}

// ResolveCreatedInterface is responsible for reconfiguring cached assignments
func (plugin *IPSecConfigurator) ResolveCreatedInterface(ifName string, swIfIdx uint32) {
	for i, entry := range plugin.SPDIfCache {
		if entry.ifaceName == ifName {
			plugin.Log.Infof("Assigning SPD %v to interface %q", entry.spdID, ifName)

			if err := vppcalls.InterfaceAddSPD(entry.spdID, swIfIdx, plugin.vppCh, plugin.Stopwatch); err != nil {
				plugin.Log.Errorf("assigning interface to SPD failed: %v", err)
				continue
			}

			plugin.Log.Infof("Assigned SPD %q to interface %q", entry.spdID, entry.ifaceName)
			plugin.SPDIfCache = append(plugin.SPDIfCache[:i], plugin.SPDIfCache[i+1:]...)
		}
	}
}

// ResolveDeletedInterface is responsible for..
func (plugin *IPSecConfigurator) ResolveDeletedInterface(ifName string, swIdx uint32) {

}
