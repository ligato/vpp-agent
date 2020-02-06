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

package vpp2001

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	vpp_bond "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/bond"
	vpp_dhcp "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/dhcp"
	vpp_gre "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/gre"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interface_types"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ipsec_types"
	vpp_memif "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/memif"
	vpp_tapv2 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/tapv2"
	vpp_vxlan "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/vxlan"
	vpp_vxlangpe "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/vxlan_gpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

const (
	// allInterfaces defines unspecified interface index
	allInterfaces = ^uint32(0)

	// prefix prepended to internal names of untagged interfaces to construct unique
	// logical names
	untaggedIfPreffix = "UNTAGGED-"
)

const (
	// Default VPP MTU value
	defaultVPPMtu = 9216

	// MAC length
	macLength = 6
)

func getMtu(vppMtu uint16) uint32 {
	// If default VPP MTU value is set, return 0 (it means MTU was not set in the NB config)
	if vppMtu == defaultVPPMtu {
		return 0
	}
	return uint32(vppMtu)
}

func (h *InterfaceVppHandler) DumpInterfacesByType(ctx context.Context, reqType ifs.Interface_Type) (map[uint32]*vppcalls.InterfaceDetails, error) {
	// Dump all
	ifs, err := h.DumpInterfaces(ctx)
	if err != nil {
		return nil, err
	}
	// Filter by type
	for ifIdx, ifData := range ifs {
		if ifData.Interface.Type != reqType {
			delete(ifs, ifIdx)
		}
	}

	return ifs, nil
}

func (h *InterfaceVppHandler) dumpInterfaces(ifIdxs ...uint32) (map[uint32]*vppcalls.InterfaceDetails, error) {
	// map for the resulting interfaces
	interfaces := make(map[uint32]*vppcalls.InterfaceDetails)

	ifIdx := allInterfaces
	if len(ifIdxs) > 0 {
		ifIdx = ifIdxs[0]
	}
	// First, dump all interfaces to create initial data.
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_ifs.SwInterfaceDump{
		SwIfIndex: vpp_ifs.InterfaceIndex(ifIdx),
	})
	for {
		ifDetails := &vpp_ifs.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(ifDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dump interface: %v", err)
		}

		ifaceName := strings.TrimRight(ifDetails.InterfaceName, "\x00")
		physAddr := make(net.HardwareAddr, macLength)
		copy(physAddr, ifDetails.L2Address[:])

		details := &vppcalls.InterfaceDetails{
			Interface: &ifs.Interface{
				Name: strings.TrimRight(ifDetails.Tag, "\x00"),
				// the type may be amended later by further dumps
				Type:        guessInterfaceType(ifaceName),
				Enabled:     isAdminStateUp(ifDetails.Flags),
				PhysAddress: net.HardwareAddr(ifDetails.L2Address[:]).String(),
				Mtu:         getMtu(ifDetails.LinkMtu),
			},
			Meta: &vppcalls.InterfaceMeta{
				SwIfIndex:      uint32(ifDetails.SwIfIndex),
				SupSwIfIndex:   ifDetails.SupSwIfIndex,
				L2Address:      physAddr,
				InternalName:   ifaceName,
				IsAdminStateUp: isAdminStateUp(ifDetails.Flags),
				IsLinkStateUp:  isLinkStateUp(ifDetails.Flags),
				LinkDuplex:     uint32(ifDetails.LinkDuplex),
				LinkMTU:        ifDetails.LinkMtu,
				MTU:            ifDetails.Mtu,
				LinkSpeed:      ifDetails.LinkSpeed,
				SubID:          ifDetails.SubID,
				Tag:            strings.TrimRight(ifDetails.Tag, "\x00"),
			},
		}

		// sub interface
		if ifDetails.SupSwIfIndex != uint32(ifDetails.SwIfIndex) {
			details.Interface.Type = ifs.Interface_SUB_INTERFACE
			details.Interface.Link = &ifs.Interface_Sub{
				Sub: &ifs.SubInterface{
					ParentName:  interfaces[ifDetails.SupSwIfIndex].Interface.Name,
					SubId:       ifDetails.SubID,
					TagRwOption: getTagRwOption(ifDetails.VtrOp),
					PushDot1Q:   uintToBool(uint8(ifDetails.VtrPushDot1q)),
					Tag1:        ifDetails.VtrTag1,
					Tag2:        ifDetails.VtrTag2,
				},
			}
		}
		// Fill name for physical interfaces (they are mostly without tag)
		switch details.Interface.Type {
		case ifs.Interface_DPDK:
			details.Interface.Name = ifaceName
		case ifs.Interface_AF_PACKET:
			details.Interface.Link = &ifs.Interface_Afpacket{
				Afpacket: &ifs.AfpacketLink{
					HostIfName: strings.TrimPrefix(ifaceName, "host-"),
				},
			}
		}
		if details.Interface.Name == "" {
			// untagged interface - generate a logical name for it
			// (apart from local0 it will get removed by resync)
			details.Interface.Name = untaggedIfPreffix + ifaceName
		}
		interfaces[uint32(ifDetails.SwIfIndex)] = details
	}

	return interfaces, nil
}

