// Copyright (c) 2017 Cisco and/or its affiliates.
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

var afPacketName = "af-packet"
var afPacketHosts = []string{"af-packet-host1", "af-packet-host2"}
var netAddresses = []string{"10.0.0.1/24", "192.168.50.1/24"}

func TestAfPacketConfigureHostNotAvail(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data[0])
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeZero())
	Expect(pending).To(BeTrue())
	// Test afpacket-by-name cache
	cached, ok := plugin.afPacketByName[afPacketName]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketName))
	Expect(cached.pending).To(BeTrue())
	// Test afpacket-by-host cache
	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketName))

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketConfigureHostAvail(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 2,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})

	// Make host available
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}

	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data[0])
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(2))
	Expect(pending).To(BeFalse())
	// Test afpacket-by-name cache
	cached, ok := plugin.afPacketByName[afPacketName]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketName))
	Expect(cached.pending).To(BeFalse())
	// Test afpacket-by-host cache
	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketName))

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketConfigureHostAvailErr(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		Retval:    1,
		SwIfIndex: 2,
	})

	// Make host available
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}

	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data[0])
	Expect(err).ToNot(BeNil())
	Expect(swIfIdx).To(BeZero())
	Expect(pending).To(BeTrue())
	// Should be cached anyway
	cached, ok := plugin.afPacketByName[afPacketName]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())

	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketConfigureIncorrectType(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := &interfaces.Interfaces_Interface{
		Name:    afPacketName,
		Type:    interfaces.InterfaceType_SOFTWARE_LOOPBACK,
		Enabled: true,
	}

	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data)
	Expect(err).ToNot(BeNil())
	Expect(swIfIdx).To(BeZero())
	Expect(pending).To(BeFalse())
	// Should not be cached
	_, ok := plugin.afPacketByName[afPacketName]
	Expect(ok).To(BeFalse())

	_, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeFalse())

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketModifyRecreateChangedHost(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(2)

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})

	// Make host available
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}

	// Configure first data
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data[0])
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())

	// Modify
	recreate, err := plugin.ModifyAfPacketInterface(data[1], data[0])
	Expect(err).To(BeNil())
	Expect(recreate).To(BeTrue())

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketModifyRecreatePending(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})

	// Configure first data
	_, pending, err := plugin.ConfigureAfPacketInterface(data[0])
	Expect(err).To(BeNil())
	Expect(pending).To(BeTrue())

	// Modify
	recreate, err := plugin.ModifyAfPacketInterface(data[0], data[0])
	Expect(err).To(BeNil())
	Expect(recreate).To(BeTrue())

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketModifyRecreatenotFound(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	// Modify
	recreate, err := plugin.ModifyAfPacketInterface(data[0], data[0])
	Expect(err).To(BeNil())
	Expect(recreate).To(BeTrue())

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketModifyNoRecreate(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(2)
	data[1].Afpacket.HostIfName = afPacketHosts[0] // Do not change host name

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})

	// Make host available
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}

	// Configure first data
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data[0])
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())

	// Modify (use same data)
	recreate, err := plugin.ModifyAfPacketInterface(data[1], data[0])
	Expect(err).To(BeNil())
	Expect(recreate).To(BeFalse())

	// Check updated data in cache
	cached, ok := plugin.afPacketByName[afPacketName]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.IpAddresses[0]).To(BeEquivalentTo(netAddresses[1]))

	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.IpAddresses[0]).To(BeEquivalentTo(netAddresses[1]))

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketModifyInvalidType(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(2)
	data[1].Type = interfaces.InterfaceType_SOFTWARE_LOOPBACK // Set incorrect type

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})

	// Make host available
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}

	// Configure first data
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data[0])
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())

	// Modify
	_, err = plugin.ModifyAfPacketInterface(data[1], data[0])
	Expect(err).ToNot(BeNil())

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketDelete(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&ap_api.AfPacketDeleteReply{})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})

	// Make host available
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}

	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data[0])
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())
	// Test afpacket-by-name cache
	cached, ok := plugin.afPacketByName[afPacketName]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketName))
	Expect(cached.pending).To(BeFalse())
	// Test afpacket-by-host cache
	cached, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeTrue())
	Expect(cached).ToNot(BeNil())
	Expect(cached.config.Name).To(BeEquivalentTo(afPacketName))

	// Delete
	err = plugin.DeleteAfPacketInterface(data[0], 1)
	Expect(err).To(BeNil())
	_, ok = plugin.afPacketByName[afPacketName]
	Expect(ok).To(BeFalse())
	_, ok = plugin.afPacketByHostIf[afPacketHosts[0]]
	Expect(ok).To(BeFalse())

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketDeleteInvalidType(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(2)
	data[1].Type = interfaces.InterfaceType_SOFTWARE_LOOPBACK // Set incorrect type

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})

	// Make host available
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}

	// Configure first data
	swIfIdx, pending, err := plugin.ConfigureAfPacketInterface(data[0])
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	Expect(pending).To(BeFalse())

	// Delete
	err = plugin.DeleteAfPacketInterface(data[1], 1)
	Expect(err).ToNot(BeNil())

	err = safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketNewLinuxInterfaceHostFound(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	// Put afpacket to cache
	plugin.afPacketByHostIf[data[0].Afpacket.HostIfName] = &AfPacketConfig{
		config:  data[0],
		pending: true,
	}

	_, ok := plugin.hostInterfaces[data[0].Afpacket.HostIfName]
	Expect(ok).To(BeFalse())

	config := plugin.ResolveCreatedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	Expect(config).ToNot(BeNil())
	Expect(config.Afpacket.HostIfName).To(BeEquivalentTo(afPacketHosts[0]))
	_, ok = plugin.hostInterfaces[data[0].Afpacket.HostIfName]
	Expect(ok).To(BeTrue())

	err := safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

// Note: this is a case which should NOT happen
func TestAfPacketNewLinuxInterfaceHostFoundPending(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketDeleteReply{})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})

	// Put afpacket to cache
	plugin.afPacketByHostIf[data[0].Afpacket.HostIfName] = &AfPacketConfig{
		config:  data[0],
		pending: false,
	}

	_, ok := plugin.hostInterfaces[data[0].Afpacket.HostIfName]
	Expect(ok).To(BeFalse())

	config := plugin.ResolveCreatedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	Expect(config).ToNot(BeNil())
	Expect(config.Afpacket.HostIfName).To(BeEquivalentTo(afPacketHosts[0]))
	_, ok = plugin.hostInterfaces[data[0].Afpacket.HostIfName]
	Expect(ok).To(BeTrue())

	err := safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketNewLinuxInterfaceHostNotFound(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	_, ok := plugin.hostInterfaces[data[0].Afpacket.HostIfName]
	Expect(ok).To(BeFalse())

	config := plugin.ResolveCreatedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	Expect(config).To(BeNil())
	_, ok = plugin.hostInterfaces[data[0].Afpacket.HostIfName]
	Expect(ok).To(BeTrue())

	err := safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketNewLinuxInterfaceNoLinux(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())
	plugin.Linux = nil

	config := plugin.ResolveCreatedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	Expect(config).To(BeNil())

	err := safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketDeletedLinuxInterfaceHostFound(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(1)

	// Replies
	ctx.MockVpp.MockReply(&ap_api.AfPacketDeleteReply{})
	ctx.MockVpp.MockReply(&if_api.SwInterfaceTagAddDelReply{})

	// Put afpacket to caches
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}
	plugin.afPacketByName[afPacketName] = &AfPacketConfig{
		config:  data[0],
		pending: false,
	}
	plugin.afPacketByHostIf[data[0].Afpacket.HostIfName] = &AfPacketConfig{
		config:  data[0],
		pending: true,
	}

	plugin.ResolveDeletedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	// Host cache should be empty
	_, ok := plugin.hostInterfaces[afPacketHosts[0]]
	Expect(ok).To(BeFalse())
	_, ok = plugin.afPacketByName[data[0].Name]
	Expect(ok).To(BeTrue())
	_, ok = plugin.afPacketByHostIf[data[0].Afpacket.HostIfName]
	Expect(ok).To(BeTrue())

	err := safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketDeletedLinuxInterfaceHostNotFound(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	// Put afpacket to caches
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}

	plugin.ResolveDeletedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	// Host cache should be empty
	_, ok := plugin.hostInterfaces[afPacketHosts[0]]
	Expect(ok).To(BeFalse())

	err := safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketDeleteLinuxInterfaceNoLinux(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())
	plugin.Linux = nil

	// Put afpacket to caches
	plugin.hostInterfaces[afPacketHosts[0]] = struct{}{}

	plugin.ResolveDeletedLinuxInterface(afPacketHosts[0], afPacketHosts[0], 1)
	_, ok := plugin.hostInterfaces[afPacketHosts[0]]
	Expect(ok).To(BeTrue())

	err := safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

func TestAfPacketIsPending(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getAfPacketConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	data := afPacketData(2)
	plugin.afPacketByName[data[0].Name] = &AfPacketConfig{
		config:  data[0],
		pending: true,
	}
	data[1].Name = afPacketName + "2"
	plugin.afPacketByName[data[1].Name] = &AfPacketConfig{
		config:  data[1],
		pending: false,
	}

	isPending := plugin.IsPendingAfPacket(data[0])
	Expect(isPending).To(BeTrue())
	isPending = plugin.IsPendingAfPacket(data[1])
	Expect(isPending).To(BeFalse())

	err := safeclose.Close(ctx)
	Expect(err).To(BeNil())
}

// Auxiliary

func getAfPacketConfigurator(ctx *vppcallmock.TestCtx) (*AFPacketConfigurator, ifaceidx.SwIfIndexRW) {
	// Logger
	log := logrus.DefaultLogger()
	log.SetLevel(logging.DebugLevel)

	// Interface indices
	swIfIndices := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(log, "nat-configurator-test", "nat", nil))

	return &AFPacketConfigurator{
		Logger:           log,
		SwIfIndexes:      swIfIndices,
		Linux:            1, // Just a flag, cannot be nil
		vppCh:            ctx.MockChannel,
		afPacketByHostIf: make(map[string]*AfPacketConfig),
		afPacketByName:   make(map[string]*AfPacketConfig),
		hostInterfaces:   make(map[string]struct{}),
	}, swIfIndices
}

func afPacketData(num int) (ifaces []*interfaces.Interfaces_Interface) {
	for i := 0; i < num; i++ {
		var ipAddress []string
		ifaces = append(ifaces, &interfaces.Interfaces_Interface{
			Name:        afPacketName, // Keep the same name
			Type:        interfaces.InterfaceType_AF_PACKET_INTERFACE,
			Enabled:     true,
			IpAddresses: append(ipAddress, netAddresses[i]),
			Afpacket: &interfaces.Interfaces_Interface_Afpacket{
				HostIfName: afPacketHosts[i],
			},
		})
	}

	return
}
