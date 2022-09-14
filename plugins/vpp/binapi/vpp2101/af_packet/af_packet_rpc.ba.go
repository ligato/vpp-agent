// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package af_packet

import (
	"context"
	"fmt"
	"io"

	api "go.fd.io/govpp/api"
	vpe "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/vpe"
)

// RPCService defines RPC service  af_packet.
type RPCService interface {
	AfPacketCreate(ctx context.Context, in *AfPacketCreate) (*AfPacketCreateReply, error)
	AfPacketDelete(ctx context.Context, in *AfPacketDelete) (*AfPacketDeleteReply, error)
	AfPacketDump(ctx context.Context, in *AfPacketDump) (RPCService_AfPacketDumpClient, error)
	AfPacketSetL4CksumOffload(ctx context.Context, in *AfPacketSetL4CksumOffload) (*AfPacketSetL4CksumOffloadReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) AfPacketCreate(ctx context.Context, in *AfPacketCreate) (*AfPacketCreateReply, error) {
	out := new(AfPacketCreateReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) AfPacketDelete(ctx context.Context, in *AfPacketDelete) (*AfPacketDeleteReply, error) {
	out := new(AfPacketDeleteReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) AfPacketDump(ctx context.Context, in *AfPacketDump) (RPCService_AfPacketDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_AfPacketDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_AfPacketDumpClient interface {
	Recv() (*AfPacketDetails, error)
	api.Stream
}

type serviceClient_AfPacketDumpClient struct {
	api.Stream
}

func (c *serviceClient_AfPacketDumpClient) Recv() (*AfPacketDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *AfPacketDetails:
		return m, nil
	case *vpe.ControlPingReply:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}

func (c *serviceClient) AfPacketSetL4CksumOffload(ctx context.Context, in *AfPacketSetL4CksumOffload) (*AfPacketSetL4CksumOffloadReply, error) {
	out := new(AfPacketSetL4CksumOffloadReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
