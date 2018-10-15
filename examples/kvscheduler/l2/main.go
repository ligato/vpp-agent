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
	linux_ns "github.com/ligato/vpp-agent/plugins/linuxv2/model/namespace"
	vpp_ifplugin "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	vpp_l2plugin "github.com/ligato/vpp-agent/plugins/vppv2/l2plugin"
	vpp_interfaces "github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	vpp_l2 "github.com/ligato/vpp-agent/plugins/vppv2/model/l2"
)

/*
	This example demonstrates L2 Plugin v.2.
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

	vppIfPlugin := vpp_ifplugin.NewPlugin(
		vpp_ifplugin.UseDeps(func(deps *vpp_ifplugin.Deps) {
			deps.PublishStatistics = etcdDataSync
		}),
	)

	linuxIfPlugin := linux_ifplugin.NewPlugin(
		linux_ifplugin.UseDeps(func(deps *linux_ifplugin.Deps) {
			deps.VppIfPlugin = vppIfPlugin
		}),
	)

	vppIfPlugin.LinuxIfPlugin = linuxIfPlugin

	vppL2Plugin := vpp_l2plugin.NewPlugin(
		vpp_l2plugin.UseDeps(func(deps *vpp_l2plugin.Deps) {
			deps.IfPlugin = vppIfPlugin
		}),
	)

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
		VPPL2Plugin:   vppL2Plugin,
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
	VPPL2Plugin   *vpp_l2plugin.L2Plugin
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
		bdNetPrefix = "10.11.1."
		bdNetMask = "/24"

		veth1LogicalName = "myVETH1"
		veth1HostName    = "veth1"
		veth1IPAddr      = bdNetPrefix + "1"
		veth1HwAddr      = "66:66:66:66:66:66"

		veth2LogicalName = "myVETH2"
		veth2HostName    = "veth2"

		afPacketLogicalName = "myAFPacket"
		afPacketHwAddr      = "a7:35:45:55:65:75"

		vppTapLogicalName = "myVPPTap"
		vppTapHwAddr      = "b3:12:12:45:A7:B7"
		vppTapVersion     = 2

		linuxTapLogicalName = "myLinuxTAP"
		linuxTapHostName    = "tap_to_vpp"
		linuxTapIPAddr      = bdNetPrefix + "2"
		linuxTapHwAddr      = "88:88:88:88:88:88"

		mycroservice1       = "microservice1"
		mycroservice2       = "microservice2"

		bviLoopName = "myLoopback1"
		bviLoopIP = bdNetPrefix + "3"
		bviLoopHwAddr = "cd:cd:cd:cd:cd:cd"

		loop2Name = "myLoopback2"
		loop2HwAddr = "ef:ef:ef:ef:ef:ef"

		bdName = "myBridgeDomain"
		bdFlood = true
		bdUnknownUnicastFlood = true
		bdForward = true
		bdLearn = false /* Learning turned off, FIBs are needed for connectivity */
		bdArpTermination = true
		bdMacAge = 0
	)

	/* microservice1 <-> VPP */
	veth1 := &linux_interfaces.LinuxInterface{
		Name:        veth1LogicalName,
		Type:        linux_interfaces.LinuxInterface_VETH,
		Enabled:     true,
		PhysAddress: veth1HwAddr,
		IpAddresses: []string{
			veth1IPAddr + bdNetMask,
		},
		HostIfName: veth1HostName,
		Link: &linux_interfaces.LinuxInterface_Veth{
			Veth: &linux_interfaces.LinuxInterface_VethLink{PeerIfName: veth2LogicalName},
		},
		Namespace: &linux_ns.LinuxNetNamespace{
			Type:      linux_ns.LinuxNetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroservice1,
		},
	}

	veth2 := &linux_interfaces.LinuxInterface{
		Name:        veth2LogicalName,
		Type:        linux_interfaces.LinuxInterface_VETH,
		Enabled:     true,
		HostIfName:  veth2HostName,
		Link: &linux_interfaces.LinuxInterface_Veth{
			Veth: &linux_interfaces.LinuxInterface_VethLink{PeerIfName: veth1LogicalName},
		},
	}

	afpacket := &vpp_interfaces.Interface{
		Name:        afPacketLogicalName,
		Type:        vpp_interfaces.Interface_AF_PACKET_INTERFACE,
		Enabled:     true,
		PhysAddress: afPacketHwAddr,
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
			linuxTapIPAddr + bdNetMask,
		},
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
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.Interface_TapLink{
				Version:        vppTapVersion,
				ToMicroservice: mycroservice2,
			},
		},
	}

	/* Bridge domain */

	bd := &vpp_l2.BridgeDomain{
		Name:                bdName,
		Flood:               bdFlood,
		UnknownUnicastFlood: bdUnknownUnicastFlood,
		Forward:             bdForward,
		Learn:               bdLearn,
		ArpTermination:      bdArpTermination,
		MacAge:              bdMacAge,
		Interfaces:          []*vpp_l2.BridgeDomain_Interface{
			{
				Name: vppTapLogicalName,
			},
			{
				Name: afPacketLogicalName,
			},
			{
				Name: bviLoopName,
				BridgedVirtualInterface: true,
			},
		},
	}

	bviLoop := &vpp_interfaces.Interface{
		Name:        bviLoopName,
		Type:        vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		PhysAddress: bviLoopHwAddr,
		IpAddresses: []string{
			bviLoopIP + bdNetMask,
		},
	}

	loop2 := &vpp_interfaces.Interface{
		Name:        loop2Name,
		Type:        vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		PhysAddress: loop2HwAddr,
	}

	/* FIB entries */

	fibForLoop := &vpp_l2.FIBEntry{
		PhysAddress:             bviLoopHwAddr,
		BridgeDomain:            bdName,
		Action:                  vpp_l2.FIBEntry_FORWARD,
		OutgoingInterface:       bviLoopName,
		BridgedVirtualInterface: true,
		StaticConfig:            true,
	}

	fibForVETH := &vpp_l2.FIBEntry{
		PhysAddress:             veth1HwAddr,
		BridgeDomain:            bdName,
		Action:                  vpp_l2.FIBEntry_FORWARD,
		OutgoingInterface:       afPacketLogicalName,
	}

	fibForTAP := &vpp_l2.FIBEntry{
		PhysAddress:             linuxTapHwAddr,
		BridgeDomain:            bdName,
		Action:                  vpp_l2.FIBEntry_FORWARD,
		OutgoingInterface:       vppTapLogicalName,
	}

	dropFIB := &vpp_l2.FIBEntry{
		PhysAddress:             loop2HwAddr,
		BridgeDomain:            bdName,
		Action:                  vpp_l2.FIBEntry_DROP,
	}

	/* XConnect */

	xConnectMs1ToMs2 := &vpp_l2.XConnectPair{
		ReceiveInterface:  afPacketLogicalName,
		TransmitInterface: vppTapLogicalName,
	}

	xConnectMs2ToMs1 := &vpp_l2.XConnectPair{
		ReceiveInterface:  vppTapLogicalName,
		TransmitInterface: afPacketLogicalName,
	}

	// resync

	time.Sleep(time.Second * 2)
	fmt.Println("=== RESYNC 0 (using bridge domain) ===")

	txn := localclient.DataResyncRequest("example")
	err := txn.
		LinuxInterface(veth2).
		LinuxInterface(veth1).
		LinuxInterface(linuxTap).
		VppInterface(afpacket).
		VppInterface(vppTap).
		VppInterface(bviLoop).
		VppInterface(loop2).
		BD(bd).
		BDFIB(fibForLoop).
		BDFIB(fibForTAP).
		BDFIB(fibForVETH).
		BDFIB(dropFIB).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// data changes

	time.Sleep(time.Second * 60)
	fmt.Printf("=== CHANGE 1 (switching to XConnect) ===\n")

	txn3 := localclient.DataChangeRequest("example")
	err = txn3.Delete().
		BD(bd.Name). // FIBs will be pending
		Put().
		XConnect(xConnectMs1ToMs2).
		XConnect(xConnectMs2ToMs1).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}
}