//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package vpp2001_324_test

import (
	"testing"

	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_ifs "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001_324/interfaces"
	. "github.com/onsi/gomega"
)

func TestSetRxPlacementForWorker(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetRxPlacementReply{})

	err := ifHandler.SetRxPlacement(1, &ifs.Interface_RxPlacement{
		Queue:      1,
		Worker:     2,
		MainThread: false,
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetRxPlacement)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.QueueID).To(BeEquivalentTo(1))
	Expect(vppMsg.WorkerID).To(BeEquivalentTo(uint32(2)))
	Expect(vppMsg.IsMain).To(BeFalse())
}

func TestSetRxPlacementForMainThread(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetRxPlacementReply{})

	err := ifHandler.SetRxPlacement(3, &ifs.Interface_RxPlacement{
		Queue:      6,
		Worker:     2, // ignored on the VPP side
		MainThread: true,
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetRxPlacement)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(3))
	Expect(vppMsg.QueueID).To(BeEquivalentTo(6))
	Expect(vppMsg.WorkerID).To(BeEquivalentTo(uint32(2)))
	Expect(vppMsg.IsMain).To(BeTrue())
}

func TestSetRxPlacementRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetRxPlacementReply{
		Retval: 1,
	})

	err := ifHandler.SetRxPlacement(1, &ifs.Interface_RxPlacement{
		Queue:      1,
		Worker:     2,
		MainThread: false,
	})

	Expect(err).ToNot(BeNil())
}

func TestSetRxPlacementError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetRxPlacement{})

	err := ifHandler.SetRxPlacement(1, &ifs.Interface_RxPlacement{
		Queue:      1,
		Worker:     2,
		MainThread: false,
	})

	Expect(err).ToNot(BeNil())
}
