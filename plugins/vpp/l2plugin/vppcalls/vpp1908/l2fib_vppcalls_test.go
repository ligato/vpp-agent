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

package vpp1908_test

import (
	"testing"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls/vpp1908"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	l2ba "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	l2nb "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

var testDataInFib = []*l2nb.FIBEntry{
	{PhysAddress: "FF:FF:FF:FF:FF:FF", BridgeDomain: "bd1", OutgoingInterface: "if1", Action: l2nb.FIBEntry_FORWARD, StaticConfig: true, BridgedVirtualInterface: true},
	{PhysAddress: "AA:AA:AA:AA:AA:AA", BridgeDomain: "bd1", OutgoingInterface: "if1", Action: l2nb.FIBEntry_FORWARD, StaticConfig: true},
	{PhysAddress: "BB:BB:BB:BB:BB:BB", BridgeDomain: "bd1", Action: l2nb.FIBEntry_DROP},
	{PhysAddress: "CC:CC:CC:CC:CC:CC", BridgeDomain: "bd1", OutgoingInterface: "if1", Action: l2nb.FIBEntry_FORWARD},
}

var testDatasOutFib = []*l2ba.L2fibAddDel{
	{BdID: 5, SwIfIndex: 55, BviMac: 1, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: 1, FilterMac: 0},
	{BdID: 5, SwIfIndex: 55, BviMac: 0, Mac: []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA}, StaticMac: 1, FilterMac: 0},
	{BdID: 5, SwIfIndex: ^uint32(0), BviMac: 0, Mac: []byte{0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB}, StaticMac: 0, FilterMac: 1},
	{BdID: 5, SwIfIndex: 55, BviMac: 0, Mac: []byte{0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC}, StaticMac: 0, FilterMac: 0},
}

func TestL2FibAdd(t *testing.T) {
	ctx, fibHandler, ifaceIdx, bdIndexes := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifaceIdx.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 55})
	bdIndexes.Put("bd1", &idxvpp.OnlyIndex{Index: 5})

	for i := 0; i < len(testDataInFib); i++ {
		ctx.MockVpp.MockReply(&l2ba.L2fibAddDelReply{})
		err := fibHandler.AddL2FIB(testDataInFib[i])
		Expect(err).ShouldNot(HaveOccurred())
		testDatasOutFib[i].IsAdd = 1
		Expect(ctx.MockChannel.Msg).To(Equal(testDatasOutFib[i]))
	}
}

func TestL2FibAddError(t *testing.T) {
	ctx, fibHandler, ifaceIdx, bdIndexes := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifaceIdx.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 55})
	bdIndexes.Put("bd1", &idxvpp.OnlyIndex{Index: 5})

	err := fibHandler.AddL2FIB(&l2nb.FIBEntry{PhysAddress: "not:mac:addr", BridgeDomain: "bd1", OutgoingInterface: "if1"})
	Expect(err).Should(HaveOccurred())

	ctx.MockVpp.MockReply(&l2ba.L2fibAddDelReply{Retval: 1})
	err = fibHandler.AddL2FIB(testDataInFib[0])
	Expect(err).Should(HaveOccurred())

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	err = fibHandler.AddL2FIB(testDataInFib[0])
	Expect(err).Should(HaveOccurred())

	err = fibHandler.AddL2FIB(&l2nb.FIBEntry{PhysAddress: "CC:CC:CC:CC:CC:CC", BridgeDomain: "non-existing-bd", OutgoingInterface: "if1"})
	Expect(err).Should(HaveOccurred())

	err = fibHandler.AddL2FIB(&l2nb.FIBEntry{PhysAddress: "CC:CC:CC:CC:CC:CC", BridgeDomain: "bd1", OutgoingInterface: "non-existing-iface"})
	Expect(err).Should(HaveOccurred())
}

func TestL2FibDelete(t *testing.T) {
	ctx, fibHandler, ifaceIdx, bdIndexes := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifaceIdx.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 55})
	bdIndexes.Put("bd1", &idxvpp.OnlyIndex{Index: 5})

	for i := 0; i < len(testDataInFib); i++ {
		ctx.MockVpp.MockReply(&l2ba.L2fibAddDelReply{})
		err := fibHandler.DeleteL2FIB(testDataInFib[i])
		Expect(err).ShouldNot(HaveOccurred())
		testDatasOutFib[i].IsAdd = 0
		Expect(ctx.MockChannel.Msg).To(Equal(testDatasOutFib[i]))
	}
}

func fibTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.FIBVppAPI, ifaceidx.IfaceMetadataIndexRW, idxvpp.NameToIndexRW) {
	ctx := vppmock.SetupTestCtx(t)
	logger := logrus.NewLogger("test-log")
	ifaceIdx := ifaceidx.NewIfaceIndex(logger, "fib-if-idx")
	bdIndexes := idxvpp.NewNameToIndex(logger, "fib-bd-idx", nil)
	fibHandler := vpp1908.NewL2VppHandler(ctx.MockChannel, ifaceIdx, bdIndexes, logger)
	return ctx, fibHandler, ifaceIdx, bdIndexes
}
