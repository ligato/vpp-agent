// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: ligato/vpp/l2/bridge_domain.proto

package vpp_l2

import (
	_ "go.ligato.io/vpp-agent/v3/proto/ligato"
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

type BridgeDomain struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name                string                              `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`                                                              // bridge domain name (can be any string)
	Flood               bool                                `protobuf:"varint,2,opt,name=flood,proto3" json:"flood,omitempty"`                                                           // enable/disable broadcast/multicast flooding in the BD
	UnknownUnicastFlood bool                                `protobuf:"varint,3,opt,name=unknown_unicast_flood,json=unknownUnicastFlood,proto3" json:"unknown_unicast_flood,omitempty"`  // enable/disable unknown unicast flood in the BD
	Forward             bool                                `protobuf:"varint,4,opt,name=forward,proto3" json:"forward,omitempty"`                                                       // enable/disable forwarding on all interfaces in the BD
	Learn               bool                                `protobuf:"varint,5,opt,name=learn,proto3" json:"learn,omitempty"`                                                           // enable/disable learning on all interfaces in the BD
	ArpTermination      bool                                `protobuf:"varint,6,opt,name=arp_termination,json=arpTermination,proto3" json:"arp_termination,omitempty"`                   // enable/disable ARP termination in the BD
	MacAge              uint32                              `protobuf:"varint,7,opt,name=mac_age,json=macAge,proto3" json:"mac_age,omitempty"`                                           // MAC aging time in min, 0 for disabled aging
	Interfaces          []*BridgeDomain_Interface           `protobuf:"bytes,100,rep,name=interfaces,proto3" json:"interfaces,omitempty"`                                                // list of interfaces
	ArpTerminationTable []*BridgeDomain_ArpTerminationEntry `protobuf:"bytes,102,rep,name=arp_termination_table,json=arpTerminationTable,proto3" json:"arp_termination_table,omitempty"` // list of ARP termination entries
}

func (x *BridgeDomain) Reset() {
	*x = BridgeDomain{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_l2_bridge_domain_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BridgeDomain) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BridgeDomain) ProtoMessage() {}

func (x *BridgeDomain) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_l2_bridge_domain_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BridgeDomain.ProtoReflect.Descriptor instead.
func (*BridgeDomain) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_l2_bridge_domain_proto_rawDescGZIP(), []int{0}
}

func (x *BridgeDomain) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *BridgeDomain) GetFlood() bool {
	if x != nil {
		return x.Flood
	}
	return false
}

func (x *BridgeDomain) GetUnknownUnicastFlood() bool {
	if x != nil {
		return x.UnknownUnicastFlood
	}
	return false
}

func (x *BridgeDomain) GetForward() bool {
	if x != nil {
		return x.Forward
	}
	return false
}

func (x *BridgeDomain) GetLearn() bool {
	if x != nil {
		return x.Learn
	}
	return false
}

func (x *BridgeDomain) GetArpTermination() bool {
	if x != nil {
		return x.ArpTermination
	}
	return false
}

func (x *BridgeDomain) GetMacAge() uint32 {
	if x != nil {
		return x.MacAge
	}
	return 0
}

func (x *BridgeDomain) GetInterfaces() []*BridgeDomain_Interface {
	if x != nil {
		return x.Interfaces
	}
	return nil
}

func (x *BridgeDomain) GetArpTerminationTable() []*BridgeDomain_ArpTerminationEntry {
	if x != nil {
		return x.ArpTerminationTable
	}
	return nil
}

