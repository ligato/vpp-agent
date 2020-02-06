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
	vpp_bond "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/bond"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func (h *InterfaceVppHandler) AddBondInterface(ifName string, mac string, bondLink *ifs.BondLink) (uint32, error) {
	req := &vpp_bond.BondCreate{
		ID:   bondLink.Id,
		Mode: getBondMode(bondLink.Mode),
		Lb:   getLoadBalance(bondLink.Lb),
	}
	if mac != "" {
		parsedMac, err := ParseMAC(mac)
		if err != nil {
			return 0, err
		}
		req.UseCustomMac = true
		req.MacAddress = parsedMac
	}

	reply := &vpp_bond.BondCreateReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}

	return uint32(reply.SwIfIndex), h.SetInterfaceTag(ifName, uint32(reply.SwIfIndex))
}

func (h *InterfaceVppHandler) DeleteBondInterface(ifName string, ifIdx uint32) error {
	req := &vpp_bond.BondDelete{
		SwIfIndex: vpp_bond.InterfaceIndex(ifIdx),
	}
	reply := &vpp_bond.BondDeleteReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return h.RemoveInterfaceTag(ifName, ifIdx)
}

func getBondMode(mode ifs.BondLink_Mode) vpp_bond.BondMode {
	switch mode {
	case ifs.BondLink_ROUND_ROBIN:
		return vpp_bond.BOND_API_MODE_ROUND_ROBIN
	case ifs.BondLink_ACTIVE_BACKUP:
		return vpp_bond.BOND_API_MODE_ACTIVE_BACKUP
	case ifs.BondLink_XOR:
		return vpp_bond.BOND_API_MODE_XOR
	case ifs.BondLink_BROADCAST:
		return vpp_bond.BOND_API_MODE_BROADCAST
	case ifs.BondLink_LACP:
		return vpp_bond.BOND_API_MODE_LACP
	default:
		// UNKNOWN
		return 0
	}
}

func (h *InterfaceVppHandler) AttachInterfaceToBond(ifIdx, bondIfIdx uint32, isPassive, isLongTimeout bool) error {
	req := &vpp_bond.BondEnslave{
		SwIfIndex:     vpp_bond.InterfaceIndex(ifIdx),
		BondSwIfIndex: vpp_bond.InterfaceIndex(bondIfIdx),
		IsPassive:     isPassive,
		IsLongTimeout: isLongTimeout,
	}
	reply := &vpp_bond.BondEnslaveReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceVppHandler) DetachInterfaceFromBond(ifIdx uint32) error {
	req := &vpp_bond.BondDetachSlave{
		SwIfIndex: vpp_bond.InterfaceIndex(ifIdx),
	}
	reply := &vpp_bond.BondDetachSlaveReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return nil
}

func getLoadBalance(lb ifs.BondLink_LoadBalance) vpp_bond.BondLbAlgo {
	switch lb {
	case ifs.BondLink_L34:
		return vpp_bond.BOND_API_LB_ALGO_L34
	case ifs.BondLink_L23:
		return vpp_bond.BOND_API_LB_ALGO_L23
	case ifs.BondLink_RR:
		return vpp_bond.BOND_API_LB_ALGO_RR
	case ifs.BondLink_BC:
		return vpp_bond.BOND_API_LB_ALGO_BC
	case ifs.BondLink_AB:
		return vpp_bond.BOND_API_LB_ALGO_AB
	default:
		// L2
		return 0
	}
}
