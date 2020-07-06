// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/ip6_nd.api.json

/*
Package ip6_nd is a generated VPP binary API for 'ip6_nd' module.

It consists of:
	 10 enums
	  6 aliases
	  7 types
	  1 union
	 13 messages
	  6 services
*/
package ip6_nd

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
	ModuleName = "ip6_nd"
	// APIVersion is the API version of this module.
	APIVersion = "1.0.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0xbb8ff0e9
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

// IP6RaPrefixInfo represents VPP binary API type 'ip6_ra_prefix_info'.
type IP6RaPrefixInfo struct {
	Prefix        Prefix
	Flags         uint8
	ValidTime     uint32
	PreferredTime uint32
}

func (*IP6RaPrefixInfo) GetTypeName() string { return "ip6_ra_prefix_info" }

type Mprefix = ip_types.Mprefix

type Prefix = ip_types.Prefix

type PrefixMatcher = ip_types.PrefixMatcher

type AddressUnion = ip_types.AddressUnion

// IP6RaEvent represents VPP binary API message 'ip6_ra_event'.
type IP6RaEvent struct {
	PID                                                 uint32
	SwIfIndex                                           InterfaceIndex
	RouterAddr                                          IP6Address
	CurrentHopLimit                                     uint8
	Flags                                               uint8
	RouterLifetimeInSec                                 uint16
	NeighborReachableTimeInMsec                         uint32
	TimeInMsecBetweenRetransmittedNeighborSolicitations uint32
	NPrefixes                                           uint32 `struc:"sizeof=Prefixes"`
	Prefixes                                            []IP6RaPrefixInfo
}

func (m *IP6RaEvent) Reset()                        { *m = IP6RaEvent{} }
func (*IP6RaEvent) GetMessageName() string          { return "ip6_ra_event" }
func (*IP6RaEvent) GetCrcString() string            { return "47e8cfbe" }
func (*IP6RaEvent) GetMessageType() api.MessageType { return api.EventMessage }

// IP6ndProxyAddDel represents VPP binary API message 'ip6nd_proxy_add_del'.
type IP6ndProxyAddDel struct {
	SwIfIndex InterfaceIndex
	IsAdd     bool
	IP        IP6Address
}

func (m *IP6ndProxyAddDel) Reset()                        { *m = IP6ndProxyAddDel{} }
func (*IP6ndProxyAddDel) GetMessageName() string          { return "ip6nd_proxy_add_del" }
func (*IP6ndProxyAddDel) GetCrcString() string            { return "3fdf6659" }
func (*IP6ndProxyAddDel) GetMessageType() api.MessageType { return api.RequestMessage }

// IP6ndProxyAddDelReply represents VPP binary API message 'ip6nd_proxy_add_del_reply'.
type IP6ndProxyAddDelReply struct {
	Retval int32
}

