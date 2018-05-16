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

package vppcalls_test

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/dhcp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestSetInterfaceAsDHCPClient(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&dhcp.DhcpClientConfigReply{})

	err := vppcalls.SetInterfaceAsDHCPClient(1, "hostName", ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*dhcp.DhcpClientConfig)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Hostname).To(BeEquivalentTo([]byte("hostName")))
	Expect(vppMsg.WantDhcpEvent).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
}

func TestSetInterfaceAsDHCPClientError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&dhcp.DhcpComplEvent{})

	err := vppcalls.SetInterfaceAsDHCPClient(1, "hostName", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestSetInterfaceAsDHCPClientRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&dhcp.DhcpClientConfigReply{
		Retval: 1,
	})

	err := vppcalls.SetInterfaceAsDHCPClient(1, "hostName", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestUnsetInterfaceAsDHCPClient(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&dhcp.DhcpClientConfigReply{})

	err := vppcalls.UnsetInterfaceAsDHCPClient(1, "hostName", ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*dhcp.DhcpClientConfig)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Hostname).To(BeEquivalentTo([]byte("hostName")))
	Expect(vppMsg.WantDhcpEvent).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(0))
}

func TestUnsetInterfaceAsDHCPClientError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&dhcp.DhcpComplEvent{})

	err := vppcalls.UnsetInterfaceAsDHCPClient(1, "hostName", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestUnsetInterfaceAsDHCPClientRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&dhcp.DhcpClientConfigReply{
		Retval: 1,
	})

	err := vppcalls.UnsetInterfaceAsDHCPClient(1, "hostName", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}
