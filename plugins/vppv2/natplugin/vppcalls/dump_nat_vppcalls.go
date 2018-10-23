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

package vppcalls

import (
	"bytes"
	"fmt"
	"net"
	"sort"

	bin_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/nat"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/nat"
)

// Nat44GlobalConfigDump dumps global NAT config in NB format.
func (h *NatVppHandler) Nat44GlobalConfigDump() (*nat.Nat44Global, error) {
	isEnabled, err := h.isNat44ForwardingEnabled()
	if err != nil {
		return nil, err
	}
	natInterfaces, err := h.nat44InterfaceDump()
	if err != nil {
		return nil, err
	}
	natAddressPool, err := h.nat44AddressDump()
	if err != nil {
		return nil, err
	}
	vrIPv4, vrIPv6, err := h.virtualReassemblyDump()
	if err != nil {
		return nil, err
	}

	// combine into the global NAT configuration
	return &nat.Nat44Global{
		Forwarding:            isEnabled,
		NatInterfaces:         natInterfaces,
		AddressPool:           natAddressPool,
		VirtualReassemblyIpv4: vrIPv4,
		VirtualReassemblyIpv6: vrIPv6,
	}, nil
}

// NAT44NatDump dumps all configured DNAT configurations ordered by label.
func (h *NatVppHandler) Nat44DNatDump() (dnats []*nat.Nat44DNat, err error) {
	dnatMap := make(map[string]*nat.Nat44DNat)

	// Static mappings
	natStMappings, err := h.nat44StaticMappingDump()
	if err != nil {
		return nil, fmt.Errorf("failed to dump NAT44 static mappings: %v", err)
	}
	for label, mapping := range natStMappings {
		dnat := getOrCreateDNAT(dnatMap, label)
		dnat.StMappings = append(dnat.StMappings, mapping)
	}

	// Static mappings with load balancer
	natStLbMappings, err := h.nat44StaticMappingLbDump()
	if err != nil {
		return nil, fmt.Errorf("failed to dump NAT44 static mappings with load balancer: %v", err)
	}
	for label, mapping := range natStLbMappings {
		dnat := getOrCreateDNAT(dnatMap, label)
		dnat.StMappings = append(dnat.StMappings, mapping)
	}

	// Identity mappings
	natIDMappings, err := h.nat44IdentityMappingDump()
	if err != nil {
		return nil, fmt.Errorf("failed to dump NAT44 identity mappings: %v", err)
	}
	for label, mapping := range natIDMappings {
		dnat := getOrCreateDNAT(dnatMap, label)
		dnat.IdMappings = append(dnat.IdMappings, mapping)
	}

	// Convert map of DNAT configurations into a list.
	for _, dnat := range dnatMap {
		dnats = append(dnats, dnat)
	}

	// sort to simplify testing
	sort.Slice(dnats, func(i, j int) bool { return dnats[i].Label < dnats[j].Label })

	return dnats, nil
}

// nat44AddressDump returns a list of NAT44 address pools configured in the VPP
func (h *NatVppHandler) nat44AddressDump() (addressPool []*nat.Nat44Global_NatAddress, err error) {
	req := &bin_api.Nat44AddressDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44AddressDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 Address pool: %v", err)
		}
		if stop {
			break
		}

		ipAddress := net.IP(msg.IPAddress)

		addressPool = append(addressPool, &nat.Nat44Global_NatAddress{
			Address:  ipAddress.To4().String(),
			VrfId:    msg.VrfID,
			TwiceNat: uintToBool(msg.TwiceNat),
		})
	}

	return
}

// virtualReassemblyDump returns current NAT44 virtual-reassembly configuration. The output config may be nil.
func (h *NatVppHandler) virtualReassemblyDump() (vrIPv4 *nat.Nat44Global_VirtualReassembly, vrIPv6 *nat.Nat44Global_VirtualReassembly, err error) {
	req := &bin_api.NatGetReass{}
	reply := &bin_api.NatGetReassReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, nil, fmt.Errorf("failed to get NAT44 virtual reassembly configuration: %v", err)
	}
	if reply.Retval != 0 {
		return nil, nil, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	vrIPv4 = &nat.Nat44Global_VirtualReassembly{
		Timeout:  reply.IP4Timeout,
		MaxReass: uint32(reply.IP4MaxReass),
		MaxFrag:  uint32(reply.IP4MaxFrag),
		DropFrag: uintToBool(reply.IP4DropFrag),
	}
	vrIPv6 = &nat.Nat44Global_VirtualReassembly{
		Timeout:  reply.IP6Timeout,
		MaxReass: uint32(reply.IP6MaxReass),
		MaxFrag:  uint32(reply.IP6MaxFrag),
		DropFrag: uintToBool(reply.IP6DropFrag),
	}

	return
}

