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
	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
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
	req := &vpp_ip.IPTableAddDel{
		Table: vpp_ip.IPTable{
			TableID: table.Id,
			IsIP6:   table.GetProtocol() == l3.VrfTable_IPV6,
			Name:    table.Label,
		},
		IsAdd: isAdd,
	}
	reply := &vpp_ip.IPTableAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// SetVrfFlowHashSettings sets IP flow hash settings for a VRF table.
func (h *VrfTableHandler) SetVrfFlowHashSettings(vrfID uint32, isIPv6 bool, hashFields *l3.VrfTable_FlowHashSettings) error {
	req := &vpp_ip.SetIPFlowHash{
		VrfID:     vrfID,
		IsIPv6:    isIPv6,
		Src:       hashFields.UseSrcIp,
		Dst:       hashFields.UseDstIp,
		Sport:     hashFields.UseSrcPort,
		Dport:     hashFields.UseDstPort,
		Proto:     hashFields.UseProtocol,
		Reverse:   hashFields.Reverse,
		Symmetric: hashFields.Symmetric,
	}
	reply := &vpp_ip.SetIPFlowHashReply{}

	err := h.callsChannel.SendRequest(req).ReceiveReply(reply)
	return err
}