func (h *InterfaceVppHandler) DumpInterfaces(ctx context.Context) (map[uint32]*vppcalls.InterfaceDetails, error) {
	interfaces, err := h.dumpInterfaces()
	if err != nil {
		return nil, err
	}

	// Get DHCP clients
	dhcpClients, err := h.DumpDhcpClients()
	if err != nil {
		return nil, fmt.Errorf("failed to dump interface DHCP clients: %v", err)
	}

	// Get IP addresses before VRF
	err = h.dumpIPAddressDetails(interfaces, false, dhcpClients)
	if err != nil {
		return nil, err
	}
	err = h.dumpIPAddressDetails(interfaces, true, dhcpClients)
	if err != nil {
		return nil, err
	}

	// Get unnumbered interfaces
	unnumbered, err := h.dumpUnnumberedDetails()
	if err != nil {
		return nil, fmt.Errorf("failed to dump unnumbered interfaces: %v", err)
	}

	// dump VXLAN details before VRFs (used by isIpv6Interface)
	err = h.dumpVxlanDetails(interfaces)
	if err != nil {
		return nil, err
	}

	err = h.dumpVxLanGpeDetails(interfaces)
	if err != nil {
		return nil, err
	}

	// Get interface VRF for every IP family, fill DHCP if set and resolve unnumbered interface setup
	for _, ifData := range interfaces {
		// VRF is stored in metadata for both, IPv4 and IPv6. If the interface is an IPv6 interface (it contains at least
		// one IPv6 address), appropriate VRF is stored also in modelled data
		ipv4Vrf, err := h.GetInterfaceVrf(ifData.Meta.SwIfIndex)
		if err != nil {
			return nil, fmt.Errorf("interface dump: failed to get IPv4 VRF from interface %d: %v",
				ifData.Meta.SwIfIndex, err)
		}
		ifData.Meta.VrfIPv4 = ipv4Vrf
		ipv6Vrf, err := h.GetInterfaceVrfIPv6(ifData.Meta.SwIfIndex)
		if err != nil {
			return nil, fmt.Errorf("interface dump: failed to get IPv6 VRF from interface %d: %v",
				ifData.Meta.SwIfIndex, err)
		}
		ifData.Meta.VrfIPv6 = ipv6Vrf
		if isIPv6If, err := isIpv6Interface(ifData.Interface); err != nil {
			return interfaces, err
		} else if isIPv6If {
			ifData.Interface.Vrf = ipv6Vrf
		} else {
			ifData.Interface.Vrf = ipv4Vrf
		}

		// DHCP
		dhcpData, ok := dhcpClients[ifData.Meta.SwIfIndex]
		if ok {
			ifData.Interface.SetDhcpClient = true
			ifData.Meta.Dhcp = dhcpData
		}
		// Unnumbered
		ifWithIPIdx, ok := unnumbered[ifData.Meta.SwIfIndex]
		if ok {
			// Find unnumbered interface
			var ifWithIPName string
			ifWithIP, ok := interfaces[ifWithIPIdx]
			if ok {
				ifWithIPName = ifWithIP.Interface.Name
			} else {
				h.log.Debugf("cannot find name of the ip-interface for unnumbered %s", ifData.Interface.Name)
				ifWithIPName = "<unknown>"
			}
			ifData.Interface.Unnumbered = &ifs.Interface_Unnumbered{
				InterfaceWithIp: ifWithIPName,
			}
		}
	}

	err = h.dumpMemifDetails(ctx, interfaces)
	if err != nil {
		return nil, err
	}

	err = h.dumpTapDetails(interfaces)
	if err != nil {
		return nil, err
	}

	err = h.dumpIPSecTunnelDetails(interfaces)
	if err != nil {
		return nil, err
	}

	err = h.dumpVmxNet3Details(interfaces)
	if err != nil {
		return nil, err
	}

	err = h.dumpBondDetails(interfaces)
	if err != nil {
		return nil, err
	}

	err = h.dumpGreDetails(interfaces)
	if err != nil {
		return nil, err
	}

	err = h.dumpGtpuDetails(interfaces)
	if err != nil {
		return nil, err
	}

	// Rx-placement dump is last since it uses interface type-specific data
	err = h.dumpRxPlacement(interfaces)
	if err != nil {
		return nil, err
	}

	return interfaces, nil
}

