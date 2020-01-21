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
	"context"
	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/pkg/errors"

	vpevppcalls "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	vpe_vpp1904 "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls/vpp1904"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
)

var (
	ErrUnsupported = errors.New("unsupported")
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, ip.AllMessages()...)
	msgs = append(msgs, vpe.AllMessages()...)
	msgs = append(msgs, dhcp.AllMessages()...)

	vppcalls.AddHandlerVersion(vpp1904.Version, msgs, NewL3VppHandler)
}

type L3VppHandler struct {
	*ArpVppHandler
	*ProxyArpVppHandler
	*RouteHandler
	*IPNeighHandler
	*VrfTableHandler
	*DHCPProxyHandler
	*L3XCHandlerUnsupported
}

func NewL3VppHandler(
	c vpp.Client,
	ifIdx ifaceidx.IfaceMetadataIndex,
	vrfIdx vrfidx.VRFMetadataIndex,
	addrAlloc netalloc.AddressAllocator,
	log logging.Logger,
) vppcalls.L3VppAPI {
	ch, err := c.NewAPIChannel()
	if err != nil {
		logging.Warnf("creating channel failed: %v", err)
		return nil
	}
	return &L3VppHandler{
		ArpVppHandler:          NewArpVppHandler(ch, ifIdx, log),
		ProxyArpVppHandler:     NewProxyArpVppHandler(ch, ifIdx, log),
		RouteHandler:           NewRouteVppHandler(ch, ifIdx, addrAlloc, log),
		IPNeighHandler:         NewIPNeighVppHandler(ch, log),
		VrfTableHandler:        NewVrfTableVppHandler(ch, log),
		DHCPProxyHandler:       NewDHCPProxyHandler(ch, log),
		L3XCHandlerUnsupported: &L3XCHandlerUnsupported{},
	}
}

// ArpVppHandler is accessor for ARP-related vppcalls methods
type ArpVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// DHCPProxyHandler is accessor for DHCP proxy-related vppcalls methods
type DHCPProxyHandler struct {
	callsChannel govppapi.Channel
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
	addrAlloc    netalloc.AddressAllocator
	log          logging.Logger
}

// IPNeighHandler is accessor for ip-neighbor-related vppcalls methods
type IPNeighHandler struct {
	callsChannel govppapi.Channel
	log          logging.Logger
	vpevppcalls.VppCoreAPI
}

// VrfTableHandler is accessor for vrf-related vppcalls methods
type VrfTableHandler struct {
	callsChannel govppapi.Channel
	log          logging.Logger
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
func NewRouteVppHandler(callsChan govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex,
	addrAlloc netalloc.AddressAllocator, log logging.Logger) *RouteHandler {
	if log == nil {
		log = logrus.NewLogger("route-handler")
	}
	return &RouteHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		addrAlloc:    addrAlloc,
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
		VppCoreAPI:   vpe_vpp1904.NewVpeHandler(callsChan),
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

// NewDHCPProxyHandler creates new instance of vrf-table vppcalls handler
func NewDHCPProxyHandler(callsChan govppapi.Channel, log logging.Logger) *DHCPProxyHandler {
	if log == nil {
		log = logrus.NewLogger("dhcp-proxy-handler")
	}
	return &DHCPProxyHandler{
		callsChannel: callsChan,
		log:          log,
	}
}

type L3XCHandlerUnsupported struct{}

func (l L3XCHandlerUnsupported) DumpAllL3XC(ctx context.Context) ([]vppcalls.L3XC, error) {
	return nil, ErrUnsupported
}

func (l L3XCHandlerUnsupported) DumpL3XC(ctx context.Context, index uint32) ([]vppcalls.L3XC, error) {
	return nil, ErrUnsupported
}

func (l L3XCHandlerUnsupported) UpdateL3XC(ctx context.Context, l3xc *vppcalls.L3XC) error {
	return ErrUnsupported
}

func (l L3XCHandlerUnsupported) DeleteL3XC(ctx context.Context, index uint32, ipv6 bool) error {
	return ErrUnsupported
}

func ipToAddress(ipstr string) (addr ip.Address, err error) {
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
