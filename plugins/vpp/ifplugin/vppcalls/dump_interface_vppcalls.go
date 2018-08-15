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

package vppcalls

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/memif"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/tap"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/tapv2"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vxlan"
	ifnb "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
)

// Default VPP MTU value
const defaultVPPMtu = 9216

// InterfaceDetails is the wrapper structure for the interface northbound API structure.
type InterfaceDetails struct {
	Interface *ifnb.Interfaces_Interface `json:"interface"`
	Meta      *InterfaceMeta             `json:"interface_meta"`
}

// InterfaceMeta is combination of proto-modelled Interface data and VPP provided metadata
type InterfaceMeta struct {
	SwIfIndex    uint32 `json:"sw_if_index"`
	Tag          string `json:"tag"`
	InternalName string `json:"internal_name"`
	Dhcp         *Dhcp  `json:"dhcp"`
}

// Dhcp is helper struct for DHCP metadata, split to client and lease (similar to VPP binary API)
type Dhcp struct {
	Client *Client `json:"dhcp_client"`
	Lease  *Lease  `json:"dhcp_lease"`
}

// Client is helper struct grouping DHCP client data
type Client struct {
	SwIfIndex        uint32
	Hostname         string
	ID               string
	WantDhcpEvent    bool
	SetBroadcastFlag bool
	Pid              uint32
}

// Lease is helper struct grouping DHCP lease data
type Lease struct {
	SwIfIndex     uint32
	State         uint8
	Hostname      string
	IsIPv6        bool
	MaskWidth     uint8
	HostAddress   string
	RouterAddress string
	HostMac       string
}