// DumpDhcpClients returns a slice of DhcpMeta with all interfaces and other DHCP-related information available
func (h *InterfaceVppHandler) DumpDhcpClients() (map[uint32]*vppcalls.Dhcp, error) {
	dhcpData := make(map[uint32]*vppcalls.Dhcp)
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_dhcp.DHCPClientDump{})

	for {
		dhcpDetails := &vpp_dhcp.DHCPClientDetails{}
		last, err := reqCtx.ReceiveReply(dhcpDetails)
		if last {
			break
		}
		if err != nil {
			return nil, err
		}
		client := dhcpDetails.Client
		lease := dhcpDetails.Lease

		// DHCP client data
		dhcpClient := &vppcalls.Client{
			SwIfIndex:        uint32(client.SwIfIndex),
			Hostname:         strings.TrimRight(client.Hostname, "\x00"),
			ID:               string(bytes.SplitN(client.ID, []byte{0x00}, 2)[0]),
			WantDhcpEvent:    client.WantDHCPEvent,
			SetBroadcastFlag: client.SetBroadcastFlag,
			PID:              client.PID,
		}

		// DHCP lease data
		dhcpLease := &vppcalls.Lease{
			SwIfIndex:     uint32(lease.SwIfIndex),
			State:         uint8(lease.State),
			Hostname:      strings.TrimRight(lease.Hostname, "\x00"),
			IsIPv6:        lease.IsIPv6,
			HostAddress:   dhcpAddressToString(lease.HostAddress, uint32(lease.MaskWidth), lease.IsIPv6),
			RouterAddress: dhcpAddressToString(lease.RouterAddress, uint32(lease.MaskWidth), lease.IsIPv6),
			HostMac:       net.HardwareAddr(lease.HostMac[:]).String(),
		}

		// DHCP metadata
		dhcpData[uint32(client.SwIfIndex)] = &vppcalls.Dhcp{
			Client: dhcpClient,
			Lease:  dhcpLease,
		}
	}

	return dhcpData, nil
}

// DumpInterfaceStates dumps link and administrative state of every interface.
func (h *InterfaceVppHandler) DumpInterfaceStates(ifIdxs ...uint32) (map[uint32]*vppcalls.InterfaceState, error) {
	// Dump all interface states if not specified.
	if len(ifIdxs) == 0 {
		ifIdxs = []uint32{allInterfaces}
	}

	ifStates := make(map[uint32]*vppcalls.InterfaceState)
	for _, ifIdx := range ifIdxs {
		reqCtx := h.callsChannel.SendMultiRequest(&vpp_ifs.SwInterfaceDump{
			SwIfIndex: vpp_ifs.InterfaceIndex(ifIdx),
		})
		for {
			ifDetails := &vpp_ifs.SwInterfaceDetails{}
			stop, err := reqCtx.ReceiveReply(ifDetails)
			if stop {
				break // Break from the loop.
			}
			if err != nil {
				return nil, fmt.Errorf("failed to dump interface states: %v", err)
			}

			physAddr := make(net.HardwareAddr, macLength)
			copy(physAddr, ifDetails.L2Address[:])

			ifaceState := vppcalls.InterfaceState{
				SwIfIndex:    uint32(ifDetails.SwIfIndex),
				InternalName: strings.TrimRight(ifDetails.InterfaceName, "\x00"),
				PhysAddress:  physAddr,
				AdminState:   adminStateToInterfaceStatus(ifDetails.Flags),
				LinkState:    linkStateToInterfaceStatus(ifDetails.Flags),
				LinkDuplex:   toLinkDuplex(ifDetails.LinkDuplex),
				LinkSpeed:    toLinkSpeed(ifDetails.LinkSpeed),
				LinkMTU:      ifDetails.LinkMtu,
			}
			ifStates[uint32(ifDetails.SwIfIndex)] = &ifaceState
		}
	}

	return ifStates, nil
}

func toLinkDuplex(duplex vpp_ifs.LinkDuplex) ifs.InterfaceState_Duplex {
	switch duplex {
	case 1:
		return ifs.InterfaceState_HALF
	case 2:
		return ifs.InterfaceState_FULL
	default:
		return ifs.InterfaceState_UNKNOWN_DUPLEX
	}
}

const megabit = 1000000 // one megabit in bytes

func toLinkSpeed(speed uint32) uint64 {
	switch speed {
	case 1:
		return 10 * megabit // 10M
	case 2:
		return 100 * megabit // 100M
	case 4:
		return 1000 * megabit // 1G
	case 8:
		return 10000 * megabit // 10G
	case 16:
		return 40000 * megabit // 40G
	case 32:
		return 100000 * megabit // 100G
	default:
		return 0
	}
}

// Returns true if given interface contains at least one IPv6 address. For VxLAN, source and destination
// addresses are also checked
func isIpv6Interface(iface *ifs.Interface) (bool, error) {
	if iface.Type == ifs.Interface_VXLAN_TUNNEL && iface.GetVxlan() != nil {
		if ipAddress := net.ParseIP(iface.GetVxlan().SrcAddress); ipAddress.To4() == nil {
			return true, nil
		}
		if ipAddress := net.ParseIP(iface.GetVxlan().DstAddress); ipAddress.To4() == nil {
			return true, nil
		}
	}
	for _, ifAddress := range iface.IpAddresses {
		if ipAddress, _, err := net.ParseCIDR(ifAddress); err != nil {
			return false, err
		} else if ipAddress.To4() == nil {
			return true, nil
		}
	}
	return false, nil
}

// dumpIPAddressDetails dumps IP address details of interfaces from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpIPAddressDetails(ifs map[uint32]*vppcalls.InterfaceDetails, isIPv6 bool, dhcpClients map[uint32]*vppcalls.Dhcp) error {
	// Dump IP addresses of each interface.
	for idx := range ifs {
		reqCtx := h.callsChannel.SendMultiRequest(&vpp_ip.IPAddressDump{
			SwIfIndex: interface_types.InterfaceIndex(idx),
			IsIPv6:    isIPv6,
		})
		for {
			ipDetails := &vpp_ip.IPAddressDetails{}
			stop, err := reqCtx.ReceiveReply(ipDetails)
			if stop {
				break // Break from the loop.
			}
			if err != nil {
				return fmt.Errorf("failed to dump interface %d IP address details: %v", idx, err)
			}
			h.processIPDetails(ifs, ipDetails, dhcpClients)
		}
	}

	return nil
}

