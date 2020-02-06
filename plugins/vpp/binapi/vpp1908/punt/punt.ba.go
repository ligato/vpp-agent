// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/punt.api.json

/*
Package punt is a generated VPP binary API for 'punt' module.

It consists of:
	  5 enums
	  5 aliases
	 11 types
	  2 unions
	 10 messages
	  5 services
*/
package punt

import (
	bytes "bytes"
	context "context"
	api "git.fd.io/govpp.git/api"
	struc "github.com/lunixbochs/struc"
	io "io"
	strconv "strconv"
)

const (
	// ModuleName is the name of this module.
	ModuleName = "punt"
	// APIVersion is the API version of this module.
	APIVersion = "2.2.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0x4aa3929b
)

// AddressFamily represents VPP binary API enum 'address_family'.
type AddressFamily uint32

const (
	ADDRESS_IP4 AddressFamily = 0
	ADDRESS_IP6 AddressFamily = 1
)

var AddressFamily_name = map[uint32]string{
	0: "ADDRESS_IP4",
	1: "ADDRESS_IP6",
}

var AddressFamily_value = map[string]uint32{
	"ADDRESS_IP4": 0,
	"ADDRESS_IP6": 1,
}

func (x AddressFamily) String() string {
	s, ok := AddressFamily_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IPDscp represents VPP binary API enum 'ip_dscp'.
type IPDscp uint8

const (
	IP_API_DSCP_CS0  IPDscp = 0
	IP_API_DSCP_CS1  IPDscp = 8
	IP_API_DSCP_AF11 IPDscp = 10
	IP_API_DSCP_AF12 IPDscp = 12
	IP_API_DSCP_AF13 IPDscp = 14
	IP_API_DSCP_CS2  IPDscp = 16
	IP_API_DSCP_AF21 IPDscp = 18
	IP_API_DSCP_AF22 IPDscp = 20
	IP_API_DSCP_AF23 IPDscp = 22
	IP_API_DSCP_CS3  IPDscp = 24
	IP_API_DSCP_AF31 IPDscp = 26
	IP_API_DSCP_AF32 IPDscp = 28
	IP_API_DSCP_AF33 IPDscp = 30
	IP_API_DSCP_CS4  IPDscp = 32
	IP_API_DSCP_AF41 IPDscp = 34
	IP_API_DSCP_AF42 IPDscp = 36
	IP_API_DSCP_AF43 IPDscp = 38
	IP_API_DSCP_CS5  IPDscp = 40
	IP_API_DSCP_EF   IPDscp = 46
	IP_API_DSCP_CS6  IPDscp = 48
	IP_API_DSCP_CS7  IPDscp = 50
)

var IPDscp_name = map[uint8]string{
	0:  "IP_API_DSCP_CS0",
	8:  "IP_API_DSCP_CS1",
	10: "IP_API_DSCP_AF11",
	12: "IP_API_DSCP_AF12",
	14: "IP_API_DSCP_AF13",
	16: "IP_API_DSCP_CS2",
	18: "IP_API_DSCP_AF21",
	20: "IP_API_DSCP_AF22",
	22: "IP_API_DSCP_AF23",
	24: "IP_API_DSCP_CS3",
	26: "IP_API_DSCP_AF31",
	28: "IP_API_DSCP_AF32",
	30: "IP_API_DSCP_AF33",
	32: "IP_API_DSCP_CS4",
	34: "IP_API_DSCP_AF41",
	36: "IP_API_DSCP_AF42",
	38: "IP_API_DSCP_AF43",
	40: "IP_API_DSCP_CS5",
	46: "IP_API_DSCP_EF",
	48: "IP_API_DSCP_CS6",
	50: "IP_API_DSCP_CS7",
}

var IPDscp_value = map[string]uint8{
	"IP_API_DSCP_CS0":  0,
	"IP_API_DSCP_CS1":  8,
	"IP_API_DSCP_AF11": 10,
	"IP_API_DSCP_AF12": 12,
	"IP_API_DSCP_AF13": 14,
	"IP_API_DSCP_CS2":  16,
	"IP_API_DSCP_AF21": 18,
	"IP_API_DSCP_AF22": 20,
	"IP_API_DSCP_AF23": 22,
	"IP_API_DSCP_CS3":  24,
	"IP_API_DSCP_AF31": 26,
	"IP_API_DSCP_AF32": 28,
	"IP_API_DSCP_AF33": 30,
	"IP_API_DSCP_CS4":  32,
	"IP_API_DSCP_AF41": 34,
	"IP_API_DSCP_AF42": 36,
	"IP_API_DSCP_AF43": 38,
	"IP_API_DSCP_CS5":  40,
	"IP_API_DSCP_EF":   46,
	"IP_API_DSCP_CS6":  48,
	"IP_API_DSCP_CS7":  50,
}

func (x IPDscp) String() string {
	s, ok := IPDscp_name[uint8(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IPEcn represents VPP binary API enum 'ip_ecn'.
type IPEcn uint8

const (
	IP_API_ECN_NONE IPEcn = 0
	IP_API_ECN_ECT0 IPEcn = 1
	IP_API_ECN_ECT1 IPEcn = 2
	IP_API_ECN_CE   IPEcn = 3
)

var IPEcn_name = map[uint8]string{
	0: "IP_API_ECN_NONE",
	1: "IP_API_ECN_ECT0",
	2: "IP_API_ECN_ECT1",
	3: "IP_API_ECN_CE",
}

var IPEcn_value = map[string]uint8{
	"IP_API_ECN_NONE": 0,
	"IP_API_ECN_ECT0": 1,
	"IP_API_ECN_ECT1": 2,
	"IP_API_ECN_CE":   3,
}

func (x IPEcn) String() string {
	s, ok := IPEcn_name[uint8(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IPProto represents VPP binary API enum 'ip_proto'.
type IPProto uint32

const (
	IP_API_PROTO_HOPOPT   IPProto = 0
	IP_API_PROTO_ICMP     IPProto = 1
	IP_API_PROTO_IGMP     IPProto = 2
	IP_API_PROTO_TCP      IPProto = 6
	IP_API_PROTO_UDP      IPProto = 17
	IP_API_PROTO_GRE      IPProto = 47
	IP_API_PROTO_AH       IPProto = 50
	IP_API_PROTO_ESP      IPProto = 51
	IP_API_PROTO_EIGRP    IPProto = 88
	IP_API_PROTO_OSPF     IPProto = 89
	IP_API_PROTO_SCTP     IPProto = 132
	IP_API_PROTO_RESERVED IPProto = 255
)

var IPProto_name = map[uint32]string{
	0:   "IP_API_PROTO_HOPOPT",
	1:   "IP_API_PROTO_ICMP",
	2:   "IP_API_PROTO_IGMP",
	6:   "IP_API_PROTO_TCP",
	17:  "IP_API_PROTO_UDP",
	47:  "IP_API_PROTO_GRE",
	50:  "IP_API_PROTO_AH",
	51:  "IP_API_PROTO_ESP",
	88:  "IP_API_PROTO_EIGRP",
	89:  "IP_API_PROTO_OSPF",
	132: "IP_API_PROTO_SCTP",
	255: "IP_API_PROTO_RESERVED",
}

var IPProto_value = map[string]uint32{
	"IP_API_PROTO_HOPOPT":   0,
	"IP_API_PROTO_ICMP":     1,
	"IP_API_PROTO_IGMP":     2,
	"IP_API_PROTO_TCP":      6,
	"IP_API_PROTO_UDP":      17,
	"IP_API_PROTO_GRE":      47,
	"IP_API_PROTO_AH":       50,
	"IP_API_PROTO_ESP":      51,
	"IP_API_PROTO_EIGRP":    88,
	"IP_API_PROTO_OSPF":     89,
	"IP_API_PROTO_SCTP":     132,
	"IP_API_PROTO_RESERVED": 255,
}

func (x IPProto) String() string {
	s, ok := IPProto_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// PuntType represents VPP binary API enum 'punt_type'.
type PuntType uint32

const (
	PUNT_API_TYPE_L4        PuntType = 1
	PUNT_API_TYPE_IP_PROTO  PuntType = 2
	PUNT_API_TYPE_EXCEPTION PuntType = 3
)

var PuntType_name = map[uint32]string{
	1: "PUNT_API_TYPE_L4",
	2: "PUNT_API_TYPE_IP_PROTO",
	3: "PUNT_API_TYPE_EXCEPTION",
}

var PuntType_value = map[string]uint32{
	"PUNT_API_TYPE_L4":        1,
	"PUNT_API_TYPE_IP_PROTO":  2,
	"PUNT_API_TYPE_EXCEPTION": 3,
}

func (x PuntType) String() string {
	s, ok := PuntType_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// AddressWithPrefix represents VPP binary API alias 'address_with_prefix'.
type AddressWithPrefix Prefix

// IP4Address represents VPP binary API alias 'ip4_address'.
type IP4Address [4]uint8

// IP4AddressWithPrefix represents VPP binary API alias 'ip4_address_with_prefix'.
type IP4AddressWithPrefix IP4Prefix

// IP6Address represents VPP binary API alias 'ip6_address'.
type IP6Address [16]uint8

// IP6AddressWithPrefix represents VPP binary API alias 'ip6_address_with_prefix'.
type IP6AddressWithPrefix IP6Prefix

// Address represents VPP binary API type 'address'.
type Address struct {
	Af AddressFamily
	Un AddressUnion
}

func (*Address) GetTypeName() string { return "address" }

// IP4Prefix represents VPP binary API type 'ip4_prefix'.
type IP4Prefix struct {
	Address IP4Address
	Len     uint8
}

func (*IP4Prefix) GetTypeName() string { return "ip4_prefix" }

// IP6Prefix represents VPP binary API type 'ip6_prefix'.
type IP6Prefix struct {
	Address IP6Address
	Len     uint8
}

func (*IP6Prefix) GetTypeName() string { return "ip6_prefix" }

// Mprefix represents VPP binary API type 'mprefix'.
type Mprefix struct {
	Af               AddressFamily
	GrpAddressLength uint16
	GrpAddress       AddressUnion
	SrcAddress       AddressUnion
}

func (*Mprefix) GetTypeName() string { return "mprefix" }

// Prefix represents VPP binary API type 'prefix'.
type Prefix struct {
	Address Address
	Len     uint8
}

func (*Prefix) GetTypeName() string { return "prefix" }

// PrefixMatcher represents VPP binary API type 'prefix_matcher'.
type PrefixMatcher struct {
	Le uint8
	Ge uint8
}

func (*PrefixMatcher) GetTypeName() string { return "prefix_matcher" }

// Punt represents VPP binary API type 'punt'.
type Punt struct {
	Type PuntType
	Punt PuntUnion
}

func (*Punt) GetTypeName() string { return "punt" }

// PuntException represents VPP binary API type 'punt_exception'.
type PuntException struct {
	ID uint32
}

func (*PuntException) GetTypeName() string { return "punt_exception" }

// PuntIPProto represents VPP binary API type 'punt_ip_proto'.
type PuntIPProto struct {
	Af       AddressFamily
	Protocol IPProto
}

func (*PuntIPProto) GetTypeName() string { return "punt_ip_proto" }

// PuntL4 represents VPP binary API type 'punt_l4'.
type PuntL4 struct {
	Af       AddressFamily
	Protocol IPProto
	Port     uint16
}

func (*PuntL4) GetTypeName() string { return "punt_l4" }

// PuntReason represents VPP binary API type 'punt_reason'.
type PuntReason struct {
	ID          uint32
	XXX_NameLen uint32 `struc:"sizeof=Name"`
	Name        string
}

func (*PuntReason) GetTypeName() string { return "punt_reason" }

// AddressUnion represents VPP binary API union 'address_union'.
type AddressUnion struct {
	XXX_UnionData [16]byte
}

func (*AddressUnion) GetTypeName() string { return "address_union" }

func AddressUnionIP4(a IP4Address) (u AddressUnion) {
	u.SetIP4(a)
	return
}
func (u *AddressUnion) SetIP4(a IP4Address) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.XXX_UnionData[:], b.Bytes())
}
func (u *AddressUnion) GetIP4() (a IP4Address) {
	var b = bytes.NewReader(u.XXX_UnionData[:])
	struc.Unpack(b, &a)
	return
}

func AddressUnionIP6(a IP6Address) (u AddressUnion) {
	u.SetIP6(a)
	return
}
func (u *AddressUnion) SetIP6(a IP6Address) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.XXX_UnionData[:], b.Bytes())
}
func (u *AddressUnion) GetIP6() (a IP6Address) {
	var b = bytes.NewReader(u.XXX_UnionData[:])
	struc.Unpack(b, &a)
	return
}

// PuntUnion represents VPP binary API union 'punt_union'.
type PuntUnion struct {
	XXX_UnionData [10]byte
}

func (*PuntUnion) GetTypeName() string { return "punt_union" }

func PuntUnionException(a PuntException) (u PuntUnion) {
	u.SetException(a)
	return
}
func (u *PuntUnion) SetException(a PuntException) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.XXX_UnionData[:], b.Bytes())
}
func (u *PuntUnion) GetException() (a PuntException) {
	var b = bytes.NewReader(u.XXX_UnionData[:])
	struc.Unpack(b, &a)
	return
}

func PuntUnionL4(a PuntL4) (u PuntUnion) {
	u.SetL4(a)
	return
}
func (u *PuntUnion) SetL4(a PuntL4) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.XXX_UnionData[:], b.Bytes())
}
func (u *PuntUnion) GetL4() (a PuntL4) {
	var b = bytes.NewReader(u.XXX_UnionData[:])
	struc.Unpack(b, &a)
	return
}

func PuntUnionIPProto(a PuntIPProto) (u PuntUnion) {
	u.SetIPProto(a)
	return
}
func (u *PuntUnion) SetIPProto(a PuntIPProto) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.XXX_UnionData[:], b.Bytes())
}
func (u *PuntUnion) GetIPProto() (a PuntIPProto) {
	var b = bytes.NewReader(u.XXX_UnionData[:])
	struc.Unpack(b, &a)
	return
}

