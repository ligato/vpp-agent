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

	. "github.com/onsi/gomega"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/interfaces"
)

func TestGetInterfaceVRF(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceGetTableReply{
		VrfID: 1,
	})

	vrfID, err := ifHandler.GetInterfaceVrf(1)
	Expect(err).To(BeNil())
	Expect(vrfID).To(BeEquivalentTo(1))
}

func TestGetInterfaceIPv6VRF(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceGetTableReply{
		VrfID: 1,
	})

	vrfID, err := ifHandler.GetInterfaceVrfIPv6(1)
	Expect(err).To(BeNil())
	Expect(vrfID).To(BeEquivalentTo(1))
}

func TestGetInterfaceVRFError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceGetTable{})

	_, err := ifHandler.GetInterfaceVrf(1)
	Expect(err).ToNot(BeNil())
}

func TestGetInterfaceVRFRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceGetTableReply{
		Retval: 1,
	})

	_, err := ifHandler.GetInterfaceVrf(1)
	Expect(err).ToNot(BeNil())
}

func TestSetInterfaceVRF(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})

	err := ifHandler.SetInterfaceVrf(1, 2)
	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceSetTable)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.VrfID).To(BeEquivalentTo(2))
}

func TestSetInterfaceIPv6VRF(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})

	err := ifHandler.SetInterfaceVrfIPv6(1, 2)
	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceSetTable)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.VrfID).To(BeEquivalentTo(2))
	Expect(vppMsg.IsIPv6).To(BeEquivalentTo(1))
}

func TestSetInterfaceVRFError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTable{})

	err := ifHandler.SetInterfaceVrf(1, 2)
	Expect(err).To(HaveOccurred())
}

func TestSetInterfaceVRFRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{
		Retval: 1,
	})

	err := ifHandler.SetInterfaceVrf(1, 2)
	Expect(err).ToNot(BeNil())
}
