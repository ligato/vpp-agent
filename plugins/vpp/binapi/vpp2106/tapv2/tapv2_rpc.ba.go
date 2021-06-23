// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package tapv2

import (
	"context"
	"fmt"
	"io"

	api "git.fd.io/govpp.git/api"
	vpe "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vpe"
)

// RPCService defines RPC service tapv2.
type RPCService interface {
	SwInterfaceTapV2Dump(ctx context.Context, in *SwInterfaceTapV2Dump) (RPCService_SwInterfaceTapV2DumpClient, error)
	TapCreateV2(ctx context.Context, in *TapCreateV2) (*TapCreateV2Reply, error)
	TapDeleteV2(ctx context.Context, in *TapDeleteV2) (*TapDeleteV2Reply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) SwInterfaceTapV2Dump(ctx context.Context, in *SwInterfaceTapV2Dump) (RPCService_SwInterfaceTapV2DumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_SwInterfaceTapV2DumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_SwInterfaceTapV2DumpClient interface {
	Recv() (*SwInterfaceTapV2Details, error)
	api.Stream
}

type serviceClient_SwInterfaceTapV2DumpClient struct {
	api.Stream
}

func (c *serviceClient_SwInterfaceTapV2DumpClient) Recv() (*SwInterfaceTapV2Details, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *SwInterfaceTapV2Details:
		return m, nil
	case *vpe.ControlPingReply:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}

func (c *serviceClient) TapCreateV2(ctx context.Context, in *TapCreateV2) (*TapCreateV2Reply, error) {
	out := new(TapCreateV2Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) TapDeleteV2(ctx context.Context, in *TapDeleteV2) (*TapDeleteV2Reply, error) {
	out := new(TapDeleteV2Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}
