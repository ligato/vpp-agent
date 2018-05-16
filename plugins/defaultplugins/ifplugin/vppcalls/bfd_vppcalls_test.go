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
	"net"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	bfd_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vpe"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestAddBfdUDPSession(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)
	bfdKeyIndexes.RegisterName(string(1), 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSession(&bfd.SingleHopBFD_Session{
		SourceAddress:         "10.0.0.1",
		DestinationAddress:    "20.0.0.1",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
		Authentication: &bfd.SingleHopBFD_Session_Authentication{
			KeyId:           1,
			AdvertisedKeyId: 1,
		},
	}, 1, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPAdd)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.DesiredMinTx).To(BeEquivalentTo(10))
	Expect(vppMsg.RequiredMinRx).To(BeEquivalentTo(15))
	Expect(vppMsg.DetectMult).To(BeEquivalentTo(2))
	Expect(vppMsg.IsIpv6).To(BeEquivalentTo(0))
	Expect(vppMsg.IsAuthenticated).To(BeEquivalentTo(1))
}

func TestAddBfdUDPSessionIPv6(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)
	bfdKeyIndexes.RegisterName(string(1), 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSession(&bfd.SingleHopBFD_Session{
		SourceAddress:         "2001:db8::1",
		DestinationAddress:    "2001:db8:0:1:1:1:1:1",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
		Authentication: &bfd.SingleHopBFD_Session_Authentication{
			KeyId:           1,
			AdvertisedKeyId: 1,
		},
	}, 1, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPAdd)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.DesiredMinTx).To(BeEquivalentTo(10))
	Expect(vppMsg.RequiredMinRx).To(BeEquivalentTo(15))
	Expect(vppMsg.DetectMult).To(BeEquivalentTo(2))
	Expect(vppMsg.IsIpv6).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAuthenticated).To(BeEquivalentTo(1))
}

func TestAddBfdUDPSessionAuthKeyNotFound(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSession(&bfd.SingleHopBFD_Session{
		SourceAddress:         "10.0.0.1",
		DestinationAddress:    "20.0.0.1",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
		Authentication: &bfd.SingleHopBFD_Session_Authentication{
			KeyId:           1,
			AdvertisedKeyId: 1,
		},
	}, 1, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPAdd)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.DesiredMinTx).To(BeEquivalentTo(10))
	Expect(vppMsg.RequiredMinRx).To(BeEquivalentTo(15))
	Expect(vppMsg.DetectMult).To(BeEquivalentTo(2))
	Expect(vppMsg.IsIpv6).To(BeEquivalentTo(0))
	Expect(vppMsg.IsAuthenticated).To(BeEquivalentTo(0))
}

func TestAddBfdUDPSessionNoAuthKey(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})
	err := vppcalls.AddBfdUDPSession(&bfd.SingleHopBFD_Session{
		SourceAddress:         "10.0.0.1",
		DestinationAddress:    "20.0.0.1",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
	}, 1, nil, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPAdd)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.DesiredMinTx).To(BeEquivalentTo(10))
	Expect(vppMsg.RequiredMinRx).To(BeEquivalentTo(15))
	Expect(vppMsg.DetectMult).To(BeEquivalentTo(2))
	Expect(vppMsg.IsIpv6).To(BeEquivalentTo(0))
	Expect(vppMsg.IsAuthenticated).To(BeEquivalentTo(0))
}

