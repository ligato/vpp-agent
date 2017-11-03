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

	"net"

	"github.com/ligato/cn-infra/core"
	agent "github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins/remoteclient"
	"github.com/ligato/vpp-agent/flavors/rpc/model/vppsvc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"

	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
	"google.golang.org/grpc"
)

const defaultAddress = "localhost:9111"

var address = defaultAddress

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

	flavor := local.FlavorLocal{}

	flag.StringVar(&address, "address", defaultAddress, "address of GRPC server")

	// Example plugin
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{}}
	// Create new agent
	agentVar := agent.NewAgent(log.DefaultLogger(), 15*time.Second, append(flavor.Plugins(), examplePlugin)...)

	// End when the localhost example is finished
	go closeExample("localhost example finished", closeChannel)

	agent.EventLoopWithInterrupt(agentVar, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(25 * time.Second)
	log.DefaultLogger().Info(message)
	closeChannel <- struct{}{}
}

/******************
 * Example plugin *
 ******************/

// PluginID of example plugin
const PluginID core.PluginName = "example-plugin"

// ExamplePlugin demonstrates the use of the remoteclient to locally transport example configuration into the default VPP plugins.
type ExamplePlugin struct {
	wg     sync.WaitGroup
	cancel context.CancelFunc
	conn   *grpc.ClientConn
}

// Init initializes example plugin.
func (plugin *ExamplePlugin) Init() (err error) {
	// Set up a connection to the server.
	plugin.conn, err = grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}

	// apply initial VPP configuration
	plugin.resyncVPP()

	// schedule reconfiguration
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())
	plugin.wg.Add(1)
	go plugin.reconfigureVPP(ctx)

	log.DefaultLogger().Info("Initialization of the example plugin has completed")
	return nil
}

// Close cleans up the resources.
func (plugin *ExamplePlugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	err := safeclose.Close(plugin.conn)
	if err != nil {
		return err
	}

	log.DefaultLogger().Info("Closed example plugin")
	return nil
}

// resyncVPP propagates snapshot of the whole initial configuration to VPP plugins.
func (plugin *ExamplePlugin) resyncVPP() {
	err := remoteclient.DataResyncRequestGRPC(vppsvc.NewResyncConfigServiceClient(plugin.conn)).
		Interface(&memif1AsMaster).
		Interface(&tap1Disabled).
		Interface(&loopback1).
		StaticRoute(&routeThroughMemif1).
		Send().ReceiveReply()
	if err != nil {
		log.DefaultLogger().Errorf("Failed to apply initial VPP configuration: %v", err)
	} else {
		log.DefaultLogger().Info("Successfully applied initial VPP configuration")
	}
}

