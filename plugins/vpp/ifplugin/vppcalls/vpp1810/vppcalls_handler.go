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

package vpp1810

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/l2"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/af_packet"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/bond"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/memif"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/tap"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/tapv2"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/vmxnet3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/vxlan"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

type (
	InterfaceDetails = vppcalls.InterfaceDetails
	InterfaceMeta    = vppcalls.InterfaceMeta
	InterfaceEvent   = vppcalls.InterfaceEvent
	Dhcp             = vppcalls.Dhcp
	Client           = vppcalls.Client
	Lease            = vppcalls.Lease
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, af_packet.Messages...)
	msgs = append(msgs, bond.Messages...)
	msgs = append(msgs, dhcp.Messages...)
	msgs = append(msgs, interfaces.Messages...)
	msgs = append(msgs, ip.Messages...)
	msgs = append(msgs, ipsec.Messages...)
	msgs = append(msgs, l2.Messages...)
	msgs = append(msgs, memif.Messages...)
	msgs = append(msgs, tap.Messages...)
	msgs = append(msgs, tapv2.Messages...)
	msgs = append(msgs, vmxnet3.Messages...)
	msgs = append(msgs, vxlan.Messages...)

	vppcalls.Versions["vpp1810"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(ch govppapi.Channel, log logging.Logger) vppcalls.InterfaceVppAPI {
			return &InterfaceVppHandler{ch, log}
		},
	}
}

// InterfaceVppHandler is accessor for interface-related vppcalls methods
type InterfaceVppHandler struct {
	callsChannel govppapi.Channel
	log          logging.Logger
}
