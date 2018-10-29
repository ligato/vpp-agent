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

	"github.com/ligato/vpp-agent/clientv2/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/kvscheduler"
	vpp_aclplugin "github.com/ligato/vpp-agent/plugins/vppv2/aclplugin"
	vpp_ifplugin "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/acl"
)

/*
	This example demonstrates KVScheduler-based aclplugin.
*/

func main() {
	// Set watcher for KVScheduler.
	kvscheduler.DefaultPlugin.Watcher = local.Get()

	vppIfPlugin := vpp_ifplugin.NewPlugin()
	vppACLPlugin := vpp_aclplugin.NewPlugin()
	vppACLPlugin.IfPlugin = vppIfPlugin

	ep := &ExamplePlugin{
		Scheduler:    &kvscheduler.DefaultPlugin,
		VPPIfPlugin:  vppIfPlugin,
		VPPACLPlugin: vppACLPlugin,
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
	Scheduler    *kvscheduler.Scheduler
	VPPIfPlugin  *vpp_ifplugin.IfPlugin
	VPPACLPlugin *vpp_aclplugin.ACLPlugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "acls-example"
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
		Name:    "memif0",
		Enabled: true,
		Type:    interfaces.Interface_MEMORY_INTERFACE,
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.Interface_MemifLink{
				Id:             1,
				Master:         true,
				Secret:         "secret",
				SocketFilename: "/tmp/memif1.sock",
			},
		},
	}
	acl0 := &acl.Acl{
		Name: "acl0",
		Rules: []*acl.Acl_Rule{
			{
				Action: acl.Acl_Rule_PERMIT,
				IpRule: &acl.Acl_Rule_IpRule{
					Ip: &acl.Acl_Rule_IpRule_Ip{
						SourceNetwork:      "10.0.0.0/24",
						DestinationNetwork: "20.0.0.0/24",
					},
				},
			},
		},
		Interfaces: &acl.Acl_Interfaces{
			Ingress: []string{"memif0"},
			Egress:  []string{"memif0"},
		},
	}
	acl1 := &acl.Acl{
		Name: "acl1",
		Rules: []*acl.Acl_Rule{
			{
				Action: acl.Acl_Rule_PERMIT,
				MacipRule: &acl.Acl_Rule_MacIpRule{
					SourceAddress:        "192.168.0.1",
					SourceAddressPrefix:  16,
					SourceMacAddress:     "b2:74:8c:12:67:d2",
					SourceMacAddressMask: "ff:ff:ff:ff:00:00",
				},
			},
		},
		Interfaces: &acl.Acl_Interfaces{
			Ingress: []string{"memif0"},
		},
	}

	// resync

	time.Sleep(time.Second / 2)
	fmt.Println("=== RESYNC 0 ===")

	txn := localclient.DataResyncRequest("example")
	err := txn.
		VppInterface(memif0).
		ACL(acl0).
		ACL(acl1).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// data change
	time.Sleep(time.Second * 1)
	fmt.Println("=== CHANGE 1 ===")

	acl1.Interfaces = nil
	acl0.Interfaces.Egress = nil

	txn2 := localclient.DataChangeRequest("example")
	err = txn2.Put().
		ACL(acl0).
		ACL(acl1).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}
}
