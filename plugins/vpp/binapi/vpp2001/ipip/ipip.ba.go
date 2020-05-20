// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/ipip.api.json

/*
Package ipip is a generated VPP binary API for 'ipip' module.

It consists of:
	 11 enums
	  6 aliases
	  7 types
	  1 union
	 10 messages
	  5 services
*/
package ipip

import (
	"bytes"
	"context"
	"io"
	"strconv"

	api "git.fd.io/govpp.git/api"
	struc "github.com/lunixbochs/struc"

	interface_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interface_types"
	ip_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
)

const (
	// ModuleName is the name of this module.
	ModuleName = "ipip"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0xf108649c
)

type AddressFamily = ip_types.AddressFamily

type IfStatusFlags = interface_types.IfStatusFlags

type IfType = interface_types.IfType

type IPDscp = ip_types.IPDscp

type IPEcn = ip_types.IPEcn

type IPProto = ip_types.IPProto

// IpipTunnelFlags represents VPP binary API enum 'ipip_tunnel_flags'.
type IpipTunnelFlags uint8

const (
	IPIP_TUNNEL_API_FLAG_NONE            IpipTunnelFlags = 0
	IPIP_TUNNEL_API_FLAG_ENCAP_COPY_DF   IpipTunnelFlags = 1
	IPIP_TUNNEL_API_FLAG_ENCAP_SET_DF    IpipTunnelFlags = 2
	IPIP_TUNNEL_API_FLAG_ENCAP_COPY_DSCP IpipTunnelFlags = 4
	IPIP_TUNNEL_API_FLAG_ENCAP_COPY_ECN  IpipTunnelFlags = 8
	IPIP_TUNNEL_API_FLAG_DECAP_COPY_ECN  IpipTunnelFlags = 16
)

var IpipTunnelFlags_name = map[uint8]string{
	0:  "IPIP_TUNNEL_API_FLAG_NONE",
	1:  "IPIP_TUNNEL_API_FLAG_ENCAP_COPY_DF",
	2:  "IPIP_TUNNEL_API_FLAG_ENCAP_SET_DF",
	4:  "IPIP_TUNNEL_API_FLAG_ENCAP_COPY_DSCP",
	8:  "IPIP_TUNNEL_API_FLAG_ENCAP_COPY_ECN",
	16: "IPIP_TUNNEL_API_FLAG_DECAP_COPY_ECN",
}

var IpipTunnelFlags_value = map[string]uint8{
	"IPIP_TUNNEL_API_FLAG_NONE":            0,
	"IPIP_TUNNEL_API_FLAG_ENCAP_COPY_DF":   1,
	"IPIP_TUNNEL_API_FLAG_ENCAP_SET_DF":    2,
	"IPIP_TUNNEL_API_FLAG_ENCAP_COPY_DSCP": 4,
	"IPIP_TUNNEL_API_FLAG_ENCAP_COPY_ECN":  8,
	"IPIP_TUNNEL_API_FLAG_DECAP_COPY_ECN":  16,
}

