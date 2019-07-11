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
	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	_ "github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	_ "github.com/ligato/vpp-agent/plugins/vpp/l3plugin"
	l3plugin_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vrfidx"
)

func TestLoopbackInterface(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.Chan, logrus.NewLogger("test"))

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
	t.Logf("interface: %+v", iface.Interface)
	if iface.Interface.Name != "loop1" {
		t.Fatalf("expected interface name to be loop1, got %v", iface.Interface.Name)
	}
	if iface.Interface.Type != vpp_interfaces.Interface_SOFTWARE_LOOPBACK {
		t.Fatalf("expected interface type to be loopback, got %v", iface.Interface.Type)
	}
}

func TestRoutes(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv4", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV4})
	vrfIndexes.Put("vrf1-ipv6", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV6})

	h := l3plugin_vppcalls.CompatibleL3VppHandler(ctx.Chan, ifIndexes, vrfIndexes, logrus.NewLogger("test"))

	routes, err := h.DumpRoutes()
	if err != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	t.Logf("%d routes dumped", len(routes))

	var hasIPv4, hasIPv6 bool
	for _, route := range routes {
		t.Logf(" - route: %+v", route.Route)

		ip, _, err := net.ParseCIDR(route.Route.DstNetwork)
		if err != nil {
			t.Fatalf("invalid dst network: %v", route.Route.DstNetwork)
		}
		if ip.To4() == nil {
			hasIPv4 = true
		} else {
			hasIPv6 = true
		}
	}

	if !hasIPv4 || !hasIPv6 {
		t.Fatalf("expected dump to contain both IPv4 and IPv6 routes")
	}
}
