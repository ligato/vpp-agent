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

package ifplugin

import (
	"testing"

	"git.fd.io/govpp.git/adapter/mock"
	"git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	ap_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/af_packet"
	if_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

var afPacketNames = []string{"af-packet", "af-packet-2"}
var afPacketHosts = []string{"af-packet-host1", "af-packet-host2"}
var netAddresses = []string{"10.0.0.1/24", "10.0.0.2/24", "192.168.50.1/24"}

/* AF_PACKET configurator init */

// Test init function
func TestAfPacketConfiguratorInit(t *testing.T) {
	RegisterTestingT(t)
	connection, err := core.Connect(&mock.VppAdapter{})
	Expect(err).To(BeNil())
	plugin := &AFPacketConfigurator{
		Logger: logrus.DefaultLogger(),
	}
	vppCh, err := connection.NewAPIChannel()
	Expect(err).To(BeNil())
	err = plugin.Init(vppCh)
	Expect(err).To(BeNil())
	Expect(plugin.vppCh).ToNot(BeNil())
	Expect(plugin.afPacketByHostIf).ToNot(BeNil())
	Expect(plugin.afPacketByName).ToNot(BeNil())
	connection.Disconnect()
}

/* AF_PACKET test cases */

// Configure af packet interface with unavailable host
func TestAfPacketConfigureHostNotAvail(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Data
	var addresses []string
	data := getTestAfPacketData(afPacketNames[0], append(addresses, netAddresses[0]), afPacketHosts[0])
	// Test configure af packet with host unavailable
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeZero())
	Expect(pending).To(BeTrue())
	cached, ok := plugin.afPacketByName[afPacketNames[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketNames[0]))
	Expect(cached.pending).To(BeTrue())
	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketNames[0]))
}

// Configure af packet interface
func TestAfPacketConfigureHostAvail(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 2,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Data
	var addresses []string
	data := getTestAfPacketData(afPacketNames[0], append(addresses, netAddresses[0]), afPacketHosts[0])
	// Test af packet
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(2))
	Expect(pending).To(BeFalse())
	cached, ok := plugin.afPacketByName[afPacketNames[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketNames[0]))
	Expect(cached.pending).To(BeFalse())
	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketNames[0]))
}

// Configure af packet with error reply from VPP API
func TestAfPacketConfigureHostAvailError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		Retval:    1,
		SwIfIndex: 2,
	})
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Data
	var addresses []string
	data := getTestAfPacketData(afPacketNames[0], append(addresses, netAddresses[0]), afPacketHosts[0])
	// Test configure af packet with return value != 0
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data)
	Expect(err).ToNot(BeNil())
	Expect(swIfIdx).To(BeZero())
	Expect(pending).To(BeTrue())
	cached, ok := plugin.afPacketByName[afPacketNames[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
}

// Configure af packet as incorrect interface type
func TestAfPacketConfigureIncorrectTypeError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Data
	var addresses []string
	data := getTestAfPacketData(afPacketNames[0], append(addresses, netAddresses[0]), afPacketHosts[0])
	data.Type = interfaces.InterfaceType_SOFTWARE_LOOPBACK
	// Test configure af packet with incorrect type
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data)
	Expect(err).ToNot(BeNil())
	Expect(swIfIdx).To(BeZero())
	Expect(pending).To(BeFalse())
	_, ok := plugin.afPacketByName[afPacketNames[0]]
	Expect(ok).To(BeFalse())
	_, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeFalse())
}

// Call af packet modification which causes recreation of the interface
func TestAfPacketModifyRecreateChangedHost(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Data
	var oldAddresses, newAddresses []string
	oldData := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	newData := getTestAfPacketData(afPacketNames[0], append(newAddresses, netAddresses[1]), afPacketHosts[1])
	// Test configure initial af packet data
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(oldData)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())
	// Test modify af packet
	recreate, err := plugin.ModifyAfPacketInterface(newData, oldData)
	Expect(err).To(BeNil())
	Expect(recreate).To(BeTrue())
}

// Test modify pending af packet interface
func TestAfPacketModifyRecreatePending(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Data
	var oldAddresses, newAddresses []string
	oldData := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	newData := getTestAfPacketData(afPacketNames[0], append(newAddresses, netAddresses[0]), afPacketHosts[0])
	// Test configure initial af packet data
	_, pending, err := plugin.ConfigureAfPacketInterface(oldData)
	Expect(err).To(BeNil())
	Expect(pending).To(BeTrue())
	// Test modify
	recreate, err := plugin.ModifyAfPacketInterface(newData, oldData)
	Expect(err).To(BeNil())
	Expect(recreate).To(BeTrue())
}

