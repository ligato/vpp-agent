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
	"bytes"
	"fmt"
	"net"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/ip"
)

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

func uintToBool(value uint8) bool {
	return value != 0
}

func cleanString(b []byte) string {
	return string(bytes.SplitN(b, []byte{0x00}, 2)[0])
}
