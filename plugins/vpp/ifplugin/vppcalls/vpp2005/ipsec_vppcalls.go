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
	"context"
	"encoding/hex"

	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ipsec"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// AddIPSecTunnelInterface adds a new IPSec tunnel interface.
func (h *InterfaceVppHandler) AddIPSecTunnelInterface(ctx context.Context, ifName string, ipSecLink *ifs.IPSecLink) (uint32, error) {
	return h.tunnelIfAddDel(ctx, ifName, ipSecLink, true)
}

// DeleteIPSecTunnelInterface removes existing IPSec tunnel interface.
func (h *InterfaceVppHandler) DeleteIPSecTunnelInterface(ctx context.Context, ifName string, idx uint32, ipSecLink *ifs.IPSecLink) error {
	// Note: ifIdx is not used now, tunnel should be matched based on parameters
	_, err := h.tunnelIfAddDel(ctx, ifName, ipSecLink, false)
	return err
}

func (h *InterfaceVppHandler) tunnelIfAddDel(ctx context.Context, ifName string, ipSecLink *ifs.IPSecLink, isAdd bool) (uint32, error) {
	localCryptoKey, err := hex.DecodeString(ipSecLink.LocalCryptoKey)
	if err != nil {
		return 0, err
	}
	remoteCryptoKey, err := hex.DecodeString(ipSecLink.RemoteCryptoKey)
	if err != nil {
		return 0, err
	}
	localIntegKey, err := hex.DecodeString(ipSecLink.LocalIntegKey)
	if err != nil {
		return 0, err
	}
	remoteIntegKey, err := hex.DecodeString(ipSecLink.RemoteIntegKey)
	if err != nil {
		return 0, err
	}

	localIP, err := IPToAddress(ipSecLink.LocalIp)
	if err != nil {
		return 0, err
	}
	remoteIP, err := IPToAddress(ipSecLink.RemoteIp)
	if err != nil {
		return 0, err
	}

	req := &vpp_ipsec.IpsecTunnelIfAddDel{
		IsAdd:              isAdd,
		Esn:                ipSecLink.Esn,
		AntiReplay:         ipSecLink.AntiReplay,
		LocalIP:            localIP,
		RemoteIP:           remoteIP,
		LocalSpi:           ipSecLink.LocalSpi,
		RemoteSpi:          ipSecLink.RemoteSpi,
		CryptoAlg:          uint8(ipSecLink.CryptoAlg),
		LocalCryptoKey:     localCryptoKey,
		LocalCryptoKeyLen:  uint8(len(localCryptoKey)),
		RemoteCryptoKey:    remoteCryptoKey,
		RemoteCryptoKeyLen: uint8(len(remoteCryptoKey)),
		IntegAlg:           uint8(ipSecLink.IntegAlg),
		LocalIntegKey:      localIntegKey,
		LocalIntegKeyLen:   uint8(len(localIntegKey)),
		RemoteIntegKey:     remoteIntegKey,
		RemoteIntegKeyLen:  uint8(len(remoteIntegKey)),
		UDPEncap:           ipSecLink.EnableUdpEncap,
	}
	reply, err := h.ipsec.IpsecTunnelIfAddDel(ctx, req)
	if err != nil {
		return 0, err
	}

	return uint32(reply.SwIfIndex), nil
}
