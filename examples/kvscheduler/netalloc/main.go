//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ligato/cn-infra/agent"
	"go.ligato.io/vpp-agent/v3/client"
	linux_ifplugin "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	linux_l3plugin "go.ligato.io/vpp-agent/v3/plugins/linux/l3plugin"
	linux_nsplugin "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	vpp_ifplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	linux_ns "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

/*
	This example demonstrates netalloc plugin (topology disassociated from the addressing).
*/

func main() {
	// Set inter-dependency between VPP & Linux plugins
	vpp_ifplugin.DefaultPlugin.LinuxIfPlugin = &linux_ifplugin.DefaultPlugin
	vpp_ifplugin.DefaultPlugin.NsPlugin = &linux_nsplugin.DefaultPlugin
	linux_ifplugin.DefaultPlugin.VppIfPlugin = &vpp_ifplugin.DefaultPlugin

	ep := &ExamplePlugin{
		Orchestrator:  &orchestrator.DefaultPlugin,
		LinuxIfPlugin: &linux_ifplugin.DefaultPlugin,
		LinuxL3Plugin: &linux_l3plugin.DefaultPlugin,
		VPPIfPlugin:   &vpp_ifplugin.DefaultPlugin,
	}

	a := agent.NewAgent(
		agent.AllPlugins(ep),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExamplePlugin is the main plugin which
// handles resync and changes in this example.
type ExamplePlugin struct {
	LinuxIfPlugin *linux_ifplugin.IfPlugin
	LinuxL3Plugin *linux_l3plugin.L3Plugin
	VPPIfPlugin   *vpp_ifplugin.IfPlugin
	Orchestrator  *orchestrator.Plugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "netalloc-example"
}

// Init handles initialization phase.
func (p *ExamplePlugin) Init() error {
	return nil
}

// AfterInit handles phase after initialization.
func (p *ExamplePlugin) AfterInit() error {
	go demonstrateNetalloc()
	return nil
}

// Close cleans up the resources.
func (p *ExamplePlugin) Close() error {
	return nil
}

func demonstrateNetalloc() {
	// initial resync
	time.Sleep(time.Second)
	fmt.Println("=== RESYNC ===")

	err := client.LocalClient.ResyncConfig(
		// addresses
		veth1Addr, afpacketAddr, linuxTapAddr, vppTapAddr,
		// topology
		veth2, veth1, linuxTap, arpForVeth1, arpForLinuxTap,
		linkRouteToMs1, routeToMs1, linkRouteToMs2, routeToMs2,
		afpacket, vppTap)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("=== CHANGE ===")
	time.Sleep(time.Second * 5)
	err = client.LocalClient.ChangeRequest().
		Delete(veth1Addr).
		Send(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("=== CHANGE (revert of the previous change) ===")
	time.Sleep(time.Second * 5)
	err = client.LocalClient.ChangeRequest().
		Update(veth1Addr).
		Send(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
}

const (
	networkName = "example-net"

	veth1LogicalName = "myVETH1"
	veth1IPAddr      = "10.11.1.1"
	veth1IPAddr2     = "10.11.1.10"

	veth2LogicalName = "myVETH2"
	veth2HostName    = "veth2"

	afPacketLogicalName = "myAFPacket"
	afPacketIPAddr      = "10.11.1.2"
	afPacketHWAddr      = "a7:35:45:55:65:75"

	vppTapLogicalName = "myVPPTap"
	vppTapIPAddr      = "10.11.2.2"
	vppTapHwAddr      = "b3:12:12:45:A7:B7"

	linuxTapLogicalName = "myLinuxTAP"

	linuxTapIPAddr = "10.11.2.1"
	linuxTapHwAddr = "88:88:88:88:88:88"

	microserviceNetMask = "/30"
	mycroservice1       = "microservice1"
	mycroservice2       = "microservice2"

	mycroservice2Mtu = 1700

	routeMetric = 50
)

// ADRESSING

var (
	veth1Addr = &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: veth1LogicalName,
		Address:       veth1IPAddr + microserviceNetMask,
		Gw:            vppTapIPAddr,
	}

	veth1Addr2 = &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: veth1LogicalName,
		Address:       veth1IPAddr2 + microserviceNetMask,
		Gw:            vppTapIPAddr,
	}

	afpacketAddr = &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: afPacketLogicalName,
		Address:       afPacketIPAddr + microserviceNetMask,
	}

	linuxTapAddr = &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: linuxTapLogicalName,
		Address:       linuxTapIPAddr + microserviceNetMask,
		Gw:            afPacketIPAddr,
	}

	vppTapAddr = &netalloc.IPAllocation{
		NetworkName:   networkName,
		InterfaceName: vppTapLogicalName,
		Address:       vppTapIPAddr + microserviceNetMask,
	}
)

