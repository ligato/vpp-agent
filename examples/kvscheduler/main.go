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
	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	"github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin"
	"github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linuxv2/model/l3"
	"github.com/ligato/vpp-agent/plugins/linuxv2/model/namespace"
)

/*
	This example demonstrates example using KVScheduler.
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

	ep := &ExamplePlugin{
		Scheduler: &kvscheduler.DefaultPlugin,
		IfPlugin:  &ifplugin.DefaultPlugin,
		L3Plugin:  &l3plugin.DefaultPlugin,
		Datasync:  etcdDataSync,
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
	Scheduler *kvscheduler.Scheduler
	IfPlugin  *ifplugin.IfPlugin
	L3Plugin  *l3plugin.L3Plugin
	Datasync  *kvdbsync.Plugin
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

		veth2LogicalName = "myVETH2"
		veth2HostName    = "veth2"

		veth3LogicalName = "myVETH3"
		veth3HostName    = "veth3"

		veth4LogicalName = "myVETH4"
		veth4HostName    = "veth4"

		mycroservice1Name = "myMicroservice1"
		mycroservice1ID   = "microservice1"

		mycroservice2Name = "myMicroservice2"
		mycroservice2ID   = "microservice2"

		namedNs1 = "ns1"
		namedNs2 = "ns2"
	)

	veth1 := &interfaces.LinuxInterface{
		Name:        veth1LogicalName,
		Type:        interfaces.LinuxInterface_VETH,
		Enabled:     true,
		PhysAddress: "66:66:66:66:66:66",
		IpAddresses: []string{
			"192.168.20.1/24",
		},
		Mtu:        1800,
		HostIfName: veth1HostName,
		Link: &interfaces.LinuxInterface_Veth{
			Veth: &interfaces.LinuxInterface_VethLink{PeerIfName: veth2LogicalName},
		},
		/*Namespace: &namespace.LinuxNetNamespace{
			Type:      namespace.LinuxNetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroservice1ID,
		},*/
	}

	arpForVeth1 := &l3.LinuxStaticARPEntry{
		Interface: veth1LogicalName,
		IpAddress: "192.168.20.100",
		HwAddress: "b3:12:12:45:A7:B7",
	}

	routeForVeth1 := &l3.LinuxStaticRoute{
		OutgoingInterface: veth1LogicalName,
		Scope:             l3.LinuxStaticRoute_LINK,
		DstNetwork:        "10.8.0.0/16",
	}

	route2ForVeth1 := &l3.LinuxStaticRoute{
		OutgoingInterface: veth1LogicalName,
		Scope:             l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        "11.11.11.0/24",
		GwAddr:            "10.8.14.14",
		Metric:            50,
	}

	veth2 := &interfaces.LinuxInterface{
		Name:        veth2LogicalName,
		Type:        interfaces.LinuxInterface_VETH,
		Enabled:     true,
		PhysAddress: "44:44:44:55:55:55",
		IpAddresses: []string{
			"192.168.20.2/24",
		},
		Mtu:        1500,
		HostIfName: veth2HostName,
		Link: &interfaces.LinuxInterface_Veth{
			Veth: &interfaces.LinuxInterface_VethLink{PeerIfName: veth1LogicalName},
		},
		/*Namespace: &namespace.LinuxNetNamespace{
			Type:      namespace.LinuxNetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroservice2ID,
		},*/
	}

	routeForVeth2 := &l3.LinuxStaticRoute{
		OutgoingInterface: veth2LogicalName,
		Scope:             l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        "12.12.12.0/24",
		GwAddr:            "192.168.20.200",
		Metric:            50,
	}

	arpForVeth2 := &l3.LinuxStaticARPEntry{
		Interface: veth2LogicalName,
		IpAddress: "192.168.20.130",
		HwAddress: "b3:12:66:66:c7:c7",
	}

	veth3 := &interfaces.LinuxInterface{
		Name:    veth3LogicalName,
		Type:    interfaces.LinuxInterface_VETH,
		Enabled: true,
		IpAddresses: []string{
			"192.168.40.3/24",
		},
		Mtu:        1800,
		HostIfName: veth3HostName,
		Link: &interfaces.LinuxInterface_Veth{
			Veth: &interfaces.LinuxInterface_VethLink{PeerIfName: veth4LogicalName},
		},
		Namespace: &namespace.LinuxNetNamespace{
			Type:      namespace.LinuxNetNamespace_NETNS_REF_NSID,
			Reference: namedNs1,
		},
	}

	veth4 := &interfaces.LinuxInterface{
		Name:    veth4LogicalName,
		Type:    interfaces.LinuxInterface_VETH,
		Enabled: true,
		IpAddresses: []string{
			"192.168.40.4/24",
		},
		HostIfName: veth4HostName,
		Link: &interfaces.LinuxInterface_Veth{
			Veth: &interfaces.LinuxInterface_VethLink{PeerIfName: veth3LogicalName},
		},
	}

	// resync

	time.Sleep(time.Second * 2)
	fmt.Println("=== RESYNC 0 ===")

	txn := localclient.DataResyncRequest("example")
	err := txn.
		LinuxInterface(veth2).
		LinuxInterface(veth1).
		LinuxInterface(veth3).
		LinuxInterface(veth4).
		LinuxArpEntry(arpForVeth1).
		LinuxArpEntry(arpForVeth2).
		LinuxRoute(routeForVeth1).
		LinuxRoute(route2ForVeth1).
		LinuxRoute(routeForVeth2).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

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

	// Test interface metadata map
	interfaceIndex := plugin.IfPlugin.GetInterfaceIndex()
	metadata, exists := interfaceIndex.LookupByName(veth1HostName)
	fmt.Printf("Interface %s: found=%t, meta=%v\n", veth1HostName, exists, metadata)
}
