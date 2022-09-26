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
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// connect VPP with a microservice via TAP interface
func TestInterfaceConnTap(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		vppTapName       = "vpp-tap"
		linuxTapName     = "linux-tap"
		linuxTapHostname = "tap"
		vppTapIP         = "192.168.1.1"
		linuxTapIP       = "192.168.1.2"
		netMask          = "/30"
		msName           = "microservice1"
	)

	vppTap := &vpp_interfaces.Interface{
		Name:        vppTapName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTapIP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: MsNamePrefix + msName,
			},
		},
	}
	linuxTap := &linux_interfaces.Interface{
		Name:        linuxTapName,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTapIP + netMask},
		HostIfName:  linuxTapHostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapName,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName,
		},
	}

	ctx.StartMicroservice(msName)

	// configure TAPs
	err := ctx.GenericClient().ChangeRequest().Update(
		vppTap,
		linuxTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.PingFromVPP(linuxTapIP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, vppTapIP)).To(Succeed())

	// resync TAPs
	err = ctx.GenericClient().ResyncConfig(
		vppTap,
		linuxTap,
	)
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.PingFromVPP(linuxTapIP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, vppTapIP)).To(Succeed())

	// restart microservice twice
	for i := 0; i < 2; i++ {
		ctx.StopMicroservice(msName)
		ctx.Eventually(ctx.GetValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Eventually(ctx.GetValueStateClb(linuxTap)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Expect(ctx.PingFromVPP(linuxTapIP)).NotTo(Succeed())
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())

		ctx.StartMicroservice(msName)
		ctx.Eventually(ctx.GetValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.GetValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.PingFromVPP(linuxTapIP)).To(Succeed())
		ctx.Expect(ctx.PingFromMs(msName, vppTapIP)).To(Succeed())
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	}

	// re-create VPP TAP
	err = ctx.GenericClient().ChangeRequest().
		Delete(vppTap).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.PingFromVPP(linuxTapIP)).NotTo(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, vppTapIP)).NotTo(Succeed())

	err = ctx.GenericClient().ChangeRequest().Update(
		vppTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.PingFromVPPClb(linuxTapIP)).Should(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, vppTapIP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	// re-create Linux TAP
	err = ctx.GenericClient().ChangeRequest().Delete(
		linuxTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.PingFromVPP(linuxTapIP)).NotTo(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, vppTapIP)).NotTo(Succeed())

	err = ctx.GenericClient().ChangeRequest().Update(
		linuxTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.PingFromVPPClb(linuxTapIP)).Should(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, vppTapIP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())
}

