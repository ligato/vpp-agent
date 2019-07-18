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
	ifplugin_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

func TestInterfaceDump(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.Chan, logrus.NewLogger("test"))

	ifIdx0, err := h.AddLoopbackInterface("loop0")
	if err != nil {
		t.Fatalf("creating loopback interface failed: %v", err)
	}
	t.Logf("loop0 index: %+v", ifIdx0)

	ifIdx, err := h.AddLoopbackInterface("loop1")
	if err != nil {
		t.Fatalf("creating loopback interface failed: %v", err)
	}
	t.Logf("loop1 index: %+v", ifIdx)

	ifaces, err := h.DumpInterfaces()
	if err != nil {
		t.Fatalf("dumping interfaces failed: %v", err)
	}
	if len(ifaces) != 3 {
		t.Fatalf("expected 3 interfaces in dump, got %d", len(ifaces))
	}

	iface, err := h.DumpInterface(ifIdx)
	if err != nil {
		t.Fatalf("dumping interface failed: %v", err)
	}
	t.Logf("interface: %+v", iface.Interface)
	if iface.Interface == nil {
		t.Fatalf("expected interface, got nil")
	}
	if iface.Interface.Name != "loop1" {
		t.Errorf("expected interface name to be loop1, got %v", iface.Interface.Name)
	}
	if iface.Interface.Type != vpp_interfaces.Interface_SOFTWARE_LOOPBACK {
		t.Errorf("expected interface type to be loopback, got %v", iface.Interface.Type)
	}
}

func TestLoopbackInterface(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.Chan, logrus.NewLogger("test"))

	ifIdx, err := h.AddLoopbackInterface("loop1")
	if err != nil {
		t.Fatalf("creating loopback interface failed: %v", err)
	}
	t.Logf("loopback index: %+v", ifIdx)

	ifaces, err := h.DumpInterfaces()
	if err != nil {
		t.Fatalf("dumping interfaces failed: %v", err)
	}
	iface, ok := ifaces[ifIdx]
	if !ok {
		t.Fatalf("loopback interface not found in dump")
	}
	t.Logf("interface: %+v", iface.Interface)
	if iface.Interface.Name != "loop1" {
		t.Errorf("expected interface name to be loop1, got %v", iface.Interface.Name)
	}
	if iface.Interface.Type != vpp_interfaces.Interface_SOFTWARE_LOOPBACK {
		t.Errorf("expected interface type to be loopback, got %v", iface.Interface.Type)
	}
}

func TestMemifInterface(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.Chan, logrus.NewLogger("test"))

	ifIdx, err := h.AddMemifInterface("memif1", &vpp_interfaces.MemifLink{
		Id:     1,
		Mode:   vpp_interfaces.MemifLink_IP,
		Secret: "secret",
		Master: true,
	}, 0)
	if err != nil {
		t.Fatalf("creating memif interface failed: %v", err)
	}
	t.Logf("memif index: %+v", ifIdx)

	ifaces, err := h.DumpInterfaces()
	if err != nil {
		t.Fatalf("dumping interfaces failed: %v", err)
	}
	iface, ok := ifaces[ifIdx]
	if !ok {
		t.Fatalf("Memif interface not found in dump")
	}
	t.Logf("interface: %+v", iface.Interface)
	if iface.Interface.Name != "memif1" {
		t.Errorf("expected interface name to be memif1, got %v", iface.Interface.Name)
	}
	if iface.Interface.Type != vpp_interfaces.Interface_MEMIF {
		t.Errorf("expected interface type to be memif, got %v", iface.Interface.Type)
	}
}