// nat44StaticMappingDump returns a map of static mapping tag/data pairs
func (h *NatVppHandler) nat44StaticMappingDump() (entries map[string]*nat.Nat44DNat_StaticMapping, err error) {
	entries = make(map[string]*nat.Nat44DNat_StaticMapping)
	req := &bin_api.Nat44StaticMappingDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44StaticMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 static mapping: %v", err)
		}
		if stop {
			break
		}
		lcIPAddress := net.IP(msg.LocalIPAddress)
		exIPAddress := net.IP(msg.ExternalIPAddress)

		// Parse tag (DNAT label)
		tag := string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0])

		// Add mapping into the map.
		entries[tag] = &nat.Nat44DNat_StaticMapping{
			ExternalInterface: func(ifIdx uint32) string {
				ifName, _, found := h.ifIndexes.LookupBySwIfIndex(ifIdx)
				if !found && ifIdx != NoInterface {
					h.log.Warnf("Interface with index %v not found in the mapping", ifIdx)
				}
				return ifName
			}(msg.ExternalSwIfIndex),
			ExternalIp:   exIPAddress.To4().String(),
			ExternalPort: uint32(msg.ExternalPort),
			LocalIps: []*nat.Nat44DNat_StaticMapping_LocalIP{ // single-value
				{
					VrfId:     msg.VrfID,
					LocalIp:   lcIPAddress.To4().String(),
					LocalPort: uint32(msg.LocalPort),
				},
			},
			Protocol: h.protocolNumberToNBValue(msg.Protocol),
			TwiceNat: h.getTwiceNatMode(msg.TwiceNat, msg.SelfTwiceNat),
		}
	}

	return entries, nil
}

// nat44StaticMappingLbDump returns a map of static mapping tag/data pairs with load balancer
func (h *NatVppHandler) nat44StaticMappingLbDump() (entries map[string]*nat.Nat44DNat_StaticMapping, err error) {
	entries = make(map[string]*nat.Nat44DNat_StaticMapping)
	req := &bin_api.Nat44LbStaticMappingDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44LbStaticMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 lb-static mapping: %v", err)
		}
		if stop {
			break
		}

		// Parse tag (DNAT label)
		tag := string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0])

		// Prepare localIPs
		var locals []*nat.Nat44DNat_StaticMapping_LocalIP
		for _, localIPVal := range msg.Locals {
			localIP := net.IP(localIPVal.Addr)
			locals = append(locals, &nat.Nat44DNat_StaticMapping_LocalIP{
				VrfId:       localIPVal.VrfID,
				LocalIp:     localIP.To4().String(),
				LocalPort:   uint32(localIPVal.Port),
				Probability: uint32(localIPVal.Probability),
			})
		}
		exIPAddress := net.IP(msg.ExternalAddr)

		// Add mapping into the map.
		entries[tag] = &nat.Nat44DNat_StaticMapping{
			ExternalIp:   exIPAddress.To4().String(),
			ExternalPort: uint32(msg.ExternalPort),
			LocalIps:     locals,
			Protocol:     h.protocolNumberToNBValue(msg.Protocol),
			TwiceNat:     h.getTwiceNatMode(msg.TwiceNat, msg.SelfTwiceNat),
		}
	}

	return entries, nil
}

// nat44IdentityMappingDump returns a map of identity mapping tag/data pairs
func (h *NatVppHandler) nat44IdentityMappingDump() (entries map[string]*nat.Nat44DNat_IdentityMapping, err error) {
	entries = make(map[string]*nat.Nat44DNat_IdentityMapping)
	req := &bin_api.Nat44IdentityMappingDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44IdentityMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 identity mapping: %v", err)
		}
		if stop {
			break
		}

		ipAddress := net.IP(msg.IPAddress)

		// Parse tag (DNAT label)
		tag := string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0])

		// Add mapping into the map.
		entries[tag] = &nat.Nat44DNat_IdentityMapping{
			VrfId: msg.VrfID,
			AddressedInterface: func(ifIdx uint32) string {
				ifName, _, found := h.ifIndexes.LookupBySwIfIndex(ifIdx)
				if !found && ifIdx != NoInterface {
					h.log.Warnf("Interface with index %v not found in the mapping", ifIdx)
				}
				return ifName
			}(msg.SwIfIndex),
			IpAddress: ipAddress.To4().String(),
			Port:      uint32(msg.Port),
			Protocol:  h.protocolNumberToNBValue(msg.Protocol),
		}
	}

	return entries, nil
}

