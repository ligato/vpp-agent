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

package vpp1904_test

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/ligato/cn-infra/idxmap"
	idxmap_mem "github.com/ligato/cn-infra/idxmap/mem"
	"github.com/ligato/cn-infra/logging/logrus"

	bin_api "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/nat"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls/vpp1904"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

func TestNat44GlobalConfigDump(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	// forwarding
	ctx.MockVpp.MockReply(&bin_api.Nat44ForwardingIsEnabledReply{
		Enabled: 1,
	})

	// virtual reassembly
	ctx.MockVpp.MockReply(&bin_api.NatGetReassReply{
		// IPv4
		IP4Timeout:  10,
		IP4MaxReass: 5,
		IP4MaxFrag:  7,
		IP4DropFrag: 1,
		// IPv6
		IP6Timeout:  20,
		IP6MaxReass: 8,
		IP6MaxFrag:  13,
		IP6DropFrag: 0,
	})

	// non-output interfaces
	ctx.MockVpp.MockReply(
		&bin_api.Nat44InterfaceDetails{
			SwIfIndex: 1,
			IsInside:  0,
		},
		&bin_api.Nat44InterfaceDetails{
			SwIfIndex: 2,
			IsInside:  1,
		})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// output interfaces
	ctx.MockVpp.MockReply(&bin_api.Nat44InterfaceOutputFeatureDetails{
		SwIfIndex: 3,
		IsInside:  1,
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// address pool
	ctx.MockVpp.MockReply(
		&bin_api.Nat44AddressDetails{
			IPAddress: net.ParseIP("192.168.10.1").To4(),
			TwiceNat:  1,
			VrfID:     1,
		},
		&bin_api.Nat44AddressDetails{
			IPAddress: net.ParseIP("192.168.10.2").To4(),
			TwiceNat:  0,
			VrfID:     2,
		})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})
	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})
	swIfIndexes.Put("if2", &ifaceidx.IfaceMetadata{SwIfIndex: 3})

	globalCfg, err := natHandler.Nat44GlobalConfigDump(true)
	Expect(err).To(Succeed())

	Expect(globalCfg.Forwarding).To(BeTrue())

	Expect(globalCfg.AddressPool).To(HaveLen(2))
	Expect(globalCfg.AddressPool[0].Address).To(Equal("192.168.10.1"))
	Expect(globalCfg.AddressPool[0].TwiceNat).To(BeTrue())
	Expect(globalCfg.AddressPool[0].VrfId).To(BeEquivalentTo(1))
	Expect(globalCfg.AddressPool[1].Address).To(Equal("192.168.10.2"))
	Expect(globalCfg.AddressPool[1].TwiceNat).To(BeFalse())
	Expect(globalCfg.AddressPool[1].VrfId).To(BeEquivalentTo(2))

	Expect(globalCfg.NatInterfaces).To(HaveLen(3))
	Expect(globalCfg.NatInterfaces[0].Name).To(Equal("if0"))
	Expect(globalCfg.NatInterfaces[0].IsInside).To(BeFalse())
	Expect(globalCfg.NatInterfaces[0].OutputFeature).To(BeFalse())
	Expect(globalCfg.NatInterfaces[1].Name).To(Equal("if1"))
	Expect(globalCfg.NatInterfaces[1].IsInside).To(BeTrue())
	Expect(globalCfg.NatInterfaces[1].OutputFeature).To(BeFalse())
	Expect(globalCfg.NatInterfaces[2].Name).To(Equal("if2"))
	Expect(globalCfg.NatInterfaces[2].IsInside).To(BeTrue())
	Expect(globalCfg.NatInterfaces[2].OutputFeature).To(BeTrue())

	Expect(globalCfg.VirtualReassembly).ToNot(BeNil())
	Expect(globalCfg.VirtualReassembly.Timeout).To(BeEquivalentTo(10))
	Expect(globalCfg.VirtualReassembly.MaxReassemblies).To(BeEquivalentTo(5))
	Expect(globalCfg.VirtualReassembly.MaxFragments).To(BeEquivalentTo(7))
	Expect(globalCfg.VirtualReassembly.DropFragments).To(BeTrue())
}

