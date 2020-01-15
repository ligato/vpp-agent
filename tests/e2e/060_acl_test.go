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
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func vppACLs(ctx *testCtx) (string, error) {
	return ctx.execVppctl("show", "acl-plugin", "acl")
}

// Test access control between microservices connected over VPP on the L3 layer.
func TestL3ACLs(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		// microservice1
		vppTap1Name       = "vpp-tap1"
		vppTap1IP         = "10.10.10.1"
		linuxTap1Name     = "linux-tap1"
		linuxTap1Hostname = "tap"
		linuxTap1IP       = "10.10.10.10"

		// private network
		vppTap2Name       = "vpp-tap2"
		vppTap2IP         = "192.168.1.1"
		linuxTap2Name     = "linux-tap2"
		linuxTap2Hostname = "tap"
		linuxTap2IP       = "192.168.1.2"

		netMask = "/24"
		ms1Name = "microservice1"
		ms2Name = "microservice2"

		// acl
		ms1BlockedUDPPort = 9000
		ms2BlockedTCPPort = 8000
		ms1Net            = "10.10.10.0" + netMask
		ms2Net            = "192.168.1.0" + netMask
		ms1IngressACLName = "ms1-ingress"
		ms1EgressACLName  = "ms1-egress"
		ms2IngressACLName = "ms2-ingress"
		ms2EgressACLName  = "ms2-egress"
		anyAddr           = "0.0.0.0/0"
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

	ms1DefaultRoute := &linux_l3.Route{
		OutgoingInterface: linuxTap1Name,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        "0.0.0.0/0",
		GwAddr:            vppTap1IP,
	}

	ms2DefaultRoute := &linux_l3.Route{
		OutgoingInterface: linuxTap2Name,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        "0.0.0.0/0",
		GwAddr:            vppTap2IP,
	}

	permitAll := &vpp_acl.ACL_Rule{
		Action: vpp_acl.ACL_Rule_PERMIT,
		IpRule: &vpp_acl.ACL_Rule_IpRule{
			Ip: &vpp_acl.ACL_Rule_IpRule_Ip{
				SourceNetwork:      anyAddr,
				DestinationNetwork: anyAddr,
			},
		},
	}

	anyPort := &vpp_acl.ACL_Rule_IpRule_PortRange{
		LowerPort: 0,
		UpperPort: uint32(^uint16(0)),
	}

	// connections initiated from microservice1 should not be blocked by
	// ACL on the egress side (from the VPP point of view)
	ms1IngressACL := &vpp_acl.ACL{
		Name: ms1IngressACLName,
		Rules: []*vpp_acl.ACL_Rule{
			{
				Action: vpp_acl.ACL_Rule_REFLECT,
				IpRule: &vpp_acl.ACL_Rule_IpRule{
					Ip: &vpp_acl.ACL_Rule_IpRule_Ip{
						SourceNetwork:      anyAddr,
						DestinationNetwork: anyAddr,
					},
				},
			},
		},
		Interfaces: &vpp_acl.ACL_Interfaces{
			Ingress: []string{
				vppTap1Name,
			},
		},
	}
	showMs1IngressACL := "{ms1-ingress}\r\n" +
		"          0: ipv4 permit+reflect src 0.0.0.0/0 dst 0.0.0.0/0 proto 0 sport 0 dport 0\r\n"

	// microservice2 is not allowed to ping microservice1 and it also cannot
	// send UDP packet to port 9000
	ms1EgressACL := &vpp_acl.ACL{
		Name: ms1EgressACLName,
		Rules: []*vpp_acl.ACL_Rule{
			{
				Action: vpp_acl.ACL_Rule_DENY,
				IpRule: &vpp_acl.ACL_Rule_IpRule{
					Ip: &vpp_acl.ACL_Rule_IpRule_Ip{
						SourceNetwork:      ms2Net,
						DestinationNetwork: anyAddr,
					},
					Udp: &vpp_acl.ACL_Rule_IpRule_Udp{
						SourcePortRange: anyPort,
						DestinationPortRange: &vpp_acl.ACL_Rule_IpRule_PortRange{
							LowerPort: ms1BlockedUDPPort,
							UpperPort: ms1BlockedUDPPort,
						},
					},
				},
			},
			{
				Action: vpp_acl.ACL_Rule_DENY,
				IpRule: &vpp_acl.ACL_Rule_IpRule{
					Ip: &vpp_acl.ACL_Rule_IpRule_Ip{
						SourceNetwork:      anyAddr,
						DestinationNetwork: anyAddr,
					},
					Icmp: &vpp_acl.ACL_Rule_IpRule_Icmp{
						// any ICMP packet
						IcmpCodeRange: &vpp_acl.ACL_Rule_IpRule_Icmp_Range{
							First: 0,
							Last:  255,
						},
						IcmpTypeRange: &vpp_acl.ACL_Rule_IpRule_Icmp_Range{
							First: 0,
							Last:  255,
						},
					},
				},
			},
			permitAll, // permit the rest
		},
		Interfaces: &vpp_acl.ACL_Interfaces{
			Egress: []string{
				vppTap1Name,
			},
		},
	}
	showMs1EgressACL := fmt.Sprintf(
		"{ms1-egress}\r\n"+
			"          0: ipv4 deny src %s dst 0.0.0.0/0 proto 17 sport 0-65535 dport %d\r\n"+
			"          1: ipv4 deny src 0.0.0.0/0 dst 0.0.0.0/0 proto 1 sport 0-255 dport 0-255\r\n"+
			"          2: ipv4 permit src 0.0.0.0/0 dst 0.0.0.0/0 proto 0 sport 0 dport 0\r\n",
		ms2Net, ms1BlockedUDPPort)

	// microservice2 is not allowed to initiate TCP connections to ms1 on well-known
	// ports (<1024)
	ms2IngressACL := &vpp_acl.ACL{
		Name: ms2IngressACLName,
		Rules: []*vpp_acl.ACL_Rule{
			{
				Action: vpp_acl.ACL_Rule_DENY,
				IpRule: &vpp_acl.ACL_Rule_IpRule{
					Ip: &vpp_acl.ACL_Rule_IpRule_Ip{
						SourceNetwork:      anyAddr,
						DestinationNetwork: ms1Net,
					},
					Tcp: &vpp_acl.ACL_Rule_IpRule_Tcp{
						SourcePortRange: anyPort,
						DestinationPortRange: &vpp_acl.ACL_Rule_IpRule_PortRange{
							LowerPort: 0,
							UpperPort: 1023,
						},
					},
				},
			},
			permitAll, // permit the rest
		},
		Interfaces: &vpp_acl.ACL_Interfaces{
			Ingress: []string{
				vppTap2Name,
			},
		},
	}
	showMs2IngressACL := fmt.Sprintf(
		"{ms2-ingress}\r\n"+
			"          0: ipv4 deny src 0.0.0.0/0 dst %s proto 6 sport 0-65535 dport 0-1023\r\n"+
			"          1: ipv4 permit src 0.0.0.0/0 dst 0.0.0.0/0 proto 0 sport 0 dport 0\r\n",
		ms1Net)

	// microservice1 is not allowed to connect to microservice2 on TCP port 8000
	ms2EgressACL := &vpp_acl.ACL{
		Name: ms2EgressACLName,
		Rules: []*vpp_acl.ACL_Rule{
			{
				Action: vpp_acl.ACL_Rule_DENY,
				IpRule: &vpp_acl.ACL_Rule_IpRule{
					Ip: &vpp_acl.ACL_Rule_IpRule_Ip{
						SourceNetwork:      linuxTap1IP + "/32",
						DestinationNetwork: anyAddr,
					},
					Tcp: &vpp_acl.ACL_Rule_IpRule_Tcp{
						SourcePortRange: anyPort,
						DestinationPortRange: &vpp_acl.ACL_Rule_IpRule_PortRange{
							LowerPort: ms2BlockedTCPPort,
							UpperPort: ms2BlockedTCPPort,
						},
					},
				},
			},
			permitAll, // permit the rest
		},
		Interfaces: &vpp_acl.ACL_Interfaces{
			Egress: []string{
				vppTap2Name,
			},
		},
	}
	showMs2EgressACL := fmt.Sprintf(
		"{ms2-egress}\r\n"+
			"          0: ipv4 deny src %s/32 dst 0.0.0.0/0 proto 6 sport 0-65535 dport %d\r\n"+
			"          1: ipv4 permit src 0.0.0.0/0 dst 0.0.0.0/0 proto 0 sport 0 dport 0\r\n",
		linuxTap1IP, ms2BlockedTCPPort)

	checkAccess := func(aclsConfigured bool) {
		beAllowed := Succeed()
		beBlocked := Not(Succeed())
		if !aclsConfigured {
			beBlocked = Succeed()
		}

		// ICMP
		ExpectWithOffset(1, ctx.pingFromMs(ms1Name, linuxTap2IP)).To(beAllowed) // reflected by ms1IngressACL
		ExpectWithOffset(1, ctx.pingFromMs(ms2Name, linuxTap1IP)).To(beBlocked) // blocked by ms1EgressACL

		// TCP
		ExpectWithOffset(1, ctx.testConnection(ms1Name, ms2Name, linuxTap2IP, linuxTap2IP,
			ms2BlockedTCPPort, ms2BlockedTCPPort, false, tapv2InputNode)).To(beBlocked) // blocked by ms2EgressACL
		ExpectWithOffset(1, ctx.testConnection(ms1Name, ms2Name, linuxTap2IP, linuxTap2IP,
			8080, 8080, false, tapv2InputNode)).To(beAllowed)
		ExpectWithOffset(1, ctx.testConnection(ms1Name, ms2Name, linuxTap2IP, linuxTap2IP,
			80, 80, false, tapv2InputNode)).To(beAllowed)
		ExpectWithOffset(1, ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			ms2BlockedTCPPort, ms2BlockedTCPPort, false, tapv2InputNode)).To(beAllowed)
		ExpectWithOffset(1, ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			8080, 8080, false, tapv2InputNode)).To(beAllowed)
		ExpectWithOffset(1, ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			80, 80, false, tapv2InputNode)).To(beBlocked) // blocked by ms2IngressACL

		// UDP
		ExpectWithOffset(1, ctx.testConnection(ms1Name, ms2Name, linuxTap2IP, linuxTap2IP,
			ms1BlockedUDPPort, ms1BlockedUDPPort, true, tapv2InputNode)).To(beAllowed)
		ExpectWithOffset(1, ctx.testConnection(ms1Name, ms2Name, linuxTap2IP, linuxTap2IP,
			9999, 9999, true, tapv2InputNode)).To(beAllowed)
		ExpectWithOffset(1, ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			ms1BlockedUDPPort, ms1BlockedUDPPort, true, tapv2InputNode)).To(beBlocked) // blocked by ms1EgressACL
		ExpectWithOffset(1, ctx.testConnection(ms2Name, ms1Name, linuxTap1IP, linuxTap1IP,
			9999, 9999, true, tapv2InputNode)).To(beAllowed)
	}

	ctx.startMicroservice(ms1Name)
	ctx.startMicroservice(ms2Name)
	Expect(vppACLs(ctx)).ShouldNot(SatisfyAny(
		ContainSubstring(showMs1IngressACL), ContainSubstring(showMs1EgressACL),
		ContainSubstring(showMs2IngressACL), ContainSubstring(showMs2EgressACL)))
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppTap1,
		linuxTap1,
		vppTap2,
		linuxTap2,
		ms1DefaultRoute, ms2DefaultRoute,
		ms1IngressACL, ms1EgressACL,
		ms2IngressACL, ms2EgressACL,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction connecting microservices and configuring ACLs failed")

	Eventually(ctx.getValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	Eventually(ctx.getValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	Expect(vppACLs(ctx)).Should(SatisfyAll(
		ContainSubstring(showMs1IngressACL), ContainSubstring(showMs1EgressACL),
		ContainSubstring(showMs2IngressACL), ContainSubstring(showMs2EgressACL)))
	checkAccess(true)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// remove ACL configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Delete(
		ms1IngressACL, ms1EgressACL,
		ms2IngressACL, ms2EgressACL,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction removing ACLs failed")

	Expect(vppACLs(ctx)).ShouldNot(SatisfyAny(
		ContainSubstring(showMs1IngressACL), ContainSubstring(showMs1EgressACL),
		ContainSubstring(showMs2IngressACL), ContainSubstring(showMs2EgressACL)))
	checkAccess(false)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

	// get back the ACL configuration
	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(
		ms1IngressACL, ms1EgressACL,
		ms2IngressACL, ms2EgressACL,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Transaction creating ACLs failed")

	Expect(vppACLs(ctx)).Should(SatisfyAll(
		ContainSubstring(showMs1IngressACL), ContainSubstring(showMs1EgressACL),
		ContainSubstring(showMs2IngressACL), ContainSubstring(showMs2EgressACL)))
	checkAccess(true)
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

	Expect(vppACLs(ctx)).Should(SatisfyAll(
		ContainSubstring(showMs1IngressACL), ContainSubstring(showMs1EgressACL),
		ContainSubstring(showMs2IngressACL), ContainSubstring(showMs2EgressACL)))
	checkAccess(true)
	Expect(ctx.agentInSync()).To(BeTrue(), "Agent is not in-sync")

}

// TODO: L2 ACLs
