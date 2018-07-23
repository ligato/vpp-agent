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
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/af_packet"
	bfdApi "github.com/ligato/vpp-agent/plugins/vpp/binapi/bfd"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/memif"
	natApi "github.com/ligato/vpp-agent/plugins/vpp/binapi/nat"
	stnApi "github.com/ligato/vpp-agent/plugins/vpp/binapi/stn"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/tap"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/tapv2"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
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

type vppReplyMock struct {
	Name     string
	Ping     bool
	Message  govppapi.Message
	Messages []govppapi.Message
}

func vppMockHandler(vppMock *mock.VppAdapter, dataList []*vppReplyMock) mock.ReplyHandler {
	var sendControlPing bool

	vppMock.RegisterBinAPITypes(af_packet.Types)
	vppMock.RegisterBinAPITypes(bfdApi.Types)
	vppMock.RegisterBinAPITypes(natApi.Types)
	vppMock.RegisterBinAPITypes(stnApi.Types)
	vppMock.RegisterBinAPITypes(interfaces.Types)
	vppMock.RegisterBinAPITypes(ip.Types)
	vppMock.RegisterBinAPITypes(memif.Types)
	vppMock.RegisterBinAPITypes(tap.Types)
	vppMock.RegisterBinAPITypes(tapv2.Types)
	vppMock.RegisterBinAPITypes(vpe.Types)
	vppMock.RegisterBinAPITypes(vxlan.Types)

	return func(request mock.MessageDTO) (reply []byte, msgID uint16, prepared bool) {
		// Following types are not automatically stored in mock adapter's map and will be sent with empty MsgName
		// TODO: initialize mock adapter's map with these
		switch request.MsgID {
		case 100:
			request.MsgName = "control_ping"
		case 101:
			request.MsgName = "control_ping_reply"
		case 200:
			request.MsgName = "sw_interface_dump"
		case 201:
			request.MsgName = "sw_interface_details"
		}

		if request.MsgName == "" {
			logrus.DefaultLogger().Fatalf("mockHandler received request (ID: %v) with empty MsgName, check if compatbility check is done before using this request", request.MsgID)
		}

		if sendControlPing {
			sendControlPing = false
			data := &vpe.ControlPingReply{}
			reply, err := vppMock.ReplyBytes(request, data)
			Expect(err).To(BeNil())
			msgID, err := vppMock.GetMsgID(data.GetMessageName(), data.GetCrcString())
			Expect(err).To(BeNil())
			return reply, msgID, true
		}

		for _, dataMock := range dataList {
			if request.MsgName == dataMock.Name {
				// Send control ping next iteration if set
				sendControlPing = dataMock.Ping
				if len(dataMock.Messages) > 0 {
					logrus.DefaultLogger().Infof(" MOCK HANDLER: mocking %d messages", len(dataMock.Messages))
					for _, msg := range dataMock.Messages {
						vppMock.MockReply(msg)
					}
					return nil, 0, false
				}
				msgID, err := vppMock.GetMsgID(dataMock.Message.GetMessageName(), dataMock.Message.GetCrcString())
				Expect(err).To(BeNil())
				reply, err := vppMock.ReplyBytes(request, dataMock.Message)
				Expect(err).To(BeNil())
				return reply, msgID, true
			}
		}

		replyMsg, msgID, ok := vppMock.ReplyFor(request.MsgName)
		if ok {
			reply, err := vppMock.ReplyBytes(request, replyMsg)
			Expect(err).To(BeNil())
			return reply, msgID, true
		} else {
			logrus.DefaultLogger().Warnf("no reply for %v found", request.MsgName)
		}

		return reply, 0, false
	}
}

func interfaceConfiguratorTestInitialization(t *testing.T, mocks []*vppReplyMock) (*ifplugin.InterfaceConfigurator, *govpp.Connection) {
	// Setup
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}

	ctx.MockVpp.MockReplyHandler(vppMockHandler(ctx.MockVpp, mocks))

	conn, err := govpp.Connect(ctx.MockVpp)
	Expect(err).To(BeNil())

	// Test init
	plugin := &ifplugin.InterfaceConfigurator{}

	ifVppNotifCh := make(chan govppapi.Message, 100)
	plugLog := logging.ForPlugin("tests", logrus.NewLogRegistry())

	err = plugin.Init(plugLog, conn, nil, ifVppNotifCh, 0, true)
	Expect(err).To(BeNil())

	return plugin, conn
}

