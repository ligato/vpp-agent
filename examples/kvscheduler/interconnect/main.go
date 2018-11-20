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
	"fmt"
	"log"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/vpp-agent/clientv2"

	"github.com/ligato/vpp-agent/api/models/linux/l3"
	linux_ns "github.com/ligato/vpp-agent/api/models/linux/namespace"
	"github.com/ligato/vpp-agent/plugins/kvscheduler"
	linux_ifplugin "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	linuxifaceidx "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/ifaceidx"
	linux_l3plugin "github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin"
	vpp_ifplugin "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	vppifaceidx "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	vpp_interfaces "github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
)

/*
	This example demonstrates KVScheduler-based VPP ifplugin, Linux ifplugin and Linux l3plugin.
*/

func main() {
	// Set watcher for KVScheduler.
	kvscheduler.DefaultPlugin.Watcher = local.DefaultRegistry

	// Set inter-dependency between VPP & Linux plugins
	vpp_ifplugin.DefaultPlugin.LinuxIfPlugin = &linux_ifplugin.DefaultPlugin
	linux_ifplugin.DefaultPlugin.VppIfPlugin = &vpp_ifplugin.DefaultPlugin

	ep := &ExamplePlugin{
		Scheduler:     &kvscheduler.DefaultPlugin,
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
	Scheduler     *kvscheduler.Scheduler
	LinuxIfPlugin *linux_ifplugin.IfPlugin
	LinuxL3Plugin *linux_l3plugin.L3Plugin
	VPPIfPlugin   *vpp_ifplugin.IfPlugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "vpp-linux-example"
}

// Init handles initialization phase.
func (p *ExamplePlugin) Init() error {
	return nil
}

// AfterInit handles phase after initialization.
func (p *ExamplePlugin) AfterInit() error {
	go testLocalClientWithScheduler(
		p.VPPIfPlugin.GetInterfaceIndex(),
		p.LinuxIfPlugin.GetInterfaceIndex(),
	)
	return nil
}

// Close cleans up the resources.
func (p *ExamplePlugin) Close() error {
	return nil
}

func testLocalClientWithScheduler(
	vppIfIndex vppifaceidx.IfaceMetadataIndex,
	linuxIfIndex linuxifaceidx.LinuxIfMetadataIndex,
) {
	// initial resync
	time.Sleep(time.Second * 2)
	fmt.Println("=== RESYNC ===")

	/*txn := localclient.DataResyncRequest("example")
	err := txn.
		LinuxInterface(veth2).
		LinuxInterface(veth1).
		LinuxInterface(linuxTap).
		LinuxArpEntry(arpForVeth1).
		LinuxArpEntry(arpForLinuxTap).
		LinuxRoute(linkRouteToMs1).
		LinuxRoute(routeToMs1).
		LinuxRoute(linkRouteToMs2).
		LinuxRoute(routeToMs2).
		VppInterface(afpacket).
		VppInterface(vppTap).
		Send().ReceiveReply()*/

	txn := clientv2.LocalClient.Resync()

	txn.Linux().
		Interface(veth2).
		Interface(veth1).
		Interface(linuxTap).
		ArpEntry(arpForVeth1).
		ArpEntry(arpForLinuxTap).
		Route(linkRouteToMs1).
		Route(routeToMs1).
		Route(linkRouteToMs2).
		Route(routeToMs2)

	txn.VPP().
		Interface(afpacket).
		Interface(vppTap)

	if err := txn.Commit(); err != nil {
		fmt.Println(err)
		return
	}

	// data change
	time.Sleep(time.Second * 10)
	fmt.Println("=== CHANGE ===")

	veth1.Enabled = false

	/*txn2 := localclient.DataChangeRequest("example")
	err = txn2.Put().
		LinuxInterface(veth1).
		Send().ReceiveReply()*/

	txn2 := clientv2.LocalClient.Change()
	txn2.Linux().Put().
		Interface(veth1)
	txn2.VPP().Delete().
		Interface(vppTap.Name)
	if err := txn2.Commit(); err != nil {
		fmt.Println(err)
		return
	}

	// test Linux interface metadata map
	linuxIfMeta, exists := linuxIfIndex.LookupByName(veth1LogicalName)
	fmt.Printf("Linux interface %s: found=%t, meta=%v\n", veth1LogicalName, exists, linuxIfMeta)
	linuxIfMeta, exists = linuxIfIndex.LookupByName(linuxTapLogicalName)
	fmt.Printf("Linux interface %s: found=%t, meta=%v\n", linuxTapLogicalName, exists, linuxIfMeta)

	// test VPP interface metadata map
	vppIfMeta, exists := vppIfIndex.LookupByName(afPacketLogicalName)
	fmt.Printf("VPP interface %s: found=%t, meta=%v\n", afPacketLogicalName, exists, vppIfMeta)
	vppIfMeta, exists = vppIfIndex.LookupByName(vppTapLogicalName)
	fmt.Printf("VPP interface %s: found=%t, meta=%v\n", vppTapLogicalName, exists, vppIfMeta)
}

const (
	veth1LogicalName = "myVETH1"

	veth2LogicalName = "myVETH2"
	veth2HostName    = "veth2"

	afPacketLogicalName = "myAFPacket"

	afPacketIPAddr = "10.11.1.2"

	vppTapLogicalName = "myVPPTap"
	vppTapIPAddr      = "10.11.2.2"
	vppTapHwAddr      = "b3:12:12:45:A7:B7"

	linuxTapLogicalName = "myLinuxTAP"

	linuxTapIPAddr = "10.11.2.1"
	linuxTapHwAddr = "88:88:88:88:88:88"

	microserviceNetMask = "/30"
	mycroservice1       = "microservice1"
	mycroservice2       = "microservice2"
	microservice1Net    = "10.11.1.0" + microserviceNetMask
	microservice2Net    = "10.11.2.0" + microserviceNetMask

	mycroservice2Mtu = 1700

	routeMetric = 50
)

var (
	/* microservice1 <-> VPP */
	veth1 = &linux_interfaces.Interface{
		Name:        veth1LogicalName,
		Type:        linux_interfaces.Interface_VETH,
		Enabled:     true,
		PhysAddress: "66:66:66:66:66:66",
		IpAddresses: []string{
			("10.11.1.1") + microserviceNetMask,
		},
		Mtu:        1800,
		HostIfName: "veth1",
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{PeerIfName: veth2LogicalName},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroservice1,
		},
	}

	arpForVeth1 = &linux_l3.StaticARPEntry{
		Interface: veth1LogicalName,
		IpAddress: vppTapIPAddr,
		HwAddress: vppTapHwAddr,
	}

	linkRouteToMs2 = &linux_l3.StaticRoute{
		OutgoingInterface: veth1LogicalName,
		Scope:             linux_l3.StaticRoute_LINK,
		DstNetwork:        vppTapIPAddr + "/32",
	}

	routeToMs2 = &linux_l3.StaticRoute{
		OutgoingInterface: veth1LogicalName,
		Scope:             linux_l3.StaticRoute_GLOBAL,
		DstNetwork:        microservice2Net,
		GwAddr:            vppTapIPAddr,
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
		PhysAddress: "a7:35:45:55:65:75",
		IpAddresses: []string{
			afPacketIPAddr + microserviceNetMask,
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
			linuxTapIPAddr + microserviceNetMask,
		},
		Mtu:        mycroservice2Mtu,
		HostIfName: "tap_to_vpp",
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapLogicalName,
			},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroservice2,
		},
	}

	vppTap = &vpp_interfaces.Interface{
		Name:        vppTapLogicalName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		PhysAddress: vppTapHwAddr,
		IpAddresses: []string{
			vppTapIPAddr + microserviceNetMask,
		},
		Mtu: mycroservice2Mtu,
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: mycroservice2,
			},
		},
	}

	arpForLinuxTap = &linux_l3.StaticARPEntry{
		Interface: linuxTapLogicalName,
		IpAddress: afPacketIPAddr,
		HwAddress: "a7:35:45:55:65:75",
	}

	linkRouteToMs1 = &linux_l3.StaticRoute{
		OutgoingInterface: linuxTapLogicalName,
		Scope:             linux_l3.StaticRoute_LINK,
		DstNetwork:        afPacketIPAddr + "/32",
	}

	routeToMs1 = &linux_l3.StaticRoute{
		OutgoingInterface: linuxTapLogicalName,
		Scope:             linux_l3.StaticRoute_GLOBAL,
		DstNetwork:        microservice1Net,
		GwAddr:            afPacketIPAddr,
		Metric:            routeMetric,
	}
)
