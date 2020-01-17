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

package vpp_interfaces

import (
	"testing"

	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
)

/*func TestInterfaceKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		expectedKey string
	}{
		{
			name:        "valid interface name",
			iface:       "memif0",
			expectedKey: "vpp/config/v2/interface/memif0",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			expectedKey: "vpp/config/v2/interface/<invalid>",
		},
		{
			name:        "Gbe interface",
			iface:       "GigabitEthernet0/8/0",
			expectedKey: "vpp/config/v2/interface/GigabitEthernet0/8/0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := InterfaceKey(test.iface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.expectedKey, key)
			}
		})
	}
}

func TestParseNameFromKey(t *testing.T) {
	tests := []struct {
		name               string
		key                string
		expectedIface      string
		expectedIsIfaceKey bool
	}{
		{
			name:               "valid interface name",
			key:                "vpp/config/v2/interface/memif0",
			expectedIface:      "memif0",
			expectedIsIfaceKey: true,
		},
		{
			name:               "invalid interface name",
			key:                "vpp/config/v2/interface/<invalid>",
			expectedIface:      "<invalid>",
			expectedIsIfaceKey: true,
		},
		{
			name:               "Gbe interface",
			key:                "vpp/config/v2/interface/GigabitEthernet0/8/0",
			expectedIface:      "GigabitEthernet0/8/0",
			expectedIsIfaceKey: true,
		},
		{
			name:               "not an interface key",
			key:                "vpp/config/v2/bd/bd1",
			expectedIface:      "",
			expectedIsIfaceKey: false,
		},
		{
			name:               "not an interface key (empty interface)",
			key:                "vpp/config/v2/interface/",
			expectedIface:      "",
			expectedIsIfaceKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, isInterfaceKey := models.Model(&Interface{}).ParseKey(test.key)
			if isInterfaceKey != test.expectedIsIfaceKey {
				t.Errorf("expected isInterfaceKey: %v\tgot: %v", test.expectedIsIfaceKey, isInterfaceKey)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
		})
	}
}*/

func TestInterfaceErrorKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		expectedKey string
	}{
		{
			name:        "valid interface name",
			iface:       "memif0",
			expectedKey: "vpp/status/v2/interface/error/memif0",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			expectedKey: "vpp/status/v2/interface/error/<invalid>",
		},
		{
			name:        "Gbe interface",
			iface:       "GigabitEthernet0/8/0",
			expectedKey: "vpp/status/v2/interface/error/GigabitEthernet0/8/0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := InterfaceErrorKey(test.iface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.expectedKey, key)
			}
		})
	}
}

func TestInterfaceStateKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		expectedKey string
	}{
		{
			name:        "valid interface name",
			iface:       "memif0",
			expectedKey: "vpp/status/v2/interface/memif0",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			expectedKey: "vpp/status/v2/interface/<invalid>",
		},
		{
			name:        "Gbe interface",
			iface:       "GigabitEthernet0/8/0",
			expectedKey: "vpp/status/v2/interface/GigabitEthernet0/8/0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := InterfaceStateKey(test.iface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.expectedKey, key)
			}
		})
	}
}

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
			expectedKey: "vpp/interface/memif0/address/static/192.168.1.12/24",
		},
		{
			name:        "IPv4 address from DHCP",
			iface:       "memif0",
			address:     "192.168.1.12/24",
			source:      netalloc.IPAddressSource_FROM_DHCP,
			expectedKey: "vpp/interface/memif0/address/from_dhcp/192.168.1.12/24",
		},
		{
			name:        "IPv6 address",
			iface:       "memif0",
			address:     "2001:db8::/32",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "vpp/interface/memif0/address/static/2001:db8::/32",
		},
		{
			name:        "IPv6 address from DHCP",
			iface:       "memif0",
			address:     "2001:db8::/32",
			source:      netalloc.IPAddressSource_FROM_DHCP,
			expectedKey: "vpp/interface/memif0/address/from_dhcp/2001:db8::/32",
		},
		{
			name:        "invalid interface",
			iface:       "",
			address:     "10.10.10.10/32",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "vpp/interface/<invalid>/address/static/10.10.10.10/32",
		},
		{
			name:        "undefined source",
			iface:       "memif1",
			address:     "10.10.10.10/32",
			expectedKey: "vpp/interface/memif1/address/undefined_source/10.10.10.10/32",
		},
		{
			name:        "invalid address",
			iface:       "tap0",
			address:     "invalid-addr",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "vpp/interface/tap0/address/static/invalid-addr",
		},
		{
			name:        "missing mask",
			iface:       "tap1",
			address:     "10.10.10.10",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "vpp/interface/tap1/address/static/10.10.10.10",
		},
		{
			name:        "empty address",
			iface:       "tap1",
			address:     "",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "vpp/interface/tap1/address/static/",
		},
		{
			name:        "IPv4 address requested from netalloc",
			iface:       "memif0",
			address:     "alloc:net1",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "vpp/interface/memif0/address/alloc_ref/alloc:net1",
		},
		{
			name:        "IPv6 address requested from netalloc",
			iface:       "memif0",
			address:     "alloc:net1/IPV6_ADDR",
			source:      netalloc.IPAddressSource_STATIC,
			expectedKey: "vpp/interface/memif0/address/alloc_ref/alloc:net1/IPV6_ADDR",
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
			key:               "vpp/interface/memif0/address/static/192.168.1.12/24",
			expectedIface:     "memif0",
			expectedIfaceAddr: "192.168.1.12/24",
			expectedSource:    netalloc.IPAddressSource_STATIC,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv4 address from DHCP",
			key:               "vpp/interface/memif0/address/from_dhcp/192.168.1.12/24",
			expectedIface:     "memif0",
			expectedIfaceAddr: "192.168.1.12/24",
			expectedSource:    netalloc.IPAddressSource_FROM_DHCP,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv4 address requested from Netalloc",
			key:               "vpp/interface/memif0/address/alloc_ref/alloc:net1",
			expectedIface:     "memif0",
			expectedIfaceAddr: "alloc:net1",
			expectedSource:    netalloc.IPAddressSource_ALLOC_REF,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv6 address",
			key:               "vpp/interface/tap1/address/static/2001:db8:85a3::8a2e:370:7334/48",
			expectedIface:     "tap1",
			expectedIfaceAddr: "2001:db8:85a3::8a2e:370:7334/48",
			expectedSource:    netalloc.IPAddressSource_STATIC,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv6 address requested from netalloc",
			key:               "vpp/interface/tap1/address/alloc_ref/alloc:net1/IPV6_ADDR",
			expectedIface:     "tap1",
			expectedIfaceAddr: "alloc:net1/IPV6_ADDR",
			expectedSource:    netalloc.IPAddressSource_ALLOC_REF,
			expectedIsAddrKey: true,
		},
		{
			name:              "IPv6 address from DHCP",
			key:               "vpp/interface/tap1/address/from_dhcp/2001:db8:85a3::8a2e:370:7334/48",
			expectedIface:     "tap1",
			expectedIfaceAddr: "2001:db8:85a3::8a2e:370:7334/48",
			expectedSource:    netalloc.IPAddressSource_FROM_DHCP,
			expectedIsAddrKey: true,
		},
		{
			name:              "invalid interface",
			key:               "vpp/interface/<invalid>/address/static/10.10.10.10/30",
			expectedIface:     "<invalid>",
			expectedIfaceAddr: "10.10.10.10/30",
			expectedSource:    netalloc.IPAddressSource_STATIC,
			expectedIsAddrKey: true,
		},
		{
			name:              "gbe interface",
			key:               "vpp/interface/GigabitEthernet0/8/0/address/static/192.168.5.5/16",
			expectedIface:     "GigabitEthernet0/8/0",
			expectedIfaceAddr: "192.168.5.5/16",
			expectedSource:    netalloc.IPAddressSource_STATIC,
			expectedIsAddrKey: true,
		},
		{
			name:               "missing interface",
			key:                "vpp/interface//address/static/192.168.5.5/16",
			expectedIface:      "<invalid>",
			expectedIfaceAddr:  "192.168.5.5/16",
			expectedSource:     netalloc.IPAddressSource_STATIC,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing interface (from DHCP)",
			key:                "vpp/interface//address/from_dhcp/192.168.5.5/16",
			expectedIface:      "<invalid>",
			expectedIfaceAddr:  "192.168.5.5/16",
			expectedSource:     netalloc.IPAddressSource_FROM_DHCP,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing IP",
			key:                "vpp/interface/tap3/address/static/",
			expectedIface:      "tap3",
			expectedIfaceAddr:  "",
			expectedSource:     netalloc.IPAddressSource_STATIC,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing IP (from DHCP)",
			key:                "vpp/interface/tap3/address/from_dhcp/",
			expectedIface:      "tap3",
			expectedIfaceAddr:  "",
			expectedSource:     netalloc.IPAddressSource_FROM_DHCP,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing IP for Gbe",
			key:                "vpp/interface/Gbe0/1/2/address/static/",
			expectedIface:      "Gbe0/1/2",
			expectedIfaceAddr:  "",
			expectedSource:     netalloc.IPAddressSource_STATIC,
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:              "not interface address key",
			key:               "vpp/config/v2/interface/GigabitEthernet0/8/0",
			expectedIface:     "",
			expectedIfaceAddr: "",
			expectedIsAddrKey: false,
		},
		{
			name:               "invalid address source",
			key:                "vpp/interface/memif0/address/<invalid>/192.168.1.12/24",
			expectedIface:      "memif0",
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "empty address source",
			key:                "vpp/interface/memif0/address//192.168.1.12/24",
			expectedIface:      "memif0",
			expectedInvalidKey: true,
			expectedIsAddrKey:  true,
		},
		{
			name:               "missing address source",
			key:                "vpp/interface/memif0/address/192.168.1.12/24",
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

func TestInterfaceVrfTableKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		vrf         int
		ipv4        bool
		ipv6        bool
		expectedKey string
	}{
		{
			name:        "default IPv4 VRF",
			iface:       "memif0",
			vrf:         0,
			ipv4:        true,
			expectedKey: "vpp/interface/memif0/vrf/0/ip-version/v4",
		},
		{
			name:        "default IPv6 VRF",
			iface:       "memif0",
			vrf:         0,
			ipv6:        true,
			expectedKey: "vpp/interface/memif0/vrf/0/ip-version/v6",
		},
		{
			name:        "default VRF for both versions",
			iface:       "memif0",
			vrf:         0,
			ipv4:        true,
			ipv6:        true,
			expectedKey: "vpp/interface/memif0/vrf/0/ip-version/both",
		},
		{
			name:        "IPv4 VRF 1",
			iface:       "memif0",
			vrf:         1,
			ipv4:        true,
			expectedKey: "vpp/interface/memif0/vrf/1/ip-version/v4",
		},
		{
			name:        "IPv6 VRF 1",
			iface:       "memif0",
			vrf:         1,
			ipv6:        true,
			expectedKey: "vpp/interface/memif0/vrf/1/ip-version/v6",
		},
		{
			name:        "interface name with forward slashes",
			iface:       "Gbe0/2/1",
			vrf:         10,
			ipv4:        true,
			expectedKey: "vpp/interface/Gbe0/2/1/vrf/10/ip-version/v4",
		},
		{
			name:        "missing interface name",
			iface:       "",
			vrf:         10,
			ipv4:        true,
			expectedKey: "vpp/interface/<invalid>/vrf/10/ip-version/v4",
		},
		{
			name:        "undefined version",
			iface:       "memif0",
			vrf:         10,
			expectedKey: "vpp/interface/memif0/vrf/10/ip-version/<invalid>",
		},
		{
			name:        "invalid VRF table ID",
			iface:       "memif0",
			vrf:         -5,
			ipv4:        true,
			expectedKey: "vpp/interface/memif0/vrf/<invalid>/ip-version/v4",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := InterfaceVrfKey(test.iface, test.vrf, test.ipv4, test.ipv6)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s vrf=%d ipv4=%t ipv6=%t\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.vrf, test.ipv4, test.ipv6, test.expectedKey, key)
			}
		})
	}
}

