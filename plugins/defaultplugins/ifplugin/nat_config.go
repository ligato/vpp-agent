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
	"fmt"
	"net"
	"strconv"
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	bin_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
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
	// DNAT indices for static mapping
	DNatMappingIndices idxvpp.NameToIdxRW
	NatIndexSeq        uint32
	vppChan            *govppapi.Channel

	// a map of missing interfaces which should be enabled for NAT (format ifName/isInside)
	notEnabledIfs map[string]bool
	// a map of NAT-enabled interfaces which should be disabled (format ifName/isInside)
	notDisabledIfs map[string]bool

	Stopwatch *measure.Stopwatch
}

// Init NAT configurator
func (plugin *NatConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing NAT configurator")

	// Init VPP API channel
	if plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return
	}

	// Check VPP message compatibility
	if err = plugin.checkMsgCompatibility(); err != nil {
		return
	}

	// Init local vars
	plugin.notEnabledIfs = make(map[string]bool)
	plugin.notDisabledIfs = make(map[string]bool)

	return
}

// Close used resources
func (plugin *NatConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// SetNatGlobalConfig configures common setup for all NAT use cases
func (plugin *NatConfigurator) SetNatGlobalConfig(config *nat.Nat44Global) (err error) {
	plugin.Log.Info("Setting up NAT global config")

	// Forwarding
	if err = vppcalls.SetNat44Forwarding(config.Forwarding, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(bin_api.Nat44ForwardingEnableDisable{}, plugin.Stopwatch)); err != nil {
		return
	}

	// Inside interfaces
	if len(config.In) > 0 {
		if err := plugin.enableNatInterfaces(config.In, true); err != nil {
			return err
		}
	} else {
		plugin.Log.Debug("No inside interfaces to configure")
	}

	// Outside interfaces
	if len(config.Out) > 0 {
		if err := plugin.enableNatInterfaces(config.Out, false); err != nil {
			return err
		}
	} else {
		plugin.Log.Debug("No outside interfaces to configure")
	}

	// Address pool
	for _, addressPool := range config.AddressPool {
		if addressPool.FirstSrcAddress == "" && addressPool.LastSrcAddres == "" {
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
		if addressPool.LastSrcAddres != "" {
			lastIP = net.ParseIP(addressPool.LastSrcAddres).To4()
			if lastIP == nil {
				plugin.Log.Errorf("unable to parse IP address %v", addressPool.LastSrcAddres)
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
		if err = vppcalls.AddNat44AddressPool(firstIP, lastIP, config.VrfId, addressPool.TwiceNat, plugin.Log, plugin.vppChan,
			measure.GetTimeLog(bin_api.Nat44AddDelAddressRange{}, plugin.Stopwatch)); err != nil {
			return
		}
	}

	plugin.Log.Debug("Setting up NAT global config done")

	return
}

// ModifyNatGlobalConfig modifies common setup for all NAT use cases
func (plugin *NatConfigurator) ModifyNatGlobalConfig(oldConfig, newConfig *nat.Nat44Global) (err error) {
	plugin.Log.Info("Modifying NAT global config")

	// Forwarding
	if oldConfig.Forwarding != newConfig.Forwarding {
		if err = vppcalls.SetNat44Forwarding(newConfig.Forwarding, plugin.Log, plugin.vppChan,
			measure.GetTimeLog(bin_api.Nat44ForwardingEnableDisable{}, plugin.Stopwatch)); err != nil {
			return
		}
	}

	// Inside/outside interfaces
	toSetIn, toUnsetIn := diffIfaces(oldConfig.In, newConfig.In)
	toSetOut, toUnsetOut := diffIfaces(oldConfig.Out, newConfig.Out)
	if err := plugin.disableNatInterfaces(append(toUnsetIn, toUnsetOut...), true); err != nil {
		return err
	}
	if err := plugin.enableNatInterfaces(append(toSetIn, toSetOut...), true); err != nil {
		return err
	}

	// Address pool
	// If Vrf ID is the same, address pool diff can be calculated and only changes are
	// added/removed
	if oldConfig.VrfId == newConfig.VrfId {
		toAdd, toRemove := diffAddressPools(oldConfig.AddressPool, newConfig.AddressPool)
		if err := plugin.addAddressPool(toAdd, newConfig.VrfId); err != nil {
			return err
		}
		if err := plugin.deleteAddressPool(toRemove, newConfig.VrfId); err != nil {
			return err
		}
	} else {
		// Otherwise all old address pools have to be removed and recreated
		if err := plugin.deleteAddressPool(oldConfig.AddressPool, oldConfig.VrfId); err != nil {
			return err
		}

		if err := plugin.addAddressPool(newConfig.AddressPool, newConfig.VrfId); err != nil {
			return err
		}
	}

	plugin.Log.Debug("Modifying NAT global config done")

	return nil
}

// DeleteNatGlobalConfig removes common setup for all NAT use cases
func (plugin *NatConfigurator) DeleteNatGlobalConfig(config *nat.Nat44Global) (err error) {
	plugin.Log.Info("Deleting NAT global config")

	// Inside interfaces
	if len(config.In) > 0 {
		if err := plugin.disableNatInterfaces(config.In, true); err != nil {
			return err
		}
	}
	// Outside interfaces
	if len(config.Out) > 0 {
		if err := plugin.disableNatInterfaces(config.Out, false); err != nil {
			return err
		}
	}

	// Address pools
	if len(config.AddressPool) > 0 {
		if err := plugin.deleteAddressPool(config.AddressPool, config.VrfId); err != nil {
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

	mapping := dNat.Mapping
	for mappingIdx, mappingEntry := range mapping {
		localIPList := mappingEntry.LocalIp
		if len(localIPList) == 0 {
			plugin.Log.Errorf("cannot configure DNAT %v mapping: no local address provided", mappingIdx)
			continue
		} else if len(localIPList) == 1 {
			// Case without load balance (one local address)
			if err := plugin.handleStaticMapping(mappingEntry.ExternalIP, mappingEntry.ExternalInterface, dNat.VrfId,
				dNat.SNatEnabled, mappingEntry, true); err != nil {
				plugin.Log.Errorf("DNAT mapping configuration failed: %v", err)
				continue
			}
		} else {
			// Case with load balance (more local addresses)
			if err := plugin.handleStaticMappingLb(dNat.VrfId, dNat.SNatEnabled, mappingEntry, true); err != nil {
				plugin.Log.Errorf("DNAT lb-mapping configuration failed: %v", err)
				continue
			}
		}
		// Register DNAT mapping
		mappingIdentifier := getMappingIdentifier(mappingEntry, dNat.VrfId)
		plugin.DNatMappingIndices.RegisterName(mappingIdentifier, plugin.NatIndexSeq, nil)
		plugin.NatIndexSeq++

		plugin.Log.Debugf("DNAT lb-mapping configured (ID %v)", mappingIdentifier)

	}

	plugin.Log.Infof("DNAT %v configuration done", dNat.Label)

	return nil
}

// ModifyDNat modifies existing DNAT setup
func (plugin *NatConfigurator) ModifyDNat(oldDNat, newDNat *nat.Nat44DNat_DNatConfig) error {
	plugin.Log.Infof("Modifying DNAT with label %v", newDNat.Label)

	// todo keep is simple for now, but it would be better to find different mappings and add/remove just them
	if err := plugin.DeleteDNat(oldDNat); err != nil {
		return err
	}
	if err := plugin.ConfigureDNat(newDNat); err != nil {
		return err
	}

	plugin.Log.Infof("DNAT %v modification done", newDNat.Label)

	return nil
}

// DeleteDNat removes existing DNAT setup
func (plugin *NatConfigurator) DeleteDNat(dNat *nat.Nat44DNat_DNatConfig) error {
	plugin.Log.Infof("Deleting DNAT with label %v", dNat.Label)

	// In delete case, vpp-agent attempts to reconstruct every mapping entry and remove it from the VPP
	mapping := dNat.Mapping
	for mappingIdx, mappingEntry := range mapping {
		localIPList := mappingEntry.LocalIp
		if len(localIPList) == 0 {
			plugin.Log.Warnf("DNAT mapping %v has not local IPs, cannot remove it", mappingIdx)
			continue
		} else if len(localIPList) == 1 {
			// Case without load balance (one local address)
			if err := plugin.handleStaticMapping(mappingEntry.ExternalIP, mappingEntry.ExternalInterface, dNat.VrfId,
				dNat.SNatEnabled, mappingEntry, false); err != nil {
				plugin.Log.Errorf("DNAT mapping removal failed: %v", err)
				continue
			}
		} else {
			// Case with load balance (more local addresses)
			if err := plugin.handleStaticMappingLb(dNat.VrfId, dNat.SNatEnabled, mappingEntry, false); err != nil {
				plugin.Log.Errorf("DNAT lb-mapping removal failed: %v", err)
				continue
			}
		}
		// Unregister DNAT mapping
		mappingIdentifier := getMappingIdentifier(mappingEntry, dNat.VrfId)
		plugin.DNatMappingIndices.UnregisterName(mappingIdentifier)

		plugin.Log.Debugf("DNAT lb-mapping un-configured (ID %v)", mappingIdentifier)
	}

	plugin.Log.Infof("DNAT %v removal done", dNat.Label)

	return nil
}

// enables set of interfaces as inside/outside in NAT
func (plugin *NatConfigurator) enableNatInterfaces(interfaces []string, isInside bool) (err error) {
	for _, iface := range interfaces {
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(iface)
		if !found {
			plugin.Log.Debugf("Interface %v missing, cannot enable NAT", iface)
			plugin.notEnabledIfs[iface] = isInside // cache as inside/outside interface
		} else {
			if err = vppcalls.EnableNat44Interface(iface, ifIdx, isInside, plugin.Log, plugin.vppChan,
				measure.GetTimeLog(bin_api.Nat44InterfaceAddDelFeature{}, plugin.Stopwatch)); err != nil {
				return
			}
			if isInside {
				plugin.Log.Debugf("Interface %v enabled for NAT as inside", iface)
			} else {
				plugin.Log.Debugf("Interface %v enabled for NAT as outside", iface)
			}
		}
	}

	return
}

// disables set of interfaces in NAT
func (plugin *NatConfigurator) disableNatInterfaces(interfaces []string, isInside bool) (err error) {
	for _, iface := range interfaces {
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(iface)
		if !found {
			plugin.Log.Debugf("Interface %v missing, cannot disable NAT", iface)
			plugin.notDisabledIfs[iface] = isInside // cache as inside/outside interface
		} else {
			if err = vppcalls.DisableNat44Interface(iface, ifIdx, isInside, plugin.Log, plugin.vppChan,
				measure.GetTimeLog(bin_api.Nat44InterfaceAddDelFeature{}, plugin.Stopwatch)); err != nil {
				return
			}
			if isInside {
				plugin.Log.Debugf("Interface %v disabled for NAT as inside", iface)
			} else {
				plugin.Log.Debugf("Interface %v disabled for NAT as outside", iface)
			}
		}
	}

	return
}

// adds NAT address pool
func (plugin *NatConfigurator) addAddressPool(addressPools []*nat.Nat44Global_AddressPool, vrfID uint32) (err error) {
	for _, addressPool := range addressPools {
		if addressPool.FirstSrcAddress == "" && addressPool.LastSrcAddres == "" {
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
		if addressPool.LastSrcAddres != "" {
			lastIP = net.ParseIP(addressPool.LastSrcAddres).To4()
			if lastIP == nil {
				plugin.Log.Errorf("unable to parse IP address %v", addressPool.LastSrcAddres)
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
		if err = vppcalls.AddNat44AddressPool(firstIP, lastIP, vrfID, addressPool.TwiceNat, plugin.Log, plugin.vppChan,
			measure.GetTimeLog(bin_api.Nat44AddDelAddressRange{}, plugin.Stopwatch)); err != nil {
			return
		}
	}

	return
}

// removes NAT address pool
func (plugin *NatConfigurator) deleteAddressPool(addressPools []*nat.Nat44Global_AddressPool, vrfID uint32) (err error) {
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
		if addressPool.LastSrcAddres != "" {
			lastIP = net.ParseIP(addressPool.LastSrcAddres).To4()
			if lastIP == nil {
				plugin.Log.Errorf("unable to parse IP address %v", addressPool.LastSrcAddres)
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
		if err = vppcalls.DelNat44AddressPool(firstIP, lastIP, vrfID, addressPool.TwiceNat, plugin.Log, plugin.vppChan,
			measure.GetTimeLog(bin_api.Nat44AddDelAddressRange{}, plugin.Stopwatch)); err != nil {
			return
		}
	}

	return
}

// configures single static mapping entry with load balancer
func (plugin *NatConfigurator) handleStaticMappingLb(vrf uint32, sNatEnabled bool, entry *nat.Nat44DNat_DNatConfig_Mapping,
	add bool) (err error) {

	// Resolve external IP address and port
	exIPAddr, exPort, err := splitIPPortVal(entry.ExternalIP)
	if err != nil {
		return fmt.Errorf("cannot configure DNAT mapping: %v", err)
	}
	// Parse external IP address
	exIPAddrByte := net.ParseIP(exIPAddr).To4()
	if exIPAddrByte == nil {
		return fmt.Errorf("cannot configure DNAT mapping: unable to parse external IP %v", exIPAddr)
	}

	// In this case, external port is required
	if exPort == 0 {
		return fmt.Errorf("cannot configure DNAT mapping: external port is not set")
	}

	// Address mapping with load balancer
	ctx := &vppcalls.StaticMappingLbContext{
		ExternalIP:   exIPAddrByte,
		ExternalPort: uint16(exPort),
		Protocol:     getProtocol(entry.Protocol, plugin.Log),
		LocalIPs:     getLocalIPs(entry.LocalIp, plugin.Log),
	}

	if len(ctx.LocalIPs) == 0 {
		return fmt.Errorf("cannot configure DNAT mapping: no local IP was successfully parsed")
	}

	if add {
		return plugin.configureStaticEntryLb(ctx, vrf, sNatEnabled)
	}
	return plugin.removeStaticEntryLb(ctx, vrf, sNatEnabled)
}

// Configure static mapping with load balancer
func (plugin *NatConfigurator) configureStaticEntryLb(ctx *vppcalls.StaticMappingLbContext, vrf uint32, sNatEnabled bool) (err error) {
	if err = vppcalls.AddNat44StaticMappingLb(ctx, vrf, sNatEnabled, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(bin_api.Nat44AddDelLbStaticMapping{}, plugin.Stopwatch)); err != nil {
		return fmt.Errorf("DNAT lb-mapping configuration failed: %v", err)
	}
	return
}

// Remove static mapping with load balancer
func (plugin *NatConfigurator) removeStaticEntryLb(ctx *vppcalls.StaticMappingLbContext, vrf uint32, sNatEnabled bool) (err error) {
	if err = vppcalls.DelNat44StaticMappingLb(ctx, vrf, sNatEnabled, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(bin_api.Nat44AddDelLbStaticMapping{}, plugin.Stopwatch)); err != nil {
		return fmt.Errorf("DNAT lb-mapping removal failed: %v", err)
	}
	return
}

// handler for single static mapping entry
func (plugin *NatConfigurator) handleStaticMapping(externalIP string, exIfName string, vrf uint32,
	sNatEnabled bool, entry *nat.Nat44DNat_DNatConfig_Mapping, add bool) (err error) {
	var ifIdx uint32 = 0xffffffff // default value - means no external interface is set
	var exIPAddr string
	var exIPAddrByte []byte
	var exPort int

	// Resolve local IP address and port
	var lcIPAddr string
	lcIPAddr, lcPort, err := splitIPPortVal(entry.LocalIp[0].LocalIP)
	if err != nil {
		return fmt.Errorf("cannot configure DNAT mapping: %v", err)
	}
	// Parse local IP address
	lcIPAddrByte := net.ParseIP(lcIPAddr).To4()
	if lcIPAddrByte == nil {
		return fmt.Errorf("cannot configure DNAT mapping: unable to parse local IP %v", lcIPAddr)
	}

	// Check external interface (prioritized over external IP)
	if exIfName != "" {
		// Check external interface
		var found bool
		ifIdx, _, found = plugin.SwIfIndexes.LookupIdx(exIfName)
		if !found {
			return fmt.Errorf("required external interface %v is missing", exIfName)
		}
	} else {
		// If no external interface, check external IP address/port
		exIPAddr, exPort, err = splitIPPortVal(externalIP)
		if err != nil {
			return fmt.Errorf("cannot configure DNAT mapping: %v", err)
		}
		// Parse external IP address
		exIPAddrByte = net.ParseIP(exIPAddr).To4()
		if lcIPAddrByte == nil {
			return fmt.Errorf("cannot configure DNAT mapping: unable to parse external IP %v", lcIPAddr)
		}
	}

	// Resolve mapping (address only or address and port)
	var addrOnly bool
	if lcPort == 0 || exPort == 0 {
		addrOnly = true
	}

	// Address mapping with load balancer
	ctx := &vppcalls.StaticMappingContext{
		AddressOnly:   addrOnly,
		LocalIP:       lcIPAddrByte,
		LocalPort:     uint16(lcPort),
		ExternalIP:    exIPAddrByte,
		ExternalPort:  uint16(exPort),
		ExternalIfIdx: ifIdx,
		Protocol:      getProtocol(entry.Protocol, plugin.Log),
	}

	if add {
		return plugin.configureStaticEntry(ctx, vrf, sNatEnabled)
	}
	return plugin.removeStaticEntry(ctx, vrf, sNatEnabled)
}

// Configure static mapping
func (plugin *NatConfigurator) configureStaticEntry(ctx *vppcalls.StaticMappingContext, vrf uint32, sNatEnabled bool) (err error) {
	if err = vppcalls.AddNat44StaticMapping(ctx, vrf, sNatEnabled, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(bin_api.Nat44AddDelStaticMapping{}, plugin.Stopwatch)); err != nil {
		return fmt.Errorf("DNAT mapping configuration failed: %v", err)
	}
	return
}

// Remove static mapping
func (plugin *NatConfigurator) removeStaticEntry(ctx *vppcalls.StaticMappingContext, vrf uint32, sNatEnabled bool) (err error) {
	if err = vppcalls.DelNat44StaticMapping(ctx, vrf, sNatEnabled, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(bin_api.Nat44AddDelStaticMapping{}, plugin.Stopwatch)); err != nil {
		return fmt.Errorf("DNAT mapping removal failed: %v", err)
	}
	return
}

// looks for new and obsolete interfaces
func diffIfaces(oldIfs, newIfs []string) (toSet, toUnset []string) {
	// Find new interfaces
	for _, newIf := range newIfs {
		var found bool
		for _, oldIf := range oldIfs {
			if newIf == oldIf {
				found = true
			}
		}
		if !found {
			toSet = append(toSet, newIf)
		}
	}
	// Find obsolete interfaces
	for _, oldIf := range oldIfs {
		var found bool
		for _, newIf := range newIfs {
			if oldIf == newIf {
				found = true
			}
		}
		if !found {
			toUnset = append(toUnset, oldIf)
		}
	}

	return
}

// looks for new and obsolete address pools
func diffAddressPools(oldAPs, newAPs []*nat.Nat44Global_AddressPool) (toAdd, toRemove []*nat.Nat44Global_AddressPool) {
	// Find new address pools
	for _, newAp := range newAPs {
		var found bool
		for _, oldAp := range oldAPs {
			if newAp.FirstSrcAddress == oldAp.FirstSrcAddress &&
				newAp.LastSrcAddres == oldAp.LastSrcAddres && newAp.TwiceNat == oldAp.TwiceNat {
				found = true
			}
		}
		if !found {
			toAdd = append(toAdd, newAp)
		}
	}
	// Find obsolete address pools
	for _, oldAp := range oldAPs {
		var found bool
		for _, newAp := range newAPs {
			if newAp.FirstSrcAddress == oldAp.FirstSrcAddress &&
				newAp.LastSrcAddres == oldAp.LastSrcAddres && newAp.TwiceNat == oldAp.TwiceNat {
				found = true
			}
		}
		if !found {
			toRemove = append(toRemove, oldAp)
		}
	}

	return
}

// checkMsgCompatibility verifies compatibility of used binary API calls
func (plugin *NatConfigurator) checkMsgCompatibility() error {
	msgs := []govppapi.Message{
		&bin_api.Nat44AddDelAddressRange{},
		&bin_api.Nat44AddDelAddressRangeReply{},
		&bin_api.Nat44ForwardingEnableDisable{},
		&bin_api.Nat44ForwardingEnableDisableReply{},
		&bin_api.Nat44InterfaceAddDelFeature{},
		&bin_api.Nat44InterfaceAddDelFeatureReply{},
		&bin_api.Nat44AddDelStaticMapping{},
		&bin_api.Nat44AddDelStaticMappingReply{},
		&bin_api.Nat44AddDelLbStaticMapping{},
		&bin_api.Nat44AddDelLbStaticMappingReply{},
	}
	err := plugin.vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		plugin.Log.Error(err)
	}
	return err
}

// returns ip and port if provided value is correct
func splitIPPortVal(val string) (ip string, port int, err error) {
	ipPort := strings.Split(val, "/")
	if len(ipPort) == 0 {
		err = fmt.Errorf("unable to get ip/port value, incorrect format %v", val)
	} else if len(ipPort) == 1 {
		ip = ipPort[0]
	} else if len(ipPort) == 2 {
		ip = ipPort[0]
		port, err = strconv.Atoi(ipPort[1])
	} else {
		err = fmt.Errorf("unable to get ip/port value, incorrect format %v", val)
	}
	return
}

// returns a list of validated local IP addresses with port and probability value
func getLocalIPs(ipPorts []*nat.Nat44DNat_DNatConfig_Mapping_LocalIP, log logging.Logger) (locals []*vppcalls.LocalLbAddress) {
	for _, ipPort := range ipPorts {
		localIP, localPort, err := splitIPPortVal(ipPort.LocalIP)
		if err != nil {
			log.Errorf("cannot set local IP/Port to mapping: %v", err)
			continue
		}
		if localPort == 0 {
			log.Error("cannot set local IP/Port to mapping: port is missing")
			continue
		}

		localIPByte := net.ParseIP(localIP).To4()
		if localIPByte == nil {
			log.Error("cannot set local IP/Port to mapping: unable to parse local IP %v", localIP)
			continue
		}

		locals = append(locals, &vppcalls.LocalLbAddress{
			LocalIP:     localIPByte,
			LocalPort:   uint16(localPort),
			Probability: uint8(ipPort.Probability),
		})
	}

	return
}

// returns num representation of provided protocol value
func getProtocol(protocol nat.Nat44DNat_DNatConfig_Mapping_Protocol, log logging.Logger) (proto uint8) {
	if protocol == nat.Nat44DNat_DNatConfig_Mapping_TCP {
		proto = vppcalls.TCP
	} else if protocol == nat.Nat44DNat_DNatConfig_Mapping_UDP {
		proto = vppcalls.UDP
	} else if protocol == nat.Nat44DNat_DNatConfig_Mapping_ICMP {
		proto = vppcalls.ICMP
	} else {
		log.Warnf("Unknown protocol %v, defaulting to TCP", protocol)
		proto = vppcalls.TCP
	}
	return
}

// returns unique ID of the mapping
func getMappingIdentifier(mapping *nat.Nat44DNat_DNatConfig_Mapping, vrf uint32) string {
	extIP := strings.Replace(mapping.ExternalIP, ".", "", -1)
	extIP = strings.Replace(extIP, "/", "", -1)
	locIP := strings.Replace(mapping.LocalIp[0].LocalIP, ".", "", -1)
	locIP = strings.Replace(locIP, "/", "", -1)
	return extIP + locIP + strconv.Itoa(int(vrf))
}
