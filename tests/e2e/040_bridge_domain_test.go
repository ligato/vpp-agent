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
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

func bridgeDomains(ctx *TestCtx) ([]map[string]string, error) {
	stdout, err := ctx.ExecVppctl("show", "bridge-domain")
	if err != nil {
		return nil, err
	}
	return parseVPPTable(stdout), nil
}

func bdAgeIs(min int) types.GomegaMatcher {
	if min == 0 {
		return HaveKeyWithValue("Age(min)", "off")
	}
	return HaveKeyWithValue("Age(min)", strconv.Itoa(min))
}

func bdWithFlooding() types.GomegaMatcher {
	return And(
		HaveKeyWithValue("UU-Flood", "flood"),
		HaveKeyWithValue("Flooding", "on"))
}

func bdWithForwarding() types.GomegaMatcher {
	return HaveKeyWithValue("U-Forwrd", "on")
}

func bdWithLearning() types.GomegaMatcher {
	return HaveKeyWithValue("Learning", "on")
}

// connect microservices into the same L2 network segment via bridge domain
// and TAP interfaces.
func TestBridgeDomainWithTAPs(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		vppTap1Name       = "vpp-tap1"
		linuxTap1Name     = "linux-tap1"
		linuxTap1Hostname = "tap"
		linuxTap1IP       = "192.168.1.2"

		vppTap2Name       = "vpp-tap2"
		linuxTap2Name     = "linux-tap2"
		linuxTap2Hostname = "tap"
		linuxTap2IP       = "192.168.1.3"

		vppLoopbackName = "loop1"
		vppLoopbackIP   = "192.168.1.1"

		netMask = "/24"
		ms1Name = "microservice1"
		ms2Name = "microservice2"
		bdName  = "my-bd"
	)

	vppTap1 := &vpp_interfaces.Interface{
		Name:    vppTap1Name,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
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
		Name:    vppTap2Name,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
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

	vppLoop := &vpp_interfaces.Interface{
		Name:        vppLoopbackName,
		Type:        vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		IpAddresses: []string{vppLoopbackIP + netMask},
	}

	bd := &vpp_l2.BridgeDomain{
		Name:                bdName,
		Flood:               true,
		Forward:             true,
		Learn:               true,
		UnknownUnicastFlood: true,
		Interfaces: []*vpp_l2.BridgeDomain_Interface{
			{
				Name: vppTap1Name,
			},
			{
				Name: vppTap2Name,
			},
			{
				Name:                    vppLoopbackName,
				SplitHorizonGroup:       1,
				BridgedVirtualInterface: true,
			},
		},
	}

	ctx.StartMicroservice(ms1Name)
	ctx.StartMicroservice(ms2Name)
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		vppTap1,
		linuxTap1,
		vppTap2,
		linuxTap2,
		vppLoop,
		bd,
	).Send(context.Background())
	ctx.Expect(err).To(BeNil(), "Transaction creating BD with TAPs failed")

	ctx.Expect(ctx.GetValueState(vppLoop)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"BD BVI should be configured even before microservices start")
	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice1 should be eventually configured")
	ctx.Eventually(ctx.GetValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TAP attached to a newly started microservice2 should be eventually configured")

	if ctx.VppRelease() < "21.01" { // "show bridge-domain" hard to parse in VPP>=21.01 (https://jira.fd.io/browse/VPP-1969)
		bds, err := bridgeDomains(ctx)
		ctx.Expect(err).ToNot(HaveOccurred())
		ctx.Expect(bds).To(HaveLen(1))
		ctx.Expect(bds[0]).To(SatisfyAll(
			bdAgeIs(0), bdWithFlooding(), bdWithForwarding(), bdWithLearning()))
	}

	ctx.Expect(ctx.PingFromMs(ms2Name, linuxTap1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, linuxTap2IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms2Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(linuxTap1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(linuxTap2IP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// kill one of the microservices
	// - "Eventually" is also used with linuxTap1 to wait for retry txn that
	//   will change state from RETRYING to PENDING
	ctx.StopMicroservice(ms1Name)
	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VPP-TAP should be pending")
	ctx.Eventually(ctx.GetValueStateClb(linuxTap1)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated LinuxTAP should be pending")
	ctx.Expect(ctx.GetValueState(vppTap2)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to running microservice is not configured")
	ctx.Expect(ctx.GetValueState(linuxTap2)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"Linux-TAP attached to running microservice is not configured")
	ctx.Expect(ctx.GetValueState(vppLoop)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"BD BVI interface is not configured")
	ctx.Expect(ctx.GetValueState(bd)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"BD is not configured")

	ctx.Expect(ctx.PingFromMs(ms2Name, linuxTap1IP)).ToNot(Succeed())
	ctx.Expect(ctx.PingFromMs(ms2Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(linuxTap1IP)).ToNot(Succeed())
	ctx.Expect(ctx.PingFromVPP(linuxTap2IP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart the microservice
	ctx.StartMicroservice(ms1Name)
	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"VPP-TAP attached to a re-started microservice1 should be eventually configured")
	ctx.Expect(ctx.GetValueState(linuxTap1)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"Linux-TAP attached to a re-started microservice1 is not configured")

	// Waiting for TAP interface after restart
	// See: https://github.com/ligato/vpp-agent/issues/1489
	ctx.Eventually(ctx.PingFromMsClb(ms2Name, linuxTap1IP), "18s", "2s").Should(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, linuxTap2IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms2Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(linuxTap1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(linuxTap2IP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// change bridge domain config to trigger re-creation
	bd.MacAge = 10
	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		bd,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction updating BD failed")

	ctx.Expect(ctx.PingFromMs(ms2Name, linuxTap1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, linuxTap2IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms2Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(linuxTap1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(linuxTap2IP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	if ctx.VppRelease() < "21.01" { // "show bridge-domain" hard to parse in VPP>=21.01 (https://jira.fd.io/browse/VPP-1969)
		bds, err := bridgeDomains(ctx)
		ctx.Expect(err).ToNot(HaveOccurred())
		ctx.Expect(bds).To(HaveLen(1))
		ctx.Expect(bds[0]).To(SatisfyAll(
			bdAgeIs(10), bdWithFlooding(), bdWithForwarding(), bdWithLearning()))
	}
}

// connect microservices into the same L2 network segment via bridge domain
// and AF-PACKET+VETH interfaces.
func TestBridgeDomainWithAfPackets(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		afPacket1Name  = "vpp-afpacket1"
		veth1AName     = "vpp-veth-1a"
		veth1BName     = "vpp-veth-1b"
		veth1AHostname = "veth1a"
		veth1BHostname = "veth1b"
		veth1IP        = "192.168.1.2"

		afPacket2Name  = "vpp-afpacket2"
		veth2AName     = "vpp-veth-2a"
		veth2BName     = "vpp-veth-2b"
		veth2AHostname = "veth2a"
		veth2BHostname = "veth2b"
		veth2IP        = "192.168.1.3"

		vppLoopbackName = "loop1"
		vppLoopbackIP   = "192.168.1.1"

		netMask = "/24"
		ms1Name = "microservice1"
		ms2Name = "microservice2"
		bdName  = "my-bd"
	)

	afPacket1 := &vpp_interfaces.Interface{
		Name:    afPacket1Name,
		Type:    vpp_interfaces.Interface_AF_PACKET,
		Enabled: true,
		Link: &vpp_interfaces.Interface_Afpacket{
			Afpacket: &vpp_interfaces.AfpacketLink{
				HostIfName: veth1BHostname,
			},
		},
	}

	veth1a := &linux_interfaces.Interface{
		Name:        veth1AName,
		Type:        linux_interfaces.Interface_VETH,
		Enabled:     true,
		HostIfName:  veth1AHostname,
		IpAddresses: []string{veth1IP + netMask},
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: veth1BName,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + ms1Name,
		},
	}

	veth1b := &linux_interfaces.Interface{
		Name:       veth1BName,
		Type:       linux_interfaces.Interface_VETH,
		Enabled:    true,
		HostIfName: veth1BHostname,
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: veth1AName,
			},
		},
	}

	afPacket2 := &vpp_interfaces.Interface{
		Name:    afPacket2Name,
		Type:    vpp_interfaces.Interface_AF_PACKET,
		Enabled: true,
		Link: &vpp_interfaces.Interface_Afpacket{
			Afpacket: &vpp_interfaces.AfpacketLink{
				HostIfName: veth2BHostname,
			},
		},
	}

	veth2a := &linux_interfaces.Interface{
		Name:        veth2AName,
		Type:        linux_interfaces.Interface_VETH,
		Enabled:     true,
		HostIfName:  veth2AHostname,
		IpAddresses: []string{veth2IP + netMask},
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: veth2BName,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + ms2Name,
		},
	}

	veth2b := &linux_interfaces.Interface{
		Name:       veth2BName,
		Type:       linux_interfaces.Interface_VETH,
		Enabled:    true,
		HostIfName: veth2BHostname,
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: veth2AName,
			},
		},
	}

	vppLoop := &vpp_interfaces.Interface{
		Name:        vppLoopbackName,
		Type:        vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		IpAddresses: []string{vppLoopbackIP + netMask},
	}

	bd := &vpp_l2.BridgeDomain{
		Name:                bdName,
		Flood:               true,
		Forward:             true,
		Learn:               true,
		UnknownUnicastFlood: true,
		Interfaces: []*vpp_l2.BridgeDomain_Interface{
			{
				Name: afPacket1Name,
			},
			{
				Name: afPacket2Name,
			},
			{
				Name:                    vppLoopbackName,
				SplitHorizonGroup:       1,
				BridgedVirtualInterface: true,
			},
		},
	}

	ctx.StartMicroservice(ms1Name)
	ctx.StartMicroservice(ms2Name)
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		afPacket1,
		veth1a, veth1b,
		afPacket2,
		veth2a, veth2b,
		vppLoop,
		bd,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction creating BD with AF-PACKETs failed")

	ctx.Expect(ctx.GetValueState(vppLoop)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"BD BVI should be configured even before microservices start")
	ctx.Eventually(ctx.GetValueStateClb(afPacket1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"AF-PACKET attached to a newly started microservice1 should be eventually configured")
	ctx.Eventually(ctx.GetValueStateClb(afPacket2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"AF-PACKET attached to a newly started microservice2 should be eventually configured")

	if ctx.VppRelease() < "21.01" { // "show bridge-domain" hard to parse in VPP>=21.01 (https://jira.fd.io/browse/VPP-1969)
		bds, err := bridgeDomains(ctx)
		ctx.Expect(err).ToNot(HaveOccurred())
		ctx.Expect(bds).To(HaveLen(1))
		ctx.Expect(bds[0]).To(SatisfyAll(
			bdAgeIs(0), bdWithFlooding(), bdWithForwarding(), bdWithLearning()))
	}

	ctx.Expect(ctx.PingFromMs(ms2Name, veth1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, veth2IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms2Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(veth1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(veth2IP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// kill one of the microservices
	// - both AF-PACKET and VETH use separate "Eventually" assertion since
	//   they react to different SB notifications
	ctx.StopMicroservice(ms1Name)
	ctx.Eventually(ctx.GetValueStateClb(afPacket1)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated AF-PACKET should be pending")
	ctx.Eventually(ctx.GetValueStateClb(veth1a)).Should(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VETH should be pending")
	ctx.Expect(ctx.GetValueState(veth1b)).To(Equal(kvscheduler.ValueState_PENDING),
		"Without microservice, the associated VETH should be pending")
	ctx.Expect(ctx.GetValueState(afPacket2)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"AF-PACKET attached to running microservice is not configured")
	ctx.Expect(ctx.GetValueState(veth2a)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"VETH attached to running microservice is not configured")
	ctx.Expect(ctx.GetValueState(veth2b)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"VETH attached to running microservice is not configured")
	ctx.Expect(ctx.GetValueState(vppLoop)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"BD BVI interface is not configured")
	ctx.Expect(ctx.GetValueState(bd)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"BD is not configured")

	ctx.Expect(ctx.PingFromMs(ms2Name, veth1IP)).ToNot(Succeed())
	ctx.Expect(ctx.PingFromMs(ms2Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(veth1IP)).ToNot(Succeed())
	ctx.Expect(ctx.PingFromVPP(veth2IP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// restart the microservice
	ctx.StartMicroservice(ms1Name)
	ctx.Eventually(ctx.GetValueStateClb(afPacket1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"AF-PACKET attached to a re-started microservice1 should be eventually configured")
	ctx.Expect(ctx.GetValueState(veth1a)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"VETH attached to re-started microservice1 is not configured")
	ctx.Expect(ctx.GetValueState(veth1b)).To(Equal(kvscheduler.ValueState_CONFIGURED),
		"VETH attached to re-started microservice1 is not configured")

	// Waiting for AF-PACKET interface after restart
	// See: https://github.com/ligato/vpp-agent/issues/1489
	ctx.Eventually(ctx.PingFromMsClb(ms2Name, veth1IP), "18s", "2s").Should(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, veth2IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms2Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(veth1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(veth2IP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")

	// change bridge domain config to trigger re-creation
	bd.MacAge = 10
	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		bd,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Transaction updating BD failed")

	if ctx.VppRelease() < "21.01" { // "show bridge-domain" hard to parse in VPP>=21.01 (https://jira.fd.io/browse/VPP-1969)
		bds, err := bridgeDomains(ctx)
		ctx.Expect(err).ToNot(HaveOccurred())
		ctx.Expect(bds).To(HaveLen(1))
		ctx.Expect(bds[0]).To(SatisfyAll(
			bdAgeIs(10), bdWithFlooding(), bdWithForwarding(), bdWithLearning()))
	}

	ctx.Expect(ctx.PingFromMs(ms2Name, veth1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, veth2IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms1Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(ms2Name, vppLoopbackIP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(veth1IP)).To(Succeed())
	ctx.Expect(ctx.PingFromVPP(veth2IP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue(), "Agent is not in-sync")
}
