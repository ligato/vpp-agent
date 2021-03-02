// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.4
// source: ligato/annotations.proto

package ligato

import (
	proto "github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
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

type LigatoOptions_Type int32

const (
	LigatoOptions_IP   LigatoOptions_Type = 0
	LigatoOptions_IPV4 LigatoOptions_Type = 1
	LigatoOptions_IPV6 LigatoOptions_Type = 2
)

// Enum value maps for LigatoOptions_Type.
var (
	LigatoOptions_Type_name = map[int32]string{
		0: "IP",
		1: "IPV4",
		2: "IPV6",
	}
	LigatoOptions_Type_value = map[string]int32{
		"IP":   0,
		"IPV4": 1,
		"IPV6": 2,
	}
)

func (x LigatoOptions_Type) Enum() *LigatoOptions_Type {
	p := new(LigatoOptions_Type)
	*p = x
	return p
}

func (x LigatoOptions_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LigatoOptions_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_ligato_annotations_proto_enumTypes[0].Descriptor()
}

func (LigatoOptions_Type) Type() protoreflect.EnumType {
	return &file_ligato_annotations_proto_enumTypes[0]
}

func (x LigatoOptions_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LigatoOptions_Type.Descriptor instead.
func (LigatoOptions_Type) EnumDescriptor() ([]byte, []int) {
	return file_ligato_annotations_proto_rawDescGZIP(), []int{0, 0}
}

type LigatoOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type LigatoOptions_Type `protobuf:"varint,1,opt,name=type,proto3,enum=ligato.LigatoOptions_Type" json:"type,omitempty"`
}

func (x *LigatoOptions) Reset() {
	*x = LigatoOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_annotations_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LigatoOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LigatoOptions) ProtoMessage() {}

func (x *LigatoOptions) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_annotations_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LigatoOptions.ProtoReflect.Descriptor instead.
func (*LigatoOptions) Descriptor() ([]byte, []int) {
	return file_ligato_annotations_proto_rawDescGZIP(), []int{0}
}

func (x *LigatoOptions) GetType() LigatoOptions_Type {
	if x != nil {
		return x.Type
	}
	return LigatoOptions_IP
}

var file_ligato_annotations_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptor.FieldOptions)(nil),
		ExtensionType: (*LigatoOptions)(nil),
		Field:         2000,
		Name:          "ligato.ligato_options",
		Tag:           "bytes,2000,opt,name=ligato_options",
		Filename:      "ligato/annotations.proto",
	},
}

// Extension fields to descriptor.FieldOptions.
var (
	// NOTE: used option field index(2000) is in extension index range of descriptor.proto, but  is not registered
	// in protobuf global extension registry (https://github.com/protocolbuffers/protobuf/blob/master/docs/options.md)
	//
	// optional ligato.LigatoOptions ligato_options = 2000;
	E_LigatoOptions = &file_ligato_annotations_proto_extTypes[0]
)

var File_ligato_annotations_proto protoreflect.FileDescriptor

var file_ligato_annotations_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6c, 0x69, 0x67, 0x61,
	0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x63, 0x0a, 0x0d, 0x4c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x2e, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x4c, 0x69, 0x67,
	0x61, 0x74, 0x6f, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x22, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x06, 0x0a,
	0x02, 0x49, 0x50, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x49, 0x50, 0x56, 0x34, 0x10, 0x01, 0x12,
	0x08, 0x0a, 0x04, 0x49, 0x50, 0x56, 0x36, 0x10, 0x02, 0x3a, 0x5c, 0x0a, 0x0e, 0x6c, 0x69, 0x67,
	0x61, 0x74, 0x6f, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1d, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd0, 0x0f, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x15, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x4c, 0x69, 0x67, 0x61, 0x74,
	0x6f, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x0d, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x42, 0x28, 0x5a, 0x26, 0x67, 0x6f, 0x2e, 0x6c, 0x69,
	0x67, 0x61, 0x74, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2d, 0x61, 0x67, 0x65, 0x6e,
	0x74, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x69, 0x67, 0x61, 0x74,
	0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ligato_annotations_proto_rawDescOnce sync.Once
	file_ligato_annotations_proto_rawDescData = file_ligato_annotations_proto_rawDesc
)

func file_ligato_annotations_proto_rawDescGZIP() []byte {
	file_ligato_annotations_proto_rawDescOnce.Do(func() {
		file_ligato_annotations_proto_rawDescData = protoimpl.X.CompressGZIP(file_ligato_annotations_proto_rawDescData)
	})
	return file_ligato_annotations_proto_rawDescData
}

var file_ligato_annotations_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_ligato_annotations_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_ligato_annotations_proto_goTypes = []interface{}{
	(LigatoOptions_Type)(0),         // 0: ligato.LigatoOptions.Type
	(*LigatoOptions)(nil),           // 1: ligato.LigatoOptions
	(*descriptor.FieldOptions)(nil), // 2: google.protobuf.FieldOptions
}
var file_ligato_annotations_proto_depIdxs = []int32{
	0, // 0: ligato.LigatoOptions.type:type_name -> ligato.LigatoOptions.Type
	2, // 1: ligato.ligato_options:extendee -> google.protobuf.FieldOptions
	1, // 2: ligato.ligato_options:type_name -> ligato.LigatoOptions
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	2, // [2:3] is the sub-list for extension type_name
	1, // [1:2] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_ligato_annotations_proto_init() }
func file_ligato_annotations_proto_init() {
	if File_ligato_annotations_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ligato_annotations_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LigatoOptions); i {
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
			RawDescriptor: file_ligato_annotations_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 1,
			NumServices:   0,
		},
		GoTypes:           file_ligato_annotations_proto_goTypes,
		DependencyIndexes: file_ligato_annotations_proto_depIdxs,
		EnumInfos:         file_ligato_annotations_proto_enumTypes,
		MessageInfos:      file_ligato_annotations_proto_msgTypes,
		ExtensionInfos:    file_ligato_annotations_proto_extTypes,
	}.Build()
	File_ligato_annotations_proto = out.File
	file_ligato_annotations_proto_rawDesc = nil
	file_ligato_annotations_proto_goTypes = nil
	file_ligato_annotations_proto_depIdxs = nil
}
