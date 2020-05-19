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

package vpp2005

import (
	"encoding/hex"
	"net"
	"strconv"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ip_types"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ipsec_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

// DumpIPSecSA implements IPSec handler.
func (h *IPSecVppHandler) DumpIPSecSA() (saList []*vppcalls.IPSecSaDetails, err error) {
	return h.DumpIPSecSAWithIndex(^uint32(0)) // Get everything
}

// DumpIPSecSAWithIndex implements IPSec handler.
func (h *IPSecVppHandler) DumpIPSecSAWithIndex(saID uint32) (saList []*vppcalls.IPSecSaDetails, err error) {
	saDetails, err := h.dumpSecurityAssociations(saID)
	if err != nil {
		return nil, err
	}

	for _, saData := range saDetails {
		// Skip tunnel interfaces
		if uint32(saData.SwIfIndex) != ^uint32(0) {
			continue
		}

		var tunnelSrcAddr, tunnelDstAddr net.IP
		if saData.Entry.TunnelDst.Af == ip_types.ADDRESS_IP6 {
			src := saData.Entry.TunnelSrc.Un.GetIP6()
			dst := saData.Entry.TunnelDst.Un.GetIP6()
			tunnelSrcAddr, tunnelDstAddr = net.IP(src[:]), net.IP(dst[:])
		} else {
			src := saData.Entry.TunnelSrc.Un.GetIP4()
			dst := saData.Entry.TunnelDst.Un.GetIP4()
			tunnelSrcAddr, tunnelDstAddr = net.IP(src[:]), net.IP(dst[:])
		}

		sa := &ipsec.SecurityAssociation{
			Index:          saData.Entry.SadID,
			Spi:            saData.Entry.Spi,
			Protocol:       ipsecProtoToProtocol(saData.Entry.Protocol),
			CryptoAlg:      ipsec.CryptoAlg(saData.Entry.CryptoAlgorithm),
			CryptoKey:      hex.EncodeToString(saData.Entry.CryptoKey.Data[:saData.Entry.CryptoKey.Length]),
			IntegAlg:       ipsec.IntegAlg(saData.Entry.IntegrityAlgorithm),
			IntegKey:       hex.EncodeToString(saData.Entry.IntegrityKey.Data[:saData.Entry.IntegrityKey.Length]),
			UseEsn:         (saData.Entry.Flags & ipsec_types.IPSEC_API_SAD_FLAG_USE_ESN) != 0,
			UseAntiReplay:  (saData.Entry.Flags & ipsec_types.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY) != 0,
			EnableUdpEncap: (saData.Entry.Flags & ipsec_types.IPSEC_API_SAD_FLAG_UDP_ENCAP) != 0,
		}
		if !tunnelSrcAddr.IsUnspecified() {
			sa.TunnelSrcAddr = tunnelSrcAddr.String()
		}
		if !tunnelDstAddr.IsUnspecified() {
			sa.TunnelDstAddr = tunnelDstAddr.String()
		}
		meta := &vppcalls.IPSecSaMeta{
			SaID:           saData.Entry.SadID,
			IfIdx:          uint32(saData.SwIfIndex),
			Salt:           saData.Salt,
			SeqOutbound:    saData.SeqOutbound,
			LastSeqInbound: saData.LastSeqInbound,
			ReplayWindow:   saData.ReplayWindow,
		}
		saList = append(saList, &vppcalls.IPSecSaDetails{
			Sa:   sa,
			Meta: meta,
		})
	}

	return saList, nil
}

// DumpIPSecSPD implements IPSec handler.
func (h *IPSecVppHandler) DumpIPSecSPD() (spdList []*vppcalls.IPSecSpdDetails, err error) {
	metadata := make(map[string]*vppcalls.SpdMeta)

	// TODO: add integration test for dumping SPD interfaces
	spdInterfaces, err := h.dumpSpdInterfaces()
	if err != nil {
		h.log.Warnf("dumping spd interfaces failed: %v", err)
	}
	h.log.Debugf("dumped spd interfaces: %+v", spdInterfaces)

	// Get all VPP SPD indexes
	spdIndexes, err := h.dumpSpdIndexes()
	if err != nil {
		return nil, errors.Errorf("failed to dump SPD indexes: %v", err)
	}
	for spdIdx, numPolicies := range spdIndexes {
		spd := &ipsec.SecurityPolicyDatabase{
			Index: uint32(spdIdx),
		}

		req := &vpp_ipsec.IpsecSpdDump{
			SpdID: uint32(spdIdx),
			SaID:  ^uint32(0),
		}
		requestCtx := h.callsChannel.SendMultiRequest(req)

		for {
			spdDetails := &vpp_ipsec.IpsecSpdDetails{}
			stop, err := requestCtx.ReceiveReply(spdDetails)
			if stop {
				break
			}
			if err != nil {
				return nil, err
			}

			// Addresses
			remoteStartAddr, remoteStopAddr := ipsecAddrToIP(spdDetails.Entry.RemoteAddressStart), ipsecAddrToIP(spdDetails.Entry.RemoteAddressStop)
			localStartAddr, localStopAddr := ipsecAddrToIP(spdDetails.Entry.LocalAddressStart), ipsecAddrToIP(spdDetails.Entry.LocalAddressStop)

			// Prepare policy entry and put to the SPD
			policyEntry := &ipsec.SecurityPolicyDatabase_PolicyEntry{
				SaIndex:         spdDetails.Entry.SaID,
				Priority:        spdDetails.Entry.Priority,
				IsOutbound:      spdDetails.Entry.IsOutbound,
				RemoteAddrStart: remoteStartAddr.String(),
				RemoteAddrStop:  remoteStopAddr.String(),
				LocalAddrStart:  localStartAddr.String(),
				LocalAddrStop:   localStopAddr.String(),
				Protocol:        uint32(spdDetails.Entry.Protocol),
				RemotePortStart: uint32(spdDetails.Entry.RemotePortStart),
				RemotePortStop:  resetPort(spdDetails.Entry.RemotePortStop),
				LocalPortStart:  uint32(spdDetails.Entry.LocalPortStart),
				LocalPortStop:   resetPort(spdDetails.Entry.LocalPortStop),
				Action:          ipsec.SecurityPolicyDatabase_PolicyEntry_Action(spdDetails.Entry.Policy),
			}
			spd.PolicyEntries = append(spd.PolicyEntries, policyEntry)

			// Prepare meta and put to the metadata map
			meta := &vppcalls.SpdMeta{
				SaID:   spdDetails.Entry.SaID,
				Policy: uint8(spdDetails.Entry.Policy),
			}
			metadata[strconv.Itoa(int(spdDetails.Entry.SaID))] = meta
		}
		// Store SPD in list
		spdList = append(spdList, &vppcalls.IPSecSpdDetails{
			Spd:         spd,
			PolicyMeta:  metadata,
			NumPolicies: numPolicies,
		})
	}

	return spdList, nil
}

// DumpTunnelProtections returns configured IPSec tunnel protections.
func (h *IPSecVppHandler) DumpTunnelProtections() (tpList []*ipsec.TunnelProtection, err error) {
	req := &vpp_ipsec.IpsecTunnelProtectDump{
		SwIfIndex: vpp_ipsec.InterfaceIndex(^uint32(0)),
	}
	requestCtx := h.callsChannel.SendMultiRequest(req)
	for {
		tpDetails := &vpp_ipsec.IpsecTunnelProtectDetails{}
		stop, err := requestCtx.ReceiveReply(tpDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(uint32(tpDetails.Tun.SwIfIndex))
		if !exists {
			h.log.Warnf("Tunnel protection dump: interface name for index %d not found", tpDetails.Tun.SwIfIndex)
			continue
		}
		tp := &ipsec.TunnelProtection{
			Interface: ifName,
			SaOut:     []uint32{tpDetails.Tun.SaOut},
		}
		tp.SaIn = append(tp.SaIn, tpDetails.Tun.SaIn...)
		tpList = append(tpList, tp)
	}
	return
}

// Get all interfaces of SPD configured on the VPP
func (h *IPSecVppHandler) dumpSpdInterfaces() (map[int][]uint32, error) {
	// SPD index to interface indexes
	spdInterfaces := make(map[int][]uint32)

	req := &vpp_ipsec.IpsecSpdInterfaceDump{}
	reqCtx := h.callsChannel.SendMultiRequest(req)

	for {
		spdDetails := &vpp_ipsec.IpsecSpdInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(spdDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		spdInterfaces[int(spdDetails.SpdIndex)] = append(spdInterfaces[int(spdDetails.SpdIndex)], uint32(spdDetails.SwIfIndex))
	}

	return spdInterfaces, nil
}

// Get all indexes of SPD configured on the VPP
func (h *IPSecVppHandler) dumpSpdIndexes() (map[int]uint32, error) {
	// SPD index to number of policies
	spdIndexes := make(map[int]uint32)

	req := &vpp_ipsec.IpsecSpdsDump{}
	reqCtx := h.callsChannel.SendMultiRequest(req)

	for {
		spdDetails := &vpp_ipsec.IpsecSpdsDetails{}
		stop, err := reqCtx.ReceiveReply(spdDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		spdIndexes[int(spdDetails.SpdID)] = spdDetails.Npolicies
	}

	return spdIndexes, nil
}

// Get all security association (used also for tunnel interfaces) in binary api format
func (h *IPSecVppHandler) dumpSecurityAssociations(saID uint32) (saList []*vpp_ipsec.IpsecSaDetails, err error) {
	req := &vpp_ipsec.IpsecSaDump{
		SaID: saID,
	}
	requestCtx := h.callsChannel.SendMultiRequest(req)

	for {
		saDetails := &vpp_ipsec.IpsecSaDetails{}
		stop, err := requestCtx.ReceiveReply(saDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		saList = append(saList, saDetails)
	}

	return saList, nil
}

// ResetPort returns 0 if stop port has maximum value (default VPP value if stop port is not defined)
func resetPort(port uint16) uint32 {
	if port == ^uint16(0) {
		return 0
	}
	return uint32(port)
}
