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

package vpp2106

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/abfplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/aclidx"
	vpp_abf "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/abf"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
)

func TestGetABFVersion(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfPluginGetVersionReply{
		Major: 1,
		Minor: 0,
	})
	version, err := abfHandler.GetAbfVersion()

	Expect(err).To(BeNil())
	Expect(version).To(Equal("1.0"))
}

func TestAddABFPolicy(t *testing.T) {
	ctx, abfHandler, ifIndexes := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfPolicyAddDelReply{})

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{
		SwIfIndex: 5,
	})
	ifIndexes.Put("if2", &ifaceidx.IfaceMetadata{
		SwIfIndex: 10,
	})

	err := abfHandler.AddAbfPolicy(1, 2, []*abf.ABF_ForwardingPath{
		{
			InterfaceName: "if1",
			NextHopIp:     "10.0.0.1",
		},
		{
			InterfaceName: "if2",
			NextHopIp:     "ffff::",
		},
	})

	Expect(err).To(BeNil())
	req, ok := ctx.MockChannel.Msg.(*vpp_abf.AbfPolicyAddDel)
	Expect(ok).To(BeTrue())
	Expect(req.IsAdd).To(BeTrue())
	Expect(req.Policy.PolicyID).To(Equal(uint32(1)))
	Expect(req.Policy.ACLIndex).To(Equal(uint32(2)))
	Expect(req.Policy.NPaths).To(Equal(uint8(2)))
	Expect(req.Policy.Paths[0].SwIfIndex).To(Equal(uint32(5)))
	Expect(req.Policy.Paths[0].Nh.Address.GetIP4()).To(BeEquivalentTo(ip_types.IP4Address([4]uint8{10, 0, 0, 1})))
	Expect(req.Policy.Paths[1].SwIfIndex).To(Equal(uint32(10)))
	Expect(req.Policy.Paths[1].Nh.Address.GetIP6()).To(BeEquivalentTo(ip_types.IP6Address([16]uint8{255, 255})))
}

func TestAddABFPolicyError(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfPolicyAddDelReply{
		Retval: 1,
	})

	err := abfHandler.AddAbfPolicy(1, 2, nil)

	Expect(err).ToNot(BeNil())
}

func TestDeleteABFPolicy(t *testing.T) {
	ctx, abfHandler, ifIndexes := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfPolicyAddDelReply{})

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{
		SwIfIndex: 5,
	})
	ifIndexes.Put("if2", &ifaceidx.IfaceMetadata{
		SwIfIndex: 10,
	})

	err := abfHandler.DeleteAbfPolicy(1, []*abf.ABF_ForwardingPath{
		{
			InterfaceName: "if1",
			NextHopIp:     "10.0.0.1",
		},
		{
			InterfaceName: "if2",
			NextHopIp:     "ffff::",
		},
	})

	Expect(err).To(BeNil())
	req, ok := ctx.MockChannel.Msg.(*vpp_abf.AbfPolicyAddDel)
	Expect(ok).To(BeTrue())
	Expect(req.IsAdd).To(BeFalse())
	Expect(req.Policy.PolicyID).To(Equal(uint32(1)))
	Expect(req.Policy.NPaths).To(Equal(uint8(2)))
	Expect(req.Policy.Paths[0].SwIfIndex).To(Equal(uint32(5)))
	Expect(req.Policy.Paths[0].Nh.Address.XXX_UnionData[:4]).To(BeEquivalentTo(net.ParseIP("10.0.0.1").To4()))
	Expect(req.Policy.Paths[1].SwIfIndex).To(Equal(uint32(10)))
	Expect(req.Policy.Paths[1].Nh.Address.XXX_UnionData[:]).To(BeEquivalentTo(net.ParseIP("ffff::").To16()))
}

func TestDeleteABFPolicyError(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfPolicyAddDelReply{
		Retval: 1,
	})

	err := abfHandler.DeleteAbfPolicy(1, nil)

	Expect(err).ToNot(BeNil())
}

