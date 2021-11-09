// Copyright (c) 2021 Cisco and/or its affiliates.
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

	"github.com/namsral/flag"
	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/clientv2/linux/localclient"
	"go.ligato.io/vpp-agent/v3/cmd/vpp-agent/app"
	linux_ifplugin "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	linux_nsplugin "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	vpp_ifplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	linux_intf "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	vpp_intf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

var (
	timeout = flag.Int("timeout", 20, "Timeout between applying of initial and modified configuration in seconds")
)

/* Confgiuration */

// Example configures two VRF devices. Then two TAP interfaces on the vpp with other ends in the default namespace.
// Both Linux taps are set with IP address
/**********************************************
 * Initial Data                               *
 *                                            *
 *  +--------------------------------------+  *
 *  |                                      |  *
 *  |   +-------+             +-------+    |  *
 *  |   | tap1  |             | tap2  |    |  *
 *  |   +---+---+             +---+---+    |  *
 *  |       |                     |        |  *
 *  +-------+------------------------------+  *
 *          |                     |           *
 *  +-------+---------+   +-------+--------+  *
 *  | linux-tap1      |   | linux-tap1     |  *
 *  | IP: 10.0.0.2/24 |   | IP: 20.0.0.2/24|  *
 *  +-----------------+   +----------------+  *
 *  VRF-dev-1			  VRF-dev-2           *
 *                                            *
 **********************************************/

// Next step switches Linux TAP VRFs
/**********************************************
 * Initial Data                               *
 *                                            *
 *  +--------------------------------------+  *
 *  |                                      |  *
 *  |   +-------+             +-------+    |  *
 *  |   | tap1  |             | tap2  |    |  *
 *  |   +---+---+             +---+---+    |  *
 *  |       |                     |        |  *
 *  +-------+------------------------------+  *
 *          |                     |           *
 *  +-------+---------+   +-------+--------+  *
 *  | linux-tap1      |   | linux-tap1     |  *
 *  | IP: 10.0.0.2/24 |   | IP: 20.0.0.2/24|  *
 *  +-----------------+   +----------------+  *
 *  VRF-dev-2			  VRF-dev-1           *
 *                                            *
 **********************************************/

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
	ep := &VrfExamplePlugin{
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

	go closeExample("VRF localhost example finished", exampleFinished)
}

// Stop the agent with desired info message.
func closeExample(message string, exampleFinished chan struct{}) {
	time.Sleep(time.Duration(*timeout+5) * time.Second)
	logrus.DefaultLogger().Info(message)
	close(exampleFinished)
}

/* VRF Example */

// VrfExamplePlugin uses localclient to transport example vrf, tap and its linux part
// configuration to linuxplugin or VPP plugins
type VrfExamplePlugin struct {
	Log logging.Logger
	app.VPP
	app.Linux
	Orchestrator *orchestrator.Plugin

	wg             sync.WaitGroup
	cancelResync   context.CancelFunc
	cancelModified context.CancelFunc
}

// PluginName represents name of plugin.
const PluginName = "vrf-example"

// Init initializes example plugin.
func (p *VrfExamplePlugin) Init() error {
	// Logger
	p.Log = logrus.DefaultLogger()
	p.Log.SetLevel(logging.DebugLevel)
	p.Log.Info("Initializing VRF device example")

	// Flags
	flag.Parse()
	p.Log.Infof("Timeout between create and modify set to %d", *timeout)

	p.Log.Info("VRF example initialization done")
	return nil
}

// AfterInit initializes example plugin.
func (p *VrfExamplePlugin) AfterInit() error {
	// Apply initial Linux/VPP configuration
	p.putInitialData()

	// Schedule VRF resync
	var ctx context.Context
	ctx, p.cancelResync = context.WithCancel(context.Background())
	p.wg.Add(1)
	go p.resync(ctx, *timeout)

	// Schedule reconfiguration
	ctx, p.cancelModified = context.WithCancel(context.Background())
	p.wg.Add(1)
	go p.putModifiedData(ctx, *timeout*2)

	return nil
}

// Close cleans up the resources.
func (p *VrfExamplePlugin) Close() error {
	p.cancelResync()
	p.cancelModified()
	p.wg.Wait()

	p.Log.Info("Closed VRF plugin")
	return nil
}

// String returns plugin name
func (p *VrfExamplePlugin) String() string {
	return PluginName
}

