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

	ipApi "github.com/ligato/vpp-binapi/binapi/ip"

	"github.com/ligato/cn-infra/logging/logrus"
	punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/puntplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	binapi "github.com/ligato/vpp-binapi/binapi/punt"
	. "github.com/onsi/gomega"
)

func TestAddPunt(t *testing.T) {
	ctx, puntHandler, _ := puntTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.SetPuntReply{})

	err := puntHandler.AddPunt(&punt.ToHost{
		L3Protocol: punt.L3Protocol_IPv4,
		L4Protocol: punt.L4Protocol_UDP,
		Port:       9000,
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*binapi.SetPunt)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.IsAdd).To(Equal(uint8(1)))
	Expect(vppMsg.Punt.IPv).To(Equal(uint8(4)))
	Expect(vppMsg.Punt.L4Protocol).To(Equal(uint8(17)))
	Expect(vppMsg.Punt.L4Port).To(Equal(uint16(9000)))
}

func TestDeletePunt(t *testing.T) {
	ctx, puntHandler, _ := puntTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.SetPuntReply{})

	err := puntHandler.DeletePunt(&punt.ToHost{
		L3Protocol: punt.L3Protocol_IPv4,
		L4Protocol: punt.L4Protocol_UDP,
		Port:       9000,
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*binapi.SetPunt)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.IsAdd).To(Equal(uint8(0)))
	Expect(vppMsg.Punt.IPv).To(Equal(uint8(4)))
	Expect(vppMsg.Punt.L4Protocol).To(Equal(uint8(17)))
	Expect(vppMsg.Punt.L4Port).To(Equal(uint16(9000)))
}

func TestRegisterPuntSocket(t *testing.T) {
	ctx, puntHandler, _ := puntTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.PuntSocketRegisterReply{})

	err := puntHandler.RegisterPuntSocket(&punt.ToHost{
		L3Protocol: punt.L3Protocol_IPv4,
		L4Protocol: punt.L4Protocol_UDP,
		Port:       9000,
		SocketPath: "/test/path/socket",
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*binapi.PuntSocketRegister)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.HeaderVersion).To(Equal(uint32(1)))
	Expect(vppMsg.Punt.IPv).To(Equal(uint8(4)))
	Expect(vppMsg.Punt.L4Protocol).To(Equal(uint8(17)))
	Expect(vppMsg.Punt.L4Port).To(Equal(uint16(9000)))
	Expect(vppMsg.Pathname).To(BeEquivalentTo("/test/path/socket"))
}

func TestRegisterPuntSocketAllIPv4(t *testing.T) {
	ctx, puntHandler, _ := puntTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.PuntSocketRegisterReply{})
	ctx.MockVpp.MockReply(&binapi.PuntSocketRegisterReply{})

	err := puntHandler.RegisterPuntSocket(&punt.ToHost{
		L3Protocol: punt.L3Protocol_ALL,
		L4Protocol: punt.L4Protocol_UDP,
		Port:       9000,
		SocketPath: "/test/path/socket",
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*binapi.PuntSocketRegister)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.HeaderVersion).To(Equal(uint32(1)))
	Expect(vppMsg.Punt.IPv).To(Equal(uint8(255)))
	Expect(vppMsg.Punt.L4Protocol).To(Equal(uint8(17)))
	Expect(vppMsg.Punt.L4Port).To(Equal(uint16(9000)))
	Expect(vppMsg.Pathname).To(BeEquivalentTo("/test/path/socket"))
}

func TestAddIPRedirect(t *testing.T) {
	ctx, puntHandler, ifIndexes := puntTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ipApi.IPPuntRedirectReply{})

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})
	ifIndexes.Put("if2", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	err := puntHandler.AddPuntRedirect(&punt.IPRedirect{
		L3Protocol:  punt.L3Protocol_IPv4,
		RxInterface: "if1",
		TxInterface: "if2",
		NextHop:     "10.0.0.1",
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*ipApi.IPPuntRedirect)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.IsAdd).To(Equal(uint8(1)))
	Expect(vppMsg.Punt.RxSwIfIndex).To(Equal(uint32(1)))
	Expect(vppMsg.Punt.TxSwIfIndex).To(Equal(uint32(2)))
	Expect(vppMsg.Punt.Nh).To(Equal(ipApi.Address{
		Af: ipApi.ADDRESS_IP4,
		Un: ipApi.AddressUnion{
			Union_data: [16]byte{10, 0, 0, 1},
		},
	}))
}

func TestAddIPRedirectAll(t *testing.T) {
	ctx, puntHandler, ifIndexes := puntTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ipApi.IPPuntRedirectReply{})

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	err := puntHandler.AddPuntRedirect(&punt.IPRedirect{
		L3Protocol:  punt.L3Protocol_IPv4,
		TxInterface: "if1",
		NextHop:     "30.0.0.1",
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*ipApi.IPPuntRedirect)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.IsAdd).To(Equal(uint8(1)))
	Expect(vppMsg.Punt.RxSwIfIndex).To(Equal(^uint32(0)))
	Expect(vppMsg.Punt.TxSwIfIndex).To(Equal(uint32(1)))
	Expect(vppMsg.Punt.Nh).To(Equal(ipApi.Address{
		Af: ipApi.ADDRESS_IP4,
		Un: ipApi.AddressUnion{
			Union_data: [16]byte{30, 0, 0, 1},
		},
	}))
}

func TestDeleteIPRedirect(t *testing.T) {
	ctx, puntHandler, ifIndexes := puntTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ipApi.IPPuntRedirectReply{})

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})
	ifIndexes.Put("if2", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	err := puntHandler.DeletePuntRedirect(&punt.IPRedirect{
		L3Protocol:  punt.L3Protocol_IPv4,
		RxInterface: "if1",
		TxInterface: "if2",
		NextHop:     "10.0.0.1",
	})

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*ipApi.IPPuntRedirect)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.IsAdd).To(Equal(uint8(0)))
	Expect(vppMsg.Punt.RxSwIfIndex).To(Equal(uint32(1)))
	Expect(vppMsg.Punt.TxSwIfIndex).To(Equal(uint32(2)))
	Expect(vppMsg.Punt.Nh).To(Equal(ipApi.Address{
		Af: ipApi.ADDRESS_IP4,
		Un: ipApi.AddressUnion{
			Union_data: [16]byte{10, 0, 0, 1},
		},
	}))
}

func puntTestSetup(t *testing.T) (*vppcallmock.TestCtx, vppcalls.PuntVppAPI, ifaceidx.IfaceMetadataIndexRW) {
	ctx := vppcallmock.SetupTestCtx(t)
	logger := logrus.NewLogger("test-log")
	ifIndexes := ifaceidx.NewIfaceIndex(logger, "punt-if-idx")
	puntHandler := vppcalls.NewPuntVppHandler(ctx.MockChannel, ifIndexes, logrus.DefaultLogger())
	return ctx, puntHandler, ifIndexes
}
