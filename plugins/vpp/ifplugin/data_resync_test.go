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
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
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
	ID      uint16
	Ping    bool
	Message api.Message
}

func vppMockHandler(vppMock *mock.VppAdapter, dataList []*vppReplyMock) mock.ReplyHandler {
	var sendControlPing bool

	return func(request mock.MessageDTO) (reply []byte, msgID uint16, prepared bool) {
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
			if request.MsgID == dataMock.ID {
				// Send control ping next iteration if set
				sendControlPing = dataMock.Ping
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
		}

		return reply, 0, false
	}
}

func interfaceConfiguratorTestInitialization(t *testing.T, mocks []*vppReplyMock, merge bool) (*ifplugin.InterfaceConfigurator, *core.Connection) {
	// Setup
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}

	if merge {
		mocks = append(mocks, []*vppReplyMock{
			{
				ID:      200,
				Ping:    true,
				Message: &interfaces.SwInterfaceDetails{},
			},
			{
				ID:      1001,
				Ping:    false,
				Message: &memif.MemifCreateReply{},
			},
			{
				ID:      1003,
				Message: &memif.MemifDeleteReply{},
			},
			{
				ID:      1005,
				Ping:    true,
				Message: &memif.MemifDetails{},
			},
			{
				ID:      1007,
				Message: &vxlan.VxlanAddDelTunnelReply{},
			},
			{
				ID:      1009,
				Ping:    true,
				Message: &vxlan.VxlanTunnelDetails{},
			},
			{
				ID:      1011,
				Ping:    false,
				Message: &af_packet.AfPacketCreateReply{},
			},
			{
				ID:      1013,
				Message: &af_packet.AfPacketDeleteReply{},
			},
			{
				ID:      1019,
				Ping:    true,
				Message: &tap.SwInterfaceTapDetails{},
			},
			{
				ID:      1021,
				Message: &tapv2.TapCreateV2Reply{},
			},
			{
				ID:      1023,
				Message: &tapv2.TapDeleteV2Reply{},
			},
			{
				ID:      1047,
				Ping:    true,
				Message: &tapv2.SwInterfaceTapV2Details{},
			},
			{
				ID:      1026,
				Ping:    false,
				Message: &interfaces.SwInterfaceSetFlagsReply{},
			},
			{
				ID:      1028,
				Ping:    false,
				Message: &interfaces.SwInterfaceAddDelAddressReply{},
			},
			{
				ID:      1032,
				Ping:    false,
				Message: &interfaces.SwInterfaceSetTableReply{},
			},
			{
				ID:      1034,
				Ping:    false,
				Message: &interfaces.SwInterfaceGetTableReply{},
			},
			{
				ID:      1036,
				Ping:    false,
				Message: &interfaces.SwInterfaceSetUnnumberedReply{},
			},
			{
				ID:      1038,
				Ping:    true,
				Message: &ip.IPAddressDetails{},
			},
			{
				ID:      1044,
				Ping:    true,
				Message: &memif.MemifSocketFilenameDetails{},
			},
			{
				ID:      1047,
				Ping:    true,
				Message: &tapv2.SwInterfaceTapV2Details{},
			},
			{
				ID:      1049,
				Ping:    false,
				Message: &interfaces.SwInterfaceTagAddDelReply{},
			},
		}...)
	}

	ctx.MockVpp.MockReplyHandler(vppMockHandler(ctx.MockVpp, mocks))

	connection, _ := core.Connect(ctx.MockVpp)
	plugin := &ifplugin.InterfaceConfigurator{}

	ifVppNotifChan := make(chan api.Message, 100)

	// Test init
	err := plugin.Init(
		logging.ForPlugin("test-log",
			logrus.NewLogRegistry()),
		connection,
		nil,
		ifVppNotifChan,
		0,
		false)

	Expect(err).To(BeNil())
	Expect(plugin.IsSocketFilenameCached("test")).To(BeTrue())

	return plugin, connection
}

func bfdConfiguratorTestInitialization(t *testing.T, mocks []*vppReplyMock) (*ifplugin.BFDConfigurator, *core.Connection, ifaceidx.SwIfIndexRW) {
	// Setup
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}

	ctx.MockVpp.MockReplyHandler(vppMockHandler(ctx.MockVpp, mocks))

	connection, _ := core.Connect(ctx.MockVpp)
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
		false)

	Expect(err).To(BeNil())

	return plugin, connection, index
}

func stnConfiguratorTestInitialization(t *testing.T, mocks []*vppReplyMock) (*ifplugin.StnConfigurator, *core.Connection) {
	// Setup
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}

	ctx.MockVpp.MockReplyHandler(vppMockHandler(ctx.MockVpp, mocks))

	connection, _ := core.Connect(ctx.MockVpp)
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
		false)

	Expect(err).To(BeNil())
	return plugin, connection
}

func natConfiguratorTestInitialization(t *testing.T, mocks []*vppReplyMock) (*ifplugin.NatConfigurator, ifaceidx.SwIfIndexRW, *core.Connection) {
	// Setup
	RegisterTestingT(t)

	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}

	ctx.MockVpp.MockReplyHandler(vppMockHandler(ctx.MockVpp, mocks))

	connection, _ := core.Connect(ctx.MockVpp)
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
		false)

	Expect(err).To(BeNil())
	return plugin, index, connection
}

