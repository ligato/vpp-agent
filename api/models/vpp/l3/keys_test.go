//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package vpp_l3

import (
	"testing"

	. "github.com/onsi/gomega"
)

/*func TestRouteKey(t *testing.T) {
	tests := []struct {
		name        string
		vrf         uint32
		dstNet      string
		nextHopAddr string
		expectedKey string
	}{
		{
			name:        "route-ipv4",
			vrf:         0,
			dstNet:      "10.10.0.0/24",
			nextHopAddr: "",
			expectedKey: "vpp/config/v2/route/vrf/0/dst/10.10.0.0/24/gw/0.0.0.0",
		},
		{
			name:        "dst-network-address",
			vrf:         0,
			dstNet:      "10.10.0.255/24",
			nextHopAddr: "",
			expectedKey: "vpp/config/v2/route/vrf/0/dst/10.10.0.0/24/gw/0.0.0.0",
		},
		{
			name:        "zero-next-hop",
			vrf:         0,
			dstNet:      "10.10.0.1/24",
			nextHopAddr: "0.0.0.0",
			expectedKey: "vpp/config/v2/route/vrf/0/dst/10.10.0.0/24/gw/0.0.0.0",
		},
		{
			name:        "non-zero-vrf",
			vrf:         1,
			dstNet:      "10.10.0.1/24",
			nextHopAddr: "0.0.0.0",
			expectedKey: "vpp/config/v2/route/vrf/1/dst/10.10.0.0/24/gw/0.0.0.0",
		},
		{
			name:        "invalid-dst-net-empty-gw",
			dstNet:      "INVALID",
			expectedKey: "vpp/config/v2/route/vrf/0/dst/<invalid>/<invalid>/gw/<invalid>",
		},
		{
			name:        "invalid-next-hop",
			dstNet:      "10.10.0.1/24",
			nextHopAddr: "INVALID",
			expectedKey: "vpp/config/v2/route/vrf/0/dst/10.10.0.0/24/gw/<invalid>",
		},
		{
			name:        "invalid-dst-net-valid-gw",
			dstNet:      "INVALID",
			nextHopAddr: "1.2.3.4",
			expectedKey: "vpp/config/v2/route/vrf/0/dst/<invalid>/<invalid>/gw/1.2.3.4",
		},
		{
			name:        "route-ipv6",
			dstNet:      "2001:DB8::0001/32",
			nextHopAddr: "",
			expectedKey: "vpp/config/v2/route/vrf/0/dst/2001:db8::/32/gw/::",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := RouteKey(test.vrf, test.dstNet, test.nextHopAddr)
			if key != test.expectedKey {
				t.Errorf("failed for: vrf=%d dstNet=%q nextHop=%q\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.vrf, test.dstNet, test.nextHopAddr, test.expectedKey, key)
			}
		})
	}
}
*/

// TestParseRouteKey test different cases for ParseRouteKey(...)
func TestParseRouteKey(t *testing.T) {
	tests := []struct {
		name                string
		routeKey            string
		expectedIsRouteKey  bool
		expectedVrfIndex    string
		expectedDstNetAddr  string
		expectedDstNetMask  int
		expectedNextHopAddr string
	}{
		{
			name:                "route-ipv4",
			routeKey:            "config/vpp/v2/route/vrf/0/dst/10.10.0.0/16/gw/0.0.0.0",
			expectedIsRouteKey:  true,
			expectedVrfIndex:    "0",
			expectedDstNetAddr:  "10.10.0.0",
			expectedDstNetMask:  16,
			expectedNextHopAddr: "0.0.0.0",
		},
		{
			name:                "route-ipv6",
			routeKey:            "config/vpp/v2/route/vrf/0/dst/2001:db8::/32/gw/::",
			expectedIsRouteKey:  true,
			expectedVrfIndex:    "0",
			expectedDstNetAddr:  "2001:db8::",
			expectedDstNetMask:  32,
			expectedNextHopAddr: "::",
		},
		{
			name:               "invalid-key",
			routeKey:           "config/vpp/v2/route/vrf/0/dst/2001:db8::/32/",
			expectedIsRouteKey: false,
		},
		{
			name:               "invalid-key-missing-dst",
			routeKey:           "config/vpp/v2/route/vrf/0/10.10.0.0/16/gw/0.0.0.0",
			expectedIsRouteKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RegisterTestingT(t)
			vrfIndex, dstNetAddr, dstNetMask, nextHopAddr, isRouteKey := ParseRouteKey(test.routeKey)
			Expect(isRouteKey).To(BeEquivalentTo(test.expectedIsRouteKey), "Route/Non-route key should be properly detected")
			if isRouteKey {
				Expect(vrfIndex).To(BeEquivalentTo(test.expectedVrfIndex), "VRF should be properly extracted by parsing route key")
				Expect(dstNetAddr).To(BeEquivalentTo(test.expectedDstNetAddr), "Destination network address should be properly extracted by parsing route key")
				Expect(dstNetMask).To(BeEquivalentTo(test.expectedDstNetMask), "Destination network mask should be properly extracted by parsing route key")
				Expect(nextHopAddr).To(BeEquivalentTo(test.expectedNextHopAddr), "Next hop address should be properly extracted by parsing route key")
			}
		})
	}
}
