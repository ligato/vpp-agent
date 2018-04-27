// Copyright (c) 2018 Cisco and/or its affiliates.
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

//go:generate protoc --proto_path=../common/model/nat --gogo_out=../common/model/nat ../common/model/nat/nat.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/nat.api.json --output-dir=../common/bin_api

package ifplugin

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// Mapping labels
const (
	static   = "-static-"
	staticLb = "-staticLb-"
	identity = "-identity-"
	dummyTag = "dummy-tag" // used for deletion where tag is not needed
)

// NatConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of NAT address pools and static entries with or without a load ballance,
// as modelled by the proto file "../common/model/nat/nat.proto"
// and stored in ETCD under the keys:
// - "/vnf-agent/{agent-label}/vpp/config/v1/nat/{vrf}/addrpool/" for NAT address pool
// - "/vnf-agent/{agent-label}/vpp/config/v1/nat/{vrf}/static/" for NAT static mapping
// - "/vnf-agent/{agent-label}/vpp/config/v1/nat/{vrf}/staticlb/" for NAT static mapping with
//   load balancer
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type NatConfigurator struct {
	log logging.Logger

	// Global config
	globalNAT *nat.Nat44Global

	// Mappings
	ifIndexes            ifaceidx.SwIfIndex
	sNatIndexes          idxvpp.NameToIdxRW // SNAT config indices
	sNatMappingIndexes   idxvpp.NameToIdxRW // SNAT indices for static mapping
	dNatIndexes          idxvpp.NameToIdxRW // DNAT indices
	dNatStMappingIndexes idxvpp.NameToIdxRW // DNAT indices for static mapping
	dNatIdMappingIndexes idxvpp.NameToIdxRW // DNAT indices for identity mapping
	natIndexSeq          uint32             // Nat name-to-idx mapping sequence
	natMappingTagSeq     uint32             // Static/identity mapping tag sequence

	// a map of missing interfaces which should be enabled for NAT (format ifName/data)
	notEnabledIfs map[string]*nat.Nat44Global_NatInterface
	// a map of NAT-enabled interfaces which should be disabled (format ifName/data)
	notDisabledIfs map[string]*nat.Nat44Global_NatInterface

	// VPP channels
	vppChan     vppcalls.VPPChannel
	vppDumpChan vppcalls.VPPChannel

	stopwatch *measure.Stopwatch
}

// Init NAT configurator
func (plugin *NatConfigurator) Init(pluginName core.PluginName, logger logging.PluginLogger, goVppMux govppmux.API, ifIndexes ifaceidx.SwIfIndex,
	enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-nat-conf")
	plugin.log.Debug("Initializing NAT configurator")

	// Mappings
	plugin.ifIndexes = ifIndexes
	plugin.notEnabledIfs = make(map[string]*nat.Nat44Global_NatInterface)
	plugin.notDisabledIfs = make(map[string]*nat.Nat44Global_NatInterface)
	plugin.sNatIndexes = nametoidx.NewNameToIdx(plugin.log, pluginName, "snat-indices", nil)
	plugin.sNatMappingIndexes = nametoidx.NewNameToIdx(plugin.log, pluginName, "snat-mapping-indices", nil)
	plugin.dNatIndexes = nametoidx.NewNameToIdx(plugin.log, pluginName, "dnat-indices", nil)
	plugin.dNatStMappingIndexes = nametoidx.NewNameToIdx(plugin.log, pluginName, "dnat-st-mapping-indices", nil)
	plugin.dNatIdMappingIndexes = nametoidx.NewNameToIdx(plugin.log, pluginName, "dnat-id-mapping-indices", nil)
	plugin.natIndexSeq, plugin.natMappingTagSeq = 1, 1

	// Init VPP API channel
	if plugin.vppChan, err = goVppMux.NewAPIChannel(); err != nil {
		return err
	}
	if plugin.vppDumpChan, err = goVppMux.NewAPIChannel(); err != nil {
		return err
	}

	// Stopwatch
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("InterfaceConfigurator", plugin.log)
	}

	// Check VPP message compatibility
	if err := vppcalls.CheckMsgCompatibilityForNat(plugin.vppChan); err != nil {
		return err
	}

	return nil
}

// Close used resources
func (plugin *NatConfigurator) Close() error {
	_, err := safeclose.CloseAll(plugin.vppChan, plugin.vppDumpChan)
	return err
}

// GetGlobalNat makes current global nat accessible
func (plugin *NatConfigurator) GetGlobalNat() *nat.Nat44Global {
	return plugin.globalNAT
}

// IsInNotEnabledIfCache checks if interface is present in 'notEnabledIfs' cache
func (plugin *NatConfigurator) IsInNotEnabledIfCache(ifName string) bool {
	_, ok := plugin.notEnabledIfs[ifName]
	return ok
}

// IsInNotDisabledIfCache checks if interface is present in 'notDisabledIfs' cache
func (plugin *NatConfigurator) IsInNotDisabledIfCache(ifName string) bool {
	_, ok := plugin.notDisabledIfs[ifName]
	return ok
}

// IsInNotDisabledIfCache checks if interface is present in 'notDisabledIfs' cache
func (plugin *NatConfigurator) IsDNatLabelRegistered(label string) bool {
	_, _, found := plugin.dNatIndexes.LookupIdx(label)
	return found
}

