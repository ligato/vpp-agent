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
	"net"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"

	vpp_interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	_ "github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	ifplugin_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

func TestInterfaceIP(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

	tests := []struct {
		name  string
		ipnet net.IPNet
	}{
		{"basic ipv4", net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.IPMask{255, 255, 255, 0}}},
		{"basic ipv6", net.IPNet{IP: net.ParseIP("::1"), Mask: net.IPMask{255, 255, 255, 0}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ifIdx, err := h.AddLoopbackInterface("loop0")
			if err != nil {
				t.Fatalf("creating loopback interface failed: %v", err)
			}
			t.Logf("loop0 index: %+v", ifIdx)

			if err := h.AddInterfaceIP(ifIdx, &test.ipnet); err != nil {
				t.Fatalf("adding interface IP failed: %v", err)
			}
		})
	}
}

func TestInterfaceDumpState(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

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

	ifaces, err := h.DumpInterfaceStates()
	if err != nil {
		t.Fatalf("dumping interface states failed: %v", err)
	}
	if len(ifaces) != 3 {
		t.Errorf("expected 3 interface states in dump, got: %d", len(ifaces))
	}

	ifaces, err = h.DumpInterfaceStates(ifIdx)
	if err != nil {
		t.Fatalf("dumping interface states failed: %v", err)
	}
	iface := ifaces[ifIdx]
	t.Logf("interface state: %+v", iface)

	if iface == nil {
		t.Fatalf("expected interface, got: nil")
	}
	if iface.InternalName != "loop1" {
		t.Errorf("expected interface internal name to be loop1, got: %v", iface.InternalName)
	}
	if len(iface.PhysAddress) == 0 {
		t.Errorf("expected interface phys address to not be empty, got: %q", iface.PhysAddress)
	}
}

func TestLoopbackInterface(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

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
		t.Errorf("expected interface name to be loop1, got: %v", iface.Interface.Name)
	}
	if iface.Interface.PhysAddress == "" {
		t.Errorf("expected interface phys address to not be empty, got: %v", iface.Interface.PhysAddress)
	}
	if iface.Interface.Enabled == true {
		t.Errorf("expected interface to not be enabled")
	}
	if iface.Interface.Type != vpp_interfaces.Interface_SOFTWARE_LOOPBACK {
		t.Errorf("expected interface type to be SOFTWARE_LOOPBACK, got: %v", iface.Interface.Type)
	}
	if iface.Interface.Link != nil {
		t.Errorf("expected interface link to be nil, got: %T", iface.Interface.Link)
	}
}

func TestMemifInterface(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

	ifIdx, err := h.AddMemifInterface("memif1", &vpp_interfaces.MemifLink{
		Id:     1,
		Mode:   vpp_interfaces.MemifLink_ETHERNET,
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
		t.Errorf("expected interface name to be memif1, got: %v", iface.Interface.Name)
	}
	if iface.Interface.Type != vpp_interfaces.Interface_MEMIF {
		t.Errorf("expected interface type to be memif, got: %v", iface.Interface.Type)
	}
	link, ok := iface.Interface.Link.(*vpp_interfaces.Interface_Memif)
	if !ok {
		t.Fatalf("expected interface link to be memif, got: %T", iface.Interface.Link)
	}
	if link.Memif.Id != 1 {
		t.Errorf("expected memif ID to be 1, got: %v", link.Memif.Id)
	}
	if link.Memif.Mode != vpp_interfaces.MemifLink_ETHERNET {
		t.Errorf("expected memif mode to be ETHERNET, got: %v", link.Memif.Mode)
	}
}
