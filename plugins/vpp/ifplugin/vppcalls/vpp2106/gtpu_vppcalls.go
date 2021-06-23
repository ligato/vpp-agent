//  Copyright (c) 2019 EMnify
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

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/gtpu"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

const defaultDecapNextIndex = 0xFFFFFFFF

func (h *InterfaceVppHandler) gtpuAddDelTunnel(isAdd bool, gtpuLink *interfaces.GtpuLink, multicastIf uint32) (uint32, error) {
	var decapNextNode uint32 = defaultDecapNextIndex
	if gtpuLink.DecapNextNode != 0 {
		decapNextNode = gtpuLink.DecapNextNode
	} else {
		// backwards compatible fallback
		if gtpuLink.DecapNext != interfaces.GtpuLink_DEFAULT {
			decapNextNode = uint32(gtpuLink.DecapNext)
		}
	}

	req := &gtpu.GtpuAddDelTunnel{
		IsAdd:          isAdd,
		McastSwIfIndex: interface_types.InterfaceIndex(multicastIf),
		EncapVrfID:     gtpuLink.EncapVrfId,
		Teid:           gtpuLink.Teid,
		Tteid:          gtpuLink.RemoteTeid,
		DecapNextIndex: decapNextNode,
	}

	srcAddr := net.ParseIP(gtpuLink.SrcAddr)
	if srcAddr == nil {
		err := errors.New("bad source address for GTPU tunnel")
		return 0, err
	}

	dstAddr := net.ParseIP(gtpuLink.DstAddr)
	if dstAddr == nil {
		err := errors.New("bad destination address for GTPU tunnel")
		return 0, err
	}

	if gtpuLink.SrcAddr == gtpuLink.DstAddr {
		err := errors.New("source and destination are the same")
		return 0, err
	}

	var isSrcIPv6, isDstIPv6 bool

	if srcAddr.To4() == nil {
		isSrcIPv6 = true
	}
	if dstAddr.To4() == nil {
		isDstIPv6 = true
	}

	if !isSrcIPv6 && !isDstIPv6 {
		var src, dst [4]uint8
		copy(src[:], srcAddr.To4())
		copy(dst[:], dstAddr.To4())
		req.SrcAddress = ip_types.Address{
			Af: ip_types.ADDRESS_IP4,
			Un: ip_types.AddressUnionIP4(src),
		}
		req.DstAddress = ip_types.Address{
			Af: ip_types.ADDRESS_IP4,
			Un: ip_types.AddressUnionIP4(dst),
		}
	} else if isSrcIPv6 && isDstIPv6 {
		var src, dst [16]uint8
		copy(src[:], srcAddr.To16())
		copy(dst[:], dstAddr.To16())
		req.SrcAddress = ip_types.Address{
			Af: ip_types.ADDRESS_IP6,
			Un: ip_types.AddressUnionIP6(src),
		}
		req.DstAddress = ip_types.Address{
			Af: ip_types.ADDRESS_IP6,
			Un: ip_types.AddressUnionIP6(dst),
		}
	} else {
		return 0, errors.New("source and destination addresses must be both either IPv4 or IPv6")
	}

	reply := &gtpu.GtpuAddDelTunnelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	return uint32(reply.SwIfIndex), nil
}

// AddGtpuTunnel adds new GTPU interface.
func (h *InterfaceVppHandler) AddGtpuTunnel(ifName string, gtpuLink *interfaces.GtpuLink, multicastIf uint32) (uint32, error) {
	if h.gtpu == nil {
		return 0, errors.WithMessage(vpp.ErrPluginDisabled, "gtpu")
	}
	if gtpuLink == nil {
		return 0, errors.New("missing GTPU tunnel information")
	}

	swIfIndex, err := h.gtpuAddDelTunnel(true, gtpuLink, multicastIf)
	if err != nil {
		return 0, err
	}
	return swIfIndex, h.SetInterfaceTag(ifName, swIfIndex)
}

// DelGtpuTunnel removes GTPU interface.
func (h *InterfaceVppHandler) DelGtpuTunnel(ifName string, gtpuLink *interfaces.GtpuLink) error {
	if h.gtpu == nil {
		return errors.WithMessage(vpp.ErrPluginDisabled, "gtpu")
	}
	if gtpuLink == nil {
		return errors.New("missing GTPU tunnel information")
	}

	swIfIndex, err := h.gtpuAddDelTunnel(false, gtpuLink, 0)
	if err != nil {
		return err
	}
	return h.RemoveInterfaceTag(ifName, swIfIndex)
}

// dumpGtpuDetails dumps GTP-U interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpGtpuDetails(ifc map[uint32]*vppcalls.InterfaceDetails) error {
	if h.gtpu == nil {
		// no-op when disabled
		return nil
	}

	reqCtx := h.callsChannel.SendMultiRequest(&gtpu.GtpuTunnelDump{
		SwIfIndex: ^interface_types.InterfaceIndex(0),
	})
	for {
		gtpuDetails := &gtpu.GtpuTunnelDetails{}
		stop, err := reqCtx.ReceiveReply(gtpuDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump GTP-U tunnel interface details: %v", err)
		}
		_, ifIdxExists := ifc[uint32(gtpuDetails.SwIfIndex)]
		if !ifIdxExists {
			continue
		}
		// Multicast interface
		var multicastIfName string
		_, exists := ifc[uint32(gtpuDetails.McastSwIfIndex)]
		if exists {
			multicastIfName = ifc[uint32(gtpuDetails.McastSwIfIndex)].Interface.Name
		}

		gtpuLink := &interfaces.GtpuLink{
			Multicast:     multicastIfName,
			EncapVrfId:    gtpuDetails.EncapVrfID,
			Teid:          gtpuDetails.Teid,
			RemoteTeid:    gtpuDetails.Tteid,
			DecapNextNode: gtpuDetails.DecapNextIndex,
		}

		if gtpuDetails.SrcAddress.Af == ip_types.ADDRESS_IP6 {
			srcAddrArr := gtpuDetails.SrcAddress.Un.GetIP6()
			gtpuLink.SrcAddr = net.IP(srcAddrArr[:]).To16().String()
			dstAddrArr := gtpuDetails.DstAddress.Un.GetIP6()
			gtpuLink.DstAddr = net.IP(dstAddrArr[:]).To16().String()
		} else {
			srcAddrArr := gtpuDetails.SrcAddress.Un.GetIP4()
			gtpuLink.SrcAddr = net.IP(srcAddrArr[:4]).To4().String()
			dstAddrArr := gtpuDetails.DstAddress.Un.GetIP4()
			gtpuLink.DstAddr = net.IP(dstAddrArr[:4]).To4().String()
		}

		ifc[uint32(gtpuDetails.SwIfIndex)].Interface.Link = &interfaces.Interface_Gtpu{Gtpu: gtpuLink}
		ifc[uint32(gtpuDetails.SwIfIndex)].Interface.Type = interfaces.Interface_GTPU_TUNNEL
	}

	return nil
}
