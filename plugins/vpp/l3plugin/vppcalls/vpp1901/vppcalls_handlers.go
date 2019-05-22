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
	"github.com/ligato/cn-infra/logging/logrus"
	vpevppcalls "github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls/vpp1901"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, ip.Messages...)
	msgs = append(msgs, vpe.Messages...)
	msgs = append(msgs, dhcp.Messages...)

	vppcalls.Versions["vpp1901"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger,
		) vppcalls.L3VppAPI {
			return NewL3VppHandler(ch, ifIdx, log)
		},
	}
}

type L3VppHandler struct {
	*ArpVppHandler
	*ProxyArpVppHandler
	*RouteHandler
	*IPNeighHandler
	*VrfTableHandler
	*DHCPProxyHandler
}

func NewL3VppHandler(
	ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger,
) *L3VppHandler {
	return &L3VppHandler{
		ArpVppHandler:      NewArpVppHandler(ch, ifIdx, log),
		ProxyArpVppHandler: NewProxyArpVppHandler(ch, ifIdx, log),
		RouteHandler:       NewRouteVppHandler(ch, ifIdx, log),
		IPNeighHandler:     NewIPNeighVppHandler(ch, log),
		VrfTableHandler:    NewVrfTableVppHandler(ch, log),
		DHCPProxyHandler:   NewDHCPProxyHandler(ch, log),
	}
}

// ArpVppHandler is accessor for ARP-related vppcalls methods
type ArpVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// ProxyArpVppHandler is accessor for proxy ARP-related vppcalls methods
type ProxyArpVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// RouteHandler is accessor for route-related vppcalls methods
type RouteHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// IPNeighHandler is accessor for ip-neighbor-related vppcalls methods
type IPNeighHandler struct {
	callsChannel govppapi.Channel
	log          logging.Logger
	vpevppcalls.VpeVppAPI
}

// VrfTableHandler is accessor for vrf-related vppcalls methods
type VrfTableHandler struct {
	callsChannel govppapi.Channel
	log          logging.Logger
}

// DHCPProxyHandler is accessor for DHCP proxy-related vppcalls methods
type DHCPProxyHandler struct {
	callsChannel govppapi.Channel
	log          logging.Logger
}

// NewVrfTableVppHandler creates new instance of vrf-table vppcalls handler
func NewDHCPProxyHandler(callsChan govppapi.Channel, log logging.Logger) *DHCPProxyHandler {
	if log == nil {
		log = logrus.NewLogger("dhcp-proxy-handler")
	}
	return &DHCPProxyHandler{
		callsChannel: callsChan,
		log:          log,
	}
}

// NewArpVppHandler creates new instance of IPsec vppcalls handler
func NewArpVppHandler(callsChan govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger) *ArpVppHandler {
	if log == nil {
		log = logrus.NewLogger("arp-handler")
	}
	return &ArpVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}

// NewProxyArpVppHandler creates new instance of proxy ARP vppcalls handler
func NewProxyArpVppHandler(callsChan govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger) *ProxyArpVppHandler {
	if log == nil {
		log = logrus.NewLogger("proxy-arp-handler")
	}
	return &ProxyArpVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}

// NewRouteVppHandler creates new instance of route vppcalls handler
func NewRouteVppHandler(callsChan govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger) *RouteHandler {
	if log == nil {
		log = logrus.NewLogger("route-handler")
	}
	return &RouteHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}

// NewIPNeighVppHandler creates new instance of ip neighbor vppcalls handler
func NewIPNeighVppHandler(callsChan govppapi.Channel, log logging.Logger) *IPNeighHandler {
	if log == nil {
		log = logrus.NewLogger("ip-neigh")
	}
	return &IPNeighHandler{
		callsChannel: callsChan,
		log:          log,
		VpeVppAPI:    vpp1901.NewVpeHandler(callsChan),
	}
}

// NewVrfTableVppHandler creates new instance of vrf-table vppcalls handler
func NewVrfTableVppHandler(callsChan govppapi.Channel, log logging.Logger) *VrfTableHandler {
	if log == nil {
		log = logrus.NewLogger("vrf-table-handler")
	}
	return &VrfTableHandler{
		callsChannel: callsChan,
		log:          log,
	}
}

func uintToBool(value uint8) bool {
	if value == 0 {
		return false
	}
	return true
}

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}
