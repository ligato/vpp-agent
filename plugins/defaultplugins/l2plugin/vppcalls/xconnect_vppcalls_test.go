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

package vppcalls

import (
	"testing"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/logrus"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/l2"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

var setTestDataInXConnect = []struct {
	receiveIfaceIndex  uint32
	transmitIfaceIndex uint32
	message            govppapi.Message
}{
	{100, 200, &l2ba.SwInterfaceSetL2XconnectReply{}},
	{100, 200, &l2ba.SwInterfaceSetL2XconnectReply{Retval: 1}},
	{100, 200, &l2ba.BridgeDomainAddDelReply{}},
}

var setTestDataOutXConnect = []struct {
	outData    *l2ba.SwInterfaceSetL2Xconnect
	isResultOk bool
}{
	{&l2ba.SwInterfaceSetL2Xconnect{100, 200, 1}, true},
	{&l2ba.SwInterfaceSetL2Xconnect{100, 200, 1}, false},
	{&l2ba.SwInterfaceSetL2Xconnect{100, 200, 1}, false},
}

/**
scenarios:
- enabling xconnect
	- ok
	- retvalue != 0
	- returned VPP message != what is expected
*/
//TestVppSetL2XConnect tests VppSetL2XConnect method
func TestVppSetL2XConnect(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	for i := 0; i < len(setTestDataInXConnect); i++ {
		ctx.MockVpp.MockReply(setTestDataInXConnect[i].message)
		err := VppSetL2XConnect(setTestDataInXConnect[i].receiveIfaceIndex,
			setTestDataInXConnect[i].transmitIfaceIndex, logrus.DefaultLogger(), ctx.MockChannel, nil)

		if setTestDataOutXConnect[i].isResultOk {
			Expect(err).To(BeNil())
		} else {
			Expect(err).NotTo(BeNil())
		}
		Expect(ctx.MockChannel.Msg).To(Equal(setTestDataOutXConnect[i].outData))
	}
}

var unsetTestDataInXConnect = []struct {
	receiveIfaceIndex  uint32
	transmitIfaceIndex uint32
	message            govppapi.Message
}{
	{100, 200, &l2ba.SwInterfaceSetL2XconnectReply{}},
	{100, 200, &l2ba.SwInterfaceSetL2XconnectReply{Retval: 1}},
	{100, 200, &l2ba.BridgeDomainAddDelReply{}},
}

var unsetTestDataOutXConnect = []struct {
	outData    *l2ba.SwInterfaceSetL2Xconnect
	isResultOk bool
}{
	{&l2ba.SwInterfaceSetL2Xconnect{100, 200, 0}, true},
	{&l2ba.SwInterfaceSetL2Xconnect{100, 200, 0}, false},
	{&l2ba.SwInterfaceSetL2Xconnect{100, 200, 0}, false},
}

/**
scenarios:
- enabling xconnect
	- ok
	- retvalue != 0
	- returned VPP message != what is expected
*/
//TestVppUnsetL2XConnect tests VppUnsetL2XConnect method
func TestVppUnsetL2XConnect(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	for i := 0; i < len(unsetTestDataInXConnect); i++ {
		ctx.MockVpp.MockReply(unsetTestDataInXConnect[i].message)
		err := VppUnsetL2XConnect(unsetTestDataInXConnect[i].receiveIfaceIndex,
			unsetTestDataInXConnect[i].transmitIfaceIndex, logrus.DefaultLogger(), ctx.MockChannel, nil)

		if unsetTestDataOutXConnect[i].isResultOk {
			Expect(err).To(BeNil())
		} else {
			Expect(err).NotTo(BeNil())
		}
		Expect(ctx.MockChannel.Msg).To(Equal(unsetTestDataOutXConnect[i].outData))
	}
}
