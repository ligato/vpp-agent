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
	"encoding/hex"
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/utils/addrs"

	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ipsec_types"
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

// AddSPDEntry implements IPSec handler.
func (h *IPSecVppHandler) AddSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabase_PolicyEntry) error {
	return h.spdAddDelEntry(spdID, saID, spd, true)
}

// DeleteSPDEntry implements IPSec handler.
func (h *IPSecVppHandler) DeleteSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabase_PolicyEntry) error {
	return h.spdAddDelEntry(spdID, saID, spd, false)
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
		IsAdd: boolToUint(isAdd),
		SpdID: spdID,
	}
	reply := &vpp_ipsec.IpsecSpdAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *IPSecVppHandler) spdAddDelEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabase_PolicyEntry, isAdd bool) error {
	req := &vpp_ipsec.IpsecSpdEntryAddDel{
		IsAdd: boolToUint(isAdd),
		Entry: vpp_ipsec.IpsecSpdEntry{
			SpdID:           spdID,
			Priority:        spd.Priority,
			IsOutbound:      boolToUint(spd.IsOutbound),
			Protocol:        uint8(spd.Protocol),
			RemotePortStart: uint16(spd.RemotePortStart),
			RemotePortStop:  uint16(spd.RemotePortStop),
			LocalPortStart:  uint16(spd.LocalPortStart),
			LocalPortStop:   uint16(spd.LocalPortStop),
			Policy:          vpp_ipsec.IpsecSpdAction(spd.Action),
			SaID:            saID,
		},
	}
	if req.Entry.RemotePortStop == 0 {
		req.Entry.RemotePortStop = ^req.Entry.RemotePortStop
	}
	if req.Entry.LocalPortStop == 0 {
		req.Entry.LocalPortStop = ^req.Entry.LocalPortStop
	}

	var err error
	req.Entry.RemoteAddressStart, err = IPToAddress(ipOr(spd.RemoteAddrStart, "0.0.0.0"))
	if err != nil {
		return err
	}
	req.Entry.RemoteAddressStop, err = IPToAddress(ipOr(spd.RemoteAddrStop, "255.255.255.255"))
	if err != nil {
		return err
	}
	req.Entry.LocalAddressStart, err = IPToAddress(ipOr(spd.LocalAddrStart, "0.0.0.0"))
	if err != nil {
		return err
	}
	req.Entry.LocalAddressStop, err = IPToAddress(ipOr(spd.LocalAddrStop, "255.255.255.255"))
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
		IsAdd:     boolToUint(isAdd),
		SwIfIndex: swIfIdx,
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

	var flags vpp_ipsec.IpsecSadFlags
	if sa.UseEsn {
		flags |= ipsec_types.IPSEC_API_SAD_FLAG_USE_ESN
	}
	if sa.UseAntiReplay {
		flags |= ipsec_types.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY
	}
	if sa.EnableUdpEncap {
		flags |= ipsec_types.IPSEC_API_SAD_FLAG_UDP_ENCAP
	}
	var tunnelSrc, tunnelDst ipsec_types.Address
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

	req := &vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: boolToUint(isAdd),
		Entry: vpp_ipsec.IpsecSadEntry{
			SadID:           sa.Index,
			Spi:             sa.Spi,
			Protocol:        vpp_ipsec.IpsecProto(sa.Protocol),
			CryptoAlgorithm: vpp_ipsec.IpsecCryptoAlg(sa.CryptoAlg),
			CryptoKey: vpp_ipsec.Key{
				Data:   cryptoKey,
				Length: uint8(len(cryptoKey)),
			},
			IntegrityAlgorithm: vpp_ipsec.IpsecIntegAlg(sa.IntegAlg),
			IntegrityKey: vpp_ipsec.Key{
				Data:   integKey,
				Length: uint8(len(integKey)),
			},
			TunnelSrc: tunnelSrc,
			TunnelDst: tunnelDst,
			Flags:     flags,
		},
	}
	reply := &vpp_ipsec.IpsecSadEntryAddDelReply{}

	if err = h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *IPSecVppHandler) tunProtectAddUpdateEntry(tp *ipsec.TunnelProtection, swIfIndex uint32) error {
	if len(tp.SaOut) == 0 || len(tp.SaIn) == 0 {
		return errors.New("missing outbound/inbound SA")
	}
	if len(tp.SaIn) > int(^uint8(0)) {
		return errors.New("invalid number of inbound SAs")
	}
	req := &vpp_ipsec.IpsecTunnelProtectUpdate{Tunnel: vpp_ipsec.IpsecTunnelProtect{
		SwIfIndex: vpp_ipsec.InterfaceIndex(swIfIndex),
		SaOut:     tp.SaOut[0],
		SaIn:      tp.SaIn,
		NSaIn:     uint8(len(tp.SaIn)),
	}}
	reply := &vpp_ipsec.IpsecTunnelProtectUpdateReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return nil
}

func (h *IPSecVppHandler) tunProtectDelEntry(tp *ipsec.TunnelProtection, swIfIndex uint32) error {
	req := &vpp_ipsec.IpsecTunnelProtectDel{
		SwIfIndex: vpp_ipsec.InterfaceIndex(swIfIndex),
	}
	reply := &vpp_ipsec.IpsecTunnelProtectDelReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return nil
}

func boolToUint(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}