func TestAddBfdUDPSessionIncorrectSrcIPError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)
	bfdKeyIndexes.RegisterName(string(1), 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSession(&bfd.SingleHopBFD_Session{
		SourceAddress:         "incorrect-ip",
		DestinationAddress:    "20.0.0.1",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
		Authentication: &bfd.SingleHopBFD_Session_Authentication{
			KeyId:           1,
			AdvertisedKeyId: 1,
		},
	}, 1, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestAddBfdUDPSessionIncorrectDstIPError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)
	bfdKeyIndexes.RegisterName(string(1), 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSession(&bfd.SingleHopBFD_Session{
		SourceAddress:         "10.0.0.1",
		DestinationAddress:    "incorrect-ip",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
		Authentication: &bfd.SingleHopBFD_Session_Authentication{
			KeyId:           1,
			AdvertisedKeyId: 1,
		},
	}, 1, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestAddBfdUDPSessionIPVerError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)
	bfdKeyIndexes.RegisterName(string(1), 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSession(&bfd.SingleHopBFD_Session{
		SourceAddress:         "10.0.0.1",
		DestinationAddress:    "2001:db8:0:1:1:1:1:1",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
		Authentication: &bfd.SingleHopBFD_Session_Authentication{
			KeyId:           1,
			AdvertisedKeyId: 1,
		},
	}, 1, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestAddBfdUDPSessionError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{})
	err := vppcalls.AddBfdUDPSession(&bfd.SingleHopBFD_Session{
		SourceAddress:         "10.0.0.1",
		DestinationAddress:    "20.0.0.1",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
	}, 1, nil, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestAddBfdUDPSessionRetvalError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{
		Retval: 1,
	})
	err := vppcalls.AddBfdUDPSession(&bfd.SingleHopBFD_Session{
		SourceAddress:         "10.0.0.1",
		DestinationAddress:    "20.0.0.1",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
	}, 1, nil, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestAddBfdUDPSessionFromDetails(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)
	bfdKeyIndexes.RegisterName(string(1), 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSessionFromDetails(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex:       1,
		LocalAddr:       net.ParseIP("10.0.0.1"),
		PeerAddr:        net.ParseIP("20.0.0.1"),
		IsIpv6:          0,
		IsAuthenticated: 1,
		BfdKeyID:        1,
		RequiredMinRx:   15,
		DesiredMinTx:    10,
		DetectMult:      2,
	}, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPAdd)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.DesiredMinTx).To(BeEquivalentTo(10))
	Expect(vppMsg.RequiredMinRx).To(BeEquivalentTo(15))
	Expect(vppMsg.DetectMult).To(BeEquivalentTo(2))
	Expect(vppMsg.IsIpv6).To(BeEquivalentTo(0))
	Expect(vppMsg.IsAuthenticated).To(BeEquivalentTo(1))
}

func TestAddBfdUDPSessionFromDetailsAuthKeyNotFound(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSessionFromDetails(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex:       1,
		LocalAddr:       net.ParseIP("10.0.0.1"),
		PeerAddr:        net.ParseIP("20.0.0.1"),
		IsIpv6:          0,
		IsAuthenticated: 1,
		BfdKeyID:        1,
		RequiredMinRx:   15,
		DesiredMinTx:    10,
		DetectMult:      2,
	}, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPAdd)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.IsAuthenticated).To(BeEquivalentTo(0))
}

func TestAddBfdUDPSessionFromDetailsNoAuth(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSessionFromDetails(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex:       1,
		LocalAddr:       net.ParseIP("10.0.0.1"),
		PeerAddr:        net.ParseIP("20.0.0.1"),
		IsIpv6:          0,
		IsAuthenticated: 0,
		BfdKeyID:        1,
		RequiredMinRx:   15,
		DesiredMinTx:    10,
		DetectMult:      2,
	}, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPAdd)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.IsAuthenticated).To(BeEquivalentTo(0))
}

