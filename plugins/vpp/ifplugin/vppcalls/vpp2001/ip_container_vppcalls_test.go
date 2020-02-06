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
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp2001"
)

func ipToAddr(ip string) vpp_ip.Address {
	addr, err := vpp2001.IPToAddress(ip)
	if err != nil {
		panic(fmt.Sprintf("invalid IP: %s", ip))
	}
	return addr
}

func TestAddContainerIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPContainerProxyAddDelReply{})

	err := ifHandler.AddContainerIP(1, "10.0.0.1/24")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ip.IPContainerProxyAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Pfx).To(BeEquivalentTo(vpp_ip.Prefix{
		Address: ipToAddr("10.0.0.1"),
		Len:     24,
	}))
	Expect(vppMsg.IsAdd).To(BeTrue())
}

func TestAddContainerIPv6(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPContainerProxyAddDelReply{})

	err := ifHandler.AddContainerIP(1, "2001:db8:0:1:1:1:1:1/128")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ip.IPContainerProxyAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Pfx).To(BeEquivalentTo(vpp_ip.Prefix{
		Address: ipToAddr("2001:db8:0:1:1:1:1:1"),
		Len:     128,
	}))
	Expect(vppMsg.IsAdd).To(BeTrue())
}

func TestAddContainerIPInvalidIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPAddressDetails{})

	err := ifHandler.AddContainerIP(1, "invalid-ip")

	Expect(err).ToNot(BeNil())
}

func TestAddContainerIPError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPAddressDetails{})

	err := ifHandler.AddContainerIP(1, "10.0.0.1/24")

	Expect(err).ToNot(BeNil())
}

func TestAddContainerIPRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPContainerProxyAddDelReply{
		Retval: 1,
	})

	err := ifHandler.AddContainerIP(1, "10.0.0.1/24")

	Expect(err).ToNot(BeNil())
}

func TestDelContainerIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPContainerProxyAddDelReply{})

	err := ifHandler.DelContainerIP(1, "10.0.0.1/24")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ip.IPContainerProxyAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Pfx).To(BeEquivalentTo(vpp_ip.Prefix{
		Address: ipToAddr("10.0.0.1"),
		Len:     24,
	}))
	Expect(vppMsg.IsAdd).To(BeFalse())
}

func TestDelContainerIPv6(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPContainerProxyAddDelReply{})

	err := ifHandler.DelContainerIP(1, "2001:db8:0:1:1:1:1:1/128")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ip.IPContainerProxyAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Pfx).To(BeEquivalentTo(vpp_ip.Prefix{
		Address: ipToAddr("2001:db8:0:1:1:1:1:1"),
		Len:     128,
	}))
	Expect(vppMsg.IsAdd).To(BeFalse())
}

func TestDelContainerIPError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPAddressDetails{})

	err := ifHandler.DelContainerIP(1, "10.0.0.1/24")

	Expect(err).ToNot(BeNil())
}

func TestDelContainerIPRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ip.IPContainerProxyAddDelReply{
		Retval: 1,
	})

	err := ifHandler.DelContainerIP(1, "10.0.0.1/24")

	Expect(err).ToNot(BeNil())
}
