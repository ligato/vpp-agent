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

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

func TestRouteKey(t *testing.T) {
	tests := []struct {
		name        string
		route       Route
		expectedKey string
	}{
		{
			"route-ipv4",
			Route{
				VrfId:             0,
				DstNetwork:        "10.10.0.0/24",
				NextHopAddr:       "0.0.0.0",
				OutgoingInterface: "",
			},
			"config/vpp/v2/route/vrf/0/dst/10.10.0.0/24/gw/0.0.0.0",
		},
		{
			"route-ipv6",
			Route{
				VrfId:             0,
				DstNetwork:        "2001:DB8::0001/32",
				NextHopAddr:       "0.0.0.0",
				OutgoingInterface: "",
			},
			"config/vpp/v2/route/vrf/0/dst/2001:db8::/32/gw/0.0.0.0",
		},
		{
			"route-ipv4-interface",
			Route{
				VrfId:             0,
				DstNetwork:        "10.10.0.0/24",
				NextHopAddr:       "0.0.0.0",
				OutgoingInterface: "iface1",
			},
			"config/vpp/v2/route/if/iface1/vrf/0/dst/10.10.0.0/24/gw/0.0.0.0",
		},
		{
			"route-ipv6-interface",
			Route{
				VrfId:             0,
				DstNetwork:        "2001:DB8::0001/32",
				NextHopAddr:       "0.0.0.0",
				OutgoingInterface: "iface1",
			},
			"config/vpp/v2/route/if/iface1/vrf/0/dst/2001:db8::/32/gw/0.0.0.0",
		},
		{
			"route-invalid-ip",
			Route{
				VrfId:             0,
				DstNetwork:        "INVALID",
				NextHopAddr:       "0.0.0.0",
				OutgoingInterface: "",
			},
			"config/vpp/v2/route/vrf/0/dst/<invalid>/0/gw/0.0.0.0",
		},
		{
			"route-invalid-gw",
			Route{
				VrfId:             0,
				DstNetwork:        "10.10.10.0/32",
				NextHopAddr:       "INVALID",
				OutgoingInterface: "",
			},
			"config/vpp/v2/route/vrf/0/dst/10.10.10.0/32/gw/INVALID",
		},
		{
			"route-dstnetwork",
			Route{
				VrfId:             0,
				DstNetwork:        "10.10.0.5/24",
				NextHopAddr:       "0.0.0.0",
				OutgoingInterface: "",
			},
			"config/vpp/v2/route/vrf/0/dst/10.10.0.0/24/gw/0.0.0.0",
		},
		{
			"route-gw-empty",
			Route{
				VrfId:             0,
				DstNetwork:        "10.0.0.0/8",
				NextHopAddr:       "",
				OutgoingInterface: "",
			},
			"config/vpp/v2/route/vrf/0/dst/10.0.0.0/8",
		},
		{
			"route-vrf",
			Route{
				VrfId:             3,
				DstNetwork:        "10.0.0.0/8",
				NextHopAddr:       "",
				OutgoingInterface: "",
			},
			"config/vpp/v2/route/vrf/3/dst/10.0.0.0/8",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := models.Key(&test.route)
			if key != test.expectedKey {
				t.Errorf("failed key for route: %+v\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.route, test.expectedKey, key)
			}
		})
	}
}

// TestParseRouteKey test different cases for ParseRouteKey(...)
func TestParseRouteKey(t *testing.T) {
	tests := []struct {
		name                string
		routeKey            string
		expectedIsRouteKey  bool
		expectedOutIface    string
		expectedVrfIndex    string
		expectedDstNet      string
		expectedNextHopAddr string
	}{
		{
			name:                "route-ipv4",
			routeKey:            "config/vpp/v2/route/vrf/0/dst/10.10.0.0/16/gw/0.0.0.0",
			expectedIsRouteKey:  true,
			expectedVrfIndex:    "0",
			expectedDstNet:      "10.10.0.0/16",
			expectedNextHopAddr: "0.0.0.0",
		},
		{
			name:                "route-ipv4 with interface",
			routeKey:            "config/vpp/v2/route/if/Gbe0/8/0/vrf/0/dst/10.10.0.0/16/gw/0.0.0.0",
			expectedIsRouteKey:  true,
			expectedOutIface:    "Gbe0/8/0",
			expectedVrfIndex:    "0",
			expectedDstNet:      "10.10.0.0/16",
			expectedNextHopAddr: "0.0.0.0",
		},
		{
			name:                "route-ipv6",
			routeKey:            "config/vpp/v2/route/vrf/0/dst/2001:db8::/32/gw/::",
			expectedIsRouteKey:  true,
			expectedVrfIndex:    "0",
			expectedDstNet:      "2001:db8::/32",
			expectedNextHopAddr: "::",
		},
		{
			name:               "undefined interface and GW",
			routeKey:           "config/vpp/v2/route/vrf/0/dst/2001:db8::/32/",
			expectedIsRouteKey: true,
			expectedVrfIndex:   "0",
			expectedDstNet:     "2001:db8::/32",
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
			outIface, vrfIndex, dstNet, nextHopAddr, isRouteKey := ParseRouteKey(test.routeKey)
			Expect(isRouteKey).To(BeEquivalentTo(test.expectedIsRouteKey), "Route/Non-route key should be properly detected")
			if isRouteKey {
				Expect(outIface).To(BeEquivalentTo(test.expectedOutIface), "outgoing interface should be properly extracted by parsing route key")
				Expect(vrfIndex).To(BeEquivalentTo(test.expectedVrfIndex), "VRF should be properly extracted by parsing route key")
				Expect(dstNet).To(BeEquivalentTo(test.expectedDstNet), "Destination network should be properly extracted by parsing route key")
				Expect(nextHopAddr).To(BeEquivalentTo(test.expectedNextHopAddr), "Next hop address should be properly extracted by parsing route key")
			}
		})
	}
}
