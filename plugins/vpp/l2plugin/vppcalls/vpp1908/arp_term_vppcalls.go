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
	"net"

	l2ba "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/l2"
)

func (h *BridgeDomainVppHandler) callBdIPMacAddDel(isAdd bool, bdID uint32, mac string, ip string) error {
	ipAddr, err := ipToAddress(ip)
	if err != nil {
		return err
	}
	macAddr, err := net.ParseMAC(mac)
	if err != nil {
		return err
	}
	bdEntry := l2ba.BdIPMac{
		BdID: bdID,
		IP:   ipAddr,
	}
	copy(bdEntry.Mac[:], macAddr)

	req := &l2ba.BdIPMacAddDel{
		IsAdd: boolToUint(isAdd),
		Entry: bdEntry,
	}

	reply := &l2ba.BdIPMacAddDelReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// AddArpTerminationTableEntry creates ARP termination entry for bridge domain.
func (h *BridgeDomainVppHandler) AddArpTerminationTableEntry(bdID uint32, mac string, ip string) error {
	err := h.callBdIPMacAddDel(true, bdID, mac, ip)
	if err != nil {
		return err
	}
	return nil
}

// RemoveArpTerminationTableEntry removes ARP termination entry from bridge domain.
func (h *BridgeDomainVppHandler) RemoveArpTerminationTableEntry(bdID uint32, mac string, ip string) error {
	err := h.callBdIPMacAddDel(false, bdID, mac, ip)
	if err != nil {
		return err
	}
	return nil
}
