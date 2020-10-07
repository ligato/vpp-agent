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

package vpp2009

import (
	"net"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/vrrp"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// DumpVrrpEntries dumps all configured VRRP entries.
func (h *VrrpVppHandler) DumpVrrpEntries() (entries []*l3.VRRPEntry, err error) {
	req := &vrrp.VrrpVrDump{
		SwIfIndex: 0xffffffff, // Send multirequest to get all VRRP entries
	}
	reqCtx := h.callsChannel.SendMultiRequest(req)
	for {
		vrrpDetails := &vrrp.VrrpVrDetails{}
		stop, err := reqCtx.ReceiveReply(vrrpDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		// VRRP interface
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(uint32(vrrpDetails.Config.SwIfIndex))
		if !exists {
			h.log.Warnf("VRRP dump: interface name not found for index %d", vrrpDetails.Config.SwIfIndex)
		}

		var isEnabled, isIpv6, isPreempt, isUnicast, isAccept bool
		if vrrpDetails.Runtime.State != vrrp.VRRP_API_VR_STATE_INIT {
			isEnabled = true
		}
		if uintToBool(uint8(vrrpDetails.Config.Flags & vrrp.VRRP_API_VR_PREEMPT)) {
			isPreempt = true
		}
		if uintToBool(uint8(vrrpDetails.Config.Flags & vrrp.VRRP_API_VR_ACCEPT)) {
			isAccept = true
		}
		if uintToBool(uint8(vrrpDetails.Config.Flags & vrrp.VRRP_API_VR_UNICAST)) {
			isUnicast = true
		}
		if uintToBool(uint8(vrrpDetails.Config.Flags & vrrp.VRRP_API_VR_IPV6)) {
			isIpv6 = true
		}

		ipStrs := make([]string, 0, len(vrrpDetails.Addrs))
		for _, v := range vrrpDetails.Addrs {
			if v.Af == ip_types.ADDRESS_IP4 {
				addr := v.Un.GetIP4()
				ipStrs = append(ipStrs, net.IP(addr[:]).To4().String())
			} else {
				addr := v.Un.GetIP6()
				ipStrs = append(ipStrs, net.IP(addr[:]).To16().String())
			}
		}

		// VRRP entry
		vrrp := &l3.VRRPEntry{
			Interface: ifName,
			VrId:      uint32(vrrpDetails.Config.VrID),
			Priority:  uint32(vrrpDetails.Config.Priority),
			Interval:  uint32(vrrpDetails.Config.Interval),
			Ipv6:      isIpv6,
			Preempt:   isPreempt,
			Accept:    isAccept,
			Unicast:   isUnicast,
			Addrs:     ipStrs,
			Enabled:   isEnabled,
		}

		entries = append(entries, vrrp)
	}

	return entries, nil
}
