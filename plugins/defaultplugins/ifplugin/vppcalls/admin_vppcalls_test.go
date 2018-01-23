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

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

const (
	dummyInterfaceIndex uint32 = 42
)

var testDataInterfaceAdminDown = &interfaces.SwInterfaceSetFlags{
	SwIfIndex:   dummyInterfaceIndex,
	AdminUpDown: 0,
}

var testDataInterfaceAdminUp = &interfaces.SwInterfaceSetFlags{
	SwIfIndex:   dummyInterfaceIndex,
	AdminUpDown: 1,
}

func TestInterfaceAdminDown(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})

	err := vppcalls.InterfaceAdminDown(dummyInterfaceIndex, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceSetFlags)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg).To(Equal(testDataInterfaceAdminDown))
}

func TestInterfaceAdminUp(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})

	err := vppcalls.InterfaceAdminUp(dummyInterfaceIndex, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceSetFlags)
	Expect(ok).To(BeTrue())

	Expect(vppMsg).NotTo(BeNil())
	Expect(vppMsg).To(Equal(testDataInterfaceAdminUp))

}
