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
	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/af_packet"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/bond"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/l2"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/memif"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/tapv2"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/vmxnet3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/vxlan"
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
	msgs = append(msgs, tapv2.Messages...)
	msgs = append(msgs, vmxnet3.Messages...)
	msgs = append(msgs, vxlan.Messages...)

	vppcalls.Versions["vpp1904"] = vppcalls.HandlerVersion{
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

func IPToAddress(ipstr string) (addr ip.Address, err error) {
	netIP := net.ParseIP(ipstr)
	if netIP == nil {
		return ip.Address{}, fmt.Errorf("invalid IP: %q", ipstr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = ip.ADDRESS_IP6
		var ip6addr ip.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ip.ADDRESS_IP4
		var ip4addr ip.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}
