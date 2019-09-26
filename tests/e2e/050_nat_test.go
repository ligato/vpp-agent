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
	"github.com/onsi/gomega/types"

	"github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/l3"
	"github.com/ligato/vpp-agent/api/models/linux/namespace"
	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/api/models/vpp/nat"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
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

	checkConfig := func(isConfigured bool) {
		substringMatcher := func(substring string, negative bool) types.GomegaMatcher {
			if negative {
				return Not(ContainSubstring(substring))
			} else {
				return ContainSubstring(substring)
			}
		}
		stdout, err := ctx.execVppctl("show", "nat44", "addresses")
		Expect(err).To(BeNil(), "Running `vppctl show nat44 addresses` failed")
		Expect(stdout).To(substringMatcher(sNatAddr1, !isConfigured),
			"Unexpected S-NAT address configuration`",
		)
		Expect(stdout).To(substringMatcher(sNatAddr2, !isConfigured),
			"Unexpected S-NAT address configuration`",
		)
		Expect(stdout).To(substringMatcher(sNatAddr3, !isConfigured),
			"Unexpected S-NAT address configuration`",
		)
	}

	checkConn := func(shouldSucceed bool) {
		expected := BeNil()
		if !shouldSucceed {
			expected = Not(BeNil())
		}
		Expect(ctx.pingFromMs(ms2Name, linuxTap1IP)).To(expected)

		// TODO: remove trace
		_, err := ctx.execVppctl("clear trace")
		Expect(err).To(BeNil())
		_, err = ctx.execVppctl("trace add", "virtio-input 100")
		Expect(err).To(BeNil())

		tcpConnStatus := ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, false)
		udpConnStatus := ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8000, 8000, true)

		stdout, err := ctx.execVppctl("show trace")
		Expect(err).To(BeNil())
		fmt.Println(stdout)

		Expect(tcpConnStatus).To(expected)
		fmt.Println(">> TCP CONNECTION STATUS", tcpConnStatus)
		Expect(udpConnStatus).To(expected)
		fmt.Println(">> UDP CONNECTION STATUS", udpConnStatus)
	}

	ctx.startMicroservice(ms1Name)
	ctx.startMicroservice(ms2Name)
	checkConfig(false)
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppTap1,
		linuxTap1,
		vppTap2,
		linuxTap2,
		ms2DefaultRoute,
		sourceNat,
	).Send(context.Background())
	Expect(err).To(BeNil(), "Transaction creating public and private networks failed")

	Eventually(ctx.getValueStateClb(vppTap1), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	Eventually(ctx.getValueStateClb(vppTap2), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	// check configuration
	checkConfig(true)
	checkConn(true)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// remove S-NAT configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Delete(
		sourceNat,
	).Send(context.Background())
	Expect(err).To(BeNil(), "Transaction removing S-NAT failed")

	// check configuration
	checkConfig(false)
	checkConn(false)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// get back the S-NAT configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(
		sourceNat,
	).Send(context.Background())
	Expect(err).To(BeNil(), "Transaction creating S-NAT failed")
	checkConn(true)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart microservice with S-NAT attached
	ctx.stopMicroservice(ms1Name)
	Eventually(ctx.getValueStateClb(vppTap1), msUpdateTimeout).Should(Equal(kvs.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.startMicroservice(ms1Name)
	Eventually(ctx.getValueStateClb(vppTap1), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")

	// check configuration
	checkConfig(true)
	checkConn(true)
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
		udpSvcExtIP     = "80.80.80.30"
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

	natGlobal := &vpp_nat.Nat44Global{
		Forwarding: true,
		NatInterfaces: []*vpp_nat.Nat44Global_Interface{
			{
				Name:          vppTap1Name,
				IsInside:      false,
			},
			{
				Name:          vppTap2Name,
				IsInside:      true,
			},
		},
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

	checkConfig := func(isConfigured bool) {
		substringMatcher := func(substring string, negative bool) types.GomegaMatcher {
			if negative {
				return Not(ContainSubstring(substring))
			} else {
				return ContainSubstring(substring)
			}
		}
		stdout, err := ctx.execVppctl("show", "nat44", "static", "mappings")
		Expect(err).To(BeNil(), "Running `vppctl show nat44 addresses` failed")
		Expect(stdout).To(substringMatcher("bla bla", !isConfigured),
			"Unexpected S-NAT address configuration`",
		)
		Expect(stdout).To(substringMatcher("bla bla", !isConfigured),
			"Unexpected S-NAT address configuration`",
		)
	}

	checkConn := func(shouldSucceed bool) {
		expected := BeNil()
		if !shouldSucceed {
			expected = Not(BeNil())
		}

		// TODO: remove trace
		_, err := ctx.execVppctl("clear trace")
		Expect(err).To(BeNil())
		_, err = ctx.execVppctl("trace add", "virtio-input 100")
		Expect(err).To(BeNil())

		tcpConnStatus := ctx.testConnection(ms1Name, ms2Name, tcpSvcExtIP, linuxTap2IP,
			tcpSvcExtPort, tcpSvcLocalPort, false)
		udpConnStatus := ctx.testConnection(ms1Name, ms2Name, udpSvcExtIP, linuxTap2IP,
			udpSvcExtPort, udpSvcLocalPort, false)

		stdout, err := ctx.execVppctl("show trace")
		Expect(err).To(BeNil())
		fmt.Println(stdout)

		Expect(tcpConnStatus).To(expected)
		fmt.Println(">> TCP CONNECTION STATUS", tcpConnStatus)
		Expect(udpConnStatus).To(expected)
		fmt.Println(">> UDP CONNECTION STATUS", udpConnStatus)
	}

	ctx.startMicroservice(ms1Name)
	ctx.startMicroservice(ms2Name)
	checkConfig(false)
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppTap1,
		linuxTap1,
		vppTap2,
		linuxTap2,
		natGlobal,
		tcpSvc,	udpSvc,
	).Send(context.Background())
	Expect(err).To(BeNil(), "Transaction creating public and private networks failed")

	Eventually(ctx.getValueStateClb(vppTap1), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	Eventually(ctx.getValueStateClb(vppTap2), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	// check configuration
	checkConfig(true)
	checkConn(true)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// remove static mappings
	req = ctx.grpcClient.ChangeRequest()
	err = req.Delete(
		tcpSvc, udpSvc,
	).Send(context.Background())
	Expect(err).To(BeNil(), "Transaction removing NAT static mappings failed")

	// check configuration
	checkConfig(false)
	checkConn(false)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// get back the NAT configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(
		tcpSvc, udpSvc,
	).Send(context.Background())
	Expect(err).To(BeNil(), "Transaction creating NAT static mappings failed")
	checkConn(true)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart both microservices
	ctx.stopMicroservice(ms1Name)
	ctx.stopMicroservice(ms2Name)
	Eventually(ctx.getValueStateClb(vppTap1), msUpdateTimeout).Should(Equal(kvs.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	Eventually(ctx.getValueStateClb(vppTap2), msUpdateTimeout).Should(Equal(kvs.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.startMicroservice(ms1Name)
	ctx.startMicroservice(ms2Name)
	Eventually(ctx.getValueStateClb(vppTap1), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")
	Eventually(ctx.getValueStateClb(vppTap2), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")

	// check configuration
	checkConfig(true)
	checkConn(true)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")
}

