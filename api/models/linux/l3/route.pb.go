// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: models/linux/l3/route.proto

package linux_l3

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
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
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type Route_Scope int32

const (
	Route_UNDEFINED Route_Scope = 0
	Route_GLOBAL    Route_Scope = 1
	Route_SITE      Route_Scope = 2
	Route_LINK      Route_Scope = 3
	Route_HOST      Route_Scope = 4
)

var Route_Scope_name = map[int32]string{
	0: "UNDEFINED",
	1: "GLOBAL",
	2: "SITE",
	3: "LINK",
	4: "HOST",
}

var Route_Scope_value = map[string]int32{
	"UNDEFINED": 0,
	"GLOBAL":    1,
	"SITE":      2,
	"LINK":      3,
	"HOST":      4,
}

func (x Route_Scope) String() string {
	return proto.EnumName(Route_Scope_name, int32(x))
}

func (Route_Scope) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_ebf8500d01fc3585, []int{0, 0}
}

type Route struct {
	// Outgoing interface logical name (mandatory).
	OutgoingInterface string `protobuf:"bytes,1,opt,name=outgoing_interface,json=outgoingInterface,proto3" json:"outgoing_interface,omitempty"`
	// The scope of the area where the link is valid.
	Scope Route_Scope `protobuf:"varint,2,opt,name=scope,proto3,enum=linux.l3.Route_Scope" json:"scope,omitempty"`
	// Destination network address in the format <address>/<prefix> (mandatory)
	// Address can be also allocated via netalloc plugin and referenced here,
	// see: api/models/netalloc/netalloc.proto
	DstNetwork string `protobuf:"bytes,3,opt,name=dst_network,json=dstNetwork,proto3" json:"dst_network,omitempty"`
	// Gateway IP address (without mask, optional).
	// Address can be also allocated via netalloc plugin and referenced here,
	// see: api/models/netalloc/netalloc.proto
	GwAddr string `protobuf:"bytes,4,opt,name=gw_addr,json=gwAddr,proto3" json:"gw_addr,omitempty"`
	// routing metric (weight)
	Metric               uint32   `protobuf:"varint,5,opt,name=metric,proto3" json:"metric,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Route) Reset()         { *m = Route{} }
func (m *Route) String() string { return proto.CompactTextString(m) }
func (*Route) ProtoMessage()    {}
func (*Route) Descriptor() ([]byte, []int) {
	return fileDescriptor_ebf8500d01fc3585, []int{0}
}
func (m *Route) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Route.Unmarshal(m, b)
}
func (m *Route) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Route.Marshal(b, m, deterministic)
}
func (m *Route) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Route.Merge(m, src)
}
func (m *Route) XXX_Size() int {
	return xxx_messageInfo_Route.Size(m)
}
func (m *Route) XXX_DiscardUnknown() {
	xxx_messageInfo_Route.DiscardUnknown(m)
}

var xxx_messageInfo_Route proto.InternalMessageInfo

func (m *Route) GetOutgoingInterface() string {
	if m != nil {
		return m.OutgoingInterface
	}
	return ""
}

func (m *Route) GetScope() Route_Scope {
	if m != nil {
		return m.Scope
	}
	return Route_UNDEFINED
}

func (m *Route) GetDstNetwork() string {
	if m != nil {
		return m.DstNetwork
	}
	return ""
}

func (m *Route) GetGwAddr() string {
	if m != nil {
		return m.GwAddr
	}
	return ""
}

func (m *Route) GetMetric() uint32 {
	if m != nil {
		return m.Metric
	}
	return 0
}

func (*Route) XXX_MessageName() string {
	return "linux.l3.Route"
}
func init() {
	proto.RegisterEnum("linux.l3.Route_Scope", Route_Scope_name, Route_Scope_value)
	proto.RegisterType((*Route)(nil), "linux.l3.Route")
}

func init() { proto.RegisterFile("models/linux/l3/route.proto", fileDescriptor_ebf8500d01fc3585) }

var fileDescriptor_ebf8500d01fc3585 = []byte{
	// 326 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0x41, 0x4b, 0xeb, 0x40,
	0x14, 0x85, 0xdf, 0xb4, 0x4d, 0x5e, 0x7b, 0x1f, 0x7d, 0xe4, 0x0d, 0x3c, 0x0d, 0x0a, 0x5a, 0xba,
	0x2a, 0x48, 0x33, 0x60, 0x36, 0x82, 0x20, 0xb6, 0xb4, 0x6a, 0xb0, 0xa4, 0x90, 0xd6, 0x8d, 0x9b,
	0x90, 0x66, 0xa6, 0xe3, 0x60, 0x9a, 0x09, 0x93, 0x89, 0xf5, 0x27, 0xfa, 0x3f, 0xfc, 0x11, 0x6e,
	0x25, 0x53, 0x0b, 0xe2, 0xee, 0x9c, 0xfb, 0xdd, 0x39, 0x73, 0x39, 0x70, 0xbc, 0x91, 0x94, 0x65,
	0x25, 0xc9, 0x44, 0x5e, 0xbd, 0x92, 0xcc, 0x27, 0x4a, 0x56, 0x9a, 0x79, 0x85, 0x92, 0x5a, 0xe2,
	0xb6, 0x99, 0x7a, 0x99, 0x7f, 0x34, 0xe4, 0x42, 0x3f, 0x55, 0x2b, 0x2f, 0x95, 0x1b, 0xc2, 0x25,
	0x97, 0xc4, 0x2c, 0xac, 0xaa, 0xb5, 0x71, 0xc6, 0x18, 0xb5, 0x7b, 0xd8, 0xff, 0x40, 0x60, 0x45,
	0x75, 0x10, 0x1e, 0x02, 0x96, 0x95, 0xe6, 0x52, 0xe4, 0x3c, 0x16, 0xb9, 0x66, 0x6a, 0x9d, 0xa4,
	0xcc, 0x45, 0x3d, 0x34, 0xe8, 0x44, 0xff, 0xf6, 0x24, 0xd8, 0x03, 0x7c, 0x06, 0x56, 0x99, 0xca,
	0x82, 0xb9, 0x8d, 0x1e, 0x1a, 0xfc, 0x3d, 0xff, 0xef, 0xed, 0x2f, 0xf0, 0x4c, 0x9c, 0xb7, 0xa8,
	0x61, 0xb4, 0xdb, 0xc1, 0xa7, 0xf0, 0x87, 0x96, 0x3a, 0xce, 0x99, 0xde, 0x4a, 0xf5, 0xec, 0x36,
	0x4d, 0x28, 0xd0, 0x52, 0x87, 0xbb, 0x09, 0x3e, 0x84, 0xdf, 0x7c, 0x1b, 0x27, 0x94, 0x2a, 0xb7,
	0x65, 0xa0, 0xcd, 0xb7, 0x23, 0x4a, 0x15, 0x3e, 0x00, 0x7b, 0xc3, 0xb4, 0x12, 0xa9, 0x6b, 0xf5,
	0xd0, 0xa0, 0x1b, 0x7d, 0xb9, 0xfe, 0x35, 0x58, 0xe6, 0x07, 0xdc, 0x85, 0xce, 0x43, 0x38, 0x99,
	0xde, 0x04, 0xe1, 0x74, 0xe2, 0xfc, 0xc2, 0x00, 0xf6, 0xed, 0x6c, 0x3e, 0x1e, 0xcd, 0x1c, 0x84,
	0xdb, 0xd0, 0x5a, 0x04, 0xcb, 0xa9, 0xd3, 0xa8, 0xd5, 0x2c, 0x08, 0xef, 0x9d, 0x66, 0xad, 0xee,
	0xe6, 0x8b, 0xa5, 0xd3, 0x1a, 0x5f, 0xbd, 0xbd, 0x9f, 0xa0, 0xc7, 0x8b, 0x6f, 0x75, 0x65, 0x82,
	0x27, 0x5a, 0x92, 0x97, 0xa2, 0x18, 0x26, 0x9c, 0xe5, 0x9a, 0x24, 0x85, 0x20, 0x3f, 0x5a, 0xbf,
	0x34, 0x22, 0xce, 0xfc, 0x95, 0x6d, 0x0a, 0xf4, 0x3f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x63, 0x31,
	0x49, 0xba, 0x98, 0x01, 0x00, 0x00,
}
