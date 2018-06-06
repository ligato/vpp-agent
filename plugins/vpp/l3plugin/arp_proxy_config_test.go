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
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
	"testing"
)

// Test ProxyArpConfigurator initialization
func TestProxyArpConfiguratorInit(t *testing.T) {
	RegisterTestingT(t)
	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}
	connection, _ := core.Connect(ctx.MockVpp)
	defer connection.Disconnect()

	plugin := &l3plugin.ProxyArpConfigurator{}

	// Test init
	err := plugin.Init(logging.ForPlugin("test-log", logrus.NewLogRegistry()), connection, nil, false)
	Expect(err).To(BeNil())
	err = plugin.Close()
	Expect(err).To(BeNil())
}

var proxyarpAddEntries = []l3.ProxyArpInterfaces_InterfaceList{
	{
		Label: "proxyArpIf1",
		Interfaces: []*l3.ProxyArpInterfaces_InterfaceList_Interface{
			{
				Name: "",
			},
		},
	},
	{
		Label: "proxyArpIf2",
		Interfaces: []*l3.ProxyArpInterfaces_InterfaceList_Interface{
			{
				Name: "tap3",
			},
		},
	},
	{
		Label: "proxyArpIf3",
		Interfaces: []*l3.ProxyArpInterfaces_InterfaceList_Interface{
			{
				Name: "tap2",
			},
		},
	},
	{
		Label: "proxyArpIf4",
		Interfaces: []*l3.ProxyArpInterfaces_InterfaceList_Interface{
			{
				Name: "tap1",
			},
		},
	},
}

var proxyarpRangeEntries = []l3.ProxyArpRanges_RangeList{
	{
		Label: "proxyArpIf1",
		Ranges: []*l3.ProxyArpRanges_RangeList_Range{
			{
				FirstIp: "124.168.10.5",
				LastIp:  "124.168.10.10",
			},
			{
				FirstIp: "124.168.20.0/24",
				LastIp:  "124.168.20.0/24",
			},
		},
	},
	{
		Label: "proxyArpIfErr",
		Ranges: []*l3.ProxyArpRanges_RangeList_Range{
			{
				FirstIp: "124.168.20.0/24/32",
				LastIp:  "124.168.30.10",
			},
			{
				FirstIp: "124.168.20.5",
				LastIp:  "124.168.30.5/16/24",
			},
		},
	},
	{
		Label: "proxyArpIf2",
		Ranges: []*l3.ProxyArpRanges_RangeList_Range{
			{
				FirstIp: "124.168.10.5",
				LastIp:  "124.168.10.10",
			},
			{
				FirstIp: "172.154.100.0/24",
				LastIp:  "172.154.200.0/24",
			},
		},
	},
}

// Test adding of ARP proxy entry
func TestAddInterface(t *testing.T) {
	// Setup
	ctx, connection, plugin := proxyarpTestSetup(t)
	defer proxyarpTestTeardown(connection, plugin)

	err := plugin.AddInterface(&proxyarpAddEntries[0])
	Expect(err).To(Not(BeNil()))
	err = plugin.AddInterface(&proxyarpAddEntries[1])
	Expect(err).To(BeNil())
	err = plugin.AddInterface(&proxyarpAddEntries[2])
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&ip.ProxyArpIntfcEnableDisableReply{})
	err = plugin.AddInterface(&proxyarpAddEntries[3])
	Expect(err).To(BeNil())
}

// Test deleting of ARP proxy entry
func TestDeleteInterface(t *testing.T) {
	// Setup
	ctx, connection, plugin := proxyarpTestSetup(t)
	defer proxyarpTestTeardown(connection, plugin)

	//DeleteInterface
	err := plugin.DeleteInterface(&proxyarpAddEntries[0])
	Expect(err).To(BeNil())
	err = plugin.DeleteInterface(&proxyarpAddEntries[2])
	Expect(err).To(Not(BeNil()))
	err = plugin.AddInterface(&proxyarpAddEntries[1])
	Expect(err).To(BeNil())
	err = plugin.DeleteInterface(&proxyarpAddEntries[1])
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&ip.ProxyArpIntfcEnableDisableReply{})
	err = plugin.DeleteInterface(&proxyarpAddEntries[3])
	Expect(err).To(BeNil())
}