func TestDNATDump(t *testing.T) {
	ctx, natHandler, swIfIndexes, dhcpIndexes := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	// non-LB static mappings
	ctx.MockVpp.MockReply(
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    net.ParseIP("10.10.11.120").To4(),
			ExternalIPAddress: net.ParseIP("10.36.20.20").To4(),
			Protocol:          6,
			LocalPort:         8080,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1904.NoInterface,
			VrfID:             1,
			TwiceNat:          1,
			SelfTwiceNat:      0,
			Tag:               []byte("DNAT 1"),
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    net.ParseIP("10.10.11.120").To4(),
			Protocol:          6,
			LocalPort:         8080,
			ExternalPort:      80,
			ExternalSwIfIndex: 1,
			VrfID:             1,
			TwiceNat:          1,
			SelfTwiceNat:      0,
			Tag:               []byte("DNAT 1"),
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    net.ParseIP("10.10.11.140").To4(),
			Protocol:          6,
			LocalPort:         8081,
			ExternalPort:      80,
			ExternalSwIfIndex: 2,
			VrfID:             1,
			TwiceNat:          0,
			SelfTwiceNat:      1,
			Tag:               []byte("DNAT 2"),
		},
		// auto-generated mappings with interface replaced by all assigned IP addresses
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    net.ParseIP("10.10.11.120").To4(),
			ExternalIPAddress: net.ParseIP("10.36.20.30").To4(),
			Protocol:          6,
			LocalPort:         8080,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1904.NoInterface,
			VrfID:             1,
			TwiceNat:          1,
			SelfTwiceNat:      0,
			Tag:               []byte("DNAT 1"),
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    net.ParseIP("10.10.11.120").To4(),
			ExternalIPAddress: net.ParseIP("10.36.20.31").To4(),
			Protocol:          6,
			LocalPort:         8080,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1904.NoInterface,
			VrfID:             1,
			TwiceNat:          1,
			SelfTwiceNat:      0,
			Tag:               []byte("DNAT 1"),
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    net.ParseIP("10.10.11.140").To4(),
			ExternalIPAddress: net.ParseIP("10.36.40.10").To4(),
			Protocol:          6,
			LocalPort:         8081,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1904.NoInterface,
			VrfID:             1,
			TwiceNat:          0,
			SelfTwiceNat:      1,
			Tag:               []byte("DNAT 2"),
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    net.ParseIP("10.10.11.140").To4(),
			ExternalIPAddress: net.ParseIP("10.36.40.20").To4(),
			Protocol:          6,
			LocalPort:         8081,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1904.NoInterface,
			VrfID:             1,
			TwiceNat:          0,
			SelfTwiceNat:      1,
			Tag:               []byte("DNAT 2"),
		},
	)

	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// LB static mappings
	ctx.MockVpp.MockReply(&bin_api.Nat44LbStaticMappingDetails{
		ExternalAddr: net.ParseIP("10.36.20.60").To4(),
		ExternalPort: 53,
		Protocol:     17,
		TwiceNat:     0,
		SelfTwiceNat: 0,
		Out2inOnly:   1,
		Tag:          []byte("DNAT 2"),
		LocalNum:     2,
		Locals: []bin_api.Nat44LbAddrPort{
			{
				Addr:        net.ParseIP("10.10.11.161").To4(),
				Port:        53,
				Probability: 1,
				VrfID:       0,
			},
			{
				Addr:        net.ParseIP("10.10.11.162").To4(),
				Port:        153,
				Probability: 2,
				VrfID:       0,
			},
		},
	})

	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// identity mappings
	ctx.MockVpp.MockReply(
		&bin_api.Nat44IdentityMappingDetails{
			AddrOnly:  1,
			Protocol:  17,
			IPAddress: net.ParseIP("10.10.11.200").To4(),
			SwIfIndex: vpp1904.NoInterface,
			VrfID:     1,
			Tag:       []byte("DNAT 3"),
		},
		&bin_api.Nat44IdentityMappingDetails{
			AddrOnly:  1,
			Protocol:  17,
			SwIfIndex: 2,
			VrfID:     1,
			Tag:       []byte("DNAT 3"),
		},
		// auto-generated mappings with interface replaced by all assigned IP addresses
		&bin_api.Nat44IdentityMappingDetails{
			AddrOnly:  1,
			Protocol:  17,
			IPAddress: net.ParseIP("10.36.40.10").To4(),
			SwIfIndex: vpp1904.NoInterface,
			VrfID:     1,
			Tag:       []byte("DNAT 3"),
		},
		&bin_api.Nat44IdentityMappingDetails{
			AddrOnly:  1,
			Protocol:  17,
			IPAddress: net.ParseIP("10.36.40.20").To4(),
			SwIfIndex: vpp1904.NoInterface,
			VrfID:     1,
			Tag:       []byte("DNAT 3"),
		},
	)

	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// interfaces and their IP addresses
	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1, IPAddresses: []string{"10.36.20.30", "10.36.20.31"}})
	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2, IPAddresses: []string{"10.36.40.10"}})
	dhcpIndexes.Put("if1", &interfaces.DHCPLease{InterfaceName: "if0", HostIpAddress: "10.36.40.20"})

	dnats, err := natHandler.DNat44Dump()
	Expect(err).To(Succeed())

	Expect(dnats).To(HaveLen(3))

	dnat := dnats[0]
	Expect(dnat.Label).To(Equal("DNAT 1"))
	Expect(dnat.IdMappings).To(HaveLen(0))
	Expect(dnat.StMappings).To(HaveLen(2))
	// 1st mapping
	Expect(dnat.StMappings[0].TwiceNat).To(Equal(nat.DNat44_StaticMapping_ENABLED))
	Expect(dnat.StMappings[0].Protocol).To(Equal(nat.DNat44_TCP))
	Expect(dnat.StMappings[0].ExternalInterface).To(BeEmpty())
	Expect(dnat.StMappings[0].ExternalIp).To(Equal("10.36.20.20"))
	Expect(dnat.StMappings[0].ExternalPort).To(BeEquivalentTo(80))
	Expect(dnat.StMappings[0].LocalIps).To(HaveLen(1))
	Expect(dnat.StMappings[0].LocalIps[0].VrfId).To(BeEquivalentTo(1))
	Expect(dnat.StMappings[0].LocalIps[0].LocalIp).To(Equal("10.10.11.120"))
	Expect(dnat.StMappings[0].LocalIps[0].LocalPort).To(BeEquivalentTo(8080))
	Expect(dnat.StMappings[0].LocalIps[0].Probability).To(BeEquivalentTo(0))
	// 2nd mapping
	Expect(dnat.StMappings[1].TwiceNat).To(Equal(nat.DNat44_StaticMapping_ENABLED))
	Expect(dnat.StMappings[1].Protocol).To(Equal(nat.DNat44_TCP))
	Expect(dnat.StMappings[1].ExternalInterface).To(BeEquivalentTo("if0"))
	Expect(dnat.StMappings[1].ExternalIp).To(BeEmpty())
	Expect(dnat.StMappings[1].ExternalPort).To(BeEquivalentTo(80))
	Expect(dnat.StMappings[1].LocalIps).To(HaveLen(1))
	Expect(dnat.StMappings[1].LocalIps[0].VrfId).To(BeEquivalentTo(1))
	Expect(dnat.StMappings[1].LocalIps[0].LocalIp).To(Equal("10.10.11.120"))
	Expect(dnat.StMappings[1].LocalIps[0].LocalPort).To(BeEquivalentTo(8080))
	Expect(dnat.StMappings[1].LocalIps[0].Probability).To(BeEquivalentTo(0))

	dnat = dnats[1]
	// -> non-LB mapping
	Expect(dnat.Label).To(Equal("DNAT 2"))
	Expect(dnat.IdMappings).To(HaveLen(0))
	Expect(dnat.StMappings).To(HaveLen(2))
	Expect(dnat.StMappings[0].TwiceNat).To(Equal(nat.DNat44_StaticMapping_SELF))
	Expect(dnat.StMappings[0].Protocol).To(Equal(nat.DNat44_TCP))
	Expect(dnat.StMappings[0].ExternalInterface).To(Equal("if1"))
	Expect(dnat.StMappings[0].ExternalIp).To(BeEmpty())
	Expect(dnat.StMappings[0].ExternalPort).To(BeEquivalentTo(80))
	Expect(dnat.StMappings[0].LocalIps).To(HaveLen(1))
	Expect(dnat.StMappings[0].LocalIps[0].VrfId).To(BeEquivalentTo(1))
	Expect(dnat.StMappings[0].LocalIps[0].LocalIp).To(Equal("10.10.11.140"))
	Expect(dnat.StMappings[0].LocalIps[0].LocalPort).To(BeEquivalentTo(8081))
	Expect(dnat.StMappings[0].LocalIps[0].Probability).To(BeEquivalentTo(0))
	// -> LB mapping
	Expect(dnat.StMappings[1].TwiceNat).To(Equal(nat.DNat44_StaticMapping_DISABLED))
	Expect(dnat.StMappings[1].Protocol).To(Equal(nat.DNat44_UDP))
	Expect(dnat.StMappings[1].ExternalInterface).To(BeEmpty())
	Expect(dnat.StMappings[1].ExternalIp).To(Equal("10.36.20.60"))
	Expect(dnat.StMappings[1].ExternalPort).To(BeEquivalentTo(53))
	Expect(dnat.StMappings[1].LocalIps).To(HaveLen(2))
	Expect(dnat.StMappings[1].LocalIps[0].VrfId).To(BeEquivalentTo(0))
	Expect(dnat.StMappings[1].LocalIps[0].LocalIp).To(Equal("10.10.11.161"))
	Expect(dnat.StMappings[1].LocalIps[0].LocalPort).To(BeEquivalentTo(53))
	Expect(dnat.StMappings[1].LocalIps[0].Probability).To(BeEquivalentTo(1))
	Expect(dnat.StMappings[1].LocalIps[1].VrfId).To(BeEquivalentTo(0))
	Expect(dnat.StMappings[1].LocalIps[1].LocalIp).To(Equal("10.10.11.162"))
	Expect(dnat.StMappings[1].LocalIps[1].LocalPort).To(BeEquivalentTo(153))
	Expect(dnat.StMappings[1].LocalIps[1].Probability).To(BeEquivalentTo(2))

	dnat = dnats[2]
	Expect(dnat.Label).To(Equal("DNAT 3"))
	Expect(dnat.StMappings).To(HaveLen(0))
	Expect(dnat.IdMappings).To(HaveLen(2))
	// 1st mapping
	Expect(dnat.IdMappings[0].VrfId).To(BeEquivalentTo(1))
	Expect(dnat.IdMappings[0].Protocol).To(Equal(nat.DNat44_UDP))
	Expect(dnat.IdMappings[0].Port).To(BeEquivalentTo(0))
	Expect(dnat.IdMappings[0].IpAddress).To(Equal("10.10.11.200"))
	Expect(dnat.IdMappings[0].Interface).To(BeEmpty())
	// 2nd mapping
	Expect(dnat.IdMappings[1].VrfId).To(BeEquivalentTo(1))
	Expect(dnat.IdMappings[1].Protocol).To(Equal(nat.DNat44_UDP))
	Expect(dnat.IdMappings[1].Port).To(BeEquivalentTo(0))
	Expect(dnat.IdMappings[1].IpAddress).To(BeEmpty())
	Expect(dnat.IdMappings[1].Interface).To(BeEquivalentTo("if1"))
}

func natTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.NatVppAPI, ifaceidx.IfaceMetadataIndexRW, idxmap.NamedMappingRW) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	swIfIndexes := ifaceidx.NewIfaceIndex(logrus.DefaultLogger(), "test-sw_if_indexes")
	dhcpIndexes := idxmap_mem.NewNamedMapping(logrus.DefaultLogger(), "test-dhcp_indexes", nil)
	natHandler := vpp1904.NewNatVppHandler(ctx.MockChannel, swIfIndexes, dhcpIndexes, log)
	return ctx, natHandler, swIfIndexes, dhcpIndexes
}
