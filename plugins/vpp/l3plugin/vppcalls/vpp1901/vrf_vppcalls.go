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

package vpp1901

import (
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/ip"
)

// AddVrfTable adds new VRF table.
func (h *VrfTableHandler) AddVrfTable(table *l3.VrfTable) error {
	return h.addDelVrfTable(table, true)
}

// DelVrfTable deletes existing VRF table.
func (h *VrfTableHandler) DelVrfTable(table *l3.VrfTable) error {
	return h.addDelVrfTable(table, false)
}

func (h *VrfTableHandler) addDelVrfTable(table *l3.VrfTable, isAdd bool) error {
	req := &ip.IPTableAddDel{
		TableID: table.Id,
		IsIPv6:  boolToUint(table.GetProtocol() == l3.VrfTable_IPV6),
		Name:    []byte(table.Label),
		IsAdd:   boolToUint(isAdd),
	}
	reply := &ip.IPTableAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}
