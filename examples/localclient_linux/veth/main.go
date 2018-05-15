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

// Initial Data configures single veth pair where both linux interfaces veth11 and veth12 are configured in
// default namespace. Af packet interface is attached to veth11 and put to the bridge domain. The bridge domain
// will contain a second af packet which will be created in the second iteration (modify).
/**********************************************
 * Initial Data                               *
 *                                            *
 *  +--------------------------------------+  *
 *  |       +-- Bridge Domain --+          |  *
 *  |       |                   |          |  *
 *  | +------------+            |          |  *
 *  | | afpacket1  |        (afpacket2)    |  *
 *  | +-----+------+                       |  *
 *  |       |                              |  *
 *  +-------+------------------------------+  *
 *          |                                 *
 *  +-------+--------+                        *
 *  | veth11         |                        *
 *  | DEFAULT CONFIG |                        *
 *  +-------+--------+                        *
 *          |                                 *
 *  +-------+---------+                       *
 *  | veth12          |                       *
 *  | IP: 10.0.0.1/24 |                       *
 *  | DEFAULT NS      |                       *
 *  +-----------------+                       *
 *                                            *
 **********************************************/

// Modify changes MTU of the veth11, moves veth12 to the namespace ns1 and configures IP address to it. Also second
// branch veth21 - veth22 is configured including afpacket2. The new af packet is in the same bridge domain. This
// configuration allows to ping between veth12 and veth22 interfaces
/***********************************************
 * Modified Data                               *
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
 *  | veth11         |  | veth21         |     *
 *  | MTU: 1000      |  | DEFAULT CONFIG |     *
 *  +-------+--------+  +-------+--------+     *
 *          |                   |              *
 *  +-------+---------+ +-------+---------+    *
 *  | veth12          | | veth22          |    *
 *  | IP: 10.0.0.1/24 | | IP: 10.0.0.2/24 |    *
 *  | NAMESPACE: ns1  | | NAMESPACE: ns2  |    *
 *  +-----------------+ +-----------------+    *
 ***********************************************/

/* Vpp-agent Init and Close*/

// Start Agent plugins selected for this example.
func main() {
	// Init close channel to stop the example.
	closeChannel := make(chan struct{}, 1)

	agent := local.NewAgent(local.WithPlugins(func(flavor *local.FlavorVppLocal) []*core.NamedPlugin {
		examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &VethExamplePlugin{}}

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

/* VETH Example */

// PluginID of an example plugin.
const PluginID core.PluginName = "veth-example-plugin"

// VethExamplePlugin uses localclient to transport example veth and af-packet
// configuration to linuxplugin, eventually VPP plugins
type VethExamplePlugin struct {
	log    logging.Logger
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// Init initializes example plugin.
func (plugin *VethExamplePlugin) Init() error {
	// Logger
	plugin.log = log.DefaultLogger()
	plugin.log.SetLevel(logging.DebugLevel)
	plugin.log.Info("Initializing Veth example")

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

	plugin.log.Info("Veth example initialization done")
	return nil
}

// Close cleans up the resources.
func (plugin *VethExamplePlugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	log.DefaultLogger().Info("Closed Veth plugin")
	return nil
}

// Configure initial data
func (plugin *VethExamplePlugin) putInitialData() {
	plugin.log.Infof("Applying initial configuration")
	err := localclient.DataResyncRequest(PluginID).
		LinuxInterface(initialVeth11()).
		LinuxInterface(initialVeth12()).
		VppInterface(afPacket1()).
		BD(bridgeDomain()).
		Send().ReceiveReply()
	if err != nil {
		plugin.log.Errorf("Initial configuration failed: %v", err)
	} else {
		plugin.log.Info("Initial configuration successful")
	}
}

// Configure modified data
func (plugin *VethExamplePlugin) putModifiedData(ctx context.Context, timeout int) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		plugin.log.Infof("Applying modified configuration")
		// Simulate configuration change after timeout
		err := localclient.DataChangeRequest(PluginID).
			Put().
			LinuxInterface(modifiedVeth11()).
			LinuxInterface(modifiedVeth12()).
			LinuxInterface(veth21()).
			LinuxInterface(veth22()).
			VppInterface(afPacket2()).
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

func initialVeth11() *linux_intf.LinuxInterfaces_Interface {
	return &linux_intf.LinuxInterfaces_Interface{
		Name:    "veth11",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth12",
		},
	}
}

func modifiedVeth11() *linux_intf.LinuxInterfaces_Interface {
	return &linux_intf.LinuxInterfaces_Interface{
		Name:    "veth11",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth12",
		},
		Mtu: 1000,
	}
}

func initialVeth12() *linux_intf.LinuxInterfaces_Interface {
	return &linux_intf.LinuxInterfaces_Interface{
		Name:    "veth12",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth11",
		},
	}
}

func modifiedVeth12() *linux_intf.LinuxInterfaces_Interface {
	return &linux_intf.LinuxInterfaces_Interface{
		Name:    "veth12",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth11",
		},
		IpAddresses: []string{"10.0.0.1/24"},
		PhysAddress: "D2:74:8C:12:67:D2",
		Namespace: &linux_intf.LinuxInterfaces_Interface_Namespace{
			Type: linux_intf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
			Name: "ns1",
		},
	}
}

func veth21() *linux_intf.LinuxInterfaces_Interface {
	return &linux_intf.LinuxInterfaces_Interface{
		Name:    "veth21",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth22",
		},
	}
}

func veth22() *linux_intf.LinuxInterfaces_Interface {
	return &linux_intf.LinuxInterfaces_Interface{
		Name:    "veth22",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth21",
		},
		IpAddresses: []string{"10.0.0.2/24"},
		PhysAddress: "92:C7:42:67:AB:CD",
		Namespace: &linux_intf.LinuxInterfaces_Interface_Namespace{
			Type: linux_intf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
			Name: "ns2",
		},
	}
}

func afPacket1() *vpp_intf.Interfaces_Interface {
	return &vpp_intf.Interfaces_Interface{
		Name:    "afpacket1",
		Type:    vpp_intf.InterfaceType_AF_PACKET_INTERFACE,
		Enabled: true,
		Afpacket: &vpp_intf.Interfaces_Interface_Afpacket{
			HostIfName: "veth11",
		},
	}
}

func afPacket2() *vpp_intf.Interfaces_Interface {
	return &vpp_intf.Interfaces_Interface{
		Name:    "afpacket2",
		Type:    vpp_intf.InterfaceType_AF_PACKET_INTERFACE,
		Enabled: true,
		Afpacket: &vpp_intf.Interfaces_Interface_Afpacket{
			HostIfName: "veth21",
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
				Name: "afpacket1",
				BridgedVirtualInterface: false,
			}, {
				Name: "afpacket2",
				BridgedVirtualInterface: false,
			},
		},
	}
}
