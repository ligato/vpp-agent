// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/vpp/l3/vrf.proto

package vpp_l3

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

type VrfTable_Protocol int32

const (
	VrfTable_IPV4 VrfTable_Protocol = 0
	VrfTable_IPV6 VrfTable_Protocol = 1
)

var VrfTable_Protocol_name = map[int32]string{
	0: "IPV4",
	1: "IPV6",
}

var VrfTable_Protocol_value = map[string]int32{
	"IPV4": 0,
	"IPV6": 1,
}

func (x VrfTable_Protocol) String() string {
	return proto.EnumName(VrfTable_Protocol_name, int32(x))
}

func (VrfTable_Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_53788f30600e5758, []int{0, 0}
}

type VrfTable struct {
	Id                   uint32            `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Protocol             VrfTable_Protocol `protobuf:"varint,2,opt,name=protocol,proto3,enum=ligato.vpp_agent.vpp.l3.VrfTable_Protocol" json:"protocol,omitempty"`
	Label                string            `protobuf:"bytes,3,opt,name=label,proto3" json:"label,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *VrfTable) Reset()         { *m = VrfTable{} }
func (m *VrfTable) String() string { return proto.CompactTextString(m) }
func (*VrfTable) ProtoMessage()    {}
func (*VrfTable) Descriptor() ([]byte, []int) {
	return fileDescriptor_53788f30600e5758, []int{0}
}

func (m *VrfTable) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VrfTable.Unmarshal(m, b)
}
func (m *VrfTable) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VrfTable.Marshal(b, m, deterministic)
}
func (m *VrfTable) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VrfTable.Merge(m, src)
}
func (m *VrfTable) XXX_Size() int {
	return xxx_messageInfo_VrfTable.Size(m)
}
func (m *VrfTable) XXX_DiscardUnknown() {
	xxx_messageInfo_VrfTable.DiscardUnknown(m)
}

var xxx_messageInfo_VrfTable proto.InternalMessageInfo

func (m *VrfTable) GetId() uint32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *VrfTable) GetProtocol() VrfTable_Protocol {
	if m != nil {
		return m.Protocol
	}
	return VrfTable_IPV4
}

func (m *VrfTable) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func init() {
	proto.RegisterEnum("ligato.vpp_agent.vpp.l3.VrfTable_Protocol", VrfTable_Protocol_name, VrfTable_Protocol_value)
	proto.RegisterType((*VrfTable)(nil), "ligato.vpp_agent.vpp.l3.VrfTable")
}

func init() { proto.RegisterFile("ligato/vpp-agent/vpp/l3/vrf.proto", fileDescriptor_53788f30600e5758) }

var fileDescriptor_53788f30600e5758 = []byte{
	// 200 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0xcc, 0xc9, 0x4c, 0x4f,
	0x2c, 0xc9, 0xd7, 0x2f, 0x2b, 0x28, 0xd0, 0x4d, 0x4c, 0x4f, 0xcd, 0x2b, 0x01, 0xb1, 0xf4, 0x73,
	0x8c, 0xf5, 0xcb, 0x8a, 0xd2, 0xf4, 0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0xc4, 0x21, 0x4a, 0xf4,
	0xca, 0x0a, 0x0a, 0xe2, 0xc1, 0x4a, 0x40, 0x2c, 0xbd, 0x1c, 0x63, 0xa5, 0x19, 0x8c, 0x5c, 0x1c,
	0x61, 0x45, 0x69, 0x21, 0x89, 0x49, 0x39, 0xa9, 0x42, 0x7c, 0x5c, 0x4c, 0x99, 0x29, 0x12, 0x8c,
	0x0a, 0x8c, 0x1a, 0xbc, 0x41, 0x4c, 0x99, 0x29, 0x42, 0x6e, 0x5c, 0x1c, 0x60, 0xed, 0xc9, 0xf9,
	0x39, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0x7c, 0x46, 0x5a, 0x7a, 0x38, 0x0c, 0xd2, 0x83, 0x19, 0xa2,
	0x17, 0x00, 0xd5, 0x11, 0x04, 0xd7, 0x2b, 0x24, 0xc2, 0xc5, 0x9a, 0x93, 0x98, 0x94, 0x9a, 0x23,
	0xc1, 0xac, 0xc0, 0xa8, 0xc1, 0x19, 0x04, 0xe1, 0x28, 0xc9, 0x71, 0x71, 0xc0, 0xd4, 0x0a, 0x71,
	0x70, 0xb1, 0x78, 0x06, 0x84, 0x99, 0x08, 0x30, 0x40, 0x59, 0x66, 0x02, 0x8c, 0x4e, 0x0e, 0x51,
	0x76, 0xe9, 0xf9, 0x30, 0xfb, 0x32, 0x51, 0xbc, 0x67, 0xa4, 0x0f, 0x36, 0x5b, 0x1f, 0x87, 0xc7,
	0xad, 0x41, 0xae, 0xcb, 0x31, 0x4e, 0x62, 0x03, 0xab, 0x32, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff,
	0xd9, 0x91, 0x01, 0x8d, 0x21, 0x01, 0x00, 0x00,
}
