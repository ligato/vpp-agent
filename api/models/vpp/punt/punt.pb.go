// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/api/models/vpp/punt/punt.proto

package vpp_punt

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

// L3Protocol defines Layer 3 protocols.
type L3Protocol int32

const (
	L3Protocol_UNDEFINED_L3 L3Protocol = 0
	L3Protocol_IPv4         L3Protocol = 4
	L3Protocol_IPv6         L3Protocol = 6
	L3Protocol_ALL          L3Protocol = 10
)

var L3Protocol_name = map[int32]string{
	0:  "UNDEFINED_L3",
	4:  "IPv4",
	6:  "IPv6",
	10: "ALL",
}

var L3Protocol_value = map[string]int32{
	"UNDEFINED_L3": 0,
	"IPv4":         4,
	"IPv6":         6,
	"ALL":          10,
}

func (x L3Protocol) String() string {
	return proto.EnumName(L3Protocol_name, int32(x))
}

func (L3Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_55e6bedee92e220c, []int{0}
}

// L4Protocol defines Layer 4 protocols.
type L4Protocol int32

const (
	L4Protocol_UNDEFINED_L4 L4Protocol = 0
	L4Protocol_TCP          L4Protocol = 6
	L4Protocol_UDP          L4Protocol = 17
)

var L4Protocol_name = map[int32]string{
	0:  "UNDEFINED_L4",
	6:  "TCP",
	17: "UDP",
}

var L4Protocol_value = map[string]int32{
	"UNDEFINED_L4": 0,
	"TCP":          6,
	"UDP":          17,
}

func (x L4Protocol) String() string {
	return proto.EnumName(L4Protocol_name, int32(x))
}

func (L4Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_55e6bedee92e220c, []int{1}
}

// IPRedirect allows otherwise dropped packet which destination IP address
// matching some of the VPP addresses to redirect to the defined next hop address
// via the TX interface.
type IPRedirect struct {
	// L3 protocol to be redirected
	L3Protocol L3Protocol `protobuf:"varint,1,opt,name=l3_protocol,json=l3Protocol,proto3,enum=vpp.punt.L3Protocol" json:"l3_protocol,omitempty"`
	// Receive interface name. Optional, only redirect traffic incoming from this interface
	RxInterface string `protobuf:"bytes,2,opt,name=rx_interface,json=rxInterface,proto3" json:"rx_interface,omitempty"`
	// Transmit interface name
	TxInterface string `protobuf:"bytes,3,opt,name=tx_interface,json=txInterface,proto3" json:"tx_interface,omitempty"`
	// Next hop IP where the traffic is redirected
	NextHop              string   `protobuf:"bytes,4,opt,name=next_hop,json=nextHop,proto3" json:"next_hop,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IPRedirect) Reset()         { *m = IPRedirect{} }
func (m *IPRedirect) String() string { return proto.CompactTextString(m) }
func (*IPRedirect) ProtoMessage()    {}
func (*IPRedirect) Descriptor() ([]byte, []int) {
	return fileDescriptor_55e6bedee92e220c, []int{0}
}

func (m *IPRedirect) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IPRedirect.Unmarshal(m, b)
}
func (m *IPRedirect) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IPRedirect.Marshal(b, m, deterministic)
}
func (m *IPRedirect) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IPRedirect.Merge(m, src)
}
func (m *IPRedirect) XXX_Size() int {
	return xxx_messageInfo_IPRedirect.Size(m)
}
func (m *IPRedirect) XXX_DiscardUnknown() {
	xxx_messageInfo_IPRedirect.DiscardUnknown(m)
}

var xxx_messageInfo_IPRedirect proto.InternalMessageInfo

func (m *IPRedirect) GetL3Protocol() L3Protocol {
	if m != nil {
		return m.L3Protocol
	}
	return L3Protocol_UNDEFINED_L3
}

func (m *IPRedirect) GetRxInterface() string {
	if m != nil {
		return m.RxInterface
	}
	return ""
}

func (m *IPRedirect) GetTxInterface() string {
	if m != nil {
		return m.TxInterface
	}
	return ""
}

func (m *IPRedirect) GetNextHop() string {
	if m != nil {
		return m.NextHop
	}
	return ""
}

// ToHost allows otherwise dropped packet which destination IP address matching
// some of the VPP interface IP addresses to be punted to the host.
// L3 and L4 protocols can be used for filtering */
type ToHost struct {
	// L3 destination protocol a packet has to match in order to be punted.
	L3Protocol L3Protocol `protobuf:"varint,2,opt,name=l3_protocol,json=l3Protocol,proto3,enum=vpp.punt.L3Protocol" json:"l3_protocol,omitempty"`
	// L4 destination protocol a packet has to match.
	// Currently VPP only supports UDP.
	L4Protocol L4Protocol `protobuf:"varint,3,opt,name=l4_protocol,json=l4Protocol,proto3,enum=vpp.punt.L4Protocol" json:"l4_protocol,omitempty"`
	// Destination port
	Port uint32 `protobuf:"varint,4,opt,name=port,proto3" json:"port,omitempty"`
	// SocketPath defines path to unix domain socket
	// used for punt packets to the host.
	// In dumps, it will actually contain the socket
	// defined in VPP config under punt section.
	SocketPath           string   `protobuf:"bytes,5,opt,name=socket_path,json=socketPath,proto3" json:"socket_path,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ToHost) Reset()         { *m = ToHost{} }
func (m *ToHost) String() string { return proto.CompactTextString(m) }
func (*ToHost) ProtoMessage()    {}
func (*ToHost) Descriptor() ([]byte, []int) {
	return fileDescriptor_55e6bedee92e220c, []int{1}
}

func (m *ToHost) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ToHost.Unmarshal(m, b)
}
func (m *ToHost) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ToHost.Marshal(b, m, deterministic)
}
func (m *ToHost) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ToHost.Merge(m, src)
}
func (m *ToHost) XXX_Size() int {
	return xxx_messageInfo_ToHost.Size(m)
}
func (m *ToHost) XXX_DiscardUnknown() {
	xxx_messageInfo_ToHost.DiscardUnknown(m)
}

