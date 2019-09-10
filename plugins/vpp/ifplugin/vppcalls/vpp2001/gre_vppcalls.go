package vpp2001

import (
	"errors"
	"net"

	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_gre "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001/gre"
)

func (h *InterfaceVppHandler) greAddDelTunnel(isAdd bool, greLink *ifs.GreLink) (uint32, error) {
	if greLink.TunnelType == ifs.GreLink_UNKNOWN {
		err := errors.New("bad GRE tunnel type")
		return 0, err
	}

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

	if greLink.TunnelType == ifs.GreLink_ERSPAN && greLink.SessionId > 1023 {
		err := errors.New("set session id for ERSPAN tunnel type")
		return 0, err
	}

	var tt vpp_gre.GreTunnelType
	switch uint8(greLink.TunnelType - 1) {
	case 0:
		tt = vpp_gre.GRE_API_TUNNEL_TYPE_L3
	case 1:
		tt = vpp_gre.GRE_API_TUNNEL_TYPE_TEB
	case 2:
		tt = vpp_gre.GRE_API_TUNNEL_TYPE_ERSPAN
	default:
		return 0, errors.New("bad GRE tunnel type")
	}

	tunnel := vpp_gre.GreTunnel{
		Type:       tt,
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
		var src, dst [16]uint8
		copy(src[:], []byte(greSource.To16()))
		copy(dst[:], []byte(greDestination.To16()))
		tunnel.Src = vpp_gre.Address{
			Af: vpp_gre.ADDRESS_IP6,
			Un: vpp_gre.AddressUnionIP6(vpp_gre.IP6Address(src)),
		}
		tunnel.Dst = vpp_gre.Address{
			Af: vpp_gre.ADDRESS_IP6,
			Un: vpp_gre.AddressUnionIP6(vpp_gre.IP6Address(dst)),
		}
	} else {
		var src, dst [4]uint8
		copy(src[:], []byte(greSource.To4()))
		copy(dst[:], []byte(greDestination.To4()))
		tunnel.Src = vpp_gre.Address{
			Af: vpp_gre.ADDRESS_IP4,
			Un: vpp_gre.AddressUnionIP4(vpp_gre.IP4Address(src)),
		}
		tunnel.Dst = vpp_gre.Address{
			Af: vpp_gre.ADDRESS_IP4,
			Un: vpp_gre.AddressUnionIP4(vpp_gre.IP4Address(dst)),
		}
	}

	req := &vpp_gre.GreTunnelAddDel{
		IsAdd:  isAdd,
		Tunnel: tunnel,
	}
	reply := &vpp_gre.GreTunnelAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	return uint32(reply.SwIfIndex), nil
}

// AddGreTunnel adds new GRE interface.
func (h *InterfaceVppHandler) AddGreTunnel(ifName string, greLink *ifs.GreLink) (uint32, error) {
	swIfIndex, err := h.greAddDelTunnel(true, greLink)
	if err != nil {
		return 0, err
	}
	return swIfIndex, h.SetInterfaceTag(ifName, swIfIndex)
}

// DelGreTunnel removes GRE interface.
func (h *InterfaceVppHandler) DelGreTunnel(ifName string, greLink *ifs.GreLink) (uint32, error) {
	swIfIndex, err := h.greAddDelTunnel(false, greLink)
	if err != nil {
		return 0, err
	}
	return swIfIndex, h.RemoveInterfaceTag(ifName, swIfIndex)
}
