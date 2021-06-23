//  Copyright (c) 2020 Pantheon.tech
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp2106_test

import (
	"testing"

	. "github.com/onsi/gomega"

	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	vpp_rdma "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/rdma"

	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestAddRdmaInterface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_rdma.RdmaCreateReply{
		SwIfIndex: 2,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	const qSize = 4096
	const qNum = 4
	rdmaLink := &ifs.RDMALink{
		HostIfName: "ens4",
		Mode:       ifs.RDMALink_DV,
		RxqNum:     qNum,
		RxqSize:    qSize,
		TxqSize:    qSize,
	}

	index, err := ifHandler.AddRdmaInterface(ctx.Context, "rdma1", rdmaLink)
	Expect(err).To(BeNil())
	Expect(index).To(Equal(uint32(2)))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_rdma.RdmaCreate)
		if ok {
			Expect(vppMsg.HostIf).To(BeEquivalentTo("ens4"))
			Expect(vppMsg.Name).To(BeEquivalentTo("rdma1"))
			Expect(vppMsg.Mode).To(BeEquivalentTo(vpp_rdma.RDMA_API_MODE_DV))
			Expect(vppMsg.RxqNum).To(BeEquivalentTo(qNum))
			Expect(vppMsg.RxqSize).To(BeEquivalentTo(qSize))
			Expect(vppMsg.TxqSize).To(BeEquivalentTo(qSize))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestDeleteRdmaInterface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_rdma.RdmaDeleteReply{})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DeleteRdmaInterface(ctx.Context, "rdma1", 2)
	Expect(err).To(BeNil())

	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_rdma.RdmaDelete)
		if ok {
			Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(2))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}
