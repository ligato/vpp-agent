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
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
)

// CreateSubif creates sub interface.
func (h *InterfaceVppHandler) CreateSubif(ifIdx, vlanID uint32) (uint32, error) {
	req := &vpp_ifs.CreateVlanSubif{
		SwIfIndex: vpp_ifs.InterfaceIndex(ifIdx),
		VlanID:    vlanID,
	}

	reply := &vpp_ifs.CreateVlanSubifReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}

	return uint32(reply.SwIfIndex), nil
}

// DeleteSubif deletes sub interface.
func (h *InterfaceVppHandler) DeleteSubif(ifIdx uint32) error {
	req := &vpp_ifs.DeleteSubif{
		SwIfIndex: vpp_ifs.InterfaceIndex(ifIdx),
	}

	reply := &vpp_ifs.DeleteSubifReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}
