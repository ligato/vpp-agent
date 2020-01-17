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

package vpp_abf_test

import (
	"testing"

	vpp_abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
)

func TestABFKey(t *testing.T) {
	tests := []struct {
		name        string
		abfIndex    uint32
		expectedKey string
	}{
		{
			name:        "valid ABF index",
			abfIndex:    0,
			expectedKey: "config/vpp/abfs/v2/abf/0",
		},
		{
			name:        "valid ABF index",
			abfIndex:    1,
			expectedKey: "config/vpp/abfs/v2/abf/1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := vpp_abf.Key(test.abfIndex)
			if key != test.expectedKey {
				t.Errorf("failed for: abfIndex=%d\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.abfIndex, test.expectedKey, key)
			}
		})
	}
}

func TestParseNameFromKey(t *testing.T) {
	tests := []struct {
		name             string
		key              string
		expectedABFIndex string
		expectedIsABFKey bool
	}{
		{
			name:             "valid ABF index",
			key:              "config/vpp/abfs/v2/abf/1",
			expectedABFIndex: "1",
			expectedIsABFKey: true,
		},
		{
			name:             "invalid ABF index",
			key:              "config/vpp/abfs/v2/abf/<invalid>",
			expectedABFIndex: "<invalid>",
			expectedIsABFKey: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			abfIndex, isABFKey := vpp_abf.ModelABF.ParseKey(test.key)
			if isABFKey != test.expectedIsABFKey {
				t.Errorf("expected isABFKey: %v\tgot: %v", test.expectedIsABFKey, isABFKey)
			}
			if abfIndex != test.expectedABFIndex {
				t.Errorf("expected abfIndex: %s\tgot: %s", test.expectedABFIndex, abfIndex)
			}
		})
	}
}

func TestABFToInterfaceKey(t *testing.T) {
	tests := []struct {
		name        string
		abfIndex    uint32
		iface       string
		expectedKey string
	}{
		{
			name:        "interface",
			abfIndex:    1,
			iface:       "tap0",
			expectedKey: "vpp/abf/1/interface/tap0",
		},
		{
			name:        "empty interface",
			abfIndex:    2,
			iface:       "",
			expectedKey: "vpp/abf/2/interface/<invalid>",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := vpp_abf.ToInterfaceKey(test.abfIndex, test.iface)
			if key != test.expectedKey {
				t.Errorf("failed for: abfIndex=%d iface=%s\n"+
					"expected key:\n\t%q\ngot key:\n\t%q",
					test.abfIndex, test.iface, test.expectedKey, key)
			}
		})
	}
}

func TestParseACLToInterfaceKey(t *testing.T) {
	tests := []struct {
		name                  string
		key                   string
		expectedABFIndex      string
		expectedIface         string
		expectedIsABFIfaceKey bool
	}{
		{
			name:                  "interface",
			key:                   "vpp/abf/1/interface/tap0",
			expectedABFIndex:      "1",
			expectedIface:         "tap0",
			expectedIsABFIfaceKey: true,
		},
		{
			name:                  "invalid abf index",
			key:                   "vpp/abf/<invalid>/interface/tap0",
			expectedABFIndex:      "<invalid>",
			expectedIface:         "tap0",
			expectedIsABFIfaceKey: true,
		},
		{
			name:                  "invalid interface",
			key:                   "vpp/abf/1/interface/<invalid>",
			expectedABFIndex:      "1",
			expectedIface:         "<invalid>",
			expectedIsABFIfaceKey: true,
		},
		{
			name:                  "all parameters invalid",
			key:                   "vpp/abf/<invalid>/interface/<invalid>",
			expectedABFIndex:      "<invalid>",
			expectedIface:         "<invalid>",
			expectedIsABFIfaceKey: true,
		},
		{
			name:                  "not ABFToInterface key",
			key:                   "vpp/acl/acl1/interface/ingress/tap0",
			expectedABFIndex:      "",
			expectedIface:         "",
			expectedIsABFIfaceKey: false,
		},
		{
			name:                  "not ABFToInterface key (cut after interface)",
			key:                   "vpp/abf/<invalid>/interface/",
			expectedABFIndex:      "",
			expectedIface:         "",
			expectedIsABFIfaceKey: false,
		},
		{
			name:                  "empty key",
			key:                   "",
			expectedABFIndex:      "",
			expectedIface:         "",
			expectedIsABFIfaceKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			abfIndex, iface, isABFIfaceKey := vpp_abf.ParseToInterfaceKey(test.key)
			if isABFIfaceKey != test.expectedIsABFIfaceKey {
				t.Errorf("expected isABFKey: %v\tgot: %v", test.expectedIsABFIfaceKey, isABFIfaceKey)
			}
			if abfIndex != test.expectedABFIndex {
				t.Errorf("expected abfIndex: %s\tgot: %s", test.expectedABFIndex, abfIndex)
			}
			if iface != test.expectedIface {
				t.Errorf("expected iface: %s\tgot: %s", test.expectedIface, iface)
			}
		})
	}
}
