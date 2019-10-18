// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/models/vpp/l2/fib.proto

package vpp_l2

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

type FIBEntry_Action int32

const (
	FIBEntry_FORWARD FIBEntry_Action = 0
	FIBEntry_DROP    FIBEntry_Action = 1
)

var FIBEntry_Action_name = map[int32]string{
	0: "FORWARD",
	1: "DROP",
}

var FIBEntry_Action_value = map[string]int32{
	"FORWARD": 0,
	"DROP":    1,
}

func (x FIBEntry_Action) String() string {
	return proto.EnumName(FIBEntry_Action_name, int32(x))
}

func (FIBEntry_Action) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_4cdc3bc94799ee60, []int{0, 0}
}

type FIBEntry struct {
	PhysAddress             string          `protobuf:"bytes,1,opt,name=phys_address,json=physAddress,proto3" json:"phys_address,omitempty"`
	BridgeDomain            string          `protobuf:"bytes,2,opt,name=bridge_domain,json=bridgeDomain,proto3" json:"bridge_domain,omitempty"`
	Action                  FIBEntry_Action `protobuf:"varint,3,opt,name=action,proto3,enum=vpp.l2.FIBEntry_Action" json:"action,omitempty"`
	OutgoingInterface       string          `protobuf:"bytes,4,opt,name=outgoing_interface,json=outgoingInterface,proto3" json:"outgoing_interface,omitempty"`
	StaticConfig            bool            `protobuf:"varint,5,opt,name=static_config,json=staticConfig,proto3" json:"static_config,omitempty"`
	BridgedVirtualInterface bool            `protobuf:"varint,6,opt,name=bridged_virtual_interface,json=bridgedVirtualInterface,proto3" json:"bridged_virtual_interface,omitempty"`
	XXX_NoUnkeyedLiteral    struct{}        `json:"-"`
	XXX_unrecognized        []byte          `json:"-"`
	XXX_sizecache           int32           `json:"-"`
}

func (m *FIBEntry) Reset()         { *m = FIBEntry{} }
func (m *FIBEntry) String() string { return proto.CompactTextString(m) }
func (*FIBEntry) ProtoMessage()    {}
func (*FIBEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_4cdc3bc94799ee60, []int{0}
}

func (m *FIBEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FIBEntry.Unmarshal(m, b)
}
func (m *FIBEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FIBEntry.Marshal(b, m, deterministic)
}
func (m *FIBEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FIBEntry.Merge(m, src)
}
func (m *FIBEntry) XXX_Size() int {
	return xxx_messageInfo_FIBEntry.Size(m)
}
func (m *FIBEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_FIBEntry.DiscardUnknown(m)
}

var xxx_messageInfo_FIBEntry proto.InternalMessageInfo

func (m *FIBEntry) GetPhysAddress() string {
	if m != nil {
		return m.PhysAddress
	}
	return ""
}

func (m *FIBEntry) GetBridgeDomain() string {
	if m != nil {
		return m.BridgeDomain
	}
	return ""
}

func (m *FIBEntry) GetAction() FIBEntry_Action {
	if m != nil {
		return m.Action
	}
	return FIBEntry_FORWARD
}

func (m *FIBEntry) GetOutgoingInterface() string {
	if m != nil {
		return m.OutgoingInterface
	}
	return ""
}

func (m *FIBEntry) GetStaticConfig() bool {
	if m != nil {
		return m.StaticConfig
	}
	return false
}

func (m *FIBEntry) GetBridgedVirtualInterface() bool {
	if m != nil {
		return m.BridgedVirtualInterface
	}
	return false
}

func init() {
	proto.RegisterEnum("vpp.l2.FIBEntry_Action", FIBEntry_Action_name, FIBEntry_Action_value)
	proto.RegisterType((*FIBEntry)(nil), "vpp.l2.FIBEntry")
}

func init() { proto.RegisterFile("api/models/vpp/l2/fib.proto", fileDescriptor_4cdc3bc94799ee60) }

