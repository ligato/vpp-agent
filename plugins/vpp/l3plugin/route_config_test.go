// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package l3plugin_test

import (
	"git.fd.io/govpp.git/adapter/mock"
	"git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
	"testing"
)

// Test RouteConfigurator initialization
func TestRouteConfiguratorInit(t *testing.T) {
	RegisterTestingT(t)
	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}
	connection, _ := core.Connect(ctx.MockVpp)
	defer connection.Disconnect()

	plugin := &l3plugin.RouteConfigurator{}

	// Test init
	err := plugin.Init(logging.ForPlugin("test-log", logrus.NewLogRegistry()), connection, nil, false)
	Expect(err).To(BeNil())
	err = plugin.Close()
	Expect(err).To(BeNil())
}

// Test adding of routes entry
func TestConfigureRoute(t *testing.T) {
	// Setup
	ctx, connection, plugin, ifIndexes := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	// add interface
	ifIndexes.GetMapping().RegisterName("tap1", 1, nil)

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err := plugin.ConfigureRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(plugin.GetCachedRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
	routes := plugin.GetRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-10.1.1.3/32-192.168.1.13"))
}

// Test adding of routes entry with invalid "VrfFromKey"
func TestConfigureRouteValidateVrfFromKey(t *testing.T) {
	// Setup
	ctx, connection, plugin, ifIndexes := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	// add interface
	ifIndexes.GetMapping().RegisterName("tap1", 1, nil)

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err := plugin.ConfigureRoute(&l3.StaticRoutes_Route{
		VrfId:             1,
		DstIpAddr:         "20.2.2.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	routes := plugin.GetRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-20.2.2.3/32-192.168.1.13"))
}

// Test adding of routes entry to cached indexes
func TestConfigureCachedRoute(t *testing.T) {
	// Setup
	_, connection, plugin, _ := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	err := plugin.ConfigureRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	routes := plugin.GetCachedRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-10.1.1.3/32-192.168.1.13"))
}

// Test deletion of route entry
func TestDeleteRoute(t *testing.T) {
	// Setup
	ctx, connection, plugin, ifIndexes := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	// add interface
	ifIndexes.GetMapping().RegisterName("tap1", 1, nil)

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err := plugin.ConfigureRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(plugin.GetCachedRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
	Expect(plugin.GetRouteIndexes().GetMapping().ListNames()).To(HaveLen(1))

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err = plugin.DeleteRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(plugin.GetCachedRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
	Expect(plugin.GetRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
}

// Test deletion of cached route entry
func TestDeleteCachedRoute(t *testing.T) {
	// Setup
	ctx, connection, plugin, ifIndexes := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	err := plugin.ConfigureRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	routes := plugin.GetCachedRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-10.1.1.3/32-192.168.1.13"))

	// add interface
	ifIndexes.GetMapping().RegisterName("tap1", 1, nil)
	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err = plugin.DeleteRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(plugin.GetCachedRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
	Expect(plugin.GetRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
}

// Test modify of existing route
func TestModifyRoute(t *testing.T) {
	// Setup
	ctx, connection, plugin, ifIndexes := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	// add interfaces
	ifIndexes.GetMapping().RegisterName("tap1", 1, nil)
	plugin.GetRouteIndexes().GetMapping().RegisterName("vrf0-10.1.1.3/32-192.168.1.13", 1, nil)

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err := plugin.ModifyRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "fd21:dead:abcd::/48",
		NextHopAddr:       "fd21:cdef:dead::",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, &l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())

	routes := plugin.GetRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-fd21:dead:abcd::/48-fd21:cdef:dead::"))
}

// Test modify of cached route
func TestModifyCachedRoute(t *testing.T) {
	// Setup
	ctx, connection, plugin, _ := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	err := plugin.ConfigureRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	routes := plugin.GetCachedRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-10.1.1.3/32-192.168.1.13"))

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err = plugin.ModifyRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "fd21:dead:abcd::/48",
		NextHopAddr:       "fd21:cdef:dead::",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, &l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).Should(HaveOccurred())

	Expect(plugin.GetRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
	routes = plugin.GetCachedRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-fd21:dead:abcd::/48-fd21:cdef:dead::"))
}

// Test modify of cached route
func TestModifyCachedRouteInterface(t *testing.T) {
	// Setup
	ctx, connection, plugin, ifIndexes := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	// add interface
	ifIndexes.GetMapping().RegisterName("tap2", 1, nil)

	err := plugin.ConfigureRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	routes := plugin.GetCachedRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-10.1.1.3/32-192.168.1.13"))

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err = plugin.ModifyRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "fd21:dead:abcd::/48",
		NextHopAddr:       "fd21:cdef:dead::",
		Weight:            6,
		OutgoingInterface: "tap2",
	}, &l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())

	Expect(plugin.GetCachedRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
	routes = plugin.GetRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-fd21:dead:abcd::/48-fd21:cdef:dead::"))
}

// Test modify of existing route from no-default VRF
func TestModifyRouteVrf(t *testing.T) {
	// Setup
	ctx, connection, plugin, ifIndexes := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	// add interface
	ifIndexes.GetMapping().RegisterName("tap1", 1, nil)
	plugin.GetRouteIndexes().GetMapping().RegisterName("vrf1-fd21:dead:abcd::/48-fd21:cdef:dead::", 1,
		&l3.StaticRoutes_Route{
			VrfId:             1,
			DstIpAddr:         "fd21:dead:abcd::/48",
			NextHopAddr:       "fd21:cdef:dead::",
			Weight:            6,
			OutgoingInterface: "tap1",
		})

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	ctx.MockVpp.MockReply(&ip.IPFibDetails{
		Path: []ip.FibPath{{SwIfIndex: 1}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err := plugin.ModifyRoute(&l3.StaticRoutes_Route{
		VrfId:             1,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, &l3.StaticRoutes_Route{
		VrfId:             1,
		DstIpAddr:         "fd21:dead:abcd::/48",
		NextHopAddr:       "fd21:cdef:dead::",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "1")
	Expect(err).ShouldNot(HaveOccurred())

	routes := plugin.GetRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf1-10.1.1.3/32-192.168.1.13"))
}

// Test deletion of cached route entry
func TestConfigureAndResolveRoute(t *testing.T) {
	// Setup
	ctx, connection, plugin, ifIndexes := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	err := plugin.ConfigureRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "fd21:dead:abcd::/48",
		NextHopAddr:       "fd21:cdef:dead::",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(plugin.GetCachedRouteIndexes().GetMapping().ListNames()).To(HaveLen(1))
	Expect(plugin.GetRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))

	// add interface
	ifIndexes.GetMapping().RegisterName("tap1", 1, nil)
	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	// reslove interfaces
	plugin.ResolveCreatedInterface("tap1", 1)
	Expect(plugin.GetCachedRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
	routes := plugin.GetRouteIndexes().GetMapping().ListNames()
	Expect(routes).To(HaveLen(1))
	Expect(routes[0]).To(Equal("vrf0-fd21:dead:abcd::/48-fd21:cdef:dead::"))
}

// Test resolving routes of deleted interface
func TestResolveDeletedRoute(t *testing.T) {
	// Setup
	ctx, connection, plugin, ifIndexes := routeTestSetup(t)
	defer routeTestTeardown(connection, plugin)

	// add interface
	ifIndexes.GetMapping().RegisterName("tap1", 1, nil)
	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err := plugin.ConfigureRoute(&l3.StaticRoutes_Route{
		VrfId:             0,
		DstIpAddr:         "10.1.1.3/32",
		NextHopAddr:       "192.168.1.13",
		Weight:            6,
		OutgoingInterface: "tap1",
	}, "0")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(plugin.GetCachedRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
	Expect(plugin.GetRouteIndexes().GetMapping().ListNames()).To(HaveLen(1))

	ifIndexes.GetMapping().UnregisterName("tap1")
	plugin.ResolveDeletedInterface("tap1", 1)
	Expect(plugin.GetCachedRouteIndexes().GetMapping().ListNames()).To(HaveLen(1))
	Expect(plugin.GetRouteIndexes().GetMapping().ListNames()).To(HaveLen(0))
}

// Rotue Test Setup
func routeTestSetup(t *testing.T) (*vppcallmock.TestCtx, *core.Connection, *l3plugin.RouteConfigurator, ifaceidx.SwIfIndex) {
	RegisterTestingT(t)
	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}
	connection, err := core.Connect(ctx.MockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	plugin := &l3plugin.RouteConfigurator{}
	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logging.ForPlugin("test-log", logrus.NewLogRegistry()), "l3-plugin", nil))

	err = plugin.Init(logging.ForPlugin("test-log", logrus.NewLogRegistry()), connection, ifIndexes, false)
	Expect(err).To(BeNil())

	return ctx, connection, plugin, ifIndexes
}

// Test Teardown
func routeTestTeardown(connection *core.Connection, plugin *l3plugin.RouteConfigurator) {
	connection.Disconnect()
	err := plugin.Close()
	Expect(err).To(BeNil())
}
