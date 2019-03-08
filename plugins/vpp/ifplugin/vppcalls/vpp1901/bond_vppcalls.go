package vpp1901

import (
	"net"

	if_model "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/bond"
)

// AddBondInterface implements interface handler.
func (h *InterfaceVppHandler) AddBondInterface(ifName string, bondLink *if_model.BondLink) (uint32, error) {
	req := &bond.BondCreate{
		ID:   bondLink.Id,
		Mode: getBondMode(bondLink.Mode),
		Lb:   getLoadBalance(bondLink.Lb),
	}
	if bondLink.Mac != "" {
		var err error
		req.UseCustomMac = 1
		req.MacAddress, err = net.ParseMAC(bondLink.Mac)
		if err != nil {
			return 0, err
		}
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
