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
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/l3"

	"github.com/ligato/vpp-agent/clientv2/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/kvscheduler"
	vpp_ifplugin "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	vpp_l3plugin "github.com/ligato/vpp-agent/plugins/vppv2/l3plugin"
)

/*
	This example demonstrates example using KVScheduler.
*/

func main() {
	// Set watcher for KVScheduler.
	kvscheduler.DefaultPlugin.Watcher = local.Get()

	vppIfPlugin := vpp_ifplugin.NewPlugin()
	vppL3Plugin := vpp_l3plugin.NewPlugin()
	vppL3Plugin.IfPlugin = vppIfPlugin

	ep := &ExamplePlugin{
		Scheduler:   &kvscheduler.DefaultPlugin,
		VPPIfPlugin: vppIfPlugin,
		VPPL3Plugin: vppL3Plugin,
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
	Scheduler   *kvscheduler.Scheduler
	VPPIfPlugin *vpp_ifplugin.IfPlugin
	VPPL3Plugin *vpp_l3plugin.L3Plugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "l3-example"
}

// Init handles initialization phase.
func (p *ExamplePlugin) Init() error {
	return nil
}

// AfterInit handles phase after initialization.
func (p *ExamplePlugin) AfterInit() error {
	go p.testLocalClientWithScheduler()
	return nil
}

// Close cleans up the resources.
func (p *ExamplePlugin) Close() error {
	return nil
}
func (p *ExamplePlugin) testLocalClientWithScheduler() {
	memif0 := &interfaces.Interface{
		Name:        "memif0",
		Enabled:     true,
		Type:        interfaces.Interface_MEMORY_INTERFACE,
		IpAddresses: []string{"3.3.0.1/16"},
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.Interface_MemifLink{
				Id:             1,
				Master:         true,
				Secret:         "secret",
				SocketFilename: "/tmp/memif1.sock",
			},
		},
	}
	route0 := &l3.StaticRoute{
		DstNetwork:        "10.10.1.0/24",
		OutgoingInterface: "memif0",
		Weight:            200,
	}
	route1 := &l3.StaticRoute{
		DstNetwork:        "2001:DB8::0001/32",
		OutgoingInterface: "memif0",
		Weight:            100,
	}
	arp0 := &l3.ARPEntry{
		Interface:   "memif0",
		PhysAddress: "33:33:33:33:33:33",
		IpAddress:   "3.3.3.3",
		Static:      true,
	}
	proxyArp := &l3.ProxyARP{
		Ranges: []*l3.ProxyARP_Range{
			{FirstIpAddr: "10.10.1.1", LastIpAddr: "10.10.1.255"},
		},
		Interfaces: []*l3.ProxyARP_Interface{
			{Name: "memif0"},
		},
	}
	ipScanNeighbor := &l3.IPScanNeighbor{
		Mode:           l3.IPScanNeighbor_IPv4,
		ScanInterval:   1,
		ScanIntDelay:   1,
		MaxProcTime:    20,
		MaxUpdate:      10,
		StaleThreshold: 4,
	}

	// resync

	time.Sleep(time.Second / 2)
	fmt.Println("=== RESYNC 0 ===")

	txn := localclient.DataResyncRequest("example")
	err := txn.
		VppInterface(memif0).
		StaticRoute(route0).
		StaticRoute(route1).
		Arp(arp0).
		ProxyArp(proxyArp).
		IPScanNeighbor(ipScanNeighbor).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// data change
	time.Sleep(time.Second * 1)
	fmt.Println("=== CHANGE 1 ===")

	route0.OutgoingInterface = ""
	arp0.PhysAddress = "22:22:22:22:22:22"
	proxyArp.Ranges = append(proxyArp.Ranges, &l3.ProxyARP_Range{
		FirstIpAddr: "10.10.2.1", LastIpAddr: "10.10.2.255",
	})
	proxyArp.Interfaces = nil

	txn2 := localclient.DataChangeRequest("example")
	err = txn2.
		Put().
		StaticRoute(route0).
		Delete().
		StaticRoute(route1.VrfId, route1.DstNetwork, route1.NextHopAddr).
		Put().
		Arp(arp0).ProxyArp(proxyArp).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}
}
