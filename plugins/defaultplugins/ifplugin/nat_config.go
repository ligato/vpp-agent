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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
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
	Log logging.Logger

	GoVppmux    govppmux.API
	SwIfIndexes ifaceidx.SwIfIndex
	// SNAT config indices
	SNatIndices idxvpp.NameToIdxRW
	// SNAT indices for static mapping
	SNatMappingIndices idxvpp.NameToIdxRW
	// DNAT indices
	DNatIndices idxvpp.NameToIdxRW
	// DNAT indices for static/identity mapping
	DNatStMappingIndices idxvpp.NameToIdxRW
	DNatIdMappingIndices idxvpp.NameToIdxRW

	// Nat name-to-idx mapping sequence
	NatIndexSeq uint32
	// Static/identity mapping tag sequence
	NatMappingTagSeq uint32
	vppChan          *govppapi.Channel
	vppDumpChan      *govppapi.Channel

	// a map of missing interfaces which should be enabled for NAT (format ifName/isInside)
	notEnabledIfs []*nat.Nat44Global_NatInterfaces
	// a map of NAT-enabled interfaces which should be disabled (format ifName/isInside)
	notDisabledIfs []*nat.Nat44Global_NatInterfaces

	Stopwatch *measure.Stopwatch
}

// Init NAT configurator
func (plugin *NatConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing NAT configurator")

	// Init VPP API channel
	if plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}
	if plugin.vppDumpChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
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

