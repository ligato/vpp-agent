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

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, ipsec.Messages...)

	vppcalls.Versions["vpp1904"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) vppcalls.IPSecVppAPI {
			return NewIPSecVppHandler(ch, ifIdx, log)
		},
	}
}

// IPSecVppHandler is accessor for IPSec-related vppcalls methods
type IPSecVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

func NewIPSecVppHandler(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) *IPSecVppHandler {
	return &IPSecVppHandler{ch, ifIdx, log}
}

func ipsecAddrToIP(addr ipsec.Address) net.IP {
	if addr.Af == ipsec.ADDRESS_IP6 {
		addrIP := addr.Un.GetIP6()
		return net.IP(addrIP[:])
	} else {
		addrIP := addr.Un.GetIP4()
		return net.IP(addrIP[:])
	}
}

func IPToAddress(ipstr string) (addr ipsec.Address, err error) {
	netIP := net.ParseIP(ipstr)
	if netIP == nil {
		return ipsec.Address{}, fmt.Errorf("invalid IP: %q", ipstr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = ipsec.ADDRESS_IP6
		var ip6addr ipsec.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ipsec.ADDRESS_IP4
		var ip4addr ipsec.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}