// connect VPP with a microservice via TAP tunnel interface
// TODO: fix topology setup for a traffic (ping) test
func TestInterfaceTapTunnel(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		vppTapName       = "vpp-taptun"
		linuxTapName     = "linux-taptun"
		linuxTapHostname = "taptun"
		vppTapIP         = "192.168.1.1"
		linuxTapIP       = "192.168.1.2"
		netMask          = "/30"
		msName           = "microservice1"
	)

	vppTap := &vpp_interfaces.Interface{
		Name:        vppTapName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTapIP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				EnableTunnel:   true,
				ToMicroservice: MsNamePrefix + msName,
			},
		},
	}
	linuxTap := &linux_interfaces.Interface{
		Name:        linuxTapName,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTapIP + netMask},
		HostIfName:  linuxTapHostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapName,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName,
		},
	}

	ctx.StartMicroservice(msName)

	// configure TAPs
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		vppTap,
		linuxTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	//ctx.Expect(ctx.pingFromVPP(linuxTapIP)).To(Succeed())
	//ctx.Expect(ctx.pingFromMs(msName, vppTapIP)).To(Succeed())

	ctx.ExecVppctl("show", "int")
	ctx.ExecVppctl("show", "int", "addr")
	ctx.ExecCmd("ip", "link")
	ctx.ExecCmd("ip", "addr")
	ctx.PingFromVPP(linuxTapIP)

	// resync TAPs
	err = ctx.GenericClient().ResyncConfig(
		vppTap,
		linuxTap,
	)
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.ExecVppctl("show", "int")
	ctx.ExecVppctl("show", "int", "addr")
	ctx.ExecCmd("ip", "link")
	ctx.ExecCmd("ip", "addr")

	ctx.Eventually(ctx.GetValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	//ctx.Expect(ctx.pingFromVPP(linuxTapIP)).To(Succeed())
	//ctx.Expect(ctx.pingFromMs(msName, vppTapIP)).To(Succeed())

	// restart microservice twice
	/*for i := 0; i < 2; i++ {
		ctx.stopMicroservice(msName)
		ctx.Eventually(ctx.getValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Eventually(ctx.getValueStateClb(linuxTap)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Expect(ctx.pingFromVPP(linuxTapIP)).NotTo(Succeed())
		ctx.Expect(ctx.agentInSync()).To(BeTrue())

		ctx.startMicroservice(msName)
		ctx.Eventually(ctx.getValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.getValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.pingFromVPP(linuxTapIP)).To(Succeed())
		ctx.Expect(ctx.pingFromMs(msName, vppTapIP)).To(Succeed())
		ctx.Expect(ctx.agentInSync()).To(BeTrue())
	}

	// re-create VPP TAP
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		vppTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.pingFromVPP(linuxTapIP)).NotTo(Succeed())
	ctx.Expect(ctx.pingFromMs(msName, vppTapIP)).NotTo(Succeed())

	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		vppTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.pingFromVPPClb(linuxTapIP)).Should(Succeed())
	ctx.Expect(ctx.pingFromMs(msName, vppTapIP)).To(Succeed())
	ctx.Expect(ctx.agentInSync()).To(BeTrue())

	// re-create Linux TAP
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		linuxTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.pingFromVPP(linuxTapIP)).NotTo(Succeed())
	ctx.Expect(ctx.pingFromMs(msName, vppTapIP)).NotTo(Succeed())

	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		linuxTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.pingFromVPPClb(linuxTapIP)).Should(Succeed())
	ctx.Expect(ctx.pingFromMs(msName, vppTapIP)).To(Succeed())
	ctx.Expect(ctx.agentInSync()).To(BeTrue())*/
}

// connect VPP with a microservice via AF-PACKET + VETH interfaces
func TestInterfaceConnAfPacket(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		afPacketName  = "vpp-afpacket"
		veth1Name     = "linux-veth1"
		veth2Name     = "linux-veth2"
		veth1Hostname = "veth1"
		veth2Hostname = "veth2"
		afPacketIP    = "192.168.1.1"
		veth2IP       = "192.168.1.2"
		netMask       = "/30"
		msName        = "microservice1"
	)

	afPacket := &vpp_interfaces.Interface{
		Name:        afPacketName,
		Type:        vpp_interfaces.Interface_AF_PACKET,
		Enabled:     true,
		IpAddresses: []string{afPacketIP + netMask},
		Link: &vpp_interfaces.Interface_Afpacket{
			Afpacket: &vpp_interfaces.AfpacketLink{
				HostIfName: veth1Hostname,
			},
		},
	}
	veth1 := &linux_interfaces.Interface{
		Name:       veth1Name,
		Type:       linux_interfaces.Interface_VETH,
		Enabled:    true,
		HostIfName: veth1Hostname,
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: veth2Name,
			},
		},
	}
	veth2 := &linux_interfaces.Interface{
		Name:        veth2Name,
		Type:        linux_interfaces.Interface_VETH,
		Enabled:     true,
		HostIfName:  veth2Hostname,
		IpAddresses: []string{veth2IP + netMask},
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: veth1Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName,
		},
	}

	ctx.StartMicroservice(msName)
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		afPacket,
		veth1,
		veth2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(afPacket)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(veth1)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(veth2)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.PingFromVPP(veth2IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).To(Succeed())

	// restart microservice twice
	for i := 0; i < 2; i++ {
		ctx.StopMicroservice(msName)
		ctx.Eventually(ctx.GetValueStateClb(afPacket)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Eventually(ctx.GetValueStateClb(veth1)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Eventually(ctx.GetValueStateClb(veth2)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Expect(ctx.PingFromVPP(veth2IP)).NotTo(Succeed())
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())

		ctx.StartMicroservice(msName)
		ctx.Eventually(ctx.GetValueStateClb(afPacket)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.GetValueState(veth1)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.GetValueState(veth2)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.PingFromVPP(veth2IP)).To(Succeed())
		ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).To(Succeed())
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	}

	// re-create AF-PACKET
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		afPacket,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.PingFromVPP(veth2IP)).NotTo(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).NotTo(Succeed())

	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		afPacket,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.PingFromVPPClb(veth2IP)).Should(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	// re-create VETH
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		veth2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.PingFromVPP(veth2IP)).NotTo(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).NotTo(Succeed())

	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		veth2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.PingFromVPPClb(veth2IP)).Should(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())
}

