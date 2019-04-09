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

package vpp_nat

import (
	"testing"
)

/*func TestDNAT44Key(t *testing.T) {
	tests := []struct {
		name        string
		label       string
		expectedKey string
	}{
		{
			name:        "valid DNAT44 label",
			label:       "dnat1",
			expectedKey: "vpp/config/v2/nat44/dnat/dnat1",
		},
		{
			name:        "invalid DNAT44 label",
			label:       "",
			expectedKey: "vpp/config/v2/nat44/dnat/<invalid>",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := DNAT44Key(test.label)
			if key != test.expectedKey {
				t.Errorf("failed for: label=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.label, test.expectedKey, key)
			}
		})
	}
}*/

func TestInterfaceNAT44Key(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		isInside    bool
		expectedKey string
	}{
		{
			name:        "interface-with-IN-feature",
			iface:       "tap0",
			isInside:    true,
			expectedKey: "vpp/nat44/interface/tap0/feature/in",
		},
		{
			name:        "interface-with-OUT-feature",
			iface:       "tap1",
			isInside:    false,
			expectedKey: "vpp/nat44/interface/tap1/feature/out",
		},
		{
			name:        "gbe-interface-OUT",
			iface:       "GigabitEthernet0/8/0",
			isInside:    false,
			expectedKey: "vpp/nat44/interface/GigabitEthernet0/8/0/feature/out",
		},
		{
			name:        "gbe-interface-IN",
			iface:       "GigabitEthernet0/8/0",
			isInside:    true,
			expectedKey: "vpp/nat44/interface/GigabitEthernet0/8/0/feature/in",
		},
		{
			name:        "invalid-interface-with-IN-feature",
			iface:       "",
			isInside:    true,
			expectedKey: "vpp/nat44/interface/<invalid>/feature/in",
		},
		{
			name:        "invalid-interface-with-OUT-feature",
			iface:       "",
			isInside:    false,
			expectedKey: "vpp/nat44/interface/<invalid>/feature/out",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := InterfaceNAT44Key(test.iface, test.isInside)
			if key != test.expectedKey {
				t.Errorf("failed for: iface=%s isInside=%t\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.iface, test.isInside, test.expectedKey, key)
			}
		})
	}
}

func TestParseInterfaceNAT44Key(t *testing.T) {
	tests := []struct {
		name                        string
		key                         string
		expectedIface               string
		expectedIsInside            bool
		expectedIsInterfaceNAT44Key bool
	}{
		{
			name:                        "interface-with-IN-feature",
			key:                         "vpp/nat44/interface/tap0/feature/in",
			expectedIface:               "tap0",
			expectedIsInside:            true,
			expectedIsInterfaceNAT44Key: true,
		},
		{
			name:                        "interface-with-OUT-feature",
			key:                         "vpp/nat44/interface/tap1/feature/out",
			expectedIface:               "tap1",
			expectedIsInside:            false,
			expectedIsInterfaceNAT44Key: true,
		},
		{
			name:                        "gbe-interface-OUT",
			key:                         "vpp/nat44/interface/GigabitEthernet0/8/0/feature/out",
			expectedIface:               "GigabitEthernet0/8/0",
			expectedIsInside:            false,
			expectedIsInterfaceNAT44Key: true,
		},
		{
			name:                        "gbe-interface-IN",
			key:                         "vpp/nat44/interface/GigabitEthernet0/8/0/feature/in",
			expectedIface:               "GigabitEthernet0/8/0",
			expectedIsInside:            true,
			expectedIsInterfaceNAT44Key: true,
		},
		{
			name:                        "invalid-interface",
			key:                         "vpp/nat44/interface/<invalid>/feature/in",
			expectedIface:               "<invalid>",
			expectedIsInside:            true,
			expectedIsInterfaceNAT44Key: true,
		},
		{
			name:                        "not interface key 1",
			key:                         "vpp/nat44/address/192.168.1.1/twice-nat/on",
			expectedIface:               "",
			expectedIsInside:            false,
			expectedIsInterfaceNAT44Key: false,
		},
		{
			name:                        "not interface key 2",
			key:                         "vpp/config/v2/nat44/dnat/dnat1",
			expectedIface:               "",
			expectedIsInside:            false,
			expectedIsInterfaceNAT44Key: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iface, isInside, isInterfaceNAT44Key := ParseInterfaceNAT44Key(test.key)
			if isInterfaceNAT44Key != test.expectedIsInterfaceNAT44Key {
				t.Errorf("expected isInterfaceNAT44Key: %v\tgot: %v", test.expectedIsInterfaceNAT44Key, isInterfaceNAT44Key)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
			if isInside != test.expectedIsInside {
				t.Errorf("expected isInside: %t\tgot: %t", test.expectedIsInside, isInside)
			}
		})
	}
}