// IsInNotDisabledIfCache checks if DNAT static mapping with provided id is registered
func (plugin *NatConfigurator) IsDNatLabelStMappingRegistered(id string) bool {
	_, _, found := plugin.dNatStMappingIndexes.LookupIdx(id)
	return found
}

// IsDNatLabelIdMappingRegistered checks if DNAT identity mapping with provided id is registered
func (plugin *NatConfigurator) IsDNatLabelIdMappingRegistered(id string) bool {
	_, _, found := plugin.dNatIdMappingIndexes.LookupIdx(id)
	return found
}

// SetNatGlobalConfig configures common setup for all NAT use cases
func (plugin *NatConfigurator) SetNatGlobalConfig(config *nat.Nat44Global) error {
	plugin.log.Info("Setting up NAT global config")

	// Store global NAT configuration (serves as cache)
	plugin.globalNAT = config

	// Forwarding
	if err := vppcalls.SetNat44Forwarding(config.Forwarding, plugin.vppChan, plugin.stopwatch); err != nil {
		return err
	}
	if config.Forwarding {
		plugin.log.Debugf("NAT forwarding enabled")
	} else {
		plugin.log.Debugf("NAT forwarding disabled")
	}

	// / Inside/outside interfaces
	if len(config.NatInterfaces) > 0 {
		if err := plugin.enableNatInterfaces(config.NatInterfaces); err != nil {
			return err
		}
	} else {
		plugin.log.Debug("No NAT interfaces to configure")
	}

	// Address pool
	var wasErr error
	for _, pool := range config.AddressPools {
		if pool.FirstSrcAddress == "" && pool.LastSrcAddress == "" {
			wasErr = fmt.Errorf("invalid address pool config, no IP address provided")
			plugin.log.Error(wasErr)
			continue
		}
		var firstIP []byte
		var lastIP []byte
		if pool.FirstSrcAddress != "" {
			firstIP = net.ParseIP(pool.FirstSrcAddress).To4()
			if firstIP == nil {
				wasErr = fmt.Errorf("unable to parse IP address %v", pool.FirstSrcAddress)
				plugin.log.Error(wasErr)
				continue
			}
		}
		if pool.LastSrcAddress != "" {
			lastIP = net.ParseIP(pool.LastSrcAddress).To4()
			if lastIP == nil {
				wasErr = fmt.Errorf("unable to parse IP address %v", pool.LastSrcAddress)
				plugin.log.Error(wasErr)
				continue
			}
		}
		// Both fields have to be set, at least at the same value if only one of them is set
		if firstIP == nil {
			firstIP = lastIP
		} else if lastIP == nil {
			lastIP = firstIP
		}

		// Configure address pool
		if err := vppcalls.AddNat44AddressPool(firstIP, lastIP, pool.VrfId, pool.TwiceNat, plugin.vppChan, plugin.stopwatch); err != nil {
			wasErr = err
			plugin.log.Error(wasErr)
			continue
		}
	}

	plugin.log.Debug("Setting up NAT global config done")

	return wasErr
}

// ModifyNatGlobalConfig modifies common setup for all NAT use cases
func (plugin *NatConfigurator) ModifyNatGlobalConfig(oldConfig, newConfig *nat.Nat44Global) (err error) {
	plugin.log.Info("Modifying NAT global config")

	// Replace global NAT config
	plugin.globalNAT = newConfig

	// Forwarding
	if oldConfig.Forwarding != newConfig.Forwarding {
		if err := vppcalls.SetNat44Forwarding(newConfig.Forwarding, plugin.vppChan, plugin.stopwatch); err != nil {
			return err
		}
	}

	// Inside/outside interfaces
	toSetIn, toSetOut, toUnsetIn, toUnsetOut := diffInterfaces(oldConfig.NatInterfaces, newConfig.NatInterfaces)
	if err := plugin.disableNatInterfaces(toUnsetIn); err != nil {
		return err
	}
	if err := plugin.disableNatInterfaces(toUnsetOut); err != nil {
		return err
	}
	if err := plugin.enableNatInterfaces(toSetIn); err != nil {
		return err
	}
	if err := plugin.enableNatInterfaces(toSetOut); err != nil {
		return err
	}

	// Address pool
	toAdd, toRemove := diffAddressPools(oldConfig.AddressPools, newConfig.AddressPools)
	if err := plugin.addDelAddressPool(toRemove, false); err != nil {
		return err
	}
	if err := plugin.addDelAddressPool(toAdd, true); err != nil {
		return err
	}

	plugin.log.Debug("Modifying NAT global config done")

	return nil
}

// DeleteNatGlobalConfig removes common setup for all NAT use cases
func (plugin *NatConfigurator) DeleteNatGlobalConfig(config *nat.Nat44Global) (err error) {
	plugin.log.Info("Deleting NAT global config")

	// Remove global NAT config
	plugin.globalNAT = nil

	// Inside/outside interfaces
	if len(config.NatInterfaces) > 0 {
		if err := plugin.disableNatInterfaces(config.NatInterfaces); err != nil {
			return err
		}
	}

	// Address pools
	if len(config.AddressPools) > 0 {
		if err := plugin.addDelAddressPool(config.AddressPools, false); err != nil {
			return err
		}
	}

	plugin.log.Debug("Deleting NAT global config done")

	return nil
}

// ConfigureSNat configures new SNAT setup
func (plugin *NatConfigurator) ConfigureSNat(sNat *nat.Nat44SNat_SNatConfig) error {
	plugin.log.Warn("SNAT CREATE not implemented")
	return nil
}

