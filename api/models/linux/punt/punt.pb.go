// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: models/linux/punt/punt.proto

package linux_punt // import "github.com/ligato/vpp-agent/api/models/linux/punt"

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type PortBased_L4Protocol int32

const (
	PortBased_UNDEFINED_L4 PortBased_L4Protocol = 0
	PortBased_TCP          PortBased_L4Protocol = 6
	PortBased_UDP          PortBased_L4Protocol = 17
)

var PortBased_L4Protocol_name = map[int32]string{
	0:  "UNDEFINED_L4",
	6:  "TCP",
	17: "UDP",
}
var PortBased_L4Protocol_value = map[string]int32{
	"UNDEFINED_L4": 0,
	"TCP":          6,
	"UDP":          17,
}

func (x PortBased_L4Protocol) String() string {
	return proto.EnumName(PortBased_L4Protocol_name, int32(x))
}
func (PortBased_L4Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_punt_808f67b1735e53e4, []int{1, 0}
}

type PortBased_L3Protocol int32

const (
	PortBased_UNDEFINED_L3 PortBased_L3Protocol = 0
	PortBased_IPv4         PortBased_L3Protocol = 1
	PortBased_IPv6         PortBased_L3Protocol = 2
	PortBased_ALL          PortBased_L3Protocol = 3
)

var PortBased_L3Protocol_name = map[int32]string{
	0: "UNDEFINED_L3",
	1: "IPv4",
	2: "IPv6",
	3: "ALL",
}
var PortBased_L3Protocol_value = map[string]int32{
	"UNDEFINED_L3": 0,
	"IPv4":         1,
	"IPv6":         2,
	"ALL":          3,
}

func (x PortBased_L3Protocol) String() string {
	return proto.EnumName(PortBased_L3Protocol_name, int32(x))
}
func (PortBased_L3Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_punt_808f67b1735e53e4, []int{1, 1}
}

