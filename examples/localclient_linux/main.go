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
	"os"
	"sync"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/vpp-agent/clientv1/linux/localclient"

	linux_intf "github.com/ligato/vpp-agent/plugins/linuxplugin/model/interfaces"

	vpp_intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	vpp_l2 "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"

	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/flavors/local"
)

// init sets the default logging level
func init() {
	log.DefaultLogger().SetOutput(os.Stdout)
	log.DefaultLogger().SetLevel(logging.DebugLevel)
}

/********
 * Main *
 ********/

// Start Agent plugins selected for this example
func main() {
	// Init close channel to stop the example
	closeChannel := make(chan struct{}, 1)

	flavor := local.FlavorVppLocal{}
	// Example plugin and dependencies
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{}}
	// Create new agent
	agentVar := core.NewAgent(log.DefaultLogger(), 15*time.Second, append(flavor.Plugins(), examplePlugin)...)

	// End when the localhost example is finished
	go closeExample("localhost example finished", closeChannel)

	core.EventLoopWithInterrupt(agentVar, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(40 * time.Second)
	log.DefaultLogger().Info(message)
	closeChannel <- struct{}{}
}

/******************
 * Example plugin *
 ******************/

// PluginID of example plugin
const PluginID core.PluginName = "example-plugin"

// ExamplePlugin demonstrates the use of the localclient to locally transport example configuration
// into linuxplugin and default VPP plugins.
type ExamplePlugin struct {
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// Init initializes example plugin.
func (plugin *ExamplePlugin) Init() error {
	// apply initial VPP configuration
	plugin.resyncLinuxAndVpp()

	// schedule reconfiguration
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())
	plugin.wg.Add(1)
	go plugin.reconfigureLinuxAndVPP(ctx)

	log.DefaultLogger().Info("Initialization of the example plugin has completed")
	return nil
}

// Close cleans up the resources.
func (plugin *ExamplePlugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	log.DefaultLogger().Info("Closed example plugin")
	return nil
}

// resyncLinuxAndVPP propagates snapshot of the whole initial configuration to linuxplugin and VPP plugins.
func (plugin *ExamplePlugin) resyncLinuxAndVpp() {
	err := localclient.DataResyncRequest(PluginID).
		LinuxInterface(&veth11DefaultNs).
		LinuxInterface(&veth12).
		VppInterface(&afpacket1).
		VppInterface(&tap1).
		Send().ReceiveReply()
	if err != nil {
		log.DefaultLogger().Errorf("Failed to apply initial Linux&VPP configuration: %v", err)
	} else {
		log.DefaultLogger().Info("Successfully applied initial Linux&VPP configuration")
	}
}

// reconfigureLinuxAndVPP simulates a set of changes in the configuration related to linuxplugin and VPP plugins.
func (plugin *ExamplePlugin) reconfigureLinuxAndVPP(ctx context.Context) {
	select {
	case <-time.After(20 * time.Second):
		// simulate configuration change exactly 20seconds after resync
		err := localclient.DataChangeRequest(PluginID).
			Put().
			LinuxInterface(&veth11Ns1).     /* move veth11 into the namespace "ns1" */
			LinuxInterface(&veth12WithMtu). /* reconfigure veth12 -- explicitly set Mtu to 1000 */
			LinuxInterface(&veth21Ns2).     /* create veth21-veth22 pair, put veth21 into the namespace "ns2" */
			LinuxInterface(&veth22).        /* enable veth22, keep default configuration */
			VppInterface(&afpacket2).       /* create afpacket2 interface and attach it to veth2 */
			BD(&BDAfpackets).               /* put afpacket1 and afpacket2 into the same bridge domain */
			Delete().
			VppInterface(tap1.Name). /* remove the tap interface */
			Send().ReceiveReply()
		if err != nil {
			log.DefaultLogger().Errorf("Failed to reconfigure Linux&VPP: %v", err)
		} else {
			log.DefaultLogger().Info("Successfully reconfigured Linux&VPP")
		}
	case <-ctx.Done():
		// cancel the scheduled re-configuration
		log.DefaultLogger().Info("Planned Linux&VPP re-configuration was canceled")
	}
	plugin.wg.Done()
}

/*************************
 * Example plugin config *
 *************************/

/**********************************************
 * After Resync                               *
 *                                            *
 *  +--------------------------------------+  *
 *  |                                      |  *
 *  |                                      |  *
 *  | +------------+          +-------+    |  *
 *  | | afpacket1  |          | tap1  |    |  *
 *  | +-----+------+          +---+---+    |  *
 *  |       |                     |        |  *
 *  +-------+---------------------+--------+  *
 *          |                     |           *
 *  +-------+--------+      +-----+------+    *
 *  | veth12         |      | linux-tap1 |    *
 *  | DEFAULT CONFIG |      +------------+    *
 *  +-------+--------+                        *
 *          |                                 *
 *  +-------+---------+                       *
 *  | veth11          |                       *
 *  | IP: 10.0.0.1/24 |                       *
 *  | DEFAULT NS      |                       *
 *  +-----------------+                       *
 *                                            *
 **********************************************/