func TestAttachABFInterfaceIPv4(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfItfAttachAddDelReply{})

	err := abfHandler.AbfAttachInterfaceIPv4(1, 2, 3)

	Expect(err).To(BeNil())
	req, ok := ctx.MockChannel.Msg.(*vpp_abf.AbfItfAttachAddDel)
	Expect(ok).To(BeTrue())
	Expect(req.IsAdd).To(BeTrue())
	Expect(req.Attach.PolicyID).To(Equal(uint32(1)))
	Expect(req.Attach.SwIfIndex).To(BeEquivalentTo(uint32(2)))
	Expect(req.Attach.Priority).To(Equal(uint32(3)))
	Expect(req.Attach.IsIPv6).To(BeFalse())
}

func TestAttachABFInterfaceIPv4Error(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfItfAttachAddDelReply{
		Retval: -1,
	})

	err := abfHandler.AbfAttachInterfaceIPv4(1, 2, 3)

	Expect(err).ToNot(BeNil())
}

func TestAttachABFInterfaceIPv6(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfItfAttachAddDelReply{})

	err := abfHandler.AbfAttachInterfaceIPv6(1, 2, 3)

	Expect(err).To(BeNil())
	req, ok := ctx.MockChannel.Msg.(*vpp_abf.AbfItfAttachAddDel)
	Expect(ok).To(BeTrue())
	Expect(req.IsAdd).To(BeTrue())
	Expect(req.Attach.PolicyID).To(Equal(uint32(1)))
	Expect(req.Attach.SwIfIndex).To(BeEquivalentTo(uint32(2)))
	Expect(req.Attach.Priority).To(Equal(uint32(3)))
	Expect(req.Attach.IsIPv6).To(BeTrue())
}

func TestAttachABFInterfaceIPv6Error(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfItfAttachAddDelReply{
		Retval: -1,
	})

	err := abfHandler.AbfAttachInterfaceIPv6(1, 2, 3)

	Expect(err).ToNot(BeNil())
}

func TestDetachABFInterfaceIPv4(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfItfAttachAddDelReply{})

	err := abfHandler.AbfDetachInterfaceIPv4(1, 2, 3)

	Expect(err).To(BeNil())
	req, ok := ctx.MockChannel.Msg.(*vpp_abf.AbfItfAttachAddDel)
	Expect(ok).To(BeTrue())
	Expect(req.IsAdd).To(BeFalse())
	Expect(req.Attach.PolicyID).To(Equal(uint32(1)))
	Expect(req.Attach.SwIfIndex).To(BeEquivalentTo(uint32(2)))
	Expect(req.Attach.Priority).To(Equal(uint32(3)))
	Expect(req.Attach.IsIPv6).To(BeFalse())
}

func TestDetachABFInterfaceIPv4Error(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfItfAttachAddDelReply{
		Retval: -1,
	})

	err := abfHandler.AbfDetachInterfaceIPv4(1, 2, 3)

	Expect(err).ToNot(BeNil())
}

func TestDetachABFInterfaceIPv6(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfItfAttachAddDelReply{})

	err := abfHandler.AbfDetachInterfaceIPv6(1, 2, 3)

	Expect(err).To(BeNil())
	req, ok := ctx.MockChannel.Msg.(*vpp_abf.AbfItfAttachAddDel)
	Expect(ok).To(BeTrue())
	Expect(req.IsAdd).To(BeFalse())
	Expect(req.Attach.PolicyID).To(Equal(uint32(1)))
	Expect(req.Attach.SwIfIndex).To(BeEquivalentTo(uint32(2)))
	Expect(req.Attach.Priority).To(Equal(uint32(3)))
	Expect(req.Attach.IsIPv6).To(BeTrue())
}

func TestDetachABFInterfaceIPv6Error(t *testing.T) {
	ctx, abfHandler, _ := abfTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_abf.AbfItfAttachAddDelReply{
		Retval: -1,
	})

	err := abfHandler.AbfDetachInterfaceIPv6(1, 2, 3)

	Expect(err).ToNot(BeNil())
}

func abfTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.ABFVppAPI, ifaceidx.IfaceMetadataIndexRW) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	aclIdx := aclidx.NewACLIndex(log, "acl-index")
	ifIdx := ifaceidx.NewIfaceIndex(log, "if-index")
	abfHandler := NewABFVppHandler(ctx.MockChannel, aclIdx, ifIdx, log)
	return ctx, abfHandler, ifIdx
}
