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

package ifplugin_test

import (
	"testing"

	"git.fd.io/govpp.git/adapter/mock"
	govppapi "git.fd.io/govpp.git/api"
	govpp "git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	bfdApi "github.com/ligato/vpp-agent/plugins/vpp/binapi/bfd"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/memif"
	natApi "github.com/ligato/vpp-agent/plugins/vpp/binapi/nat"
	stnApi "github.com/ligato/vpp-agent/plugins/vpp/binapi/stn"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/tapv2"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vxlan"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/bfd"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/model/nat"
	"github.com/ligato/vpp-agent/plugins/vpp/model/stn"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

// TODO: use configurator initializers from other files which do the same thing

func interfaceConfiguratorTestInitialization(t *testing.T) (*vppcallmock.TestCtx, *ifplugin.InterfaceConfigurator, *govpp.Connection) {
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: mock.NewVppAdapter(),
	}

	conn, err := govpp.Connect(ctx.MockVpp)
	Expect(err).To(BeNil())

	// Test init
	plugin := &ifplugin.InterfaceConfigurator{}

	ifVppNotifCh := make(chan govppapi.Message, 100)
	plugLog := logging.ForPlugin("tests")

	err = plugin.Init(plugLog, conn, nil, ifVppNotifCh, 0, true)
	Expect(err).To(BeNil())

	return ctx, plugin, conn
}

func interfaceConfiguratorTestTeardown(plugin *ifplugin.InterfaceConfigurator, conn *govpp.Connection) {
	conn.Disconnect()
	Expect(plugin.Close()).To(BeNil())
	logging.DefaultRegistry.ClearRegistry()
}

func bfdConfiguratorTestInitialization(t *testing.T) (*vppcallmock.TestCtx, *ifplugin.BFDConfigurator, *govpp.Connection, ifaceidx.SwIfIndexRW) {
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: mock.NewVppAdapter(),
	}

	c, err := govpp.Connect(ctx.MockVpp)
	Expect(err).To(BeNil())

	// initialize index
	nameToIdx := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "sw_if_index_test", ifaceidx.IndexMetadata)
	index := ifaceidx.NewSwIfIndex(nameToIdx)
	names := nameToIdx.ListNames()
	Expect(names).To(BeEmpty())

	// Test init
	plugin := &ifplugin.BFDConfigurator{}
	err = plugin.Init(logging.ForPlugin("test-log"), c, index, true)
	Expect(err).To(BeNil())

	return ctx, plugin, c, index
}

func bfdConfiguratorTestTeardown(plugin *ifplugin.BFDConfigurator, conn *govpp.Connection) {
	conn.Disconnect()
	Expect(plugin.Close()).To(BeNil())
	logging.DefaultRegistry.ClearRegistry()
}

func stnConfiguratorTestInitialization(t *testing.T) (*vppcallmock.TestCtx, *ifplugin.StnConfigurator, *govpp.Connection) {
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: mock.NewVppAdapter(),
	}
	c, err := govpp.Connect(ctx.MockVpp)
	Expect(err).To(BeNil())

	// initialize index
	nameToIdx := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "sw_if_index_test", ifaceidx.IndexMetadata)
	index := ifaceidx.NewSwIfIndex(nameToIdx)
	names := nameToIdx.ListNames()
	Expect(names).To(BeEmpty())

	// Test init
	plugin := &ifplugin.StnConfigurator{}
	err = plugin.Init(logging.ForPlugin("test-log"), c, index, true)
	Expect(err).To(BeNil())

	return ctx, plugin, c
}

func stnConfiguratorTestTeardown(plugin *ifplugin.StnConfigurator, conn *govpp.Connection) {
	conn.Disconnect()
	Expect(plugin.Close()).To(BeNil())
	logging.DefaultRegistry.ClearRegistry()
}

func natConfiguratorTestInitialization(t *testing.T) (*vppcallmock.TestCtx, *ifplugin.NatConfigurator, *govpp.Connection, ifaceidx.SwIfIndexRW) {
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: mock.NewVppAdapter(),
	}
	c, err := govpp.Connect(ctx.MockVpp)
	Expect(err).To(BeNil())

	// initialize index
	nameToIdx := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "sw_if_index_test", ifaceidx.IndexMetadata)
	index := ifaceidx.NewSwIfIndex(nameToIdx)
	names := nameToIdx.ListNames()
	Expect(names).To(BeEmpty())

	// Test init
	plugin := &ifplugin.NatConfigurator{}
	err = plugin.Init(logging.ForPlugin("test-log"), c, index, true)
	Expect(err).To(BeNil())

	return ctx, plugin, c, index
}

func natConfiguratorTestTeardown(plugin *ifplugin.NatConfigurator, conn *govpp.Connection) {
	conn.Disconnect()
	Expect(plugin.Close()).To(BeNil())
	logging.DefaultRegistry.ClearRegistry()
}

