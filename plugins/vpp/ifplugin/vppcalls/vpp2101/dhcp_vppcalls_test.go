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

	vpp_dhcp "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/dhcp"
)

func TestSetInterfaceAsDHCPClient(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_dhcp.DHCPClientConfigReply{})

	err := ifHandler.SetInterfaceAsDHCPClient(1, "hostName")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_dhcp.DHCPClientConfig)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.Client.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Client.Hostname).To(BeEquivalentTo([]byte("hostName")))
	Expect(vppMsg.Client.WantDHCPEvent).To(BeTrue())
	Expect(vppMsg.IsAdd).To(BeTrue())
}

func TestSetInterfaceAsDHCPClientError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_dhcp.DHCPComplEvent{})

	err := ifHandler.SetInterfaceAsDHCPClient(1, "hostName")

	Expect(err).ToNot(BeNil())
}

func TestSetInterfaceAsDHCPClientRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_dhcp.DHCPClientConfigReply{
		Retval: 1,
	})

	err := ifHandler.SetInterfaceAsDHCPClient(1, "hostName")

	Expect(err).ToNot(BeNil())
}

func TestUnsetInterfaceAsDHCPClient(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_dhcp.DHCPClientConfigReply{})

	err := ifHandler.UnsetInterfaceAsDHCPClient(1, "hostName")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_dhcp.DHCPClientConfig)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.Client.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Client.Hostname).To(BeEquivalentTo([]byte("hostName")))
	Expect(vppMsg.Client.WantDHCPEvent).To(BeTrue())
	Expect(vppMsg.IsAdd).To(BeFalse())
}

func TestUnsetInterfaceAsDHCPClientError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_dhcp.DHCPComplEvent{})

	err := ifHandler.UnsetInterfaceAsDHCPClient(1, "hostName")

	Expect(err).ToNot(BeNil())
}

func TestUnsetInterfaceAsDHCPClientRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_dhcp.DHCPClientConfigReply{
		Retval: 1,
	})

	err := ifHandler.UnsetInterfaceAsDHCPClient(1, "hostName")

	Expect(err).ToNot(BeNil())
}