func bfdConfiguratorTestInitialization(t *testing.T, mocks []*vppReplyMock) (*ifplugin.BFDConfigurator, *govpp.Connection, ifaceidx.SwIfIndexRW) {
	// Setup
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}

	ctx.MockVpp.MockReplyHandler(vppMockHandler(ctx.MockVpp, mocks))

	connection, _ := govpp.Connect(ctx.MockVpp)
	plugin := &ifplugin.BFDConfigurator{}

	// initialize index
	nameToIdx := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "sw_if_index_test", ifaceidx.IndexMetadata)
	index := ifaceidx.NewSwIfIndex(nameToIdx)
	names := nameToIdx.ListNames()

	// check if names were empty
	Expect(names).To(BeEmpty())

	// Test init
	err := plugin.Init(
		logging.ForPlugin("test-log",
			logrus.NewLogRegistry()),
		connection,
		index,
		true)

	Expect(err).To(BeNil())

	return plugin, connection, index
}

func stnConfiguratorTestInitialization(t *testing.T, mocks []*vppReplyMock) (*ifplugin.StnConfigurator, *govpp.Connection) {
	// Setup
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}

	ctx.MockVpp.MockReplyHandler(vppMockHandler(ctx.MockVpp, mocks))

	connection, _ := govpp.Connect(ctx.MockVpp)
	plugin := &ifplugin.StnConfigurator{}

	// initialize index
	nameToIdx := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "sw_if_index_test", ifaceidx.IndexMetadata)
	index := ifaceidx.NewSwIfIndex(nameToIdx)
	names := nameToIdx.ListNames()

	// check if names were empty
	Expect(names).To(BeEmpty())

	// Test init
	err := plugin.Init(
		logging.ForPlugin("test-log",
			logrus.NewLogRegistry()),
		connection,
		index,
		true)

	Expect(err).To(BeNil())
	return plugin, connection
}

func natConfiguratorTestInitialization(t *testing.T, mocks []*vppReplyMock) (*ifplugin.NatConfigurator, ifaceidx.SwIfIndexRW, *govpp.Connection) {
	// Setup
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}

	ctx.MockVpp.MockReplyHandler(vppMockHandler(ctx.MockVpp, mocks))

	connection, _ := govpp.Connect(ctx.MockVpp)
	plugin := &ifplugin.NatConfigurator{}

	// initialize index
	nameToIdx := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "sw_if_index_test", ifaceidx.IndexMetadata)
	index := ifaceidx.NewSwIfIndex(nameToIdx)
	names := nameToIdx.ListNames()

	// check if names were empty
	Expect(names).To(BeEmpty())

	// Test init
	err := plugin.Init(
		logging.ForPlugin("test-log",
			logrus.NewLogRegistry()),
		connection,
		index,
		true)

	Expect(err).To(BeNil())
	return plugin, index, connection
}

// Tests InterfaceConfigurator resync
func TestDataResyncResync(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

	Expect(plugin.IsSocketFilenameCached("testsocket")).To(BeTrue())

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

	errs := plugin.Resync(intfaces)
	Expect(errs).To(BeEmpty())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_MEMORY_INTERFACE))
}

// Tests InterfaceConfigurator resync with SwIfIndex 0
func TestDataResyncResyncIdx0(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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

	errs := plugin.Resync(intfaces)
	Expect(errs).To(BeEmpty())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_MEMORY_INTERFACE))
}

// Tests InterfaceConfigurator resync with same interface name/tag
func TestDataResyncResyncSameName(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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

	errs := plugin.Resync(intfaces)
	Expect(errs).To(BeEmpty())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_AF_PACKET_INTERFACE))
}

// Tests InterfaceConfigurator resync with unnamed interface
func TestDataResyncResyncUnnamed(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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

	errs := plugin.Resync(intfaces)
	Expect(errs).To(BeEmpty())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_MEMORY_INTERFACE))
}

// Tests InterfaceConfigurator resync with unnumbered VXLAN interface
func TestDataResyncResyncUnnumbered(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

	// Test
	intfaces := []*intf.Interfaces_Interface{
		{
			Name:        "test",
			Type:        intf.InterfaceType_VXLAN_TUNNEL,
			Enabled:     true,
			IpAddresses: []string{"192.168.0.1/24"},
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

	errs := plugin.Resync(intfaces)
	Expect(errs).To(BeEmpty())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_VXLAN_TUNNEL))
}

// Tests InterfaceConfigurator resync with unnumbered tap interface
func TestDataResyncResyncUnnumberedTap(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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

	errs := plugin.Resync(intfaces)
	Expect(errs).To(BeEmpty())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_TAP_INTERFACE))
}

