// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/models/vpp/l3/vrf.proto

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
	return fileDescriptor_ebba13541c92d5a6, []int{0, 0}
}

type VrfTable struct {
	Id                   uint32            `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Protocol             VrfTable_Protocol `protobuf:"varint,2,opt,name=protocol,proto3,enum=vpp.l3.VrfTable_Protocol" json:"protocol,omitempty"`
	Label                string            `protobuf:"bytes,3,opt,name=label,proto3" json:"label,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *VrfTable) Reset()         { *m = VrfTable{} }
func (m *VrfTable) String() string { return proto.CompactTextString(m) }
func (*VrfTable) ProtoMessage()    {}
func (*VrfTable) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebba13541c92d5a6, []int{0}
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
	proto.RegisterEnum("vpp.l3.VrfTable_Protocol", VrfTable_Protocol_name, VrfTable_Protocol_value)
	proto.RegisterType((*VrfTable)(nil), "vpp.l3.VrfTable")
}

func init() { proto.RegisterFile("api/models/vpp/l3/vrf.proto", fileDescriptor_ebba13541c92d5a6) }

var fileDescriptor_ebba13541c92d5a6 = []byte{
	// 203 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4e, 0x2c, 0xc8, 0xd4,
	0xcf, 0xcd, 0x4f, 0x49, 0xcd, 0x29, 0xd6, 0x2f, 0x2b, 0x28, 0xd0, 0xcf, 0x31, 0xd6, 0x2f, 0x2b,
	0x4a, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2b, 0x2b, 0x28, 0xd0, 0xcb, 0x31, 0x56,
	0x6a, 0x67, 0xe4, 0xe2, 0x08, 0x2b, 0x4a, 0x0b, 0x49, 0x4c, 0xca, 0x49, 0x15, 0xe2, 0xe3, 0x62,
	0xca, 0x4c, 0x91, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0d, 0x62, 0xca, 0x4c, 0x11, 0x32, 0xe5, 0xe2,
	0x00, 0xab, 0x4e, 0xce, 0xcf, 0x91, 0x60, 0x52, 0x60, 0xd4, 0xe0, 0x33, 0x92, 0xd4, 0x83, 0xe8,
	0xd3, 0x83, 0xe9, 0xd1, 0x0b, 0x80, 0x2a, 0x08, 0x82, 0x2b, 0x15, 0x12, 0xe1, 0x62, 0xcd, 0x49,
	0x4c, 0x4a, 0xcd, 0x91, 0x60, 0x56, 0x60, 0xd4, 0xe0, 0x0c, 0x82, 0x70, 0x94, 0xe4, 0xb8, 0x38,
	0x60, 0x6a, 0x85, 0x38, 0xb8, 0x58, 0x3c, 0x03, 0xc2, 0x4c, 0x04, 0x18, 0xa0, 0x2c, 0x33, 0x01,
	0x46, 0x27, 0xb3, 0x28, 0x93, 0xf4, 0xcc, 0x92, 0x8c, 0xd2, 0x24, 0xbd, 0xe4, 0xfc, 0x5c, 0xfd,
	0x9c, 0xcc, 0xf4, 0xc4, 0x92, 0x7c, 0x90, 0xbb, 0x75, 0x13, 0xd3, 0x53, 0xf3, 0x4a, 0xf4, 0x31,
	0x3c, 0x63, 0x5d, 0x56, 0x50, 0x10, 0x9f, 0x63, 0x9c, 0xc4, 0x06, 0xb6, 0xd7, 0x18, 0x10, 0x00,
	0x00, 0xff, 0xff, 0x17, 0x09, 0x19, 0xbf, 0xef, 0x00, 0x00, 0x00,
}
