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

	_ "github.com/ligato/vpp-agent/plugins/vpp/binapi/nat"
	"github.com/ligato/vpp-agent/plugins/vpp/model/bfd"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/model/nat"
	"github.com/ligato/vpp-agent/plugins/vpp/model/stn"
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
	plugin.log.WithField("cfg", plugin).Debugf("RESYNC Interface begin for %v", nbIfs)
	// Calculate and log interface resync
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Re-initialize cache
	if err := plugin.clearMapping(); err != nil {
		return []error{err}
	}
	plugin.afPacketConfigurator.clearMapping()

	var err error
	if plugin.memifScCache, err = plugin.ifHandler.DumpMemifSocketDetails(); err != nil {
		return []error{err}
	}

	// Dump current state of the VPP interfaces
	vppIfs, err := plugin.ifHandler.DumpInterfaces()
	if err != nil {
		return []error{err}
	}

	// Cache for untagged interfaces. All un-named interfaces have to be correlated
	unnamedVppIfs := make(map[uint32]*intf.Interfaces_Interface)

	// Iterate over VPP interfaces and try to correlate NB config
	for vppIfIdx, vppIf := range vppIfs {
		if vppIfIdx == 0 {
			// Register local0 interface with zero index
			if err := plugin.registerInterface(vppIf.Meta.InternalName, vppIfIdx, vppIf.Interface); err != nil {
				errs = append(errs, err)
			}
			continue
		}
		if vppIf.Interface.Name == "" {
			// If interface has no name, it is stored as unnamed and resolved later
			plugin.log.Debugf("RESYNC interfaces: interface %v has no name (tag)", vppIfIdx)
			unnamedVppIfs[vppIfIdx] = vppIf.Interface
			continue
		}
		var correlated bool
		for _, nbIf := range nbIfs {
			if vppIf.Interface.Name == nbIf.Name {
				correlated = true
				// Register interface to mapping and VPP tag/index
				if err := plugin.registerInterface(vppIf.Interface.Name, vppIfIdx, nbIf); err != nil {
					errs = append(errs, err)
				}
				// Calculate whether modification is needed
				if plugin.isIfModified(nbIf, vppIf.Interface) {
					plugin.log.Debugf("RESYNC interfaces: modifying interface %v", vppIf.Interface.Name)
					if err = plugin.ModifyVPPInterface(nbIf, vppIf.Interface); err != nil {
						plugin.log.Errorf("Error while modifying interface: %v", err)
						errs = append(errs, err)
					}
				} else {
					plugin.log.Debugf("RESYNC interfaces: %v registered without additional changes", vppIf.Interface.Name)
				}
				break
			}
		}
		if !correlated {
			// Register interface before removal (to keep state consistent)
			if err := plugin.registerInterface(vppIf.Interface.Name, vppIfIdx, vppIf.Interface); err != nil {
				errs = append(errs, err)
			}
			// VPP interface is obsolete and will be removed (un-configured if physical device)
			plugin.log.Debugf("RESYNC interfaces: removing obsolete interface %v", vppIf.Interface.Name)
			if err = plugin.deleteVPPInterface(vppIf.Interface, vppIfIdx); err != nil {
				plugin.log.Errorf("Error while removing interface: %v", err)
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
				plugin.log.Debugf("RESYNC interfaces: modifying correlated interface %v", vppIf.Name)
				if err = plugin.ModifyVPPInterface(correlatedIf, vppIf); err != nil {
					plugin.log.Errorf("Error while modifying correlated interface: %v", err)
					errs = append(errs, err)
				}
			} else {
				plugin.log.Debugf("RESYNC interfaces: correlated %v registered without additional changes", vppIf.Name)
			}
		} else {
			// Register interface  with temporary name (will be unregistered during removal)
			if err := plugin.registerInterface(ifTempName, vppIfIdx, vppIf); err != nil {
				errs = append(errs, err)
			}
			// VPP interface cannot be correlated and will be removed
			plugin.log.Debugf("RESYNC interfaces: removing interface %v", vppIf.Name)
			if err = plugin.deleteVPPInterface(vppIf, vppIfIdx); err != nil {
				plugin.log.Errorf("Error while removing interface: %v", err)
				errs = append(errs, err)
			}
		}
	}

	// Last step is to configure all new (not-yet-registered) interfaces
	for _, nbIf := range nbIfs {
		// If interface is registered, it was already processed
		_, _, found := plugin.swIfIndexes.LookupIdx(nbIf.Name)
		if !found {
			plugin.log.Debugf("RESYNC interfaces: configuring new interface %v", nbIf.Name)
			if err := plugin.ConfigureVPPInterface(nbIf); err != nil {
				plugin.log.Errorf("Error while configuring interface: %v", err)
				errs = append(errs, err)
			}
		}
	}

	// update the interfaces state data in memory
	plugin.PropagateIfDetailsToStatus()

	plugin.log.WithField("cfg", plugin).Debug("RESYNC Interface end.")

	return
}

