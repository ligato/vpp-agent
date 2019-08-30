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
	"time"

	. "github.com/onsi/gomega"

	"github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/namespace"
	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/client"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

// TODO: running downstream resync in-between restarts/re-creates seems to break stuff (for now commented out)

const (
	msUpdateTimeout  = time.Second * 6
	recreateTimeout  = time.Second * 6
)

// connect VPP with a microservice via TAP interface
func TestTapInterfaceConn(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

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
				ToMicroservice: msNamePrefix + msName,
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
			Reference: msNamePrefix + msName,
		},
	}

	ctx.startMicroservice(msName)
	req := client.LocalClient.ChangeRequest()
	err := req.Update(
		vppTap,
		linuxTap,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Eventually(ctx.getValueStateClb(vppTap), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED))
	Expect(ctx.getValueState(linuxTap)).To(Equal(kvs.ValueState_CONFIGURED))
	Expect(ctx.pingFromVPP(linuxTapIP)).To(BeNil())
	Expect(ctx.pingFromMs(msName, vppTapIP)).To(BeNil())

	// restart microservice twice
	for i := 0; i < 2; i++ {
		ctx.stopMicroservice(msName)
		Eventually(ctx.getValueStateClb(vppTap), msUpdateTimeout).Should(Equal(kvs.ValueState_PENDING))
		Eventually(ctx.getValueStateClb(linuxTap), msUpdateTimeout).Should(Equal(kvs.ValueState_PENDING))
		Expect(ctx.pingFromVPP(linuxTapIP)).ToNot(BeNil())

		//Expect(ctx.agentInSync()).To(BeTrue())

		ctx.startMicroservice(msName)
		Eventually(ctx.getValueStateClb(vppTap), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED))
		Expect(ctx.getValueState(linuxTap)).To(Equal(kvs.ValueState_CONFIGURED))
		Expect(ctx.pingFromVPP(linuxTapIP)).To(BeNil())
		Expect(ctx.pingFromMs(msName, vppTapIP)).To(BeNil())

		//Expect(ctx.agentInSync()).To(BeTrue())
	}

	// re-create VPP TAP
	req = client.LocalClient.ChangeRequest()
	err = req.Delete(
		vppTap,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Expect(ctx.pingFromVPP(linuxTapIP)).ToNot(BeNil())
	Expect(ctx.pingFromMs(msName, vppTapIP)).ToNot(BeNil())

	req = client.LocalClient.ChangeRequest()
	err = req.Update(
		vppTap,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Eventually(ctx.pingFromVPPClb(linuxTapIP), recreateTimeout).Should(BeNil())
	Expect(ctx.pingFromMs(msName, vppTapIP)).To(BeNil())
	//Expect(ctx.agentInSync()).To(BeTrue())

	// re-create Linux TAP
	req = client.LocalClient.ChangeRequest()
	err = req.Delete(
		linuxTap,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Expect(ctx.pingFromVPP(linuxTapIP)).ToNot(BeNil())
	Expect(ctx.pingFromMs(msName, vppTapIP)).ToNot(BeNil())

	req = client.LocalClient.ChangeRequest()
	err = req.Update(
		linuxTap,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Eventually(ctx.pingFromVPPClb(linuxTapIP), recreateTimeout).Should(BeNil())
	Expect(ctx.pingFromMs(msName, vppTapIP)).To(BeNil())
	//Expect(ctx.agentInSync()).To(BeTrue())
}

// connect VPP with a microservice via AF-PACKET + VETH interfaces
func TestAfPacketInterfaceConn(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		afPacketName     = "vpp-afpacket"
		veth1Name        = "linux-veth1"
		veth2Name        = "linux-veth2"
		veth1Hostname    = "veth1"
		veth2Hostname    = "veth2"
		afPacketIP       = "192.168.1.1"
		veth2IP          = "192.168.1.2"
		netMask          = "/30"
		msName           = "microservice1"
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
			Reference: msNamePrefix + msName,
		},
	}

	ctx.startMicroservice(msName)
	req := client.LocalClient.ChangeRequest()
	err := req.Update(
		afPacket,
		veth1,
		veth2,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Eventually(ctx.getValueStateClb(afPacket), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED))
	Expect(ctx.getValueState(veth1)).To(Equal(kvs.ValueState_CONFIGURED))
	Expect(ctx.getValueState(veth2)).To(Equal(kvs.ValueState_CONFIGURED))
	Expect(ctx.pingFromVPP(veth2IP)).To(BeNil())
	Expect(ctx.pingFromMs(msName, afPacketIP)).To(BeNil())

	// restart microservice twice
	for i := 0; i < 2; i++ {
		ctx.stopMicroservice(msName)
		Eventually(ctx.getValueStateClb(afPacket), msUpdateTimeout).Should(Equal(kvs.ValueState_PENDING))
		Eventually(ctx.getValueStateClb(veth1), msUpdateTimeout).Should(Equal(kvs.ValueState_PENDING))
		Eventually(ctx.getValueStateClb(veth2), msUpdateTimeout).Should(Equal(kvs.ValueState_PENDING))
		Expect(ctx.pingFromVPP(veth2IP)).ToNot(BeNil())

		//Expect(ctx.agentInSync()).To(BeTrue())

		ctx.startMicroservice(msName)
		Eventually(ctx.getValueStateClb(afPacket), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED))
		Expect(ctx.getValueState(veth1)).To(Equal(kvs.ValueState_CONFIGURED))
		Expect(ctx.getValueState(veth2)).To(Equal(kvs.ValueState_CONFIGURED))
		Expect(ctx.pingFromVPP(veth2IP)).To(BeNil())
		Expect(ctx.pingFromMs(msName, afPacketIP)).To(BeNil())

		//Expect(ctx.agentInSync()).To(BeTrue())
	}

	// re-create AF-PACKET
	req = client.LocalClient.ChangeRequest()
	err = req.Delete(
		afPacket,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Expect(ctx.pingFromVPP(veth2IP)).ToNot(BeNil())
	Expect(ctx.pingFromMs(msName, afPacketIP)).ToNot(BeNil())

	req = client.LocalClient.ChangeRequest()
	err = req.Update(
		afPacket,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Eventually(ctx.pingFromVPPClb(veth2IP), recreateTimeout).Should(BeNil())
	Expect(ctx.pingFromMs(msName, afPacketIP)).To(BeNil())
	//Expect(ctx.agentInSync()).To(BeTrue())

	// re-create VETH
	req = client.LocalClient.ChangeRequest()
	err = req.Delete(
		veth2,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Expect(ctx.pingFromVPP(veth2IP)).ToNot(BeNil())
	Expect(ctx.pingFromMs(msName, afPacketIP)).ToNot(BeNil())

	req = client.LocalClient.ChangeRequest()
	err = req.Update(
		veth2,
	).Send(context.Background())
	Expect(err).To(BeNil())

	Eventually(ctx.pingFromVPPClb(veth2IP), recreateTimeout).Should(BeNil())
	Expect(ctx.pingFromMs(msName, afPacketIP)).To(BeNil())
	//Expect(ctx.agentInSync()).To(BeTrue())
}

