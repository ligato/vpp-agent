//  Copyright (c) 2021 Cisco and/or its affiliates.
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
	govppapi "go.fd.io/govpp/api"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ethernet_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	vpp_l2 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

var testDataInMessagesBDs = []govppapi.Message{
	&vpp_l2.BridgeDomainDetails{
		BdID:  4,
		Flood: true, UuFlood: true, Forward: true, Learn: true, ArpTerm: true, MacAge: 140,
		SwIfDetails: []vpp_l2.BridgeDomainSwIf{
			{SwIfIndex: 5},
			{SwIfIndex: 7},
		},
	},
	&vpp_l2.BridgeDomainDetails{
		BdID:  5,
		Flood: false, UuFlood: false, Forward: false, Learn: false, ArpTerm: false, MacAge: 141,
		SwIfDetails: []vpp_l2.BridgeDomainSwIf{
			{SwIfIndex: 5},
			{SwIfIndex: 8},
		},
	},
}

var testDataOutMessage = []*vppcalls.BridgeDomainDetails{
	{
		Bd: &l2.BridgeDomain{
			Flood:               true,
			UnknownUnicastFlood: true,
			Forward:             true,
			Learn:               true,
			ArpTermination:      true,
			MacAge:              140,
			Interfaces: []*l2.BridgeDomain_Interface{
				{
					Name: "if1",
				},
				{
					Name: "if2",
				},
			},
		},
		Meta: &vppcalls.BridgeDomainMeta{
			BdID: 4,
		},
	}, {
		Bd: &l2.BridgeDomain{
			Flood:               false,
			UnknownUnicastFlood: false,
			Forward:             false,
			Learn:               false,
			ArpTermination:      false,
			MacAge:              141,
			Interfaces: []*l2.BridgeDomain_Interface{
				{
					Name: "if1",
				},
				{
					Name: "if3",
				},
			},
			ArpTerminationTable: []*l2.BridgeDomain_ArpTerminationEntry{
				{
					IpAddress:   "192.168.0.1",
					PhysAddress: "aa:aa:aa:aa:aa:aa",
				},
			},
		},
		Meta: &vppcalls.BridgeDomainMeta{
			BdID: 5,
		},
	},
}

// TestDumpBridgeDomains tests DumpBridgeDomains method
func TestDumpBridgeDomains(t *testing.T) {
	ctx, bdHandler, ifIndexes := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 5})
	ifIndexes.Put("if2", &ifaceidx.IfaceMetadata{SwIfIndex: 7})

	ctx.MockReplies([]*vppmock.HandleReplies{
		{
			Name:    (&vpp_l2.BdIPMacDump{}).GetMessageName(),
			Ping:    true,
			Message: &vpp_l2.BdIPMacDetails{},
		},
		{
			Name:    (&vpp_l2.BridgeDomainDump{}).GetMessageName(),
			Ping:    true,
			Message: testDataInMessagesBDs[0],
		},
	})

	bridgeDomains, err := bdHandler.DumpBridgeDomains()

	Expect(err).To(BeNil())
	Expect(bridgeDomains).To(HaveLen(1))
	Expect(bridgeDomains[0]).To(Equal(testDataOutMessage[0]))

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{})
	_, err = bdHandler.DumpBridgeDomains()
	Expect(err).Should(HaveOccurred())
}

// TestDumpBridgeDomains tests DumpBridgeDomains method
func TestDumpBridgeDomainsWithARP(t *testing.T) {
	ctx, bdHandler, ifIndexes := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 5})
	ifIndexes.Put("if3", &ifaceidx.IfaceMetadata{SwIfIndex: 8})

	ctx.MockReplies([]*vppmock.HandleReplies{
		{
			Name: (&vpp_l2.BdIPMacDump{}).GetMessageName(),
			Ping: true,
			Message: &vpp_l2.BdIPMacDetails{
				Entry: vpp_l2.BdIPMac{
					BdID: 5,
					IP: ip_types.Address{
						Af: ip_types.ADDRESS_IP4,
						Un: ip_types.AddressUnionIP4(
							ip_types.IP4Address{192, 168, 0, 1},
						),
					},
					Mac: ethernet_types.MacAddress{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA},
				},
			},
		},
		{
			Name:    (&vpp_l2.BridgeDomainDump{}).GetMessageName(),
			Ping:    true,
			Message: testDataInMessagesBDs[1],
		},
	})

	bridgeDomains, err := bdHandler.DumpBridgeDomains()

	Expect(err).To(BeNil())
	Expect(bridgeDomains).To(HaveLen(1))
	Expect(bridgeDomains[0]).To(Equal(testDataOutMessage[1]))

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{})
	_, err = bdHandler.DumpBridgeDomains()
	Expect(err).Should(HaveOccurred())
}

