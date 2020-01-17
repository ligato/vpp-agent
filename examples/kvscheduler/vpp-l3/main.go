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

	"go.ligato.io/vpp-agent/v3/clientv2/linux/localclient"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

/*
	This example demonstrates example with VPP L3Plugin using KVScheduler.
*/

func main() {
	ep := &ExamplePlugin{
		Orchestrator: &orchestrator.DefaultPlugin,
		VPPIfPlugin:  &ifplugin.DefaultPlugin,
		VPPL3Plugin:  &l3plugin.DefaultPlugin,
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
	VPPIfPlugin  *ifplugin.IfPlugin
	VPPL3Plugin  *l3plugin.L3Plugin
	Orchestrator *orchestrator.Plugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "vpp-l3-example"
}

// Init handles initialization phase.
func (p *ExamplePlugin) Init() error {
	return nil
}

// AfterInit handles phase after initialization.
func (p *ExamplePlugin) AfterInit() error {
	go testLocalClientWithScheduler()
	return nil
}

// Close cleans up the resources.
func (p *ExamplePlugin) Close() error {
	return nil
}

func testLocalClientWithScheduler() {
	// initial resync
	time.Sleep(time.Second * 2)
	fmt.Println("=== RESYNC ===")

	txn := localclient.DataResyncRequest("example")
	err := txn.
		VppInterface(memif0).
		VppInterface(memif0_10).
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
	time.Sleep(time.Second * 10)
	fmt.Println("=== CHANGE ===")

	route0.OutgoingInterface = ""
	arp0.PhysAddress = "22:22:22:22:22:22"
	proxyArp.Ranges = append(proxyArp.Ranges, &l3.ProxyARP_Range{
		FirstIpAddr: "10.10.2.1", LastIpAddr: "10.10.2.255",
	})
	proxyArp.Interfaces = nil

	txn2 := localclient.DataChangeRequest("example")
	err = txn2.
		Put().
		VppInterface(memif0_10).
		StaticRoute(route0).
		Delete().
		VppInterface(memif0.Name).
		StaticRoute(route1.OutgoingInterface, route1.VrfId, route1.DstNetwork, route1.NextHopAddr).
		Put().
		Arp(arp0).
		ProxyArp(proxyArp).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}
}

var (
	memif0 = &interfaces.Interface{
		Name:        "memif0",
		Enabled:     true,
		Type:        interfaces.Interface_MEMIF,
		IpAddresses: []string{"3.3.0.1/16"},
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             1,
				Master:         true,
				Secret:         "secret",
				SocketFilename: "/tmp/memif1.sock",
			},
		},
	}
	memif0_10 = &interfaces.Interface{
		Name:        "memif0/10",
		Enabled:     true,
		Type:        interfaces.Interface_SUB_INTERFACE,
		IpAddresses: []string{"3.10.0.10/32"},
		Link: &interfaces.Interface_Sub{
			Sub: &interfaces.SubInterface{
				ParentName: "memif0",
				SubId:      10,
			},
		},
	}
	route0 = &l3.Route{
		DstNetwork:        "10.10.1.0/24",
		OutgoingInterface: "memif0",
		Weight:            200,
	}
	route1 = &l3.Route{
		DstNetwork:        "2001:DB8::0001/32",
		OutgoingInterface: "memif0",
		Weight:            100,
	}
	arp0 = &l3.ARPEntry{
		Interface:   "memif0",
		PhysAddress: "33:33:33:33:33:33",
		IpAddress:   "3.3.3.3",
		Static:      true,
	}
	proxyArp = &l3.ProxyARP{
		Ranges: []*l3.ProxyARP_Range{
			{FirstIpAddr: "10.10.1.1", LastIpAddr: "10.10.1.255"},
		},
		Interfaces: []*l3.ProxyARP_Interface{
			{Name: "memif0"},
		},
	}
	ipScanNeighbor = &l3.IPScanNeighbor{
		Mode:           l3.IPScanNeighbor_IPV4,
		ScanInterval:   1,
		ScanIntDelay:   1,
		MaxProcTime:    20,
		MaxUpdate:      0,
		StaleThreshold: 4,
	}
)