// ModifySNat modifies existing SNAT setup
func (plugin *NatConfigurator) ModifySNat(oldSNat, newSNat *nat.Nat44SNat_SNatConfig) error {
	plugin.log.Warn("SNAT MODIFY not implemented")
	return nil
}

// DeleteSNat removes existing SNAT setup
func (plugin *NatConfigurator) DeleteSNat(sNat *nat.Nat44SNat_SNatConfig) error {
	plugin.log.Warn("SNAT DELETE not implemented")
	return nil
}

// ConfigureDNat configures new DNAT setup
func (plugin *NatConfigurator) ConfigureDNat(dNat *nat.Nat44DNat_DNatConfig) error {
	plugin.log.Infof("Configuring DNAT with label %v", dNat.Label)

	var wasErr error

	// Resolve static mapping
	if err := plugin.configureStaticMappings(dNat.Label, dNat.StMappings); err != nil {
		wasErr = err
		plugin.log.Errorf("Failed to configure static mapping for DNAT %s: %v", dNat.Label, err)
	}

	// Resolve identity mapping
	if err := plugin.configureIdentityMappings(dNat.Label, dNat.IdMappings); err != nil {
		wasErr = err
		plugin.log.Errorf("Failed to configure identity mapping for DNAT %s: %v", dNat.Label, err)
	}

	// Register DNAT configuration
	plugin.dNatIndexes.RegisterName(dNat.Label, plugin.natIndexSeq, nil)
	plugin.natIndexSeq++
	plugin.log.Debugf("DNAT configuration registered (label: %v)", dNat.Label)

	plugin.log.Infof("DNAT %v configuration done", dNat.Label)

	return wasErr
}

// ModifyDNat modifies existing DNAT setup
func (plugin *NatConfigurator) ModifyDNat(oldDNat, newDNat *nat.Nat44DNat_DNatConfig) error {
	plugin.log.Infof("Modifying DNAT with label %v", newDNat.Label)

	var wasErr error

	// Static mappings
	stmToAdd, stmToRemove := plugin.diffStatic(oldDNat.StMappings, newDNat.StMappings)

	if err := plugin.unconfigureStaticMappings(stmToRemove); err != nil {
		wasErr = err
		plugin.log.Errorf("Failed to remove static mapping from DNAT %s: %v", newDNat.Label, err)
	}

	if err := plugin.configureStaticMappings(newDNat.Label, stmToAdd); err != nil {
		wasErr = err
		plugin.log.Errorf("Failed to configure static mapping for DNAT %s: %v", newDNat.Label, err)
	}

	// Identity mappings
	idToAdd, idToRemove := plugin.diffIdentity(oldDNat.IdMappings, newDNat.IdMappings)

	if err := plugin.unconfigureIdentityMappings(idToRemove); err != nil {
		wasErr = err
		plugin.log.Errorf("Failed to remove identity mapping from DNAT %s: %v", newDNat.Label, err)
	}

	if err := plugin.configureIdentityMappings(newDNat.Label, idToAdd); err != nil {
		wasErr = err
		plugin.log.Errorf("Failed to configure identity mapping for DNAT %s: %v", newDNat.Label, err)
	}

	plugin.log.Infof("DNAT %v modification done", newDNat.Label)

	return wasErr
}

// DeleteDNat removes existing DNAT setup
func (plugin *NatConfigurator) DeleteDNat(dNat *nat.Nat44DNat_DNatConfig) error {
	plugin.log.Infof("Deleting DNAT with label %v", dNat.Label)

	var wasErr error

	// In delete case, vpp-agent attempts to reconstruct every static mapping entry and remove it from the VPP
	if err := plugin.unconfigureStaticMappings(dNat.StMappings); err != nil {
		wasErr = err
		plugin.log.Errorf("Failed to remove static mapping from DNAT %s: %v", dNat.Label, err)
	}

	// Do the same also for identity apping
	if err := plugin.unconfigureIdentityMappings(dNat.IdMappings); err != nil {
		wasErr = err
		plugin.log.Errorf("Failed to remove identity mapping from DNAT %s: %v", dNat.Label, err)
	}

	// Unregister DNAT configuration
	plugin.dNatIndexes.UnregisterName(dNat.Label)
	plugin.log.Debugf("DNAT configuration unregistered (label: %v)", dNat.Label)

	plugin.log.Infof("DNAT %v removal done", dNat.Label)

	return wasErr
}

// ResolveCreatedInterface looks for cache of interfaces which should be enabled or disabled
// for NAT
func (plugin *NatConfigurator) ResolveCreatedInterface(ifName string, ifIdx uint32) error {
	plugin.log.Debugf("NAT configurator: resolving registered interface %s", ifName)

	var wasErr error

	// Check for interfaces which should be enabled
	var enabledIf []*nat.Nat44Global_NatInterface
	for cachedName, data := range plugin.notEnabledIfs {
		if cachedName == ifName {
			delete(plugin.notEnabledIfs, cachedName)
			if err := plugin.enableNatInterfaces(append(enabledIf, data)); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
		}
	}
	// Check for interfaces which could be disabled
	var disabledIf []*nat.Nat44Global_NatInterface
	for cachedName, data := range plugin.notDisabledIfs {
		if cachedName == ifName {
			delete(plugin.notDisabledIfs, cachedName)
			if err := plugin.disableNatInterfaces(append(disabledIf, data)); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
		}
	}

	return wasErr
}

