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

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/client"
	"github.com/ligato/vpp-agent/cmd/vpp-agent/app/v2"
	"github.com/ligato/vpp-agent/examples/custom_model/pb"
	"github.com/ligato/vpp-agent/plugins/orchestrator"
	"github.com/namsral/flag"
	"google.golang.org/grpc"

	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/l3"
	"github.com/ligato/vpp-agent/api/models/vpp"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/api/models/vpp/l2"
)

var (
	address    = flag.String("address", "172.17.0.2:9111", "address of GRPC server")
	socketType = flag.String("socket-type", "tcp", "socket type [tcp, tcp4, tcp6, unix, unixpacket]")

	dialTimeout = time.Second * 2
)

var exampleFinished = make(chan struct{})

func main() {
	ep := &ExamplePlugin{}
	ep.Deps = Deps{
		Orchestrator: &orchestrator.DefaultPlugin,
		VPP:          appv2.DefaultVPP(),
		Linux:        appv2.DefaultLinux(),
	}
	ep.SetName("custom-model-example")
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
	Deps

	conn *grpc.ClientConn

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

type Deps struct {
	infra.PluginDeps
	Orchestrator *orchestrator.Plugin
	appv2.VPP
	appv2.Linux
}

// Init initializes example plugin.
func (p *ExamplePlugin) Init() (err error) {
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

		//c := remoteclient.NewClientGRPC(api.NewGenericManagerClient(conn))
		c := client.LocalClient

		demonstrateClient(c)

		time.Sleep(time.Second * 3)

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

func demonstrateClient(c client.ConfigClient) {
	// ==========================================
	// List known models
	// ==========================================
	knownModels, err := c.KnownModels()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Listing %d known models..\n", len(knownModels))
	for _, model := range knownModels {
		fmt.Printf(" - %v\n", model.String())
	}
	time.Sleep(time.Second * 3)

	// ==========================================
	// Resync config
	// ==========================================
	fmt.Printf("Requesting config resync..\n")
	err = c.ResyncConfig(
		memif1, memif2,
		veth1, veth2,
		routeX,
	)
	if err != nil {
		log.Fatalln(err)
	}
	time.Sleep(time.Second * 5)

	// ==========================================
	// Change config
	// ==========================================
	fmt.Printf("Requesting config change..\n")
	memif1.Enabled = false
	memif1.Mtu = 666

	custom := &mymodel.MyModel{
		Name:  "my1",
		Mynum: 33,
	}

	req := c.ChangeRequest()
	req.Update(afp1, memif1, bd1, vppRoute1, custom)
	req.Delete(memif2)
	if err := req.Send(context.Background()); err != nil {
		log.Fatalln(err)
	}
	time.Sleep(time.Second * 5)

	// ==========================================
	// Get config
	// ==========================================
	fmt.Printf("Retrieving config..\n")
	data := &configurator.Config{
		VppConfig:   &vpp.ConfigData{},
		LinuxConfig: &linux.ConfigData{},
	}
	if err := c.GetConfig(data.VppConfig, data.LinuxConfig); err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Retrieved config:\n%+v\n", proto.MarshalTextString(data))
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
		IpAddresses: []string{"3.10.0.1/24"},
		Link: &interfaces.Interface_Sub{
			Sub: &interfaces.SubInterface{
				ParentName: "memif1",
				SubId:      10,
			},
		},
	}
	bd1 = &vpp.BridgeDomain{
		Name: "bd1",
		Interfaces: []*vpp_l2.BridgeDomain_Interface{
			{Name: "memif1"},
		},
	}
	vppRoute1 = &vpp.Route{
		OutgoingInterface: "memif1",
		DstNetwork:        "4.4.10.0/24",
		NextHopAddr:       "3.10.0.5",
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
	routeX = &linux.Route{
		DstNetwork:        "192.168.5.0/24",
		OutgoingInterface: "myVETH1",
		GwAddr:            "10.10.3.254",
		Scope:             linux_l3.Route_GLOBAL,
	}
)
