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

// Test ArpConfigurator initialization
func TestArpConfiguratorInit(t *testing.T) {
	RegisterTestingT(t)
	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}
	connection, _ := core.Connect(ctx.MockVpp)
	defer connection.Disconnect()

	plugin := &l3plugin.ArpConfigurator{}

	// Test init
	err := plugin.Init(logging.ForPlugin("test-log", logrus.NewLogRegistry()), connection, nil, false)
	Expect(err).To(BeNil())
	err = plugin.Close()
	Expect(err).To(BeNil())
}

var arpTableEntries = []l3.ArpTable_ArpEntry{
	{
		Interface:   "tap1",
		IpAddress:   "192.168.10.21",
		PhysAddress: "59:6C:DE:AD:00:01",
		Static:      true,
	},
	{
		Interface:   "tap2",
		IpAddress:   "192.168.10.22",
		PhysAddress: "59:6C:DE:AD:00:02",
		Static:      true,
	},
	{
		Interface:   "tap3",
		IpAddress:   "dead::01",
		PhysAddress: "59:6C:DE:AD:00:03",
		Static:      false,
	},
	{
		Interface:   "tap4",
		IpAddress:   "dead::02",
		PhysAddress: "59:6C:DE:AD:00:04",
		Static:      false,
	},
}

// Test adding of ARP entry
func TestAddArp(t *testing.T) {
	// Setup
	ctx, connection, plugin := arpTestSetup(t)
	defer arpTestTeardown(connection, plugin)

	err := plugin.AddArp(&arpTableEntries[2])
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	err = plugin.AddArp(&arpTableEntries[0])
	Expect(err).To(BeNil())

	// Test isValidARP
	err = plugin.AddArp(&l3.ArpTable_ArpEntry{})
	Expect(err).To(Not(BeNil()))
	err = plugin.AddArp(&l3.ArpTable_ArpEntry{Interface: "tap5"})
	Expect(err).To(Not(BeNil()))
	err = plugin.AddArp(&l3.ArpTable_ArpEntry{Interface: "tap6", IpAddress: "192.168.10.33"})
	Expect(err).To(Not(BeNil()))
}

// Test deleting of ARP entry
func TestDeleteArp(t *testing.T) {
	// Setup
	ctx, connection, plugin := arpTestSetup(t)
	defer arpTestTeardown(connection, plugin)

	// delete arp for non existing intf
	err := plugin.DeleteArp(&arpTableEntries[3])
	Expect(err).To(BeNil())

	// delete arp-chached ARP entry
	err = plugin.AddArp(&arpTableEntries[2])
	Expect(err).To(BeNil())
	err = plugin.DeleteArp(&arpTableEntries[2])
	Expect(err).To(BeNil())

	// delete added ARP entry
	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	err = plugin.AddArp(&arpTableEntries[0])
	Expect(err).To(BeNil())
	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	err = plugin.DeleteArp(&arpTableEntries[0])
	Expect(err).To(BeNil())
}

// Test changing of ARP entry
func TestChangeArp(t *testing.T) {
	// Setup
	ctx, connection, plugin := arpTestSetup(t)
	defer arpTestTeardown(connection, plugin)

	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	err := plugin.AddArp(&arpTableEntries[0])
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	err = plugin.ChangeArp(&arpTableEntries[1], &arpTableEntries[0])
	Expect(err).To(BeNil())
}

// Test resolving of created ARPs
func TestArpResolveCreatedInterface(t *testing.T) {
	// Setup
	_, connection, plugin := arpTestSetup(t)
	defer arpTestTeardown(connection, plugin)

	err := plugin.AddArp(&arpTableEntries[2])
	Expect(err).To(BeNil())
	err = plugin.DeleteArp(&arpTableEntries[3])
	Expect(err).To(BeNil())

	err = plugin.ResolveCreatedInterface("tap3")
	Expect(err).To(BeNil())
	err = plugin.ResolveCreatedInterface("tap4")
	Expect(err).To(BeNil())
}

// Test resolving of created ARPs
func TestArpResolveDeletedInterface(t *testing.T) {
	// Setup
	_, connection, plugin := arpTestSetup(t)
	defer arpTestTeardown(connection, plugin)

	err := plugin.ResolveDeletedInterface("tap4", 3)
	Expect(err).To(BeNil())
}

// ARP Test Setup
func arpTestSetup(t *testing.T) (*vppcallmock.TestCtx, *core.Connection, *l3plugin.ArpConfigurator) {
	RegisterTestingT(t)
	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}
	connection, err := core.Connect(ctx.MockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	plugin := &l3plugin.ArpConfigurator{}
	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logging.ForPlugin("test-log", logrus.NewLogRegistry()), "l3-plugin", nil))
	ifIndexes.RegisterName("tap1", 1, nil)
	ifIndexes.RegisterName("tap2", 2, nil)

	err = plugin.Init(logging.ForPlugin("test-log", logrus.NewLogRegistry()), connection, ifIndexes, false)
	Expect(err).To(BeNil())

	return ctx, connection, plugin
}

// Test Teardown
func arpTestTeardown(connection *core.Connection, plugin *l3plugin.ArpConfigurator) {
	connection.Disconnect()
	err := plugin.Close()
	Expect(err).To(BeNil())
}
