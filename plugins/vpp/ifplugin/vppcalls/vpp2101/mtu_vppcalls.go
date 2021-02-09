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
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/interface_types"
)

func (h *InterfaceVppHandler) SetInterfaceMtu(ifIdx uint32, mtu uint32) error {
	req := &vpp_ifs.HwInterfaceSetMtu{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
		Mtu:       uint16(mtu),
	}
	reply := &vpp_ifs.HwInterfaceSetMtuReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}
