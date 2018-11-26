// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: vpp/stn/stn.proto

package vpp_stn // import "github.com/ligato/vpp-agent/api/models/vpp/stn"

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

type Rule struct {
	IpAddress            string   `protobuf:"bytes,1,opt,name=ip_address,json=ipAddress,proto3" json:"ip_address,omitempty"`
	Interface            string   `protobuf:"bytes,2,opt,name=interface,proto3" json:"interface,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Rule) Reset()         { *m = Rule{} }
func (m *Rule) String() string { return proto.CompactTextString(m) }
func (*Rule) ProtoMessage()    {}
func (*Rule) Descriptor() ([]byte, []int) {
	return fileDescriptor_stn_ff7f71da879f1d98, []int{0}
}
func (m *Rule) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Rule.Unmarshal(m, b)
}
func (m *Rule) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Rule.Marshal(b, m, deterministic)
}
func (dst *Rule) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Rule.Merge(dst, src)
}
func (m *Rule) XXX_Size() int {
	return xxx_messageInfo_Rule.Size(m)
}
func (m *Rule) XXX_DiscardUnknown() {
	xxx_messageInfo_Rule.DiscardUnknown(m)
}

var xxx_messageInfo_Rule proto.InternalMessageInfo

func (m *Rule) GetIpAddress() string {
	if m != nil {
		return m.IpAddress
	}
	return ""
}

func (m *Rule) GetInterface() string {
	if m != nil {
		return m.Interface
	}
	return ""
}

func (*Rule) XXX_MessageName() string {
	return "vpp.stn.Rule"
}
func init() {
	proto.RegisterType((*Rule)(nil), "vpp.stn.Rule")
}

func init() { proto.RegisterFile("vpp/stn/stn.proto", fileDescriptor_stn_ff7f71da879f1d98) }

var fileDescriptor_stn_ff7f71da879f1d98 = []byte{
	// 187 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2c, 0x2b, 0x28, 0xd0,
	0x2f, 0x2e, 0xc9, 0x03, 0x61, 0xbd, 0x82, 0xa2, 0xfc, 0x92, 0x7c, 0x21, 0xf6, 0xb2, 0x82, 0x02,
	0xbd, 0xe2, 0x92, 0x3c, 0x29, 0xdd, 0xf4, 0xcc, 0x92, 0x8c, 0xd2, 0x24, 0xbd, 0xe4, 0xfc, 0x5c,
	0xfd, 0xf4, 0xfc, 0xf4, 0x7c, 0x7d, 0xb0, 0x7c, 0x52, 0x69, 0x1a, 0x98, 0x07, 0xe6, 0x80, 0x59,
	0x10, 0x7d, 0x4a, 0xce, 0x5c, 0x2c, 0x41, 0xa5, 0x39, 0xa9, 0x42, 0xb2, 0x5c, 0x5c, 0x99, 0x05,
	0xf1, 0x89, 0x29, 0x29, 0x45, 0xa9, 0xc5, 0xc5, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x9c,
	0x99, 0x05, 0x8e, 0x10, 0x01, 0x21, 0x19, 0x2e, 0xce, 0xcc, 0xbc, 0x92, 0xd4, 0xa2, 0xb4, 0xc4,
	0xe4, 0x54, 0x09, 0x26, 0xa8, 0x2c, 0x4c, 0xc0, 0xc9, 0xe6, 0xc4, 0x63, 0x39, 0xc6, 0x28, 0x33,
	0x24, 0x9b, 0x73, 0x32, 0xd3, 0x13, 0x4b, 0xf2, 0xf5, 0xcb, 0x0a, 0x0a, 0x74, 0x13, 0xd3, 0x53,
	0xf3, 0x4a, 0xf4, 0x13, 0x0b, 0x32, 0xf5, 0x73, 0xf3, 0x53, 0x52, 0x73, 0x8a, 0xf5, 0xa1, 0x8e,
	0xb7, 0x2e, 0x2b, 0x28, 0x88, 0x2f, 0x2e, 0xc9, 0x4b, 0x62, 0x03, 0xbb, 0xc4, 0x18, 0x10, 0x00,
	0x00, 0xff, 0xff, 0x62, 0xa3, 0x2a, 0xa8, 0xd6, 0x00, 0x00, 0x00,
}