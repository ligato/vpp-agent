// Code generated by protoc-gen-go. DO NOT EDIT.
// source: plugins/kvscheduler/internal/test/model/values.proto

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
	return fileDescriptor_13c433c6c49f4d7d, []int{0}
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
	return fileDescriptor_13c433c6c49f4d7d, []int{1}
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
	proto.RegisterFile("plugins/kvscheduler/internal/test/model/values.proto", fileDescriptor_13c433c6c49f4d7d)
}

var fileDescriptor_13c433c6c49f4d7d = []byte{
	// 162 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x32, 0x29, 0xc8, 0x29, 0x4d,
	0xcf, 0xcc, 0x2b, 0xd6, 0xcf, 0x2e, 0x2b, 0x4e, 0xce, 0x48, 0x4d, 0x29, 0xcd, 0x49, 0x2d, 0xd2,
	0xcf, 0xcc, 0x2b, 0x49, 0x2d, 0xca, 0x4b, 0xcc, 0xd1, 0x2f, 0x49, 0x2d, 0x2e, 0xd1, 0xcf, 0xcd,
	0x4f, 0x49, 0xcd, 0xd1, 0x2f, 0x4b, 0xcc, 0x29, 0x4d, 0x2d, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9,
	0x17, 0x62, 0x05, 0x8b, 0x29, 0x39, 0x73, 0x71, 0x39, 0x16, 0x15, 0x25, 0x56, 0x86, 0x81, 0xe4,
	0x84, 0x44, 0xb8, 0x58, 0x33, 0x4b, 0x52, 0x73, 0x8b, 0x25, 0x18, 0x15, 0x98, 0x35, 0x38, 0x83,
	0x20, 0x1c, 0x21, 0x79, 0x2e, 0x6e, 0x10, 0x23, 0xbe, 0xb8, 0x34, 0x2d, 0x2d, 0xb3, 0x42, 0x82,
	0x49, 0x81, 0x51, 0x83, 0x33, 0x88, 0x0b, 0x24, 0x14, 0x0c, 0x16, 0x51, 0x52, 0xe6, 0xe2, 0x0e,
	0x2e, 0x29, 0xca, 0xcc, 0x4b, 0x87, 0x9b, 0x02, 0xb6, 0x4a, 0x82, 0x11, 0xac, 0x12, 0xc2, 0x49,
	0x62, 0x03, 0xdb, 0x6b, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x1b, 0x09, 0x57, 0x5d, 0xaf, 0x00,
	0x00, 0x00,
}
