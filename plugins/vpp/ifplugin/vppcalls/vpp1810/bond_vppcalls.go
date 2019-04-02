// Copyright (c) 2019 PANTHEON.tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vpp1810

import (
	"net"

	if_model "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/bond"
)

// AddBondInterface implements interface handler.
func (h *InterfaceVppHandler) AddBondInterface(ifName string, mac string, bondLink *if_model.BondLink) (uint32, error) {
	req := &bond.BondCreate{
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

	reply := &bond.BondCreateReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}

	return reply.SwIfIndex, h.SetInterfaceTag(ifName, reply.SwIfIndex)
}

// DeleteBondInterface implements interface handler.
func (h *InterfaceVppHandler) DeleteBondInterface(ifName string, ifIdx uint32) error {
	req := &bond.BondDelete{
		SwIfIndex: ifIdx,
	}
	reply := &bond.BondDeleteReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return h.RemoveInterfaceTag(ifName, ifIdx)
}

func getBondMode(mode if_model.BondLink_Mode) uint8 {
	switch mode {
	case if_model.BondLink_ROUND_ROBIN:
		return 1
	case if_model.BondLink_ACTIVE_BACKUP:
		return 2
	case if_model.BondLink_XOR:
		return 3
	case if_model.BondLink_BROADCAST:
		return 4
	case if_model.BondLink_LACP:
		return 5
	default:
		// UNKNOWN
		return 0
	}
}

// AttachInterfaceToBond implements interface handler.
func (h *InterfaceVppHandler) AttachInterfaceToBond(ifIdx, bondIfIdx uint32, isPassive, isLongTimeout bool) error {
	req := &bond.BondEnslave{
		SwIfIndex:     ifIdx,
		BondSwIfIndex: bondIfIdx,
		IsPassive:     boolToUint(isPassive),
		IsLongTimeout: boolToUint(isLongTimeout),
	}
	reply := &bond.BondEnslaveReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// DetachInterfaceFromBond implements interface handler
func (h *InterfaceVppHandler) DetachInterfaceFromBond(ifIdx uint32) error {
	req := &bond.BondDetachSlave{
		SwIfIndex: ifIdx,
	}
	reply := &bond.BondDetachSlaveReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return nil
}

func getLoadBalance(lb if_model.BondLink_LoadBalance) uint8 {
	switch lb {
	case if_model.BondLink_L34:
		return 1
	case if_model.BondLink_L23:
		return 2
	default:
		// L2
		return 0
	}
}
