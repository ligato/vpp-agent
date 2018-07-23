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

//go:generate protoc --proto_path=../model/ipsec --gogo_out=../model/ipsec ../model/ipsec/ipsec.proto

// Package ipsecplugin implements the IPSec plugin that handles management of IPSec for VPP.
package ipsecplugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	iface_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/ipsecidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"
)

// SPDIfCacheEntry contains info about cached assignment of interface to SPD
type SPDIfCacheEntry struct {
	spdID     uint32
	ifaceName string
}

// IPSecConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/ipsec/ipsec.proto"
// and stored in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/ipsec".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type IPSecConfigurator struct {
	log logging.Logger

	// In-memory mappings
	ifIndexes        ifaceidx.SwIfIndexRW
	spdIndexes       ipsecidx.SPDIndexRW
	cachedSpdIndexes ipsecidx.SPDIndexRW
	spdIndexSeq      uint32
	saIndexes        idxvpp.NameToIdxRW
	saIndexSeq       uint32

	// SPC interface cache
	spdIfCache []SPDIfCacheEntry

	// VPP channel
	vppCh govppapi.Channel

	// VPP API handlers
	ifHandler iface_vppcalls.IfVppAPI

	// Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// Init members (channels...) and start go routines
func (plugin *IPSecConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, swIfIndexes ifaceidx.SwIfIndexRW,
	enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-ipsec-plugin")
	plugin.log.Debug("Initializing IPSec configurator")

	// Mappings
	plugin.ifIndexes = swIfIndexes
	plugin.spdIndexes = ipsecidx.NewSPDIndex(nametoidx.NewNameToIdx(plugin.log, "ipsec_spd_indexes", nil))
	plugin.cachedSpdIndexes = ipsecidx.NewSPDIndex(nametoidx.NewNameToIdx(plugin.log, "ipsec_cached_spd_indexes", nil))
	plugin.saIndexes = nametoidx.NewNameToIdx(plugin.log, "ipsec_sa_indexes", ifaceidx.IndexMetadata)
	plugin.spdIndexSeq = 1
	plugin.saIndexSeq = 1

	// VPP channel
	plugin.vppCh, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Stopwatch
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("IPSecConfigurator", plugin.log)
	}

	// VPP API handlers
	if plugin.ifHandler, err = iface_vppcalls.NewIfVppHandler(plugin.vppCh, plugin.log, plugin.stopwatch); err != nil {
		return err
	}

	// Message compatibility
	if err = plugin.vppCh.CheckMessageCompatibility(vppcalls.IPSecMessages...); err != nil {
		plugin.log.Error(err)
		return err
	}

	return nil
}

