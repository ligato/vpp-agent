// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/govppmux/metrics.proto

package govppmux

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

type Metrics struct {
	ChannelsCreated      uint64   `protobuf:"varint,1,opt,name=channels_created,json=channelsCreated,proto3" json:"channels_created,omitempty"`
	ChannelsOpen         uint64   `protobuf:"varint,2,opt,name=channels_open,json=channelsOpen,proto3" json:"channels_open,omitempty"`
	RequestsSent         uint64   `protobuf:"varint,3,opt,name=requests_sent,json=requestsSent,proto3" json:"requests_sent,omitempty"`
	RequestsDone         uint64   `protobuf:"varint,4,opt,name=requests_done,json=requestsDone,proto3" json:"requests_done,omitempty"`
	RequestsFail         uint64   `protobuf:"varint,5,opt,name=requests_fail,json=requestsFail,proto3" json:"requests_fail,omitempty"`
	RequestReplies       uint64   `protobuf:"varint,6,opt,name=request_replies,json=requestReplies,proto3" json:"request_replies,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Metrics) Reset()         { *m = Metrics{} }
func (m *Metrics) String() string { return proto.CompactTextString(m) }
func (*Metrics) ProtoMessage()    {}
func (*Metrics) Descriptor() ([]byte, []int) {
	return fileDescriptor_9db95072bd392a48, []int{0}
}

func (m *Metrics) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Metrics.Unmarshal(m, b)
}
func (m *Metrics) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Metrics.Marshal(b, m, deterministic)
}
func (m *Metrics) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metrics.Merge(m, src)
}
func (m *Metrics) XXX_Size() int {
	return xxx_messageInfo_Metrics.Size(m)
}
func (m *Metrics) XXX_DiscardUnknown() {
	xxx_messageInfo_Metrics.DiscardUnknown(m)
}

var xxx_messageInfo_Metrics proto.InternalMessageInfo

func (m *Metrics) GetChannelsCreated() uint64 {
	if m != nil {
		return m.ChannelsCreated
	}
	return 0
}

func (m *Metrics) GetChannelsOpen() uint64 {
	if m != nil {
		return m.ChannelsOpen
	}
	return 0
}

func (m *Metrics) GetRequestsSent() uint64 {
	if m != nil {
		return m.RequestsSent
	}
	return 0
}

func (m *Metrics) GetRequestsDone() uint64 {
	if m != nil {
		return m.RequestsDone
	}
	return 0
}

func (m *Metrics) GetRequestsFail() uint64 {
	if m != nil {
		return m.RequestsFail
	}
	return 0
}

func (m *Metrics) GetRequestReplies() uint64 {
	if m != nil {
		return m.RequestReplies
	}
	return 0
}

func init() {
	proto.RegisterType((*Metrics)(nil), "ligato.vpp_agent.govppmux.Metrics")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/govppmux/metrics.proto", fileDescriptor_9db95072bd392a48)
}

var fileDescriptor_9db95072bd392a48 = []byte{
	// 240 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0xd0, 0xb1, 0x4b, 0xc4, 0x30,
	0x14, 0xc7, 0x71, 0xaa, 0xe7, 0x09, 0x41, 0x3d, 0xe9, 0x14, 0x37, 0xd1, 0xe1, 0x74, 0xb0, 0x05,
	0x9d, 0xe4, 0x36, 0x15, 0x37, 0x11, 0xce, 0xcd, 0xa5, 0xc4, 0xf6, 0x19, 0x03, 0xb9, 0xf7, 0x9e,
	0x49, 0x5a, 0xfc, 0xb7, 0xfd, 0x0f, 0xc4, 0xa6, 0x39, 0x2e, 0xe3, 0xef, 0xcb, 0x67, 0xfa, 0x89,
	0xa5, 0x35, 0x5a, 0x05, 0xaa, 0x07, 0xe6, 0x1b, 0xa5, 0x01, 0x43, 0xad, 0x69, 0x60, 0xde, 0xf4,
	0x3f, 0xf5, 0x06, 0x82, 0x33, 0xad, 0xaf, 0xd8, 0x51, 0xa0, 0xf2, 0x2c, 0xc2, 0x6a, 0x60, 0x6e,
	0x46, 0x58, 0x25, 0x78, 0xf1, 0x5b, 0x88, 0xc3, 0x97, 0x88, 0xcb, 0x6b, 0x71, 0xda, 0x7e, 0x29,
	0x44, 0xb0, 0xbe, 0x69, 0x1d, 0xa8, 0x00, 0x9d, 0x2c, 0xce, 0x8b, 0xab, 0xd9, 0x7a, 0x91, 0xfa,
	0x63, 0xcc, 0xe5, 0xa5, 0x38, 0xde, 0x52, 0x62, 0x40, 0xb9, 0x37, 0xba, 0xa3, 0x14, 0x5f, 0x19,
	0xf0, 0x1f, 0x39, 0xf8, 0xee, 0xc1, 0x07, 0xdf, 0x78, 0xc0, 0x20, 0xf7, 0x23, 0x4a, 0xf1, 0x0d,
	0x30, 0x64, 0xa8, 0x23, 0x04, 0x39, 0xcb, 0xd1, 0x13, 0x21, 0x64, 0xe8, 0x53, 0x19, 0x2b, 0x0f,
	0x72, 0xf4, 0xac, 0x8c, 0x2d, 0x97, 0x62, 0x31, 0xed, 0xc6, 0x01, 0x5b, 0x03, 0x5e, 0xce, 0x47,
	0x76, 0x32, 0xe5, 0x75, 0xac, 0x0f, 0xab, 0xf7, 0x7b, 0x4d, 0xd5, 0xf4, 0x89, 0xd9, 0xfd, 0x6f,
	0xb8, 0xad, 0xd9, 0xf6, 0xda, 0xa0, 0xdf, 0xb9, 0x92, 0x3a, 0xb0, 0xab, 0x34, 0x3f, 0xe6, 0xe3,
	0xa5, 0x77, 0x7f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x72, 0x8d, 0x55, 0x1f, 0x7d, 0x01, 0x00, 0x00,
}
