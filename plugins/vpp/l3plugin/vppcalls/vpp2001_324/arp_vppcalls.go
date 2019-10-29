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

	"github.com/pkg/errors"
	vpp_ip "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/ip"
	l3 "go.ligato.io/vpp-agent/v2/proto/ligato/vpp-agent/vpp/l3"
)

// vppAddDelArp adds or removes ARP entry according to provided input
func (h *ArpVppHandler) vppAddDelArp(entry *l3.ARPEntry, delete bool) error {
	meta, found := h.ifIndexes.LookupByName(entry.Interface)
	if !found {
		return errors.Errorf("interface %s not found", entry.Interface)
	}

	var flags vpp_ip.IPNeighborFlags
	flags |= vpp_ip.IP_API_NEIGHBOR_FLAG_NO_FIB_ENTRY
	if entry.Static {
		flags |= vpp_ip.IP_API_NEIGHBOR_FLAG_STATIC
	}

	req := &vpp_ip.IPNeighborAddDel{
		IsAdd: boolToUint(!delete),
		Neighbor: vpp_ip.IPNeighbor{
			SwIfIndex: meta.SwIfIndex,
			Flags:     flags,
		},
	}

	var err error
	req.Neighbor.IPAddress, err = ipToAddress(entry.IpAddress)
	if err != nil {
		return errors.WithStack(err)
	}

	macAddr, err := net.ParseMAC(entry.PhysAddress)
	if err != nil {
		return err
	}
	copy(req.Neighbor.MacAddress[:], macAddr)

	reply := &vpp_ip.IPNeighborAddDelReply{}
	if err = h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// VppAddArp implements arp handler.
func (h *ArpVppHandler) VppAddArp(entry *l3.ARPEntry) error {
	return h.vppAddDelArp(entry, false)
}

// VppDelArp implements arp handler.
func (h *ArpVppHandler) VppDelArp(entry *l3.ARPEntry) error {
	return h.vppAddDelArp(entry, true)
}
