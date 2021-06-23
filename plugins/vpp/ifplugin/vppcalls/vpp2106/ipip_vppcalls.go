//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package vpp2106

import (
	"fmt"
	"net"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ipip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/tunnel_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// AddIpipTunnel adds new IPIP tunnel interface.
func (h *InterfaceVppHandler) AddIpipTunnel(ifName string, vrf uint32, ipipLink *interfaces.IPIPLink) (uint32, error) {
	req := &ipip.IpipAddTunnel{
		Tunnel: ipip.IpipTunnel{
			Instance: ^uint32(0),
			TableID:  vrf,
		},
	}

	if ipipLink == nil {
		return 0, errors.New("missing IPIP tunnel information")
	}
	var srcAddr, dstAddr net.IP
	var isSrcIPv6, isDstIPv6 bool

	srcAddr = net.ParseIP(ipipLink.SrcAddr)
	if srcAddr == nil {
		err := errors.New("bad source address for IPIP tunnel")
		return 0, err
	}
	if srcAddr.To4() == nil {
		isSrcIPv6 = true
	}

	if ipipLink.TunnelMode == interfaces.IPIPLink_POINT_TO_POINT {
		dstAddr = net.ParseIP(ipipLink.DstAddr)
		if dstAddr == nil {
			err := errors.New("bad destination address for IPIP tunnel")
			return 0, err
		}
		if dstAddr.To4() == nil {
			isDstIPv6 = true
		}
	}

	if !isSrcIPv6 && (dstAddr == nil || !isDstIPv6) {
		var src, dst [4]uint8
		copy(src[:], srcAddr.To4())
		req.Tunnel.Src = ip_types.Address{
			Af: ip_types.ADDRESS_IP4,
			Un: ip_types.AddressUnionIP4(src),
		}
		if dstAddr != nil {
			copy(dst[:], dstAddr.To4())
			req.Tunnel.Dst = ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(dst),
			}
		}
	} else if isSrcIPv6 && (dstAddr == nil || isDstIPv6) {
		var src, dst [16]uint8
		copy(src[:], srcAddr.To16())
		req.Tunnel.Src = ip_types.Address{
			Af: ip_types.ADDRESS_IP6,
			Un: ip_types.AddressUnionIP6(src),
		}
		if dstAddr != nil {
			copy(dst[:], dstAddr.To16())
			req.Tunnel.Dst = ip_types.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ip_types.AddressUnionIP6(dst),
			}
		}
	} else {
		return 0, errors.New("source and destination addresses must be both either IPv4 or IPv6")
	}

	if ipipLink.TunnelMode == interfaces.IPIPLink_POINT_TO_MULTIPOINT {
		req.Tunnel.Mode = tunnel_types.TUNNEL_API_MODE_MP
	}

	reply := &ipip.IpipAddTunnelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	swIfIndex := uint32(reply.SwIfIndex)
	return swIfIndex, h.SetInterfaceTag(ifName, swIfIndex)
}

// DelIpipTunnel removes IPIP tunnel interface.
func (h *InterfaceVppHandler) DelIpipTunnel(ifName string, ifIdx uint32) error {
	req := &ipip.IpipDelTunnel{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
	}
	reply := &ipip.IpipDelTunnelReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return h.RemoveInterfaceTag(ifName, ifIdx)
}

// dumpIpipDetails dumps IPIP interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpIpipDetails(ifc map[uint32]*vppcalls.InterfaceDetails) error {

	reqCtx := h.callsChannel.SendMultiRequest(&ipip.IpipTunnelDump{
		SwIfIndex: ^interface_types.InterfaceIndex(0),
	})
	for {
		ipipDetails := &ipip.IpipTunnelDetails{}
		stop, err := reqCtx.ReceiveReply(ipipDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump IPIP tunnel interface details: %v", err)
		}
		_, ifIdxExists := ifc[uint32(ipipDetails.Tunnel.SwIfIndex)]
		if !ifIdxExists {
			continue
		}

		ipipLink := &interfaces.IPIPLink{}
		if ipipDetails.Tunnel.Src.Af == ip_types.ADDRESS_IP6 {
			srcAddrArr := ipipDetails.Tunnel.Src.Un.GetIP6()
			ipipLink.SrcAddr = net.IP(srcAddrArr[:]).To16().String()
			if ipipDetails.Tunnel.Mode == tunnel_types.TUNNEL_API_MODE_P2P {
				dstAddrArr := ipipDetails.Tunnel.Dst.Un.GetIP6()
				ipipLink.DstAddr = net.IP(dstAddrArr[:]).To16().String()
			}
		} else {
			srcAddrArr := ipipDetails.Tunnel.Src.Un.GetIP4()
			ipipLink.SrcAddr = net.IP(srcAddrArr[:4]).To4().String()
			if ipipDetails.Tunnel.Mode == tunnel_types.TUNNEL_API_MODE_P2P {
				dstAddrArr := ipipDetails.Tunnel.Dst.Un.GetIP4()
				ipipLink.DstAddr = net.IP(dstAddrArr[:4]).To4().String()
			}
		}

		if ipipDetails.Tunnel.Mode == tunnel_types.TUNNEL_API_MODE_MP {
			ipipLink.TunnelMode = interfaces.IPIPLink_POINT_TO_MULTIPOINT
		}

		// TODO: temporary fix since VPP does not dump the tunnel mode properly.
		// If dst address is empty, this must be a multipoint tunnel.
		if ipipLink.DstAddr == "0.0.0.0" {
			ipipLink.TunnelMode = interfaces.IPIPLink_POINT_TO_MULTIPOINT
			ipipLink.DstAddr = ""
		}

		ifc[uint32(ipipDetails.Tunnel.SwIfIndex)].Interface.Link = &interfaces.Interface_Ipip{Ipip: ipipLink}
		ifc[uint32(ipipDetails.Tunnel.SwIfIndex)].Interface.Type = interfaces.Interface_IPIP_TUNNEL
	}
	return nil
}
