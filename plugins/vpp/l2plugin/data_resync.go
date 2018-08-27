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

	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/l2idx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
)

// Resync writes missing BDs to the VPP and removes obsolete ones.
func (plugin *BDConfigurator) Resync(nbBDs []*l2.BridgeDomains_BridgeDomain) error {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC BDs begin.")
	// Calculate and log bd resync.
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Re-initialize cache
	plugin.clearMapping()

	// Dump current state of the VPP bridge domains
	vppBDs, err := plugin.bdHandler.DumpBridgeDomains()
	if err != nil {
		return err
	}

	// Correlate with NB config
	var wasErr error
	for vppBDIdx, vppBD := range vppBDs {
		// tag is bridge domain name (unique identifier)
		tag := vppBD.Bd.Name
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
				plugin.log.Error(err)
				continue
			}
			plugin.log.Debugf("RESYNC bridge domain: obsolete config %v (ID %v) removed", tag, vppBDIdx)
		} else {
			// Bridge domain exists, validate
			valid, recreate := plugin.vppValidateBridgeDomainBVI(nbBD, &l2.BridgeDomains_BridgeDomain{
				Name:                tag,
				Learn:               vppBD.Bd.Learn,
				Flood:               vppBD.Bd.Flood,
				Forward:             vppBD.Bd.Forward,
				UnknownUnicastFlood: vppBD.Bd.UnknownUnicastFlood,
				ArpTermination:      vppBD.Bd.ArpTermination,
				MacAge:              vppBD.Bd.MacAge,
			})
			if !valid {
				plugin.log.Errorf("RESYNC bridge domain: new config %v is invalid", nbBD.Name)
				continue
			}
			if recreate {
				// Internal bridge domain parameters changed, cannot be modified and will be recreated
				if err := plugin.deleteBridgeDomain(&l2.BridgeDomains_BridgeDomain{
					Name: tag,
				}, vppBDIdx); err != nil {
					plugin.log.Error(err)
					continue
				}
				if err := plugin.ConfigureBridgeDomain(nbBD); err != nil {
					plugin.log.Error(err)
					continue
				}
				plugin.log.Debugf("RESYNC bridge domains: config %v recreated", nbBD.Name)
				continue
			}

			// todo currently it is not possible to dump interfaces. In order to prevent BD removal, unset all available interfaces
			// Dump all interfaces
			interfaceMap, err := plugin.ifHandler.DumpInterfaces()
			if err != nil {
				plugin.log.Error(err)
				wasErr = err
				continue
			}
			// Prepare a list of interface objects with proper name
			var interfacesToUnset []*l2.BridgeDomains_BridgeDomain_Interfaces
			for _, iface := range interfaceMap {
				interfacesToUnset = append(interfacesToUnset, &l2.BridgeDomains_BridgeDomain_Interfaces{
					Name: iface.Interface.Name,
				})
			}
			// Remove interfaces from bridge domain. Attempt to unset interface which does not belong to the bridge domain
			// does not cause an error
			if _, err := plugin.bdHandler.UnsetInterfacesFromBridgeDomain(nbBD.Name, vppBDIdx, interfacesToUnset, plugin.ifIndexes); err != nil {
				return err
			}
			// Set all new interfaces to the bridge domain
			// todo there is no need to calculate diff from configured interfaces, because currently all available interfaces are set here
			configuredIfs, err := plugin.bdHandler.SetInterfacesToBridgeDomain(nbBD.Name, vppBDIdx, nbBD.Interfaces, plugin.ifIndexes)
			if err != nil {
				return err
			}

			// todo VPP does not support ARP dump, they can be only added at this time
			// Resolve new ARP entries
			for _, arpEntry := range nbBD.ArpTerminationTable {
				if err := plugin.bdHandler.VppAddArpTerminationTableEntry(vppBDIdx, arpEntry.PhysAddress, arpEntry.IpAddress); err != nil {
					plugin.log.Error(err)
					wasErr = err
				}
			}

			// Register bridge domain
			plugin.bdIndexes.RegisterName(nbBD.Name, plugin.bdIDSeq, l2idx.NewBDMetadata(nbBD, configuredIfs))
			plugin.bdIDSeq++

			plugin.log.Debugf("RESYNC Bridge domain: config %v (ID %v) modified", tag, vppBDIdx)
		}
	}

	// Configure new bridge domains
	for _, newBD := range nbBDs {
		_, _, found := plugin.bdIndexes.LookupIdx(newBD.Name)
		if !found {
			if err := plugin.ConfigureBridgeDomain(newBD); err != nil {
				plugin.log.Error(err)
				continue
			}
			plugin.log.Debugf("RESYNC bridge domains: new config %v added", newBD.Name)
		}
	}

	return wasErr
}

