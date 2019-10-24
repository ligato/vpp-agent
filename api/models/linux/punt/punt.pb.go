// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/api/models/linux/punt/punt.proto

package linux_punt

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

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
	return fileDescriptor_aea0ef71cde48a4c, []int{1, 0}
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
	return fileDescriptor_aea0ef71cde48a4c, []int{1, 1}
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
	return fileDescriptor_aea0ef71cde48a4c, []int{0}
}

func (m *Proxy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Proxy.Unmarshal(m, b)
}
func (m *Proxy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Proxy.Marshal(b, m, deterministic)
}
func (m *Proxy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Proxy.Merge(m, src)
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

type Proxy_RxPort struct {
	RxPort *PortBased `protobuf:"bytes,1,opt,name=rx_port,json=rxPort,proto3,oneof"`
}

type Proxy_RxSocket struct {
	RxSocket *SocketBased `protobuf:"bytes,2,opt,name=rx_socket,json=rxSocket,proto3,oneof"`
}

func (*Proxy_RxPort) isProxy_Rx() {}

func (*Proxy_RxSocket) isProxy_Rx() {}

func (m *Proxy) GetRx() isProxy_Rx {
	if m != nil {
		return m.Rx
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

type isProxy_Tx interface {
	isProxy_Tx()
}

type Proxy_TxPort struct {
	TxPort *PortBased `protobuf:"bytes,3,opt,name=tx_port,json=txPort,proto3,oneof"`
}

type Proxy_TxSocket struct {
	TxSocket *SocketBased `protobuf:"bytes,4,opt,name=tx_socket,json=txSocket,proto3,oneof"`
}

func (*Proxy_TxPort) isProxy_Tx() {}

func (*Proxy_TxSocket) isProxy_Tx() {}

func (m *Proxy) GetTx() isProxy_Tx {
	if m != nil {
		return m.Tx
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

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Proxy) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Proxy_RxPort)(nil),
		(*Proxy_RxSocket)(nil),
		(*Proxy_TxPort)(nil),
		(*Proxy_TxSocket)(nil),
	}
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
	return fileDescriptor_aea0ef71cde48a4c, []int{1}
}

func (m *PortBased) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PortBased.Unmarshal(m, b)
}
func (m *PortBased) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PortBased.Marshal(b, m, deterministic)
}
func (m *PortBased) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PortBased.Merge(m, src)
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
	return fileDescriptor_aea0ef71cde48a4c, []int{2}
}

func (m *SocketBased) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SocketBased.Unmarshal(m, b)
}
func (m *SocketBased) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SocketBased.Marshal(b, m, deterministic)
}
func (m *SocketBased) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SocketBased.Merge(m, src)
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

func init() {
	proto.RegisterEnum("linux.punt.PortBased_L4Protocol", PortBased_L4Protocol_name, PortBased_L4Protocol_value)
	proto.RegisterEnum("linux.punt.PortBased_L3Protocol", PortBased_L3Protocol_name, PortBased_L3Protocol_value)
	proto.RegisterType((*Proxy)(nil), "linux.punt.Proxy")
	proto.RegisterType((*PortBased)(nil), "linux.punt.PortBased")
	proto.RegisterType((*SocketBased)(nil), "linux.punt.SocketBased")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/api/models/linux/punt/punt.proto", fileDescriptor_aea0ef71cde48a4c)
}