// VerifyVPPConfigPresence dumps VPP interface configuration on the vpp. If there are any interfaces configured (except
// the local0), it returns false (do not interrupt the resto of the resync), otherwise returns true
func (plugin *InterfaceConfigurator) VerifyVPPConfigPresence(nbIfaces []*intf.Interfaces_Interface) bool {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC Interface begin.")
	// notify that the resync should be stopped
	var stop bool

	// Step 0: Dump actual state of the VPP
	vppIfaces, err := plugin.ifHandler.DumpInterfaces()
	if err != nil {
		return stop
	}

	plugin.log.Infof("dumped %d interfaces", len(vppIfaces))

	// The strategy is optimize-cold-start, so look over all dumped VPP interfaces and check for the configured ones
	// (leave out the local0). If there are any other interfaces, return true (resync will continue).
	// If not, return a false flag which cancels the VPP resync operation.
	plugin.log.Info("optimize-cold-start VPP resync strategy chosen, resolving...")
	if len(vppIfaces) == 0 {
		stop = true
		plugin.log.Infof("...VPP resync interrupted assuming there is no configuration on the VPP (no interface was found)")
		return stop
	}
	// if interface exists, try to find local0 interface (index 0)
	_, ok := vppIfaces[0]
	// in case local0 is the only interface on the vpp, stop the resync
	if len(vppIfaces) == 1 && ok {
		stop = true
		plugin.log.Infof("...VPP resync interrupted assuming there is no configuration on the VPP (only local0 was found)")
		return stop
	}
	// otherwise continue normally
	plugin.log.Infof("... VPP configuration found, continue with VPP resync")

	return stop
}

// ResyncSession writes BFD sessions to the empty VPP
func (plugin *BFDConfigurator) ResyncSession(nbSessions []*bfd.SingleHopBFD_Session) error {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC BFD Session begin.")
	// Calculate and log bfd resync
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Re-initialize cache
	plugin.clearMapping()

	// Dump all BFD vppSessions
	vppBfdSessions, err := plugin.bfdHandler.DumpBfdSessions()
	if err != nil {
		return err
	}

	// Correlate existing BFD sessions from the VPP and NB config, configure new ones
	var wasErr error
	for _, nbSession := range nbSessions {
		// look for configured session
		var found bool
		for _, vppSession := range vppBfdSessions.Session {
			// compare fixed fields
			if nbSession.Interface == vppSession.Interface && nbSession.SourceAddress == vppSession.SourceAddress &&
				nbSession.DestinationAddress == vppSession.DestinationAddress {
				plugin.log.Debugf("found configured BFD session for interface %v", nbSession.Interface)
				plugin.sessionsIndexes.RegisterName(nbSession.Interface, plugin.bfdIDSeq, nil)
				if err := plugin.ModifyBfdSession(vppSession, nbSession); err != nil {
					plugin.log.Errorf("BFD resync: error modifying BFD session for interface %v: %v", nbSession.Interface, err)
					wasErr = err
				}
				found = true
			}
		}
		if !found {
			// configure new BFD session
			if err := plugin.ConfigureBfdSession(nbSession); err != nil {
				plugin.log.Errorf("BFD resync: error creating a new BFD session for interface %v: %v", nbSession.Interface, err)
				wasErr = err
			}
		}
	}

	// Remove old sessions
	for _, vppSession := range vppBfdSessions.Session {
		// remove every not-yet-registered session
		_, _, found := plugin.sessionsIndexes.LookupIdx(vppSession.Interface)
		if !found {
			if err := plugin.DeleteBfdSession(vppSession); err != nil {
				plugin.log.Errorf("BFD resync: error removing BFD session for interface %v: %v", vppSession.Interface, err)
				wasErr = err
			}
		}
	}

	plugin.log.WithField("cfg", plugin).Debug("RESYNC BFD Session end. ", wasErr)

	return wasErr
}

