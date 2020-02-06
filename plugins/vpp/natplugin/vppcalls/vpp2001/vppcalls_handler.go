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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001"
	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip"
	vpp_nat "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/nat"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_nat.AllMessages()...)

	vppcalls.AddNatHandlerVersion(vpp2001.Version, msgs, NewNatVppHandler)
}

// NatVppHandler is accessor for NAT-related vppcalls methods.
type NatVppHandler struct {
	callsChannel govppapi.Channel
	ip           vpp_ip.RPCService
	nat          vpp_nat.RPCService
	ifIndexes    ifaceidx.IfaceMetadataIndex
	dhcpIndex    idxmap.NamedMapping
	log          logging.Logger
}

// NewNatVppHandler creates new instance of NAT vppcalls handler.
func NewNatVppHandler(callsChan govppapi.Channel,
	ifIndexes ifaceidx.IfaceMetadataIndex, dhcpIndex idxmap.NamedMapping, log logging.Logger,
) vppcalls.NatVppAPI {
	return &NatVppHandler{
		callsChannel: callsChan,
		ip:           vpp_ip.NewServiceClient(callsChan),
		nat:          vpp_nat.NewServiceClient(callsChan),
		ifIndexes:    ifIndexes,
		dhcpIndex:    dhcpIndex,
		log:          log,
	}
}