// SetNatGlobalConfig configures common setup for all NAT use cases
func (plugin *NatConfigurator) SetNatGlobalConfig(config *nat.Nat44Global) error {
	plugin.Log.Info("Setting up NAT global config")

	// Forwarding
	if err := vppcalls.SetNat44Forwarding(config.Forwarding, plugin.vppChan, plugin.Stopwatch); err != nil {
		return err
	}
	if config.Forwarding {
		plugin.Log.Debugf("NAT forwarding enabled")
	} else {
		plugin.Log.Debugf("NAT forwarding disabled")
	}

	// Inside/outside interfaces
	if len(config.NatInterfaces) > 0 {
		if err := plugin.enableNatInterfaces(config.NatInterfaces); err != nil {
			return err
		}
	} else {
		plugin.Log.Debug("No NAT interfaces to configure")
	}

	// Address pool
	for _, pool := range config.AddressPools {
		if pool.FirstSrcAddress == "" && pool.LastSrcAddress == "" {
			plugin.Log.Warn("Invalid address pool config, no IP address provided")
			continue
		}
		var firstIP []byte
		var lastIP []byte
		if pool.FirstSrcAddress != "" {
			firstIP = net.ParseIP(pool.FirstSrcAddress).To4()
			if firstIP == nil {
				plugin.Log.Errorf("unable to parse IP address %v", pool.FirstSrcAddress)
				continue
			}
		}
		if pool.LastSrcAddress != "" {
			lastIP = net.ParseIP(pool.LastSrcAddress).To4()
			if lastIP == nil {
				plugin.Log.Errorf("unable to parse IP address %v", pool.LastSrcAddress)
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
		if err := vppcalls.AddNat44AddressPool(firstIP, lastIP, pool.VrfId, pool.TwiceNat, plugin.vppChan, plugin.Stopwatch); err != nil {
			return err
		}
	}

	plugin.Log.Debug("Setting up NAT global config done")

	return nil
}

// ModifyNatGlobalConfig modifies common setup for all NAT use cases
func (plugin *NatConfigurator) ModifyNatGlobalConfig(oldConfig, newConfig *nat.Nat44Global) (err error) {
	plugin.Log.Info("Modifying NAT global config")

	// Forwarding
	if oldConfig.Forwarding != newConfig.Forwarding {
		if err := vppcalls.SetNat44Forwarding(newConfig.Forwarding, plugin.vppChan, plugin.Stopwatch); err != nil {
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
	if err := plugin.deleteAddressPool(toRemove); err != nil {
		return err
	}
	if err := plugin.addAddressPool(toAdd); err != nil {
		return err
	}

	plugin.Log.Debug("Modifying NAT global config done")

	return nil
}

// DeleteNatGlobalConfig removes common setup for all NAT use cases
func (plugin *NatConfigurator) DeleteNatGlobalConfig(config *nat.Nat44Global) (err error) {
	plugin.Log.Info("Deleting NAT global config")

	// Inside/outside interfaces
	if len(config.NatInterfaces) > 0 {
		if err := plugin.disableNatInterfaces(config.NatInterfaces); err != nil {
			return err
		}
	}

	// Address pools
	if len(config.AddressPools) > 0 {
		if err := plugin.deleteAddressPool(config.AddressPools); err != nil {
			return err
		}
	}

	plugin.Log.Debug("Deleting NAT global config done")

	return nil
}

// ConfigureSNat configures new SNAT setup
func (plugin *NatConfigurator) ConfigureSNat(sNat *nat.Nat44SNat_SNatConfig) error {
	plugin.Log.Warn("SNAT CREATE not implemented")
	return nil
}

// ModifySNat modifies existing SNAT setup
func (plugin *NatConfigurator) ModifySNat(oldSNat, newSNat *nat.Nat44SNat_SNatConfig) error {
	plugin.Log.Warn("SNAT MODIFY not implemented")
	return nil
}

// DeleteSNat removes existing SNAT setup
func (plugin *NatConfigurator) DeleteSNat(sNat *nat.Nat44SNat_SNatConfig) error {
	plugin.Log.Warn("SNAT DELETE not implemented")
	return nil
}

// ConfigureDNat configures new DNAT setup
func (plugin *NatConfigurator) ConfigureDNat(dNat *nat.Nat44DNat_DNatConfig) error {
	plugin.Log.Infof("Configuring DNAT with label %v", dNat.Label)

	var wasErr error

	// Resolve static mapping
	if err := plugin.configureStaticMappings(dNat.Label, dNat.StMappings); err != nil {
		wasErr = err
		plugin.Log.Errorf("Failed to configure static mapping for DNAT %s: %v", dNat.Label, err)
	}

	// Resolve identity mapping
	if err := plugin.configureIdentityMappings(dNat.Label, dNat.IdMappings); err != nil {
		wasErr = err
		plugin.Log.Errorf("Failed to configure identity mapping for DNAT %s: %v", dNat.Label, err)
	}

	// Register DNAT configuration
	plugin.DNatIndices.RegisterName(dNat.Label, plugin.NatIndexSeq, nil)
	plugin.NatIndexSeq++
	plugin.Log.Debugf("DNAT configuration registered (label: %v)", dNat.Label)

	plugin.Log.Infof("DNAT %v configuration done", dNat.Label)

	return wasErr
}

// ModifyDNat modifies existing DNAT setup
func (plugin *NatConfigurator) ModifyDNat(oldDNat, newDNat *nat.Nat44DNat_DNatConfig) error {
	plugin.Log.Infof("Modifying DNAT with label %v", newDNat.Label)

	var wasErr error

	// Static mappings
	stmToAdd, stmToRemove := plugin.diffStatic(oldDNat.StMappings, newDNat.StMappings)

	if err := plugin.unconfigureStaticMappings(stmToRemove); err != nil {
		wasErr = err
		plugin.Log.Errorf("Failed to remove static mapping from DNAT %s: %v", newDNat.Label, err)
	}

	if err := plugin.configureStaticMappings(newDNat.Label, stmToAdd); err != nil {
		wasErr = err
		plugin.Log.Errorf("Failed to configure static mapping for DNAT %s: %v", newDNat.Label, err)
	}

	// Identity mappings
	idToAdd, idToRemove := plugin.diffIdentity(oldDNat.IdMappings, newDNat.IdMappings)

	if err := plugin.unconfigureIdentityMappings(idToRemove); err != nil {
		wasErr = err
		plugin.Log.Errorf("Failed to remove identity mapping from DNAT %s: %v", newDNat.Label, err)
	}

	if err := plugin.configureIdentityMappings(newDNat.Label, idToAdd); err != nil {
		wasErr = err
		plugin.Log.Errorf("Failed to configure identity mapping for DNAT %s: %v", newDNat.Label, err)
	}

	plugin.Log.Infof("DNAT %v modification done", newDNat.Label)

	return wasErr
}

// DeleteDNat removes existing DNAT setup
func (plugin *NatConfigurator) DeleteDNat(dNat *nat.Nat44DNat_DNatConfig) error {
	plugin.Log.Infof("Deleting DNAT with label %v", dNat.Label)

	var wasErr error

	// In delete case, vpp-agent attempts to reconstruct every static mapping entry and remove it from the VPP
	if err := plugin.unconfigureStaticMappings(dNat.StMappings); err != nil {
		wasErr = err
		plugin.Log.Errorf("Failed to remove static mapping from DNAT %s: %v", dNat.Label, err)
	}

	// Do the same also for identity apping
	if err := plugin.unconfigureIdentityMappings(dNat.IdMappings); err != nil {
		wasErr = err
		plugin.Log.Errorf("Failed to remove identity mapping from DNAT %s: %v", dNat.Label, err)
	}

	// Unregister DNAT configuration
	plugin.DNatIndices.UnregisterName(dNat.Label)
	plugin.Log.Debugf("DNAT configuration unregistered (label: %v)", dNat.Label)

	plugin.Log.Infof("DNAT %v removal done", dNat.Label)

	return wasErr
}

// DumpNatGlobal returns the current NAT44 global config
func (plugin *NatConfigurator) DumpNatGlobal() (*nat.Nat44Global, error) {
	return vppdump.Nat44GlobalConfigDump(plugin.SwIfIndexes, plugin.Log, plugin.vppDumpChan, plugin.Stopwatch)
}

// DumpNatDNat returns the current NAT44 DNAT config
func (plugin *NatConfigurator) DumpNatDNat() (*nat.Nat44DNat, error) {
	return vppdump.NAT44DNatDump(plugin.SwIfIndexes, plugin.Log, plugin.vppDumpChan, plugin.Stopwatch)
}

// enables set of interfaces as inside/outside in NAT
func (plugin *NatConfigurator) enableNatInterfaces(natInterfaces []*nat.Nat44Global_NatInterfaces) (err error) {
	for _, natInterface := range natInterfaces {
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(natInterface.Name)
		if !found {
			plugin.Log.Debugf("Interface %v missing, cannot enable NAT", natInterface.Name)
			plugin.notEnabledIfs = append(plugin.notEnabledIfs, natInterface) // cache interface
		} else {
			if natInterface.OutputFeature {
				// enable nat interface and output feature
				if err = vppcalls.EnableNat44InterfaceOutput(ifIdx, natInterface.IsInside, plugin.vppChan, plugin.Stopwatch); err != nil {
					return
				}
				if natInterface.IsInside {
					plugin.Log.Debugf("Interface %v output-feature enabled for NAT as inside", natInterface.Name)
				} else {
					plugin.Log.Debugf("Interface %v output-feature enabled for NAT as outside", natInterface.Name)
				}
			} else {
				// enable interface only
				if err = vppcalls.EnableNat44Interface(ifIdx, natInterface.IsInside, plugin.vppChan, plugin.Stopwatch); err != nil {
					return
				}
				if natInterface.IsInside {
					plugin.Log.Debugf("Interface %v enabled for NAT as inside", natInterface.Name)
				} else {
					plugin.Log.Debugf("Interface %v enabled for NAT as outside", natInterface.Name)
				}
			}
		}
	}

	return
}

// disables set of interfaces in NAT
func (plugin *NatConfigurator) disableNatInterfaces(natInterfaces []*nat.Nat44Global_NatInterfaces) (err error) {
	for _, natInterface := range natInterfaces {
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(natInterface.Name)
		if !found {
			plugin.Log.Debugf("Interface %v missing, cannot disable NAT", natInterface)
			plugin.notDisabledIfs = append(plugin.notDisabledIfs, natInterface) // cache interface
		} else {
			if natInterface.OutputFeature {
				// disable nat interface and output feature
				if err = vppcalls.DisableNat44InterfaceOutput(ifIdx, natInterface.IsInside, plugin.vppChan, plugin.Stopwatch); err != nil {
					return
				}
				if natInterface.IsInside {
					plugin.Log.Debugf("Interface %v output-feature disabled for NAT as inside", natInterface.Name)
				} else {
					plugin.Log.Debugf("Interface %v output-feature disabled for NAT as outside", natInterface.Name)
				}
			} else {
				// disable interface
				if err = vppcalls.DisableNat44Interface(ifIdx, natInterface.IsInside, plugin.vppChan, plugin.Stopwatch); err != nil {
					return
				}
				if natInterface.IsInside {
					plugin.Log.Debugf("Interface %v disabled for NAT as inside", natInterface)
				} else {
					plugin.Log.Debugf("Interface %v disabled for NAT as outside", natInterface)
				}
			}
		}
	}

	return
}

// adds NAT address pool
func (plugin *NatConfigurator) addAddressPool(addressPools []*nat.Nat44Global_AddressPools) (err error) {
	for _, addressPool := range addressPools {
		if addressPool.FirstSrcAddress == "" && addressPool.LastSrcAddress == "" {
			plugin.Log.Warn("Invalid address pool config, no IP address provided")
			continue
		}
		var firstIP []byte
		var lastIP []byte
		if addressPool.FirstSrcAddress != "" {
			firstIP = net.ParseIP(addressPool.FirstSrcAddress).To4()
			if firstIP == nil {
				plugin.Log.Errorf("unable to parse IP address %v", addressPool.FirstSrcAddress)
				continue
			}
		}
		if addressPool.LastSrcAddress != "" {
			lastIP = net.ParseIP(addressPool.LastSrcAddress).To4()
			if lastIP == nil {
				plugin.Log.Errorf("unable to parse IP address %v", addressPool.LastSrcAddress)
				continue
			}
		}
		// Both fields have to be set, at least at the same value if only one of them is set
		if firstIP == nil {
			firstIP = lastIP
		} else if lastIP == nil {
			lastIP = firstIP
		}

		// configure address pool
		if err = vppcalls.AddNat44AddressPool(firstIP, lastIP, addressPool.VrfId, addressPool.TwiceNat, plugin.vppChan, plugin.Stopwatch); err != nil {
			return
		}
	}

	return
}

// removes NAT address pool
func (plugin *NatConfigurator) deleteAddressPool(addressPools []*nat.Nat44Global_AddressPools) (err error) {
	for _, addressPool := range addressPools {
		var firstIP []byte
		var lastIP []byte
		if addressPool.FirstSrcAddress != "" {
			firstIP = net.ParseIP(addressPool.FirstSrcAddress).To4()
			if firstIP == nil {
				plugin.Log.Errorf("unable to parse IP address %v", addressPool.FirstSrcAddress)
				continue
			}
		}
		if addressPool.LastSrcAddress != "" {
			lastIP = net.ParseIP(addressPool.LastSrcAddress).To4()
			if lastIP == nil {
				plugin.Log.Errorf("unable to parse IP address %v", addressPool.LastSrcAddress)
				continue
			}
		}
		// Both fields have to be set, at least at the same value if only one of them is set
		if firstIP == nil {
			firstIP = lastIP
		} else if lastIP == nil {
			lastIP = firstIP
		}

		// configure address pool
		if err = vppcalls.DelNat44AddressPool(firstIP, lastIP, addressPool.VrfId, addressPool.TwiceNat, plugin.vppChan, plugin.Stopwatch); err != nil {
			return
		}
	}

	return
}

// configures a list of static mappings for provided label
func (plugin *NatConfigurator) configureStaticMappings(label string, mappings []*nat.Nat44DNat_DNatConfig_StaticMappings) error {
	var wasErr error
	for _, mappingEntry := range mappings {
		var tag string
		localIPList := mappingEntry.LocalIps
		if len(localIPList) == 0 {
			wasErr = fmt.Errorf("cannot configure DNAT static mapping: no local address provided")
			plugin.Log.Error(wasErr)
			continue
		} else if len(localIPList) == 1 {
			// Case without load balance (one local address)
			tag = plugin.getMappingTag(label, static)
			if err := plugin.handleStaticMapping(mappingEntry, tag, true); err != nil {
				wasErr = fmt.Errorf("DNAT static mapping configuration failed: %v", err)
				plugin.Log.Error(wasErr)
				continue
			}
		} else {
			// Case with load balance (more local addresses)
			tag = plugin.getMappingTag(label, staticLb)
			if err := plugin.handleStaticMappingLb(mappingEntry, tag, true); err != nil {
				wasErr = fmt.Errorf("DNAT static lb-mapping configuration failed: %v", err)
				plugin.Log.Error(wasErr)
				continue
			}
		}
		// Register DNAT static mapping
		mappingIdentifier := getStMappingIdentifier(mappingEntry)
		plugin.DNatStMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
		plugin.NatIndexSeq++

		plugin.Log.Debugf("DNAT static (lb)mapping configured (ID: %s, Tag: %s)", mappingIdentifier, tag)
	}

	return wasErr
}

// removes static mappings from configuration with provided label
func (plugin *NatConfigurator) unconfigureStaticMappings(mappings []*nat.Nat44DNat_DNatConfig_StaticMappings) error {
	var wasErr error
	for mappingIdx, mappingEntry := range mappings {
		localIPList := mappingEntry.LocalIps
		if len(localIPList) == 0 {
			plugin.Log.Warnf("DNAT mapping %v has not local IPs, cannot remove it", mappingIdx)
			continue
		} else if len(localIPList) == 1 {
			// Case without load balance (one local address)
			if err := plugin.handleStaticMapping(mappingEntry, dummyTag, false); err != nil {
				wasErr = fmt.Errorf("DNAT mapping removal failed: %v", err)
				plugin.Log.Error(wasErr)
				continue
			}
		} else {
			// Case with load balance (more local addresses)
			if err := plugin.handleStaticMappingLb(mappingEntry, dummyTag, false); err != nil {
				wasErr = fmt.Errorf("DNAT lb-mapping removal failed: %v", err)
				plugin.Log.Error(wasErr)
				continue
			}
		}
		// Unregister DNAT mapping
		mappingIdentifier := getStMappingIdentifier(mappingEntry)
		plugin.DNatStMappingIndices.UnregisterName(mappingIdentifier)

		plugin.Log.Debugf("DNAT lb-mapping un-configured (ID %v)", mappingIdentifier)
	}

	return wasErr
}

// configures single static mapping entry with load balancer
func (plugin *NatConfigurator) handleStaticMappingLb(staticMappingLb *nat.Nat44DNat_DNatConfig_StaticMappings, tag string, add bool) (err error) {
	// Validate tag
	if tag == dummyTag && add {
		plugin.Log.Warn("Static mapping will be configured with generic tag")
	}
	// Parse external IP address
	exIPAddrByte := net.ParseIP(staticMappingLb.ExternalIP).To4()
	if exIPAddrByte == nil {
		return fmt.Errorf("cannot configure DNAT mapping: unable to parse external IP %v", staticMappingLb.ExternalIP)
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
		Protocol:     getProtocol(staticMappingLb.Protocol, plugin.Log),
		LocalIPs:     getLocalIPs(staticMappingLb.LocalIps, plugin.Log),
		Vrf:          staticMappingLb.VrfId,
		TwiceNat:     staticMappingLb.TwiceNat,
	}

	if len(ctx.LocalIPs) == 0 {
		return fmt.Errorf("cannot configure DNAT mapping: no local IP was successfully parsed")
	}

	if add {
		return vppcalls.AddNat44StaticMappingLb(ctx, plugin.vppChan, plugin.Stopwatch)
	}
	return vppcalls.DelNat44StaticMappingLb(ctx, plugin.vppChan, plugin.Stopwatch)
}

// handler for single static mapping entry
func (plugin *NatConfigurator) handleStaticMapping(staticMapping *nat.Nat44DNat_DNatConfig_StaticMappings, tag string, add bool) (err error) {
	var ifIdx uint32 = 0xffffffff // default value - means no external interface is set
	var exIPAddr []byte

	// Validate tag
	if tag == dummyTag && add {
		plugin.Log.Warn("Static mapping will be configured with generic tag")
	}

	// Parse local IP address and port
	lcIPAddr := net.ParseIP(staticMapping.LocalIps[0].LocalIP).To4()
	lcPort := staticMapping.LocalIps[0].LocalPort
	if lcIPAddr == nil {
		return fmt.Errorf("cannot configure DNAT mapping: unable to parse local IP %v", lcIPAddr)
	}

	// Check external interface (prioritized over external IP)
	if staticMapping.ExternalInterface != "" {
		// Check external interface
		var found bool
		ifIdx, _, found = plugin.SwIfIndexes.LookupIdx(staticMapping.ExternalInterface)
		if !found {
			return fmt.Errorf("required external interface %v is missing", staticMapping.ExternalInterface)
		}
	} else {
		// Parse external IP address
		exIPAddr = net.ParseIP(staticMapping.ExternalIP).To4()
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
		Protocol:      getProtocol(staticMapping.Protocol, plugin.Log),
		Vrf:           staticMapping.VrfId,
		TwiceNat:      staticMapping.TwiceNat,
	}

	if add {
		return vppcalls.AddNat44StaticMapping(ctx, plugin.vppChan, plugin.Stopwatch)
	}
	return vppcalls.DelNat44StaticMapping(ctx, plugin.vppChan, plugin.Stopwatch)
}

// configures a list of identity mappings with label
func (plugin *NatConfigurator) configureIdentityMappings(label string, mappings []*nat.Nat44DNat_DNatConfig_IdentityMappings) error {
	var wasErr error
	for _, idMapping := range mappings {
		if idMapping.IpAddress == "" && idMapping.AddressedInterface == "" {
			wasErr = fmt.Errorf("cannot configure DNAT identity mapping: no IP address or interface provided")
			plugin.Log.Error(wasErr)
			continue
		}
		// Case without load balance (one local address)
		tag := plugin.getMappingTag(label, identity)
		if err := plugin.handleIdentityMapping(idMapping, tag, true); err != nil {
			wasErr = err
			plugin.Log.Error(err)
			continue
		}

		// Register DNAT identity mapping
		mappingIdentifier := getIdMappingIdentifier(idMapping)
		plugin.DNatIdMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
		plugin.NatIndexSeq++

		plugin.Log.Debugf("DNAT identity mapping configured (ID: %s, Tag: %s)", mappingIdentifier, tag)
	}

	return wasErr
}

// removes identity mappings from configuration with provided label
func (plugin *NatConfigurator) unconfigureIdentityMappings(mappings []*nat.Nat44DNat_DNatConfig_IdentityMappings) error {
	var wasErr error
	for mappingIdx, idMapping := range mappings {
		if idMapping.IpAddress == "" && idMapping.AddressedInterface == "" {
			wasErr = fmt.Errorf("cannot remove DNAT %v identity mapping: no IP address or interface provided", mappingIdx)
			plugin.Log.Error(wasErr)
			continue
		}
		if err := plugin.handleIdentityMapping(idMapping, dummyTag, false); err != nil {
			wasErr = err
			plugin.Log.Error(err)
			continue
		}

		// Unregister DNAT identity mapping
		mappingIdentifier := getIdMappingIdentifier(idMapping)
		plugin.DNatIdMappingIndices.UnregisterName(mappingIdentifier)
		plugin.NatIndexSeq++

		plugin.Log.Debugf("DNAT identity (lb)mapping un-configured (ID: %v)", mappingIdentifier)
	}

	return wasErr
}

// handler for single identity mapping entry
func (plugin *NatConfigurator) handleIdentityMapping(idMapping *nat.Nat44DNat_DNatConfig_IdentityMappings, tag string, isAdd bool) (err error) {
	// Verify interface if exists
	var ifIdx uint32
	if idMapping.AddressedInterface != "" {
		var found bool
		ifIdx, _, found = plugin.SwIfIndexes.LookupIdx(idMapping.AddressedInterface)
		if !found {
			// TODO: use cache to configure later
			plugin.Log.Warnf("Identity mapping config: provided interface %v does not exist", idMapping.AddressedInterface)
			return
		}
	}

	// Identity mapping (common fields)
	ctx := &vppcalls.IdentityMappingContext{
		Tag:      tag,
		Protocol: getProtocol(idMapping.Protocol, plugin.Log),
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
		return vppcalls.AddNat44IdentityMapping(ctx, plugin.vppChan, plugin.Stopwatch)
	}
	return vppcalls.DelNat44IdentityMapping(ctx, plugin.vppChan, plugin.Stopwatch)
}

// looks for new and obsolete IN interfaces
func diffInterfaces(oldIfs, newIfs []*nat.Nat44Global_NatInterfaces) (toSetIn, toSetOut, toUnsetIn, toUnsetOut []*nat.Nat44Global_NatInterfaces) {
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
func diffAddressPools(oldAPs, newAPs []*nat.Nat44Global_AddressPools) (toAdd, toRemove []*nat.Nat44Global_AddressPools) {
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
func (plugin *NatConfigurator) diffStatic(oldMappings, newMappings []*nat.Nat44DNat_DNatConfig_StaticMappings) (toAdd, toRemove []*nat.Nat44DNat_DNatConfig_StaticMappings) {
	// Find missing mappings
	for _, newMap := range newMappings {
		var found bool
		for _, oldMap := range oldMappings {
			// VRF, protocol and twice map
			if newMap.VrfId != oldMap.VrfId || newMap.Protocol != oldMap.Protocol || newMap.TwiceNat != oldMap.TwiceNat {
				continue
			}
			// External interface, IP and port
			if newMap.ExternalInterface != oldMap.ExternalInterface || newMap.ExternalIP != oldMap.ExternalIP ||
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
			if newMap.ExternalInterface != oldMap.ExternalInterface || newMap.ExternalIP != oldMap.ExternalIP ||
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
func (plugin *NatConfigurator) diffIdentity(oldMappings, newMappings []*nat.Nat44DNat_DNatConfig_IdentityMappings) (toAdd, toRemove []*nat.Nat44DNat_DNatConfig_IdentityMappings) {
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
func (plugin *NatConfigurator) compareLocalIPs(oldIPs, newIPs []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs) bool {
	if len(oldIPs) != len(newIPs) {
		return false
	}
	for _, newIP := range newIPs {
		var found bool
		for _, oldIP := range oldIPs {
			if newIP.LocalIP == oldIP.LocalIP && newIP.LocalPort == oldIP.LocalPort && newIP.Probability == oldIP.Probability {
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
func getLocalIPs(ipPorts []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs, log logging.Logger) (locals []*vppcalls.LocalLbAddress) {
	for _, ipPort := range ipPorts {
		if ipPort.LocalPort == 0 {
			log.Error("cannot set local IP/Port to mapping: port is missing")
			continue
		}

		localIP := net.ParseIP(ipPort.LocalIP).To4()
		if localIP == nil {
			log.Errorf("cannot set local IP/Port to mapping: unable to parse local IP %v", ipPort.LocalIP)
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

// returns unique ID of the mapping
func getStMappingIdentifier(mapping *nat.Nat44DNat_DNatConfig_StaticMappings) string {
	extIP := strings.Replace(mapping.ExternalIP, ".", "", -1)
	extIP = strings.Replace(extIP, "/", "", -1)
	locIP := strings.Replace(mapping.LocalIps[0].LocalIP, ".", "", -1)
	locIP = strings.Replace(locIP, "/", "", -1)
	return extIP + locIP + strconv.Itoa(int(mapping.VrfId))
}

// returns unique ID of the mapping
func getIdMappingIdentifier(mapping *nat.Nat44DNat_DNatConfig_IdentityMappings) string {
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
	buffer.WriteString(strconv.Itoa(int(plugin.NatMappingTagSeq)))
	plugin.NatMappingTagSeq++

	return buffer.String()
}
