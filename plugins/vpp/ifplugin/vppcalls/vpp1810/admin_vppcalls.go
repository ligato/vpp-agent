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

package vppcalls

import (
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/interfaces"
)

// InterfaceAdminDown implements interface handler.
func (h *InterfaceVppHandler) InterfaceAdminDown(ifIdx uint32) error {
	return h.interfaceSetFlags(ifIdx, false)
}

// InterfaceAdminUp implements interface handler.
func (h *InterfaceVppHandler) InterfaceAdminUp(ifIdx uint32) error {
	return h.interfaceSetFlags(ifIdx, true)
}

// SetInterfaceTag implements interface handler.
func (h *InterfaceVppHandler) SetInterfaceTag(tag string, ifIdx uint32) error {
	return h.handleInterfaceTag(tag, ifIdx, true)
}

// RemoveInterfaceTag implements interface handler.
func (h *InterfaceVppHandler) RemoveInterfaceTag(tag string, ifIdx uint32) error {
	return h.handleInterfaceTag(tag, ifIdx, false)
}

func (h *InterfaceVppHandler) interfaceSetFlags(ifIdx uint32, adminUp bool) error {
	req := &interfaces.SwInterfaceSetFlags{
		SwIfIndex:   ifIdx,
		AdminUpDown: boolToUint(adminUp),
	}
	reply := &interfaces.SwInterfaceSetFlagsReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceVppHandler) handleInterfaceTag(tag string, ifIdx uint32, isAdd bool) error {
	req := &interfaces.SwInterfaceTagAddDel{
		Tag:   []byte(tag),
		IsAdd: boolToUint(isAdd),
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

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}
