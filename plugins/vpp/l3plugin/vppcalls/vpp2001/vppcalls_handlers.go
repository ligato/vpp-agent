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
	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"

	corevppcalls "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	vpe_vpp2001 "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls/vpp2001"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001"
	vpp_dhcp "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/dhcp"
	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip"
	vpp_ip_neighbor "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_neighbor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/l3xc"
	vpp_vpe "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_ip.AllMessages()...)
	msgs = append(msgs, vpp_ip_neighbor.AllMessages()...)
	msgs = append(msgs, vpp_vpe.AllMessages()...)
	msgs = append(msgs, vpp_dhcp.AllMessages()...)

	vppcalls.AddHandlerVersion(vpp2001.Version, msgs, NewL3VppHandler)
}

type L3VppHandler struct {
	*ArpVppHandler
	*ProxyArpVppHandler
	*RouteHandler
	*IPNeighHandler
	*VrfTableHandler
	*DHCPProxyHandler
	*L3XCHandler
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
		ArpVppHandler:      NewArpVppHandler(ch, ifIdx, log),
		ProxyArpVppHandler: NewProxyArpVppHandler(ch, ifIdx, log),
		RouteHandler:       NewRouteVppHandler(ch, ifIdx, vrfIdx, addrAlloc, log),
		IPNeighHandler:     NewIPNeighVppHandler(ch, log),
		VrfTableHandler:    NewVrfTableVppHandler(ch, log),
		DHCPProxyHandler:   NewDHCPProxyHandler(ch, log),
		L3XCHandler:        NewL3XCHandler(c, ifIdx, log),
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
	vrfIndexes   vrfidx.VRFMetadataIndex
	addrAlloc    netalloc.AddressAllocator
	log          logging.Logger
}

// IPNeighHandler is accessor for ip-neighbor-related vppcalls methods
type IPNeighHandler struct {
	callsChannel govppapi.Channel
	log          logging.Logger
	corevppcalls.VppCoreAPI
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
	vrfIdx vrfidx.VRFMetadataIndex, addrAlloc netalloc.AddressAllocator, log logging.Logger) *RouteHandler {
	if log == nil {
		log = logrus.NewLogger("route-handler")
	}
	return &RouteHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		vrfIndexes:   vrfIdx,
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
		VppCoreAPI:   vpe_vpp2001.NewVpeHandler(callsChan),
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

type L3XCHandler struct {
	l3xc      l3xc.RPCService
	ifIndexes ifaceidx.IfaceMetadataIndex
	log       logging.Logger
}

// NewL3XCHandler creates new instance of L3XC vppcalls handler
func NewL3XCHandler(c vpp.Client, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger) *L3XCHandler {
	if log == nil {
		log = logrus.NewLogger("l3xc-handler")
	}
	h := &L3XCHandler{
		ifIndexes: ifIndexes,
		log:       log,
	}
	if c.IsPluginLoaded(l3xc.ModuleName) {
		ch, err := c.NewAPIChannel()
		if err != nil {
			logging.Warnf("creating channel failed: %v", err)
			return nil
		}
		h.l3xc = l3xc.NewServiceClient(ch)
	}
	return h
}

func ipToAddress(ipstr string) (addr vpp_ip.Address, err error) {
	netIP := net.ParseIP(ipstr)
	if netIP == nil {
		return vpp_ip.Address{}, fmt.Errorf("invalid IP: %q", ipstr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = ip_types.ADDRESS_IP6
		var ip6addr vpp_ip.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ip_types.ADDRESS_IP4
		var ip4addr vpp_ip.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}

func networkToPrefix(dstNetwork *net.IPNet) vpp_ip.Prefix {
	var addr vpp_ip.Address
	if dstNetwork.IP.To4() == nil {
		addr.Af = ip_types.ADDRESS_IP6
		var ip6addr vpp_ip.IP6Address
		copy(ip6addr[:], dstNetwork.IP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ip_types.ADDRESS_IP4
		var ip4addr vpp_ip.IP4Address
		copy(ip4addr[:], dstNetwork.IP.To4())
		addr.Un.SetIP4(ip4addr)
	}
	mask, _ := dstNetwork.Mask.Size()
	return vpp_ip.Prefix{
		Address: addr,
		Len:     uint8(mask),
	}
}

func uintToBool(value uint8) bool {
	if value == 0 {
		return false
	}
	return true
}
