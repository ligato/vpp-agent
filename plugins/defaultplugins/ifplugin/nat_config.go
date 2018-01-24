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
	"net"

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
//   load ballancer
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type NatConfigurator struct {
	Log logging.Logger

	GoVppmux    govppmux.API
	SwIfIndexes ifaceidx.SwIfIndex
	NatIndices  idxvpp.NameToIdxRW
	NatIndexSeq uint32
	vppChan     *govppapi.Channel

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
	plugin.Log.Info("Deleteing NAT global config")

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
func (plugin *NatConfigurator) ConfigureSNat(sNat *nat.Nat44SNat) error {
	plugin.Log.Warn("SNAT CREATE not implemented")
	return nil
}

// ModifySNat modifies existing SNAT setup
func (plugin *NatConfigurator) ModifySNat(oldSNat, newSNat *nat.Nat44SNat) error {
	plugin.Log.Warn("SNAT MODIFY not implemented")
	return nil
}

// DeleteSNat removes existing SNAT setup
func (plugin *NatConfigurator) DeleteSNat(sNat *nat.Nat44SNat) error {
	plugin.Log.Warn("SNAT DELETE not implemented")
	return nil
}

// ConfigureDNat configures new DNAT setup
func (plugin *NatConfigurator) ConfigureDNat(sNat *nat.Nat44DNat) error {
	plugin.Log.Warn("DNAT CREATE not implemented")
	return nil
}

// ModifyDNat modifies existing DNAT setup
func (plugin *NatConfigurator) ModifyDNat(oldSNat, newSNat *nat.Nat44DNat) error {
	plugin.Log.Warn("DNAT MODIFY not implemented")
	return nil
}

// DeleteDNat removes existing DNAT setup
func (plugin *NatConfigurator) DeleteDNat(sNat *nat.Nat44DNat) error {
	plugin.Log.Warn("DNAT DELETE not implemented")
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
	}
	err := plugin.vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		plugin.Log.Error(err)
	}
	return err
}
