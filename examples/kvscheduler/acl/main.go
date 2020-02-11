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

	"go.ligato.io/vpp-agent/v3/clientv2/vpp/localclient"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	vpp_aclplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin"
	vpp_ifplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

/*
	This example demonstrates KVScheduler-based ACLPlugin.
*/

func main() {
	ep := &ExamplePlugin{
		Orchestrator: &orchestrator.DefaultPlugin,
		VPPIfPlugin:  &vpp_ifplugin.DefaultPlugin,
		VPPACLPlugin: &vpp_aclplugin.DefaultPlugin,
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
	VPPIfPlugin  *vpp_ifplugin.IfPlugin
	VPPACLPlugin *vpp_aclplugin.ACLPlugin
	Orchestrator *orchestrator.Plugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "acl-example"
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
		Interface(memif0).
		ACL(acl0).
		ACL(acl1).
		ACL(acl3).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// data change
	time.Sleep(time.Second * 10)
	fmt.Println("=== CHANGE ===")

	acl1.Interfaces = nil
	acl0.Interfaces.Egress = nil
	acl3.Rules[0].IpRule.Ip.SourceNetwork = "0.0.0.0/0" // this is actually equivalent to unspecified field

	txn2 := localclient.DataChangeRequest("example")
	err = txn2.Put().
		ACL(acl0).
		ACL(acl1).
		ACL(acl3).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}
}

var (
	memif0 = &interfaces.Interface{
		Name:    "memif0",
		Enabled: true,
		Type:    interfaces.Interface_MEMIF,
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             1,
				Master:         true,
				Secret:         "secret",
				SocketFilename: "/tmp/memif1.sock",
			},
		},
	}
	acl0 = &acl.ACL{
		Name: "acl0",
		Rules: []*acl.ACL_Rule{
			{
				Action: acl.ACL_Rule_PERMIT,
				IpRule: &acl.ACL_Rule_IpRule{
					Ip: &acl.ACL_Rule_IpRule_Ip{
						SourceNetwork:      "10.0.0.0/24",
						DestinationNetwork: "20.0.0.0/24",
					},
				},
			},
		},
		Interfaces: &acl.ACL_Interfaces{
			Ingress: []string{"memif0"},
			Egress:  []string{"memif0"},
		},
	}
	acl1 = &acl.ACL{
		Name: "acl1",
		Rules: []*acl.ACL_Rule{
			{
				Action: acl.ACL_Rule_PERMIT,
				MacipRule: &acl.ACL_Rule_MacIpRule{
					SourceAddress:        "192.168.0.1",
					SourceAddressPrefix:  16,
					SourceMacAddress:     "b2:74:8c:12:67:d2",
					SourceMacAddressMask: "ff:ff:ff:ff:00:00",
				},
			},
		},
		Interfaces: &acl.ACL_Interfaces{
			Ingress: []string{"memif0"},
		},
	}
	acl3 = &acl.ACL{
		Name: "acl3",
		Rules: []*acl.ACL_Rule{
			{
				Action: acl.ACL_Rule_DENY,
				IpRule: &acl.ACL_Rule_IpRule{
					Ip: &acl.ACL_Rule_IpRule_Ip{
						// SourceNetwork is unspecified (ANY)
						DestinationNetwork: "30.0.0.0/8",
					},
				},
			},
		},
		Interfaces: &acl.ACL_Interfaces{
			Egress: []string{"memif0"},
		},
	}
)
