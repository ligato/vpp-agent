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
	"bytes"
	"net"
	"testing"

	vpp_nat "github.com/ligato/vpp-agent/api/models/vpp/nat"
	binapi "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/nat"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin/vppcalls/vpp1908"
	. "github.com/onsi/gomega"
)

func TestSetNat44Forwarding(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.Nat44ForwardingEnableDisableReply{})
	err := natHandler.SetNat44Forwarding(true)

	Expect(err).ShouldNot(HaveOccurred())

	t.Logf("Msg: %+v (%#v)", ctx.MockChannel.Msg, ctx.MockChannel.Msg)
	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44ForwardingEnableDisable)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.Enable).To(BeTrue())
}

func TestUnsetNat44Forwarding(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.Nat44ForwardingEnableDisableReply{})
	err := natHandler.SetNat44Forwarding(false)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44ForwardingEnableDisable)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.Enable).To(BeFalse())
}

func TestSetNat44ForwardingError(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{})
	err := natHandler.SetNat44Forwarding(true)

	Expect(err).Should(HaveOccurred())
}

func TestSetNat44ForwardingRetval(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.Nat44ForwardingEnableDisableReply{
		Retval: 1,
	})
	err := natHandler.SetNat44Forwarding(true)

	Expect(err).Should(HaveOccurred())
}

func TestEnableNat44InterfaceAsInside(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelFeatureReply{})
	err := natHandler.EnableNat44Interface("if0", true, false)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44InterfaceAddDelFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeTrue())
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_INSIDE))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
}

func TestEnableNat44InterfaceAsOutside(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelFeatureReply{})
	err := natHandler.EnableNat44Interface("if1", false, false)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44InterfaceAddDelFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeTrue())
	Expect(msg.Flags).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(2))
}

func TestEnableNat44InterfaceError(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	// Incorrect reply object
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelAddressRangeReply{})
	err := natHandler.EnableNat44Interface("if1", false, false)

	Expect(err).Should(HaveOccurred())
}

func TestEnableNat44InterfaceRetval(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelFeatureReply{
		Retval: 1,
	})
	err := natHandler.EnableNat44Interface("if1", false, false)

	Expect(err).Should(HaveOccurred())
}

func TestDisableNat44InterfaceAsInside(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelFeatureReply{})
	err := natHandler.DisableNat44Interface("if0", true, false)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44InterfaceAddDelFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeFalse())
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_INSIDE))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
}

func TestDisableNat44InterfaceAsOutside(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelFeatureReply{})
	err := natHandler.DisableNat44Interface("if1", false, false)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44InterfaceAddDelFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeFalse())
	Expect(msg.Flags).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(2))
}

func TestEnableNat44InterfaceOutputAsInside(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelOutputFeatureReply{})
	err := natHandler.EnableNat44Interface("if0", true, true)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44InterfaceAddDelOutputFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeTrue())
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_INSIDE))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
}

func TestEnableNat44InterfaceOutputAsOutside(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelOutputFeatureReply{})
	err := natHandler.EnableNat44Interface("if1", false, true)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44InterfaceAddDelOutputFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeTrue())
	Expect(msg.Flags).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(2))
}

func TestEnableNat44InterfaceOutputError(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	// Incorrect reply object
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{})
	err := natHandler.EnableNat44Interface("if1", false, true)

	Expect(err).Should(HaveOccurred())
}

func TestEnableNat44InterfaceOutputRetval(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelOutputFeatureReply{
		Retval: 1,
	})
	err := natHandler.EnableNat44Interface("if1", false, true)

	Expect(err).Should(HaveOccurred())
}

func TestDisableNat44InterfaceOutputAsInside(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelOutputFeatureReply{})
	err := natHandler.DisableNat44Interface("if0", true, true)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44InterfaceAddDelOutputFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeFalse())
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_INSIDE))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
}

func TestDisableNat44InterfaceOutputAsOutside(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	ctx.MockVpp.MockReply(&binapi.Nat44InterfaceAddDelOutputFeatureReply{})
	err := natHandler.DisableNat44Interface("if1", false, true)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44InterfaceAddDelOutputFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeFalse())
	Expect(msg.Flags).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(2))
}

func TestAddNat44Address(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	addr := net.ParseIP("10.0.0.1").To4()

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelAddressRangeReply{})
	err := natHandler.AddNat44Address(addr.String(), 0, false)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelAddressRange)
	Expect(ok).To(BeTrue())
	Expect(msg.IsAdd).To(BeTrue())
	Expect(addressTo4IP(msg.FirstIPAddress)).To(BeEquivalentTo(addr.String()))
	Expect(addressTo4IP(msg.LastIPAddress)).To(BeEquivalentTo(addr.String()))
	Expect(msg.VrfID).To(BeEquivalentTo(0))
	Expect(msg.Flags).To(BeEquivalentTo(0))
}

