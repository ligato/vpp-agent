// Copyright (c) 2019 Cisco and/or its affiliates.
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

package vpp2001_test

import (
	"testing"

	. "github.com/onsi/gomega"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
)

func TestCreateSubif(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ifs.CreateVlanSubifReply{
		SwIfIndex: 2,
	})
	swifindex, err := ifHandler.CreateSubif(5, 32)
	Expect(err).To(BeNil())
	Expect(swifindex).To(Equal(uint32(2)))
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.CreateVlanSubif)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).ToNot(BeNil())
	Expect(vppMsg.SwIfIndex).To(Equal(vpp_ifs.InterfaceIndex(5)))
	Expect(vppMsg.VlanID).To(Equal(uint32(32)))
}

func TestCreateSubifError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ifs.CreateVlanSubifReply{
		SwIfIndex: 2,
		Retval:    9,
	})
	swifindex, err := ifHandler.CreateSubif(5, 32)
	Expect(err).ToNot(BeNil())
	Expect(swifindex).To(Equal(uint32(0)))
}

func TestDeleteSubif(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ifs.DeleteSubifReply{
		Retval: 2,
	})
	err := ifHandler.DeleteSubif(5)
	Expect(err).ToNot(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.DeleteSubif)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).ToNot(BeNil())
	Expect(vppMsg.SwIfIndex).To(Equal(vpp_ifs.InterfaceIndex(5)))
}

func TestDeleteSubifError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ifs.DeleteSubifReply{
		Retval: 2,
	})
	err := ifHandler.DeleteSubif(5)
	Expect(err).ToNot(BeNil())
}
