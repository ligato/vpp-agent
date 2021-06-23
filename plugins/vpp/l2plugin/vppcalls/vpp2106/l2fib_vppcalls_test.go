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

package vpp2106_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ethernet_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	vpp_l2 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	vpp2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls/vpp2106"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"

	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

var testDataInFib = []*l2.FIBEntry{
	{PhysAddress: "FF:FF:FF:FF:FF:FF", BridgeDomain: "bd1", OutgoingInterface: "if1", Action: l2.FIBEntry_FORWARD, StaticConfig: true, BridgedVirtualInterface: true},
	{PhysAddress: "AA:AA:AA:AA:AA:AA", BridgeDomain: "bd1", OutgoingInterface: "if1", Action: l2.FIBEntry_FORWARD, StaticConfig: true},
	{PhysAddress: "BB:BB:BB:BB:BB:BB", BridgeDomain: "bd1", Action: l2.FIBEntry_DROP},
	{PhysAddress: "CC:CC:CC:CC:CC:CC", BridgeDomain: "bd1", OutgoingInterface: "if1", Action: l2.FIBEntry_FORWARD},
}

var testDatasOutFib = []*vpp_l2.L2fibAddDel{
	{BdID: 5, SwIfIndex: 55, BviMac: true, Mac: ethernet_types.MacAddress{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: true, FilterMac: false},
	{BdID: 5, SwIfIndex: 55, BviMac: false, Mac: ethernet_types.MacAddress{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA}, StaticMac: true, FilterMac: false},
	{BdID: 5, SwIfIndex: ^interface_types.InterfaceIndex(0), BviMac: false, Mac: ethernet_types.MacAddress{0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB}, StaticMac: false, FilterMac: true},
	{BdID: 5, SwIfIndex: 55, BviMac: false, Mac: ethernet_types.MacAddress{0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC}, StaticMac: false, FilterMac: false},
}

func TestL2FibAdd(t *testing.T) {
	ctx, fibHandler, ifaceIdx, bdIndexes := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifaceIdx.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 55})
	bdIndexes.Put("bd1", &idxvpp.OnlyIndex{Index: 5})

	for i := 0; i < len(testDataInFib); i++ {
		ctx.MockVpp.MockReply(&vpp_l2.L2fibAddDelReply{})
		err := fibHandler.AddL2FIB(testDataInFib[i])
		Expect(err).ShouldNot(HaveOccurred())
		testDatasOutFib[i].IsAdd = true
		Expect(ctx.MockChannel.Msg).To(Equal(testDatasOutFib[i]))
	}
}

func TestL2FibAddError(t *testing.T) {
	ctx, fibHandler, ifaceIdx, bdIndexes := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifaceIdx.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 55})
	bdIndexes.Put("bd1", &idxvpp.OnlyIndex{Index: 5})

	err := fibHandler.AddL2FIB(&l2.FIBEntry{PhysAddress: "not:mac:addr", BridgeDomain: "bd1", OutgoingInterface: "if1"})
	Expect(err).Should(HaveOccurred())

	ctx.MockVpp.MockReply(&vpp_l2.L2fibAddDelReply{Retval: 1})
	err = fibHandler.AddL2FIB(testDataInFib[0])
	Expect(err).Should(HaveOccurred())

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{})
	err = fibHandler.AddL2FIB(testDataInFib[0])
	Expect(err).Should(HaveOccurred())

	err = fibHandler.AddL2FIB(&l2.FIBEntry{PhysAddress: "CC:CC:CC:CC:CC:CC", BridgeDomain: "non-existing-bd", OutgoingInterface: "if1"})
	Expect(err).Should(HaveOccurred())

	err = fibHandler.AddL2FIB(&l2.FIBEntry{PhysAddress: "CC:CC:CC:CC:CC:CC", BridgeDomain: "bd1", OutgoingInterface: "non-existing-iface"})
	Expect(err).Should(HaveOccurred())
}

func TestL2FibDelete(t *testing.T) {
	ctx, fibHandler, ifaceIdx, bdIndexes := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifaceIdx.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 55})
	bdIndexes.Put("bd1", &idxvpp.OnlyIndex{Index: 5})

	for i := 0; i < len(testDataInFib); i++ {
		ctx.MockVpp.MockReply(&vpp_l2.L2fibAddDelReply{})
		err := fibHandler.DeleteL2FIB(testDataInFib[i])
		Expect(err).ShouldNot(HaveOccurred())
		testDatasOutFib[i].IsAdd = false
		Expect(ctx.MockChannel.Msg).To(Equal(testDatasOutFib[i]))
	}
}

func fibTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.FIBVppAPI, ifaceidx.IfaceMetadataIndexRW, idxvpp.NameToIndexRW) {
	ctx := vppmock.SetupTestCtx(t)
	logger := logrus.NewLogger("test-log")
	ifaceIdx := ifaceidx.NewIfaceIndex(logger, "fib-if-idx")
	bdIndexes := idxvpp.NewNameToIndex(logger, "fib-bd-idx", nil)
	fibHandler := vpp2106.NewL2VppHandler(ctx.MockChannel, ifaceIdx, bdIndexes, logger)
	return ctx, fibHandler, ifaceIdx, bdIndexes
}
