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

package linux_interfaces

import (
	"testing"

	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
)

func TestInterfaceAddressKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		address     string
		source      netalloc.IPAddressSource
		expectedKey string
	}{
		{
			name:        "IPv4 address",
			iface:       "memif0",
			address:     "192.168.1.12/24",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "linux/interface/memif0/address/static/192.168.1.12/24",
		},
		{
			name:        "IPv4 address from DHCP",
			iface:       "memif0",
			address:     "192.168.1.12/24",
			source:      netalloc.IPAddressSource_FROM_DHCP,
			expectedKey: "linux/interface/memif0/address/from_dhcp/192.168.1.12/24",
		},
		{
			name:        "IPv6 address",
			iface:       "memif0",
			address:     "2001:db8::/32",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "linux/interface/memif0/address/static/2001:db8::/32",
		},
		{
			name:        "IPv6 address from DHCP",
			iface:       "memif0",
			address:     "2001:db8::/32",
			source:      netalloc.IPAddressSource_FROM_DHCP,
			expectedKey: "linux/interface/memif0/address/from_dhcp/2001:db8::/32",
		},
		{
			name:        "invalid interface",
			iface:       "",
			address:     "10.10.10.10/32",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "linux/interface/<invalid>/address/static/10.10.10.10/32",
		},
		{
			name:        "undefined source",
			iface:       "memif1",
			address:     "10.10.10.10/32",
			expectedKey: "linux/interface/memif1/address/undefined_source/10.10.10.10/32",
		},
		{
			name:        "invalid address",
			iface:       "tap0",
			address:     "invalid-addr",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "linux/interface/tap0/address/static/invalid-addr",
		},
		{
			name:        "missing mask",
			iface:       "tap1",
			address:     "10.10.10.10",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "linux/interface/tap1/address/static/10.10.10.10",
		},
		{
			name:        "empty address",
			iface:       "tap1",
			address:     "",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "linux/interface/tap1/address/static/",
		},
		{
			name:        "IPv4 address requested from netalloc",
			iface:       "memif0",
			address:     "alloc:net1",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "linux/interface/memif0/address/alloc_ref/alloc:net1",
		},
		{
			name:        "IPv6 address requested from netalloc",
			iface:       "memif0",
			address:     "alloc:net1/IPV6_ADDR",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "linux/interface/memif0/address/alloc_ref/alloc:net1/IPV6_ADDR",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := InterfaceAddressKey(test.iface, test.address, test.source)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s address=%s source=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.address, string(test.source), test.expectedKey, key)
			}
		})
	}
}