// Tests InterfaceConfigurator resync
func TestDataResyncResync(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: []byte("memif0"),
				Tag:           []byte("test2"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
				SwIfIndex:     1,
			},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("testsocket"),
			},
		},
	})

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_MEMORY_INTERFACE,
			Enabled:     true,
			IpAddresses: []string{"192.168.0.1/24"},
			Memif: &intf.Interfaces_Interface_Memif{
				Id:             1,
				SocketFilename: "testsocket",
				Master:         true,
			},
		},
	}

	err := plugin.Resync(intfaces)
	Expect(err).To(BeNil())
	Expect(plugin.IsSocketFilenameCached("testsocket")).To(BeTrue())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_MEMORY_INTERFACE))
}

// Tests InterfaceConfigurator resync with SwIfIndex 0
func TestDataResyncResyncIdx0(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: []byte("memif0"),
				Tag:           []byte("test2"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
				SwIfIndex:     0,
			},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	})

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_MEMORY_INTERFACE,
			Enabled:     true,
			IpAddresses: []string{"192.168.0.1/24"},
			Memif: &intf.Interfaces_Interface_Memif{
				SocketFilename: "test",
				Master:         true,
				Id:             0,
			},
		},
	}

	err := plugin.Resync(intfaces)
	Expect(err).To(BeNil())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_MEMORY_INTERFACE))
}

// Tests InterfaceConfigurator resync with same interface name/tag
func TestDataResyncResyncSameName(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: []byte("host0"),
				Tag:           []byte("test"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
				SwIfIndex:     1,
			},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	})

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_AF_PACKET_INTERFACE,
			Enabled:     true,
			IpAddresses: []string{"192.168.0.1/24"},
			Afpacket: &intf.Interfaces_Interface_Afpacket{
				HostIfName: "host0",
			},
		},
	}

	err := plugin.Resync(intfaces)
	Expect(err).To(BeNil())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_AF_PACKET_INTERFACE))
}

// Tests InterfaceConfigurator resync with unnamed interface
func TestDataResyncResyncUnnamed(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: []byte("memif0"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
				SwIfIndex:     1,
			},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	})

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_MEMORY_INTERFACE,
			Enabled:     true,
			IpAddresses: []string{"192.168.0.1/24"},
			Memif: &intf.Interfaces_Interface_Memif{
				SocketFilename: "test",
				Master:         true,
				Id:             1,
			},
		},
	}

	err := plugin.Resync(intfaces)
	Expect(err).To(BeNil())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_MEMORY_INTERFACE))
}

// Tests InterfaceConfigurator resync with unnumbered VXLAN interface
func TestDataResyncResyncUnnumbered(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				Tag:           []byte("test"),
				InterfaceName: []byte("vxlan0"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
				SwIfIndex:     1,
			},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
		{
			Name: (&vxlan.VxlanTunnelDump{}).GetMessageName(),
			Ping: true,
			Message: &vxlan.VxlanTunnelDetails{
				SwIfIndex:  1,
				Vni:        12,
				SrcAddress: []byte("192.168.0.1"),
				DstAddress: []byte("192.168.10.1"),
			},
		},
	})

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_VXLAN_TUNNEL,
			Enabled:     true,
			Unnumbered: &intf.Interfaces_Interface_Unnumbered{
				IsUnnumbered:    true,
				InterfaceWithIp: "test",
			},
			Vxlan: &intf.Interfaces_Interface_Vxlan{
				SrcAddress: "192.168.0.1",
				DstAddress: "192.168.10.1",
				Vni:        12,
			},
		},
	}

	err := plugin.Resync(intfaces)
	Expect(err).To(BeNil())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_VXLAN_TUNNEL))
}

// Tests InterfaceConfigurator resync with unnumbered tap interface
func TestDataResyncResyncUnnumberedTap(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				Tag:           []byte("test"),
				InterfaceName: []byte("tap0"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
				SwIfIndex:     1,
			},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
		{
			Name: (&tapv2.SwInterfaceTapV2Dump{}).GetMessageName(),
			Ping: true,
			Message: &tapv2.SwInterfaceTapV2Details{
				SwIfIndex:  1,
				HostIfName: []byte("if0"),
				TxRingSz:   20,
			},
		},
	})

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_TAP_INTERFACE,
			Enabled:     true,
			IpAddresses: []string{"192.168.0.1/24"},
			Unnumbered: &intf.Interfaces_Interface_Unnumbered{
				IsUnnumbered:    true,
				InterfaceWithIp: "test",
			},
			Tap: &intf.Interfaces_Interface_Tap{
				Version:    2,
				HostIfName: "if0",
				TxRingSize: 15,
			},
		},
	}

	err := plugin.Resync(intfaces)
	Expect(err).To(BeNil())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_TAP_INTERFACE))
}

// Tests InterfaceConfigurator resync with unnumbered AF_PACKET interface
func TestDataResyncResyncUnnumberedAfPacket(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				Tag:           []byte("test"),
				InterfaceName: []byte("host-test"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
				SwIfIndex:     1,
			},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	})

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_AF_PACKET_INTERFACE,
			Enabled:     true,
			IpAddresses: []string{"192.168.0.1/24"},
			Unnumbered: &intf.Interfaces_Interface_Unnumbered{
				IsUnnumbered:    true,
				InterfaceWithIp: "test",
			},
			Afpacket: &intf.Interfaces_Interface_Afpacket{
				HostIfName: "host-test",
			},
		},
	}

	err := plugin.Resync(intfaces)
	Expect(err).To(BeNil())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_AF_PACKET_INTERFACE))
}