func TestAddBfdUDPSessionFromDetailsError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{})

	err := vppcalls.AddBfdUDPSessionFromDetails(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex: 1,
		LocalAddr: net.ParseIP("10.0.0.1"),
		PeerAddr:  net.ParseIP("20.0.0.1"),
		IsIpv6:    0,
	}, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestAddBfdUDPSessionFromDetailsRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	bfdKeyIndexes := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "bfd", nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{
		Retval: 1,
	})

	err := vppcalls.AddBfdUDPSessionFromDetails(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex: 1,
		LocalAddr: net.ParseIP("10.0.0.1"),
		PeerAddr:  net.ParseIP("20.0.0.1"),
		IsIpv6:    0,
	}, bfdKeyIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestModifyBfdUDPSession(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{})

	err := vppcalls.ModifyBfdUDPSession(&bfd.SingleHopBFD_Session{
		Interface:             "if1",
		SourceAddress:         "10.0.0.1",
		DestinationAddress:    "20.0.0.1",
		DesiredMinTxInterval:  10,
		RequiredMinRxInterval: 15,
		DetectMultiplier:      2,
	}, ifIndexes, ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPMod)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.DesiredMinTx).To(BeEquivalentTo(10))
	Expect(vppMsg.RequiredMinRx).To(BeEquivalentTo(15))
	Expect(vppMsg.DetectMult).To(BeEquivalentTo(2))
	Expect(vppMsg.IsIpv6).To(BeEquivalentTo(0))
}

func TestModifyBfdUDPSessionIPv6(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{})

	err := vppcalls.ModifyBfdUDPSession(&bfd.SingleHopBFD_Session{
		Interface:          "if1",
		SourceAddress:      "2001:db8::1",
		DestinationAddress: "2001:db8:0:1:1:1:1:1",
	}, ifIndexes, ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPMod)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.IsIpv6).To(BeEquivalentTo(1))
}

func TestModifyBfdUDPSessionDifferentIPVer(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{})

	err := vppcalls.ModifyBfdUDPSession(&bfd.SingleHopBFD_Session{
		Interface:          "if1",
		SourceAddress:      "10.0.0.1",
		DestinationAddress: "2001:db8:0:1:1:1:1:1",
	}, ifIndexes, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestModifyBfdUDPSessionNoInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{})

	err := vppcalls.ModifyBfdUDPSession(&bfd.SingleHopBFD_Session{
		Interface:          "if1",
		SourceAddress:      "10.0.0.1",
		DestinationAddress: "20.0.0.1",
	}, ifIndexes, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestModifyBfdUDPSessionInvalidSrcIP(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{})

	err := vppcalls.ModifyBfdUDPSession(&bfd.SingleHopBFD_Session{
		Interface:          "if1",
		SourceAddress:      "invalid-ip",
		DestinationAddress: "20.0.0.1",
	}, ifIndexes, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestModifyBfdUDPSessionInvalidDstIP(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{})

	err := vppcalls.ModifyBfdUDPSession(&bfd.SingleHopBFD_Session{
		Interface:          "if1",
		SourceAddress:      "10.0.0.1",
		DestinationAddress: "invalid-ip",
	}, ifIndexes, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestModifyBfdUDPSessionError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})

	err := vppcalls.ModifyBfdUDPSession(&bfd.SingleHopBFD_Session{
		Interface:          "if1",
		SourceAddress:      "10.0.0.1",
		DestinationAddress: "20.0.0.1",
	}, ifIndexes, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestModifyBfdUDPSessionRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{
		Retval: 1,
	})

	err := vppcalls.ModifyBfdUDPSession(&bfd.SingleHopBFD_Session{
		Interface:          "if1",
		SourceAddress:      "10.0.0.1",
		DestinationAddress: "20.0.0.1",
	}, ifIndexes, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestDeleteBfdUDPSession(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPDelReply{})

	err := vppcalls.DeleteBfdUDPSession(1, "10.0.0.1", "20.0.0.1", ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.LocalAddr).To(BeEquivalentTo(net.ParseIP("10.0.0.1").To4()))
	Expect(vppMsg.PeerAddr).To(BeEquivalentTo(net.ParseIP("20.0.0.1").To4()))
	Expect(vppMsg.IsIpv6).To(BeEquivalentTo(0))
}

func TestDeleteBfdUDPSessionError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{})

	err := vppcalls.DeleteBfdUDPSession(1, "10.0.0.1", "20.0.0.1", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestDeleteBfdUDPSessionRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPDelReply{
		Retval: 1,
	})

	err := vppcalls.DeleteBfdUDPSession(1, "10.0.0.1", "20.0.0.1", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestDumpBfdUDPSessions(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex: 1,
		LocalAddr: net.ParseIP("10.0.0.1"),
		PeerAddr:  net.ParseIP("20.0.0.1"),
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	sessions, err := vppcalls.DumpBfdUDPSessions(ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(sessions).To(HaveLen(1))
}

func TestDumpBfdUDPSessionsWithID(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Authenticated wiht ID 1
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex:       1,
		LocalAddr:       net.ParseIP("10.0.0.1"),
		PeerAddr:        net.ParseIP("20.0.0.1"),
		IsAuthenticated: 1,
		BfdKeyID:        1,
	})
	// Authenticated with ID 2 (filtered)
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex:       2,
		LocalAddr:       net.ParseIP("10.0.0.2"),
		PeerAddr:        net.ParseIP("20.0.0.2"),
		IsAuthenticated: 1,
		BfdKeyID:        2,
	})
	// Not authenticated
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex:       3,
		LocalAddr:       net.ParseIP("10.0.0.3"),
		PeerAddr:        net.ParseIP("20.0.0.3"),
		IsAuthenticated: 0,
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	sessions, err := vppcalls.DumpBfdUDPSessionsWithID(1, ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(sessions).To(HaveLen(1))
}

