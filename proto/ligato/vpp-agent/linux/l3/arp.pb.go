// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/linux/l3/arp.proto

package linux_l3

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

type ARPEntry struct {
	Interface            string   `protobuf:"bytes,1,opt,name=interface,proto3" json:"interface,omitempty"`
	IpAddress            string   `protobuf:"bytes,2,opt,name=ip_address,json=ipAddress,proto3" json:"ip_address,omitempty"`
	HwAddress            string   `protobuf:"bytes,3,opt,name=hw_address,json=hwAddress,proto3" json:"hw_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ARPEntry) Reset()         { *m = ARPEntry{} }
func (m *ARPEntry) String() string { return proto.CompactTextString(m) }
func (*ARPEntry) ProtoMessage()    {}
func (*ARPEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_e94dcdb3c70bc8da, []int{0}
}

func (m *ARPEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ARPEntry.Unmarshal(m, b)
}
func (m *ARPEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ARPEntry.Marshal(b, m, deterministic)
}
func (m *ARPEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ARPEntry.Merge(m, src)
}
func (m *ARPEntry) XXX_Size() int {
	return xxx_messageInfo_ARPEntry.Size(m)
}
func (m *ARPEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_ARPEntry.DiscardUnknown(m)
}

var xxx_messageInfo_ARPEntry proto.InternalMessageInfo

func (m *ARPEntry) GetInterface() string {
	if m != nil {
		return m.Interface
	}
	return ""
}

func (m *ARPEntry) GetIpAddress() string {
	if m != nil {
		return m.IpAddress
	}
	return ""
}

func (m *ARPEntry) GetHwAddress() string {
	if m != nil {
		return m.HwAddress
	}
	return ""
}

func init() {
	proto.RegisterType((*ARPEntry)(nil), "ligato.vpp_agent.linux.l3.ARPEntry")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/linux/l3/arp.proto", fileDescriptor_e94dcdb3c70bc8da)
}

var fileDescriptor_e94dcdb3c70bc8da = []byte{
	// 178 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0xce, 0xc9, 0x4c, 0x4f,
	0x2c, 0xc9, 0xd7, 0x2f, 0x2b, 0x28, 0xd0, 0x4d, 0x4c, 0x4f, 0xcd, 0x2b, 0xd1, 0xcf, 0xc9, 0xcc,
	0x2b, 0xad, 0xd0, 0xcf, 0x31, 0xd6, 0x4f, 0x2c, 0x2a, 0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0x92, 0x84, 0x28, 0xd2, 0x2b, 0x2b, 0x28, 0x88, 0x07, 0x2b, 0xd2, 0x03, 0x2b, 0xd2, 0xcb, 0x31,
	0x56, 0x4a, 0xe3, 0xe2, 0x70, 0x0c, 0x0a, 0x70, 0xcd, 0x2b, 0x29, 0xaa, 0x14, 0x92, 0xe1, 0xe2,
	0xcc, 0xcc, 0x2b, 0x49, 0x2d, 0x4a, 0x4b, 0x4c, 0x4e, 0x95, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c,
	0x42, 0x08, 0x08, 0xc9, 0x72, 0x71, 0x65, 0x16, 0xc4, 0x27, 0xa6, 0xa4, 0x14, 0xa5, 0x16, 0x17,
	0x4b, 0x30, 0x41, 0xa5, 0x0b, 0x1c, 0x21, 0x02, 0x20, 0xe9, 0x8c, 0x72, 0xb8, 0x34, 0x33, 0x44,
	0x3a, 0xa3, 0x1c, 0x2a, 0xed, 0xe4, 0x12, 0xe5, 0x94, 0x9e, 0xaf, 0x07, 0x75, 0x47, 0x26, 0xb2,
	0x7b, 0xcb, 0x8c, 0xf4, 0xc1, 0xae, 0xd4, 0xc7, 0xe9, 0x13, 0x6b, 0x30, 0x23, 0x3e, 0xc7, 0x38,
	0x89, 0x0d, 0xac, 0xd2, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x31, 0xe5, 0x4b, 0x28, 0xf6, 0x00,
	0x00, 0x00,
}