func TestAddNat44AddressError(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	addr := net.ParseIP("10.0.0.1").To4()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelIdentityMappingReply{})
	err := natHandler.AddNat44Address(addr.String(), 0, false)

	Expect(err).Should(HaveOccurred())
}

func TestAddNat44AddressPoolRetval(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	addr := net.ParseIP("10.0.0.1").To4()

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelAddressRangeReply{
		Retval: 1,
	})
	err := natHandler.AddNat44Address(addr.String(), 0, false)

	Expect(err).Should(HaveOccurred())
}

func TestDelNat44Address(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	addr := net.ParseIP("10.0.0.1").To4()

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelAddressRangeReply{})
	err := natHandler.DelNat44Address(addr.String(), 0, false)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelAddressRange)
	Expect(ok).To(BeTrue())
	Expect(msg.IsAdd).To(BeFalse())
	Expect(addressTo4IP(msg.FirstIPAddress)).To(BeEquivalentTo(addr.String()))
	Expect(addressTo4IP(msg.LastIPAddress)).To(BeEquivalentTo(addr.String()))
	Expect(msg.VrfID).To(BeEquivalentTo(0))
	Expect(msg.Flags).To(BeEquivalentTo(0))
}

func TestSetNat44VirtualReassemblyIPv4(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.NatSetReassReply{})
	err := natHandler.SetVirtualReassemblyIPv4(&vpp_nat.VirtualReassembly{
		Timeout:         10,
		MaxFragments:    20,
		MaxReassemblies: 30,
		DropFragments:   true,
	})

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.NatSetReass)
	Expect(ok).To(BeTrue())
	Expect(msg.Timeout).To(BeEquivalentTo(10))
	Expect(msg.MaxFrag).To(BeEquivalentTo(20))
	Expect(msg.MaxReass).To(BeEquivalentTo(30))
	Expect(msg.DropFrag).To(BeEquivalentTo(1))
}

func TestSetNat44VirtualReassemblyIPv6(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.NatSetReassReply{})
	err := natHandler.SetVirtualReassemblyIPv6(&vpp_nat.VirtualReassembly{
		Timeout:         5,
		MaxFragments:    10,
		MaxReassemblies: 15,
		DropFragments:   true,
	})

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.NatSetReass)
	Expect(ok).To(BeTrue())
	Expect(msg.Timeout).To(BeEquivalentTo(5))
	Expect(msg.MaxFrag).To(BeEquivalentTo(10))
	Expect(msg.MaxReass).To(BeEquivalentTo(15))
	Expect(msg.DropFrag).To(BeEquivalentTo(1))
}

func TestAddNat44StaticMapping(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	localIP := net.ParseIP("10.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()

	// DataContext
	mapping := &vpp_nat.DNat44_StaticMapping{
		ExternalIp:        externalIP.String(),
		ExternalPort:      8080,
		ExternalInterface: "if0", // overrides external IP
		Protocol:          vpp_nat.DNat44_TCP,
		TwiceNat:          vpp_nat.DNat44_StaticMapping_ENABLED,
		LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
			{
				LocalIp:   localIP.String(),
				VrfId:     1,
				LocalPort: 24,
			},
		},
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{})
	err := natHandler.AddNat44StaticMapping(mapping, "DNAT 1")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 1"))
	Expect(msg.VrfID).To(BeEquivalentTo(1))
	Expect(msg.IsAdd).To(BeTrue())
	Expect(msg.LocalPort).To(BeEquivalentTo(24))
	Expect(msg.ExternalPort).To(BeEquivalentTo(8080))
	Expect(msg.Protocol).To(BeEquivalentTo(6))
	Expect(addressTo4IP(msg.ExternalIPAddress)).To(BeEquivalentTo("0.0.0.0"))
	Expect(msg.ExternalSwIfIndex).To(BeEquivalentTo(1))
	Expect(addressTo4IP(msg.LocalIPAddress)).To(BeEquivalentTo(localIP.String()))
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_TWICE_NAT + binapi.NAT_IS_OUT2IN_ONLY))
}

