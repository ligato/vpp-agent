// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/examples/kvscheduler/plugin_skeleton/with_metadata/model/model.proto

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

type ValueSkeleton struct {
	//
	//logical name is often defined to build unique keys for value instances
	//- alternativelly, in the value model (keys.go), you may apply the
	//WithNameTemplate() option to generate value instance name using a golang
	//template, combining multiple value attributes that collectively
	//guarantee unique value identification (i.e. primary key)
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValueSkeleton) Reset()         { *m = ValueSkeleton{} }
func (m *ValueSkeleton) String() string { return proto.CompactTextString(m) }
func (*ValueSkeleton) ProtoMessage()    {}
func (*ValueSkeleton) Descriptor() ([]byte, []int) {
	return fileDescriptor_f25cb7e1ea01c299, []int{0}
}

func (m *ValueSkeleton) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValueSkeleton.Unmarshal(m, b)
}
func (m *ValueSkeleton) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValueSkeleton.Marshal(b, m, deterministic)
}
func (m *ValueSkeleton) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValueSkeleton.Merge(m, src)
}
func (m *ValueSkeleton) XXX_Size() int {
	return xxx_messageInfo_ValueSkeleton.Size(m)
}
func (m *ValueSkeleton) XXX_DiscardUnknown() {
	xxx_messageInfo_ValueSkeleton.DiscardUnknown(m)
}

var xxx_messageInfo_ValueSkeleton proto.InternalMessageInfo

func (m *ValueSkeleton) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterType((*ValueSkeleton)(nil), "model.ValueSkeleton")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/examples/kvscheduler/plugin_skeleton/with_metadata/model/model.proto", fileDescriptor_f25cb7e1ea01c299)
}

var fileDescriptor_f25cb7e1ea01c299 = []byte{
	// 142 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x2c, 0xcb, 0xbd, 0x0a, 0xc2, 0x30,
	0x14, 0xc5, 0x71, 0x0a, 0x2a, 0x58, 0x70, 0xe9, 0xe4, 0x28, 0xba, 0xb8, 0xe8, 0x1d, 0x7c, 0x13,
	0x45, 0xd7, 0x72, 0x35, 0x87, 0x34, 0xf4, 0xe6, 0x83, 0xe6, 0xa6, 0xfa, 0xf8, 0x42, 0x74, 0x39,
	0x9c, 0xff, 0xf0, 0x6b, 0xef, 0xe2, 0x2c, 0x6b, 0xa4, 0x39, 0xa5, 0x13, 0x5b, 0x04, 0x25, 0x7c,
	0xd8, 0x27, 0x41, 0xa6, 0x71, 0xce, 0xaf, 0x01, 0xa6, 0x08, 0x26, 0x4a, 0x52, 0xac, 0x0b, 0x7d,
	0x1e, 0x21, 0xd0, 0x18, 0xe8, 0xed, 0x74, 0xe8, 0x3d, 0x94, 0x0d, 0x2b, 0x93, 0x8f, 0x06, 0xf2,
	0xdb, 0x73, 0x9a, 0xa2, 0xc6, 0x6e, 0x59, 0x63, 0x7f, 0x68, 0x37, 0x0f, 0x96, 0x82, 0xdb, 0xdf,
	0x75, 0x5d, 0xbb, 0x08, 0xec, 0xb1, 0x6d, 0x76, 0xcd, 0x71, 0x7d, 0xad, 0xff, 0xb9, 0xaa, 0xe4,
	0xf2, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x87, 0xbc, 0x46, 0x8d, 0x8b, 0x00, 0x00, 0x00,
}
