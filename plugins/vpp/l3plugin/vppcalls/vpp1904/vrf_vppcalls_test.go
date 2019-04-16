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

	"github.com/ligato/cn-infra/logging/logrus"
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls/vpp1904"
	"github.com/ligato/vpp-agent/plugins/vpp/vppcallmock"
	. "github.com/onsi/gomega"
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

func vrfTableTestSetup(t *testing.T) (*vppcallmock.TestCtx, vppcalls.VrfTableVppAPI) {
	ctx := vppcallmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	vtHandler := vpp1904.NewVrfTableVppHandler(ctx.MockChannel, log)
	return ctx, vtHandler
}