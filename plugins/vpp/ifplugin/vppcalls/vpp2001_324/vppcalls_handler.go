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

package vpp2001_324

import (
	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	vpp_afpacket "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/af_packet"
	vpp_bond "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/bond"
	vpp_dhcp "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/dhcp"
	vpp_gre "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/gre"
	vpp_ifs "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/interfaces"
	vpp_ip "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/ip"
	vpp_ipsec "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/ipsec"
	vpp_l2 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/l2"
	vpp_memif "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/memif"
	vpp_span "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/span"
	vpp_tapv2 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/tapv2"
	vpp_vmxnet3 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/vmxnet3"
	vpp_vpe "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/vpe"
	vpp_vxlan "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/vxlan"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_afpacket.AllMessages()...)
	msgs = append(msgs, vpp_bond.AllMessages()...)
	msgs = append(msgs, vpp_dhcp.AllMessages()...)
	msgs = append(msgs, vpp_gre.AllMessages()...)
	msgs = append(msgs, vpp_ifs.AllMessages()...)
	msgs = append(msgs, vpp_ip.AllMessages()...)
	msgs = append(msgs, vpp_ipsec.AllMessages()...)
	msgs = append(msgs, vpp_l2.AllMessages()...)
	msgs = append(msgs, vpp_memif.AllMessages()...)
	msgs = append(msgs, vpp_span.AllMessages()...)
	msgs = append(msgs, vpp_tapv2.AllMessages()...)
	msgs = append(msgs, vpp_vmxnet3.AllMessages()...)
	msgs = append(msgs, vpp_vpe.AllMessages()...)
	msgs = append(msgs, vpp_vxlan.AllMessages()...)

	vppcalls.Versions["vpp2001_324"] = vppcalls.HandlerVersion{
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

// IPToAddress converts string type IP address to VPP ip.api address representation
func IPToAddress(ipStr string) (addr vpp_ip.Address, err error) {
	netIP := net.ParseIP(ipStr)
	if netIP == nil {
		return vpp_ip.Address{}, fmt.Errorf("invalid IP: %q", ipStr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = vpp_ip.ADDRESS_IP6
		var ip6addr vpp_ip.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = vpp_ip.ADDRESS_IP4
		var ip4addr vpp_ip.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}