func (x IpipTunnelFlags) String() string {
	s, ok := IpipTunnelFlags_name[uint8(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

type LinkDuplex = interface_types.LinkDuplex

type MtuProto = interface_types.MtuProto

type RxMode = interface_types.RxMode

type SubIfFlags = interface_types.SubIfFlags

type AddressWithPrefix = ip_types.AddressWithPrefix

type InterfaceIndex = interface_types.InterfaceIndex

type IP4Address = ip_types.IP4Address

type IP4AddressWithPrefix = ip_types.IP4AddressWithPrefix

type IP6Address = ip_types.IP6Address

type IP6AddressWithPrefix = ip_types.IP6AddressWithPrefix

type Address = ip_types.Address

type IP4Prefix = ip_types.IP4Prefix

type IP6Prefix = ip_types.IP6Prefix

// IpipTunnel represents VPP binary API type 'ipip_tunnel'.
type IpipTunnel struct {
	Instance  uint32
	Src       Address
	Dst       Address
	SwIfIndex InterfaceIndex
	TableID   uint32
	Flags     IpipTunnelFlags
	Dscp      IPDscp
}

func (*IpipTunnel) GetTypeName() string { return "ipip_tunnel" }

type Mprefix = ip_types.Mprefix

type Prefix = ip_types.Prefix

type PrefixMatcher = ip_types.PrefixMatcher

type AddressUnion = ip_types.AddressUnion

// Ipip6rdAddTunnel represents VPP binary API message 'ipip_6rd_add_tunnel'.
type Ipip6rdAddTunnel struct {
	IP6TableID    uint32
	IP4TableID    uint32
	IP6Prefix     IP6Prefix
	IP4Prefix     IP4Prefix
	IP4Src        IP4Address
	SecurityCheck bool
	TcTos         uint8
}

func (m *Ipip6rdAddTunnel) Reset()                        { *m = Ipip6rdAddTunnel{} }
func (*Ipip6rdAddTunnel) GetMessageName() string          { return "ipip_6rd_add_tunnel" }
func (*Ipip6rdAddTunnel) GetCrcString() string            { return "56e93cc0" }
func (*Ipip6rdAddTunnel) GetMessageType() api.MessageType { return api.RequestMessage }

// Ipip6rdAddTunnelReply represents VPP binary API message 'ipip_6rd_add_tunnel_reply'.
type Ipip6rdAddTunnelReply struct {
	Retval    int32
	SwIfIndex InterfaceIndex
}

func (m *Ipip6rdAddTunnelReply) Reset()                        { *m = Ipip6rdAddTunnelReply{} }
func (*Ipip6rdAddTunnelReply) GetMessageName() string          { return "ipip_6rd_add_tunnel_reply" }
func (*Ipip6rdAddTunnelReply) GetCrcString() string            { return "5383d31f" }
func (*Ipip6rdAddTunnelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// Ipip6rdDelTunnel represents VPP binary API message 'ipip_6rd_del_tunnel'.
type Ipip6rdDelTunnel struct {
	SwIfIndex InterfaceIndex
}

func (m *Ipip6rdDelTunnel) Reset()                        { *m = Ipip6rdDelTunnel{} }
func (*Ipip6rdDelTunnel) GetMessageName() string          { return "ipip_6rd_del_tunnel" }
func (*Ipip6rdDelTunnel) GetCrcString() string            { return "f9e6675e" }
func (*Ipip6rdDelTunnel) GetMessageType() api.MessageType { return api.RequestMessage }

// Ipip6rdDelTunnelReply represents VPP binary API message 'ipip_6rd_del_tunnel_reply'.
type Ipip6rdDelTunnelReply struct {
	Retval int32
}

func (m *Ipip6rdDelTunnelReply) Reset()                        { *m = Ipip6rdDelTunnelReply{} }
func (*Ipip6rdDelTunnelReply) GetMessageName() string          { return "ipip_6rd_del_tunnel_reply" }
func (*Ipip6rdDelTunnelReply) GetCrcString() string            { return "e8d4e804" }
func (*Ipip6rdDelTunnelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// IpipAddTunnel represents VPP binary API message 'ipip_add_tunnel'.
type IpipAddTunnel struct {
	Tunnel IpipTunnel
}

func (m *IpipAddTunnel) Reset()                        { *m = IpipAddTunnel{} }
func (*IpipAddTunnel) GetMessageName() string          { return "ipip_add_tunnel" }
func (*IpipAddTunnel) GetCrcString() string            { return "ef93caab" }
func (*IpipAddTunnel) GetMessageType() api.MessageType { return api.RequestMessage }

// IpipAddTunnelReply represents VPP binary API message 'ipip_add_tunnel_reply'.
type IpipAddTunnelReply struct {
	Retval    int32
	SwIfIndex InterfaceIndex
}

func (m *IpipAddTunnelReply) Reset()                        { *m = IpipAddTunnelReply{} }
func (*IpipAddTunnelReply) GetMessageName() string          { return "ipip_add_tunnel_reply" }
func (*IpipAddTunnelReply) GetCrcString() string            { return "5383d31f" }
func (*IpipAddTunnelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// IpipDelTunnel represents VPP binary API message 'ipip_del_tunnel'.
type IpipDelTunnel struct {
	SwIfIndex InterfaceIndex
}

func (m *IpipDelTunnel) Reset()                        { *m = IpipDelTunnel{} }
func (*IpipDelTunnel) GetMessageName() string          { return "ipip_del_tunnel" }
func (*IpipDelTunnel) GetCrcString() string            { return "f9e6675e" }
func (*IpipDelTunnel) GetMessageType() api.MessageType { return api.RequestMessage }

// IpipDelTunnelReply represents VPP binary API message 'ipip_del_tunnel_reply'.
type IpipDelTunnelReply struct {
	Retval int32
}

func (m *IpipDelTunnelReply) Reset()                        { *m = IpipDelTunnelReply{} }
func (*IpipDelTunnelReply) GetMessageName() string          { return "ipip_del_tunnel_reply" }
func (*IpipDelTunnelReply) GetCrcString() string            { return "e8d4e804" }
func (*IpipDelTunnelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// IpipTunnelDetails represents VPP binary API message 'ipip_tunnel_details'.
type IpipTunnelDetails struct {
	Tunnel IpipTunnel
}

func (m *IpipTunnelDetails) Reset()                        { *m = IpipTunnelDetails{} }
func (*IpipTunnelDetails) GetMessageName() string          { return "ipip_tunnel_details" }
func (*IpipTunnelDetails) GetCrcString() string            { return "7f7b5b85" }
func (*IpipTunnelDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// IpipTunnelDump represents VPP binary API message 'ipip_tunnel_dump'.
type IpipTunnelDump struct {
	SwIfIndex InterfaceIndex
}

func (m *IpipTunnelDump) Reset()                        { *m = IpipTunnelDump{} }
func (*IpipTunnelDump) GetMessageName() string          { return "ipip_tunnel_dump" }
func (*IpipTunnelDump) GetCrcString() string            { return "f9e6675e" }
func (*IpipTunnelDump) GetMessageType() api.MessageType { return api.RequestMessage }

func init() {
	api.RegisterMessage((*Ipip6rdAddTunnel)(nil), "ipip.Ipip6rdAddTunnel")
	api.RegisterMessage((*Ipip6rdAddTunnelReply)(nil), "ipip.Ipip6rdAddTunnelReply")
	api.RegisterMessage((*Ipip6rdDelTunnel)(nil), "ipip.Ipip6rdDelTunnel")
	api.RegisterMessage((*Ipip6rdDelTunnelReply)(nil), "ipip.Ipip6rdDelTunnelReply")
	api.RegisterMessage((*IpipAddTunnel)(nil), "ipip.IpipAddTunnel")
	api.RegisterMessage((*IpipAddTunnelReply)(nil), "ipip.IpipAddTunnelReply")
	api.RegisterMessage((*IpipDelTunnel)(nil), "ipip.IpipDelTunnel")
	api.RegisterMessage((*IpipDelTunnelReply)(nil), "ipip.IpipDelTunnelReply")
	api.RegisterMessage((*IpipTunnelDetails)(nil), "ipip.IpipTunnelDetails")
	api.RegisterMessage((*IpipTunnelDump)(nil), "ipip.IpipTunnelDump")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*Ipip6rdAddTunnel)(nil),
		(*Ipip6rdAddTunnelReply)(nil),
		(*Ipip6rdDelTunnel)(nil),
		(*Ipip6rdDelTunnelReply)(nil),
		(*IpipAddTunnel)(nil),
		(*IpipAddTunnelReply)(nil),
		(*IpipDelTunnel)(nil),
		(*IpipDelTunnelReply)(nil),
		(*IpipTunnelDetails)(nil),
		(*IpipTunnelDump)(nil),
	}
}

// RPCService represents RPC service API for ipip module.
type RPCService interface {
	DumpIpipTunnel(ctx context.Context, in *IpipTunnelDump) (RPCService_DumpIpipTunnelClient, error)
	Ipip6rdAddTunnel(ctx context.Context, in *Ipip6rdAddTunnel) (*Ipip6rdAddTunnelReply, error)
	Ipip6rdDelTunnel(ctx context.Context, in *Ipip6rdDelTunnel) (*Ipip6rdDelTunnelReply, error)
	IpipAddTunnel(ctx context.Context, in *IpipAddTunnel) (*IpipAddTunnelReply, error)
	IpipDelTunnel(ctx context.Context, in *IpipDelTunnel) (*IpipDelTunnelReply, error)
}

type serviceClient struct {
	ch api.Channel
}

func NewServiceClient(ch api.Channel) RPCService {
	return &serviceClient{ch}
}

func (c *serviceClient) DumpIpipTunnel(ctx context.Context, in *IpipTunnelDump) (RPCService_DumpIpipTunnelClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpipTunnelClient{stream}
	return x, nil
}

type RPCService_DumpIpipTunnelClient interface {
	Recv() (*IpipTunnelDetails, error)
}

type serviceClient_DumpIpipTunnelClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpipTunnelClient) Recv() (*IpipTunnelDetails, error) {
	m := new(IpipTunnelDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) Ipip6rdAddTunnel(ctx context.Context, in *Ipip6rdAddTunnel) (*Ipip6rdAddTunnelReply, error) {
	out := new(Ipip6rdAddTunnelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) Ipip6rdDelTunnel(ctx context.Context, in *Ipip6rdDelTunnel) (*Ipip6rdDelTunnelReply, error) {
	out := new(Ipip6rdDelTunnelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpipAddTunnel(ctx context.Context, in *IpipAddTunnel) (*IpipAddTunnelReply, error) {
	out := new(IpipAddTunnelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpipDelTunnel(ctx context.Context, in *IpipDelTunnel) (*IpipDelTunnelReply, error) {
	out := new(IpipDelTunnelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion1 // please upgrade the GoVPP api package

// Reference imports to suppress errors if they are not otherwise used.
var _ = api.RegisterMessage
var _ = bytes.NewBuffer
var _ = context.Background
var _ = io.Copy
var _ = strconv.Itoa
var _ = struc.Pack