// TOPOLOGY

var (
	/* microservice1 <-> VPP */
	veth1 = &linux_interfaces.Interface{
		Name:        veth1LogicalName,
		Type:        linux_interfaces.Interface_VETH,
		Enabled:     true,
		PhysAddress: "66:66:66:66:66:66",
		IpAddresses: []string{
			"alloc:" + networkName,
		},
		Mtu:        1800,
		HostIfName: "veth1",
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{PeerIfName: veth2LogicalName},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_MICROSERVICE,
			Reference: mycroservice1,
		},
	}

	arpForVeth1 = &linux_l3.ARPEntry{
		Interface: veth1LogicalName,
		IpAddress: "alloc:" + networkName + "/" + vppTapLogicalName,
		HwAddress: vppTapHwAddr,
	}

	linkRouteToMs2 = &linux_l3.Route{
		OutgoingInterface: veth1LogicalName,
		Scope:             linux_l3.Route_LINK,
		DstNetwork:        "alloc:" + networkName + "/" + veth1LogicalName + "/GW",
	}

	routeToMs2 = &linux_l3.Route{
		OutgoingInterface: veth1LogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        "alloc:" + networkName + "/" + linuxTapLogicalName,
		GwAddr:            "alloc:" + networkName + "/GW",
		Metric:            routeMetric,
	}

	veth2 = &linux_interfaces.Interface{
		Name:       veth2LogicalName,
		Type:       linux_interfaces.Interface_VETH,
		Enabled:    true,
		Mtu:        1800,
		HostIfName: veth2HostName,
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{PeerIfName: veth1LogicalName},
		},
	}

	afpacket = &vpp_interfaces.Interface{
		Name:        afPacketLogicalName,
		Type:        vpp_interfaces.Interface_AF_PACKET,
		Enabled:     true,
		PhysAddress: afPacketHWAddr,
		IpAddresses: []string{
			"alloc:" + networkName,
		},
		Mtu: 1800,
		Link: &vpp_interfaces.Interface_Afpacket{
			Afpacket: &vpp_interfaces.AfpacketLink{
				HostIfName: veth2HostName,
			},
		},
	}

	/* microservice2 <-> VPP */

	linuxTap = &linux_interfaces.Interface{
		Name:        linuxTapLogicalName,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: linuxTapHwAddr,
		IpAddresses: []string{
			"alloc:" + networkName,
		},
		Mtu:        mycroservice2Mtu,
		HostIfName: "tap_to_vpp",
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapLogicalName,
			},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_MICROSERVICE,
			Reference: mycroservice2,
		},
	}

	vppTap = &vpp_interfaces.Interface{
		Name:        vppTapLogicalName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		PhysAddress: vppTapHwAddr,
		IpAddresses: []string{
			"alloc:" + networkName,
		},
		Mtu: mycroservice2Mtu,
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: mycroservice2,
			},
		},
	}

	arpForLinuxTap = &linux_l3.ARPEntry{
		Interface: linuxTapLogicalName,
		IpAddress: "alloc:" + networkName + "/" + afPacketLogicalName,
		HwAddress: afPacketHWAddr,
	}

	linkRouteToMs1 = &linux_l3.Route{
		OutgoingInterface: linuxTapLogicalName,
		Scope:             linux_l3.Route_LINK,
		DstNetwork:        "alloc:" + networkName + "/" + linuxTapLogicalName + "/GW",
	}

	routeToMs1 = &linux_l3.Route{
		OutgoingInterface: linuxTapLogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        "alloc:" + networkName + "/" + veth1LogicalName,
		GwAddr:            "alloc:" + networkName + "/GW",
		Metric:            routeMetric,
	}
)
