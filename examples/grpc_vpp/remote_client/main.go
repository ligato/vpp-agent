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

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/namsral/flag"
	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

var (
	address    = flag.String("address", "172.17.0.2:9111", "address of GRPC server")
	socketType = flag.String("socket-type", "tcp", "socket type [tcp, tcp4, tcp6, unix, unixpacket]")

	dialTimeout = time.Second * 2
)

var exampleFinished = make(chan struct{})

func main() {
	ep := &ExamplePlugin{}
	ep.SetName("remote-client-example")
	ep.Setup()

	a := agent.NewAgent(
		agent.AllPlugins(ep),
		agent.QuitOnClose(exampleFinished),
	)
	if err := a.Run(); err != nil {
		log.Fatal()
	}
}

// ExamplePlugin demonstrates the use of the remoteclient to locally transport example configuration into the default VPP plugins.
type ExamplePlugin struct {
	infra.PluginDeps

	conn *grpc.ClientConn

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// Init initializes example plugin.
func (p *ExamplePlugin) Init() (err error) {
	// Set up connection to the server.
	p.conn, err = grpc.Dial("unix",
		grpc.WithInsecure(),
		grpc.WithDialer(dialer(*socketType, *address, dialTimeout)),
	)
	if err != nil {
		return err
	}

	client := configurator.NewConfiguratorServiceClient(p.conn)

	// Apply initial VPP configuration.
	go p.demonstrateClient(client)

	// Schedule reconfiguration.
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())
	_ = ctx
	/*plugin.wg.Add(1)
	go plugin.reconfigureVPP(ctx)*/

	go func() {
		time.Sleep(time.Second * 30)
		close(exampleFinished)
	}()

	return nil
}

// Close cleans up the resources.
func (p *ExamplePlugin) Close() error {
	logrus.DefaultLogger().Info("Closing example plugin")

	p.cancel()
	p.wg.Wait()

	if err := p.conn.Close(); err != nil {
		return err
	}

	return nil
}

// demonstrateClient propagates snapshot of the whole initial configuration to VPP plugins.
func (p *ExamplePlugin) demonstrateClient(client configurator.ConfiguratorServiceClient) {
	time.Sleep(time.Second * 2)
	p.Log.Infof("Requesting resync..")

	config := &configurator.Config{
		VppConfig: &vpp.ConfigData{
			Interfaces: []*interfaces.Interface{
				memif1,
			},
			IpscanNeighbor: ipScanNeigh,
			IpsecSas:       []*vpp_ipsec.SecurityAssociation{sa10},
			IpsecSpds:      []*vpp_ipsec.SecurityPolicyDatabase{spd1},
		},
		LinuxConfig: &linux.ConfigData{
			Interfaces: []*linux_interfaces.Interface{
				veth1, veth2,
			},
		},
	}
	_, err := client.Update(context.Background(), &configurator.UpdateRequest{
		Update:     config,
		FullResync: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	time.Sleep(time.Second * 5)
	p.Log.Infof("Requesting change..")

	ifaces := []*interfaces.Interface{memif1, memif2, afpacket}
	_, err = client.Update(context.Background(), &configurator.UpdateRequest{
		Update: &configurator.Config{
			VppConfig: &vpp.ConfigData{
				Interfaces: ifaces,
			},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	time.Sleep(time.Second * 5)
	p.Log.Infof("Requesting delete..")

	ifaces = []*interfaces.Interface{memif1}
	_, err = client.Delete(context.Background(), &configurator.DeleteRequest{
		Delete: &configurator.Config{
			VppConfig: &vpp.ConfigData{
				Interfaces: ifaces,
			},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	time.Sleep(time.Second * 5)
	p.Log.Infof("Requesting get..")

	cfg, err := client.Get(context.Background(), &configurator.GetRequest{})
	if err != nil {
		log.Fatalln(err)
	}
	out, _ := (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(cfg)
	fmt.Printf("Config:\n %+v\n", out)

	time.Sleep(time.Second * 5)
	p.Log.Infof("Requesting dump..")

	dump, err := client.Dump(context.Background(), &configurator.DumpRequest{})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Dump:\n %+v\n", proto.MarshalTextString(dump))
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

var (
	sa10 = &vpp.IPSecSA{
		Index:     "10",
		Spi:       1001,
		Protocol:  1,
		CryptoAlg: 1,
		CryptoKey: "4a506a794f574265564551694d653768",
		IntegAlg:  2,
		IntegKey:  "4339314b55523947594d6d3547666b45764e6a58",
	}
	spd1 = &vpp.IPSecSPD{
		Index: "1",
		PolicyEntries: []*vpp_ipsec.SecurityPolicyDatabase_PolicyEntry{
			{
				Priority:   100,
				IsOutbound: false,
				Action:     0,
				Protocol:   50,
				SaIndex:    "10",
			},
		},
	}
	memif1 = &vpp.Interface{
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
	memif2 = &vpp.Interface{
		Name:        "memif2",
		Enabled:     true,
		IpAddresses: []string{"4.3.0.1/16"},
		Type:        interfaces.Interface_MEMIF,
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             2,
				Master:         true,
				Secret:         "secret",
				SocketFilename: "/tmp/memif2.sock",
			},
		},
	}
	ipScanNeigh = &vpp.IPScanNeigh{
		Mode: vpp_l3.IPScanNeighbor_BOTH,
	}
	veth1 = &linux.Interface{
		Name:        "myVETH1",
		Type:        linux_interfaces.Interface_VETH,
		Enabled:     true,
		HostIfName:  "veth1",
		IpAddresses: []string{"10.10.3.1/24"},
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: "myVETH2",
			},
		},
	}
	veth2 = &linux.Interface{
		Name:       "myVETH2",
		Type:       linux_interfaces.Interface_VETH,
		Enabled:    true,
		HostIfName: "veth2",
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{
				PeerIfName: "myVETH1",
			},
		},
	}
	afpacket = &vpp.Interface{
		Name:        "myAFpacket",
		Type:        interfaces.Interface_AF_PACKET,
		Enabled:     true,
		PhysAddress: "a7:35:45:55:65:75",
		IpAddresses: []string{
			"10.20.30.40/24",
		},
		Mtu: 1800,
		Link: &interfaces.Interface_Afpacket{
			Afpacket: &interfaces.AfpacketLink{
				HostIfName: "veth2",
			},
		},
	}
)

/*
// demonstrateClient propagates snapshot of the whole initial configuration to VPP plugins.
func (plugin *ExamplePlugin) demonstrateClient() {
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
