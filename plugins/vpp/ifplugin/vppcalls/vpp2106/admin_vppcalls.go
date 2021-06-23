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

package vpp2106

import (
	"context"

	interfaces "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
)

func (h *InterfaceVppHandler) InterfaceAdminDown(ctx context.Context, ifIdx uint32) error {
	return h.interfaceSetFlags(ifIdx, false)
}

func (h *InterfaceVppHandler) InterfaceAdminUp(ctx context.Context, ifIdx uint32) error {
	return h.interfaceSetFlags(ifIdx, true)
}

func (h *InterfaceVppHandler) SetInterfaceTag(tag string, ifIdx uint32) error {
	return h.handleInterfaceTag(tag, interface_types.InterfaceIndex(ifIdx), true)
}

func (h *InterfaceVppHandler) RemoveInterfaceTag(tag string, ifIdx uint32) error {
	return h.handleInterfaceTag(tag, interface_types.InterfaceIndex(ifIdx), false)
}

func (h *InterfaceVppHandler) interfaceSetFlags(ifIdx uint32, adminUp bool) error {
	req := &interfaces.SwInterfaceSetFlags{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
		Flags:     setAdminUpFlag(adminUp),
	}
	reply := &interfaces.SwInterfaceSetFlagsReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceVppHandler) handleInterfaceTag(tag string, ifIdx interface_types.InterfaceIndex, isAdd bool) error {
	req := &interfaces.SwInterfaceTagAddDel{
		Tag:   tag,
		IsAdd: isAdd,
	}
	// For some reason, if deleting tag, the software interface index has to be 0 and only name should be set.
	// Otherwise reply returns with error core -2 (incorrect sw_if_idx)
	if isAdd {
		req.SwIfIndex = ifIdx
	}
	reply := &interfaces.SwInterfaceTagAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func setAdminUpFlag(adminUp bool) interface_types.IfStatusFlags {
	if adminUp {
		return interface_types.IF_STATUS_API_FLAG_ADMIN_UP
	}
	return 0
}