func TestParseInterfaceVrfKey(t *testing.T) {
	tests := []struct {
		name                  string
		key                   string
		expectedIface         string
		expectedVrf           int
		expectedIpv4          bool
		expectedIpv6          bool
		expectedIsIfaceVrfKey bool
	}{
		{
			name:                  "default IPv4 VRF",
			key:                   "vpp/interface/memif0/vrf/0/ip-version/v4",
			expectedIface:         "memif0",
			expectedVrf:           0,
			expectedIpv4:          true,
			expectedIsIfaceVrfKey: true,
		},
		{
			name:                  "default IPv6 VRF",
			key:                   "vpp/interface/memif0/vrf/0/ip-version/v6",
			expectedIface:         "memif0",
			expectedVrf:           0,
			expectedIpv6:          true,
			expectedIsIfaceVrfKey: true,
		},
		{
			name:                  "IPv4 VRF 1",
			key:                   "vpp/interface/memif0/vrf/1/ip-version/v4",
			expectedIface:         "memif0",
			expectedVrf:           1,
			expectedIpv4:          true,
			expectedIsIfaceVrfKey: true,
		},
		{
			name:                  "IPv6 VRF 1",
			key:                   "vpp/interface/memif0/vrf/1/ip-version/v6",
			expectedIface:         "memif0",
			expectedVrf:           1,
			expectedIpv6:          true,
			expectedIsIfaceVrfKey: true,
		},
		{
			name:                  "invalid interface name",
			key:                   "vpp/interface/<invalid>/vrf/1/ip-version/v6",
			expectedIface:         "<invalid>",
			expectedVrf:           1,
			expectedIpv6:          true,
			expectedIsIfaceVrfKey: true,
		},
		{
			name:                  "invalid ip version",
			key:                   "vpp/interface/memif0/vrf/1/ip-version/<invalid>",
			expectedIface:         "memif0",
			expectedVrf:           1,
			expectedIsIfaceVrfKey: true,
		},
		{
			name:                  "missing table ID",
			key:                   "vpp/interface/memif0/vrf//ip-version/v6",
			expectedIface:         "memif0",
			expectedVrf:           -1,
			expectedIpv6:          true,
			expectedIsIfaceVrfKey: true,
		},
		{
			name:                  "invalid table ID",
			key:                   "vpp/interface/memif0/vrf/<invalid>/ip-version/v6",
			expectedIface:         "memif0",
			expectedVrf:           -1,
			expectedIpv6:          true,
			expectedIsIfaceVrfKey: true,
		},
		{
			name:                  "interface name with forward slashes",
			key:                   "vpp/interface/Gbe1/2/3/vrf/12/ip-version/v6",
			expectedIface:         "Gbe1/2/3",
			expectedVrf:           12,
			expectedIpv6:          true,
			expectedIsIfaceVrfKey: true,
		},
		{
			name:                  "not vrf table key",
			key:                   "vpp/config/v2/interface/GigabitEthernet0/8/0",
			expectedIsIfaceVrfKey: false,
		},
		{
			name:                  "not vrf table key (inherited VRF)",
			key:                   "vpp/interface/Gbe1/2/3/vrf/from-interface/memif0",
			expectedIsIfaceVrfKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, vrf, ipv4, ipv6, isIfaceVrfKey := ParseInterfaceVrfKey(test.key)
			if isIfaceVrfKey != test.expectedIsIfaceVrfKey {
				t.Errorf("expected isVrfTableKey: %v\tgot: %v", test.expectedIsIfaceVrfKey, isIfaceVrfKey)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
			if vrf != test.expectedVrf {
				t.Errorf("expected vrf: %d\tgot: %d", test.expectedVrf, vrf)
			}
			if ipv4 != test.expectedIpv4 {
				t.Errorf("expected ipv4: %t\tgot: %t", test.expectedIpv4, ipv4)
			}
			if ipv6 != test.expectedIpv6 {
				t.Errorf("expected ipv6: %t\tgot: %t", test.expectedIpv6, ipv6)
			}
		})
	}
}

func TestInterfaceInheritedVrfKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		fromIface   string
		expectedKey string
	}{
		{
			name:        "memifs",
			iface:       "memif0",
			fromIface:   "memif1",
			expectedKey: "vpp/interface/memif0/vrf/from-interface/memif1",
		},
		{
			name:        "memif and Gbe",
			iface:       "memif0",
			fromIface:   "Gbe0/1/2",
			expectedKey: "vpp/interface/memif0/vrf/from-interface/Gbe0/1/2",
		},
		{
			name:        "Gbe-s",
			iface:       "Gbe3/4/5",
			fromIface:   "Gbe0/1/2",
			expectedKey: "vpp/interface/Gbe3/4/5/vrf/from-interface/Gbe0/1/2",
		},
		{
			name:        "missing interface",
			iface:       "",
			fromIface:   "memif1",
			expectedKey: "vpp/interface/<invalid>/vrf/from-interface/memif1",
		},
		{
			name:        "missing from-interface",
			iface:       "memif0",
			fromIface:   "",
			expectedKey: "vpp/interface/memif0/vrf/from-interface/<invalid>",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := InterfaceInheritedVrfKey(test.iface, test.fromIface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s fromIface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.fromIface, test.expectedKey, key)
			}
		})
	}
}

func TestParseInterfaceInheritedVrfKey(t *testing.T) {
	tests := []struct {
		name                       string
		key                        string
		expectedIface              string
		expectedFromIface          string
		expectedIsIfaceInherVrfKey bool
	}{
		{
			name:                       "memifs",
			key:                        "vpp/interface/memif0/vrf/from-interface/memif1",
			expectedIface:              "memif0",
			expectedFromIface:          "memif1",
			expectedIsIfaceInherVrfKey: true,
		},
		{
			name:                       "Gbe-s",
			key:                        "vpp/interface/Gbe1/2/3/vrf/from-interface/Gbe4/5/6",
			expectedIface:              "Gbe1/2/3",
			expectedFromIface:          "Gbe4/5/6",
			expectedIsIfaceInherVrfKey: true,
		},
		{
			name:                       "invalid interface",
			key:                        "vpp/interface/<invalid>/vrf/from-interface/Gbe4/5/6",
			expectedIface:              "<invalid>",
			expectedFromIface:          "Gbe4/5/6",
			expectedIsIfaceInherVrfKey: true,
		},
		{
			name:                       "invalid from-interface",
			key:                        "vpp/interface/Gbe1/2/3/vrf/from-interface/<invalid>",
			expectedIface:              "Gbe1/2/3",
			expectedFromIface:          "<invalid>",
			expectedIsIfaceInherVrfKey: true,
		},
		{
			name:                       "missing interface",
			key:                        "vpp/interface//vrf/from-interface/Gbe4/5/6",
			expectedIface:              "<invalid>",
			expectedFromIface:          "Gbe4/5/6",
			expectedIsIfaceInherVrfKey: true,
		},
		{
			name:                       "missing from-interface",
			key:                        "vpp/interface/Gbe1/2/3/vrf/from-interface/",
			expectedIface:              "Gbe1/2/3",
			expectedFromIface:          "<invalid>",
			expectedIsIfaceInherVrfKey: true,
		},
		{
			name:                       "not interface inherited-vrf key",
			key:                        "vpp/interface/memif0/vrf/1/ip-version/v6",
			expectedIsIfaceInherVrfKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, fromIface, isIfaceInherVrfKey := ParseInterfaceInheritedVrfKey(test.key)
			if isIfaceInherVrfKey != test.expectedIsIfaceInherVrfKey {
				t.Errorf("expected isIfaceInherVrfKey: %v\tgot: %v", test.expectedIsIfaceInherVrfKey, isIfaceInherVrfKey)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
			if fromIface != test.expectedFromIface {
				t.Errorf("expected fromIface: %s\tgot: %s", test.expectedFromIface, fromIface)
			}
		})
	}
}

func TestUnnumberedKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		expectedKey string
	}{
		{
			name:        "valid interface name",
			iface:       "memif0",
			expectedKey: "vpp/interface/unnumbered/memif0",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			expectedKey: "vpp/interface/unnumbered/<invalid>",
		},
		{
			name:        "Gbe interface",
			iface:       "GigabitEthernet0/8/0",
			expectedKey: "vpp/interface/unnumbered/GigabitEthernet0/8/0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := UnnumberedKey(test.iface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.expectedKey, key)
			}
		})
	}
}

func TestDHCPClientKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		expectedKey string
	}{
		{
			name:        "valid interface name",
			iface:       "memif0",
			expectedKey: "vpp/interface/dhcp-client/memif0",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			expectedKey: "vpp/interface/dhcp-client/<invalid>",
		},
		{
			name:        "Gbe interface",
			iface:       "GigabitEthernet0/8/0",
			expectedKey: "vpp/interface/dhcp-client/GigabitEthernet0/8/0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := DHCPClientKey(test.iface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.expectedKey, key)
			}
		})
	}
}

func TestParseNameFromDHCPClientKey(t *testing.T) {
	tests := []struct {
		name                    string
		key                     string
		expectedIface           string
		expectedIsDHCPClientKey bool
	}{
		{
			name:                    "valid interface name",
			key:                     "vpp/interface/dhcp-client/memif0",
			expectedIface:           "memif0",
			expectedIsDHCPClientKey: true,
		},
		{
			name:                    "invalid interface name",
			key:                     "vpp/interface/dhcp-client/<invalid>",
			expectedIface:           "<invalid>",
			expectedIsDHCPClientKey: true,
		},
		{
			name:                    "Gbe interface",
			key:                     "vpp/interface/dhcp-client/GigabitEthernet0/8/0",
			expectedIface:           "GigabitEthernet0/8/0",
			expectedIsDHCPClientKey: true,
		},
		{
			name:                    "not DHCP client key",
			key:                     "vpp/config/v2/bd/bd1",
			expectedIface:           "",
			expectedIsDHCPClientKey: false,
		},
		{
			name:                    "not DHCP client key (empty interface)",
			key:                     "vpp/interface/dhcp-client/",
			expectedIface:           "",
			expectedIsDHCPClientKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, isDHCPClientKey := ParseNameFromDHCPClientKey(test.key)
			if isDHCPClientKey != test.expectedIsDHCPClientKey {
				t.Errorf("expected isInterfaceKey: %v\tgot: %v", test.expectedIsDHCPClientKey, isDHCPClientKey)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
		})
	}
}

func TestDHCPLeaseKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		expectedKey string
	}{
		{
			name:        "valid interface name",
			iface:       "memif0",
			expectedKey: "vpp/interface/dhcp-lease/memif0",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			expectedKey: "vpp/interface/dhcp-lease/<invalid>",
		},
		{
			name:        "Gbe interface",
			iface:       "GigabitEthernet0/8/0",
			expectedKey: "vpp/interface/dhcp-lease/GigabitEthernet0/8/0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := DHCPLeaseKey(test.iface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.expectedKey, key)
			}
		})
	}
}

func TestLinkStateKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		linkIsUp    bool
		expectedKey string
	}{
		{
			name:        "link is UP",
			iface:       "memif0",
			linkIsUp:    true,
			expectedKey: "vpp/interface/memif0/link-state/UP",
		},
		{
			name:        "link is DOWN",
			iface:       "memif0",
			linkIsUp:    false,
			expectedKey: "vpp/interface/memif0/link-state/DOWN",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			linkIsUp:    true,
			expectedKey: "vpp/interface/<invalid>/link-state/UP",
		},
		{
			name:        "Gbe interface",
			iface:       "GigabitEthernet0/8/0",
			linkIsUp:    false,
			expectedKey: "vpp/interface/GigabitEthernet0/8/0/link-state/DOWN",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := LinkStateKey(test.iface, test.linkIsUp)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s linkIsUp=%v\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.linkIsUp, test.expectedKey, key)
			}
		})
	}
}