var fileDescriptor_aea0ef71cde48a4c = []byte{
	// 365 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x5f, 0x4f, 0xf2, 0x30,
	0x18, 0xc5, 0xd9, 0x9f, 0x77, 0xc0, 0xc3, 0xab, 0xa9, 0x4d, 0x8c, 0x5c, 0xe2, 0xae, 0xbc, 0x71,
	0x43, 0xb6, 0x70, 0x83, 0x31, 0x01, 0xc1, 0x48, 0xb2, 0x90, 0x65, 0xca, 0x8d, 0x37, 0xcb, 0x80,
	0x05, 0x16, 0x07, 0x5d, 0x4a, 0x21, 0xf5, 0x7b, 0xf8, 0x1d, 0xfd, 0x1a, 0xa6, 0x75, 0x6e, 0xc4,
	0x18, 0xe2, 0x4d, 0x73, 0xf6, 0xac, 0xe7, 0xfc, 0xda, 0x93, 0xc2, 0x4d, 0x9a, 0x2c, 0x23, 0x46,
	0xec, 0x7d, 0x96, 0x5d, 0x47, 0xcb, 0x78, 0xc3, 0xec, 0x28, 0x4b, 0xec, 0x35, 0x59, 0xc4, 0xe9,
	0xd6, 0x4e, 0x93, 0xcd, 0x8e, 0xdb, 0xd9, 0x6e, 0xc3, 0xe4, 0x62, 0x65, 0x94, 0x30, 0x82, 0x41,
	0x8e, 0x2d, 0x31, 0x31, 0x3f, 0x14, 0xf8, 0xe7, 0x53, 0xc2, 0xdf, 0x70, 0x1b, 0xaa, 0x94, 0x87,
	0x19, 0xa1, 0xac, 0xa9, 0xb4, 0x94, 0xab, 0x46, 0xe7, 0xdc, 0x2a, 0xf7, 0x59, 0x3e, 0xa1, 0x6c,
	0x10, 0x6d, 0xe3, 0xc5, 0x63, 0x25, 0x30, 0x28, 0x17, 0x9f, 0xb8, 0x0b, 0x75, 0xca, 0xc3, 0x2d,
	0x99, 0xbf, 0xc6, 0xac, 0xa9, 0x4a, 0xcf, 0xc5, 0xa1, 0xe7, 0x49, 0xfe, 0xf9, 0x76, 0xd5, 0x28,
	0xff, 0x1a, 0x08, 0x12, 0xcb, 0x49, 0xda, 0x31, 0x92, 0x12, 0x18, 0xac, 0x20, 0xb1, 0x82, 0xa4,
	0x1f, 0x27, 0x29, 0x41, 0x8d, 0xe5, 0xa4, 0x81, 0x0e, 0x2a, 0xe5, 0x62, 0x65, 0xdc, 0x7c, 0x57,
	0xa1, 0x5e, 0x64, 0xe3, 0x3e, 0x34, 0x52, 0x37, 0x94, 0x7d, 0xcc, 0x49, 0x2a, 0x6f, 0x7c, 0xda,
	0x69, 0xfd, 0x7a, 0x0e, 0xcb, 0x73, 0xfd, 0x7c, 0x5f, 0x00, 0x69, 0xa1, 0x65, 0x84, 0x53, 0x46,
	0x68, 0x47, 0x23, 0x9c, 0x83, 0x88, 0x42, 0x63, 0x0c, 0xba, 0xac, 0x41, 0x5c, 0xe9, 0x24, 0x90,
	0xda, 0x6c, 0x03, 0x94, 0x40, 0x8c, 0xe0, 0xff, 0x74, 0x32, 0x1c, 0x3d, 0x8c, 0x27, 0xa3, 0x61,
	0xe8, 0xb9, 0xa8, 0x82, 0xab, 0xa0, 0x3d, 0xdf, 0xfb, 0xc8, 0x10, 0x62, 0x3a, 0xf4, 0xd1, 0x99,
	0xd9, 0x03, 0x28, 0xf3, 0x7f, 0x38, 0x1c, 0x54, 0xc1, 0x35, 0xd0, 0xc7, 0xfe, 0xde, 0x45, 0x4a,
	0xae, 0xba, 0x48, 0x15, 0xe6, 0xbe, 0xe7, 0x21, 0xcd, 0xbc, 0x84, 0xc6, 0x41, 0x7b, 0xf2, 0x44,
	0x11, 0x5b, 0xc9, 0x42, 0xea, 0x81, 0xd4, 0x83, 0xbb, 0x97, 0xdb, 0x65, 0xc2, 0x56, 0xbb, 0x99,
	0x35, 0x27, 0x6b, 0xfb, 0x6f, 0xef, 0xad, 0x27, 0x65, 0x28, 0xe4, 0xcc, 0x90, 0x1d, 0x39, 0x9f,
	0x01, 0x00, 0x00, 0xff, 0xff, 0xee, 0xe1, 0xb6, 0xcc, 0xab, 0x02, 0x00, 0x00,
}
