// Copyright (c) 2017 Cisco and/or its affiliates.
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

	"github.com/ligato/cn-infra/logging/logrus"

	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/l2"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/idxvpp2"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

var testDataInFib = []struct {
	mac    string
	bdID   uint32
	ifIdx  uint32
	bvi    bool
	static bool
}{
	{"FF:FF:FF:FF:FF:FF", 5, 55, true, true},
	{"FF:FF:FF:FF:FF:FF", 5, 55, false, true},
	{"FF:FF:FF:FF:FF:FF", 5, 55, true, false},
	{"FF:FF:FF:FF:FF:FF", 5, 55, false, false},
}

var createTestDatasOutFib = []*l2ba.L2fibAddDel{
	{BdID: 5, IsAdd: 1, SwIfIndex: 55, BviMac: 1, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: 1},
	{BdID: 5, IsAdd: 1, SwIfIndex: 55, BviMac: 0, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: 1},
	{BdID: 5, IsAdd: 1, SwIfIndex: 55, BviMac: 1, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: 0},
	{BdID: 5, IsAdd: 1, SwIfIndex: 55, BviMac: 0, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: 0},
}

var deleteTestDataOutFib = &l2ba.L2fibAddDel{
	BdID: 5, IsAdd: 0, SwIfIndex: 55, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
}

func TestL2FibAdd(t *testing.T) {
	ctx, fibHandler, _, _ := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	for i := 0; i < len(testDataInFib); i++ {
		ctx.MockVpp.MockReply(&l2ba.L2fibAddDelReply{})
		err := fibHandler.AddL2FIB(testDataInFib[i].mac, testDataInFib[i].bdID, testDataInFib[i].ifIdx,
			testDataInFib[i].bvi, testDataInFib[i].static)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ctx.MockChannel.Msg).To(Equal(createTestDatasOutFib[i]))
	}
}

func TestL2FibAddError(t *testing.T) {
	ctx, fibHandler, _, _ := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	err := fibHandler.AddL2FIB("not:mac:addr", 4, 10, false, false)
	Expect(err).Should(HaveOccurred())

	ctx.MockVpp.MockReply(&l2ba.L2fibAddDelReply{Retval: 1})
	err = fibHandler.AddL2FIB("FF:FF:FF:FF:FF:FF", 4, 10, false, false)
	Expect(err).Should(HaveOccurred())

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	err = fibHandler.AddL2FIB("FF:FF:FF:FF:FF:FF", 4, 10, false, false)
	Expect(err).Should(HaveOccurred())
}

func TestL2FibDelete(t *testing.T) {
	ctx, fibHandler, _, _ := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	for i := 0; i < len(testDataInFib); i++ {
		ctx.MockVpp.MockReply(&l2ba.L2fibAddDelReply{})
		err := fibHandler.DeleteL2FIB(testDataInFib[i].mac, testDataInFib[i].bdID, testDataInFib[i].ifIdx)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ctx.MockChannel.Msg).To(Equal(deleteTestDataOutFib))
	}
}

func fibTestSetup(t *testing.T) (*vppcallmock.TestCtx, vppcalls.FIBVppAPI, ifaceidx.IfaceMetadataIndexRW, idxvpp2.NameToIndexRW) {
	ctx := vppcallmock.SetupTestCtx(t)
	logger := logrus.NewLogger("test-log")
	ifIndexes := ifaceidx.NewIfaceIndex(logger, "fib-if-idx")
	bdIndexes := idxvpp2.NewNameToIndex(logger, "fib-bd-idx", nil)
	fibHandler := vppcalls.NewFIBVppHandler(ctx.MockChannel, ifIndexes, bdIndexes, logger)
	return ctx, fibHandler, ifIndexes, bdIndexes
}
