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

package vpp2101_test

import (
	"testing"

	. "github.com/onsi/gomega"
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	vpp_l2 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	vpp2101 "go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls/vpp2101"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
)

var inTestDataXConnect = []struct {
	receiveIfaceIndex  string
	transmitIfaceIndex string
	message            govppapi.Message
}{
	{"rxIf1", "txIf1", &vpp_l2.SwInterfaceSetL2XconnectReply{}},
	{"rxIf2", "txIf2", &vpp_l2.SwInterfaceSetL2XconnectReply{Retval: 1}},
	{"rxIf2", "txIf2", &vpp_l2.BridgeDomainAddDelReply{}},
}

var outTestDataXConnect = []struct {
	outData    *vpp_l2.SwInterfaceSetL2Xconnect
	isResultOk bool
}{
	{&vpp_l2.SwInterfaceSetL2Xconnect{
		RxSwIfIndex: 100,
		TxSwIfIndex: 200,
	}, true},
	{&vpp_l2.SwInterfaceSetL2Xconnect{
		RxSwIfIndex: 101,
		TxSwIfIndex: 201,
	}, false},
	{&vpp_l2.SwInterfaceSetL2Xconnect{
		RxSwIfIndex: 101,
		TxSwIfIndex: 201,
	}, false},
}

/**
scenarios:
- enabling xconnect
	- ok
	- retvalue != 0
	- returned VPP message != what is expected
*/
// TestVppSetL2XConnect tests VppSetL2XConnect method
func TestVppSetL2XConnect(t *testing.T) {
	ctx, xcHandler, ifaceIdx := xcTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifaceIdx.Put("rxIf1", &ifaceidx.IfaceMetadata{SwIfIndex: 100})
	ifaceIdx.Put("rxIf2", &ifaceidx.IfaceMetadata{SwIfIndex: 101})
	ifaceIdx.Put("txIf1", &ifaceidx.IfaceMetadata{SwIfIndex: 200})
	ifaceIdx.Put("txIf2", &ifaceidx.IfaceMetadata{SwIfIndex: 201})

	for i := 0; i < len(inTestDataXConnect); i++ {
		ctx.MockVpp.MockReply(inTestDataXConnect[i].message)
		err := xcHandler.AddL2XConnect(inTestDataXConnect[i].receiveIfaceIndex,
			inTestDataXConnect[i].transmitIfaceIndex)

		if outTestDataXConnect[i].isResultOk {
			Expect(err).To(BeNil())
		} else {
			Expect(err).NotTo(BeNil())
		}
		outTestDataXConnect[i].outData.Enable = true
		Expect(ctx.MockChannel.Msg).To(Equal(outTestDataXConnect[i].outData))
	}
}

/**
scenarios:
- disabling xconnect
	- ok
	- retvalue != 0
	- returned VPP message != what is expected
*/
// TestVppUnsetL2XConnect tests VppUnsetL2XConnect method
func TestVppUnsetL2XConnect(t *testing.T) {
	ctx, xcHandler, ifaceIdx := xcTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifaceIdx.Put("rxIf1", &ifaceidx.IfaceMetadata{SwIfIndex: 100})
	ifaceIdx.Put("rxIf2", &ifaceidx.IfaceMetadata{SwIfIndex: 101})
	ifaceIdx.Put("txIf1", &ifaceidx.IfaceMetadata{SwIfIndex: 200})
	ifaceIdx.Put("txIf2", &ifaceidx.IfaceMetadata{SwIfIndex: 201})

	for i := 0; i < len(inTestDataXConnect); i++ {
		ctx.MockVpp.MockReply(inTestDataXConnect[i].message)
		err := xcHandler.DeleteL2XConnect(inTestDataXConnect[i].receiveIfaceIndex,
			inTestDataXConnect[i].transmitIfaceIndex)

		if outTestDataXConnect[i].isResultOk {
			Expect(err).To(BeNil())
		} else {
			Expect(err).NotTo(BeNil())
		}
		outTestDataXConnect[i].outData.Enable = false
		Expect(ctx.MockChannel.Msg).To(Equal(outTestDataXConnect[i].outData))
	}
}

func xcTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.XConnectVppAPI, ifaceidx.IfaceMetadataIndexRW) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	ifaceIdx := ifaceidx.NewIfaceIndex(log, "xc-if-idx")
	xcHandler := vpp2101.NewL2VppHandler(ctx.MockChannel, ifaceIdx, nil, log)
	return ctx, xcHandler, ifaceIdx
}
