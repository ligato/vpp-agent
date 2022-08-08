//  Copyright (c) 2021 Cisco and/or its affiliates.
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

package vpp2106

import (
	"fmt"
	"net"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	vpp2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_ipsec.AllMessages()...)

	vppcalls.AddHandlerVersion(vpp2106.Version, msgs, NewIPSecVppHandler)
}

// IPSecVppHandler is accessor for IPSec-related vppcalls methods
type IPSecVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

func NewIPSecVppHandler(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) vppcalls.IPSecVppAPI {
	return &IPSecVppHandler{ch, ifIdx, log}
}

func ipsecAddrToIP(addr ip_types.Address) net.IP {
	if addr.Af == ip_types.ADDRESS_IP6 {
		addrIP := addr.Un.GetIP6()
		return net.IP(addrIP[:])
	}
	addrIP := addr.Un.GetIP4()
	return net.IP(addrIP[:])
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
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}