// processIPDetails processes ip.IPAddressDetails binary API message and fills the details into the provided interface map.
func (h *InterfaceVppHandler) processIPDetails(ifs map[uint32]*vppcalls.InterfaceDetails, ipDetails *vpp_ip.IPAddressDetails, dhcpClients map[uint32]*vppcalls.Dhcp) {
	ifDetails, ifIdxExists := ifs[uint32(ipDetails.SwIfIndex)]
	if !ifIdxExists {
		return
	}

	var ipAddr string
	ipByte := make([]byte, 16)
	copy(ipByte[:], ipDetails.Prefix.Address.Un.XXX_UnionData[:])
	if ipDetails.Prefix.Address.Af == ip_types.ADDRESS_IP6 {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipByte).To16().String(), uint32(ipDetails.Prefix.Len))
	} else {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipByte[:4]).To4().String(), uint32(ipDetails.Prefix.Len))
	}

	// skip IP addresses given by DHCP
	if dhcpClient, hasDhcpClient := dhcpClients[uint32(ipDetails.SwIfIndex)]; hasDhcpClient {
		if dhcpClient.Lease != nil && dhcpClient.Lease.HostAddress == ipAddr {
			return
		}
	}

	ifDetails.Interface.IpAddresses = append(ifDetails.Interface.IpAddresses, ipAddr)
}

// dumpTapDetails dumps tap interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpTapDetails(interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	// Original TAP v1 was DEPRECATED

	// TAP v2
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_tapv2.SwInterfaceTapV2Dump{
		SwIfIndex: ^vpp_tapv2.InterfaceIndex(0),
	})
	for {
		tapDetails := &vpp_tapv2.SwInterfaceTapV2Details{}
		stop, err := reqCtx.ReceiveReply(tapDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump TAPv2 interface details: %v", err)
		}
		_, ifIdxExists := interfaces[tapDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		interfaces[tapDetails.SwIfIndex].Interface.Link = &ifs.Interface_Tap{
			Tap: &ifs.TapLink{
				Version:    2,
				HostIfName: cleanString(tapDetails.HostIfName),
				RxRingSize: uint32(tapDetails.RxRingSz),
				TxRingSize: uint32(tapDetails.TxRingSz),
				EnableGso:  tapDetails.TapFlags&vpp_tapv2.TAP_FLAG_GSO == vpp_tapv2.TAP_FLAG_GSO,
			},
		}
		interfaces[tapDetails.SwIfIndex].Interface.Type = ifs.Interface_TAP
	}

	return nil
}

// dumpVxlanDetails dumps VXLAN interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpVxlanDetails(interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_vxlan.VxlanTunnelDump{
		SwIfIndex: ^uint32(0),
	})
	for {
		vxlanDetails := &vpp_vxlan.VxlanTunnelDetails{}
		stop, err := reqCtx.ReceiveReply(vxlanDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump VxLAN tunnel interface details: %v", err)
		}
		_, ifIdxExists := interfaces[vxlanDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		// Multicast interface
		var multicastIfName string
		_, exists := interfaces[vxlanDetails.McastSwIfIndex]
		if exists {
			multicastIfName = interfaces[vxlanDetails.McastSwIfIndex].Interface.Name
		}

		if vxlanDetails.IsIPv6 == 1 {
			interfaces[vxlanDetails.SwIfIndex].Interface.Link = &ifs.Interface_Vxlan{
				Vxlan: &ifs.VxlanLink{
					Multicast:  multicastIfName,
					SrcAddress: net.IP(vxlanDetails.SrcAddress).To16().String(),
					DstAddress: net.IP(vxlanDetails.DstAddress).To16().String(),
					Vni:        vxlanDetails.Vni,
				},
			}
		} else {
			interfaces[vxlanDetails.SwIfIndex].Interface.Link = &ifs.Interface_Vxlan{
				Vxlan: &ifs.VxlanLink{
					Multicast:  multicastIfName,
					SrcAddress: net.IP(vxlanDetails.SrcAddress[:4]).To4().String(),
					DstAddress: net.IP(vxlanDetails.DstAddress[:4]).To4().String(),
					Vni:        vxlanDetails.Vni,
				},
			}
		}
		interfaces[vxlanDetails.SwIfIndex].Interface.Type = ifs.Interface_VXLAN_TUNNEL
	}

	return nil
}

