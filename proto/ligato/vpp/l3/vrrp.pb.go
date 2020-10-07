// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.4
// source: ligato/vpp/l3/vrrp.proto

package vpp_l3

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

type VRRPEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Interface string `protobuf:"bytes,1,opt,name=interface,proto3" json:"interface,omitempty"`
	VrId      uint32 `protobuf:"varint,2,opt,name=vr_id,json=vrId,proto3" json:"vr_id,omitempty"`
	Priority  uint32 `protobuf:"varint,3,opt,name=priority,proto3" json:"priority,omitempty"`
	// VR advertisement interval.
	Interval uint32 `protobuf:"varint,4,opt,name=interval,proto3" json:"interval,omitempty"`
	// Controls whether a (starting or restarting)
	// higher-priority Backup router preempts a lower-priority Master router.
	Preempt bool `protobuf:"varint,5,opt,name=preempt,proto3" json:"preempt,omitempty"`
	// Controls whether a virtual router in Master state will accept packets
	// addressed to the address owner's IPvX address as its own if it is not the IPvX address owner.
	Accept bool `protobuf:"varint,6,opt,name=accept,proto3" json:"accept,omitempty"`
	// Unicast mode may be used to take
	// advantage of newer token ring adapter implementations that support
	// non-promiscuous reception for multiple unicast MAC addresses and to
	// avoid both the multicast traffic and usage conflicts associated with
	// the use of token ring functional addresses.
	Unicast bool     `protobuf:"varint,7,opt,name=unicast,proto3" json:"unicast,omitempty"`
	Ipv6    bool     `protobuf:"varint,8,opt,name=ipv6,proto3" json:"ipv6,omitempty"`
	Addrs   []string `protobuf:"bytes,9,rep,name=addrs,proto3" json:"addrs,omitempty"`
	Enabled bool     `protobuf:"varint,10,opt,name=enabled,proto3" json:"enabled,omitempty"`
}

func (x *VRRPEntry) Reset() {
	*x = VRRPEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_l3_vrrp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VRRPEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VRRPEntry) ProtoMessage() {}

func (x *VRRPEntry) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_l3_vrrp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VRRPEntry.ProtoReflect.Descriptor instead.
func (*VRRPEntry) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_l3_vrrp_proto_rawDescGZIP(), []int{0}
}

func (x *VRRPEntry) GetInterface() string {
	if x != nil {
		return x.Interface
	}
	return ""
}

func (x *VRRPEntry) GetVrId() uint32 {
	if x != nil {
		return x.VrId
	}
	return 0
}

func (x *VRRPEntry) GetPriority() uint32 {
	if x != nil {
		return x.Priority
	}
	return 0
}

func (x *VRRPEntry) GetInterval() uint32 {
	if x != nil {
		return x.Interval
	}
	return 0
}

func (x *VRRPEntry) GetPreempt() bool {
	if x != nil {
		return x.Preempt
	}
	return false
}

func (x *VRRPEntry) GetAccept() bool {
	if x != nil {
		return x.Accept
	}
	return false
}

func (x *VRRPEntry) GetUnicast() bool {
	if x != nil {
		return x.Unicast
	}
	return false
}

func (x *VRRPEntry) GetIpv6() bool {
	if x != nil {
		return x.Ipv6
	}
	return false
}

func (x *VRRPEntry) GetAddrs() []string {
	if x != nil {
		return x.Addrs
	}
	return nil
}

func (x *VRRPEntry) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

var File_ligato_vpp_l3_vrrp_proto protoreflect.FileDescriptor

var file_ligato_vpp_l3_vrrp_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x6c, 0x33, 0x2f,
	0x76, 0x72, 0x72, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x6c, 0x69, 0x67, 0x61,
	0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x6c, 0x33, 0x22, 0x86, 0x02, 0x0a, 0x09, 0x56, 0x52,
	0x52, 0x50, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x66, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x13, 0x0a, 0x05, 0x76, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x76, 0x72, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72,
	0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x70, 0x72,
	0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76,
	0x61, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76,
	0x61, 0x6c, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x65, 0x65, 0x6d, 0x70, 0x74, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x07, 0x70, 0x72, 0x65, 0x65, 0x6d, 0x70, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x61, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x75, 0x6e, 0x69, 0x63, 0x61, 0x73, 0x74, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x75, 0x6e, 0x69, 0x63, 0x61, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x69, 0x70, 0x76, 0x36, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x69, 0x70,
	0x76, 0x36, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x64, 0x64, 0x72, 0x73, 0x18, 0x09, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x05, 0x61, 0x64, 0x64, 0x72, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62,
	0x6c, 0x65, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c,
	0x65, 0x64, 0x42, 0x36, 0x5a, 0x34, 0x67, 0x6f, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e,
	0x69, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2d, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x76, 0x33, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70,
	0x2f, 0x6c, 0x33, 0x3b, 0x76, 0x70, 0x70, 0x5f, 0x6c, 0x33, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_ligato_vpp_l3_vrrp_proto_rawDescOnce sync.Once
	file_ligato_vpp_l3_vrrp_proto_rawDescData = file_ligato_vpp_l3_vrrp_proto_rawDesc
)

func file_ligato_vpp_l3_vrrp_proto_rawDescGZIP() []byte {
	file_ligato_vpp_l3_vrrp_proto_rawDescOnce.Do(func() {
		file_ligato_vpp_l3_vrrp_proto_rawDescData = protoimpl.X.CompressGZIP(file_ligato_vpp_l3_vrrp_proto_rawDescData)
	})
	return file_ligato_vpp_l3_vrrp_proto_rawDescData
}

var file_ligato_vpp_l3_vrrp_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_ligato_vpp_l3_vrrp_proto_goTypes = []interface{}{
	(*VRRPEntry)(nil), // 0: ligato.vpp.l3.VRRPEntry
}
var file_ligato_vpp_l3_vrrp_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_ligato_vpp_l3_vrrp_proto_init() }
func file_ligato_vpp_l3_vrrp_proto_init() {
	if File_ligato_vpp_l3_vrrp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ligato_vpp_l3_vrrp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VRRPEntry); i {
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
			RawDescriptor: file_ligato_vpp_l3_vrrp_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ligato_vpp_l3_vrrp_proto_goTypes,
		DependencyIndexes: file_ligato_vpp_l3_vrrp_proto_depIdxs,
		MessageInfos:      file_ligato_vpp_l3_vrrp_proto_msgTypes,
	}.Build()
	File_ligato_vpp_l3_vrrp_proto = out.File
	file_ligato_vpp_l3_vrrp_proto_rawDesc = nil
	file_ligato_vpp_l3_vrrp_proto_goTypes = nil
	file_ligato_vpp_l3_vrrp_proto_depIdxs = nil
}
