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

package models_test

import (
	"testing"

	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/l3"
	"github.com/ligato/vpp-agent/api/models/vpp/nat"
)

func TestKeys(t *testing.T) {
	tests := []struct {
		name        string
		model       models.ProtoModel
		expectedKey string
	}{
		{
			name: "linux iface",
			model: &linux_interfaces.Interface{
				Name: "testName",
				Type: linux_interfaces.Interface_VETH,
			},
			expectedKey: "linux/config/v2/interface/testName",
		},
		{
			name: "linux route",
			model: &linux_l3.StaticRoute{
				DstNetwork:        "1.1.1.1/24",
				OutgoingInterface: "eth0",
				GwAddr:            "9.9.9.9",
			},
			expectedKey: "linux/config/v2/route/1.1.1.0/24/eth0",
		},
		{
			name: "linux arp",
			model: &linux_l3.StaticARPEntry{
				Interface: "if1",
				IpAddress: "1.2.3.4",
				HwAddress: "11:22:33:44:55:66",
			},
			expectedKey: "linux/config/v2/arp/if1/1.2.3.4",
		},
		{
			name: "vpp dnat",
			model: &vpp_nat.DNat44{
				Label: "mynat1",
			},
			expectedKey: "vpp/config/v2/nat44/dnat/mynat1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := models.Key(test.model)
			t.Logf("key: %q", key)

			if key != test.expectedKey {
				t.Fatalf("expected key: %q, got: %q", test.expectedKey, key)
			}
		})
	}
}
