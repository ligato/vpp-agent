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
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

var dummyLogger = logrus.NewLogger("dummy")

func TestVppAddArpTerminationTableEntry(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := vppcalls.VppAddArpTerminationTableEntry(
		4, "FF:FF:FF:FF:FF:FF", "192.168.4.4",
		dummyLogger, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(&l2ba.BdIPMacAddDel{
		BdID:       4,
		IsAdd:      1,
		IsIpv6:     0,
		IPAddress:  []byte{192, 168, 4, 4},
		MacAddress: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}))
}

func TestVppAddArpTerminationTableEntryIPv6(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := vppcalls.VppAddArpTerminationTableEntry(
		4, "FF:FF:FF:FF:FF:FF", "2001:db9::54",
		dummyLogger, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(&l2ba.BdIPMacAddDel{
		BdID:       4,
		IsAdd:      1,
		IsIpv6:     1,
		IPAddress:  []byte{32, 1, 13, 185, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 84},
		MacAddress: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}))
}

func TestVppRemoveArpTerminationTableEntry(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := vppcalls.VppRemoveArpTerminationTableEntry(
		4, "FF:FF:FF:FF:FF:FF", "192.168.4.4",
		dummyLogger, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(&l2ba.BdIPMacAddDel{
		BdID:       4,
		IsAdd:      0,
		IsIpv6:     0,
		IPAddress:  []byte{192, 168, 4, 4},
		MacAddress: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}))
}

func TestVppArpTerminationTableEntryMacError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := vppcalls.VppAddArpTerminationTableEntry(
		4, "in:va:li:d:ma:c", "192.168.4.4",
		dummyLogger, ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())

	err = vppcalls.VppRemoveArpTerminationTableEntry(
		4, "in:va:li:d:ma:c", "192.168.4.4",
		dummyLogger, ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())
}

func TestVppArpTerminationTableEntryIpError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})

	err := vppcalls.VppAddArpTerminationTableEntry(
		4, "FF:FF:FF:FF:FF:FF", "",
		dummyLogger, ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())

	err = vppcalls.VppRemoveArpTerminationTableEntry(
		4, "FF:FF:FF:FF:FF:FF", "",
		dummyLogger, ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())
}

func TestVppArpTerminationTableEntryError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{
		Retval: 1,
	})

	err := vppcalls.VppAddArpTerminationTableEntry(
		4, "FF:FF:FF:FF:FF:FF", "192.168.4.4",
		dummyLogger, ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})

	err = vppcalls.VppRemoveArpTerminationTableEntry(
		4, "FF:FF:FF:FF:FF:FF", "192.168.4.4",
		dummyLogger, ctx.MockChannel, nil)
	Expect(err).Should(HaveOccurred())
}
