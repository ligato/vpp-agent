// Copyright (c) 2018 Cisco and/or its affiliates.
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
	"net"
	"testing"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestSetInterfaceMac(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})

	mac, _ := net.ParseMAC("65:77:BF:72:C9:8D")
	err := SetInterfaceMac(1, "65:77:BF:72:C9:8D", ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceSetMacAddress)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.MacAddress).To(BeEquivalentTo(mac))
}

func TestSetInterfaceInvalidMac(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddress{})

	err := SetInterfaceMac(1, "invalid-mac", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestSetInterfaceMacError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddress{})

	err := SetInterfaceMac(1, "65:77:BF:72:C9:8D", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestSetInterfaceMacRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{
		Retval: 1,
	})

	err := SetInterfaceMac(1, "65:77:BF:72:C9:8D", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}
