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
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/namsral/flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/pkg/version"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

var (
	reg              = prometheus.NewRegistry()
	grpcMetrics      = grpc_prometheus.NewClientMetrics()
	perfTestSettings *prometheus.GaugeVec
)

func init() {
	flag.Parse()

	grpcMetrics.EnableClientHandlingTimeHistogram()
	reg.MustRegister(grpcMetrics)
	perfTestSettings = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "ligato",
		Subsystem: "perf_test",
		Name:      "client_settings",
		Help:      "",
		ConstLabels: map[string]string{
			"num_clients": fmt.Sprint(*numClients),
			"num_tunnels": fmt.Sprint(*numTunnels),
			"num_per_req": fmt.Sprint(*numPerRequest),
		},
	}, []string{"start_time"})
	reg.MustRegister(perfTestSettings)
}

var (
	address        = flag.String("address", "127.0.0.1:9111", "address of GRPC server")
	socketType     = flag.String("socket-type", "tcp", "socket type [tcp, tcp4, tcp6, unix, unixpacket]")
	numClients     = flag.Int("clients", 1, "number of concurrent grpc clients")
	numTunnels     = flag.Int("tunnels", 100, "number of tunnels to stress per client")
	numPerRequest  = flag.Int("numperreq", 1, "number of tunnels/routes per grpc request")
	withIPs        = flag.Bool("with-ips", false, "configure IP address for each tunnel on memif at the end")
	debug          = flag.Bool("debug", false, "turn on debug dump")
	dumpMetrics    = flag.Bool("dumpmetrics", false, "Dump metrics before exit.")
	timeout        = flag.Uint("timeout", 300, "timeout for requests (in seconds)")
	reportProgress = flag.Uint("progress", 20, "percent of progress to report")

	dialTimeout = time.Second * 3
	reqTimeout  = time.Second * 300
)

func main() {
	if *debug {
		logging.DefaultLogger.SetLevel(logging.DebugLevel)
	}

	perfTestSettings.WithLabelValues(time.Now().Format(time.Stamp)).Set(1)

	go serveMetrics()

	quit := make(chan struct{})

	ep := NewGRPCStressPlugin()

	ver, rev, date := version.Data()
	a := agent.NewAgent(
		agent.AllPlugins(ep),
		agent.QuitOnClose(quit),
		agent.Version(ver, date, rev),
	)

	if err := a.Start(); err != nil {
		log.Fatalln(err)
	}

	ep.setupInitial()
	ep.runAllClients()

	if err := a.Stop(); err != nil {
		log.Fatalln(err)
	}

	if *dumpMetrics {
		resp, err := http.Get("http://localhost:9094/metrics")
		if err != nil {
			log.Fatalln(err)
		}
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			log.Fatalln(err)
		} else {
			fmt.Println("----------------------")
			fmt.Println("-> CLIENT METRICS")
			fmt.Println("----------------------")
			fmt.Print(string(b))
			fmt.Println("----------------------")
		}
	}

	time.Sleep(time.Second * 5)
}

func serveMetrics() {
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	// Create a HTTP server for prometheus.
	httpServer := &http.Server{
		Handler: h,
		Addr:    fmt.Sprintf(":%d", 9094),
	}

	// Start your http server for prometheus.
	if err := httpServer.ListenAndServe(); err != nil {
		log.Println("Unable to start a http server.")
	}
}

