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

package vpp1908

import (
	l2ba "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1908/l2"
	l2 "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/l2"
)

// AddBridgeDomain adds new bridge domain.
func (h *BridgeDomainVppHandler) AddBridgeDomain(bdIdx uint32, bd *l2.BridgeDomain) error {
	req := &l2ba.BridgeDomainAddDel{
		IsAdd:   1,
		BdID:    bdIdx,
		Learn:   boolToUint(bd.Learn),
		ArpTerm: boolToUint(bd.ArpTermination),
		Flood:   boolToUint(bd.Flood),
		UuFlood: boolToUint(bd.UnknownUnicastFlood),
		Forward: boolToUint(bd.Forward),
		MacAge:  uint8(bd.MacAge),
		BdTag:   []byte(bd.Name),
	}
	reply := &l2ba.BridgeDomainAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// DeleteBridgeDomain removes existing bridge domain.
func (h *BridgeDomainVppHandler) DeleteBridgeDomain(bdIdx uint32) error {
	req := &l2ba.BridgeDomainAddDel{
		IsAdd: 0,
		BdID:  bdIdx,
	}
	reply := &l2ba.BridgeDomainAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func boolToUint(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}