var xxx_messageInfo_ToHost proto.InternalMessageInfo

func (m *ToHost) GetL3Protocol() L3Protocol {
	if m != nil {
		return m.L3Protocol
	}
	return L3Protocol_UNDEFINED_L3
}

func (m *ToHost) GetL4Protocol() L4Protocol {
	if m != nil {
		return m.L4Protocol
	}
	return L4Protocol_UNDEFINED_L4
}

func (m *ToHost) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *ToHost) GetSocketPath() string {
	if m != nil {
		return m.SocketPath
	}
	return ""
}

// Exception allows specifying punt exceptions used for punting packets.
// The type of exception is defined by reason name.
type Exception struct {
	// Name should contain reason name, e.g. `ipsec4-spi-0`.
	Reason string `protobuf:"bytes,1,opt,name=reason,proto3" json:"reason,omitempty"`
	// SocketPath defines path to unix domain socket
	// used for punt packets to the host.
	// In dumps, it will actually contain the socket
	// defined in VPP config under punt section.
	SocketPath           string   `protobuf:"bytes,2,opt,name=socket_path,json=socketPath,proto3" json:"socket_path,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Exception) Reset()         { *m = Exception{} }
func (m *Exception) String() string { return proto.CompactTextString(m) }
func (*Exception) ProtoMessage()    {}
func (*Exception) Descriptor() ([]byte, []int) {
	return fileDescriptor_55e6bedee92e220c, []int{2}
}

func (m *Exception) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Exception.Unmarshal(m, b)
}
func (m *Exception) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Exception.Marshal(b, m, deterministic)
}
func (m *Exception) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Exception.Merge(m, src)
}
func (m *Exception) XXX_Size() int {
	return xxx_messageInfo_Exception.Size(m)
}
func (m *Exception) XXX_DiscardUnknown() {
	xxx_messageInfo_Exception.DiscardUnknown(m)
}

var xxx_messageInfo_Exception proto.InternalMessageInfo

func (m *Exception) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

func (m *Exception) GetSocketPath() string {
	if m != nil {
		return m.SocketPath
	}
	return ""
}

// Reason represents punt reason used in exceptions.
// List of known exceptions can be retrieved in VPP CLI
// with following command:
//
// vpp# show punt reasons
//    [0] ipsec4-spi-0 from:[ipsec ]
//    [1] ipsec6-spi-0 from:[ipsec ]
//    [2] ipsec4-spi-o-udp-0 from:[ipsec ]
//    [3] ipsec4-no-such-tunnel from:[ipsec ]
//    [4] ipsec6-no-such-tunnel from:[ipsec ]
//    [5] VXLAN-GBP-no-such-v4-tunnel from:[vxlan-gbp ]
//    [6] VXLAN-GBP-no-such-v6-tunnel from:[vxlan-gbp ]
//
type Reason struct {
	// Name contains reason name.
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Reason) Reset()         { *m = Reason{} }
func (m *Reason) String() string { return proto.CompactTextString(m) }
func (*Reason) ProtoMessage()    {}
func (*Reason) Descriptor() ([]byte, []int) {
	return fileDescriptor_55e6bedee92e220c, []int{3}
}

func (m *Reason) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Reason.Unmarshal(m, b)
}
func (m *Reason) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Reason.Marshal(b, m, deterministic)
}
func (m *Reason) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Reason.Merge(m, src)
}
func (m *Reason) XXX_Size() int {
	return xxx_messageInfo_Reason.Size(m)
}
func (m *Reason) XXX_DiscardUnknown() {
	xxx_messageInfo_Reason.DiscardUnknown(m)
}

var xxx_messageInfo_Reason proto.InternalMessageInfo

func (m *Reason) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterEnum("vpp.punt.L3Protocol", L3Protocol_name, L3Protocol_value)
	proto.RegisterEnum("vpp.punt.L4Protocol", L4Protocol_name, L4Protocol_value)
	proto.RegisterType((*IPRedirect)(nil), "vpp.punt.IPRedirect")
	proto.RegisterType((*ToHost)(nil), "vpp.punt.ToHost")
	proto.RegisterType((*Exception)(nil), "vpp.punt.Exception")
	proto.RegisterType((*Reason)(nil), "vpp.punt.Reason")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/api/models/vpp/punt/punt.proto", fileDescriptor_55e6bedee92e220c)
}

var fileDescriptor_55e6bedee92e220c = []byte{
	// 385 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x4f, 0xab, 0x9b, 0x40,
	0x14, 0xc5, 0x63, 0x62, 0x4d, 0x72, 0x93, 0x16, 0x3b, 0x94, 0x62, 0xa1, 0xd0, 0x34, 0xab, 0x10,
	0xa8, 0x96, 0x6a, 0x4b, 0x69, 0x56, 0x6d, 0x4d, 0x89, 0x20, 0x41, 0x24, 0xd9, 0x74, 0x23, 0x13,
	0x33, 0x2f, 0xca, 0x33, 0xce, 0xa0, 0x13, 0xf1, 0x03, 0xbd, 0xdd, 0xfb, 0x92, 0x0f, 0x47, 0xf3,
	0x87, 0x64, 0x93, 0x8d, 0x9c, 0x7b, 0xe7, 0x77, 0x2e, 0xe7, 0x5e, 0x04, 0x23, 0x89, 0x77, 0x98,
	0x53, 0xa3, 0x60, 0xec, 0x0b, 0xde, 0x91, 0x94, 0x1b, 0x98, 0xc5, 0xc6, 0x9e, 0x6e, 0x49, 0x92,
	0x57, 0x4d, 0x83, 0x1d, 0x52, 0x2e, 0x3e, 0x3a, 0xcb, 0x28, 0xa7, 0xa8, 0x57, 0x30, 0xa6, 0x57,
	0xf5, 0xf8, 0x49, 0x02, 0x70, 0x3c, 0x9f, 0x6c, 0xe3, 0x8c, 0x84, 0x1c, 0x7d, 0x87, 0x41, 0x62,
	0x06, 0x02, 0x0a, 0x69, 0xa2, 0x49, 0x23, 0x69, 0xf2, 0xe6, 0xdb, 0x3b, 0xfd, 0x88, 0xeb, 0xae,
	0xe9, 0x35, 0x6f, 0x3e, 0x24, 0x27, 0x8d, 0x3e, 0xc3, 0x30, 0x2b, 0x83, 0x38, 0xe5, 0x24, 0x7b,
	0xc0, 0x21, 0xd1, 0xda, 0x23, 0x69, 0xd2, 0xf7, 0x07, 0x59, 0xe9, 0x1c, 0x5b, 0x15, 0xc2, 0x2f,
	0x91, 0x4e, 0x8d, 0xf0, 0x0b, 0xe4, 0x03, 0xf4, 0x52, 0x52, 0xf2, 0x20, 0xa2, 0x4c, 0x93, 0xc5,
	0x73, 0xb7, 0xaa, 0x17, 0x94, 0x8d, 0x9f, 0x25, 0x50, 0x56, 0x74, 0x41, 0xf3, 0x9b, 0x88, 0xed,
	0x3b, 0x23, 0x56, 0x36, 0xeb, 0x6c, 0xeb, 0xdc, 0xd8, 0xac, 0x0b, 0xdb, 0x49, 0x23, 0x04, 0x32,
	0xa3, 0x19, 0x17, 0x79, 0x5e, 0xfb, 0x42, 0xa3, 0x4f, 0x30, 0xc8, 0x69, 0xf8, 0x48, 0x78, 0xc0,
	0x30, 0x8f, 0xb4, 0x57, 0x22, 0x2a, 0xd4, 0x2d, 0x0f, 0xf3, 0x68, 0x6c, 0x43, 0x7f, 0x5e, 0x86,
	0x84, 0xf1, 0x98, 0xa6, 0xe8, 0x3d, 0x28, 0x19, 0xc1, 0x39, 0x4d, 0xc5, 0x35, 0xfb, 0x7e, 0x53,
	0x5d, 0x4f, 0x69, 0xdf, 0x4c, 0xf9, 0x08, 0x8a, 0x5f, 0xa3, 0x08, 0xe4, 0x14, 0xef, 0x49, 0x33,
	0x40, 0xe8, 0xe9, 0x0c, 0xe0, 0xbc, 0x29, 0x52, 0x61, 0xb8, 0x5e, 0xda, 0xf3, 0x7f, 0xce, 0x72,
	0x6e, 0x07, 0xae, 0xa9, 0xb6, 0x50, 0x0f, 0x64, 0xc7, 0x2b, 0x2c, 0x55, 0x6e, 0xd4, 0x0f, 0x55,
	0x41, 0x5d, 0xe8, 0xfc, 0x76, 0x5d, 0x15, 0xa6, 0x5f, 0x01, 0xce, 0xfb, 0x5e, 0x99, 0x2d, 0xb5,
	0x55, 0x81, 0xab, 0xbf, 0x5e, 0xed, 0x58, 0xdb, 0x9e, 0xfa, 0xf6, 0xcf, 0xaf, 0xff, 0x3f, 0x77,
	0x31, 0x8f, 0x0e, 0x1b, 0x3d, 0xa4, 0xfb, 0xbb, 0xfe, 0xb7, 0x59, 0xc1, 0x58, 0x50, 0x89, 0x8d,
	0x22, 0xae, 0x6e, 0xbe, 0x04, 0x00, 0x00, 0xff, 0xff, 0x29, 0x31, 0x0a, 0x97, 0xa7, 0x02, 0x00,
	0x00,
}
