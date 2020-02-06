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
	"net"
	"testing"

	. "github.com/onsi/gomega"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
)

func TestAddInterfaceIP(t *testing.T) {
	var ipv4Addr [4]uint8

	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddressReply{})

	_, ipNet, err := net.ParseCIDR("10.0.0.1/24")
	Expect(err).To(BeNil())
	err = ifHandler.AddInterfaceIP(1, ipNet)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceAddDelAddress)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Prefix.Address.Af).To(BeEquivalentTo(ip_types.ADDRESS_IP4))
	copy(ipv4Addr[:], ipNet.IP.To4())
	Expect(vppMsg.Prefix.Address.Un.GetIP4()).To(BeEquivalentTo(ipv4Addr))
	Expect(vppMsg.Prefix.Len).To(BeEquivalentTo(24))
	Expect(vppMsg.DelAll).To(BeFalse())
	Expect(vppMsg.IsAdd).To(BeTrue())
}

func TestAddInterfaceIPv6(t *testing.T) {
	var ipv6Addr [16]uint8

	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddressReply{})

	_, ipNet, err := net.ParseCIDR("2001:db8:0:1:1:1:1:1/128")
	Expect(err).To(BeNil())
	err = ifHandler.AddInterfaceIP(1, ipNet)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceAddDelAddress)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Prefix.Address.Af).To(BeEquivalentTo(ip_types.ADDRESS_IP6))
	copy(ipv6Addr[:], ipNet.IP.To16())
	Expect(vppMsg.Prefix.Address.Un.GetIP6()).To(BeEquivalentTo(ipv6Addr))
	Expect(vppMsg.Prefix.Len).To(BeEquivalentTo(128))
	Expect(vppMsg.DelAll).To(BeFalse())
	Expect(vppMsg.IsAdd).To(BeTrue())
}

func TestAddInterfaceInvalidIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddressReply{})

	err := ifHandler.AddInterfaceIP(1, &net.IPNet{
		IP: []byte("invalid-ip"),
	})

	Expect(err).ToNot(BeNil())
}

func TestAddInterfaceIPError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	_, ipNet, err := net.ParseCIDR("10.0.0.1/24")
	Expect(err).To(BeNil())
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddress{})

	err = ifHandler.AddInterfaceIP(1, ipNet)

	Expect(err).ToNot(BeNil())
}

func TestAddInterfaceIPRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	_, ipNet, err := net.ParseCIDR("10.0.0.1/24")
	Expect(err).To(BeNil())
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddressReply{
		Retval: 1,
	})

	err = ifHandler.AddInterfaceIP(1, ipNet)

	Expect(err).ToNot(BeNil())
}

func TestDelInterfaceIP(t *testing.T) {
	var ipv4Addr [4]uint8

	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddressReply{})

	_, ipNet, err := net.ParseCIDR("10.0.0.1/24")
	Expect(err).To(BeNil())
	err = ifHandler.DelInterfaceIP(1, ipNet)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceAddDelAddress)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Prefix.Address.Af).To(BeEquivalentTo(ip_types.ADDRESS_IP4))
	copy(ipv4Addr[:], ipNet.IP.To4())
	Expect(vppMsg.Prefix.Address.Un.GetIP4()).To(BeEquivalentTo(ipv4Addr))
	Expect(vppMsg.Prefix.Len).To(BeEquivalentTo(24))
	Expect(vppMsg.DelAll).To(BeFalse())
	Expect(vppMsg.IsAdd).To(BeFalse())
}

func TestDelInterfaceIPv6(t *testing.T) {
	var ipv6Addr [16]uint8

	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddressReply{})

	_, ipNet, err := net.ParseCIDR("2001:db8:0:1:1:1:1:1/128")
	Expect(err).To(BeNil())
	err = ifHandler.DelInterfaceIP(1, ipNet)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceAddDelAddress)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.Prefix.Address.Af).To(BeEquivalentTo(ip_types.ADDRESS_IP6))
	copy(ipv6Addr[:], ipNet.IP.To16())
	Expect(vppMsg.Prefix.Address.Un.GetIP6()).To(BeEquivalentTo(ipv6Addr))
	Expect(vppMsg.Prefix.Len).To(BeEquivalentTo(128))
	Expect(vppMsg.DelAll).To(BeFalse())
	Expect(vppMsg.IsAdd).To(BeFalse())
}

func TestDelInterfaceInvalidIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddressReply{})

	err := ifHandler.DelInterfaceIP(1, &net.IPNet{
		IP: []byte("invalid-ip"),
	})

	Expect(err).ToNot(BeNil())
}

func TestDelInterfaceIPError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	_, ipNet, err := net.ParseCIDR("10.0.0.1/24")
	Expect(err).To(BeNil())
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddress{})

	err = ifHandler.DelInterfaceIP(1, ipNet)

	Expect(err).ToNot(BeNil())
}

func TestDelInterfaceIPRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	_, ipNet, err := net.ParseCIDR("10.0.0.1/24")
	Expect(err).To(BeNil())
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceAddDelAddressReply{
		Retval: 1,
	})

	err = ifHandler.DelInterfaceIP(1, ipNet)

	Expect(err).ToNot(BeNil())
}

func TestSetUnnumberedIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetUnnumberedReply{})

	err := ifHandler.SetUnnumberedIP(ctx.Context, 1, 2)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetUnnumbered)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(2))
	Expect(vppMsg.UnnumberedSwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeTrue())
}

func TestSetUnnumberedIPError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetUnnumbered{})

	err := ifHandler.SetUnnumberedIP(ctx.Context, 1, 2)

	Expect(err).ToNot(BeNil())
}

func TestSetUnnumberedIPRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetUnnumberedReply{
		Retval: 1,
	})

	err := ifHandler.SetUnnumberedIP(ctx.Context, 1, 2)

	Expect(err).ToNot(BeNil())
}

func TestUnsetUnnumberedIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetUnnumberedReply{})

	err := ifHandler.UnsetUnnumberedIP(ctx.Context, 1)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ifs.SwInterfaceSetUnnumbered)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(0))
	Expect(vppMsg.UnnumberedSwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.IsAdd).To(BeFalse())
}

func TestUnsetUnnumberedIPError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetUnnumbered{})

	err := ifHandler.UnsetUnnumberedIP(ctx.Context, 1)

	Expect(err).ToNot(BeNil())
}

func TestUnsetUnnumberedIPRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceSetUnnumberedReply{
		Retval: 1,
	})

	err := ifHandler.UnsetUnnumberedIP(ctx.Context, 1)

	Expect(err).ToNot(BeNil())
}
