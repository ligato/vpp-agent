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
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/namsral/flag"
	"google.golang.org/grpc"

	"github.com/ligato/vpp-agent/api"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/client/remoteclient"
)

var (
	address    = flag.String("address", "172.17.0.2:9111", "address of GRPC server")
	socketType = flag.String("socket-type", "tcp", "socket type [tcp, tcp4, tcp6, unix, unixpacket]")

	dialTimeout = time.Second * 2
)

var exampleFinished = make(chan struct{})

func main() {
	// Init close channel to stop the example.

	// Inject dependencies to example plugin
	ep := &ExamplePlugin{}

	// Start Agent
	a := agent.NewAgent(
		agent.AllPlugins(ep),
		agent.QuitOnClose(exampleFinished),
	)
	if err := a.Run(); err != nil {
		log.Fatal()
	}

	// End when the localhost example is finished.
	//go closeExample("localhost example finished", exampleFinished)

}

// Stop the agent with desired info message.
/*func closeExample(message string, exampleFinished chan struct{}) {
	time.Sleep(25 * time.Second)
	logrus.DefaultLogger().Info(message)
	close(exampleFinished)
}*/

/******************
 * Example plugin *
 ******************/

// ExamplePlugin demonstrates the use of the remoteclient to locally transport example configuration into the default VPP plugins.
type ExamplePlugin struct {
	conn *grpc.ClientConn

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// String returns plugin name
func (plugin *ExamplePlugin) String() string {
	return "grpc-config-example"
}

// Init initializes example plugin.
func (plugin *ExamplePlugin) Init() (err error) {
	switch *socketType {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
	default:
		return fmt.Errorf("unknown gRPC socket type: %s", socketType)
	}

	// Set up connection to the server.
	plugin.conn, err = grpc.Dial("unix",
		grpc.WithInsecure(),
		grpc.WithDialer(dialer(*socketType, *address, dialTimeout)),
	)
	if err != nil {
		return err
	}

	go my(plugin.conn)

	// Apply initial VPP configuration.
	//plugin.resyncVPP()

	// Schedule reconfiguration.
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())
	_ = ctx
	/*plugin.wg.Add(1)
	go plugin.reconfigureVPP(ctx)*/

	logrus.DefaultLogger().Info("Initialization of the example plugin has completed")
	return nil
}

func my(conn *grpc.ClientConn) {
	time.Sleep(time.Second)

	c := remoteclient.NewClientGRPC(api.NewSyncServiceClient(conn))

	req := c.ResyncRequest()
	req.Put(memif1, memif2)
	if err := req.Send(context.Background()); err != nil {
		log.Fatalln(err)
	}

	time.Sleep(time.Second * 5)

	req2 := c.ChangeRequest()
	req2.Delete(memif1)
	if err := req2.Send(context.Background()); err != nil {
		log.Fatalln(err)
	}

	time.Sleep(time.Second)
	close(exampleFinished)
}

var memif1 = &interfaces.Interface{
	Name:        "memif1",
	Enabled:     true,
	IpAddresses: []string{"3.3.0.1/16"},
	Type:        interfaces.Interface_MEMIF,
	Link: &interfaces.Interface_Memif{
		Memif: &interfaces.MemifLink{
			Id:             1,
			Master:         true,
			Secret:         "secret",
			SocketFilename: "/tmp/memif1.sock",
		},
	},
}
var memif2 = &interfaces.Interface{
	Name:        "memif0/10",
	Enabled:     true,
	Type:        interfaces.Interface_SUB_INTERFACE,
	IpAddresses: []string{"3.10.0.10/32"},
	Link: &interfaces.Interface_Sub{
		Sub: &interfaces.SubInterface{
			ParentName: "memif1",
			SubId:      10,
		},
	},
}