// Tests InterfaceConfigurator resync with unnumbered MEMIF interface
func TestDataResyncResyncUnnumberedMemif(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				Tag:           []byte("test"),
				InterfaceName: []byte("memif0"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
				SwIfIndex:     1,
			},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
		{
			Name: (&memif.MemifDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifDetails{
				SwIfIndex: 1,
				ID:        1,
				SocketID:  1,
				Role:      1,
				Mode:      1,
			},
		},
	})

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_MEMORY_INTERFACE,
			Enabled:     true,
			IpAddresses: []string{"192.168.0.1/24"},
			Unnumbered: &intf.Interfaces_Interface_Unnumbered{
				IsUnnumbered:    true,
				InterfaceWithIp: "test",
			},
			Memif: &intf.Interfaces_Interface_Memif{
				SocketFilename: "test",
				Master:         false,
				Id:             1,
				Mode:           intf.Interfaces_Interface_Memif_IP,
			},
		},
	}

	err := plugin.Resync(intfaces)
	Expect(err).To(BeNil())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_MEMORY_INTERFACE))
}

// Tests if InterfaceConfigurator VPP config is present
func TestDataResyncVerifyVPPConfigPresence(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: []byte("memif1"),
				Tag:           []byte("test"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
			},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	})

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_AF_PACKET_INTERFACE,
			Enabled:     true,
			IpAddresses: []string{"192.168.0.1/24"},
			Afpacket: &intf.Interfaces_Interface_Afpacket{
				HostIfName: "host0",
			},
		},
	}

	ok := plugin.VerifyVPPConfigPresence(intfaces)
	Expect(ok).To(BeTrue())
}

// Tests if InterfaceConfigurator VPP config is not present
func TestDataResyncVerifyVPPConfigPresenceNegative(t *testing.T) {
	ctx, plugin, conn := interfaceConfiguratorTestInitialization(t)
	defer interfaceConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Messages: []govppapi.Message{
				&interfaces.SwInterfaceDetails{
					SwIfIndex:     0,
					InterfaceName: []byte("local0"),
					AdminUpDown:   1,
					LinkMtu:       9216, // Default MTU
				},
				&interfaces.SwInterfaceDetails{
					SwIfIndex:     1,
					InterfaceName: []byte("testif0"),
					AdminUpDown:   0,
				},
			},
		}, {
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	})

	// Test
	ok := plugin.VerifyVPPConfigPresence([]*intf.Interfaces_Interface{})
	Expect(ok).To(BeFalse())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeFalse())
	Expect(meta).To(BeNil())
}

// Tests BFDConfigurator session resync
func TestDataResyncResyncSession(t *testing.T) {
	ctx, plugin, conn, index := bfdConfiguratorTestInitialization(t)
	defer bfdConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name:    (&bfdApi.BfdUDPAdd{}).GetMessageName(),
			Message: &bfdApi.BfdUDPAddReply{},
		},
		{
			Name:    (&bfdApi.BfdUDPMod{}).GetMessageName(),
			Message: &bfdApi.BfdUDPModReply{},
		},
		{
			Name:    (&bfdApi.BfdUDPDel{}).GetMessageName(),
			Message: &bfdApi.BfdUDPDelReply{},
		},
		{
			Name:    (&bfdApi.BfdUDPSessionDump{}).GetMessageName(),
			Ping:    true,
			Message: &bfdApi.BfdUDPSessionDetails{},
		},
	})

	index.RegisterName("if0", 0, &intf.Interfaces_Interface{
		Name:        "if0",
		IpAddresses: []string{"192.168.1.10", "192.168.2.10"},
	})

	// Test
	sessions := []*bfd.SingleHopBFD_Session{
		{
			Interface:          "if0",
			SourceAddress:      "192.168.1.10",
			DestinationAddress: "192.168.2.10",
		},
	}

	err := plugin.ResyncSession(sessions)
	Expect(err).To(BeNil())

	_, _, found := plugin.GetBfdSessionIndexes().LookupIdx("if0")
	Expect(found).To(BeTrue())
}

// Tests BFDConfigurator session resync
func TestDataResyncResyncSessionSameData(t *testing.T) {
	ctx, plugin, conn, index := bfdConfiguratorTestInitialization(t)
	defer bfdConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name:    (&bfdApi.BfdUDPAdd{}).GetMessageName(),
			Message: &bfdApi.BfdUDPAddReply{},
		},
		{
			Name:    (&bfdApi.BfdUDPMod{}).GetMessageName(),
			Message: &bfdApi.BfdUDPModReply{},
		},
		{
			Name:    (&bfdApi.BfdUDPDel{}).GetMessageName(),
			Message: &bfdApi.BfdUDPDelReply{},
		},
		{
			Name: (&bfdApi.BfdUDPSessionDump{}).GetMessageName(),
			Ping: true,
			Message: &bfdApi.BfdUDPSessionDetails{
				IsAuthenticated: 1,
				LocalAddr:       []byte{192, 168, 1, 10},
				PeerAddr:        []byte{192, 168, 2, 10},
				SwIfIndex:       1,
			},
		},
	})

	index.RegisterName("if0", 1, &intf.Interfaces_Interface{
		Name:        "if0",
		IpAddresses: []string{"192.168.1.10", "192.168.2.10"},
	})

	// Test
	sessions := []*bfd.SingleHopBFD_Session{
		{
			Interface:          "if0",
			SourceAddress:      "192.168.1.10",
			DestinationAddress: "192.168.2.10",
		},
	}

	err := plugin.ResyncSession(sessions)
	Expect(err).To(BeNil())

	_, _, found := plugin.GetBfdSessionIndexes().LookupIdx("if0")
	Expect(found).To(BeTrue())
}

