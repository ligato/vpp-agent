// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package vxlan_gpe

import (
	"context"
	"fmt"
	"io"

	api "git.fd.io/govpp.git/api"
	vpe "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/vpe"
)

// RPCService defines RPC service  vxlan_gpe.
type RPCService interface {
	SwInterfaceSetVxlanGpeBypass(ctx context.Context, in *SwInterfaceSetVxlanGpeBypass) (*SwInterfaceSetVxlanGpeBypassReply, error)
	VxlanGpeAddDelTunnel(ctx context.Context, in *VxlanGpeAddDelTunnel) (*VxlanGpeAddDelTunnelReply, error)
	VxlanGpeTunnelDump(ctx context.Context, in *VxlanGpeTunnelDump) (RPCService_VxlanGpeTunnelDumpClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) SwInterfaceSetVxlanGpeBypass(ctx context.Context, in *SwInterfaceSetVxlanGpeBypass) (*SwInterfaceSetVxlanGpeBypassReply, error) {
	out := new(SwInterfaceSetVxlanGpeBypassReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) VxlanGpeAddDelTunnel(ctx context.Context, in *VxlanGpeAddDelTunnel) (*VxlanGpeAddDelTunnelReply, error) {
	out := new(VxlanGpeAddDelTunnelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) VxlanGpeTunnelDump(ctx context.Context, in *VxlanGpeTunnelDump) (RPCService_VxlanGpeTunnelDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_VxlanGpeTunnelDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_VxlanGpeTunnelDumpClient interface {
	Recv() (*VxlanGpeTunnelDetails, error)
	api.Stream
}

type serviceClient_VxlanGpeTunnelDumpClient struct {
	api.Stream
}

func (c *serviceClient_VxlanGpeTunnelDumpClient) Recv() (*VxlanGpeTunnelDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *VxlanGpeTunnelDetails:
		return m, nil
	case *vpe.ControlPingReply:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}
