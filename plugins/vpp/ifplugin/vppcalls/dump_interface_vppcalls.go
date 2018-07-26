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

	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
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

	// Get vrf for every interface
	for _, ifData := range ifs {
		vrf, err := handler.GetInterfaceVRF(ifData.Meta.SwIfIndex)
		if err != nil {
			handler.log.Warnf("Interface dump: failed to get VRF from interface %d", ifData.Meta.SwIfIndex)
			continue
		}
		ifData.Interface.Vrf = vrf
	}

	handler.log.Debugf("dumped %d interfaces", len(ifs))

	// SwInterfaceDump time
	timeLog := measure.GetTimeLog(interfaces.SwInterfaceDump{}, handler.stopwatch)
	if timeLog != nil {
		timeLog.LogTimeEntry(time.Since(start))
	}

	for idx := range ifs {
		vrfID, err := handler.GetInterfaceVRF(idx)
		if err != nil {
			return nil, err
		}
		ifs[idx].Interface.Vrf = vrfID
	}

	timeLog = measure.GetTimeLog(ip.IPAddressDump{}, handler.stopwatch)
	err := handler.dumpIPAddressDetails(ifs, 0, timeLog)
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
	// TODO: workaround for incorrect ip.IPAddressDetails message
	notifChan := make(chan api.Message, 100)
	subs, _ := handler.callsChannel.SubscribeNotification(notifChan, ip.NewIPAddressDetails)

	// Dump IP addresses of each interface.
	for idx := range ifs {
		// IPAddressDetails time measurement
		start := time.Now()

		reqCtx := handler.callsChannel.SendMultiRequest(&ip.IPAddressDump{SwIfIndex: idx, IsIpv6: isIPv6})
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

		// TODO: workaround for incorrect ip.IPAddressDetails message
		for len(notifChan) > 0 {
			notifMsg := <-notifChan
			handler.processIPDetails(ifs, notifMsg.(*ip.IPAddressDetails))
		}

		// IPAddressDump time
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}

	// TODO: workaround for incorrect ip.IPAddressDetails message
	handler.callsChannel.UnsubscribeNotification(subs)

	return nil
}

// processIPDetails processes ip.IPAddressDetails binary API message and fills the details into the provided interface map.
func (handler *ifVppHandler) processIPDetails(ifs map[uint32]*InterfaceDetails, ipDetails *ip.IPAddressDetails) {
	_, ifIdxExists := ifs[ipDetails.SwIfIndex]
	if !ifIdxExists {
		return
	}
	if ifs[ipDetails.SwIfIndex].Interface.IpAddresses == nil {
		ifs[ipDetails.SwIfIndex].Interface.IpAddresses = make([]string, 0)
	}
	var ipAddr string
	if ipDetails.IsIpv6 == 1 {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipDetails.IP).To16().String(), uint32(ipDetails.PrefixLength))
	} else {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipDetails.IP[:4]).To4().String(), uint32(ipDetails.PrefixLength))
	}
	ifs[ipDetails.SwIfIndex].Interface.IpAddresses = append(ifs[ipDetails.SwIfIndex].Interface.IpAddresses, ipAddr)
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

		if vxlanDetails.IsIpv6 == 1 {
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