// Connect VPP with a microservice via AF-PACKET + VETH interfaces.
// Configure AF-PACKET with logical reference to the target VETH interface.
func TestInterfaceAfPacketWithLogicalReference(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		afPacketName  = "vpp-afpacket"
		veth1Name     = "linux-veth1"
		veth2Name     = "linux-veth2"
		veth1Hostname = "veth1"
		veth2Hostname = "veth2"
		afPacketIP    = "192.168.1.1"
		veth2IP       = "192.168.1.2"
		netMask       = "/30"
		msName        = "microservice1"
	)

	afPacket := &vpp_interfaces.Interface{
		Name:        afPacketName,
		Type:        vpp_interfaces.Interface_AF_PACKET,
		Enabled:     true,
		IpAddresses: []string{afPacketIP + netMask},
		Link: &vpp_interfaces.Interface_Afpacket{
			Afpacket: &vpp_interfaces.AfpacketLink{
				LinuxInterface: veth1Name,
			},
		},
	}
	veth1 := &linux_interfaces.Interface{
		Name:       veth1Name,
		Type:       linux_interfaces.Interface_VETH,
		Enabled:    true,
		HostIfName: veth1Hostname,
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: veth2Name,
			},
		},
	}
	veth2 := &linux_interfaces.Interface{
		Name:        veth2Name,
		Type:        linux_interfaces.Interface_VETH,
		Enabled:     true,
		HostIfName:  veth2Hostname,
		IpAddresses: []string{veth2IP + netMask},
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: veth1Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName,
		},
	}

	ctx.StartMicroservice(msName)
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		afPacket,
		veth1,
		veth2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(afPacket)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(veth1)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(veth2)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.PingFromVPP(veth2IP)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).To(Succeed())

	// restart microservice twice
	for i := 0; i < 2; i++ {
		ctx.StopMicroservice(msName)
		ctx.Eventually(ctx.GetValueStateClb(afPacket)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Eventually(ctx.GetValueStateClb(veth1)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Eventually(ctx.GetValueStateClb(veth2)).Should(Equal(kvscheduler.ValueState_PENDING))
		ctx.Expect(ctx.PingFromVPP(veth2IP)).NotTo(Succeed())
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())

		ctx.StartMicroservice(msName)
		ctx.Eventually(ctx.GetValueStateClb(afPacket)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.GetValueState(veth1)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.GetValueState(veth2)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		ctx.Expect(ctx.PingFromVPP(veth2IP)).To(Succeed())
		ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).To(Succeed())
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	}

	// re-create AF-PACKET
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		afPacket,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.PingFromVPP(veth2IP)).NotTo(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).NotTo(Succeed())

	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		afPacket,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.PingFromVPPClb(veth2IP)).Should(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	// re-create VETH
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		veth2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.PingFromVPP(veth2IP)).NotTo(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).NotTo(Succeed())

	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		veth2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.PingFromVPPClb(veth2IP)).Should(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, afPacketIP)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())
}