// Tests BFDConfigurator authorization key resync
func TestDataResyncResyncAuthKey(t *testing.T) {
	ctx, plugin, conn, swIfIdx := bfdConfiguratorTestInitialization(t)
	defer bfdConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name:    (&bfdApi.BfdAuthKeysDump{}).GetMessageName(),
			Ping:    true,
			Message: &bfdApi.BfdAuthKeysDetails{},
		},
		{
			Name:    (&bfdApi.BfdUDPSessionDump{}).GetMessageName(),
			Ping:    true,
			Message: &bfdApi.BfdUDPSessionDetails{},
		},
		{
			Name:    (&bfdApi.BfdAuthDelKey{}).GetMessageName(),
			Message: &bfdApi.BfdAuthDelKeyReply{},
		},
		{
			Name:    (&bfdApi.BfdAuthSetKey{}).GetMessageName(),
			Message: &bfdApi.BfdAuthSetKeyReply{},
		},
	})
	// Register
	swIfIdx.RegisterName("if1", 0, nil)
	// Test
	authKey := []*bfd.SingleHopBFD_Key{
		{
			Name: "test",
		},
	}

	err := plugin.ResyncAuthKey(authKey)
	Expect(err).To(BeNil())

	_, _, found := plugin.GetBfdKeyIndexes().LookupIdx(ifplugin.AuthKeyIdentifier(0))
	Expect(found).To(BeTrue())
}

// Tests BFDConfigurator authorization key resync
func TestDataResyncResyncAuthKeyNoMatch(t *testing.T) {
	ctx, plugin, conn, _ := bfdConfiguratorTestInitialization(t)
	defer bfdConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name: (&bfdApi.BfdAuthKeysDump{}).GetMessageName(),
			Ping: true,
			Message: &bfdApi.BfdAuthKeysDetails{
				ConfKeyID: 2,
			},
		},
		{
			Name:    (&bfdApi.BfdUDPSessionDump{}).GetMessageName(),
			Ping:    true,
			Message: &bfdApi.BfdUDPSessionDetails{},
		},
		{
			Name:    (&bfdApi.BfdAuthDelKey{}).GetMessageName(),
			Message: &bfdApi.BfdAuthDelKeyReply{},
		},
		{
			Name:    (&bfdApi.BfdAuthSetKey{}).GetMessageName(),
			Message: &bfdApi.BfdAuthSetKeyReply{},
		},
	})

	// Test
	authKey := []*bfd.SingleHopBFD_Key{
		{
			Name: "test",
			Id:   1,
		},
	}

	err := plugin.ResyncAuthKey(authKey)
	Expect(err).To(BeNil())

	_, _, found := plugin.GetBfdKeyIndexes().LookupIdx(ifplugin.AuthKeyIdentifier(1))
	Expect(found).To(BeTrue())
}

// Tests BFDConfigurator echo resync
func TestDataResyncResyncEchoFunction(t *testing.T) {
	ctx, plugin, conn, index := bfdConfiguratorTestInitialization(t)
	defer bfdConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name:    (&bfdApi.BfdUDPSetEchoSource{}).GetMessageName(),
			Message: &bfdApi.BfdUDPSetEchoSourceReply{},
		},
	})

	index.RegisterName("if0", 0, &intf.Interfaces_Interface{
		Name:        "if0",
		IpAddresses: []string{"192.168.1.10/24"},
	})

	// Test
	echoFunctions := []*bfd.SingleHopBFD_EchoFunction{
		{
			Name:                "test",
			EchoSourceInterface: "if0",
		},
	}

	err := plugin.ResyncEchoFunction(echoFunctions)
	Expect(err).To(BeNil())

	_, _, found := plugin.GetBfdEchoFunctionIndexes().LookupIdx("if0")
	Expect(found).To(BeTrue())
}

// Tests StnConfigurator resync
func TestDataResyncResyncStn(t *testing.T) {
	ctx, plugin, conn := stnConfiguratorTestInitialization(t)
	defer stnConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name:    (&stnApi.StnAddDelRule{}).GetMessageName(),
			Message: &stnApi.StnAddDelRuleReply{},
		},
	})

	// Test
	nbStnRules := []*stn.STN_Rule{
		{
			RuleName:  "test",
			Interface: "if0",
			IpAddress: "192.168.0.1",
		},
	}

	ok := plugin.Resync(nbStnRules)
	Expect(ok).To(BeNil())

	Expect(plugin.IndexExistsFor(ifplugin.StnIdentifier("if0"))).To(BeTrue())
}

