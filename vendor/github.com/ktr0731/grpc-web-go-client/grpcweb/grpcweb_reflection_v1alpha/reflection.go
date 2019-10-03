package grpcweb_reflection_v1alpha

import (
	"errors"

	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	pb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type serverReflectionClient struct {
	cc *grpcweb.ClientConn
}

// NewServerReflectionClient instantiates a new server reflection client.
// most part of the implementation is same as the original grpc_reflection_v1alpha package's.
//
// the version (like v1alpha) is corrensponding to grpc_reflection_v1alpha package
func NewServerReflectionClient(cc *grpcweb.ClientConn) pb.ServerReflectionClient {
	return &serverReflectionClient{cc}
}

func (c *serverReflectionClient) ServerReflectionInfo(ctx context.Context, opts ...grpc.CallOption) (pb.ServerReflection_ServerReflectionInfoClient, error) {
	if len(opts) != 0 {
		return nil, errors.New("currently, ktr0731/grpc-web-go-client does not support grpc.CallOption")
	}

	stream, err := c.cc.NewBidiStream(
		&grpc.StreamDesc{ServerStreams: true, ClientStreams: true},
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo")
	if err != nil {
		return nil, err
	}

	return &serverReflectionServerReflectionInfoClient{ctx: ctx, stream: stream}, nil
}

type serverReflectionServerReflectionInfoClient struct {
	ctx    context.Context
	stream grpcweb.BidiStream

	// To satisfy pb.ServerReflection_ServerReflectionInfoClient
	grpc.ClientStream
}

func (x *serverReflectionServerReflectionInfoClient) Send(m *pb.ServerReflectionRequest) error {
	return x.stream.Send(x.ctx, m)
}

func (x *serverReflectionServerReflectionInfoClient) Recv() (*pb.ServerReflectionResponse, error) {
	var res pb.ServerReflectionResponse
	err := x.stream.Receive(x.ctx, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (x *serverReflectionServerReflectionInfoClient) CloseSend() error {
	return x.stream.CloseSend()
}
