// Copyright (c) 2017 Cisco and/or its affiliates.
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

package vppcalls

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	ipsec_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ipsec"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/ipsec"
)

func tunnelIfAddDel(tunnel *ipsec.TunnelInterfaces_Tunnel, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (uint32, error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(ipsec_api.IpsecTunnelIfAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	localCryptoKey, err := hex.DecodeString(tunnel.LocalCryptoKey)
	if err != nil {
		return 0, err
	}
	remoteCryptoKey, err := hex.DecodeString(tunnel.RemoteCryptoKey)
	if err != nil {
		return 0, err
	}
	localIntegKey, err := hex.DecodeString(tunnel.LocalIntegKey)
	if err != nil {
		return 0, err
	}
	remoteIntegKey, err := hex.DecodeString(tunnel.RemoteIntegKey)
	if err != nil {
		return 0, err
	}

	req := &ipsec_api.IpsecTunnelIfAddDel{
		IsAdd:              boolToUint(isAdd),
		Esn:                boolToUint(tunnel.Esn),
		AntiReplay:         boolToUint(tunnel.AntiReplay),
		LocalIP:            net.ParseIP(tunnel.LocalIp).To4(),
		RemoteIP:           net.ParseIP(tunnel.RemoteIp).To4(),
		LocalSpi:           tunnel.LocalSpi,
		RemoteSpi:          tunnel.RemoteSpi,
		CryptoAlg:          uint8(tunnel.CryptoAlg),
		LocalCryptoKey:     localCryptoKey,
		LocalCryptoKeyLen:  uint8(len(localCryptoKey)),
		RemoteCryptoKey:    remoteCryptoKey,
		RemoteCryptoKeyLen: uint8(len(remoteCryptoKey)),
		IntegAlg:           uint8(tunnel.IntegAlg),
		LocalIntegKey:      localIntegKey,
		LocalIntegKeyLen:   uint8(len(localIntegKey)),
		RemoteIntegKey:     remoteIntegKey,
		RemoteIntegKeyLen:  uint8(len(remoteIntegKey)),
	}

	reply := &ipsec_api.IpsecTunnelIfAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return reply.SwIfIndex, nil
}

// AddSPD adds SPD to VPP via binary API
func AddTunnelInterface(tunnel *ipsec.TunnelInterfaces_Tunnel, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (uint32, error) {
	return tunnelIfAddDel(tunnel, true, vppChan, stopwatch)
}

// DelSPD deletes SPD from VPP via binary API
func DelTunnelInterface(ifIdx uint32, tunnel *ipsec.TunnelInterfaces_Tunnel, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	// Note: ifIdx is not used now, tunnel shiould be matched based on paramters
	_, err := tunnelIfAddDel(tunnel, false, vppChan, stopwatch)
	return err
}

func spdAddDel(spdID uint32, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ipsec_api.IpsecSpdAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ipsec_api.IpsecSpdAddDel{
		IsAdd: boolToUint(isAdd),
		SpdID: spdID,
	}

	reply := &ipsec_api.IpsecSpdAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// AddSPD adds SPD to VPP via binary API
func AddSPD(spdID uint32, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return spdAddDel(spdID, true, vppChan, stopwatch)
}

// DelSPD deletes SPD from VPP via binary API
func DelSPD(spdID uint32, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return spdAddDel(spdID, false, vppChan, stopwatch)
}

func interfaceAddDelSpd(spdID, swIfIdx uint32, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ipsec_api.IpsecInterfaceAddDelSpd{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ipsec_api.IpsecInterfaceAddDelSpd{
		IsAdd:     boolToUint(isAdd),
		SwIfIndex: swIfIdx,
		SpdID:     spdID,
	}

	reply := &ipsec_api.IpsecInterfaceAddDelSpdReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// InterfaceAddSPD adds SPD interface assignment to VPP via binary API
func InterfaceAddSPD(spdID, swIfIdx uint32, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return interfaceAddDelSpd(spdID, swIfIdx, true, vppChan, stopwatch)
}

// InterfaceDelSPD deletes SPD interface assignment from VPP via binary API
func InterfaceDelSPD(spdID, swIfIdx uint32, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return interfaceAddDelSpd(spdID, swIfIdx, false, vppChan, stopwatch)
}

func spdAddDelEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabases_SPD_PolicyEntry, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ipsec_api.IpsecSpdAddDelEntry{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ipsec_api.IpsecSpdAddDelEntry{
		IsAdd:           boolToUint(isAdd),
		SpdID:           spdID,
		Priority:        spd.Priority,
		IsOutbound:      boolToUint(spd.IsOutbound),
		Protocol:        uint8(spd.Protocol),
		RemotePortStart: uint16(spd.RemotePortStart),
		RemotePortStop:  uint16(spd.RemotePortStop),
		LocalPortStart:  uint16(spd.LocalPortStart),
		LocalPortStop:   uint16(spd.LocalPortStop),
		Policy:          uint8(spd.Action),
		SaID:            saID,
	}
	if req.RemotePortStop == 0 {
		req.RemotePortStop = ^req.RemotePortStop
	}
	if req.LocalPortStop == 0 {
		req.LocalPortStop = ^req.LocalPortStop
	}
	if spd.RemoteAddrStart != "" {
		isIPv6, err := addrs.IsIPv6(spd.RemoteAddrStart)
		if err != nil {
			return err
		}
		if isIPv6 {
			req.IsIpv6 = 1
			req.RemoteAddressStart = net.ParseIP(spd.RemoteAddrStart).To16()
			req.RemoteAddressStop = net.ParseIP(spd.RemoteAddrStop).To16()
			req.LocalAddressStart = net.ParseIP(spd.LocalAddrStart).To16()
			req.LocalAddressStop = net.ParseIP(spd.LocalAddrStop).To16()
		} else {
			req.IsIpv6 = 0
			req.RemoteAddressStart = net.ParseIP(spd.RemoteAddrStart).To4()
			req.RemoteAddressStop = net.ParseIP(spd.RemoteAddrStop).To4()
			req.LocalAddressStart = net.ParseIP(spd.LocalAddrStart).To4()
			req.LocalAddressStop = net.ParseIP(spd.LocalAddrStop).To4()
		}
	} else {
		req.RemoteAddressStart = net.ParseIP("0.0.0.0").To4()
		req.RemoteAddressStop = net.ParseIP("255.255.255.255").To4()
		req.LocalAddressStart = net.ParseIP("0.0.0.0").To4()
		req.LocalAddressStop = net.ParseIP("255.255.255.255").To4()
	}

	reply := &ipsec_api.IpsecSpdAddDelEntryReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// AddSPDEntry adds SPD policy entry to VPP via binary API
func AddSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabases_SPD_PolicyEntry, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return spdAddDelEntry(spdID, saID, spd, true, vppChan, stopwatch)
}

// DelSPDEntry deletes SPD policy entry from VPP via binary API
func DelSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabases_SPD_PolicyEntry, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return spdAddDelEntry(spdID, saID, spd, false, vppChan, stopwatch)
}

func sadAddDelEntry(saID uint32, sa *ipsec.SecurityAssociations_SA, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ipsec_api.IpsecSadAddDelEntry{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	cryptoKey, err := hex.DecodeString(sa.CryptoKey)
	if err != nil {
		return err
	}
	integKey, err := hex.DecodeString(sa.IntegKey)
	if err != nil {
		return err
	}

	req := &ipsec_api.IpsecSadAddDelEntry{
		IsAdd:                     boolToUint(isAdd),
		SadID:                     saID,
		Spi:                       sa.Spi,
		Protocol:                  uint8(sa.Protocol),
		CryptoAlgorithm:           uint8(sa.CryptoAlg),
		CryptoKey:                 cryptoKey,
		CryptoKeyLength:           uint8(len(cryptoKey)),
		IntegrityAlgorithm:        uint8(sa.IntegAlg),
		IntegrityKey:              integKey,
		IntegrityKeyLength:        uint8(len(integKey)),
		UseExtendedSequenceNumber: boolToUint(sa.UseEsn),
		UseAntiReplay:             boolToUint(sa.UseAntiReplay),
	}
	if sa.TunnelSrcAddr != "" {
		req.IsTunnel = 1
		isIPv6, err := addrs.IsIPv6(sa.TunnelSrcAddr)
		if err != nil {
			return err
		}
		if isIPv6 {
			req.IsTunnelIpv6 = 1
			req.TunnelSrcAddress = net.ParseIP(sa.TunnelSrcAddr).To16()
			req.TunnelDstAddress = net.ParseIP(sa.TunnelDstAddr).To16()
		} else {
			req.IsTunnelIpv6 = 0
			req.TunnelSrcAddress = net.ParseIP(sa.TunnelSrcAddr).To4()
			req.TunnelDstAddress = net.ParseIP(sa.TunnelDstAddr).To4()
		}
	}

	reply := &ipsec_api.IpsecSadAddDelEntryReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// AddSAEntry adds SA to VPP via binary API
func AddSAEntry(saID uint32, sa *ipsec.SecurityAssociations_SA, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return sadAddDelEntry(saID, sa, true, vppChan, stopwatch)
}

// DelSAEntry deletes SA from VPP via binary API
func DelSAEntry(saID uint32, sa *ipsec.SecurityAssociations_SA, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return sadAddDelEntry(saID, sa, false, vppChan, stopwatch)
}

// CheckMsgCompatibilityForIPSec verifies compatibility of used binary API calls
func CheckMsgCompatibilityForIPSec(vppChan *govppapi.Channel) error {
	msgs := []govppapi.Message{
		&ipsec_api.IpsecSpdAddDel{},
		&ipsec_api.IpsecSpdAddDelReply{},
		&ipsec_api.IpsecInterfaceAddDelSpd{},
		&ipsec_api.IpsecInterfaceAddDelSpdReply{},
		&ipsec_api.IpsecSpdAddDelEntry{},
		&ipsec_api.IpsecSpdAddDelEntryReply{},
		&ipsec_api.IpsecSadAddDelEntry{},
		&ipsec_api.IpsecSadAddDelEntryReply{},
		&ipsec_api.IpsecSpdDump{},
		&ipsec_api.IpsecSpdDetails{},
		&ipsec_api.IpsecTunnelIfAddDel{},
		&ipsec_api.IpsecTunnelIfAddDelReply{},
		&ipsec_api.IpsecSaDump{},
		&ipsec_api.IpsecSaDetails{},
		&ipsec_api.IpsecTunnelIfSetKey{},
		&ipsec_api.IpsecTunnelIfSetKeyReply{},
		&ipsec_api.IpsecTunnelIfSetSa{},
		&ipsec_api.IpsecTunnelIfSetSaReply{},
	}
	return vppChan.CheckMessageCompatibility(msgs...)
}

func boolToUint(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}
