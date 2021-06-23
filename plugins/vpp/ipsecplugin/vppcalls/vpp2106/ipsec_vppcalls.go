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

package vpp2106

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/utils/addrs"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ipsec_types"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

// AddSPD implements IPSec handler.
func (h *IPSecVppHandler) AddSPD(spdID uint32) error {
	return h.spdAddDel(spdID, true)
}

// DeleteSPD implements IPSec handler.
func (h *IPSecVppHandler) DeleteSPD(spdID uint32) error {
	return h.spdAddDel(spdID, false)
}

// AddSP implements IPSec handler.
func (h *IPSecVppHandler) AddSP(sp *ipsec.SecurityPolicy) error {
	return h.spdAddDelEntry(sp, true)
}

// DeleteSP implements IPSec handler.
func (h *IPSecVppHandler) DeleteSP(sp *ipsec.SecurityPolicy) error {
	return h.spdAddDelEntry(sp, false)
}

// AddSPDInterface implements IPSec handler.
func (h *IPSecVppHandler) AddSPDInterface(spdID uint32, ifaceCfg *ipsec.SecurityPolicyDatabase_Interface) error {
	ifaceMeta, found := h.ifIndexes.LookupByName(ifaceCfg.Name)
	if !found {
		return errors.New("failed to get interface metadata")
	}
	return h.interfaceAddDelSpd(spdID, ifaceMeta.SwIfIndex, true)
}

// DeleteSPDInterface implements IPSec handler.
func (h *IPSecVppHandler) DeleteSPDInterface(spdID uint32, ifaceCfg *ipsec.SecurityPolicyDatabase_Interface) error {
	ifaceMeta, found := h.ifIndexes.LookupByName(ifaceCfg.Name)
	if !found {
		return errors.New("failed to get interface metadata")
	}
	return h.interfaceAddDelSpd(spdID, ifaceMeta.SwIfIndex, false)
}

// AddSA implements IPSec handler.
func (h *IPSecVppHandler) AddSA(sa *ipsec.SecurityAssociation) error {
	return h.sadAddDelEntry(sa, true)
}

// DeleteSA implements IPSec handler.
func (h *IPSecVppHandler) DeleteSA(sa *ipsec.SecurityAssociation) error {
	return h.sadAddDelEntry(sa, false)
}

// AddTunnelProtection implements IPSec handler for adding a tunnel protection.
func (h *IPSecVppHandler) AddTunnelProtection(tp *ipsec.TunnelProtection) error {
	ifaceMeta, found := h.ifIndexes.LookupByName(tp.Interface)
	if !found {
		return errors.New("failed to get interface metadata")
	}
	return h.tunProtectAddUpdateEntry(tp, ifaceMeta.SwIfIndex)
}

// UpdateTunnelProtection implements IPSec handler for updating a tunnel protection.
func (h *IPSecVppHandler) UpdateTunnelProtection(tp *ipsec.TunnelProtection) error {
	ifaceMeta, found := h.ifIndexes.LookupByName(tp.Interface)
	if !found {
		return errors.New("failed to get interface metadata")
	}
	return h.tunProtectAddUpdateEntry(tp, ifaceMeta.SwIfIndex)
}

// DeleteTunnelProtection implements IPSec handler for deleting a tunnel protection.
func (h *IPSecVppHandler) DeleteTunnelProtection(tp *ipsec.TunnelProtection) error {
	ifaceMeta, found := h.ifIndexes.LookupByName(tp.Interface)
	if !found {
		return errors.New("failed to get interface metadata")
	}
	return h.tunProtectDelEntry(tp, ifaceMeta.SwIfIndex)
}

