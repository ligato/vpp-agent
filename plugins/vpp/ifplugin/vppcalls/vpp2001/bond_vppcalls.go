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
	"net"

	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_bond "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001/bond"
)

func (h *InterfaceVppHandler) AddBondInterface(ifName string, mac string, bondLink *ifs.BondLink) (uint32, error) {
	req := &vpp_bond.BondCreate{
		ID:   bondLink.Id,
		Mode: getBondMode(bondLink.Mode),
		Lb:   getLoadBalance(bondLink.Lb),
	}
	if mac != "" {
		parsedMac, err := net.ParseMAC(mac)
		if err != nil {
			return 0, err
		}
		req.UseCustomMac = 1
		req.MacAddress = parsedMac
	}

	reply := &vpp_bond.BondCreateReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}

	return reply.SwIfIndex, h.SetInterfaceTag(ifName, reply.SwIfIndex)
}

func (h *InterfaceVppHandler) DeleteBondInterface(ifName string, ifIdx uint32) error {
	req := &vpp_bond.BondDelete{
		SwIfIndex: ifIdx,
	}
	reply := &vpp_bond.BondDeleteReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return h.RemoveInterfaceTag(ifName, ifIdx)
}

func getBondMode(mode ifs.BondLink_Mode) uint8 {
	switch mode {
	case ifs.BondLink_ROUND_ROBIN:
		return 1
	case ifs.BondLink_ACTIVE_BACKUP:
		return 2
	case ifs.BondLink_XOR:
		return 3
	case ifs.BondLink_BROADCAST:
		return 4
	case ifs.BondLink_LACP:
		return 5
	default:
		// UNKNOWN
		return 0
	}
}

func (h *InterfaceVppHandler) AttachInterfaceToBond(ifIdx, bondIfIdx uint32, isPassive, isLongTimeout bool) error {
	req := &vpp_bond.BondEnslave{
		SwIfIndex:     ifIdx,
		BondSwIfIndex: bondIfIdx,
		IsPassive:     boolToUint(isPassive),
		IsLongTimeout: boolToUint(isLongTimeout),
	}
	reply := &vpp_bond.BondEnslaveReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceVppHandler) DetachInterfaceFromBond(ifIdx uint32) error {
	req := &vpp_bond.BondDetachSlave{
		SwIfIndex: ifIdx,
	}
	reply := &vpp_bond.BondDetachSlaveReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return nil
}

func getLoadBalance(lb ifs.BondLink_LoadBalance) uint8 {
	switch lb {
	case ifs.BondLink_L34:
		return 1
	case ifs.BondLink_L23:
		return 2
	default:
		// L2
		return 0
	}
}

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}