// dumpVxlanDetails dumps VXLAN-GPE interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpVxLanGpeDetails(interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_vxlangpe.VxlanGpeTunnelDump{SwIfIndex: ^uint32(0)})
	for {
		vxlanGpeDetails := &vpp_vxlangpe.VxlanGpeTunnelDetails{}
		stop, err := reqCtx.ReceiveReply(vxlanGpeDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump VxLAN-GPE tunnel interface details: %v", err)
		}
		_, ifIdxExists := interfaces[vxlanGpeDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		// Multicast interface
		var multicastIfName string
		_, exists := interfaces[vxlanGpeDetails.McastSwIfIndex]
		if exists {
			multicastIfName = interfaces[vxlanGpeDetails.McastSwIfIndex].Interface.Name
		}

		vxLan := &ifs.VxlanLink{
			Multicast: multicastIfName,
			Vni:       vxlanGpeDetails.Vni,
			Gpe: &ifs.VxlanLink_Gpe{
				DecapVrfId: vxlanGpeDetails.DecapVrfID,
				Protocol:   getVxLanGpeProtocol(vxlanGpeDetails.Protocol),
			},
		}

		if vxlanGpeDetails.IsIPv6 == 1 {
			vxLan.SrcAddress = net.IP(vxlanGpeDetails.Local).To16().String()
			vxLan.DstAddress = net.IP(vxlanGpeDetails.Remote).To16().String()
		} else {
			vxLan.SrcAddress = net.IP(vxlanGpeDetails.Local[:4]).To4().String()
			vxLan.DstAddress = net.IP(vxlanGpeDetails.Remote[:4]).To4().String()
		}

		interfaces[vxlanGpeDetails.SwIfIndex].Interface.Link = &ifs.Interface_Vxlan{Vxlan: vxLan}
		interfaces[vxlanGpeDetails.SwIfIndex].Interface.Type = ifs.Interface_VXLAN_TUNNEL
	}

	return nil
}

// dumpIPSecTunnelDetails dumps IPSec tunnel interfaces from the VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpIPSecTunnelDetails(interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	// tunnel interfaces are a part of security association dump
	var tunnels []*vpp_ipsec.IpsecSaDetails
	req := &vpp_ipsec.IpsecSaDump{
		SaID: ^uint32(0),
	}
	requestCtx := h.callsChannel.SendMultiRequest(req)

	for {
		saDetails := &vpp_ipsec.IpsecSaDetails{}
		stop, err := requestCtx.ReceiveReply(saDetails)
		if stop {
			break
		}
		if err != nil {
			return err
		}
		// skip non-tunnel security associations
		if saDetails.SwIfIndex != ^uint32(0) {
			tunnels = append(tunnels, saDetails)
		}
	}

	// every tunnel interface is returned in two API calls. To reconstruct the correct proto-modelled data,
	// first appearance is cached, and when the second part arrives, data are completed and stored.
	tunnelParts := make(map[uint32]*vpp_ipsec.IpsecSaDetails)

	for _, tunnel := range tunnels {
		// first appearance is stored in the map, the second one is used in configuration.
		firstSaData, ok := tunnelParts[tunnel.SwIfIndex]
		if !ok {
			tunnelParts[tunnel.SwIfIndex] = tunnel
			continue
		}

		local := firstSaData
		remote := tunnel

		// verify data for local & remote
		if err := verifyIPSecTunnelDetails(local, remote); err != nil {
			h.log.Warnf("IPSec SA dump for tunnel interface data does not match: %v", err)
			continue
		}

		var localIP, remoteIP net.IP
		if tunnel.Entry.TunnelDst.Af == ip_types.ADDRESS_IP6 {
			localSrc := local.Entry.TunnelSrc.Un.GetIP6()
			remoteSrc := remote.Entry.TunnelSrc.Un.GetIP6()
			localIP, remoteIP = net.IP(localSrc[:]), net.IP(remoteSrc[:])
		} else {
			localSrc := local.Entry.TunnelSrc.Un.GetIP4()
			remoteSrc := remote.Entry.TunnelSrc.Un.GetIP4()
			localIP, remoteIP = net.IP(localSrc[:]), net.IP(remoteSrc[:])
		}

		ifDetails, ok := interfaces[tunnel.SwIfIndex]
		if !ok {
			h.log.Warnf("ipsec SA dump returned unrecognized swIfIndex: %v", tunnel.SwIfIndex)
			continue
		}
		ifDetails.Interface.Type = ifs.Interface_IPSEC_TUNNEL
		ifDetails.Interface.Link = &ifs.Interface_Ipsec{
			Ipsec: &ifs.IPSecLink{
				Esn:             (tunnel.Entry.Flags & ipsec_types.IPSEC_API_SAD_FLAG_USE_ESN) != 0,
				AntiReplay:      (tunnel.Entry.Flags & ipsec_types.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY) != 0,
				LocalIp:         localIP.String(),
				RemoteIp:        remoteIP.String(),
				LocalSpi:        local.Entry.Spi,
				RemoteSpi:       remote.Entry.Spi,
				CryptoAlg:       ipsec.CryptoAlg(tunnel.Entry.CryptoAlgorithm),
				LocalCryptoKey:  hex.EncodeToString(local.Entry.CryptoKey.Data[:local.Entry.CryptoKey.Length]),
				RemoteCryptoKey: hex.EncodeToString(remote.Entry.CryptoKey.Data[:remote.Entry.CryptoKey.Length]),
				IntegAlg:        ipsec.IntegAlg(tunnel.Entry.IntegrityAlgorithm),
				LocalIntegKey:   hex.EncodeToString(local.Entry.IntegrityKey.Data[:local.Entry.IntegrityKey.Length]),
				RemoteIntegKey:  hex.EncodeToString(remote.Entry.IntegrityKey.Data[:remote.Entry.IntegrityKey.Length]),
				EnableUdpEncap:  (tunnel.Entry.Flags & ipsec_types.IPSEC_API_SAD_FLAG_UDP_ENCAP) != 0,
			},
		}
	}

	return nil
}

