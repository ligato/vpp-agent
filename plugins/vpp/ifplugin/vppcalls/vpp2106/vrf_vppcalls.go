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
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
)

func (h *InterfaceVppHandler) SetInterfaceVrf(ifIdx, vrfID uint32) error {
	return h.setInterfaceVrf(ifIdx, vrfID, false)
}

func (h *InterfaceVppHandler) SetInterfaceVrfIPv6(ifIdx, vrfID uint32) error {
	return h.setInterfaceVrf(ifIdx, vrfID, true)
}

func (h *InterfaceVppHandler) GetInterfaceVrf(ifIdx uint32) (vrfID uint32, err error) {
	return h.getInterfaceVrf(ifIdx, false)
}

func (h *InterfaceVppHandler) GetInterfaceVrfIPv6(ifIdx uint32) (vrfID uint32, err error) {
	return h.getInterfaceVrf(ifIdx, true)
}

// Interface is set to VRF table. Table IP version has to be defined.
func (h *InterfaceVppHandler) setInterfaceVrf(ifIdx, vrfID uint32, isIPv6 bool) error {
	req := &vpp_ifs.SwInterfaceSetTable{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
		VrfID:     vrfID,
		IsIPv6:    isIPv6,
	}
	reply := &vpp_ifs.SwInterfaceSetTableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	h.log.Debugf("Interface %d set to VRF %d", ifIdx, vrfID)

	return nil
}

// Returns VRF ID for provided interface.
func (h *InterfaceVppHandler) getInterfaceVrf(ifIdx uint32, isIPv6 bool) (vrfID uint32, err error) {
	req := &vpp_ifs.SwInterfaceGetTable{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
		IsIPv6:    isIPv6,
	}
	reply := &vpp_ifs.SwInterfaceGetTableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}

	return reply.VrfID, nil
}