func TestAddressNAT44Key(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		twiceNat    bool
		expectedKey string
	}{
		{
			name:        "twice NAT is disabled",
			address:     "192.168.1.1",
			twiceNat:    false,
			expectedKey: "vpp/nat44/address/192.168.1.1/twice-nat/off",
		},
		{
			name:        "twice NAT is enabled",
			address:     "192.168.1.1",
			twiceNat:    true,
			expectedKey: "vpp/nat44/address/192.168.1.1/twice-nat/on",
		},
		{
			name:        "invalid address",
			address:     "invalid",
			twiceNat:    true,
			expectedKey: "vpp/nat44/address/invalid/twice-nat/on",
		},
		{
			name:        "empty address",
			address:     "",
			twiceNat:    true,
			expectedKey: "vpp/nat44/address//twice-nat/on",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := AddressNAT44Key(test.address, test.twiceNat)
			if key != test.expectedKey {
				t.Errorf("failed for: address=%s twiceNat=%t\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.address, test.twiceNat, test.expectedKey, key)
			}
		})
	}
}

func TestParseAddressNAT44Key(t *testing.T) {
	tests := []struct {
		name                      string
		key                       string
		expectedAddress           string
		expectedTwiceNat          bool
		expectedIsAddressNAT44Key bool
	}{
		{
			name:                      "twice NAT is disabled",
			key:                       "vpp/nat44/address/192.168.1.1/twice-nat/off",
			expectedAddress:           "192.168.1.1",
			expectedTwiceNat:          false,
			expectedIsAddressNAT44Key: true,
		},
		{
			name:                      "twice NAT is enabled",
			key:                       "vpp/nat44/address/192.168.1.1/twice-nat/on",
			expectedAddress:           "192.168.1.1",
			expectedTwiceNat:          true,
			expectedIsAddressNAT44Key: true,
		},
		{
			name:                      "invalid address (not validated)",
			key:                       "vpp/nat44/address/invalid/twice-nat/on",
			expectedAddress:           "invalid",
			expectedTwiceNat:          true,
			expectedIsAddressNAT44Key: true,
		},
		{
			name:                      "empty address (not validated)",
			key:                       "vpp/nat44/address//twice-nat/on",
			expectedAddress:           "",
			expectedTwiceNat:          true,
			expectedIsAddressNAT44Key: true,
		},
		{
			name:                      "not address key",
			key:                       "vpp/nat44/interface/tap0/feature/in",
			expectedIsAddressNAT44Key: false,
		},
		{
			name:                      "not address key (missing twice-nat flag)",
			key:                       "vpp/nat44/address/192.168.1.1",
			expectedIsAddressNAT44Key: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			address, twiceNat, isAddressNAT44Key := ParseAddressNAT44Key(test.key)
			if isAddressNAT44Key != test.expectedIsAddressNAT44Key {
				t.Errorf("expected isAddressNAT44Key: %v\tgot: %v", test.expectedIsAddressNAT44Key, isAddressNAT44Key)
			}
			if address != test.expectedAddress {
				t.Errorf("expected address: %s\tgot: %s", test.expectedAddress, address)
			}
			if twiceNat != test.expectedTwiceNat {
				t.Errorf("expected twiceNat: %t\tgot: %t", test.expectedTwiceNat, twiceNat)
			}
		})
	}
}