func TestAddNat44IdentityMappingWithInterface(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	localIP := net.ParseIP("10.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()

	// DataContext
	mapping := &vpp_nat.DNat44_StaticMapping{
		ExternalIp: externalIP.String(),
		Protocol:   vpp_nat.DNat44_TCP,
		LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
			{
				LocalIp: localIP.String(),
			},
		},
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{})
	err := natHandler.AddNat44StaticMapping(mapping, "DNAT 1")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 1"))
	Expect(msg.IsAdd).To(BeTrue())
	Expect(addressTo4IP(msg.ExternalIPAddress)).To(BeEquivalentTo(externalIP.String()))
	Expect(addressTo4IP(msg.LocalIPAddress)).To(BeEquivalentTo(localIP.String()))
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_OUT2IN_ONLY + binapi.NAT_IS_ADDR_ONLY))
}

func TestAddNat44StaticMappingError(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelLbStaticMappingReply{})
	err := natHandler.AddNat44StaticMapping(&vpp_nat.DNat44_StaticMapping{}, "")

	Expect(err).Should(HaveOccurred())
}

func TestAddNat44StaticMappingRetval(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{
		Retval: 1,
	})
	err := natHandler.AddNat44StaticMapping(&vpp_nat.DNat44_StaticMapping{}, "")

	Expect(err).Should(HaveOccurred())
}

func TestDelNat44StaticMapping(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	localIP := net.ParseIP("10.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()

	mapping := &vpp_nat.DNat44_StaticMapping{
		ExternalIp:        externalIP.String(),
		ExternalPort:      8080,
		ExternalInterface: "if0", // overrides external IP
		Protocol:          vpp_nat.DNat44_TCP,
		TwiceNat:          vpp_nat.DNat44_StaticMapping_ENABLED,
		LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
			{
				LocalIp:   localIP.String(),
				VrfId:     1,
				LocalPort: 24,
			},
		},
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{})
	err := natHandler.DelNat44StaticMapping(mapping, "DNAT 1")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 1"))
	Expect(msg.VrfID).To(BeEquivalentTo(1))
	Expect(msg.IsAdd).To(BeFalse())
	Expect(msg.LocalPort).To(BeEquivalentTo(24))
	Expect(msg.ExternalPort).To(BeEquivalentTo(8080))
	Expect(msg.Protocol).To(BeEquivalentTo(6))
	Expect(addressTo4IP(msg.ExternalIPAddress)).To(BeEquivalentTo("0.0.0.0"))
	Expect(msg.ExternalSwIfIndex).To(BeEquivalentTo(1))
	Expect(addressTo4IP(msg.LocalIPAddress)).To(BeEquivalentTo(localIP.String()))
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_TWICE_NAT + binapi.NAT_IS_OUT2IN_ONLY))
}

func TestDelNat44StaticMappingAddrOnly(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	localIP := net.ParseIP("10.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()

	mapping := &vpp_nat.DNat44_StaticMapping{
		ExternalIp: externalIP.String(),
		Protocol:   vpp_nat.DNat44_TCP,
		LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
			{
				LocalIp: localIP.String(),
			},
		},
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{})
	err := natHandler.DelNat44StaticMapping(mapping, "DNAT 1")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 1"))
	Expect(msg.IsAdd).To(BeFalse())
	Expect(addressTo4IP(msg.ExternalIPAddress)).To(BeEquivalentTo(externalIP.String()))
	Expect(addressTo4IP(msg.LocalIPAddress)).To(BeEquivalentTo(localIP.String()))
}

func TestAddNat44StaticMappingLb(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	externalIP := net.ParseIP("10.0.0.1").To4()
	localIP1 := net.ParseIP("10.0.0.2").To4()
	localIP2 := net.ParseIP("10.0.0.3").To4()

	mapping := &vpp_nat.DNat44_StaticMapping{
		ExternalIp:        externalIP.String(),
		ExternalPort:      8080,
		ExternalInterface: "if0",
		Protocol:          vpp_nat.DNat44_TCP,
		TwiceNat:          vpp_nat.DNat44_StaticMapping_ENABLED,
		LocalIps:          localIPs(localIP1, localIP2),
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelLbStaticMappingReply{})
	err := natHandler.AddNat44StaticMapping(mapping, "DNAT 1")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelLbStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 1"))
	Expect(msg.IsAdd).To(BeTrue())
	Expect(addressTo4IP(msg.ExternalAddr)).To(BeEquivalentTo(externalIP.String()))
	Expect(msg.ExternalPort).To(BeEquivalentTo(8080))
	Expect(msg.Protocol).To(BeEquivalentTo(6))
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_TWICE_NAT + binapi.NAT_IS_OUT2IN_ONLY))

	// Local IPs
	Expect(msg.Locals).To(HaveLen(2))
	expectedCount := 0
	for _, local := range msg.Locals {
		localAddr := make(net.IP, net.IPv4len)
		copy(localAddr, local.Addr[:])
		if bytes.Compare(localAddr, localIP1) == 0 && local.Port == 8080 && local.Probability == 35 {
			expectedCount++
		}
		copy(localAddr, local.Addr[:])
		if bytes.Compare(localAddr, localIP2) == 0 && local.Port == 8181 && local.Probability == 65 {
			expectedCount++
		}
	}
	Expect(expectedCount).To(BeEquivalentTo(2))
}