// GRPCStressPlugin makes use of the remoteclient to locally CRUD ipsec tunnels and routes.
type GRPCStressPlugin struct {
	infra.PluginName
	Log *logrus.Logger

	conn  *grpc.ClientConn
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
	if p.conn != nil {
		return p.conn.Close()
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

func (p *GRPCStressPlugin) setupInitial() {
	conn, err := grpc.Dial("unix",
		grpc.WithInsecure(),
		grpc.WithDialer(dialer(*socketType, *address, dialTimeout)),
		grpc.WithUnaryInterceptor(grpcMetrics.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(grpcMetrics.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatal(err)
	}
	p.conn = conn

	reqTimeout = time.Second * time.Duration(*timeout)

	client := configurator.NewConfiguratorServiceClient(conn)

	if *debug {
		p.Log.Infof("Requesting get..")
		cfg, err := client.Get(context.Background(), &configurator.GetRequest{})
		if err != nil {
			log.Fatalln(err)
		}
		out, _ := (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(cfg)
		fmt.Printf("Config:\n %+v\n", out)

		p.Log.Infof("Requesting dump..")
		dump, err := client.Dump(context.Background(), &configurator.DumpRequest{})
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Dump:\n %+v\n", proto.MarshalTextString(dump))
	}

	time.Sleep(time.Second * 1)

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
	p.Log.Infof("----------------------------------------")
	p.Log.Infof(" SETTINGS:")
	p.Log.Infof("----------------------------------------")
	p.Log.Infof(" -> Clients: %d", *numClients)
	p.Log.Infof(" -> Requests: %d", *numTunnels)
	p.Log.Infof(" -> Tunnels per request: %d", *numPerRequest)
	p.Log.Infof("----------------------------------------")
	p.Log.Infof("Launching all clients..")

	t := time.Now()

	p.wg.Add(*numClients)
	for i := 0; i < *numClients; i++ {
		// Set up connection to the server.
		/*conn, err := grpc.Dial("unix",
			grpc.WithInsecure(),
			grpc.WithDialer(dialer(*socketType, *address, dialTimeout)),
		)
		if err != nil {
			log.Fatal(err)
		}
		p.conns = append(p.conns, conn)
		client := configurator.NewConfiguratorServiceClient(p.conns[i])*/

		client := configurator.NewConfiguratorServiceClient(p.conn)

		go p.runGRPCStressCreate(i, client, *numTunnels)
	}

	p.Log.Debugf("Waiting for clients..")

	p.wg.Wait()

	took := time.Since(t)
	perSec := float64((*numTunnels)*(*numClients)) / took.Seconds()

	p.Log.Infof("All clients done!")
	p.Log.Infof("========================================")
	p.Log.Infof(" RESULTS:")
	p.Log.Infof("========================================")
	p.Log.Infof("	Elapsed: %.2f sec", took.Seconds())
	p.Log.Infof("	Average: %.1f req/sec", perSec)
	p.Log.Infof("========================================")

	/*for i := 0; i < *numClients; i++ {
		if err := p.conns[i].Close(); err != nil {
			log.Fatal(err)
		}
	}*/

	if *debug {
		client := configurator.NewConfiguratorServiceClient(p.conn)

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

// runGRPCStressCreate creates 1 tunnel and 1 route ... emulating what strongswan does on a per remote warrior
func (p *GRPCStressPlugin) runGRPCStressCreate(clientId int, client configurator.ConfiguratorServiceClient, numTunnels int) {
	defer p.wg.Done()

	p.Log.Debugf("Creating %d tunnels/routes ... for client %d, ", numTunnels, clientId)

	startTime := time.Now()

	ips := []string{"10.0.0.1/24"}

	report := 0.0
	lastNumTunnels := 0
	lastReport := startTime

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

			tunID := clientId*numTunnels + tunNum
			tunNum++

			ipsecTunnelName := fmt.Sprintf("ipsec-%d", tunID)

			ipPart0 := 100 + (uint32(tunID)>>16)&0xFF
			ipPart := gen2octets(uint32(tunID))
			localIP := fmt.Sprintf("%d.%s.1", ipPart0, ipPart)
			remoteIP := fmt.Sprintf("%d.%s.254", ipPart0, ipPart)

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

			//p.Log.Infof("Creating %s ... client: %d, tunNum: %d", ipsecTunnelName, clientId, tunNum)

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
			log.Fatalf("Error creating tun/route: clientId/tun=%d/%d, err: %s", clientId, tunNum, err)
		}

		progress := (float64(tunNum) / float64(numTunnels)) * 100
		if uint(progress-report) >= *reportProgress {
			tunNumReport := tunNum - lastNumTunnels

			took := time.Since(lastReport)
			perSec := float64(tunNumReport) / took.Seconds()

			p.Log.Infof("client #%d - progress % 3.0f%% -> %d tunnels took %.3fs (%.1f tunnels/sec)",
				clientId, progress, tunNumReport, took.Seconds(), perSec)

			report = progress
			lastReport = time.Now()
			lastNumTunnels = tunNum
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

	took := time.Since(startTime)
	perSec := float64(numTunnels) / took.Seconds()

	p.Log.Infof("client #%d done => %d tunnels took %.3fs (%.1f tunnels/sec)",
		clientId, numTunnels, took.Seconds(), perSec)
}

func gen3octets(num uint32) string {
	return fmt.Sprintf("%d.%d.%d", (num>>16)&0xFF, (num>>8)&0xFF, (num)&0xFF)
}

func gen2octets(num uint32) string {
	return fmt.Sprintf("%d.%d", (num>>8)&0xFF, (num)&0xFF)
}
