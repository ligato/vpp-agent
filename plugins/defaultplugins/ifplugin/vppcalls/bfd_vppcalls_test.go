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
	"fmt"
	"testing"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	bfd_model "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

const (
	dummyTextSourceAddress             = "192.168.1.1"
	dummyTextDestinationAddress        = "10.10.10.5"
	dummyDesiredMinTx           uint32 = 0
	dummyDetectMultiplier              = 0
	dummyRequiredMinRx                 = 0
	dummyLoggerName                    = "dummyLogger"
	dummyPluginName                    = "dummyPluginName"
	dummyFunctionName                  = "dummyFunctionName"
)

var testDataInDummySwIfIndex = initSwIfIndex().(ifaceidx.SwIfIndexRW)
var testDataInDummyBfdIndex = initBfdKeyIndex().(idxvpp.NameToIdxRW)

var dummyAddBfdUDPSession = &bfd_model.SingleHopBFD_Session{
	Interface:             dummyHostIFName,
	Authentication:        &bfd_model.SingleHopBFD_Session_Authentication{KeyId: 42, AdvertisedKeyId: 1},
	DesiredMinTxInterval:  dummyDesiredMinTx,
	DestinationAddress:    dummyTextDestinationAddress,
	DetectMultiplier:      dummyDetectMultiplier,
	Enabled:               true,
	RequiredMinRxInterval: dummyRequiredMinRx,
	SourceAddress:         dummyTextSourceAddress,
}

var dummyAddBfdEcho = &bfd_model.SingleHopBFD_EchoFunction{
	Name:                dummyFunctionName,
	EchoSourceInterface: dummyHostIFName,
}

var dummyAddBfdUDPSessionFromDetails = &bfd.BfdUDPSessionDetails{
	SwIfIndex:       dummyInterfaceIndex,
	BfdKeyID:        42,
	ConfKeyID:       1,
	DesiredMinTx:    dummyDesiredMinTx,
	DetectMult:      dummyDetectMultiplier,
	IsAuthenticated: 1,
	IsIpv6:          0,
	LocalAddr:       []byte{192, 168, 1, 1},
	PeerAddr:        []byte{10, 10, 10, 5},
	RequiredMinRx:   dummyRequiredMinRx,
	State:           0,
}

var testDataAddBfdUDPSession = &bfd.BfdUDPAdd{
	SwIfIndex:       dummyInterfaceIndex,
	BfdKeyID:        42,
	ConfKeyID:       1,
	DesiredMinTx:    dummyDesiredMinTx,
	DetectMult:      dummyDetectMultiplier,
	IsAuthenticated: 1,
	IsIpv6:          0,
	LocalAddr:       []byte{192, 168, 1, 1},
	PeerAddr:        []byte{10, 10, 10, 5},
	RequiredMinRx:   dummyRequiredMinRx,
}

var testDataSwIfIndex = &bfd.BfdUDPSetEchoSource{
	SwIfIndex: dummyInterfaceIndex,
}

func initSwIfIndex() interface{} {
	result := ifaceidx.NewSwIfIndex(
		nametoidx.NewNameToIdx(
			logrus.DefaultLogger(),
			core.PluginName(dummyPluginName),
			fmt.Sprintf("iface-cache-test"),
			nil),
	)
	result.RegisterName(dummyHostIFName, 42, nil)
	return result
}

func initBfdKeyIndex() interface{} {
	result := nametoidx.NewNameToIdx(
		logrus.DefaultLogger(),
		core.PluginName(dummyPluginName),
		fmt.Sprintf("bfd-index-test"),
		nil)
	result.RegisterName("*", 1, nil)
	return result
}

func TestAddBfdUDPSession(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSession(dummyAddBfdUDPSession, testDataInDummySwIfIndex, testDataInDummyBfdIndex, logrus.NewLogger(dummyLoggerName), ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd.BfdUDPAdd)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg).To(Equal(testDataAddBfdUDPSession))
}

func TestAddBfdUDPSessionFromDetails(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd.BfdUDPAddReply{})

	err := vppcalls.AddBfdUDPSessionFromDetails(dummyAddBfdUDPSessionFromDetails, testDataInDummyBfdIndex, logrus.NewLogger(dummyLoggerName), ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*bfd.BfdUDPAdd)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg).To(Equal(testDataAddBfdUDPSession))
}

func TestAddBfdEchoFunction(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bfd.BfdUDPAddReply{})

	err := vppcalls.AddBfdEchoFunction(dummyAddBfdEcho, testDataInDummySwIfIndex, ctx.MockChannel, nil)

	// Returns error on behalf: Bug 762
	// TODO: After bug fix, test need to be fixed too
	Expect(err).Should(HaveOccurred())

	vppMsg, ok := ctx.MockChannel.Msg.(*bfd.BfdUDPSetEchoSource)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg).To(Equal(testDataSwIfIndex))
}