func (m *IP6ndProxyAddDelReply) Reset()                        { *m = IP6ndProxyAddDelReply{} }
func (*IP6ndProxyAddDelReply) GetMessageName() string          { return "ip6nd_proxy_add_del_reply" }
func (*IP6ndProxyAddDelReply) GetCrcString() string            { return "e8d4e804" }
func (*IP6ndProxyAddDelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// IP6ndProxyDetails represents VPP binary API message 'ip6nd_proxy_details'.
type IP6ndProxyDetails struct {
	SwIfIndex InterfaceIndex
	IP        IP6Address
}

func (m *IP6ndProxyDetails) Reset()                        { *m = IP6ndProxyDetails{} }
func (*IP6ndProxyDetails) GetMessageName() string          { return "ip6nd_proxy_details" }
func (*IP6ndProxyDetails) GetCrcString() string            { return "d35be8ff" }
func (*IP6ndProxyDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// IP6ndProxyDump represents VPP binary API message 'ip6nd_proxy_dump'.
type IP6ndProxyDump struct{}

func (m *IP6ndProxyDump) Reset()                        { *m = IP6ndProxyDump{} }
func (*IP6ndProxyDump) GetMessageName() string          { return "ip6nd_proxy_dump" }
func (*IP6ndProxyDump) GetCrcString() string            { return "51077d14" }
func (*IP6ndProxyDump) GetMessageType() api.MessageType { return api.RequestMessage }

// IP6ndSendRouterSolicitation represents VPP binary API message 'ip6nd_send_router_solicitation'.
type IP6ndSendRouterSolicitation struct {
	Irt       uint32
	Mrt       uint32
	Mrc       uint32
	Mrd       uint32
	SwIfIndex InterfaceIndex
	Stop      bool
}

func (m *IP6ndSendRouterSolicitation) Reset()                        { *m = IP6ndSendRouterSolicitation{} }
func (*IP6ndSendRouterSolicitation) GetMessageName() string          { return "ip6nd_send_router_solicitation" }
func (*IP6ndSendRouterSolicitation) GetCrcString() string            { return "e5de609c" }
func (*IP6ndSendRouterSolicitation) GetMessageType() api.MessageType { return api.RequestMessage }

// IP6ndSendRouterSolicitationReply represents VPP binary API message 'ip6nd_send_router_solicitation_reply'.
type IP6ndSendRouterSolicitationReply struct {
	Retval int32
}

func (m *IP6ndSendRouterSolicitationReply) Reset() { *m = IP6ndSendRouterSolicitationReply{} }
func (*IP6ndSendRouterSolicitationReply) GetMessageName() string {
	return "ip6nd_send_router_solicitation_reply"
}
func (*IP6ndSendRouterSolicitationReply) GetCrcString() string            { return "e8d4e804" }
func (*IP6ndSendRouterSolicitationReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// SwInterfaceIP6ndRaConfig represents VPP binary API message 'sw_interface_ip6nd_ra_config'.
type SwInterfaceIP6ndRaConfig struct {
	SwIfIndex       InterfaceIndex
	Suppress        uint8
	Managed         uint8
	Other           uint8
	LlOption        uint8
	SendUnicast     uint8
	Cease           uint8
	IsNo            bool
	DefaultRouter   uint8
	MaxInterval     uint32
	MinInterval     uint32
	Lifetime        uint32
	InitialCount    uint32
	InitialInterval uint32
}

func (m *SwInterfaceIP6ndRaConfig) Reset()                        { *m = SwInterfaceIP6ndRaConfig{} }
func (*SwInterfaceIP6ndRaConfig) GetMessageName() string          { return "sw_interface_ip6nd_ra_config" }
func (*SwInterfaceIP6ndRaConfig) GetCrcString() string            { return "3eb00b1c" }
func (*SwInterfaceIP6ndRaConfig) GetMessageType() api.MessageType { return api.RequestMessage }

// SwInterfaceIP6ndRaConfigReply represents VPP binary API message 'sw_interface_ip6nd_ra_config_reply'.
type SwInterfaceIP6ndRaConfigReply struct {
	Retval int32
}

func (m *SwInterfaceIP6ndRaConfigReply) Reset() { *m = SwInterfaceIP6ndRaConfigReply{} }
func (*SwInterfaceIP6ndRaConfigReply) GetMessageName() string {
	return "sw_interface_ip6nd_ra_config_reply"
}
func (*SwInterfaceIP6ndRaConfigReply) GetCrcString() string            { return "e8d4e804" }
func (*SwInterfaceIP6ndRaConfigReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// SwInterfaceIP6ndRaPrefix represents VPP binary API message 'sw_interface_ip6nd_ra_prefix'.
type SwInterfaceIP6ndRaPrefix struct {
	SwIfIndex    InterfaceIndex
	Prefix       Prefix
	UseDefault   bool
	NoAdvertise  bool
	OffLink      bool
	NoAutoconfig bool
	NoOnlink     bool
	IsNo         bool
	ValLifetime  uint32
	PrefLifetime uint32
}

func (m *SwInterfaceIP6ndRaPrefix) Reset()                        { *m = SwInterfaceIP6ndRaPrefix{} }
func (*SwInterfaceIP6ndRaPrefix) GetMessageName() string          { return "sw_interface_ip6nd_ra_prefix" }
func (*SwInterfaceIP6ndRaPrefix) GetCrcString() string            { return "e098785f" }
func (*SwInterfaceIP6ndRaPrefix) GetMessageType() api.MessageType { return api.RequestMessage }

// SwInterfaceIP6ndRaPrefixReply represents VPP binary API message 'sw_interface_ip6nd_ra_prefix_reply'.
type SwInterfaceIP6ndRaPrefixReply struct {
	Retval int32
}

func (m *SwInterfaceIP6ndRaPrefixReply) Reset() { *m = SwInterfaceIP6ndRaPrefixReply{} }
func (*SwInterfaceIP6ndRaPrefixReply) GetMessageName() string {
	return "sw_interface_ip6nd_ra_prefix_reply"
}
func (*SwInterfaceIP6ndRaPrefixReply) GetCrcString() string            { return "e8d4e804" }
func (*SwInterfaceIP6ndRaPrefixReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// WantIP6RaEvents represents VPP binary API message 'want_ip6_ra_events'.
type WantIP6RaEvents struct {
	Enable bool
	PID    uint32
}

func (m *WantIP6RaEvents) Reset()                        { *m = WantIP6RaEvents{} }
func (*WantIP6RaEvents) GetMessageName() string          { return "want_ip6_ra_events" }
func (*WantIP6RaEvents) GetCrcString() string            { return "3ec6d6c2" }
func (*WantIP6RaEvents) GetMessageType() api.MessageType { return api.RequestMessage }

// WantIP6RaEventsReply represents VPP binary API message 'want_ip6_ra_events_reply'.
type WantIP6RaEventsReply struct {
	Retval int32
}

func (m *WantIP6RaEventsReply) Reset()                        { *m = WantIP6RaEventsReply{} }
func (*WantIP6RaEventsReply) GetMessageName() string          { return "want_ip6_ra_events_reply" }
func (*WantIP6RaEventsReply) GetCrcString() string            { return "e8d4e804" }
func (*WantIP6RaEventsReply) GetMessageType() api.MessageType { return api.ReplyMessage }

func init() {
	api.RegisterMessage((*IP6RaEvent)(nil), "ip6_nd.IP6RaEvent")
	api.RegisterMessage((*IP6ndProxyAddDel)(nil), "ip6_nd.IP6ndProxyAddDel")
	api.RegisterMessage((*IP6ndProxyAddDelReply)(nil), "ip6_nd.IP6ndProxyAddDelReply")
	api.RegisterMessage((*IP6ndProxyDetails)(nil), "ip6_nd.IP6ndProxyDetails")
	api.RegisterMessage((*IP6ndProxyDump)(nil), "ip6_nd.IP6ndProxyDump")
	api.RegisterMessage((*IP6ndSendRouterSolicitation)(nil), "ip6_nd.IP6ndSendRouterSolicitation")
	api.RegisterMessage((*IP6ndSendRouterSolicitationReply)(nil), "ip6_nd.IP6ndSendRouterSolicitationReply")
	api.RegisterMessage((*SwInterfaceIP6ndRaConfig)(nil), "ip6_nd.SwInterfaceIP6ndRaConfig")
	api.RegisterMessage((*SwInterfaceIP6ndRaConfigReply)(nil), "ip6_nd.SwInterfaceIP6ndRaConfigReply")
	api.RegisterMessage((*SwInterfaceIP6ndRaPrefix)(nil), "ip6_nd.SwInterfaceIP6ndRaPrefix")
	api.RegisterMessage((*SwInterfaceIP6ndRaPrefixReply)(nil), "ip6_nd.SwInterfaceIP6ndRaPrefixReply")
	api.RegisterMessage((*WantIP6RaEvents)(nil), "ip6_nd.WantIP6RaEvents")
	api.RegisterMessage((*WantIP6RaEventsReply)(nil), "ip6_nd.WantIP6RaEventsReply")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*IP6RaEvent)(nil),
		(*IP6ndProxyAddDel)(nil),
		(*IP6ndProxyAddDelReply)(nil),
		(*IP6ndProxyDetails)(nil),
		(*IP6ndProxyDump)(nil),
		(*IP6ndSendRouterSolicitation)(nil),
		(*IP6ndSendRouterSolicitationReply)(nil),
		(*SwInterfaceIP6ndRaConfig)(nil),
		(*SwInterfaceIP6ndRaConfigReply)(nil),
		(*SwInterfaceIP6ndRaPrefix)(nil),
		(*SwInterfaceIP6ndRaPrefixReply)(nil),
		(*WantIP6RaEvents)(nil),
		(*WantIP6RaEventsReply)(nil),
	}
}

// RPCService represents RPC service API for ip6_nd module.
type RPCService interface {
	DumpIP6ndProxy(ctx context.Context, in *IP6ndProxyDump) (RPCService_DumpIP6ndProxyClient, error)
	IP6ndProxyAddDel(ctx context.Context, in *IP6ndProxyAddDel) (*IP6ndProxyAddDelReply, error)
	IP6ndSendRouterSolicitation(ctx context.Context, in *IP6ndSendRouterSolicitation) (*IP6ndSendRouterSolicitationReply, error)
	SwInterfaceIP6ndRaConfig(ctx context.Context, in *SwInterfaceIP6ndRaConfig) (*SwInterfaceIP6ndRaConfigReply, error)
	SwInterfaceIP6ndRaPrefix(ctx context.Context, in *SwInterfaceIP6ndRaPrefix) (*SwInterfaceIP6ndRaPrefixReply, error)
	WantIP6RaEvents(ctx context.Context, in *WantIP6RaEvents) (*WantIP6RaEventsReply, error)
}

type serviceClient struct {
	ch api.Channel
}

func NewServiceClient(ch api.Channel) RPCService {
	return &serviceClient{ch}
}

func (c *serviceClient) DumpIP6ndProxy(ctx context.Context, in *IP6ndProxyDump) (RPCService_DumpIP6ndProxyClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIP6ndProxyClient{stream}
	return x, nil
}

type RPCService_DumpIP6ndProxyClient interface {
	Recv() (*IP6ndProxyDetails, error)
}

type serviceClient_DumpIP6ndProxyClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIP6ndProxyClient) Recv() (*IP6ndProxyDetails, error) {
	m := new(IP6ndProxyDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) IP6ndProxyAddDel(ctx context.Context, in *IP6ndProxyAddDel) (*IP6ndProxyAddDelReply, error) {
	out := new(IP6ndProxyAddDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IP6ndSendRouterSolicitation(ctx context.Context, in *IP6ndSendRouterSolicitation) (*IP6ndSendRouterSolicitationReply, error) {
	out := new(IP6ndSendRouterSolicitationReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SwInterfaceIP6ndRaConfig(ctx context.Context, in *SwInterfaceIP6ndRaConfig) (*SwInterfaceIP6ndRaConfigReply, error) {
	out := new(SwInterfaceIP6ndRaConfigReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SwInterfaceIP6ndRaPrefix(ctx context.Context, in *SwInterfaceIP6ndRaPrefix) (*SwInterfaceIP6ndRaPrefixReply, error) {
	out := new(SwInterfaceIP6ndRaPrefixReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) WantIP6RaEvents(ctx context.Context, in *WantIP6RaEvents) (*WantIP6RaEventsReply, error) {
	out := new(WantIP6RaEventsReply)
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
