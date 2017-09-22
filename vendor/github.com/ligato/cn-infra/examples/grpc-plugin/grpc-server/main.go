package main

import (
	"time"

	"errors"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/logroot"
	"golang.org/x/net/context"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

// *************************************************************************
// This file contains GRPC service exposure example. To register service use
// Server.RegisterService(descriptor, service)
// ************************************************************************/

func main() {
	// Init close channel to stop the example after everything was logged
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor
	// (combination of ExamplePlugin & reused cn-infra plugins).
	flavor := ExampleFlavor{ExamplePlugin: ExamplePlugin{exampleFinished: exampleFinished}}
	agent := core.NewAgent(logroot.StandardLogger(), 15*time.Second, flavor.Plugins()...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// ExamplePlugin presents the PluginLogger API.
type ExamplePlugin struct {
	Deps
	exampleFinished chan struct{}
}

// Init demonstrates the usage of PluginLogger API.
func (plugin *ExamplePlugin) Init() (err error) {
	plugin.Log.Info("Example Init")

	helloworld.RegisterGreeterServer(plugin.GRPC.Server(), &GreeterService{})

	return nil
}

// GreeterService implements GRPC GreeterServer interface (interface generated from protobuf definition file).
// It is a simple implementation for testing/demo only purposes.
type GreeterService struct{}

// SayHello returns error if request.name was not filled otherwise: "hello " + request.Name
func (*GreeterService) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	if request.Name == "" {
		return nil, errors.New("not filled name in the request")
	}

	return &helloworld.HelloReply{Message: "hello " + request.Name}, nil
}
