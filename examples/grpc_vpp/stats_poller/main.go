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
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/namsral/flag"
	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/infra"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
)

var (
	address    = flag.String("address", "localhost:9111", "address of GRPC server")
	socketType = flag.String("socket-type", "tcp", "[tcp, tcp4, tcp6, unix, unixpacket]")

	period = flag.Uint("period", 3, "Polling period (in seconds)")
	polls  = flag.Uint("polls", 0, "Number of pollings")
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
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.DialTimeout(*socketType, *address, time.Second*3)
		}),
		//grpc.WithContextDialer(dialer(*socketType, *address, time.Second*3)),
	)

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
	ctx := context.Background()

	req := &configurator.PollStatsRequest{
		PeriodSec: uint32(*period),
		NumPolls:  uint32(*polls),
	}
	fmt.Printf("Polling stats: %v\n", req)

	stream, err := client.PollStats(ctx, req)
	if err != nil {
		p.Log.Fatalln("PollStats failed:", err)
	}

	var lastSeq uint32
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			p.Log.Infof("Polling has completed.")
			os.Exit(0)
		} else if err != nil {
			p.Log.Fatalln("Recv failed:", err)
		}

		if resp.PollSeq != lastSeq {
			fmt.Printf(" --- Poll sequence: %-3v\n", resp.PollSeq)
		}
		lastSeq = resp.PollSeq

		vppStats := resp.GetStats().GetVppStats()
		fmt.Printf("VPP stats: %v\n", vppStats)
	}
}

// Dialer for unix domain socket
func dialer(socket, address string, timeoutVal time.Duration) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		return net.DialTimeout(socket, address, timeoutVal)
	}
}
