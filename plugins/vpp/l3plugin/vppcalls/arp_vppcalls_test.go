//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package vppcalls_test

import (
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
	"testing"
)

var arpEntries = []vppcalls.ArpEntry{
	{
		Interface:  1,
		IPAddress:  []byte{192, 168, 10, 21},
		MacAddress: []byte{0x59, 0x6C, 0xde, 0xad, 0x00, 0x01},
		Static:     true,
	},
	{
		Interface:  1,
		IPAddress:  []byte{192, 168, 10, 22},
		MacAddress: []byte{0x59, 0x6C, 0xde, 0xad, 0x00, 0x02},
		Static:     false,
	},
	{
		Interface:  1,
		IPAddress:  []byte{0xde, 0xad, 0, 0, 0, 0, 0, 0, 0xde, 0xad, 0, 0, 0, 0, 0, 1},
		MacAddress: []byte{0x59, 0x6C, 0xde, 0xad, 0x00, 0x02},
		Static:     false,
	},
}

// Test adding of ARP
func TestAddArp(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	err := vppcalls.VppAddArp(&arpEntries[0], ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	err = vppcalls.VppAddArp(&arpEntries[1], ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	err = vppcalls.VppAddArp(&arpEntries[2], ctx.MockChannel, nil)
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{Retval: 1})
	err = vppcalls.VppAddArp(&arpEntries[0], ctx.MockChannel, nil)
	Expect(err).To(Not(BeNil()))
}

// Test deleting of ARP
func TestDelArp(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPNeighborAddDelReply{})
	err := vppcalls.VppDelArp(&arpEntries[0], ctx.MockChannel, nil)
	Expect(err).To(Succeed())
}
