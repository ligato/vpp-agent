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

package vpp2001_324_test

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
	vpp_ip "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/ip"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/l3plugin/vppcalls/vpp2001_324"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/vppcallmock"
	l3 "go.ligato.io/vpp-agent/v2/proto/ligato/vpp-agent/vpp/l3"
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

	ctx.MockVpp.MockReply(&vpp_ip.IPTableAddDelReply{})
	err := vtHandler.AddVrfTable(vrfTables[0])
	Expect(err).To(Succeed())

	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.Table.TableID).To(BeEquivalentTo(1))
	Expect(vppMsg.Table.IsIP6).To(BeEquivalentTo(0))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
	Expect(vppMsg.Table.Name).To(BeEquivalentTo([]byte("table1")))

	ctx.MockVpp.MockReply(&vpp_ip.IPTableAddDelReply{})
	err = vtHandler.AddVrfTable(vrfTables[1])
	Expect(err).To(Succeed())

	vppMsg, ok = ctx.MockChannel.Msg.(*vpp_ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.Table.TableID).To(BeEquivalentTo(1))
	Expect(vppMsg.Table.IsIP6).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
	Expect(vppMsg.Table.Name).To(BeEquivalentTo([]byte("table1")))

	ctx.MockVpp.MockReply(&vpp_ip.IPTableAddDelReply{})
	err = vtHandler.AddVrfTable(vrfTables[2])
	Expect(err).To(Succeed())

	vppMsg, ok = ctx.MockChannel.Msg.(*vpp_ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.Table.TableID).To(BeEquivalentTo(2))
	Expect(vppMsg.Table.IsIP6).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
	Expect(vppMsg.Table.Name).To(BeEquivalentTo([]byte("table2")))

	ctx.MockVpp.MockReply(&vpp_ip.IPTableAddDelReply{Retval: 1})
	err = vtHandler.AddVrfTable(vrfTables[0])
	Expect(err).To(Not(BeNil()))
}

// Test deleting routes
func TestDeleteVrfTable(t *testing.T) {
	ctx, vtHandler := vrfTableTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPTableAddDelReply{})
	err := vtHandler.DelVrfTable(vrfTables[0])
	Expect(err).To(Succeed())

	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.Table.TableID).To(BeEquivalentTo(1))
	Expect(vppMsg.Table.IsIP6).To(BeEquivalentTo(0))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(0))
	Expect(vppMsg.Table.Name).To(BeEquivalentTo([]byte("table1")))

	ctx.MockVpp.MockReply(&vpp_ip.IPTableAddDelReply{})
	err = vtHandler.DelVrfTable(vrfTables[1])
	Expect(err).To(Succeed())

	vppMsg, ok = ctx.MockChannel.Msg.(*vpp_ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.Table.TableID).To(BeEquivalentTo(1))
	Expect(vppMsg.Table.IsIP6).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(0))
	Expect(vppMsg.Table.Name).To(BeEquivalentTo([]byte("table1")))

	ctx.MockVpp.MockReply(&vpp_ip.IPTableAddDelReply{})
	err = vtHandler.DelVrfTable(vrfTables[2])
	Expect(err).To(Succeed())

	vppMsg, ok = ctx.MockChannel.Msg.(*vpp_ip.IPTableAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.Table.TableID).To(BeEquivalentTo(2))
	Expect(vppMsg.Table.IsIP6).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(0))
	Expect(vppMsg.Table.Name).To(BeEquivalentTo([]byte("table2")))

	ctx.MockVpp.MockReply(&vpp_ip.IPTableAddDelReply{Retval: 1})
	err = vtHandler.DelVrfTable(vrfTables[0])
	Expect(err).To(Not(BeNil()))
}

func vrfTableTestSetup(t *testing.T) (*vppcallmock.TestCtx, vppcalls.VrfTableVppAPI) {
	ctx := vppcallmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	vtHandler := vpp2001_324.NewVrfTableVppHandler(ctx.MockChannel, log)
	return ctx, vtHandler
}
