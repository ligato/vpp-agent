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

// AddDelVxlanTunnelReq prepares the request for bin API calls.
func AddDelVxlanTunnelReq(vxlanIntf *intf.Interfaces_Interface_Vxlan, add uint8) (req *vxlan.VxlanAddDelTunnel, err error) {
	req = &vxlan.VxlanAddDelTunnel{}

	req.IsAdd = add

	address := net.ParseIP(vxlanIntf.SrcAddress).To4()
	if address == nil {
		address = net.ParseIP(vxlanIntf.SrcAddress).To16()
		if nil == address {
			return nil, fmt.Errorf("VXLAN source address is neither IPv4 nor IPv6 address")
		}
		req.IsIpv6 = 1
	} else {
		req.IsIpv6 = 0
	}
	req.SrcAddress = []byte(address)

	if req.IsIpv6 == 0 {
		address = net.ParseIP(vxlanIntf.DstAddress).To4()
	} else {
		address = net.ParseIP(vxlanIntf.DstAddress).To16()
	}
	if address == nil {
		return nil, fmt.Errorf("VXLAN destination and source addresses differ in IP version")
	}
	req.DstAddress = []byte(address)

	req.Vni = vxlanIntf.Vni
	req.DecapNextIndex = 0xFFFFFFFF
	return req, nil
}

// AddVxlanTunnel calls AddDelVxlanTunnelReq with flag add=1.
func AddVxlanTunnel(ifName string, vxlanIntf *intf.Interfaces_Interface_Vxlan, encapVrf uint32, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (swIndex uint32, err error) {
	// VxlanAddDelTunnelReply time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// this is temporary fix to solve creation of VRF table for VXLAN
	if err := CreateVrfIfNeeded(encapVrf, vppChan); err != nil {
		return 0, err
	}

	req, err := AddDelVxlanTunnelReq(vxlanIntf, 1)
	if err != nil {
		return 0, err
	}
	req.EncapVrfID = encapVrf

	reply := &vxlan.VxlanAddDelTunnelReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("add VXLAN tunnel returned %d", reply.Retval)
	}

	return reply.SwIfIndex, SetInterfaceTag(ifName, reply.SwIfIndex, vppChan, timeLog)
}

// DeleteVxlanTunnel calls AddDelVxlanTunnelReq with flag add=0.
func DeleteVxlanTunnel(ifName string, idx uint32, vxlanIntf *intf.Interfaces_Interface_Vxlan, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// VxlanAddDelTunnelReply time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req, err := AddDelVxlanTunnelReq(vxlanIntf, 0)
	if err != nil {
		return err
	}

	reply := &vxlan.VxlanAddDelTunnelReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("deleting of VXLAN tunnel returned %d", reply.Retval)
	}

	return RemoveInterfaceTag(ifName, idx, vppChan, timeLog)
}
