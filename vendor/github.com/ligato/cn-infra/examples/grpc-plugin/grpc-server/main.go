package main

import (
	"crypto/tls"
	"errors"
	"log"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/examples/grpc-plugin/insecure"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/cn-infra/rpc/rest"
	"golang.org/x/net/context"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

// *************************************************************************
// This file contains GRPC service exposure example. To register service use
// Server.RegisterService(descriptor, service)
// ************************************************************************/

// PluginName represents name of plugin.
const PluginName = "myPlugin"

func main() {
	grpcPlug := grpc.NewPlugin(
		grpc.UseHTTP(&rest.DefaultPlugin),
		grpc.UseConf(grpc.Config{
			Endpoint: "localhost:9111",
		}),
		grpc.UseAuth(&grpc.Authenticator{
			Username: "testuser",
			Password: "testpwd",
			Token:    "testtoken",
		}),
		grpc.UseTLS(&tls.Config{
			Certificates: []tls.Certificate{insecure.Cert},
			ClientCAs:    insecure.CertPool,
			ClientAuth:   tls.VerifyClientCertIfGiven,
		}),
	)
	p := &ExamplePlugin{
		GRPC: grpcPlug,
		Log:  logging.ForPlugin(PluginName),
	}

	a := agent.NewAgent(agent.AllPlugins(p))

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExamplePlugin presents main plugin.
type ExamplePlugin struct {
	Log  logging.PluginLogger
	GRPC grpc.Server
}

// String return name of the plugin.
func (p *ExamplePlugin) String() string {
	return PluginName
}

// Init demonstrates the usage of PluginLogger API.
func (p *ExamplePlugin) Init() error {
	p.Log.Info("Registering greeter")

	helloworld.RegisterGreeterServer(p.GRPC.GetServer(), &GreeterService{})

	return nil
}

// Close closes the plugin.
func (p *ExamplePlugin) Close() error {
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

	logging.Infof("Greeting client: %v", request.Name)

	return &helloworld.HelloReply{Message: "Greetings " + request.Name}, nil
}
