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
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// test IP address allocation using the netalloc plugin for VPP+Linux interfaces,
// Linux routes and Linux ARPs in a topology where the interface is a neighbour
// of the associated GW.
//
// topology + addressing:
//  VPP loop (192.168.10.1/24) <--> VPP tap (192.168.11.1/24) <--> Linux tap (192.168.11.2/24)
// topology + addressing AFTER CHANGE:
//  VPP loop (192.168.20.1/24) <--> VPP tap (192.168.12.1/24) <--> Linux tap (192.168.12.2/24)
func TestIPWithNeighGW(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		networkName      = "net1"
		vppLoopName      = "vpp-loop"
		vppTapName       = "vpp-tap"
		linuxTapName     = "linux-tap"
		linuxTapHostname = "tap"
		vppLoopIP        = "192.168.10.1"
		vppLoopIP2       = "192.168.20.1"
		vppTapIP         = "192.168.11.1"
		vppTapIP2        = "192.168.12.1"
		vppTapHw         = "aa:aa:aa:bb:bb:bb"
		linuxTapIP       = "192.168.11.2"
		linuxTapIP2      = "192.168.12.2"
		linuxTapHw       = "cc:cc:cc:dd:dd:dd"
		netMask          = "/24"
		msName           = "microservice1"
	)

	// ------- addresses:

	vppLoopAddr := &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: vppLoopName,
		Address:       vppLoopIP + netMask,
	}

	vppTapAddr := &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: vppTapName,
		Address:       vppTapIP + netMask,
		Gw:            linuxTapIP,
	}

	linuxTapAddr := &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: linuxTapName,
		Address:       linuxTapIP + netMask,
		Gw:            vppTapIP,
	}

	// ------- network items:

	vppLoop := &vpp_interfaces.Interface{
		Name:        vppLoopName,
		Type:        vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		IpAddresses: []string{"alloc:" + networkName},
	}

	vppTap := &vpp_interfaces.Interface{
		Name:        vppTapName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{"alloc:" + networkName},
		PhysAddress: vppTapHw,
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
		IpAddresses: []string{"alloc:" + networkName},
		HostIfName:  linuxTapHostname,
		PhysAddress: linuxTapHw,
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

	linuxArp := &linux_l3.ARPEntry{
		Interface: linuxTapName,
		IpAddress: "alloc:" + networkName + "/" + vppTapName,
		HwAddress: vppTapHw,
	}

	linuxRoute := &linux_l3.Route{
		OutgoingInterface: linuxTapName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        "alloc:" + networkName + "/" + vppLoopName,
		GwAddr:            "alloc:" + networkName + "/GW",
	}

	ctx.startMicroservice(msName)
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppLoopAddr, vppTapAddr, linuxTapAddr,
		vppLoop, vppTap, linuxTap,
		linuxArp, linuxRoute,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	checkItemsAreConfigured := func(msRestart, withLoopAddr bool) {
		// configured immediately:
		if withLoopAddr {
			Expect(ctx.getValueState(vppLoopAddr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		}
		Expect(ctx.getValueState(vppTapAddr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(linuxTapAddr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(vppLoop)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		// the rest depends on the microservice
		if msRestart {
			Eventually(ctx.getValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
		} else {
			Expect(ctx.getValueState(vppTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		}
		Expect(ctx.getValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(linuxArp)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		if withLoopAddr {
			Expect(ctx.getValueState(linuxRoute)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		} else {
			Expect(ctx.getValueState(linuxRoute)).To(Equal(kvscheduler.ValueState_PENDING))
		}
	}
	checkItemsAreConfigured(true, true)

	// check connection with ping
	Expect(ctx.pingFromVPP(linuxTapIP)).To(Succeed())
	Expect(ctx.pingFromMs(msName, vppLoopIP)).To(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())

	// restart microservice
	ctx.stopMicroservice(msName)
	ctx.startMicroservice(msName)
	checkItemsAreConfigured(true, true)

	// check connection with ping (few packets will get lost before tables are refreshed)
	ctx.pingFromVPP(linuxTapIP)
	Expect(ctx.pingFromVPP(linuxTapIP)).To(Succeed())
	Expect(ctx.pingFromMs(msName, vppLoopIP)).To(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())

	// change IP addresses - the network items should be re-created
	vppLoopAddr.Address = vppLoopIP2 + netMask
	vppTapAddr.Address = vppTapIP2 + netMask
	vppTapAddr.Gw = linuxTapIP2
	linuxTapAddr.Address = linuxTapIP2 + netMask
	linuxTapAddr.Gw = vppTapIP2

	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(vppLoopAddr, vppTapAddr, linuxTapAddr).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())
	checkItemsAreConfigured(false, true)

	// check connection with ping
	Expect(ctx.pingFromVPP(linuxTapIP)).NotTo(Succeed())

	Expect(ctx.pingFromMs(msName, vppLoopIP)).NotTo(Succeed())
	Expect(ctx.pingFromVPP(linuxTapIP2)).To(Succeed())
	Expect(ctx.pingFromMs(msName, vppLoopIP2)).To(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())

	// de-allocate loopback IP - the connection should not work anymore
	req = ctx.grpcClient.ChangeRequest()
	err = req.Delete(vppLoopAddr).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	// loopback is still created but without IP and route is pending
	checkItemsAreConfigured(false, false)

	// can ping linux TAP from VPP, but cannot ping loopback
	Expect(ctx.pingFromVPP(linuxTapIP2)).To(Succeed())
	Expect(ctx.pingFromMs(msName, vppLoopIP2)).NotTo(Succeed())

	// TODO: not in-sync - the list of IP addresses is updated in the metadata
	//  - we need to figure out how to get rid of this and how to solve VRF-related
	//    dependencies with netalloc'd IP addresses
	//Expect(ctx.agentInSync()).To(BeTrue())
}

// test IP address allocation using the netalloc plugin for VPP+Linux interfaces,
// Linux routes and Linux ARPs in a topology where the interface is NOT a neighbour
// of the associated GW.
//
// topology + addressing (note the single-host network mask for Linux TAP):
//  VPP loop (192.168.10.1/24) <--> VPP tap (192.168.11.1/24) <--> Linux tap (192.168.11.2/32)
// topology + addressing AFTER CHANGE:
//  VPP loop (192.168.20.1/24) <--> VPP tap (192.168.12.1/24) <--> Linux tap (192.168.12.2/32)
func TestIPWithNonLocalGW(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		networkName      = "net1"
		vppLoopName      = "vpp-loop"
		vppTapName       = "vpp-tap"
		linuxTapName     = "linux-tap"
		linuxTapHostname = "tap"
		vppLoopIP        = "192.168.10.1"
		vppLoopIP2       = "192.168.20.1"
		vppTapIP         = "192.168.11.1"
		vppTapIP2        = "192.168.12.1"
		vppTapHw         = "aa:aa:aa:bb:bb:bb"
		linuxTapIP       = "192.168.11.2"
		linuxTapIP2      = "192.168.12.2"
		linuxTapHw       = "cc:cc:cc:dd:dd:dd"
		vppNetMask       = "/24"
		linuxNetMask     = "/32"
		msName           = "microservice1"
	)

	// ------- addresses:

	vppLoopAddr := &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: vppLoopName,
		Address:       vppLoopIP + vppNetMask,
	}

	vppTapAddr := &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: vppTapName,
		Address:       vppTapIP + vppNetMask,
		Gw:            linuxTapIP,
	}

	linuxTapAddr := &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: linuxTapName,
		Address:       linuxTapIP + linuxNetMask,
		Gw:            vppTapIP,
	}

	// ------- network items:

	vppLoop := &vpp_interfaces.Interface{
		Name:        vppLoopName,
		Type:        vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		IpAddresses: []string{"alloc:" + networkName},
	}

	vppTap := &vpp_interfaces.Interface{
		Name:        vppTapName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{"alloc:" + networkName},
		PhysAddress: vppTapHw,
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
		IpAddresses: []string{"alloc:" + networkName},
		HostIfName:  linuxTapHostname,
		PhysAddress: linuxTapHw,
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

	linuxArp := &linux_l3.ARPEntry{
		Interface: linuxTapName,
		IpAddress: "alloc:" + networkName + "/" + vppTapName,
		HwAddress: vppTapHw,
	}

	linuxRoute := &linux_l3.Route{
		OutgoingInterface: linuxTapName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        "alloc:" + networkName + "/" + vppLoopName,
		GwAddr:            "alloc:" + networkName + "/GW",
	}

	// link route is necessary to route the GW of the linux TAP interface
	linuxLinkRoute := &linux_l3.Route{
		OutgoingInterface: linuxTapName,
		Scope:             linux_l3.Route_LINK,
		DstNetwork:        "alloc:" + networkName + "/" + linuxTapName + "/GW",
	}

	ctx.startMicroservice(msName)
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppLoopAddr, vppTapAddr, linuxTapAddr,
		vppLoop, vppTap, linuxTap,
		linuxArp, linuxRoute, linuxLinkRoute,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	checkItemsAreConfigured := func(msRestart, withLinkRoute bool) {
		// configured immediately:
		Expect(ctx.getValueState(vppLoopAddr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(vppTapAddr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(linuxTapAddr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(vppLoop)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		// the rest depends on the microservice
		if msRestart {
			Eventually(ctx.getValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
		} else {
			Expect(ctx.getValueState(vppTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		}
		Expect(ctx.getValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(linuxArp)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		if withLinkRoute {
			Expect(ctx.getValueState(linuxRoute)).To(Equal(kvscheduler.ValueState_CONFIGURED))
			Expect(ctx.getValueState(linuxLinkRoute)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		} else {
			Expect(ctx.getValueState(linuxRoute)).To(Equal(kvscheduler.ValueState_PENDING))
		}
	}
	checkItemsAreConfigured(true, true)

	// check connection with ping
	Expect(ctx.pingFromVPP(linuxTapIP)).To(Succeed())
	Expect(ctx.pingFromMs(msName, vppLoopIP)).To(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())

	// restart microservice
	ctx.stopMicroservice(msName)
	ctx.startMicroservice(msName)
	checkItemsAreConfigured(true, true)

	// check connection with ping (few packets will get lost before tables are refreshed)
	ctx.pingFromVPP(linuxTapIP)
	Expect(ctx.pingFromVPP(linuxTapIP)).To(Succeed())
	Expect(ctx.pingFromMs(msName, vppLoopIP)).To(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())

	// change IP addresses - the network items should be re-created
	vppLoopAddr.Address = vppLoopIP2 + vppNetMask
	vppTapAddr.Address = vppTapIP2 + vppNetMask
	vppTapAddr.Gw = linuxTapIP2
	linuxTapAddr.Address = linuxTapIP2 + linuxNetMask
	linuxTapAddr.Gw = vppTapIP2

	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(vppLoopAddr, vppTapAddr, linuxTapAddr).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())
	checkItemsAreConfigured(false, true)

	// check connection with ping
	Expect(ctx.pingFromVPP(linuxTapIP)).NotTo(Succeed())

	Expect(ctx.pingFromMs(msName, vppLoopIP)).NotTo(Succeed())
	Expect(ctx.pingFromVPP(linuxTapIP2)).To(Succeed())
	Expect(ctx.pingFromMs(msName, vppLoopIP2)).To(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())

	// remove link route - this should make the GW for linux TAP non-routable
	req = ctx.grpcClient.ChangeRequest()
	err = req.Delete(linuxLinkRoute).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	// the route to VPP is pending
	checkItemsAreConfigured(false, false)

	// cannot ping anymore from any of the sides
	Expect(ctx.pingFromVPP(linuxTapIP2)).NotTo(Succeed())
	Expect(ctx.pingFromMs(msName, vppLoopIP2)).NotTo(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())
}

// test IP address allocation using the netalloc plugin for VPP routes mainly.
//
// topology + addressing:
//  VPP tap (192.168.11.1/24) <--> Linux tap (192.168.11.2/24) <--> Linux loop (192.168.20.1/24, 10.10.10.10/32)
//
// topology + addressing AFTER CHANGE:
//  VPP tap (192.168.12.1/24) <--> Linux tap (192.168.12.2/24) <--> Linux loop (192.168.30.1/24, 10.10.10.10/32)
func TestVPPRoutesWithNetalloc(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		network1Name     = "net1"
		network2Name     = "net2"
		vppTapName       = "vpp-tap"
		linuxTapName     = "linux-tap"
		linuxTapHostname = "tap"
		linuxLoopName    = "linux-loop"
		vppTapIP         = "192.168.11.1"
		vppTapIP2        = "192.168.12.1"
		linuxTapIP       = "192.168.11.2"
		linuxTapIP2      = "192.168.12.2"
		linuxLoopNet1IP  = "192.168.20.1"
		linuxLoopNet1IP2 = "192.168.30.1"
		linuxLoopNet2IP  = "10.10.10.10"
		net1Mask         = "/24"
		net2Mask         = "/32"
		msName           = "microservice1"
	)

	// ------- addresses:

	vppTapAddr := &netalloc.IPAllocation{
		NetworkName:   network1Name,
		InterfaceName: vppTapName,
		Address:       vppTapIP + net1Mask,
		Gw:            linuxTapIP,
	}

	linuxTapAddr := &netalloc.IPAllocation{
		NetworkName:   network1Name,
		InterfaceName: linuxTapName,
		Address:       linuxTapIP + net1Mask,
		Gw:            vppTapIP,
	}

	linuxLoopNet1Addr := &netalloc.IPAllocation{
		NetworkName:   network1Name,
		InterfaceName: linuxLoopName,
		Address:       linuxLoopNet1IP + net1Mask,
	}

	linuxLoopNet2Addr := &netalloc.IPAllocation{
		NetworkName:   network2Name,
		InterfaceName: linuxLoopName,
		Address:       linuxLoopNet2IP + net2Mask,
	}

	// ------- network items:

	vppTap := &vpp_interfaces.Interface{
		Name:        vppTapName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{"alloc:" + network1Name},
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
		IpAddresses: []string{"alloc:" + network1Name},
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

	linuxLoop := &linux_interfaces.Interface{
		Name:    linuxLoopName,
		Type:    linux_interfaces.Interface_LOOPBACK,
		Enabled: true,
		IpAddresses: []string{
			"127.0.0.1/8", "alloc:" + network1Name, "alloc:" + network2Name},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: msNamePrefix + msName,
		},
	}

	vppRouteLoopNet1 := &vpp_l3.Route{
		OutgoingInterface: vppTapName,
		DstNetwork:        "alloc:" + network1Name + "/" + linuxLoopName,
		NextHopAddr:       "alloc:" + network1Name + "/GW",
	}

	vppRouteLoopNet2 := &vpp_l3.Route{
		OutgoingInterface: vppTapName,
		DstNetwork:        "alloc:" + network2Name + "/" + linuxLoopName,
		NextHopAddr:       "alloc:" + network1Name + "/GW",
	}

	ctx.startMicroservice(msName)
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppTapAddr, linuxTapAddr, linuxLoopNet1Addr, linuxLoopNet2Addr,
		vppTap, linuxTap, linuxLoop,
		vppRouteLoopNet1, vppRouteLoopNet2,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	checkItemsAreConfigured := func(msRestart, withLoopNet2Addr bool) {
		// configured immediately:
		if withLoopNet2Addr {
			Expect(ctx.getValueState(linuxLoopNet2Addr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		}
		Expect(ctx.getValueState(vppTapAddr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(linuxTapAddr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(linuxLoopNet1Addr)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		// the rest depends on the microservice
		if msRestart {
			Eventually(ctx.getValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
		} else {
			Expect(ctx.getValueState(vppTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		}
		Expect(ctx.getValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(linuxLoop)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		Expect(ctx.getValueState(vppRouteLoopNet1)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		if withLoopNet2Addr {
			Expect(ctx.getValueState(vppRouteLoopNet2)).To(Equal(kvscheduler.ValueState_CONFIGURED))
		} else {
			Expect(ctx.getValueState(vppRouteLoopNet2)).To(Equal(kvscheduler.ValueState_PENDING))
		}
	}
	checkItemsAreConfigured(true, true)

	// check connection with ping
	Expect(ctx.pingFromVPP(linuxLoopNet1IP)).To(Succeed())
	Expect(ctx.pingFromVPP(linuxLoopNet2IP)).To(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())

	// restart microservice
	ctx.stopMicroservice(msName)
	ctx.startMicroservice(msName)
	checkItemsAreConfigured(true, true)

	// check connection with ping (few packets will get lost before tables are refreshed)
	ctx.pingFromVPP(linuxLoopNet1IP)
	ctx.pingFromVPP(linuxLoopNet2IP)
	Expect(ctx.pingFromVPP(linuxLoopNet1IP)).To(Succeed())
	Expect(ctx.pingFromVPP(linuxLoopNet2IP)).To(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())

	// change IP addresses - the network items should be re-created
	linuxLoopNet1Addr.Address = linuxLoopNet1IP2 + net1Mask
	vppTapAddr.Address = vppTapIP2 + net1Mask
	vppTapAddr.Gw = linuxTapIP2
	linuxTapAddr.Address = linuxTapIP2 + net1Mask
	linuxTapAddr.Gw = vppTapIP2

	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(linuxLoopNet1Addr, vppTapAddr, linuxTapAddr).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())
	checkItemsAreConfigured(false, true)

	// check connection with ping
	Expect(ctx.pingFromVPP(linuxLoopNet1IP)).NotTo(Succeed())
	Expect(ctx.pingFromVPP(linuxLoopNet1IP2)).To(Succeed())
	Expect(ctx.pingFromVPP(linuxLoopNet2IP)).To(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())

	// de-allocate loopback IP in net2 - the connection to that IP should not work anymore
	req = ctx.grpcClient.ChangeRequest()
	err = req.Delete(linuxLoopNet2Addr).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	// loopback is still created but without IP and route is pending
	checkItemsAreConfigured(false, false)

	// can ping loop1, but cannot ping loop2
	Expect(ctx.pingFromVPP(linuxLoopNet1IP2)).To(Succeed())
	Expect(ctx.pingFromVPP(linuxLoopNet2IP)).NotTo(Succeed())
	Expect(ctx.agentInSync()).To(BeTrue())
}
