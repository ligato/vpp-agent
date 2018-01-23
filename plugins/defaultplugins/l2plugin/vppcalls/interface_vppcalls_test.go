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
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

const (
	ifaceA = "A"
	ifaceB = "B"
	ifaceC = "C"
	ifaceD = "D"
	ifaceE = "E"

	swIndexA uint32 = 1
	swIndexB uint32 = 2
	swIndexC uint32 = 3
	swIndexD uint32 = 4

	splitHorizonGroupA = 2104
	splitHorizonGroupB = 1903

	dummyPluginName = "dummy plugin name"
	dummyRetVal     = 4
)

var testDataInDummySwIfIndex = initSwIfIndex().(ifaceidx.SwIfIndexRW)

var testDataInBDIfaces = []*l2.BridgeDomains_BridgeDomain{
	{
		Name: dummyBridgeDomainName,
		Interfaces: []*l2.BridgeDomains_BridgeDomain_Interfaces{
			{Name: ifaceA, BridgedVirtualInterface: true, SplitHorizonGroup: splitHorizonGroupA},
			{Name: ifaceB, BridgedVirtualInterface: false, SplitHorizonGroup: splitHorizonGroupA},
			{Name: ifaceC, BridgedVirtualInterface: false, SplitHorizonGroup: splitHorizonGroupB},
			{Name: ifaceD, BridgedVirtualInterface: false, SplitHorizonGroup: splitHorizonGroupB},
			{Name: ifaceE, BridgedVirtualInterface: false, SplitHorizonGroup: splitHorizonGroupB}},
	},
	{
		Name: dummyBridgeDomainName,
	},
}

var testDataOutBDIfaces = []*l2ba.SwInterfaceSetL2Bridge{
	{
		BdID:        dummyBridgeDomain,
		RxSwIfIndex: swIndexA,
		Enable:      1,
		Bvi:         1,
	},
	{
		BdID:        dummyBridgeDomain,
		RxSwIfIndex: swIndexB,
		Enable:      1,
	},
	{
		BdID:        dummyBridgeDomain,
		RxSwIfIndex: swIndexA,
		Enable:      0,
	},
	{
		BdID:        dummyBridgeDomain,
		RxSwIfIndex: swIndexB,
		Enable:      0,
	},
}

/**
covers scenarios
- 5 provided interfaces - A..E
	- interface A - common interface
	- interface B - BVI interface
	- interface C - vpp binary call returns dummy ret value
	- interface D - vpp binary call returns incorrect return value
	- interface E - isn't specified sw index
*/
//TestVppSetAllInterfacesToBridgeDomainWithInterfaces tests method VppSetAllInterfacesToBridgeDomain
func TestVppSetAllInterfacesToBridgeDomainWithInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2ba.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2ba.SwInterfaceSetL2BridgeReply{Retval: dummyRetVal})
	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})

	//call testing method
	allBdInterfaces, configuredBdInterfaces, bviInterfaceName := vppcalls.VppSetAllInterfacesToBridgeDomain(
		testDataInBDIfaces[0], dummyBridgeDomain, testDataInDummySwIfIndex, logrus.DefaultLogger(),
		ctx.MockChannel, nil)

	//evaluate allBdInterfaces
	Expect(allBdInterfaces).To(HaveLen(3))
	Expect(allBdInterfaces).To(Equal([]string{ifaceA, ifaceB, ifaceE}))

	//Four VPP call - only two of them are successfull
	Expect(ctx.MockChannel.Msgs).To(HaveLen(4))
	Expect(ctx.MockChannel.Msgs[0]).To(Equal(testDataOutBDIfaces[0]))
	Expect(ctx.MockChannel.Msgs[1]).To(Equal(testDataOutBDIfaces[1]))

	//evaluate configuredBdInterfaces
	Expect(configuredBdInterfaces).To(HaveLen(2))
	Expect(configuredBdInterfaces).To(ContainElement(ifaceA))
	Expect(configuredBdInterfaces).To(ContainElement(ifaceB))
	Expect(configuredBdInterfaces).To(HaveLen(2))

	//evaluate bviInterfaceName
	Expect(bviInterfaceName).To(Equal(ifaceA))
}