func TestSetBfdUDPAuthenticationKeySha1(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{})

	err := vppcalls.SetBfdUDPAuthenticationKey(&bfd.SingleHopBFD_Key{
		Name:               "bfd-key",
		AuthKeyIndex:       1,
		Id:                 1,
		AuthenticationType: bfd.SingleHopBFD_Key_KEYED_SHA1,
		Secret:             "secret",
	}, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdAuthSetKey)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.ConfKeyID).To(BeEquivalentTo(1))
	Expect(vppMsg.KeyLen).To(BeEquivalentTo(len("secret")))
	Expect(vppMsg.AuthType).To(BeEquivalentTo(4)) // Keyed SHA1
	Expect(vppMsg.Key).To(BeEquivalentTo([]byte("secret")))
}

func TestSetBfdUDPAuthenticationKeyMeticulous(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{})

	err := vppcalls.SetBfdUDPAuthenticationKey(&bfd.SingleHopBFD_Key{
		Name:               "bfd-key",
		AuthKeyIndex:       1,
		Id:                 1,
		AuthenticationType: bfd.SingleHopBFD_Key_METICULOUS_KEYED_SHA1,
		Secret:             "secret",
	}, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdAuthSetKey)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.ConfKeyID).To(BeEquivalentTo(1))
	Expect(vppMsg.KeyLen).To(BeEquivalentTo(len("secret")))
	Expect(vppMsg.AuthType).To(BeEquivalentTo(5)) // METICULOUS
	Expect(vppMsg.Key).To(BeEquivalentTo([]byte("secret")))
}

func TestSetBfdUDPAuthenticationKeyUnknown(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{})

	err := vppcalls.SetBfdUDPAuthenticationKey(&bfd.SingleHopBFD_Key{
		Name:               "bfd-key",
		AuthKeyIndex:       1,
		Id:                 1,
		AuthenticationType: 2, // Unknown type
		Secret:             "secret",
	}, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdAuthSetKey)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.ConfKeyID).To(BeEquivalentTo(1))
	Expect(vppMsg.KeyLen).To(BeEquivalentTo(len("secret")))
	Expect(vppMsg.AuthType).To(BeEquivalentTo(4)) // Keyed SHA1 as default
	Expect(vppMsg.Key).To(BeEquivalentTo([]byte("secret")))
}

