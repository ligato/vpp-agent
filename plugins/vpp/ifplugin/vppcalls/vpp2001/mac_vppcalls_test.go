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

package vpp2001_test

import (
	"testing"

	. "github.com/onsi/gomega"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp2001"
)

func TestSetInterfaceMac(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetMacAddressReply{})

	mac, err := vpp2001.ParseMAC("65:77:BF:72:C9:8D")
	Expect(err).To(BeNil())
	err = ifHandler.SetInterfaceMac(1, "65:77:BF:72:C9:8D")
	Expect(err).To(BeNil())

	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetMacAddress)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.MacAddress).To(BeEquivalentTo(mac))
}

func TestSetInterfaceInvalidMac(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetMacAddress{})

	err := ifHandler.SetInterfaceMac(1, "invalid-mac")

	Expect(err).ToNot(BeNil())
}

func TestSetInterfaceMacError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetMacAddress{})

	err := ifHandler.SetInterfaceMac(1, "65:77:BF:72:C9:8D")

	Expect(err).ToNot(BeNil())
}

func TestSetInterfaceMacRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetMacAddressReply{
		Retval: 1,
	})

	err := ifHandler.SetInterfaceMac(1, "65:77:BF:72:C9:8D")

	Expect(err).ToNot(BeNil())
}