// Tests InterfaceConfigurator resync with unnumbered AF_PACKET interface
func TestDataResyncResyncUnnumberedAfPacket(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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

	errs := plugin.Resync(intfaces)
	Expect(errs).To(BeEmpty())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_AF_PACKET_INTERFACE))
}

// Tests InterfaceConfigurator resync with unnumbered MEMIF interface
func TestDataResyncResyncUnnumberedMemif(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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

	errs := plugin.Resync(intfaces)
	Expect(errs).To(BeEmpty())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeTrue())
	Expect(meta).To(Not(BeNil()))
	Expect(meta.Type).To(BeEquivalentTo(intf.InterfaceType_MEMORY_INTERFACE))
}

// Tests if InterfaceConfigurator VPP config is present
func TestDataResyncVerifyVPPConfigPresence(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

	// Test
	ok := plugin.VerifyVPPConfigPresence([]*intf.Interfaces_Interface{})
	Expect(ok).To(BeFalse())

	_, meta, found := plugin.GetSwIfIndexes().LookupIdx("test")
	Expect(found).To(BeFalse())
	Expect(meta).To(BeNil())
}

// Tests BFDConfigurator session resync
func TestDataResyncResyncSession(t *testing.T) {
	// Setup
	plugin, conn, index := bfdConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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
	// Setup
	plugin, conn, index := bfdConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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
	// Setup
	plugin, conn, _ := bfdConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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
	// Setup
	plugin, conn, _ := bfdConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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
	// Setup
	plugin, conn, index := bfdConfiguratorTestInitialization(t, []*vppReplyMock{
		{
			Name:    (&bfdApi.BfdUDPSetEchoSource{}).GetMessageName(),
			Message: &bfdApi.BfdUDPSetEchoSourceReply{},
		},
	})

	defer plugin.Close()
	defer conn.Disconnect()

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
	// Setup
	plugin, conn := stnConfiguratorTestInitialization(t, []*vppReplyMock{
		{
			Name:    (&stnApi.StnRulesDump{}).GetMessageName(),
			Ping:    true,
			Message: &stnApi.StnRulesDetails{},
		},
		{
			Name:    (&stnApi.StnAddDelRule{}).GetMessageName(),
			Message: &stnApi.StnAddDelRuleReply{},
		},
	})

	defer plugin.Close()
	defer conn.Disconnect()

	// Test
	nbStnRules := []*stn.STN_Rule{
		{
			RuleName:  "test",
			Interface: "if0",
			IpAddress: "192.168.0.1/24",
		},
	}

	ok := plugin.Resync(nbStnRules)
	Expect(ok).To(BeNil())

	Expect(plugin.IndexExistsFor(ifplugin.StnIdentifier("if0"))).To(BeTrue())
}

// Tests NATConfigurator NAT global resync
func TestDataResyncResyncNatGlobal(t *testing.T) {
	// Setup
	plugin, _, conn := natConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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
	// Setup
	plugin, _, conn := natConfiguratorTestInitialization(t, []*vppReplyMock{})
	defer plugin.Close()
	defer conn.Disconnect()

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
	// Setup
	plugin, index, conn := natConfiguratorTestInitialization(t, []*vppReplyMock{
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

	defer plugin.Close()
	defer conn.Disconnect()

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

	idIdent := ifplugin.GetIdMappingIdentifier(&nat.Nat44DNat_DNatConfig_IdentityMapping{
		Protocol:  nat.Protocol_TCP,
		IpAddress: "192.168.0.1",
	})

	Expect(plugin.IsDNatLabelIdMappingRegistered(idIdent)).To(BeTrue())
	Expect(plugin.IsDNatLabelStMappingRegistered(stIdent)).To(BeTrue())
}

// Tests NATConfigurator DNAT resync
func TestDataResyncResyncDNatMultipleIPs(t *testing.T) {
	// Setup
	plugin, index, conn := natConfiguratorTestInitialization(t, []*vppReplyMock{
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
				Protocol: 6,
				Tag:      []byte("smap|lbstat|idmap"),
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

	defer plugin.Close()
	defer conn.Disconnect()

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

	idIdent := ifplugin.GetIdMappingIdentifier(&nat.Nat44DNat_DNatConfig_IdentityMapping{
		Protocol:  nat.Protocol_TCP,
		IpAddress: "192.168.0.1",
	})

	Expect(plugin.IsDNatLabelIdMappingRegistered(idIdent)).To(BeTrue())
	Expect(plugin.IsDNatLabelStMappingRegistered(stIdent)).To(BeTrue())
}