func TestParseLinkStateKey(t *testing.T) {
	tests := []struct {
		name                   string
		key                    string
		expectedIface          string
		expectedIsLinkUp       bool
		expectedIsLinkStateKey bool
	}{
		{
			name:                   "link is UP",
			key:                    "vpp/interface/memif0/link-state/UP",
			expectedIface:          "memif0",
			expectedIsLinkUp:       true,
			expectedIsLinkStateKey: true,
		},
		{
			name:                   "link is DOWN",
			key:                    "vpp/interface/memif0/link-state/DOWN",
			expectedIface:          "memif0",
			expectedIsLinkUp:       false,
			expectedIsLinkStateKey: true,
		},
		{
			name:                   "invalid interface name",
			key:                    "vpp/interface/<invalid>/link-state/DOWN",
			expectedIsLinkStateKey: false,
		},
		{
			name:                   "Gbe interface",
			key:                    "vpp/interface/GigabitEthernet0/8/0/link-state/UP",
			expectedIface:          "GigabitEthernet0/8/0",
			expectedIsLinkUp:       true,
			expectedIsLinkStateKey: true,
		},
		{
			name:                   "invalid link state",
			key:                    "vpp/interface/GigabitEthernet0/8/0/link-state/<invalid>",
			expectedIsLinkStateKey: false,
		},
		{
			name:                   "missing link state",
			key:                    "vpp/interface/GigabitEthernet0/8/0/link-state",
			expectedIsLinkStateKey: false,
		},
		{
			name:                   "not link state key",
			key:                    "vpp/interface/unnumbered/GigabitEthernet0/8/0",
			expectedIsLinkStateKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, isLinkUp, isLinkStateKey := ParseLinkStateKey(test.key)
			if isLinkStateKey != test.expectedIsLinkStateKey {
				t.Errorf("expected isLinkStateKey: %v\tgot: %v",
					test.expectedIsLinkStateKey, isLinkStateKey)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
			if isLinkUp != test.expectedIsLinkUp {
				t.Errorf("expected isLinkUp: %t\tgot: %t", test.expectedIsLinkUp, isLinkUp)
			}
		})
	}
}

func TestRxPlacementKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		queue       uint32
		expectedKey string
	}{
		{
			name:        "queue 0",
			iface:       "memif0",
			queue:       0,
			expectedKey: "vpp/interface/memif0/rx-placement/queue/0",
		},
		{
			name:        "queue 1",
			iface:       "memif0",
			queue:       1,
			expectedKey: "vpp/interface/memif0/rx-placement/queue/1",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			queue:       1,
			expectedKey: "vpp/interface/<invalid>/rx-placement/queue/1",
		},
		{
			name:        "Gbe interface",
			iface:       "GigabitEthernet0/8/0",
			queue:       2,
			expectedKey: "vpp/interface/GigabitEthernet0/8/0/rx-placement/queue/2",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := RxPlacementKey(test.iface, test.queue)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s queue=%d\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.queue, test.expectedKey, key)
			}
		})
	}
}

func TestParseRxPlacementKey(t *testing.T) {
	tests := []struct {
		name                     string
		key                      string
		expectedIface            string
		expectedQueue            uint32
		expectedIsRxPlacementKey bool
	}{
		{
			name:                     "queue 0",
			key:                      "vpp/interface/memif0/rx-placement/queue/0",
			expectedIface:            "memif0",
			expectedQueue:            0,
			expectedIsRxPlacementKey: true,
		},
		{
			name:                     "queue 1",
			key:                      "vpp/interface/memif0/rx-placement/queue/1",
			expectedIface:            "memif0",
			expectedQueue:            1,
			expectedIsRxPlacementKey: true,
		},
		{
			name:                     "invalid interface name",
			key:                      "vpp/interface/<invalid>/rx-placement/queue/1",
			expectedIsRxPlacementKey: false,
		},
		{
			name:                     "invalid queue",
			key:                      "vpp/interface/memif0/rx-placement/queue/<invalid>",
			expectedIsRxPlacementKey: false,
		},
		{
			name:                     "missing queue",
			key:                      "vpp/interface/memif0/rx-placement/",
			expectedIsRxPlacementKey: false,
		},
		{
			name:                     "missing queue ID",
			key:                      "vpp/interface/memif0/rx-placement/queue",
			expectedIsRxPlacementKey: false,
		},
		{
			name:                     "Gbe interface",
			key:                      "vpp/interface/GigabitEthernet0/8/0/rx-placement/queue/3",
			expectedIface:            "GigabitEthernet0/8/0",
			expectedQueue:            3,
			expectedIsRxPlacementKey: true,
		},
		{
			name:                     "not rx placement key",
			key:                      "vpp/interface/memif0/link-state/DOWN",
			expectedIsRxPlacementKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, queue, isRxPlacementKey := ParseRxPlacementKey(test.key)
			if isRxPlacementKey != test.expectedIsRxPlacementKey {
				t.Errorf("expected isRxPlacementKey: %v\tgot: %v",
					test.expectedIsRxPlacementKey, isRxPlacementKey)
			}
			if queue != test.expectedQueue {
				t.Errorf("expected queue: %d\tgot: %d", test.expectedQueue, queue)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
		})
	}
}

func TestRxModesKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		expectedKey string
	}{
		{
			name:        "memif",
			iface:       "memif0",
			expectedKey: "vpp/interface/memif0/rx-modes",
		},
		{
			name:        "Gbe",
			iface:       "GigabitEthernet0/8/0",
			expectedKey: "vpp/interface/GigabitEthernet0/8/0/rx-modes",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			expectedKey: "vpp/interface/<invalid>/rx-modes",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := RxModesKey(test.iface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.expectedKey, key)
			}
		})
	}
}

func TestParseRxModesKey(t *testing.T) {
	tests := []struct {
		name                 string
		key                  string
		expectedIface        string
		expectedIsRxModesKey bool
	}{
		{
			name:                 "memif",
			key:                  "vpp/interface/memif0/rx-modes",
			expectedIface:        "memif0",
			expectedIsRxModesKey: true,
		},
		{
			name:                 "Gbe",
			key:                  "vpp/interface/GigabitEthernet0/8/0/rx-modes",
			expectedIface:        "GigabitEthernet0/8/0",
			expectedIsRxModesKey: true,
		},
		{
			name:                 "invalid interface name",
			key:                  "vpp/interface/<invalid>/rx-modes",
			expectedIsRxModesKey: false,
		},
		{
			name:                 "missing rx-mode suffix",
			key:                  "vpp/interface/<invalid>",
			expectedIsRxModesKey: false,
		},
		{
			name:                 "not rx-mode key",
			key:                  "vpp/interface/memif0/address/192.168.1.12/24",
			expectedIsRxModesKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, isRxModesKey := ParseRxModesKey(test.key)
			if isRxModesKey != test.expectedIsRxModesKey {
				t.Errorf("expected isRxModesKey: %v\tgot: %v",
					test.expectedIsRxModesKey, isRxModesKey)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
		})
	}
}

func TestInterfaceWithIPKey(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		expectedKey string
	}{
		{
			name:        "memif",
			iface:       "memif0",
			expectedKey: "vpp/interface/memif0/has-IP-address",
		},
		{
			name:        "invalid interface name",
			iface:       "",
			expectedKey: "vpp/interface/<invalid>/has-IP-address",
		},
		{
			name:        "Gbe interface",
			iface:       "GigabitEthernet0/8/0",
			expectedKey: "vpp/interface/GigabitEthernet0/8/0/has-IP-address",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := InterfaceWithIPKey(test.iface)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.expectedKey, key)
			}
		})
	}
}

func TestParseInterfaceWithIPKey(t *testing.T) {
	tests := []struct {
		name                     string
		key                      string
		expectedIface            string
		expectedIsIfaceWithIPKey bool
	}{
		{
			name:                     "memif",
			key:                      "vpp/interface/memif0/has-IP-address",
			expectedIface:            "memif0",
			expectedIsIfaceWithIPKey: true,
		},
		{
			name:                     "Gbe",
			key:                      "vpp/interface/GigabitEthernet0/8/0/has-IP-address",
			expectedIface:            "GigabitEthernet0/8/0",
			expectedIsIfaceWithIPKey: true,
		},
		{
			name:                     "invalid interface name",
			key:                      "vpp/interface/<invalid>/has-IP-address",
			expectedIsIfaceWithIPKey: false,
		},
		{
			name:                     "missing has-IP-address suffix",
			key:                      "vpp/interface/<invalid>",
			expectedIsIfaceWithIPKey: false,
		},
		{
			name:                     "not has-IP-address key",
			key:                      "vpp/interface/memif0/address/192.168.1.12/24",
			expectedIsIfaceWithIPKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, isInterfaceWithIPKey := ParseInterfaceWithIPKey(test.key)
			if isInterfaceWithIPKey != test.expectedIsIfaceWithIPKey {
				t.Errorf("expected isInterfaceWithIPKey: %v\tgot: %v",
					test.expectedIsIfaceWithIPKey, isInterfaceWithIPKey)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
		})
	}
}