func verifyIPSecTunnelDetails(local, remote *vpp_ipsec.IpsecSaDetails) error {
	if local.SwIfIndex != remote.SwIfIndex {
		return fmt.Errorf("swIfIndex data mismatch (local: %v, remote: %v)",
			local.SwIfIndex, remote.SwIfIndex)
	}
	localIsTunnel := local.Entry.Flags & ipsec_types.IPSEC_API_SAD_FLAG_IS_TUNNEL
	remoteIsTunnel := remote.Entry.Flags & ipsec_types.IPSEC_API_SAD_FLAG_IS_TUNNEL
	if localIsTunnel != remoteIsTunnel {
		return fmt.Errorf("tunnel data mismatch (local: %v, remote: %v)",
			localIsTunnel, remoteIsTunnel)
	}

	localSrc, localDst := local.Entry.TunnelSrc.Un.XXX_UnionData, local.Entry.TunnelDst.Un.XXX_UnionData
	remoteSrc, remoteDst := remote.Entry.TunnelSrc.Un.XXX_UnionData, remote.Entry.TunnelDst.Un.XXX_UnionData
	if (local.Entry.Flags&ipsec_types.IPSEC_API_SAD_FLAG_IS_TUNNEL_V6) != (remote.Entry.Flags&ipsec_types.IPSEC_API_SAD_FLAG_IS_TUNNEL_V6) ||
		!bytes.Equal(localSrc[:], remoteDst[:]) ||
		!bytes.Equal(localDst[:], remoteSrc[:]) {
		return fmt.Errorf("src/dst IP mismatch (local: %+v, remote: %+v)",
			local.Entry, remote.Entry)
	}

	return nil
}

// dumpBondDetails dumps bond interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpBondDetails(interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	bondIndexes := make([]uint32, 0)
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_bond.SwInterfaceBondDump{})
	for {
		bondDetails := &vpp_bond.SwInterfaceBondDetails{}
		stop, err := reqCtx.ReceiveReply(bondDetails)
		if err != nil {
			return fmt.Errorf("failed to dump bond interface details: %v", err)
		}
		if stop {
			break
		}
		_, ifIdxExists := interfaces[uint32(bondDetails.SwIfIndex)]
		if !ifIdxExists {
			continue
		}
		interfaces[uint32(bondDetails.SwIfIndex)].Interface.Link = &ifs.Interface_Bond{
			Bond: &ifs.BondLink{
				Id:   bondDetails.ID,
				Mode: getBondIfMode(bondDetails.Mode),
				Lb:   getBondLoadBalance(bondDetails.Lb),
			},
		}
		interfaces[uint32(bondDetails.SwIfIndex)].Interface.Type = ifs.Interface_BOND_INTERFACE
		bondIndexes = append(bondIndexes, uint32(bondDetails.SwIfIndex))
	}

	// get slave interfaces for bonds
	for _, bondIdx := range bondIndexes {
		var bondSlaves []*ifs.BondLink_BondedInterface
		reqSlCtx := h.callsChannel.SendMultiRequest(&vpp_bond.SwInterfaceSlaveDump{
			SwIfIndex: vpp_bond.InterfaceIndex(bondIdx),
		})
		for {
			slaveDetails := &vpp_bond.SwInterfaceSlaveDetails{}
			stop, err := reqSlCtx.ReceiveReply(slaveDetails)
			if err != nil {
				return fmt.Errorf("failed to dump bond slave details: %v", err)
			}
			if stop {
				break
			}
			slaveIf, ifIdxExists := interfaces[uint32(slaveDetails.SwIfIndex)]
			if !ifIdxExists {
				continue
			}
			bondSlaves = append(bondSlaves, &ifs.BondLink_BondedInterface{
				Name:          slaveIf.Interface.Name,
				IsPassive:     slaveDetails.IsPassive,
				IsLongTimeout: slaveDetails.IsLongTimeout,
			})
			interfaces[bondIdx].Interface.GetBond().BondedInterfaces = bondSlaves
		}
	}

	return nil
}

