//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package vpp1904_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls/vpp1904"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

var vrfTables = []*l3.VrfTable{
	{
		Id:       1,
		Protocol: l3.VrfTable_IPV4,
		Label:    "table1",
	},
	{
		Id:       1,
		Protocol: l3.VrfTable_IPV6,
		Label:    "table1",
	},
	{
		Id:       2,
		Protocol: l3.VrfTable_IPV6,
		Label:    "table2",
	},
}

// Test adding routes
func TestAddVrfTable(t *testing.T) {
	ctx, vtHandler := vrfTableTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	err := vtHandler.AddVrfTable(vrfTables[0])
	Expect(err).To(Succeed())

	vppMsg, ok := ctx.MockChannel.Msg.(*ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.TableID).To(BeEquivalentTo(1))
	Expect(vppMsg.IsIPv6).To(BeEquivalentTo(0))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
	Expect(vppMsg.Name).To(BeEquivalentTo([]byte("table1")))

	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	err = vtHandler.AddVrfTable(vrfTables[1])
	Expect(err).To(Succeed())

	vppMsg, ok = ctx.MockChannel.Msg.(*ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.TableID).To(BeEquivalentTo(1))
	Expect(vppMsg.IsIPv6).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
	Expect(vppMsg.Name).To(BeEquivalentTo([]byte("table1")))

	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	err = vtHandler.AddVrfTable(vrfTables[2])
	Expect(err).To(Succeed())

	vppMsg, ok = ctx.MockChannel.Msg.(*ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.TableID).To(BeEquivalentTo(2))
	Expect(vppMsg.IsIPv6).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
	Expect(vppMsg.Name).To(BeEquivalentTo([]byte("table2")))

	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{Retval: 1})
	err = vtHandler.AddVrfTable(vrfTables[0])
	Expect(err).To(Not(BeNil()))
}

// Test deleting routes
func TestDeleteVrfTable(t *testing.T) {
	ctx, vtHandler := vrfTableTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	err := vtHandler.DelVrfTable(vrfTables[0])
	Expect(err).To(Succeed())

	vppMsg, ok := ctx.MockChannel.Msg.(*ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.TableID).To(BeEquivalentTo(1))
	Expect(vppMsg.IsIPv6).To(BeEquivalentTo(0))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(0))
	Expect(vppMsg.Name).To(BeEquivalentTo([]byte("table1")))

	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	err = vtHandler.DelVrfTable(vrfTables[1])
	Expect(err).To(Succeed())

	vppMsg, ok = ctx.MockChannel.Msg.(*ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.TableID).To(BeEquivalentTo(1))
	Expect(vppMsg.IsIPv6).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(0))
	Expect(vppMsg.Name).To(BeEquivalentTo([]byte("table1")))

	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	err = vtHandler.DelVrfTable(vrfTables[2])
	Expect(err).To(Succeed())

	vppMsg, ok = ctx.MockChannel.Msg.(*ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.TableID).To(BeEquivalentTo(2))
	Expect(vppMsg.IsIPv6).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(0))
	Expect(vppMsg.Name).To(BeEquivalentTo([]byte("table2")))

	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{Retval: 1})
	err = vtHandler.DelVrfTable(vrfTables[0])
	Expect(err).To(Not(BeNil()))
}

// Test VRF flow hash settings
func TestVrfFlowHashSettings(t *testing.T) {
	ctx, vtHandler := vrfTableTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.SetIPFlowHashReply{})
	err := vtHandler.SetVrfFlowHashSettings(5, true,
		&l3.VrfTable_FlowHashSettings{
			UseSrcIp:   true,
			UseSrcPort: true,
			Symmetric:  true,
		})
	Expect(err).To(Succeed())

	vppMsg, ok := ctx.MockChannel.Msg.(*ip.SetIPFlowHash)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.VrfID).To(BeEquivalentTo(5))
	Expect(vppMsg.IsIPv6).To(BeEquivalentTo(1))
	Expect(vppMsg.Src).To(BeEquivalentTo(1))
	Expect(vppMsg.Dst).To(BeEquivalentTo(0))
	Expect(vppMsg.Sport).To(BeEquivalentTo(1))
	Expect(vppMsg.Dport).To(BeEquivalentTo(0))
	Expect(vppMsg.Proto).To(BeEquivalentTo(0))
	Expect(vppMsg.Symmetric).To(BeEquivalentTo(1))
	Expect(vppMsg.Reverse).To(BeEquivalentTo(0))
}

func vrfTableTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.VrfTableVppAPI) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	vtHandler := vpp1904.NewVrfTableVppHandler(ctx.MockChannel, log)
	return ctx, vtHandler
}
