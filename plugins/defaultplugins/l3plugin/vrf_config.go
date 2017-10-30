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

//go:generate binapi-generator --input-file=/usr/share/vpp/api/interface.api.json --output-dir=bin_api

// Package l3plugin implements the L3 plugin that handles L3 FIBs.
package l3plugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/prometheus/common/log"
)

var msgCompatiblityVRF = []govppapi.Message{
	&ip.IPTableAddDel{},
	&ip.IPTableAddDelReply{},
	&interfaces.SwInterfaceSetTable{},
	&interfaces.SwInterfaceSetTableReply{},
	&interfaces.SwInterfaceGetTable{},
	&interfaces.SwInterfaceGetTableReply{},
}

// VrfConfigurator is for managing VRF tables
type VrfConfigurator struct {
	Log logging.Logger

	GoVppmux      govppmux.API
	TableIndexes  idxvpp.NameToIdxRW
	TableIndexSeq uint32
	SwIfIndexes   ifaceidx.SwIfIndex

	vppChan *govppapi.Channel
}

// Init members (channels...) and start go routines
func (plugin *VrfConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing L3 VRF")

	// Init VPP API channel
	if plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}

	return plugin.vppChan.CheckMessageCompatibility(msgCompatiblityVRF...)
}

// Close GOVPP channel
func (plugin *VrfConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// AddTable creates VRF table
func (plugin *VrfConfigurator) AddTable(table *l3.VRFTable) error {
	l := plugin.Log.WithField("name", table.Name)
	l.Debugf("Creating VRF table.")

	// Create and register VRF table
	vrfIdx := plugin.TableIndexSeq

	l = plugin.Log.WithFields(map[string]interface{}{
		"vrfName": table.Name,
		"vrfIdx":  vrfIdx,
	})

	if err := vppcalls.VppAddIPTable(vrfIdx, table.Name, plugin.vppChan); err != nil {
		return err
	}
	plugin.TableIndexes.RegisterName(table.Name, vrfIdx, nil)
	plugin.TableIndexSeq++
	l.Infof("VRF table registered")

	// Set interfaces to VRF
	for _, iface := range table.Interfaces {
		ifaceIdx, _, found := plugin.SwIfIndexes.LookupIdx(iface.Name)
		if !found {
			log.Infof("Interface %v not found", iface.Name)
			continue
		}
		if err := vppcalls.VppSetInterfaceToVRF(vrfIdx, ifaceIdx, plugin.Log, plugin.vppChan); err != nil {
			log.Error("Set interface to VRF failed:", err)
			continue
		}
	}

	return nil
}

// DeleteTable deletes VRF table
func (plugin *VrfConfigurator) DeleteTable(table *l3.VRFTable) error {
	l := plugin.Log.WithField("name", table.Name)
	l.Debugf("Deleting VRF table.")

	vrfIdx, _, found := plugin.TableIndexes.LookupIdx(table.Name)
	if !found {
		l.Debug("Unable to find index for VRF table to be deleted.")
		return nil
	}

	l = plugin.Log.WithFields(map[string]interface{}{
		"vrfName": table.Name,
		"vrfIdx":  vrfIdx,
	})

	// Delete and unregister VRF table
	if err := vppcalls.VppDelIPTable(vrfIdx, table.Name, plugin.vppChan); err != nil {
		return err
	}
	plugin.TableIndexes.UnregisterName(table.Name)
	l.Infof("VRF table unregistered.")

	return nil
}
