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

package vpp1908

import (
	"fmt"
	"net"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/gtpu"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

const defaultDecapNextIndex = 0xFFFFFFFF

func (h *InterfaceVppHandler) gtpuAddDelTunnel(isAdd uint8, gtpuLink *interfaces.GtpuLink, multicastIf uint32) (uint32, error) {
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
		McastSwIfIndex: multicastIf,
		EncapVrfID:     gtpuLink.EncapVrfId,
		Teid:           gtpuLink.Teid,
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
		req.IsIPv6 = 0
		req.SrcAddress = []byte(srcAddr.To4())
		req.DstAddress = []byte(dstAddr.To4())
	} else if isSrcIPv6 && isDstIPv6 {
		req.IsIPv6 = 1
		req.SrcAddress = []byte(srcAddr.To16())
		req.DstAddress = []byte(dstAddr.To16())
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

	swIfIndex, err := h.gtpuAddDelTunnel(1, gtpuLink, multicastIf)
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

	swIfIndex, err := h.gtpuAddDelTunnel(0, gtpuLink, 0)
	if err != nil {
		return err
	}
	return h.RemoveInterfaceTag(ifName, swIfIndex)
}

// dumpGtpuDetails dumps GTP-U interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpGtpuDetails(ifs map[uint32]*vppcalls.InterfaceDetails) error {
	if h.gtpu == nil {
		// no-op when disabled
		return nil
	}

	reqCtx := h.callsChannel.SendMultiRequest(&gtpu.GtpuTunnelDump{
		SwIfIndex: ^uint32(0),
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
		_, ifIdxExists := ifs[gtpuDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		// Multicast interface
		var multicastIfName string
		_, exists := ifs[gtpuDetails.McastSwIfIndex]
		if exists {
			multicastIfName = ifs[gtpuDetails.McastSwIfIndex].Interface.Name
		}

		gtpuLink := &interfaces.GtpuLink{
			Multicast:     multicastIfName,
			EncapVrfId:    gtpuDetails.EncapVrfID,
			Teid:          gtpuDetails.Teid,
			DecapNext:     interfaces.GtpuLink_NextNode(gtpuDetails.DecapNextIndex),
			DecapNextNode: gtpuDetails.DecapNextIndex,
		}

		if gtpuDetails.IsIPv6 == 1 {
			gtpuLink.SrcAddr = net.IP(gtpuDetails.SrcAddress).To16().String()
			gtpuLink.DstAddr = net.IP(gtpuDetails.DstAddress).To16().String()
		} else {
			gtpuLink.SrcAddr = net.IP(gtpuDetails.SrcAddress[:4]).To4().String()
			gtpuLink.DstAddr = net.IP(gtpuDetails.DstAddress[:4]).To4().String()
		}

		ifs[gtpuDetails.SwIfIndex].Interface.Link = &interfaces.Interface_Gtpu{Gtpu: gtpuLink}
		ifs[gtpuDetails.SwIfIndex].Interface.Type = interfaces.Interface_GTPU_TUNNEL
	}

	return nil
}