// ResolveDeletedInterface handles removed interface from NAT perspective
func (plugin *NatConfigurator) ResolveDeletedInterface(ifName string, ifIdx uint32) error {
	plugin.log.Debugf("NAT configurator: resolving deleted interface %s", ifName)

	// Check global NAT for interfaces
	if plugin.globalNAT != nil {
		for _, natIf := range plugin.globalNAT.NatInterfaces {
			if natIf.Name == ifName {
				// This interface was removed and it is not possible to determine its state, so agent handles it as
				// not enabled
				plugin.notEnabledIfs[natIf.Name] = natIf
				return nil
			}
		}
	}

	return nil
}

// DumpNatGlobal returns the current NAT44 global config
func (plugin *NatConfigurator) DumpNatGlobal() (*nat.Nat44Global, error) {
	return vppdump.Nat44GlobalConfigDump(plugin.ifIndexes, plugin.log, plugin.vppDumpChan, plugin.stopwatch)
}

// DumpNatDNat returns the current NAT44 DNAT config
func (plugin *NatConfigurator) DumpNatDNat() (*nat.Nat44DNat, error) {
	return vppdump.NAT44DNatDump(plugin.ifIndexes, plugin.log, plugin.vppDumpChan, plugin.stopwatch)
}

// enables set of interfaces as inside/outside in NAT
func (plugin *NatConfigurator) enableNatInterfaces(natInterfaces []*nat.Nat44Global_NatInterface) (err error) {
	for _, natInterface := range natInterfaces {
		ifIdx, _, found := plugin.ifIndexes.LookupIdx(natInterface.Name)
		if !found {
			plugin.log.Debugf("Interface %v missing, cannot enable NAT", natInterface.Name)
			plugin.notEnabledIfs[natInterface.Name] = natInterface // cache interface
		} else {
			if natInterface.OutputFeature {
				// enable nat interface and output feature
				if err = vppcalls.EnableNat44InterfaceOutput(ifIdx, natInterface.IsInside, plugin.vppChan, plugin.stopwatch); err != nil {
					return
				}
				if natInterface.IsInside {
					plugin.log.Debugf("Interface %v output-feature enabled for NAT as inside", natInterface.Name)
				} else {
					plugin.log.Debugf("Interface %v output-feature enabled for NAT as outside", natInterface.Name)
				}
			} else {
				// enable interface only
				if err = vppcalls.EnableNat44Interface(ifIdx, natInterface.IsInside, plugin.vppChan, plugin.stopwatch); err != nil {
					return
				}
				if natInterface.IsInside {
					plugin.log.Debugf("Interface %v enabled for NAT as inside", natInterface.Name)
				} else {
					plugin.log.Debugf("Interface %v enabled for NAT as outside", natInterface.Name)
				}
			}
		}
	}

	return
}

// disables set of interfaces in NAT
func (plugin *NatConfigurator) disableNatInterfaces(natInterfaces []*nat.Nat44Global_NatInterface) (err error) {
	for _, natInterface := range natInterfaces {
		// Check if interface is not in the cache
		for ifName := range plugin.notEnabledIfs {
			if ifName == natInterface.Name {
				delete(plugin.notEnabledIfs, ifName)
			}
		}
		// Check if interface exists
		ifIdx, _, found := plugin.ifIndexes.LookupIdx(natInterface.Name)
		if !found {
			plugin.log.Debugf("Interface %v missing, cannot disable NAT", natInterface)
			plugin.notDisabledIfs[natInterface.Name] = natInterface // cache interface
		} else {
			if natInterface.OutputFeature {
				// disable nat interface and output feature
				if err = vppcalls.DisableNat44InterfaceOutput(ifIdx, natInterface.IsInside, plugin.vppChan, plugin.stopwatch); err != nil {
					return
				}
				if natInterface.IsInside {
					plugin.log.Debugf("Interface %v output-feature disabled for NAT as inside", natInterface.Name)
				} else {
					plugin.log.Debugf("Interface %v output-feature disabled for NAT as outside", natInterface.Name)
				}
			} else {
				// disable interface
				if err = vppcalls.DisableNat44Interface(ifIdx, natInterface.IsInside, plugin.vppChan, plugin.stopwatch); err != nil {
					return
				}
				if natInterface.IsInside {
					plugin.log.Debugf("Interface %v disabled for NAT as inside", natInterface)
				} else {
					plugin.log.Debugf("Interface %v disabled for NAT as outside", natInterface)
				}
			}
		}
	}

	return
}

// adds NAT address pool
func (plugin *NatConfigurator) addDelAddressPool(addressPools []*nat.Nat44Global_AddressPool, isAdd bool) (err error) {
	var wasErr error
	for _, addressPool := range addressPools {
		if addressPool.FirstSrcAddress == "" && addressPool.LastSrcAddress == "" {
			err := fmt.Errorf("invalid address pool config, no IP address provided")
			plugin.log.Error(err)
			wasErr = err
			continue
		}
		var firstIP []byte
		var lastIP []byte
		if addressPool.FirstSrcAddress != "" {
			firstIP = net.ParseIP(addressPool.FirstSrcAddress).To4()
			if firstIP == nil {
				err := fmt.Errorf("unable to parse IP address %v", addressPool.FirstSrcAddress)
				plugin.log.Error(err)
				wasErr = err
				continue
			}
		}
		if addressPool.LastSrcAddress != "" {
			lastIP = net.ParseIP(addressPool.LastSrcAddress).To4()
			if lastIP == nil {
				err := fmt.Errorf("unable to parse IP address %v", addressPool.LastSrcAddress)
				plugin.log.Error(err)
				wasErr = err
				continue
			}
		}
		// Both fields have to be set, at least at the same value if only one of them is set
		if firstIP == nil {
			firstIP = lastIP
		} else if lastIP == nil {
			lastIP = firstIP
		}

		// configure or remove address pool
		if isAdd {
			if err = vppcalls.AddNat44AddressPool(firstIP, lastIP, addressPool.VrfId, addressPool.TwiceNat, plugin.vppChan, plugin.stopwatch); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
		} else {
			if err = vppcalls.DelNat44AddressPool(firstIP, lastIP, addressPool.VrfId, addressPool.TwiceNat, plugin.vppChan, plugin.stopwatch); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
		}
	}

	return wasErr
}