// nat44InterfaceDump dumps NAT interface features.
func (h *NatVppHandler) nat44InterfaceDump() (interfaces []*nat.Nat44Global_NatInterface, err error) {

	/* dump non-Output interfaces first */
	req1 := &bin_api.Nat44InterfaceDump{}
	reqContext := h.callsChannel.SendMultiRequest(req1)

	for {
		msg := &bin_api.Nat44InterfaceDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface: %v", err)
		}
		if stop {
			break
		}

		// Find interface name
		ifName, _, found := h.ifIndexes.LookupBySwIfIndex(msg.SwIfIndex)
		if !found {
			h.log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}

		if msg.IsInside == 0 || msg.IsInside == 2 {
			interfaces = append(interfaces, &nat.Nat44Global_NatInterface{
				Name:     ifName,
				IsInside: false,
			})
		}
		if msg.IsInside == 1 || msg.IsInside == 2 {
			interfaces = append(interfaces, &nat.Nat44Global_NatInterface{
				Name:     ifName,
				IsInside: true,
			})
		}
	}

	/* dump Output interfaces next */
	req2 := &bin_api.Nat44InterfaceOutputFeatureDump{}
	reqContext = h.callsChannel.SendMultiRequest(req2)

	for {
		msg := &bin_api.Nat44InterfaceOutputFeatureDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
		}
		if stop {
			break
		}

		// Find interface name
		ifName, _, found := h.ifIndexes.LookupBySwIfIndex(msg.SwIfIndex)
		if !found {
			h.log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}

		interfaces = append(interfaces, &nat.Nat44Global_NatInterface{
			Name:          ifName,
			IsInside:      uintToBool(msg.IsInside),
			OutputFeature: true,
		})
	}

	return interfaces, nil
}

// Nat44IsForwardingEnabled checks if the NAT forwarding is enabled.
func (h *NatVppHandler) isNat44ForwardingEnabled() (isEnabled bool, err error) {
	req := &bin_api.Nat44ForwardingIsEnabled{}

	reply := &bin_api.Nat44ForwardingIsEnabledReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return false, fmt.Errorf("failed to dump NAT44 forwarding: %v", err)
	}

	isEnabled = uintToBool(reply.Enabled)
	return isEnabled, nil
}

// protocolNumberToNBValue converts protocol numeric representation into the corresponding enum
// enum value from the NB model.
func (h *NatVppHandler) protocolNumberToNBValue(protocol uint8) (proto nat.Nat44DNat_Protocol) {
	switch protocol {
	case TCP:
		return nat.Nat44DNat_TCP
	case UDP:
		return nat.Nat44DNat_UDP
	case ICMP:
		return nat.Nat44DNat_ICMP
	default:
		h.log.Warnf("Unknown protocol %v", protocol)
		return 0
	}
}

// protocolNBValueToNumber converts protocol enum value from the NB model into the
// corresponding numeric representation.
func (h *NatVppHandler) protocolNBValueToNumber(protocol nat.Nat44DNat_Protocol) (proto uint8) {
	switch protocol {
	case nat.Nat44DNat_TCP:
		return TCP
	case nat.Nat44DNat_UDP:
		return UDP
	case nat.Nat44DNat_ICMP:
		return ICMP
	default:
		h.log.Warnf("Unknown protocol %v, defaulting to TCP", protocol)
		return TCP
	}
}

func (h *NatVppHandler) getTwiceNatMode(twiceNat, selfTwiceNat uint8) nat.Nat44DNat_StaticMapping_TwiceNatMode {
	if twiceNat > 0 {
		if selfTwiceNat > 0 {
			h.log.Warnf("Both TwiceNAT and self-TwiceNAT are enabled")
			return 0
		}
		return nat.Nat44DNat_StaticMapping_ENABLED
	}
	if selfTwiceNat > 0 {
		return nat.Nat44DNat_StaticMapping_SELF
	}
	return nat.Nat44DNat_StaticMapping_DISABLED
}

func getOrCreateDNAT(dnats map[string]*nat.Nat44DNat, label string) *nat.Nat44DNat {
	if _, created := dnats[label]; !created {
		dnats[label] = &nat.Nat44DNat{Label: label}
	}
	return dnats[label]
}

func uintToBool(value uint8) bool {
	if value == 0 {
		return false
	}
	return true
}