var testDataInMessagesFIBs = []govppapi.Message{
	&vpp_l2.L2FibTableDetails{
		BdID:   10,
		Mac:    ethernet_types.MacAddress{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA},
		BviMac: false, SwIfIndex: ^interface_types.InterfaceIndex(0), FilterMac: true, StaticMac: false,
	},
	&vpp_l2.L2FibTableDetails{
		BdID:   20,
		Mac:    ethernet_types.MacAddress{0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB},
		BviMac: true, SwIfIndex: 1, FilterMac: false, StaticMac: true,
	},
}

var testDataOutFIBs = []*vppcalls.FibTableDetails{
	{
		Fib: &l2.FIBEntry{
			PhysAddress:             "aa:aa:aa:aa:aa:aa",
			BridgeDomain:            "bd1",
			Action:                  l2.FIBEntry_DROP,
			StaticConfig:            false,
			BridgedVirtualInterface: false,
			OutgoingInterface:       "",
		},
		Meta: &vppcalls.FibMeta{
			BdID:  10,
			IfIdx: ^uint32(0),
		},
	},
	{
		Fib: &l2.FIBEntry{
			PhysAddress:             "bb:bb:bb:bb:bb:bb",
			BridgeDomain:            "bd2",
			Action:                  l2.FIBEntry_FORWARD,
			StaticConfig:            true,
			BridgedVirtualInterface: true,
			OutgoingInterface:       "if1",
		},
		Meta: &vppcalls.FibMeta{
			BdID:  20,
			IfIdx: 1,
		},
	},
}

// Scenario:
// - 2 FIB entries in VPP
// TestDumpFIBTableEntries tests DumpFIBTableEntries method
func TestDumpFIBTableEntries(t *testing.T) {
	ctx, fibHandler, ifIndexes, bdIndexes := fibTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})
	bdIndexes.Put("bd1", &idxvpp.OnlyIndex{Index: 10})
	bdIndexes.Put("bd2", &idxvpp.OnlyIndex{Index: 20})

	ctx.MockVpp.MockReply(testDataInMessagesFIBs...)
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	fibTable, err := fibHandler.DumpL2FIBs()
	Expect(err).To(BeNil())
	Expect(fibTable).To(HaveLen(2))
	Expect(fibTable["aa:aa:aa:aa:aa:aa"]).To(Equal(testDataOutFIBs[0]))
	Expect(fibTable["bb:bb:bb:bb:bb:bb"]).To(Equal(testDataOutFIBs[1]))

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{})
	_, err = fibHandler.DumpL2FIBs()
	Expect(err).Should(HaveOccurred())
}

var testDataInXConnect = []govppapi.Message{
	&vpp_l2.L2XconnectDetails{
		RxSwIfIndex: 1,
		TxSwIfIndex: 2,
	},
	&vpp_l2.L2XconnectDetails{
		RxSwIfIndex: 3,
		TxSwIfIndex: 4,
	},
}

var testDataOutXconnect = []*vppcalls.XConnectDetails{
	{
		Xc: &l2.XConnectPair{
			ReceiveInterface:  "if1",
			TransmitInterface: "if2",
		},
		Meta: &vppcalls.XcMeta{
			ReceiveInterfaceSwIfIdx:  1,
			TransmitInterfaceSwIfIdx: 2,
		},
	},
	{
		Xc: &l2.XConnectPair{
			ReceiveInterface:  "if3",
			TransmitInterface: "if4",
		},
		Meta: &vppcalls.XcMeta{
			ReceiveInterfaceSwIfIdx:  3,
			TransmitInterfaceSwIfIdx: 4,
		},
	},
}

/*
TODO: re-enable the test once l2_xconnect_dump is fixed in VPP (it crashes in 21.01, see https://jira.fd.io/browse/VPP-1968)
// Scenario:
// - 2 Xconnect entries in VPP
// TestDumpXConnectPairs tests DumpXConnectPairs method
func TestDumpXConnectPairs(t *testing.T) {
	ctx, xcHandler, ifIndex := xcTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndex.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})
	ifIndex.Put("if2", &ifaceidx.IfaceMetadata{SwIfIndex: 2})
	ifIndex.Put("if3", &ifaceidx.IfaceMetadata{SwIfIndex: 3})
	ifIndex.Put("if4", &ifaceidx.IfaceMetadata{SwIfIndex: 4})

	ctx.MockVpp.MockReply(testDataInXConnect...)
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	xConnectPairs, err := xcHandler.DumpXConnectPairs()

	Expect(err).To(BeNil())
	Expect(xConnectPairs).To(HaveLen(2))
	Expect(xConnectPairs[1]).To(Equal(testDataOutXconnect[0]))
	Expect(xConnectPairs[3]).To(Equal(testDataOutXconnect[1]))

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{})
	_, err = xcHandler.DumpXConnectPairs()
	Expect(err).Should(HaveOccurred())
}
*/
