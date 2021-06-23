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

package vpp2106_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	vpp_l2 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	vpp2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls/vpp2106"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

const (
	dummyBridgeDomain     = 4
	dummyBridgeDomainName = "bridge_domain"
)

// Input test data for creating bridge domain
var createTestDataInBD = &l2.BridgeDomain{
	Name:                dummyBridgeDomainName,
	Flood:               true,
	UnknownUnicastFlood: true,
	Forward:             true,
	Learn:               true,
	ArpTermination:      true,
	MacAge:              45,
}

// Output test data for creating bridge domain
var createTestDataOutBD = &vpp_l2.BridgeDomainAddDel{
	BdID:    dummyBridgeDomain,
	Flood:   true,
	UuFlood: true,
	Forward: true,
	Learn:   true,
	ArpTerm: true,
	MacAge:  45,
	BdTag:   dummyBridgeDomainName,
	IsAdd:   true,
}

// Output test data for deleting bridge domain
var deleteTestDataOutBd = &vpp_l2.BridgeDomainAddDel{
	BdID:  dummyBridgeDomain,
	IsAdd: false,
}

func TestVppAddBridgeDomain(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{})
	err := bdHandler.AddBridgeDomain(dummyBridgeDomain, createTestDataInBD)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(createTestDataOutBD))
}

func TestVppAddBridgeDomainError(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{Retval: 1})
	ctx.MockVpp.MockReply(&vpp_l2.SwInterfaceSetL2Bridge{})

	err := bdHandler.AddBridgeDomain(dummyBridgeDomain, createTestDataInBD)
	Expect(err).Should(HaveOccurred())

	err = bdHandler.AddBridgeDomain(dummyBridgeDomain, createTestDataInBD)
	Expect(err).Should(HaveOccurred())
}

func TestVppDeleteBridgeDomain(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{})
	err := bdHandler.DeleteBridgeDomain(dummyBridgeDomain)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(Equal(deleteTestDataOutBd))
}

func TestVppDeleteBridgeDomainError(t *testing.T) {
	ctx, bdHandler, _ := bdTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_l2.BridgeDomainAddDelReply{Retval: 1})
	ctx.MockVpp.MockReply(&vpp_l2.SwInterfaceSetL2Bridge{})

	err := bdHandler.DeleteBridgeDomain(dummyBridgeDomain)
	Expect(err).Should(HaveOccurred())

	err = bdHandler.DeleteBridgeDomain(dummyBridgeDomain)
	Expect(err).Should(HaveOccurred())
}

func bdTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.BridgeDomainVppAPI, ifaceidx.IfaceMetadataIndexRW) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	ifIndex := ifaceidx.NewIfaceIndex(log, "bd-test-ifidx")
	bdHandler := vpp2106.NewL2VppHandler(ctx.MockChannel, ifIndex, nil, log)
	return ctx, bdHandler, ifIndex
}
