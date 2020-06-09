// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/custom/vpp/syslog/syslog.proto

package vpp_syslog

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

type Sender struct {
	Source               string   `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
	Collector            string   `protobuf:"bytes,2,opt,name=collector,proto3" json:"collector,omitempty"`
	Port                 int32    `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Sender) Reset()         { *m = Sender{} }
func (m *Sender) String() string { return proto.CompactTextString(m) }
func (*Sender) ProtoMessage()    {}
func (*Sender) Descriptor() ([]byte, []int) {
	return fileDescriptor_947997541bf72f4b, []int{0}
}

func (m *Sender) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Sender.Unmarshal(m, b)
}
func (m *Sender) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Sender.Marshal(b, m, deterministic)
}
func (m *Sender) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Sender.Merge(m, src)
}
func (m *Sender) XXX_Size() int {
	return xxx_messageInfo_Sender.Size(m)
}
func (m *Sender) XXX_DiscardUnknown() {
	xxx_messageInfo_Sender.DiscardUnknown(m)
}

var xxx_messageInfo_Sender proto.InternalMessageInfo

func (m *Sender) GetSource() string {
	if m != nil {
		return m.Source
	}
	return ""
}

func (m *Sender) GetCollector() string {
	if m != nil {
		return m.Collector
	}
	return ""
}

func (m *Sender) GetPort() int32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func init() {
	proto.RegisterType((*Sender)(nil), "custom.vpp.syslog.Sender")
}

func init() {
	proto.RegisterFile("proto/custom/vpp/syslog/syslog.proto", fileDescriptor_947997541bf72f4b)
}

var fileDescriptor_947997541bf72f4b = []byte{
	// 195 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x8f, 0x3d, 0x6f, 0x83, 0x30,
	0x10, 0x86, 0x45, 0x3f, 0x90, 0xf0, 0x56, 0x0f, 0x15, 0x43, 0x07, 0x54, 0x75, 0x60, 0xa9, 0x3d,
	0x30, 0x66, 0xcb, 0x4f, 0x20, 0x5b, 0x86, 0x10, 0x62, 0x4e, 0x16, 0x92, 0xf1, 0x9d, 0x6c, 0x63,
	0x91, 0x7f, 0x1f, 0x61, 0x90, 0x32, 0x65, 0xba, 0xf7, 0xee, 0x79, 0x87, 0xe7, 0xd8, 0x1f, 0x39,
	0x0c, 0x28, 0xd5, 0xec, 0x03, 0x4e, 0x32, 0x12, 0x49, 0x7f, 0xf7, 0x06, 0xf5, 0x3e, 0x44, 0xc2,
	0xfc, 0x6b, 0xe3, 0x22, 0x12, 0x89, 0x0d, 0xfc, 0xb6, 0x2c, 0x3f, 0x81, 0x1d, 0xc0, 0xf1, 0x6f,
	0x96, 0x7b, 0x9c, 0x9d, 0x82, 0x32, 0xab, 0xb2, 0xba, 0x68, 0xf7, 0x8d, 0xff, 0xb0, 0x42, 0xa1,
	0x31, 0xa0, 0x02, 0xba, 0xf2, 0x2d, 0xa1, 0xe7, 0x81, 0x73, 0xf6, 0x41, 0xe8, 0x42, 0xf9, 0x5e,
	0x65, 0xf5, 0x67, 0x9b, 0xf2, 0xf1, 0x7a, 0xbe, 0x68, 0x14, 0x66, 0xd4, 0x7d, 0x40, 0x31, 0xe2,
	0xaa, 0xf3, 0xdf, 0x6b, 0xb0, 0x41, 0xc6, 0x46, 0xc2, 0xd2, 0x4f, 0x64, 0xc0, 0x4b, 0x58, 0x02,
	0xd8, 0x61, 0x57, 0xee, 0x22, 0x51, 0x47, 0x66, 0xd6, 0xa3, 0x95, 0x2f, 0x5e, 0x39, 0xac, 0x95,
	0x2d, 0xde, 0xf2, 0xd4, 0x69, 0x1e, 0x01, 0x00, 0x00, 0xff, 0xff, 0x36, 0x5b, 0xb8, 0xf4, 0xf7,
	0x00, 0x00, 0x00,
}
