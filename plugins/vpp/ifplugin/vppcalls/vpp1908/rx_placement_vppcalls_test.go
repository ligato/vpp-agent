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

package vpp1908_test

import (
	"testing"

	ifApi "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/interfaces"

	. "github.com/onsi/gomega"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestSetRxPlacementForWorker(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ifApi.SwInterfaceSetRxPlacementReply{})

	err := ifHandler.SetRxPlacement(1, &interfaces.Interface_RxPlacement{
		Queue:      1,
		Worker:     2,
		MainThread: false,
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*ifApi.SwInterfaceSetRxPlacement)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.QueueID).To(BeEquivalentTo(1))
	Expect(vppMsg.WorkerID).To(BeEquivalentTo(uint32(2)))
	Expect(vppMsg.IsMain).To(BeEquivalentTo(uint32(0)))
}

func TestSetRxPlacementForMainThread(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ifApi.SwInterfaceSetRxPlacementReply{})

	err := ifHandler.SetRxPlacement(3, &interfaces.Interface_RxPlacement{
		Queue:      6,
		Worker:     2, // ignored on the VPP side
		MainThread: true,
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*ifApi.SwInterfaceSetRxPlacement)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(3))
	Expect(vppMsg.QueueID).To(BeEquivalentTo(6))
	Expect(vppMsg.WorkerID).To(BeEquivalentTo(uint32(2)))
	Expect(vppMsg.IsMain).To(BeEquivalentTo(uint32(1)))
}

func TestSetRxPlacementRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ifApi.SwInterfaceSetRxPlacementReply{
		Retval: 1,
	})

	err := ifHandler.SetRxPlacement(1, &interfaces.Interface_RxPlacement{
		Queue:      1,
		Worker:     2,
		MainThread: false,
	})

	Expect(err).ToNot(BeNil())
}

func TestSetRxPlacementError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ifApi.SwInterfaceSetRxPlacement{})

	err := ifHandler.SetRxPlacement(1, &interfaces.Interface_RxPlacement{
		Queue:      1,
		Worker:     2,
		MainThread: false,
	})

	Expect(err).ToNot(BeNil())
}