// Modify recreate of af packet interface which was not found
func TestAfPacketModifyRecreateNotFound(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Data
	var oldAddresses, newAddresses []string
	oldData := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	newData := getTestAfPacketData(afPacketNames[0], append(newAddresses, netAddresses[1]), afPacketHosts[1])
	// Test af packet modify
	recreate, err := plugin.ModifyAfPacketInterface(newData, oldData)
	Expect(err).To(BeNil())
	Expect(recreate).To(BeTrue())
}

// Modify af packet interface without recreation
func TestAfPacketModifyNoRecreate(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Data
	var oldAddresses, newAddresses []string
	oldData := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	newData := getTestAfPacketData(afPacketNames[0], append(newAddresses, netAddresses[1]), afPacketHosts[0])
	// Test configure initial data
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(oldData)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())
	// Test modify
	recreate, err := plugin.ModifyAfPacketInterface(newData, oldData)
	Expect(err).To(BeNil())
	Expect(recreate).To(BeFalse())
	cached, ok := plugin.afPacketByName[afPacketNames[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.IpAddresses[0]).To(BeEquivalentTo(netAddresses[1]))
	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.IpAddresses[0]).To(BeEquivalentTo(netAddresses[1]))
}

// Modify af packet with incorrect interface type
func TestAfPacketModifyIncorrectType(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Data
	var oldAddresses, newAddresses []string
	oldData := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	newData := getTestAfPacketData(afPacketNames[0], append(newAddresses, netAddresses[1]), afPacketHosts[0])
	newData.Type = interfaces.InterfaceType_SOFTWARE_LOOPBACK
	// Test configure initial data
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(oldData)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())
	// Test modify with incorrect type
	_, err = plugin.ModifyAfPacketInterface(newData, oldData)
	Expect(err).ToNot(BeNil())
}

// Af packet delete
func TestAfPacketDelete(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{ // Create
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&ap_api.AfPacketDeleteReply{}) // Delete
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Data
	var oldAddresses []string
	oldData := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	// Test configure initial af packet data
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(oldData)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())
	cached, ok := plugin.afPacketByName[afPacketNames[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketNames[0]))
	Expect(cached.pending).To(BeFalse())
	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketNames[0]))
	// Test af packet delete
	err = plugin.DeleteAfPacketInterface(oldData, 1)
	Expect(err).To(BeNil())
	_, ok = plugin.afPacketByName[afPacketNames[0]]
	Expect(ok).To(BeFalse())
	_, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeFalse())
}

// Delete af packet with incorrect interface type data
func TestAfPacketDeleteIncorrectType(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Data
	var oldAddresses []string
	data := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	modifiedData := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	modifiedData.Type = interfaces.InterfaceType_SOFTWARE_LOOPBACK
	// Test configure initial af packet
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())
	// Test delete with incorrect type
	err = plugin.DeleteAfPacketInterface(modifiedData, 1)
	Expect(err).ToNot(BeNil())
}

// Register new linux interface and test af packet behaviour
func TestAfPacketNewLinuxInterfaceHostFound(t *testing.T) {
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Data
	var oldAddresses []string
	data := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	// Fill af packet cache
	plugin.afPacketByHostIf[data.Afpacket.HostIfName] = &AfPacketConfig{
		config:  data,
		pending: true,
	}
	_, ok := plugin.hostInterfaces[data.Afpacket.HostIfName]
	Expect(ok).To(BeFalse())
	// Test registered linux interface
	config := plugin.ResolveCreatedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	Expect(config).ToNot(BeNil())
	Expect(config.Afpacket.HostIfName).To(BeEquivalentTo(afPacketHosts[0]))
	_, ok = plugin.hostInterfaces[data.Afpacket.HostIfName]
	Expect(ok).To(BeTrue())
}

// Register new linux interface while af packet is pending. Note: this is a case which should NOT happen
func TestAfPacketNewLinuxInterfaceHostFoundPending(t *testing.T) {
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketDeleteReply{})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Data
	var oldAddresses []string
	data := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	// Fill af packet cache
	plugin.afPacketByHostIf[data.Afpacket.HostIfName] = &AfPacketConfig{
		config:  data,
		pending: false,
	}
	_, ok := plugin.hostInterfaces[data.Afpacket.HostIfName]
	Expect(ok).To(BeFalse())
	// Test registered linux interface
	config := plugin.ResolveCreatedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	Expect(config).ToNot(BeNil())
	Expect(config.Afpacket.HostIfName).To(BeEquivalentTo(afPacketHosts[0]))
	_, ok = plugin.hostInterfaces[data.Afpacket.HostIfName]
	Expect(ok).To(BeTrue())
}

// Test new linux interface which is not a host
func TestAfPacketNewLinuxInterfaceHostNotFound(t *testing.T) {
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Data
	var oldAddresses []string
	data := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	_, ok := plugin.hostInterfaces[data.Afpacket.HostIfName]
	Expect(ok).To(BeFalse())
	// Test registered linux interface
	config := plugin.ResolveCreatedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	Expect(config).To(BeNil())
	_, ok = plugin.hostInterfaces[data.Afpacket.HostIfName]
	Expect(ok).To(BeTrue())
}

