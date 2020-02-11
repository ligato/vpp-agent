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

	"log"

	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/clientv2/vpp/localclient"
	"go.ligato.io/vpp-agent/v3/cmd/vpp-agent/app"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// init sets the default logging level.
func init() {
	logrus.DefaultLogger().SetOutput(os.Stdout)
	logrus.DefaultLogger().SetLevel(logging.DebugLevel)
}

/********
 * Main *
 ********/

// Start Agent plugins selected for this example.
func main() {
	// Init close channel to stop the example.
	exampleFinished := make(chan struct{})

	// Inject dependencies to example plugin
	ep := &ExamplePlugin{
		Log:          logging.DefaultLogger,
		VPP:          app.DefaultVPP(),
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
	time.Sleep(25 * time.Second)
	logrus.DefaultLogger().Info(message)
	close(exampleFinished)
}

/******************
 * Example plugin *
 ******************/

// ExamplePlugin demonstrates the use of the localclient to locally transport example configuration into the default VPP plugins.
type ExamplePlugin struct {
	Log logging.Logger
	app.VPP
	Orchestrator *orchestrator.Plugin

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// PluginName represents name of plugin.
const PluginName = "plugin-example"

// Init initializes example plugin.
func (p *ExamplePlugin) Init() error {
	// Logger
	p.Log = logrus.DefaultLogger()
	p.Log.SetLevel(logging.DebugLevel)
	p.Log.Info("Initializing VPP example")

	logrus.DefaultLogger().Info("Initialization of the example plugin has completed")
	return nil
}

// AfterInit initializes example plugin.
func (p *ExamplePlugin) AfterInit() error {
	// Apply initial VPP configuration.
	p.resyncVPP()

	// Schedule reconfiguration.
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())
	p.wg.Add(1)
	go p.reconfigureVPP(ctx)

	return nil
}

// Close cleans up the resources.
func (p *ExamplePlugin) Close() error {
	p.cancel()
	p.wg.Wait()

	logrus.DefaultLogger().Info("Closed example plugin")
	return nil
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return PluginName
}

// resyncVPP propagates snapshot of the whole initial configuration to VPP plugins.
func (p *ExamplePlugin) resyncVPP() {
	err := localclient.DataResyncRequest(PluginName).
		Interface(&memif1AsMaster).
		Interface(&tap1Disabled).
		Interface(&loopback1).
		StaticRoute(&routeThroughMemif1).
		Send().ReceiveReply()
	if err != nil {
		logrus.DefaultLogger().Errorf("Failed to apply initial VPP configuration: %v", err)
	} else {
		logrus.DefaultLogger().Info("Successfully applied initial VPP configuration")
	}
}

// reconfigureVPP simulates a set of changes in the configuration related to VPP plugins.
func (p *ExamplePlugin) reconfigureVPP(ctx context.Context) {
	select {
	case <-time.After(15 * time.Second):
		// Simulate configuration change exactly 15seconds after resync.
		err := localclient.DataChangeRequest(PluginName).
			Put().
			Interface(&memif1AsSlave).     /* turn memif1 into slave, remove the IP address */
			Interface(&memif2).            /* newly added memif interface */
			Interface(&tap1Enabled).       /* enable tap1 interface */
			Interface(&loopback1WithAddr). /* assign IP address to loopback1 interface */
			ACL(&acl1).                    /* declare ACL for the traffic leaving tap1 interface */
			XConnect(&XConMemif1ToMemif2). /* xconnect memif interfaces */
			BD(&BDLoopback1ToTap1).        /* put loopback and tap1 into the same bridge domain */
			Delete().
			StaticRoute("", 0, "192.168.2.1/32", "192.168.1.1"). /* remove the route going through memif1 */
			Send().ReceiveReply()
		if err != nil {
			logrus.DefaultLogger().Errorf("Failed to reconfigure VPP: %v", err)
		} else {
			logrus.DefaultLogger().Info("Successfully reconfigured VPP")
		}
	case <-ctx.Done():
		// cancel the scheduled re-configuration
		logrus.DefaultLogger().Info("Planned VPP re-configuration was canceled")
	}
	p.wg.Done()
}

/*************************
 * Example plugin config *
 *************************/

/*****************************************************
 * After Resync                                      *
 *                                                   *
 *  +---------------------------------------------+  *
 *  |                                             |  *
 *  +-----------+           +---------------------+  *
 *  | tap1      |           |  memif1             |  *
 *  | DISABLED  |      +--> |  MASTER             |  *
 *  +-----------+      |    |  IP: 192.168.1.1/24 |  *
 *  |                  |    +---------------------+  *
 *  |  +-----------+   |                          |  *
 *  |  | loopback1 |   +                          |  *
 *  |  +-----------+   route for 192.168.2.1      |  *
 *  |                                             |  *
 *  +---------------------------------------------+  *
 *                                                   *
 *****************************************************/

/********************************************************
 * After Data Change Request                            *
 *                                                      *
 *  +------------------------------------------------+  *
 *  |                                                |  *
 *  +---------+ +------+                  +----------+  *
 *  | tap1    |-| acl1 |-+         +------| memif1   |  *
 *  | ENABLED | +------+ |         |      | SLAVE    |  *
 *  +---------+          |         |      +----------+  *
 *  |                  Bridge   xconnect             |  *
 *  |                  domain      |      +----------+  *
 *  |                    |         |      | memif2   |  *
 *  |  +------------+    |         +------| SLAVE    |  *
 *  |  | loopback1  |----+                +----------|  *
 *  |  +------------+                                |  *
 *  |                                                |  *
 *  +------------------------------------------------+  *
 *                                                      *
 ********************************************************/

var (
	// memif1AsMaster is an example of a memory interface configuration. (Master=true, with IPv4 address).
	memif1AsMaster = interfaces.Interface{
		Name:        "memif1",
		Type:        interfaces.Interface_MEMIF,
		Enabled:     true,
		Mtu:         1500,
		IpAddresses: []string{"192.168.1.1/24"},
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             1,
				Master:         true,
				SocketFilename: "/tmp/memif1.sock",
			},
		},
	}

	// memif1AsSlave is the original memif1 turned into slave and stripped of the IP address.
	memif1AsSlave = interfaces.Interface{
		Name:    "memif1",
		Type:    interfaces.Interface_MEMIF,
		Enabled: true,
		Mtu:     1500,
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             1,
				Master:         false,
				SocketFilename: "/tmp/memif1.sock",
			},
		},
	}

	// Memif2 is a slave memif without IP address and to be xconnected with memif1.
	memif2 = interfaces.Interface{
		Name:    "memif2",
		Type:    interfaces.Interface_MEMIF,
		Enabled: true,
		Mtu:     1500,
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             2,
				Master:         false,
				SocketFilename: "/tmp/memif2.sock",
			},
		},
	}

	// XConMemif1ToMemif2 defines xconnect between memifs.
	XConMemif1ToMemif2 = l2.XConnectPair{
		ReceiveInterface:  memif1AsSlave.Name,
		TransmitInterface: memif2.Name,
	}

	// tap1Disabled is a disabled tap interface.
	tap1Disabled = interfaces.Interface{
		Name:    "tap1",
		Type:    interfaces.Interface_TAP,
		Enabled: false,
		Link: &interfaces.Interface_Tap{
			Tap: &interfaces.TapLink{
				Version:    2,
				HostIfName: "linux-tap1",
			},
		},
		Mtu: 1500,
	}

	// tap1Enabled is an enabled tap1 interface.
	tap1Enabled = interfaces.Interface{
		Name:    "tap1",
		Type:    interfaces.Interface_TAP,
		Enabled: true,
		Link: &interfaces.Interface_Tap{
			Tap: &interfaces.TapLink{
				Version:    2,
				HostIfName: "linux-tap1",
			},
		},
		Mtu: 1500,
	}

	acl1 = acl.ACL{
		Name: "acl1",
		Rules: []*acl.ACL_Rule{
			{
				Action: acl.ACL_Rule_DENY,
				IpRule: &acl.ACL_Rule_IpRule{
					Ip: &acl.ACL_Rule_IpRule_Ip{
						DestinationNetwork: "10.1.1.0/24",
						SourceNetwork:      "10.1.2.0/24",
					},
					Tcp: &acl.ACL_Rule_IpRule_Tcp{
						DestinationPortRange: &acl.ACL_Rule_IpRule_PortRange{
							LowerPort: 50,
							UpperPort: 150,
						},
						SourcePortRange: &acl.ACL_Rule_IpRule_PortRange{
							LowerPort: 1000,
							UpperPort: 2000,
						},
					},
				},
			},
		},
		Interfaces: &acl.ACL_Interfaces{
			Egress: []string{"tap1"},
		},
	}

	// loopback1 is an example of a loopback interface configuration (without IP address assigned).
	loopback1 = interfaces.Interface{
		Name:    "loopback1",
		Type:    interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled: true,
		Mtu:     1500,
	}

	// loopback1WithAddr extends loopback1 definition with an IP address.
	loopback1WithAddr = interfaces.Interface{
		Name:        "loopback1",
		Type:        interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		Mtu:         1500,
		IpAddresses: []string{"10.0.0.1/24"},
	}

	// BDLoopback1ToTap1 is a bridge domain with tap1 and loopback1 interfaces in it.
	// Loopback is set to be BVI.
	BDLoopback1ToTap1 = l2.BridgeDomain{
		Name:                "br1",
		Flood:               false,
		UnknownUnicastFlood: false,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0, /* means disable aging */
		Interfaces: []*l2.BridgeDomain_Interface{
			{
				Name:                    "loopback1",
				BridgedVirtualInterface: true,
			}, {
				Name:                    "tap1",
				BridgedVirtualInterface: false,
			},
		},
	}

	// routeThroughMemif1 is an example route configuration with memif1 being the next hop.
	routeThroughMemif1 = l3.Route{
		VrfId:       0,
		DstNetwork:  "192.168.2.1/32",
		NextHopAddr: "192.168.1.1", // Memif1AsMaster
		Weight:      5,
	}
)