var fileDescriptor_4cdc3bc94799ee60 = []byte{
	// 309 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0xc1, 0x4f, 0xfa, 0x30,
	0x1c, 0xc5, 0x7f, 0xe3, 0x87, 0x13, 0x0b, 0x1a, 0xec, 0x85, 0x19, 0x0f, 0x22, 0x5e, 0xb8, 0xb0,
	0x26, 0xd3, 0x78, 0xd0, 0x13, 0x88, 0x24, 0x9c, 0x30, 0x3b, 0x68, 0xe2, 0x65, 0xe9, 0xb6, 0x52,
	0xbe, 0x49, 0x69, 0x9b, 0xae, 0x2c, 0xe1, 0xff, 0xf2, 0x0f, 0x34, 0xb4, 0xcc, 0x98, 0x78, 0xfd,
	0xbc, 0xf7, 0xfd, 0xbe, 0x97, 0x87, 0xae, 0xa9, 0x06, 0xb2, 0x55, 0x25, 0x13, 0x15, 0xa9, 0xb5,
	0x26, 0x22, 0x21, 0x6b, 0xc8, 0x63, 0x6d, 0x94, 0x55, 0x38, 0xac, 0xb5, 0x8e, 0x45, 0x32, 0xfa,
	0x6a, 0xa1, 0xce, 0x62, 0x39, 0x7b, 0x95, 0xd6, 0xec, 0xf1, 0x2d, 0xea, 0xe9, 0xcd, 0xbe, 0xca,
	0x68, 0x59, 0x1a, 0x56, 0x55, 0x51, 0x30, 0x0c, 0xc6, 0x67, 0x69, 0xf7, 0xc0, 0xa6, 0x1e, 0xe1,
	0x3b, 0x74, 0x9e, 0x1b, 0x28, 0x39, 0xcb, 0x4a, 0xb5, 0xa5, 0x20, 0xa3, 0x96, 0xf3, 0xf4, 0x3c,
	0x9c, 0x3b, 0x86, 0x09, 0x0a, 0x69, 0x61, 0x41, 0xc9, 0xe8, 0xff, 0x30, 0x18, 0x5f, 0x24, 0x83,
	0xd8, 0xa7, 0xc5, 0x4d, 0x52, 0x3c, 0x75, 0x72, 0x7a, 0xb4, 0xe1, 0x09, 0xc2, 0x6a, 0x67, 0xb9,
	0x02, 0xc9, 0x33, 0x90, 0x96, 0x99, 0x35, 0x2d, 0x58, 0xd4, 0x76, 0xaf, 0x2f, 0x1b, 0x65, 0xd9,
	0x08, 0x87, 0x12, 0x95, 0xa5, 0x16, 0x8a, 0xac, 0x50, 0x72, 0x0d, 0x3c, 0x3a, 0x19, 0x06, 0xe3,
	0x4e, 0xda, 0xf3, 0xf0, 0xc5, 0x31, 0xfc, 0x84, 0xae, 0x7c, 0xa9, 0x32, 0xab, 0xc1, 0xd8, 0x1d,
	0x15, 0xbf, 0x5e, 0x87, 0xee, 0x60, 0x70, 0x34, 0xbc, 0x7b, 0xfd, 0x27, 0x60, 0x74, 0x83, 0x42,
	0xdf, 0x10, 0x77, 0xd1, 0xe9, 0x62, 0x95, 0x7e, 0x4c, 0xd3, 0x79, 0xff, 0x1f, 0xee, 0xa0, 0xf6,
	0x3c, 0x5d, 0xbd, 0xf5, 0x83, 0xd9, 0xe3, 0xe7, 0x03, 0x07, 0xbb, 0xd9, 0xe5, 0x71, 0xa1, 0xb6,
	0x44, 0x00, 0xa7, 0x56, 0x1d, 0x46, 0x9e, 0x50, 0xce, 0xa4, 0x25, 0x7f, 0x96, 0x7f, 0xae, 0xb5,
	0xce, 0x44, 0x92, 0x87, 0x6e, 0xfd, 0xfb, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0x98, 0xdf, 0x69,
	0x91, 0x9c, 0x01, 0x00, 0x00,
}