// configures a list of static mappings for provided label
func (plugin *NatConfigurator) configureStaticMappings(label string, mappings []*nat.Nat44DNat_DNatConfig_StaticMapping) error {
	var wasErr error
	for _, mappingEntry := range mappings {
		var tag string
		localIPList := mappingEntry.LocalIps
		if len(localIPList) == 0 {
			wasErr = fmt.Errorf("cannot configure DNAT static mapping: no local address provided")
			plugin.log.Error(wasErr)
			continue
		} else if len(localIPList) == 1 {
			// Case without load balance (one local address)
			tag = plugin.getMappingTag(label, static)
			if err := plugin.handleStaticMapping(mappingEntry, tag, true); err != nil {
				wasErr = fmt.Errorf("DNAT static mapping configuration failed: %v", err)
				plugin.log.Error(wasErr)
				continue
			}
		} else {
			// Case with load balance (more local addresses)
			tag = plugin.getMappingTag(label, staticLb)
			if err := plugin.handleStaticMappingLb(mappingEntry, tag, true); err != nil {
				wasErr = fmt.Errorf("DNAT static lb-mapping configuration failed: %v", err)
				plugin.log.Error(wasErr)
				continue
			}
		}
		// Register DNAT static mapping
		mappingIdentifier := GetStMappingIdentifier(mappingEntry)
		plugin.dNatStMappingIndexes.RegisterName(mappingIdentifier, plugin.natIndexSeq, nil)
		plugin.natIndexSeq++

		plugin.log.Debugf("DNAT static (lb)mapping configured (ID: %s, Tag: %s)", mappingIdentifier, tag)
	}

	return wasErr
}

// removes static mappings from configuration with provided label
func (plugin *NatConfigurator) unconfigureStaticMappings(mappings []*nat.Nat44DNat_DNatConfig_StaticMapping) error {
	var wasErr error
	for mappingIdx, mappingEntry := range mappings {
		localIPList := mappingEntry.LocalIps
		if len(localIPList) == 0 {
			plugin.log.Warnf("DNAT mapping %v has not local IPs, cannot remove it", mappingIdx)
			continue
		} else if len(localIPList) == 1 {
			// Case without load balance (one local address)
			if err := plugin.handleStaticMapping(mappingEntry, dummyTag, false); err != nil {
				wasErr = fmt.Errorf("DNAT mapping removal failed: %v", err)
				plugin.log.Error(wasErr)
				continue
			}
		} else {
			// Case with load balance (more local addresses)
			if err := plugin.handleStaticMappingLb(mappingEntry, dummyTag, false); err != nil {
				wasErr = fmt.Errorf("DNAT lb-mapping removal failed: %v", err)
				plugin.log.Error(wasErr)
				continue
			}
		}
		// Unregister DNAT mapping
		mappingIdentifier := GetStMappingIdentifier(mappingEntry)
		plugin.dNatStMappingIndexes.UnregisterName(mappingIdentifier)

		plugin.log.Debugf("DNAT lb-mapping un-configured (ID %v)", mappingIdentifier)
	}

	return wasErr
}

// configures single static mapping entry with load balancer
func (plugin *NatConfigurator) handleStaticMappingLb(staticMappingLb *nat.Nat44DNat_DNatConfig_StaticMapping, tag string, add bool) (err error) {
	// Validate tag
	if tag == dummyTag && add {
		plugin.log.Warn("Static mapping will be configured with generic tag")
	}
	// Parse external IP address
	exIPAddrByte := net.ParseIP(staticMappingLb.ExternalIp).To4()
	if exIPAddrByte == nil {
		return fmt.Errorf("cannot configure DNAT mapping: unable to parse external IP %v", staticMappingLb.ExternalIp)
	}

	// In this case, external port is required
	if staticMappingLb.ExternalPort == 0 {
		return fmt.Errorf("cannot configure DNAT mapping: external port is not set")
	}

	// Address mapping with load balancer
	ctx := &vppcalls.StaticMappingLbContext{
		Tag:          tag,
		ExternalIP:   exIPAddrByte,
		ExternalPort: uint16(staticMappingLb.ExternalPort),
		Protocol:     getProtocol(staticMappingLb.Protocol, plugin.log),
		LocalIPs:     getLocalIPs(staticMappingLb.LocalIps, plugin.log),
		Vrf:          staticMappingLb.VrfId,
		TwiceNat:     staticMappingLb.TwiceNat == nat.TwiceNatMode_ENABLED,
		SelfTwiceNat: staticMappingLb.TwiceNat == nat.TwiceNatMode_SELF,
	}

	if len(ctx.LocalIPs) == 0 {
		return fmt.Errorf("cannot configure DNAT mapping: no local IP was successfully parsed")
	}

	if add {
		return vppcalls.AddNat44StaticMappingLb(ctx, plugin.vppChan, plugin.stopwatch)
	}
	return vppcalls.DelNat44StaticMappingLb(ctx, plugin.vppChan, plugin.stopwatch)
}