/**
scenario
- input list of interfaces which should be added to BD is empty
*/
//TestVppSetAllInterfacesToBridgeDomainWithoutInterfaces tests method VppSetAllInterfacesToBridgeDomain
func TestVppSetAllInterfacesToBridgeDomainWithoutInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	//call testing method
	allBdInterfaces, configuredBdInterfaces, bviInterfaceName := vppcalls.VppSetAllInterfacesToBridgeDomain(
		testDataInBDIfaces[1], dummyBridgeDomain, testDataInDummySwIfIndex, logrus.DefaultLogger(),
		ctx.MockChannel, nil)

	//evaluate allBdInterfaces
	Expect(allBdInterfaces).To(BeEmpty())

	//evaluate configuredBdInterfaces
	Expect(configuredBdInterfaces).To(BeEmpty())

	//evaluate bviInterfaceName
	Expect(bviInterfaceName).To(Equal(""))
}

/**
covers scenarios
- 5 provided interfaces - A..E
	- interface A - common interface
	- interface B - common interface
	- interface C - vpp binary call returns dummy ret value
	- interface D - vpp binary call returns incorrect return value
	- interface E - isn't specified sw index
*/
//TestVppUnsetAllInterfacesFromBridgeDomain tests method VppUnsetAllInterfacesFromBridgeDomain
func TestVppUnsetAllInterfacesFromBridgeDomain(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2ba.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2ba.SwInterfaceSetL2BridgeReply{Retval: dummyRetVal})
	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})

	//call testing method
	allBdInterfaces := vppcalls.VppUnsetAllInterfacesFromBridgeDomain(
		testDataInBDIfaces[0], dummyBridgeDomain, testDataInDummySwIfIndex, logrus.DefaultLogger(),
		ctx.MockChannel, nil)

	Expect(allBdInterfaces).To(HaveLen(5))
	Expect(allBdInterfaces).To(Equal([]string{ifaceA, ifaceB, ifaceC, ifaceD, ifaceE}))

	Expect(ctx.MockChannel.Msgs).To(HaveLen(4))
	Expect(ctx.MockChannel.Msgs[0]).To(Equal(testDataOutBDIfaces[2]))
	Expect(ctx.MockChannel.Msgs[1]).To(Equal(testDataOutBDIfaces[3]))
}

/**
scenario
- input list of interfaces which should be added to BD is empty
*/
//TestVppSetAllInterfacesToBridgeDomainWithoutInterfaces tests method VppSetAllInterfacesToBridgeDomain
func TestVppUnsetAllInterfacesToBridgeDomainWithoutInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	//call testing method
	allBdInterfaces := vppcalls.VppUnsetAllInterfacesFromBridgeDomain(
		testDataInBDIfaces[1], dummyBridgeDomain, testDataInDummySwIfIndex, logrus.DefaultLogger(),
		ctx.MockChannel, nil)

	//evaluate allBdInterfaces
	Expect(allBdInterfaces).To(BeEmpty())
}

var testDatasInInterfaceToBd = []struct {
	bdIndex   uint32
	swIfIndex uint32
	bvi       bool
}{
	{dummyBridgeDomain, 1, true},
	{dummyBridgeDomain, 1, false},
}

var testDatasOutInterfaceToBd = []*l2ba.SwInterfaceSetL2Bridge{

	{RxSwIfIndex: 1, BdID: dummyBridgeDomain, Bvi: 1, Enable: 1},
	{RxSwIfIndex: 1, BdID: dummyBridgeDomain, Bvi: 0, Enable: 1},
}

/**
scenarios:
- BVI - true
- BVI - false
*/
//TestVppSetInterfaceToBridgeDomain tests VppSetInterfaceToBridgeDomain method
func TestVppSetInterfaceToBridgeDomain(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	for idx, testDataIn := range testDatasInInterfaceToBd {
		ctx.MockVpp.MockReply(&l2ba.SwInterfaceSetL2BridgeReply{})
		vppcalls.VppSetInterfaceToBridgeDomain(testDataIn.bdIndex, testDataIn.swIfIndex, testDataIn.bvi,
			logrus.DefaultLogger(), ctx.MockChannel, nil)
		Expect(ctx.MockChannel.Msg).To(Equal(testDatasOutInterfaceToBd[idx]))
	}
}

func initSwIfIndex() interface{} {
	result := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), dummyPluginName,
		"sw_if_indexes", ifaceidx.IndexMetadata))
	result.RegisterName(ifaceA, swIndexA, nil)
	result.RegisterName(ifaceB, swIndexB, nil)
	result.RegisterName(ifaceC, swIndexC, nil)
	result.RegisterName(ifaceD, swIndexD, nil)
	return result
}
