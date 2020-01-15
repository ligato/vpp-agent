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
	"log"
	"sync"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/namsral/flag"

	"go.ligato.io/vpp-agent/v3/clientv2/linux/localclient"
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
 * Modified Data                                *
 *                                              *
 *  +----------------------------------------+  *
 *  |       +-- Bridge Domain --+            |  *
 *  |       |                   |            |  *
 *  | +-----+--------+    *------+-------+   |  *
 *  | | tap1         |    | (tap2)       |   |  *
 *  | | 10.0.0.11/24 |    | 20.0.0.11/24 |   |  *
 *  | +-----+--------+    +------+-------+   |  *
 *  |       |                   |            |  *
 *  +-------+-------------------+------------+  *
 *          |                   |               *
 *  +-------+----------+   +-----+------------+ *
 *  | linux-tap1       |   | linux-tap2       | *
 *  | IP: 10.0.0.12/24 |   | IP: 20.0.0.12\24 | *
 *  | Namespace: ns1   |   | Namespace: ns2   | *
 *  +------------------+   +------------------+ *
 *                                              *
 *                                              *
 ************************************************/

/* Vpp-agent Init and Close*/

// Start Agent plugins selected for this example.
func main() {
	// Set inter-dependency between VPP & Linux plugins
	vpp_ifplugin.DefaultPlugin.LinuxIfPlugin = &linux_ifplugin.DefaultPlugin
	vpp_ifplugin.DefaultPlugin.NsPlugin = &linux_nsplugin.DefaultPlugin
	linux_ifplugin.DefaultPlugin.VppIfPlugin = &vpp_ifplugin.DefaultPlugin

	// Init close channel to stop the example.
	exampleFinished := make(chan struct{})

	// Inject dependencies to example plugin
	ep := &TapExamplePlugin{
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
		log.Fatal(err)
	}

	go closeExample("localhost example finished", exampleFinished)
}

// Stop the agent with desired info message.
func closeExample(message string, exampleFinished chan struct{}) {
	time.Sleep(time.Duration(*timeout+5) * time.Second)
	logrus.DefaultLogger().Info(message)
	close(exampleFinished)
}

/* TAP Example */

// TapExamplePlugin uses localclient to transport example tap and its linux end
// configuration to linuxplugin or VPP plugins
type TapExamplePlugin struct {
	Log logging.Logger
	app.VPP
	app.Linux
	Orchestrator *orchestrator.Plugin

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// PluginName represents name of plugin.
const PluginName = "tap-example"

// Init initializes example plugin.
func (p *TapExamplePlugin) Init() error {
	// Logger
	p.Log = logrus.DefaultLogger()
	p.Log.SetLevel(logging.DebugLevel)
	p.Log.Info("Initializing Tap example")

	// Flags
	flag.Parse()
	p.Log.Infof("Timeout between create and modify set to %d", *timeout)

	p.Log.Info("Tap example initialization done")
	return nil
}

// AfterInit initializes example plugin.
func (p *TapExamplePlugin) AfterInit() error {
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
func (p *TapExamplePlugin) Close() error {
	p.cancel()
	p.wg.Wait()

	p.Log.Info("Closed Tap plugin")
	return nil
}

// String returns plugin name
func (p *TapExamplePlugin) String() string {
	return PluginName
}

// Configure initial data
func (p *TapExamplePlugin) putInitialData() {
	p.Log.Infof("Applying initial configuration")
	err := localclient.DataResyncRequest(PluginName).
		VppInterface(initialTap1()).
		LinuxInterface(initialLinuxTap1()).
		BD(bridgeDomain()).
		Send().ReceiveReply()
	if err != nil {
		p.Log.Errorf("Initial configuration failed: %v", err)
	} else {
		p.Log.Info("Initial configuration successful")
	}
}

// Configure modified data
func (p *TapExamplePlugin) putModifiedData(ctx context.Context, timeout int) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		p.Log.Infof("Applying modified configuration")
		// Simulate configuration change after timeout
		err := localclient.DataChangeRequest(PluginName).
			Put().
			VppInterface(modifiedTap1()).
			VppInterface(tap2()).
			LinuxInterface(modifiedLinuxTap1()).
			LinuxInterface(linuxTap2()).
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

func initialTap1() *vpp_intf.Interface {
	return &vpp_intf.Interface{
		Name:    "tap1",
		Type:    vpp_intf.Interface_TAP,
		Enabled: true,
		Link: &vpp_intf.Interface_Tap{
			Tap: &vpp_intf.TapLink{
				Version: 2,
			},
		},
	}
}

func modifiedTap1() *vpp_intf.Interface {
	return &vpp_intf.Interface{
		Name:        "tap1",
		Type:        vpp_intf.Interface_TAP,
		Enabled:     true,
		PhysAddress: "12:E4:0E:D5:BC:DC",
		IpAddresses: []string{
			"10.0.0.11/24",
		},
		Link: &vpp_intf.Interface_Tap{
			Tap: &vpp_intf.TapLink{
				Version: 2,
			},
		},
	}
}

func tap2() *vpp_intf.Interface {
	return &vpp_intf.Interface{
		Name:        "tap2",
		Type:        vpp_intf.Interface_TAP,
		Enabled:     true,
		PhysAddress: "D5:BC:DC:12:E4:0E",
		IpAddresses: []string{
			"20.0.0.11/24",
		},
		Link: &vpp_intf.Interface_Tap{
			Tap: &vpp_intf.TapLink{
				Version: 2,
			},
		},
	}
}

func initialLinuxTap1() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:        "linux-tap1",
		Type:        linux_intf.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: "88:88:88:88:88:88",
		IpAddresses: []string{
			"10.0.0.2/24",
		},
		HostIfName: "tap_to_vpp1",
		Link: &linux_intf.Interface_Tap{
			Tap: &linux_intf.TapLink{
				VppTapIfName: "tap1",
			},
		},
	}
}

func modifiedLinuxTap1() *linux_intf.Interface {
	return &linux_intf.Interface{

		Name:        "linux-tap1",
		Type:        linux_intf.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: "BC:FE:E9:5E:07:04",
		Namespace: &linux_namespace.NetNamespace{
			Reference: "ns1",
			Type:      linux_namespace.NetNamespace_NSID,
		},
		Mtu: 1500,
		IpAddresses: []string{
			"10.0.0.12/24",
		},
		HostIfName: "tap_to_vpp1",
		Link: &linux_intf.Interface_Tap{
			Tap: &linux_intf.TapLink{
				VppTapIfName: "tap1",
			},
		},
	}
}

func linuxTap2() *linux_intf.Interface {
	return &linux_intf.Interface{

		Name:        "linux-tap2",
		Type:        linux_intf.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: "5E:07:04:BC:FE:E9",
		Namespace: &linux_namespace.NetNamespace{
			Reference: "ns2",
			Type:      linux_namespace.NetNamespace_NSID,
		},
		Mtu: 1500,
		IpAddresses: []string{
			"20.0.0.12/24",
		},
		HostIfName: "tap_to_vpp2",
		Link: &linux_intf.Interface_Tap{
			Tap: &linux_intf.TapLink{
				VppTapIfName: "tap2",
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
				Name:                    "tap1",
				BridgedVirtualInterface: false,
			},
			{
				Name:                    "tap2",
				BridgedVirtualInterface: false,
			},
			{
				Name:                    "loop1",
				BridgedVirtualInterface: true,
			},
		},
	}
}