// handler for single static mapping entry
func (plugin *NatConfigurator) handleStaticMapping(staticMapping *nat.Nat44DNat_DNatConfig_StaticMapping, tag string, add bool) (err error) {
	var ifIdx uint32 = 0xffffffff // default value - means no external interface is set
	var exIPAddr []byte

	// Validate tag
	if tag == dummyTag && add {
		plugin.log.Warn("Static mapping will be configured with generic tag")
	}

	// Parse local IP address and port
	lcIPAddr := net.ParseIP(staticMapping.LocalIps[0].LocalIp).To4()
	lcPort := staticMapping.LocalIps[0].LocalPort
	if lcIPAddr == nil {
		return fmt.Errorf("cannot configure DNAT mapping: unable to parse local IP %v", lcIPAddr)
	}

	// Check external interface (prioritized over external IP)
	if staticMapping.ExternalInterface != "" {
		// Check external interface
		var found bool
		ifIdx, _, found = plugin.ifIndexes.LookupIdx(staticMapping.ExternalInterface)
		if !found {
			return fmt.Errorf("required external interface %v is missing", staticMapping.ExternalInterface)
		}
	} else {
		// Parse external IP address
		exIPAddr = net.ParseIP(staticMapping.ExternalIp).To4()
		if exIPAddr == nil {
			return fmt.Errorf("cannot configure DNAT mapping: unable to parse external IP %v", exIPAddr)
		}
	}

	// Resolve mapping (address only or address and port)
	var addrOnly bool
	if lcPort == 0 || staticMapping.ExternalPort == 0 {
		addrOnly = true
	}

	// Address mapping with load balancer
	ctx := &vppcalls.StaticMappingContext{
		Tag:           tag,
		AddressOnly:   addrOnly,
		LocalIP:       lcIPAddr,
		LocalPort:     uint16(lcPort),
		ExternalIP:    exIPAddr,
		ExternalPort:  uint16(staticMapping.ExternalPort),
		ExternalIfIdx: ifIdx,
		Protocol:      getProtocol(staticMapping.Protocol, plugin.log),
		Vrf:           staticMapping.VrfId,
		TwiceNat:      staticMapping.TwiceNat == nat.TwiceNatMode_ENABLED,
		SelfTwiceNat:  staticMapping.TwiceNat == nat.TwiceNatMode_SELF,
	}

	if add {
		return vppcalls.AddNat44StaticMapping(ctx, plugin.vppChan, plugin.stopwatch)
	}
	return vppcalls.DelNat44StaticMapping(ctx, plugin.vppChan, plugin.stopwatch)
}

// configures a list of identity mappings with label
func (plugin *NatConfigurator) configureIdentityMappings(label string, mappings []*nat.Nat44DNat_DNatConfig_IdentityMapping) error {
	var wasErr error
	for _, idMapping := range mappings {
		if idMapping.IpAddress == "" && idMapping.AddressedInterface == "" {
			wasErr = fmt.Errorf("cannot configure DNAT identity mapping: no IP address or interface provided")
			plugin.log.Error(wasErr)
			continue
		}
		// Case without load balance (one local address)
		tag := plugin.getMappingTag(label, identity)
		if err := plugin.handleIdentityMapping(idMapping, tag, true); err != nil {
			wasErr = err
			plugin.log.Error(err)
			continue
		}

		// Register DNAT identity mapping
		mappingIdentifier := GetIdMappingIdentifier(idMapping)
		plugin.dNatIdMappingIndexes.RegisterName(mappingIdentifier, plugin.natIndexSeq, nil)
		plugin.natIndexSeq++

		plugin.log.Debugf("DNAT identity mapping configured (ID: %s, Tag: %s)", mappingIdentifier, tag)
	}

	return wasErr
}

// removes identity mappings from configuration with provided label
func (plugin *NatConfigurator) unconfigureIdentityMappings(mappings []*nat.Nat44DNat_DNatConfig_IdentityMapping) error {
	var wasErr error
	for mappingIdx, idMapping := range mappings {
		if idMapping.IpAddress == "" && idMapping.AddressedInterface == "" {
			wasErr = fmt.Errorf("cannot remove DNAT %v identity mapping: no IP address or interface provided", mappingIdx)
			plugin.log.Error(wasErr)
			continue
		}
		if err := plugin.handleIdentityMapping(idMapping, dummyTag, false); err != nil {
			wasErr = err
			plugin.log.Error(err)
			continue
		}

		// Unregister DNAT identity mapping
		mappingIdentifier := GetIdMappingIdentifier(idMapping)
		plugin.dNatIdMappingIndexes.UnregisterName(mappingIdentifier)
		plugin.natIndexSeq++

		plugin.log.Debugf("DNAT identity (lb)mapping un-configured (ID: %v)", mappingIdentifier)
	}

	return wasErr
}