// PuntReasonDetails represents VPP binary API message 'punt_reason_details'.
type PuntReasonDetails struct {
	Reason PuntReason
}

func (m *PuntReasonDetails) Reset()                        { *m = PuntReasonDetails{} }
func (*PuntReasonDetails) GetMessageName() string          { return "punt_reason_details" }
func (*PuntReasonDetails) GetCrcString() string            { return "2c9d4a40" }
func (*PuntReasonDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// PuntReasonDump represents VPP binary API message 'punt_reason_dump'.
type PuntReasonDump struct {
	Reason PuntReason
}

func (m *PuntReasonDump) Reset()                        { *m = PuntReasonDump{} }
func (*PuntReasonDump) GetMessageName() string          { return "punt_reason_dump" }
func (*PuntReasonDump) GetCrcString() string            { return "5c0dd4fe" }
func (*PuntReasonDump) GetMessageType() api.MessageType { return api.RequestMessage }

// PuntSocketDeregister represents VPP binary API message 'punt_socket_deregister'.
type PuntSocketDeregister struct {
	Punt Punt
}

func (m *PuntSocketDeregister) Reset()                        { *m = PuntSocketDeregister{} }
func (*PuntSocketDeregister) GetMessageName() string          { return "punt_socket_deregister" }
func (*PuntSocketDeregister) GetCrcString() string            { return "98a444f4" }
func (*PuntSocketDeregister) GetMessageType() api.MessageType { return api.RequestMessage }

// PuntSocketDeregisterReply represents VPP binary API message 'punt_socket_deregister_reply'.
type PuntSocketDeregisterReply struct {
	Retval int32
}

func (m *PuntSocketDeregisterReply) Reset()                        { *m = PuntSocketDeregisterReply{} }
func (*PuntSocketDeregisterReply) GetMessageName() string          { return "punt_socket_deregister_reply" }
func (*PuntSocketDeregisterReply) GetCrcString() string            { return "e8d4e804" }
func (*PuntSocketDeregisterReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// PuntSocketDetails represents VPP binary API message 'punt_socket_details'.
type PuntSocketDetails struct {
	Punt     Punt
	Pathname []byte `struc:"[108]byte"`
}

func (m *PuntSocketDetails) Reset()                        { *m = PuntSocketDetails{} }
func (*PuntSocketDetails) GetMessageName() string          { return "punt_socket_details" }
func (*PuntSocketDetails) GetCrcString() string            { return "25100aad" }
func (*PuntSocketDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// PuntSocketDump represents VPP binary API message 'punt_socket_dump'.
type PuntSocketDump struct {
	Type PuntType
}

func (m *PuntSocketDump) Reset()                        { *m = PuntSocketDump{} }
func (*PuntSocketDump) GetMessageName() string          { return "punt_socket_dump" }
func (*PuntSocketDump) GetCrcString() string            { return "52974935" }
func (*PuntSocketDump) GetMessageType() api.MessageType { return api.RequestMessage }

// PuntSocketRegister represents VPP binary API message 'punt_socket_register'.
type PuntSocketRegister struct {
	HeaderVersion uint32
	Punt          Punt
	Pathname      []byte `struc:"[108]byte"`
}

func (m *PuntSocketRegister) Reset()                        { *m = PuntSocketRegister{} }
func (*PuntSocketRegister) GetMessageName() string          { return "punt_socket_register" }
func (*PuntSocketRegister) GetCrcString() string            { return "ddc0d8e0" }
func (*PuntSocketRegister) GetMessageType() api.MessageType { return api.RequestMessage }

// PuntSocketRegisterReply represents VPP binary API message 'punt_socket_register_reply'.
type PuntSocketRegisterReply struct {
	Retval   int32
	Pathname []byte `struc:"[64]byte"`
}

func (m *PuntSocketRegisterReply) Reset()                        { *m = PuntSocketRegisterReply{} }
func (*PuntSocketRegisterReply) GetMessageName() string          { return "punt_socket_register_reply" }
func (*PuntSocketRegisterReply) GetCrcString() string            { return "42dc0ee6" }
func (*PuntSocketRegisterReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// SetPunt represents VPP binary API message 'set_punt'.
type SetPunt struct {
	IsAdd uint8
	Punt  Punt
}

func (m *SetPunt) Reset()                        { *m = SetPunt{} }
func (*SetPunt) GetMessageName() string          { return "set_punt" }
func (*SetPunt) GetCrcString() string            { return "032a42ef" }
func (*SetPunt) GetMessageType() api.MessageType { return api.RequestMessage }

// SetPuntReply represents VPP binary API message 'set_punt_reply'.
type SetPuntReply struct {
	Retval int32
}

func (m *SetPuntReply) Reset()                        { *m = SetPuntReply{} }
func (*SetPuntReply) GetMessageName() string          { return "set_punt_reply" }
func (*SetPuntReply) GetCrcString() string            { return "e8d4e804" }
func (*SetPuntReply) GetMessageType() api.MessageType { return api.ReplyMessage }

func init() {
	api.RegisterMessage((*PuntReasonDetails)(nil), "punt.PuntReasonDetails")
	api.RegisterMessage((*PuntReasonDump)(nil), "punt.PuntReasonDump")
	api.RegisterMessage((*PuntSocketDeregister)(nil), "punt.PuntSocketDeregister")
	api.RegisterMessage((*PuntSocketDeregisterReply)(nil), "punt.PuntSocketDeregisterReply")
	api.RegisterMessage((*PuntSocketDetails)(nil), "punt.PuntSocketDetails")
	api.RegisterMessage((*PuntSocketDump)(nil), "punt.PuntSocketDump")
	api.RegisterMessage((*PuntSocketRegister)(nil), "punt.PuntSocketRegister")
	api.RegisterMessage((*PuntSocketRegisterReply)(nil), "punt.PuntSocketRegisterReply")
	api.RegisterMessage((*SetPunt)(nil), "punt.SetPunt")
	api.RegisterMessage((*SetPuntReply)(nil), "punt.SetPuntReply")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*PuntReasonDetails)(nil),
		(*PuntReasonDump)(nil),
		(*PuntSocketDeregister)(nil),
		(*PuntSocketDeregisterReply)(nil),
		(*PuntSocketDetails)(nil),
		(*PuntSocketDump)(nil),
		(*PuntSocketRegister)(nil),
		(*PuntSocketRegisterReply)(nil),
		(*SetPunt)(nil),
		(*SetPuntReply)(nil),
	}
}

// RPCService represents RPC service API for punt module.
type RPCService interface {
	DumpPuntReason(ctx context.Context, in *PuntReasonDump) (RPCService_DumpPuntReasonClient, error)
	DumpPuntSocket(ctx context.Context, in *PuntSocketDump) (RPCService_DumpPuntSocketClient, error)
	PuntSocketDeregister(ctx context.Context, in *PuntSocketDeregister) (*PuntSocketDeregisterReply, error)
	PuntSocketRegister(ctx context.Context, in *PuntSocketRegister) (*PuntSocketRegisterReply, error)
	SetPunt(ctx context.Context, in *SetPunt) (*SetPuntReply, error)
}

type serviceClient struct {
	ch api.Channel
}

func NewServiceClient(ch api.Channel) RPCService {
	return &serviceClient{ch}
}

func (c *serviceClient) DumpPuntReason(ctx context.Context, in *PuntReasonDump) (RPCService_DumpPuntReasonClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpPuntReasonClient{stream}
	return x, nil
}

type RPCService_DumpPuntReasonClient interface {
	Recv() (*PuntReasonDetails, error)
}

type serviceClient_DumpPuntReasonClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpPuntReasonClient) Recv() (*PuntReasonDetails, error) {
	m := new(PuntReasonDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpPuntSocket(ctx context.Context, in *PuntSocketDump) (RPCService_DumpPuntSocketClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpPuntSocketClient{stream}
	return x, nil
}

type RPCService_DumpPuntSocketClient interface {
	Recv() (*PuntSocketDetails, error)
}

type serviceClient_DumpPuntSocketClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpPuntSocketClient) Recv() (*PuntSocketDetails, error) {
	m := new(PuntSocketDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) PuntSocketDeregister(ctx context.Context, in *PuntSocketDeregister) (*PuntSocketDeregisterReply, error) {
	out := new(PuntSocketDeregisterReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) PuntSocketRegister(ctx context.Context, in *PuntSocketRegister) (*PuntSocketRegisterReply, error) {
	out := new(PuntSocketRegisterReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SetPunt(ctx context.Context, in *SetPunt) (*SetPuntReply, error) {
	out := new(SetPuntReply)
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
