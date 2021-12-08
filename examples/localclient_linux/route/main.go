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
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
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
	timeout = flag.Int("timeout", 5, "Timeout between applying of initial and modified configuration in seconds")
)

/* Vpp-agent Init and Close*/

// Start Agent plugins selected for this example.
func main() {
	// Set inter-dependency between VPP & Linux plugins
	vpp_ifplugin.DefaultPlugin.LinuxIfPlugin = &linux_ifplugin.DefaultPlugin
	vpp_ifplugin.DefaultPlugin.NsPlugin = &linux_nsplugin.DefaultPlugin
	linux_ifplugin.DefaultPlugin.VppIfPlugin = &vpp_ifplugin.DefaultPlugin

	// Init close channel to stop the example.
	exampleFinished := make(chan struct{})
	go closeExample("Route example finished", exampleFinished)

	// Inject dependencies to example plugin
	ep := &RouteExamplePlugin{
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
}

// Stop the agent with desired info message.
func closeExample(message string, exampleFinished chan struct{}) {
	time.Sleep(time.Duration(*timeout*4) * time.Second)
	logrus.DefaultLogger().Info(message)
	close(exampleFinished)
}

/* Route Example */

// RouteExamplePlugin uses localclient to transport example route, tap and its linux part
// configuration to linuxplugin or VPP plugins
type RouteExamplePlugin struct {
	Log logging.Logger
	app.VPP
	app.Linux
	Orchestrator *orchestrator.Plugin

	wg             sync.WaitGroup
	cancelResync   context.CancelFunc
	cancelModified context.CancelFunc
}

// PluginName represents name of plugin.
const PluginName = "route-example"

// Init initializes example plugin.
func (p *RouteExamplePlugin) Init() error {
	// Logger
	p.Log = logrus.DefaultLogger()
	p.Log.SetLevel(logging.DebugLevel)
	p.Log.Info("Initializing route device example")

	// Flags
	flag.Parse()
	p.Log.Infof("Timeout between create and modify set to %d", *timeout)

	p.Log.Info("Route example initialization done")
	return nil
}

// AfterInit initializes example plugin.
func (p *RouteExamplePlugin) AfterInit() error {
	// apply initial Linux/VPP configuration
	p.putInitialData()

	// schedule route resync
	var ctx context.Context
	ctx, p.cancelResync = context.WithCancel(context.Background())
	p.wg.Add(1)
	go p.resync(ctx, *timeout)

	// schedule route reconfiguration
	ctx, p.cancelModified = context.WithCancel(context.Background())
	p.wg.Add(1)
	go p.putModifiedData(ctx, *timeout*2)

	return nil
}

// Close cleans up the resources.
func (p *RouteExamplePlugin) Close() error {
	p.cancelResync()
	p.cancelModified()
	p.wg.Wait()

	p.Log.Info("Closed route plugin")
	return nil
}

// String returns plugin name
func (p *RouteExamplePlugin) String() string {
	return PluginName
}

// Configure initial data
func (p *RouteExamplePlugin) putInitialData() {
	p.Log.Infof("Applying initial configuration")
	err := localclient.DataResyncRequest(PluginName).
		VppInterface(tap()).
		VppInterface(tapIPv6()).
		LinuxInterface(linuxTap()).
		LinuxInterface(linuxTapIpv6()).
		LinuxRoute(route()).
		LinuxRoute(routeIPv6()).
		Send().ReceiveReply()
	if err != nil {
		p.Log.Errorf("Initial configuration failed: %v", err)
	} else {
		p.Log.Info("Initial configuration successful")
	}
}

// Configure modified data
// This step serves as a route resync check. After the resync, IPv4 route
// is expected to be re-created (metric change) while the IPv6 should be
// updated because of modified scope
func (p *RouteExamplePlugin) resync(ctx context.Context, timeout int) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		p.Log.Infof("Applying resync routes")
		// Simulate configuration change after timeout
		err := localclient.DataResyncRequest(PluginName).
			VppInterface(tap()).
			VppInterface(tapIPv6()).
			LinuxInterface(linuxTap()).
			LinuxInterface(linuxTapIpv6()).
			LinuxRoute(routeResync()).     // recreate
			LinuxRoute(routeResyncIPv6()). // update
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

// Calling modify should cause IPv4 route to update (modified scope) while
// the IPv6 is expected to be recreated (changed metric)
func (p *RouteExamplePlugin) putModifiedData(ctx context.Context, timeout int) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		p.Log.Infof("Applying modified configuration")
		// Simulate configuration change after timeout
		err := localclient.DataChangeRequest(PluginName).
			Put().
			VppInterface(tap()).
			VppInterface(tapIPv6()).
			LinuxInterface(linuxTap()).
			LinuxInterface(linuxTapIpv6()).
			LinuxRoute(routeModified()).     // update
			LinuxRoute(routeModifiedIPv6()). // recreate
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

func tap() *vpp_intf.Interface {
	return &vpp_intf.Interface{
		Name:        "tap1",
		Type:        vpp_intf.Interface_TAP,
		Enabled:     true,
		PhysAddress: "d5:bc:dc:12:e4:0e",
		IpAddresses: []string{
			"10.0.0.10/24",
		},
		Link: &vpp_intf.Interface_Tap{
			Tap: &vpp_intf.TapLink{
				Version: 2,
			},
		},
	}
}

func tapIPv6() *vpp_intf.Interface {
	return &vpp_intf.Interface{
		Name:        "tap2",
		Type:        vpp_intf.Interface_TAP,
		Enabled:     true,
		PhysAddress: "d5:bc:dc:12:e4:0d",
		IpAddresses: []string{
			"abc1::10/64",
		},
		Link: &vpp_intf.Interface_Tap{
			Tap: &vpp_intf.TapLink{
				Version: 2,
			},
		},
	}
}

func linuxTap() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:        "linux-tap1",
		Type:        linux_intf.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: "12:e4:0e:d5:bc:dc",
		IpAddresses: []string{
			"11.0.0.20/24",
		},
		Link: &linux_intf.Interface_Tap{
			Tap: &linux_intf.TapLink{
				VppTapIfName: "tap1",
			},
		},
	}
}