// Proxy allows to listen on network socket or unix domain socket, and resend to another network/unix domain socket
type Proxy struct {
	// Types that are valid to be assigned to Rx:
	//	*Proxy_RxPort
	//	*Proxy_RxSocket
	Rx isProxy_Rx `protobuf_oneof:"rx"`
	// Types that are valid to be assigned to Tx:
	//	*Proxy_TxPort
	//	*Proxy_TxSocket
	Tx                   isProxy_Tx `protobuf_oneof:"tx"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Proxy) Reset()         { *m = Proxy{} }
func (m *Proxy) String() string { return proto.CompactTextString(m) }
func (*Proxy) ProtoMessage()    {}
func (*Proxy) Descriptor() ([]byte, []int) {
	return fileDescriptor_punt_808f67b1735e53e4, []int{0}
}
func (m *Proxy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Proxy.Unmarshal(m, b)
}
func (m *Proxy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Proxy.Marshal(b, m, deterministic)
}
func (dst *Proxy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Proxy.Merge(dst, src)
}
func (m *Proxy) XXX_Size() int {
	return xxx_messageInfo_Proxy.Size(m)
}
func (m *Proxy) XXX_DiscardUnknown() {
	xxx_messageInfo_Proxy.DiscardUnknown(m)
}

var xxx_messageInfo_Proxy proto.InternalMessageInfo

type isProxy_Rx interface {
	isProxy_Rx()
}
type isProxy_Tx interface {
	isProxy_Tx()
}

type Proxy_RxPort struct {
	RxPort *PortBased `protobuf:"bytes,1,opt,name=rx_port,json=rxPort,proto3,oneof"`
}
type Proxy_RxSocket struct {
	RxSocket *SocketBased `protobuf:"bytes,2,opt,name=rx_socket,json=rxSocket,proto3,oneof"`
}
type Proxy_TxPort struct {
	TxPort *PortBased `protobuf:"bytes,3,opt,name=tx_port,json=txPort,proto3,oneof"`
}
type Proxy_TxSocket struct {
	TxSocket *SocketBased `protobuf:"bytes,4,opt,name=tx_socket,json=txSocket,proto3,oneof"`
}

func (*Proxy_RxPort) isProxy_Rx()   {}
func (*Proxy_RxSocket) isProxy_Rx() {}
func (*Proxy_TxPort) isProxy_Tx()   {}
func (*Proxy_TxSocket) isProxy_Tx() {}

func (m *Proxy) GetRx() isProxy_Rx {
	if m != nil {
		return m.Rx
	}
	return nil
}
func (m *Proxy) GetTx() isProxy_Tx {
	if m != nil {
		return m.Tx
	}
	return nil
}

func (m *Proxy) GetRxPort() *PortBased {
	if x, ok := m.GetRx().(*Proxy_RxPort); ok {
		return x.RxPort
	}
	return nil
}

func (m *Proxy) GetRxSocket() *SocketBased {
	if x, ok := m.GetRx().(*Proxy_RxSocket); ok {
		return x.RxSocket
	}
	return nil
}

func (m *Proxy) GetTxPort() *PortBased {
	if x, ok := m.GetTx().(*Proxy_TxPort); ok {
		return x.TxPort
	}
	return nil
}

func (m *Proxy) GetTxSocket() *SocketBased {
	if x, ok := m.GetTx().(*Proxy_TxSocket); ok {
		return x.TxSocket
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Proxy) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Proxy_OneofMarshaler, _Proxy_OneofUnmarshaler, _Proxy_OneofSizer, []interface{}{
		(*Proxy_RxPort)(nil),
		(*Proxy_RxSocket)(nil),
		(*Proxy_TxPort)(nil),
		(*Proxy_TxSocket)(nil),
	}
}

func _Proxy_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Proxy)
	// rx
	switch x := m.Rx.(type) {
	case *Proxy_RxPort:
		_ = b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.RxPort); err != nil {
			return err
		}
	case *Proxy_RxSocket:
		_ = b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.RxSocket); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Proxy.Rx has unexpected type %T", x)
	}
	// tx
	switch x := m.Tx.(type) {
	case *Proxy_TxPort:
		_ = b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.TxPort); err != nil {
			return err
		}
	case *Proxy_TxSocket:
		_ = b.EncodeVarint(4<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.TxSocket); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Proxy.Tx has unexpected type %T", x)
	}
	return nil
}

func _Proxy_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Proxy)
	switch tag {
	case 1: // rx.rx_port
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(PortBased)
		err := b.DecodeMessage(msg)
		m.Rx = &Proxy_RxPort{msg}
		return true, err
	case 2: // rx.rx_socket
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(SocketBased)
		err := b.DecodeMessage(msg)
		m.Rx = &Proxy_RxSocket{msg}
		return true, err
	case 3: // tx.tx_port
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(PortBased)
		err := b.DecodeMessage(msg)
		m.Tx = &Proxy_TxPort{msg}
		return true, err
	case 4: // tx.tx_socket
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(SocketBased)
		err := b.DecodeMessage(msg)
		m.Tx = &Proxy_TxSocket{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Proxy_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Proxy)
	// rx
	switch x := m.Rx.(type) {
	case *Proxy_RxPort:
		s := proto.Size(x.RxPort)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Proxy_RxSocket:
		s := proto.Size(x.RxSocket)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	// tx
	switch x := m.Tx.(type) {
	case *Proxy_TxPort:
		s := proto.Size(x.TxPort)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Proxy_TxSocket:
		s := proto.Size(x.TxSocket)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

func (*Proxy) XXX_MessageName() string {
	return "linux.punt.Proxy"
}

// Define network socket type
type PortBased struct {
	L4Protocol           PortBased_L4Protocol `protobuf:"varint,1,opt,name=l4_protocol,json=l4Protocol,proto3,enum=linux.punt.PortBased_L4Protocol" json:"l4_protocol,omitempty"`
	L3Protocol           PortBased_L3Protocol `protobuf:"varint,3,opt,name=l3_protocol,json=l3Protocol,proto3,enum=linux.punt.PortBased_L3Protocol" json:"l3_protocol,omitempty"`
	Port                 uint32               `protobuf:"varint,4,opt,name=port,proto3" json:"port,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *PortBased) Reset()         { *m = PortBased{} }
func (m *PortBased) String() string { return proto.CompactTextString(m) }
func (*PortBased) ProtoMessage()    {}
func (*PortBased) Descriptor() ([]byte, []int) {
	return fileDescriptor_punt_808f67b1735e53e4, []int{1}
}
func (m *PortBased) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PortBased.Unmarshal(m, b)
}
func (m *PortBased) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PortBased.Marshal(b, m, deterministic)
}
func (dst *PortBased) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PortBased.Merge(dst, src)
}
func (m *PortBased) XXX_Size() int {
	return xxx_messageInfo_PortBased.Size(m)
}
func (m *PortBased) XXX_DiscardUnknown() {
	xxx_messageInfo_PortBased.DiscardUnknown(m)
}

