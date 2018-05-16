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
// limitations under the License

package vppcalls_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestSetNat44Forwarding(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44ForwardingEnableDisableReply{})
	err := vppcalls.SetNat44Forwarding(true, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44ForwardingEnableDisable)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.Enable).To(BeEquivalentTo(1))
}

func TestUnsetNat44Forwarding(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44ForwardingEnableDisableReply{})
	err := vppcalls.SetNat44Forwarding(false, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44ForwardingEnableDisable)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.Enable).To(BeEquivalentTo(0))
}

func TestSetNat44ForwardingError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&nat.Nat44AddDelStaticMappingReply{})
	err := vppcalls.SetNat44Forwarding(true, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestSetNat44ForwardingRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44ForwardingEnableDisableReply{
		Retval: 1,
	})
	err := vppcalls.SetNat44Forwarding(true, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestEnableNat44InterfaceAsInside(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelFeatureReply{})
	err := vppcalls.EnableNat44Interface(1, true, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44InterfaceAddDelFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
	Expect(msg.IsInside).To(BeEquivalentTo(1))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
}

func TestEnableNat44InterfaceAsOutside(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelFeatureReply{})
	err := vppcalls.EnableNat44Interface(2, false, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44InterfaceAddDelFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
	Expect(msg.IsInside).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(2))
}

func TestEnableNat44InterfaceError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&nat.Nat44AddDelAddressRangeReply{})
	err := vppcalls.EnableNat44Interface(2, false, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestEnableNat44InterfaceRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelFeatureReply{
		Retval: 1,
	})
	err := vppcalls.EnableNat44Interface(2, false, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestDisableNat44InterfaceAsInside(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelFeatureReply{})
	err := vppcalls.DisableNat44Interface(1, true, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44InterfaceAddDelFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeEquivalentTo(0))
	Expect(msg.IsInside).To(BeEquivalentTo(1))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
}

func TestDisableNat44InterfaceAsOutside(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelFeatureReply{})
	err := vppcalls.DisableNat44Interface(2, false, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44InterfaceAddDelFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeEquivalentTo(0))
	Expect(msg.IsInside).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(2))
}

func TestEnableNat44InterfaceOutputAsInside(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelOutputFeatureReply{})
	err := vppcalls.EnableNat44InterfaceOutput(1, true, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44InterfaceAddDelOutputFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
	Expect(msg.IsInside).To(BeEquivalentTo(1))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
}

func TestEnableNat44InterfaceOutputAsOutside(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelOutputFeatureReply{})
	err := vppcalls.EnableNat44InterfaceOutput(2, false, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44InterfaceAddDelOutputFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
	Expect(msg.IsInside).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(2))
}

func TestEnableNat44InterfaceOutputError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&nat.Nat44AddDelStaticMappingReply{})
	err := vppcalls.EnableNat44InterfaceOutput(2, false, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestEnableNat44InterfaceOutputRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelOutputFeatureReply{
		Retval: 1,
	})
	err := vppcalls.EnableNat44InterfaceOutput(2, false, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestDisableNat44InterfaceOutputAsInside(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelOutputFeatureReply{})
	err := vppcalls.DisableNat44InterfaceOutput(1, true, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44InterfaceAddDelOutputFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeEquivalentTo(0))
	Expect(msg.IsInside).To(BeEquivalentTo(1))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
}

func TestDisableNat44InterfaceOutputAsOutside(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44InterfaceAddDelOutputFeatureReply{})
	err := vppcalls.DisableNat44InterfaceOutput(2, false, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44InterfaceAddDelOutputFeature)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.IsAdd).To(BeEquivalentTo(0))
	Expect(msg.IsInside).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(2))
}

func TestAddNat44AddressPool(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	firstIP := net.ParseIP("10.0.0.1").To4()
	lastIP := net.ParseIP("10.0.0.2").To4()

	ctx.MockVpp.MockReply(&nat.Nat44AddDelAddressRangeReply{})
	err := vppcalls.AddNat44AddressPool(firstIP, lastIP, 0, false, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelAddressRange)
	Expect(ok).To(BeTrue())
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
	Expect(msg.FirstIPAddress).To(BeEquivalentTo(firstIP))
	Expect(msg.LastIPAddress).To(BeEquivalentTo(lastIP))
	Expect(msg.VrfID).To(BeEquivalentTo(0))
	Expect(msg.TwiceNat).To(BeEquivalentTo(0))
}

