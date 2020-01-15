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

package main

import (
	"log"
	"net"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/infra"
	"github.com/namsral/flag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
)

var (
	address    = flag.String("address", "localhost:9111", "address of GRPC server")
	socketType = flag.String("socket-type", "tcp", "[tcp, tcp4, tcp6, unix, unixpacket]")
	period     = flag.Uint("period", 3, "Polling period (in seconds)")
)

func main() {
	ep := &ExamplePlugin{}
	ep.SetName("stats-poller-example")
	ep.Setup()

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
		grpc.WithDialer(dialer(*socketType, *address, time.Second*3)))

	if err != nil {
		return err
	}

	client := configurator.NewStatsPollerServiceClient(p.conn)

	// Start stats poller.
	go p.pollStats(client)

	return err
}

// Get is an implementation of client-side statistics streaming.
func (p *ExamplePlugin) pollStats(client configurator.StatsPollerServiceClient) {
	p.Log.Infof("Polling every %v seconds..", *period)

	req := &configurator.PollStatsRequest{
		PeriodSec: uint32(*period),
	}

	ctx := context.Background()
	stream, err := client.PollStats(ctx, req)
	if err != nil {
		p.Log.Fatalln("PollStats failed:", err)
	}

	var lastSeq uint32
	for {
		resp, err := stream.Recv()
		if err != nil {
			p.Log.Fatalln("Recv failed:", err)
		}

		if resp.PollSeq != lastSeq {
			p.Log.Infof(" --- Poll sequence: %-3v", resp.PollSeq)
		}
		lastSeq = resp.PollSeq

		vppStats := resp.GetStats().GetVppStats()
		p.Log.Infof("VPP stats: %v", vppStats)
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