// ResyncAuthKey writes BFD keys to the empty VPP
func (plugin *BFDConfigurator) ResyncAuthKey(nbKeys []*bfd.SingleHopBFD_Key) error {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC BFD Keys begin.")
	// Calculate and log bfd resync
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// lookup BFD auth keys
	vppBfdKeys, err := plugin.bfdHandler.DumpBfdAuthKeys()
	if err != nil {
		return err
	}

	// Correlate existing BFD keys from the VPP and NB config, configure new ones
	var wasErr error
	for _, nbKey := range nbKeys {
		// look for configured keys
		var found bool
		for _, vppKey := range vppBfdKeys.AuthKeys {
			// compare key ID
			if nbKey.Id == vppKey.Id {
				plugin.log.Debugf("found configured BFD auth key with ID %v", nbKey.Id)
				plugin.keysIndexes.RegisterName(AuthKeyIdentifier(nbKey.Id), plugin.bfdIDSeq, nil)
				if err := plugin.ModifyBfdAuthKey(vppKey, nbKey); err != nil {
					plugin.log.Errorf("BFD resync: error modifying BFD auth key with ID %v: %v", nbKey.Id, err)
					wasErr = err
				}
				found = true
			}
		}
		if !found {
			// configure new BFD authentication key
			if err := plugin.ConfigureBfdAuthKey(nbKey); err != nil {
				plugin.log.Errorf("BFD resync: error creating a new BFD auth key with ID %v: %v", nbKey.Id, err)
				wasErr = err
			}
		}
	}

	// Remove old keys
	for _, vppKey := range vppBfdKeys.AuthKeys {
		// remove every not-yet-registered keys
		_, _, found := plugin.keysIndexes.LookupIdx(AuthKeyIdentifier(vppKey.Id))
		if !found {
			if err := plugin.DeleteBfdAuthKey(vppKey); err != nil {
				plugin.log.Errorf("BFD resync: error removing BFD auth key with ID %v: %v", vppKey.Id, err)
				wasErr = err
			}
		}
	}

	plugin.log.WithField("cfg", plugin).Debug("RESYNC BFD Keys end. ", wasErr)

	return wasErr
}

// ResyncEchoFunction writes BFD echo function to the empty VPP
func (plugin *BFDConfigurator) ResyncEchoFunction(echoFunctions []*bfd.SingleHopBFD_EchoFunction) error {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC BFD Echo source begin.")

	if len(echoFunctions) == 0 {
		// Nothing to do here. Currently VPP does not support BFD echo dump so agent does not know
		// whether there is echo function already configured and cannot remove it
		return nil
	}
	// Only one config can be used to set an echo source. If there are multiple configurations,
	// use the first one
	if len(echoFunctions) > 1 {
		plugin.log.Warn("Multiple configurations of BFD echo function found. Setting up %v as source",
			echoFunctions[0].EchoSourceInterface)
	}
	if err := plugin.ConfigureBfdEchoFunction(echoFunctions[0]); err != nil {
		return err
	}

	return nil
}

// Resync writes stn rule to the the empty VPP
func (plugin *StnConfigurator) Resync(nbStnRules []*stn.STN_Rule) error {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC stn rules begin. ")
	// Calculate and log stn rules resync
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Re-initialize cache
	plugin.clearMapping()

	// Dump existing STN Rules
	vppStnDetails, err := plugin.Dump()
	if err != nil {
		return err
	}

	// Correlate configuration, and remove obsolete rules STN rules
	var wasErr error
	for _, vppStnRule := range vppStnDetails.Rules {
		// Parse parameters
		var vppStnIP net.IP
		var vppStnIPStr string

		vppStnIfIdx, _, found := plugin.ifIndexes.LookupIdx(vppStnRule.Interface)
		if !found {
			// The rule is attached to non existing interface but it can be removed. If there is a similar
			// rule in NB config, it will be configured (or cached)
			if err := plugin.stnHandler.DelStnRule(vppStnIfIdx, &vppStnIP); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
			plugin.log.Debugf("RESYNC STN: rule IP: %v ifIdx: %v removed due to missing interface, will be reconfigured if needed",
				vppStnIPStr, vppStnIfIdx)
			continue
		}

		// Look for equal rule in NB configuration
		var match bool
		for _, nbStnRule := range nbStnRules {
			if nbStnRule.IpAddress == vppStnIPStr && nbStnRule.Interface == vppStnRule.Interface {
				// Register existing rule
				plugin.indexSTNRule(nbStnRule, false)
				match = true
			}
			plugin.log.Debugf("RESYNC STN: registered already existing rule %v", nbStnRule.RuleName)
		}

		// If STN rule does not exist, it is obsolete
		if !match {
			if err := plugin.stnHandler.DelStnRule(vppStnIfIdx, &vppStnIP); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
			plugin.log.Debugf("RESYNC STN: rule IP: %v ifName: %v removed as obsolete", vppStnIPStr, vppStnRule.Interface)
		}
	}

	// Configure missing rules
	for _, nbStnRule := range nbStnRules {
		identifier := StnIdentifier(nbStnRule.Interface)
		_, _, found := plugin.allIndexes.LookupIdx(identifier)
		if !found {
			if err := plugin.Add(nbStnRule); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
			plugin.log.Debugf("RESYNC STN: rule %v added", nbStnRule.RuleName)
		}
	}

	return wasErr
}

