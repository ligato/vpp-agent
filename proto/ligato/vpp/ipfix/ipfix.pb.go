// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.1
// source: ligato/vpp/ipfix/ipfix.proto

package vpp_ipfix

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

// IPFIX defines the IP Flow Information eXport (IPFIX) configuration.
type IPFIX struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Collector        *IPFIX_Collector `protobuf:"bytes,1,opt,name=collector,proto3" json:"collector,omitempty"`
	SourceAddress    string           `protobuf:"bytes,2,opt,name=source_address,json=sourceAddress,proto3" json:"source_address,omitempty"`
	VrfId            uint32           `protobuf:"varint,3,opt,name=vrf_id,json=vrfId,proto3" json:"vrf_id,omitempty"`
	PathMtu          uint32           `protobuf:"varint,4,opt,name=path_mtu,json=pathMtu,proto3" json:"path_mtu,omitempty"`
	TemplateInterval uint32           `protobuf:"varint,5,opt,name=template_interval,json=templateInterval,proto3" json:"template_interval,omitempty"`
}

func (x *IPFIX) Reset() {
	*x = IPFIX{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_ipfix_ipfix_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IPFIX) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPFIX) ProtoMessage() {}

func (x *IPFIX) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_ipfix_ipfix_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPFIX.ProtoReflect.Descriptor instead.
func (*IPFIX) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_ipfix_ipfix_proto_rawDescGZIP(), []int{0}
}

func (x *IPFIX) GetCollector() *IPFIX_Collector {
	if x != nil {
		return x.Collector
	}
	return nil
}

func (x *IPFIX) GetSourceAddress() string {
	if x != nil {
		return x.SourceAddress
	}
	return ""
}

func (x *IPFIX) GetVrfId() uint32 {
	if x != nil {
		return x.VrfId
	}
	return 0
}

func (x *IPFIX) GetPathMtu() uint32 {
	if x != nil {
		return x.PathMtu
	}
	return 0
}

func (x *IPFIX) GetTemplateInterval() uint32 {
	if x != nil {
		return x.TemplateInterval
	}
	return 0
}

type IPFIX_Collector struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port    uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
}

func (x *IPFIX_Collector) Reset() {
	*x = IPFIX_Collector{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_ipfix_ipfix_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IPFIX_Collector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPFIX_Collector) ProtoMessage() {}

func (x *IPFIX_Collector) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_ipfix_ipfix_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPFIX_Collector.ProtoReflect.Descriptor instead.
func (*IPFIX_Collector) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_ipfix_ipfix_proto_rawDescGZIP(), []int{0, 0}
}

func (x *IPFIX_Collector) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *IPFIX_Collector) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

var File_ligato_vpp_ipfix_ipfix_proto protoreflect.FileDescriptor

var file_ligato_vpp_ipfix_ipfix_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x69, 0x70, 0x66,
	0x69, 0x78, 0x2f, 0x69, 0x70, 0x66, 0x69, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10,
	0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x69, 0x70, 0x66, 0x69, 0x78,
	0x22, 0x89, 0x02, 0x0a, 0x05, 0x49, 0x50, 0x46, 0x49, 0x58, 0x12, 0x3f, 0x0a, 0x09, 0x63, 0x6f,
	0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e,
	0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x69, 0x70, 0x66, 0x69, 0x78,
	0x2e, 0x49, 0x50, 0x46, 0x49, 0x58, 0x2e, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x52, 0x09, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x25, 0x0a, 0x0e, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x12, 0x15, 0x0a, 0x06, 0x76, 0x72, 0x66, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x05, 0x76, 0x72, 0x66, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x61, 0x74,
	0x68, 0x5f, 0x6d, 0x74, 0x75, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x70, 0x61, 0x74,
	0x68, 0x4d, 0x74, 0x75, 0x12, 0x2b, 0x0a, 0x11, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65,
	0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x10, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61,
	0x6c, 0x1a, 0x39, 0x0a, 0x09, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x18,
	0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x42, 0x3c, 0x5a, 0x3a,
	0x67, 0x6f, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x76, 0x70, 0x70,
	0x2d, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x69, 0x70, 0x66, 0x69, 0x78,
	0x3b, 0x76, 0x70, 0x70, 0x5f, 0x69, 0x70, 0x66, 0x69, 0x78, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_ligato_vpp_ipfix_ipfix_proto_rawDescOnce sync.Once
	file_ligato_vpp_ipfix_ipfix_proto_rawDescData = file_ligato_vpp_ipfix_ipfix_proto_rawDesc
)

func file_ligato_vpp_ipfix_ipfix_proto_rawDescGZIP() []byte {
	file_ligato_vpp_ipfix_ipfix_proto_rawDescOnce.Do(func() {
		file_ligato_vpp_ipfix_ipfix_proto_rawDescData = protoimpl.X.CompressGZIP(file_ligato_vpp_ipfix_ipfix_proto_rawDescData)
	})
	return file_ligato_vpp_ipfix_ipfix_proto_rawDescData
}

var file_ligato_vpp_ipfix_ipfix_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_ligato_vpp_ipfix_ipfix_proto_goTypes = []interface{}{
	(*IPFIX)(nil),           // 0: ligato.vpp.ipfix.IPFIX
	(*IPFIX_Collector)(nil), // 1: ligato.vpp.ipfix.IPFIX.Collector
}
var file_ligato_vpp_ipfix_ipfix_proto_depIdxs = []int32{
	1, // 0: ligato.vpp.ipfix.IPFIX.collector:type_name -> ligato.vpp.ipfix.IPFIX.Collector
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_ligato_vpp_ipfix_ipfix_proto_init() }
func file_ligato_vpp_ipfix_ipfix_proto_init() {
	if File_ligato_vpp_ipfix_ipfix_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ligato_vpp_ipfix_ipfix_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IPFIX); i {
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
		file_ligato_vpp_ipfix_ipfix_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IPFIX_Collector); i {
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
			RawDescriptor: file_ligato_vpp_ipfix_ipfix_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ligato_vpp_ipfix_ipfix_proto_goTypes,
		DependencyIndexes: file_ligato_vpp_ipfix_ipfix_proto_depIdxs,
		MessageInfos:      file_ligato_vpp_ipfix_ipfix_proto_msgTypes,
	}.Build()
	File_ligato_vpp_ipfix_ipfix_proto = out.File
	file_ligato_vpp_ipfix_ipfix_proto_rawDesc = nil
	file_ligato_vpp_ipfix_ipfix_proto_goTypes = nil
	file_ligato_vpp_ipfix_ipfix_proto_depIdxs = nil
}