var xxx_messageInfo_PortBased proto.InternalMessageInfo

func (m *PortBased) GetL4Protocol() PortBased_L4Protocol {
	if m != nil {
		return m.L4Protocol
	}
	return PortBased_UNDEFINED_L4
}

func (m *PortBased) GetL3Protocol() PortBased_L3Protocol {
	if m != nil {
		return m.L3Protocol
	}
	return PortBased_UNDEFINED_L3
}

func (m *PortBased) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (*PortBased) XXX_MessageName() string {
	return "linux.punt.PortBased"
}

// Define unix domain socket type for IPC
type SocketBased struct {
	Path                 string   `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SocketBased) Reset()         { *m = SocketBased{} }
func (m *SocketBased) String() string { return proto.CompactTextString(m) }
func (*SocketBased) ProtoMessage()    {}
func (*SocketBased) Descriptor() ([]byte, []int) {
	return fileDescriptor_punt_808f67b1735e53e4, []int{2}
}
func (m *SocketBased) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SocketBased.Unmarshal(m, b)
}
func (m *SocketBased) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SocketBased.Marshal(b, m, deterministic)
}
func (dst *SocketBased) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SocketBased.Merge(dst, src)
}
func (m *SocketBased) XXX_Size() int {
	return xxx_messageInfo_SocketBased.Size(m)
}
func (m *SocketBased) XXX_DiscardUnknown() {
	xxx_messageInfo_SocketBased.DiscardUnknown(m)
}

var xxx_messageInfo_SocketBased proto.InternalMessageInfo

func (m *SocketBased) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (*SocketBased) XXX_MessageName() string {
	return "linux.punt.SocketBased"
}
func init() {
	proto.RegisterType((*Proxy)(nil), "linux.punt.Proxy")
	proto.RegisterType((*PortBased)(nil), "linux.punt.PortBased")
	proto.RegisterType((*SocketBased)(nil), "linux.punt.SocketBased")
	proto.RegisterEnum("linux.punt.PortBased_L4Protocol", PortBased_L4Protocol_name, PortBased_L4Protocol_value)
	proto.RegisterEnum("linux.punt.PortBased_L3Protocol", PortBased_L3Protocol_name, PortBased_L3Protocol_value)
}

func init() { proto.RegisterFile("models/linux/punt/punt.proto", fileDescriptor_punt_808f67b1735e53e4) }

var fileDescriptor_punt_808f67b1735e53e4 = []byte{
	// 394 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0xcf, 0x8e, 0xd3, 0x30,
	0x10, 0xc6, 0x9b, 0x3f, 0x64, 0xdb, 0x29, 0x20, 0x63, 0x09, 0xb1, 0x42, 0x08, 0x2d, 0x39, 0x71,
	0xd9, 0x64, 0xb5, 0x89, 0xf6, 0xb2, 0x5c, 0x1a, 0x5a, 0x44, 0xa5, 0xa8, 0x8a, 0x02, 0xbd, 0x70,
	0x89, 0xd2, 0x36, 0xa4, 0x11, 0x69, 0x1d, 0xb9, 0x4e, 0x65, 0xde, 0x83, 0x87, 0xe2, 0x3d, 0x38,
	0xf0, 0x1a, 0xc8, 0xd3, 0x90, 0x44, 0xb0, 0xea, 0xc5, 0xfa, 0xcd, 0xd8, 0xdf, 0xf7, 0xd9, 0x23,
	0xc3, 0xab, 0x1d, 0xdb, 0x64, 0xe5, 0xc1, 0x2d, 0x8b, 0x7d, 0x2d, 0xdd, 0xaa, 0xde, 0x0b, 0x5c,
	0x9c, 0x8a, 0x33, 0xc1, 0x28, 0x60, 0xdb, 0x51, 0x9d, 0x97, 0xd7, 0x79, 0x21, 0xb6, 0xf5, 0xca,
	0x59, 0xb3, 0x9d, 0x9b, 0xb3, 0x9c, 0xb9, 0x78, 0x64, 0x55, 0x7f, 0xc5, 0x0a, 0x0b, 0xa4, 0x93,
	0xd4, 0xfe, 0xad, 0xc1, 0xa3, 0x88, 0x33, 0xf9, 0x9d, 0xde, 0xc0, 0x05, 0x97, 0x49, 0xc5, 0xb8,
	0xb8, 0xd4, 0xae, 0xb4, 0xb7, 0xe3, 0xdb, 0xe7, 0x4e, 0x67, 0xeb, 0x44, 0x8c, 0x8b, 0x20, 0x3d,
	0x64, 0x9b, 0x8f, 0x83, 0xd8, 0xe2, 0x52, 0x95, 0xf4, 0x0e, 0x46, 0x5c, 0x26, 0x07, 0xb6, 0xfe,
	0x96, 0x89, 0x4b, 0x1d, 0x35, 0x2f, 0xfa, 0x9a, 0x4f, 0xb8, 0xf3, 0x57, 0x35, 0xe4, 0xf2, 0xd4,
	0x50, 0x49, 0xa2, 0x49, 0x32, 0xce, 0x25, 0x69, 0xb1, 0x25, 0xda, 0x24, 0xd1, 0x26, 0x99, 0xe7,
	0x93, 0xb4, 0x78, 0x28, 0x9a, 0xa4, 0xc0, 0x04, 0x9d, 0x4b, 0xb5, 0x0a, 0x69, 0xff, 0xd0, 0x61,
	0xd4, 0x7a, 0xd3, 0x09, 0x8c, 0x4b, 0x3f, 0xc1, 0x19, 0xac, 0x59, 0x89, 0x2f, 0x7e, 0x7a, 0x7b,
	0xf5, 0xe0, 0x3d, 0x9c, 0xd0, 0x8f, 0x9a, 0x73, 0x31, 0x94, 0x2d, 0xa3, 0x85, 0xd7, 0x59, 0x18,
	0x67, 0x2d, 0xbc, 0x9e, 0x45, 0xcb, 0x94, 0x82, 0x89, 0x63, 0x50, 0x4f, 0x7a, 0x12, 0x23, 0xdb,
	0x37, 0x00, 0x5d, 0x20, 0x25, 0xf0, 0x78, 0xb9, 0x98, 0xce, 0x3e, 0xcc, 0x17, 0xb3, 0x69, 0x12,
	0xfa, 0x64, 0x40, 0x2f, 0xc0, 0xf8, 0xfc, 0x3e, 0x22, 0x96, 0x82, 0xe5, 0x34, 0x22, 0xcf, 0xec,
	0x7b, 0x80, 0xce, 0xff, 0x1f, 0x85, 0x47, 0x06, 0x74, 0x08, 0xe6, 0x3c, 0x3a, 0xfa, 0x44, 0x6b,
	0xe8, 0x8e, 0xe8, 0x4a, 0x3c, 0x09, 0x43, 0x62, 0xd8, 0x6f, 0x60, 0xdc, 0x9b, 0x1e, 0xde, 0x28,
	0x15, 0x5b, 0x1c, 0xc8, 0x28, 0x46, 0x0e, 0x82, 0x9f, 0xbf, 0x5e, 0x6b, 0x5f, 0xde, 0xf5, 0x3e,
	0x56, 0x59, 0xe4, 0xa9, 0x60, 0xee, 0xb1, 0xaa, 0xae, 0xd3, 0x3c, 0xdb, 0x0b, 0x37, 0xad, 0x0a,
	0xf7, 0xbf, 0x2f, 0x7a, 0x8f, 0x98, 0x28, 0x5c, 0x59, 0x38, 0x27, 0xef, 0x4f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x54, 0x1e, 0xcc, 0xdc, 0xc9, 0x02, 0x00, 0x00,
}