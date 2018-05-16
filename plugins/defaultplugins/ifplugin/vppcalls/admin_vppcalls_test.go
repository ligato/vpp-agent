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

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestInterfaceAdminDown(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	err := vppcalls.InterfaceAdminDown(1, ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceSetFlags)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.SwIfIndex).To(Equal(uint32(1)))
	Expect(vppMsg.AdminUpDown).To(Equal(uint8(0)))
}

func TestInterfaceAdminDownError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	err := vppcalls.InterfaceAdminDown(1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceAdminDownRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{
		Retval: 1,
	})
	err := vppcalls.InterfaceAdminDown(1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceAdminUp(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	err := vppcalls.InterfaceAdminUp(1, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceSetFlags)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.SwIfIndex).To(Equal(uint32(1)))
	Expect(vppMsg.AdminUpDown).To(Equal(uint8(1)))
}

func TestInterfaceAdminUpError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	err := vppcalls.InterfaceAdminDown(1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceAdminUpRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{
		Retval: 1,
	})
	err := vppcalls.InterfaceAdminDown(1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceSetTag(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	err := vppcalls.SetInterfaceTag("tag", 1, ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceTagAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.Tag).To(BeEquivalentTo("tag"))
	Expect(vppMsg.SwIfIndex).To(Equal(uint32(1)))
}

func TestInterfaceSetTagError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	err := vppcalls.SetInterfaceTag("tag", 1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceSetTagRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{
		Retval: 1,
	})
	err := vppcalls.SetInterfaceTag("tag", 1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceRemoveTag(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	err := vppcalls.RemoveInterfaceTag("tag", 1, ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceTagAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.Tag).To(BeEquivalentTo("tag"))
	Expect(vppMsg.IsAdd).To(Equal(uint8(0)))
}

func TestInterfaceRemoveTagError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	err := vppcalls.RemoveInterfaceTag("tag", 1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceRemoveTagRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{
		Retval: 1,
	})
	err := vppcalls.RemoveInterfaceTag("tag", 1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}