// Resync writes missing FIBs to the VPP and removes obsolete ones.
func (plugin *FIBConfigurator) Resync(nbFIBs []*l2.FibTable_FibEntry) error {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC FIBs begin.")
	// Calculate and log fib resync.
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Re-initialize cache
	plugin.clearMapping()

	// Get all FIB entries configured on the VPP
	vppFIBs, err := plugin.fibHandler.DumpFIBTableEntries()
	if err != nil {
		return err
	}

	// Correlate existing config with the NB
	var wasErr error
	for vppFIBmac, vppFIBdata := range vppFIBs {
		exists, meta := func(nbFIBs []*l2.FibTable_FibEntry) (bool, *l2.FibTable_FibEntry) {
			for _, nbFIB := range nbFIBs {
				// Physical address
				if strings.ToUpper(vppFIBmac) != strings.ToUpper(nbFIB.PhysAddress) {
					continue
				}
				// Bridge domain
				bdIdx, _, found := plugin.bdIndexes.LookupIdx(nbFIB.BridgeDomain)
				if !found || vppFIBdata.Meta.BdID != bdIdx {
					continue
				}
				// BVI
				if vppFIBdata.Fib.BridgedVirtualInterface != nbFIB.BridgedVirtualInterface {
					continue
				}
				// Interface
				swIdx, _, found := plugin.ifIndexes.LookupIdx(nbFIB.OutgoingInterface)
				if !found || vppFIBdata.Meta.IfIdx != swIdx {
					continue
				}
				// Is static
				if vppFIBdata.Fib.StaticConfig != nbFIB.StaticConfig {
					continue
				}

				return true, nbFIB
			}
			return false, nil
		}(nbFIBs)

		// Register existing entries, Remove entries missing in NB config (except non-static)
		if exists {
			plugin.fibIndexes.RegisterName(vppFIBmac, plugin.fibIndexSeq, meta)
			plugin.fibIndexSeq++
		} else if vppFIBdata.Fib.StaticConfig {
			// Get appropriate interface/bridge domain names
			ifIdx, _, ifFound := plugin.ifIndexes.LookupName(vppFIBdata.Meta.IfIdx)
			bdIdx, _, bdFound := plugin.bdIndexes.LookupName(vppFIBdata.Meta.BdID)
			if !ifFound || !bdFound {
				// FIB entry cannot be removed without these informations and
				// it should be removed by the VPP
				continue
			}

			plugin.Delete(&l2.FibTable_FibEntry{
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
		_, _, found := plugin.fibIndexes.LookupIdx(nbFIB.PhysAddress)
		if !found {
			plugin.Add(nbFIB, func(callbackErr error) {
				if callbackErr != nil {
					wasErr = callbackErr
				}
			})
		}
	}

	plugin.log.WithField("cfg", plugin).Debug("RESYNC FIBs end.")

	return wasErr
}

// Resync writes missing XCons to the VPP and removes obsolete ones.
func (plugin *XConnectConfigurator) Resync(nbXConns []*l2.XConnectPairs_XConnectPair) error {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC XConnect begin.")
	// Calculate and log xConnect resync.
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Re-initialize cache
	plugin.clearMapping()

	// Read cross connects from the VPP
	vppXConns, err := plugin.xcHandler.DumpXConnectPairs()
	if err != nil {
		return err
	}

	// Correlate with NB config
	var wasErr error
	for _, vppXConn := range vppXConns {
		var existsInNB bool
		var rxIfName, txIfName string
		for _, nbXConn := range nbXConns {
			// Find receive and transmit interface
			rxIfName, _, rxIfFound := plugin.ifIndexes.LookupName(vppXConn.Meta.ReceiveInterfaceSwIfIdx)
			txIfName, _, txIfFound := plugin.ifIndexes.LookupName(vppXConn.Meta.TransmitInterfaceSwIfIdx)
			if !rxIfFound || !txIfFound {
				continue
			}
			if rxIfName == nbXConn.ReceiveInterface && txIfName == nbXConn.TransmitInterface {
				// NB XConnect correlated with VPP
				plugin.xcIndexes.RegisterName(nbXConn.ReceiveInterface, plugin.xcIndexSeq, nbXConn)
				plugin.xcIndexSeq++
				existsInNB = true
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

	// Configure new XConnect pairs
	for _, nbXConn := range nbXConns {
		_, _, found := plugin.xcIndexes.LookupIdx(nbXConn.ReceiveInterface)
		if !found {
			if err := plugin.ConfigureXConnectPair(nbXConn); err != nil {
				wasErr = err
			}
		}
	}

	plugin.log.WithField("cfg", plugin).Debug("RESYNC XConnect end. ", wasErr)

	return wasErr
}
