// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.1
// source: ligato/vpp/punt/punt.proto

package vpp_punt

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// L3Protocol defines Layer 3 protocols.
type L3Protocol int32

const (
	L3Protocol_UNDEFINED_L3 L3Protocol = 0
	L3Protocol_IPV4         L3Protocol = 4
	L3Protocol_IPV6         L3Protocol = 6
	L3Protocol_ALL          L3Protocol = 10
)

// Enum value maps for L3Protocol.
var (
	L3Protocol_name = map[int32]string{
		0:  "UNDEFINED_L3",
		4:  "IPV4",
		6:  "IPV6",
		10: "ALL",
	}
	L3Protocol_value = map[string]int32{
		"UNDEFINED_L3": 0,
		"IPV4":         4,
		"IPV6":         6,
		"ALL":          10,
	}
)

func (x L3Protocol) Enum() *L3Protocol {
	p := new(L3Protocol)
	*p = x
	return p
}

func (x L3Protocol) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (L3Protocol) Descriptor() protoreflect.EnumDescriptor {
	return file_ligato_vpp_punt_punt_proto_enumTypes[0].Descriptor()
}

func (L3Protocol) Type() protoreflect.EnumType {
	return &file_ligato_vpp_punt_punt_proto_enumTypes[0]
}

func (x L3Protocol) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use L3Protocol.Descriptor instead.
func (L3Protocol) EnumDescriptor() ([]byte, []int) {
	return file_ligato_vpp_punt_punt_proto_rawDescGZIP(), []int{0}
}

// L4Protocol defines Layer 4 protocols.
type L4Protocol int32

const (
	L4Protocol_UNDEFINED_L4 L4Protocol = 0
	L4Protocol_TCP          L4Protocol = 6
	L4Protocol_UDP          L4Protocol = 17
)

// Enum value maps for L4Protocol.
var (
	L4Protocol_name = map[int32]string{
		0:  "UNDEFINED_L4",
		6:  "TCP",
		17: "UDP",
	}
	L4Protocol_value = map[string]int32{
		"UNDEFINED_L4": 0,
		"TCP":          6,
		"UDP":          17,
	}
)

func (x L4Protocol) Enum() *L4Protocol {
	p := new(L4Protocol)
	*p = x
	return p
}

func (x L4Protocol) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (L4Protocol) Descriptor() protoreflect.EnumDescriptor {
	return file_ligato_vpp_punt_punt_proto_enumTypes[1].Descriptor()
}

func (L4Protocol) Type() protoreflect.EnumType {
	return &file_ligato_vpp_punt_punt_proto_enumTypes[1]
}

func (x L4Protocol) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use L4Protocol.Descriptor instead.
func (L4Protocol) EnumDescriptor() ([]byte, []int) {
	return file_ligato_vpp_punt_punt_proto_rawDescGZIP(), []int{1}
}

// IPRedirect allows otherwise dropped packet which destination IP address
// matching some of the VPP addresses to redirect to the defined next hop address
// via the TX interface.
type IPRedirect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// L3 protocol to be redirected
	L3Protocol L3Protocol `protobuf:"varint,1,opt,name=l3_protocol,json=l3Protocol,proto3,enum=ligato.vpp.punt.L3Protocol" json:"l3_protocol,omitempty"`
	// Receive interface name. Optional, only redirect traffic incoming from this interface
	RxInterface string `protobuf:"bytes,2,opt,name=rx_interface,json=rxInterface,proto3" json:"rx_interface,omitempty"`
	// Transmit interface name
	TxInterface string `protobuf:"bytes,3,opt,name=tx_interface,json=txInterface,proto3" json:"tx_interface,omitempty"`
	// Next hop IP where the traffic is redirected
	NextHop string `protobuf:"bytes,4,opt,name=next_hop,json=nextHop,proto3" json:"next_hop,omitempty"`
}

func (x *IPRedirect) Reset() {
	*x = IPRedirect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_punt_punt_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IPRedirect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPRedirect) ProtoMessage() {}

func (x *IPRedirect) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_punt_punt_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPRedirect.ProtoReflect.Descriptor instead.
func (*IPRedirect) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_punt_punt_proto_rawDescGZIP(), []int{0}
}

func (x *IPRedirect) GetL3Protocol() L3Protocol {
	if x != nil {
		return x.L3Protocol
	}
	return L3Protocol_UNDEFINED_L3
}

func (x *IPRedirect) GetRxInterface() string {
	if x != nil {
		return x.RxInterface
	}
	return ""
}

func (x *IPRedirect) GetTxInterface() string {
	if x != nil {
		return x.TxInterface
	}
	return ""
}

func (x *IPRedirect) GetNextHop() string {
	if x != nil {
		return x.NextHop
	}
	return ""
}

