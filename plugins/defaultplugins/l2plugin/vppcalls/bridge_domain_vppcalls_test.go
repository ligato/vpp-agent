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
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

const (
	dummyBridgeDomainName = "bridge_domain"
)

//Input test data for creating bridge domain
var createTestDataInBD *l2.BridgeDomains_BridgeDomain = &l2.BridgeDomains_BridgeDomain{
	Name:                dummyBridgeDomainName,
	Flood:               true,
	UnknownUnicastFlood: true,
	Forward:             true,
	Learn:               true,
	ArpTermination:      true,
	MacAge:              45,
}

//Output test data for creating bridge domain
var createTestDataOutBD *l2ba.BridgeDomainAddDel = &l2ba.BridgeDomainAddDel{
	BdID:    dummyBridgeDomain,
	Flood:   1,
	UuFlood: 1,
	Forward: 1,
	Learn:   1,
	ArpTerm: 1,
	MacAge:  45,
	IsAdd:   1,
}

//Input test data for updating bridge domain
var updateTestDataInBd *l2.BridgeDomains_BridgeDomain = &l2.BridgeDomains_BridgeDomain{
	Name:                dummyBridgeDomainName,
	Flood:               false,
	UnknownUnicastFlood: false,
	Forward:             false,
	Learn:               false,
	ArpTermination:      false,
	MacAge:              50,
}

//Output test data for updating bridge domain
var updateTestDataOutBd *l2ba.BridgeDomainAddDel = &l2ba.BridgeDomainAddDel{
	BdID:    dummyBridgeDomain,
	Flood:   0,
	UuFlood: 0,
	Forward: 0,
	Learn:   0,
	ArpTerm: 0,
	MacAge:  50,
	IsAdd:   1,
}

//Output test data for deleting bridge domain
var deleteTestDataOutBd *l2ba.BridgeDomainAddDel = &l2ba.BridgeDomainAddDel{
	BdID:  dummyBridgeDomain,
	IsAdd: 0,
}

func TestVppAddBridgeDomain(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	err := vppcalls.VppAddBridgeDomain(dummyBridgeDomain, createTestDataInBD,
		logrus.NewLogger(dummyLoggerName), ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*l2ba.BridgeDomainAddDel)
	Expect(ok).To(BeTrue())
	Expect(msg).To(Equal(createTestDataOutBD))
}

func TestVppUpdateBridgeDomain(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	err := vppcalls.VppUpdateBridgeDomain(dummyBridgeDomain, dummyBridgeDomain, updateTestDataInBd,
		logrus.NewLogger(dummyLoggerName), ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	Expect(ctx.MockChannel.Msgs).To(HaveLen(2))

	//delete msg
	msg, ok := ctx.MockChannel.Msgs[0].(*l2ba.BridgeDomainAddDel)
	Expect(ok).To(BeTrue())
	Expect(msg).To(Equal(deleteTestDataOutBd))

	//add msg
	msg, ok = ctx.MockChannel.Msgs[1].(*l2ba.BridgeDomainAddDel)
	Expect(ok).To(BeTrue())
	Expect(msg).To(Equal(updateTestDataOutBd))
}

func TestVppDeleteBridgeDomain(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&l2ba.BridgeDomainAddDelReply{})
	err := vppcalls.VppDeleteBridgeDomain(dummyBridgeDomain,
		logrus.NewLogger(dummyLoggerName), ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*l2ba.BridgeDomainAddDel)
	Expect(ok).To(BeTrue())
	Expect(msg).To(Equal(deleteTestDataOutBd))
}
