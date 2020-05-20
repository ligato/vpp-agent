// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/vxlan_gpe.api.json

/*
Package vxlan_gpe is a generated VPP binary API for 'vxlan_gpe' module.

It consists of:
	 10 enums
	  6 aliases
	  6 types
	  1 union
	  6 messages
	  3 services
*/
package vxlan_gpe

import (
	"bytes"
	"context"
	"io"
	"strconv"

	api "git.fd.io/govpp.git/api"
	struc "github.com/lunixbochs/struc"

	interface_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/interface_types"
	ip_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ip_types"
)

const (
	// ModuleName is the name of this module.
	ModuleName = "vxlan_gpe"
	// APIVersion is the API version of this module.
	APIVersion = "2.0.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0x8295f28f
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

// SwInterfaceSetVxlanGpeBypass represents VPP binary API message 'sw_interface_set_vxlan_gpe_bypass'.
type SwInterfaceSetVxlanGpeBypass struct {
	SwIfIndex InterfaceIndex
	IsIPv6    bool
	Enable    bool
}

func (m *SwInterfaceSetVxlanGpeBypass) Reset() { *m = SwInterfaceSetVxlanGpeBypass{} }
func (*SwInterfaceSetVxlanGpeBypass) GetMessageName() string {
	return "sw_interface_set_vxlan_gpe_bypass"
}
func (*SwInterfaceSetVxlanGpeBypass) GetCrcString() string            { return "65247409" }
func (*SwInterfaceSetVxlanGpeBypass) GetMessageType() api.MessageType { return api.RequestMessage }

// SwInterfaceSetVxlanGpeBypassReply represents VPP binary API message 'sw_interface_set_vxlan_gpe_bypass_reply'.
type SwInterfaceSetVxlanGpeBypassReply struct {
	Retval int32
}

func (m *SwInterfaceSetVxlanGpeBypassReply) Reset() { *m = SwInterfaceSetVxlanGpeBypassReply{} }
func (*SwInterfaceSetVxlanGpeBypassReply) GetMessageName() string {
	return "sw_interface_set_vxlan_gpe_bypass_reply"
}
func (*SwInterfaceSetVxlanGpeBypassReply) GetCrcString() string            { return "e8d4e804" }
func (*SwInterfaceSetVxlanGpeBypassReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// VxlanGpeAddDelTunnel represents VPP binary API message 'vxlan_gpe_add_del_tunnel'.
type VxlanGpeAddDelTunnel struct {
	Local          Address
	Remote         Address
	McastSwIfIndex InterfaceIndex
	EncapVrfID     uint32
	DecapVrfID     uint32
	Protocol       IPProto
	Vni            uint32
	IsAdd          bool
}

func (m *VxlanGpeAddDelTunnel) Reset()                        { *m = VxlanGpeAddDelTunnel{} }
func (*VxlanGpeAddDelTunnel) GetMessageName() string          { return "vxlan_gpe_add_del_tunnel" }
func (*VxlanGpeAddDelTunnel) GetCrcString() string            { return "7c6da6ae" }
func (*VxlanGpeAddDelTunnel) GetMessageType() api.MessageType { return api.RequestMessage }

// VxlanGpeAddDelTunnelReply represents VPP binary API message 'vxlan_gpe_add_del_tunnel_reply'.
type VxlanGpeAddDelTunnelReply struct {
	Retval    int32
	SwIfIndex InterfaceIndex
}

func (m *VxlanGpeAddDelTunnelReply) Reset()                        { *m = VxlanGpeAddDelTunnelReply{} }
func (*VxlanGpeAddDelTunnelReply) GetMessageName() string          { return "vxlan_gpe_add_del_tunnel_reply" }
func (*VxlanGpeAddDelTunnelReply) GetCrcString() string            { return "5383d31f" }
func (*VxlanGpeAddDelTunnelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// VxlanGpeTunnelDetails represents VPP binary API message 'vxlan_gpe_tunnel_details'.
type VxlanGpeTunnelDetails struct {
	SwIfIndex      InterfaceIndex
	Local          Address
	Remote         Address
	Vni            uint32
	Protocol       IPProto
	McastSwIfIndex InterfaceIndex
	EncapVrfID     uint32
	DecapVrfID     uint32
	IsIPv6         bool
}

func (m *VxlanGpeTunnelDetails) Reset()                        { *m = VxlanGpeTunnelDetails{} }
func (*VxlanGpeTunnelDetails) GetMessageName() string          { return "vxlan_gpe_tunnel_details" }
func (*VxlanGpeTunnelDetails) GetCrcString() string            { return "57712346" }
func (*VxlanGpeTunnelDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// VxlanGpeTunnelDump represents VPP binary API message 'vxlan_gpe_tunnel_dump'.
type VxlanGpeTunnelDump struct {
	SwIfIndex InterfaceIndex
}

func (m *VxlanGpeTunnelDump) Reset()                        { *m = VxlanGpeTunnelDump{} }
func (*VxlanGpeTunnelDump) GetMessageName() string          { return "vxlan_gpe_tunnel_dump" }
func (*VxlanGpeTunnelDump) GetCrcString() string            { return "f9e6675e" }
func (*VxlanGpeTunnelDump) GetMessageType() api.MessageType { return api.RequestMessage }

func init() {
	api.RegisterMessage((*SwInterfaceSetVxlanGpeBypass)(nil), "vxlan_gpe.SwInterfaceSetVxlanGpeBypass")
	api.RegisterMessage((*SwInterfaceSetVxlanGpeBypassReply)(nil), "vxlan_gpe.SwInterfaceSetVxlanGpeBypassReply")
	api.RegisterMessage((*VxlanGpeAddDelTunnel)(nil), "vxlan_gpe.VxlanGpeAddDelTunnel")
	api.RegisterMessage((*VxlanGpeAddDelTunnelReply)(nil), "vxlan_gpe.VxlanGpeAddDelTunnelReply")
	api.RegisterMessage((*VxlanGpeTunnelDetails)(nil), "vxlan_gpe.VxlanGpeTunnelDetails")
	api.RegisterMessage((*VxlanGpeTunnelDump)(nil), "vxlan_gpe.VxlanGpeTunnelDump")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*SwInterfaceSetVxlanGpeBypass)(nil),
		(*SwInterfaceSetVxlanGpeBypassReply)(nil),
		(*VxlanGpeAddDelTunnel)(nil),
		(*VxlanGpeAddDelTunnelReply)(nil),
		(*VxlanGpeTunnelDetails)(nil),
		(*VxlanGpeTunnelDump)(nil),
	}
}

// RPCService represents RPC service API for vxlan_gpe module.
type RPCService interface {
	DumpVxlanGpeTunnel(ctx context.Context, in *VxlanGpeTunnelDump) (RPCService_DumpVxlanGpeTunnelClient, error)
	SwInterfaceSetVxlanGpeBypass(ctx context.Context, in *SwInterfaceSetVxlanGpeBypass) (*SwInterfaceSetVxlanGpeBypassReply, error)
	VxlanGpeAddDelTunnel(ctx context.Context, in *VxlanGpeAddDelTunnel) (*VxlanGpeAddDelTunnelReply, error)
}

type serviceClient struct {
	ch api.Channel
}

func NewServiceClient(ch api.Channel) RPCService {
	return &serviceClient{ch}
}

func (c *serviceClient) DumpVxlanGpeTunnel(ctx context.Context, in *VxlanGpeTunnelDump) (RPCService_DumpVxlanGpeTunnelClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpVxlanGpeTunnelClient{stream}
	return x, nil
}

type RPCService_DumpVxlanGpeTunnelClient interface {
	Recv() (*VxlanGpeTunnelDetails, error)
}

type serviceClient_DumpVxlanGpeTunnelClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpVxlanGpeTunnelClient) Recv() (*VxlanGpeTunnelDetails, error) {
	m := new(VxlanGpeTunnelDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) SwInterfaceSetVxlanGpeBypass(ctx context.Context, in *SwInterfaceSetVxlanGpeBypass) (*SwInterfaceSetVxlanGpeBypassReply, error) {
	out := new(SwInterfaceSetVxlanGpeBypassReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) VxlanGpeAddDelTunnel(ctx context.Context, in *VxlanGpeAddDelTunnel) (*VxlanGpeAddDelTunnelReply, error) {
	out := new(VxlanGpeAddDelTunnelReply)
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
