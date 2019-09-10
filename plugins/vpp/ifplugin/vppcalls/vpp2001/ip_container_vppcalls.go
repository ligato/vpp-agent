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
	"github.com/ligato/cn-infra/utils/addrs"
	vpp_ip "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001/ip"
)

const (
	addContainerIP    uint8 = 1
	removeContainerIP uint8 = 0
)

func (h *InterfaceVppHandler) sendAndLogMessageForVpp(ifIdx uint32, addr string, isAdd uint8) error {
	req := &vpp_ip.IPContainerProxyAddDel{
		SwIfIndex: ifIdx,
		IsAdd:     isAdd,
	}

	IPaddr, isIPv6, err := addrs.ParseIPWithPrefix(addr)
	if err != nil {
		return err
	}

	prefix, _ := IPaddr.Mask.Size()
	req.Pfx.Len = byte(prefix)
	if isIPv6 {
		copy(req.Pfx.Address.Un.XXX_UnionData[:], IPaddr.IP.To16())
		req.Pfx.Address.Af = vpp_ip.ADDRESS_IP6
	} else {
		copy(req.Pfx.Address.Un.XXX_UnionData[:], IPaddr.IP.To4())
		req.Pfx.Address.Af = vpp_ip.ADDRESS_IP4
	}
	reply := &vpp_ip.IPContainerProxyAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceVppHandler) AddContainerIP(ifIdx uint32, addr string) error {
	return h.sendAndLogMessageForVpp(ifIdx, addr, addContainerIP)
}

func (h *InterfaceVppHandler) DelContainerIP(ifIdx uint32, addr string) error {
	return h.sendAndLogMessageForVpp(ifIdx, addr, removeContainerIP)
}
