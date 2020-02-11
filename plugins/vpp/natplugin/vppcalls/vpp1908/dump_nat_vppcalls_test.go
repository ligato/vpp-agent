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
	"net"
	"testing"

	. "github.com/onsi/gomega"

	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"

	"go.ligato.io/cn-infra/v2/idxmap"
	idxmap_mem "go.ligato.io/cn-infra/v2/idxmap/mem"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	bin_api "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/nat"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls/vpp1908"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
)

func TestNat44GlobalConfigDump(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	// forwarding
	ctx.MockVpp.MockReply(&bin_api.Nat44ForwardingIsEnabledReply{
		Enabled: true,
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
		},
		&bin_api.Nat44InterfaceDetails{
			SwIfIndex: 2,
			Flags:     bin_api.NAT_IS_INSIDE,
		})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// output interfaces
	ctx.MockVpp.MockReply(&bin_api.Nat44InterfaceOutputFeatureDetails{
		SwIfIndex: 3,
		Flags:     bin_api.NAT_IS_INSIDE,
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// address pool
	ctx.MockVpp.MockReply(
		&bin_api.Nat44AddressDetails{
			IPAddress: ipTo4Address("192.168.10.1"),
			Flags:     bin_api.NAT_IS_TWICE_NAT,
			VrfID:     1,
		},
		&bin_api.Nat44AddressDetails{
			IPAddress: ipTo4Address("192.168.10.2"),
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

func TestNat44InterfacesDump(t *testing.T) {
	ctx, natHandler, swIfIndexes, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	// non-output interfaces
	ctx.MockVpp.MockReply(
		&bin_api.Nat44InterfaceDetails{
			SwIfIndex: 1,
			Flags:     bin_api.NAT_IS_OUTSIDE,
		},
		&bin_api.Nat44InterfaceDetails{
			SwIfIndex: 2,
			Flags:     bin_api.NAT_IS_INSIDE,
		})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// output interfaces
	ctx.MockVpp.MockReply(&bin_api.Nat44InterfaceOutputFeatureDetails{
		SwIfIndex: 3,
		Flags:     bin_api.NAT_IS_INSIDE | bin_api.NAT_IS_OUTSIDE,
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1})
	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})
	swIfIndexes.Put("if2", &ifaceidx.IfaceMetadata{SwIfIndex: 3})

	interfaces, err := natHandler.Nat44InterfacesDump()
	Expect(err).To(Succeed())

	Expect(interfaces).To(HaveLen(3))

	Expect(interfaces[0].Name).To(Equal("if0"))
	Expect(interfaces[0].NatInside).To(BeFalse())
	Expect(interfaces[0].NatOutside).To(BeTrue())
	Expect(interfaces[0].OutputFeature).To(BeFalse())

	Expect(interfaces[1].Name).To(Equal("if1"))
	Expect(interfaces[1].NatInside).To(BeTrue())
	Expect(interfaces[1].NatOutside).To(BeFalse())
	Expect(interfaces[1].OutputFeature).To(BeFalse())

	Expect(interfaces[2].Name).To(Equal("if2"))
	Expect(interfaces[2].NatInside).To(BeTrue())
	Expect(interfaces[2].NatOutside).To(BeTrue())
	Expect(interfaces[2].OutputFeature).To(BeTrue())
}

func TestNat44AddressPoolsDump(t *testing.T) {
	ctx, natHandler, _, _ := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	// address pool
	ctx.MockVpp.MockReply(
		&bin_api.Nat44AddressDetails{
			IPAddress: ipTo4Address("192.168.10.1"),
			Flags:     bin_api.NAT_IS_TWICE_NAT,
			VrfID:     1,
		},
		&bin_api.Nat44AddressDetails{
			IPAddress: ipTo4Address("192.168.10.2"),
			VrfID:     2,
		},
		&bin_api.Nat44AddressDetails{
			IPAddress: ipTo4Address("192.168.10.3"),
			VrfID:     2,
		},
		&bin_api.Nat44AddressDetails{
			IPAddress: ipTo4Address("192.168.10.4"),
			VrfID:     2,
		})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	pools, err := natHandler.Nat44AddressPoolsDump()
	Expect(err).To(Succeed())

	Expect(pools).To(HaveLen(2))

	Expect(pools[0].FirstIp).To(Equal("192.168.10.1"))
	Expect(pools[0].LastIp).To(Equal(""))
	Expect(pools[0].TwiceNat).To(BeTrue())
	Expect(pools[0].VrfId).To(BeEquivalentTo(1))

	Expect(pools[1].FirstIp).To(Equal("192.168.10.2"))
	Expect(pools[1].LastIp).To(Equal("192.168.10.4"))
	Expect(pools[1].TwiceNat).To(BeFalse())
	Expect(pools[1].VrfId).To(BeEquivalentTo(2))
}

func TestDNATDump(t *testing.T) {
	ctx, natHandler, swIfIndexes, dhcpIndexes := natTestSetup(t)
	defer ctx.TeardownTestCtx()

	// non-LB static mappings
	ctx.MockVpp.MockReply(
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    ipTo4Address("10.10.11.120"),
			ExternalIPAddress: ipTo4Address("10.36.20.20"),
			Protocol:          6,
			LocalPort:         8080,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1908.NoInterface,
			VrfID:             1,
			Tag:               "DNAT 1",
			Flags:             bin_api.NAT_IS_TWICE_NAT,
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    ipTo4Address("10.10.11.120"),
			Protocol:          6,
			LocalPort:         8080,
			ExternalPort:      80,
			ExternalSwIfIndex: 1,
			VrfID:             1,
			Tag:               "DNAT 1",
			Flags:             bin_api.NAT_IS_TWICE_NAT,
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    ipTo4Address("10.10.11.140"),
			Protocol:          6,
			LocalPort:         8081,
			ExternalPort:      80,
			ExternalSwIfIndex: 2,
			VrfID:             1,
			Tag:               "DNAT 2",
			Flags:             bin_api.NAT_IS_SELF_TWICE_NAT,
		},
		// auto-generated mappings with interface replaced by all assigned IP addresses
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    ipTo4Address("10.10.11.120"),
			ExternalIPAddress: ipTo4Address("10.36.20.30"),
			Protocol:          6,
			LocalPort:         8080,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1908.NoInterface,
			VrfID:             1,
			Tag:               "DNAT 1",
			Flags:             bin_api.NAT_IS_TWICE_NAT,
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    ipTo4Address("10.10.11.120"),
			ExternalIPAddress: ipTo4Address("10.36.20.31"),
			Protocol:          6,
			LocalPort:         8080,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1908.NoInterface,
			VrfID:             1,
			Tag:               "DNAT 1",
			Flags:             bin_api.NAT_IS_TWICE_NAT,
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    ipTo4Address("10.10.11.140"),
			ExternalIPAddress: ipTo4Address("10.36.40.10"),
			Protocol:          6,
			LocalPort:         8081,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1908.NoInterface,
			VrfID:             1,
			Tag:               "DNAT 2",
			Flags:             bin_api.NAT_IS_SELF_TWICE_NAT,
		},
		&bin_api.Nat44StaticMappingDetails{
			LocalIPAddress:    ipTo4Address("10.10.11.140"),
			ExternalIPAddress: ipTo4Address("10.36.40.20"),
			Protocol:          6,
			LocalPort:         8081,
			ExternalPort:      80,
			ExternalSwIfIndex: vpp1908.NoInterface,
			VrfID:             1,
			Tag:               "DNAT 2",
			Flags:             bin_api.NAT_IS_SELF_TWICE_NAT,
		},
	)

	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// LB static mappings
	ctx.MockVpp.MockReply(&bin_api.Nat44LbStaticMappingDetails{
		ExternalAddr: ipTo4Address("10.36.20.60"),
		ExternalPort: 53,
		Protocol:     17,
		Flags:        bin_api.NAT_IS_OUT2IN_ONLY,
		Tag:          "DNAT 2",
		LocalNum:     2,
		Locals: []bin_api.Nat44LbAddrPort{
			{
				Addr:        ipTo4Address("10.10.11.161"),
				Port:        53,
				Probability: 1,
				VrfID:       0,
			},
			{
				Addr:        ipTo4Address("10.10.11.162"),
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
			Flags:     bin_api.NAT_IS_ADDR_ONLY,
			Protocol:  17,
			IPAddress: ipTo4Address("10.10.11.200"),
			SwIfIndex: vpp1908.NoInterface,
			VrfID:     1,
			Tag:       "DNAT 3",
		},
		&bin_api.Nat44IdentityMappingDetails{
			Flags:     bin_api.NAT_IS_ADDR_ONLY,
			Protocol:  17,
			SwIfIndex: 2,
			VrfID:     1,
			Tag:       "DNAT 3",
		},
		// auto-generated mappings with interface replaced by all assigned IP addresses
		&bin_api.Nat44IdentityMappingDetails{
			Flags:     bin_api.NAT_IS_ADDR_ONLY,
			Protocol:  17,
			IPAddress: ipTo4Address("10.36.40.10"),
			SwIfIndex: vpp1908.NoInterface,
			VrfID:     1,
			Tag:       "DNAT 3",
		},
		&bin_api.Nat44IdentityMappingDetails{
			Flags:     bin_api.NAT_IS_ADDR_ONLY,
			Protocol:  17,
			IPAddress: ipTo4Address("10.36.40.20"),
			SwIfIndex: vpp1908.NoInterface,
			VrfID:     1,
			Tag:       "DNAT 3",
		},
	)

	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	// interfaces and their IP addresses
	swIfIndexes.Put("if0", &ifaceidx.IfaceMetadata{SwIfIndex: 1, IPAddresses: []string{"10.36.20.30", "10.36.20.31"}})
	swIfIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2, IPAddresses: []string{"10.36.40.10"}})
	dhcpIndexes.Put("if1", &vpp_interfaces.DHCPLease{InterfaceName: "if0", HostIpAddress: "10.36.40.20"})

	dnats, err := natHandler.DNat44Dump()
	Expect(err).To(Succeed())

	Expect(dnats).To(HaveLen(3))

	dnat := dnats[0]
	Expect(dnat.Label).To(Equal("DNAT 1"))
	Expect(dnat.IdMappings).To(HaveLen(0))
	Expect(dnat.StMappings).To(HaveLen(2))
	// 1st mapping
	Expect(dnat.StMappings[0].TwiceNat).To(Equal(vpp_nat.DNat44_StaticMapping_ENABLED))
	Expect(dnat.StMappings[0].Protocol).To(Equal(vpp_nat.DNat44_TCP))
	Expect(dnat.StMappings[0].ExternalInterface).To(BeEmpty())
	Expect(dnat.StMappings[0].ExternalIp).To(Equal("10.36.20.20"))
	Expect(dnat.StMappings[0].ExternalPort).To(BeEquivalentTo(80))
	Expect(dnat.StMappings[0].LocalIps).To(HaveLen(1))
	Expect(dnat.StMappings[0].LocalIps[0].VrfId).To(BeEquivalentTo(1))
	Expect(dnat.StMappings[0].LocalIps[0].LocalIp).To(Equal("10.10.11.120"))
	Expect(dnat.StMappings[0].LocalIps[0].LocalPort).To(BeEquivalentTo(8080))
	Expect(dnat.StMappings[0].LocalIps[0].Probability).To(BeEquivalentTo(0))
	// 2nd mapping
	Expect(dnat.StMappings[1].TwiceNat).To(Equal(vpp_nat.DNat44_StaticMapping_ENABLED))
	Expect(dnat.StMappings[1].Protocol).To(Equal(vpp_nat.DNat44_TCP))
	Expect(dnat.StMappings[1].ExternalInterface).To(BeEquivalentTo("if0"))
	Expect(dnat.StMappings[1].ExternalIp).To(BeEquivalentTo("0.0.0.0"))
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
	Expect(dnat.StMappings[0].TwiceNat).To(Equal(vpp_nat.DNat44_StaticMapping_SELF))
	Expect(dnat.StMappings[0].Protocol).To(Equal(vpp_nat.DNat44_TCP))
	Expect(dnat.StMappings[0].ExternalInterface).To(Equal("if1"))
	Expect(dnat.StMappings[0].ExternalIp).To(BeEquivalentTo("0.0.0.0"))
	Expect(dnat.StMappings[0].ExternalPort).To(BeEquivalentTo(80))
	Expect(dnat.StMappings[0].LocalIps).To(HaveLen(1))
	Expect(dnat.StMappings[0].LocalIps[0].VrfId).To(BeEquivalentTo(1))
	Expect(dnat.StMappings[0].LocalIps[0].LocalIp).To(Equal("10.10.11.140"))
	Expect(dnat.StMappings[0].LocalIps[0].LocalPort).To(BeEquivalentTo(8081))
	Expect(dnat.StMappings[0].LocalIps[0].Probability).To(BeEquivalentTo(0))
	// -> LB mapping
	Expect(dnat.StMappings[1].TwiceNat).To(Equal(vpp_nat.DNat44_StaticMapping_DISABLED))
	Expect(dnat.StMappings[1].Protocol).To(Equal(vpp_nat.DNat44_UDP))
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
	Expect(dnat.IdMappings[0].Protocol).To(Equal(vpp_nat.DNat44_UDP))
	Expect(dnat.IdMappings[0].Port).To(BeEquivalentTo(0))
	Expect(dnat.IdMappings[0].IpAddress).To(Equal("10.10.11.200"))
	Expect(dnat.IdMappings[0].Interface).To(BeEmpty())
	// 2nd mapping
	Expect(dnat.IdMappings[1].VrfId).To(BeEquivalentTo(1))
	Expect(dnat.IdMappings[1].Protocol).To(Equal(vpp_nat.DNat44_UDP))
	Expect(dnat.IdMappings[1].Port).To(BeEquivalentTo(0))
	Expect(dnat.IdMappings[1].IpAddress).To(BeEquivalentTo("0.0.0.0"))
	Expect(dnat.IdMappings[1].Interface).To(BeEquivalentTo("if1"))
}

func natTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.NatVppAPI, ifaceidx.IfaceMetadataIndexRW, idxmap.NamedMappingRW) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	swIfIndexes := ifaceidx.NewIfaceIndex(logrus.DefaultLogger(), "test-sw_if_indexes")
	dhcpIndexes := idxmap_mem.NewNamedMapping(logrus.DefaultLogger(), "test-dhcp_indexes", nil)
	natHandler := vpp1908.NewNatVppHandler(ctx.MockChannel, swIfIndexes, dhcpIndexes, log)
	return ctx, natHandler, swIfIndexes, dhcpIndexes
}

func ipTo4Address(ipStr string) (addr bin_api.IP4Address) {
	netIP := net.ParseIP(ipStr)
	if ip4 := netIP.To4(); ip4 != nil {
		copy(addr[:], ip4)
	}
	return
}