func (handler *ifVppHandler) DumpInterfacesByType(reqType ifnb.InterfaceType) (map[uint32]*InterfaceDetails, error) {
	// Dump all
	ifs, err := handler.DumpInterfaces()
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

func (handler *ifVppHandler) DumpInterfaces() (map[uint32]*InterfaceDetails, error) {
	start := time.Now()
	// map for the resulting interfaces
	ifs := make(map[uint32]*InterfaceDetails)

	// First, dump all interfaces to create initial data.
	reqCtx := handler.callsChannel.SendMultiRequest(&interfaces.SwInterfaceDump{})

	for {
		ifDetails := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(ifDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dump interface: %v", err)
		}

		details := &InterfaceDetails{
			Interface: &ifnb.Interfaces_Interface{
				Name:        string(bytes.SplitN(ifDetails.Tag, []byte{0x00}, 2)[0]),
				Type:        guessInterfaceType(string(ifDetails.InterfaceName)), // the type may be amended later by further dumps
				Enabled:     ifDetails.AdminUpDown > 0,
				PhysAddress: net.HardwareAddr(ifDetails.L2Address[:ifDetails.L2AddressLength]).String(),
				Mtu: func(vppMtu uint16) uint32 {
					// If default VPP MTU value is set, return 0 (it means MTU was not set in the NB config)
					if vppMtu == defaultVPPMtu {
						return 0
					}
					return uint32(vppMtu)
				}(ifDetails.LinkMtu),
			},
			Meta: &InterfaceMeta{
				SwIfIndex:    ifDetails.SwIfIndex,
				Tag:          string(bytes.SplitN(ifDetails.Tag, []byte{0x00}, 2)[0]),
				InternalName: string(bytes.SplitN(ifDetails.InterfaceName, []byte{0x00}, 2)[0]),
			},
		}
		// Fill name for physical interfaces (they are mostly without tag)
		if details.Interface.Type == ifnb.InterfaceType_ETHERNET_CSMACD {
			details.Interface.Name = details.Meta.InternalName
		}
		ifs[ifDetails.SwIfIndex] = details

		if details.Interface.Type == ifnb.InterfaceType_AF_PACKET_INTERFACE {
			fillAFPacketDetails(ifs, ifDetails.SwIfIndex, details.Meta.InternalName)
		}
	}

	// Get DHCP clients
	dhcpClients, err := handler.dumpDhcpClients()
	if err != nil {
		return nil, fmt.Errorf("failed to dump interface DHCP clients: %v", err)
	}
	// Get vrf for every interface and fill DHCP if set
	for _, ifData := range ifs {
		// VRF
		vrf, err := handler.GetInterfaceVRF(ifData.Meta.SwIfIndex)
		if err != nil {
			handler.log.Warnf("Interface dump: failed to get VRF from interface %d: %v", ifData.Meta.SwIfIndex, err)
			continue
		}
		ifData.Interface.Vrf = vrf
		// DHCP
		dhcpData, ok := dhcpClients[ifData.Meta.SwIfIndex]
		if ok {
			ifData.Interface.SetDhcpClient = true
			ifData.Meta.Dhcp = dhcpData
		}
	}

	handler.log.Debugf("dumped %d interfaces", len(ifs))

	// SwInterfaceDump time
	timeLog := measure.GetTimeLog(interfaces.SwInterfaceDump{}, handler.stopwatch)
	if timeLog != nil {
		timeLog.LogTimeEntry(time.Since(start))
	}

	timeLog = measure.GetTimeLog(ip.IPAddressDump{}, handler.stopwatch)
	err = handler.dumpIPAddressDetails(ifs, 0, timeLog)
	if err != nil {
		return nil, err
	}
	err = handler.dumpIPAddressDetails(ifs, 1, timeLog)
	if err != nil {
		return nil, err
	}

	err = handler.dumpMemifDetails(ifs)
	if err != nil {
		return nil, err
	}

	err = handler.dumpTapDetails(ifs)
	if err != nil {
		return nil, err
	}

	err = handler.dumpVxlanDetails(ifs)
	if err != nil {
		return nil, err
	}

	return ifs, nil
}

func (handler *ifVppHandler) DumpMemifSocketDetails() (map[string]uint32, error) {
	// MemifSocketFilenameDump time measurement
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(memif.MemifSocketFilenameDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	memifSocketMap := make(map[string]uint32)

	reqCtx := handler.callsChannel.SendMultiRequest(&memif.MemifSocketFilenameDump{})
	for {
		socketDetails := &memif.MemifSocketFilenameDetails{}
		stop, err := reqCtx.ReceiveReply(socketDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return memifSocketMap, fmt.Errorf("failed to dump memif socket filename details: %v", err)
		}

		filename := string(bytes.SplitN(socketDetails.SocketFilename, []byte{0x00}, 2)[0])
		memifSocketMap[filename] = socketDetails.SocketID
	}

	handler.log.Debugf("Memif socket dump completed, found %d entries", len(memifSocketMap))

	return memifSocketMap, nil
}

// dumpIPAddressDetails dumps IP address details of interfaces from VPP and fills them into the provided interface map.
func (handler *ifVppHandler) dumpIPAddressDetails(ifs map[uint32]*InterfaceDetails, isIPv6 uint8, timeLog measure.StopWatchEntry) error {
	// Dump IP addresses of each interface.
	for idx := range ifs {
		// IPAddressDetails time measurement
		start := time.Now()

		reqCtx := handler.callsChannel.SendMultiRequest(&ip.IPAddressDump{
			SwIfIndex: idx,
			IsIPv6:    isIPv6,
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
			handler.processIPDetails(ifs, ipDetails)
		}

		// IPAddressDump time
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

// processIPDetails processes ip.IPAddressDetails binary API message and fills the details into the provided interface map.
func (handler *ifVppHandler) processIPDetails(ifs map[uint32]*InterfaceDetails, ipDetails *ip.IPAddressDetails) {
	ifDetails, ifIdxExists := ifs[ipDetails.SwIfIndex]
	if !ifIdxExists {
		return
	}
	var ipAddr string
	if ipDetails.IsIPv6 == 1 {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipDetails.IP).To16().String(), uint32(ipDetails.PrefixLength))
	} else {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipDetails.IP[:4]).To4().String(), uint32(ipDetails.PrefixLength))
	}
	ifDetails.Interface.IpAddresses = append(ifDetails.Interface.IpAddresses, ipAddr)
}

// fillAFPacketDetails fills af_packet interface details into the provided interface map.
func fillAFPacketDetails(ifs map[uint32]*InterfaceDetails, swIfIndex uint32, ifName string) {
	ifs[swIfIndex].Interface.Afpacket = &ifnb.Interfaces_Interface_Afpacket{
		HostIfName: strings.TrimPrefix(ifName, "host-"),
	}
	ifs[swIfIndex].Interface.Type = ifnb.InterfaceType_AF_PACKET_INTERFACE
}

// dumpMemifDetails dumps memif interface details from VPP and fills them into the provided interface map.
func (handler *ifVppHandler) dumpMemifDetails(ifs map[uint32]*InterfaceDetails) error {
	// MemifDetails time measurement
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(memif.MemifDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Dump all memif sockets
	memifSocketMap, err := handler.DumpMemifSocketDetails()
	if err != nil {
		return err
	}

	reqCtx := handler.callsChannel.SendMultiRequest(&memif.MemifDump{})
	for {
		memifDetails := &memif.MemifDetails{}
		stop, err := reqCtx.ReceiveReply(memifDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump memif interface: %v", err)
		}
		_, ifIdxExists := ifs[memifDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		ifs[memifDetails.SwIfIndex].Interface.Memif = &ifnb.Interfaces_Interface_Memif{
			Master: memifDetails.Role == 0,
			Mode:   memifModetoNB(memifDetails.Mode),
			Id:     memifDetails.ID,
			// TODO: Secret - not available in the binary API
			SocketFilename: func(socketMap map[string]uint32) (filename string) {
				for filename, id := range socketMap {
					if memifDetails.SocketID == id {
						return filename
					}
				}
				// Socket for configured memif should exist
				handler.log.Warnf("Socket ID not found for memif %v", memifDetails.SwIfIndex)
				return
			}(memifSocketMap),
			RingSize:   memifDetails.RingSize,
			BufferSize: uint32(memifDetails.BufferSize),
			// TODO: RxQueues, TxQueues - not available in the binary API
		}
		ifs[memifDetails.SwIfIndex].Interface.Type = ifnb.InterfaceType_MEMORY_INTERFACE
	}

	return nil
}

// dumpTapDetails dumps tap interface details from VPP and fills them into the provided interface map.
func (handler *ifVppHandler) dumpTapDetails(ifs map[uint32]*InterfaceDetails) error {
	// SwInterfaceTapDump time measurement
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(tap.SwInterfaceTapDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Original TAP.
	reqCtx := handler.callsChannel.SendMultiRequest(&tap.SwInterfaceTapDump{})
	for {
		tapDetails := &tap.SwInterfaceTapDetails{}
		stop, err := reqCtx.ReceiveReply(tapDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump TAP interface details: %v", err)
		}
		_, ifIdxExists := ifs[tapDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		ifs[tapDetails.SwIfIndex].Interface.Tap = &ifnb.Interfaces_Interface_Tap{
			Version:    1,
			HostIfName: string(bytes.SplitN(tapDetails.DevName, []byte{0x00}, 2)[0]),
		}
		ifs[tapDetails.SwIfIndex].Interface.Type = ifnb.InterfaceType_TAP_INTERFACE
	}

	// TAP v.2
	reqCtx = handler.callsChannel.SendMultiRequest(&tapv2.SwInterfaceTapV2Dump{})
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
		ifs[tapDetails.SwIfIndex].Interface.Tap = &ifnb.Interfaces_Interface_Tap{
			Version:    2,
			HostIfName: string(bytes.SplitN(tapDetails.HostIfName, []byte{0x00}, 2)[0]),
			// Other parameters are not not yet part of the dump.

		}
		ifs[tapDetails.SwIfIndex].Interface.Type = ifnb.InterfaceType_TAP_INTERFACE
	}

	return nil
}

// dumpVxlanDetails dumps VXLAN interface details from VPP and fills them into the provided interface map.
func (handler *ifVppHandler) dumpVxlanDetails(ifs map[uint32]*InterfaceDetails) error {
	// VxlanTunnelDump time measurement
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(vxlan.VxlanTunnelDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	reqCtx := handler.callsChannel.SendMultiRequest(&vxlan.VxlanTunnelDump{SwIfIndex: ^uint32(0)})
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
			ifs[vxlanDetails.SwIfIndex].Interface.Vxlan = &ifnb.Interfaces_Interface_Vxlan{
				Multicast:  multicastIfName,
				SrcAddress: net.IP(vxlanDetails.SrcAddress).To16().String(),
				DstAddress: net.IP(vxlanDetails.DstAddress).To16().String(),
				Vni:        vxlanDetails.Vni,
			}
		} else {
			ifs[vxlanDetails.SwIfIndex].Interface.Vxlan = &ifnb.Interfaces_Interface_Vxlan{
				Multicast:  multicastIfName,
				SrcAddress: net.IP(vxlanDetails.SrcAddress[:4]).To4().String(),
				DstAddress: net.IP(vxlanDetails.DstAddress[:4]).To4().String(),
				Vni:        vxlanDetails.Vni,
			}
		}
		ifs[vxlanDetails.SwIfIndex].Interface.Type = ifnb.InterfaceType_VXLAN_TUNNEL
	}

	return nil
}

// dumpDhcpClients returns a slice of DhcpMeta with all interfaces and other DHCP-related information available
func (handler *ifVppHandler) dumpDhcpClients() (map[uint32]*Dhcp, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(dhcp.DHCPClientDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	dhcpData := make(map[uint32]*Dhcp)
	reqCtx := handler.callsChannel.SendMultiRequest(&dhcp.DHCPClientDump{})

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
		dhcpClient := &Client{
			SwIfIndex:        client.SwIfIndex,
			Hostname:         string(bytes.SplitN(client.Hostname, []byte{0x00}, 2)[0]),
			ID:               string(bytes.SplitN(client.ID, []byte{0x00}, 2)[0]),
			WantDhcpEvent:    uintToBool(client.WantDHCPEvent),
			SetBroadcastFlag: uintToBool(client.SetBroadcastFlag),
			Pid:              client.PID,
		}

		// DHCP lease data
		dhcpLease := &Lease{
			SwIfIndex:     lease.SwIfIndex,
			State:         lease.State,
			Hostname:      string(bytes.SplitN(lease.Hostname, []byte{0x00}, 2)[0]),
			IsIPv6:        uintToBool(lease.IsIPv6),
			MaskWidth:     lease.MaskWidth,
			HostAddress:   hostAddr,
			RouterAddress: routerAddr,
			HostMac:       hostMac.String(),
		}

		// DHCP metadata
		dhcpData[client.SwIfIndex] = &Dhcp{
			Client: dhcpClient,
			Lease:  dhcpLease,
		}
	}

	return dhcpData, nil
}

// guessInterfaceType attempts to guess the correct interface type from its internal name (as given by VPP).
// This is required mainly for those interface types, that do not provide dump binary API,
// such as loopback of af_packet.
func guessInterfaceType(ifName string) ifnb.InterfaceType {
	switch {
	case strings.HasPrefix(ifName, "loop"):
		return ifnb.InterfaceType_SOFTWARE_LOOPBACK
	case strings.HasPrefix(ifName, "local"):
		return ifnb.InterfaceType_SOFTWARE_LOOPBACK
	case strings.HasPrefix(ifName, "memif"):
		return ifnb.InterfaceType_MEMORY_INTERFACE
	case strings.HasPrefix(ifName, "tap"):
		return ifnb.InterfaceType_TAP_INTERFACE
	case strings.HasPrefix(ifName, "host"):
		return ifnb.InterfaceType_AF_PACKET_INTERFACE
	case strings.HasPrefix(ifName, "vxlan"):
		return ifnb.InterfaceType_VXLAN_TUNNEL
	}
	return ifnb.InterfaceType_ETHERNET_CSMACD
}

// memifModetoNB converts binary API type of memif mode to the northbound API type memif mode.
func memifModetoNB(mode uint8) ifnb.Interfaces_Interface_Memif_MemifMode {
	switch mode {
	case 0:
		return ifnb.Interfaces_Interface_Memif_ETHERNET
	case 1:
		return ifnb.Interfaces_Interface_Memif_IP
	case 2:
		return ifnb.Interfaces_Interface_Memif_PUNT_INJECT
	}
	return ifnb.Interfaces_Interface_Memif_ETHERNET
}
