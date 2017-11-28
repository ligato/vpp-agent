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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/logrus"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls/test/impl"
	. "github.com/onsi/gomega"
)

const (
	dummyBridgeDomain uint32 = 4
	dummyMACAddress          = "FF:FF:FF:FF:FF:FF"
	dummyIPAddress           = "192.168.4.4"
	dummyLoggerName          = "dummyLogger"
)

//TestVppAddArpTerminationTableEntry tests VppAddArpTerminationTableEntry
func TestVppAddArpTerminationTableEntry(t *testing.T) {
	RegisterTestingT(t)
	mockedChannel := &impl.MockedChannel{Channel: govppapi.Channel{}}
	vppcalls.VppAddArpTerminationTableEntry(dummyBridgeDomain, dummyMACAddress, dummyIPAddress, logrus.NewLogger(dummyLoggerName),
		mockedChannel, nil)

	vppMsg, ok := mockedChannel.Msg.(*l2ba.BdIPMacAddDel)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	//check values which will be send to VPP
	Expect(vppMsg.BdID).To(Equal(dummyBridgeDomain))
	Expect(vppMsg.MacAddress).To(Equal([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}))
	Expect(vppMsg.IPAddress).To(Equal([]byte{192, 168, 4, 4}))
	Expect(vppMsg.IsIpv6).To(Equal(uint8(0)))
	Expect(vppMsg.IsAdd).To(Equal(uint8(1)))
}

// TestVppRemoveArpTerminationTableEntry tests VppRemoveArpTerminationTableEntry method
func TestVppRemoveArpTerminationTableEntry(t *testing.T) {
	RegisterTestingT(t)
	mockedChannel := &impl.MockedChannel{Channel: govppapi.Channel{}}
	vppcalls.VppRemoveArpTerminationTableEntry(dummyBridgeDomain, dummyMACAddress, dummyIPAddress, logrus.NewLogger(dummyLoggerName),
		mockedChannel, nil)
	vppMsg, ok := mockedChannel.Msg.(*l2ba.BdIPMacAddDel)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	//check values which will be send to VPP
	Expect(vppMsg.BdID).To(Equal(dummyBridgeDomain))
	Expect(vppMsg.MacAddress).To(Equal([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}))
	Expect(vppMsg.IPAddress).To(Equal([]byte{192, 168, 4, 4}))
	Expect(vppMsg.IsIpv6).To(Equal(uint8(0)))
	Expect(vppMsg.IsAdd).To(Equal(uint8(0)))
}
