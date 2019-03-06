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

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/namsral/flag"
	"google.golang.org/grpc"

	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/api/models/vpp"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	"github.com/ligato/vpp-agent/api/models/vpp/l3"
)

var (
	address       = flag.String("address", "127.0.0.1:9111", "address of GRPC server")
	socketType    = flag.String("socket-type", "tcp", "socket type [tcp, tcp4, tcp6, unix, unixpacket]")
	numTunnels    = flag.Int("tunnels", 1, "number of tunnels to stress per client")
	numPerRequest = flag.Int("numperreq", 1, "number of tunnels/routes per grpc request")
	debug         = flag.Bool("debug", false, "turn on debug dump")
	numClients    = flag.Int("clients", 1, "number of concurrent grpc clients")

	dialTimeout = time.Second * 2
)

var stressTestFinished = make(chan struct{})

func main() {

	ep := &GRPCStressPlugin{}
	ep.SetName("grpc-stress-test-client")
	ep.Setup()

	a := agent.NewAgent(
		agent.AllPlugins(ep),
		agent.QuitOnClose(stressTestFinished))
		//agent.StartTimeout(time.Second*1000000))
	if err := a.Run(); err != nil {
		log.Fatal()
	}
}

// GRPCStressPlugin makes use of the remoteclient to locally CRUD ipsec tunnels and routes.
type GRPCStressPlugin struct {
	infra.PluginDeps

	conns []*grpc.ClientConn

	wg     sync.WaitGroup
}

// Init initializes  plugin.
func (p *GRPCStressPlugin) Init() (err error) {

	log.Printf("numTunnels: %d, numPerRequest: %d, numClients=%d\n",
		*numTunnels, *numPerRequest, *numClients)

	// create a conn/client to create the red/black interfaces that each tunnel will reference
	conn, err := grpc.Dial("unix",
		grpc.WithInsecure(),
		grpc.WithDialer(dialer(*socketType, *address, dialTimeout)),
	)
	client := configurator.NewConfiguratorClient(conn)
	p.runGRPCCreateRedBlackMemifs(client)


	p.wg.Add(*numClients)

	for i := 0; i < *numClients; i++ {
		// Set up connection to the server.
		conn, err := grpc.Dial("unix",
			grpc.WithInsecure(),
			grpc.WithDialer(dialer(*socketType, *address, dialTimeout)),
		)

		if err != nil {
			return err
		}
		p.conns = append(p.conns, conn)
		client := configurator.NewConfiguratorClient(p.conns[i])
		go p.runGRPCStressCreate(i, client, *numTunnels)
	}

	log.Printf("GRPCStressPlugin: init done")

	return nil
}

// Close cleans up the resources.
func (p *GRPCStressPlugin) Close() error {
	logrus.DefaultLogger().Info("Closing example plugin")

	p.wg.Wait()

	for i := 0; i < *numClients; i++ {
		if err := p.conns[i].Close(); err != nil {
			return err
		}
	}

	logrus.DefaultLogger().Info("Closing example plugin: DONE")

	stressTestFinished <- struct{}{}

	return nil
}

