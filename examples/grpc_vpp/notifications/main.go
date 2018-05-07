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
	"flag"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"os"
	"sync"
	"time"
)

const defaultAddress = "localhost:9112"

var address = defaultAddress

// init sets the default logging level
func init() {
	log.DefaultLogger().SetOutput(os.Stdout)
	log.DefaultLogger().SetLevel(logging.DebugLevel)
}

// Start Agent plugins selected for this example.
func main() {
	// Init close channel to stop the example.
	closeChannel := make(chan struct{}, 1)

	flag.StringVar(&address, "address", defaultAddress, "address of GRPC server")

	// Example plugin
	agent := local.NewAgent(local.WithPlugins(func(flavor *local.FlavorLocal) []*core.NamedPlugin {
		examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{}}

		return []*core.NamedPlugin{{examplePlugin.PluginName, examplePlugin}}
	}))

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// PluginID of example plugin
const PluginID core.PluginName = "example-plugin"

// ExamplePlugin demonstrates the use of the remoteclient to locally transport example configuration into the default VPP plugins.
type ExamplePlugin struct {
	wg sync.WaitGroup
	// GRPC server instance
	grpcServer *grpc.Server
	//
	listener net.Listener
}

type StatisticsService struct {
}

// Init initializes example plugin.
func (plugin *ExamplePlugin) Init() (err error) {
	log.DefaultLogger().Infof("Initializing GRPC server on tcp://%s", address)
	// Initialize new GRPC server
	plugin.grpcServer = grpc.NewServer()
	// Register statistics service to the server
	rpc.RegisterStatisticsServiceServer(plugin.grpcServer, &StatisticsService{})
	// Start GRPC listener
	plugin.listener, err = Listen(plugin.grpcServer)

	return err
}

// Close cleans up the resources.
func (plugin *ExamplePlugin) Close() error {
	_, err := safeclose.CloseAll(plugin.listener, plugin.grpcServer)
	if err != nil {
		return err
	}

	log.DefaultLogger().Info("Closed example plugin")
	return nil
}

// Send is an implementation of client-side statistics streaming.
func (svc *StatisticsService) Send(ctx context.Context, stats *rpc.Statistics) (*rpc.StatisticsResponse, error) {
	if stats.IfNotif != nil {
		log.DefaultLogger().Infof("Received interface notification (type %s) for interface %s:\n,%v",
			stats.IfNotif.Type, stats.IfNotif.State.Name, stats.IfNotif.State)
	}
	// todo add other notification types
	return &rpc.StatisticsResponse{}, nil
}

// Listen on provided address
func Listen(grpc *grpc.Server) (net.Listener, error) {
	netListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	var errCh chan error
	go func() {
		if err := grpc.Serve(netListener); err != nil {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	select {
	case err := <-errCh:
		return nil, err
	case <-time.After(100 * time.Millisecond):
		return netListener, nil
	}
}
