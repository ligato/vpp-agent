// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/vxlan.api.json

/*
Package vxlan is a generated VPP binary API for 'vxlan' module.

It consists of:
	 10 enums
	  6 aliases
	  6 types
	  1 union
	  8 messages
	  4 services
*/
package vxlan

import (
	bytes "bytes"
	context "context"
	api "git.fd.io/govpp.git/api"
	struc "github.com/lunixbochs/struc"
	io "io"
	strconv "strconv"

	interface_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/interface_types"
	ip_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ip_types"
)

const (
	// ModuleName is the name of this module.
	ModuleName = "vxlan"
	// APIVersion is the API version of this module.
	APIVersion = "2.0.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0xf11ad29f
)

type AddressFamily = ip_types.AddressFamily

type IfStatusFlags = interface_types.IfStatusFlags

type IfType = interface_types.IfType

type IPDscp = ip_types.IPDscp

type IPEcn = ip_types.IPEcn

type IPProto = ip_types.IPProto

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

type Mprefix = ip_types.Mprefix

type Prefix = ip_types.Prefix

type PrefixMatcher = ip_types.PrefixMatcher

type AddressUnion = ip_types.AddressUnion

// SwInterfaceSetVxlanBypass represents VPP binary API message 'sw_interface_set_vxlan_bypass'.
type SwInterfaceSetVxlanBypass struct {
	SwIfIndex InterfaceIndex
	IsIPv6    bool
	Enable    bool
}

func (m *SwInterfaceSetVxlanBypass) Reset()                        { *m = SwInterfaceSetVxlanBypass{} }
func (*SwInterfaceSetVxlanBypass) GetMessageName() string          { return "sw_interface_set_vxlan_bypass" }
func (*SwInterfaceSetVxlanBypass) GetCrcString() string            { return "65247409" }
func (*SwInterfaceSetVxlanBypass) GetMessageType() api.MessageType { return api.RequestMessage }

// SwInterfaceSetVxlanBypassReply represents VPP binary API message 'sw_interface_set_vxlan_bypass_reply'.
type SwInterfaceSetVxlanBypassReply struct {
	Retval int32
}