// ResyncNatGlobal writes NAT address pool config to the the empty VPP
func (plugin *NatConfigurator) ResyncNatGlobal(nbGlobal *nat.Nat44Global) error {
	plugin.log.Debug("RESYNC nat global config.")

	// Re-initialize cache
	plugin.clearMapping()

	vppNatGlobal, err := plugin.natHandler.Nat44GlobalConfigDump()
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
	plugin.log.Debug("RESYNC DNAT config.")

	vppDNatCfg, err := plugin.natHandler.NAT44DNatDump()
	if err != nil {
		return fmt.Errorf("failed to dump DNAT config: %v", err)
	}
	if len(vppDNatCfg.DnatConfigs) == 0 {
		return nil
	}

	// Correlate with existing config
	for _, nbDNat := range nbDNatConfig {
		for _, vppDNat := range vppDNatCfg.DnatConfigs {
			if nbDNat.Label != vppDNat.Label {
				continue
			}
			// Compare all VPP mappings with the NB, register existing ones
			plugin.resolveMappings(nbDNat, &vppDNat.StMappings, &vppDNat.IdMappings)
			// Configure all missing DNAT mappings
			for _, nbMapping := range nbDNat.StMappings {
				mappingIdentifier := GetStMappingIdentifier(nbMapping)
				_, _, found := plugin.dNatStMappingIndexes.LookupIdx(mappingIdentifier)
				if !found {
					// Configure missing mapping
					if len(nbMapping.LocalIps) == 1 {
						if err := plugin.handleStaticMapping(nbMapping, "", true); err != nil {
							plugin.log.Errorf("NAT44 resync: failed to configure static mapping: %v", err)
							continue
						}
					} else {
						if err := plugin.handleStaticMappingLb(nbMapping, "", true); err != nil {
							plugin.log.Errorf("NAT44 resync: failed to configure lb-static mapping: %v", err)
							continue
						}
					}
					// Register new DNAT mapping
					plugin.dNatStMappingIndexes.RegisterName(mappingIdentifier, plugin.natIndexSeq, nil)
					plugin.natIndexSeq++
					plugin.log.Debugf("NAT44 resync: new (lb)static mapping %v configured", mappingIdentifier)
				}
			}
			// Configure all missing DNAT identity mappings
			for _, nbIdMapping := range nbDNat.IdMappings {
				mappingIdentifier := GetIdMappingIdentifier(nbIdMapping)
				_, _, found := plugin.dNatIdMappingIndexes.LookupIdx(mappingIdentifier)
				if !found {
					// Configure missing mapping
					if err := plugin.handleIdentityMapping(nbIdMapping, "", true); err != nil {
						plugin.log.Errorf("NAT44 resync: failed to configure identity mapping: %v", err)
						continue
					}

					// Register new DNAT mapping
					plugin.dNatIdMappingIndexes.RegisterName(mappingIdentifier, plugin.natIndexSeq, nil)
					plugin.natIndexSeq++
					plugin.log.Debugf("NAT44 resync: new identity mapping %v configured", mappingIdentifier)
				}
			}
			// Remove obsolete mappings from DNAT
			for _, vppMapping := range vppDNat.StMappings {
				// Static mapping
				if len(vppMapping.LocalIps) == 1 {

					if err := plugin.handleStaticMapping(vppMapping, "", false); err != nil {
						plugin.log.Errorf("NAT44 resync: failed to remove static mapping: %v", err)
						continue
					}
				} else {
					// Lb-static mapping
					if err := plugin.handleStaticMappingLb(vppMapping, "", false); err != nil {
						plugin.log.Errorf("NAT44 resync: failed to remove static mapping: %v", err)
						continue
					}
				}
			}
			for _, vppIdMapping := range vppDNat.IdMappings {
				// Identity mapping
				if err := plugin.handleIdentityMapping(vppIdMapping, "", false); err != nil {
					plugin.log.Errorf("NAT44 resync: failed to remove identity mapping: %v", err)
					continue
				}
			}
			// At this point, the DNAT is completely configured and can be registered
			plugin.dNatIndexes.RegisterName(nbDNat.Label, plugin.natIndexSeq, nil)
			plugin.natIndexSeq++
			plugin.log.Debugf("NAT44 resync: DNAT %v synced", nbDNat.Label)
		}
	}

	// Remove obsolete DNAT configurations which are not registered
	for _, vppDNat := range vppDNatCfg.DnatConfigs {
		_, _, found := plugin.dNatIndexes.LookupIdx(vppDNat.Label)
		if !found {
			if err := plugin.DeleteDNat(vppDNat); err != nil {
				plugin.log.Errorf("NAT44 resync: failed to remove obsolete DNAT configuration: %v", vppDNat.Label)
				continue
			}
		}
	}

	plugin.log.WithField("cfg", plugin).Debug("RESYNC DNAT config done.")

	return nil
}

