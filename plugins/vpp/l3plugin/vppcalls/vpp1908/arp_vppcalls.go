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

package vpp1908

import (
	"net"

	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/ip"
	"github.com/pkg/errors"
)

// vppAddDelArp adds or removes ARP entry according to provided input
func (h *ArpVppHandler) vppAddDelArp(entry *l3.ARPEntry, delete bool) error {
	meta, found := h.ifIndexes.LookupByName(entry.Interface)
	if !found {
		return errors.Errorf("interface %s not found", entry.Interface)
	}

	var flags ip.IPNeighborFlags
	flags |= ip.IP_API_NEIGHBOR_FLAG_NO_FIB_ENTRY
	if entry.Static {
		flags |= ip.IP_API_NEIGHBOR_FLAG_STATIC
	}

	req := &ip.IPNeighborAddDel{
		IsAdd: boolToUint(!delete),
		Neighbor: ip.IPNeighbor{
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

	reply := &ip.IPNeighborAddDelReply{}
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