// Configure initial data
func (p *VrfExamplePlugin) putInitialData() {
	p.Log.Infof("Applying initial configuration")
	err := localclient.DataResyncRequest(PluginName).
		LinuxInterface(vrf1()).
		LinuxInterface(vrf2()).
		VppInterface(tap1()).
		LinuxInterface(linuxTap1()).
		VppInterface(tap2()).
		LinuxInterface(linuxTap2()).
		Send().ReceiveReply()
	if err != nil {
		p.Log.Errorf("Initial configuration failed: %v", err)
	} else {
		p.Log.Info("Initial configuration successful")
	}
}

// Configure modified data
// This step serves as a MTU check for various kernels. After the resync, vrf1's
// MTU should be 65575 and vrf2's MTU 65536
func (p *VrfExamplePlugin) resync(ctx context.Context, timeout int) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		p.Log.Infof("Applying resync VRFs")
		// Simulate configuration change after timeout
		err := localclient.DataResyncRequest(PluginName).
			LinuxInterface(vrf1()).
			LinuxInterface(vrf2()).
			VppInterface(tap1()).
			LinuxInterface(linuxTap1()).
			VppInterface(tap2()).
			LinuxInterface(linuxTap2()).
			Send().ReceiveReply()
		if err != nil {
			p.Log.Errorf("Resync failed: %v", err)
		} else {
			p.Log.Info("Resync successful")
		}
	case <-ctx.Done():
		// Cancel the scheduled re-configuration.
		p.Log.Info("Resync of configuration canceled")
	}
	p.wg.Done()
}

// Configure modified data
func (p *VrfExamplePlugin) putModifiedData(ctx context.Context, timeout int) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		p.Log.Infof("Applying modified configuration")
		// Simulate configuration change after timeout
		err := localclient.DataChangeRequest(PluginName).
			Put().
			LinuxInterface(modifiedLinuxTap1()).
			LinuxInterface(modifiedLinuxTap2()).
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

func vrf1() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:    "vrf1",
		Type:    linux_intf.Interface_VRF_DEVICE,
		Enabled: true,
		Link: &linux_intf.Interface_VrfDev{
			VrfDev: &linux_intf.VrfDevLink{
				RoutingTable: 1,
			},
		},
	}
}

func vrf2() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:    "vrf2",
		Type:    linux_intf.Interface_VRF_DEVICE,
		Enabled: true,
		Mtu:     65536, // simulate old kernel's default
		Link: &linux_intf.Interface_VrfDev{
			VrfDev: &linux_intf.VrfDevLink{
				RoutingTable: 2,
			},
		},
	}
}

func tap1() *vpp_intf.Interface {
	return &vpp_intf.Interface{
		Name:        "tap1",
		Type:        vpp_intf.Interface_TAP,
		Enabled:     true,
		PhysAddress: "D5:BC:DC:12:E4:0E",
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

func linuxTap1() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:        "linux-tap1",
		Type:        linux_intf.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: "88:88:88:88:88:88",
		IpAddresses: []string{
			"10.0.0.2/24",
		},
		HostIfName:         "tap_to_vpp1",
		VrfMasterInterface: "vrf1",
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
		PhysAddress: "88:88:88:88:88:88",
		IpAddresses: []string{
			"10.0.0.2/24",
		},
		HostIfName:         "tap_to_vpp1",
		VrfMasterInterface: "vrf2",
		Link: &linux_intf.Interface_Tap{
			Tap: &linux_intf.TapLink{
				VppTapIfName: "tap1",
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

func linuxTap2() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:        "linux-tap2",
		Type:        linux_intf.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: "88:88:88:88:88:88",
		IpAddresses: []string{
			"20.0.0.2/24",
		},
		HostIfName:         "tap_to_vpp2",
		VrfMasterInterface: "vrf2",
		Link: &linux_intf.Interface_Tap{
			Tap: &linux_intf.TapLink{
				VppTapIfName: "tap2",
			},
		},
	}
}

func modifiedLinuxTap2() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:        "linux-tap2",
		Type:        linux_intf.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: "88:88:88:88:88:88",
		IpAddresses: []string{
			"20.0.0.2/24",
		},
		HostIfName:         "tap_to_vpp2",
		VrfMasterInterface: "vrf1",
		Link: &linux_intf.Interface_Tap{
			Tap: &linux_intf.TapLink{
				VppTapIfName: "tap2",
			},
		},
	}
}
