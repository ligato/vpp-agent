// Copyright (c) 2019 Cisco and/or its affiliates.
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
	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

var (
	address       = flag.String("address", "127.0.0.1:9111", "address of GRPC server")
	socketType    = flag.String("socket-type", "tcp", "socket type [tcp, tcp4, tcp6, unix, unixpacket]")
	numClients    = flag.Int("clients", 1, "number of concurrent grpc clients")
	numTunnels    = flag.Int("tunnels", 100, "number of tunnels to stress per client")
	numPerRequest = flag.Int("numperreq", 10, "number of tunnels/routes per grpc request")
	withIPs       = flag.Bool("with-ips", false, "configure IP address for each tunnel on memif at the end")
	debug         = flag.Bool("debug", false, "turn on debug dump")
	timeout       = flag.Uint("timeout", 300, "timeout for requests (in seconds)")

	dialTimeout = time.Second * 3
	reqTimeout  = time.Second * 300
)

func main() {
	if *debug {
		logging.DefaultLogger.SetLevel(logging.DebugLevel)
	}

	quit := make(chan struct{})

	ep := NewGRPCStressPlugin()

	a := agent.NewAgent(
		agent.AllPlugins(ep),
		agent.QuitOnClose(quit),
	)

	if err := a.Start(); err != nil {
		log.Fatalln(err)
	}

	ep.setupInitial()
	ep.runAllClients()

	if err := a.Stop(); err != nil {
		log.Fatalln(err)
	}
}

// GRPCStressPlugin makes use of the remoteclient to locally CRUD ipsec tunnels and routes.
type GRPCStressPlugin struct {
	infra.PluginName
	Log *logrus.Logger

	conns []*grpc.ClientConn

	wg sync.WaitGroup
}

func NewGRPCStressPlugin() *GRPCStressPlugin {
	p := &GRPCStressPlugin{}
	p.SetName("grpc-stress-test-client")
	p.Log = logrus.New()
	p.Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
	})
	return p
}

func (p *GRPCStressPlugin) Init() error {
	return nil
}
func (p *GRPCStressPlugin) Close() error {
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

func (p *GRPCStressPlugin) setupInitial() {
	conn, err := grpc.Dial("unix",
		grpc.WithInsecure(),
		grpc.WithDialer(dialer(*socketType, *address, dialTimeout)),
	)
	if err != nil {
		log.Fatal(err)
	}

	reqTimeout = time.Second * time.Duration(*timeout)

	client := configurator.NewConfiguratorServiceClient(conn)

	// create a conn/client to create the red/black interfaces
	// that each tunnel will reference
	p.runGRPCCreateRedBlackMemifs(client)
}

// create the initial red and black memif's that kiknos uses ...
// ipsec wil ref the red ONLY i guess we dont need the black yet
// but maybe there will be a reason
func (p *GRPCStressPlugin) runGRPCCreateRedBlackMemifs(client configurator.ConfiguratorServiceClient) {
	p.Log.Infof("Configuring memif interfaces..")

	memIFRed := &interfaces.Interface{
		Name:        "red",
		Type:        interfaces.Interface_MEMIF,
		IpAddresses: []string{"100.0.0.1/24"},
		Mtu:         9200,
		Enabled:     true,
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             1,
				Master:         false,
				SocketFilename: "/var/run/memif_k8s-master.sock",
			},
		},
	}
	memIFBlack := &interfaces.Interface{
		Name:        "black",
		Type:        interfaces.Interface_MEMIF,
		IpAddresses: []string{"192.168.20.1/24"},
		Mtu:         9200,
		Enabled:     true,
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             2,
				Master:         false,
				SocketFilename: "/var/run/memif_k8s-master.sock",
			},
		},
	}
	ifaces := []*interfaces.Interface{memIFRed, memIFBlack}

	ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
	_, err := client.Update(ctx, &configurator.UpdateRequest{
		Update: &configurator.Config{
			VppConfig: &vpp.ConfigData{
				Interfaces: ifaces,
			},
		},
		FullResync: true,
	})
	if err != nil {
		log.Fatalln(err)
	}
	cancel()

	if *debug {
		p.Log.Infof("Requesting get..")

		cfg, err := client.Get(context.Background(), &configurator.GetRequest{})
		if err != nil {
			log.Fatalln(err)
		}
		out, _ := (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(cfg)
		fmt.Printf("Config:\n %+v\n", out)
	}

}

func (p *GRPCStressPlugin) runAllClients() {
	p.Log.Debugf("numTunnels: %d, numPerRequest: %d, numClients=%d",
		*numTunnels, *numPerRequest, *numClients)

	p.Log.Infof("Running for %d clients", *numClients)

	t := time.Now()

	p.wg.Add(*numClients)
	for i := 0; i < *numClients; i++ {
		// Set up connection to the server.
		conn, err := grpc.Dial("unix",
			grpc.WithInsecure(),
			grpc.WithDialer(dialer(*socketType, *address, dialTimeout)),
		)
		if err != nil {
			log.Fatal(err)
		}
		p.conns = append(p.conns, conn)
		client := configurator.NewConfiguratorServiceClient(p.conns[i])

		go p.runGRPCStressCreate(i, client, *numTunnels)
	}

	p.Log.Debugf("Waiting for clients..")
	p.wg.Wait()
	took := time.Since(t)
	perSec := float64(*numTunnels) / took.Seconds()

	p.Log.Infof("All clients done!")
	p.Log.Infof("----------------------------------------")
	p.Log.Infof(" -> Took: %.3fs", took.Seconds())
	p.Log.Infof(" -> Clients: %d", *numClients)
	p.Log.Infof(" -> Requests: %d", *numTunnels)
	p.Log.Infof(" -> PERFORMANCE: %.1f req/sec", perSec)
	p.Log.Infof("----------------------------------------")

	for i := 0; i < *numClients; i++ {
		if err := p.conns[i].Close(); err != nil {
			log.Fatal(err)
		}
	}

}