// Tests NATConfigurator NAT global resync
func TestDataResyncResyncNatGlobal(t *testing.T) {
	ctx, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name:    (&natApi.Nat44ForwardingIsEnabled{}).GetMessageName(),
			Message: &natApi.Nat44ForwardingIsEnabledReply{},
		},
		{
			Name:    (&natApi.Nat44InterfaceDump{}).GetMessageName(),
			Ping:    true,
			Message: &natApi.Nat44InterfaceDetails{},
		},
		{
			Name:    (&natApi.Nat44InterfaceOutputFeatureDump{}).GetMessageName(),
			Ping:    true,
			Message: &natApi.Nat44InterfaceOutputFeatureDetails{},
		},
		{
			Name:    (&natApi.Nat44InterfaceOutputFeatureDump{}).GetMessageName(),
			Ping:    true,
			Message: &natApi.Nat44InterfaceOutputFeatureDetails{},
		},
		{
			Name:    (&natApi.Nat44AddressDump{}).GetMessageName(),
			Ping:    true,
			Message: &natApi.Nat44AddressDetails{},
		},
		{
			Name:    (&natApi.Nat44AddDelAddressRange{}).GetMessageName(),
			Message: &natApi.Nat44AddDelAddressRangeReply{},
		},
	})

	// Test
	nbGlobal := &nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterface{
			{
				Name: "test",
			},
		},
	}

	err := plugin.ResyncNatGlobal(nbGlobal)
	Expect(err).To(BeNil())

	Expect(plugin.IsInNotEnabledIfCache("test")).To(BeTrue())
}

// Tests NATConfigurator SNAT resync
func TestDataResyncResyncSNat(t *testing.T) {
	ctx, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies(nil)

	// Test
	sNatConf := []*nat.Nat44SNat_SNatConfig{
		{
			Label: "test",
		},
	}

	// Unfinished, this method does nothing atm
	err := plugin.ResyncSNat(sNatConf)
	Expect(err).To(BeNil())
}

// Tests NATConfigurator DNAT resync
func TestDataResyncResyncDNat(t *testing.T) {
	ctx, plugin, conn, index := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name:    (&natApi.Nat44AddDelStaticMapping{}).GetMessageName(),
			Message: &natApi.Nat44AddDelStaticMappingReply{},
		},
		{
			Name:    (&natApi.Nat44AddDelIdentityMapping{}).GetMessageName(),
			Message: &natApi.Nat44AddDelIdentityMappingReply{},
		},
		{
			Name: (&natApi.Nat44StaticMappingDump{}).GetMessageName(),
			Ping: true,
			Message: &natApi.Nat44StaticMappingDetails{
				Protocol:          6,
				Tag:               []byte("smap|lbstat|idmap"),
				LocalIPAddress:    []byte{192, 168, 10, 0},
				ExternalIPAddress: []byte{192, 168, 0, 1},
				LocalPort:         88,
			},
		},
		{
			Name: (&natApi.Nat44LbStaticMappingDump{}).GetMessageName(),
			Ping: true,
			Message: &natApi.Nat44LbStaticMappingDetails{
				Protocol:     6,
				Tag:          []byte("smap|lbstat|idmap"),
				ExternalAddr: []byte{192, 168, 10, 0},
				ExternalPort: 88,
				Locals: []natApi.Nat44LbAddrPort{
					{
						Addr: []byte{192., 168, 10, 0},
						Port: 89,
					},
					{
						Addr: []byte{192., 168, 20, 0},
						Port: 90,
					},
				},
			},
		},
		{
			Name: (&natApi.Nat44IdentityMappingDump{}).GetMessageName(),
			Ping: true,
			Message: &natApi.Nat44IdentityMappingDetails{
				Protocol:  6,
				Tag:       []byte("smap|lbstat|idmap"),
				IPAddress: []byte{192, 168, 10, 0},
			},
		},
	})

	// Register index
	index.RegisterName("if0", 0, &intf.Interfaces_Interface{
		Name:        "if0",
		IpAddresses: []string{"192.168.0.1", "192.168.10.0"},
	})

	// Test
	dNatConf := []*nat.Nat44DNat_DNatConfig{
		{
			Label: "smap",
			StMappings: []*nat.Nat44DNat_DNatConfig_StaticMapping{
				{
					Protocol:   nat.Protocol_TCP,
					ExternalIp: "192.168.0.1",
					LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
						{
							LocalIp:     "192.168.10.0",
							LocalPort:   88,
							Probability: 12,
						},
					},
				},
			},
			IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMapping{
				{
					Protocol:  nat.Protocol_TCP,
					IpAddress: "192.168.0.1",
				},
			},
		},
		{
			Label: "lbstat",
			StMappings: []*nat.Nat44DNat_DNatConfig_StaticMapping{
				{
					Protocol:   nat.Protocol_TCP,
					ExternalIp: "192.168.0.1",
					LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
						{
							LocalIp:     "192.168.10.0",
							LocalPort:   88,
							Probability: 12,
						},
					},
				},
			},
			IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMapping{
				{
					Protocol:  nat.Protocol_TCP,
					IpAddress: "192.168.0.1",
				},
			},
		},
		{
			Label: "idmap",
			StMappings: []*nat.Nat44DNat_DNatConfig_StaticMapping{
				{
					Protocol:   nat.Protocol_TCP,
					ExternalIp: "192.168.0.1",
					LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
						{
							LocalIp:     "192.168.10.0",
							LocalPort:   88,
							Probability: 12,
						},
					},
				},
			},
			IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMapping{
				{
					Protocol:  nat.Protocol_TCP,
					IpAddress: "192.168.0.1",
				},
			},
		},
	}

	err := plugin.ResyncDNat(dNatConf)
	Expect(err).To(BeNil())

	stIdent := ifplugin.GetStMappingIdentifier(&nat.Nat44DNat_DNatConfig_StaticMapping{
		ExternalIp: "192.168.0.1",
		Protocol:   nat.Protocol_TCP,
		LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
			{
				LocalIp:     "192.168.10.0",
				LocalPort:   88,
				Probability: 12,
			},
		},
	})

	idIdent := ifplugin.GetIDMappingIdentifier(&nat.Nat44DNat_DNatConfig_IdentityMapping{
		Protocol:  nat.Protocol_TCP,
		IpAddress: "192.168.0.1",
	})

	Expect(plugin.IsDNatLabelIDMappingRegistered(idIdent)).To(BeTrue())
	Expect(plugin.IsDNatLabelStMappingRegistered(stIdent)).To(BeTrue())
}

