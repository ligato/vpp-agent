// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: punt.proto

package punt

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// L3 protocol definition
type L3Protocol int32

const (
	L3Protocol_UNDEFINED_L3 L3Protocol = 0
	L3Protocol_IPv4         L3Protocol = 4
	L3Protocol_IPv6         L3Protocol = 6
	L3Protocol_ALL          L3Protocol = 10
)

var L3Protocol_name = map[int32]string{
	0:  "UNDEFINED_L3",
	4:  "IPv4",
	6:  "IPv6",
	10: "ALL",
}
var L3Protocol_value = map[string]int32{
	"UNDEFINED_L3": 0,
	"IPv4":         4,
	"IPv6":         6,
	"ALL":          10,
}

func (x L3Protocol) String() string {
	return proto.EnumName(L3Protocol_name, int32(x))
}
func (L3Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_punt_e71f973f1eec15e5, []int{0}
}

// L4 protocol definition
type L4Protocol int32

const (
	L4Protocol_UNDEFINED_L4 L4Protocol = 0
	L4Protocol_TCP          L4Protocol = 6
	L4Protocol_UDP          L4Protocol = 17
)

var L4Protocol_name = map[int32]string{
	0:  "UNDEFINED_L4",
	6:  "TCP",
	17: "UDP",
}
var L4Protocol_value = map[string]int32{
	"UNDEFINED_L4": 0,
	"TCP":          6,
	"UDP":          17,
}

func (x L4Protocol) String() string {
	return proto.EnumName(L4Protocol_name, int32(x))
}
func (L4Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_punt_e71f973f1eec15e5, []int{1}
}

// Allows otherwise dropped packet which destination IP address matching some of the VPP interface IP addresses to be
// punted to the host via socket. L3 and L4 protocols can be used for filtering
type Punt struct {
	Name                 string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	L3Protocol           L3Protocol `protobuf:"varint,2,opt,name=l3_protocol,json=l3Protocol,proto3,enum=punt.L3Protocol" json:"l3_protocol,omitempty"`
	L4Protocol           L4Protocol `protobuf:"varint,3,opt,name=l4_protocol,json=l4Protocol,proto3,enum=punt.L4Protocol" json:"l4_protocol,omitempty"`
	Port                 uint32     `protobuf:"varint,4,opt,name=port,proto3" json:"port,omitempty"`
	SocketPath           string     `protobuf:"bytes,5,opt,name=socket_path,json=socketPath,proto3" json:"socket_path,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Punt) Reset()         { *m = Punt{} }
func (m *Punt) String() string { return proto.CompactTextString(m) }
func (*Punt) ProtoMessage()    {}
func (*Punt) Descriptor() ([]byte, []int) {
	return fileDescriptor_punt_e71f973f1eec15e5, []int{0}
}
func (m *Punt) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Punt.Unmarshal(m, b)
}
func (m *Punt) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Punt.Marshal(b, m, deterministic)
}
func (dst *Punt) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Punt.Merge(dst, src)
}
func (m *Punt) XXX_Size() int {
	return xxx_messageInfo_Punt.Size(m)
}
func (m *Punt) XXX_DiscardUnknown() {
	xxx_messageInfo_Punt.DiscardUnknown(m)
}

var xxx_messageInfo_Punt proto.InternalMessageInfo

func (m *Punt) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Punt) GetL3Protocol() L3Protocol {
	if m != nil {
		return m.L3Protocol
	}
	return L3Protocol_UNDEFINED_L3
}

func (m *Punt) GetL4Protocol() L4Protocol {
	if m != nil {
		return m.L4Protocol
	}
	return L4Protocol_UNDEFINED_L4
}

func (m *Punt) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *Punt) GetSocketPath() string {
	if m != nil {
		return m.SocketPath
	}
	return ""
}

func init() {
	proto.RegisterType((*Punt)(nil), "punt.Punt")
	proto.RegisterEnum("punt.L3Protocol", L3Protocol_name, L3Protocol_value)
	proto.RegisterEnum("punt.L4Protocol", L4Protocol_name, L4Protocol_value)
}

func init() { proto.RegisterFile("punt.proto", fileDescriptor_punt_e71f973f1eec15e5) }

var fileDescriptor_punt_e71f973f1eec15e5 = []byte{
	// 229 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x8f, 0xc1, 0x4b, 0xc3, 0x30,
	0x18, 0xc5, 0x17, 0x17, 0xa7, 0xbe, 0xa9, 0x7c, 0x7e, 0xa7, 0xde, 0x2c, 0x9e, 0xca, 0x0e, 0x43,
	0x4d, 0xf0, 0xe2, 0x49, 0xec, 0x84, 0x41, 0x18, 0xa1, 0xb8, 0x73, 0xa9, 0x63, 0x30, 0xb0, 0x36,
	0x61, 0x66, 0xfe, 0x67, 0xfe, 0x7f, 0xd2, 0x6c, 0xb4, 0xea, 0xed, 0xc7, 0xe3, 0xfd, 0x92, 0xf7,
	0x01, 0x7e, 0xd7, 0x84, 0xa9, 0xdf, 0xba, 0xe0, 0x58, 0xb6, 0x7c, 0xf3, 0x2d, 0x20, 0xed, 0xae,
	0x09, 0xcc, 0x90, 0x4d, 0xf5, 0xb1, 0x4e, 0x44, 0x2a, 0xb2, 0xb3, 0x22, 0x32, 0xdf, 0x61, 0x5c,
	0xab, 0x32, 0xd6, 0x57, 0xae, 0x4e, 0x8e, 0x52, 0x91, 0x5d, 0xde, 0xd3, 0x34, 0x3e, 0x62, 0x94,
	0x3d, 0xe4, 0x05, 0xea, 0x8e, 0xa3, 0xa2, 0x7b, 0x65, 0xf8, 0x47, 0xd1, 0xbf, 0x94, 0x8e, 0xdb,
	0x9f, 0xbd, 0xdb, 0x86, 0x44, 0xa6, 0x22, 0xbb, 0x28, 0x22, 0xf3, 0x35, 0xc6, 0x9f, 0x6e, 0xf5,
	0xbe, 0x0e, 0xa5, 0xaf, 0xc2, 0x26, 0x39, 0x8e, 0xa3, 0xb0, 0x8f, 0x6c, 0x15, 0x36, 0x93, 0x47,
	0xa0, 0x5f, 0xc0, 0x84, 0xf3, 0xe5, 0x22, 0x9f, 0xbd, 0xcc, 0x17, 0xb3, 0xbc, 0x34, 0x8a, 0x06,
	0x7c, 0x0a, 0x39, 0xb7, 0x5f, 0x9a, 0xe4, 0x81, 0x1e, 0x68, 0xc4, 0x27, 0x18, 0x3e, 0x19, 0x43,
	0x98, 0xdc, 0x02, 0xfd, 0x96, 0x7f, 0xb2, 0xa6, 0x41, 0x5b, 0x7c, 0x7d, 0xb6, 0x7b, 0x63, 0x99,
	0x5b, 0xba, 0x7a, 0x1b, 0xc5, 0x8b, 0xd4, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x58, 0xa3, 0x8b,
	0x4e, 0x41, 0x01, 0x00, 0x00,
}