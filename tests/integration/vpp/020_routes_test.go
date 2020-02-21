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

	"go.ligato.io/cn-infra/v2/logging/logrus"

	netalloc_mock "go.ligato.io/vpp-agent/v3/plugins/netalloc/mock"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	l3plugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
)

func TestRoutes(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv4", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV4})
	vrfIndexes.Put("vrf1-ipv6", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV6})

	h := l3plugin_vppcalls.CompatibleL3VppHandler(ctx.vppClient, ifIndexes, vrfIndexes,
		netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test"))

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

func TestCRUDIPv4Route(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(test.vppClient, logrus.NewLogger("test"))
	const ifName = "loop1"
	ifIdx, err := ih.AddLoopbackInterface(ifName)
	if err != nil {
		t.Fatalf("creating interface failed: %v", err)
	}
	t.Logf("interface created %v", ifIdx)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-iface1"), "test-iface1")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{
		SwIfIndex: ifIdx,
	})

	var vrfMetaIdx uint32
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv4-vrf0", &vrfidx.VRFMetadata{Index: vrfMetaIdx, Protocol: vpp_l3.VrfTable_IPV4})

	h := l3plugin_vppcalls.CompatibleL3VppHandler(test.vppClient, ifIndexes, vrfIndexes,
		netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test"))

	routes, errx := h.DumpRoutes()
	if errx != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	routesCnt := len(routes)
	t.Logf("%d routes dumped", routesCnt)

	newRoute := vpp_l3.Route{VrfId: 0, DstNetwork: "192.168.10.0/24", NextHopAddr: "192.168.30.1", OutgoingInterface: ifName}
	err = h.VppAddRoute(test.Ctx, &newRoute)
	if err != nil {
		t.Fatalf("adding route failed: %v", err)
	}
	t.Logf("route added: %+v", newRoute)

	routes, err = h.DumpRoutes()
	routesCnt2 := len(routes)
	if err != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	t.Logf("%d routes dumped", routesCnt2)

	if routesCnt+1 != routesCnt2 {
		t.Errorf("Number of routes after adding of one route is not incremented by 1")
	}

	newRouteIsPresent := false
	for _, route := range routes {
		if (route.Route.DstNetwork == newRoute.DstNetwork) && (route.Route.NextHopAddr == newRoute.NextHopAddr) && (route.Route.OutgoingInterface == newRoute.OutgoingInterface) {
			newRouteIsPresent = true
			break
		}
	}
	if !newRouteIsPresent {
		t.Error("Added route is not present in route dump")
	}

	err = h.VppDelRoute(test.Ctx, &newRoute)
	if err != nil {
		t.Fatalf("deleting route failed: %v", err)
	}
	t.Logf("route deleted")

	routes, err = h.DumpRoutes()
	routesCnt3 := len(routes)
	if err != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	t.Logf("%d routes dumped", routesCnt3)
	if routesCnt2-1 != routesCnt3 {
		t.Errorf("Number of routes after deleting of one route is not decremented by 1")
	}

	for _, route := range routes {
		if (route.Route.DstNetwork == newRoute.DstNetwork) && (route.Route.NextHopAddr == newRoute.NextHopAddr) && (route.Route.OutgoingInterface == newRoute.OutgoingInterface) {
			t.Error("Added route is still present in route dump - should be deleted")
		}
	}

	vrfMetaIdx = 2
	err = h.AddVrfTable(&vpp_l3.VrfTable{Id: vrfMetaIdx, Protocol: vpp_l3.VrfTable_IPV4, Label: "table1"})
	if err != nil {
		t.Fatalf("creating vrf table failed: %v", err)
	}
	t.Logf("vrf table 2 created")
	vrfIndexes.Put("vrf1-ipv4-vrf2", &vrfidx.VRFMetadata{Index: vrfMetaIdx, Protocol: vpp_l3.VrfTable_IPV4})

	routes, errx = h.DumpRoutes()
	if errx != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	routesCnt = len(routes)
	t.Logf("%d routes dumped", routesCnt)

	newRoute = vpp_l3.Route{VrfId: 2, DstNetwork: "192.168.10.0/24", NextHopAddr: "192.168.30.1", OutgoingInterface: ifName}
	err = h.VppAddRoute(test.Ctx, &newRoute)
	if err != nil {
		t.Fatalf("adding route failed: %v", err)
	}
	t.Logf("route added: %+v", newRoute)

	routes, err = h.DumpRoutes()
	routesCnt2 = len(routes)
	if err != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	t.Logf("%d routes dumped", routesCnt2)

	if routesCnt+1 != routesCnt2 {
		t.Errorf("Number of routes after adding of one route is not incremented by 1")
	}

	newRouteIsPresent = false
	for _, route := range routes {
		if (route.Route.DstNetwork == newRoute.DstNetwork) && (route.Route.NextHopAddr == newRoute.NextHopAddr) && (route.Route.OutgoingInterface == newRoute.OutgoingInterface) {
			newRouteIsPresent = true
			break
		}
	}
	if !newRouteIsPresent {
		t.Error("Added route is not present in route dump")
	}

	err = h.VppDelRoute(test.Ctx, &newRoute)
	if err != nil {
		t.Fatalf("deleting route failed: %v", err)
	}
	t.Logf("route deleted")

	routes, err = h.DumpRoutes()
	routesCnt3 = len(routes)
	if err != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	t.Logf("%d routes dumped", routesCnt3)
	if routesCnt2-1 != routesCnt3 {
		t.Errorf("Number of routes after deleting of one route is not decremented by 1")
	}

	for _, route := range routes {
		if (route.Route.DstNetwork == newRoute.DstNetwork) && (route.Route.NextHopAddr == newRoute.NextHopAddr) && (route.Route.OutgoingInterface == newRoute.OutgoingInterface) {
			t.Error("Added route is still present in route dump - should be deleted")
		}
	}
}