func TestAddNat44AddressPoolError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	firstIP := net.ParseIP("10.0.0.1").To4()
	lastIP := net.ParseIP("10.0.0.2").To4()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&nat.Nat44AddDelIdentityMappingReply{})
	err := vppcalls.AddNat44AddressPool(firstIP, lastIP, 0, false, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestAddNat44AddressPoolRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	firstIP := net.ParseIP("10.0.0.1").To4()
	lastIP := net.ParseIP("10.0.0.2").To4()

	ctx.MockVpp.MockReply(&nat.Nat44AddDelAddressRangeReply{
		Retval: 1,
	})
	err := vppcalls.AddNat44AddressPool(firstIP, lastIP, 0, false, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestDelNat44AddressPool(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	firstIP := net.ParseIP("10.0.0.1").To4()
	lastIP := net.ParseIP("10.0.0.2").To4()

	ctx.MockVpp.MockReply(&nat.Nat44AddDelAddressRangeReply{})
	err := vppcalls.DelNat44AddressPool(firstIP, lastIP, 0, false, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelAddressRange)
	Expect(ok).To(BeTrue())
	Expect(msg.IsAdd).To(BeEquivalentTo(0))
	Expect(msg.FirstIPAddress).To(BeEquivalentTo(firstIP))
	Expect(msg.LastIPAddress).To(BeEquivalentTo(lastIP))
	Expect(msg.VrfID).To(BeEquivalentTo(0))
	Expect(msg.TwiceNat).To(BeEquivalentTo(0))
}

func TestAddNat44StaticMapping(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	localIP := net.ParseIP("10.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()

	// DataContext
	stmCtx := &vppcalls.StaticMappingContext{
		Tag:           "tag1",
		AddressOnly:   false,
		LocalIP:       localIP,
		LocalPort:     24,
		ExternalIP:    externalIP,
		ExternalPort:  8080,
		ExternalIfIdx: 1,
		Protocol:      16,
		Vrf:           1,
		TwiceNat:      true,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelStaticMappingReply{})
	err := vppcalls.AddNat44StaticMapping(stmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.VrfID).To(BeEquivalentTo(1))
	Expect(msg.TwiceNat).To(BeEquivalentTo(1))
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
	Expect(msg.LocalPort).To(BeEquivalentTo(24))
	Expect(msg.ExternalPort).To(BeEquivalentTo(8080))
	Expect(msg.Protocol).To(BeEquivalentTo(16))
	Expect(msg.AddrOnly).To(BeEquivalentTo(0))
	Expect(msg.ExternalIPAddress).To(BeEquivalentTo(externalIP))
	Expect(msg.ExternalSwIfIndex).To(BeEquivalentTo(1))
	Expect(msg.LocalIPAddress).To(BeEquivalentTo(localIP))
	Expect(msg.Out2inOnly).To(BeEquivalentTo(1))
}

func TestAddNat44StaticMappingAddrOnly(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	localIP := net.ParseIP("10.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()

	// DataContext
	stmCtx := &vppcalls.StaticMappingContext{
		Tag:         "tag1",
		AddressOnly: true,
		LocalIP:     localIP,
		ExternalIP:  externalIP,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelStaticMappingReply{})
	err := vppcalls.AddNat44StaticMapping(stmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
	Expect(msg.AddrOnly).To(BeEquivalentTo(1))
	Expect(msg.ExternalIPAddress).To(BeEquivalentTo(externalIP))
	Expect(msg.LocalIPAddress).To(BeEquivalentTo(localIP))
}

func TestAddNat44StaticMappingError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&nat.Nat44AddDelLbStaticMappingReply{})
	err := vppcalls.AddNat44StaticMapping(&vppcalls.StaticMappingContext{}, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestAddNat44StaticMappingRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44AddDelStaticMappingReply{
		Retval: 1,
	})
	err := vppcalls.AddNat44StaticMapping(&vppcalls.StaticMappingContext{}, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestDelNat44StaticMapping(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	localIP := net.ParseIP("10.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()

	// DataContext
	stmCtx := &vppcalls.StaticMappingContext{
		Tag:         "tag1",
		AddressOnly: false,
		LocalIP:     localIP,
		ExternalIP:  externalIP,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelStaticMappingReply{})
	err := vppcalls.DelNat44StaticMapping(stmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.IsAdd).To(BeEquivalentTo(0))
	Expect(msg.AddrOnly).To(BeEquivalentTo(0))
	Expect(msg.ExternalIPAddress).To(BeEquivalentTo(externalIP))
	Expect(msg.LocalIPAddress).To(BeEquivalentTo(localIP))
}

func TestDelNat44StaticMappingAddrOnly(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	localIP := net.ParseIP("10.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()

	// DataContext
	stmCtx := &vppcalls.StaticMappingContext{
		Tag:         "tag1",
		AddressOnly: true,
		LocalIP:     localIP,
		ExternalIP:  externalIP,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelStaticMappingReply{})
	err := vppcalls.DelNat44StaticMapping(stmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.IsAdd).To(BeEquivalentTo(0))
	Expect(msg.AddrOnly).To(BeEquivalentTo(1))
	Expect(msg.ExternalIPAddress).To(BeEquivalentTo(externalIP))
	Expect(msg.LocalIPAddress).To(BeEquivalentTo(localIP))
}

func TestAddNat44StaticMappingLb(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	externalIP := net.ParseIP("10.0.0.1").To4()
	localIP1 := net.ParseIP("10.0.0.2").To4()
	localIP2 := net.ParseIP("10.0.0.3").To4()

	// DataContext
	stmCtx := &vppcalls.StaticMappingLbContext{
		Tag:          "tag1",
		LocalIPs:     localIPs(localIP1, localIP2),
		ExternalIP:   externalIP,
		ExternalPort: 8080,
		Protocol:     16,
		Vrf:          1,
		TwiceNat:     true,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelLbStaticMappingReply{})
	err := vppcalls.AddNat44StaticMappingLb(stmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelLbStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.VrfID).To(BeEquivalentTo(1))
	Expect(msg.TwiceNat).To(BeEquivalentTo(1))
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
	Expect(msg.ExternalAddr).To(BeEquivalentTo(externalIP))
	Expect(msg.ExternalPort).To(BeEquivalentTo(8080))
	Expect(msg.Protocol).To(BeEquivalentTo(16))
	Expect(msg.Out2inOnly).To(BeEquivalentTo(1))

	// Local IPs
	Expect(msg.Locals).To(HaveLen(2))
	expectedCount := 0
	for _, local := range msg.Locals {
		if bytes.Compare(local.Addr, localIP1) == 0 && local.Port == 8080 && local.Probability == 35 {
			expectedCount++
		}
		if bytes.Compare(local.Addr, localIP2) == 0 && local.Port == 8181 && local.Probability == 65 {
			expectedCount++
		}
	}
	Expect(expectedCount).To(BeEquivalentTo(2))
}

func TestAddNat44StaticMappingLbError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&nat.Nat44AddDelIdentityMappingReply{})
	err := vppcalls.AddNat44StaticMappingLb(&vppcalls.StaticMappingLbContext{}, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestAddNat44StaticMappingLbRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44AddDelLbStaticMappingReply{
		Retval: 1,
	})
	err := vppcalls.AddNat44StaticMappingLb(&vppcalls.StaticMappingLbContext{}, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestDelNat44StaticMappingLb(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	externalIP := net.ParseIP("10.0.0.1").To4()
	localIP1 := net.ParseIP("10.0.0.2").To4()
	localIP2 := net.ParseIP("10.0.0.3").To4()

	// DataContext
	stmCtx := &vppcalls.StaticMappingLbContext{
		Tag:          "tag1",
		LocalIPs:     localIPs(localIP1, localIP2),
		ExternalIP:   externalIP,
		ExternalPort: 8080,
		Protocol:     16,
		Vrf:          1,
		TwiceNat:     true,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelLbStaticMappingReply{})
	err := vppcalls.DelNat44StaticMappingLb(stmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelLbStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.VrfID).To(BeEquivalentTo(1))
	Expect(msg.TwiceNat).To(BeEquivalentTo(1))
	Expect(msg.IsAdd).To(BeEquivalentTo(0))
	Expect(msg.ExternalAddr).To(BeEquivalentTo(externalIP))
	Expect(msg.ExternalPort).To(BeEquivalentTo(8080))
	Expect(msg.Protocol).To(BeEquivalentTo(16))
	Expect(msg.Out2inOnly).To(BeEquivalentTo(1))

	// Local IPs
	Expect(msg.Locals).To(HaveLen(2))
	expectedCount := 0
	for _, local := range msg.Locals {
		if bytes.Compare(local.Addr, localIP1) == 0 && local.Port == 8080 && local.Probability == 35 {
			expectedCount++
		}
		if bytes.Compare(local.Addr, localIP2) == 0 && local.Port == 8181 && local.Probability == 65 {
			expectedCount++
		}
	}
	Expect(expectedCount).To(BeEquivalentTo(2))
}

func TestAddNat44IdentityMapping(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	address := net.ParseIP("10.0.0.1").To4()

	// DataContext
	idmCtx := &vppcalls.IdentityMappingContext{
		Tag:       "tag1",
		IPAddress: address,
		Protocol:  16,
		Vrf:       1,
		IfIdx:     1,
		Port:      9000,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelIdentityMappingReply{})
	err := vppcalls.AddNat44IdentityMapping(idmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.VrfID).To(BeEquivalentTo(1))
	Expect(msg.IPAddress).To(BeEquivalentTo(address))
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(msg.Protocol).To(BeEquivalentTo(16))
	Expect(msg.Port).To(BeEquivalentTo(9000))
	Expect(msg.AddrOnly).To(BeEquivalentTo(0))
}

func TestAddNat44IdentityMappingAddrOnly(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DataContext (IPAddress == nil and Port == 0 means it's address only)
	idmCtx := &vppcalls.IdentityMappingContext{
		Tag:      "tag1",
		Protocol: 16,
		Vrf:      1,
		IfIdx:    1,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelIdentityMappingReply{})
	err := vppcalls.AddNat44IdentityMapping(idmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.AddrOnly).To(BeEquivalentTo(1))
	Expect(msg.IsAdd).To(BeEquivalentTo(1))
}

func TestAddNat44IdentityMappingNoInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	address := net.ParseIP("10.0.0.1").To4()

	// DataContext (IPAddress == nil and Port == 0 means it's address only)
	idmCtx := &vppcalls.IdentityMappingContext{
		Tag:       "tag1",
		Protocol:  16,
		Vrf:       1,
		IPAddress: address,
		Port:      8989,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelIdentityMappingReply{})
	err := vppcalls.AddNat44IdentityMapping(idmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.IPAddress).To(BeEquivalentTo(address))
	Expect(msg.Port).To(BeEquivalentTo(8989))
	Expect(msg.AddrOnly).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(vppcalls.NoInterface))
}

func TestAddNat44IdentityMappingError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Incorrect reply object
	ctx.MockVpp.MockReply(&nat.Nat44AddDelStaticMappingReply{})
	err := vppcalls.AddNat44IdentityMapping(&vppcalls.IdentityMappingContext{}, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestAddNat44IdentityMappingRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&nat.Nat44AddDelIdentityMappingReply{
		Retval: 1,
	})
	err := vppcalls.AddNat44IdentityMapping(&vppcalls.IdentityMappingContext{}, ctx.MockChannel, nil)

	Expect(err).Should(HaveOccurred())
}

func TestDelNat44IdentityMapping(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	address := net.ParseIP("10.0.0.1").To4()

	// DataContext
	idmCtx := &vppcalls.IdentityMappingContext{
		Tag:       "tag1",
		IPAddress: address,
		Protocol:  16,
		Vrf:       1,
		IfIdx:     1,
	}

	ctx.MockVpp.MockReply(&nat.Nat44AddDelIdentityMappingReply{})
	err := vppcalls.DelNat44IdentityMapping(idmCtx, ctx.MockChannel, nil)

	Expect(err).ShouldNot(HaveOccurred())

	msg, ok := ctx.MockChannel.Msg.(*nat.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg.Tag).To(BeEquivalentTo("tag1"))
	Expect(msg.VrfID).To(BeEquivalentTo(1))
	Expect(msg.IPAddress).To(BeEquivalentTo(address))
	Expect(msg.IsAdd).To(BeEquivalentTo(0))
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(msg.Protocol).To(BeEquivalentTo(16))
}

func localIPs(addr1, addr2 []byte) []*vppcalls.LocalLbAddress {
	return []*vppcalls.LocalLbAddress{
		{
			Tag:         "tag2",
			LocalIP:     addr1,
			LocalPort:   8080,
			Probability: 35,
		},
		{
			Tag:         "tag3",
			LocalIP:     addr2,
			LocalPort:   8181,
			Probability: 65,
		},
	}
}
