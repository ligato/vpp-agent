//  Copyright (c) 2020 Doc.ai and/or its affiliates.
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
	"encoding/base64"
	vpp_wg "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/wireguard"
	wg "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/wireguard"
)

// DumpWgPeers implements wg handler.
func (h *WgVppHandler) DumpWgPeers() (peerList []*wg.Peer, err error) {
	req := &vpp_wg.WireguardPeersDump{}
	requestCtx := h.callsChannel.SendMultiRequest(req)

	var vppPeerList []*vpp_wg.WireguardPeersDetails
	for {
		vppPeerDetails := &vpp_wg.WireguardPeersDetails{}
		stop, err := requestCtx.ReceiveReply(vppPeerDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		vppPeerList = append(vppPeerList, vppPeerDetails)
	}

	for _, vppPeerDetails := range vppPeerList {
		peerDetails := &wg.Peer{
			Port:                uint32(vppPeerDetails.Peer.Port),
			PersistentKeepalive: uint32(vppPeerDetails.Peer.PersistentKeepalive),
			Flags:               uint32(vppPeerDetails.Peer.Flags),
		}

		peerDetails.PublicKey = base64.StdEncoding.EncodeToString(vppPeerDetails.Peer.PublicKey)

		for _, prefix := range vppPeerDetails.Peer.AllowedIps {
			peerDetails.AllowedIps = append(peerDetails.AllowedIps, prefix.String())
		}

		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(uint32(vppPeerDetails.Peer.SwIfIndex))
		if !exists {
			h.log.Warnf("Wireguard peers dump: interface name for index %d not found", vppPeerDetails.Peer.SwIfIndex)
			continue
		}

		peerDetails.WgIfName = ifName;

		endpointAddr := vppPeerDetails.Peer.Endpoint.ToIP()
		if !endpointAddr.IsUnspecified() {
			peerDetails.Endpoint = endpointAddr.String()
		}

		peerList = append(peerList, peerDetails)
	}

	return
}
