//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package grpc

import (
	"context"
	"io"
	"net"
	"time"

	"go.ligato.io/cn-infra/v2/infra"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// Client is aGRPC plugin structure
type Client struct {
	Deps // external dependencies

	connection *grpc.ClientConn
	client     configurator.ConfiguratorServiceClient
}

// Deps is a structure for the plugin external dependencies
type Deps struct {
	infra.PluginDeps
}

// Inti is the initialization function, called when the agent in started
func (p *Client) Init() (err error) {
	p.connection, err = grpc.Dial("unix",
		grpc.WithInsecure(),
		grpc.WithDialer(dialer("tcp", "0.0.0.0:9111", time.Second*2)),
	)
	if err != nil {
		return err
	}

	p.client = configurator.NewConfiguratorServiceClient(p.connection)

	p.Log.Info("GRPC client is connected")
	// Start notification watcher
	go p.watchNotif()
	go p.configure()

	return nil
}

// Close function, called on the shutdown
func (p *Client) Close() (err error) {
	return nil
}

// String is the GRPC plugin string representation
func (p *Client) String() string {
	return "GRPC-client"
}

// Dialer function used as a parameter for 'grpc.WithDialer'
func dialer(socket, address string, timeoutVal time.Duration) func(string, time.Duration) (net.Conn, error) {
	return func(addr string, timeout time.Duration) (net.Conn, error) {
		addr, timeout = address, timeoutVal
		return net.DialTimeout(socket, addr, timeoutVal)
	}
}

// Configure is a helper struct to demonstrate plugin functionality
func (p *Client) configure() {
	time.Sleep(2 * time.Second)
	_, err := p.client.Update(context.Background(), &configurator.UpdateRequest{
		Update: &configurator.Config{
			VppConfig: &vpp.ConfigData{
				Interfaces: []*vpp.Interface{
					{
						Name:        "interface1",
						Type:        vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
						Enabled:     true,
						IpAddresses: []string{"10.0.0.1/24"},
					},
				},
			},
		},
		FullResync: true,
	})
	if err != nil {
		p.Log.Errorf("Error putting GRPC data: %v", err)
		return
	}
	p.Log.Infof("GRPC data sent")
}

// WatchNotif shows how to implement GRPC notification watcher
func (p *Client) watchNotif() {
	p.Log.Info("Notification watcher started")
	var nextIdx uint32
	for {
		request := &configurator.NotifyRequest{
			Idx: nextIdx,
		}
		stream, err := p.client.Notify(context.Background(), request)
		if err != nil {
			p.Log.Error(err)
			return
		}
		var recvNotifs int
		for {
			notif, err := stream.Recv()
			if err == io.EOF {
				if recvNotifs == 0 {
					// Nothing to do
				} else {
					p.Log.Infof("%d new notifications received", recvNotifs)
				}
				break
			}
			if err != nil {
				p.Log.Error(err)
				return
			}

			p.Log.Infof("Notification[%d]: %v", notif.NextIdx-1, notif.Notification)
			nextIdx = notif.NextIdx
			recvNotifs++
		}

		time.Sleep(time.Second * 1)
	}
}
