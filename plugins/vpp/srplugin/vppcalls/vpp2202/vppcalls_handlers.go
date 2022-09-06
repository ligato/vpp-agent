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

package vpp2202

import (
	"fmt"
	"net"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	core_vppcalls "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	core_vpp2202 "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls/vpp2202"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	vpp2202 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/sr"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls"
)

func init() {
	msgs := vpp.Messages(
		sr.AllMessages,
		vpe.AllMessages, // using also vpe -> need to have correct vpp version also for vpe
	)
	vppcalls.AddHandlerVersion(vpp2202.Version, msgs.AllMessages(), NewSRv6VppHandler)
}

// SRv6VppHandler is accessor for SRv6-related vppcalls methods
type SRv6VppHandler struct {
	core_vppcalls.VppCoreAPI

	log          logging.Logger
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
}

// NewSRv6VppHandler creates new instance of SRv6 vppcalls handler
func NewSRv6VppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) vppcalls.SRv6VppAPI {
	vppChan, err := c.NewAPIChannel()
	if err != nil {
		logging.Warnf("failed to create API channel")
		return nil
	}
	return &SRv6VppHandler{
		callsChannel: vppChan,
		ifIndexes:    ifIdx,
		log:          log,
		VppCoreAPI:   core_vpp2202.NewVpeHandler(c),
	}
}

func addressToIP(address ip_types.Address) net.IP {
	if address.Af == ip_types.ADDRESS_IP6 {
		ipAddr := address.Un.GetIP6()
		return net.IP(ipAddr[:]).To16()
	}
	ipAddr := address.Un.GetIP4()
	return net.IP(ipAddr[:]).To4()
}

// parseIPv6 parses string <str> to IPv6 address (including IPv4 address converted to IPv6 address)
func parseIPv6(str string) (net.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return nil, fmt.Errorf(" %q is not ip address", str)
	}
	ipv6 := ip.To16()
	if ipv6 == nil {
		return nil, fmt.Errorf(" %q is not ipv6 address", str)
	}
	return ipv6, nil
}

func IPToAddress(ipstr string) (addr ip_types.Address, err error) {
	netIP := net.ParseIP(ipstr)
	if netIP == nil {
		return ip_types.Address{}, fmt.Errorf("invalid IP: %q", ipstr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = ip_types.ADDRESS_IP6
		var ip6addr ip_types.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ip_types.ADDRESS_IP4
		var ip4addr ip_types.IP4Address
		copy(ip4addr[:], ip4.To4())
		addr.Un.SetIP4(ip4addr)
	}
	return
}