// ToHost allows otherwise dropped packet which destination IP address matching
// some of the VPP interface IP addresses to be punted to the host.
// L3 and L4 protocols can be used for filtering */
type ToHost struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// L3 destination protocol a packet has to match in order to be punted.
	L3Protocol L3Protocol `protobuf:"varint,2,opt,name=l3_protocol,json=l3Protocol,proto3,enum=ligato.vpp.punt.L3Protocol" json:"l3_protocol,omitempty"`
	// L4 destination protocol a packet has to match.
	// Currently VPP only supports UDP.
	L4Protocol L4Protocol `protobuf:"varint,3,opt,name=l4_protocol,json=l4Protocol,proto3,enum=ligato.vpp.punt.L4Protocol" json:"l4_protocol,omitempty"`
	// Destination port
	Port uint32 `protobuf:"varint,4,opt,name=port,proto3" json:"port,omitempty"`
	// SocketPath defines path to unix domain socket
	// used for punt packets to the host.
	// In dumps, it will actually contain the socket
	// defined in VPP config under punt section.
	SocketPath string `protobuf:"bytes,5,opt,name=socket_path,json=socketPath,proto3" json:"socket_path,omitempty"`
}

func (x *ToHost) Reset() {
	*x = ToHost{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_punt_punt_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ToHost) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToHost) ProtoMessage() {}

func (x *ToHost) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_punt_punt_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToHost.ProtoReflect.Descriptor instead.
func (*ToHost) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_punt_punt_proto_rawDescGZIP(), []int{1}
}

func (x *ToHost) GetL3Protocol() L3Protocol {
	if x != nil {
		return x.L3Protocol
	}
	return L3Protocol_UNDEFINED_L3
}

func (x *ToHost) GetL4Protocol() L4Protocol {
	if x != nil {
		return x.L4Protocol
	}
	return L4Protocol_UNDEFINED_L4
}

func (x *ToHost) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ToHost) GetSocketPath() string {
	if x != nil {
		return x.SocketPath
	}
	return ""
}

// Exception allows specifying punt exceptions used for punting packets.
// The type of exception is defined by reason name.
type Exception struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name should contain reason name, e.g. `ipsec4-spi-0`.
	Reason string `protobuf:"bytes,1,opt,name=reason,proto3" json:"reason,omitempty"`
	// SocketPath defines path to unix domain socket
	// used for punt packets to the host.
	// In dumps, it will actually contain the socket
	// defined in VPP config under punt section.
	SocketPath string `protobuf:"bytes,2,opt,name=socket_path,json=socketPath,proto3" json:"socket_path,omitempty"`
}

func (x *Exception) Reset() {
	*x = Exception{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_punt_punt_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Exception) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Exception) ProtoMessage() {}

func (x *Exception) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_punt_punt_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Exception.ProtoReflect.Descriptor instead.
func (*Exception) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_punt_punt_proto_rawDescGZIP(), []int{2}
}

func (x *Exception) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *Exception) GetSocketPath() string {
	if x != nil {
		return x.SocketPath
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
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name contains reason name.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Reason) Reset() {
	*x = Reason{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_punt_punt_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Reason) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reason) ProtoMessage() {}

func (x *Reason) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_punt_punt_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Reason.ProtoReflect.Descriptor instead.
func (*Reason) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_punt_punt_proto_rawDescGZIP(), []int{3}
}

func (x *Reason) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

var File_ligato_vpp_punt_punt_proto protoreflect.FileDescriptor