func TestCRUDIPv6Route(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(test.vppClient, logrus.NewLogger("test"))
	const ifName = "loop1"
	ifIdx, err := ih.AddLoopbackInterface(ifName)
	if err != nil {
		t.Fatalf("creating interface failed: %v", err)
	}
	t.Logf("interface created %v", ifIdx)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-iface1"), "test-iface1")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{
		SwIfIndex: ifIdx,
	})

	var vrfMetaIdx uint32
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv6-vrf0", &vrfidx.VRFMetadata{Index: vrfMetaIdx, Protocol: vpp_l3.VrfTable_IPV6})

	h := l3plugin_vppcalls.CompatibleL3VppHandler(test.vppClient, ifIndexes, vrfIndexes,
		netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test"))

	routes, errx := h.DumpRoutes()
	if errx != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	routesCnt := len(routes)
	t.Logf("%d routes dumped", routesCnt)

	newRoute := vpp_l3.Route{VrfId: 0, DstNetwork: "fd30:0:0:1::/64", NextHopAddr: "fd31::1:1:0:0:1", OutgoingInterface: ifName}
	err = h.VppAddRoute(test.Ctx, &newRoute)
	if err != nil {
		t.Fatalf("adding route failed: %v", err)
	}
	t.Logf("route added: %+v", newRoute)

	routes, err = h.DumpRoutes()
	routesCnt2 := len(routes)
	if err != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	t.Logf("%d routes dumped", routesCnt2)

	if routesCnt+1 != routesCnt2 {
		t.Errorf("Number of routes after adding of one route is not incremented by 1")
	}

	newRouteIsPresent := false
	for _, route := range routes {
		if (route.Route.DstNetwork == newRoute.DstNetwork) && (route.Route.NextHopAddr == newRoute.NextHopAddr) && (route.Route.OutgoingInterface == newRoute.OutgoingInterface) {
			newRouteIsPresent = true
		}
	}
	if !newRouteIsPresent {
		t.Error("Added route is not present in route dump")
	}

	err = h.VppDelRoute(test.Ctx, &newRoute)
	if err != nil {
		t.Fatalf("deleting route failed: %v", err)
	}
	t.Logf("route deleted")

	routes, err = h.DumpRoutes()
	routesCnt3 := len(routes)
	if err != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	t.Logf("%d routes dumped", routesCnt3)
	if routesCnt2-1 != routesCnt3 {
		t.Errorf("Number of routes after deleting of one route is not decremented by 1")
	}

	for _, route := range routes {
		if (route.Route.DstNetwork == newRoute.DstNetwork) && (route.Route.NextHopAddr == newRoute.NextHopAddr) && (route.Route.OutgoingInterface == newRoute.OutgoingInterface) {
			t.Error("Added route is still present in route dump - should be deleted")
		}
	}

	vrfMetaIdx = 2
	err = h.AddVrfTable(&vpp_l3.VrfTable{Id: vrfMetaIdx, Protocol: vpp_l3.VrfTable_IPV6, Label: "table1"})
	if err != nil {
		t.Fatalf("creating vrf table failed: %v", err)
	}
	t.Logf("vrf table 2 created")
	vrfIndexes.Put("vrf1-ipv6-vrf2", &vrfidx.VRFMetadata{Index: vrfMetaIdx, Protocol: vpp_l3.VrfTable_IPV6})

	routes, errx = h.DumpRoutes()
	if errx != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	routesCnt = len(routes)
	t.Logf("%d routes dumped", routesCnt)

	newRoute = vpp_l3.Route{VrfId: 2, DstNetwork: "fd30:0:0:1::/64", NextHopAddr: "fd31::1:1:0:0:1", OutgoingInterface: ifName}
	err = h.VppAddRoute(test.Ctx, &newRoute)
	if err != nil {
		t.Fatalf("adding route failed: %v", err)
	}
	t.Logf("route added: %+v", newRoute)

	routes, err = h.DumpRoutes()
	routesCnt2 = len(routes)
	if err != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	t.Logf("%d routes dumped", routesCnt2)

	if routesCnt+1 != routesCnt2 {
		t.Errorf("Number of routes after adding of one route is not incremented by 1")
	}

	newRouteIsPresent = false
	for _, route := range routes {
		if (route.Route.DstNetwork == newRoute.DstNetwork) && (route.Route.NextHopAddr == newRoute.NextHopAddr) && (route.Route.OutgoingInterface == newRoute.OutgoingInterface) {
			newRouteIsPresent = true
		}
	}
	if !newRouteIsPresent {
		t.Error("Added route is not present in route dump")
	}

	err = h.VppDelRoute(test.Ctx, &newRoute)
	if err != nil {
		t.Fatalf("deleting route failed: %v", err)
	}
	t.Logf("route deleted")

	routes, err = h.DumpRoutes()
	routesCnt3 = len(routes)
	if err != nil {
		t.Fatalf("dumping routes failed: %v", err)
	}
	t.Logf("%d routes dumped", routesCnt3)
	if routesCnt2-1 != routesCnt3 {
		t.Errorf("Number of routes after deleting of one route is not decremented by 1")
	}

	for _, route := range routes {
		if (route.Route.DstNetwork == newRoute.DstNetwork) && (route.Route.NextHopAddr == newRoute.NextHopAddr) && (route.Route.OutgoingInterface == newRoute.OutgoingInterface) {
			t.Error("Added route is still present in route dump - should be deleted")
		}
	}
}
