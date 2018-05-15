// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"sync"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/clientv1/linux/localclient"
	"github.com/ligato/vpp-agent/flavors/local"
	linux_intf "github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
	vpp_intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	vpp_l2 "github.com/ligato/vpp-agent/plugins/vpp/model/l2"
	"github.com/namsral/flag"
)

var (
	timeout = flag.Int("timeout", 20, "Timeout between applying of initial and modified configuration in seconds")
)

/* Confgiuration */

// Initial Data configures TAP interface on the vpp with other end in the default namespace. Linux-tap is then set with
// IP address. Also bridge domain is created for linux TAP interfaces
/**********************************************
 * Initial Data                               *
 *                                            *
 *  +--------------------------------------+  *
 *  |       +-- Bridge Domain --+          |  *
 *  |       |                   |          |  *
 *  |   +-------+               |          |  *
 *  |   | tap1  |             (tap2)       |  *
 *  |   +---+---+                          |  *
 *  |       |                              |  *
 *  +-------+------------------------------+  *
 *          |                                 *
 *  +-------+---------+                       *
 *  | linux-tap1      |                       *
 *  | IP: 10.0.0.2/24 |                       *
 *  +-----------------+                       *
 *                                            *
 *                                            *
 **********************************************/

// Modify sets IP address for tap1, moves linux host to namespace ns1 and configures second TAP interface with linux
// host in namespace ns2
/************************************************
 * Initial Data                                 *
 *                                              *
 *  +----------------------------------------+  *
 *  |       +-- Bridge Domain --+            |  *
 *  |       |                   |            |  *
 *  | +-----+--------+    *------+-------+   |  *
 *  | | tap1         |    | (tap2)       |   |  *
 *  | | 10.0.0.11/24 |    | 10.0.0.21/24 |   |  *
 *  | +-----+--------+    +------+-------+   |  *
 *  |       |                   |            |  *
 *  +-------+-------------------+------------+  *
 *          |                   |               *
 *  +-------+----------+   +-----+------------+ *
 *  | linux-tap1       |   | linux-tap2       | *
 *  | IP: 10.0.0.12/24 |   | IP: 10.0.0.22\24 | *
 *  | Namespace: ns1   |   | Namespace: ns2   | *
 *  +------------------+   +------------------+ *
 *                                              *
 *                                              *
 ************************************************/

/* Vpp-agent Init and Close*/

// Start Agent plugins selected for this example.
func main() {
	// Init close channel to stop the example.
	closeChannel := make(chan struct{}, 1)

	agent := local.NewAgent(local.WithPlugins(func(flavor *local.FlavorVppLocal) []*core.NamedPlugin {
		examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &TapExamplePlugin{}}

		return []*core.NamedPlugin{{examplePlugin.PluginName, examplePlugin}}
	}))

	// End when the localhost example is finished.
	go closeExample("localhost example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message.
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(time.Duration(*timeout+5) * time.Second)
	log.DefaultLogger().Info(message)
	closeChannel <- struct{}{}
}

/* TAP Example */

// PluginID of an example plugin.
const PluginID core.PluginName = "tap-example-plugin"

// TapExamplePlugin uses localclient to transport example tap and its linux end
// configuration to linuxplugin or VPP plugins
type TapExamplePlugin struct {
	log    logging.Logger
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// Init initializes example plugin.
func (plugin *TapExamplePlugin) Init() error {
	// Logger
	plugin.log = log.DefaultLogger()
	plugin.log.SetLevel(logging.DebugLevel)
	plugin.log.Info("Initializing Tap example")

	// Flags
	flag.Parse()
	plugin.log.Infof("Timeout between create and modify set to %d", *timeout)

	// Apply initial Linux/VPP configuration.
	plugin.putInitialData()

	// Schedule reconfiguration.
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())
	plugin.wg.Add(1)
	go plugin.putModifiedData(ctx, *timeout)

	plugin.log.Info("Tap example initialization done")
	return nil
}

// Close cleans up the resources.
func (plugin *TapExamplePlugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	log.DefaultLogger().Info("Closed Tap plugin")
	return nil
}

// Configure initial data
func (plugin *TapExamplePlugin) putInitialData() {
	plugin.log.Infof("Applying initial configuration")
	err := localclient.DataResyncRequest(PluginID).
		VppInterface(initialTap1()).
		LinuxInterface(initialLinuxTap1()).
		BD(bridgeDomain()).
		Send().ReceiveReply()
	if err != nil {
		plugin.log.Errorf("Initial configuration failed: %v", err)
	} else {
		plugin.log.Info("Initial configuration successful")
	}
}

