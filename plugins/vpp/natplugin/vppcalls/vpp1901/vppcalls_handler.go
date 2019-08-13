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

package vpp1901

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin/vppcalls"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/nat"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, nat.Messages...)

	vppcalls.Versions["vpp1901"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(
			ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, dhcpIdx idxmap.NamedMapping, log logging.Logger,
		) vppcalls.NatVppAPI {
			return NewNatVppHandler(ch, ifIdx, dhcpIdx, log)
		},
	}
}

// NatVppHandler is accessor for NAT-related vppcalls methods.
type NatVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	dhcpIndex    idxmap.NamedMapping
	log          logging.Logger
}

// NewNatVppHandler creates new instance of NAT vppcalls handler.
func NewNatVppHandler(callsChan govppapi.Channel,
	ifIndexes ifaceidx.IfaceMetadataIndex, dhcpIndex idxmap.NamedMapping, log logging.Logger,
) *NatVppHandler {
	return &NatVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		dhcpIndex:    dhcpIndex,
		log:          log,
	}
}
