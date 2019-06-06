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

	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/l2"
	. "github.com/onsi/gomega"
)

func TestVppAddArpTerminationTableEntry(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := bdHandler.AddArpTerminationTableEntry(
		4, "FF:FF:FF:FF:FF:FF", "192.168.4.4")

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(&l2ba.BdIPMacAddDel{
		BdID:  4,
		IsAdd: 1,
		IP: l2ba.Address{
			Af: l2ba.ADDRESS_IP4,
			Un: l2ba.AddressUnionIP4(
				l2ba.IP4Address{192, 168, 4, 4},
			),
		},
		Mac: l2ba.MacAddress{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}))
}

func TestVppAddArpTerminationTableEntryIPv6(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := bdHandler.AddArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "2001:db9::54")

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(&l2ba.BdIPMacAddDel{
		BdID:  4,
		IsAdd: 1,
		IP: l2ba.Address{
			Af: l2ba.ADDRESS_IP6,
			Un: l2ba.AddressUnionIP6(
				l2ba.IP6Address{32, 1, 13, 185, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 84},
			),
		},
		Mac: l2ba.MacAddress{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}))
}

func TestVppRemoveArpTerminationTableEntry(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := bdHandler.RemoveArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "192.168.4.4")

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(&l2ba.BdIPMacAddDel{
		BdID:  4,
		IsAdd: 0,
		IP: l2ba.Address{
			Af: l2ba.ADDRESS_IP4,
			Un: l2ba.AddressUnionIP4(
				l2ba.IP4Address{192, 168, 4, 4},
			),
		},
		Mac: l2ba.MacAddress{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}))
}

func TestVppArpTerminationTableEntryMacError(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := bdHandler.AddArpTerminationTableEntry(4, "in:va:li:d:ma:c", "192.168.4.4")
	Expect(err).Should(HaveOccurred())

	err = bdHandler.RemoveArpTerminationTableEntry(4, "in:va:li:d:ma:c", "192.168.4.4")
	Expect(err).Should(HaveOccurred())
}

func TestVppArpTerminationTableEntryIpError(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := bdHandler.AddArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "")
	Expect(err).Should(HaveOccurred())

	err = bdHandler.RemoveArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "")
	Expect(err).Should(HaveOccurred())
}

func TestVppArpTerminationTableEntryError(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{
		Retval: 1,
	})

	err := bdHandler.AddArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "192.168.4.4")
	Expect(err).Should(HaveOccurred())

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})

	err = bdHandler.RemoveArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "192.168.4.4")
	Expect(err).Should(HaveOccurred())
}
