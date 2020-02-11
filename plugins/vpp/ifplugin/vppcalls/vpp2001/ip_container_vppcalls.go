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
	"go.ligato.io/cn-infra/v2/utils/addrs"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interface_types"
	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
)

func (h *InterfaceVppHandler) sendAndLogMessageForVpp(ifIdx uint32, addr string, isAdd bool) error {
	req := &vpp_ip.IPContainerProxyAddDel{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
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
		req.Pfx.Address.Af = ip_types.ADDRESS_IP6
	} else {
		copy(req.Pfx.Address.Un.XXX_UnionData[:], IPaddr.IP.To4())
		req.Pfx.Address.Af = ip_types.ADDRESS_IP4
	}
	reply := &vpp_ip.IPContainerProxyAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceVppHandler) AddContainerIP(ifIdx uint32, addr string) error {
	return h.sendAndLogMessageForVpp(ifIdx, addr, true)
}

func (h *InterfaceVppHandler) DelContainerIP(ifIdx uint32, addr string) error {
	return h.sendAndLogMessageForVpp(ifIdx, addr, false)
}
