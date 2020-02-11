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

	"log"

	"github.com/namsral/flag"
	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	localclient2 "go.ligato.io/vpp-agent/v3/clientv2/linux/localclient"
	"go.ligato.io/vpp-agent/v3/cmd/vpp-agent/app"
	linux_ifplugin "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	linux_nsplugin "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	vpp_ifplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	linux_intf "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_intf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
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

// PluginName represents name of plugin.
const PluginName = "veth-example"

// Start Agent plugins selected for this example.
func main() {
	// Set inter-dependency between VPP & Linux plugins
	vpp_ifplugin.DefaultPlugin.LinuxIfPlugin = &linux_ifplugin.DefaultPlugin
	vpp_ifplugin.DefaultPlugin.NsPlugin = &linux_nsplugin.DefaultPlugin
	linux_ifplugin.DefaultPlugin.VppIfPlugin = &vpp_ifplugin.DefaultPlugin

	// Init close channel to stop the example.
	exampleFinished := make(chan struct{})

	// Inject dependencies to example plugin
	ep := &VethExamplePlugin{
		Log:          logging.DefaultLogger,
		VPP:          app.DefaultVPP(),
		Linux:        app.DefaultLinux(),
		Orchestrator: &orchestrator.DefaultPlugin,
	}

	// Start Agent
	a := agent.NewAgent(
		agent.AllPlugins(ep),
		agent.QuitOnClose(exampleFinished),
	)
	if err := a.Run(); err != nil {
		log.Fatal()
	}

	go closeExample("localhost example finished", exampleFinished)
}

// Stop the agent with desired info message.
func closeExample(message string, exampleFinished chan struct{}) {
	time.Sleep(time.Duration(*timeout+5) * time.Second)
	logrus.DefaultLogger().Info(message)
	close(exampleFinished)
}

/* VETH Example */

// VethExamplePlugin uses localclient to transport example veth and af-packet
// configuration to linuxplugin, eventually VPP plugins
type VethExamplePlugin struct {
	Log logging.Logger
	app.VPP
	app.Linux
	Orchestrator *orchestrator.Plugin

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// String returns plugin name
func (p *VethExamplePlugin) String() string {
	return PluginName
}

// Init initializes example plugin.
func (p *VethExamplePlugin) Init() error {
	// Logger
	p.Log = logrus.DefaultLogger()
	p.Log.SetLevel(logging.DebugLevel)
	p.Log.Info("Initializing Veth example")

	// Flags
	flag.Parse()
	p.Log.Infof("Timeout between create and modify set to %d", *timeout)

	p.Log.Info("Veth example initialization done")
	return nil
}

// AfterInit initializes example plugin.
func (p *VethExamplePlugin) AfterInit() error {
	// Apply initial Linux/VPP configuration.
	p.putInitialData()

	// Schedule reconfiguration.
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())
	p.wg.Add(1)
	go p.putModifiedData(ctx, *timeout)

	return nil
}

// Close cleans up the resources.
func (p *VethExamplePlugin) Close() error {
	p.cancel()
	p.wg.Wait()

	p.Log.Info("Closed Veth plugin")
	return nil
}

// Configure initial data
func (p *VethExamplePlugin) putInitialData() {
	p.Log.Infof("Applying initial configuration")
	err := localclient2.DataResyncRequest(PluginName).
		LinuxInterface(initialVeth11()).
		LinuxInterface(initialVeth12()).
		VppInterface(afPacket1()).
		BD(bridgeDomain()).
		Send().ReceiveReply()
	if err != nil {
		p.Log.Errorf("Initial configuration failed: %v", err)
	} else {
		p.Log.Info("Initial configuration successful")
	}
}

