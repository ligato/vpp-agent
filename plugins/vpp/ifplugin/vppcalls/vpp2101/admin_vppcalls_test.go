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

package vpp2101_test

import (
	"path"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/core"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	vpp2101 "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp2101"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
)

func TestInterfaceAdminDown(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetFlagsReply{})
	err := ifHandler.InterfaceAdminDown(ctx.Context, 1)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetFlags)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.SwIfIndex).To(Equal(interface_types.InterfaceIndex(1)))
	Expect(vppMsg.Flags).To(Equal(interface_types.IfStatusFlags(0)))
}

func TestInterfaceAdminDownError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.HwInterfaceSetMtuReply{})
	err := ifHandler.InterfaceAdminDown(ctx.Context, 1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceAdminDownRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetFlagsReply{
		Retval: 1,
	})
	err := ifHandler.InterfaceAdminDown(ctx.Context, 1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceAdminUp(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetFlagsReply{})
	err := ifHandler.InterfaceAdminUp(ctx.Context, 1)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetFlags)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg.SwIfIndex).To(Equal(interface_types.InterfaceIndex(1)))
	Expect(vppMsg.Flags).To(Equal(interface_types.IF_STATUS_API_FLAG_ADMIN_UP))
}

func TestInterfaceAdminUpError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.HwInterfaceSetMtuReply{})
	err := ifHandler.InterfaceAdminDown(ctx.Context, 1)

	Expect(err).ToNot(BeNil())
}

func TestInterfaceAdminUpRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetFlagsReply{
		Retval: 1,
	})
	err := ifHandler.InterfaceAdminDown(ctx.Context, 1)

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
	Expect(vppMsg.SwIfIndex).To(Equal(interface_types.InterfaceIndex(1)))
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

func ifTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.InterfaceVppAPI) {
	// FIXME: this control pings below are hacked to avoid issues in tests
	// that do not properly handle all replies, affecting tests that run after
	// causing failures, because of unexpected ControlPingReply type from core
	controlPingMsg := &vpe.ControlPingReply{}
	binapiPath := path.Dir(reflect.TypeOf(controlPingMsg).Elem().PkgPath())
	api.GetRegisteredMessages()[binapiPath]["control_ping_reply"] = controlPingMsg
	api.GetRegisteredMessages()[binapiPath]["control_ping_reply_f6b0b8ca"] = controlPingMsg
	core.SetControlPingReply(controlPingMsg)
	ctx := vppmock.SetupTestCtx(t)
	core.SetControlPingReply(controlPingMsg)
	ctx.PingReplyMsg = controlPingMsg
	log := logrus.NewLogger("test-log")
	ifHandler := vpp2101.NewInterfaceVppHandler(ctx.MockVPPClient, log)
	return ctx, ifHandler
}
