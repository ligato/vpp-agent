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

package vpp2001_test

import (
	"testing"

	. "github.com/onsi/gomega"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestSetRxMode(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetRxModeReply{})

	err := ifHandler.SetRxMode(1, &ifs.Interface_RxMode{
		Mode:        ifs.Interface_RxMode_DEFAULT,
		Queue:       1,
		DefaultMode: false,
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetRxMode)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Mode).To(BeEquivalentTo(4))
	Expect(vppMsg.QueueID).To(BeEquivalentTo(1))
	Expect(vppMsg.QueueIDValid).To(BeTrue())
}

func TestSetRxModeError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetRxMode{})

	err := ifHandler.SetRxMode(1, &ifs.Interface_RxMode{
		Mode:        ifs.Interface_RxMode_DEFAULT,
		Queue:       1,
		DefaultMode: false,
	})

	Expect(err).ToNot(BeNil())
}

func TestSetRxModeRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetRxModeReply{
		Retval: 1,
	})

	err := ifHandler.SetRxMode(1, &ifs.Interface_RxMode{
		Mode:        ifs.Interface_RxMode_DEFAULT,
		Queue:       1,
		DefaultMode: false,
	})

	Expect(err).ToNot(BeNil())
}

func TestSetDefaultRxMode(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetRxModeReply{})

	err := ifHandler.SetRxMode(5, &ifs.Interface_RxMode{
		Mode:        ifs.Interface_RxMode_POLLING,
		Queue:       10, // ignored on the VPP side
		DefaultMode: true,
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetRxMode)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(5))
	Expect(vppMsg.Mode).To(BeEquivalentTo(1))
	Expect(vppMsg.QueueID).To(BeEquivalentTo(10))
	Expect(vppMsg.QueueIDValid).To(BeFalse())
}
