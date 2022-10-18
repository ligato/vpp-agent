//  Copyright (c) 2022 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp2202

import (
	"fmt"
	"net"

	"go.fd.io/govpp/api"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ip_types"
	vpp_nat_ed "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/nat44_ed"
	vpp_nat_ei "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/nat44_ei"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/nat_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

// Num protocol representation
const (
	ICMP uint8 = 1
	TCP  uint8 = 6
	UDP  uint8 = 17
)

const (
	// NoInterface is sw-if-idx which means 'no interface'
	NoInterface = interface_types.InterfaceIndex(^uint32(0))
	// Maximal length of tag
	maxTagLen = 64
)

// holds a list of NAT44 ED flags set
type nat44EdFlags struct {
	isTwiceNat     bool
	isSelfTwiceNat bool
	isOut2In       bool
	isAddrOnly     bool
	isOutside      bool
	isInside       bool
	isStatic       bool
	isExtHostValid bool
}

// holds a list of NAT44 EI flags set
type nat44EiFlags struct {
	eiStaticMappingOnly  bool
	eiConnectionTracking bool
	eiOut2InDpo          bool
	eiAddrOnlyMapping    bool
	eiIfInside           bool
	eiIfOutside          bool
	eiStaticMapping      bool
}

func (h *NatVppHandler) enableNAT44EdPlugin(opts vppcalls.Nat44InitOpts) error {
	var flags vpp_nat_ed.Nat44ConfigFlags
	if opts.ConnectionTracking {
		flags |= vpp_nat_ed.NAT44_IS_CONNECTION_TRACKING
	}
	if opts.StaticMappingOnly {
		flags |= vpp_nat_ed.NAT44_IS_STATIC_MAPPING_ONLY
	}

	req := &vpp_nat_ed.Nat44EdPluginEnableDisable{
		Enable: true,
		Flags:  flags,
	}
	reply := &vpp_nat_ed.Nat44EdPluginEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return nil
}

func (h *NatVppHandler) enableNAT44EiPlugin(opts vppcalls.Nat44InitOpts) error {
	var flags vpp_nat_ei.Nat44EiConfigFlags
	if opts.ConnectionTracking {
		flags |= vpp_nat_ei.NAT44_EI_CONNECTION_TRACKING
	}
	if opts.StaticMappingOnly {
		flags |= vpp_nat_ei.NAT44_EI_STATIC_MAPPING_ONLY
	}
	if opts.OutToInDPO {
		flags |= vpp_nat_ei.NAT44_EI_OUT2IN_DPO
	}

	req := &vpp_nat_ei.Nat44EiPluginEnableDisable{
		Enable: true,
		Flags:  flags,
	}
	reply := &vpp_nat_ei.Nat44EiPluginEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return nil
}

// EnableNAT44plugin and apply the given set of options.
func (h *NatVppHandler) EnableNAT44Plugin(opts vppcalls.Nat44InitOpts) error {
	// VPP since version 21.06 supports running both NAT EI and NAT ED simultaneously,
	// but not vpp-agent yet.
	// TODO: separate vpp-agent natplugin into 2 separate plugins
	// (for ED and EI NAT) or create a separate handlers inside one plugin,
	// this have number of advantages and will probably also become necessary as VPP
	// NAT plugins will differ more and more over time.
	if opts.EndpointDependent {
		h.ed = true
		return h.enableNAT44EdPlugin(opts)
	} else {
		h.ed = false
		return h.enableNAT44EiPlugin(opts)
	}
}

func (h *NatVppHandler) disableNAT44EdPlugin() error {
	req := &vpp_nat_ed.Nat44EdPluginEnableDisable{
		Enable: false,
	}
	reply := &vpp_nat_ed.Nat44EdPluginEnableDisableReply{}
	err := h.callsChannel.SendRequest(req).ReceiveReply(reply)
	if err == api.VPPApiError(1) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (h *NatVppHandler) disableNAT44EiPlugin() error {
	req := &vpp_nat_ei.Nat44EiPluginEnableDisable{
		Enable: false,
	}
	reply := &vpp_nat_ei.Nat44EiPluginEnableDisableReply{}
	err := h.callsChannel.SendRequest(req).ReceiveReply(reply)
	if err == api.VPPApiError(1) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

// DisableNAT44Plugin disables NAT44 plugin.
func (h *NatVppHandler) DisableNAT44Plugin() error {
	if h.ed {
		return h.disableNAT44EdPlugin()
	} else {
		return h.disableNAT44EiPlugin()
	}
}

func (h *NatVppHandler) setNat44EdForwarding(enableFwd bool) error {
	req := &vpp_nat_ed.Nat44ForwardingEnableDisable{
		Enable: enableFwd,
	}
	reply := &vpp_nat_ed.Nat44ForwardingEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *NatVppHandler) setNat44EiForwarding(enableFwd bool) error {
	req := &vpp_nat_ei.Nat44EiForwardingEnableDisable{
		Enable: enableFwd,
	}
	reply := &vpp_nat_ei.Nat44EiForwardingEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// SetNat44Forwarding configures NAT44 forwarding.
func (h *NatVppHandler) SetNat44Forwarding(enableFwd bool) error {
	if h.ed {
		return h.setNat44EdForwarding(enableFwd)
	} else {
		return h.setNat44EiForwarding(enableFwd)
	}
}

// EnableNat44Interface enables NAT44 feature for provided interface.
func (h *NatVppHandler) EnableNat44Interface(iface string, isInside, isOutput bool) error {
	if isOutput {
		return h.handleNat44InterfaceOutputFeature(iface, isInside, true)
	}
	return h.handleNat44Interface(iface, isInside, true)
}

// DisableNat44Interface disables NAT44 feature for provided interface.
func (h *NatVppHandler) DisableNat44Interface(iface string, isInside, isOutput bool) error {
	if isOutput {
		return h.handleNat44InterfaceOutputFeature(iface, isInside, false)
	}
	return h.handleNat44Interface(iface, isInside, false)
}

// AddNat44AddressPool adds new IPV4 address pool into the NAT pools.
func (h *NatVppHandler) AddNat44AddressPool(vrf uint32, firstIP, lastIP string, twiceNat bool) error {
	return h.handleNat44AddressPool(vrf, firstIP, lastIP, twiceNat, true)
}

// DelNat44AddressPool removes existing IPv4 address pool from the NAT pools.
func (h *NatVppHandler) DelNat44AddressPool(vrf uint32, firstIP, lastIP string, twiceNat bool) error {
	return h.handleNat44AddressPool(vrf, firstIP, lastIP, twiceNat, false)
}

// SetVirtualReassemblyIPv4 configures NAT virtual reassembly for IPv4 packets.
func (h *NatVppHandler) SetVirtualReassemblyIPv4(vrCfg *nat.VirtualReassembly) error {
	return h.handleNatVirtualReassembly(vrCfg, false)
}

// SetVirtualReassemblyIPv6 configures NAT virtual reassembly for IPv6 packets.
func (h *NatVppHandler) SetVirtualReassemblyIPv6(vrCfg *nat.VirtualReassembly) error {
	return h.handleNatVirtualReassembly(vrCfg, true)
}

// AddNat44IdentityMapping adds new NAT44 identity mapping
func (h *NatVppHandler) AddNat44IdentityMapping(mapping *nat.DNat44_IdentityMapping, dnatLabel string) error {
	return h.handleNat44IdentityMapping(mapping, dnatLabel, true)
}

// DelNat44IdentityMapping removes existing NAT44 identity mapping
func (h *NatVppHandler) DelNat44IdentityMapping(mapping *nat.DNat44_IdentityMapping, dnatLabel string) error {
	return h.handleNat44IdentityMapping(mapping, dnatLabel, false)
}

// AddNat44StaticMapping creates new NAT44 static mapping entry.
func (h *NatVppHandler) AddNat44StaticMapping(mapping *nat.DNat44_StaticMapping, dnatLabel string) error {
	if len(mapping.LocalIps) == 0 {
		return errors.Errorf("cannot configure static mapping for DNAT %s: no local address provided", dnatLabel)
	}
	if len(mapping.LocalIps) == 1 {
		return h.handleNat44StaticMapping(mapping, dnatLabel, true)
	}
	return h.handleNat44StaticMappingLb(mapping, dnatLabel, true)
}

// DelNat44StaticMapping removes existing NAT44 static mapping entry.
func (h *NatVppHandler) DelNat44StaticMapping(mapping *nat.DNat44_StaticMapping, dnatLabel string) error {
	if len(mapping.LocalIps) == 0 {
		return errors.Errorf("cannot un-configure static mapping from DNAT %s: no local address provided", dnatLabel)
	}
	if len(mapping.LocalIps) == 1 {
		return h.handleNat44StaticMapping(mapping, dnatLabel, false)
	}
	return h.handleNat44StaticMappingLb(mapping, dnatLabel, false)
}

func (h *NatVppHandler) handleNatEd44Interface(iface string, isInside, isAdd bool) error {
	// get interface metadata
	ifaceMeta, found := h.ifIndexes.LookupByName(iface)
	if !found {
		return errors.New("failed to get interface metadata")
	}

	req := &vpp_nat_ed.Nat44InterfaceAddDelFeature{
		SwIfIndex: interface_types.InterfaceIndex(ifaceMeta.SwIfIndex),
		Flags:     setNat44EdFlags(&nat44EdFlags{isInside: isInside}),
		IsAdd:     isAdd,
	}
	reply := &vpp_nat_ed.Nat44InterfaceAddDelFeatureReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *NatVppHandler) handleNat44EiInterface(iface string, isInside, isAdd bool) error {
	// get interface metadata
	ifaceMeta, found := h.ifIndexes.LookupByName(iface)
	if !found {
		return errors.New("failed to get interface metadata")
	}

	req := &vpp_nat_ei.Nat44EiInterfaceAddDelFeature{
		SwIfIndex: interface_types.InterfaceIndex(ifaceMeta.SwIfIndex),
		Flags:     setNat44EiFlags(&nat44EiFlags{eiIfInside: isInside}),
		IsAdd:     isAdd,
	}
	reply := &vpp_nat_ei.Nat44EiInterfaceAddDelFeatureReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to set/unset interface NAT44 feature.
func (h *NatVppHandler) handleNat44Interface(iface string, isInside, isAdd bool) error {
	if h.ed {
		return h.handleNatEd44Interface(iface, isInside, isAdd)
	} else {
		return h.handleNat44EiInterface(iface, isInside, isAdd)
	}
}

func (h *NatVppHandler) handleNat44EdInterfaceOutputFeature(iface string, isInside, isAdd bool) error {
	// get interface metadata
	ifaceMeta, found := h.ifIndexes.LookupByName(iface)
	if !found {
		return errors.New("failed to get interface metadata")
	}

	req := &vpp_nat_ed.Nat44InterfaceAddDelOutputFeature{
		SwIfIndex: interface_types.InterfaceIndex(ifaceMeta.SwIfIndex),
		Flags:     setNat44EdFlags(&nat44EdFlags{isInside: isInside}),
		IsAdd:     isAdd,
	}
	reply := &vpp_nat_ed.Nat44InterfaceAddDelOutputFeatureReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *NatVppHandler) handleNat44EiInterfaceOutputFeature(iface string, isInside, isAdd bool) error {
	// get interface metadata
	ifaceMeta, found := h.ifIndexes.LookupByName(iface)
	if !found {
		return errors.New("failed to get interface metadata")
	}

	req := &vpp_nat_ei.Nat44EiInterfaceAddDelFeature{
		SwIfIndex: interface_types.InterfaceIndex(ifaceMeta.SwIfIndex),
		Flags:     setNat44EiFlags(&nat44EiFlags{eiIfInside: isInside}),
		IsAdd:     isAdd,
	}
	reply := &vpp_nat_ei.Nat44EiInterfaceAddDelFeatureReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to set/unset interface NAT44 output feature
func (h *NatVppHandler) handleNat44InterfaceOutputFeature(iface string, isInside, isAdd bool) error {
	if h.ed {
		return h.handleNat44EdInterfaceOutputFeature(iface, isInside, isAdd)
	} else {
		return h.handleNat44EiInterfaceOutputFeature(iface, isInside, isAdd)
	}
}

func (h *NatVppHandler) handleNat44EdAddressPool(vrf uint32, firstIP, lastIP string, twiceNat, isAdd bool) error {
	firstAddr, err := ipTo4Address(firstIP)
	if err != nil {
		return errors.Errorf("unable to parse address %s from the NAT pool: %v", firstIP, err)
	}
	lastAddr := firstAddr
	if lastIP != "" {
		lastAddr, err = ipTo4Address(lastIP)
		if err != nil {
			return errors.Errorf("unable to parse address %s from the NAT pool: %v", lastIP, err)
		}
	}

	req := &vpp_nat_ed.Nat44AddDelAddressRange{
		FirstIPAddress: firstAddr,
		LastIPAddress:  lastAddr,
		VrfID:          vrf,
		Flags:          setNat44EdFlags(&nat44EdFlags{isTwiceNat: twiceNat}),
		IsAdd:          isAdd,
	}
	reply := &vpp_nat_ed.Nat44AddDelAddressRangeReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *NatVppHandler) handleNat44EiAddressPool(vrf uint32, firstIP, lastIP string, twiceNat, isAdd bool) error {
	firstAddr, err := ipTo4Address(firstIP)
	if err != nil {
		return errors.Errorf("unable to parse address %s from the NAT pool: %v", firstIP, err)
	}
	lastAddr := firstAddr
	if lastIP != "" {
		lastAddr, err = ipTo4Address(lastIP)
		if err != nil {
			return errors.Errorf("unable to parse address %s from the NAT pool: %v", lastIP, err)
		}
	}

	req := &vpp_nat_ei.Nat44EiAddDelAddressRange{
		FirstIPAddress: firstAddr,
		LastIPAddress:  lastAddr,
		VrfID:          vrf,
		IsAdd:          isAdd,
	}
	reply := &vpp_nat_ei.Nat44EiAddDelAddressRangeReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to add/remove addresses to/from the NAT44 pool.
func (h *NatVppHandler) handleNat44AddressPool(vrf uint32, firstIP, lastIP string, twiceNat, isAdd bool) error {
	if h.ed {
		return h.handleNat44EdAddressPool(vrf, firstIP, lastIP, twiceNat, isAdd)
	} else {
		return h.handleNat44EiAddressPool(vrf, firstIP, lastIP, twiceNat, isAdd)
	}
}

// Calls VPP binary API to setup NAT virtual reassembly
func (h *NatVppHandler) handleNatVirtualReassembly(vrCfg *nat.VirtualReassembly, isIpv6 bool) error {
	// Virtual Reassembly has been removed from NAT API in VPP (moved to IP API)
	// TODO: define IPReassembly model in L3 plugin
	return nil
	/*req := &vpp_nat.NatSetReass{
	  	Timeout:  vrCfg.Timeout,
	  	MaxReass: uint16(vrCfg.MaxReassemblies),
	  	MaxFrag:  uint8(vrCfg.MaxFragments),
	  	DropFrag: boolToUint(vrCfg.DropFragments),
	  	IsIP6:    isIpv6,
	  }
	  reply := &vpp_nat.NatSetReassReply{}
	  if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
	  	return err
	  }*/
}

// Calls VPP binary API to add/remove NAT44 static mapping
func (h *NatVppHandler) handleNat44StaticMapping(mapping *nat.DNat44_StaticMapping, dnatLabel string, isAdd bool) error {
	var ifIdx interface_types.InterfaceIndex // NOTE: This is a workaround, because NoInterface crashes VPP 22.02
	var exIPAddr ip_types.IP4Address

	// check tag length limit
	if err := checkTagLength(dnatLabel); err != nil {
		return err
	}

	// parse local endpoint
	lcIPAddr, err := ipTo4Address(mapping.LocalIps[0].LocalIp)
	if err != nil {
		return errors.Errorf("cannot configure DNAT static mapping %s: unable to parse local IP %s: %v",
			dnatLabel, mapping.LocalIps[0].LocalIp, err)
	}
	lcPort := uint16(mapping.LocalIps[0].LocalPort)
	lcVrf := mapping.LocalIps[0].VrfId

	// Check external interface (prioritized over external IP)
	if mapping.ExternalInterface != "" {
		// Check external interface
		ifMeta, found := h.ifIndexes.LookupByName(mapping.ExternalInterface)
		if !found {
			return errors.Errorf("cannot configure static mapping for DNAT %s: required external interface %s is missing",
				dnatLabel, mapping.ExternalInterface)
		}
		ifIdx = interface_types.InterfaceIndex(ifMeta.SwIfIndex)
	} else {
		// Parse external IP address
		exIPAddr, err = ipTo4Address(mapping.ExternalIp)
		if err != nil {
			return errors.Errorf("cannot configure static mapping for DNAT %s: unable to parse external IP %s: %v",
				dnatLabel, mapping.ExternalIp, err)
		}
	}

	// Resolve mapping (address only or address and port)
	var addrOnly bool
	if lcPort == 0 || mapping.ExternalPort == 0 {
		addrOnly = true
	}

	if h.ed {
		req := &vpp_nat_ed.Nat44AddDelStaticMappingV2{
			Tag:               dnatLabel,
			LocalIPAddress:    lcIPAddr,
			ExternalIPAddress: exIPAddr,
			Protocol:          h.protocolNBValueToNumber(mapping.Protocol),
			ExternalSwIfIndex: ifIdx,
			VrfID:             lcVrf,
			Flags: setNat44EdFlags(&nat44EdFlags{
				isTwiceNat:     mapping.TwiceNat == nat.DNat44_StaticMapping_ENABLED,
				isSelfTwiceNat: mapping.TwiceNat == nat.DNat44_StaticMapping_SELF,
				isOut2In:       true,
				isAddrOnly:     addrOnly,
			}),
			IsAdd: isAdd,
		}

		if !addrOnly {
			req.LocalPort = lcPort
			req.ExternalPort = uint16(mapping.ExternalPort)
		}

		// Applying(if needed) the override of IP address picking from twice-NAT address pool
		if mapping.TwiceNatPoolIp != "" {
			req.MatchPool = true
			req.PoolIPAddress, err = ipTo4Address(mapping.TwiceNatPoolIp)
			if err != nil {
				return errors.Errorf("cannot configure static mapping for DNAT %s: unable to parse "+
					"twice-NAT pool IP %s: %v", dnatLabel, mapping.TwiceNatPoolIp, err)
			}
		}

		reply := &vpp_nat_ed.Nat44AddDelStaticMappingV2Reply{}

		if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
			return err
		}
	} else {
		req := &vpp_nat_ei.Nat44EiAddDelStaticMapping{
			Tag:               dnatLabel,
			LocalIPAddress:    lcIPAddr,
			ExternalIPAddress: exIPAddr,
			Protocol:          h.protocolNBValueToNumber(mapping.Protocol),
			ExternalSwIfIndex: ifIdx,
			VrfID:             lcVrf,
			Flags: setNat44EiFlags(&nat44EiFlags{
				eiAddrOnlyMapping: addrOnly,
			}),
			IsAdd: isAdd,
		}

		if !addrOnly {
			req.LocalPort = lcPort
			req.ExternalPort = uint16(mapping.ExternalPort)
		}

		reply := &vpp_nat_ei.Nat44EiAddDelStaticMappingReply{}

		if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
			return err
		}
	}

	return nil
}

func (h *NatVppHandler) handleNat44EdStaticMappingLb(mapping *nat.DNat44_StaticMapping, dnatLabel string, isAdd bool) error {
	// check tag length limit
	if err := checkTagLength(dnatLabel); err != nil {
		return err
	}

	// parse external IP address
	exIPAddrByte, err := ipTo4Address(mapping.ExternalIp)
	if err != nil {
		return errors.Errorf("cannot configure LB static mapping for DNAT %s: unable to parse external IP %s: %v",
			dnatLabel, mapping.ExternalIp, err)
	}

	// In this case, external port is required
	if mapping.ExternalPort == 0 {
		return errors.Errorf("cannot configure LB static mapping for DNAT %s: external port is not set", dnatLabel)
	}

	// Transform local IP/Ports
	var locals []vpp_nat_ed.Nat44LbAddrPort
	for _, local := range mapping.LocalIps {
		if local.LocalPort == 0 {
			return errors.Errorf("cannot set local IP/Port for DNAT mapping %s: port is missing",
				dnatLabel)
		}

		localIP, err := ipTo4Address(local.LocalIp)
		if err != nil {
			return errors.Errorf("cannot set local IP/Port for DNAT mapping %s: unable to parse local IP %v: %v",
				dnatLabel, local.LocalIp, err)
		}

		locals = append(locals, vpp_nat_ed.Nat44LbAddrPort{
			Addr:        localIP,
			Port:        uint16(local.LocalPort),
			Probability: uint8(local.Probability),
			VrfID:       local.VrfId,
		})
	}

	req := &vpp_nat_ed.Nat44AddDelLbStaticMapping{
		Tag:    dnatLabel,
		Locals: locals,
		// LocalNum:     uint32(len(locals)), // should not be needed (will be set by struc)
		ExternalAddr: exIPAddrByte,
		ExternalPort: uint16(mapping.ExternalPort),
		Protocol:     h.protocolNBValueToNumber(mapping.Protocol),
		Flags: setNat44EdFlags(&nat44EdFlags{
			isTwiceNat:     mapping.TwiceNat == nat.DNat44_StaticMapping_ENABLED,
			isSelfTwiceNat: mapping.TwiceNat == nat.DNat44_StaticMapping_SELF,
			isOut2In:       true,
		}),
		IsAdd:    isAdd,
		Affinity: mapping.SessionAffinity,
	}

	reply := &vpp_nat_ed.Nat44AddDelLbStaticMappingReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to add/remove NAT44 static mapping with load balancing.
func (h *NatVppHandler) handleNat44StaticMappingLb(mapping *nat.DNat44_StaticMapping, dnatLabel string, isAdd bool) error {
	if h.ed {
		return h.handleNat44EdStaticMappingLb(mapping, dnatLabel, isAdd)
	} else {
		// no static mapping with load balancing implemented for EI nat yet
		return nil
	}
}

// Calls VPP binary API to add/remove NAT44 identity mapping.
func (h *NatVppHandler) handleNat44IdentityMapping(mapping *nat.DNat44_IdentityMapping, dnatLabel string, isAdd bool) (err error) {
	var ifIdx = NoInterface
	var ipAddr ip_types.IP4Address

	// check tag length limit
	if err := checkTagLength(dnatLabel); err != nil {
		return err
	}

	// get interface index
	if mapping.Interface != "" {
		ifMeta, found := h.ifIndexes.LookupByName(mapping.Interface)
		if !found {
			return errors.Errorf("failed to configure identity mapping for DNAT %s: addressed interface %s does not exist",
				dnatLabel, mapping.Interface)
		}
		ifIdx = interface_types.InterfaceIndex(ifMeta.SwIfIndex)
	}

	if ifIdx == NoInterface {
		// Case with IP (optionally port). Verify and parse input IP/port
		ipAddr, err = ipTo4Address(mapping.IpAddress)
		if err != nil {
			return errors.Errorf("failed to configure identity mapping for DNAT %s: unable to parse IP address %s: %v",
				dnatLabel, mapping.IpAddress, err)
		}
	}

	var addrOnly bool
	if mapping.Port == 0 {
		addrOnly = true
	}

	if h.ed {
		req := &vpp_nat_ed.Nat44AddDelIdentityMapping{
			Tag:       dnatLabel,
			Flags:     setNat44EdFlags(&nat44EdFlags{isAddrOnly: addrOnly}),
			IPAddress: ipAddr,
			Port:      uint16(mapping.Port),
			Protocol:  h.protocolNBValueToNumber(mapping.Protocol),
			SwIfIndex: ifIdx,
			VrfID:     mapping.VrfId,
			IsAdd:     isAdd,
		}

		reply := &vpp_nat_ed.Nat44AddDelIdentityMappingReply{}

		if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
			return err
		}
	} else {
		req := &vpp_nat_ei.Nat44EiAddDelIdentityMapping{
			Tag:       dnatLabel,
			Flags:     setNat44EiFlags(&nat44EiFlags{eiAddrOnlyMapping: addrOnly}),
			IPAddress: ipAddr,
			Port:      uint16(mapping.Port),
			Protocol:  h.protocolNBValueToNumber(mapping.Protocol),
			SwIfIndex: ifIdx,
			VrfID:     mapping.VrfId,
			IsAdd:     isAdd,
		}

		reply := &vpp_nat_ei.Nat44EiAddDelIdentityMappingReply{}

		if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
			return err
		}
	}

	return nil
}

func setNat44EdFlags(flags *nat44EdFlags) nat_types.NatConfigFlags {
	var flagsCfg nat_types.NatConfigFlags
	if flags.isTwiceNat {
		flagsCfg |= nat_types.NAT_IS_TWICE_NAT
	}
	if flags.isSelfTwiceNat {
		flagsCfg |= nat_types.NAT_IS_SELF_TWICE_NAT
	}
	if flags.isOut2In {
		flagsCfg |= nat_types.NAT_IS_OUT2IN_ONLY
	}
	if flags.isAddrOnly {
		flagsCfg |= nat_types.NAT_IS_ADDR_ONLY
	}
	if flags.isOutside {
		flagsCfg |= nat_types.NAT_IS_OUTSIDE
	}
	if flags.isInside {
		flagsCfg |= nat_types.NAT_IS_INSIDE
	}
	if flags.isStatic {
		flagsCfg |= nat_types.NAT_IS_STATIC
	}
	if flags.isExtHostValid {
		flagsCfg |= nat_types.NAT_IS_EXT_HOST_VALID
	}
	return flagsCfg
}

func setNat44EiFlags(flags *nat44EiFlags) vpp_nat_ei.Nat44EiConfigFlags {
	var flagsCfg vpp_nat_ei.Nat44EiConfigFlags
	if flags.eiStaticMappingOnly {
		flagsCfg |= vpp_nat_ei.NAT44_EI_STATIC_MAPPING_ONLY
	}
	if flags.eiConnectionTracking {
		flagsCfg |= vpp_nat_ei.NAT44_EI_CONNECTION_TRACKING
	}
	if flags.eiOut2InDpo {
		flagsCfg |= vpp_nat_ei.NAT44_EI_OUT2IN_DPO
	}
	if flags.eiAddrOnlyMapping {
		flagsCfg |= vpp_nat_ei.NAT44_EI_ADDR_ONLY_MAPPING
	}
	if flags.eiIfInside {
		flagsCfg |= vpp_nat_ei.NAT44_EI_IF_INSIDE
	}
	if flags.eiIfOutside {
		flagsCfg |= vpp_nat_ei.NAT44_EI_IF_OUTSIDE
	}
	if flags.eiStaticMapping {
		flagsCfg |= vpp_nat_ei.NAT44_EI_STATIC_MAPPING
	}
	return flagsCfg
}

func ipTo4Address(ipStr string) (addr ip_types.IP4Address, err error) {
	netIP := net.ParseIP(ipStr)
	if netIP == nil {
		return ip_types.IP4Address{}, fmt.Errorf("invalid IP: %q", ipStr)
	}
	if ip4 := netIP.To4(); ip4 != nil {
		var ip4Addr ip_types.IP4Address
		copy(ip4Addr[:], netIP.To4())
		addr = ip4Addr
	} else {
		return ip_types.IP4Address{}, fmt.Errorf("required IPv4, provided: %q", ipStr)
	}
	return
}

// checkTagLength serves as a validator for static/identity mapping tag length
func checkTagLength(tag string) error {
	if len(tag) > maxTagLen {
		return errors.Errorf("DNAT label '%s' has %d bytes, max allowed is %d",
			tag, len(tag), maxTagLen)
	}
	return nil
}
