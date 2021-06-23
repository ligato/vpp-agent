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
	"fmt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	vpp_wg "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/wireguard"
	wg "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/wireguard"
)

func (h *WgVppHandler) AddPeer(peer *wg.Peer) (uint32, error) {
	invalidIdx := ^uint32(0)

	peer_vpp := vpp_wg.WireguardPeer{
		Port:                uint16(peer.Port),
		PersistentKeepalive: uint16(peer.PersistentKeepalive),
	}

	publicKeyBin, err := base64.StdEncoding.DecodeString(peer.PublicKey)
	if err != nil {
		return invalidIdx, err
	}
	peer_vpp.PublicKey = publicKeyBin

	ifaceMeta, found := h.ifIndexes.LookupByName(peer.WgIfName)
	if !found {
		return invalidIdx, fmt.Errorf("failed to get interface metadata")
	}
	peer_vpp.SwIfIndex = interface_types.InterfaceIndex(ifaceMeta.SwIfIndex)
	peer_vpp.TableID = ifaceMeta.Vrf

	peer_vpp.Endpoint, err = ip_types.ParseAddress(peer.Endpoint)
	if err != nil {
		return invalidIdx, err
	}

	for _, allowedIp := range peer.AllowedIps {
		prefix, err := ip_types.ParsePrefix(allowedIp);
		if err != nil {
			return invalidIdx, err
		}
		peer_vpp.AllowedIps = append(peer_vpp.AllowedIps, prefix);
	}

	request := &vpp_wg.WireguardPeerAdd {
		Peer: peer_vpp,
	};
	// prepare reply
	reply := &vpp_wg.WireguardPeerAddReply{}
	// send request and obtain reply
	if err := h.callsChannel.SendRequest(request).ReceiveReply(reply); err != nil {
		return invalidIdx, err
	}
	return reply.PeerIndex, nil;
}

func (h *WgVppHandler) RemovePeer(peer_idx uint32) error {
	// prepare request
	request := &vpp_wg.WireguardPeerRemove{
		PeerIndex: peer_idx,
	}
	// prepare reply
	reply := &vpp_wg.WireguardPeerRemoveReply{}

	// send request and obtain reply
	if err := h.callsChannel.SendRequest(request).ReceiveReply(reply); err != nil {
		return err
	}
	return nil;
}
