//  Copyright (c) 2022 Cisco and/or its affiliates.
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

package vpp2306

import (
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2306"
	vpp_nat_ed "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2306/nat44_ed"
	vpp_nat_ei "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2306/nat44_ei"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_nat_ed.AllMessages()...)
	msgs = append(msgs, vpp_nat_ei.AllMessages()...)

	vppcalls.AddNatHandlerVersion(vpp2306.Version, msgs, NewNatVppHandler)
}

// NatVppHandler is accessor for NAT-related vppcalls methods.
type NatVppHandler struct {
	callsChannel govppapi.Channel
	natEd        vpp_nat_ed.RPCService
	natEi        vpp_nat_ei.RPCService
	ifIndexes    ifaceidx.IfaceMetadataIndex
	dhcpIndex    idxmap.NamedMapping
	log          logging.Logger
	ed           bool
}

// NewNatVppHandler creates new instance of NAT vppcalls handler.
func NewNatVppHandler(c vpp.Client,
	ifIndexes ifaceidx.IfaceMetadataIndex, dhcpIndex idxmap.NamedMapping, log logging.Logger,
) vppcalls.NatVppAPI {
	callsChan, _ := c.NewAPIChannel()
	return &NatVppHandler{
		callsChannel: callsChan,
		natEd:        vpp_nat_ed.NewServiceClient(c),
		natEi:        vpp_nat_ei.NewServiceClient(c),
		ifIndexes:    ifIndexes,
		dhcpIndex:    dhcpIndex,
		log:          log,
		ed:           true,
	}
}