func TestDelNat44StaticMappingLb(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	externalIP := net.ParseIP("10.0.0.1").To4()
	localIP1 := net.ParseIP("10.0.0.2").To4()
	localIP2 := net.ParseIP("10.0.0.3").To4()

	mapping := &vpp_nat.DNat44_StaticMapping{
		ExternalIp:        externalIP.String(),
		ExternalPort:      8080,
		ExternalInterface: "if0",
		Protocol:          vpp_nat.DNat44_TCP,
		TwiceNat:          vpp_nat.DNat44_StaticMapping_ENABLED,
		LocalIps:          localIPs(localIP1, localIP2),
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelLbStaticMappingReply{})
	err := natHandler.DelNat44StaticMapping(mapping, "DNAT 1")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelLbStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 1"))
	Expect(msg.IsAdd).To(BeFalse())
	Expect(addressTo4IP(msg.ExternalAddr)).To(BeEquivalentTo(externalIP.String()))
	Expect(msg.ExternalPort).To(BeEquivalentTo(8080))
	Expect(msg.Protocol).To(BeEquivalentTo(6))
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_TWICE_NAT + binapi.NAT_IS_OUT2IN_ONLY))

	// Local IPs
	Expect(msg.Locals).To(HaveLen(2))
	expectedCount := 0
	for _, local := range msg.Locals {
		localAddr := make(net.IP, net.IPv4len)
		copy(localAddr, local.Addr[:])
		if bytes.Compare(localAddr, localIP1) == 0 && local.Port == 8080 && local.Probability == 35 {
			expectedCount++
		}
		copy(localAddr, local.Addr[:])
		if bytes.Compare(localAddr, localIP2) == 0 && local.Port == 8181 && local.Probability == 65 {
			expectedCount++
		}
	}
	Expect(expectedCount).To(BeEquivalentTo(2))
}

func TestAddNat44IdentityMapping(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	address := net.ParseIP("10.0.0.1").To4()

	mapping := &vpp_nat.DNat44_IdentityMapping{
		VrfId:     1,
		Interface: "if0", // overrides IP address
		IpAddress: address.String(),
		Port:      9000,
		Protocol:  vpp_nat.DNat44_UDP,
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelIdentityMappingReply{})
	err := natHandler.AddNat44IdentityMapping(mapping, "DNAT 1")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 1"))
	Expect(msg.VrfID).To(BeEquivalentTo(1))
	Expect(addressTo4IP(msg.IPAddress)).To(BeEquivalentTo("0.0.0.0"))
	Expect(msg.IsAdd).To(BeTrue())
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(msg.Protocol).To(BeEquivalentTo(17))
	Expect(msg.Port).To(BeEquivalentTo(9000))
	Expect(msg.Flags).To(BeEquivalentTo(0))
}

func TestAddNat44IdentityMappingAddrOnly(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	// IPAddress == nil and Port == 0 means it's address only
	mapping := &vpp_nat.DNat44_IdentityMapping{
		VrfId:     1,
		Interface: "if0", // overrides IP address
		Protocol:  vpp_nat.DNat44_UDP,
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelIdentityMappingReply{})
	err := natHandler.AddNat44IdentityMapping(mapping, "DNAT 1")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 1"))
	Expect(addressTo4IP(msg.IPAddress)).To(BeEquivalentTo("0.0.0.0"))
	Expect(msg.IsAdd).To(BeTrue())
	Expect(msg.Protocol).To(BeEquivalentTo(17))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_ADDR_ONLY))
}

func TestAddNat44IdentityMappingNoInterface(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	address := net.ParseIP("10.0.0.1").To4()

	mapping := &vpp_nat.DNat44_IdentityMapping{
		VrfId:     1,
		Protocol:  vpp_nat.DNat44_UDP,
		IpAddress: address.String(),
		Port:      8989,
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelIdentityMappingReply{})
	err := natHandler.AddNat44IdentityMapping(mapping, "DNAT 2")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 2"))
	Expect(addressTo4IP(msg.IPAddress)).To(BeEquivalentTo(address.String()))
	Expect(msg.Port).To(BeEquivalentTo(8989))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(vpp1908.NoInterface))
	Expect(msg.Flags).To(BeEquivalentTo(0))
}

