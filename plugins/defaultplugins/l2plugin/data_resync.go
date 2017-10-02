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

package l2plugin

import (
	"fmt"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppdump"
)

// Resync writes BDs to the empty VPP
func (plugin *BDConfigurator) Resync(nbBDs []*l2.BridgeDomains_BridgeDomain) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BDs begin.")

	// Step 0: Dump actual state of the VPP
	vppBDs, err := vppdump.DumpBridgeDomains(plugin.Log, plugin.vppChan)
	if err != nil {
		return err
	}

	pluginID := core.PluginName("defaultvppplugins-l2plugin")

	var wasError error

	// Step 1: delete existing vpp configuration (current ModifyBridgeDomain does it also... need to improve that first)
	for vppIdx, vppBD := range vppBDs {
		hackIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(plugin.Log, pluginID,
			"hack_sw_if_indexes", ifaceidx.IndexMetadata))

		// hack to reuse existing binary call wrappers
		hackBD := l2.BridgeDomains_BridgeDomain(vppBD.BridgeDomains_BridgeDomain)
		for _, vppBDIface := range vppBD.Interfaces {
			hackIfaceName := fmt.Sprintf("%d", vppBDIface.SwIfIndex)
			hackIfIndexes.RegisterName(hackIfaceName, vppBDIface.SwIfIndex, nil)
			hackBDIface := l2.BridgeDomains_BridgeDomain_Interfaces(vppBDIface.BridgeDomains_BridgeDomain_Interfaces)
			hackBDIface.Name = hackIfaceName
			hackBD.Interfaces = append(hackBD.Interfaces, &hackBDIface)
		}

		vppcalls.VppUnsetAllInterfacesFromBridgeDomain(&hackBD, vppIdx,
			hackIfIndexes, plugin.Log, plugin.vppChan)
		err := plugin.deleteBridgeDomain(&hackBD, vppIdx)
		// TODO check if it is ok to delete the initial BD
		if err != nil {
			wasError = err
		}
	}

	// Step 2: create missing vpp configuration
	for _, nbBD := range nbBDs {
		err := plugin.ConfigureBridgeDomain(nbBD)
		if err != nil {
			wasError = err
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BDs end. ", wasError)

	return wasError
}

// Resync writes FIBs to the empty VPP
func (plugin *FIBConfigurator) Resync(fibConfig []*l2.FibTableEntries_FibTableEntry) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC FIBs begin.")

	activeDomains, err := vppdump.DumpBridgeDomainIDs(plugin.Log, plugin.syncVppChannel)
	if err != nil {
		return err
	}
	for _, domainID := range activeDomains {
		plugin.LookupFIBEntries(domainID)
	}

	for _, fib := range fibConfig {
		plugin.Add(fib, func(err2 error) {
			if err2 != nil {
				plugin.Log.Error(err2)
			}
		})
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC FIBs end.")

	return nil
}

// Resync writes XCons to the empty VPP
func (plugin *XConnectConfigurator) Resync(xcConfig []*l2.XConnectPairs_XConnectPair) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC XConnect begin.")

	err := plugin.LookupXConnectPairs()
	if err != nil {
		return err
	}

	var wasError error
	for _, xcon := range xcConfig {
		err = plugin.ConfigureXConnectPair(xcon)
		if err != nil {
			wasError = err
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC XConnect end. ", wasError)

	return wasError
}
