// Testing the concept of KVScheduler on a mock VPP-ifplugin to see how it would
// look like.
package main

import (
	"fmt"
	"log"

	"github.com/ligato/cn-infra/agent"

	"github.com/ligato/vpp-agent/clientv1/vpp/localclient"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/examples/scheduler_example/ifplugin"
	"time"
)

func main() {
	// Inject dependencies to example plugin
	ep := &TapExamplePlugin{}
	ep.IfPlugin = &ifplugin.DefaultPlugin

	// Start Agent
	a := agent.NewAgent(
		agent.AllPlugins(ep),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

/* TAP Example */

// TapExamplePlugin uses localclient to transport example TAP configuration
// to (mock) ifplugin based on KVScheduler.
type TapExamplePlugin struct {
	Deps
}

// Deps is example plugin dependencies. Keep order of fields.
type Deps struct {
	IfPlugin   *ifplugin.IfPlugin
}

// PluginName represents name of plugin.
const PluginName = "tap-example"

// Init initializes example plugin.
func (plugin *TapExamplePlugin) Init() error {
	return nil
}

// AfterInit sends an example TAP configuration to mock IfPlugin using localclient + scheduler.
func (plugin *TapExamplePlugin) AfterInit() error {
	go plugin.testLocalClientWithScheduler()
	return nil
}

// Close cleans up the resources.
func (plugin *TapExamplePlugin) Close() error {
	return nil
}

// String returns plugin name
func (plugin *TapExamplePlugin) String() string {
	return PluginName
}


func (plugin *TapExamplePlugin) testLocalClientWithScheduler() {
	const tapName= "myTap"
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

	time.Sleep(time.Second*3)

	// create TAP interface
	txn := localclient.DataChangeRequest("example")
	err := txn.Put().Interface(tap1).Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second*3)

	// Update TAP config without any change
	err = txn.Put().Interface(tap1).Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Test interface metadata map
	interfaceIndex := plugin.IfPlugin.GetInterfaceIndex()
	ifName := interfaceIndex.LookupByIP("192.168.20.3/24")[0]
	fmt.Printf("IP address 192.168.20.3/24 is used by interface %s\n", ifName)
	ifMeta, exists := interfaceIndex.LookupByName(ifName)
	if exists {
		fmt.Printf("Interface %s has sw_if_index=%d\n", ifName, ifMeta.GetIndex())
	}

	time.Sleep(time.Second*3)

	// change TAP MAC address
	err = txn.Put().Interface(tap2).Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second*3)

	// Revert TAP MAC address
	err = txn.Put().Interface(tap1).Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second*3)

	// Change MAC address + TAP host name => requires re-create
	err = txn.Put().Interface(tap3).Send().ReceiveReply() /* need to re-create */
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second*3)

	// Delete the TAP interface
	err = txn.Delete().Interface(tap1.Name).Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}
}
