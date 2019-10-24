// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/plugins/kvscheduler/internal/test/model/values.proto

package model

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

type ArrayValue struct {
	Items                []string `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	ItemSuffix           string   `protobuf:"bytes,2,opt,name=item_suffix,json=itemSuffix,proto3" json:"item_suffix,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ArrayValue) Reset()         { *m = ArrayValue{} }
func (m *ArrayValue) String() string { return proto.CompactTextString(m) }
func (*ArrayValue) ProtoMessage()    {}
func (*ArrayValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_04183cdf1f63b588, []int{0}
}

func (m *ArrayValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ArrayValue.Unmarshal(m, b)
}
func (m *ArrayValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ArrayValue.Marshal(b, m, deterministic)
}
func (m *ArrayValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ArrayValue.Merge(m, src)
}
func (m *ArrayValue) XXX_Size() int {
	return xxx_messageInfo_ArrayValue.Size(m)
}
func (m *ArrayValue) XXX_DiscardUnknown() {
	xxx_messageInfo_ArrayValue.DiscardUnknown(m)
}

var xxx_messageInfo_ArrayValue proto.InternalMessageInfo

func (m *ArrayValue) GetItems() []string {
	if m != nil {
		return m.Items
	}
	return nil
}

func (m *ArrayValue) GetItemSuffix() string {
	if m != nil {
		return m.ItemSuffix
	}
	return ""
}

type StringValue struct {
	Value                string   `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StringValue) Reset()         { *m = StringValue{} }
func (m *StringValue) String() string { return proto.CompactTextString(m) }
func (*StringValue) ProtoMessage()    {}
func (*StringValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_04183cdf1f63b588, []int{1}
}

func (m *StringValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StringValue.Unmarshal(m, b)
}
func (m *StringValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StringValue.Marshal(b, m, deterministic)
}
func (m *StringValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StringValue.Merge(m, src)
}
func (m *StringValue) XXX_Size() int {
	return xxx_messageInfo_StringValue.Size(m)
}
func (m *StringValue) XXX_DiscardUnknown() {
	xxx_messageInfo_StringValue.DiscardUnknown(m)
}

var xxx_messageInfo_StringValue proto.InternalMessageInfo

func (m *StringValue) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func init() {
	proto.RegisterType((*ArrayValue)(nil), "model.ArrayValue")
	proto.RegisterType((*StringValue)(nil), "model.StringValue")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/plugins/kvscheduler/internal/test/model/values.proto", fileDescriptor_04183cdf1f63b588)
}

var fileDescriptor_04183cdf1f63b588 = []byte{
	// 175 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x8d, 0xb1, 0x0a, 0xc2, 0x30,
	0x10, 0x40, 0xa9, 0x52, 0xa1, 0xe9, 0x56, 0x1c, 0xba, 0x59, 0xea, 0xd2, 0x45, 0x6f, 0xf0, 0x0b,
	0x44, 0xfc, 0x81, 0x16, 0x5c, 0x25, 0xda, 0x6b, 0x0c, 0xa6, 0x49, 0x48, 0x2e, 0x45, 0xff, 0x5e,
	0x9a, 0x82, 0xdb, 0xbd, 0x77, 0xc7, 0x3d, 0x76, 0x55, 0x52, 0x70, 0x32, 0x30, 0x59, 0x7b, 0xe0,
	0x02, 0x35, 0x81, 0x55, 0x41, 0x48, 0xed, 0xe1, 0x3d, 0xf9, 0xe7, 0x0b, 0xfb, 0xa0, 0xd0, 0x81,
	0xd4, 0x84, 0x4e, 0x73, 0x05, 0x84, 0x9e, 0x60, 0x34, 0x3d, 0x2a, 0x98, 0xb8, 0x0a, 0xe8, 0x8f,
	0xd6, 0x19, 0x32, 0x45, 0x1a, 0x5d, 0x7d, 0x61, 0xec, 0xec, 0x1c, 0xff, 0xde, 0xe6, 0x5d, 0xb1,
	0x65, 0xa9, 0x24, 0x1c, 0x7d, 0x99, 0x54, 0xeb, 0x26, 0x6b, 0x17, 0x28, 0x76, 0x2c, 0x9f, 0x87,
	0xbb, 0x0f, 0xc3, 0x20, 0x3f, 0xe5, 0xaa, 0x4a, 0x9a, 0xac, 0x65, 0xb3, 0xea, 0xa2, 0xa9, 0xf7,
	0x2c, 0xef, 0xc8, 0x49, 0x2d, 0xfe, 0x5f, 0x62, 0xaa, 0x4c, 0xe2, 0xe5, 0x02, 0x8f, 0x4d, 0xec,
	0x9e, 0x7e, 0x01, 0x00, 0x00, 0xff, 0xff, 0xe4, 0x3d, 0x69, 0xc8, 0xc0, 0x00, 0x00, 0x00,
}
