package main

import (
	"flag"

	"github.com/ligato/cn-infra/logging/logroot"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	defaultAddress = "localhost:9111"
	defaultName    = "world"
)

var address = defaultAddress
var name = defaultName

func main() {
	flag.StringVar(&address, "address", defaultAddress, "address of GRPC server")
	flag.StringVar(&name, "name", defaultName, "name used in GRPC request")
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		logroot.StandardLogger().Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name})
	if err != nil {
		logroot.StandardLogger().Fatalf("could not greet: %v", err)
	}
	logroot.StandardLogger().Printf("Greeting: %s (received from server)", r.Message)
}
