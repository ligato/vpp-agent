//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package vpp1908

import (
	"fmt"
	"net"

	"github.com/pkg/errors"

	nat "github.com/ligato/vpp-agent/api/models/vpp/nat"
	natba "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/nat"
)

// Num protocol representation
const (
	ICMP uint8 = 1
	TCP  uint8 = 6
	UDP  uint8 = 17
)

const (
	// NoInterface is sw-if-idx which means 'no interface'
	NoInterface = natba.InterfaceIndex(^uint32(0))
	// Maximal length of tag
	maxTagLen = 64
)

// holds a list of NAT44 flags set
type nat44Flags struct {
	isTwiceNat     bool
	isSelfTwiceNat bool
	isOut2In       bool
	isAddrOnly     bool
	isOutside      bool
	isInside       bool
	isStatic       bool
	isExtHostValid bool
}

// SetNat44Forwarding configures NAT44 forwarding.
func (h *NatVppHandler) SetNat44Forwarding(enableFwd bool) error {
	req := &natba.Nat44ForwardingEnableDisable{
		Enable: enableFwd,
	}
	reply := &natba.Nat44ForwardingEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
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

// AddNat44Address adds new IPv4 address into the NAT44 pool.
func (h *NatVppHandler) AddNat44Address(address string, vrf uint32, twiceNat bool) error {
	return h.handleNat44AddressPool(address, vrf, twiceNat, true)
}

// DelNat44Address removes existing IPv4 address from the NAT44 pool.
func (h *NatVppHandler) DelNat44Address(address string, vrf uint32, twiceNat bool) error {
	return h.handleNat44AddressPool(address, vrf, twiceNat, false)
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

// Calls VPP binary API to set/unset interface NAT44 feature.
func (h *NatVppHandler) handleNat44Interface(iface string, isInside, isAdd bool) error {
	// get interface metadata
	ifaceMeta, found := h.ifIndexes.LookupByName(iface)
	if !found {
		return errors.New("failed to get interface metadata")
	}

	req := &natba.Nat44InterfaceAddDelFeature{
		SwIfIndex: natba.InterfaceIndex(ifaceMeta.SwIfIndex),
		Flags:     setNat44Flags(&nat44Flags{isInside: isInside}),
		IsAdd:     isAdd,
	}
	reply := &natba.Nat44InterfaceAddDelFeatureReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to set/unset interface NAT44 output feature
func (h *NatVppHandler) handleNat44InterfaceOutputFeature(iface string, isInside, isAdd bool) error {
	// get interface metadata
	ifaceMeta, found := h.ifIndexes.LookupByName(iface)
	if !found {
		return errors.New("failed to get interface metadata")
	}

	req := &natba.Nat44InterfaceAddDelOutputFeature{
		SwIfIndex: natba.InterfaceIndex(ifaceMeta.SwIfIndex),
		Flags:     setNat44Flags(&nat44Flags{isInside: isInside}),
		IsAdd:     isAdd,
	}
	reply := &natba.Nat44InterfaceAddDelOutputFeatureReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to add/remove address to/from the NAT44 pool.
func (h *NatVppHandler) handleNat44AddressPool(address string, vrf uint32, twiceNat, isAdd bool) error {
	ipAddr, err := ipTo4Address(address)
	if err != nil {
		return errors.Errorf("unable to parse address %s from the NAT pool: %v", address, err)
	}

	req := &natba.Nat44AddDelAddressRange{
		FirstIPAddress: ipAddr,
		LastIPAddress:  ipAddr,
		VrfID:          vrf,
		Flags:          setNat44Flags(&nat44Flags{isTwiceNat: twiceNat}),
		IsAdd:          isAdd,
	}
	reply := &natba.Nat44AddDelAddressRangeReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to setup NAT virtual reassembly
func (h *NatVppHandler) handleNatVirtualReassembly(vrCfg *nat.VirtualReassembly, isIpv6 bool) error {
	req := &natba.NatSetReass{
		Timeout:  vrCfg.Timeout,
		MaxReass: uint16(vrCfg.MaxReassemblies),
		MaxFrag:  uint8(vrCfg.MaxFragments),
		DropFrag: boolToUint(vrCfg.DropFragments),
		IsIP6:    isIpv6,
	}
	reply := &natba.NatSetReassReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to add/remove NAT44 static mapping
func (h *NatVppHandler) handleNat44StaticMapping(mapping *nat.DNat44_StaticMapping, dnatLabel string, isAdd bool) error {
	var ifIdx = NoInterface
	var exIPAddr natba.IP4Address

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
		ifIdx = natba.InterfaceIndex(ifMeta.SwIfIndex)
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

	req := &natba.Nat44AddDelStaticMapping{
		Tag:               dnatLabel,
		LocalIPAddress:    lcIPAddr,
		ExternalIPAddress: exIPAddr,
		Protocol:          h.protocolNBValueToNumber(mapping.Protocol),
		ExternalSwIfIndex: ifIdx,
		VrfID:             lcVrf,
		Flags: setNat44Flags(&nat44Flags{
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

	reply := &natba.Nat44AddDelStaticMappingReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to add/remove NAT44 static mapping with load balancing.
func (h *NatVppHandler) handleNat44StaticMappingLb(mapping *nat.DNat44_StaticMapping, dnatLabel string, isAdd bool) error {
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
	var locals []natba.Nat44LbAddrPort
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

		locals = append(locals, natba.Nat44LbAddrPort{
			Addr:        localIP,
			Port:        uint16(local.LocalPort),
			Probability: uint8(local.Probability),
			VrfID:       local.VrfId,
		})
	}

	req := &natba.Nat44AddDelLbStaticMapping{
		Tag:    dnatLabel,
		Locals: locals,
		//LocalNum:     uint32(len(locals)), // should not be needed (will be set by struc)
		ExternalAddr: exIPAddrByte,
		ExternalPort: uint16(mapping.ExternalPort),
		Protocol:     h.protocolNBValueToNumber(mapping.Protocol),
		Flags: setNat44Flags(&nat44Flags{
			isTwiceNat:     mapping.TwiceNat == nat.DNat44_StaticMapping_ENABLED,
			isSelfTwiceNat: mapping.TwiceNat == nat.DNat44_StaticMapping_SELF,
			isOut2In:       true,
		}),
		IsAdd:    isAdd,
		Affinity: mapping.SessionAffinity,
	}

	reply := &natba.Nat44AddDelLbStaticMappingReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// Calls VPP binary API to add/remove NAT44 identity mapping.
func (h *NatVppHandler) handleNat44IdentityMapping(mapping *nat.DNat44_IdentityMapping, dnatLabel string, isAdd bool) (err error) {
	var ifIdx = NoInterface
	var ipAddr natba.IP4Address

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
		ifIdx = natba.InterfaceIndex(ifMeta.SwIfIndex)
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

	req := &natba.Nat44AddDelIdentityMapping{
		Tag:       dnatLabel,
		Flags:     setNat44Flags(&nat44Flags{isAddrOnly: addrOnly}),
		IPAddress: ipAddr,
		Port:      uint16(mapping.Port),
		Protocol:  h.protocolNBValueToNumber(mapping.Protocol),
		SwIfIndex: ifIdx,
		VrfID:     mapping.VrfId,
		IsAdd:     isAdd,
	}

	reply := &natba.Nat44AddDelIdentityMappingReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func setNat44Flags(flags *nat44Flags) natba.NatConfigFlags {
	var flagsCfg natba.NatConfigFlags
	if flags.isTwiceNat {
		flagsCfg |= natba.NAT_IS_TWICE_NAT
	}
	if flags.isSelfTwiceNat {
		flagsCfg |= natba.NAT_IS_SELF_TWICE_NAT
	}
	if flags.isOut2In {
		flagsCfg |= natba.NAT_IS_OUT2IN_ONLY
	}
	if flags.isAddrOnly {
		flagsCfg |= natba.NAT_IS_ADDR_ONLY
	}
	if flags.isOutside {
		flagsCfg |= natba.NAT_IS_OUTSIDE
	}
	if flags.isInside {
		flagsCfg |= natba.NAT_IS_INSIDE
	}
	if flags.isStatic {
		flagsCfg |= natba.NAT_IS_STATIC
	}
	if flags.isExtHostValid {
		flagsCfg |= natba.NAT_IS_EXT_HOST_VALID
	}
	return flagsCfg
}

func ipTo4Address(ipStr string) (addr natba.IP4Address, err error) {
	netIP := net.ParseIP(ipStr)
	if netIP == nil {
		return natba.IP4Address{}, fmt.Errorf("invalid IP: %q", ipStr)
	}
	if ip4 := netIP.To4(); ip4 != nil {
		var ip4Addr natba.IP4Address
		copy(ip4Addr[:], netIP.To4())
		addr = ip4Addr
	} else {
		return natba.IP4Address{}, fmt.Errorf("required IPv4, provided: %q", ipStr)
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

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}