// Test new linux interface while linux plugin is not available
func TestAfPacketNewLinuxInterfaceNoLinux(t *testing.T) {
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	plugin.Linux = nil
	defer afPacketTestTeardown(ctx)
	// Test registered linux interface
	config := plugin.ResolveCreatedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	Expect(config).To(BeNil())
}

// Un-register linux interface
func TestAfPacketDeletedLinuxInterface(t *testing.T) {
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Reply set
	ctx.MockVpp.MockReply(&ap_api.AfPacketDeleteReply{})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Data
	var oldAddresses []string
	data := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	// Fill af packet cache
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	plugin.afPacketByName[afPacketNames[0]] = &AfPacketConfig{
		config:  data,
		pending: false,
	}
	plugin.afPacketByHostIf[data.Afpacket.HostIfName] = &AfPacketConfig{
		config:  data,
		pending: true,
	}
	// Test un-registered linux interface
	plugin.ResolveDeletedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	_, ok := plugin.hostInterfaces[afPacketHosts[0]]
	Expect(ok).To(BeFalse())
	_, ok = plugin.afPacketByName[data.Name]
	Expect(ok).To(BeTrue())
	_, ok = plugin.afPacketByHostIf[data.Afpacket.HostIfName]
	Expect(ok).To(BeTrue())
}

// Un-register linux interface while host is not found
func TestAfPacketDeletedLinuxInterfaceHostNotFound(t *testing.T) {
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Test un-registered linux interface
	plugin.ResolveDeletedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	_, ok := plugin.hostInterfaces[afPacketHosts[0]]
	Expect(ok).To(BeFalse())
}

// Un-register linux interface with linux plugin not initialized
func TestAfPacketDeleteLinuxInterfaceNoLinux(t *testing.T) {
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	plugin.Linux = nil
	defer afPacketTestTeardown(ctx)
	// Register
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Test un-registered linux interface
	plugin.ResolveDeletedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	_, ok := plugin.hostInterfaces[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
}

// Check if 'IsPending' returns correct output
func TestAfPacketIsPending(t *testing.T) {
	// Setup
	ctx, plugin, _ := afPacketTestSetup(t)
	defer afPacketTestTeardown(ctx)
	// Data
	var oldAddresses []string
	firstData := getTestAfPacketData(afPacketNames[0], append(oldAddresses, netAddresses[0]), afPacketHosts[0])
	secondData := getTestAfPacketData(afPacketNames[1], append(oldAddresses, netAddresses[1]), afPacketHosts[1])
	// Fill af packet cache
	plugin.afPacketByName[firstData.Name] = &AfPacketConfig{
		config:  firstData,
		pending: true,
	}
	plugin.afPacketByName[secondData.Name] = &AfPacketConfig{
		config:  secondData,
		pending: false,
	}
	// Test 'IsPending'
	isPending := plugin.IsPendingAfPacket(firstData)
	Expect(isPending).To(BeTrue())
	isPending = plugin.IsPendingAfPacket(secondData)
	Expect(isPending).To(BeFalse())
}

/* AF_PACKET Test Setup */

func afPacketTestSetup(t *testing.T) (*vppcallmock.TestCtx, *AFPacketConfigurator, ifaceidx.SwIfIndexRW) {
	ctx := vppcallmock.SetupTestCtx(t)
	// Logger
	log := logrus.DefaultLogger()
	log.SetLevel(logging.DebugLevel)

	// Interface indices
	swIfIndices := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(log, "afpacket-configurator-test", "afpacket", nil))

	return ctx, &AFPacketConfigurator{
		Logger:           log,
		SwIfIndexes:      swIfIndices,
		Linux:            1, // Just a flag, cannot be nil
		vppCh:            ctx.MockChannel,
		afPacketByHostIf: make(map[string]*AfPacketConfig),
		afPacketByName:   make(map[string]*AfPacketConfig),
		hostInterfaces:   make(map[string]struct{}),
	}, swIfIndices
}

func afPacketTestTeardown(ctx *vppcallmock.TestCtx) {
	ctx.TeardownTestCtx()
	err := safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

/* AF_PACKET Test Data */

func getTestAfPacketData(ifName string, addresses []string, host string) *interfaces.Interfaces_Interface {
	return &interfaces.Interfaces_Interface{
		Name:        ifName,
		Type:        interfaces.InterfaceType_AF_PACKET_INTERFACE,
		Enabled:     true,
		IpAddresses: addresses,
		Afpacket: &interfaces.Interfaces_Interface_Afpacket{
			HostIfName: host,
		},
	}

}
