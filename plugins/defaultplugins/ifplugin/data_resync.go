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
	"fmt"
	"net"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/idxvpp/persist"
	_ "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/stn"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
)

// Resync writes interfaces to the VPP
//
// - resyncs the VPP
// - temporary: (checks wether sw_if_indexes are not obsolate - this will be swapped with master ID)
// - deletes obsolate status data
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

	// Read persistent mapping
	persistentIfs := nametoidx.NewNameToIdx(plugin.Log, core.PluginName("defaultvppplugins-ifplugin"), "iface resync corr", nil)
	err = persist.Marshalling(plugin.ServiceLabel.GetAgentLabel(), plugin.swIfIndexes.GetMapping(), persistentIfs)
	if err != nil {
		return []error{err}
	}

	// Register default interface and handle physical interfaces. Other configurable interfaces are added to map for
	// further processing
	configurableVppIfs := make(map[uint32]*vppdump.Interface, 0)
	for vppIfIdx, vppIf := range vppIfs {
		if vppIfIdx == 0 {
			// Register default interface (do not add to configurable interfaces)
			plugin.swIfIndexes.RegisterName(vppIf.VPPInternalName, vppIfIdx, &vppIf.Interfaces_Interface)
			continue
		}
		// Handle physical interfaces
		if vppIf.Type == intf.InterfaceType_ETHERNET_CSMACD {
			var nbIfConfig *intf.Interfaces_Interface
			var found bool
			// Look for nb equivalent
			for _, nbIf := range nbIfs {
				if nbIf.Type == intf.InterfaceType_ETHERNET_CSMACD && nbIf.Name == vppIf.VPPInternalName {
					nbIfConfig = nbIf
					found = true
					break
				}
			}
			// If interface configuration exists in the NB, call modify to update the device
			if found {
				if err = plugin.ModifyVPPInterface(nbIfConfig, &vppIf.Interfaces_Interface); err != nil {
					plugin.Log.Errorf("Error while modifying physical interface: %v", err)
					errs = append(errs, err)
				}
			} else {
				// If not, remove configuration from physical device (the device itself cannot be removed)
				if err = plugin.deleteVPPInterface(&vppIf.Interfaces_Interface, vppIfIdx); err != nil {
					plugin.Log.Errorf("Error while removing configuration from physical interface: %v", err)
					errs = append(errs, err)
				}
			}
			// In both cases, register interface
			if nbIfConfig == nil {
				nbIfConfig = &vppIf.Interfaces_Interface
			}
			plugin.swIfIndexes.RegisterName(vppIf.VPPInternalName, vppIfIdx, nbIfConfig)

		}
		// Otherwise put to the map of all configurable VPP interfaces (except default & ethernet interfaces)
		configurableVppIfs[vppIfIdx] = vppIf
	}

	// Handle case where persistent mapping is not available for configurable interfaces
	if len(persistentIfs.ListNames()) == 0 && len(configurableVppIfs) > 0 {
		plugin.Log.Debug("Persistent mapping for interfaces is empty, %v VPP interfaces is unknown", len(configurableVppIfs))
		// In such a case, there is nothing to correlate with. All existing interfaces will be removed
		for vppIfIdx, vppIf := range configurableVppIfs {
			// register interface before deletion (to keep state consistent)
			vppAgentIf := &vppIf.Interfaces_Interface
			vppAgentIf.Name = vppIf.VPPInternalName
			// todo plugin.swIfIndexes.RegisterName(vppAgentIf.Name, vppIfIdx, vppAgentIf)
			if err := plugin.deleteVPPInterface(vppAgentIf, vppIfIdx); err != nil {
				plugin.Log.Errorf("Error while removing interface: %v", err)
				errs = append(errs, err)
			}
		}
		// Configure NB interfaces
		for _, nbIf := range nbIfs {
			if err := plugin.ConfigureVPPInterface(nbIf); err != nil {
				plugin.Log.Errorf("Error while configuring interface: %v", err)
				errs = append(errs, err)
			}
		}
		return
	}

	// Find correlation between VPP, ETCD NB and persistent mapping. Update existing interfaces
	// and configure new ones
	plugin.Log.Debugf("Using persistent mapping to resync %v interfaces", len(configurableVppIfs))
	for _, nbIf := range nbIfs {
		persistIdx, _, found := persistentIfs.LookupIdx(nbIf.Name)
		if found {
			// last interface ID is known. Let's verify that there is an interface
			// with the same ID on the vpp
			var ifPresent bool
			var ifVppData *intf.Interfaces_Interface
			for vppIfIdx, vppIf := range configurableVppIfs {
				// Check at least interface type
				if vppIfIdx == persistIdx && vppIf.Type == nbIf.Type {
					ifPresent = true
					ifVppData = &vppIf.Interfaces_Interface
				}
			}
			if ifPresent && ifVppData != nil {
				// Interface exists on the vpp. Agent assumes that the the configured interface
				// correlates with the NB one (needs to be registered manually)
				plugin.swIfIndexes.RegisterName(nbIf.Name, persistIdx, nbIf)
				plugin.Log.Debugf("Registered existing interface %v with index %v", nbIf.Name, persistIdx)
				// todo calculate diff before modifying
				plugin.ModifyVPPInterface(nbIf, ifVppData)
			} else {
				// Interface exists in mapping but not in vpp.
				if err := plugin.ConfigureVPPInterface(nbIf); err != nil {
					plugin.Log.Errorf("Error while configuring interface: %v", err)
					errs = append(errs, err)
				}
			}
		} else {
			// a new interface (missing in persistent mapping)
			if err := plugin.ConfigureVPPInterface(nbIf); err != nil {
				plugin.Log.Errorf("Error while configuring interface: %v", err)
				errs = append(errs, err)
			}
		}
	}

	// Remove obsolete config
	for vppIfIdx, vppIf := range configurableVppIfs {
		// If interface index is not registered yet, remove it
		_, _, found := plugin.swIfIndexes.LookupName(vppIfIdx)
		if !found {
			// Remove configuration
			vppAgentIf := &vppIf.Interfaces_Interface
			vppAgentIf.Name = vppIf.VPPInternalName
			// todo plugin.swIfIndexes.RegisterName(vppAgentIf.Name, vppIfIdx, vppAgentIf)
			if err := plugin.deleteVPPInterface(vppAgentIf, vppIfIdx); err != nil {
				plugin.Log.Errorf("Error while removing interface: %v", err)
				errs = append(errs, err)
			}
			plugin.Log.Debugf("Removed unknown interface with index %v", vppIfIdx)
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
			if err := vppcalls.DelStnRule(vppStnRule.SwIfIndex, &vppStnIP, plugin.Log, plugin.vppChan,
				nil); err != nil {
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
			if err := vppcalls.DelStnRule(vppStnRule.SwIfIndex, &vppStnIP, plugin.Log, plugin.vppChan,
				nil); err != nil {
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
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC nat global config.")

	vppNatGlobal, err := vppdump.Nat44GlobalConfigDump(plugin.SwIfIndexes, plugin.Log, plugin.vppChan, plugin.Stopwatch)
	if err != nil {
		return fmt.Errorf("failed to dump NAT44 global config: %v", err)
	}

	// Modify will made all the diffs needed
	return plugin.ModifyNatGlobalConfig(vppNatGlobal, nbGlobal)
}

// ResyncSNat writes NAT static mapping config to the the empty VPP
func (plugin *NatConfigurator) ResyncSNat(sNatConf []*nat.Nat44SNat_SNatConfig) error {
	// todo SNAT not implemented yet, nothing to resync
	return nil
}

// ResyncDNat writes NAT static mapping config to the the empty VPP
func (plugin *NatConfigurator) ResyncDNat(nbDNatConfig []*nat.Nat44DNat_DNatConfig) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC DNAT config.")

	vppDNat, err := vppdump.NAT44DNatDump(plugin.SwIfIndexes, plugin.Log, plugin.vppChan, plugin.Stopwatch)
	if err != nil {
		return fmt.Errorf("failed to dump DNAT config: %v", err)
	}
	if len(vppDNat.DnatConfig) == 0 {
		return nil
	}
	// For now, there is only one DNAT config
	vppDNatConfig := vppDNat.DnatConfig[0]

	// Correlate with existing config
	for _, nbDNat := range nbDNatConfig {
		if len(nbDNat.StMappings) == 0 && len(nbDNat.IdMappings) == 0 {
			plugin.Log.Warnf("NB DNAT entry %v all mappings are empty", nbDNat.Label)
			continue
		} else {
			// Compare all VPP mappings with the NB, register existing ones
			plugin.resolveMappings(nbDNat, vppDNatConfig.StMappings, vppDNatConfig.IdMappings)
			// Configure all missing DNAT mappings
			for _, nbMapping := range nbDNat.StMappings {
				mappingIdentifier := getStMappingIdentifier(nbMapping)
				_, _, found := plugin.DNatStMappingIndices.LookupIdx(mappingIdentifier)
				if !found {
					// Configure missing mapping
					if len(nbMapping.LocalIps) == 1 {
						if err := plugin.handleStaticMapping(nbMapping, true); err != nil {
							plugin.Log.Errorf("NAT44 resync: failed to configure static mapping: %v", err)
							continue
						}
					} else {
						if err := plugin.handleStaticMappingLb(nbMapping, true); err != nil {
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
					if err := plugin.handleIdentityMapping(nbIdMapping, true); err != nil {
						plugin.Log.Errorf("NAT44 resync: failed to configure identity mapping: %v", err)
						continue
					}

					// Register new DNAT mapping
					plugin.DNatIdMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
					plugin.NatIndexSeq++
					plugin.Log.Debugf("NAT44 resync: new identity mapping %v configured", mappingIdentifier)
				}
			}
			// At this point, the DNAT is completely configured and can be registered
			plugin.DNatIndices.RegisterName(nbDNat.Label, plugin.NatIndexSeq, nil)
			plugin.NatIndexSeq++
			plugin.Log.Debugf("NAT44 resync: DNAT %v synced", nbDNat.Label)
		}
	}

	// The last step is to remove obsolete mappings
	for _, vppMapping := range vppDNatConfig.StMappings {
		// Static mapping
		if len(vppMapping.LocalIps) == 1 {

			if err := plugin.handleStaticMapping(vppMapping, false); err != nil {
				plugin.Log.Errorf("NAT44 resync: failed to remove static mapping: %v", err)
				continue
			}
		} else {
			// Lb-static mapping
			if err := plugin.handleStaticMappingLb(vppMapping, false); err != nil {
				plugin.Log.Errorf("NAT44 resync: failed to remove static mapping: %v", err)
				continue
			}
		}
	}
	for _, vppIdMapping := range vppDNatConfig.IdMappings {
		// Identity mapping
		if err := plugin.handleIdentityMapping(vppIdMapping, false); err != nil {
			plugin.Log.Errorf("NAT44 resync: failed to remove identity mapping: %v", err)
			continue
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC DNAT config done.")

	return nil
}

// Looks for the same mapping in the VPP, register existing ones
func (plugin *NatConfigurator) resolveMappings(nbDNatConfig *nat.Nat44DNat_DNatConfig,
	vppMappings []*nat.Nat44DNat_DNatConfig_StaticMappigs, vppIdMappings []*nat.Nat44DNat_DNatConfig_IdentityMappings) {
	// Iterate over static mappings in NB DNAT config
	for _, nbMapping := range nbDNatConfig.StMappings {
		if len(nbMapping.LocalIps) > 1 {
			// Load balanced
		MappingCompare:
			for vppIndex, vppLbMapping := range vppMappings {
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
					for _, vppLocal := range vppLbMapping.LocalIps {
						if nbLocal.LocalIP != vppLocal.LocalIP || nbLocal.LocalPort != vppLocal.LocalPort ||
							nbLocal.Probability != vppLocal.Probability {
							continue MappingCompare
						}
					}
				}
				// At this point, the NB mapping matched the VPP one, so register it
				mappingIdentifier := getStMappingIdentifier(nbMapping)
				plugin.DNatStMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
				plugin.NatIndexSeq++

				// Remove registered entry from vpp mapping (configurator knows which mappings were registered)
				vppMappings = append(vppMappings[:vppIndex], vppMappings[vppIndex+1:]...)
				plugin.Log.Debugf("NAT44 resync: lb-mapping %v already configured", mappingIdentifier)
			}
		} else {
			// No load balancer
			for vppIndex, vppMapping := range vppMappings {
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
				vppMappings = append(vppMappings[:vppIndex], vppMappings[vppIndex+1:]...)
				plugin.Log.Debugf("NAT44 resync: mapping %v already configured", mappingIdentifier)
			}
		}
	}
	// Iterate over identity mappings in NB DNAT config
	for _, nbIdMapping := range nbDNatConfig.IdMappings {
		for vppIdIndex, vppIdMapping := range vppIdMappings {
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
			vppIdMappings = append(vppIdMappings[:vppIdIndex], vppIdMappings[vppIdIndex+1:]...)
			plugin.Log.Debugf("NAT44 resync: identity mapping %v already configured", mappingIdentifier)
		}
	}
}
