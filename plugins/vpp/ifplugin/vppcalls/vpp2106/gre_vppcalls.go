package vpp2106

import (
	"errors"
	"net"

	vpp_gre "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/gre"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/tunnel_types"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
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
	switch greLink.TunnelType {
	case ifs.GreLink_L3:
		tt = vpp_gre.GRE_API_TUNNEL_TYPE_L3
	case ifs.GreLink_TEB:
		tt = vpp_gre.GRE_API_TUNNEL_TYPE_TEB
	case ifs.GreLink_ERSPAN:
		tt = vpp_gre.GRE_API_TUNNEL_TYPE_ERSPAN
	default:
		return 0, errors.New("bad GRE tunnel type")
	}

	tunnel := vpp_gre.GreTunnel{
		Type:         tt,
		Mode:         tunnel_types.TUNNEL_API_MODE_P2P, // TODO: add mode to proto model
		Instance:     ^uint32(0),
		OuterTableID: greLink.OuterFibId,
		SessionID:    uint16(greLink.SessionId),
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
		copy(src[:], greSource.To16())
		copy(dst[:], greDestination.To16())
		tunnel.Src = ip_types.Address{
			Af: ip_types.ADDRESS_IP6,
			Un: ip_types.AddressUnionIP6(src),
		}
		tunnel.Dst = ip_types.Address{
			Af: ip_types.ADDRESS_IP6,
			Un: ip_types.AddressUnionIP6(dst),
		}
	} else {
		var src, dst [4]uint8
		copy(src[:], greSource.To4())
		copy(dst[:], greDestination.To4())
		tunnel.Src = ip_types.Address{
			Af: ip_types.ADDRESS_IP4,
			Un: ip_types.AddressUnionIP4(src),
		}
		tunnel.Dst = ip_types.Address{
			Af: ip_types.ADDRESS_IP4,
			Un: ip_types.AddressUnionIP4(dst),
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
