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
	"os"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
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

//
// +---------------------------------------------------+
// | VPP                                               |
// |                                                   |
// |     +---------------+       +---------------+     |
// |     |192.168.1.10/24|       |192.168.2.20/24|     |
// |     +-------+-------+       +-------+-------+     |
// |             | (SUBIF)               | (SUBIF)     |
// |     +-------+-----------------------+-------+     |
// |     |                                       |     |
// |     +-------------------+-------------------+     |
// +-------------------------|-------------------------+
//                           | (MEMIF)
// +-------------------------|-------------------------+
// |     +-------------------+-------------------+     |
// |     |                                       |     |
// |     +-------+-----------------------+-------+     |
// |             | (SUBIF)               | (SUBIF)     |
// |     +-------+-------+       +-------+-------+     |
// |     |192.168.1.10/24|       |192.168.2.20/24|     |
// |     +-------+-------+       +-------+-------+     |
// |             |                       |             |
// | VPP         | (BD)                  | (BD)        |
// |             |                       |             |
// |     +-------+-------+       +-------+-------+     |
// |     |192.168.1.11/24|       |192.168.2.22/24|     |
// |     +-------+-------+       +-------+-------+     |
// +-------------|-----------------------|-------------+
//               | (TAP)                 | (TAP)
// +-------------|----------+ +----------|-------------+
// |     +-------+-------+  | |  +-------+-------+     |
// |     |192.168.1.11/24|  | |  |192.168.2.22/24|     |
// |     +-------+-------+  | |  +-------+-------+     |
// |                        | |                        |
// | LINUX                  | |                  LINUX |
// +------------------------+ +------------------------+
//
func TestMemifSubinterfaceVlanConn(t *testing.T) {
	ctx := Setup(t, WithoutVPPAgent())
	defer ctx.Teardown()

	const (
		vpp1MemifName  = "vpp1-to-vpp2"
		vpp2MemifName  = "vpp2-to-vpp1"
		vpp1Subif1Name = "vpp1Subif1"
		vpp1Subif2Name = "vpp1Subif2"
		vpp1Subif1IP   = "192.168.1.10"
		vpp1Subif2IP   = "192.168.2.20"
		vpp2Subif1Name = "vpp2Subif1"
		vpp2Subif2Name = "vpp2Subif2"
		vpp2Tap1Name   = "vpp2-to-ms1"
		vpp2Tap2Name   = "vpp2-to-ms2"
		ms1TapName     = "ms1-to-vpp2"
		ms2TapName     = "ms2-to-vpp2"
		ms1TapIP       = "192.168.1.11"
		ms2TapIP       = "192.168.2.22"
		bd1Name        = "bd1"
		bd2Name        = "bd2"
		ms1Name        = "ms1"
		ms2Name        = "ms2"
		agent1Name     = "agent1"
		agent2Name     = "agent2"
		netMask        = "/24"
		msTapHostname  = "tap"
		memifFilepath  = "/test-memif-subif-vlan-conn/memif/"
		memifSockname  = "memif.sock"
	)

	if err := os.MkdirAll(shareDir+memifFilepath, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	// agent1 configuration
	vpp1Memif := &vpp_interfaces.Interface{
		Name:    vpp1MemifName,
		Type:    vpp_interfaces.Interface_MEMIF,
		Enabled: true,
		Link: &vpp_interfaces.Interface_Memif{
			Memif: &vpp_interfaces.MemifLink{
				Master:         true,
				Id:             1,
				SocketFilename: shareDir + memifFilepath + memifSockname,
			},
		},
	}
	vpp1Subif1 := &vpp_interfaces.Interface{
		Name:        vpp1Subif1Name,
		Type:        vpp_interfaces.Interface_SUB_INTERFACE,
		Enabled:     true,
		IpAddresses: []string{vpp1Subif1IP + netMask},
		Link: &vpp_interfaces.Interface_Sub{
			Sub: &vpp_interfaces.SubInterface{
				ParentName:  vpp1MemifName,
				SubId:       10,
				TagRwOption: vpp_interfaces.SubInterface_POP1,
			},
		},
	}
	vpp1Subif2 := &vpp_interfaces.Interface{
		Name:        vpp1Subif2Name,
		Type:        vpp_interfaces.Interface_SUB_INTERFACE,
		Enabled:     true,
		IpAddresses: []string{vpp1Subif2IP + netMask},
		Link: &vpp_interfaces.Interface_Sub{
			Sub: &vpp_interfaces.SubInterface{
				ParentName:  vpp1MemifName,
				SubId:       20,
				TagRwOption: vpp_interfaces.SubInterface_POP1,
			},
		},
	}

	// agent2 configuration
	vpp2Memif := &vpp_interfaces.Interface{
		Name:    vpp2MemifName,
		Type:    vpp_interfaces.Interface_MEMIF,
		Enabled: true,
		Link: &vpp_interfaces.Interface_Memif{
			Memif: &vpp_interfaces.MemifLink{
				Master:         false,
				Id:             1,
				SocketFilename: shareDir + memifFilepath + memifSockname,
			},
		},
	}
	vpp2Subif1 := &vpp_interfaces.Interface{
		Name:    vpp2Subif1Name,
		Type:    vpp_interfaces.Interface_SUB_INTERFACE,
		Enabled: true,
		Link: &vpp_interfaces.Interface_Sub{
			Sub: &vpp_interfaces.SubInterface{
				ParentName:  vpp2MemifName,
				SubId:       10,
				TagRwOption: vpp_interfaces.SubInterface_POP1,
			},
		},
	}
	vpp2Subif2 := &vpp_interfaces.Interface{
		Name:    vpp2Subif2Name,
		Type:    vpp_interfaces.Interface_SUB_INTERFACE,
		Enabled: true,
		Link: &vpp_interfaces.Interface_Sub{
			Sub: &vpp_interfaces.SubInterface{
				ParentName:  vpp2MemifName,
				SubId:       20,
				TagRwOption: vpp_interfaces.SubInterface_POP1,
			},
		},
	}

	vpp2Tap1 := &vpp_interfaces.Interface{
		Name:    vpp2Tap1Name,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: MsNamePrefix + ms1Name,
			},
		},
	}
	vpp2Tap2 := &vpp_interfaces.Interface{
		Name:    vpp2Tap2Name,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: MsNamePrefix + ms2Name,
			},
		},
	}

	ms1Tap := &linux_interfaces.Interface{
		Name:        ms1TapName,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{ms1TapIP + netMask},
		HostIfName:  msTapHostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vpp2Tap1Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + ms1Name,
		},
	}
	ms2Tap := &linux_interfaces.Interface{
		Name:        ms2TapName,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{ms2TapIP + netMask},
		HostIfName:  msTapHostname,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vpp2Tap2Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + ms2Name,
		},
	}

	bd1 := &vpp_l2.BridgeDomain{
		Name:    bd1Name,
		Flood:   true,
		Forward: true,
		Learn:   true,
		Interfaces: []*vpp_l2.BridgeDomain_Interface{
			{
				Name: vpp2Subif1Name,
			},
			{
				Name: vpp2Tap1Name,
			},
		},
	}
	bd2 := &vpp_l2.BridgeDomain{
		Name:    bd2Name,
		Flood:   true,
		Forward: true,
		Learn:   true,
		Interfaces: []*vpp_l2.BridgeDomain_Interface{
			{
				Name: vpp2Subif2Name,
			},
			{
				Name: vpp2Tap2Name,
			},
		},
	}

	ctx.StartMicroservice(ms1Name)
	ctx.StartMicroservice(ms2Name)
	agent1 := ctx.StartAgent(agent1Name)
	agent2 := ctx.StartAgent(agent2Name)

	err := agent1.GenericClient().ChangeRequest().Update(
		vpp1Memif,
		vpp1Subif1,
		vpp1Subif2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())
	err = agent2.GenericClient().ChangeRequest().Update(
		vpp2Memif,
		vpp2Subif1,
		vpp2Subif2,
		vpp2Tap1,
		vpp2Tap2,
		ms1Tap,
		ms2Tap,
		bd1,
		bd2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(agent1.GetValueStateClb(vpp1Memif)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(agent2.GetValueStateClb(ms1Tap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(agent2.GetValueStateClb(ms2Tap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	// Pings from VPP should automatically go through correct vlan
	ctx.Expect(agent1.PingFromVPP(ms1TapIP)).To(Succeed())
	ctx.Expect(agent1.PingFromVPP(ms2TapIP)).To(Succeed())
	// Pings from correct vlan should succeed
	ctx.Expect(agent1.PingFromVPP(ms1TapIP, "source", "memif1/1.10")).To(Succeed())
	ctx.Expect(agent1.PingFromVPP(ms2TapIP, "source", "memif1/1.20")).To(Succeed())
	// Pings from incorrect vlan should fail
	ctx.Expect(agent1.PingFromVPP(ms1TapIP, "source", "memif1/1.10")).NotTo(Succeed())
	ctx.Expect(agent1.PingFromVPP(ms2TapIP, "source", "memif1/1.20")).NotTo(Succeed())
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
