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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vxlan"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
)

func addDelVxlanTunnel(iface *intf.Interfaces_Interface_Vxlan, encVrf uint32, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (swIdx uint32, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(vxlan.VxlanAddDelTunnel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// this is temporary fix to solve creation of VRF table for VXLAN
	if err := CreateVrfIfNeeded(encVrf, vppChan); err != nil {
		return 0, err
	}

	req := &vxlan.VxlanAddDelTunnel{
		IsAdd:          boolToUint(isAdd),
		Vni:            iface.Vni,
		DecapNextIndex: 0xFFFFFFFF,
		EncapVrfID:     encVrf,
	}

	srcAddr := net.ParseIP(iface.SrcAddress).To4()
	if srcAddr == nil {
		srcAddr = net.ParseIP(iface.SrcAddress).To16()
		if srcAddr == nil {
			return 0, fmt.Errorf("invalid VXLAN source address")
		}
		req.IsIpv6 = 1
	} else {
		req.IsIpv6 = 0
	}
	req.SrcAddress = []byte(srcAddr)

	var dstAddr net.IP
	if req.IsIpv6 == 0 {
		dstAddr = net.ParseIP(iface.DstAddress).To4()
	} else {
		dstAddr = net.ParseIP(iface.DstAddress).To16()
	}
	if dstAddr == nil {
		return 0, fmt.Errorf("IP version mismatch for VXLAN destination and source IP addresses")
	}
	req.DstAddress = []byte(dstAddr)

	reply := &vxlan.VxlanAddDelTunnelReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return reply.SwIfIndex, nil
}

// AddVxlanTunnel calls AddDelVxlanTunnelReq with flag add=1.
func AddVxlanTunnel(ifName string, vxlanIntf *intf.Interfaces_Interface_Vxlan, encapVrf uint32, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (swIndex uint32, err error) {
	swIfIdx, err := addDelVxlanTunnel(vxlanIntf, encapVrf, true, vppChan, stopwatch)
	if err != nil {
		return 0, err
	}
	return swIfIdx, SetInterfaceTag(ifName, swIfIdx, vppChan, stopwatch)
}

// DeleteVxlanTunnel calls AddDelVxlanTunnelReq with flag add=0.
func DeleteVxlanTunnel(ifName string, idx uint32, vxlanIntf *intf.Interfaces_Interface_Vxlan, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	if _, err := addDelVxlanTunnel(vxlanIntf, 0, false, vppChan, stopwatch); err != nil {
		return err
	}
	return RemoveInterfaceTag(ifName, idx, vppChan, stopwatch)
}