// runGRPCStressCreate creates 1 tunnel and 1 route ... emulating what strongswan does on a per remote warrior
func (p *GRPCStressPlugin) runGRPCStressCreate(id int, client configurator.ConfiguratorServiceClient, numTunnels int) {
	defer p.wg.Done()

	p.Log.Debugf("Creating %d tunnels/routes ... for client %d, ", numTunnels, id)

	startTime := time.Now()

	ips := []string{"10.0.0.1/24"}

	for tunNum := 0; tunNum < numTunnels; {
		if tunNum == numTunnels {
			break
		}

		var ifaces []*interfaces.Interface
		var routes []*vpp_l3.Route

		for req := 0; req < *numPerRequest; req++ {
			if tunNum == numTunnels {
				break
			}

			tunID := id*numTunnels + tunNum
			tunNum++

			ipsecTunnelName := fmt.Sprintf("ipsec-%d", tunID)

			ipPart := gen2octets(uint32(tunID))
			localIP := fmt.Sprintf("100.%s.1", ipPart)
			remoteIP := fmt.Sprintf("100.%s.254", ipPart)

			ips = append(ips, localIP+"/24")

			ipsecInfo := &interfaces.Interface_Ipsec{
				Ipsec: &interfaces.IPSecLink{
					LocalIp:         localIP,
					RemoteIp:        remoteIP,
					LocalSpi:        200000 + uint32(tunID),
					RemoteSpi:       100000 + uint32(tunID),
					CryptoAlg:       ipsec.CryptoAlg_AES_CBC_256,
					LocalCryptoKey:  "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
					RemoteCryptoKey: "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
					IntegAlg:        ipsec.IntegAlg_SHA_512_256,
					LocalIntegKey:   "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
					RemoteIntegKey:  "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
				},
			}
			ipsecTunnel := &interfaces.Interface{
				Name:    ipsecTunnelName,
				Type:    interfaces.Interface_IPSEC_TUNNEL,
				Enabled: true,
				Mtu:     9000,
				Unnumbered: &interfaces.Interface_Unnumbered{
					InterfaceWithIp: "red",
				},
				Link: ipsecInfo,
			}

			route := &vpp_l3.Route{
				DstNetwork:        "10." + gen3octets(uint32(tunID)) + "/32",
				NextHopAddr:       remoteIP,
				OutgoingInterface: ipsecTunnelName,
			}

			//p.Log.Infof("Creating %s ... client: %d, tunNum: %d", ipsecTunnelName, id, tunNum)

			ifaces = append(ifaces, ipsecTunnel)
			routes = append(routes, route)
		}

		//p.Log.Infof("Creating %d ifaces & %d routes", len(ifaces), len(routes))

		_, err := client.Update(context.Background(), &configurator.UpdateRequest{
			Update: &configurator.Config{
				VppConfig: &vpp.ConfigData{
					Interfaces: ifaces,
					Routes:     routes,
				},
			},
		})
		if err != nil {
			log.Fatalf("Error creating tun/route: id/tun=%d/%d, err: %s", id, tunNum, err)
		}
	}

	if *withIPs {
		p.Log.Infof("updating %d ip addresses on memif", len(ips))

		memIFRed := &interfaces.Interface{
			Name:        "red",
			Type:        interfaces.Interface_MEMIF,
			IpAddresses: ips,
			Mtu:         9000,
			Enabled:     true,
			Link: &interfaces.Interface_Memif{
				Memif: &interfaces.MemifLink{
					Id:             1,
					Master:         false,
					SocketFilename: "/var/run/memif_k8s-master.sock",
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		_, err := client.Update(ctx, &configurator.UpdateRequest{
			Update: &configurator.Config{
				VppConfig: &vpp.ConfigData{
					Interfaces: []*interfaces.Interface{memIFRed},
				},
			},
		})
		cancel()
		if err != nil {
			log.Fatalln(err)
		}
	}

	endTime := time.Now()

	p.Log.Infof("Client #%d done, %d tunnels took %s",
		id, numTunnels, endTime.Sub(startTime).Round(time.Millisecond))

	if *debug {
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

}

func gen3octets(num uint32) string {
	return fmt.Sprintf("%d.%d.%d", (num>>16)&0xFF, (num>>8)&0xFF, (num)&0xFF)
}

func gen2octets(num uint32) string {
	return fmt.Sprintf("%d.%d", (num>>8)&0xFF, (num)&0xFF)
}
