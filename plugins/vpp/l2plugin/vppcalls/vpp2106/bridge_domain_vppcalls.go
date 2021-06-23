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
	vpp_l2 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/l2"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

// AddBridgeDomain adds new bridge domain.
func (h *BridgeDomainVppHandler) AddBridgeDomain(bdIdx uint32, bd *l2.BridgeDomain) error {
	req := &vpp_l2.BridgeDomainAddDel{
		IsAdd:   true,
		BdID:    bdIdx,
		Learn:   bd.Learn,
		ArpTerm: bd.ArpTermination,
		Flood:   bd.Flood,
		UuFlood: bd.UnknownUnicastFlood,
		Forward: bd.Forward,
		MacAge:  uint8(bd.MacAge),
		BdTag:   bd.Name,
	}
	reply := &vpp_l2.BridgeDomainAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// DeleteBridgeDomain removes existing bridge domain.
func (h *BridgeDomainVppHandler) DeleteBridgeDomain(bdIdx uint32) error {
	req := &vpp_l2.BridgeDomainAddDel{
		IsAdd: false,
		BdID:  bdIdx,
	}
	reply := &vpp_l2.BridgeDomainAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}
