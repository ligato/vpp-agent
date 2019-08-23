package vpp1908

import (
	"errors"
	"fmt"
	"net"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/gre"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

func (h *InterfaceVppHandler) greAddDelTunnel(isAdd uint8, greLink *interfaces.GreLink) (uint32, error) {
	greSource := net.ParseIP(greLink.SrcAddr)
	if greSource == nil {
		err := errors.New("bad source address for GRE tunnel")
		return 0, err
	}
	greDestination := net.ParseIP(greLink.DstAddr)
	if greDestination == nil {
		err := errors.New("bad destination address for GRE tunnel")
		return 0, err
	}

	if greLink.SrcAddr == greLink.DstAddr {
		err := errors.New("source and destination are the same")
		return 0, err
	}

	if greLink.TunnelType == interfaces.GreLink_ERSPAN && greLink.SessionId > 1023 {
		err := errors.New("set session id for ERSPAN tunnel type")
		return 0, err
	}
	req := &gre.GreAddDelTunnel{
		IsAdd:      isAdd,
		TunnelType: uint8(greLink.TunnelType),
		Instance:   ^uint32(0),
		OuterFibID: greLink.OuterFibId,
		SessionID:  uint16(greLink.SessionId),
	}

	var isSrcIPv6, isDstIPv6 bool

	if greSource.To4() == nil {
		isSrcIPv6 = true
	}
	if greDestination.To4() == nil {
		isDstIPv6 = true
	}
	if isSrcIPv6 != isDstIPv6 {
		return 0, errors.New("source and destination addresses must be both either in IPv4 or IPv6")
	}

	if isSrcIPv6 {
		req.SrcAddress = []byte(greSource.To16())
		req.DstAddress = []byte(greDestination.To16())
		req.IsIPv6 = 1
	} else {
		req.SrcAddress = []byte(greSource.To4())
		req.DstAddress = []byte(greDestination.To4())
	}

	reply := &gre.GreAddDelTunnelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	return reply.SwIfIndex, nil
}

// AddGreTunnel adds new GRE interface.
func (h *InterfaceVppHandler) AddGreTunnel(ifName string, greLink *interfaces.GreLink) (uint32, error) {
	swIfIndex, err := h.greAddDelTunnel(1, greLink)
	if err != nil {
		return 0, err
	}
	return swIfIndex, h.SetInterfaceTag(ifName, swIfIndex)
}

// DelGreTunnel removes GRE interface.
func (h *InterfaceVppHandler) DelGreTunnel(ifName string, greLink *interfaces.GreLink) (uint32, error) {
	swIfIndex, err := h.greAddDelTunnel(0, greLink)
	if err != nil {
		return 0, err
	}
	return swIfIndex, h.RemoveInterfaceTag(ifName, swIfIndex)
}

// DumpGre dumps GRE interface.
func (h *InterfaceVppHandler) DumpGre(ifIdx uint32) ([]*vppcalls.GreTunnelDetails, error) {
	var gres []*vppcalls.GreTunnelDetails
	reqCtx := h.callsChannel.SendMultiRequest(&gre.GreTunnelDump{
		SwIfIndex: ifIdx,
	})

	for {
		greDetails := &gre.GreTunnelDetails{}
		stop, err := reqCtx.ReceiveReply(greDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dump span: %v", err)
		}

		var srcAddr, dstAddr net.IP
		if greDetails.IsIPv6 == 1 {
			srcAddr = net.IP(greDetails.SrcAddress)
			dstAddr = net.IP(greDetails.DstAddress)
		} else {
			srcAddr = net.IP(greDetails.SrcAddress[:4])
			dstAddr = net.IP(greDetails.DstAddress[:4])
		}

		gre := &vppcalls.GreTunnelDetails{
			SwIfIndex:  greDetails.SwIfIndex,
			Instance:   greDetails.Instance,
			TunnelType: greDetails.TunnelType,
			SrcAddress: srcAddr,
			DstAddress: dstAddr,
			OuterFibID: greDetails.OuterFibID,
			SessionID:  greDetails.SessionID,
		}
		gres = append(gres, gre)
	}
	return gres, nil
}
