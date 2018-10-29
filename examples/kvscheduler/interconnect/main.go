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
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/db/keyval/etcd"

	"github.com/ligato/vpp-agent/clientv2/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/kvscheduler"
	linux_ifplugin "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	linux_l3plugin "github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin"
	linux_interfaces "github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
	linux_l3 "github.com/ligato/vpp-agent/plugins/linuxv2/model/l3"
	linux_ns "github.com/ligato/vpp-agent/plugins/linuxv2/model/namespace"
	vpp_ifplugin "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	vpp_interfaces "github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
)

/*
	This example demonstrates KVScheduler-based VPP ifplugin, Linux ifplugin and Linux l3plugin.
*/

func main() {
	etcdDataSync := kvdbsync.NewPlugin(kvdbsync.UseDeps(func(deps *kvdbsync.Deps) {
		deps.KvPlugin = &etcd.DefaultPlugin
	}))

	watchers := datasync.KVProtoWatchers{
		local.Get(),
		etcdDataSync,
	}

	// Set watcher for KVScheduler.
	kvscheduler.DefaultPlugin.Watcher = watchers

	vppIfPlugin := vpp_ifplugin.NewPlugin()

	linuxIfPlugin := linux_ifplugin.NewPlugin(
		linux_ifplugin.UseDeps(func(deps *linux_ifplugin.Deps) {
			deps.VppIfPlugin = vppIfPlugin
		}),
	)

	vppIfPlugin.LinuxIfPlugin = linuxIfPlugin

	linuxL3Plugin := linux_l3plugin.NewPlugin(
		linux_l3plugin.UseDeps(func(deps *linux_l3plugin.Deps) {
			deps.IfPlugin = linuxIfPlugin
		}),
	)

	ep := &ExamplePlugin{
		Scheduler:     &kvscheduler.DefaultPlugin,
		LinuxIfPlugin: linuxIfPlugin,
		LinuxL3Plugin: linuxL3Plugin,
		VPPIfPlugin:   vppIfPlugin,
		Datasync:      etcdDataSync,
	}

	a := agent.NewAgent(
		agent.AllPlugins(ep),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// PluginName is a constant with name of main plugin.
const PluginName = "example"

// ExamplePlugin is the main plugin which
// handles resync and changes in this example.
type ExamplePlugin struct {
	Scheduler     *kvscheduler.Scheduler
	LinuxIfPlugin *linux_ifplugin.IfPlugin
	LinuxL3Plugin *linux_l3plugin.L3Plugin
	VPPIfPlugin   *vpp_ifplugin.IfPlugin
	Datasync      *kvdbsync.Plugin
}

// String returns plugin name
func (plugin *ExamplePlugin) String() string {
	return PluginName
}

// Init handles initialization phase.
func (plugin *ExamplePlugin) Init() error {
	return nil
}

// AfterInit handles phase after initialization.
func (plugin *ExamplePlugin) AfterInit() error {
	go plugin.testLocalClientWithScheduler()
	return nil
}

// Close cleans up the resources.
func (plugin *ExamplePlugin) Close() error {
	return nil
}
func (plugin *ExamplePlugin) testLocalClientWithScheduler() {
	const (
		veth1LogicalName = "myVETH1"
		veth1HostName    = "veth1"
		veth1IPAddr      = "10.11.1.1"
		veth1HwAddr      = "66:66:66:66:66:66"

		veth2LogicalName = "myVETH2"
		veth2HostName    = "veth2"

		afPacketLogicalName = "myAFPacket"
		afPacketHwAddr      = "a7:35:45:55:65:75"
		afPacketIPAddr      = "10.11.1.2"

		vppTapLogicalName = "myVPPTap"
		vppTapIPAddr      = "10.11.2.2"
		vppTapHwAddr      = "b3:12:12:45:A7:B7"
		vppTapVersion     = 2

		linuxTapLogicalName = "myLinuxTAP"
		linuxTapHostName    = "tap_to_vpp"
		linuxTapIPAddr      = "10.11.2.1"
		linuxTapHwAddr      = "88:88:88:88:88:88"

		microserviceNetMask = "/30"
		mycroservice1       = "microservice1"
		mycroservice2       = "microservice2"
		microservice1Net    = "10.11.1.0" + microserviceNetMask
		microservice2Net    = "10.11.2.0" + microserviceNetMask
		mycroservice1Mtu    = 1800
		mycroservice2Mtu    = 1700

		routeMetric = 50
	)

	/* microservice1 <-> VPP */
	veth1 := &linux_interfaces.LinuxInterface{
		Name:        veth1LogicalName,
		Type:        linux_interfaces.LinuxInterface_VETH,
		Enabled:     true,
		PhysAddress: veth1HwAddr,
		IpAddresses: []string{
			veth1IPAddr + microserviceNetMask,
		},
		Mtu:        mycroservice1Mtu,
		HostIfName: veth1HostName,
		Link: &linux_interfaces.LinuxInterface_Veth{
			Veth: &linux_interfaces.LinuxInterface_VethLink{PeerIfName: veth2LogicalName},
		},
		Namespace: &linux_ns.LinuxNetNamespace{
			Type:      linux_ns.LinuxNetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroservice1,
		},
	}

	arpForVeth1 := &linux_l3.LinuxStaticARPEntry{
		Interface: veth1LogicalName,
		IpAddress: vppTapIPAddr,
		HwAddress: vppTapHwAddr,
	}

	linkRouteToMs2 := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: veth1LogicalName,
		Scope:             linux_l3.LinuxStaticRoute_LINK,
		DstNetwork:        vppTapIPAddr + "/32",
	}

	routeToMs2 := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: veth1LogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        microservice2Net,
		GwAddr:            vppTapIPAddr,
		Metric:            routeMetric,
	}

	veth2 := &linux_interfaces.LinuxInterface{
		Name:       veth2LogicalName,
		Type:       linux_interfaces.LinuxInterface_VETH,
		Enabled:    true,
		Mtu:        mycroservice1Mtu,
		HostIfName: veth2HostName,
		Link: &linux_interfaces.LinuxInterface_Veth{
			Veth: &linux_interfaces.LinuxInterface_VethLink{PeerIfName: veth1LogicalName},
		},
	}

	afpacket := &vpp_interfaces.Interface{
		Name:        afPacketLogicalName,
		Type:        vpp_interfaces.Interface_AF_PACKET_INTERFACE,
		Enabled:     true,
		PhysAddress: afPacketHwAddr,
		IpAddresses: []string{
			afPacketIPAddr + microserviceNetMask,
		},
		Mtu: mycroservice1Mtu,
		Link: &vpp_interfaces.Interface_Afpacket{
			Afpacket: &vpp_interfaces.Interface_AfpacketLink{
				HostIfName: veth2HostName,
			},
		},
	}

	/* microservice2 <-> VPP */

	linuxTap := &linux_interfaces.LinuxInterface{
		Name:        linuxTapLogicalName,
		Type:        linux_interfaces.LinuxInterface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: linuxTapHwAddr,
		IpAddresses: []string{
			linuxTapIPAddr + microserviceNetMask,
		},
		Mtu:        mycroservice2Mtu,
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.LinuxInterface_Tap{
			Tap: &linux_interfaces.LinuxInterface_TapLink{
				VppTapIfName: vppTapLogicalName,
			},
		},
		Namespace: &linux_ns.LinuxNetNamespace{
			Type:      linux_ns.LinuxNetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroservice2,
		},
	}

	vppTap := &vpp_interfaces.Interface{
		Name:        vppTapLogicalName,
		Type:        vpp_interfaces.Interface_TAP_INTERFACE,
		Enabled:     true,
		PhysAddress: vppTapHwAddr,
		IpAddresses: []string{
			vppTapIPAddr + microserviceNetMask,
		},
		Mtu: mycroservice2Mtu,
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.Interface_TapLink{
				Version:        vppTapVersion,
				ToMicroservice: mycroservice2,
			},
		},
	}

	arpForLinuxTap := &linux_l3.LinuxStaticARPEntry{
		Interface: linuxTapLogicalName,
		IpAddress: afPacketIPAddr,
		HwAddress: afPacketHwAddr,
	}

	linkRouteToMs1 := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapLogicalName,
		Scope:             linux_l3.LinuxStaticRoute_LINK,
		DstNetwork:        afPacketIPAddr + "/32",
	}

	routeToMs1 := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapLogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        microservice1Net,
		GwAddr:            afPacketIPAddr,
		Metric:            routeMetric,
	}

	// resync

	time.Sleep(time.Second * 2)
	fmt.Println("=== RESYNC 0 ===")

	txn := localclient.DataResyncRequest("example")
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
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	/*
		// data change

		time.Sleep(time.Second * 2)
		fmt.Println("=== CHANGE 1 ===")

		veth1.Enabled = false
		txn2 := localclient.DataChangeRequest("example")
		err = txn2.Put().
			LinuxInterface(veth1).
			Send().ReceiveReply()
		if err != nil {
			fmt.Println(err)
			return
		}
	*/

	// test Linux interface metadata map
	linuxIfIndex := plugin.LinuxIfPlugin.GetInterfaceIndex()
	linuxIfMeta, exists := linuxIfIndex.LookupByName(veth1LogicalName)
	fmt.Printf("Linux interface %s: found=%t, meta=%v\n", veth1LogicalName, exists, linuxIfMeta)
	linuxIfMeta, exists = linuxIfIndex.LookupByName(linuxTapLogicalName)
	fmt.Printf("Linux interface %s: found=%t, meta=%v\n", linuxTapLogicalName, exists, linuxIfMeta)

	// test VPP interface metadata map
	vppIfIndex := plugin.VPPIfPlugin.GetInterfaceIndex()
	vppIfMeta, exists := vppIfIndex.LookupByName(afPacketLogicalName)
	fmt.Printf("VPP interface %s: found=%t, meta=%v\n", afPacketLogicalName, exists, vppIfMeta)
	vppIfMeta, exists = vppIfIndex.LookupByName(vppTapLogicalName)
	fmt.Printf("VPP interface %s: found=%t, meta=%v\n", vppTapLogicalName, exists, vppIfMeta)
}
