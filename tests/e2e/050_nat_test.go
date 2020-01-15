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

package e2e

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	vpp_nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

// Simulate public and private networks using two microservices and test
// source-NAT in-between.
func TestSourceNAT(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		// public network
		vppTap1Name       = "vpp-tap1"
		vppTap1IP         = "80.80.80.20"
		linuxTap1Name     = "linux-tap1"
		linuxTap1Hostname = "tap"
		linuxTap1IP       = "80.80.80.10"

		// private network
		vppTap2Name       = "vpp-tap2"
		vppTap2IP         = "192.168.1.1"
		linuxTap2Name     = "linux-tap2"
		linuxTap2Hostname = "tap"
		linuxTap2IP       = "192.168.1.2"

		// nat
		sNatAddr1 = vppTap1IP
		sNatAddr2 = "80.80.80.21"
		sNatAddr3 = "80.80.80.22"

		netMask = "/24"
		ms1Name = "microservice1"
		ms2Name = "microservice2"
	)

	vppTap1 := &vpp_interfaces.Interface{
		Name:        vppTap1Name,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTap1IP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: msNamePrefix + ms1Name,
			},
		},
	}
	linuxTap1 := &linux_interfaces.Interface{
		Name:        linuxTap1Name,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTap1IP + netMask},
		HostIfName:  linuxTap1Hostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTap1Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: msNamePrefix + ms1Name,
		},
	}

	vppTap2 := &vpp_interfaces.Interface{
		Name:        vppTap2Name,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTap2IP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: msNamePrefix + ms2Name,
			},
		},
	}
	linuxTap2 := &linux_interfaces.Interface{
		Name:        linuxTap2Name,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTap2IP + netMask},
		HostIfName:  linuxTap2Hostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTap2Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: msNamePrefix + ms2Name,
		},
	}

	ms2DefaultRoute := &linux_l3.Route{
		OutgoingInterface: linuxTap2Name,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        "0.0.0.0/0",
		GwAddr:            vppTap2IP,
	}

	natGlobal := &vpp_nat.Nat44Global{
		Forwarding: true,
	}
	natInterface := &vpp_nat.Nat44Interface{
		Name:          vppTap1Name,
		NatOutside:    true,
		OutputFeature: true,
	}
	natPool := &vpp_nat.Nat44AddressPool{
		FirstIp: sNatAddr1,
		LastIp:  sNatAddr3,
	}

	nat44Addresses := func() (string, error) {
		return ctx.execVppctl("show", "nat44", "addresses")
	}
	connectTCP := func() error {
		return ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, false, tapv2InputNode)
	}
	connectUDP := func() error {
		return ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, true, tapv2InputNode)
	}
	ping := func() error {
		return ctx.pingFromMs(ms2Name, linuxTap1IP)
	}

	ctx.startMicroservice(ms1Name)
	ctx.startMicroservice(ms2Name)
	Expect(nat44Addresses()).ShouldNot(SatisfyAny(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppTap1,
		linuxTap1,
		vppTap2,
		linuxTap2,
		ms2DefaultRoute,
		natGlobal,
		natInterface,
		natPool,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction creating public and private networks failed")

	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	Eventually(ctx.getValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	Expect(ping()).Should(Succeed())
	Expect(connectTCP()).Should(Succeed())
	Expect(connectUDP()).Should(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// remove S-NAT configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Delete(
		natGlobal,
		natInterface,
		natPool,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction removing S-NAT failed")

	// check configuration
	Expect(nat44Addresses()).ShouldNot(SatisfyAny(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	Expect(ping()).ShouldNot(Succeed())
	Expect(connectTCP()).ShouldNot(Succeed())
	Expect(connectUDP()).ShouldNot(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// get back the S-NAT configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(
		natGlobal,
		natInterface,
		natPool,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction creating S-NAT failed")

	Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	Expect(ping()).Should(Succeed())
	Expect(connectTCP()).Should(Succeed())
	Expect(connectUDP()).Should(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart microservice with S-NAT attached
	ctx.stopMicroservice(ms1Name)
	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.startMicroservice(ms1Name)
	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")

	Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	Expect(ping()).Should(Succeed())
	Expect(connectTCP()).Should(Succeed())
	Expect(connectUDP()).Should(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")
}

// Simulate use-case in which a service located in a private network is published
// on a publicly accessible IP address.
func TestNATStaticMappings(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		// public network
		vppTap1Name       = "vpp-tap1"
		vppTap1IP         = "80.80.80.20"
		linuxTap1Name     = "linux-tap1"
		linuxTap1Hostname = "tap"
		linuxTap1IP       = "80.80.80.10"

		// private network
		vppTap2Name       = "vpp-tap2"
		vppTap2IP         = "192.168.1.1"
		linuxTap2Name     = "linux-tap2"
		linuxTap2Hostname = "tap"
		linuxTap2IP       = "192.168.1.2"

		// nat
		tcpSvcLabel     = "tcp-service"
		tcpSvcExtIP     = "80.80.80.30"
		tcpSvcExtPort   = 8888
		tcpSvcLocalPort = 8000
		udpSvcLabel     = "udp-service"
		udpSvcExtIP     = "80.80.80.50"
		udpSvcExtPort   = 9999
		udpSvcLocalPort = 9000

		netMask = "/24"
		ms1Name = "microservice1"
		ms2Name = "microservice2"
	)

	vppTap1 := &vpp_interfaces.Interface{
		Name:        vppTap1Name,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTap1IP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: msNamePrefix + ms1Name,
			},
		},
	}
	linuxTap1 := &linux_interfaces.Interface{
		Name:        linuxTap1Name,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTap1IP + netMask},
		HostIfName:  linuxTap1Hostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTap1Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: msNamePrefix + ms1Name,
		},
	}

	vppTap2 := &vpp_interfaces.Interface{
		Name:        vppTap2Name,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTap2IP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: msNamePrefix + ms2Name,
			},
		},
	}
	linuxTap2 := &linux_interfaces.Interface{
		Name:        linuxTap2Name,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTap2IP + netMask},
		HostIfName:  linuxTap2Hostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTap2Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: msNamePrefix + ms2Name,
		},
	}

	ms2DefaultRoute := &linux_l3.Route{
		OutgoingInterface: linuxTap2Name,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        "0.0.0.0/0",
		GwAddr:            vppTap2IP,
	}

	natGlobal := &vpp_nat.Nat44Global{
		Forwarding: true,
	}
	natInterface1 := &vpp_nat.Nat44Interface{
		Name:       vppTap1Name,
		NatOutside: true,
	}
	natInterface2 := &vpp_nat.Nat44Interface{
		Name:      vppTap2Name,
		NatInside: true,
	}

	tcpSvc := &vpp_nat.DNat44{
		Label: tcpSvcLabel,
		StMappings: []*vpp_nat.DNat44_StaticMapping{
			{
				ExternalIp:   tcpSvcExtIP,
				ExternalPort: tcpSvcExtPort,
				Protocol:     vpp_nat.DNat44_TCP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTap2IP,
						LocalPort: tcpSvcLocalPort,
					},
				},
			},
		},
	}

	udpSvc := &vpp_nat.DNat44{
		Label: udpSvcLabel,
		StMappings: []*vpp_nat.DNat44_StaticMapping{
			{
				ExternalIp:   udpSvcExtIP,
				ExternalPort: udpSvcExtPort,
				Protocol:     vpp_nat.DNat44_UDP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTap2IP,
						LocalPort: udpSvcLocalPort,
					},
				},
			},
		},
	}

	// without proxy ARP, the client running from microservice1 (public net),
	// will not get any response for ARP requests concerning service external IPs
	// from VPP
	svcExtIPsProxyArp := &vpp_l3.ProxyARP{
		Interfaces: []*vpp_l3.ProxyARP_Interface{
			{
				Name: vppTap1Name,
			},
		},
		Ranges: []*vpp_l3.ProxyARP_Range{
			{
				FirstIpAddr: tcpSvcExtIP,
				LastIpAddr:  tcpSvcExtIP,
			},
			{
				FirstIpAddr: udpSvcExtIP,
				LastIpAddr:  udpSvcExtIP,
			},
		},
	}

	staticMappings := func() (string, error) {
		return ctx.execVppctl("show", "nat44", "static", "mappings")
	}
	containTCP := ContainSubstring(
		"tcp local %s:%d external %s:%d vrf 0  out2in-only",
		linuxTap2IP, tcpSvcLocalPort, tcpSvcExtIP, tcpSvcExtPort)

	containUDP := ContainSubstring(
		"udp local %s:%d external %s:%d vrf 0  out2in-only",
		linuxTap2IP, udpSvcLocalPort, udpSvcExtIP, udpSvcExtPort)

	connectTCP := func() error {
		return ctx.testConnection(ms1Name, ms2Name, tcpSvcExtIP, linuxTap2IP,
			tcpSvcExtPort, tcpSvcLocalPort, false, tapv2InputNode)
	}
	connectUDP := func() error {
		return ctx.testConnection(ms1Name, ms2Name, udpSvcExtIP, linuxTap2IP,
			udpSvcExtPort, udpSvcLocalPort, true, tapv2InputNode)
	}

	ctx.startMicroservice(ms1Name)
	ctx.startMicroservice(ms2Name)
	Expect(staticMappings()).ShouldNot(SatisfyAny(containTCP, containUDP))
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppTap1,
		linuxTap1,
		vppTap2,
		linuxTap2,
		ms2DefaultRoute,
		svcExtIPsProxyArp,
		natGlobal,
		natInterface1,
		natInterface2,
		tcpSvc, udpSvc,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction creating public and private networks failed")

	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	Eventually(ctx.getValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	Expect(staticMappings()).Should(SatisfyAll(containTCP, containUDP))
	Expect(connectTCP()).Should(Succeed())
	Expect(connectUDP()).Should(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// remove static mappings
	req = ctx.grpcClient.ChangeRequest()
	err = req.Delete(
		tcpSvc, udpSvc,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction removing NAT static mappings failed")

	Expect(staticMappings()).ShouldNot(SatisfyAny(containTCP, containUDP))
	Expect(connectTCP()).ShouldNot(Succeed())
	Expect(connectUDP()).ShouldNot(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// get back the NAT configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(
		tcpSvc, udpSvc,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction creating NAT static mappings failed")

	Expect(staticMappings()).Should(SatisfyAll(containTCP, containUDP))
	Expect(connectTCP()).Should(Succeed())
	Expect(connectUDP()).Should(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart both microservices
	ctx.stopMicroservice(ms1Name)
	ctx.stopMicroservice(ms2Name)
	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	Eventually(ctx.getValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.startMicroservice(ms1Name)
	ctx.startMicroservice(ms2Name)
	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")
	Eventually(ctx.getValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")

	Expect(staticMappings()).Should(SatisfyAll(containTCP, containUDP))
	Expect(connectTCP()).Should(Succeed())
	Expect(connectUDP()).Should(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")
}

// Simulate public and private networks using two microservices and test
// source-NAT in-between. Uses deprecated API for NatInterfaces and AddressPool in Nat44Global.
func TestSourceNATDeprecatedAPI(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		// public network
		vppTap1Name       = "vpp-tap1"
		vppTap1IP         = "80.80.80.20"
		linuxTap1Name     = "linux-tap1"
		linuxTap1Hostname = "tap"
		linuxTap1IP       = "80.80.80.10"

		// private network
		vppTap2Name       = "vpp-tap2"
		vppTap2IP         = "192.168.1.1"
		linuxTap2Name     = "linux-tap2"
		linuxTap2Hostname = "tap"
		linuxTap2IP       = "192.168.1.2"

		// nat
		sNatAddr1 = vppTap1IP
		sNatAddr2 = "80.80.80.30"
		sNatAddr3 = "80.80.80.40"

		netMask = "/24"
		ms1Name = "microservice1"
		ms2Name = "microservice2"
	)

	vppTap1 := &vpp_interfaces.Interface{
		Name:        vppTap1Name,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTap1IP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: msNamePrefix + ms1Name,
			},
		},
	}
	linuxTap1 := &linux_interfaces.Interface{
		Name:        linuxTap1Name,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTap1IP + netMask},
		HostIfName:  linuxTap1Hostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTap1Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: msNamePrefix + ms1Name,
		},
	}

	vppTap2 := &vpp_interfaces.Interface{
		Name:        vppTap2Name,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTap2IP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: msNamePrefix + ms2Name,
			},
		},
	}
	linuxTap2 := &linux_interfaces.Interface{
		Name:        linuxTap2Name,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTap2IP + netMask},
		HostIfName:  linuxTap2Hostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTap2Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: msNamePrefix + ms2Name,
		},
	}

	ms2DefaultRoute := &linux_l3.Route{
		OutgoingInterface: linuxTap2Name,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        "0.0.0.0/0",
		GwAddr:            vppTap2IP,
	}

	sourceNat := &vpp_nat.Nat44Global{
		Forwarding: true,
		NatInterfaces: []*vpp_nat.Nat44Global_Interface{
			{
				Name:          vppTap1Name,
				IsInside:      false,
				OutputFeature: true,
			},
		},
		AddressPool: []*vpp_nat.Nat44Global_Address{
			{
				Address: sNatAddr1,
			},
			{
				Address: sNatAddr2,
			},
			{
				Address: sNatAddr3,
			},
		},
	}

	nat44Addresses := func() (string, error) {
		return ctx.execVppctl("show", "nat44", "addresses")
	}
	connectTCP := func() error {
		return ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, false, tapv2InputNode)
	}
	connectUDP := func() error {
		return ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, true, tapv2InputNode)
	}
	ping := func() error {
		return ctx.pingFromMs(ms2Name, linuxTap1IP)
	}

	ctx.startMicroservice(ms1Name)
	ctx.startMicroservice(ms2Name)
	Expect(nat44Addresses()).ShouldNot(SatisfyAny(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppTap1,
		linuxTap1,
		vppTap2,
		linuxTap2,
		ms2DefaultRoute,
		sourceNat,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction creating public and private networks failed")

	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	Eventually(ctx.getValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	Expect(ping()).Should(Succeed())
	Expect(connectTCP()).Should(Succeed())
	Expect(connectUDP()).Should(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// remove S-NAT configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Delete(
		sourceNat,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction removing S-NAT failed")

	// check configuration
	Expect(nat44Addresses()).ShouldNot(SatisfyAny(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	Expect(ping()).ShouldNot(Succeed())
	Expect(connectTCP()).ShouldNot(Succeed())
	Expect(connectUDP()).ShouldNot(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// get back the S-NAT configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(
		sourceNat,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction creating S-NAT failed")

	Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	Expect(ping()).Should(Succeed())
	Expect(connectTCP()).Should(Succeed())
	Expect(connectUDP()).Should(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart microservice with S-NAT attached
	ctx.stopMicroservice(ms1Name)
	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.startMicroservice(ms1Name)
	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")

	Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	Expect(ping()).Should(Succeed())
	Expect(connectTCP()).Should(Succeed())
	Expect(connectUDP()).Should(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")
}
