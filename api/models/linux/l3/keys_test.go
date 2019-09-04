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

package linux_l3

import (
	"testing"
)

func TestRouteKey(t *testing.T) {
	tests := []struct {
		name        string
		outIface    string
		dstNetwork  string
		expectedKey string
	}{
		{
			name:        "IPv4 dest address",
			outIface:    "memif1",
			dstNetwork:  "192.168.1.0/24",
			expectedKey: "config/linux/l3/v2/route/192.168.1.0/24/memif1",
		},
		{
			name:        "dest address obtained from netalloc",
			outIface:    "memif1",
			dstNetwork:  "alloc:net1/memif2",
			expectedKey: "config/linux/l3/v2/route/alloc:net1/memif2/memif1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := RouteKey(test.dstNetwork, test.outIface)
			if key != test.expectedKey {
				t.Errorf("failed for: outIface=%s dstNet=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.outIface, test.dstNetwork, test.expectedKey, key)
			}
		})
	}
}

func TestStaticLinkLocalRouteKey(t *testing.T) {
	tests := []struct {
		name        string
		dstAddr     string
		outIface    string
		expectedKey string
	}{
		{
			name:        "IPv4 address via memif",
			outIface:    "memif0",
			dstAddr:     "192.168.1.12/24",
			expectedKey: "linux/link-local-route/memif0/dest-address/192.168.1.12/24",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := StaticLinkLocalRouteKey(test.dstAddr, test.outIface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s address=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.outIface, test.dstAddr, test.expectedKey, key)
			}
		})
	}
}

func TestParseStaticLinkLocalRouteKey(t *testing.T) {
	tests := []struct {
		name                        string
		key                         string
		expectedIface               string
		expectedDstAddr             string
		expectedIsLinkLocalRouteKey bool
	}{
		{
			name:                        "IPv4 address via memif",
			key:                         "linux/link-local-route/memif0/dest-address/192.168.1.12/24",
			expectedIface:               "memif0",
			expectedDstAddr:             "192.168.1.12/24",
			expectedIsLinkLocalRouteKey: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dstAddr, outIface, isLinkLocalRouteKey := ParseStaticLinkLocalRouteKey(test.key)
			if isLinkLocalRouteKey != test.expectedIsLinkLocalRouteKey {
				t.Errorf("expected isLinkLocalRouteKey: %v\tgot: %v", test.expectedIsLinkLocalRouteKey, isLinkLocalRouteKey)
			}
			if outIface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, outIface)
			}
			if dstAddr != test.expectedDstAddr {
				t.Errorf("expected dstAddr: %s\tgot: %s", test.expectedDstAddr, dstAddr)
			}
		})
	}
}
