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

package vpp1904

import (
	govppapi "git.fd.io/govpp.git/api"
	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/nat"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, nat.AllMessages()...)

	vppcalls.AddNatHandlerVersion(vpp1904.Version, msgs, NewNatVppHandler)
}

// NatVppHandler is accessor for NAT-related vppcalls methods.
type NatVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	dhcpIndex    idxmap.NamedMapping
	log          logging.Logger
}

// NewNatVppHandler creates new instance of NAT vppcalls handler.
func NewNatVppHandler(
	callsChan govppapi.Channel,
	ifIndexes ifaceidx.IfaceMetadataIndex,
	dhcpIndex idxmap.NamedMapping,
	log logging.Logger,
) vppcalls.NatVppAPI {
	return &NatVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		dhcpIndex:    dhcpIndex,
		log:          log,
	}
}
