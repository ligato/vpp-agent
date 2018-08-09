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

package vppcalls_test

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	. "github.com/onsi/gomega"
)

func TestSetRxPlacement(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpe.CliInbandReply{})

	err := ifHandler.SetRxPlacement("if-internal", &interfaces.Interfaces_Interface_RxPlacementSettings{
		Queue:  1,
		Worker: 2,
	})

	expMsg := "set interface rx-placement if-internal queue 1 worker 2"
	expMsgLen := len(expMsg)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpe.CliInband)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.Cmd).To(BeEquivalentTo([]byte(expMsg)))
	Expect(vppMsg.Length).To(BeEquivalentTo(uint32(expMsgLen)))
}

func TestSetRxPlacementRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpe.CliInbandReply{
		Retval: 1,
	})

	err := ifHandler.SetRxPlacement("if-internal", &interfaces.Interfaces_Interface_RxPlacementSettings{
		Queue:  1,
		Worker: 2,
	})

	Expect(err).ToNot(BeNil())
}

func TestSetRxPlacementReply(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpe.CliInbandReply{
		Reply: []byte("dummy-reply"),
	})

	err := ifHandler.SetRxPlacement("if-internal", &interfaces.Interfaces_Interface_RxPlacementSettings{
		Queue:  1,
		Worker: 2,
	})

	Expect(err).ToNot(BeNil())
}

func TestSetRxPlacementError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpe.CliInband{})

	err := ifHandler.SetRxPlacement("if-internal", &interfaces.Interfaces_Interface_RxPlacementSettings{
		Queue:  1,
		Worker: 2,
	})

	Expect(err).ToNot(BeNil())
}