func (m *SwInterfaceSetVxlanBypassReply) Reset() { *m = SwInterfaceSetVxlanBypassReply{} }
func (*SwInterfaceSetVxlanBypassReply) GetMessageName() string {
	return "sw_interface_set_vxlan_bypass_reply"
}
func (*SwInterfaceSetVxlanBypassReply) GetCrcString() string            { return "e8d4e804" }
func (*SwInterfaceSetVxlanBypassReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// VxlanAddDelTunnel represents VPP binary API message 'vxlan_add_del_tunnel'.
type VxlanAddDelTunnel struct {
	IsAdd          bool
	Instance       uint32
	SrcAddress     Address
	DstAddress     Address
	McastSwIfIndex InterfaceIndex
	EncapVrfID     uint32
	DecapNextIndex uint32
	Vni            uint32
}

func (m *VxlanAddDelTunnel) Reset()                        { *m = VxlanAddDelTunnel{} }
func (*VxlanAddDelTunnel) GetMessageName() string          { return "vxlan_add_del_tunnel" }
func (*VxlanAddDelTunnel) GetCrcString() string            { return "a35dc8f5" }
func (*VxlanAddDelTunnel) GetMessageType() api.MessageType { return api.RequestMessage }

// VxlanAddDelTunnelReply represents VPP binary API message 'vxlan_add_del_tunnel_reply'.
type VxlanAddDelTunnelReply struct {
	Retval    int32
	SwIfIndex InterfaceIndex
}

func (m *VxlanAddDelTunnelReply) Reset()                        { *m = VxlanAddDelTunnelReply{} }
func (*VxlanAddDelTunnelReply) GetMessageName() string          { return "vxlan_add_del_tunnel_reply" }
func (*VxlanAddDelTunnelReply) GetCrcString() string            { return "5383d31f" }
func (*VxlanAddDelTunnelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// VxlanOffloadRx represents VPP binary API message 'vxlan_offload_rx'.
type VxlanOffloadRx struct {
	HwIfIndex InterfaceIndex
	SwIfIndex InterfaceIndex
	Enable    bool
}

func (m *VxlanOffloadRx) Reset()                        { *m = VxlanOffloadRx{} }
func (*VxlanOffloadRx) GetMessageName() string          { return "vxlan_offload_rx" }
func (*VxlanOffloadRx) GetCrcString() string            { return "89a1564b" }
func (*VxlanOffloadRx) GetMessageType() api.MessageType { return api.RequestMessage }

// VxlanOffloadRxReply represents VPP binary API message 'vxlan_offload_rx_reply'.
type VxlanOffloadRxReply struct {
	Retval int32
}

func (m *VxlanOffloadRxReply) Reset()                        { *m = VxlanOffloadRxReply{} }
func (*VxlanOffloadRxReply) GetMessageName() string          { return "vxlan_offload_rx_reply" }
func (*VxlanOffloadRxReply) GetCrcString() string            { return "e8d4e804" }
func (*VxlanOffloadRxReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// VxlanTunnelDetails represents VPP binary API message 'vxlan_tunnel_details'.
type VxlanTunnelDetails struct {
	SwIfIndex      InterfaceIndex
	Instance       uint32
	SrcAddress     Address
	DstAddress     Address
	McastSwIfIndex InterfaceIndex
	EncapVrfID     uint32
	DecapNextIndex uint32
	Vni            uint32
}

func (m *VxlanTunnelDetails) Reset()                        { *m = VxlanTunnelDetails{} }
func (*VxlanTunnelDetails) GetMessageName() string          { return "vxlan_tunnel_details" }
func (*VxlanTunnelDetails) GetCrcString() string            { return "e782f70f" }
func (*VxlanTunnelDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// VxlanTunnelDump represents VPP binary API message 'vxlan_tunnel_dump'.
type VxlanTunnelDump struct {
	SwIfIndex InterfaceIndex
}

func (m *VxlanTunnelDump) Reset()                        { *m = VxlanTunnelDump{} }
func (*VxlanTunnelDump) GetMessageName() string          { return "vxlan_tunnel_dump" }
func (*VxlanTunnelDump) GetCrcString() string            { return "f9e6675e" }
func (*VxlanTunnelDump) GetMessageType() api.MessageType { return api.RequestMessage }

func init() {
	api.RegisterMessage((*SwInterfaceSetVxlanBypass)(nil), "vxlan.SwInterfaceSetVxlanBypass")
	api.RegisterMessage((*SwInterfaceSetVxlanBypassReply)(nil), "vxlan.SwInterfaceSetVxlanBypassReply")
	api.RegisterMessage((*VxlanAddDelTunnel)(nil), "vxlan.VxlanAddDelTunnel")
	api.RegisterMessage((*VxlanAddDelTunnelReply)(nil), "vxlan.VxlanAddDelTunnelReply")
	api.RegisterMessage((*VxlanOffloadRx)(nil), "vxlan.VxlanOffloadRx")
	api.RegisterMessage((*VxlanOffloadRxReply)(nil), "vxlan.VxlanOffloadRxReply")
	api.RegisterMessage((*VxlanTunnelDetails)(nil), "vxlan.VxlanTunnelDetails")
	api.RegisterMessage((*VxlanTunnelDump)(nil), "vxlan.VxlanTunnelDump")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*SwInterfaceSetVxlanBypass)(nil),
		(*SwInterfaceSetVxlanBypassReply)(nil),
		(*VxlanAddDelTunnel)(nil),
		(*VxlanAddDelTunnelReply)(nil),
		(*VxlanOffloadRx)(nil),
		(*VxlanOffloadRxReply)(nil),
		(*VxlanTunnelDetails)(nil),
		(*VxlanTunnelDump)(nil),
	}
}

// RPCService represents RPC service API for vxlan module.
type RPCService interface {
	DumpVxlanTunnel(ctx context.Context, in *VxlanTunnelDump) (RPCService_DumpVxlanTunnelClient, error)
	SwInterfaceSetVxlanBypass(ctx context.Context, in *SwInterfaceSetVxlanBypass) (*SwInterfaceSetVxlanBypassReply, error)
	VxlanAddDelTunnel(ctx context.Context, in *VxlanAddDelTunnel) (*VxlanAddDelTunnelReply, error)
	VxlanOffloadRx(ctx context.Context, in *VxlanOffloadRx) (*VxlanOffloadRxReply, error)
}

type serviceClient struct {
	ch api.Channel
}

func NewServiceClient(ch api.Channel) RPCService {
	return &serviceClient{ch}
}

func (c *serviceClient) DumpVxlanTunnel(ctx context.Context, in *VxlanTunnelDump) (RPCService_DumpVxlanTunnelClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpVxlanTunnelClient{stream}
	return x, nil
}

type RPCService_DumpVxlanTunnelClient interface {
	Recv() (*VxlanTunnelDetails, error)
}

type serviceClient_DumpVxlanTunnelClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpVxlanTunnelClient) Recv() (*VxlanTunnelDetails, error) {
	m := new(VxlanTunnelDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) SwInterfaceSetVxlanBypass(ctx context.Context, in *SwInterfaceSetVxlanBypass) (*SwInterfaceSetVxlanBypassReply, error) {
	out := new(SwInterfaceSetVxlanBypassReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) VxlanAddDelTunnel(ctx context.Context, in *VxlanAddDelTunnel) (*VxlanAddDelTunnelReply, error) {
	out := new(VxlanAddDelTunnelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) VxlanOffloadRx(ctx context.Context, in *VxlanOffloadRx) (*VxlanOffloadRxReply, error) {
	out := new(VxlanOffloadRxReply)
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