// runGRPCStressCreate creates 1 tunnel and 1 route ... emulating what strongswan does on a per remote warrior
func (p *GRPCStressPlugin) runGRPCStressCreate(id int, client configurator.ConfiguratorClient, numTunnels int) {

	defer p.wg.Done()

	p.Log.Infof("Creating %d tunnels/routes ... for client %d, ", numTunnels, id)

	startTime := time.Now()

	for tunNum := 0; tunNum < numTunnels; {
		if tunNum == numTunnels {
			break
		}
		for req := 0; req < *numPerRequest; req++ {
			if tunNum == numTunnels {
				break
			}

			tunNum++

			tunID := id*numTunnels + tunNum

			ipsecInfo := &interfaces.Interface_Ipsec{
				Ipsec: &interfaces.IPSecLink{
					LocalIp:         "100.100.100.100",
					RemoteIp:        "20." + gen3octets(uint32(tunID)),
					LocalSpi:        uint32(tunID),
					RemoteSpi:       uint32(tunID),
					CryptoAlg:       ipsec.CryptoAlg_AES_CBC_256,
					LocalCryptoKey:  "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
					RemoteCryptoKey: "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
					IntegAlg:        ipsec.IntegAlg_SHA_512_256,
					LocalIntegKey:   "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
					RemoteIntegKey:  "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
				},
			}
			ipsecTunnelName := fmt.Sprintf("grpc-ipsec-%d", tunID)
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
			ifaces := []*interfaces.Interface{ipsecTunnel}

			route := &vpp_l3.Route{
				DstNetwork:        "30." + gen3octets(uint32(tunID)) + "/32",
				NextHopAddr:       "172.2.0.1",
				OutgoingInterface: ipsecTunnelName,
			}
			routes := []*vpp_l3.Route{route}

			p.Log.Infof("Creating %s ... client: %d, tunNum: %d", ipsecTunnelName, id, tunNum)

			_, err := client.Update(context.Background(), &configurator.UpdateRequest{
				Update: &configurator.Config{
					VppConfig: &vpp.ConfigData{
						Interfaces: ifaces,
						Routes:     routes,
					},
				},
			})
			if err != nil {
				log.Fatalln(err)
				log.Panicf("Error creating tun/route: id/tun=%d/%d, err: %s", id, tunNum, err)
			}
		}
	}

	endTime := time.Now()

	log.Printf("total create time for client %d, tunnels: %d, time: %s\n",
		id, numTunnels, endTime.Sub(startTime).String())

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

// create the initial red and black memif's that kiknos uses ... ipsec wil ref the red ONLY
// i guess we dont need the black yet but maybe there will be a reason
func (p *GRPCStressPlugin) runGRPCCreateRedBlackMemifs(client configurator.ConfiguratorClient) {

	p.Log.Infof("Creating the red/black memif's ...")

	memifRedInfo := &interfaces.Interface_Memif{
		Memif: &interfaces.MemifLink{
			Id:             1000,
			Master:         false,
			SocketFilename: "/var/run/memif_k8s-master.sock",
		},
	}
	memIFRed := &interfaces.Interface{
		Name:        "red",
		Type:        interfaces.Interface_MEMIF,
		Enabled:     true,
		IpAddresses: []string{"100.100.100.100/24"},
		Mtu:         9000,
		Link:        memifRedInfo,
	}
	memifBlackInfo := &interfaces.Interface_Memif{
		Memif: &interfaces.MemifLink{
			Id:             1001,
			Master:         false,
			SocketFilename: "/var/run/memif_k8s-master.sock",
		},
	}
	memIFBlack := &interfaces.Interface{
		Name:        "black",
		Type:        interfaces.Interface_MEMIF,
		Enabled:     true,
		IpAddresses: []string{"20.20.20.100/24"},
		Mtu:         9000,
		Link:        memifBlackInfo,
	}
	ifaces := []*interfaces.Interface{memIFRed, memIFBlack}

	_, err := client.Update(context.Background(), &configurator.UpdateRequest{
		Update: &configurator.Config{
			VppConfig: &vpp.ConfigData{
				Interfaces: ifaces,
			},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

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

// Dialer for unix domain socket
func dialer(socket, address string, timeoutVal time.Duration) func(string, time.Duration) (net.Conn, error) {
	return func(addr string, timeout time.Duration) (net.Conn, error) {
		// Pass values
		addr, timeout = address, timeoutVal
		// Dial with timeout
		return net.DialTimeout(socket, addr, timeoutVal)
	}
}

func gen3octets(num uint32) string {
	return fmt.Sprintf("%d.%d.%d",
		(num>>16)&0xFF,
		(num>>8)&0xFF,
		(num)&0xFF)
}
