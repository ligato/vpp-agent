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

package vpp1908_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp1908"
)

func ipToAddr(ip string) ip.Address {
	addr, err := vpp1908.IPToAddress(ip)
	if err != nil {
		panic(fmt.Sprintf("invalid IP: %s", ip))
	}
	return addr
}

func TestAddContainerIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})

	err := ifHandler.AddContainerIP(1, "10.0.0.1/24")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*ip.IPContainerProxyAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Pfx).To(BeEquivalentTo(ip.Prefix{
		Address: ipToAddr("10.0.0.1"),
		Len:     24,
	}))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
}

func TestAddContainerIPv6(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})

	err := ifHandler.AddContainerIP(1, "2001:db8:0:1:1:1:1:1/128")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*ip.IPContainerProxyAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Pfx).To(BeEquivalentTo(ip.Prefix{
		Address: ipToAddr("2001:db8:0:1:1:1:1:1"),
		Len:     128,
	}))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
}

func TestAddContainerIPInvalidIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPAddressDetails{})

	err := ifHandler.AddContainerIP(1, "invalid-ip")

	Expect(err).ToNot(BeNil())
}

func TestAddContainerIPError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPAddressDetails{})

	err := ifHandler.AddContainerIP(1, "10.0.0.1/24")

	Expect(err).ToNot(BeNil())
}

func TestAddContainerIPRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{
		Retval: 1,
	})

	err := ifHandler.AddContainerIP(1, "10.0.0.1/24")

	Expect(err).ToNot(BeNil())
}

func TestDelContainerIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})

	err := ifHandler.DelContainerIP(1, "10.0.0.1/24")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*ip.IPContainerProxyAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Pfx).To(BeEquivalentTo(ip.Prefix{
		Address: ipToAddr("10.0.0.1"),
		Len:     24,
	}))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(0))
}

func TestDelContainerIPv6(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})

	err := ifHandler.DelContainerIP(1, "2001:db8:0:1:1:1:1:1/128")

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*ip.IPContainerProxyAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Pfx).To(BeEquivalentTo(ip.Prefix{
		Address: ipToAddr("2001:db8:0:1:1:1:1:1"),
		Len:     128,
	}))
	Expect(vppMsg.IsAdd).To(BeEquivalentTo(0))
}

func TestDelContainerIPError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPAddressDetails{})

	err := ifHandler.DelContainerIP(1, "10.0.0.1/24")

	Expect(err).ToNot(BeNil())
}

func TestDelContainerIPRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{
		Retval: 1,
	})

	err := ifHandler.DelContainerIP(1, "10.0.0.1/24")

	Expect(err).ToNot(BeNil())
}
