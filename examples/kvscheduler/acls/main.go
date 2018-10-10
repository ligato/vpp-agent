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

	"github.com/ligato/vpp-agent/clientv2/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/kvscheduler"
	vpp_aclplugin "github.com/ligato/vpp-agent/plugins/vppv2/aclplugin"
	vpp_ifplugin "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/acl"
)

/*
	This example demonstrates example using KVScheduler.
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

// PluginName is a constant with name of main plugin.
const PluginName = "acls-example"

// ExamplePlugin is the main plugin which
// handles resync and changes in this example.
type ExamplePlugin struct {
	Scheduler    *kvscheduler.Scheduler
	VPPIfPlugin  *vpp_ifplugin.IfPlugin
	VPPACLPlugin *vpp_aclplugin.AclPlugin
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
	acl1 := &acl.Acl{
		Name: "acl1",
		Rules: []*acl.Acl_Rule{
			{
				AclAction: acl.Acl_Rule_PERMIT,
				Match: &acl.Acl_Rule_Match{
					MacipRule: &acl.Acl_Rule_Match_MacIpRule{
						SourceAddress:        "192.168.0.1",
						SourceAddressPrefix:  16,
						SourceMacAddress:     "b2:74:8c:12:67:d2",
						SourceMacAddressMask: "ff:ff:ff:ff:00:00",
					},
				},
			},
		},
	}

	// resync

	time.Sleep(time.Second * 1)
	fmt.Println("=== RESYNC 0 ===")

	txn := localclient.DataResyncRequest("example")
	err := txn.
		ACL(acl1).
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
	/*linuxIfIndex := plugin.LinuxIfPlugin.GetInterfaceIndex()
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
	*/
}
