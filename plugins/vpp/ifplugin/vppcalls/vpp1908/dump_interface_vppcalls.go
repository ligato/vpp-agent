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
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/bond"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/gre"
	binapi_interface "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/tapv2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vxlan"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vxlan_gpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

const (
	// allInterfaces defines unspecified interface index
	allInterfaces = ^uint32(0)

	// prefix prepended to internal names of untagged interfaces to construct unique
	// logical names
	untaggedIfPreffix = "UNTAGGED-"
)

// Default VPP MTU value
const defaultVPPMtu = 9216

func getMtu(vppMtu uint16) uint32 {
	// If default VPP MTU value is set, return 0 (it means MTU was not set in the NB config)
	if vppMtu == defaultVPPMtu {
		return 0
	}
	return uint32(vppMtu)
}

// DumpInterfacesByType implements interface handler.
func (h *InterfaceVppHandler) DumpInterfacesByType(ctx context.Context, reqType interfaces.Interface_Type) (map[uint32]*vppcalls.InterfaceDetails, error) {
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
	ifs := make(map[uint32]*vppcalls.InterfaceDetails)

	ifIdx := allInterfaces
	if len(ifIdxs) > 0 {
		ifIdx = ifIdxs[0]
	}
	// First, dump all interfaces to create initial data.
	reqCtx := h.callsChannel.SendMultiRequest(&binapi_interface.SwInterfaceDump{
		SwIfIndex: binapi_interface.InterfaceIndex(ifIdx),
	})
	for {
		ifDetails := &binapi_interface.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(ifDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dump interface: %v", err)
		}

		ifaceName := strings.TrimRight(ifDetails.InterfaceName, "\x00")
		l2addr := net.HardwareAddr(ifDetails.L2Address[:ifDetails.L2AddressLength])

		details := &vppcalls.InterfaceDetails{
			Interface: &interfaces.Interface{
				Name: strings.TrimRight(ifDetails.Tag, "\x00"),
				// the type may be amended later by further dumps
				Type:        guessInterfaceType(ifaceName),
				Enabled:     ifDetails.AdminUpDown > 0,
				PhysAddress: net.HardwareAddr(ifDetails.L2Address[:ifDetails.L2AddressLength]).String(),
				Mtu:         getMtu(ifDetails.LinkMtu),
			},
			Meta: &vppcalls.InterfaceMeta{
				SwIfIndex:      ifDetails.SwIfIndex,
				SupSwIfIndex:   ifDetails.SupSwIfIndex,
				L2Address:      l2addr,
				InternalName:   ifaceName,
				IsAdminStateUp: uintToBool(ifDetails.AdminUpDown),
				IsLinkStateUp:  uintToBool(ifDetails.LinkUpDown),
				LinkDuplex:     uint32(ifDetails.LinkDuplex),
				LinkMTU:        ifDetails.LinkMtu,
				MTU:            ifDetails.Mtu,
				LinkSpeed:      ifDetails.LinkSpeed,
				SubID:          ifDetails.SubID,
				Tag:            strings.TrimRight(ifDetails.Tag, "\x00"),
			},
		}

		// sub interface
		if ifDetails.SupSwIfIndex != ifDetails.SwIfIndex {
			details.Interface.Type = interfaces.Interface_SUB_INTERFACE
			details.Interface.Link = &interfaces.Interface_Sub{
				Sub: &interfaces.SubInterface{
					ParentName:  ifs[ifDetails.SupSwIfIndex].Interface.Name,
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
		case interfaces.Interface_DPDK:
			details.Interface.Name = ifaceName
		case interfaces.Interface_AF_PACKET:
			details.Interface.Link = &interfaces.Interface_Afpacket{
				Afpacket: &interfaces.AfpacketLink{
					HostIfName: strings.TrimPrefix(ifaceName, "host-"),
				},
			}
		}
		if details.Interface.Name == "" {
			// untagged interface - generate a logical name for it
			// (apart from local0 it will get removed by resync)
			details.Interface.Name = untaggedIfPreffix + ifaceName
		}
		ifs[ifDetails.SwIfIndex] = details
	}

	return ifs, nil
}

// DumpInterfaces implements interface handler.
func (h *InterfaceVppHandler) DumpInterfaces(ctx context.Context) (map[uint32]*vppcalls.InterfaceDetails, error) {
	ifs, err := h.dumpInterfaces()
	if err != nil {
		return nil, err
	}

	// Get DHCP clients
	dhcpClients, err := h.DumpDhcpClients()
	if err != nil {
		return nil, fmt.Errorf("failed to dump interface DHCP clients: %v", err)
	}

	// Get IP addresses before VRF
	err = h.dumpIPAddressDetails(ifs, false, dhcpClients)
	if err != nil {
		return nil, err
	}
	err = h.dumpIPAddressDetails(ifs, true, dhcpClients)
	if err != nil {
		return nil, err
	}

	// Get unnumbered interfaces
	unnumbered, err := h.dumpUnnumberedDetails()
	if err != nil {
		return nil, fmt.Errorf("failed to dump unnumbered interfaces: %v", err)
	}

	// dump VXLAN details before VRFs (used by isIpv6Interface)
	err = h.dumpVxlanDetails(ifs)
	if err != nil {
		return nil, err
	}
	err = h.dumpVxLanGpeDetails(ifs)
	if err != nil {
		return nil, err
	}

	// Get interface VRF for every IP family, fill DHCP if set and resolve unnumbered interface setup
	for _, ifData := range ifs {
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
			return ifs, err
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
			ifWithIP, ok := ifs[ifWithIPIdx]
			if ok {
				ifWithIPName = ifWithIP.Interface.Name
			} else {
				h.log.Debugf("cannot find name of the ip-interface for unnumbered %s", ifData.Interface.Name)
				ifWithIPName = "<unknown>"
			}
			ifData.Interface.Unnumbered = &interfaces.Interface_Unnumbered{
				InterfaceWithIp: ifWithIPName,
			}
		}
	}

	err = h.dumpMemifDetails(ctx, ifs)
	if err != nil {
		return nil, err
	}

	err = h.dumpTapDetails(ifs)
	if err != nil {
		return nil, err
	}

	err = h.dumpIPSecTunnelDetails(ifs)
	if err != nil {
		return nil, err
	}

	err = h.dumpVmxNet3Details(ifs)
	if err != nil {
		return nil, err
	}

	err = h.dumpBondDetails(ifs)
	if err != nil {
		return nil, err
	}

	err = h.dumpGreDetails(ifs)
	if err != nil {
		return nil, err
	}

	err = h.dumpGtpuDetails(ifs)
	if err != nil {
		return nil, err
	}

	err = h.dumpIpipDetails(ifs)
	if err != nil {
		return nil, err
	}

	// Rx-placement dump is last since it uses interface type-specific data
	err = h.dumpRxPlacement(ifs)
	if err != nil {
		return nil, err
	}

	return ifs, nil
}

// DumpDhcpClients returns a slice of DhcpMeta with all interfaces and other DHCP-related information available
func (h *InterfaceVppHandler) DumpDhcpClients() (map[uint32]*vppcalls.Dhcp, error) {
	dhcpData := make(map[uint32]*vppcalls.Dhcp)
	reqCtx := h.callsChannel.SendMultiRequest(&dhcp.DHCPClientDump{})

	for {
		dhcpDetails := &dhcp.DHCPClientDetails{}
		last, err := reqCtx.ReceiveReply(dhcpDetails)
		if last {
			break
		}
		if err != nil {
			return nil, err
		}
		client := dhcpDetails.Client
		lease := dhcpDetails.Lease

		var hostMac net.HardwareAddr = lease.HostMac
		var hostAddr, routerAddr string
		if uintToBool(lease.IsIPv6) {
			hostAddr = fmt.Sprintf("%s/%d", net.IP(lease.HostAddress).To16().String(), uint32(lease.MaskWidth))
			routerAddr = fmt.Sprintf("%s/%d", net.IP(lease.RouterAddress).To16().String(), uint32(lease.MaskWidth))
		} else {
			hostAddr = fmt.Sprintf("%s/%d", net.IP(lease.HostAddress[:4]).To4().String(), uint32(lease.MaskWidth))
			routerAddr = fmt.Sprintf("%s/%d", net.IP(lease.RouterAddress[:4]).To4().String(), uint32(lease.MaskWidth))
		}

		// DHCP client data
		dhcpClient := &vppcalls.Client{
			SwIfIndex:        client.SwIfIndex,
			Hostname:         string(bytes.SplitN(client.Hostname, []byte{0x00}, 2)[0]),
			ID:               string(bytes.SplitN(client.ID, []byte{0x00}, 2)[0]),
			WantDhcpEvent:    uintToBool(client.WantDHCPEvent),
			SetBroadcastFlag: uintToBool(client.SetBroadcastFlag),
			PID:              client.PID,
		}

		// DHCP lease data
		dhcpLease := &vppcalls.Lease{
			SwIfIndex:     lease.SwIfIndex,
			State:         lease.State,
			Hostname:      string(bytes.SplitN(lease.Hostname, []byte{0x00}, 2)[0]),
			IsIPv6:        uintToBool(lease.IsIPv6),
			HostAddress:   hostAddr,
			RouterAddress: routerAddr,
			HostMac:       hostMac.String(),
		}

		// DHCP metadata
		dhcpData[client.SwIfIndex] = &vppcalls.Dhcp{
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

	ifs := make(map[uint32]*vppcalls.InterfaceState)
	for _, ifIdx := range ifIdxs {
		reqCtx := h.callsChannel.SendMultiRequest(&binapi_interface.SwInterfaceDump{
			SwIfIndex: binapi_interface.InterfaceIndex(ifIdx),
		})
		for {
			ifDetails := &binapi_interface.SwInterfaceDetails{}
			stop, err := reqCtx.ReceiveReply(ifDetails)
			if stop {
				break // Break from the loop.
			}
			if err != nil {
				return nil, fmt.Errorf("failed to dump interface states: %v", err)
			}

			physAddr := make(net.HardwareAddr, ifDetails.L2AddressLength)
			copy(physAddr, ifDetails.L2Address[:])

			ifaceState := vppcalls.InterfaceState{
				SwIfIndex:    ifDetails.SwIfIndex,
				InternalName: strings.TrimRight(ifDetails.InterfaceName, "\x00"),
				PhysAddress:  physAddr,
				AdminState:   toInterfaceStatus(ifDetails.AdminUpDown),
				LinkState:    toInterfaceStatus(ifDetails.LinkUpDown),
				LinkDuplex:   toLinkDuplex(ifDetails.LinkDuplex),
				LinkSpeed:    toLinkSpeed(ifDetails.LinkSpeed),
				LinkMTU:      ifDetails.LinkMtu,
			}
			ifs[ifDetails.SwIfIndex] = &ifaceState
		}
	}

	return ifs, nil
}

func toInterfaceStatus(upDown uint8) interfaces.InterfaceState_Status {
	switch upDown {
	case 0:
		return interfaces.InterfaceState_DOWN
	case 1:
		return interfaces.InterfaceState_UP
	default:
		return interfaces.InterfaceState_UNKNOWN_STATUS
	}
}

func toLinkDuplex(duplex uint8) interfaces.InterfaceState_Duplex {
	switch duplex {
	case 1:
		return interfaces.InterfaceState_HALF
	case 2:
		return interfaces.InterfaceState_FULL
	default:
		return interfaces.InterfaceState_UNKNOWN_DUPLEX
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
func isIpv6Interface(iface *interfaces.Interface) (bool, error) {
	if iface.Type == interfaces.Interface_VXLAN_TUNNEL && iface.GetVxlan() != nil {
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
		reqCtx := h.callsChannel.SendMultiRequest(&ip.IPAddressDump{
			SwIfIndex: idx,
			IsIPv6:    boolToUint(isIPv6),
		})
		for {
			ipDetails := &ip.IPAddressDetails{}
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
func (h *InterfaceVppHandler) processIPDetails(ifs map[uint32]*vppcalls.InterfaceDetails, ipDetails *ip.IPAddressDetails, dhcpClients map[uint32]*vppcalls.Dhcp) {
	ifDetails, ifIdxExists := ifs[ipDetails.SwIfIndex]
	if !ifIdxExists {
		return
	}

	var ipAddr string
	ipByte := make([]byte, 16)
	copy(ipByte[:], ipDetails.Prefix.Address.Un.XXX_UnionData[:])
	if ipDetails.Prefix.Address.Af == ip.ADDRESS_IP6 {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipByte).To16().String(), uint32(ipDetails.Prefix.Len))
	} else {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipByte[:4]).To4().String(), uint32(ipDetails.Prefix.Len))
	}

	// skip IP addresses given by DHCP
	if dhcpClient, hasDhcpClient := dhcpClients[ipDetails.SwIfIndex]; hasDhcpClient {
		if dhcpClient.Lease != nil && dhcpClient.Lease.HostAddress == ipAddr {
			return
		}
	}

	ifDetails.Interface.IpAddresses = append(ifDetails.Interface.IpAddresses, ipAddr)
}

// dumpTapDetails dumps tap interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpTapDetails(ifs map[uint32]*vppcalls.InterfaceDetails) error {
	// Original TAP v1 was DEPRECATED

	// TAP v2
	reqCtx := h.callsChannel.SendMultiRequest(&tapv2.SwInterfaceTapV2Dump{})
	for {
		tapDetails := &tapv2.SwInterfaceTapV2Details{}
		stop, err := reqCtx.ReceiveReply(tapDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump TAPv2 interface details: %v", err)
		}
		_, ifIdxExists := ifs[tapDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		ifs[tapDetails.SwIfIndex].Interface.Link = &interfaces.Interface_Tap{
			Tap: &interfaces.TapLink{
				Version:    2,
				HostIfName: cleanString(tapDetails.HostIfName),
				RxRingSize: uint32(tapDetails.RxRingSz),
				TxRingSize: uint32(tapDetails.TxRingSz),
				EnableGso:  tapDetails.TapFlags&TapFlagGSO == TapFlagGSO,
			},
		}
		ifs[tapDetails.SwIfIndex].Interface.Type = interfaces.Interface_TAP
	}

	return nil
}

// dumpVxlanDetails dumps VXLAN interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpVxlanDetails(ifs map[uint32]*vppcalls.InterfaceDetails) error {
	reqCtx := h.callsChannel.SendMultiRequest(&vxlan.VxlanTunnelDump{
		SwIfIndex: ^uint32(0),
	})
	for {
		vxlanDetails := &vxlan.VxlanTunnelDetails{}
		stop, err := reqCtx.ReceiveReply(vxlanDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump VxLAN tunnel interface details: %v", err)
		}
		_, ifIdxExists := ifs[vxlanDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		// Multicast interface
		var multicastIfName string
		_, exists := ifs[vxlanDetails.McastSwIfIndex]
		if exists {
			multicastIfName = ifs[vxlanDetails.McastSwIfIndex].Interface.Name
		}

		if vxlanDetails.IsIPv6 == 1 {
			ifs[vxlanDetails.SwIfIndex].Interface.Link = &interfaces.Interface_Vxlan{
				Vxlan: &interfaces.VxlanLink{
					Multicast:  multicastIfName,
					SrcAddress: net.IP(vxlanDetails.SrcAddress).To16().String(),
					DstAddress: net.IP(vxlanDetails.DstAddress).To16().String(),
					Vni:        vxlanDetails.Vni,
				},
			}
		} else {
			ifs[vxlanDetails.SwIfIndex].Interface.Link = &interfaces.Interface_Vxlan{
				Vxlan: &interfaces.VxlanLink{
					Multicast:  multicastIfName,
					SrcAddress: net.IP(vxlanDetails.SrcAddress[:4]).To4().String(),
					DstAddress: net.IP(vxlanDetails.DstAddress[:4]).To4().String(),
					Vni:        vxlanDetails.Vni,
				},
			}
		}
		ifs[vxlanDetails.SwIfIndex].Interface.Type = interfaces.Interface_VXLAN_TUNNEL
	}

	return nil
}

// dumpVxlanDetails dumps VXLAN-GPE interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpVxLanGpeDetails(ifs map[uint32]*vppcalls.InterfaceDetails) error {
	reqCtx := h.callsChannel.SendMultiRequest(&vxlan_gpe.VxlanGpeTunnelDump{SwIfIndex: ^uint32(0)})
	for {
		vxlanGpeDetails := &vxlan_gpe.VxlanGpeTunnelDetails{}
		stop, err := reqCtx.ReceiveReply(vxlanGpeDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump VxLAN-GPE tunnel interface details: %v", err)
		}
		_, ifIdxExists := ifs[vxlanGpeDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		// Multicast interface
		var multicastIfName string
		_, exists := ifs[vxlanGpeDetails.McastSwIfIndex]
		if exists {
			multicastIfName = ifs[vxlanGpeDetails.McastSwIfIndex].Interface.Name
		}

		vxLan := &interfaces.VxlanLink{
			Multicast: multicastIfName,
			Vni:       vxlanGpeDetails.Vni,
			Gpe: &interfaces.VxlanLink_Gpe{
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

		ifs[vxlanGpeDetails.SwIfIndex].Interface.Link = &interfaces.Interface_Vxlan{Vxlan: vxLan}
		ifs[vxlanGpeDetails.SwIfIndex].Interface.Type = interfaces.Interface_VXLAN_TUNNEL
	}

	return nil
}

// dumpIPSecTunnelDetails dumps IPSec tunnel interfaces from the VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpIPSecTunnelDetails(ifs map[uint32]*vppcalls.InterfaceDetails) error {
	// tunnel interfaces are a part of security association dump
	var tunnels []*ipsec.IpsecSaDetails
	req := &ipsec.IpsecSaDump{
		SaID: ^uint32(0),
	}
	requestCtx := h.callsChannel.SendMultiRequest(req)

	for {
		saDetails := &ipsec.IpsecSaDetails{}
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
	tunnelParts := make(map[uint32]*ipsec.IpsecSaDetails)

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
		if tunnel.Entry.TunnelDst.Af == ipsec.ADDRESS_IP6 {
			localSrc := local.Entry.TunnelSrc.Un.GetIP6()
			remoteSrc := remote.Entry.TunnelSrc.Un.GetIP6()
			localIP, remoteIP = net.IP(localSrc[:]), net.IP(remoteSrc[:])
		} else {
			localSrc := local.Entry.TunnelSrc.Un.GetIP4()
			remoteSrc := remote.Entry.TunnelSrc.Un.GetIP4()
			localIP, remoteIP = net.IP(localSrc[:]), net.IP(remoteSrc[:])
		}

		ifDetails, ok := ifs[tunnel.SwIfIndex]
		if !ok {
			h.log.Warnf("ipsec SA dump returned unrecognized swIfIndex: %v", tunnel.SwIfIndex)
			continue
		}
		ifDetails.Interface.Type = interfaces.Interface_IPSEC_TUNNEL
		ifDetails.Interface.Link = &interfaces.Interface_Ipsec{
			Ipsec: &interfaces.IPSecLink{
				Esn:             (tunnel.Entry.Flags & ipsec.IPSEC_API_SAD_FLAG_USE_ESN) != 0,
				AntiReplay:      (tunnel.Entry.Flags & ipsec.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY) != 0,
				LocalIp:         localIP.String(),
				RemoteIp:        remoteIP.String(),
				LocalSpi:        local.Entry.Spi,
				RemoteSpi:       remote.Entry.Spi,
				CryptoAlg:       vpp_ipsec.CryptoAlg(tunnel.Entry.CryptoAlgorithm),
				LocalCryptoKey:  hex.EncodeToString(local.Entry.CryptoKey.Data[:local.Entry.CryptoKey.Length]),
				RemoteCryptoKey: hex.EncodeToString(remote.Entry.CryptoKey.Data[:remote.Entry.CryptoKey.Length]),
				IntegAlg:        vpp_ipsec.IntegAlg(tunnel.Entry.IntegrityAlgorithm),
				LocalIntegKey:   hex.EncodeToString(local.Entry.IntegrityKey.Data[:local.Entry.IntegrityKey.Length]),
				RemoteIntegKey:  hex.EncodeToString(remote.Entry.IntegrityKey.Data[:remote.Entry.IntegrityKey.Length]),
				EnableUdpEncap:  (tunnel.Entry.Flags & ipsec.IPSEC_API_SAD_FLAG_UDP_ENCAP) != 0,
			},
		}
	}

	return nil
}

func verifyIPSecTunnelDetails(local, remote *ipsec.IpsecSaDetails) error {
	if local.SwIfIndex != remote.SwIfIndex {
		return fmt.Errorf("swIfIndex data mismatch (local: %v, remote: %v)",
			local.SwIfIndex, remote.SwIfIndex)
	}
	localIsTunnel, remoteIsTunnel := local.Entry.Flags&ipsec.IPSEC_API_SAD_FLAG_IS_TUNNEL, remote.Entry.Flags&ipsec.IPSEC_API_SAD_FLAG_IS_TUNNEL
	if localIsTunnel != remoteIsTunnel {
		return fmt.Errorf("tunnel data mismatch (local: %v, remote: %v)",
			localIsTunnel, remoteIsTunnel)
	}

	localSrc, localDst := local.Entry.TunnelSrc.Un.XXX_UnionData, local.Entry.TunnelDst.Un.XXX_UnionData
	remoteSrc, remoteDst := remote.Entry.TunnelSrc.Un.XXX_UnionData, remote.Entry.TunnelDst.Un.XXX_UnionData
	if (local.Entry.Flags&ipsec.IPSEC_API_SAD_FLAG_IS_TUNNEL_V6) != (remote.Entry.Flags&ipsec.IPSEC_API_SAD_FLAG_IS_TUNNEL_V6) ||
		!bytes.Equal(localSrc[:], remoteDst[:]) ||
		!bytes.Equal(localDst[:], remoteSrc[:]) {
		return fmt.Errorf("src/dst IP mismatch (local: %+v, remote: %+v)",
			local.Entry, remote.Entry)
	}

	return nil
}

// dumpBondDetails dumps bond interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpBondDetails(ifs map[uint32]*vppcalls.InterfaceDetails) error {
	bondIndexes := make([]uint32, 0)
	reqCtx := h.callsChannel.SendMultiRequest(&bond.SwInterfaceBondDump{})
	for {
		bondDetails := &bond.SwInterfaceBondDetails{}
		stop, err := reqCtx.ReceiveReply(bondDetails)
		if err != nil {
			return fmt.Errorf("failed to dump bond interface details: %v", err)
		}
		if stop {
			break
		}
		_, ifIdxExists := ifs[bondDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		ifs[bondDetails.SwIfIndex].Interface.Link = &interfaces.Interface_Bond{
			Bond: &interfaces.BondLink{
				Id:   bondDetails.ID,
				Mode: getBondIfMode(bondDetails.Mode),
				Lb:   getBondLoadBalance(bondDetails.Lb),
			},
		}
		ifs[bondDetails.SwIfIndex].Interface.Type = interfaces.Interface_BOND_INTERFACE
		bondIndexes = append(bondIndexes, bondDetails.SwIfIndex)
	}

	// get slave interfaces for bonds
	for _, bondIdx := range bondIndexes {
		var bondSlaves []*interfaces.BondLink_BondedInterface
		reqSlCtx := h.callsChannel.SendMultiRequest(&bond.SwInterfaceSlaveDump{SwIfIndex: bondIdx})
		for {
			slaveDetails := &bond.SwInterfaceSlaveDetails{}
			stop, err := reqSlCtx.ReceiveReply(slaveDetails)
			if err != nil {
				return fmt.Errorf("failed to dump bond slave details: %v", err)
			}
			if stop {
				break
			}
			slaveIf, ifIdxExists := ifs[slaveDetails.SwIfIndex]
			if !ifIdxExists {
				continue
			}
			bondSlaves = append(bondSlaves, &interfaces.BondLink_BondedInterface{
				Name:          slaveIf.Interface.Name,
				IsPassive:     uintToBool(slaveDetails.IsPassive),
				IsLongTimeout: uintToBool(slaveDetails.IsLongTimeout),
			})
			ifs[bondIdx].Interface.GetBond().BondedInterfaces = bondSlaves
		}
	}

	return nil
}

func (h *InterfaceVppHandler) dumpGreDetails(ifs map[uint32]*vppcalls.InterfaceDetails) error {
	msg := &gre.GreTunnelDump{SwIfIndex: gre.InterfaceIndex(^uint32(0))}
	reqCtx := h.callsChannel.SendMultiRequest(msg)
	for {
		greDetails := &gre.GreTunnelDetails{}
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

		if tunnel.Src.Af == gre.ADDRESS_IP4 {
			srcAddrArr := tunnel.Src.Un.GetIP4()
			srcAddr = net.IP(srcAddrArr[:])
		} else {
			srcAddrArr := tunnel.Src.Un.GetIP6()
			srcAddr = net.IP(srcAddrArr[:])
		}
		if tunnel.Dst.Af == gre.ADDRESS_IP4 {
			dstAddrArr := tunnel.Dst.Un.GetIP4()
			dstAddr = net.IP(dstAddrArr[:])
		} else {
			dstAddrArr := tunnel.Dst.Un.GetIP6()
			dstAddr = net.IP(dstAddrArr[:])
		}

		ifs[swIfIndex].Interface.Link = &interfaces.Interface_Gre{
			Gre: &interfaces.GreLink{
				TunnelType: getGreTunnelType(tunnel.Type),
				SrcAddr:    srcAddr.String(),
				DstAddr:    dstAddr.String(),
				OuterFibId: tunnel.OuterFibID,
				SessionId:  uint32(tunnel.SessionID),
			},
		}
		ifs[swIfIndex].Interface.Type = interfaces.Interface_GRE_TUNNEL
	}
	return nil
}

// dumpUnnumberedDetails returns a map of unnumbered interface indexes, every with interface index of element with IP
func (h *InterfaceVppHandler) dumpUnnumberedDetails() (map[uint32]uint32, error) {
	unIfMap := make(map[uint32]uint32) // unnumbered/ip-interface
	reqCtx := h.callsChannel.SendMultiRequest(&ip.IPUnnumberedDump{
		SwIfIndex: ^uint32(0),
	})

	for {
		unDetails := &ip.IPUnnumberedDetails{}
		last, err := reqCtx.ReceiveReply(unDetails)
		if last {
			break
		}
		if err != nil {
			return nil, err
		}

		unIfMap[unDetails.SwIfIndex] = unDetails.IPSwIfIndex
	}

	return unIfMap, nil
}

func (h *InterfaceVppHandler) dumpRxPlacement(ifs map[uint32]*vppcalls.InterfaceDetails) error {
	reqCtx := h.callsChannel.SendMultiRequest(&binapi_interface.SwInterfaceRxPlacementDump{
		SwIfIndex: ^uint32(0),
	})
	for {
		rxDetails := &binapi_interface.SwInterfaceRxPlacementDetails{}
		stop, err := reqCtx.ReceiveReply(rxDetails)
		if err != nil {
			return fmt.Errorf("failed to dump rx-placement details: %v", err)
		}
		if stop {
			break
		}

		ifData, ok := ifs[rxDetails.SwIfIndex]
		if !ok {
			h.log.Warnf("Received rx-placement data for unknown interface with index %d", rxDetails.SwIfIndex)
			continue
		}

		ifData.Interface.RxModes = append(ifData.Interface.RxModes,
			&interfaces.Interface_RxMode{
				Queue: rxDetails.QueueID,
				Mode:  getRxModeType(rxDetails.Mode),
			})

		var worker uint32
		if rxDetails.WorkerID > 0 {
			worker = rxDetails.WorkerID - 1
		}
		ifData.Interface.RxPlacements = append(ifData.Interface.RxPlacements,
			&interfaces.Interface_RxPlacement{
				Queue:      rxDetails.QueueID,
				Worker:     worker,
				MainThread: rxDetails.WorkerID == 0,
			})
	}
	return nil
}

// guessInterfaceType attempts to guess the correct interface type from its internal name (as given by VPP).
// This is required mainly for those interface types, that do not provide dump binary API,
// such as loopback of af_packet.
func guessInterfaceType(ifName string) interfaces.Interface_Type {
	switch {
	case strings.HasPrefix(ifName, "loop"),
		strings.HasPrefix(ifName, "local"):
		return interfaces.Interface_SOFTWARE_LOOPBACK

	case strings.HasPrefix(ifName, "memif"):
		return interfaces.Interface_MEMIF

	case strings.HasPrefix(ifName, "tap"):
		return interfaces.Interface_TAP

	case strings.HasPrefix(ifName, "host"):
		return interfaces.Interface_AF_PACKET

	case strings.HasPrefix(ifName, "vxlan"):
		return interfaces.Interface_VXLAN_TUNNEL

	case strings.HasPrefix(ifName, "ipsec"):
		return interfaces.Interface_IPSEC_TUNNEL

	case strings.HasPrefix(ifName, "vmxnet3"):
		return interfaces.Interface_VMXNET3_INTERFACE

	case strings.HasPrefix(ifName, "Bond"):
		return interfaces.Interface_BOND_INTERFACE

	case strings.HasPrefix(ifName, "gre"):
		return interfaces.Interface_GRE_TUNNEL

	case strings.HasPrefix(ifName, "gtpu"):
		return interfaces.Interface_GTPU_TUNNEL

	case strings.HasPrefix(ifName, "ipip"):
		return interfaces.Interface_IPIP_TUNNEL

	default:
		return interfaces.Interface_DPDK
	}
}

// memifModetoNB converts binary API type of memif mode to the northbound API type memif mode.
func memifModetoNB(mode uint8) interfaces.MemifLink_MemifMode {
	switch mode {
	case 0:
		return interfaces.MemifLink_ETHERNET
	case 1:
		return interfaces.MemifLink_IP
	case 2:
		return interfaces.MemifLink_PUNT_INJECT
	default:
		return interfaces.MemifLink_ETHERNET
	}
}

// Convert binary API rx-mode to northbound representation
func getRxModeType(mode uint8) interfaces.Interface_RxMode_Type {
	switch mode {
	case 1:
		return interfaces.Interface_RxMode_POLLING
	case 2:
		return interfaces.Interface_RxMode_INTERRUPT
	case 3:
		return interfaces.Interface_RxMode_ADAPTIVE
	case 4:
		return interfaces.Interface_RxMode_DEFAULT
	default:
		return interfaces.Interface_RxMode_UNKNOWN
	}
}

func getBondIfMode(mode uint8) interfaces.BondLink_Mode {
	switch mode {
	case 1:
		return interfaces.BondLink_ROUND_ROBIN
	case 2:
		return interfaces.BondLink_ACTIVE_BACKUP
	case 3:
		return interfaces.BondLink_XOR
	case 4:
		return interfaces.BondLink_BROADCAST
	case 5:
		return interfaces.BondLink_LACP
	default:
		// UNKNOWN
		return 0
	}
}

func getBondLoadBalance(lb uint8) interfaces.BondLink_LoadBalance {
	switch lb {
	case 1:
		return interfaces.BondLink_L34
	case 2:
		return interfaces.BondLink_L23
	default:
		return interfaces.BondLink_L2
	}
}

func getTagRwOption(op uint32) interfaces.SubInterface_TagRewriteOptions {
	switch op {
	case 1:
		return interfaces.SubInterface_PUSH1
	case 2:
		return interfaces.SubInterface_PUSH2
	case 3:
		return interfaces.SubInterface_POP1
	case 4:
		return interfaces.SubInterface_POP2
	case 5:
		return interfaces.SubInterface_TRANSLATE11
	case 6:
		return interfaces.SubInterface_TRANSLATE12
	case 7:
		return interfaces.SubInterface_TRANSLATE21
	case 8:
		return interfaces.SubInterface_TRANSLATE22
	default: // disabled
		return interfaces.SubInterface_DISABLED
	}
}

func getGreTunnelType(tt gre.GreTunnelType) interfaces.GreLink_Type {
	switch tt {
	case gre.GRE_API_TUNNEL_TYPE_L3:
		return interfaces.GreLink_L3
	case gre.GRE_API_TUNNEL_TYPE_TEB:
		return interfaces.GreLink_TEB
	case gre.GRE_API_TUNNEL_TYPE_ERSPAN:
		return interfaces.GreLink_ERSPAN
	default:
		return interfaces.GreLink_UNKNOWN
	}
}

func getVxLanGpeProtocol(p uint8) interfaces.VxlanLink_Gpe_Protocol {
	switch p {
	case 1:
		return interfaces.VxlanLink_Gpe_IP4
	case 2:
		return interfaces.VxlanLink_Gpe_IP6
	case 3:
		return interfaces.VxlanLink_Gpe_ETHERNET
	case 4:
		return interfaces.VxlanLink_Gpe_NSH
	default:
		return interfaces.VxlanLink_Gpe_UNKNOWN
	}
}
