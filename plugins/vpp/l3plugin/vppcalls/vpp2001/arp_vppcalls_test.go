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

package vpp2001_test

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"

	vpp_ip_neighbor "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_neighbor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls/vpp2001"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

var arpEntries = []*l3.ARPEntry{
	{
		Interface:   "if1",
		IpAddress:   "192.168.10.21",
		PhysAddress: "59:6C:45:59:8E:BD",
		Static:      true,
	},
	{
		Interface:   "if1",
		IpAddress:   "192.168.10.22",
		PhysAddress: "6C:45:59:59:8E:BD",
		Static:      false,
	},
	{
		Interface:   "if1",
		IpAddress:   "dead::1",
		PhysAddress: "8E:BD:6C:45:59:59",
		Static:      false,
	},
}

// Test adding of ARP
func TestAddArp(t *testing.T) {
	ctx, ifIndexes, arpHandler := arpTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&vpp_ip_neighbor.IPNeighborAddDelReply{})
	err := arpHandler.VppAddArp(arpEntries[0])
	Expect(err).To(Succeed())
	ctx.MockVpp.MockReply(&vpp_ip_neighbor.IPNeighborAddDelReply{})
	err = arpHandler.VppAddArp(arpEntries[1])
	Expect(err).To(Succeed())
	ctx.MockVpp.MockReply(&vpp_ip_neighbor.IPNeighborAddDelReply{})
	err = arpHandler.VppAddArp(arpEntries[2])
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&vpp_ip_neighbor.IPNeighborAddDelReply{Retval: 1})
	err = arpHandler.VppAddArp(arpEntries[0])
	Expect(err).NotTo(BeNil())
}

// Test deleting of ARP
func TestDelArp(t *testing.T) {
	ctx, ifIndexes, arpHandler := arpTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&vpp_ip_neighbor.IPNeighborAddDelReply{})
	err := arpHandler.VppDelArp(arpEntries[0])
	Expect(err).To(Succeed())
}

func arpTestSetup(t *testing.T) (*vppmock.TestCtx, ifaceidx.IfaceMetadataIndexRW, vppcalls.ArpVppAPI) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test"), "test")
	arpHandler := vpp2001.NewArpVppHandler(ctx.MockChannel, ifIndexes, log)
	return ctx, ifIndexes, arpHandler
}