/***********************************************
 * After Data Change Request                   *
 *                                             *
 *  +---------------------------------------+  *
 *  |       +-- Bridge domain --+           |  *
 *  |       |                   |           |  *
 *  | +-----+------+      +-----+------+    |  *
 *  | | afpacket1  |      | afpacket2  |    |  *
 *  | +-----+------+      +-----+------+    |  *
 *  |       |                   |           |  *
 *  +-------+-------------------+-----------+  *
 *          |                   |              *
 *  +-------+--------+  +-------+--------+     *
 *  | veth12         |  | veth22         |     *
 *  | MTU: 1000      |  | DEFAULT CONFIG |     *
 *  +-------+--------+  +-------+--------+     *
 *          |                   |              *
 *  +-------+---------+ +-------+---------+    *
 *  | veth11          | | veth21          |    *
 *  | IP: 10.0.0.1/24 | | IP: 10.0.0.2/24 |    *
 *  | NAMESPACE: ns1  | | NAMESPACE: ns2  |    *
 *  +-----------------+ +-----------------+    *
 ***********************************************/

var (
	// tap1 is an example tap interface configuration.
	tap1 = vpp_intf.Interfaces_Interface{
		Name:    "tap1",
		Type:    vpp_intf.InterfaceType_TAP_INTERFACE,
		Enabled: false,
		Tap: &vpp_intf.Interfaces_Interface_Tap{
			HostIfName: "linux-tap1",
		},
		Mtu: 1500,
	}

	// veth11DefaultNs is one end of the veth11-veth12 VETH pair, put into the default namespace and NOT attached to VPP
	veth11DefaultNs = linux_intf.LinuxInterfaces_Interface{
		Name:       "veth11",
		HostIfName: "veth11",
		Type:       linux_intf.LinuxInterfaces_VETH,
		Enabled:    true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth12",
		},
		IpAddresses: []string{"10.0.0.1/24"},
	}

	// veth11Ns1 is veth11DefaultNs moved to the namespace "ns1"
	veth11Ns1 = linux_intf.LinuxInterfaces_Interface{
		Name:       "veth11",
		HostIfName: "veth11",
		Type:       linux_intf.LinuxInterfaces_VETH,
		Enabled:    true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth12",
		},
		IpAddresses: []string{"10.0.0.1/24"},
		Namespace: &linux_intf.LinuxInterfaces_Interface_Namespace{
			Type: linux_intf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
			Name: "ns1",
		},
	}

	// veth12 is one end of the veth11-veth12 VETH pair, put into the default namespace and attached to VPP
	veth12 = linux_intf.LinuxInterfaces_Interface{
		Name:       "veth12",
		HostIfName: "veth12",
		Type:       linux_intf.LinuxInterfaces_VETH,
		Enabled:    true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth11",
		},
	}

	// veth12WithMtu is like veth12, but MTU is reconfigured
	veth12WithMtu = linux_intf.LinuxInterfaces_Interface{
		Name:       "veth12",
		HostIfName: "veth12",
		Type:       linux_intf.LinuxInterfaces_VETH,
		Enabled:    true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth11",
		},
		Mtu: 1000,
	}

	// veth21Ns2 is one end of the veth21-veth22 VETH pair, put into the namespace "ns2" and NOT attached to VPP
	veth21Ns2 = linux_intf.LinuxInterfaces_Interface{
		Name:       "veth21",
		HostIfName: "veth21",
		Type:       linux_intf.LinuxInterfaces_VETH,
		Enabled:    true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth22",
		},
		IpAddresses: []string{"10.0.0.2/24"},
		Namespace: &linux_intf.LinuxInterfaces_Interface_Namespace{
			Type: linux_intf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
			Name: "ns2",
		},
	}

	// veth22 is one end of the veth21-veth22 VETH pair, put into the default namespace and attached to VPP
	veth22 = linux_intf.LinuxInterfaces_Interface{
		Name:       "veth22",
		HostIfName: "veth22",
		Type:       linux_intf.LinuxInterfaces_VETH,
		Enabled:    true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth21",
		},
	}

	// afpacket1 is attached to veth12 interface through the AF_PACKET socket.
	afpacket1 = vpp_intf.Interfaces_Interface{
		Name:    "afpacket1",
		Type:    vpp_intf.InterfaceType_AF_PACKET_INTERFACE,
		Enabled: true,
		Afpacket: &vpp_intf.Interfaces_Interface_Afpacket{
			HostIfName: "veth12",
		},
	}

	// afpacket2 is attached to veth22 interface through the AF_PACKET socket.
	afpacket2 = vpp_intf.Interfaces_Interface{
		Name:    "afpacket2",
		Type:    vpp_intf.InterfaceType_AF_PACKET_INTERFACE,
		Enabled: true,
		Afpacket: &vpp_intf.Interfaces_Interface_Afpacket{
			HostIfName: "veth22",
		},
	}

	// BDAfpackets is a bridge domain with both afpacket interfaces in it.
	BDAfpackets = vpp_l2.BridgeDomains_BridgeDomain{
		Name:                "br1",
		Flood:               true,
		UnknownUnicastFlood: true,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0, /* means disable aging */
		Interfaces: []*vpp_l2.BridgeDomains_BridgeDomain_Interfaces{
			{
				Name: "afpacket1",
				BridgedVirtualInterface: false,
			}, {
				Name: "afpacket2",
				BridgedVirtualInterface: false,
			},
		},
	}
)