// Configure modified data
func (p *VethExamplePlugin) putModifiedData(ctx context.Context, timeout int) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		p.Log.Infof("Applying modified configuration")
		// Simulate configuration change after timeout
		err := localclient2.DataChangeRequest(PluginName).
			Put().
			LinuxInterface(modifiedVeth11()).
			LinuxInterface(modifiedVeth12()).
			LinuxInterface(veth21()).
			LinuxInterface(veth22()).
			VppInterface(afPacket2()).
			Send().ReceiveReply()
		if err != nil {
			p.Log.Errorf("Modified configuration failed: %v", err)
		} else {
			p.Log.Info("Modified configuration successful")
		}
	case <-ctx.Done():
		// Cancel the scheduled re-configuration.
		p.Log.Info("Modification of configuration canceled")
	}
	p.wg.Done()
}

/* Example Data */

func initialVeth11() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:    "veth11",
		Type:    linux_intf.Interface_VETH,
		Enabled: true,
		Link: &linux_intf.Interface_Veth{
			Veth: &linux_intf.VethLink{PeerIfName: "veth12"},
		},
	}
}

func modifiedVeth11() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:    "veth11",
		Type:    linux_intf.Interface_VETH,
		Enabled: true,
		Link: &linux_intf.Interface_Veth{
			Veth: &linux_intf.VethLink{PeerIfName: "veth12"},
		},
		Mtu: 1000,
	}
}

func initialVeth12() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:    "veth12",
		Type:    linux_intf.Interface_VETH,
		Enabled: true,
		Link: &linux_intf.Interface_Veth{
			Veth: &linux_intf.VethLink{PeerIfName: "veth11"},
		},
	}
}

func modifiedVeth12() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:    "veth12",
		Type:    linux_intf.Interface_VETH,
		Enabled: true,
		Link: &linux_intf.Interface_Veth{
			Veth: &linux_intf.VethLink{PeerIfName: "veth11"},
		},
		IpAddresses: []string{"10.0.0.1/24"},
		PhysAddress: "D2:74:8C:12:67:D2",
		Namespace: &linux_namespace.NetNamespace{
			Reference: "ns1",
			Type:      linux_namespace.NetNamespace_NSID,
		},
	}
}

func veth21() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:    "veth21",
		Type:    linux_intf.Interface_VETH,
		Enabled: true,
		Link: &linux_intf.Interface_Veth{
			Veth: &linux_intf.VethLink{PeerIfName: "veth22"},
		},
	}
}

func veth22() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:    "veth22",
		Type:    linux_intf.Interface_VETH,
		Enabled: true,
		Link: &linux_intf.Interface_Veth{
			Veth: &linux_intf.VethLink{PeerIfName: "veth21"},
		},
		IpAddresses: []string{"10.0.0.2/24"},
		PhysAddress: "92:C7:42:67:AB:CD",
		Namespace: &linux_namespace.NetNamespace{
			Reference: "ns2",
			Type:      linux_namespace.NetNamespace_NSID,
		},
	}
}

func afPacket1() *vpp_intf.Interface {
	return &vpp_intf.Interface{
		Name:    "afpacket1",
		Type:    vpp_intf.Interface_AF_PACKET,
		Enabled: true,
		Link: &vpp_intf.Interface_Afpacket{
			Afpacket: &vpp_intf.AfpacketLink{
				HostIfName: "veth11",
			},
		},
	}
}

func afPacket2() *vpp_intf.Interface {
	return &vpp_intf.Interface{
		Name:    "afpacket2",
		Type:    vpp_intf.Interface_AF_PACKET,
		Enabled: true,
		Link: &vpp_intf.Interface_Afpacket{
			Afpacket: &vpp_intf.AfpacketLink{
				HostIfName: "veth21",
			},
		},
	}
}

func bridgeDomain() *vpp_l2.BridgeDomain {
	return &vpp_l2.BridgeDomain{
		Name:                "br1",
		Flood:               true,
		UnknownUnicastFlood: true,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0, /* means disable aging */
		Interfaces: []*vpp_l2.BridgeDomain_Interface{
			{
				Name:                    "afpacket1",
				BridgedVirtualInterface: false,
			}, {
				Name:                    "afpacket2",
				BridgedVirtualInterface: false,
			},
		},
	}
}