type BridgeDomain_Interface struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name                    string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`                                                                         // interface name belonging to this bridge domain
	BridgedVirtualInterface bool   `protobuf:"varint,2,opt,name=bridged_virtual_interface,json=bridgedVirtualInterface,proto3" json:"bridged_virtual_interface,omitempty"` // true if this is a BVI interface
	SplitHorizonGroup       uint32 `protobuf:"varint,3,opt,name=split_horizon_group,json=splitHorizonGroup,proto3" json:"split_horizon_group,omitempty"`                   // VXLANs in the same BD need the same non-zero SHG
}

func (x *BridgeDomain_Interface) Reset() {
	*x = BridgeDomain_Interface{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_l2_bridge_domain_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BridgeDomain_Interface) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BridgeDomain_Interface) ProtoMessage() {}

func (x *BridgeDomain_Interface) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_l2_bridge_domain_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BridgeDomain_Interface.ProtoReflect.Descriptor instead.
func (*BridgeDomain_Interface) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_l2_bridge_domain_proto_rawDescGZIP(), []int{0, 0}
}

func (x *BridgeDomain_Interface) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *BridgeDomain_Interface) GetBridgedVirtualInterface() bool {
	if x != nil {
		return x.BridgedVirtualInterface
	}
	return false
}

func (x *BridgeDomain_Interface) GetSplitHorizonGroup() uint32 {
	if x != nil {
		return x.SplitHorizonGroup
	}
	return 0
}

type BridgeDomain_ArpTerminationEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IpAddress   string `protobuf:"bytes,1,opt,name=ip_address,json=ipAddress,proto3" json:"ip_address,omitempty"`       // IP address
	PhysAddress string `protobuf:"bytes,2,opt,name=phys_address,json=physAddress,proto3" json:"phys_address,omitempty"` // MAC address matching to the IP
}

func (x *BridgeDomain_ArpTerminationEntry) Reset() {
	*x = BridgeDomain_ArpTerminationEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_l2_bridge_domain_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BridgeDomain_ArpTerminationEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BridgeDomain_ArpTerminationEntry) ProtoMessage() {}

func (x *BridgeDomain_ArpTerminationEntry) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_l2_bridge_domain_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BridgeDomain_ArpTerminationEntry.ProtoReflect.Descriptor instead.
func (*BridgeDomain_ArpTerminationEntry) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_l2_bridge_domain_proto_rawDescGZIP(), []int{0, 1}
}

func (x *BridgeDomain_ArpTerminationEntry) GetIpAddress() string {
	if x != nil {
		return x.IpAddress
	}
	return ""
}

func (x *BridgeDomain_ArpTerminationEntry) GetPhysAddress() string {
	if x != nil {
		return x.PhysAddress
	}
	return ""
}

var File_ligato_vpp_l2_bridge_domain_proto protoreflect.FileDescriptor

var file_ligato_vpp_l2_bridge_domain_proto_rawDesc = []byte{
	0x0a, 0x21, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x6c, 0x32, 0x2f,
	0x62, 0x72, 0x69, 0x64, 0x67, 0x65, 0x5f, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e,
	0x6c, 0x32, 0x1a, 0x18, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf8, 0x04, 0x0a,
	0x0c, 0x42, 0x72, 0x69, 0x64, 0x67, 0x65, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x6c, 0x6f, 0x6f, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x05, 0x66, 0x6c, 0x6f, 0x6f, 0x64, 0x12, 0x32, 0x0a, 0x15, 0x75, 0x6e, 0x6b, 0x6e, 0x6f,
	0x77, 0x6e, 0x5f, 0x75, 0x6e, 0x69, 0x63, 0x61, 0x73, 0x74, 0x5f, 0x66, 0x6c, 0x6f, 0x6f, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x13, 0x75, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x55,
	0x6e, 0x69, 0x63, 0x61, 0x73, 0x74, 0x46, 0x6c, 0x6f, 0x6f, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x66,
	0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x66, 0x6f,
	0x72, 0x77, 0x61, 0x72, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x65, 0x61, 0x72, 0x6e, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x6c, 0x65, 0x61, 0x72, 0x6e, 0x12, 0x27, 0x0a, 0x0f, 0x61,
	0x72, 0x70, 0x5f, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x0e, 0x61, 0x72, 0x70, 0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x17, 0x0a, 0x07, 0x6d, 0x61, 0x63, 0x5f, 0x61, 0x67, 0x65, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x6d, 0x61, 0x63, 0x41, 0x67, 0x65, 0x12, 0x45, 0x0a,
	0x0a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x73, 0x18, 0x64, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x25, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x6c,
	0x32, 0x2e, 0x42, 0x72, 0x69, 0x64, 0x67, 0x65, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x49,
	0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x52, 0x0a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66,
	0x61, 0x63, 0x65, 0x73, 0x12, 0x63, 0x0a, 0x15, 0x61, 0x72, 0x70, 0x5f, 0x74, 0x65, 0x72, 0x6d,
	0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x66, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70,
	0x2e, 0x6c, 0x32, 0x2e, 0x42, 0x72, 0x69, 0x64, 0x67, 0x65, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e,
	0x2e, 0x41, 0x72, 0x70, 0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x13, 0x61, 0x72, 0x70, 0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x1a, 0x8b, 0x01, 0x0a, 0x09, 0x49, 0x6e,
	0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x3a, 0x0a, 0x19, 0x62,
	0x72, 0x69, 0x64, 0x67, 0x65, 0x64, 0x5f, 0x76, 0x69, 0x72, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x17,
	0x62, 0x72, 0x69, 0x64, 0x67, 0x65, 0x64, 0x56, 0x69, 0x72, 0x74, 0x75, 0x61, 0x6c, 0x49, 0x6e,
	0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x2e, 0x0a, 0x13, 0x73, 0x70, 0x6c, 0x69, 0x74,
	0x5f, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x6f, 0x6e, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x11, 0x73, 0x70, 0x6c, 0x69, 0x74, 0x48, 0x6f, 0x72, 0x69, 0x7a,
	0x6f, 0x6e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x1a, 0x5e, 0x0a, 0x13, 0x41, 0x72, 0x70, 0x54, 0x65,
	0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x24,
	0x0a, 0x0a, 0x69, 0x70, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x05, 0x82, 0x7d, 0x02, 0x08, 0x01, 0x52, 0x09, 0x69, 0x70, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x70, 0x68, 0x79, 0x73, 0x5f, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x68, 0x79, 0x73,
	0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x42, 0x36, 0x5a, 0x34, 0x67, 0x6f, 0x2e, 0x6c, 0x69,
	0x67, 0x61, 0x74, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2d, 0x61, 0x67, 0x65, 0x6e,
	0x74, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x69, 0x67, 0x61, 0x74,
	0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x6c, 0x32, 0x3b, 0x76, 0x70, 0x70, 0x5f, 0x6c, 0x32, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ligato_vpp_l2_bridge_domain_proto_rawDescOnce sync.Once
	file_ligato_vpp_l2_bridge_domain_proto_rawDescData = file_ligato_vpp_l2_bridge_domain_proto_rawDesc
)

func file_ligato_vpp_l2_bridge_domain_proto_rawDescGZIP() []byte {
	file_ligato_vpp_l2_bridge_domain_proto_rawDescOnce.Do(func() {
		file_ligato_vpp_l2_bridge_domain_proto_rawDescData = protoimpl.X.CompressGZIP(file_ligato_vpp_l2_bridge_domain_proto_rawDescData)
	})
	return file_ligato_vpp_l2_bridge_domain_proto_rawDescData
}

var file_ligato_vpp_l2_bridge_domain_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_ligato_vpp_l2_bridge_domain_proto_goTypes = []interface{}{
	(*BridgeDomain)(nil),                     // 0: ligato.vpp.l2.BridgeDomain
	(*BridgeDomain_Interface)(nil),           // 1: ligato.vpp.l2.BridgeDomain.Interface
	(*BridgeDomain_ArpTerminationEntry)(nil), // 2: ligato.vpp.l2.BridgeDomain.ArpTerminationEntry
}
var file_ligato_vpp_l2_bridge_domain_proto_depIdxs = []int32{
	1, // 0: ligato.vpp.l2.BridgeDomain.interfaces:type_name -> ligato.vpp.l2.BridgeDomain.Interface
	2, // 1: ligato.vpp.l2.BridgeDomain.arp_termination_table:type_name -> ligato.vpp.l2.BridgeDomain.ArpTerminationEntry
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_ligato_vpp_l2_bridge_domain_proto_init() }
func file_ligato_vpp_l2_bridge_domain_proto_init() {
	if File_ligato_vpp_l2_bridge_domain_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ligato_vpp_l2_bridge_domain_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BridgeDomain); i {
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
		file_ligato_vpp_l2_bridge_domain_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BridgeDomain_Interface); i {
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
		file_ligato_vpp_l2_bridge_domain_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BridgeDomain_ArpTerminationEntry); i {
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
			RawDescriptor: file_ligato_vpp_l2_bridge_domain_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ligato_vpp_l2_bridge_domain_proto_goTypes,
		DependencyIndexes: file_ligato_vpp_l2_bridge_domain_proto_depIdxs,
		MessageInfos:      file_ligato_vpp_l2_bridge_domain_proto_msgTypes,
	}.Build()
	File_ligato_vpp_l2_bridge_domain_proto = out.File
	file_ligato_vpp_l2_bridge_domain_proto_rawDesc = nil
	file_ligato_vpp_l2_bridge_domain_proto_goTypes = nil
	file_ligato_vpp_l2_bridge_domain_proto_depIdxs = nil
}
