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

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/vrrp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// DumpVrrpEntries dumps all configured VRRP entries.
func (h *VrrpVppHandler) DumpVrrpEntries() (entries []*vppcalls.VrrpDetails, err error) {
	req := &vrrp.VrrpVrDump{
		SwIfIndex: 0xffffffff, // Send multirequest to get all VRRP entries
	}
	reqCtx := h.callsChannel.SendMultiRequest(req)
	for {
		vrrpDetails := &vrrp.VrrpVrPeerDetails{}
		stop, err := reqCtx.ReceiveReply(vrrpDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		// VRRP interface
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(uint32(vrrpDetails.SwIfIndex))
		if !exists {
			h.log.Warnf("VRRP dump: interface name not found for index %d", vrrpDetails.SwIfIndex)
		}

		ipStrs := make([]string, 0, 0)
		for _, v := range vrrpDetails.PeerAddrs {
			ipStrs = append(ipStrs, net.IP(v.Un.XXX_UnionData[:]).String())
		}

		// VRRP entry
		vrrp := &l3.VRRPEntry{
			Interface: ifName,
			VrId:      uint32(vrrpDetails.VrID),
			Ipv6Flag:  uintToBool(vrrpDetails.IsIPv6),
			Addrs:     ipStrs,
		}
		// VRRP meta
		meta := &vppcalls.VrrpMeta{
			SwIfIndex: uint32(vrrpDetails.SwIfIndex),
		}

		entries = append(entries, &vppcalls.VrrpDetails{
			Vrrp: vrrp,
			Meta: meta,
		})
	}

	return entries, nil
}
