package vpp2001_379

import (
	"errors"
	"net"

	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_gtpu "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001_379/gtpu"
)

func (h *InterfaceVppHandler) gtpuAddDelTunnel(isAdd uint8, gtpuLink *ifs.GtpuLink, multicastIf uint32) (uint32, error) {
	req := &vpp_gtpu.GtpuAddDelTunnel{
        IsAdd:          isAdd,
		McastSwIfIndex: multicastIf,
        EncapVrfID:     gtpuLink.EncapVrfId,
		Teid:           gtpuLink.Teid,
    }

    if gtpuLink.DecapNext == ifs.GtpuLink_DEFAULT {
        req.DecapNextIndex = 0xFFFFFFFF
    } else {
        req.DecapNextIndex = uint32(gtpuLink.DecapNext)
    }

	srcAddr := net.ParseIP(gtpuLink.SrcAddr)
	if srcAddr == nil {
		err := errors.New("bad source address for GTPU tunnel")
		return 0, err
	}

	dstAddr := net.ParseIP(gtpuLink.DstAddr)
	if dstAddr == nil {
		err := errors.New("bad destination address for GTPU tunnel")
		return 0, err
	}

	if gtpuLink.SrcAddr == gtpuLink.DstAddr {
		err := errors.New("source and destination are the same")
		return 0, err
	}

    if srcAddr.To4() != nil && dstAddr.To4() != nil {
        req.IsIPv6 = 0
        req.SrcAddress = []byte(srcAddr.To4())
        req.DstAddress = []byte(dstAddr.To4())
    } else if srcAddr.To16() != nil && dstAddr.To16() != nil {
        req.IsIPv6 = 1
        req.SrcAddress = []byte(srcAddr.To16())
        req.DstAddress = []byte(dstAddr.To16())
    } else {
		return 0, errors.New("source and destination addresses must be both either IPv4 or IPv6")
    }

	reply := &vpp_gtpu.GtpuAddDelTunnelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	return uint32(reply.SwIfIndex), nil
}

// AddGtpuTunnel adds new GTPU interface.
func (h *InterfaceVppHandler) AddGtpuTunnel(ifName string, gtpuLink *ifs.GtpuLink, multicastIf uint32) (uint32, error) {
	swIfIndex, err := h.gtpuAddDelTunnel(1, gtpuLink, multicastIf)
	if err != nil {
		return 0, err
	}
	return swIfIndex, h.SetInterfaceTag(ifName, swIfIndex)
}

// DelGtpuTunnel removes GTPU interface.
func (h *InterfaceVppHandler) DelGtpuTunnel(ifName string, gtpuLink *ifs.GtpuLink) error {
	swIfIndex, err := h.gtpuAddDelTunnel(0, gtpuLink, 0)
	if err != nil {
		return err
	}
	return h.RemoveInterfaceTag(ifName, swIfIndex)
}