/*
func my(conn *grpc.ClientConn) {
	time.Sleep(time.Second)

	anyMemif1, err := types.MarshalAny(memif1)
	if err != nil {
		log.Fatal(err)
	}
	anyMemif2, err := types.MarshalAny(memif2)
	if err != nil {
		log.Fatal(err)
	}

	req := &api.ResyncRequest{}

	req.Models = append(req.Models, &api.Model{
		Key:   interfaces.InterfaceKey(memif1.Name),
		Value: anyMemif1,
	})
	req.Models = append(req.Models, &api.Model{
		Key:   interfaces.InterfaceKey(memif2.Name),
		Value: anyMemif2,
	})

	c := api.NewModelServiceClient(conn)

	t0 := time.Now()

	resp, err := c.Resync(context.Background(), req)
	if err != nil {
		//s := status.Convert(err)
		s, ok := status.FromError(err)
		if !ok {
			log.Fatal("not ok!!!!")
		}
		var details []string
		for _, d := range s.Details() {
			switch dd := d.(type) {
			case *rpc.DebugInfo:
				details = append(details, dd.Detail)
			default:
				details = append(details, fmt.Sprint(dd))
			}
		}
		log.Printf("gRPC status: CODE: %v MSG: %v DETAILS: %v (%#v)",
			s.Code(), s.Message(), details, s.Details())

		log.Fatalf("RESYNC FAILED: %v", err)
	}

	log.Printf("Resync took %s, response:\n%+v",
		time.Since(t0).Round(time.Millisecond), *resp)
}
*/
// Close cleans up the resources.
func (plugin *ExamplePlugin) Close() error {
	logrus.DefaultLogger().Info("Closing example plugin")

	plugin.cancel()
	plugin.wg.Wait()

	if err := plugin.conn.Close(); err != nil {
		return err
	}

	return nil
}

// Dialer for unix domain socket
func dialer(socket, address string, timeoutVal time.Duration) func(string, time.Duration) (net.Conn, error) {
	return func(addr string, timeout time.Duration) (net.Conn, error) {
		// Pass values
		addr, timeout = address, timeoutVal
		// Dial with timeout
		return net.DialTimeout(socket, addr, timeoutVal)
	}
}

/*
// resyncVPP propagates snapshot of the whole initial configuration to VPP plugins.
func (plugin *ExamplePlugin) resyncVPP() {
	err := remoteclient.DataResyncRequestGRPC(rpc.NewDataResyncServiceClient(plugin.conn)).
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
func (plugin *ExamplePlugin) reconfigureVPP(ctx context.Context) {
	return
	_, dstNetAddr, err := net.ParseCIDR("192.168.2.1/32")
	if err != nil {
		return
	}
	nextHopAddr := net.ParseIP("192.168.1.1")

	select {
	case <-time.After(3 * time.Second):
		// Simulate configuration change exactly 15seconds after resync.
		err := remoteclient.DataChangeRequestGRPC(rpc.NewDataChangeServiceClient(plugin.conn)).
			Put().
			Interface(&memif1AsSlave).
			Interface(&memif2).
			Interface(&tap1Enabled).
			Interface(&loopback1WithAddr).
			ACL(&acl1).
			XConnect(&XConMemif1ToMemif2).
			BD(&BDLoopback1ToTap1).
			Delete().
			StaticRoute(0, dstNetAddr.String(), nextHopAddr.String()).
			Send().ReceiveReply()
		if err != nil {
			logrus.DefaultLogger().Errorf("Failed to reconfigure VPP: %v", err)
		} else {
			logrus.DefaultLogger().Info("Successfully reconfigured VPP")
		}
	case <-ctx.Done():
		// Cancel the scheduled re-configuration.
		logrus.DefaultLogger().Info("Planned VPP re-configuration was canceled")
	}
	plugin.wg.Done()
}
*/
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
/*
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
	// XConMemif1ToMemif2 defines xconnect between memifs.
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
				RuleName:  "rule1",
				AclAction: acl.AclAction_DENY,
				Match: &acl.AccessLists_Acl_Rule_Match{
					IpRule: &acl.AccessLists_Acl_Rule_Match_IpRule{
						Ip: &acl.AccessLists_Acl_Rule_Match_IpRule_Ip{
							DestinationNetwork: "10.1.1.0/24",
							SourceNetwork:      "10.1.2.0/24",
						},
						Tcp: &acl.AccessLists_Acl_Rule_Match_IpRule_Tcp{
							DestinationPortRange: &acl.AccessLists_Acl_Rule_Match_IpRule_PortRange{
								LowerPort: 50,
								UpperPort: 150,
							},
							SourcePortRange: &acl.AccessLists_Acl_Rule_Match_IpRule_PortRange{
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
		MacAge:              0,
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
*/
