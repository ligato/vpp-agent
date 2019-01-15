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
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/namsral/flag"
	"google.golang.org/grpc"

	"github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/l3"
	"github.com/ligato/vpp-agent/api/models/vpp"
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
	ep := &ExamplePlugin{}
	ep.SetName("grpc-client-example")
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
	// Validate socket type
	switch *socketType {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
	default:
		return fmt.Errorf("unknown gRPC socket type: %s", socketType)
	}

	_, p.cancel = context.WithCancel(context.Background())

	// Set up connection to the server.
	p.conn, err = grpc.Dial("unix",
		grpc.WithInsecure(),
		grpc.WithDialer(dialer(*socketType, *address, dialTimeout)),
	)
	if err != nil {
		return err
	}

	p.Log.Info("Init complete")
	return nil
}

// AfterInit executes client demo.
func (p *ExamplePlugin) AfterInit() (err error) {
	go func() {
		time.Sleep(time.Second)

		demonstrateClient(p.conn)

		time.Sleep(time.Second)

		logrus.DefaultLogger().Info("Closing example")
		close(exampleFinished)
	}()
	return nil
}

// Close cleans up the resources.
func (p *ExamplePlugin) Close() error {
	logrus.DefaultLogger().Info("Closing example")

	p.cancel()
	p.wg.Wait()

	if err := p.conn.Close(); err != nil {
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

func demonstrateClient(conn *grpc.ClientConn) {
	c := remoteclient.NewClientGRPC(api.NewGenericConfiguratorClient(conn))

	// List supported model specs
	modules, err := c.ActiveModels()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Listing %d supported modules\n", len(modules))
	for module, models := range modules {
		fmt.Printf("* %s module (%d models)\n", module, len(models))
		for _, spec := range models {
			fmt.Printf(" - %v\n", spec.String())
		}
	}

	ctx := context.Background()

	fmt.Printf("Requesting resync\n")

	// Resync
	req := c.SetConfig(true)
	req.Update(
		memif1, memif2,
		veth1, veth2,
		routeX,
	)
	if err = req.Send(ctx); err != nil {
		log.Fatalln(err)
	}

	time.Sleep(time.Second * 5)
	fmt.Printf("Requesting change\n")

	memif1.Enabled = false
	memif1.Mtu = 666

	req2 := c.SetConfig(false)
	req2.Update(afp1, memif1)
	req2.Delete(memif2)
	//req2.Update(memif2)
	if err := req2.Send(context.Background()); err != nil {
		log.Fatalln(err)
	}

}

var (
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
		Name:        "memif1.1",
		Enabled:     true,
		Type:        interfaces.Interface_SUB_INTERFACE,
		IpAddresses: []string{"3.10.0.1/32"},
		Link: &interfaces.Interface_Sub{
			Sub: &interfaces.SubInterface{
				ParentName: "memif1",
				SubId:      10,
			},
		},
	}
	afp1 = &vpp.Interface{
		Name:        "afp1",
		Enabled:     true,
		Type:        interfaces.Interface_AF_PACKET,
		IpAddresses: []string{"10.10.3.5/24"},
		Link: &interfaces.Interface_Afpacket{
			Afpacket: &interfaces.AfpacketLink{
				HostIfName: "veth1",
			},
		},
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
	routeX = &linux.StaticRoute{
		DstNetwork:        "192.168.5.0/24",
		OutgoingInterface: "myVETH1",
		GwAddr:            "10.10.3.254",
		Scope:             linux_l3.StaticRoute_GLOBAL,
	}
)