func TestAddNat44IdentityMappingError(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{})
	err := natHandler.AddNat44IdentityMapping(&vpp_nat.DNat44_IdentityMapping{}, "")

	Expect(err).Should(HaveOccurred())
}

func TestAddNat44IdentityMappingRetval(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelIdentityMappingReply{
		Retval: 1,
	})
	err := natHandler.AddNat44IdentityMapping(&vpp_nat.DNat44_IdentityMapping{}, "")

	Expect(err).Should(HaveOccurred())
}

func TestDelNat44IdentityMapping(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	address := net.ParseIP("10.0.0.1").To4()

	mapping := &vpp_nat.DNat44_IdentityMapping{
		Interface: "if0",
		IpAddress: address.String(),
		Protocol:  vpp_nat.DNat44_TCP,
		VrfId:     1,
	}

	ctx.MockVpp.MockReply(&binapi.Nat44AddDelIdentityMappingReply{})
	err := natHandler.DelNat44IdentityMapping(mapping, "DNAT 2")

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*binapi.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("DNAT 2"))
	Expect(msg.VrfID).To(BeEquivalentTo(1))
	Expect(addressTo4IP(msg.IPAddress)).To(BeEquivalentTo("0.0.0.0"))
	Expect(msg.IsAdd).To(BeFalse())
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(msg.Protocol).To(BeEquivalentTo(6))
	Expect(msg.Flags).To(BeEquivalentTo(binapi.NAT_IS_ADDR_ONLY))
}

func TestNat44MappingLongTag(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	normalTag := "normalTag"
	longTag := "some-weird-tag-which-is-much-longer-than-allowed-sixty-four-bytes"

	localIP1 := net.ParseIP("10.0.0.1").To4()
	localIP2 := net.ParseIP("20.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()

	mapping := &vpp_nat.DNat44_StaticMapping{
		LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
			{
				LocalIp: localIP1.String(),
			},
		},
		ExternalIp: externalIP.String(),
	}
	lbMapping := &vpp_nat.DNat44_StaticMapping{
		LocalIps:     localIPs(localIP1, localIP2),
		ExternalIp:   externalIP.String(),
		ExternalPort: 8080,
		Protocol:     vpp_nat.DNat44_TCP,
		TwiceNat:     vpp_nat.DNat44_StaticMapping_ENABLED,
	}
	idMapping := &vpp_nat.DNat44_IdentityMapping{
		IpAddress: localIP1.String(),
		Protocol:  vpp_nat.DNat44_UDP,
		VrfId:     1,
		Interface: "if0",
	}

	// 1. test
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{})
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelLbStaticMappingReply{})
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelIdentityMappingReply{})
	// 2. test
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelStaticMappingReply{})
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelLbStaticMappingReply{})
	ctx.MockVpp.MockReply(&binapi.Nat44AddDelIdentityMappingReply{})

	// Successful scenario (to ensure there is no other error)
	err := natHandler.AddNat44StaticMapping(mapping, normalTag)
	Expect(err).To(BeNil())
	err = natHandler.AddNat44StaticMapping(lbMapping, normalTag)
	Expect(err).To(BeNil())
	err = natHandler.AddNat44IdentityMapping(idMapping, normalTag)
	Expect(err).To(BeNil())

	// Replace tags and test again
	err = natHandler.AddNat44StaticMapping(mapping, longTag)
	Expect(err).ToNot(BeNil())
	err = natHandler.AddNat44StaticMapping(lbMapping, longTag)
	Expect(err).ToNot(BeNil())
	err = natHandler.AddNat44IdentityMapping(idMapping, longTag)
	Expect(err).ToNot(BeNil())
}

func localIPs(addr1, addr2 net.IP) []*vpp_nat.DNat44_StaticMapping_LocalIP {
	return []*vpp_nat.DNat44_StaticMapping_LocalIP{
		{
			LocalIp:     addr1.String(),
			LocalPort:   8080,
			Probability: 35,
		},
		{
			LocalIp:     addr2.String(),
			LocalPort:   8181,
			Probability: 65,
		},
	}
}

func addressTo4IP(address binapi.IP4Address) string {
	ipAddr := make(net.IP, net.IPv4len)
	copy(ipAddr[:], address[:])
	if ipAddr.To4() == nil {
		return ""
	}
	return ipAddr.To4().String()
}
