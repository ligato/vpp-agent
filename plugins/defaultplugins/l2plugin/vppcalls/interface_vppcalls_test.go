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

	"github.com/ligato/vpp-agent/tests/vppcallmock"

	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	l2Api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	. "github.com/onsi/gomega"
)

func TestSetInterfacesToBridgeDomain(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))
	swIfIndexes.RegisterName("if1", 1, nil) // Metadata are not required for test purpose
	swIfIndexes.RegisterName("if2", 2, nil)
	swIfIndexes.RegisterName("if3", 3, nil)

	err := vppcalls.SetInterfacesToBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{
		{
			Name: "if1",
			BridgedVirtualInterface: true,
			SplitHorizonGroup:       0,
		},
		{
			Name: "if2",
			BridgedVirtualInterface: false,
			SplitHorizonGroup:       1,
		},
		{
			Name: "if3",
			BridgedVirtualInterface: false,
			SplitHorizonGroup:       2,
		},
	}, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(len(ctx.MockChannel.Msgs)).To(BeEquivalentTo(3))
	for i, msg := range ctx.MockChannel.Msgs {
		var bvi uint8
		if i == 0 {
			bvi = 1
		}
		Expect(msg).To(Equal(&l2Api.SwInterfaceSetL2Bridge{
			RxSwIfIndex: uint32(i + 1),
			BdID:        1,
			Shg:         uint8(i),
			Bvi:         bvi,
			Enable:      1,
		}))
	}
}

func TestSetInterfacesToBridgeDomainNoInterfaceToSet(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))

	err := vppcalls.SetInterfacesToBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{},
		swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(len(ctx.MockChannel.Msgs)).To(BeEquivalentTo(0))
}

func TestSetInterfacesToBridgeDomainMissingInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))
	swIfIndexes.RegisterName("if1", 1, nil) // Metadata are not required for test purpose
	// Interface "if2" is not registered

	err := vppcalls.SetInterfacesToBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{
		{
			Name: "if1",
		},
		{
			Name: "if2",
		},
	}, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(len(ctx.MockChannel.Msgs)).To(BeEquivalentTo(1))
}

func TestSetInterfacesToBridgeDomainError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2Bridge{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))
	swIfIndexes.RegisterName("if1", 1, nil) // Metadata are not required for test purpose

	err := vppcalls.SetInterfacesToBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{
		{
			Name: "if1",
		},
	}, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestSetInterfacesToBridgeDomainRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{
		Retval: 1,
	})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))
	swIfIndexes.RegisterName("if1", 1, nil) // Metadata are not required for test purpose

	err := vppcalls.SetInterfacesToBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{
		{
			Name: "if1",
		},
	}, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestUnsetInterfacesFromBridgeDomain(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))
	swIfIndexes.RegisterName("if1", 1, nil) // Metadata are not required for test purpose
	swIfIndexes.RegisterName("if2", 2, nil)
	swIfIndexes.RegisterName("if3", 3, nil)

	err := vppcalls.UnsetInterfacesFromBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{
		{
			Name:              "if1",
			SplitHorizonGroup: 0,
		},
		{
			Name:              "if2",
			SplitHorizonGroup: 1,
		},
		{
			Name:              "if3",
			SplitHorizonGroup: 2,
		},
	}, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(len(ctx.MockChannel.Msgs)).To(BeEquivalentTo(3))
	for i, msg := range ctx.MockChannel.Msgs {
		Expect(msg).To(Equal(&l2Api.SwInterfaceSetL2Bridge{
			RxSwIfIndex: uint32(i + 1),
			BdID:        1,
			Shg:         uint8(i),
			Enable:      0,
		}))
	}
}

func TestUnsetInterfacesFromBridgeDomainNoInterfaceToUnset(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))

	err := vppcalls.UnsetInterfacesFromBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{},
		swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(len(ctx.MockChannel.Msgs)).To(BeEquivalentTo(0))
}

func TestUnsetInterfacesFromBridgeDomainMissingInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})
	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))
	swIfIndexes.RegisterName("if1", 1, nil) // Metadata are not required for test purpose
	// Interface "if2" is not registered

	err := vppcalls.UnsetInterfacesFromBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{
		{
			Name: "if1",
		},
		{
			Name: "if2",
		},
	}, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(len(ctx.MockChannel.Msgs)).To(BeEquivalentTo(1))
}

func TestUnsetInterfacesFromBridgeDomainError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2Bridge{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))
	swIfIndexes.RegisterName("if1", 1, nil) // Metadata are not required for test purpose

	err := vppcalls.UnsetInterfacesFromBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{
		{
			Name: "if1",
		},
	}, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestUnsetInterfacesFromBridgeDomainRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2Api.SwInterfaceSetL2BridgeReply{
		Retval: 1,
	})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "bd", nil))
	swIfIndexes.RegisterName("if1", 1, nil) // Metadata are not required for test purpose

	err := vppcalls.UnsetInterfacesFromBridgeDomain("bd1", 1, []*l2.BridgeDomains_BridgeDomain_Interfaces{
		{
			Name: "if1",
		},
	}, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}
