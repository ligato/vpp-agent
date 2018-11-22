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
	_ "github.com/ligato/vpp-agent/api/models/linux"
	"github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/l3"
)

func TestKey(t *testing.T) {
	testIf := &linux_interfaces.Interface{
		Name: "testName",
		Type: linux_interfaces.Interface_VETH,
	}

	key := models.Key(testIf)
	t.Logf("key=%q", key)
	if key != "linux/config/v2/interface/testName" {
		t.Fatalf("key is: %q", key)
	}
}

func TestKey2(t *testing.T) {
	testIf := &linux_l3.StaticRoute{
		DstNetwork:        "1.1.1.1/24",
		OutgoingInterface: "eth0",
		GwAddr:            "9.9.9.9",
	}

	key := models.Key(testIf)
	t.Logf("key=%q", key)
	if key != "linux/config/v2/route/1.1.1.0/24/eth0" {
		t.Fatalf("key is: %q", key)
	}
}
