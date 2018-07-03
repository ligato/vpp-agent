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

package vppdump

import (
	"testing"

	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/l2"
	l2nb "github.com/ligato/vpp-agent/plugins/vpp/model/l2"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

var testDataInMessagesBDs = []govppapi.Message{
	&l2ba.BridgeDomainDetails{
		BdID:  4,
		Flood: 1, UuFlood: 1, Forward: 1, Learn: 1, ArpTerm: 1, MacAge: 140,
		SwIfDetails: []l2ba.BridgeDomainSwIf{
			{SwIfIndex: 5},
			{SwIfIndex: 7},
		},
	},
	&l2ba.BridgeDomainDetails{
		BdID:  5,
		Flood: 0, UuFlood: 0, Forward: 0, Learn: 0, ArpTerm: 0, MacAge: 141,
		SwIfDetails: []l2ba.BridgeDomainSwIf{
			{SwIfIndex: 5},
			{SwIfIndex: 8},
		},
	},
}

var testDataOutMessage = []*BridgeDomain{
	{
		Interfaces: []*BridgeDomainInterface{
			{SwIfIndex: 5},
			{SwIfIndex: 7},
		},
		BridgeDomains_BridgeDomain: l2nb.BridgeDomains_BridgeDomain{
			Flood:               true,
			UnknownUnicastFlood: true,
			Forward:             true,
			Learn:               true,
			ArpTermination:      true,
			MacAge:              140},
	}, {
		Interfaces: []*BridgeDomainInterface{
			{SwIfIndex: 5},
			{SwIfIndex: 8},
		},
		BridgeDomains_BridgeDomain: l2nb.BridgeDomains_BridgeDomain{
			Flood:               false,
			UnknownUnicastFlood: false,
			Forward:             false,
			Learn:               false,
			ArpTermination:      false,
			MacAge:              141},
	},
}

// Scenario:
// - 2 bridge domains + 1 default in VPP
// TestDumpBridgeDomainIDs tests DumpBridgeDomainIDs method
func TestDumpBridgeDomainIDs(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(testDataInMessagesBDs...)
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	activeDomains, err := DumpBridgeDomainIDs(ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(activeDomains).To(Equal([]uint32{0, 4, 5}))

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	_, err = DumpBridgeDomainIDs(ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())
}

// Scenario:
// - 2 bridge domains + 1 default in VPP
// TestDumpBridgeDomains tests DumpBridgeDomains method
func TestDumpBridgeDomains(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(testDataInMessagesBDs...)
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	bridgeDomains, err := DumpBridgeDomains(ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(bridgeDomains).To(HaveLen(2))
	Expect(bridgeDomains[4]).To(Equal(testDataOutMessage[0]))
	Expect(bridgeDomains[5]).To(Equal(testDataOutMessage[1]))

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	_, err = DumpBridgeDomains(ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())
}

var testDataInMessagesFIBs = []govppapi.Message{
	&l2ba.L2FibTableDetails{
		BdID:   10,
		Mac:    []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA},
		BviMac: 1, SwIfIndex: 1, FilterMac: 1, StaticMac: 1,
	},
	&l2ba.L2FibTableDetails{
		BdID:   20,
		Mac:    []byte{0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB},
		BviMac: 0, SwIfIndex: 2, FilterMac: 0, StaticMac: 0,
	},
}

var testDataOutFIBs = []*FIBTableEntry{
	{
		BridgeDomainIdx:          10,
		OutgoingInterfaceSwIfIdx: 1,
		FibTable_FibEntry: l2nb.FibTable_FibEntry{
			PhysAddress:             "aa:aa:aa:aa:aa:aa",
			Action:                  l2nb.FibTable_FibEntry_DROP,
			StaticConfig:            true,
			BridgedVirtualInterface: true,
		},
	},
	{
		BridgeDomainIdx:          20,
		OutgoingInterfaceSwIfIdx: 2,
		FibTable_FibEntry: l2nb.FibTable_FibEntry{
			PhysAddress:             "bb:bb:bb:bb:bb:bb",
			Action:                  l2nb.FibTable_FibEntry_FORWARD,
			StaticConfig:            false,
			BridgedVirtualInterface: false,
		},
	},
}

// Scenario:
// - 2 FIB entries in VPP
// TestDumpFIBTableEntries tests DumpFIBTableEntries method
func TestDumpFIBTableEntries(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(testDataInMessagesFIBs...)
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	fibTable, err := DumpFIBTableEntries(ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	Expect(fibTable).To(HaveLen(2))
	Expect(fibTable["aa:aa:aa:aa:aa:aa"]).To(Equal(testDataOutFIBs[0]))
	Expect(fibTable["bb:bb:bb:bb:bb:bb"]).To(Equal(testDataOutFIBs[1]))

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	_, err = DumpFIBTableEntries(ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())
}

var testDataInXConnect = []govppapi.Message{
	&l2ba.L2XconnectDetails{1, 2},
	&l2ba.L2XconnectDetails{3, 4},
}

var testDataOutXconnect = []*XConnectPairs{
	{1, 2},
	{3, 4},
}

// Scenario:
// - 2 Xconnect entries in VPP
// TestDumpXConnectPairs tests DumpXConnectPairs method
func TestDumpXConnectPairs(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(testDataInXConnect...)
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	xConnectPairs, err := DumpXConnectPairs(ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(xConnectPairs).To(HaveLen(2))
	Expect(xConnectPairs[1]).To(Equal(testDataOutXconnect[0]))
	Expect(xConnectPairs[3]).To(Equal(testDataOutXconnect[1]))

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	_, err = DumpXConnectPairs(ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())
}

// TestDumpL2 probably needs for run also running VPP
/*func DumpL2(t *testing.T) {
	// Connect to VPP.
	conn, err := govpp.Connect()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer conn.Disconnect()

	// Create an API channel that will be used in the examples.
	ch, err := conn.NewAPIChannel()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer ch.Close()

	res, err := DumpBridgeDomains(logrus.DefaultLogger(), ch, nil)
	fmt.Printf("%+v\n", res)

	res2, err := DumpFIBTableEntries(logrus.DefaultLogger(), ch, nil)
	fmt.Printf("%+v\n", res2)
	for _, fib := range res2 {
		fmt.Printf("%+v\n", fib)
	}

	res3, _ := DumpXConnectPairs(logrus.DefaultLogger(), ch, nil)
	fmt.Printf("%+v\n", res3)
	for _, xconn := range res3 {
		fmt.Printf("%+v\n", xconn)
	}
}*/