// reconfigureVPP simulates a set of changes in the configuration related to VPP plugins.
func (plugin *ExamplePlugin) reconfigureVPP(ctx context.Context) {
	_, dstNetAddr, err := net.ParseCIDR("192.168.2.1/32")
	if err != nil {
		return
	}
	nextHopAddr := net.ParseIP("192.168.1.1")

	select {
	case <-time.After(15 * time.Second):
		// simulate configuration change exactly 15seconds after resync
		err := remoteclient.DataChangeRequestGRPC(vppsvc.NewChangeConfigServiceClient(plugin.conn)).
			Put().
			Interface(&memif1AsSlave).     /* turn memif1 into slave, remove the IP address */
			Interface(&memif2).            /* newly added memif interface */
			Interface(&tap1Enabled).       /* enable tap1 interface */
			Interface(&loopback1WithAddr). /* assign IP address to loopback1 interface */
			ACL(&acl1).                    /* declare ACL for the traffic leaving tap1 interface */
			XConnect(&XConMemif1ToMemif2). /* xconnect memif interfaces */
			BD(&BDLoopback1ToTap1).        /* put loopback and tap1 into the same bridge domain */
			Delete().
			StaticRoute(0, dstNetAddr, nextHopAddr). /* remove the route going through memif1 */
			Send().ReceiveReply()
		if err != nil {
			log.DefaultLogger().Errorf("Failed to reconfigure VPP: %v", err)
		} else {
			log.DefaultLogger().Info("Successfully reconfigured VPP")
		}
	case <-ctx.Done():
		// cancel the scheduled re-configuration
		log.DefaultLogger().Info("Planned VPP re-configuration was canceled")
	}
	plugin.wg.Done()
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
	memif1AsMaster = interfaces.Interfaces_Interface{
		Name:    "memif1",
		Type:    interfaces.InterfaceType_MEMORY_INTERFACE,
		Enabled: true,
		Memif: &interfaces.Interfaces_Interface_Memif{
			Id:             1,
			Master:         true,
			SocketFilename: "/tmp/memif1.sock",
		},
		Mtu:         1500,
		IpAddresses: []string{"192.168.1.1/24"},
	}

	// memif1AsSlave is the original memif1 turned into slave and stripped of the IP address.
	memif1AsSlave = interfaces.Interfaces_Interface{
		Name:    "memif1",
		Type:    interfaces.InterfaceType_MEMORY_INTERFACE,
		Enabled: true,
		Memif: &interfaces.Interfaces_Interface_Memif{
			Id:             1,
			Master:         false,
			SocketFilename: "/tmp/memif1.sock",
		},
		Mtu: 1500,
	}

	// Memif2 is a slave memif without IP address and to be xconnected with memif1.
	memif2 = interfaces.Interfaces_Interface{
		Name:    "memif2",
		Type:    interfaces.InterfaceType_MEMORY_INTERFACE,
		Enabled: true,
		Memif: &interfaces.Interfaces_Interface_Memif{
			Id:             2,
			Master:         false,
			SocketFilename: "/tmp/memif2.sock",
		},
		Mtu: 1500,
	}
	// XConMemif1ToMemif2 defines xconnect between memifs
	XConMemif1ToMemif2 = l2.XConnectPairs_XConnectPair{
		ReceiveInterface:  memif1AsSlave.Name,
		TransmitInterface: memif2.Name,
	}

	// tap1Disabled is a disabled tap interface.
	tap1Disabled = interfaces.Interfaces_Interface{
		Name:    "tap1",
		Type:    interfaces.InterfaceType_TAP_INTERFACE,
		Enabled: false,
		Tap: &interfaces.Interfaces_Interface_Tap{
			HostIfName: "linux-tap1",
		},
		Mtu: 1500,
	}

	// tap1Enabled is an enabled tap1 interface.
	tap1Enabled = interfaces.Interfaces_Interface{
		Name:    "tap1",
		Type:    interfaces.InterfaceType_TAP_INTERFACE,
		Enabled: true,
		Tap: &interfaces.Interfaces_Interface_Tap{
			HostIfName: "linux-tap1",
		},
		Mtu: 1500,
	}

	acl1 = acl.AccessLists_Acl{
		AclName: "acl1",
		Rules: []*acl.AccessLists_Acl_Rule{
			{
				RuleName: "rule1",
				Actions: &acl.AccessLists_Acl_Rule_Actions{
					AclAction: acl.AclAction_DENY,
				},
				Matches: &acl.AccessLists_Acl_Rule_Matches{
					IpRule: &acl.AccessLists_Acl_Rule_Matches_IpRule{
						Ip: &acl.AccessLists_Acl_Rule_Matches_IpRule_Ip{
							DestinationNetwork: "10.1.1.0/24",
							SourceNetwork:      "10.1.2.0/24",
						},
						Tcp: &acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp{
							DestinationPortRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_DestinationPortRange{
								LowerPort: 50,
								UpperPort: 150,
							},
							SourcePortRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_SourcePortRange{
								LowerPort: 1000,
								UpperPort: 2000,
							},
						},
					},
				},
			},
		},
		Interfaces: &acl.AccessLists_Acl_Interfaces{
			Egress: []string{"tap1"},
		},
	}

	// loopback1 is an example of a loopback interface configuration (without IP address assigned).
	loopback1 = interfaces.Interfaces_Interface{
		Name:    "loopback1",
		Type:    interfaces.InterfaceType_SOFTWARE_LOOPBACK,
		Enabled: true,
		Mtu:     1500,
	}

	// loopback1WithAddr extends loopback1 definition with an IP address.
	loopback1WithAddr = interfaces.Interfaces_Interface{
		Name:        "loopback1",
		Type:        interfaces.InterfaceType_SOFTWARE_LOOPBACK,
		Enabled:     true,
		Mtu:         1500,
		IpAddresses: []string{"10.0.0.1/24"},
	}

	// BDLoopback1ToTap1 is a bridge domain with tap1 and loopback1 interfaces in it.
	// Loopback is set to be BVI.
	BDLoopback1ToTap1 = l2.BridgeDomains_BridgeDomain{
		Name:                "br1",
		Flood:               false,
		UnknownUnicastFlood: false,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0, /* means disable aging */
		Interfaces: []*l2.BridgeDomains_BridgeDomain_Interfaces{
			{
				Name: "loopback1",
				BridgedVirtualInterface: true,
			}, {
				Name: "tap1",
				BridgedVirtualInterface: false,
			},
		},
	}

	// routeThroughMemif1 is an example route configuration, with memif1 being the next hop.
	routeThroughMemif1 = l3.StaticRoutes_Route{
		Description: "Description",
		VrfId:       0,
		DstIpAddr:   "192.168.2.1/32",
		NextHopAddr: "192.168.1.1", // Memif1AsMaster
		Weight:      5,
	}
)