func (h *InterfaceVppHandler) dumpGreDetails(interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	msg := &vpp_gre.GreTunnelDump{SwIfIndex: vpp_gre.InterfaceIndex(^uint32(0))}
	reqCtx := h.callsChannel.SendMultiRequest(msg)
	for {
		greDetails := &vpp_gre.GreTunnelDetails{}
		stop, err := reqCtx.ReceiveReply(greDetails)
		if stop {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to dump span: %v", err)
		}

		tunnel := greDetails.Tunnel
		swIfIndex := uint32(tunnel.SwIfIndex)

		var srcAddr, dstAddr net.IP
		if tunnel.Src.Af == ip_types.ADDRESS_IP4 {
			srcAddrArr := tunnel.Src.Un.GetIP4()
			srcAddr = net.IP(srcAddrArr[:])
		} else {
			srcAddrArr := tunnel.Src.Un.GetIP6()
			srcAddr = net.IP(srcAddrArr[:])
		}
		if tunnel.Dst.Af == ip_types.ADDRESS_IP4 {
			dstAddrArr := tunnel.Dst.Un.GetIP4()
			dstAddr = net.IP(dstAddrArr[:])
		} else {
			dstAddrArr := tunnel.Dst.Un.GetIP6()
			dstAddr = net.IP(dstAddrArr[:])
		}

		interfaces[swIfIndex].Interface.Link = &ifs.Interface_Gre{
			Gre: &ifs.GreLink{
				TunnelType: getGreTunnelType(tunnel.Type),
				SrcAddr:    srcAddr.String(),
				DstAddr:    dstAddr.String(),
				OuterFibId: tunnel.OuterTableID,
				SessionId:  uint32(tunnel.SessionID),
			},
		}
		interfaces[swIfIndex].Interface.Type = ifs.Interface_GRE_TUNNEL
	}
	return nil
}

// dumpUnnumberedDetails returns a map of unnumbered interface indexes, every with interface index of element with IP
func (h *InterfaceVppHandler) dumpUnnumberedDetails() (map[uint32]uint32, error) {
	unIfMap := make(map[uint32]uint32) // unnumbered/ip-interface
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_ip.IPUnnumberedDump{
		SwIfIndex: ^interface_types.InterfaceIndex(0),
	})

	for {
		unDetails := &vpp_ip.IPUnnumberedDetails{}
		last, err := reqCtx.ReceiveReply(unDetails)
		if last {
			break
		}
		if err != nil {
			return nil, err
		}

		unIfMap[uint32(unDetails.SwIfIndex)] = uint32(unDetails.IPSwIfIndex)
	}

	return unIfMap, nil
}

func (h *InterfaceVppHandler) dumpRxPlacement(interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_ifs.SwInterfaceRxPlacementDump{
		SwIfIndex: vpp_ifs.InterfaceIndex(^uint32(0)),
	})
	for {
		rxDetails := &vpp_ifs.SwInterfaceRxPlacementDetails{}
		stop, err := reqCtx.ReceiveReply(rxDetails)
		if err != nil {
			return fmt.Errorf("failed to dump rx-placement details: %v", err)
		}
		if stop {
			break
		}

		ifData, ok := interfaces[uint32(rxDetails.SwIfIndex)]
		if !ok {
			h.log.Warnf("Received rx-placement data for unknown interface with index %d", rxDetails.SwIfIndex)
			continue
		}

		ifData.Interface.RxModes = append(ifData.Interface.RxModes,
			&ifs.Interface_RxMode{
				Queue: rxDetails.QueueID,
				Mode:  getRxModeType(rxDetails.Mode),
			})

		var worker uint32
		if rxDetails.WorkerID > 0 {
			worker = rxDetails.WorkerID - 1
		}
		ifData.Interface.RxPlacements = append(ifData.Interface.RxPlacements,
			&ifs.Interface_RxPlacement{
				Queue:      rxDetails.QueueID,
				Worker:     worker,
				MainThread: rxDetails.WorkerID == 0,
			})
	}
	return nil
}

func dhcpAddressToString(address vpp_dhcp.Address, maskWidth uint32, isIPv6 bool) string {
	dhcpIPByte := make([]byte, 16)
	copy(dhcpIPByte[:], address.Un.XXX_UnionData[:])
	if isIPv6 {
		return fmt.Sprintf("%s/%d", net.IP(dhcpIPByte).To16().String(), maskWidth)
	}
	return fmt.Sprintf("%s/%d", net.IP(dhcpIPByte[:4]).To4().String(), maskWidth)
}

// guessInterfaceType attempts to guess the correct interface type from its internal name (as given by VPP).
// This is required mainly for those interface types, that do not provide dump binary API,
// such as loopback of af_packet.
func guessInterfaceType(ifName string) ifs.Interface_Type {
	switch {
	case strings.HasPrefix(ifName, "loop"),
		strings.HasPrefix(ifName, "local"):
		return ifs.Interface_SOFTWARE_LOOPBACK

	case strings.HasPrefix(ifName, "memif"):
		return ifs.Interface_MEMIF

	case strings.HasPrefix(ifName, "tap"):
		return ifs.Interface_TAP

	case strings.HasPrefix(ifName, "host"):
		return ifs.Interface_AF_PACKET

	case strings.HasPrefix(ifName, "vxlan"):
		return ifs.Interface_VXLAN_TUNNEL

	case strings.HasPrefix(ifName, "ipsec"):
		return ifs.Interface_IPSEC_TUNNEL

	case strings.HasPrefix(ifName, "vmxnet3"):
		return ifs.Interface_VMXNET3_INTERFACE

	case strings.HasPrefix(ifName, "Bond"):
		return ifs.Interface_BOND_INTERFACE

	case strings.HasPrefix(ifName, "gtpu"):
		return ifs.Interface_GTPU_TUNNEL

	default:
		return ifs.Interface_DPDK
	}
}