// Configure modified data
func (plugin *TapExamplePlugin) putModifiedData(ctx context.Context, timeout int) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		plugin.log.Infof("Applying modified configuration")
		// Simulate configuration change after timeout
		err := localclient.DataChangeRequest(PluginID).
			Put().
			VppInterface(modifiedTap1()).
			VppInterface(tap2()).
			LinuxInterface(modifiedLinuxTap1()).
			LinuxInterface(linuxTap2()).
			Send().ReceiveReply()
		if err != nil {
			plugin.log.Errorf("Modified configuration failed: %v", err)
		} else {
			plugin.log.Info("Modified configuration successful")
		}
	case <-ctx.Done():
		// Cancel the scheduled re-configuration.
		plugin.log.Info("Modification of configuration canceled")
	}
	plugin.wg.Done()
}

/* Example Data */

func initialTap1() *vpp_intf.Interfaces_Interface {
	return &vpp_intf.Interfaces_Interface{
		Name:    "tap1",
		Type:    vpp_intf.InterfaceType_TAP_INTERFACE,
		Enabled: true,
		Tap: &vpp_intf.Interfaces_Interface_Tap{
			HostIfName: "linux-tap1",
		},
	}
}

func modifiedTap1() *vpp_intf.Interfaces_Interface {
	return &vpp_intf.Interfaces_Interface{
		Name:        "tap1",
		Type:        vpp_intf.InterfaceType_TAP_INTERFACE,
		Enabled:     true,
		PhysAddress: "12:E4:0E:D5:BC:DC",
		IpAddresses: []string{
			"10.0.0.11/24",
		},
		Tap: &vpp_intf.Interfaces_Interface_Tap{
			HostIfName: "linux-tap1",
		},
	}
}

func tap2() *vpp_intf.Interfaces_Interface {
	return &vpp_intf.Interfaces_Interface{
		Name:        "tap2",
		Type:        vpp_intf.InterfaceType_TAP_INTERFACE,
		Enabled:     true,
		PhysAddress: "D5:BC:DC:12:E4:0E",
		IpAddresses: []string{
			"10.0.0.21/24",
		},
		Tap: &vpp_intf.Interfaces_Interface_Tap{
			HostIfName: "linux-tap2",
		},
	}
}

func initialLinuxTap1() *linux_intf.LinuxInterfaces_Interface {
	return &linux_intf.LinuxInterfaces_Interface{

		Name:        "linux-tap1",
		Type:        linux_intf.LinuxInterfaces_AUTO_TAP,
		Enabled:     true,
		PhysAddress: "BC:FE:E9:5E:07:04",
		Mtu:         1500,
		IpAddresses: []string{
			"10.0.0.12/24",
		},
	}
}

func modifiedLinuxTap1() *linux_intf.LinuxInterfaces_Interface {
	return &linux_intf.LinuxInterfaces_Interface{

		Name:        "linux-tap1",
		Type:        linux_intf.LinuxInterfaces_AUTO_TAP,
		Enabled:     true,
		PhysAddress: "BC:FE:E9:5E:07:04",
		Namespace: &linux_intf.LinuxInterfaces_Interface_Namespace{
			Name: "ns1",
			Type: linux_intf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
		},
		Mtu: 1500,
		IpAddresses: []string{
			"10.0.0.12/24",
		},
	}
}

func linuxTap2() *linux_intf.LinuxInterfaces_Interface {
	return &linux_intf.LinuxInterfaces_Interface{

		Name:        "linux-tap2",
		Type:        linux_intf.LinuxInterfaces_AUTO_TAP,
		Enabled:     true,
		PhysAddress: "5E:07:04:BC:FE:E9",
		Namespace: &linux_intf.LinuxInterfaces_Interface_Namespace{
			Name: "ns2",
			Type: linux_intf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
		},
		Mtu: 1500,
		IpAddresses: []string{
			"10.0.0.22/24",
		},
	}
}

func bridgeDomain() *vpp_l2.BridgeDomains_BridgeDomain {
	return &vpp_l2.BridgeDomains_BridgeDomain{
		Name:                "br1",
		Flood:               true,
		UnknownUnicastFlood: true,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0, /* means disable aging */
		Interfaces: []*vpp_l2.BridgeDomains_BridgeDomain_Interfaces{
			{
				Name: "tap1",
				BridgedVirtualInterface: false,
			},
			{
				Name: "tap2",
				BridgedVirtualInterface: false,
			},
			{
				Name: "loop1",
				BridgedVirtualInterface: true,
			},
		},
	}
}