// Tests InterfaceConfigurator resync
func TestDataResyncResync(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
		{
			ID:   200,
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
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	}, true)

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

// Tests InterfaceConfigurator resync with SwIfIndex 0
func TestDataResyncResyncIdx0(t *testing.T) {
	// Setup
	plugin, conn := interfaceConfiguratorTestInitialization(t, []*vppReplyMock{
		{
			ID:   200,
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
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	}, true)

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
			ID:   200,
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
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	}, true)

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
			ID:   200,
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: []byte("memif0"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
				SwIfIndex:     1,
			},
		},
		{
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	}, true)

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
			ID:   200,
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
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
		{
			ID:   1009,
			Ping: true,
			Message: &vxlan.VxlanTunnelDetails{
				SwIfIndex:  1,
				Vni:        12,
				SrcAddress: []byte("192.168.0.1"),
				DstAddress: []byte("192.168.10.1"),
			},
		},
	}, true)

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
			ID:   200,
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
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
		{
			ID:   1047,
			Ping: true,
			Message: &tapv2.SwInterfaceTapV2Details{
				SwIfIndex:  1,
				HostIfName: []byte("if0"),
				TxRingSz:   20,
			},
		},
	}, true)

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
			ID:   200,
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
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	}, true)

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
			ID:   200,
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
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
		{
			ID:   1005,
			Ping: true,
			Message: &memif.MemifDetails{
				SwIfIndex: 1,
				ID:        1,
				SocketID:  1,
				Role:      1,
				Mode:      1,
			},
		},
	}, true)

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
			ID:   200,
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: []byte("memif1"),
				Tag:           []byte("test"),
				AdminUpDown:   1,
				LinkMtu:       9216, // Default MTU
			},
		},
		{
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	}, true)

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
			ID:   1044,
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
	}, false)

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
			ID:      1001,
			Message: &bfdApi.BfdUDPAddReply{},
		},
		{
			ID:      1003,
			Message: &bfdApi.BfdUDPModReply{},
		},
		{
			ID:      1005,
			Message: &bfdApi.BfdUDPDelReply{},
		},
		{
			ID:      1011,
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
			ID:      1001,
			Message: &bfdApi.BfdUDPAddReply{},
		},
		{
			ID:      1003,
			Message: &bfdApi.BfdUDPModReply{},
		},
		{
			ID:      1005,
			Message: &bfdApi.BfdUDPDelReply{},
		},
		{
			ID:   1011,
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
			ID:      1011,
			Ping:    true,
			Message: &bfdApi.BfdAuthKeysDetails{},
		},
		{
			ID:      1013,
			Ping:    true,
			Message: &bfdApi.BfdUDPSessionDetails{},
		},
		{
			ID:      1009,
			Message: &bfdApi.BfdAuthDelKeyReply{},
		},
		{
			ID:      1007,
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
			ID:   1011,
			Ping: true,
			Message: &bfdApi.BfdAuthKeysDetails{
				ConfKeyID: 2,
			},
		},
		{
			ID:      1013,
			Ping:    true,
			Message: &bfdApi.BfdUDPSessionDetails{},
		},
		{
			ID:      1009,
			Message: &bfdApi.BfdAuthDelKeyReply{},
		},
		{
			ID:      1007,
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
			ID:      1011,
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
			ID:      1003,
			Ping:    true,
			Message: &stnApi.StnRulesDetails{},
		},
		{
			ID:      1001,
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
			ID:      1011,
			Message: &natApi.Nat44ForwardingIsEnabledReply{},
		},
		{
			ID:      1013,
			Ping:    true,
			Message: &natApi.Nat44InterfaceDetails{},
		},
		{
			ID:      1014,
			Ping:    true,
			Message: &natApi.Nat44InterfaceOutputFeatureDetails{},
		},
		{
			ID:      1015,
			Ping:    true,
			Message: &natApi.Nat44InterfaceOutputFeatureDetails{},
		},
		{
			ID:      1017,
			Ping:    true,
			Message: &natApi.Nat44AddressDetails{},
		},
		{
			ID:      1001,
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
			ID:      1007,
			Message: &natApi.Nat44AddDelStaticMappingReply{},
		},
		{
			ID:      1017,
			Message: &natApi.Nat44AddDelIdentityMappingReply{},
		},
		{
			ID:   1011,
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
			ID:   1013,
			Ping: true,
			Message: &natApi.Nat44LbStaticMappingDetails{
				Protocol:     6,
				Tag:          []byte("smap|lbstat|idmap"),
				ExternalAddr: []byte{192, 168, 10, 0},
				ExternalPort: 88,
			},
		},
		{
			ID:   1015,
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
			ID:      1007,
			Message: &natApi.Nat44AddDelStaticMappingReply{},
		},
		{
			ID:      1009,
			Message: &natApi.Nat44AddDelLbStaticMappingReply{},
		},
		{
			ID:      1017,
			Message: &natApi.Nat44AddDelIdentityMappingReply{},
		},
		{
			ID:   1011,
			Ping: true,
			Message: &natApi.Nat44StaticMappingDetails{
				Protocol: 6,
				Tag:      []byte("smap|lbstat|idmap"),
			},
		},
		{
			ID:   1013,
			Ping: true,
			Message: &natApi.Nat44LbStaticMappingDetails{
				Protocol: 6,
				Tag:      []byte("smap|lbstat|idmap"),
			},
		},
		{
			ID:   1015,
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
