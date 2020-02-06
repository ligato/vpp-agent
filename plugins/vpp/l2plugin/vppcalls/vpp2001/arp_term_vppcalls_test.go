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

	. "github.com/onsi/gomega"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	vpp_l2 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/l2"
)

func TestVppAddArpTerminationTableEntry(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BdIPMacAddDelReply{})

	err := bdHandler.AddArpTerminationTableEntry(
		4, "FF:FF:FF:FF:FF:FF", "192.168.4.4")

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(&vpp_l2.BdIPMacAddDel{
		Entry: vpp_l2.BdIPMac{
			BdID: 4,
			IP: vpp_l2.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(
					vpp_l2.IP4Address{192, 168, 4, 4},
				),
			},
			Mac: vpp_l2.MacAddress{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		IsAdd: 1,
	}))
}

func TestVppAddArpTerminationTableEntryIPv6(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BdIPMacAddDelReply{})

	err := bdHandler.AddArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "2001:db9::54")

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(&vpp_l2.BdIPMacAddDel{
		Entry: vpp_l2.BdIPMac{
			BdID: 4,
			IP: vpp_l2.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ip_types.AddressUnionIP6(
					vpp_l2.IP6Address{32, 1, 13, 185, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 84},
				),
			},
			Mac: vpp_l2.MacAddress{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		IsAdd: 1,
	}))
}

func TestVppRemoveArpTerminationTableEntry(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BdIPMacAddDelReply{})

	err := bdHandler.RemoveArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "192.168.4.4")

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(&vpp_l2.BdIPMacAddDel{
		Entry: vpp_l2.BdIPMac{
			BdID: 4,
			IP: vpp_l2.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(
					vpp_l2.IP4Address{192, 168, 4, 4},
				),
			},
			Mac: vpp_l2.MacAddress{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		IsAdd: 0,
	}))
}

func TestVppArpTerminationTableEntryMacError(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BdIPMacAddDelReply{})

	err := bdHandler.AddArpTerminationTableEntry(4, "in:va:li:d:ma:c", "192.168.4.4")
	Expect(err).Should(HaveOccurred())

	err = bdHandler.RemoveArpTerminationTableEntry(4, "in:va:li:d:ma:c", "192.168.4.4")
	Expect(err).Should(HaveOccurred())
}

func TestVppArpTerminationTableEntryIpError(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BdIPMacAddDelReply{})

	err := bdHandler.AddArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "")
	Expect(err).Should(HaveOccurred())

	err = bdHandler.RemoveArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "")
	Expect(err).Should(HaveOccurred())
}

func TestVppArpTerminationTableEntryError(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BdIPMacAddDelReply{
		Retval: 1,
	})

	err := bdHandler.AddArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "192.168.4.4")
	Expect(err).Should(HaveOccurred())

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{})

	err = bdHandler.RemoveArpTerminationTableEntry(4, "FF:FF:FF:FF:FF:FF", "192.168.4.4")
	Expect(err).Should(HaveOccurred())
}