// Looks for the same mapping in the VPP, register existing ones
func (plugin *NatConfigurator) resolveMappings(nbDNatConfig *nat.Nat44DNat_DNatConfig,
	vppMappings *[]*nat.Nat44DNat_DNatConfig_StaticMapping, vppIdMappings *[]*nat.Nat44DNat_DNatConfig_IdentityMapping) {
	// Iterate over static mappings in NB DNAT config
	for _, nbMapping := range nbDNatConfig.StMappings {
		if len(nbMapping.LocalIps) > 1 {
			// Load balanced
		MappingCompare:
			for vppIndex, vppLbMapping := range *vppMappings {
				// Compare VRF/SNAT fields
				if nbMapping.TwiceNat != vppLbMapping.TwiceNat {
					continue
				}
				// Compare external IP/Port
				if nbMapping.ExternalIp != vppLbMapping.ExternalIp || nbMapping.ExternalPort != vppLbMapping.ExternalPort {
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
						if nbLocal.LocalIp == vppLocal.LocalIp && nbLocal.LocalPort == vppLocal.LocalPort &&
							nbLocal.Probability == vppLocal.Probability && nbLocal.VrfId == vppLocal.VrfId {
							found = true
						}
					}
					if !found {
						continue MappingCompare
					}
				}
				// At this point, the NB mapping matched the VPP one, so register it
				mappingIdentifier := GetStMappingIdentifier(nbMapping)
				plugin.dNatStMappingIndexes.RegisterName(mappingIdentifier, plugin.natIndexSeq, nil)
				plugin.natIndexSeq++

				// Remove registered entry from vpp mapping (configurator knows which mappings were registered)
				dMappings := *vppMappings
				*vppMappings = append(dMappings[:vppIndex], dMappings[vppIndex+1:]...)
				plugin.log.Debugf("NAT44 resync: lb-mapping %v already configured", mappingIdentifier)
			}
		} else {
			// No load balancer
			for vppIndex, vppMapping := range *vppMappings {
				// Compare VRF/SNAT fields
				if nbMapping.TwiceNat != vppMapping.TwiceNat {
					continue
				}
				// Compare external IP/Port and interface
				if nbMapping.ExternalIp != vppMapping.ExternalIp || nbMapping.ExternalPort != vppMapping.ExternalPort {
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
					plugin.log.Warnf("NAT44 resync: mapping without load balancer contains more than 1 local IP address")
					continue
				}
				nbLocal := nbMapping.LocalIps[0]
				vppLocal := vppMapping.LocalIps[0]
				if nbLocal.LocalIp != vppLocal.LocalIp || nbLocal.LocalPort != vppLocal.LocalPort ||
					nbLocal.Probability != vppLocal.Probability {
					continue
				}

				// At this point, the NB mapping matched the VPP one, so register it
				mappingIdentifier := GetStMappingIdentifier(nbMapping)
				plugin.dNatStMappingIndexes.RegisterName(mappingIdentifier, plugin.natIndexSeq, nil)
				plugin.natIndexSeq++

				// Remove registered entry from vpp mapping (so configurator knows which mappings were registered)
				dMappings := *vppMappings
				*vppMappings = append(dMappings[:vppIndex], dMappings[vppIndex+1:]...)
				plugin.log.Debugf("NAT44 resync: mapping %v already configured", mappingIdentifier)
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
			mappingIdentifier := GetIdMappingIdentifier(nbIdMapping)
			plugin.dNatIdMappingIndexes.RegisterName(mappingIdentifier, plugin.natIndexSeq, nil)
			plugin.natIndexSeq++

			// Remove registered entry from vpp mapping (configurator knows which mappings were registered)
			dIdMappings := *vppIdMappings
			*vppIdMappings = append(dIdMappings[:vppIdIndex], dIdMappings[vppIdIndex+1:]...)
			plugin.log.Debugf("NAT44 resync: identity mapping %v already configured", mappingIdentifier)
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
					plugin.log.Error(err)
					continue
				}
				pVppIP, vppIPNet, err := net.ParseCIDR(vppIP)
				if err != nil {
					plugin.log.Error(err)
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
	plugin.log.Debugf("Interface RESYNC comparison started for interface %s", nbIf.Name)
	// Type
	if nbIf.Type != vppIf.Type {
		plugin.log.Debugf("Interface RESYNC comparison: type changed (NB: %v, VPP: %v)",
			nbIf.Type, vppIf.Type)
		return true
	}
	// Enabled
	if nbIf.Enabled != vppIf.Enabled {
		plugin.log.Debugf("Interface RESYNC comparison: enabled state changed (NB: %t, VPP: %t)",
			nbIf.Enabled, vppIf.Enabled)
		return true
	}
	// VRF
	if nbIf.Vrf != vppIf.Vrf {
		plugin.log.Debugf("Interface RESYNC comparison: VRF changed (NB: %d, VPP: %d)",
			nbIf.Vrf, vppIf.Vrf)
		return true
	}
	// Container IP address
	if nbIf.ContainerIpAddress != vppIf.ContainerIpAddress {
		plugin.log.Debugf("Interface RESYNC comparison: container IP changed (NB: %s, VPP: %s)",
			nbIf.ContainerIpAddress, vppIf.ContainerIpAddress)
		return true
	}
	// DHCP setup
	if nbIf.SetDhcpClient != vppIf.SetDhcpClient {
		plugin.log.Debugf("Interface RESYNC comparison: DHCP setup changed (NB: %t, VPP: %t)",
			nbIf.SetDhcpClient, vppIf.SetDhcpClient)
		return true
	}
	//  MTU value (not valid for VxLAN)
	if nbIf.Mtu != vppIf.Mtu && nbIf.Type != intf.InterfaceType_VXLAN_TUNNEL {
		plugin.log.Debugf("Interface RESYNC comparison: MTU changed (NB: %d, VPP: %d)",
			nbIf.Mtu, vppIf.Mtu)
		return true
	}
	// MAC address (compare only if it is set in the NB configuration)
	nbMac := strings.ToUpper(nbIf.PhysAddress)
	vppMac := strings.ToUpper(vppIf.PhysAddress)
	if nbMac != "" && nbMac != vppMac {
		plugin.log.Debugf("Interface RESYNC comparison: Physical address changed (NB: %s, VPP: %s)", nbMac, vppMac)
		return true
	}
	// Unnumbered settings. If interface is unnumbered, do not compare ip addresses.
	// todo dump unnumbered data
	if nbIf.Unnumbered != nil {
		plugin.log.Debugf("RESYNC interfaces: interface %s is unnumbered, result of the comparison may not be correct", nbIf.Name)
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
			plugin.log.Debugf("Interface RESYNC comparison: IP address count changed (NB: %d, VPP: %d)",
				len(nbIf.IpAddresses), len(vppIf.IpAddresses))
			return true
		}
		// Compare every single IP address. If equal, every address should have identical counterpart
		for _, nbIP := range nbIf.IpAddresses {
			var ipFound bool
			for _, vppIP := range vppIf.IpAddresses {
				pNbIP, nbIPNet, err := net.ParseCIDR(nbIP)
				if err != nil {
					plugin.log.Error(err)
					continue
				}
				pVppIP, vppIPNet, err := net.ParseCIDR(vppIP)
				if err != nil {
					plugin.log.Error(err)
					continue
				}
				if nbIPNet.Mask.String() == vppIPNet.Mask.String() && bytes.Compare(pNbIP, pVppIP) == 0 {
					ipFound = true
					break
				}
			}
			if !ipFound {
				plugin.log.Debugf("Interface RESYNC comparison: VPP interface %s does not contain IP %s", nbIf.Name, nbIP)
				return true
			}
		}
	}
	// RxMode settings
	if nbIf.RxModeSettings == nil && vppIf.RxModeSettings != nil || nbIf.RxModeSettings != nil && vppIf.RxModeSettings == nil {
		plugin.log.Debugf("Interface RESYNC comparison: RxModeSettings changed (NB: %v, VPP: %v)",
			nbIf.RxModeSettings, vppIf.RxModeSettings)
		return true
	}
	if nbIf.RxModeSettings != nil && vppIf.RxModeSettings != nil {
		// RxMode
		if nbIf.RxModeSettings.RxMode != vppIf.RxModeSettings.RxMode {
			plugin.log.Debugf("Interface RESYNC comparison: RxMode changed (NB: %v, VPP: %v)",
				nbIf.RxModeSettings.RxMode, vppIf.RxModeSettings.RxMode)
			return true

		}
		// QueueID
		if nbIf.RxModeSettings.QueueId != vppIf.RxModeSettings.QueueId {
			plugin.log.Debugf("Interface RESYNC comparison: QueueID changed (NB: %d, VPP: %d)",
				nbIf.RxModeSettings.QueueId, vppIf.RxModeSettings.QueueId)
			return true

		}
		// QueueIDValid
		if nbIf.RxModeSettings.QueueIdValid != vppIf.RxModeSettings.QueueIdValid {
			plugin.log.Debugf("Interface RESYNC comparison: QueueIDValid changed (NB: %d, VPP: %d)",
				nbIf.RxModeSettings.QueueIdValid, vppIf.RxModeSettings.QueueIdValid)
			return true

		}
	}

	switch nbIf.Type {
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		if nbIf.Afpacket == nil && vppIf.Afpacket != nil || nbIf.Afpacket != nil && vppIf.Afpacket == nil {
			plugin.log.Debugf("Interface RESYNC comparison: AF-packet setup changed (NB: %v, VPP: %v)",
				nbIf.Afpacket, vppIf.Afpacket)
			return true
		}
		if nbIf.Afpacket != nil && vppIf.Afpacket != nil {
			// AF-packet host name
			if nbIf.Afpacket.HostIfName != vppIf.Afpacket.HostIfName {
				plugin.log.Debugf("Interface RESYNC comparison: AF-packet host name changed (NB: %s, VPP: %s)",
					nbIf.Afpacket.HostIfName, vppIf.Afpacket.HostIfName)
				return true
			}
		}
	case intf.InterfaceType_MEMORY_INTERFACE:
		if nbIf.Memif == nil && vppIf.Memif != nil || nbIf.Memif != nil && vppIf.Memif == nil {
			plugin.log.Debugf("Interface RESYNC comparison: memif setup changed (NB: %v, VPP: %v)",
				nbIf.Memif, vppIf.Memif)
			return true
		}
		if nbIf.Memif != nil && vppIf.Memif != nil {
			// Memif ID
			if nbIf.Memif.Id != vppIf.Memif.Id {
				plugin.log.Debugf("Interface RESYNC comparison: memif ID changed (NB: %d, VPP: %d)",
					nbIf.Memif.Id, vppIf.Memif.Id)
				return true
			}

			// Memif socket
			if nbIf.Memif.SocketFilename != vppIf.Memif.SocketFilename {
				plugin.log.Debugf("Interface RESYNC comparison: memif socket filename changed (NB: %s, VPP: %s)",
					nbIf.Memif.SocketFilename, vppIf.Memif.SocketFilename)
				return true
			}
			// Master
			if nbIf.Memif.Master != vppIf.Memif.Master {
				plugin.log.Debugf("Interface RESYNC comparison: memif master setup changed (NB: %t, VPP: %t)",
					nbIf.Memif.Master, vppIf.Memif.Master)
				return true
			}
			// Mode
			if nbIf.Memif.Mode != vppIf.Memif.Mode {
				plugin.log.Debugf("Interface RESYNC comparison: memif mode setup changed (NB: %v, VPP: %v)",
					nbIf.Memif.Mode, vppIf.Memif.Mode)
				return true
			}
			// Rx queues
			if nbIf.Memif.RxQueues != vppIf.Memif.RxQueues {
				plugin.log.Debugf("Interface RESYNC comparison: RxQueues changed (NB: %d, VPP: %d)",
					nbIf.Memif.RxQueues, vppIf.Memif.RxQueues)
				return true
			}
			// Tx queues
			if nbIf.Memif.TxQueues != vppIf.Memif.TxQueues {
				plugin.log.Debugf("Interface RESYNC comparison: TxQueues changed (NB: %d, VPP: %d)",
					nbIf.Memif.TxQueues, vppIf.Memif.TxQueues)
				return true
			}
			// todo secret, buffer size and ring size is not compared. VPP always returns 0 for buffer size
			// and 1 for ring size. Secret cannot be dumped at all.
		}
	case intf.InterfaceType_TAP_INTERFACE:
		if nbIf.Tap == nil && vppIf.Tap != nil || nbIf.Tap != nil && vppIf.Tap == nil {
			plugin.log.Debugf("Interface RESYNC comparison: tap setup changed (NB: %v, VPP: %v)",
				nbIf.Tap, vppIf.Tap)
			return true
		}
		if nbIf.Tap != nil && vppIf.Tap != nil {
			// Tap version
			if nbIf.Tap.Version == 2 && nbIf.Tap.Version != vppIf.Tap.Version {
				plugin.log.Debugf("Interface RESYNC comparison: tap version changed (NB: %d, VPP: %d)",
					nbIf.Tap.Version, vppIf.Tap.Version)
				return true
			}
			// Namespace and host name
			if nbIf.Tap.Namespace != vppIf.Tap.Namespace {
				plugin.log.Debugf("Interface RESYNC comparison: tap namespace changed (NB: %s, VPP: %s)",
					nbIf.Tap.Namespace, vppIf.Tap.Namespace)
				return true
			}
			// Namespace and host name
			if nbIf.Tap.HostIfName != vppIf.Tap.HostIfName {
				plugin.log.Debugf("Interface RESYNC comparison: tap host name changed (NB: %s, VPP: %s)",
					nbIf.Tap.HostIfName, vppIf.Tap.HostIfName)
				return true
			}
			// Rx ring size
			if nbIf.Tap.RxRingSize != nbIf.Tap.RxRingSize {
				plugin.log.Debugf("Interface RESYNC comparison: tap Rx ring size changed (NB: %d, VPP: %d)",
					nbIf.Tap.RxRingSize, vppIf.Tap.RxRingSize)
				return true
			}
			// Tx ring size
			if nbIf.Tap.TxRingSize != nbIf.Tap.TxRingSize {
				plugin.log.Debugf("Interface RESYNC comparison: tap Tx ring size changed (NB: %d, VPP: %d)",
					nbIf.Tap.TxRingSize, vppIf.Tap.TxRingSize)
				return true
			}
		}
	case intf.InterfaceType_VXLAN_TUNNEL:
		if nbIf.Vxlan == nil && vppIf.Vxlan != nil || nbIf.Vxlan != nil && vppIf.Vxlan == nil {
			plugin.log.Debugf("Interface RESYNC comparison: VxLAN setup changed (NB: %v, VPP: %v)",
				nbIf.Vxlan, vppIf.Vxlan)
			return true
		}
		if nbIf.Vxlan != nil && vppIf.Vxlan != nil {
			// VxLAN Vni
			if nbIf.Vxlan.Vni != vppIf.Vxlan.Vni {
				plugin.log.Debugf("Interface RESYNC comparison: VxLAN Vni changed (NB: %d, VPP: %d)",
					nbIf.Vxlan.Vni, vppIf.Vxlan.Vni)
				return true
			}
			// VxLAN Src Address
			if nbIf.Vxlan.SrcAddress != vppIf.Vxlan.SrcAddress {
				plugin.log.Debugf("Interface RESYNC comparison: VxLAN src address changed (NB: %s, VPP: %s)",
					nbIf.Vxlan.SrcAddress, vppIf.Vxlan.SrcAddress)
				return true
			}
			// VxLAN Dst Address
			if nbIf.Vxlan.DstAddress != vppIf.Vxlan.DstAddress {
				plugin.log.Debugf("Interface RESYNC comparison: VxLAN dst address changed (NB: %s, VPP: %s)",
					nbIf.Vxlan.DstAddress, vppIf.Vxlan.DstAddress)
				return true
			}
			// VxLAN Multicast
			if nbIf.Vxlan.Multicast != vppIf.Vxlan.Multicast {
				plugin.log.Debugf("Interface RESYNC comparison: VxLAN multicast address changed (NB: %s, VPP: %s)",
					nbIf.Vxlan.Multicast, vppIf.Vxlan.Multicast)
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
	if err := plugin.ifHandler.SetInterfaceTag(ifName, ifIdx); err != nil {
		return fmt.Errorf("error while adding interface tag %s, index %d: %v", ifName, ifIdx, err)
	}
	// Add AF-packet type interface to local cache
	if ifData.Type == intf.InterfaceType_AF_PACKET_INTERFACE {
		if plugin.linux != nil && plugin.afPacketConfigurator != nil && ifData.Afpacket != nil {
			// Interface is already present on the VPP so it cannot be marked as pending.
			plugin.afPacketConfigurator.addToCache(ifData, false)
		}
	}
	plugin.log.Debugf("RESYNC interfaces: registered interface %s (index %d)", ifName, ifIdx)
	return nil
}
