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

	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/af_packet"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

const dummyHostIFName = "TestIFName"

var dummyAfPacket = &interfaces.Interfaces_Interface_Afpacket{
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

	ifIndex, err := vppcalls.AddAfPacketInterface(dummyAfPacket, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ifIndex).ShouldNot(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*af_packet.AfPacketCreate)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg).To(Equal(testAfPacketAddData))
}

func TestDeleteAfPacketInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&af_packet.AfPacketDeleteReply{})

	err := vppcalls.DeleteAfPacketInterface(dummyAfPacket, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*af_packet.AfPacketDelete)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg).To(Equal(testAfPacketDelData))
}
