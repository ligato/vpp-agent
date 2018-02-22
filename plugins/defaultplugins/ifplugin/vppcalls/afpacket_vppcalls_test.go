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

package vppcalls_test

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/af_packet"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

const (
	dummyIfName     = "TestIfName"
	dummyHostIFName = "TestHostIfName"
)

var dummyAfPacket = &intf.Interfaces_Interface_Afpacket{
	HostIfName: dummyHostIFName,
}

var testAfPacketAddData = &af_packet.AfPacketCreate{
	HostIfName:      []byte(dummyHostIFName),
	HwAddr:          nil,
	UseRandomHwAddr: 1,
}

var testAfPacketDelData = &af_packet.AfPacketDelete{
	HostIfName: []byte(dummyHostIFName),
}

func TestAddAfPacketInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&af_packet.AfPacketCreateReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	ifIndex, err := vppcalls.AddAfPacketInterface(dummyIfName, dummyAfPacket, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ifIndex).ShouldNot(BeNil())
	Expect(len(ctx.MockChannel.Msgs)).To(BeEquivalentTo(2))
	for i, msg := range ctx.MockChannel.Msgs {
		if i == 0 {
			vppMsg, ok := msg.(*af_packet.AfPacketCreate)
			Expect(ok).To(BeTrue())
			Expect(vppMsg).To(Equal(testAfPacketAddData))
		}
	}
}

func TestDeleteAfPacketInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&af_packet.AfPacketDeleteReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	err := vppcalls.DeleteAfPacketInterface(dummyIfName, 0, dummyAfPacket, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(len(ctx.MockChannel.Msgs)).To(BeEquivalentTo(2))
	for i, msg := range ctx.MockChannel.Msgs {
		if i == 0 {
			vppMsg, ok := msg.(*af_packet.AfPacketDelete)
			Expect(ok).To(BeTrue())
			Expect(vppMsg).To(Equal(testAfPacketDelData))
		}
	}
}
