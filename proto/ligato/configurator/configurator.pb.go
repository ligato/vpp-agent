// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: ligato/configurator/configurator.proto

package configurator

import (
	linux "go.ligato.io/vpp-agent/v3/proto/ligato/linux"
	netalloc "go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	vpp "go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
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

// Config describes all supported configs into a single config message.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	VppConfig      *vpp.ConfigData      `protobuf:"bytes,1,opt,name=vpp_config,json=vppConfig,proto3" json:"vpp_config,omitempty"`
	LinuxConfig    *linux.ConfigData    `protobuf:"bytes,2,opt,name=linux_config,json=linuxConfig,proto3" json:"linux_config,omitempty"`
	NetallocConfig *netalloc.ConfigData `protobuf:"bytes,3,opt,name=netalloc_config,json=netallocConfig,proto3" json:"netalloc_config,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetVppConfig() *vpp.ConfigData {
	if x != nil {
		return x.VppConfig
	}
	return nil
}

func (x *Config) GetLinuxConfig() *linux.ConfigData {
	if x != nil {
		return x.LinuxConfig
	}
	return nil
}

func (x *Config) GetNetallocConfig() *netalloc.ConfigData {
	if x != nil {
		return x.NetallocConfig
	}
	return nil
}

// Notification describes all known notifications into a single message.
type Notification struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Notification:
	//	*Notification_VppNotification
	//	*Notification_LinuxNotification
	Notification isNotification_Notification `protobuf_oneof:"notification"`
}

func (x *Notification) Reset() {
	*x = Notification{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Notification) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Notification) ProtoMessage() {}

func (x *Notification) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Notification.ProtoReflect.Descriptor instead.
func (*Notification) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{1}
}

func (m *Notification) GetNotification() isNotification_Notification {
	if m != nil {
		return m.Notification
	}
	return nil
}

func (x *Notification) GetVppNotification() *vpp.Notification {
	if x, ok := x.GetNotification().(*Notification_VppNotification); ok {
		return x.VppNotification
	}
	return nil
}

func (x *Notification) GetLinuxNotification() *linux.Notification {
	if x, ok := x.GetNotification().(*Notification_LinuxNotification); ok {
		return x.LinuxNotification
	}
	return nil
}

type isNotification_Notification interface {
	isNotification_Notification()
}

type Notification_VppNotification struct {
	VppNotification *vpp.Notification `protobuf:"bytes,1,opt,name=vpp_notification,json=vppNotification,proto3,oneof"`
}

type Notification_LinuxNotification struct {
	LinuxNotification *linux.Notification `protobuf:"bytes,2,opt,name=linux_notification,json=linuxNotification,proto3,oneof"`
}

func (*Notification_VppNotification) isNotification_Notification() {}

func (*Notification_LinuxNotification) isNotification_Notification() {}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Update is a config data to be updated.
	Update *Config `protobuf:"bytes,1,opt,name=update,proto3" json:"update,omitempty"`
	// FullResync option can be used to overwrite
	// all existing config with config update.
	//
	// NOTE: Using FullResync with empty config update will
	// remove all existing config.
	FullResync bool `protobuf:"varint,2,opt,name=full_resync,json=fullResync,proto3" json:"full_resync,omitempty"`
	// WaitDone option can be used to block until either
	// config update is done (non-pending) or request times out.
	//
	// NOTE: WaitDone is intended to be used for config updates
	// that depend on some event from dataplane to fully configure.
	// Using this with incomplete config updates will require
	// another update request to unblock.
	WaitDone bool `protobuf:"varint,3,opt,name=wait_done,json=waitDone,proto3" json:"wait_done,omitempty"`
}

func (x *UpdateRequest) Reset() {
	*x = UpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRequest) ProtoMessage() {}

func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRequest.ProtoReflect.Descriptor instead.
func (*UpdateRequest) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateRequest) GetUpdate() *Config {
	if x != nil {
		return x.Update
	}
	return nil
}

func (x *UpdateRequest) GetFullResync() bool {
	if x != nil {
		return x.FullResync
	}
	return false
}

func (x *UpdateRequest) GetWaitDone() bool {
	if x != nil {
		return x.WaitDone
	}
	return false
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UpdateResponse) Reset() {
	*x = UpdateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateResponse) ProtoMessage() {}

func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateResponse.ProtoReflect.Descriptor instead.
func (*UpdateResponse) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{3}
}

type DeleteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Delete is a config data to be deleted.
	Delete *Config `protobuf:"bytes,1,opt,name=delete,proto3" json:"delete,omitempty"`
	// WaitDone option can be used to block until either
	// config delete is done (non-pending) or request times out.
	//
	// NOTE: WaitDone is intended to be used for config updates
	// that depend on some event from dataplane to fully configure.
	// Using this with incomplete config updates will require
	// another update request to unblock.
	WaitDone bool `protobuf:"varint,3,opt,name=wait_done,json=waitDone,proto3" json:"wait_done,omitempty"`
}

func (x *DeleteRequest) Reset() {
	*x = DeleteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteRequest) ProtoMessage() {}

func (x *DeleteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteRequest.ProtoReflect.Descriptor instead.
func (*DeleteRequest) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteRequest) GetDelete() *Config {
	if x != nil {
		return x.Delete
	}
	return nil
}

func (x *DeleteRequest) GetWaitDone() bool {
	if x != nil {
		return x.WaitDone
	}
	return false
}

type DeleteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteResponse) Reset() {
	*x = DeleteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteResponse) ProtoMessage() {}

func (x *DeleteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteResponse.ProtoReflect.Descriptor instead.
func (*DeleteResponse) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{5}
}

type GetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetRequest) Reset() {
	*x = GetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRequest) ProtoMessage() {}

func (x *GetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRequest.ProtoReflect.Descriptor instead.
func (*GetRequest) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{6}
}

type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Config describes desired config retrieved from agent.
	Config *Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *GetResponse) Reset() {
	*x = GetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetResponse) ProtoMessage() {}

func (x *GetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetResponse.ProtoReflect.Descriptor instead.
func (*GetResponse) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{7}
}

func (x *GetResponse) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

type DumpRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DumpRequest) Reset() {
	*x = DumpRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DumpRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DumpRequest) ProtoMessage() {}

func (x *DumpRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DumpRequest.ProtoReflect.Descriptor instead.
func (*DumpRequest) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{8}
}

type DumpResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Dump is a running config.
	Dump *Config `protobuf:"bytes,1,opt,name=dump,proto3" json:"dump,omitempty"`
}

func (x *DumpResponse) Reset() {
	*x = DumpResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DumpResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DumpResponse) ProtoMessage() {}

func (x *DumpResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DumpResponse.ProtoReflect.Descriptor instead.
func (*DumpResponse) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{9}
}

func (x *DumpResponse) GetDump() *Config {
	if x != nil {
		return x.Dump
	}
	return nil
}

type NotifyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Idx     uint32          `protobuf:"varint,1,opt,name=idx,proto3" json:"idx,omitempty"`
	Filters []*Notification `protobuf:"bytes,2,rep,name=filters,proto3" json:"filters,omitempty"`
}

func (x *NotifyRequest) Reset() {
	*x = NotifyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyRequest) ProtoMessage() {}

func (x *NotifyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyRequest.ProtoReflect.Descriptor instead.
func (*NotifyRequest) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{10}
}

func (x *NotifyRequest) GetIdx() uint32 {
	if x != nil {
		return x.Idx
	}
	return 0
}

func (x *NotifyRequest) GetFilters() []*Notification {
	if x != nil {
		return x.Filters
	}
	return nil
}

type NotifyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Index of next notification
	NextIdx uint32 `protobuf:"varint,1,opt,name=next_idx,json=nextIdx,proto3" json:"next_idx,omitempty"`
	// Notification contains notification data.
	Notification *Notification `protobuf:"bytes,2,opt,name=notification,proto3" json:"notification,omitempty"`
}

func (x *NotifyResponse) Reset() {
	*x = NotifyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_configurator_configurator_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyResponse) ProtoMessage() {}

func (x *NotifyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_configurator_configurator_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyResponse.ProtoReflect.Descriptor instead.
func (*NotifyResponse) Descriptor() ([]byte, []int) {
	return file_ligato_configurator_configurator_proto_rawDescGZIP(), []int{11}
}

func (x *NotifyResponse) GetNextIdx() uint32 {
	if x != nil {
		return x.NextIdx
	}
	return 0
}

func (x *NotifyResponse) GetNotification() *Notification {
	if x != nil {
		return x.Notification
	}
	return nil
}

var File_ligato_configurator_configurator_proto protoreflect.FileDescriptor

var file_ligato_configurator_configurator_proto_rawDesc = []byte{
	0x0a, 0x26, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74,
	0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f,
	0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x1a, 0x14, 0x6c,
	0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x76, 0x70, 0x70, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x6c, 0x69, 0x6e, 0x75,
	0x78, 0x2f, 0x6c, 0x69, 0x6e, 0x75, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x6c,
	0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x2f, 0x6e,
	0x65, 0x74, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc2, 0x01,
	0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x35, 0x0a, 0x0a, 0x76, 0x70, 0x70, 0x5f,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6c,
	0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x44, 0x61, 0x74, 0x61, 0x52, 0x09, 0x76, 0x70, 0x70, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12,
	0x3b, 0x0a, 0x0c, 0x6c, 0x69, 0x6e, 0x75, 0x78, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x6c,
	0x69, 0x6e, 0x75, 0x78, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x44, 0x61, 0x74, 0x61, 0x52,
	0x0b, 0x6c, 0x69, 0x6e, 0x75, 0x78, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x44, 0x0a, 0x0f,
	0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x6e,
	0x65, 0x74, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x44, 0x61,
	0x74, 0x61, 0x52, 0x0e, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x22, 0xb2, 0x01, 0x0a, 0x0c, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x45, 0x0a, 0x10, 0x76, 0x70, 0x70, 0x5f, 0x6e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e,
	0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x0f, 0x76, 0x70, 0x70, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x4b, 0x0a, 0x12, 0x6c, 0x69,
	0x6e, 0x75, 0x78, 0x5f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e,
	0x6c, 0x69, 0x6e, 0x75, 0x78, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x48, 0x00, 0x52, 0x11, 0x6c, 0x69, 0x6e, 0x75, 0x78, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x0e, 0x0a, 0x0c, 0x6e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x82, 0x01, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x33, 0x0a, 0x06, 0x75, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6c, 0x69, 0x67, 0x61,
	0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1f,
	0x0a, 0x0b, 0x66, 0x75, 0x6c, 0x6c, 0x5f, 0x72, 0x65, 0x73, 0x79, 0x6e, 0x63, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x0a, 0x66, 0x75, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x79, 0x6e, 0x63, 0x12,
	0x1b, 0x0a, 0x09, 0x77, 0x61, 0x69, 0x74, 0x5f, 0x64, 0x6f, 0x6e, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x08, 0x77, 0x61, 0x69, 0x74, 0x44, 0x6f, 0x6e, 0x65, 0x22, 0x10, 0x0a, 0x0e,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x61,
	0x0a, 0x0d, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x33, 0x0a, 0x06, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1b, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x64, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x77, 0x61, 0x69, 0x74, 0x5f, 0x64, 0x6f, 0x6e,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x77, 0x61, 0x69, 0x74, 0x44, 0x6f, 0x6e,
	0x65, 0x22, 0x10, 0x0a, 0x0e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x0c, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x22, 0x42, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x33, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1b, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x0d, 0x0a, 0x0b, 0x44, 0x75, 0x6d, 0x70, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x3f, 0x0a, 0x0c, 0x44, 0x75, 0x6d, 0x70, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2f, 0x0a, 0x04, 0x64, 0x75, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52,
	0x04, 0x64, 0x75, 0x6d, 0x70, 0x22, 0x5e, 0x0a, 0x0d, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x78, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x03, 0x69, 0x64, 0x78, 0x12, 0x3b, 0x0a, 0x07, 0x66, 0x69, 0x6c, 0x74,
	0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6c, 0x69, 0x67, 0x61,
	0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x66, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x73, 0x22, 0x72, 0x0a, 0x0e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x6e, 0x65, 0x78, 0x74, 0x5f,
	0x69, 0x64, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x6e, 0x65, 0x78, 0x74, 0x49,
	0x64, 0x78, 0x12, 0x45, 0x0a, 0x0c, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74,
	0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e,
	0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x6e, 0x6f, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x32, 0xa7, 0x03, 0x0a, 0x13, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x48, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x1f, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74,
	0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x6c, 0x69, 0x67, 0x61,
	0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e,
	0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x51, 0x0a, 0x06, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x22, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x6c, 0x69, 0x67, 0x61,
	0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x51,
	0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x22, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74,
	0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x44,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x6c,
	0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74,
	0x6f, 0x72, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x4b, 0x0a, 0x04, 0x44, 0x75, 0x6d, 0x70, 0x12, 0x20, 0x2e, 0x6c, 0x69, 0x67, 0x61,
	0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e,
	0x44, 0x75, 0x6d, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x6c, 0x69,
	0x67, 0x61, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f,
	0x72, 0x2e, 0x44, 0x75, 0x6d, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x53,
	0x0a, 0x06, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x12, 0x22, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74,
	0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e,
	0x6f, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x6c,
	0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74,
	0x6f, 0x72, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x30, 0x01, 0x42, 0x42, 0x5a, 0x40, 0x67, 0x6f, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f,
	0x2e, 0x69, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2d, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x76, 0x33,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x3b, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x75, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ligato_configurator_configurator_proto_rawDescOnce sync.Once
	file_ligato_configurator_configurator_proto_rawDescData = file_ligato_configurator_configurator_proto_rawDesc
)

func file_ligato_configurator_configurator_proto_rawDescGZIP() []byte {
	file_ligato_configurator_configurator_proto_rawDescOnce.Do(func() {
		file_ligato_configurator_configurator_proto_rawDescData = protoimpl.X.CompressGZIP(file_ligato_configurator_configurator_proto_rawDescData)
	})
	return file_ligato_configurator_configurator_proto_rawDescData
}

var file_ligato_configurator_configurator_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_ligato_configurator_configurator_proto_goTypes = []interface{}{
	(*Config)(nil),              // 0: ligato.configurator.Config
	(*Notification)(nil),        // 1: ligato.configurator.Notification
	(*UpdateRequest)(nil),       // 2: ligato.configurator.UpdateRequest
	(*UpdateResponse)(nil),      // 3: ligato.configurator.UpdateResponse
	(*DeleteRequest)(nil),       // 4: ligato.configurator.DeleteRequest
	(*DeleteResponse)(nil),      // 5: ligato.configurator.DeleteResponse
	(*GetRequest)(nil),          // 6: ligato.configurator.GetRequest
	(*GetResponse)(nil),         // 7: ligato.configurator.GetResponse
	(*DumpRequest)(nil),         // 8: ligato.configurator.DumpRequest
	(*DumpResponse)(nil),        // 9: ligato.configurator.DumpResponse
	(*NotifyRequest)(nil),       // 10: ligato.configurator.NotifyRequest
	(*NotifyResponse)(nil),      // 11: ligato.configurator.NotifyResponse
	(*vpp.ConfigData)(nil),      // 12: ligato.vpp.ConfigData
	(*linux.ConfigData)(nil),    // 13: ligato.linux.ConfigData
	(*netalloc.ConfigData)(nil), // 14: ligato.netalloc.ConfigData
	(*vpp.Notification)(nil),    // 15: ligato.vpp.Notification
	(*linux.Notification)(nil),  // 16: ligato.linux.Notification
}
var file_ligato_configurator_configurator_proto_depIdxs = []int32{
	12, // 0: ligato.configurator.Config.vpp_config:type_name -> ligato.vpp.ConfigData
	13, // 1: ligato.configurator.Config.linux_config:type_name -> ligato.linux.ConfigData
	14, // 2: ligato.configurator.Config.netalloc_config:type_name -> ligato.netalloc.ConfigData
	15, // 3: ligato.configurator.Notification.vpp_notification:type_name -> ligato.vpp.Notification
	16, // 4: ligato.configurator.Notification.linux_notification:type_name -> ligato.linux.Notification
	0,  // 5: ligato.configurator.UpdateRequest.update:type_name -> ligato.configurator.Config
	0,  // 6: ligato.configurator.DeleteRequest.delete:type_name -> ligato.configurator.Config
	0,  // 7: ligato.configurator.GetResponse.config:type_name -> ligato.configurator.Config
	0,  // 8: ligato.configurator.DumpResponse.dump:type_name -> ligato.configurator.Config
	1,  // 9: ligato.configurator.NotifyRequest.filters:type_name -> ligato.configurator.Notification
	1,  // 10: ligato.configurator.NotifyResponse.notification:type_name -> ligato.configurator.Notification
	6,  // 11: ligato.configurator.ConfiguratorService.Get:input_type -> ligato.configurator.GetRequest
	2,  // 12: ligato.configurator.ConfiguratorService.Update:input_type -> ligato.configurator.UpdateRequest
	4,  // 13: ligato.configurator.ConfiguratorService.Delete:input_type -> ligato.configurator.DeleteRequest
	8,  // 14: ligato.configurator.ConfiguratorService.Dump:input_type -> ligato.configurator.DumpRequest
	10, // 15: ligato.configurator.ConfiguratorService.Notify:input_type -> ligato.configurator.NotifyRequest
	7,  // 16: ligato.configurator.ConfiguratorService.Get:output_type -> ligato.configurator.GetResponse
	3,  // 17: ligato.configurator.ConfiguratorService.Update:output_type -> ligato.configurator.UpdateResponse
	5,  // 18: ligato.configurator.ConfiguratorService.Delete:output_type -> ligato.configurator.DeleteResponse
	9,  // 19: ligato.configurator.ConfiguratorService.Dump:output_type -> ligato.configurator.DumpResponse
	11, // 20: ligato.configurator.ConfiguratorService.Notify:output_type -> ligato.configurator.NotifyResponse
	16, // [16:21] is the sub-list for method output_type
	11, // [11:16] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_ligato_configurator_configurator_proto_init() }
func file_ligato_configurator_configurator_proto_init() {
	if File_ligato_configurator_configurator_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ligato_configurator_configurator_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Notification); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRequest); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateResponse); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteRequest); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteResponse); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRequest); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetResponse); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DumpRequest); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DumpResponse); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyRequest); i {
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
		file_ligato_configurator_configurator_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyResponse); i {
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
	file_ligato_configurator_configurator_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*Notification_VppNotification)(nil),
		(*Notification_LinuxNotification)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ligato_configurator_configurator_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ligato_configurator_configurator_proto_goTypes,
		DependencyIndexes: file_ligato_configurator_configurator_proto_depIdxs,
		MessageInfos:      file_ligato_configurator_configurator_proto_msgTypes,
	}.Build()
	File_ligato_configurator_configurator_proto = out.File
	file_ligato_configurator_configurator_proto_rawDesc = nil
	file_ligato_configurator_configurator_proto_goTypes = nil
	file_ligato_configurator_configurator_proto_depIdxs = nil
}
