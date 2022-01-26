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
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	vpp_nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

// Simulate public and private networks using two microservices and test
// source-NAT in-between.
func TestSourceNAT(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

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
				ToMicroservice: MsNamePrefix + ms1Name,
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
			Reference: MsNamePrefix + ms1Name,
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
				ToMicroservice: MsNamePrefix + ms2Name,
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
			Reference: MsNamePrefix + ms2Name,
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
		return ctx.ExecVppctl("show", "nat44", "addresses")
	}
	connectTCP := func() error {
		return ctx.TestConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, false, tapv2InputNode)
	}
	connectUDP := func() error {
		return ctx.TestConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, true, tapv2InputNode)
	}
	ping := func() error {
		return ctx.PingFromMs(ms2Name, linuxTap1IP)
	}

	ctx.StartMicroservice(ms1Name)
	ctx.StartMicroservice(ms2Name)
	ctx.Expect(nat44Addresses()).ShouldNot(SatisfyAny(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	req := ctx.GenericClient().ChangeRequest()
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
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction creating public and private networks failed")

	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	ctx.Eventually(ctx.GetValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	ctx.Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	ctx.Expect(ping()).Should(Succeed())
	ctx.Expect(connectTCP()).Should(Succeed())
	ctx.Expect(connectUDP()).Should(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// remove S-NAT configuration
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		natGlobal,
		natInterface,
		natPool,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction removing S-NAT failed")

	// check configuration
	ctx.Expect(nat44Addresses()).ShouldNot(SatisfyAny(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	ctx.Expect(ping()).ShouldNot(Succeed())
	ctx.Expect(connectTCP()).ShouldNot(Succeed())
	ctx.Expect(connectUDP()).ShouldNot(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// get back the S-NAT configuration
	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		natGlobal,
		natInterface,
		natPool,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction creating S-NAT failed")

	ctx.Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	ctx.Expect(ping()).Should(Succeed())
	ctx.Expect(connectTCP()).Should(Succeed())
	ctx.Expect(connectUDP()).Should(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart microservice with S-NAT attached
	ctx.StopMicroservice(ms1Name)
	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.StartMicroservice(ms1Name)
	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")

	ctx.Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	ctx.Expect(ping()).Should(Succeed())
	ctx.Expect(connectTCP()).Should(Succeed())
	ctx.Expect(connectUDP()).Should(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")
}

// Tests NAT pool CRUD operation and NAT pool resync.
func TestNATPools(t *testing.T) {
	// variable/helper method initialization
	const addressCount = 7
	addresses := make([]string, addressCount)
	for i := 0; i < addressCount; i++ {
		addresses[i] = fmt.Sprintf("80.80.80.%d", i)
	}
	nat44Addresses := func(ctx *TestCtx) func() (string, error) {
		return func() (string, error) {
			return ctx.ExecVppctl("show", "nat44", "addresses")
		}
	}
	mustBeInVPP := func(ctx *TestCtx, addrs []string) {
		for _, addr := range addrs { //Eventually is needed due to VPP-Agent to VPP configuration delay
			ctx.Eventually(nat44Addresses(ctx)).Should(ContainSubstring(addr))
		}
	}
	cantBeInVPP := func(ctx *TestCtx, addrs []string) {
		for _, addr := range addrs { //Eventually is needed due to VPP-Agent to VPP configuration delay
			ctx.Eventually(nat44Addresses(ctx)).ShouldNot(ContainSubstring(addr))
		}
	}

	// tests
	tests := []struct {
		name               string
		createNATPool      *vpp_nat.Nat44AddressPool
		checkAfterCreation func(ctx *TestCtx)
		updateNATPool      *vpp_nat.Nat44AddressPool
		checkAfterUpdate   func(ctx *TestCtx)
		checkAfterDelete   func(ctx *TestCtx)
	}{
		{
			name: "CRUD for named pool",
			createNATPool: &vpp_nat.Nat44AddressPool{
				Name:    "myPool",
				FirstIp: addresses[0],
				LastIp:  addresses[2],
			},
			checkAfterCreation: func(ctx *TestCtx) {
				mustBeInVPP(ctx, addresses[0:3]) // initial creation
			},
			updateNATPool: &vpp_nat.Nat44AddressPool{
				Name:    "myPool",
				FirstIp: addresses[4],
				LastIp:  addresses[6],
			},
			checkAfterUpdate: func(ctx *TestCtx) {
				cantBeInVPP(ctx, addresses[0:3]) // old addresses are deleted
				mustBeInVPP(ctx, addresses[4:7]) // new addresses are created
			},
			checkAfterDelete: func(ctx *TestCtx) {
				cantBeInVPP(ctx, addresses) // empty pool after deleting only pool
			},
		},
		{
			name: "CRUD for unnamed pool",
			createNATPool: &vpp_nat.Nat44AddressPool{
				FirstIp: addresses[0],
				LastIp:  addresses[2],
			},
			checkAfterCreation: func(ctx *TestCtx) {
				mustBeInVPP(ctx, addresses[0:3]) // initial creation
			},
			updateNATPool: &vpp_nat.Nat44AddressPool{
				FirstIp: addresses[4],
				LastIp:  addresses[6],
			},
			checkAfterUpdate: func(ctx *TestCtx) {
				// unnamed pools are not tied by name in key
				// -> update with different ip addresses == create another pool
				mustBeInVPP(ctx, addresses[0:3])
				mustBeInVPP(ctx, addresses[4:7])
			},
			checkAfterDelete: func(ctx *TestCtx) {
				mustBeInVPP(ctx, addresses[0:3]) // original create won't get deleted
				cantBeInVPP(ctx, addresses[4:7]) // the updated is deleted
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := Setup(t)
			defer ctx.Teardown()

			// check empty pool state
			cantBeInVPP(ctx, addresses)

			// create NAT pool
			req := ctx.GenericClient().ChangeRequest()
			err := req.Update(
				&vpp_nat.Nat44Global{},
				test.createNATPool,
			).Send(context.Background())
			ctx.Expect(err).ToNot(HaveOccurred(), "Transaction creating nat pool failed")
			test.checkAfterCreation(ctx)
			ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync") // resync

			// update NAT pool
			req = ctx.GenericClient().ChangeRequest()
			err = req.Update(
				test.updateNATPool,
			).Send(context.Background())
			ctx.Expect(err).ToNot(HaveOccurred(), "Transaction updating nat pool failed")
			test.checkAfterUpdate(ctx)
			ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync") // resync

			// delete NAT pool
			req = ctx.GenericClient().ChangeRequest()
			err = req.Delete(
				test.updateNATPool,
			).Send(context.Background())
			ctx.Expect(err).ToNot(HaveOccurred(), "Transaction deleting NAT pool failed")
			test.checkAfterDelete(ctx)
			ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync") // resync
		})
	}
}

// Simulate use-case in which a service located in a private network is published
// on a publicly accessible IP address.
func TestNATStaticMappings(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

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
				ToMicroservice: MsNamePrefix + ms1Name,
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
			Reference: MsNamePrefix + ms1Name,
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
				ToMicroservice: MsNamePrefix + ms2Name,
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
			Reference: MsNamePrefix + ms2Name,
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
		return ctx.ExecVppctl("show", "nat44", "static", "mappings")
	}
	containTCP := ContainSubstring(
		"tcp local %s:%d external %s:%d vrf 0  out2in-only",
		linuxTap2IP, tcpSvcLocalPort, tcpSvcExtIP, tcpSvcExtPort)

	containUDP := ContainSubstring(
		"udp local %s:%d external %s:%d vrf 0  out2in-only",
		linuxTap2IP, udpSvcLocalPort, udpSvcExtIP, udpSvcExtPort)

	connectTCP := func() error {
		return ctx.TestConnection(ms1Name, ms2Name, tcpSvcExtIP, linuxTap2IP,
			tcpSvcExtPort, tcpSvcLocalPort, false, tapv2InputNode)
	}
	connectUDP := func() error {
		return ctx.TestConnection(ms1Name, ms2Name, udpSvcExtIP, linuxTap2IP,
			udpSvcExtPort, udpSvcLocalPort, true, tapv2InputNode)
	}

	ctx.StartMicroservice(ms1Name)
	ctx.StartMicroservice(ms2Name)
	ctx.Expect(staticMappings()).ShouldNot(SatisfyAny(containTCP, containUDP))
	req := ctx.GenericClient().ChangeRequest()
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
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction creating public and private networks failed")

	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	ctx.Eventually(ctx.GetValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	ctx.Expect(staticMappings()).Should(SatisfyAll(containTCP, containUDP))
	ctx.Expect(connectTCP()).Should(Succeed())
	ctx.Expect(connectUDP()).Should(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// remove static mappings
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		tcpSvc, udpSvc,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction removing NAT static mappings failed")

	ctx.Expect(staticMappings()).ShouldNot(SatisfyAny(containTCP, containUDP))
	ctx.Expect(connectTCP()).ShouldNot(Succeed())
	ctx.Expect(connectUDP()).ShouldNot(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// get back the NAT configuration
	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		tcpSvc, udpSvc,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction creating NAT static mappings failed")

	ctx.Expect(staticMappings()).Should(SatisfyAll(containTCP, containUDP))
	ctx.Expect(connectTCP()).Should(Succeed())
	ctx.Expect(connectUDP()).Should(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart both microservices
	ctx.StopMicroservice(ms1Name)
	ctx.StopMicroservice(ms2Name)
	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.Eventually(ctx.GetValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.StartMicroservice(ms1Name)
	ctx.StartMicroservice(ms2Name)
	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")
	ctx.Eventually(ctx.GetValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")

	ctx.Expect(staticMappings()).Should(SatisfyAll(containTCP, containUDP))
	ctx.Expect(connectTCP()).Should(Succeed())
	ctx.Expect(connectUDP()).Should(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")
}

// Simulate public and private networks using two microservices and test
// source-NAT in-between. Uses deprecated API for NatInterfaces and AddressPool in Nat44Global.
func TestSourceNATDeprecatedAPI(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

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
				ToMicroservice: MsNamePrefix + ms1Name,
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
			Reference: MsNamePrefix + ms1Name,
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
				ToMicroservice: MsNamePrefix + ms2Name,
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
			Reference: MsNamePrefix + ms2Name,
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
		return ctx.ExecVppctl("show", "nat44", "addresses")
	}
	connectTCP := func() error {
		return ctx.TestConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, false, tapv2InputNode)
	}
	connectUDP := func() error {
		return ctx.TestConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, true, tapv2InputNode)
	}
	ping := func() error {
		return ctx.PingFromMs(ms2Name, linuxTap1IP)
	}

	ctx.StartMicroservice(ms1Name)
	ctx.StartMicroservice(ms2Name)
	ctx.Expect(nat44Addresses()).ShouldNot(SatisfyAny(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		vppTap1,
		linuxTap1,
		vppTap2,
		linuxTap2,
		ms2DefaultRoute,
		sourceNat,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction creating public and private networks failed")

	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	ctx.Eventually(ctx.GetValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	ctx.Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	ctx.Expect(ping()).Should(Succeed())
	ctx.Expect(connectTCP()).Should(Succeed())
	ctx.Expect(connectUDP()).Should(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// remove S-NAT configuration
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		sourceNat,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction removing S-NAT failed")

	// check configuration
	ctx.Expect(nat44Addresses()).ShouldNot(SatisfyAny(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	ctx.Expect(ping()).ShouldNot(Succeed())
	ctx.Expect(connectTCP()).ShouldNot(Succeed())
	ctx.Expect(connectUDP()).ShouldNot(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// get back the S-NAT configuration
	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		sourceNat,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction creating S-NAT failed")

	ctx.Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	ctx.Expect(ping()).Should(Succeed())
	ctx.Expect(connectTCP()).Should(Succeed())
	ctx.Expect(connectUDP()).Should(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart microservice with S-NAT attached
	ctx.StopMicroservice(ms1Name)
	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.StartMicroservice(ms1Name)
	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")

	ctx.Expect(nat44Addresses()).Should(SatisfyAll(
		ContainSubstring(sNatAddr1), ContainSubstring(sNatAddr2), ContainSubstring(sNatAddr3)))
	ctx.Expect(ping()).Should(Succeed())
	ctx.Expect(connectTCP()).Should(Succeed())
	ctx.Expect(connectUDP()).Should(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")
}