// Test deleting of ARP proxy entry
func TestModifyInterface(t *testing.T) {
	// Setup
	ctx, connection, plugin := proxyarpTestSetup(t)
	defer proxyarpTestTeardown(connection, plugin)

	err := plugin.AddInterface(&proxyarpAddEntries[1])
	Expect(err).To(BeNil())
	err = plugin.ModifyInterface(&proxyarpAddEntries[2], &proxyarpAddEntries[1])
	Expect(err).To(Not(BeNil()))
	err = plugin.ModifyInterface(&proxyarpAddEntries[2], &proxyarpAddEntries[3])
	Expect(err).To(Not(BeNil()))
	ctx.MockVpp.MockReply(&ip.ProxyArpIntfcEnableDisableReply{})
	err = plugin.ModifyInterface(&proxyarpAddEntries[1], &proxyarpAddEntries[2])
	Expect(err).To(BeNil())
	ctx.MockVpp.MockReply(&ip.ProxyArpIntfcEnableDisableReply{})
	ctx.MockVpp.MockReply(&ip.ProxyArpIntfcEnableDisableReply{})
	err = plugin.ModifyInterface(&proxyarpAddEntries[2], &proxyarpAddEntries[3])
	Expect(err).To(BeNil())
}

// Test adding of ARP proxy range
func TestAddRange(t *testing.T) {
	// Setup
	ctx, connection, plugin := proxyarpTestSetup(t)
	defer proxyarpTestTeardown(connection, plugin)

	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err := plugin.AddRange(&proxyarpRangeEntries[0])
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err = plugin.AddRange(&proxyarpRangeEntries[0])
	Expect(err).To(Not(BeNil()))

	err = plugin.AddRange(&proxyarpRangeEntries[1])
	Expect(err).To(Not(BeNil()))
}

// Test deleting of ARP proxy range
func TestDeleteRange(t *testing.T) {
	// Setup
	ctx, connection, plugin := proxyarpTestSetup(t)
	defer proxyarpTestTeardown(connection, plugin)

	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err := plugin.DeleteRange(&proxyarpRangeEntries[0])
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err = plugin.DeleteRange(&proxyarpRangeEntries[0])
	Expect(err).To(Not(BeNil()))

	err = plugin.DeleteRange(&proxyarpRangeEntries[1])
	Expect(err).To(Not(BeNil()))
}

// Test modification of ARP proxy range
func TestModifyRange(t *testing.T) {
	// Setup
	ctx, connection, plugin := proxyarpTestSetup(t)
	defer proxyarpTestTeardown(connection, plugin)

	err := plugin.ModifyRange(&proxyarpRangeEntries[2], &proxyarpRangeEntries[0])
	Expect(err).To(Not(BeNil()))
	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err = plugin.ModifyRange(&proxyarpRangeEntries[2], &proxyarpRangeEntries[0])
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err = plugin.ModifyRange(&proxyarpRangeEntries[2], &proxyarpRangeEntries[0])
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err = plugin.ModifyRange(&proxyarpRangeEntries[1], &proxyarpRangeEntries[2])
	Expect(err).To(Not(BeNil()))
	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err = plugin.ModifyRange(&proxyarpRangeEntries[2], &proxyarpRangeEntries[1])
	Expect(err).To(Not(BeNil()))
}

// Test resolution of new registered interface for proxy ARP
func TestArpProxyResolveCreatedInterface(t *testing.T) {
	// Setup
	_, connection, plugin := proxyarpTestSetup(t)
	defer proxyarpTestTeardown(connection, plugin)

	err := plugin.AddInterface(&proxyarpAddEntries[1])
	Expect(err).To(BeNil())
	err = plugin.ResolveCreatedInterface("tap3", 1)
	Expect(err).To(Not(BeNil()))
}

// Test resolution of new registered interface for proxy ARP
func TestArpProxyResolveDeletedInterface(t *testing.T) {
	// Setup
	ctx, connection, plugin := proxyarpTestSetup(t)
	defer proxyarpTestTeardown(connection, plugin)

	ctx.MockVpp.MockReply(&ip.ProxyArpIntfcEnableDisableReply{})
	err := plugin.AddInterface(&proxyarpAddEntries[3])
	Expect(err).To(BeNil())
	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err = plugin.ResolveDeletedInterface("proxyArpIf4")
	Expect(err).To(BeNil())
}

// Test Setup
func proxyarpTestSetup(t *testing.T) (*vppcallmock.TestCtx, *core.Connection, *l3plugin.ProxyArpConfigurator) {
	RegisterTestingT(t)
	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}
	connection, err := core.Connect(ctx.MockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	plugin := &l3plugin.ProxyArpConfigurator{}
	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logging.ForPlugin("test-log", logrus.NewLogRegistry()), "l3-plugin", nil))
	ifIndexes.RegisterName("tap1", 1, nil)
	ifIndexes.RegisterName("tap2", 2, nil)

	err = plugin.Init(logging.ForPlugin("test-log", logrus.NewLogRegistry()), connection, ifIndexes, false)
	Expect(err).To(BeNil())

	return ctx, connection, plugin
}

// Test Teardown
func proxyarpTestTeardown(connection *core.Connection, plugin *l3plugin.ProxyArpConfigurator) {
	connection.Disconnect()
	err := plugin.Close()
	Expect(err).To(BeNil())
}
