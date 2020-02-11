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

package vpp1904

import (
	"go.ligato.io/cn-infra/v2/utils/addrs"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/ip"
)

const (
	addContainerIP    uint8 = 1
	removeContainerIP uint8 = 0
)

func (h *InterfaceVppHandler) sendAndLogMessageForVpp(ifIdx uint32, addr string, isAdd uint8) error {
	req := &ip.IPContainerProxyAddDel{
		SwIfIndex: ifIdx,
		IsAdd:     isAdd,
	}

	IPaddr, isIPv6, err := addrs.ParseIPWithPrefix(addr)
	if err != nil {
		return err
	}

	prefix, _ := IPaddr.Mask.Size()
	req.Pfx.AddressLength = byte(prefix)
	if isIPv6 {
		copy(req.Pfx.Address.Un.XXX_UnionData[:], IPaddr.IP.To16())
		req.Pfx.Address.Af = ip.ADDRESS_IP6
	} else {
		copy(req.Pfx.Address.Un.XXX_UnionData[:], IPaddr.IP.To4())
		req.Pfx.Address.Af = ip.ADDRESS_IP4
	}
	reply := &ip.IPContainerProxyAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// AddContainerIP implements interface handler.
func (h *InterfaceVppHandler) AddContainerIP(ifIdx uint32, addr string) error {
	return h.sendAndLogMessageForVpp(ifIdx, addr, addContainerIP)
}

// DelContainerIP implements interface handler.
func (h *InterfaceVppHandler) DelContainerIP(ifIdx uint32, addr string) error {
	return h.sendAndLogMessageForVpp(ifIdx, addr, removeContainerIP)
}
