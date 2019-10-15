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

package vpp2001_379_test

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	vpp_ifs "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001_379/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls/vpp2001_379"
	"github.com/ligato/vpp-agent/plugins/vpp/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestInterfaceAdminDown(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetFlagsReply{})
	err := ifHandler.InterfaceAdminDown(1)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetFlags)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.SwIfIndex).To(Equal(vpp_ifs.InterfaceIndex(1)))
	Expect(vppMsg.Flags).To(Equal(vpp_ifs.IfStatusFlags(0)))
}

func TestInterfaceAdminDownError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.HwInterfaceSetMtuReply{})
	err := ifHandler.InterfaceAdminDown(1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceAdminDownRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetFlagsReply{
		Retval: 1,
	})
	err := ifHandler.InterfaceAdminDown(1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceAdminUp(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetFlagsReply{})
	err := ifHandler.InterfaceAdminUp(1)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetFlags)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.SwIfIndex).To(Equal(vpp_ifs.InterfaceIndex(1)))
	Expect(vppMsg.Flags).To(Equal(vpp_ifs.IF_STATUS_API_FLAG_ADMIN_UP))
}

func TestInterfaceAdminUpError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.HwInterfaceSetMtuReply{})
	err := ifHandler.InterfaceAdminDown(1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceAdminUpRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetFlagsReply{
		Retval: 1,
	})
	err := ifHandler.InterfaceAdminDown(1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceSetTag(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})
	err := ifHandler.SetInterfaceTag("tag", 1)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceTagAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.Tag).To(BeEquivalentTo("tag"))
	Expect(vppMsg.SwIfIndex).To(Equal(vpp_ifs.InterfaceIndex(1)))
	Expect(vppMsg.IsAdd).To(BeTrue())
}

func TestInterfaceSetTagError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.HwInterfaceSetMtuReply{})
	err := ifHandler.SetInterfaceTag("tag", 1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceSetTagRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{
		Retval: 1,
	})
	err := ifHandler.SetInterfaceTag("tag", 1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceRemoveTag(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})
	err := ifHandler.RemoveInterfaceTag("tag", 1)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceTagAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.Tag).To(BeEquivalentTo("tag"))
	Expect(vppMsg.IsAdd).To(BeFalse())
}

func TestInterfaceRemoveTagError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.HwInterfaceSetMtuReply{})
	err := ifHandler.RemoveInterfaceTag("tag", 1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceRemoveTagRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{
		Retval: 1,
	})
	err := ifHandler.RemoveInterfaceTag("tag", 1)

	Expect(err).ToNot(BeNil())
}

func ifTestSetup(t *testing.T) (*vppcallmock.TestCtx, vppcalls.InterfaceVppAPI) {
	ctx := vppcallmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	ifHandler := vpp2001_379.NewInterfaceVppHandler(ctx.MockChannel, log)
	return ctx, ifHandler
}
