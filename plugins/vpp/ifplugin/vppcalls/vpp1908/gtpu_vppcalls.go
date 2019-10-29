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
	"errors"
	"net"

	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_gtpu "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/gtpu"
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
	swIfIndex, err := h.gtpuAddDelTunnel(1, gtpuLink, multicastIf)
	if err != nil {
		return 0, err
	}
	return swIfIndex, h.SetInterfaceTag(ifName, swIfIndex)
}

// DelGtpuTunnel removes GTPU interface.
func (h *InterfaceVppHandler) DelGtpuTunnel(ifName string, gtpuLink *ifs.GtpuLink) error {
	swIfIndex, err := h.gtpuAddDelTunnel(0, gtpuLink, 0)
	if err != nil {
		return err
	}
	return h.RemoveInterfaceTag(ifName, swIfIndex)
}