// Close GOVPP channel
func (plugin *IPSecConfigurator) Close() error {
	return safeclose.Close(plugin.vppCh)
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (plugin *IPSecConfigurator) clearMapping() {
	plugin.spdIndexes.Clear()
	plugin.cachedSpdIndexes.Clear()
	plugin.saIndexes.Clear()
}

// GetSaIndexes returns security association indexes
func (plugin *IPSecConfigurator) GetSaIndexes() idxvpp.NameToIdxRW {
	return plugin.saIndexes
}

// GetSaIndexes returns security policy database indexes
func (plugin *IPSecConfigurator) GetSpdIndexes() ipsecidx.SPDIndex {
	return plugin.spdIndexes
}

// ConfigureSPD configures Security Policy Database in VPP
func (plugin *IPSecConfigurator) ConfigureSPD(spd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.log.Debugf("Configuring SPD %v", spd.Name)

	spdID := plugin.spdIndexSeq
	plugin.spdIndexSeq++

	for _, entry := range spd.PolicyEntries {
		if entry.Sa != "" {
			if _, _, exists := plugin.saIndexes.LookupIdx(entry.Sa); !exists {
				plugin.log.Warnf("SA %q for SPD %q not found, caching SPD configuration", entry.Sa, spd.Name)
				plugin.cachedSpdIndexes.RegisterName(spd.Name, spdID, spd)
				return nil
			}
		}
	}

	return plugin.configureSPD(spdID, spd)
}

func (plugin *IPSecConfigurator) configureSPD(spdID uint32, spd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.log.Debugf("configuring SPD %v (%d)", spd.Name, spdID)

	if err := vppcalls.AddSPD(spdID, plugin.vppCh, plugin.stopwatch); err != nil {
		return err
	}

	plugin.spdIndexes.RegisterName(spd.Name, spdID, spd)
	plugin.log.Infof("Registered SPD %v (%d)", spd.Name, spdID)

	for _, iface := range spd.Interfaces {
		plugin.log.Debugf("Assigning SPD to interface %v", iface)

		swIfIdx, _, exists := plugin.ifIndexes.LookupIdx(iface.Name)
		if !exists {
			plugin.log.Infof("Interface %q for SPD %q not found, caching assignment of interface to SPD", iface.Name, spd.Name)
			plugin.cacheSPDInterfaceAssignment(spdID, iface.Name)
			continue
		}

		if err := vppcalls.InterfaceAddSPD(spdID, swIfIdx, plugin.vppCh, plugin.stopwatch); err != nil {
			plugin.log.Errorf("assigning interface to SPD failed: %v", err)
			continue
		}

		plugin.log.Infof("Assigned SPD %q to interface %q", spd.Name, iface.Name)
	}

	for _, entry := range spd.PolicyEntries {
		plugin.log.Infof("Adding SPD policy entry %v", entry)

		var saID uint32
		if entry.Sa != "" {
			var exists bool
			if saID, _, exists = plugin.saIndexes.LookupIdx(entry.Sa); !exists {
				plugin.log.Warnf("SA %q for SPD %q not found, skipping SPD policy entry configuration", entry.Sa, spd.Name)
				continue
			}
		}

		if err := vppcalls.AddSPDEntry(spdID, saID, entry, plugin.vppCh, plugin.stopwatch); err != nil {
			plugin.log.Errorf("adding SPD policy entry failed: %v", err)
			continue
		}

		plugin.log.Infof("Added SPD policy entry")
	}

	plugin.log.Infof("Configured SPD %v", spd.Name)

	return nil
}

// ModifySPD modifies Security Policy Database in VPP
func (plugin *IPSecConfigurator) ModifySPD(oldSpd *ipsec.SecurityPolicyDatabases_SPD, newSpd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.log.Debugf("Modifying SPD %v", oldSpd.Name)

	if err := plugin.DeleteSPD(oldSpd); err != nil {
		plugin.log.Error("deleting old SPD failed:", err)
		return err
	}
	if err := plugin.ConfigureSPD(newSpd); err != nil {
		plugin.log.Error("configuring new SPD failed:", err)
		return err
	}

	return nil
}

// DeleteSPD deletes Security Policy Database in VPP
func (plugin *IPSecConfigurator) DeleteSPD(oldSpd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.log.Debugf("Deleting SPD %v", oldSpd.Name)

	if spdID, _, found := plugin.cachedSpdIndexes.LookupIdx(oldSpd.Name); found {
		plugin.log.Debugf("removing cached SPD %v", spdID)
		plugin.cachedSpdIndexes.UnregisterName(oldSpd.Name)
		return nil
	}

	spdID, _, exists := plugin.spdIndexes.LookupIdx(oldSpd.Name)
	if !exists {
		plugin.log.Warnf("SPD %q not found", oldSpd.Name)
		return nil
	}
	if err := vppcalls.DelSPD(spdID, plugin.vppCh, plugin.stopwatch); err != nil {
		return err
	}

	// remove cache entries related to the SPD
	for i, entry := range plugin.spdIfCache {
		if entry.spdID == spdID {
			plugin.log.Debugf("Removing cache entry for assignment of SPD %q to interface %q", entry.spdID, entry.ifaceName)
			plugin.spdIfCache = append(plugin.spdIfCache[:i], plugin.spdIfCache[i+1:]...)
		}
	}

	plugin.spdIndexes.UnregisterName(oldSpd.Name)
	plugin.log.Infof("Deleted SPD %v", oldSpd.Name)

	return nil
}

// ConfigureSA configures Security Association in VPP
func (plugin *IPSecConfigurator) ConfigureSA(sa *ipsec.SecurityAssociations_SA) error {
	plugin.log.Debugf("Configuring SA %v", sa.Name)

	saID := plugin.saIndexSeq
	plugin.saIndexSeq++

	if err := vppcalls.AddSAEntry(saID, sa, plugin.vppCh, plugin.stopwatch); err != nil {
		return err
	}

	plugin.saIndexes.RegisterName(sa.Name, saID, nil)
	plugin.log.Infof("Registered SA %v (%d)", sa.Name, saID)

	for _, cached := range plugin.cachedSpdIndexes.LookupBySA(sa.Name) {
		for _, entry := range cached.SPD.PolicyEntries {
			if entry.Sa != "" {
				if _, _, exists := plugin.saIndexes.LookupIdx(entry.Sa); !exists {
					plugin.log.Warnf("SA %q for SPD %q not found, keeping SPD in cache", entry.Sa, cached.SPD.Name)
					return nil
				}
			}
		}
		if err := plugin.configureSPD(cached.SpdID, cached.SPD); err != nil {
			plugin.log.Errorf("configuring cached SPD failed: %v", err)
		} else {
			plugin.cachedSpdIndexes.UnregisterName(cached.SPD.Name)
		}
	}

	return nil
}

// ModifySA modifies Security Association in VPP
func (plugin *IPSecConfigurator) ModifySA(oldSa *ipsec.SecurityAssociations_SA, newSa *ipsec.SecurityAssociations_SA) error {
	plugin.log.Debugf("Modifying SA %v", oldSa.Name)

	// TODO: check if only keys change and use IpsecSaSetKey vpp call

	if err := plugin.DeleteSA(oldSa); err != nil {
		plugin.log.Error("deleting old SPD failed:", err)
		return err
	}
	if err := plugin.ConfigureSA(newSa); err != nil {
		plugin.log.Error("configuring new SPD failed:", err)
		return err
	}

	return nil
}

// DeleteSA deletes Security Association in VPP
func (plugin *IPSecConfigurator) DeleteSA(oldSa *ipsec.SecurityAssociations_SA) error {
	plugin.log.Debugf("Deleting SA %v", oldSa.Name)

	saID, _, exists := plugin.saIndexes.LookupIdx(oldSa.Name)
	if !exists {
		plugin.log.Warnf("SA %q not found", oldSa.Name)
		return nil
	}

	for _, entry := range plugin.spdIndexes.LookupBySA(oldSa.Name) {
		if err := plugin.DeleteSPD(entry.SPD); err != nil {
			plugin.log.Errorf("deleting SPD to be cached failed: %v", err)
			continue
		}
		plugin.cachedSpdIndexes.RegisterName(entry.SPD.Name, entry.SpdID, entry.SPD)
		plugin.log.Warnf("caching SPD %v due removed SA %v", entry.SPD.Name, oldSa.Name)
	}

	if err := vppcalls.DelSAEntry(saID, oldSa, plugin.vppCh, plugin.stopwatch); err != nil {
		return err
	}

	plugin.saIndexes.UnregisterName(oldSa.Name)
	plugin.log.Infof("Deleted SA %v", oldSa.Name)

	return nil
}

// ConfigureTunnel configures Tunnel interface in VPP
func (plugin *IPSecConfigurator) ConfigureTunnel(tunnel *ipsec.TunnelInterfaces_Tunnel) error {
	plugin.log.Debugf("Configuring Tunnel %v", tunnel.Name)

	ifIdx, err := vppcalls.AddTunnelInterface(tunnel, plugin.vppCh, plugin.stopwatch)
	if err != nil {
		return err
	}

	plugin.ifIndexes.RegisterName(tunnel.Name, ifIdx, nil)
	plugin.log.Infof("Registered Tunnel %v (%d)", tunnel.Name, ifIdx)

	if err := plugin.ifHandler.SetInterfaceVRF(ifIdx, tunnel.Vrf); err != nil {
		return err
	}

	ipAddrs, err := addrs.StrAddrsToStruct(tunnel.IpAddresses)
	if err != nil {
		return err
	}
	for _, ip := range ipAddrs {
		if err := plugin.ifHandler.AddInterfaceIP(ifIdx, ip); err != nil {
			plugin.log.Errorf("adding interface IP address failed: %v", err)
			return err
		}
	}

	if tunnel.Enabled {
		if err := plugin.ifHandler.InterfaceAdminUp(ifIdx); err != nil {
			plugin.log.Debugf("setting interface up failed: %v", err)
			return err
		}
	}

	return nil
}

// ModifyTunnel modifies Tunnel interface in VPP
func (plugin *IPSecConfigurator) ModifyTunnel(oldTunnel *ipsec.TunnelInterfaces_Tunnel, newTunnel *ipsec.TunnelInterfaces_Tunnel) error {
	plugin.log.Debugf("Modifying Tunnel %v", oldTunnel.Name)

	if err := plugin.DeleteTunnel(oldTunnel); err != nil {
		plugin.log.Error("deleting old Tunnel failed:", err)
		return err
	}
	if err := plugin.ConfigureTunnel(newTunnel); err != nil {
		plugin.log.Error("configuring new Tunnel failed:", err)
		return err
	}

	return nil
}

// DeleteTunnel deletes Tunnel interface in VPP
func (plugin *IPSecConfigurator) DeleteTunnel(oldTunnel *ipsec.TunnelInterfaces_Tunnel) error {
	plugin.log.Debugf("Deleting Tunnel %v", oldTunnel.Name)

	ifIdx, _, exists := plugin.ifIndexes.LookupIdx(oldTunnel.Name)
	if !exists {
		plugin.log.Warnf("Tunnel %q not found", oldTunnel.Name)
		return nil
	}

	if err := vppcalls.DelTunnelInterface(ifIdx, oldTunnel, plugin.vppCh, plugin.stopwatch); err != nil {
		return err
	}

	plugin.ifIndexes.UnregisterName(oldTunnel.Name)
	plugin.log.Infof("Deleted Tunnel %v", oldTunnel.Name)

	return nil
}

// ResolveCreatedInterface is responsible for reconfiguring cached assignments
func (plugin *IPSecConfigurator) ResolveCreatedInterface(ifName string, swIfIdx uint32) {
	for i, entry := range plugin.spdIfCache {
		if entry.ifaceName == ifName {
			plugin.log.Infof("Assigning SPD %v to interface %q", entry.spdID, ifName)

			// TODO: loop through stored deletes, this is now needed because old assignment might still exist
			if err := vppcalls.InterfaceDelSPD(entry.spdID, swIfIdx, plugin.vppCh, plugin.stopwatch); err != nil {
				plugin.log.Errorf("unassigning interface from SPD failed: %v", err)
			} else {
				plugin.log.Infof("Unassigned SPD %v from interface %q", entry.spdID, ifName)
			}

			if err := vppcalls.InterfaceAddSPD(entry.spdID, swIfIdx, plugin.vppCh, plugin.stopwatch); err != nil {
				plugin.log.Errorf("assigning interface to SPD failed: %v", err)
				continue
			} else {
				plugin.log.Infof("Assigned SPD %v to interface %q", entry.spdID, entry.ifaceName)
			}

			plugin.spdIfCache = append(plugin.spdIfCache[:i], plugin.spdIfCache[i+1:]...)
		}
	}
}

// ResolveDeletedInterface is responsible for caching assignments for future reconfiguration
func (plugin *IPSecConfigurator) ResolveDeletedInterface(ifName string, swIfIdx uint32) {
	for _, assign := range plugin.spdIndexes.LookupByInterface(ifName) {
		plugin.log.Infof("Unassigning SPD %v from interface %q", assign.SpdID, ifName)

		// TODO: just store this for future, because this will fail since swIfIdx no longer exists
		if err := vppcalls.InterfaceDelSPD(assign.SpdID, swIfIdx, plugin.vppCh, plugin.stopwatch); err != nil {
			plugin.log.Errorf("unassigning interface from SPD failed: %v", err)
		} else {
			plugin.log.Infof("Unassigned SPD %v from interface %q", assign.SpdID, ifName)
		}

		plugin.cacheSPDInterfaceAssignment(assign.SpdID, ifName)
	}
}

func (plugin *IPSecConfigurator) cacheSPDInterfaceAssignment(spdID uint32, ifaceName string) {
	plugin.log.Debugf("caching SPD %v interface assignment to %v", spdID, ifaceName)
	plugin.spdIfCache = append(plugin.spdIfCache, SPDIfCacheEntry{
		ifaceName: ifaceName,
		spdID:     spdID,
	})
}