// Tests NATConfigurator DNAT resync
func TestDataResyncResyncDNatMultipleIPs(t *testing.T) {
	ctx, plugin, conn, index := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	ctx.MockReplies([]*vppcallmock.HandleReplies{
		{
			Name:    (&natApi.Nat44AddDelStaticMapping{}).GetMessageName(),
			Message: &natApi.Nat44AddDelStaticMappingReply{},
		},
		{
			Name:    (&natApi.Nat44AddDelLbStaticMapping{}).GetMessageName(),
			Message: &natApi.Nat44AddDelLbStaticMappingReply{},
		},
		{
			Name:    (&natApi.Nat44AddDelIdentityMappingReply{}).GetMessageName(),
			Message: &natApi.Nat44AddDelIdentityMappingReply{},
		},
		{
			Name: (&natApi.Nat44StaticMappingDump{}).GetMessageName(),
			Ping: true,
			Message: &natApi.Nat44StaticMappingDetails{
				Protocol: 6,
				Tag:      []byte("smap|lbstat|idmap"),
			},
		},
		{
			Name: (&natApi.Nat44LbStaticMappingDump{}).GetMessageName(),
			Ping: true,
			Message: &natApi.Nat44LbStaticMappingDetails{
				ExternalPort: 25,
				Protocol:     6,
				Tag:          []byte("smap|lbstat|idmap"),
				Locals: []natApi.Nat44LbAddrPort{
					{
						Addr: []byte{192., 168, 10, 0},
						Port: 89,
					},
					{
						Addr: []byte{192., 168, 20, 0},
						Port: 90,
					},
				},
			},
		},
		{
			Name: (&natApi.Nat44IdentityMappingDump{}).GetMessageName(),
			Ping: true,
			Message: &natApi.Nat44IdentityMappingDetails{
				Protocol: 6,
				Tag:      []byte("smap|lbstat|idmap"),
			},
		},
	})

	// Register index
	index.RegisterName("if0", 0, &intf.Interfaces_Interface{
		Name:        "if0",
		IpAddresses: []string{"192.168.0.1", "192.168.10.0", "192.168.12.0", "192.168.14.0"},
	})

	// Test
	dNatConf := []*nat.Nat44DNat_DNatConfig{
		{
			Label: "smap",
			StMappings: []*nat.Nat44DNat_DNatConfig_StaticMapping{
				{
					Protocol:     nat.Protocol_TCP,
					ExternalIp:   "192.168.0.1",
					ExternalPort: 88,
					LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
						{
							LocalIp:     "192.168.10.0",
							LocalPort:   88,
							Probability: 12,
						},
						{
							LocalIp:     "192.168.12.0",
							LocalPort:   88,
							Probability: 13,
						},
						{
							LocalIp:     "192.168.14.0",
							LocalPort:   88,
							Probability: 14,
						},
					},
				},
			},
			IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMapping{
				{
					Protocol:  nat.Protocol_TCP,
					IpAddress: "192.168.0.1",
				},
			},
		},
		{
			Label: "lbstat",
			StMappings: []*nat.Nat44DNat_DNatConfig_StaticMapping{
				{
					Protocol:     nat.Protocol_TCP,
					ExternalIp:   "192.168.0.1",
					ExternalPort: 88,
					LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
						{
							LocalIp:     "192.168.10.0",
							LocalPort:   88,
							Probability: 12,
						},
						{
							LocalIp:     "192.168.12.0",
							LocalPort:   88,
							Probability: 16,
						},
						{
							LocalIp:     "192.168.14.0",
							LocalPort:   88,
							Probability: 17,
						},
					},
				},
			},
			IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMapping{
				{
					Protocol:  nat.Protocol_TCP,
					IpAddress: "192.168.0.1",
				},
			},
		},
		{
			Label: "idmap",
			StMappings: []*nat.Nat44DNat_DNatConfig_StaticMapping{
				{
					ExternalIp:   "192.168.0.1",
					ExternalPort: 88,
					Protocol:     nat.Protocol_TCP,
					LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
						{
							LocalIp:     "192.168.10.0",
							LocalPort:   88,
							Probability: 12,
						},
						{
							LocalIp:     "192.168.12.0",
							LocalPort:   88,
							Probability: 8,
						},
						{
							LocalIp:     "192.168.14.0",
							LocalPort:   88,
							Probability: 2,
						},
					},
				},
			},
			IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMapping{
				{
					Protocol:  nat.Protocol_TCP,
					IpAddress: "192.168.0.1",
				},
			},
		},
	}

	err := plugin.ResyncDNat(dNatConf)
	Expect(err).To(BeNil())

	stIdent := ifplugin.GetStMappingIdentifier(&nat.Nat44DNat_DNatConfig_StaticMapping{
		ExternalIp: "192.168.0.1",
		Protocol:   nat.Protocol_TCP,
		LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
			{
				LocalIp:     "192.168.10.0",
				LocalPort:   88,
				Probability: 12,
			},
			{
				LocalIp:     "192.168.12.0",
				LocalPort:   88,
				Probability: 8,
			},
			{
				LocalIp:     "192.168.14.0",
				LocalPort:   88,
				Probability: 2,
			},
		},
	})

	idIdent := ifplugin.GetIDMappingIdentifier(&nat.Nat44DNat_DNatConfig_IdentityMapping{
		Protocol:  nat.Protocol_TCP,
		IpAddress: "192.168.0.1",
	})

	Expect(plugin.IsDNatLabelIDMappingRegistered(idIdent)).To(BeTrue())
	Expect(plugin.IsDNatLabelStMappingRegistered(stIdent)).To(BeTrue())
}

