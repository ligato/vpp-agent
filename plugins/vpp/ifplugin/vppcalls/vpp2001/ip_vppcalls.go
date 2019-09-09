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

package vpp2001

import (
	"net"

	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001/interfaces"
)

func (h *InterfaceVppHandler) addDelInterfaceIP(ifIdx uint32, addr *net.IPNet, isAdd bool) error {
	req := &interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: interfaces.InterfaceIndex(ifIdx),
		IsAdd:     isAdd,
	}

	isIPv6, err := addrs.IsIPv6(addr.IP.String())
	if err != nil {
		return err
	}
	if isIPv6 {
		copy(req.Prefix.Address.Un.XXX_UnionData[:], addr.IP.To16())
		req.Prefix.Address.Af = interfaces.ADDRESS_IP6
	} else {
		copy(req.Prefix.Address.Un.XXX_UnionData[:], addr.IP.To4())
		req.Prefix.Address.Af = interfaces.ADDRESS_IP4
	}
	prefix, _ := addr.Mask.Size()
	req.Prefix.Len = byte(prefix)

	reply := &interfaces.SwInterfaceAddDelAddressReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// AddInterfaceIP implements interface handler.
func (h *InterfaceVppHandler) AddInterfaceIP(ifIdx uint32, addr *net.IPNet) error {
	return h.addDelInterfaceIP(ifIdx, addr, true)
}

// DelInterfaceIP implements interface handler.
func (h *InterfaceVppHandler) DelInterfaceIP(ifIdx uint32, addr *net.IPNet) error {
	return h.addDelInterfaceIP(ifIdx, addr, false)
}

func (h *InterfaceVppHandler) setUnsetUnnumberedIP(uIfIdx uint32, ifIdxWithIP uint32, isAdd bool) error {
	// Prepare the message.
	req := &interfaces.SwInterfaceSetUnnumbered{
		SwIfIndex:           interfaces.InterfaceIndex(ifIdxWithIP),
		UnnumberedSwIfIndex: interfaces.InterfaceIndex(uIfIdx),
		IsAdd:               isAdd,
	}
	reply := &interfaces.SwInterfaceSetUnnumberedReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// SetUnnumberedIP implements interface handler.
func (h *InterfaceVppHandler) SetUnnumberedIP(uIfIdx uint32, ifIdxWithIP uint32) error {
	return h.setUnsetUnnumberedIP(uIfIdx, ifIdxWithIP, true)
}

// UnsetUnnumberedIP implements interface handler.
func (h *InterfaceVppHandler) UnsetUnnumberedIP(uIfIdx uint32) error {
	return h.setUnsetUnnumberedIP(uIfIdx, 0, false)
}
