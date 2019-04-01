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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/bond"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/l2"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/af_packet"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/memif"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/tap"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/tapv2"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/vmxnet3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/vxlan"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
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

	vppcalls.Versions["vpp1901"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(ch govppapi.Channel, log logging.Logger) vppcalls.InterfaceVppAPI {
			return NewInterfaceVppHandler(ch, log)
		},
	}
}

// InterfaceVppHandler is accessor for interface-related vppcalls methods
type InterfaceVppHandler struct {
	callsChannel govppapi.Channel
	log          logging.Logger
}

// NewInterfaceVppHandler returns new InterfaceVppHandler.
func NewInterfaceVppHandler(ch govppapi.Channel, log logging.Logger) *InterfaceVppHandler {
	return &InterfaceVppHandler{ch, log}
}
