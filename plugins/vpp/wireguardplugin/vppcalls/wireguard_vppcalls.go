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

package vppcalls

import (
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	wg "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/wireguard"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
)

type WgVppAPI interface {
	WgVppRead

	// Set peer via binary API
	AddPeer(peer *wg.Peer) (uint32, error)
	// Remove peer via binary API
	RemovePeer(peer_idx uint32) error
}

// WgVPPRead provides read methods for wireguard
type WgVppRead interface {
	// DumpWgPeer returns a peers state
	DumpWgPeers() (peerList []*wg.Peer, err error)
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "wireguard",
	HandlerAPI: (*WgVppAPI)(nil),
})

type NewHandlerFunc func(ch govppapi.Channel, ifDdx ifaceidx.IfaceMetadataIndex, log logging.Logger) WgVppAPI

func AddHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return ch.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return h(ch, a[0].(ifaceidx.IfaceMetadataIndex), a[1].(logging.Logger))
		},
	})
}

func CompatibleWgVppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) WgVppAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx, log).(WgVppAPI)
	}
	return nil
}
