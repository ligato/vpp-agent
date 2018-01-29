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
	"strconv"

	"net"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/idxvpp/persist"
	bin_api_nat "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
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

	// Register default and ethernet interfaces
	configurableVppIfs := make(map[uint32]*vppdump.Interface, 0)
	for vppIfIdx, vppIf := range vppIfs {
		if vppIfIdx == 0 || vppIf.Type == intf.InterfaceType_ETHERNET_CSMACD {
			plugin.swIfIndexes.RegisterName(vppIf.VPPInternalName, vppIfIdx, &vppIf.Interfaces_Interface)
			continue
		}
		// fill map of all configurable VPP interfaces (skip default & ethernet interfaces)
		configurableVppIfs[vppIfIdx] = vppIf
	}

	// Handle case where persistent mapping is not available
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
			// To keep interface state consistent, interface has to be
			// temporary registered before removal
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

	addresses, err := vppdump.Nat44AddressDump(plugin.Log, plugin.vppChan, nil)
	if err != nil {
		plugin.Log.Errorf("address dump: %v", err)
	}
	interfaces, err := vppdump.Nat44InterfaceDump(plugin.Log, plugin.vppChan, nil)
	if err != nil {
		plugin.Log.Errorf("iface dump: %v", err)
	}
	staticMappings, err := vppdump.Nat44StaticMappingDump(plugin.Log, plugin.vppChan, nil)
	if err != nil {
		plugin.Log.Errorf("sm dump: %v", err)
	}
	lbStaticMappings, err := vppdump.Nat44StaticMappingLbDump(plugin.Log, plugin.vppChan, nil)
	if err != nil {
		plugin.Log.Errorf("lbsm dump: %v", err)
	}

	plugin.Log.Warn("Addresses:")
	for i, address := range addresses {
		plugin.Log.Warnf("%v: IP:%v, twice-nat:%v, vrf:%v", i, address.IPAddress, address.TwiceNat, address.VrfID)
	}

	plugin.Log.Warn("Interfaces:")
	for i, iface := range interfaces {
		plugin.Log.Warnf("%v: IfIdx:%v, inside:%v", i, iface.IfIdx, iface.IsInside)
	}

	plugin.Log.Warn("Static mappings:")
	for i, staticMapping := range staticMappings {
		plugin.Log.Warnf("%v: lcIP:%v:%v, exIP:%v:%v, exIdx:%v, proto:%v, addrOnly:%v, twice-nat:%v, vrf:%v ", i, staticMapping.LocalIPs[0].LocalIP, staticMapping.LocalIPs[0].LocalPort, staticMapping.ExternalIP,
			staticMapping.ExternalPort, staticMapping.ExternalIfIdx, staticMapping.Protocol, staticMapping.AddressOnly, staticMapping.TwiceNat, staticMapping.VrfID)
	}

	plugin.Log.Warn("LB-Static mappings:")
	for i, lbStaticMapping := range lbStaticMappings {
		plugin.Log.Warnf("%v: exIP:%v:%v, exIdx:%v, proto:%v, addrOnly:%v, twice-nat:%v, vrf:%v ", i, lbStaticMapping.ExternalIP,
			lbStaticMapping.ExternalPort, lbStaticMapping.ExternalIfIdx, lbStaticMapping.Protocol, lbStaticMapping.AddressOnly, lbStaticMapping.TwiceNat, lbStaticMapping.VrfID)
		for _, local := range lbStaticMapping.LocalIPs {
			plugin.Log.Warnf("	IP:%v:%v, prob.:%v", local.LocalIP, local.LocalPort, local.Probability)
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC nat global config.")

	// Dump existing global configuration (forwarding, address pool and interfaces)
	forwarding, err := vppdump.Nat44IsForwardingEnabled(plugin.Log, plugin.vppChan,
		measure.GetTimeLog(&bin_api_nat.Nat44ForwardingIsEnabled{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("failed to dump NAT44 forwarding: %v", err)
	}
	addressPools, err := vppdump.Nat44AddressDump(plugin.Log, plugin.vppChan,
		measure.GetTimeLog(&bin_api_nat.Nat44AddressDump{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("failed to dump NAT44 address pools: %v", err)
	}
	interfaces, err = vppdump.Nat44InterfaceDump(plugin.Log, plugin.vppChan,
		measure.GetTimeLog(&bin_api_nat.Nat44InterfaceDump{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("failed to dump NAT44 interfaces: %v", err)
	}

	// Reconstruct existing global config
	var vrf uint32
	var vppAddressPools []*nat.Nat44Global_AddressPool
	for _, addressPool := range addressPools {
		vppAddressPools = append(vppAddressPools, &nat.Nat44Global_AddressPool{
			FirstSrcAddress: addressPool.IPAddress,
			TwiceNat:        addressPool.TwiceNat,
		})
		vrf = addressPool.VrfID // VRF ID is the same for every entry in address pool
	}

	var vppInInterfaces []string
	var vppOutInterfaces []string
	for _, vppIface := range interfaces {
		// Look for interface
		ifName, _, found := plugin.SwIfIndexes.LookupName(vppIface.IfIdx)
		if !found {
			plugin.Log.Warnf("NAT44 interface dump: interface %v not found in the mapping")
			continue
		}
		if vppIface.IsInside {
			vppInInterfaces = append(vppInInterfaces, ifName)
			continue
		}
		vppOutInterfaces = append(vppOutInterfaces, ifName)
	}

	// VPP global config
	vppNatGlobal := &nat.Nat44Global{
		AddressPool: vppAddressPools,
		VrfId:       vrf,
		Forwarding:  forwarding,
		In:          vppInInterfaces,
		Out:         vppOutInterfaces,
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

	// Read existing DNAT configuration from the VPP
	vppDNatMappings, err := vppdump.Nat44StaticMappingDump(plugin.Log, plugin.vppChan,
		measure.GetTimeLog(&bin_api_nat.Nat44StaticMappingDump{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	vppDNatLbMappings, err := vppdump.Nat44StaticMappingLbDump(plugin.Log, plugin.vppChan,
		measure.GetTimeLog(&bin_api_nat.Nat44LbStaticMappingDump{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	// Correlate with existing config
	for _, nbDNat := range nbDNatConfig {
		if len(nbDNat.Mapping) == 0 {
			plugin.Log.Warnf("NB DNAT entry %v mapping is empty", nbDNat.Label)
			continue
		} else {
			// Compare all VPP mappings with the NB, register existing ones
			plugin.resolveMappings(nbDNat, vppDNatMappings, vppDNatLbMappings)
			// Configure all missing DNAT mappings
			for _, nbMapping := range nbDNat.Mapping {
				mappingIdentifier := getMappingIdentifier(nbMapping, nbDNat.VrfId)
				_, _, found := plugin.DNatMappingIndices.LookupIdx(mappingIdentifier)
				if !found {
					// Configure missing mapping
					if len(nbMapping.LocalIp) == 1 {
						if err := plugin.handleStaticMapping(nbMapping.ExternalIP, nbMapping.ExternalInterface, nbDNat.VrfId,
							nbDNat.SNatEnabled, nbMapping, true); err != nil {
							plugin.Log.Errorf("NAT44 resync: failed to configure static mapping: %v", err)
							continue
						}
					} else {
						if err := plugin.handleStaticMappingLb(nbDNat.VrfId, nbDNat.SNatEnabled, nbMapping, true); err != nil {
							plugin.Log.Errorf("NAT44 resync: failed to configure lb-static mapping: %v", err)
							continue
						}
					}
					// Register new DNAT mapping
					plugin.DNatMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
					plugin.NatIndexSeq++
				}
			}
			// At this point, the DNAT is completely configured and can be registered
			plugin.DNatIndices.RegisterName(nbDNat.Label, plugin.NatIndexSeq, nil)
			plugin.NatIndexSeq++
			plugin.Log.Debugf("NAT44 resync: DNAT %v synced", nbDNat.Label)
		}
	}

	// The last step is to remove obsolete mappings
	// Static mapping
	for _, vppMapping := range vppDNatMappings {
		// Get interface name
		ifName, _, found := plugin.SwIfIndexes.LookupName(vppMapping.ExternalIfIdx)
		if !found {
			plugin.Log.Errorf("NAT44 resync: failed to remove mapping, interface %v is missing", vppMapping.ExternalIfIdx)
			continue
		}
		data := plugin.getModelConfig(vppMapping, ifName)

		if err := plugin.handleStaticMapping(vppMapping.ExternalIP, ifName, vppMapping.VrfID,
			vppMapping.TwiceNat, data, false); err != nil {
			plugin.Log.Errorf("NAT44 resync: failed to remove static mapping: %v", err)
			continue
		}
	}

	// Lb-static mapping
	for _, vppLbMapping := range vppDNatLbMappings {
		data := plugin.getModelConfig(vppLbMapping, "")

		if err := plugin.handleStaticMappingLb(vppLbMapping.VrfID, vppLbMapping.TwiceNat, data, false); err != nil {
			plugin.Log.Errorf("NAT44 resync: failed to remove static mapping: %v", err)
			continue
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC DNAT config done.")

	return nil
}

// Looks for the same mapping in the VPP, register existing ones
func (plugin *NatConfigurator) resolveMappings(nbDNatConfig *nat.Nat44DNat_DNatConfig,
	vppMappings, vppLbMappings []*vppdump.Nat44StaticMappingEntry) {
	// Store VRF and SNAT values (they are the same for every mapping in DNAT config)
	vrf := nbDNatConfig.VrfId
	sNat := nbDNatConfig.SNatEnabled
	// Iterate over mappings in NB DNAT config
	for _, nbMapping := range nbDNatConfig.Mapping {
		if len(nbMapping.LocalIp) > 1 {
			// Load balanced
		MappingCompare:
			for vppIndex, vppLbMapping := range vppMappings {
				// Compare VRF/SNAT fields
				if vrf != vppLbMapping.VrfID || sNat != vppLbMapping.TwiceNat {
					continue
				}
				// Compare external IP/Port
				nbExIP, nbExPort, err := splitIPPortVal(nbMapping.ExternalIP)
				if err != nil {
					plugin.Log.Errorf("NAT44 resync: %v", err)
					continue
				}
				if nbExIP != vppLbMapping.ExternalIP || uint32(nbExPort) != vppLbMapping.ExternalPort {
					continue
				}
				// Compare protocol
				if nbMapping.Protocol != vppLbMapping.Protocol {
					continue
				}
				// Compare Local IP/Port and probability addresses
				if len(nbMapping.LocalIp) != len(vppLbMapping.LocalIPs) {
					continue
				}
				for _, nbLocal := range nbMapping.LocalIp {
					for _, vppLocal := range vppLbMapping.LocalIPs {
						nbLcIP, nbLcPort, err := splitIPPortVal(nbLocal.LocalIP)
						if err != nil {
							plugin.Log.Errorf("NAT44 resync: %v", err)
							continue MappingCompare
						}
						if nbLcIP != vppLocal.LocalIP || uint32(nbLcPort) != vppLocal.LocalPort ||
							nbLocal.Probability != vppLocal.Probability {
							continue MappingCompare
						}
					}
				}
				// At this point, the NB mapping matched the VPP one, so register it
				mappingIdentifier := getMappingIdentifier(nbMapping, vrf)
				plugin.DNatMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
				plugin.NatIndexSeq++

				// Remove registered entry from vpp mapping (configurator knows which mappings were registered)
				vppMappings = append(vppMappings[:vppIndex], vppMappings[vppIndex+1:]...)
				plugin.Log.Debugf("NAT44 resync: lb-mapping %v already configured", mappingIdentifier)
			}
		} else {
			// No load balancer
			for vppIndex, vppMapping := range vppMappings {
				// Compare VRF/SNAT fields
				if vrf != vppMapping.VrfID || sNat != vppMapping.TwiceNat {
					continue
				}
				// Compare external IP/Port and interface
				nbExIP, nbExPort, err := splitIPPortVal(nbMapping.ExternalIP)
				if err != nil {
					plugin.Log.Errorf("NAT44 resync: %v", err)
					continue
				}
				if nbExIP != vppMapping.ExternalIP || uint32(nbExPort) != vppMapping.ExternalPort {
					continue
				}
				// Compare external interface
				nbExIfIdx, _, found := plugin.SwIfIndexes.LookupIdx(nbMapping.ExternalInterface)
				if !found {
					plugin.Log.Errorf("NAT44 resync: interface %v not found", nbMapping.ExternalInterface)
					continue
				}
				if nbExIfIdx != vppMapping.ExternalIfIdx {
					continue
				}
				// Compare protocol
				if nbMapping.Protocol != vppMapping.Protocol {
					continue
				}
				// Compare Local IP/Port and probability addresses (there is only one local IP address in this case)
				if len(nbMapping.LocalIp) != 1 || len(vppMapping.LocalIPs) != 1 {
					plugin.Log.Warnf("NAT44 resync: mapping without load balancer contains more than 1 local IP address")
					continue
				}
				nbLocal := nbMapping.LocalIp[0]
				vppLocal := vppMapping.LocalIPs[0]
				nbLcIP, nbLcPort, err := splitIPPortVal(nbLocal.LocalIP)
				if err != nil {
					plugin.Log.Errorf("NAT44 resync: %v", err)
					continue
				}
				if nbLcIP != vppLocal.LocalIP || uint32(nbLcPort) != vppLocal.LocalPort ||
					nbLocal.Probability != vppLocal.Probability {
					continue
				}

				// At this point, the NB mapping matched the VPP one, so register it
				mappingIdentifier := getMappingIdentifier(nbMapping, vrf)
				plugin.DNatMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
				plugin.NatIndexSeq++

				// Remove registered entry from vpp mapping (so configurator knows which mappings were registered)
				vppMappings = append(vppMappings[:vppIndex], vppMappings[vppIndex+1:]...)
				plugin.Log.Debugf("NAT44 resync: mapping %v already configured", mappingIdentifier)
			}
		}
	}
}

// Reconstruct NAT44 DNat static mapping configuration
func (plugin *NatConfigurator) getModelConfig(vppMapping *vppdump.Nat44StaticMappingEntry, ifName string) *nat.Nat44DNat_DNatConfig_Mapping {
	var localIPs []*nat.Nat44DNat_DNatConfig_Mapping_LocalIP
	for _, localIP := range vppMapping.LocalIPs {
		localIPs = append(localIPs, &nat.Nat44DNat_DNatConfig_Mapping_LocalIP{
			LocalIP: localIP.LocalIP + "/" + strconv.Itoa(int(localIP.LocalPort)),
		})
	}

	return &nat.Nat44DNat_DNatConfig_Mapping{
		ExternalInterface: ifName,
		ExternalIP:        vppMapping.ExternalIP + "/" + strconv.Itoa(int(vppMapping.ExternalPort)),
		Protocol:          vppMapping.Protocol,
		LocalIp:           localIPs,
	}
}
