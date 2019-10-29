//  Copyright (c) 2019 Cisco and/or its affiliates.
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
	"net"

	vpp_ip "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/ip"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v2/proto/ligato/vpp-agent/vpp/l3"
)

// DumpArpEntries implements arp handler.
func (h *ArpVppHandler) DumpArpEntries() ([]*vppcalls.ArpDetails, error) {
	arpV4Entries, err := h.dumpArpEntries(false)
	if err != nil {
		return nil, err
	}
	arpV6Entries, err := h.dumpArpEntries(true)
	if err != nil {
		return nil, err
	}
	return append(arpV4Entries, arpV6Entries...), nil
}

func (h *ArpVppHandler) dumpArpEntries(isIPv6 bool) ([]*vppcalls.ArpDetails, error) {
	var entries []*vppcalls.ArpDetails
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_ip.IPNeighborDump{
		SwIfIndex: 0xffffffff, // Send multirequest to get all ARP entries for given IP version
		IsIPv6:    boolToUint(isIPv6),
	})

	for {
		arpDetails := &vpp_ip.IPNeighborDetails{}
		stop, err := reqCtx.ReceiveReply(arpDetails)
		if stop {
			break
		}
		if err != nil {
			h.log.Error(err)
			return nil, err
		}

		// ARP interface
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(arpDetails.Neighbor.SwIfIndex)
		if !exists {
			h.log.Warnf("ARP dump: interface name not found for index %d", arpDetails.Neighbor.SwIfIndex)
		}
		// IP & MAC address
		var ip string
		if arpDetails.Neighbor.IPAddress.Af == vpp_ip.ADDRESS_IP6 {
			addr := arpDetails.Neighbor.IPAddress.Un.GetIP6()
			ip = net.IP(addr[:]).To16().String()
		} else {
			addr := arpDetails.Neighbor.IPAddress.Un.GetIP4()
			ip = net.IP(addr[:]).To4().String()
		}

		// ARP entry
		arp := &l3.ARPEntry{
			Interface:   ifName,
			IpAddress:   ip,
			PhysAddress: net.HardwareAddr(arpDetails.Neighbor.MacAddress[:]).String(),
			Static:      arpDetails.Neighbor.Flags&vpp_ip.IP_API_NEIGHBOR_FLAG_STATIC != 0,
		}
		// ARP meta
		meta := &vppcalls.ArpMeta{
			SwIfIndex: arpDetails.Neighbor.SwIfIndex,
		}

		entries = append(entries, &vppcalls.ArpDetails{
			Arp:  arp,
			Meta: meta,
		})
	}

	return entries, nil
}