// Test unexported method resolving NB static mapping equal to the VPP static mapping. Mapping
// is expected to be registered
func TestResolveStaticMapping(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingData()
	vppData := getNat44StaticMappingData().StMappings

	// Test where NB == VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeTrue())
}

// Test unexported method resolving NB static mapping with different local IP address as the VPP static mapping. Mapping
// is not expected to be registered
func TestResolveStaticMappingNoMatch1(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingData()
	vppData := getNat44StaticMappingData().StMappings
	vppData[0].LocalIps[0].LocalIp = "" // Change localIP

	// Tests where NB != VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeFalse())
}

// Test unexported method resolving NB static mapping with different external IP address as the VPP static mapping.
// Mapping  is not expected to be registered
func TestResolveStaticMappingNoMatch2(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingData()
	vppData := getNat44StaticMappingData().StMappings
	vppData[0].ExternalIp = "" // Change external IP

	// Tests where NB != VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeFalse())
}

// Test unexported method resolving NB static mapping with different VRF as the VPP static mapping. Mapping
// is not expected to be registered
func TestResolveStaticMappingNoMatch3(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingData()
	vppData := getNat44StaticMappingData().StMappings
	vppData[0].LocalIps[0].VrfId = 1 // Change VRF

	// Tests where NB != VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeFalse())
}

// Test unexported method resolving NB static mapping with different count of local IP addresses as the VPP static
// mapping. Mapping is not expected to be registered
func TestResolveStaticMappingNoMatch4(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingData()
	vppData := getNat44StaticMappingData().StMappings
	vppData[0].LocalIps = append(vppData[0].LocalIps, getLocalIP("10.0.0.2", 30, 15, 0)) // Change number of Local IPs

	// Tests where NB != VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeFalse())
}

// Test unexported method resolving NB load-balanced static mapping equal to the VPP load-balanced static mapping.
// Mapping is expected to be registered
func TestResolveStaticMappingLb(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingLbData()
	vppData := getNat44StaticMappingLbData().StMappings

	// Test where NB == VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeTrue())
}

// Test unexported method resolving NB load-balanced static mapping with different local IP in one of the entries.
// Mapping is expected to not be registered
func TestResolveStaticMappingLbNoMatch1(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingLbData()
	vppData := getNat44StaticMappingLbData().StMappings
	vppData[0].LocalIps[1].LocalIp = "" // Change localIP in second entry

	// Tests where NB != VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeFalse())
}

// Test unexported method resolving NB load-balanced static mapping with different external IP.
// Mapping is expected to not be registered
func TestResolveStaticMappingLbNoMatch2(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingLbData()
	vppData := getNat44StaticMappingLbData().StMappings
	vppData[0].ExternalIp = "" // Change external IP

	// Tests where NB != VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeFalse())
}

// Test unexported method resolving NB load-balanced static mapping with different VRF.
// Mapping is expected to not be registered
func TestResolveStaticMappingLbNoMatch3(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingLbData()
	vppData := getNat44StaticMappingLbData().StMappings
	vppData[0].LocalIps[1].VrfId = 1 // Change VRF

	// Tests where NB != VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeFalse())
}

