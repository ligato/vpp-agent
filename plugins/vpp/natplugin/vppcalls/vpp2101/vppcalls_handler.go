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

package vpp2101

import (
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101"
	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/ip"
	vpp_nat "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/nat44"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_nat.AllMessages()...)

	vppcalls.AddNatHandlerVersion(vpp2101.Version, msgs, NewNatVppHandler)
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
func NewNatVppHandler(c vpp.Client,
	ifIndexes ifaceidx.IfaceMetadataIndex, dhcpIndex idxmap.NamedMapping, log logging.Logger,
) vppcalls.NatVppAPI {
	callsChan, _ := c.NewAPIChannel()
	return &NatVppHandler{
		callsChannel: callsChan,
		ip:           vpp_ip.NewServiceClient(c),
		nat:          vpp_nat.NewServiceClient(c),
		ifIndexes:    ifIndexes,
		dhcpIndex:    dhcpIndex,
		log:          log,
	}
}