func (h *IPSecVppHandler) spdAddDel(spdID uint32, isAdd bool) error {
	req := &vpp_ipsec.IpsecSpdAddDel{
		IsAdd: isAdd,
		SpdID: spdID,
	}
	reply := &vpp_ipsec.IpsecSpdAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *IPSecVppHandler) spdAddDelEntry(sp *ipsec.SecurityPolicy, isAdd bool) error {
	req := &vpp_ipsec.IpsecSpdEntryAddDel{
		IsAdd: isAdd,
		Entry: vpp_ipsec.IpsecSpdEntry{
			SpdID:           sp.SpdIndex,
			Priority:        sp.Priority,
			IsOutbound:      sp.IsOutbound,
			Protocol:        uint8(sp.Protocol),
			RemotePortStart: uint16(sp.RemotePortStart),
			RemotePortStop:  uint16(sp.RemotePortStop),
			LocalPortStart:  uint16(sp.LocalPortStart),
			LocalPortStop:   uint16(sp.LocalPortStop),
			Policy:          vpp_ipsec.IpsecSpdAction(sp.Action),
			SaID:            sp.SaIndex,
		},
	}
	if req.Entry.RemotePortStop == 0 {
		req.Entry.RemotePortStop = ^req.Entry.RemotePortStop
	}
	if req.Entry.LocalPortStop == 0 {
		req.Entry.LocalPortStop = ^req.Entry.LocalPortStop
	}

	var err error
	req.Entry.RemoteAddressStart, err = IPToAddress(ipOr(sp.RemoteAddrStart, "0.0.0.0"))
	if err != nil {
		return err
	}
	req.Entry.RemoteAddressStop, err = IPToAddress(ipOr(sp.RemoteAddrStop, "255.255.255.255"))
	if err != nil {
		return err
	}
	req.Entry.LocalAddressStart, err = IPToAddress(ipOr(sp.LocalAddrStart, "0.0.0.0"))
	if err != nil {
		return err
	}
	req.Entry.LocalAddressStop, err = IPToAddress(ipOr(sp.LocalAddrStop, "255.255.255.255"))
	if err != nil {
		return err
	}

	reply := &vpp_ipsec.IpsecSpdEntryAddDelReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func ipOr(s, o string) string {
	if s != "" {
		return s
	}
	return o
}

func (h *IPSecVppHandler) interfaceAddDelSpd(spdID, swIfIdx uint32, isAdd bool) error {
	req := &vpp_ipsec.IpsecInterfaceAddDelSpd{
		IsAdd:     isAdd,
		SwIfIndex: interface_types.InterfaceIndex(swIfIdx),
		SpdID:     spdID,
	}
	reply := &vpp_ipsec.IpsecInterfaceAddDelSpdReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *IPSecVppHandler) sadAddDelEntry(sa *ipsec.SecurityAssociation, isAdd bool) error {
	cryptoKey, err := hex.DecodeString(sa.CryptoKey)
	if err != nil {
		return err
	}
	integKey, err := hex.DecodeString(sa.IntegKey)
	if err != nil {
		return err
	}

	var flags ipsec_types.IpsecSadFlags
	if sa.UseEsn {
		flags |= ipsec_types.IPSEC_API_SAD_FLAG_USE_ESN
	}
	if sa.UseAntiReplay {
		flags |= ipsec_types.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY
	}
	if sa.EnableUdpEncap {
		flags |= ipsec_types.IPSEC_API_SAD_FLAG_UDP_ENCAP
	}
	var tunnelSrc, tunnelDst ip_types.Address
	if sa.TunnelSrcAddr != "" {
		flags |= ipsec_types.IPSEC_API_SAD_FLAG_IS_TUNNEL
		isIPv6, err := addrs.IsIPv6(sa.TunnelSrcAddr)
		if err != nil {
			return err
		}
		if isIPv6 {
			flags |= ipsec_types.IPSEC_API_SAD_FLAG_IS_TUNNEL_V6
		}
		tunnelSrc, err = IPToAddress(sa.TunnelSrcAddr)
		if err != nil {
			return err
		}
		tunnelDst, err = IPToAddress(sa.TunnelDstAddr)
		if err != nil {
			return err
		}
	}
	const undefinedPort = ^uint16(0)
	udpSrcPort := undefinedPort
	if sa.TunnelSrcPort != 0 {
		udpSrcPort = uint16(sa.TunnelSrcPort)
	}
	udpDstPort := undefinedPort
	if sa.TunnelDstPort != 0 {
		udpDstPort = uint16(sa.TunnelDstPort)
	}

	req := &vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: isAdd,
		Entry: ipsec_types.IpsecSadEntry{
			SadID:           sa.Index,
			Spi:             sa.Spi,
			Protocol:        protocolToIpsecProto(sa.Protocol),
			CryptoAlgorithm: ipsec_types.IpsecCryptoAlg(sa.CryptoAlg),
			CryptoKey: ipsec_types.Key{
				Data:   cryptoKey,
				Length: uint8(len(cryptoKey)),
			},
			Salt:               sa.CryptoSalt,
			IntegrityAlgorithm: ipsec_types.IpsecIntegAlg(sa.IntegAlg),
			IntegrityKey: ipsec_types.Key{
				Data:   integKey,
				Length: uint8(len(integKey)),
			},
			TunnelSrc:  tunnelSrc,
			TunnelDst:  tunnelDst,
			Flags:      flags,
			UDPSrcPort: udpSrcPort,
			UDPDstPort: udpDstPort,
		},
	}
	reply := &vpp_ipsec.IpsecSadEntryAddDelReply{}

	if err = h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func ipsecProtoToProtocol(ipsecProto ipsec_types.IpsecProto) ipsec.SecurityAssociation_IPSecProtocol {
	switch ipsecProto {
	case ipsec_types.IPSEC_API_PROTO_AH:
		return ipsec.SecurityAssociation_AH
	case ipsec_types.IPSEC_API_PROTO_ESP:
		return ipsec.SecurityAssociation_ESP
	default:
		return 0
	}
}

func protocolToIpsecProto(protocol ipsec.SecurityAssociation_IPSecProtocol) ipsec_types.IpsecProto {
	switch protocol {
	case ipsec.SecurityAssociation_AH:
		return ipsec_types.IPSEC_API_PROTO_AH
	case ipsec.SecurityAssociation_ESP:
		return ipsec_types.IPSEC_API_PROTO_ESP
	default:
		return 0
	}
}

func (h *IPSecVppHandler) tunProtectAddUpdateEntry(tp *ipsec.TunnelProtection, swIfIndex uint32) error {
	if len(tp.SaOut) == 0 || len(tp.SaIn) == 0 {
		return errors.New("missing outbound/inbound SA")
	}
	if len(tp.SaIn) > int(^uint8(0)) {
		return errors.New("invalid number of inbound SAs")
	}
	req := &vpp_ipsec.IpsecTunnelProtectUpdate{Tunnel: vpp_ipsec.IpsecTunnelProtect{
		SwIfIndex: interface_types.InterfaceIndex(swIfIndex),
		SaOut:     tp.SaOut[0],
		SaIn:      tp.SaIn,
		NSaIn:     uint8(len(tp.SaIn)),
	}}
	if tp.NextHopAddr != "" {
		nh, err := IPToAddress(tp.NextHopAddr)
		if err != nil {
			return err
		}
		req.Tunnel.Nh = nh
	}
	reply := &vpp_ipsec.IpsecTunnelProtectUpdateReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return nil
}

func (h *IPSecVppHandler) tunProtectDelEntry(tp *ipsec.TunnelProtection, swIfIndex uint32) error {
	req := &vpp_ipsec.IpsecTunnelProtectDel{
		SwIfIndex: interface_types.InterfaceIndex(swIfIndex),
	}
	if tp.NextHopAddr != "" {
		nh, err := IPToAddress(tp.NextHopAddr)
		if err != nil {
			return err
		}
		req.Nh = nh
	}
	reply := &vpp_ipsec.IpsecTunnelProtectDelReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return nil
}