// handler for single identity mapping entry
func (plugin *NatConfigurator) handleIdentityMapping(idMapping *nat.Nat44DNat_DNatConfig_IdentityMapping, tag string, isAdd bool) (err error) {
	// Verify interface if exists
	var ifIdx uint32
	if idMapping.AddressedInterface != "" {
		var found bool
		ifIdx, _, found = plugin.ifIndexes.LookupIdx(idMapping.AddressedInterface)
		if !found {
			// TODO: use cache to configure later
			return fmt.Errorf("identity mapping config: provided interface %v does not exist", idMapping.AddressedInterface)
		}
	}

	// Identity mapping (common fields)
	ctx := &vppcalls.IdentityMappingContext{
		Tag:      tag,
		Protocol: getProtocol(idMapping.Protocol, plugin.log),
		Port:     uint16(idMapping.Port),
		IfIdx:    ifIdx,
		Vrf:      idMapping.VrfId,
	}

	if ctx.IfIdx == 0 {
		// Case with IP (optionally port). Verify and parse input IP/port
		parsedIP := net.ParseIP(idMapping.IpAddress).To4()
		if parsedIP == nil {
			return fmt.Errorf("unable to parse IP address %v", idMapping.IpAddress)
		}
		// Add IP address to context
		ctx.IPAddress = parsedIP
	}

	// Configure/remove identity mapping
	if isAdd {
		return vppcalls.AddNat44IdentityMapping(ctx, plugin.vppChan, plugin.stopwatch)
	}
	return vppcalls.DelNat44IdentityMapping(ctx, plugin.vppChan, plugin.stopwatch)
}

// looks for new and obsolete IN interfaces
func diffInterfaces(oldIfs, newIfs []*nat.Nat44Global_NatInterface) (toSetIn, toSetOut, toUnsetIn, toUnsetOut []*nat.Nat44Global_NatInterface) {
	// Find new interfaces
	for _, newIf := range newIfs {
		var found bool
		for _, oldIf := range oldIfs {
			if newIf.Name == oldIf.Name && newIf.IsInside == oldIf.IsInside && newIf.OutputFeature == oldIf.OutputFeature {
				found = true
				break
			}
		}
		if !found {
			if newIf.IsInside {
				toSetIn = append(toSetIn, newIf)
			} else {
				toSetOut = append(toSetOut, newIf)
			}
		}
	}
	// Find obsolete interfaces
	for _, oldIf := range oldIfs {
		var found bool
		for _, newIf := range newIfs {
			if oldIf.Name == newIf.Name && oldIf.IsInside == newIf.IsInside && oldIf.OutputFeature == newIf.OutputFeature {
				found = true
				break
			}
		}
		if !found {
			if oldIf.IsInside {
				toUnsetIn = append(toUnsetIn, oldIf)
			} else {
				toUnsetOut = append(toUnsetOut, oldIf)
			}
		}
	}

	return
}

// looks for new and obsolete address pools
func diffAddressPools(oldAPs, newAPs []*nat.Nat44Global_AddressPool) (toAdd, toRemove []*nat.Nat44Global_AddressPool) {
	// Find new address pools
	for _, newAp := range newAPs {
		// If new address pool is a range, add it
		if newAp.LastSrcAddress != "" {
			toAdd = append(toAdd, newAp)
			continue
		}
		// Otherwise try to find the same address pool
		var found bool
		for _, oldAp := range oldAPs {
			// Skip address pools
			if oldAp.LastSrcAddress != "" {
				continue
			}
			if newAp.FirstSrcAddress == oldAp.FirstSrcAddress && newAp.TwiceNat == oldAp.TwiceNat && newAp.VrfId == oldAp.VrfId {
				found = true
			}
		}
		if !found {
			toAdd = append(toAdd, newAp)
		}
	}
	// Find obsolete address pools
	for _, oldAp := range oldAPs {
		// If new address pool is a range, remove it
		if oldAp.LastSrcAddress != "" {
			toRemove = append(toRemove, oldAp)
			continue
		}
		// Otherwise try to find the same address pool
		var found bool
		for _, newAp := range newAPs {
			// Skip address pools (they are already added)
			if oldAp.LastSrcAddress != "" {
				continue
			}
			if oldAp.FirstSrcAddress == newAp.FirstSrcAddress && oldAp.TwiceNat == newAp.TwiceNat && oldAp.VrfId == newAp.VrfId {
				found = true
			}
		}
		if !found {
			toRemove = append(toRemove, oldAp)
		}
	}

	return
}

// returns a list of static mappings to add/remove
func (plugin *NatConfigurator) diffStatic(oldMappings, newMappings []*nat.Nat44DNat_DNatConfig_StaticMapping) (toAdd, toRemove []*nat.Nat44DNat_DNatConfig_StaticMapping) {
	// Find missing mappings
	for _, newMap := range newMappings {
		var found bool
		for _, oldMap := range oldMappings {
			// VRF, protocol and twice map
			if newMap.VrfId != oldMap.VrfId || newMap.Protocol != oldMap.Protocol || newMap.TwiceNat != oldMap.TwiceNat {
				continue
			}
			// External interface, IP and port
			if newMap.ExternalInterface != oldMap.ExternalInterface || newMap.ExternalIp != oldMap.ExternalIp ||
				newMap.ExternalPort != oldMap.ExternalPort {
				continue
			}
			// Local IPs
			if !plugin.compareLocalIPs(oldMap.LocalIps, newMap.LocalIps) {
				continue
			}
			found = true
		}
		if !found {
			toAdd = append(toAdd, newMap)
		}
	}
	// Find obsolete mappings
	for _, oldMap := range oldMappings {
		var found bool
		for _, newMap := range newMappings {
			// VRF, protocol and twice map
			if newMap.VrfId != oldMap.VrfId || newMap.Protocol != oldMap.Protocol || newMap.TwiceNat != oldMap.TwiceNat {
				continue
			}
			// External interface, IP and port
			if newMap.ExternalInterface != oldMap.ExternalInterface || newMap.ExternalIp != oldMap.ExternalIp ||
				newMap.ExternalPort != oldMap.ExternalPort {
				continue
			}
			// Local IPs
			if !plugin.compareLocalIPs(oldMap.LocalIps, newMap.LocalIps) {
				continue
			}
			found = true
		}
		if !found {
			toRemove = append(toRemove, oldMap)
		}
	}

	return
}

