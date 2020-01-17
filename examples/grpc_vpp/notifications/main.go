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
	"io"
	"log"
	"net"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/namsral/flag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
)

var (
	address    = flag.String("address", "localhost:9111", "address of GRPC server")
	socketType = flag.String("socket-type", "tcp", "[tcp, tcp4, tcp6, unix, unixpacket]")
	reqPer     = flag.Int("request-period", 3, "notification request period in seconds")
)

// Start Agent plugins selected for this example.
func main() {
	// Inject dependencies to example plugin
	ep := &ExamplePlugin{}
	ep.SetName("remote-client-example")
	ep.Setup()

	// Start Agent
	a := agent.NewAgent(
		agent.AllPlugins(ep),
	)
	if err := a.Run(); err != nil {
		log.Fatal()
	}
}

// ExamplePlugin demonstrates the use of grpc to watch on VPP notifications using vpp-agent.
type ExamplePlugin struct {
	infra.PluginDeps

	conn *grpc.ClientConn
}

// Init initializes example plugin.
func (p *ExamplePlugin) Init() (err error) {
	// Set up connection to the server.
	p.conn, err = grpc.Dial("unix",
		grpc.WithInsecure(),
		grpc.WithDialer(dialer(*socketType, *address, 2*time.Second)))

	if err != nil {
		return err
	}

	client := configurator.NewConfiguratorServiceClient(p.conn)

	// Start notification watcher.
	go p.watchNotifications(client)

	logrus.DefaultLogger().Info("Initialization of the example plugin has completed")
	return err
}

// Get is an implementation of client-side statistics streaming.
func (p *ExamplePlugin) watchNotifications(client configurator.ConfiguratorServiceClient) {
	var nextIdx uint32

	logrus.DefaultLogger().Info("Watching..")
	for {
		// Prepare request with the initial index
		request := &configurator.NotifyRequest{
			Idx: nextIdx,
		}
		// Get stream object
		stream, err := client.Notify(context.Background(), request)
		if err != nil {
			logrus.DefaultLogger().Error(err)
			return
		}
		// Receive all message from the stream
		var recvNotifs int
		for {
			notif, err := stream.Recv()
			if err == io.EOF {
				if recvNotifs == 0 {
					//logrus.DefaultLogger().Info("No new notifications")
				} else {
					logrus.DefaultLogger().Infof("%d new notifications received", recvNotifs)
				}
				break
			}
			if err != nil {
				logrus.DefaultLogger().Error(err)
				return
			}

			logrus.DefaultLogger().Infof("Notification[%d]: %v",
				notif.NextIdx-1, notif.Notification)
			nextIdx = notif.NextIdx
			recvNotifs++
		}

		// Wait till next request
		time.Sleep(time.Duration(*reqPer) * time.Second)
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