func TestParseInterfaceAddressKey(t *testing.T) {
	tests := []struct {
		name               string
		key                string
		expectedIface      string
		expectedIfaceAddr  string
		expectedSource     netalloc.IPAddressSource
		expectedInvalidKey bool
		expectedIsAddrKey  bool
	}{
		{
			name:              "IPv4 address",
			key:               "linux/interface/memif0/address/static/192.168.1.12/24",
			expectedIface:     "memif0",
			expectedIfaceAddr: "192.168.1.12/24",
			expectedSource:    netalloc.IPAddressSource_STATIC,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv4 address from DHCP",
			key:               "linux/interface/memif0/address/from_dhcp/192.168.1.12/24",
			expectedIface:     "memif0",
			expectedIfaceAddr: "192.168.1.12/24",
			expectedSource:    netalloc.IPAddressSource_FROM_DHCP,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv4 address requested from Netalloc",
			key:               "linux/interface/memif0/address/alloc_ref/alloc:net1",
			expectedIface:     "memif0",
			expectedIfaceAddr: "alloc:net1",
			expectedSource:    netalloc.IPAddressSource_ALLOC_REF,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv6 address",
			key:               "linux/interface/tap1/address/static/2001:db8:85a3::8a2e:370:7334/48",
			expectedIface:     "tap1",
			expectedIfaceAddr: "2001:db8:85a3::8a2e:370:7334/48",
			expectedSource:    netalloc.IPAddressSource_STATIC,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv6 address requested from netalloc",
			key:               "linux/interface/tap1/address/alloc_ref/alloc:net1/IPV6_ADDR",
			expectedIface:     "tap1",
			expectedIfaceAddr: "alloc:net1/IPV6_ADDR",
			expectedSource:    netalloc.IPAddressSource_ALLOC_REF,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv6 address from DHCP",
			key:               "linux/interface/tap1/address/from_dhcp/2001:db8:85a3::8a2e:370:7334/48",
			expectedIface:     "tap1",
			expectedIfaceAddr: "2001:db8:85a3::8a2e:370:7334/48",
			expectedSource:    netalloc.IPAddressSource_FROM_DHCP,
			expectedIsAddrKey: true,
		},
		{
			name:              "invalid interface",
			key:               "linux/interface/<invalid>/address/static/10.10.10.10/30",
			expectedIface:     "<invalid>",
			expectedIfaceAddr: "10.10.10.10/30",
			expectedSource:    netalloc.IPAddressSource_STATIC,
			expectedIsAddrKey: true,
		},
		{
			name:              "gbe interface",
			key:               "linux/interface/GigabitEthernet0/8/0/address/static/192.168.5.5/16",
			expectedIface:     "GigabitEthernet0/8/0",
			expectedIfaceAddr: "192.168.5.5/16",
			expectedSource:    netalloc.IPAddressSource_STATIC,
			expectedIsAddrKey: true,
		},
		{
			name:               "missing interface",
			key:                "linux/interface//address/static/192.168.5.5/16",
			expectedIface:      "<invalid>",
			expectedIfaceAddr:  "192.168.5.5/16",
			expectedSource:     netalloc.IPAddressSource_STATIC,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing interface (from DHCP)",
			key:                "linux/interface//address/from_dhcp/192.168.5.5/16",
			expectedIface:      "<invalid>",
			expectedIfaceAddr:  "192.168.5.5/16",
			expectedSource:     netalloc.IPAddressSource_FROM_DHCP,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing IP",
			key:                "linux/interface/tap3/address/static/",
			expectedIface:      "tap3",
			expectedIfaceAddr:  "",
			expectedSource:     netalloc.IPAddressSource_STATIC,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing IP (from DHCP)",
			key:                "linux/interface/tap3/address/from_dhcp/",
			expectedIface:      "tap3",
			expectedIfaceAddr:  "",
			expectedSource:     netalloc.IPAddressSource_FROM_DHCP,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing IP for Gbe",
			key:                "linux/interface/Gbe0/1/2/address/static/",
			expectedIface:      "Gbe0/1/2",
			expectedIfaceAddr:  "",
			expectedSource:     netalloc.IPAddressSource_STATIC,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:              "not interface address key",
			key:               "linux/config/v2/interface/GigabitEthernet0/8/0",
			expectedIface:     "",
			expectedIfaceAddr: "",
			expectedIsAddrKey: false,
		},
		{
			name:               "invalid address source",
			key:                "linux/interface/memif0/address/<invalid>/192.168.1.12/24",
			expectedIface:      "memif0",
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "empty address source",
			key:                "linux/interface/memif0/address//192.168.1.12/24",
			expectedIface:      "memif0",
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing address source",
			key:                "linux/interface/memif0/address/192.168.1.12/24",
			expectedIface:      "memif0",
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, ipAddr, source, invalidKey, isAddrKey := ParseInterfaceAddressKey(test.key)
			if isAddrKey != test.expectedIsAddrKey {
				t.Errorf("expected isAddrKey: %v\tgot: %v", test.expectedIsAddrKey, isAddrKey)
			}
			if source != test.expectedSource {
				t.Errorf("expected source: %v\tgot: %v", test.expectedSource, source)
			}
			if invalidKey != test.expectedInvalidKey {
				t.Errorf("expected invalidKey: %v\tgot: %v", test.expectedInvalidKey, invalidKey)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
			if ipAddr != test.expectedIfaceAddr {
				t.Errorf("expected ipAddr: %s\tgot: %s", test.expectedIfaceAddr, ipAddr)
			}
		})
	}
}
