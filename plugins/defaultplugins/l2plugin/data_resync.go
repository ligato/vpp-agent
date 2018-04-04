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
	"strings"

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	if_dump "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
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
	vppBDs, err := vppdump.DumpBridgeDomains(plugin.vppChan, plugin.Stopwatch)
	if err != nil {
		return err
	}

	// Correlate with NB config
	var wasErr error
	for vppBDIdx, vppBD := range vppBDs {
		// tag is bridge domain name (unique identifier)
		tag := vppBD.Name
		// Find NB bridge domain with the same name
		var nbBD *l2.BridgeDomains_BridgeDomain
		for _, nbBDConfig := range nbBDs {
			if tag == nbBDConfig.Name {
				nbBD = nbBDConfig
				break
			}
		}

		// NB config does not exist, VPP bridge domain is obsolete
		if nbBD == nil {
			if err := plugin.deleteBridgeDomain(&l2.BridgeDomains_BridgeDomain{
				Name: tag,
			}, vppBDIdx); err != nil {
				plugin.Log.Error(err)
				continue
			}
			plugin.Log.Debugf("RESYNC bridge domain: obsolete config %v (ID %v) removed", tag, vppBDIdx)
		} else {
			// Bridge domain exists, validate
			valid, recreate := plugin.vppValidateBridgeDomainBVI(nbBD, &l2.BridgeDomains_BridgeDomain{
				Name:                tag,
				Learn:               vppBD.Learn,
				Flood:               vppBD.Flood,
				Forward:             vppBD.Forward,
				UnknownUnicastFlood: vppBD.UnknownUnicastFlood,
				ArpTermination:      vppBD.ArpTermination,
				MacAge:              vppBD.MacAge,
			})
			if !valid {
				plugin.Log.Errorf("RESYNC bridge domain: new config %v is invalid", nbBD.Name)
				continue
			}
			if recreate {
				// Internal bridge domain parameters changed, cannot be modified and will be recreated
				if err := plugin.deleteBridgeDomain(&l2.BridgeDomains_BridgeDomain{
					Name: tag,
				}, vppBDIdx); err != nil {
					plugin.Log.Error(err)
					continue
				}
				if err := plugin.ConfigureBridgeDomain(nbBD); err != nil {
					plugin.Log.Error(err)
					continue
				}
				plugin.Log.Debugf("RESYNC bridge domains: config %v recreated", nbBD.Name)
				continue
			}

			// todo currently it is not possible to dump interfaces. In order to prevent BD removal, unset all available interfaces
			// Dump all interfaces
			interfaceMap, err := if_dump.DumpInterfaces(plugin.Log, plugin.vppChan, nil)
			if err != nil {
				plugin.Log.Error(err)
				wasErr = err
				continue
			}
			// Prepare a list of interface objects with proper name
			var interfacesToUnset []*l2.BridgeDomains_BridgeDomain_Interfaces
			for _, iface := range interfaceMap {
				interfacesToUnset = append(interfacesToUnset, &l2.BridgeDomains_BridgeDomain_Interfaces{
					Name: iface.Name,
				})
			}
			// Remove interfaces from bridge domain. Attempt to unset interface which does not belong to the bridge domain
			// does not cause an error
			vppcalls.UnsetInterfacesFromBridgeDomain(nbBD, vppBDIdx, interfacesToUnset, plugin.SwIfIndices, plugin.Log,
				plugin.vppChan, nil)

			// Set all new interfaces to the bridge domain
			vppcalls.SetInterfacesToBridgeDomain(nbBD, vppBDIdx, nbBD.Interfaces, plugin.SwIfIndices, plugin.Log,
				plugin.vppChan, nil)

			// todo VPP does not support ARP dump, they can be only added at this time
			// Resolve new ARP entries
			for _, arpEntry := range nbBD.ArpTerminationTable {
				if err := vppcalls.VppAddArpTerminationTableEntry(vppBDIdx, arpEntry.PhysAddress, arpEntry.IpAddress,
					plugin.Log, plugin.vppChan, nil); err != nil {
					plugin.Log.Error(err)
					wasErr = err
				}
			}

			// Register bridge domain
			plugin.BdIndices.RegisterName(nbBD.Name, plugin.BridgeDomainIDSeq, nbBD)
			plugin.BridgeDomainIDSeq++

			plugin.Log.Debugf("RESYNC Bridge domain: config %v (ID %v) modified", tag, vppBDIdx)
		}
	}

	// Configure new bridge domains
	for _, newBD := range nbBDs {
		_, _, found := plugin.BdIndices.LookupIdx(newBD.Name)
		if !found {
			if err := plugin.ConfigureBridgeDomain(newBD); err != nil {
				plugin.Log.Error(err)
				continue
			}
			plugin.Log.Debugf("RESYNC bridge domains: new config %v added", newBD.Name)
		}
	}

	return wasErr
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
	vppFIBs, err := vppdump.DumpFIBTableEntries(plugin.syncVppChannel, plugin.Stopwatch)
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
	vppXConns, err := vppdump.DumpXConnectPairs(plugin.vppChan, plugin.Stopwatch)
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
