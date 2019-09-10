package vpp2001

import (
	"fmt"
	"net"

	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_gpe "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001/vxlan_gpe"
)

func (h *InterfaceVppHandler) vxLanGpeAddDelTunnel(isAdd uint8, vxLan *ifs.VxlanLink, vrf, multicastIf uint32) (uint32, error) {
	req := &vpp_gpe.VxlanGpeAddDelTunnel{
		McastSwIfIndex: multicastIf,
		EncapVrfID:     vrf,
		DecapVrfID:     vxLan.Gpe.DecapVrfId,
		Protocol:       uint8(vxLan.Gpe.Protocol),
		Vni:            vxLan.Vni,
		IsAdd:          isAdd,
	}

	if vxLan.SrcAddress == vxLan.DstAddress {
		return 0, fmt.Errorf("source and destination addresses must not be the same")
	}
	srcAddr := net.ParseIP(vxLan.SrcAddress).To4()
	dstAddr := net.ParseIP(vxLan.DstAddress).To4()
	if srcAddr == nil && dstAddr == nil {
		srcAddr = net.ParseIP(vxLan.SrcAddress).To16()
		dstAddr = net.ParseIP(vxLan.DstAddress).To16()
		req.IsIPv6 = 1
		if srcAddr == nil || dstAddr == nil {
			return 0, fmt.Errorf("invalid VXLAN address, src: %s, dst: %s", srcAddr, dstAddr)
		}
	} else if srcAddr == nil && dstAddr != nil || srcAddr != nil && dstAddr == nil {
		return 0, fmt.Errorf("IP version mismatch for VXLAN destination and source IP addresses")
	}

	req.Local = []byte(srcAddr)
	req.Remote = []byte(dstAddr)

	reply := &vpp_gpe.VxlanGpeAddDelTunnelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	return reply.SwIfIndex, nil
}

// AddVxLanGpeTunnel adds new VxLAN-GPE interface.
func (h *InterfaceVppHandler) AddVxLanGpeTunnel(ifName string, vrf, multicastIf uint32, vxLan *ifs.VxlanLink) (uint32, error) {
	swIfIndex, err := h.vxLanGpeAddDelTunnel(1, vxLan, vrf, multicastIf)
	if err != nil {
		return 0, err
	}
	return swIfIndex, h.SetInterfaceTag(ifName, swIfIndex)
}

// DeleteVxLanGpeTunnel removes VxLAN-GPE interface.
func (h *InterfaceVppHandler) DeleteVxLanGpeTunnel(ifName string, vxLan *ifs.VxlanLink) error {
	swIfIndex, err := h.vxLanGpeAddDelTunnel(0, vxLan, 0, 0)
	if err != nil {
		return err
	}
	return h.RemoveInterfaceTag(ifName, swIfIndex)
}
