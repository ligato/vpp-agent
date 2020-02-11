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

	"go.ligato.io/cn-infra/v2/agent"

	"github.com/golang/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/clientv2/linux/localclient"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

/*
	This example demonstrates VPP L3Plugin using KVScheduler, with focus on VRF tables.
*/

func main() {
	ep := &ExamplePlugin{
		Orchestrator: &orchestrator.DefaultPlugin,
		VPPIfPlugin:  &ifplugin.DefaultPlugin,
		VPPL3Plugin:  &l3plugin.DefaultPlugin,
		VPPNATPlugin: &natplugin.DefaultPlugin,
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
	VPPNATPlugin *natplugin.NATPlugin
	Orchestrator *orchestrator.Plugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "vpp-vrf-example"
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
		VppInterface(tap1).
		VppInterface(tapUnnum).
		VppInterface(vxlan1).
		StaticRoute(route1).
		StaticRoute(route2).
		NAT44Global(natGlobal).
		DNAT44(dnat1).
		VrfTable(vrfV4).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// data change
	time.Sleep(time.Second * 15)
	fmt.Println("=== CHANGE #1 ===")

	tap1 = proto.Clone(tap1).(*interfaces.Interface)
	tap1.IpAddresses = append(tap1.IpAddresses, "2001:db8:a0b:12f0::1/64")
	dnat1 = proto.Clone(dnat1).(*nat.DNat44)
	dnat1.IdMappings = []*nat.DNat44_IdentityMapping{
		{
			VrfId:     4, // not available
			IpAddress: "10.11.12.13",
			Port:      443,
			Protocol:  nat.DNat44_TCP,
		},
	}

	txn2 := localclient.DataChangeRequest("example")
	err = txn2.
		Put().
		VrfTable(vrfV6).
		VppInterface(tap1). // add IPv6 address
		//		DNAT44(dnat1).      // will become pending
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	/*
		// data change #2
		time.Sleep(time.Second * 60)
		fmt.Println("=== CHANGE #2 ===")

		txn3 := localclient.DataChangeRequest("example")
		err = txn3.
			Delete().
			VrfTable(vrfV6.Id, vrfV6.Protocol).
			VrfTable(vrfV4.Id, vrfV4.Protocol).
			Send().ReceiveReply()
		if err != nil {
			fmt.Println(err)
			return
		}
	*/
}

var (
	tap1 = &interfaces.Interface{
		Name:        "tap1",
		Enabled:     true,
		Type:        interfaces.Interface_TAP,
		Vrf:         3,
		IpAddresses: []string{"3.3.0.1/16"},
		Link: &interfaces.Interface_Tap{
			Tap: &interfaces.TapLink{
				Version: 2,
			},
		},
	}
	tapUnnum = &interfaces.Interface{
		Name:    "tap-unnumbered",
		Enabled: true,
		Type:    interfaces.Interface_TAP,
		Vrf:     6, // ignored
		Unnumbered: &interfaces.Interface_Unnumbered{
			InterfaceWithIp: tap1.Name,
		},
		Link: &interfaces.Interface_Tap{
			Tap: &interfaces.TapLink{
				Version: 2,
			},
		},
	}
	vxlan1 = &interfaces.Interface{
		Name:    "vxlan1",
		Enabled: true,
		Type:    interfaces.Interface_VXLAN_TUNNEL,
		Vrf:     3,
		Link: &interfaces.Interface_Vxlan{
			Vxlan: &interfaces.VxlanLink{ // depends on VRF 3 for IPv6
				SrcAddress: "2001:db8:85a3::8a2e:370:7334",
				DstAddress: "2001:db8:85a3::7cbb:942:1234",
			},
		},
	}
	route1 = &l3.Route{
		VrfId:             3,
		DstNetwork:        "10.11.1.0/24",
		OutgoingInterface: tap1.GetName(),
		Weight:            100,
		NextHopAddr:       "0.0.0.0",
	}
	route2 = &l3.Route{
		VrfId:             3,
		DstNetwork:        "2001:db8:a0b:12f0::/64",
		OutgoingInterface: tap1.GetName(),
		Weight:            100,
		NextHopAddr:       "::",
	}
	natGlobal = &nat.Nat44Global{
		AddressPool: []*nat.Nat44Global_Address{
			{
				Address: "10.14.5.5",
			},
			{
				Address: "10.12.1.1",
				VrfId:   3,
			},
		},
	}
	dnat1 = &nat.DNat44{
		Label: "dnat1",
		StMappings: []*nat.DNat44_StaticMapping{
			{
				ExternalIp:   "5.5.10.10",
				ExternalPort: 80,
				LocalIps: []*nat.DNat44_StaticMapping_LocalIP{
					{
						VrfId:       3,
						LocalIp:     "192.168.10.5",
						LocalPort:   8080,
						Probability: 1,
					},
					{
						VrfId:       3,
						LocalIp:     "192.168.10.19",
						LocalPort:   8080,
						Probability: 2,
					},
				},
			},
			{
				ExternalInterface: tap1.Name,
				ExternalPort:      80,
				LocalIps: []*nat.DNat44_StaticMapping_LocalIP{
					{
						VrfId:     3,
						LocalIp:   "192.168.17.4",
						LocalPort: 8080,
					},
				},
			},
		},
	}
	vrfV4 = &l3.VrfTable{
		Id: 3,
		//Label: "vrf3-IPv4",
	}
	vrfV6 = &l3.VrfTable{
		Id:       3,
		Protocol: l3.VrfTable_IPV6,
		Label:    "vrf3-IPv6",
	}
)
