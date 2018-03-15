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

package ifplugin

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	_ "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/stn"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
)

const ifTempName = "temp-if-name"

// Resync writes interfaces to the VPP. VPP interfaces are usually configured with tag, which corresponds with interface
// name (exceptions are physical devices, but their name is always equal to vpp internal name). Resync consists of
// following steps:
// 1. Dump all VPP interfaces
// 2. Every VPP interface looks for NB counterpart using tag (name). If found, it is calculated whether modification is
//    needed. Otherwise, the interface is only registered. If interface does not contain tag, it is stored for now and
//    resolved later. Tagged interfaces without NB config are removed.
// 3. Untagged interfaces are correlated heuristically (mac address, ip addresses). If correlation
//    is found, interface is modified if needed and registered.
// 4. All remaining NB interfaces are configured
func (plugin *InterfaceConfigurator) Resync(nbIfs []*intf.Interfaces_Interface) (errs []error) {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface begin.")
	// Calculate and log interface resync
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Dump current state of the VPP interfaces
	vppIfs, err := vppdump.DumpInterfaces(plugin.Log, plugin.vppCh, plugin.Stopwatch)
	if err != nil {
		return []error{err}
	}

	// Cache for untagged interfaces
	unnamedVppIfs := make(map[uint32]*intf.Interfaces_Interface)

	// Iterate over VPP interfaces and try to correlate NB config
	for vppIfIdx, vppIf := range vppIfs {
		if vppIfIdx == 0 {
			// Register interface before removal (to keep state consistent)
			if err := plugin.registerInterface(vppIf.VPPInternalName, vppIfIdx, &vppIf.Interfaces_Interface); err != nil {
				errs = append(errs, err)
			}
			continue
		}
		if vppIf.Name == "" {
			// If interface has no name, it is stored as unnamed and resolved later
			plugin.Log.Debugf("RESYNC interfaces: interface %v has no name (tag)", vppIfIdx)
			unnamedVppIfs[vppIfIdx] = &vppIf.Interfaces_Interface
			continue
		}
		var correlated bool
		for _, nbIf := range nbIfs {
			if vppIf.Name == nbIf.Name {
				correlated = true
				// Register interface to mapping and VPP tag/index
				if err := plugin.registerInterface(vppIf.Name, vppIfIdx, nbIf); err != nil {
					errs = append(errs, err)
				}
				// Calculate whether modification is needed
				if plugin.isIfModified(nbIf, &vppIf.Interfaces_Interface) {
					plugin.Log.Debugf("RESYNC interfaces: modifying interface %v", vppIf.Name)
					if err = plugin.ModifyVPPInterface(nbIf, &vppIf.Interfaces_Interface); err != nil {
						plugin.Log.Errorf("Error while modifying interface: %v", err)
						errs = append(errs, err)
					}
				} else {
					plugin.Log.Debugf("RESYNC interfaces: %v registered without additional changes", vppIf.Name)
				}
				break
			}
		}
		if !correlated {
			// Register interface before removal (to keep state consistent)
			if err := plugin.registerInterface(vppIf.Name, vppIfIdx, &vppIf.Interfaces_Interface); err != nil {
				errs = append(errs, err)
			}
			// VPP interface is obsolete and will be removed (un-configured if physical device)
			plugin.Log.Debugf("RESYNC interfaces: removing obsolete interface %v", vppIf.Name)
			if err = plugin.deleteVPPInterface(&vppIf.Interfaces_Interface, vppIfIdx); err != nil {
				plugin.Log.Errorf("Error while removing interface: %v", err)
				errs = append(errs, err)
			}
		}
	}

	// Now resolve untagged interfaces
	for vppIfIdx, vppIf := range unnamedVppIfs {
		// Try to find NB config which is not registered and correlates with VPP interface
		var correlatedIf *intf.Interfaces_Interface
		for _, nbIf := range nbIfs {
			// Already registered interfaces cannot be correlated again
			_, _, found := plugin.swIfIndexes.LookupIdx(nbIf.Name)
			if found {
				continue
			}
			// Try to correlate heuristically
			correlatedIf = plugin.correlateInterface(vppIf, nbIf)
			if correlatedIf != nil {
				break
			}
		}

		if correlatedIf != nil {
			// Register interface
			if err := plugin.registerInterface(correlatedIf.Name, vppIfIdx, correlatedIf); err != nil {
				errs = append(errs, err)
			}
			// Calculate whether modification is needed
			if plugin.isIfModified(correlatedIf, vppIf) {
				plugin.Log.Debugf("RESYNC interfaces: modifying correlated interface %v", vppIf.Name)
				if err = plugin.ModifyVPPInterface(correlatedIf, vppIf); err != nil {
					plugin.Log.Errorf("Error while modifying correlated interface: %v", err)
					errs = append(errs, err)
				}
			} else {
				plugin.Log.Debugf("RESYNC interfaces: correlated %v registered without additional changes", vppIf.Name)
			}
		} else {
			// Register interface  with temporary name (will be unregistered during removal)
			if err := plugin.registerInterface(ifTempName, vppIfIdx, vppIf); err != nil {
				errs = append(errs, err)
			}
			// VPP interface cannot be correlated and will be removed
			plugin.Log.Debugf("RESYNC interfaces: removing interface %v", vppIf.Name)
			if err = plugin.deleteVPPInterface(vppIf, vppIfIdx); err != nil {
				plugin.Log.Errorf("Error while removing interface: %v", err)
				errs = append(errs, err)
			}
		}
	}

	// Last step is to configure all new (not-yet-registered) interfaces
	for _, nbIf := range nbIfs {
		// If interface is registered, it was already processed
		_, _, found := plugin.swIfIndexes.LookupIdx(nbIf.Name)
		if !found {
			plugin.Log.Debugf("RESYNC interfaces: configuring new interface %v", nbIf.Name)
			if err := plugin.ConfigureVPPInterface(nbIf); err != nil {
				plugin.Log.Errorf("Error while configuring interface: %v", err)
				errs = append(errs, err)
			}
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface end.")

	return
}

// VerifyVPPConfigPresence dumps VPP interface configuration on the vpp. If there are any interfaces configured (except
// the local0), it returns false (do not interrupt the resto of the resync), otherwise returns true
func (plugin *InterfaceConfigurator) VerifyVPPConfigPresence(nbIfaces []*intf.Interfaces_Interface) bool {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface begin.")
	// notify that the resync should be stopped
	var stop bool

	// Step 0: Dump actual state of the VPP
	vppIfaces, err := vppdump.DumpInterfaces(plugin.Log, plugin.vppCh, plugin.Stopwatch)
	if err != nil {
		return stop
	}

	// The strategy is optimize-cold-start, so look over all dumped VPP interfaces and check for the configured ones
	// (leave out the local0). If there are any other interfaces, return true (resync will continue).
	// If not, return a false flag which cancels the VPP resync operation.
	plugin.Log.Info("optimize-cold-start VPP resync strategy chosen, resolving...")
	if len(vppIfaces) == 0 {
		stop = true
		plugin.Log.Infof("...VPP resync interrupted assuming there is no configuration on the VPP (no interface was found)")
		return stop
	}
	// if interface exists, try to find local0 interface (index 0)
	_, ok := vppIfaces[0]
	// in case local0 is the only interface on the vpp, stop the resync
	if len(vppIfaces) == 1 && ok {
		stop = true
		plugin.Log.Infof("...VPP resync interrupted assuming there is no configuration on the VPP (only local0 was found)")
		return stop
	}
	// otherwise continue normally
	plugin.Log.Infof("... VPP configuration found, continue with VPP resync")

	return stop
}

// ResyncSession writes BFD sessions to the empty VPP
func (plugin *BFDConfigurator) ResyncSession(nbSessions []*bfd.SingleHopBFD_Session) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BFD Session begin.")
	// Calculate and log bfd resync
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Dump all BFD vppSessions
	vppSessions, err := plugin.DumpBfdSessions()
	if err != nil {
		return err
	}

	// Correlate existing BFD sessions from the VPP and NB config, configure new ones
	var wasErr error
	for _, nbSession := range nbSessions {
		// look for configured session
		var found bool
		for _, vppSession := range vppSessions {
			// compare fixed fields
			if nbSession.Interface == vppSession.Interface && nbSession.SourceAddress == vppSession.SourceAddress &&
				nbSession.DestinationAddress == vppSession.DestinationAddress {
				plugin.Log.Debugf("found configured BFD session for interface %v", nbSession.Interface)
				plugin.bfdSessionsIndexes.RegisterName(nbSession.Interface, plugin.BfdIDSeq, nil)
				if err := plugin.ModifyBfdSession(vppSession, nbSession); err != nil {
					plugin.Log.Errorf("BFD resync: error modifying BFD session for interface %v: %v", nbSession.Interface, err)
					wasErr = err
				}
				found = true
			}
		}
		if !found {
			// configure new BFD session
			if err := plugin.ConfigureBfdSession(nbSession); err != nil {
				plugin.Log.Errorf("BFD resync: error creating a new BFD session for interface %v: %v", nbSession.Interface, err)
				wasErr = err
			}
		}
	}

	// Remove old sessions
	for _, vppSession := range vppSessions {
		// remove every not-yet-registered session
		_, _, found := plugin.bfdSessionsIndexes.LookupIdx(vppSession.Interface)
		if !found {
			if err := plugin.DeleteBfdSession(vppSession); err != nil {
				plugin.Log.Errorf("BFD resync: error removing BFD session for interface %v: %v", vppSession.Interface, err)
				wasErr = err
			}
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BFD Session end. ", wasErr)

	return wasErr
}

// ResyncAuthKey writes BFD keys to the empty VPP
func (plugin *BFDConfigurator) ResyncAuthKey(nbKeys []*bfd.SingleHopBFD_Key) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BFD Keys begin.")
	// Calculate and log bfd resync
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// lookup BFD auth keys
	vppKeys, err := plugin.DumpBFDAuthKeys()
	if err != nil {
		return err
	}

	// Correlate existing BFD keys from the VPP and NB config, configure new ones
	var wasErr error
	for _, nbKey := range nbKeys {
		// look for configured keys
		var found bool
		for _, vppKey := range vppKeys {
			// compare key ID
			if nbKey.Id == vppKey.Id {
				plugin.Log.Debugf("found configured BFD auth key with ID %v", nbKey.Id)
				plugin.bfdKeysIndexes.RegisterName(authKeyIdentifier(nbKey.Id), plugin.BfdIDSeq, nil)
				if err := plugin.ModifyBfdAuthKey(vppKey, nbKey); err != nil {
					plugin.Log.Errorf("BFD resync: error modifying BFD auth key with ID %v: %v", nbKey.Id, err)
					wasErr = err
				}
				found = true
			}
		}
		if !found {
			// configure new BFD authentication key
			if err := plugin.ConfigureBfdAuthKey(nbKey); err != nil {
				plugin.Log.Errorf("BFD resync: error creating a new BFD auth key with ID %v: %v", nbKey.Id, err)
				wasErr = err
			}
		}
	}

	// Remove old keys
	for _, vppKey := range vppKeys {
		// remove every not-yet-registered keys
		_, _, found := plugin.bfdKeysIndexes.LookupIdx(authKeyIdentifier(vppKey.Id))
		if !found {
			if err := plugin.DeleteBfdAuthKey(vppKey); err != nil {
				plugin.Log.Errorf("BFD resync: error removing BFD auth key with ID %v: %v", vppKey.Id, err)
				wasErr = err
			}
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BFD Keys end. ", wasErr)

	return wasErr
}

// ResyncEchoFunction writes BFD echo function to the empty VPP
func (plugin *BFDConfigurator) ResyncEchoFunction(echoFunctions []*bfd.SingleHopBFD_EchoFunction) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BFD Echo source begin.")

	if len(echoFunctions) == 0 {
		// Nothing to do here. Currently VPP does not support BFD echo dump so agent does not know
		// whether there is echo function already configured and cannot remove it
		return nil
	}
	// Only one config can be used to set an echo source. If there are multiple configurations,
	// use the first one
	if len(echoFunctions) > 1 {
		plugin.Log.Warn("Multiple configurations of BFD echo function found. Setting up %v as source",
			echoFunctions[0].EchoSourceInterface)
	}
	if err := plugin.ConfigureBfdEchoFunction(echoFunctions[0]); err != nil {
		return err
	}

	return nil
}

// Resync writes stn rule to the the empty VPP
func (plugin *StnConfigurator) Resync(nbStnRules []*stn.StnRule) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC stn rules begin. ")
	// Calculate and log stn rules resync
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Dump existing STN Rules
	vppStnRules, err := plugin.Dump()
	if err != nil {
		return err
	}

	// Correlate configuration, and remove obsolete rules STN rules
	var wasErr error
	for _, vppStnRule := range vppStnRules {
		// Parse parameters
		var vppStnIP net.IP
		var vppStnIPStr string

		if vppStnRule.IsIP4 == 1 {
			vppStnIP = vppStnRule.IPAddress[:4]
		} else {
			vppStnIP = vppStnRule.IPAddress
		}
		vppStnIPStr = vppStnIP.String()

		vppStnIfName, _, found := plugin.SwIfIndexes.LookupName(vppStnRule.SwIfIndex)
		if !found {
			// The rule is attached to non existing interface but it can be removed. If there is a similar
			// rule in NB config, it will be configured (or cached)
			if err := vppcalls.DelStnRule(vppStnRule.SwIfIndex, &vppStnIP, plugin.vppChan, nil); err != nil {
				plugin.Log.Error(err)
				wasErr = err
			}
			plugin.Log.Debugf("RESYNC STN: rule IP: %v ifIdx: %v removed due to missing interface, will be reconfigured if needed",
				vppStnIPStr, vppStnRule.SwIfIndex)
			continue
		}

		// Look for equal rule in NB configuration
		var match bool
		for _, nbStnRule := range nbStnRules {
			if nbStnRule.IpAddress == vppStnIPStr && nbStnRule.Interface == vppStnIfName {
				// Register existing rule
				plugin.indexSTNRule(nbStnRule, false)
				match = true
			}
			plugin.Log.Debugf("RESYNC STN: registered already existing rule %v", nbStnRule.RuleName)
		}

		// If STN rule does not exist, it is obsolete
		if !match {
			if err := vppcalls.DelStnRule(vppStnRule.SwIfIndex, &vppStnIP, plugin.vppChan, nil); err != nil {
				plugin.Log.Error(err)
				wasErr = err
			}
			plugin.Log.Debugf("RESYNC STN: rule IP: %v ifName: %v removed as obsolete", vppStnIPStr, vppStnIfName)
		}
	}

	// Configure missing rules
	for _, nbStnRule := range nbStnRules {
		identifier := StnIdentifier(nbStnRule.Interface)
		_, _, found := plugin.StnAllIndexes.LookupIdx(identifier)
		if !found {
			if err := plugin.Add(nbStnRule); err != nil {
				plugin.Log.Error(err)
				wasErr = err
			}
			plugin.Log.Debugf("RESYNC STN: rule %v added", nbStnRule.RuleName)
		}
	}

	return wasErr
}

// ResyncNatGlobal writes NAT address pool config to the the empty VPP
func (plugin *NatConfigurator) ResyncNatGlobal(nbGlobal *nat.Nat44Global) error {
	plugin.Log.Debug("RESYNC nat global config.")

	vppNatGlobal, err := vppdump.Nat44GlobalConfigDump(plugin.SwIfIndexes, plugin.Log, plugin.vppChan, plugin.Stopwatch)
	if err != nil {
		return fmt.Errorf("failed to dump NAT44 global config: %v", err)
	}

	// Modify will made all the diffs needed (nothing if content is equal)
	return plugin.ModifyNatGlobalConfig(vppNatGlobal, nbGlobal)
}

// ResyncSNat writes NAT static mapping config to the the empty VPP
func (plugin *NatConfigurator) ResyncSNat(sNatConf []*nat.Nat44SNat_SNatConfig) error {
	// todo SNAT not implemented yet, nothing to resync
	return nil
}

// ResyncDNat writes NAT static mapping config to the the empty VPP
func (plugin *NatConfigurator) ResyncDNat(nbDNatConfig []*nat.Nat44DNat_DNatConfig) error {
	plugin.Log.Debug("RESYNC DNAT config.")

	vppDNatCfg, err := vppdump.NAT44DNatDump(plugin.SwIfIndexes, plugin.Log, plugin.vppChan, plugin.Stopwatch)
	if err != nil {
		return fmt.Errorf("failed to dump DNAT config: %v", err)
	}
	if len(vppDNatCfg.DnatConfig) == 0 {
		return nil
	}

	// Correlate with existing config
	for _, nbDNat := range nbDNatConfig {
		for _, vppDNat := range vppDNatCfg.DnatConfig {
			if nbDNat.Label != vppDNat.Label {
				continue
			}
			// Compare all VPP mappings with the NB, register existing ones
			plugin.resolveMappings(nbDNat, &vppDNat.StMappings, &vppDNat.IdMappings)
			// Configure all missing DNAT mappings
			for _, nbMapping := range nbDNat.StMappings {
				mappingIdentifier := getStMappingIdentifier(nbMapping)
				_, _, found := plugin.DNatStMappingIndices.LookupIdx(mappingIdentifier)
				if !found {
					// Configure missing mapping
					if len(nbMapping.LocalIps) == 1 {
						if err := plugin.handleStaticMapping(nbMapping, "", true); err != nil {
							plugin.Log.Errorf("NAT44 resync: failed to configure static mapping: %v", err)
							continue
						}
					} else {
						if err := plugin.handleStaticMappingLb(nbMapping, "", true); err != nil {
							plugin.Log.Errorf("NAT44 resync: failed to configure lb-static mapping: %v", err)
							continue
						}
					}
					// Register new DNAT mapping
					plugin.DNatStMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
					plugin.NatIndexSeq++
					plugin.Log.Debugf("NAT44 resync: new (lb)static mapping %v configured", mappingIdentifier)
				}
			}
			// Configure all missing DNAT identity mappings
			for _, nbIdMapping := range nbDNat.IdMappings {
				mappingIdentifier := getIdMappingIdentifier(nbIdMapping)
				_, _, found := plugin.DNatIdMappingIndices.LookupIdx(mappingIdentifier)
				if !found {
					// Configure missing mapping
					if err := plugin.handleIdentityMapping(nbIdMapping, "", true); err != nil {
						plugin.Log.Errorf("NAT44 resync: failed to configure identity mapping: %v", err)
						continue
					}

					// Register new DNAT mapping
					plugin.DNatIdMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
					plugin.NatIndexSeq++
					plugin.Log.Debugf("NAT44 resync: new identity mapping %v configured", mappingIdentifier)
				}
			}
			// Remove obsolete mappings from DNAT
			for _, vppMapping := range vppDNat.StMappings {
				// Static mapping
				if len(vppMapping.LocalIps) == 1 {

					if err := plugin.handleStaticMapping(vppMapping, "", false); err != nil {
						plugin.Log.Errorf("NAT44 resync: failed to remove static mapping: %v", err)
						continue
					}
				} else {
					// Lb-static mapping
					if err := plugin.handleStaticMappingLb(vppMapping, "", false); err != nil {
						plugin.Log.Errorf("NAT44 resync: failed to remove static mapping: %v", err)
						continue
					}
				}
			}
			for _, vppIdMapping := range vppDNat.IdMappings {
				// Identity mapping
				if err := plugin.handleIdentityMapping(vppIdMapping, "", false); err != nil {
					plugin.Log.Errorf("NAT44 resync: failed to remove identity mapping: %v", err)
					continue
				}
			}
			// At this point, the DNAT is completely configured and can be registered
			plugin.DNatIndices.RegisterName(nbDNat.Label, plugin.NatIndexSeq, nil)
			plugin.NatIndexSeq++
			plugin.Log.Debugf("NAT44 resync: DNAT %v synced", nbDNat.Label)
		}
	}

	// Remove obsolete DNAT configurations which are not registered
	for _, vppDNat := range vppDNatCfg.DnatConfig {
		_, _, found := plugin.DNatIndices.LookupIdx(vppDNat.Label)
		if !found {
			if err := plugin.DeleteDNat(vppDNat); err != nil {
				plugin.Log.Errorf("NAT44 resync: failed to remove obsolete DNAT configuration: %v", vppDNat.Label)
				continue
			}
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC DNAT config done.")

	return nil
}

// Looks for the same mapping in the VPP, register existing ones
func (plugin *NatConfigurator) resolveMappings(nbDNatConfig *nat.Nat44DNat_DNatConfig,
	vppMappings *[]*nat.Nat44DNat_DNatConfig_StaticMappings, vppIdMappings *[]*nat.Nat44DNat_DNatConfig_IdentityMappings) {
	// Iterate over static mappings in NB DNAT config
	for _, nbMapping := range nbDNatConfig.StMappings {
		if len(nbMapping.LocalIps) > 1 {
			// Load balanced
		MappingCompare:
			for vppIndex, vppLbMapping := range *vppMappings {
				// Compare VRF/SNAT fields
				if nbMapping.VrfId != vppLbMapping.VrfId || nbMapping.TwiceNat != vppLbMapping.TwiceNat {
					continue
				}
				// Compare external IP/Port
				if nbMapping.ExternalIP != vppLbMapping.ExternalIP || nbMapping.ExternalPort != vppLbMapping.ExternalPort {
					continue
				}
				// Compare protocol
				if nbMapping.Protocol != vppLbMapping.Protocol {
					continue
				}
				// Compare Local IP/Port and probability addresses
				if len(nbMapping.LocalIps) != len(vppLbMapping.LocalIps) {
					continue
				}
				for _, nbLocal := range nbMapping.LocalIps {
					var found bool
					for _, vppLocal := range vppLbMapping.LocalIps {
						if nbLocal.LocalIP == vppLocal.LocalIP || nbLocal.LocalPort == vppLocal.LocalPort ||
							nbLocal.Probability == vppLocal.Probability {
							found = true
						}
					}
					if !found {
						continue MappingCompare
					}
				}
				// At this point, the NB mapping matched the VPP one, so register it
				mappingIdentifier := getStMappingIdentifier(nbMapping)
				plugin.DNatStMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
				plugin.NatIndexSeq++

				// Remove registered entry from vpp mapping (configurator knows which mappings were registered)
				dMappings := *vppMappings
				*vppMappings = append(dMappings[:vppIndex], dMappings[vppIndex+1:]...)
				plugin.Log.Debugf("NAT44 resync: lb-mapping %v already configured", mappingIdentifier)
			}
		} else {
			// No load balancer
			for vppIndex, vppMapping := range *vppMappings {
				// Compare VRF/SNAT fields
				if nbMapping.VrfId != vppMapping.VrfId || nbMapping.TwiceNat != vppMapping.TwiceNat {
					continue
				}
				// Compare external IP/Port and interface
				if nbMapping.ExternalIP != vppMapping.ExternalIP || nbMapping.ExternalPort != vppMapping.ExternalPort {
					continue
				}
				// Compare external interface
				if nbMapping.ExternalInterface != vppMapping.ExternalInterface {
					continue
				}
				// Compare protocol
				if nbMapping.Protocol != vppMapping.Protocol {
					continue
				}
				// Compare Local IP/Port and probability addresses (there is only one local IP address in this case)
				if len(nbMapping.LocalIps) != 1 || len(vppMapping.LocalIps) != 1 {
					plugin.Log.Warnf("NAT44 resync: mapping without load balancer contains more than 1 local IP address")
					continue
				}
				nbLocal := nbMapping.LocalIps[0]
				vppLocal := vppMapping.LocalIps[0]
				if nbLocal.LocalIP != vppLocal.LocalIP || nbLocal.LocalPort != vppLocal.LocalPort ||
					nbLocal.Probability != vppLocal.Probability {
					continue
				}

				// At this point, the NB mapping matched the VPP one, so register it
				mappingIdentifier := getStMappingIdentifier(nbMapping)
				plugin.DNatStMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
				plugin.NatIndexSeq++

				// Remove registered entry from vpp mapping (so configurator knows which mappings were registered)
				dMappings := *vppMappings
				*vppMappings = append(dMappings[:vppIndex], dMappings[vppIndex+1:]...)
				plugin.Log.Debugf("NAT44 resync: mapping %v already configured", mappingIdentifier)
			}
		}
	}
	// Iterate over identity mappings in NB DNAT config
	for _, nbIdMapping := range nbDNatConfig.IdMappings {
		for vppIdIndex, vppIdMapping := range *vppIdMappings {
			// Compare VRF and address interface
			if nbIdMapping.VrfId != vppIdMapping.VrfId || nbIdMapping.AddressedInterface != vppIdMapping.AddressedInterface {
				continue
			}
			// Compare IP and port values
			if nbIdMapping.IpAddress != vppIdMapping.IpAddress || nbIdMapping.Port != vppIdMapping.Port {
				continue
			}
			// Compare protocol
			if nbIdMapping.Protocol != vppIdMapping.Protocol {
				continue
			}

			// At this point, the NB mapping matched the VPP one, so register it
			mappingIdentifier := getIdMappingIdentifier(nbIdMapping)
			plugin.DNatIdMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
			plugin.NatIndexSeq++

			// Remove registered entry from vpp mapping (configurator knows which mappings were registered)
			dIdMappings := *vppIdMappings
			*vppIdMappings = append(dIdMappings[:vppIdIndex], dIdMappings[vppIdIndex+1:]...)
			plugin.Log.Debugf("NAT44 resync: identity mapping %v already configured", mappingIdentifier)
		}
	}
}

// Correlate interfaces according to MAC address, interface addresses
func (plugin *InterfaceConfigurator) correlateInterface(vppIf, nbIf *intf.Interfaces_Interface) *intf.Interfaces_Interface {
	// Correlate MAC address
	if nbIf.PhysAddress != "" {
		if nbIf.PhysAddress == vppIf.PhysAddress {
			return nbIf
		}
	}
	// Correlate IP addresses
	if len(nbIf.IpAddresses) == len(vppIf.IpAddresses) {
		ipMatch := true

	ipComparison:
		for _, nbIP := range nbIf.IpAddresses {
			var ipFound bool
			for _, vppIP := range vppIf.IpAddresses {
				pNbIP, nbIPNet, err := net.ParseCIDR(nbIP)
				if err != nil {
					plugin.Log.Error(err)
					continue
				}
				pVppIP, vppIPNet, err := net.ParseCIDR(vppIP)
				if err != nil {
					plugin.Log.Error(err)
					continue
				}
				if nbIPNet.Mask.String() == vppIPNet.Mask.String() && bytes.Compare(pNbIP, pVppIP) == 0 {
					ipFound = true
					break
				}
			}
			if !ipFound {
				// Break comparison if there is mismatch
				ipMatch = false
				break ipComparison
			}
		}

		if ipMatch {
			return nbIf
		}
	}
	// todo correlate also unnumbered interfaces if possible

	// Otherwise there is no match between interfaces
	return nil
}

// Compares two interfaces. If there is any difference, returns true, false otherwise
func (plugin *InterfaceConfigurator) isIfModified(nbIf, vppIf *intf.Interfaces_Interface) bool {
	plugin.Log.Debugf("Interface RESYNC comparison started for interface %s", nbIf.Name)
	// Type
	if nbIf.Type != vppIf.Type {
		plugin.Log.Debugf("Interface RESYNC comparison: type changed (NB: %v, VPP: %v)",
			nbIf.Type, vppIf.Type)
		return true
	}
	// Enabled
	if nbIf.Enabled != vppIf.Enabled {
		plugin.Log.Debugf("Interface RESYNC comparison: enabled state changed (NB: %t, VPP: %t)",
			nbIf.Enabled, vppIf.Enabled)
		return true
	}
	// VRF
	if nbIf.Vrf != vppIf.Vrf {
		plugin.Log.Debugf("Interface RESYNC comparison: VRF changed (NB: %d, VPP: %d)",
			nbIf.Vrf, vppIf.Vrf)
		return true
	}
	// Container IP address
	if nbIf.ContainerIpAddress != vppIf.ContainerIpAddress {
		plugin.Log.Debugf("Interface RESYNC comparison: container IP changed (NB: %s, VPP: %s)",
			nbIf.ContainerIpAddress, vppIf.ContainerIpAddress)
		return true
	}
	// DHCP setup
	if nbIf.SetDhcpClient != vppIf.SetDhcpClient {
		plugin.Log.Debugf("Interface RESYNC comparison: DHCP setup changed (NB: %t, VPP: %t)",
			nbIf.SetDhcpClient, vppIf.SetDhcpClient)
		return true
	}
	//  MTU value
	if nbIf.Mtu != vppIf.Mtu {
		plugin.Log.Debugf("Interface RESYNC comparison: MTU changed (NB: %d, VPP: %d)",
			nbIf.Mtu, vppIf.Mtu)
		return true
	}
	// MAC address (compare only if it is set in the NB configuration)
	nbMac := strings.ToUpper(nbIf.PhysAddress)
	vppMac := strings.ToUpper(vppIf.PhysAddress)
	if nbMac != "" && nbMac != vppMac {
		plugin.Log.Debugf("Interface RESYNC comparison: Physical address changed (NB: %s, VPP: %s)", nbMac, vppMac)
		return true
	}
	// Unnumbered settings. If interface is unnumbered, do not compare ip addresses.
	// todo unnumbered data cannot be dumped
	if nbIf.Unnumbered != nil {
		plugin.Log.Debugf("RESYNC interfaces: interface %s is unnumbered, result of the comparison may not be correct", nbIf.Name)
		vppIf.IpAddresses = nil
	} else {
		// Remove IPv6 link local addresses (default values)
		for ipIdx, ipAddress := range vppIf.IpAddresses {
			if strings.HasPrefix(ipAddress, "fe80") {
				vppIf.IpAddresses = append(vppIf.IpAddresses[:ipIdx], vppIf.IpAddresses[ipIdx+1:]...)
			}
		}
		// Compare IP address count
		if len(nbIf.IpAddresses) != len(vppIf.IpAddresses) {
			plugin.Log.Debugf("Interface RESYNC comparison: IP address count changed (NB: %d, VPP: %d)",
				len(nbIf.IpAddresses), len(vppIf.IpAddresses))
			return true
		}
		// Compare every single IP address. If equal, every address should have identical counterpart
		for _, nbIP := range nbIf.IpAddresses {
			var ipFound bool
			for _, vppIP := range vppIf.IpAddresses {
				pNbIP, nbIPNet, err := net.ParseCIDR(nbIP)
				if err != nil {
					plugin.Log.Error(err)
					continue
				}
				pVppIP, vppIPNet, err := net.ParseCIDR(vppIP)
				if err != nil {
					plugin.Log.Error(err)
					continue
				}
				if nbIPNet.Mask.String() == vppIPNet.Mask.String() && bytes.Compare(pNbIP, pVppIP) == 0 {
					ipFound = true
					break
				}
			}
			if !ipFound {
				plugin.Log.Debugf("Interface RESYNC comparison: VPP interface %s does not contain IP %s", nbIf.Name, nbIP)
				return true
			}
		}
	}
	// RxMode settings
	if nbIf.RxModeSettings == nil && vppIf.RxModeSettings != nil || nbIf.RxModeSettings != nil && vppIf.RxModeSettings == nil {
		plugin.Log.Debugf("Interface RESYNC comparison: RxModeSettings changed (NB: %v, VPP: %v)",
			nbIf.RxModeSettings, vppIf.RxModeSettings)
		return true
	}
	if nbIf.RxModeSettings != nil && vppIf.RxModeSettings != nil {
		// RxMode
		if nbIf.RxModeSettings.RxMode != vppIf.RxModeSettings.RxMode {
			plugin.Log.Debugf("Interface RESYNC comparison: RxMode changed (NB: %v, VPP: %v)",
				nbIf.RxModeSettings.RxMode, vppIf.RxModeSettings.RxMode)
			return true

		}
		// QueueID
		if nbIf.RxModeSettings.QueueID != vppIf.RxModeSettings.QueueID {
			plugin.Log.Debugf("Interface RESYNC comparison: QueueID changed (NB: %d, VPP: %d)",
				nbIf.RxModeSettings.QueueID, vppIf.RxModeSettings.QueueID)
			return true

		}
		// QueueIDValid
		if nbIf.RxModeSettings.QueueIDValid != vppIf.RxModeSettings.QueueIDValid {
			plugin.Log.Debugf("Interface RESYNC comparison: QueueIDValid changed (NB: %d, VPP: %d)",
				nbIf.RxModeSettings.QueueIDValid, vppIf.RxModeSettings.QueueIDValid)
			return true

		}
	}

	switch nbIf.Type {
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		if nbIf.Afpacket == nil && vppIf.Afpacket != nil || nbIf.Afpacket != nil && vppIf.Afpacket == nil {
			plugin.Log.Debugf("Interface RESYNC comparison: AF-packet setup changed (NB: %v, VPP: %v)",
				nbIf.Afpacket, vppIf.Afpacket)
			return true
		}
		if nbIf.Afpacket != nil && vppIf.Afpacket != nil {
			// AF-packet host name
			if nbIf.Afpacket.HostIfName != vppIf.Afpacket.HostIfName {
				plugin.Log.Debugf("Interface RESYNC comparison: AF-packet host name changed (NB: %s, VPP: %s)",
					nbIf.Afpacket.HostIfName, vppIf.Afpacket.HostIfName)
				return true
			}
		}
	case intf.InterfaceType_MEMORY_INTERFACE:
		if nbIf.Memif == nil && vppIf.Memif != nil || nbIf.Memif != nil && vppIf.Memif == nil {
			plugin.Log.Debugf("Interface RESYNC comparison: memif setup changed (NB: %v, VPP: %v)",
				nbIf.Memif, vppIf.Memif)
			return true
		}
		if nbIf.Memif != nil && vppIf.Memif != nil {
			// Memif ID
			if nbIf.Memif.Id != vppIf.Memif.Id {
				plugin.Log.Debugf("Interface RESYNC comparison: memif ID changed (NB: %d, VPP: %d)",
					nbIf.Memif.Id, vppIf.Memif.Id)
				return true
			}

			// Memif socket
			if nbIf.Memif.SocketFilename != vppIf.Memif.SocketFilename {
				plugin.Log.Debugf("Interface RESYNC comparison: memif socket filename changed (NB: %s, VPP: %s)",
					nbIf.Memif.SocketFilename, vppIf.Memif.SocketFilename)
				return true
			}
			// Master
			if nbIf.Memif.Master != vppIf.Memif.Master {
				plugin.Log.Debugf("Interface RESYNC comparison: memif master setup changed (NB: %t, VPP: %t)",
					nbIf.Memif.Master, vppIf.Memif.Master)
				return true
			}
			// Mode
			if nbIf.Memif.Mode != vppIf.Memif.Mode {
				plugin.Log.Debugf("Interface RESYNC comparison: memif mode setup changed (NB: %v, VPP: %v)",
					nbIf.Memif.Mode, vppIf.Memif.Mode)
				return true
			}
			// Rx queues
			if nbIf.Memif.RxQueues != vppIf.Memif.RxQueues {
				plugin.Log.Debugf("Interface RESYNC comparison: RxQueues changed (NB: %d, VPP: %d)",
					nbIf.Memif.RxQueues, vppIf.Memif.RxQueues)
				return true
			}
			// Tx queues
			if nbIf.Memif.TxQueues != vppIf.Memif.TxQueues {
				plugin.Log.Debugf("Interface RESYNC comparison: TxQueues changed (NB: %d, VPP: %d)",
					nbIf.Memif.TxQueues, vppIf.Memif.TxQueues)
				return true
			}
			// todo secret, buffer size and ring size is not compared. VPP always returns 0 for buffer size
			// and 1 for ring size. Secret cannot be dumped at all.
		}
	case intf.InterfaceType_TAP_INTERFACE:
		if nbIf.Tap == nil && vppIf.Tap != nil || nbIf.Tap != nil && vppIf.Tap == nil {
			plugin.Log.Debugf("Interface RESYNC comparison: tap setup changed (NB: %v, VPP: %v)",
				nbIf.Tap, vppIf.Tap)
			return true
		}
		if nbIf.Tap != nil && vppIf.Tap != nil {
			// Tap version
			if nbIf.Tap.Version == 2 && nbIf.Tap.Version != vppIf.Tap.Version {
				plugin.Log.Debugf("Interface RESYNC comparison: tap version changed (NB: %d, VPP: %d)",
					nbIf.Tap.Version, vppIf.Tap.Version)
				return true
			}
			// Namespace and host name
			if nbIf.Tap.Namespace != vppIf.Tap.Namespace {
				plugin.Log.Debugf("Interface RESYNC comparison: tap namespace changed (NB: %s, VPP: %s)",
					nbIf.Tap.Namespace, vppIf.Tap.Namespace)
				return true
			}
			// Namespace and host name
			if nbIf.Tap.HostIfName != vppIf.Tap.HostIfName {
				plugin.Log.Debugf("Interface RESYNC comparison: tap host name changed (NB: %s, VPP: %s)",
					nbIf.Tap.HostIfName, vppIf.Tap.HostIfName)
				return true
			}
			// Rx ring size
			if nbIf.Tap.RxRingSize != nbIf.Tap.RxRingSize {
				plugin.Log.Debugf("Interface RESYNC comparison: tap Rx ring size changed (NB: %d, VPP: %d)",
					nbIf.Tap.RxRingSize, vppIf.Tap.RxRingSize)
				return true
			}
			// Tx ring size
			if nbIf.Tap.TxRingSize != nbIf.Tap.TxRingSize {
				plugin.Log.Debugf("Interface RESYNC comparison: tap Tx ring size changed (NB: %d, VPP: %d)",
					nbIf.Tap.TxRingSize, vppIf.Tap.TxRingSize)
				return true
			}
		}
	case intf.InterfaceType_VXLAN_TUNNEL:
		if nbIf.Vxlan == nil && vppIf.Vxlan != nil || nbIf.Vxlan != nil && vppIf.Vxlan == nil {
			plugin.Log.Debugf("Interface RESYNC comparison: VxLAN setup changed (NB: %v, VPP: %v)",
				nbIf.Vxlan, vppIf.Vxlan)
			return true
		}
		if nbIf.Vxlan != nil && vppIf.Vxlan != nil {
			// VxLAN Vni
			if nbIf.Vxlan.Vni != vppIf.Vxlan.Vni {
				plugin.Log.Debugf("Interface RESYNC comparison: VxLAN Vni changed (NB: %d, VPP: %d)",
					nbIf.Vxlan.Vni, vppIf.Vxlan.Vni)
				return true
			}
			// VxLAN Src Address
			if nbIf.Vxlan.SrcAddress != vppIf.Vxlan.SrcAddress {
				plugin.Log.Debugf("Interface RESYNC comparison: VxLAN src address changed (NB: %s, VPP: %s)",
					nbIf.Vxlan.SrcAddress, vppIf.Vxlan.SrcAddress)
				return true
			}
			// VxLAN Dst Address
			if nbIf.Vxlan.DstAddress != vppIf.Vxlan.DstAddress {
				plugin.Log.Debugf("Interface RESYNC comparison: VxLAN dst address changed (NB: %s, VPP: %s)",
					nbIf.Vxlan.DstAddress, vppIf.Vxlan.DstAddress)
				return true
			}
		}
	}

	// At last, return false if interfaces are equal
	return false
}

// Register interface to mapping and add tag/index to the VPP
func (plugin *InterfaceConfigurator) registerInterface(ifName string, ifIdx uint32, ifData *intf.Interfaces_Interface) error {
	plugin.swIfIndexes.RegisterName(ifName, ifIdx, ifData)
	if err := vppcalls.SetInterfaceTag(ifName, ifIdx, plugin.vppCh, plugin.Stopwatch); err != nil {
		return fmt.Errorf("error while adding interface tag %s, index %d: %v", ifName, ifIdx, err)
	}
	// Add AF-packet type interface to local cache
	if ifData.Type == intf.InterfaceType_AF_PACKET_INTERFACE {
		if plugin.Linux != nil && plugin.afPacketConfigurator != nil && ifData.Afpacket != nil {
			// Interface is already present on the VPP so it cannot be marked as pending.
			plugin.afPacketConfigurator.addToCache(ifData, false)
		}
	}
	plugin.Log.Debugf("RESYNC interfaces: registered interface %s (index %d)", ifName, ifIdx)
	return nil
}