// Test unexported method resolving NB load-balanced static mapping with different count of local IP entries.
// Mapping is expected to not be registered
func TestResolveStaticMappingLbNoMatch4(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := getNat44StaticMappingLbData()
	vppData := getNat44StaticMappingLbData().StMappings
	vppData[0].LocalIps = append(vppData[0].LocalIps, getLocalIP("10.0.0.3", 35, 20, 0)) // Change number of Local IPs

	// Tests where NB != VPP
	ifplugin.ResolveMappings(plugin, nbData, &vppData, &idMappings)
	Expect(plugin.IsDNatLabelStMappingRegistered(ifplugin.GetStMappingIdentifier(nbData.StMappings[0]))).To(BeFalse())
}

// Test unexported method resolving NB identity mapping equal to the VPP identity mapping.
// Mapping is expected to be registered.
func TestResolveIdentityMapping(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var stMappings []*nat.Nat44DNat_DNatConfig_StaticMapping

	nbData := getNat44IdentityMappingData()
	vppData := getNat44IdentityMappingData().IdMappings

	// Test where NB == VPP
	ifplugin.ResolveMappings(plugin, nbData, &stMappings, &vppData)
	Expect(plugin.IsDNatLabelIDMappingRegistered(ifplugin.GetIDMappingIdentifier(nbData.IdMappings[0]))).To(BeTrue())
}

// Test unexported method resolving NB identity mapping with different IP address.
// Mapping is expected to not be registered.
func TestResolveIdentityMappingNoMatch1(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var stMappings []*nat.Nat44DNat_DNatConfig_StaticMapping

	nbData := getNat44IdentityMappingData()
	vppData := getNat44IdentityMappingData().IdMappings
	vppData[0].IpAddress = "" // Ip address change

	// Test where NB == VPP
	ifplugin.ResolveMappings(plugin, nbData, &stMappings, &vppData)
	Expect(plugin.IsDNatLabelIDMappingRegistered(ifplugin.GetIDMappingIdentifier(nbData.IdMappings[0]))).To(BeFalse())
}

// Test unexported method resolving NB identity mapping with different VRF.
// Mapping is expected to not be registered.
func TestResolveIdentityMappingNoMatch2(t *testing.T) {
	_, plugin, conn, _ := natConfiguratorTestInitialization(t)
	defer natConfiguratorTestTeardown(plugin, conn)

	var stMappings []*nat.Nat44DNat_DNatConfig_StaticMapping

	nbData := getNat44IdentityMappingData()
	vppData := getNat44IdentityMappingData().IdMappings
	vppData[0].VrfId = 1 // VRF change

	// Test where NB == VPP
	ifplugin.ResolveMappings(plugin, nbData, &stMappings, &vppData)
	Expect(plugin.IsDNatLabelIDMappingRegistered(ifplugin.GetIDMappingIdentifier(nbData.IdMappings[0]))).To(BeFalse())
}

func getNat44StaticMappingData() *nat.Nat44DNat_DNatConfig {
	var stMappings []*nat.Nat44DNat_DNatConfig_StaticMapping
	var localIPs []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP

	nbData := &nat.Nat44DNat_DNatConfig{
		Label:      "test-dnat",
		StMappings: append(stMappings, getStaticMapping("10.0.0.1", 25, 6)),
	}
	nbData.StMappings[0].LocalIps = append(localIPs, getLocalIP("192.168.0.1", 9000, 35, 0))
	return nbData
}

func getNat44StaticMappingLbData() *nat.Nat44DNat_DNatConfig {
	var stMappings []*nat.Nat44DNat_DNatConfig_StaticMapping
	var localIPs []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP

	nbData := &nat.Nat44DNat_DNatConfig{
		Label:      "test-dnat",
		StMappings: append(stMappings, getStaticMapping("10.0.0.1", 25, 6)),
	}
	nbData.StMappings[0].LocalIps = append(localIPs, getLocalIP("192.168.0.1", 9000, 35, 0),
		getLocalIP("192.168.0.2", 9001, 40, 0))
	return nbData
}

func getNat44IdentityMappingData() *nat.Nat44DNat_DNatConfig {
	var idMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping

	nbData := &nat.Nat44DNat_DNatConfig{
		Label:      "test-dnat",
		IdMappings: append(idMappings, getIdentityMapping("10.0.0.1", 25, 0, 6)),
	}
	return nbData
}

func getStaticMapping(ip string, port uint32, proto nat.Protocol) *nat.Nat44DNat_DNatConfig_StaticMapping {
	return &nat.Nat44DNat_DNatConfig_StaticMapping{
		ExternalIp:   ip,
		ExternalPort: port,
		Protocol:     proto,
	}
}

func getIdentityMapping(ip string, port, vrf uint32, proto nat.Protocol) *nat.Nat44DNat_DNatConfig_IdentityMapping {
	return &nat.Nat44DNat_DNatConfig_IdentityMapping{
		VrfId:     vrf,
		IpAddress: ip,
		Port:      port,
		Protocol:  proto,
	}
}

func getLocalIP(ip string, port, probability uint32, vrf uint32) *nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP {
	return &nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
		VrfId:       vrf,
		LocalIp:     ip,
		LocalPort:   port,
		Probability: probability,
	}
}
