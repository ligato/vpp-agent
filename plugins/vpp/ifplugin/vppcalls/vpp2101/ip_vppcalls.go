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
	"context"
	"net"

	"go.ligato.io/cn-infra/v2/utils/addrs"

	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/interface_types"
)

func (h *InterfaceVppHandler) addDelInterfaceIP(ifIdx uint32, addr *net.IPNet, isAdd bool) error {
	req := &vpp_ifs.SwInterfaceAddDelAddress{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
		IsAdd:     isAdd,
	}

	isIPv6, err := addrs.IsIPv6(addr.IP.String())
	if err != nil {
		return err
	}
	req.Prefix.Address = ipToAddress(addr, isIPv6)
	prefix, _ := addr.Mask.Size()
	req.Prefix.Len = byte(prefix)

	reply := &vpp_ifs.SwInterfaceAddDelAddressReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceVppHandler) AddInterfaceIP(ifIdx uint32, addr *net.IPNet) error {
	return h.addDelInterfaceIP(ifIdx, addr, true)
}

func (h *InterfaceVppHandler) DelInterfaceIP(ifIdx uint32, addr *net.IPNet) error {
	return h.addDelInterfaceIP(ifIdx, addr, false)
}

func (h *InterfaceVppHandler) setUnsetUnnumberedIP(uIfIdx uint32, ifIdxWithIP uint32, isAdd bool) error {
	req := &vpp_ifs.SwInterfaceSetUnnumbered{
		SwIfIndex:           interface_types.InterfaceIndex(ifIdxWithIP),
		UnnumberedSwIfIndex: interface_types.InterfaceIndex(uIfIdx),
		IsAdd:               isAdd,
	}
	reply := &vpp_ifs.SwInterfaceSetUnnumberedReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceVppHandler) SetUnnumberedIP(ctx context.Context, uIfIdx uint32, ifIdxWithIP uint32) error {
	return h.setUnsetUnnumberedIP(uIfIdx, ifIdxWithIP, true)
}

func (h *InterfaceVppHandler) UnsetUnnumberedIP(ctx context.Context, uIfIdx uint32) error {
	return h.setUnsetUnnumberedIP(uIfIdx, 0, false)
}