var file_ligato_vpp_punt_punt_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x70, 0x75, 0x6e,
	0x74, 0x2f, 0x70, 0x75, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x6c, 0x69,
	0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x70, 0x75, 0x6e, 0x74, 0x22, 0xab, 0x01,
	0x0a, 0x0a, 0x49, 0x50, 0x52, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x12, 0x3c, 0x0a, 0x0b,
	0x6c, 0x33, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x1b, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x70,
	0x75, 0x6e, 0x74, 0x2e, 0x4c, 0x33, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x0a,
	0x6c, 0x33, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x21, 0x0a, 0x0c, 0x72, 0x78,
	0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x72, 0x78, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x21, 0x0a,
	0x0c, 0x74, 0x78, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x74, 0x78, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65,
	0x12, 0x19, 0x0a, 0x08, 0x6e, 0x65, 0x78, 0x74, 0x5f, 0x68, 0x6f, 0x70, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6e, 0x65, 0x78, 0x74, 0x48, 0x6f, 0x70, 0x22, 0xb9, 0x01, 0x0a, 0x06,
	0x54, 0x6f, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x3c, 0x0a, 0x0b, 0x6c, 0x33, 0x5f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x6c, 0x69,
	0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x70, 0x75, 0x6e, 0x74, 0x2e, 0x4c, 0x33,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x0a, 0x6c, 0x33, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x3c, 0x0a, 0x0b, 0x6c, 0x34, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x6c, 0x69, 0x67, 0x61,
	0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x70, 0x75, 0x6e, 0x74, 0x2e, 0x4c, 0x34, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x0a, 0x6c, 0x34, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74,
	0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x6f, 0x63,
	0x6b, 0x65, 0x74, 0x50, 0x61, 0x74, 0x68, 0x22, 0x44, 0x0a, 0x09, 0x45, 0x78, 0x63, 0x65, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x1f, 0x0a, 0x0b,
	0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x50, 0x61, 0x74, 0x68, 0x22, 0x1c, 0x0a,
	0x06, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x2a, 0x3b, 0x0a, 0x0a, 0x4c,
	0x33, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x10, 0x0a, 0x0c, 0x55, 0x4e, 0x44,
	0x45, 0x46, 0x49, 0x4e, 0x45, 0x44, 0x5f, 0x4c, 0x33, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x49,
	0x50, 0x56, 0x34, 0x10, 0x04, 0x12, 0x08, 0x0a, 0x04, 0x49, 0x50, 0x56, 0x36, 0x10, 0x06, 0x12,
	0x07, 0x0a, 0x03, 0x41, 0x4c, 0x4c, 0x10, 0x0a, 0x2a, 0x30, 0x0a, 0x0a, 0x4c, 0x34, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x10, 0x0a, 0x0c, 0x55, 0x4e, 0x44, 0x45, 0x46, 0x49,
	0x4e, 0x45, 0x44, 0x5f, 0x4c, 0x34, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x54, 0x43, 0x50, 0x10,
	0x06, 0x12, 0x07, 0x0a, 0x03, 0x55, 0x44, 0x50, 0x10, 0x11, 0x42, 0x3a, 0x5a, 0x38, 0x67, 0x6f,
	0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2d, 0x61,
	0x67, 0x65, 0x6e, 0x74, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x69,
	0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x70, 0x75, 0x6e, 0x74, 0x3b, 0x76, 0x70,
	0x70, 0x5f, 0x70, 0x75, 0x6e, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ligato_vpp_punt_punt_proto_rawDescOnce sync.Once
	file_ligato_vpp_punt_punt_proto_rawDescData = file_ligato_vpp_punt_punt_proto_rawDesc
)

func file_ligato_vpp_punt_punt_proto_rawDescGZIP() []byte {
	file_ligato_vpp_punt_punt_proto_rawDescOnce.Do(func() {
		file_ligato_vpp_punt_punt_proto_rawDescData = protoimpl.X.CompressGZIP(file_ligato_vpp_punt_punt_proto_rawDescData)
	})
	return file_ligato_vpp_punt_punt_proto_rawDescData
}

var file_ligato_vpp_punt_punt_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_ligato_vpp_punt_punt_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_ligato_vpp_punt_punt_proto_goTypes = []interface{}{
	(L3Protocol)(0),    // 0: ligato.vpp.punt.L3Protocol
	(L4Protocol)(0),    // 1: ligato.vpp.punt.L4Protocol
	(*IPRedirect)(nil), // 2: ligato.vpp.punt.IPRedirect
	(*ToHost)(nil),     // 3: ligato.vpp.punt.ToHost
	(*Exception)(nil),  // 4: ligato.vpp.punt.Exception
	(*Reason)(nil),     // 5: ligato.vpp.punt.Reason
}
var file_ligato_vpp_punt_punt_proto_depIdxs = []int32{
	0, // 0: ligato.vpp.punt.IPRedirect.l3_protocol:type_name -> ligato.vpp.punt.L3Protocol
	0, // 1: ligato.vpp.punt.ToHost.l3_protocol:type_name -> ligato.vpp.punt.L3Protocol
	1, // 2: ligato.vpp.punt.ToHost.l4_protocol:type_name -> ligato.vpp.punt.L4Protocol
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_ligato_vpp_punt_punt_proto_init() }
func file_ligato_vpp_punt_punt_proto_init() {
	if File_ligato_vpp_punt_punt_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ligato_vpp_punt_punt_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IPRedirect); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ligato_vpp_punt_punt_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ToHost); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ligato_vpp_punt_punt_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Exception); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ligato_vpp_punt_punt_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Reason); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ligato_vpp_punt_punt_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ligato_vpp_punt_punt_proto_goTypes,
		DependencyIndexes: file_ligato_vpp_punt_punt_proto_depIdxs,
		EnumInfos:         file_ligato_vpp_punt_punt_proto_enumTypes,
		MessageInfos:      file_ligato_vpp_punt_punt_proto_msgTypes,
	}.Build()
	File_ligato_vpp_punt_punt_proto = out.File
	file_ligato_vpp_punt_punt_proto_rawDesc = nil
	file_ligato_vpp_punt_punt_proto_goTypes = nil
	file_ligato_vpp_punt_punt_proto_depIdxs = nil
}
