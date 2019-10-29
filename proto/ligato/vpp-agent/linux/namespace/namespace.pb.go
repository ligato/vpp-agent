// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/linux/namespace/namespace.proto

package linux_namespace

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

type NetNamespace_ReferenceType int32

const (
	NetNamespace_UNDEFINED    NetNamespace_ReferenceType = 0
	NetNamespace_NSID         NetNamespace_ReferenceType = 1
	NetNamespace_PID          NetNamespace_ReferenceType = 2
	NetNamespace_FD           NetNamespace_ReferenceType = 3
	NetNamespace_MICROSERVICE NetNamespace_ReferenceType = 4
)

var NetNamespace_ReferenceType_name = map[int32]string{
	0: "UNDEFINED",
	1: "NSID",
	2: "PID",
	3: "FD",
	4: "MICROSERVICE",
}

var NetNamespace_ReferenceType_value = map[string]int32{
	"UNDEFINED":    0,
	"NSID":         1,
	"PID":          2,
	"FD":           3,
	"MICROSERVICE": 4,
}

func (x NetNamespace_ReferenceType) String() string {
	return proto.EnumName(NetNamespace_ReferenceType_name, int32(x))
}

func (NetNamespace_ReferenceType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_2ccf07f5f5d26ae3, []int{0, 0}
}

type NetNamespace struct {
	Type NetNamespace_ReferenceType `protobuf:"varint,1,opt,name=type,proto3,enum=ligato.vpp_agent.linux.namespace.NetNamespace_ReferenceType" json:"type,omitempty"`
	// Reference defines reference specific
	// to the namespace type:
	//  * namespace ID (NSID)
	//  * PID number (PID)
	//  * file path (FD)
	//  * microservice label (MICROSERVICE)
	Reference            string   `protobuf:"bytes,2,opt,name=reference,proto3" json:"reference,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NetNamespace) Reset()         { *m = NetNamespace{} }
func (m *NetNamespace) String() string { return proto.CompactTextString(m) }
func (*NetNamespace) ProtoMessage()    {}
func (*NetNamespace) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ccf07f5f5d26ae3, []int{0}
}

func (m *NetNamespace) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NetNamespace.Unmarshal(m, b)
}
func (m *NetNamespace) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NetNamespace.Marshal(b, m, deterministic)
}
func (m *NetNamespace) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NetNamespace.Merge(m, src)
}
func (m *NetNamespace) XXX_Size() int {
	return xxx_messageInfo_NetNamespace.Size(m)
}
func (m *NetNamespace) XXX_DiscardUnknown() {
	xxx_messageInfo_NetNamespace.DiscardUnknown(m)
}

var xxx_messageInfo_NetNamespace proto.InternalMessageInfo

func (m *NetNamespace) GetType() NetNamespace_ReferenceType {
	if m != nil {
		return m.Type
	}
	return NetNamespace_UNDEFINED
}

func (m *NetNamespace) GetReference() string {
	if m != nil {
		return m.Reference
	}
	return ""
}

func init() {
	proto.RegisterEnum("ligato.vpp_agent.linux.namespace.NetNamespace_ReferenceType", NetNamespace_ReferenceType_name, NetNamespace_ReferenceType_value)
	proto.RegisterType((*NetNamespace)(nil), "ligato.vpp_agent.linux.namespace.NetNamespace")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/linux/namespace/namespace.proto", fileDescriptor_2ccf07f5f5d26ae3)
}

var fileDescriptor_2ccf07f5f5d26ae3 = []byte{
	// 247 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x32, 0xc8, 0xc9, 0x4c, 0x4f,
	0x2c, 0xc9, 0xd7, 0x2f, 0x2b, 0x28, 0xd0, 0x4d, 0x4c, 0x4f, 0xcd, 0x2b, 0xd1, 0xcf, 0xc9, 0xcc,
	0x2b, 0xad, 0xd0, 0xcf, 0x4b, 0xcc, 0x4d, 0x2d, 0x2e, 0x48, 0x4c, 0x4e, 0x45, 0xb0, 0xf4, 0x0a,
	0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0x14, 0x20, 0x3a, 0xf4, 0xca, 0x0a, 0x0a, 0xe2, 0xc1, 0x3a, 0xf4,
	0xc0, 0x3a, 0xf4, 0xe0, 0xea, 0x94, 0x4e, 0x33, 0x72, 0xf1, 0xf8, 0xa5, 0x96, 0xf8, 0xc1, 0x04,
	0x84, 0x02, 0xb8, 0x58, 0x4a, 0x2a, 0x0b, 0x52, 0x25, 0x18, 0x15, 0x18, 0x35, 0xf8, 0x8c, 0x6c,
	0xf4, 0x08, 0x99, 0xa0, 0x87, 0xac, 0x5b, 0x2f, 0x28, 0x35, 0x2d, 0xb5, 0x28, 0x35, 0x2f, 0x39,
	0x35, 0xa4, 0xb2, 0x20, 0x35, 0x08, 0x6c, 0x92, 0x90, 0x0c, 0x17, 0x67, 0x11, 0x4c, 0x58, 0x82,
	0x49, 0x81, 0x51, 0x83, 0x33, 0x08, 0x21, 0xa0, 0xe4, 0xcd, 0xc5, 0x8b, 0xa2, 0x49, 0x88, 0x97,
	0x8b, 0x33, 0xd4, 0xcf, 0xc5, 0xd5, 0xcd, 0xd3, 0xcf, 0xd5, 0x45, 0x80, 0x41, 0x88, 0x83, 0x8b,
	0xc5, 0x2f, 0xd8, 0xd3, 0x45, 0x80, 0x51, 0x88, 0x9d, 0x8b, 0x39, 0xc0, 0xd3, 0x45, 0x80, 0x49,
	0x88, 0x8d, 0x8b, 0xc9, 0xcd, 0x45, 0x80, 0x59, 0x48, 0x80, 0x8b, 0xc7, 0xd7, 0xd3, 0x39, 0xc8,
	0x3f, 0xd8, 0x35, 0x28, 0xcc, 0xd3, 0xd9, 0x55, 0x80, 0xc5, 0x29, 0x28, 0x2a, 0x20, 0x3d, 0x1f,
	0xe6, 0xe4, 0x4c, 0xe4, 0x90, 0x2a, 0x33, 0xd2, 0x07, 0x07, 0x89, 0x3e, 0xa1, 0x30, 0xb4, 0x06,
	0xf3, 0xe3, 0xe1, 0xfc, 0x24, 0x36, 0xb0, 0x3e, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xd1,
	0x0b, 0xe0, 0x32, 0x7e, 0x01, 0x00, 0x00,
}