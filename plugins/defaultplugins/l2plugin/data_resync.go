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
	"strings"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppdump"
)

// Resync writes missing BDs to the VPP and removes obsolete ones.
func (plugin *BDConfigurator) Resync(nbBDs []*l2.BridgeDomains_BridgeDomain) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BDs begin.")
	// Calculate and log bd resync.
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Dump current state of the VPP bridge domains
	vppBDs, err := vppdump.DumpBridgeDomains(plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.BridgeDomainDump{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	// Read persistent mapping
	persistent := nametoidx.NewNameToIdx(logrus.DefaultLogger(), core.PluginName("defaultvppplugins-l2plugin"), "bd resync corr", nil)
	err = persist.Marshalling(plugin.ServiceLabel.GetAgentLabel(), plugin.BdIndices.GetMapping(), persistent)
	if err != nil {
		return err
	}

	// todo
	for _, name := range persistent.ListNames() {
		idx, _, _ := persistent.LookupIdx(name)
		plugin.Log.Warnf("persistent mapping entry %v %v", name, idx)
	}

	// Handle the case where persistent mapping is not available
	var wasErr error
	if len(persistent.ListNames()) == 0 && len(vppBDs) > 0 {
		plugin.Log.Infof("Persisten mapping for bridge domains is empty (%v unknown bridge domains)", len(vppBDs))
		// There is no way how to correlate NB and VPP configuration so remove all bridge domains from the VPP
		for bdID, unknownVppBd := range vppBDs {
			// Try to reconstruct bridge domain with interfaces
			// todo: bridge domain dump returns no interfaces. It is possible to remove bridge domain to unsetting it,
			// but it would be better to do so
			bd := l2.BridgeDomains_BridgeDomain(unknownVppBd.BridgeDomains_BridgeDomain)
			bd.Name = "unknownBD" + string(bdID)
			plugin.BdIndices.RegisterName(bd.Name, bdID, nil)
			wasErr = plugin.DeleteBridgeDomain(&bd)
		}
		// Configure NB
		for _, nbBd := range nbBDs {
			wasErr = plugin.ConfigureBridgeDomain(nbBd)
		}
		return wasErr
	}

	pluginID := core.PluginName("defaultvppplugins-l2plugin")

	var wasError error

	// todo: if the bridge domain dump is fixed, it is better to correlate existing bridge domains with NB config and
	// do partial changes only
	// Step 1: Delete existing vpp configuration (current ModifyBridgeDomain does it also... need to improve that first)
	for vppIdx, vppBD := range vppBDs {
		hackIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(plugin.Log, pluginID,
			"hack_sw_if_indexes", ifaceidx.IndexMetadata))

		// hack to reuse an existing binary call wrappers
		hackBD := l2.BridgeDomains_BridgeDomain(vppBD.BridgeDomains_BridgeDomain)
		for _, vppBDIface := range vppBD.Interfaces {
			hackIfaceName := fmt.Sprintf("%d", vppBDIface.SwIfIndex)
			hackIfIndexes.RegisterName(hackIfaceName, vppBDIface.SwIfIndex, nil)
			hackBDIface := l2.BridgeDomains_BridgeDomain_Interfaces(vppBDIface.BridgeDomains_BridgeDomain_Interfaces)
			hackBDIface.Name = hackIfaceName
			hackBD.Interfaces = append(hackBD.Interfaces, &hackBDIface)
		}

		vppcalls.UnsetInterfacesFromBridgeDomain(&hackBD, vppIdx, hackBD.Interfaces,
			hackIfIndexes, plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))
		err := plugin.deleteBridgeDomain(&hackBD, vppIdx)
		// TODO check if it is ok to delete the initial BD
		if err != nil {
			wasError = err
		}
	}

	// Step 2: Create missing vpp configuration.
	for _, nbBD := range nbBDs {
		err := plugin.ConfigureBridgeDomain(nbBD)
		if err != nil {
			wasError = err
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BDs end. ", wasError)

	return wasError
}

// Resync writes missing FIBs to the VPP and removes obsolete ones.
func (plugin *FIBConfigurator) Resync(nbFIBs []*l2.FibTableEntries_FibTableEntry) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC FIBs begin.")
	// Calculate and log fib resync.
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Get all FIB entries configured on the VPP
	vppFIBs, err := vppdump.DumpFIBTableEntries(plugin.Log, plugin.syncVppChannel,
		measure.GetTimeLog(l2ba.L2FibTableDump{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	// Correlate existing config with the NB
	var wasErr error
	for vppFIBmac, vppFIBdata := range vppFIBs {
		exists, meta := func(nbFIBs []*l2.FibTableEntries_FibTableEntry) (bool, *FIBMeta) {
			for _, nbFIB := range nbFIBs {
				// Physical address
				if strings.ToUpper(vppFIBmac) != strings.ToUpper(nbFIB.PhysAddress) {
					continue
				}
				// Bridge domain
				bdIdx, _, found := plugin.BdIndexes.LookupIdx(nbFIB.BridgeDomain)
				if !found || vppFIBdata.BridgeDomainIdx != bdIdx {
					continue
				}
				// BVI
				if vppFIBdata.BridgedVirtualInterface != nbFIB.BridgedVirtualInterface {
					continue
				}
				// Interface
				swIdx, _, found := plugin.SwIfIndexes.LookupIdx(nbFIB.OutgoingInterface)
				if !found || vppFIBdata.OutgoingInterfaceSwIfIdx != swIdx {
					continue
				}
				// Is static
				if vppFIBdata.StaticConfig != nbFIB.StaticConfig {
					continue
				}

				// Prepare FIB metadata
				meta := &FIBMeta{nbFIB.OutgoingInterface, nbFIB.BridgeDomain, nbFIB.BridgedVirtualInterface, nbFIB.StaticConfig}

				return true, meta
			}
			return false, nil
		}(nbFIBs)

		// Register existing entries, Remove entries missing in NB config (except non-static)
		if exists {
			plugin.FibIndexes.RegisterName(vppFIBmac, plugin.FibIndexSeq, meta)
			plugin.FibIndexSeq++
		} else if vppFIBdata.StaticConfig {
			// Get appropriate interface/bridge domain names
			ifIdx, _, ifFound := plugin.SwIfIndexes.LookupName(vppFIBdata.OutgoingInterfaceSwIfIdx)
			bdIdx, _, bdFound := plugin.BdIndexes.LookupName(vppFIBdata.BridgeDomainIdx)
			if !ifFound || !bdFound {
				// FIB entry cannot be removed without these informations and
				// it should be removed by the VPP
				continue
			}

			plugin.Delete(&l2.FibTableEntries_FibTableEntry{
				PhysAddress:       vppFIBmac,
				OutgoingInterface: ifIdx,
				BridgeDomain:      bdIdx,
			}, func(callbackErr error) {
				if callbackErr != nil {
					wasErr = callbackErr
				}
			})

		}
	}

	// Configure all unregistered FIB entries from NB config
	for _, nbFIB := range nbFIBs {
		_, _, found := plugin.FibIndexes.LookupIdx(nbFIB.PhysAddress)
		if !found {
			plugin.Add(nbFIB, func(callbackErr error) {
				if callbackErr != nil {
					wasErr = callbackErr
				}
			})
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC FIBs end.")

	return wasErr
}

// Resync writes missing XCons to the VPP and removes obsolete ones.
func (plugin *XConnectConfigurator) Resync(nbXConns []*l2.XConnectPairs_XConnectPair) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC XConnect begin.")
	// Calculate and log xConnect resync.
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Read cross connect from the VPP
	vppXConns, err := vppdump.DumpXConnectPairs(plugin.Log, plugin.vppChan,
		measure.GetTimeLog(l2ba.L2XconnectDump{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	// Correlate with NB config
	var wasErr error
	for _, vppXConn := range vppXConns {
		var existsInNB bool
		var rxIfName, txIfName string
		for _, nbXConn := range nbXConns {
			// find receive and transmitt interface
			rxIfName, _, rxIfExists := plugin.SwIfIndexes.LookupName(vppXConn.ReceiveInterfaceSwIfIdx)
			txIfName, _, txIfExists := plugin.SwIfIndexes.LookupName(vppXConn.TransmitInterfaceSwIfIdx)
			if !rxIfExists || !txIfExists {
				continue
			}

			if rxIfName == nbXConn.ReceiveInterface && txIfName == nbXConn.TransmitInterface {
				// NB interface already exists
				plugin.XcIndexes.RegisterName(nbXConn.ReceiveInterface, plugin.XcIndexSeq, &XConnectMeta{
					TransmitInterface: nbXConn.TransmitInterface,
					configured:        rxIfExists && txIfExists,
				})
				plugin.XcIndexSeq++
			}
		}
		if !existsInNB {
			if err := plugin.DeleteXConnectPair(&l2.XConnectPairs_XConnectPair{
				ReceiveInterface:  rxIfName,
				TransmitInterface: txIfName,
			}); err != nil {
				wasErr = err
			}
		}
	}

	// Configure new xConnect pairs
	for _, nbXConn := range nbXConns {
		_, _, found := plugin.XcIndexes.LookupIdx(nbXConn.ReceiveInterface)
		if !found {
			if err := plugin.ConfigureXConnectPair(nbXConn); err != nil {
				wasErr = err
			}
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC XConnect end. ", wasErr)

	return wasErr
}
