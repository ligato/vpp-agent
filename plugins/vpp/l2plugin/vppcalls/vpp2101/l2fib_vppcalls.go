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

package vpp2101

import (
	"errors"
	"net"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/ethernet_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/interface_types"
	vpp_l2 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/l2"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

// AddL2FIB creates L2 FIB table entry.
func (h *FIBVppHandler) AddL2FIB(fib *l2.FIBEntry) error {
	return h.l2fibAddDel(fib, true)
}

// DeleteL2FIB removes existing L2 FIB table entry.
func (h *FIBVppHandler) DeleteL2FIB(fib *l2.FIBEntry) error {
	return h.l2fibAddDel(fib, false)
}

func (h *FIBVppHandler) l2fibAddDel(fib *l2.FIBEntry, isAdd bool) (err error) {
	// get bridge domain metadata
	bdMeta, found := h.bdIndexes.LookupByName(fib.BridgeDomain)
	if !found {
		return errors.New("failed to get bridge domain metadata")
	}

	// get outgoing interface index
	swIfIndex := ^uint32(0) // ~0 is used by DROP entries
	if fib.Action == l2.FIBEntry_FORWARD {
		ifaceMeta, found := h.ifIndexes.LookupByName(fib.OutgoingInterface)
		if !found {
			return errors.New("failed to get interface metadata")
		}
		swIfIndex = ifaceMeta.GetIndex()
	}

	// parse MAC address
	var mac []byte
	if fib.PhysAddress != "" {
		mac, err = net.ParseMAC(fib.PhysAddress)
		if err != nil {
			return err
		}
	}

	var macAddr ethernet_types.MacAddress
	copy(macAddr[:], mac)

	// add L2 FIB
	req := &vpp_l2.L2fibAddDel{
		IsAdd:     isAdd,
		Mac:       macAddr,
		BdID:      bdMeta.GetIndex(),
		SwIfIndex: interface_types.InterfaceIndex(swIfIndex),
		BviMac:    fib.BridgedVirtualInterface,
		StaticMac: fib.StaticConfig,
		FilterMac: fib.Action == l2.FIBEntry_DROP,
	}
	reply := &vpp_l2.L2fibAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}