func TestSetBfdUDPAuthenticationError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthDelKeyReply{})

	err := vppcalls.SetBfdUDPAuthenticationKey(&bfd.SingleHopBFD_Key{
		Name:               "bfd-key",
		AuthKeyIndex:       1,
		Id:                 1,
		AuthenticationType: bfd.SingleHopBFD_Key_KEYED_SHA1,
		Secret:             "secret",
	}, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestSetBfdUDPAuthenticationRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{
		Retval: 1,
	})

	err := vppcalls.SetBfdUDPAuthenticationKey(&bfd.SingleHopBFD_Key{
		Name:               "bfd-key",
		AuthKeyIndex:       1,
		Id:                 1,
		AuthenticationType: bfd.SingleHopBFD_Key_KEYED_SHA1,
		Secret:             "secret",
	}, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestDeleteBfdUDPAuthenticationKey(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthDelKeyReply{})

	err := vppcalls.DeleteBfdUDPAuthenticationKey(&bfd.SingleHopBFD_Key{
		Name:               "bfd-key",
		AuthKeyIndex:       1,
		Id:                 1,
		AuthenticationType: bfd.SingleHopBFD_Key_KEYED_SHA1,
		Secret:             "secret",
	}, ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdAuthDelKey)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.ConfKeyID).To(BeEquivalentTo(1))
}

func TestDeleteBfdUDPAuthenticationKeyError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{})

	err := vppcalls.DeleteBfdUDPAuthenticationKey(&bfd.SingleHopBFD_Key{
		Name:               "bfd-key",
		AuthKeyIndex:       1,
		Id:                 1,
		AuthenticationType: bfd.SingleHopBFD_Key_KEYED_SHA1,
		Secret:             "secret",
	}, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestDeleteBfdUDPAuthenticationKeyRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthDelKeyReply{
		Retval: 1,
	})

	err := vppcalls.DeleteBfdUDPAuthenticationKey(&bfd.SingleHopBFD_Key{
		Name:               "bfd-key",
		AuthKeyIndex:       1,
		Id:                 1,
		AuthenticationType: bfd.SingleHopBFD_Key_KEYED_SHA1,
		Secret:             "secret",
	}, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestDumpBfdKeys(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdAuthKeysDetails{
		ConfKeyID: 1,
		UseCount:  0,
		AuthType:  4,
	})
	ctx.MockVpp.MockReply(&bfd_api.BfdAuthKeysDetails{
		ConfKeyID: 2,
		UseCount:  1,
		AuthType:  5,
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	keys, err := vppcalls.DumpBfdKeys(ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(keys).To(HaveLen(2))
}

func TestAddBfdEchoFunction(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSetEchoSourceReply{})

	err := vppcalls.AddBfdEchoFunction(&bfd.SingleHopBFD_EchoFunction{
		Name:                "echo",
		EchoSourceInterface: "if1",
	}, ifIndexes, ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPSetEchoSource)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
}

func TestAddBfdEchoFunctionInterfaceNotFound(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSetEchoSourceReply{})

	err := vppcalls.AddBfdEchoFunction(&bfd.SingleHopBFD_EchoFunction{
		Name:                "echo",
		EchoSourceInterface: "if1",
	}, ifIndexes, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestAddBfdEchoFunctionError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPDelEchoSourceReply{})

	err := vppcalls.AddBfdEchoFunction(&bfd.SingleHopBFD_EchoFunction{
		Name:                "echo",
		EchoSourceInterface: "if1",
	}, ifIndexes, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestAddBfdEchoFunctionRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	ifIndexes.RegisterName("if1", 1, nil)

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSetEchoSourceReply{
		Retval: 1,
	})

	err := vppcalls.AddBfdEchoFunction(&bfd.SingleHopBFD_EchoFunction{
		Name:                "echo",
		EchoSourceInterface: "if1",
	}, ifIndexes, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestDeleteBfdEchoFunction(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPDelEchoSourceReply{})

	err := vppcalls.DeleteBfdEchoFunction(ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	_, ok := ctx.MockChannel.Msg.(*bfd_api.BfdUDPDelEchoSource)
	Expect(ok).To(BeTrue())
}

func TestDeleteBfdEchoFunctionError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSetEchoSourceReply{})

	err := vppcalls.DeleteBfdEchoFunction(ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestDeleteBfdEchoFunctionRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd_api.BfdUDPDelEchoSourceReply{
		Retval: 1,
	})

	err := vppcalls.DeleteBfdEchoFunction(ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}
