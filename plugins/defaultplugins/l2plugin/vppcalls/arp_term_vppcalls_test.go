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
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

const (
	dummyBridgeDomain uint32 = 4
	dummyMACAddress          = "FF:FF:FF:FF:FF:FF"
	dummyIPAddress           = "192.168.4.4"
	dummyLoggerName          = "dummyLogger"
)

var createTestDataOutArp = &l2ba.BdIPMacAddDel{
	BdID:       dummyBridgeDomain,
	IsAdd:      1,
	IsIpv6:     0,
	IPAddress:  []byte{192, 168, 4, 4},
	MacAddress: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
}

var deleteTestDataOutArp = &l2ba.BdIPMacAddDel{
	BdID:       dummyBridgeDomain,
	IsAdd:      0,
	IsIpv6:     0,
	IPAddress:  []byte{192, 168, 4, 4},
	MacAddress: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
}

//TestVppAddArpTerminationTableEntry tests VppAddArpTerminationTableEntry
func TestVppAddArpTerminationTableEntry(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})
	err := vppcalls.VppAddArpTerminationTableEntry(dummyBridgeDomain, dummyMACAddress, dummyIPAddress,
		logrus.NewLogger(dummyLoggerName), ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*l2ba.BdIPMacAddDel)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg).To(Equal(createTestDataOutArp))
}

// TestVppRemoveArpTerminationTableEntry tests VppRemoveArpTerminationTableEntry method
func TestVppRemoveArpTerminationTableEntry(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BdIPMacAddDelReply{})
	err := vppcalls.VppRemoveArpTerminationTableEntry(dummyBridgeDomain, dummyMACAddress, dummyIPAddress,
		logrus.NewLogger(dummyLoggerName), ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*l2ba.BdIPMacAddDel)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg).To(Equal(deleteTestDataOutArp))
}