// memifModetoNB converts binary API type of memif mode to the northbound API type memif mode.
func memifModetoNB(mode vpp_memif.MemifMode) ifs.MemifLink_MemifMode {
	switch mode {
	case vpp_memif.MEMIF_MODE_API_IP:
		return ifs.MemifLink_IP
	case vpp_memif.MEMIF_MODE_API_PUNT_INJECT:
		return ifs.MemifLink_PUNT_INJECT
	default:
		return ifs.MemifLink_ETHERNET
	}
}

// Convert binary API rx-mode to northbound representation
func getRxModeType(mode vpp_ifs.RxMode) ifs.Interface_RxMode_Type {
	switch mode {
	case 1:
		return ifs.Interface_RxMode_POLLING
	case 2:
		return ifs.Interface_RxMode_INTERRUPT
	case 3:
		return ifs.Interface_RxMode_ADAPTIVE
	case 4:
		return ifs.Interface_RxMode_DEFAULT
	default:
		return ifs.Interface_RxMode_UNKNOWN
	}
}

func getBondIfMode(mode vpp_bond.BondMode) ifs.BondLink_Mode {
	switch mode {
	case vpp_bond.BOND_API_MODE_ROUND_ROBIN:
		return ifs.BondLink_ROUND_ROBIN
	case vpp_bond.BOND_API_MODE_ACTIVE_BACKUP:
		return ifs.BondLink_ACTIVE_BACKUP
	case vpp_bond.BOND_API_MODE_XOR:
		return ifs.BondLink_XOR
	case vpp_bond.BOND_API_MODE_BROADCAST:
		return ifs.BondLink_BROADCAST
	case vpp_bond.BOND_API_MODE_LACP:
		return ifs.BondLink_LACP
	default:
		// UNKNOWN
		return 0
	}
}

func getBondLoadBalance(lb vpp_bond.BondLbAlgo) ifs.BondLink_LoadBalance {
	switch lb {
	case vpp_bond.BOND_API_LB_ALGO_L34:
		return ifs.BondLink_L34
	case vpp_bond.BOND_API_LB_ALGO_L23:
		return ifs.BondLink_L23
	case vpp_bond.BOND_API_LB_ALGO_RR:
		return ifs.BondLink_RR
	case vpp_bond.BOND_API_LB_ALGO_BC:
		return ifs.BondLink_BC
	case vpp_bond.BOND_API_LB_ALGO_AB:
		return ifs.BondLink_AB
	default:
		return ifs.BondLink_L2
	}
}

func getTagRwOption(op uint32) ifs.SubInterface_TagRewriteOptions {
	switch op {
	case 1:
		return ifs.SubInterface_PUSH1
	case 2:
		return ifs.SubInterface_PUSH2
	case 3:
		return ifs.SubInterface_POP1
	case 4:
		return ifs.SubInterface_POP2
	case 5:
		return ifs.SubInterface_TRANSLATE11
	case 6:
		return ifs.SubInterface_TRANSLATE12
	case 7:
		return ifs.SubInterface_TRANSLATE21
	case 8:
		return ifs.SubInterface_TRANSLATE22
	default: // disabled
		return ifs.SubInterface_DISABLED
	}
}

func getGreTunnelType(tt vpp_gre.GreTunnelType) ifs.GreLink_Type {
	switch tt {
	case vpp_gre.GRE_API_TUNNEL_TYPE_L3:
		return ifs.GreLink_L3
	case vpp_gre.GRE_API_TUNNEL_TYPE_TEB:
		return ifs.GreLink_TEB
	case vpp_gre.GRE_API_TUNNEL_TYPE_ERSPAN:
		return ifs.GreLink_ERSPAN
	default:
		return ifs.GreLink_UNKNOWN
	}
}

func getVxLanGpeProtocol(p uint8) ifs.VxlanLink_Gpe_Protocol {
	switch p {
	case 1:
		return ifs.VxlanLink_Gpe_IP4
	case 2:
		return ifs.VxlanLink_Gpe_IP6
	case 3:
		return ifs.VxlanLink_Gpe_ETHERNET
	case 4:
		return ifs.VxlanLink_Gpe_NSH
	default:
		return ifs.VxlanLink_Gpe_UNKNOWN
	}
}

func isAdminStateUp(flags vpp_ifs.IfStatusFlags) bool {
	return flags&interface_types.IF_STATUS_API_FLAG_ADMIN_UP != 0
}

func isLinkStateUp(flags vpp_ifs.IfStatusFlags) bool {
	return flags&interface_types.IF_STATUS_API_FLAG_LINK_UP != 0
}

func adminStateToInterfaceStatus(flags vpp_ifs.IfStatusFlags) ifs.InterfaceState_Status {
	if isAdminStateUp(flags) {
		return ifs.InterfaceState_UP
	}
	return ifs.InterfaceState_DOWN
}

func linkStateToInterfaceStatus(flags vpp_ifs.IfStatusFlags) ifs.InterfaceState_Status {
	if isLinkStateUp(flags) {
		return ifs.InterfaceState_UP
	}
	return ifs.InterfaceState_DOWN
}
