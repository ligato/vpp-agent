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
	"fmt"
	"net"
	"time"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vxlan"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
)

func (handler *ifVppHandler) addDelVxlanTunnel(iface *intf.Interfaces_Interface_Vxlan, encVrf, multicastIf uint32, isAdd bool) (swIdx uint32, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(vxlan.VxlanAddDelTunnel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// this is temporary fix to solve creation of VRF table for VXLAN
	if err := handler.CreateVrfIfNeeded(encVrf); err != nil {
		return 0, err
	}

	req := &vxlan.VxlanAddDelTunnel{
		IsAdd:          boolToUint(isAdd),
		Vni:            iface.Vni,
		DecapNextIndex: 0xFFFFFFFF,
		Instance:       ^uint32(0),
		EncapVrfID:     encVrf,
		McastSwIfIndex: multicastIf,
	}

	srcAddr := net.ParseIP(iface.SrcAddress).To4()
	dstAddr := net.ParseIP(iface.DstAddress).To4()
	if srcAddr == nil && dstAddr == nil {
		srcAddr = net.ParseIP(iface.SrcAddress).To16()
		dstAddr = net.ParseIP(iface.DstAddress).To16()
		req.IsIpv6 = 1
		if srcAddr == nil || dstAddr == nil {
			return 0, fmt.Errorf("invalid VXLAN address, src: %s, dst: %s", srcAddr, dstAddr)
		}
	} else if srcAddr == nil && dstAddr != nil || srcAddr != nil && dstAddr == nil {
		return 0, fmt.Errorf("IP version mismatch for VXLAN destination and source IP addresses")
	}

	req.SrcAddress = []byte(srcAddr)
	req.DstAddress = []byte(dstAddr)

	reply := &vxlan.VxlanAddDelTunnelReply{}
	if err = handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return reply.SwIfIndex, nil
}

func (handler *ifVppHandler) AddVxlanTunnel(ifName string, vxlanIntf *intf.Interfaces_Interface_Vxlan, encapVrf, multicastIf uint32) (swIndex uint32, err error) {
	swIfIdx, err := handler.addDelVxlanTunnel(vxlanIntf, encapVrf, multicastIf, true)
	if err != nil {
		return 0, err
	}
	return swIfIdx, handler.SetInterfaceTag(ifName, swIfIdx)
}

func (handler *ifVppHandler) DeleteVxlanTunnel(ifName string, idx uint32, vxlanIntf *intf.Interfaces_Interface_Vxlan) error {
	if _, err := handler.addDelVxlanTunnel(vxlanIntf, 0, 0, false); err != nil {
		return err
	}
	return handler.RemoveInterfaceTag(ifName, idx)
}