func linuxTapIpv6() *linux_intf.Interface {
	return &linux_intf.Interface{
		Name:        "linux-tap2",
		Type:        linux_intf.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: "44:bc:12:e4:0e:aa",
		IpAddresses: []string{
			"abc2::20/64",
		},
		Link: &linux_intf.Interface_Tap{
			Tap: &linux_intf.TapLink{
				VppTapIfName: "tap2",
			},
		},
	}
}

func route() *linux_l3.Route {
	return &linux_l3.Route{
		OutgoingInterface: "linux-tap1",
		DstNetwork:        "100.10.0.0/24",
		Scope:             linux_l3.Route_GLOBAL,
		Metric:            0,
	}
}

func routeResync() *linux_l3.Route {
	return &linux_l3.Route{
		OutgoingInterface: "linux-tap1",
		DstNetwork:        "100.10.0.0/24",
		Scope:             linux_l3.Route_GLOBAL,
		Metric:            500,
	}
}

func routeModified() *linux_l3.Route {
	return &linux_l3.Route{
		OutgoingInterface: "linux-tap1",
		DstNetwork:        "100.10.0.0/24",
		Scope:             linux_l3.Route_LINK,
		Metric:            500,
	}
}

func routeIPv6() *linux_l3.Route {
	return &linux_l3.Route{
		OutgoingInterface: "linux-tap2",
		DstNetwork:        "aaa1::/64",
		Scope:             linux_l3.Route_GLOBAL,
		Metric:            0,
	}
}

func routeResyncIPv6() *linux_l3.Route {
	return &linux_l3.Route{
		OutgoingInterface: "linux-tap2",
		DstNetwork:        "aaa1::/64",
		Scope:             linux_l3.Route_LINK,
		Metric:            1024,
	}
}

func routeModifiedIPv6() *linux_l3.Route {
	return &linux_l3.Route{
		OutgoingInterface: "linux-tap2",
		DstNetwork:        "aaa1::/64",
		Scope:             linux_l3.Route_LINK,
		Metric:            500,
	}
}