// returns a list of identity mappings to add/remove
func (plugin *NatConfigurator) diffIdentity(oldMappings, newMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping) (toAdd, toRemove []*nat.Nat44DNat_DNatConfig_IdentityMapping) {
	// Find missing mappings
	for _, newMap := range newMappings {
		var found bool
		for _, oldMap := range oldMappings {
			// VRF and protocol
			if newMap.VrfId != oldMap.VrfId || newMap.Protocol != oldMap.Protocol {
				continue
			}
			// Addressed interface, IP address and port
			if newMap.AddressedInterface != oldMap.AddressedInterface || newMap.IpAddress != oldMap.IpAddress ||
				newMap.Port != oldMap.Port {
				continue
			}
			found = true
		}
		if !found {
			toAdd = append(toAdd, newMap)
		}
	}
	// Find obsolete mappings
	for _, oldMap := range oldMappings {
		var found bool
		for _, newMap := range newMappings {
			// VRF and protocol
			if newMap.VrfId != oldMap.VrfId || newMap.Protocol != oldMap.Protocol {
				continue
			}
			// Addressed interface, IP address and port
			if newMap.AddressedInterface != oldMap.AddressedInterface || newMap.IpAddress != oldMap.IpAddress ||
				newMap.Port != oldMap.Port {
				continue
			}
			found = true
		}
		if !found {
			toRemove = append(toRemove, oldMap)
		}
	}

	return
}

// comapares two lists of Local IP addresses, returns true if lists are equal, false otherwise
func (plugin *NatConfigurator) compareLocalIPs(oldIPs, newIPs []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP) bool {
	if len(oldIPs) != len(newIPs) {
		return false
	}
	for _, newIP := range newIPs {
		var found bool
		for _, oldIP := range oldIPs {
			if newIP.LocalIp == oldIP.LocalIp && newIP.LocalPort == oldIP.LocalPort && newIP.Probability == oldIP.Probability {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	// Do not need to compare old vs. new if length is the same
	return true
}

// returns a list of validated local IP addresses with port and probability value
func getLocalIPs(ipPorts []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP, log logging.Logger) (locals []*vppcalls.LocalLbAddress) {
	for _, ipPort := range ipPorts {
		if ipPort.LocalPort == 0 {
			log.Error("cannot set local IP/Port to mapping: port is missing")
			continue
		}

		localIP := net.ParseIP(ipPort.LocalIp).To4()
		if localIP == nil {
			log.Errorf("cannot set local IP/Port to mapping: unable to parse local IP %v", ipPort.LocalIp)
			continue
		}

		locals = append(locals, &vppcalls.LocalLbAddress{
			LocalIP:     localIP,
			LocalPort:   uint16(ipPort.LocalPort),
			Probability: uint8(ipPort.Probability),
		})
	}

	return
}

// returns num representation of provided protocol value
func getProtocol(protocol nat.Protocol, log logging.Logger) uint8 {
	switch protocol {
	case nat.Protocol_TCP:
		return vppcalls.TCP
	case nat.Protocol_UDP:
		return vppcalls.UDP
	case nat.Protocol_ICMP:
		return vppcalls.ICMP
	default:
		log.Warnf("Unknown protocol %v, defaulting to TCP", protocol)
		return vppcalls.TCP
	}
}

// GetStMappingIdentifier returns unique ID of the mapping
func GetStMappingIdentifier(mapping *nat.Nat44DNat_DNatConfig_StaticMapping) string {
	extIP := strings.Replace(mapping.ExternalIp, ".", "", -1)
	extIP = strings.Replace(extIP, "/", "", -1)
	locIP := strings.Replace(mapping.LocalIps[0].LocalIp, ".", "", -1)
	locIP = strings.Replace(locIP, "/", "", -1)
	return extIP + locIP + strconv.Itoa(int(mapping.VrfId))
}

// GetIdMappingIdentifier returns unique ID of the mapping
func GetIdMappingIdentifier(mapping *nat.Nat44DNat_DNatConfig_IdentityMapping) string {
	extIP := strings.Replace(mapping.IpAddress, ".", "", -1)
	extIP = strings.Replace(extIP, "/", "", -1)
	if mapping.AddressedInterface == "" {
		return extIP + "-noif-" + strconv.Itoa(int(mapping.VrfId))
	}
	return extIP + "-" + mapping.AddressedInterface + "-" + strconv.Itoa(int(mapping.VrfId))
}

// returns unique mapping tag
func (plugin *NatConfigurator) getMappingTag(label, mType string) string {
	var buffer bytes.Buffer
	buffer.WriteString(label)
	buffer.WriteString(mType)
	buffer.WriteString(strconv.Itoa(int(plugin.natMappingTagSeq)))
	plugin.natMappingTagSeq++

	return buffer.String()
}
