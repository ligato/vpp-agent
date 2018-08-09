// Testing the concept of KVScheduler on a example configurator to see how it would
// look like.
package main

import (
	"fmt"

	"github.com/ligato/cn-infra/kvscheduler"

	"github.com/ligato/vpp-agent/clientv1/vpp/localclient"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/examples/scheduler_example/ifplugin"
)

func main() {
	const tapName = "myTap"
	tap1 := &interfaces.Interfaces_Interface{
		Name:        tapName,
		Description: "this is my tap",
		Type:        interfaces.InterfaceType_TAP_INTERFACE,
		Enabled:     true,
		PhysAddress: "12:E4:0E:D5:BC:DC",
		IpAddresses: []string{
			"192.168.20.3/24",
		},
		Tap: &interfaces.Interfaces_Interface_Tap{
			HostIfName: "tap-host",
		},
	}

	tap2 := &interfaces.Interfaces_Interface{
		Name:        tapName,
		Description: "this is my tap",
		Type:        interfaces.InterfaceType_TAP_INTERFACE,
		Enabled:     true,
		PhysAddress: "BB:BB:BB:AA:AA:AA",
		IpAddresses: []string{
			"192.168.20.3/24",
		},
		Tap: &interfaces.Interfaces_Interface_Tap{
			HostIfName: "tap-host",
		},
	}

	tap3 := &interfaces.Interfaces_Interface{
		Name:        tapName,
		Description: "this is my tap",
		Type:        interfaces.InterfaceType_TAP_INTERFACE,
		Enabled:     true,
		PhysAddress: "BB:BB:BB:AA:AA:AA",
		IpAddresses: []string{
			"192.168.20.3/24",
		},
		Tap: &interfaces.Interfaces_Interface_Tap{
			HostIfName: "different-host-if-name",
		},
	}

	// dependency injection
	scheduler := &kvscheduler.DefaultPlugin
	ifPlugin := ifplugin.IfPlugin{Deps: ifplugin.Deps{
		Scheduler: scheduler,
	}}

	// init phase
	scheduler.Init()
	ifPlugin.Init()

	// agent run-time
	txn := localclient.DataChangeRequest("example")
	txn.Put().Interface(tap1).Send().ReceiveReply()
	txn.Put().Interface(tap1).Send().ReceiveReply() /* no change */

	interfaceIndex := ifPlugin.GetInterfaceIndex()
	ifName := interfaceIndex.LookupByIP("192.168.20.3/24")[0]
	fmt.Printf("IP address 192.168.20.3/24 is used by interface %s\n", ifName)

	ifMeta, exists := interfaceIndex.LookupByName(ifName)
	if exists {
		fmt.Printf("Interface %s has sw_if_index=%d\n", ifName, ifMeta.GetIndex())
	}

	txn.Put().Interface(tap2).Send().ReceiveReply()
	txn.Put().Interface(tap1).Send().ReceiveReply()
	txn.Put().Interface(tap2).Send().ReceiveReply()
	txn.Put().Interface(tap3).Send().ReceiveReply() /* need to re-create */
	txn.Delete().Interface(tap1.Name).Send().ReceiveReply()
}
