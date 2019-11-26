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

package vpp2001_324

import (
	"errors"
	"fmt"
	"net"

	"go.ligato.io/vpp-agent/v2/plugins/vpp"
	vpp_gtpu "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/gtpu"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls"
	ifs "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/interfaces"
)

func (h *InterfaceVppHandler) gtpuAddDelTunnel(isAdd uint8, gtpuLink *ifs.GtpuLink, multicastIf uint32) (uint32, error) {
	req := &vpp_gtpu.GtpuAddDelTunnel{
		IsAdd:          isAdd,
		McastSwIfIndex: multicastIf,
		EncapVrfID:     gtpuLink.EncapVrfId,
		Teid:           gtpuLink.Teid,
	}

	if gtpuLink.DecapNext == ifs.GtpuLink_DEFAULT {
		req.DecapNextIndex = 0xFFFFFFFF
	} else {
		req.DecapNextIndex = uint32(gtpuLink.DecapNext)
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

	reply := &vpp_gtpu.GtpuAddDelTunnelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	return uint32(reply.SwIfIndex), nil
}

// AddGtpuTunnel adds new GTPU interface.
func (h *InterfaceVppHandler) AddGtpuTunnel(ifName string, gtpuLink *ifs.GtpuLink, multicastIf uint32) (uint32, error) {
    if gtpuLink == nil {
        return 0, errors.New("Missing GTPU tunnel information")
    }
    if h.gtpu == nil {
		return 0, vpp.ErrPluginDisabled
	}

	swIfIndex, err := h.gtpuAddDelTunnel(1, gtpuLink, multicastIf)
	if err != nil {
		return 0, err
	}
	return swIfIndex, h.SetInterfaceTag(ifName, swIfIndex)
}

// DelGtpuTunnel removes GTPU interface.
func (h *InterfaceVppHandler) DelGtpuTunnel(ifName string, gtpuLink *ifs.GtpuLink) error {
    if gtpuLink == nil {
        return errors.New("Missing GTPU tunnel information")
    }
    if h.gtpu == nil {
		return vpp.ErrPluginDisabled
	}

	swIfIndex, err := h.gtpuAddDelTunnel(0, gtpuLink, 0)
	if err != nil {
		return err
	}
	return h.RemoveInterfaceTag(ifName, swIfIndex)
}

// dumpGtpuDetails dumps GTP-U interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpGtpuDetails(interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	if h.gtpu == nil {
		// no-op when disabled
		return nil
	}

	reqCtx := h.callsChannel.SendMultiRequest(&vpp_gtpu.GtpuTunnelDump{
		SwIfIndex: ^uint32(0),
	})
	for {
		gtpuDetails := &vpp_gtpu.GtpuTunnelDetails{}
		stop, err := reqCtx.ReceiveReply(gtpuDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump GTP-U tunnel interface details: %v", err)
		}
		_, ifIdxExists := interfaces[gtpuDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		// Multicast interface
		var multicastIfName string
		_, exists := interfaces[gtpuDetails.McastSwIfIndex]
		if exists {
			multicastIfName = interfaces[gtpuDetails.McastSwIfIndex].Interface.Name
		}

		gtpu := &ifs.GtpuLink{
			Multicast:  multicastIfName,
			EncapVrfId: gtpuDetails.EncapVrfID,
			Teid:       gtpuDetails.Teid,
			DecapNext:  ifs.GtpuLink_NextNode(gtpuDetails.DecapNextIndex),
		}

		if gtpuDetails.IsIPv6 == 1 {
			gtpu.SrcAddr = net.IP(gtpuDetails.SrcAddress).To16().String()
			gtpu.DstAddr = net.IP(gtpuDetails.DstAddress).To16().String()
		} else {
			gtpu.SrcAddr = net.IP(gtpuDetails.SrcAddress[:4]).To4().String()
			gtpu.DstAddr = net.IP(gtpuDetails.DstAddress[:4]).To4().String()
		}

		interfaces[gtpuDetails.SwIfIndex].Interface.Link = &ifs.Interface_Gtpu{Gtpu: gtpu}
		interfaces[gtpuDetails.SwIfIndex].Interface.Type = ifs.Interface_GTPU_TUNNEL
	}

	return nil
}
