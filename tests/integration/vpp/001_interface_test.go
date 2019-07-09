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

package vpp

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"

	vpp_interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	_ "github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

func TestLoopbackInterface(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := vppcalls.CompatibleInterfaceVppHandler(ctx.Chan, logrus.NewLogger("test"))

	index, err := h.AddLoopbackInterface("loop1")
	if err != nil {
		t.Fatalf("creating loopback interface failed: %v", err)
	}
	t.Logf("loopback index: %+v", index)

	ifaces, err := h.DumpInterfaces()
	if err != nil {
		t.Fatalf("dumping interfaces failed: %v", err)
	}
	iface, ok := ifaces[index]
	if !ok {
		t.Fatalf("loopback interface not found in dump")
	}
	t.Logf("interface: %+v", iface)
	if iface.Interface.Name != "loop1" {
		t.Fatalf("expected interface name to be loop1, got %v", iface.Interface.Name)
	}
	if iface.Interface.Type != vpp_interfaces.Interface_SOFTWARE_LOOPBACK {
		t.Fatalf("expected interface type to be loopback, got %v", iface.Interface.Type)
	}
}
