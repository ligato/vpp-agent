// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp/srv6/srv6.proto

package vpp_srv6

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

type LocalSID struct {
	Sid               string `protobuf:"bytes,1,opt,name=sid,proto3" json:"sid,omitempty"`
	InstallationVrfId uint32 `protobuf:"varint,2,opt,name=installation_vrf_id,json=installationVrfId,proto3" json:"installation_vrf_id,omitempty"`
	// Configuration for end functions (all end functions are mutually exclusive)
	//
	// Types that are valid to be assigned to EndFunction:
	//	*LocalSID_BaseEndFunction
	//	*LocalSID_EndFunctionX
	//	*LocalSID_EndFunctionT
	//	*LocalSID_EndFunctionDx2
	//	*LocalSID_EndFunctionDx4
	//	*LocalSID_EndFunctionDx6
	//	*LocalSID_EndFunctionDt4
	//	*LocalSID_EndFunctionDt6
	//	*LocalSID_EndFunctionAd
	EndFunction          isLocalSID_EndFunction `protobuf_oneof:"end_function"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *LocalSID) Reset()         { *m = LocalSID{} }
func (m *LocalSID) String() string { return proto.CompactTextString(m) }
func (*LocalSID) ProtoMessage()    {}
func (*LocalSID) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0}
}

func (m *LocalSID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID.Unmarshal(m, b)
}
func (m *LocalSID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID.Marshal(b, m, deterministic)
}
func (m *LocalSID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID.Merge(m, src)
}
func (m *LocalSID) XXX_Size() int {
	return xxx_messageInfo_LocalSID.Size(m)
}
func (m *LocalSID) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID proto.InternalMessageInfo

func (m *LocalSID) GetSid() string {
	if m != nil {
		return m.Sid
	}
	return ""
}

func (m *LocalSID) GetInstallationVrfId() uint32 {
	if m != nil {
		return m.InstallationVrfId
	}
	return 0
}

type isLocalSID_EndFunction interface {
	isLocalSID_EndFunction()
}

type LocalSID_BaseEndFunction struct {
	BaseEndFunction *LocalSID_End `protobuf:"bytes,3,opt,name=base_end_function,json=baseEndFunction,proto3,oneof"`
}

type LocalSID_EndFunctionX struct {
	EndFunctionX *LocalSID_EndX `protobuf:"bytes,4,opt,name=end_function_x,json=endFunctionX,proto3,oneof"`
}

type LocalSID_EndFunctionT struct {
	EndFunctionT *LocalSID_EndT `protobuf:"bytes,5,opt,name=end_function_t,json=endFunctionT,proto3,oneof"`
}

type LocalSID_EndFunctionDx2 struct {
	EndFunctionDx2 *LocalSID_EndDX2 `protobuf:"bytes,6,opt,name=end_function_dx2,json=endFunctionDx2,proto3,oneof"`
}

type LocalSID_EndFunctionDx4 struct {
	EndFunctionDx4 *LocalSID_EndDX4 `protobuf:"bytes,7,opt,name=end_function_dx4,json=endFunctionDx4,proto3,oneof"`
}

type LocalSID_EndFunctionDx6 struct {
	EndFunctionDx6 *LocalSID_EndDX6 `protobuf:"bytes,8,opt,name=end_function_dx6,json=endFunctionDx6,proto3,oneof"`
}

type LocalSID_EndFunctionDt4 struct {
	EndFunctionDt4 *LocalSID_EndDT4 `protobuf:"bytes,9,opt,name=end_function_dt4,json=endFunctionDt4,proto3,oneof"`
}

type LocalSID_EndFunctionDt6 struct {
	EndFunctionDt6 *LocalSID_EndDT6 `protobuf:"bytes,10,opt,name=end_function_dt6,json=endFunctionDt6,proto3,oneof"`
}

type LocalSID_EndFunctionAd struct {
	EndFunctionAd *LocalSID_EndAD `protobuf:"bytes,11,opt,name=end_function_ad,json=endFunctionAd,proto3,oneof"`
}

func (*LocalSID_BaseEndFunction) isLocalSID_EndFunction() {}

func (*LocalSID_EndFunctionX) isLocalSID_EndFunction() {}

func (*LocalSID_EndFunctionT) isLocalSID_EndFunction() {}

func (*LocalSID_EndFunctionDx2) isLocalSID_EndFunction() {}

func (*LocalSID_EndFunctionDx4) isLocalSID_EndFunction() {}

func (*LocalSID_EndFunctionDx6) isLocalSID_EndFunction() {}

func (*LocalSID_EndFunctionDt4) isLocalSID_EndFunction() {}

func (*LocalSID_EndFunctionDt6) isLocalSID_EndFunction() {}

func (*LocalSID_EndFunctionAd) isLocalSID_EndFunction() {}

func (m *LocalSID) GetEndFunction() isLocalSID_EndFunction {
	if m != nil {
		return m.EndFunction
	}
	return nil
}

func (m *LocalSID) GetBaseEndFunction() *LocalSID_End {
	if x, ok := m.GetEndFunction().(*LocalSID_BaseEndFunction); ok {
		return x.BaseEndFunction
	}
	return nil
}

func (m *LocalSID) GetEndFunctionX() *LocalSID_EndX {
	if x, ok := m.GetEndFunction().(*LocalSID_EndFunctionX); ok {
		return x.EndFunctionX
	}
	return nil
}

func (m *LocalSID) GetEndFunctionT() *LocalSID_EndT {
	if x, ok := m.GetEndFunction().(*LocalSID_EndFunctionT); ok {
		return x.EndFunctionT
	}
	return nil
}

func (m *LocalSID) GetEndFunctionDx2() *LocalSID_EndDX2 {
	if x, ok := m.GetEndFunction().(*LocalSID_EndFunctionDx2); ok {
		return x.EndFunctionDx2
	}
	return nil
}

func (m *LocalSID) GetEndFunctionDx4() *LocalSID_EndDX4 {
	if x, ok := m.GetEndFunction().(*LocalSID_EndFunctionDx4); ok {
		return x.EndFunctionDx4
	}
	return nil
}

func (m *LocalSID) GetEndFunctionDx6() *LocalSID_EndDX6 {
	if x, ok := m.GetEndFunction().(*LocalSID_EndFunctionDx6); ok {
		return x.EndFunctionDx6
	}
	return nil
}

func (m *LocalSID) GetEndFunctionDt4() *LocalSID_EndDT4 {
	if x, ok := m.GetEndFunction().(*LocalSID_EndFunctionDt4); ok {
		return x.EndFunctionDt4
	}
	return nil
}

func (m *LocalSID) GetEndFunctionDt6() *LocalSID_EndDT6 {
	if x, ok := m.GetEndFunction().(*LocalSID_EndFunctionDt6); ok {
		return x.EndFunctionDt6
	}
	return nil
}

func (m *LocalSID) GetEndFunctionAd() *LocalSID_EndAD {
	if x, ok := m.GetEndFunction().(*LocalSID_EndFunctionAd); ok {
		return x.EndFunctionAd
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*LocalSID) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*LocalSID_BaseEndFunction)(nil),
		(*LocalSID_EndFunctionX)(nil),
		(*LocalSID_EndFunctionT)(nil),
		(*LocalSID_EndFunctionDx2)(nil),
		(*LocalSID_EndFunctionDx4)(nil),
		(*LocalSID_EndFunctionDx6)(nil),
		(*LocalSID_EndFunctionDt4)(nil),
		(*LocalSID_EndFunctionDt6)(nil),
		(*LocalSID_EndFunctionAd)(nil),
	}
}

// End function behavior of simple endpoint
type LocalSID_End struct {
	Psp                  bool     `protobuf:"varint,1,opt,name=psp,proto3" json:"psp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSID_End) Reset()         { *m = LocalSID_End{} }
func (m *LocalSID_End) String() string { return proto.CompactTextString(m) }
func (*LocalSID_End) ProtoMessage()    {}
func (*LocalSID_End) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0, 0}
}

func (m *LocalSID_End) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID_End.Unmarshal(m, b)
}
func (m *LocalSID_End) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID_End.Marshal(b, m, deterministic)
}
func (m *LocalSID_End) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID_End.Merge(m, src)
}
func (m *LocalSID_End) XXX_Size() int {
	return xxx_messageInfo_LocalSID_End.Size(m)
}
func (m *LocalSID_End) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID_End.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID_End proto.InternalMessageInfo

func (m *LocalSID_End) GetPsp() bool {
	if m != nil {
		return m.Psp
	}
	return false
}

// End function behavior of endpoint with Layer-3 cross-connect (IPv6)
type LocalSID_EndX struct {
	Psp                  bool     `protobuf:"varint,1,opt,name=psp,proto3" json:"psp,omitempty"`
	OutgoingInterface    string   `protobuf:"bytes,2,opt,name=outgoing_interface,json=outgoingInterface,proto3" json:"outgoing_interface,omitempty"`
	NextHop              string   `protobuf:"bytes,3,opt,name=next_hop,json=nextHop,proto3" json:"next_hop,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSID_EndX) Reset()         { *m = LocalSID_EndX{} }
func (m *LocalSID_EndX) String() string { return proto.CompactTextString(m) }
func (*LocalSID_EndX) ProtoMessage()    {}
func (*LocalSID_EndX) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0, 1}
}

func (m *LocalSID_EndX) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID_EndX.Unmarshal(m, b)
}
func (m *LocalSID_EndX) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID_EndX.Marshal(b, m, deterministic)
}
func (m *LocalSID_EndX) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID_EndX.Merge(m, src)
}
func (m *LocalSID_EndX) XXX_Size() int {
	return xxx_messageInfo_LocalSID_EndX.Size(m)
}
func (m *LocalSID_EndX) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID_EndX.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID_EndX proto.InternalMessageInfo

func (m *LocalSID_EndX) GetPsp() bool {
	if m != nil {
		return m.Psp
	}
	return false
}

func (m *LocalSID_EndX) GetOutgoingInterface() string {
	if m != nil {
		return m.OutgoingInterface
	}
	return ""
}

func (m *LocalSID_EndX) GetNextHop() string {
	if m != nil {
		return m.NextHop
	}
	return ""
}

// End function behavior of endpoint with specific IPv6 table lookup
type LocalSID_EndT struct {
	Psp                  bool     `protobuf:"varint,1,opt,name=psp,proto3" json:"psp,omitempty"`
	VrfId                uint32   `protobuf:"varint,2,opt,name=vrf_id,json=vrfId,proto3" json:"vrf_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSID_EndT) Reset()         { *m = LocalSID_EndT{} }
func (m *LocalSID_EndT) String() string { return proto.CompactTextString(m) }
func (*LocalSID_EndT) ProtoMessage()    {}
func (*LocalSID_EndT) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0, 2}
}

func (m *LocalSID_EndT) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID_EndT.Unmarshal(m, b)
}
func (m *LocalSID_EndT) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID_EndT.Marshal(b, m, deterministic)
}
func (m *LocalSID_EndT) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID_EndT.Merge(m, src)
}
func (m *LocalSID_EndT) XXX_Size() int {
	return xxx_messageInfo_LocalSID_EndT.Size(m)
}
func (m *LocalSID_EndT) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID_EndT.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID_EndT proto.InternalMessageInfo

func (m *LocalSID_EndT) GetPsp() bool {
	if m != nil {
		return m.Psp
	}
	return false
}

func (m *LocalSID_EndT) GetVrfId() uint32 {
	if m != nil {
		return m.VrfId
	}
	return 0
}

// End function behavior of endpoint with decapsulation and Layer-2 cross-connect (or DX2 with egress VLAN rewrite when VLAN notzero - not supported this variant yet)
type LocalSID_EndDX2 struct {
	VlanTag              uint32   `protobuf:"varint,1,opt,name=vlan_tag,json=vlanTag,proto3" json:"vlan_tag,omitempty"`
	OutgoingInterface    string   `protobuf:"bytes,2,opt,name=outgoing_interface,json=outgoingInterface,proto3" json:"outgoing_interface,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSID_EndDX2) Reset()         { *m = LocalSID_EndDX2{} }
func (m *LocalSID_EndDX2) String() string { return proto.CompactTextString(m) }
func (*LocalSID_EndDX2) ProtoMessage()    {}
func (*LocalSID_EndDX2) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0, 3}
}

func (m *LocalSID_EndDX2) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID_EndDX2.Unmarshal(m, b)
}
func (m *LocalSID_EndDX2) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID_EndDX2.Marshal(b, m, deterministic)
}
func (m *LocalSID_EndDX2) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID_EndDX2.Merge(m, src)
}
func (m *LocalSID_EndDX2) XXX_Size() int {
	return xxx_messageInfo_LocalSID_EndDX2.Size(m)
}
func (m *LocalSID_EndDX2) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID_EndDX2.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID_EndDX2 proto.InternalMessageInfo

func (m *LocalSID_EndDX2) GetVlanTag() uint32 {
	if m != nil {
		return m.VlanTag
	}
	return 0
}

func (m *LocalSID_EndDX2) GetOutgoingInterface() string {
	if m != nil {
		return m.OutgoingInterface
	}
	return ""
}

// End function behavior of endpoint with decapsulation and IPv4 cross-connect
type LocalSID_EndDX4 struct {
	OutgoingInterface    string   `protobuf:"bytes,1,opt,name=outgoing_interface,json=outgoingInterface,proto3" json:"outgoing_interface,omitempty"`
	NextHop              string   `protobuf:"bytes,2,opt,name=next_hop,json=nextHop,proto3" json:"next_hop,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSID_EndDX4) Reset()         { *m = LocalSID_EndDX4{} }
func (m *LocalSID_EndDX4) String() string { return proto.CompactTextString(m) }
func (*LocalSID_EndDX4) ProtoMessage()    {}
func (*LocalSID_EndDX4) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0, 4}
}

func (m *LocalSID_EndDX4) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID_EndDX4.Unmarshal(m, b)
}
func (m *LocalSID_EndDX4) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID_EndDX4.Marshal(b, m, deterministic)
}
func (m *LocalSID_EndDX4) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID_EndDX4.Merge(m, src)
}
func (m *LocalSID_EndDX4) XXX_Size() int {
	return xxx_messageInfo_LocalSID_EndDX4.Size(m)
}
func (m *LocalSID_EndDX4) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID_EndDX4.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID_EndDX4 proto.InternalMessageInfo

func (m *LocalSID_EndDX4) GetOutgoingInterface() string {
	if m != nil {
		return m.OutgoingInterface
	}
	return ""
}

func (m *LocalSID_EndDX4) GetNextHop() string {
	if m != nil {
		return m.NextHop
	}
	return ""
}

// End function behavior of endpoint with decapsulation and IPv6 cross-connect
type LocalSID_EndDX6 struct {
	OutgoingInterface    string   `protobuf:"bytes,1,opt,name=outgoing_interface,json=outgoingInterface,proto3" json:"outgoing_interface,omitempty"`
	NextHop              string   `protobuf:"bytes,2,opt,name=next_hop,json=nextHop,proto3" json:"next_hop,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSID_EndDX6) Reset()         { *m = LocalSID_EndDX6{} }
func (m *LocalSID_EndDX6) String() string { return proto.CompactTextString(m) }
func (*LocalSID_EndDX6) ProtoMessage()    {}
func (*LocalSID_EndDX6) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0, 5}
}

func (m *LocalSID_EndDX6) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID_EndDX6.Unmarshal(m, b)
}
func (m *LocalSID_EndDX6) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID_EndDX6.Marshal(b, m, deterministic)
}
func (m *LocalSID_EndDX6) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID_EndDX6.Merge(m, src)
}
func (m *LocalSID_EndDX6) XXX_Size() int {
	return xxx_messageInfo_LocalSID_EndDX6.Size(m)
}
func (m *LocalSID_EndDX6) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID_EndDX6.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID_EndDX6 proto.InternalMessageInfo

func (m *LocalSID_EndDX6) GetOutgoingInterface() string {
	if m != nil {
		return m.OutgoingInterface
	}
	return ""
}

func (m *LocalSID_EndDX6) GetNextHop() string {
	if m != nil {
		return m.NextHop
	}
	return ""
}

// End function behavior of endpoint with decapsulation and specific IPv4 table lookup
type LocalSID_EndDT4 struct {
	VrfId                uint32   `protobuf:"varint,1,opt,name=vrf_id,json=vrfId,proto3" json:"vrf_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSID_EndDT4) Reset()         { *m = LocalSID_EndDT4{} }
func (m *LocalSID_EndDT4) String() string { return proto.CompactTextString(m) }
func (*LocalSID_EndDT4) ProtoMessage()    {}
func (*LocalSID_EndDT4) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0, 6}
}

func (m *LocalSID_EndDT4) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID_EndDT4.Unmarshal(m, b)
}
func (m *LocalSID_EndDT4) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID_EndDT4.Marshal(b, m, deterministic)
}
func (m *LocalSID_EndDT4) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID_EndDT4.Merge(m, src)
}
func (m *LocalSID_EndDT4) XXX_Size() int {
	return xxx_messageInfo_LocalSID_EndDT4.Size(m)
}
func (m *LocalSID_EndDT4) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID_EndDT4.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID_EndDT4 proto.InternalMessageInfo

func (m *LocalSID_EndDT4) GetVrfId() uint32 {
	if m != nil {
		return m.VrfId
	}
	return 0
}

// End function behavior of endpoint with decapsulation and specific IPv6 table lookup
type LocalSID_EndDT6 struct {
	VrfId                uint32   `protobuf:"varint,1,opt,name=vrf_id,json=vrfId,proto3" json:"vrf_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSID_EndDT6) Reset()         { *m = LocalSID_EndDT6{} }
func (m *LocalSID_EndDT6) String() string { return proto.CompactTextString(m) }
func (*LocalSID_EndDT6) ProtoMessage()    {}
func (*LocalSID_EndDT6) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0, 7}
}

func (m *LocalSID_EndDT6) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID_EndDT6.Unmarshal(m, b)
}
func (m *LocalSID_EndDT6) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID_EndDT6.Marshal(b, m, deterministic)
}
func (m *LocalSID_EndDT6) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID_EndDT6.Merge(m, src)
}
func (m *LocalSID_EndDT6) XXX_Size() int {
	return xxx_messageInfo_LocalSID_EndDT6.Size(m)
}
func (m *LocalSID_EndDT6) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID_EndDT6.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID_EndDT6 proto.InternalMessageInfo

func (m *LocalSID_EndDT6) GetVrfId() uint32 {
	if m != nil {
		return m.VrfId
	}
	return 0
}

// End function behavior of dynamic segment routing proxy endpoint
type LocalSID_EndAD struct {
	OutgoingInterface    string   `protobuf:"bytes,2,opt,name=outgoing_interface,json=outgoingInterface,proto3" json:"outgoing_interface,omitempty"`
	IncomingInterface    string   `protobuf:"bytes,3,opt,name=incoming_interface,json=incomingInterface,proto3" json:"incoming_interface,omitempty"`
	L3ServiceAddress     string   `protobuf:"bytes,4,opt,name=l3_service_address,json=l3ServiceAddress,proto3" json:"l3_service_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSID_EndAD) Reset()         { *m = LocalSID_EndAD{} }
func (m *LocalSID_EndAD) String() string { return proto.CompactTextString(m) }
func (*LocalSID_EndAD) ProtoMessage()    {}
func (*LocalSID_EndAD) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{0, 8}
}

func (m *LocalSID_EndAD) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSID_EndAD.Unmarshal(m, b)
}
func (m *LocalSID_EndAD) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSID_EndAD.Marshal(b, m, deterministic)
}
func (m *LocalSID_EndAD) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSID_EndAD.Merge(m, src)
}
func (m *LocalSID_EndAD) XXX_Size() int {
	return xxx_messageInfo_LocalSID_EndAD.Size(m)
}
func (m *LocalSID_EndAD) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSID_EndAD.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSID_EndAD proto.InternalMessageInfo

func (m *LocalSID_EndAD) GetOutgoingInterface() string {
	if m != nil {
		return m.OutgoingInterface
	}
	return ""
}

func (m *LocalSID_EndAD) GetIncomingInterface() string {
	if m != nil {
		return m.IncomingInterface
	}
	return ""
}

func (m *LocalSID_EndAD) GetL3ServiceAddress() string {
	if m != nil {
		return m.L3ServiceAddress
	}
	return ""
}

// Model for SRv6 policy (policy without at least one policy segment is only cached in ligato and not written to VPP)
type Policy struct {
	Bsid                 string                `protobuf:"bytes,1,opt,name=bsid,proto3" json:"bsid,omitempty"`
	InstallationVrfId    uint32                `protobuf:"varint,2,opt,name=installation_vrf_id,json=installationVrfId,proto3" json:"installation_vrf_id,omitempty"`
	SrhEncapsulation     bool                  `protobuf:"varint,3,opt,name=srh_encapsulation,json=srhEncapsulation,proto3" json:"srh_encapsulation,omitempty"`
	SprayBehaviour       bool                  `protobuf:"varint,4,opt,name=spray_behaviour,json=sprayBehaviour,proto3" json:"spray_behaviour,omitempty"`
	SegmentLists         []*Policy_SegmentList `protobuf:"bytes,5,rep,name=segment_lists,json=segmentLists,proto3" json:"segment_lists,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *Policy) Reset()         { *m = Policy{} }
func (m *Policy) String() string { return proto.CompactTextString(m) }
func (*Policy) ProtoMessage()    {}
func (*Policy) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{1}
}

func (m *Policy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Policy.Unmarshal(m, b)
}
func (m *Policy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Policy.Marshal(b, m, deterministic)
}
func (m *Policy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Policy.Merge(m, src)
}
func (m *Policy) XXX_Size() int {
	return xxx_messageInfo_Policy.Size(m)
}
func (m *Policy) XXX_DiscardUnknown() {
	xxx_messageInfo_Policy.DiscardUnknown(m)
}

var xxx_messageInfo_Policy proto.InternalMessageInfo

func (m *Policy) GetBsid() string {
	if m != nil {
		return m.Bsid
	}
	return ""
}

func (m *Policy) GetInstallationVrfId() uint32 {
	if m != nil {
		return m.InstallationVrfId
	}
	return 0
}

func (m *Policy) GetSrhEncapsulation() bool {
	if m != nil {
		return m.SrhEncapsulation
	}
	return false
}

func (m *Policy) GetSprayBehaviour() bool {
	if m != nil {
		return m.SprayBehaviour
	}
	return false
}

func (m *Policy) GetSegmentLists() []*Policy_SegmentList {
	if m != nil {
		return m.SegmentLists
	}
	return nil
}

// Model for SRv6 Segment List
type Policy_SegmentList struct {
	Weight               uint32   `protobuf:"varint,1,opt,name=weight,proto3" json:"weight,omitempty"`
	Segments             []string `protobuf:"bytes,2,rep,name=segments,proto3" json:"segments,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Policy_SegmentList) Reset()         { *m = Policy_SegmentList{} }
func (m *Policy_SegmentList) String() string { return proto.CompactTextString(m) }
func (*Policy_SegmentList) ProtoMessage()    {}
func (*Policy_SegmentList) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{1, 0}
}

func (m *Policy_SegmentList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Policy_SegmentList.Unmarshal(m, b)
}
func (m *Policy_SegmentList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Policy_SegmentList.Marshal(b, m, deterministic)
}
func (m *Policy_SegmentList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Policy_SegmentList.Merge(m, src)
}
func (m *Policy_SegmentList) XXX_Size() int {
	return xxx_messageInfo_Policy_SegmentList.Size(m)
}
func (m *Policy_SegmentList) XXX_DiscardUnknown() {
	xxx_messageInfo_Policy_SegmentList.DiscardUnknown(m)
}

var xxx_messageInfo_Policy_SegmentList proto.InternalMessageInfo

func (m *Policy_SegmentList) GetWeight() uint32 {
	if m != nil {
		return m.Weight
	}
	return 0
}

func (m *Policy_SegmentList) GetSegments() []string {
	if m != nil {
		return m.Segments
	}
	return nil
}

// Model for steering traffic to SRv6 policy
type Steering struct {
	Name string `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty"`
	// Referencing policy that should be used for steering traffic into (all policy references are mutual exclusive)
	//
	// Types that are valid to be assigned to PolicyRef:
	//	*Steering_PolicyBsid
	//	*Steering_PolicyIndex
	PolicyRef isSteering_PolicyRef `protobuf_oneof:"policy_ref"`
	// Traffic configuration (all traffic messages are mutual exclusive)
	//
	// Types that are valid to be assigned to Traffic:
	//	*Steering_L2Traffic_
	//	*Steering_L3Traffic_
	Traffic              isSteering_Traffic `protobuf_oneof:"traffic"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *Steering) Reset()         { *m = Steering{} }
func (m *Steering) String() string { return proto.CompactTextString(m) }
func (*Steering) ProtoMessage()    {}
func (*Steering) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{2}
}

func (m *Steering) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Steering.Unmarshal(m, b)
}
func (m *Steering) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Steering.Marshal(b, m, deterministic)
}
func (m *Steering) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Steering.Merge(m, src)
}
func (m *Steering) XXX_Size() int {
	return xxx_messageInfo_Steering.Size(m)
}
func (m *Steering) XXX_DiscardUnknown() {
	xxx_messageInfo_Steering.DiscardUnknown(m)
}

var xxx_messageInfo_Steering proto.InternalMessageInfo

func (m *Steering) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type isSteering_PolicyRef interface {
	isSteering_PolicyRef()
}

type Steering_PolicyBsid struct {
	PolicyBsid string `protobuf:"bytes,1,opt,name=policy_bsid,json=policyBsid,proto3,oneof"`
}

type Steering_PolicyIndex struct {
	PolicyIndex uint32 `protobuf:"varint,2,opt,name=policy_index,json=policyIndex,proto3,oneof"`
}

func (*Steering_PolicyBsid) isSteering_PolicyRef() {}

func (*Steering_PolicyIndex) isSteering_PolicyRef() {}

func (m *Steering) GetPolicyRef() isSteering_PolicyRef {
	if m != nil {
		return m.PolicyRef
	}
	return nil
}

func (m *Steering) GetPolicyBsid() string {
	if x, ok := m.GetPolicyRef().(*Steering_PolicyBsid); ok {
		return x.PolicyBsid
	}
	return ""
}

func (m *Steering) GetPolicyIndex() uint32 {
	if x, ok := m.GetPolicyRef().(*Steering_PolicyIndex); ok {
		return x.PolicyIndex
	}
	return 0
}

type isSteering_Traffic interface {
	isSteering_Traffic()
}

type Steering_L2Traffic_ struct {
	L2Traffic *Steering_L2Traffic `protobuf:"bytes,3,opt,name=l2_traffic,json=l2Traffic,proto3,oneof"`
}

type Steering_L3Traffic_ struct {
	L3Traffic *Steering_L3Traffic `protobuf:"bytes,4,opt,name=l3_traffic,json=l3Traffic,proto3,oneof"`
}

func (*Steering_L2Traffic_) isSteering_Traffic() {}

func (*Steering_L3Traffic_) isSteering_Traffic() {}

func (m *Steering) GetTraffic() isSteering_Traffic {
	if m != nil {
		return m.Traffic
	}
	return nil
}

func (m *Steering) GetL2Traffic() *Steering_L2Traffic {
	if x, ok := m.GetTraffic().(*Steering_L2Traffic_); ok {
		return x.L2Traffic
	}
	return nil
}

func (m *Steering) GetL3Traffic() *Steering_L3Traffic {
	if x, ok := m.GetTraffic().(*Steering_L3Traffic_); ok {
		return x.L3Traffic
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Steering) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Steering_PolicyBsid)(nil),
		(*Steering_PolicyIndex)(nil),
		(*Steering_L2Traffic_)(nil),
		(*Steering_L3Traffic_)(nil),
	}
}

type Steering_L2Traffic struct {
	InterfaceName        string   `protobuf:"bytes,1,opt,name=interface_name,json=interfaceName,proto3" json:"interface_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Steering_L2Traffic) Reset()         { *m = Steering_L2Traffic{} }
func (m *Steering_L2Traffic) String() string { return proto.CompactTextString(m) }
func (*Steering_L2Traffic) ProtoMessage()    {}
func (*Steering_L2Traffic) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{2, 0}
}

func (m *Steering_L2Traffic) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Steering_L2Traffic.Unmarshal(m, b)
}
func (m *Steering_L2Traffic) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Steering_L2Traffic.Marshal(b, m, deterministic)
}
func (m *Steering_L2Traffic) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Steering_L2Traffic.Merge(m, src)
}
func (m *Steering_L2Traffic) XXX_Size() int {
	return xxx_messageInfo_Steering_L2Traffic.Size(m)
}
func (m *Steering_L2Traffic) XXX_DiscardUnknown() {
	xxx_messageInfo_Steering_L2Traffic.DiscardUnknown(m)
}

var xxx_messageInfo_Steering_L2Traffic proto.InternalMessageInfo

func (m *Steering_L2Traffic) GetInterfaceName() string {
	if m != nil {
		return m.InterfaceName
	}
	return ""
}

type Steering_L3Traffic struct {
	InstallationVrfId    uint32   `protobuf:"varint,1,opt,name=installation_vrf_id,json=installationVrfId,proto3" json:"installation_vrf_id,omitempty"`
	PrefixAddress        string   `protobuf:"bytes,2,opt,name=prefix_address,json=prefixAddress,proto3" json:"prefix_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Steering_L3Traffic) Reset()         { *m = Steering_L3Traffic{} }
func (m *Steering_L3Traffic) String() string { return proto.CompactTextString(m) }
func (*Steering_L3Traffic) ProtoMessage()    {}
func (*Steering_L3Traffic) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{2, 1}
}

func (m *Steering_L3Traffic) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Steering_L3Traffic.Unmarshal(m, b)
}
func (m *Steering_L3Traffic) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Steering_L3Traffic.Marshal(b, m, deterministic)
}
func (m *Steering_L3Traffic) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Steering_L3Traffic.Merge(m, src)
}
func (m *Steering_L3Traffic) XXX_Size() int {
	return xxx_messageInfo_Steering_L3Traffic.Size(m)
}
func (m *Steering_L3Traffic) XXX_DiscardUnknown() {
	xxx_messageInfo_Steering_L3Traffic.DiscardUnknown(m)
}

var xxx_messageInfo_Steering_L3Traffic proto.InternalMessageInfo

func (m *Steering_L3Traffic) GetInstallationVrfId() uint32 {
	if m != nil {
		return m.InstallationVrfId
	}
	return 0
}

func (m *Steering_L3Traffic) GetPrefixAddress() string {
	if m != nil {
		return m.PrefixAddress
	}
	return ""
}

// Global SRv6 config
type SRv6Global struct {
	EncapSourceAddress   string   `protobuf:"bytes,1,opt,name=encap_source_address,json=encapSourceAddress,proto3" json:"encap_source_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SRv6Global) Reset()         { *m = SRv6Global{} }
func (m *SRv6Global) String() string { return proto.CompactTextString(m) }
func (*SRv6Global) ProtoMessage()    {}
func (*SRv6Global) Descriptor() ([]byte, []int) {
	return fileDescriptor_f16c9933790ee176, []int{3}
}

func (m *SRv6Global) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SRv6Global.Unmarshal(m, b)
}
func (m *SRv6Global) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SRv6Global.Marshal(b, m, deterministic)
}
func (m *SRv6Global) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SRv6Global.Merge(m, src)
}
func (m *SRv6Global) XXX_Size() int {
	return xxx_messageInfo_SRv6Global.Size(m)
}
func (m *SRv6Global) XXX_DiscardUnknown() {
	xxx_messageInfo_SRv6Global.DiscardUnknown(m)
}

var xxx_messageInfo_SRv6Global proto.InternalMessageInfo

func (m *SRv6Global) GetEncapSourceAddress() string {
	if m != nil {
		return m.EncapSourceAddress
	}
	return ""
}

func init() {
	proto.RegisterType((*LocalSID)(nil), "ligato.vpp.srv6.LocalSID")
	proto.RegisterType((*LocalSID_End)(nil), "ligato.vpp.srv6.LocalSID.End")
	proto.RegisterType((*LocalSID_EndX)(nil), "ligato.vpp.srv6.LocalSID.EndX")
	proto.RegisterType((*LocalSID_EndT)(nil), "ligato.vpp.srv6.LocalSID.EndT")
	proto.RegisterType((*LocalSID_EndDX2)(nil), "ligato.vpp.srv6.LocalSID.EndDX2")
	proto.RegisterType((*LocalSID_EndDX4)(nil), "ligato.vpp.srv6.LocalSID.EndDX4")
	proto.RegisterType((*LocalSID_EndDX6)(nil), "ligato.vpp.srv6.LocalSID.EndDX6")
	proto.RegisterType((*LocalSID_EndDT4)(nil), "ligato.vpp.srv6.LocalSID.EndDT4")
	proto.RegisterType((*LocalSID_EndDT6)(nil), "ligato.vpp.srv6.LocalSID.EndDT6")
	proto.RegisterType((*LocalSID_EndAD)(nil), "ligato.vpp.srv6.LocalSID.EndAD")
	proto.RegisterType((*Policy)(nil), "ligato.vpp.srv6.Policy")
	proto.RegisterType((*Policy_SegmentList)(nil), "ligato.vpp.srv6.Policy.SegmentList")
	proto.RegisterType((*Steering)(nil), "ligato.vpp.srv6.Steering")
	proto.RegisterType((*Steering_L2Traffic)(nil), "ligato.vpp.srv6.Steering.L2Traffic")
	proto.RegisterType((*Steering_L3Traffic)(nil), "ligato.vpp.srv6.Steering.L3Traffic")
	proto.RegisterType((*SRv6Global)(nil), "ligato.vpp.srv6.SRv6Global")
}

func init() { proto.RegisterFile("ligato/vpp/srv6/srv6.proto", fileDescriptor_f16c9933790ee176) }

var fileDescriptor_f16c9933790ee176 = []byte{
	// 877 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x96, 0x6d, 0x6f, 0xea, 0x36,
	0x14, 0xc7, 0xa1, 0xb4, 0x94, 0x1c, 0x1e, 0x5a, 0xbc, 0xa7, 0x0c, 0x69, 0xbb, 0xac, 0x57, 0x57,
	0x43, 0xda, 0x0a, 0x13, 0x20, 0x34, 0x6d, 0xd2, 0xa4, 0x56, 0xf4, 0x2e, 0x68, 0xd5, 0x34, 0x19,
	0x34, 0x55, 0x7b, 0x63, 0x19, 0x62, 0x82, 0xa5, 0xd4, 0x89, 0x62, 0x93, 0xd1, 0xef, 0xb0, 0x77,
	0xfb, 0x08, 0xfb, 0x00, 0xfb, 0x8a, 0x53, 0x9c, 0x04, 0x42, 0xd3, 0x76, 0x54, 0xba, 0x6f, 0x90,
	0xcf, 0x83, 0x7f, 0x39, 0x3e, 0xf9, 0x1f, 0x13, 0x68, 0xb9, 0xdc, 0xa1, 0xca, 0xeb, 0x85, 0xbe,
	0xdf, 0x93, 0x41, 0x38, 0xd2, 0x3f, 0x5d, 0x3f, 0xf0, 0x94, 0x87, 0xce, 0xe2, 0x58, 0x37, 0xf4,
	0xfd, 0x6e, 0xe4, 0xbe, 0xf8, 0xab, 0x0a, 0x95, 0x5b, 0x6f, 0x41, 0xdd, 0xe9, 0x64, 0x8c, 0xce,
	0xa1, 0x24, 0xb9, 0x6d, 0x16, 0xdb, 0xc5, 0x8e, 0x81, 0xa3, 0x25, 0xea, 0xc2, 0x47, 0x5c, 0x48,
	0x45, 0x5d, 0x97, 0x2a, 0xee, 0x09, 0x12, 0x06, 0x4b, 0xc2, 0x6d, 0xf3, 0xa8, 0x5d, 0xec, 0xd4,
	0x71, 0x33, 0x1b, 0xfa, 0x3d, 0x58, 0x4e, 0x6c, 0xf4, 0x0b, 0x34, 0xe7, 0x54, 0x32, 0xc2, 0x84,
	0x4d, 0x96, 0x6b, 0xb1, 0x88, 0x22, 0x66, 0xa9, 0x5d, 0xec, 0x54, 0xfb, 0x5f, 0x74, 0x1f, 0x3d,
	0xbb, 0x9b, 0x3e, 0xb7, 0x7b, 0x23, 0x6c, 0xab, 0x80, 0xcf, 0xa2, 0x9d, 0x37, 0xc2, 0x7e, 0x9f,
	0xec, 0x43, 0xef, 0xa1, 0x91, 0xe5, 0x90, 0x8d, 0x79, 0xac, 0x49, 0x5f, 0xbe, 0x48, 0xba, 0xb3,
	0x0a, 0xb8, 0xc6, 0x76, 0x98, 0xbb, 0x1c, 0x47, 0x99, 0x27, 0x07, 0x70, 0x66, 0x8f, 0x38, 0x33,
	0x74, 0x0b, 0xe7, 0x7b, 0x1c, 0x7b, 0xd3, 0x37, 0xcb, 0x9a, 0xd4, 0x7e, 0x91, 0x34, 0xbe, 0xeb,
	0x5b, 0x05, 0xdc, 0xc8, 0xb0, 0xc6, 0x9b, 0xfe, 0x13, 0xb4, 0xa1, 0x79, 0x7a, 0x10, 0x6d, 0x98,
	0xa3, 0x0d, 0x9f, 0xa0, 0x8d, 0xcc, 0xca, 0x41, 0xb4, 0x51, 0x8e, 0x36, 0xca, 0xd3, 0xd4, 0xd0,
	0x34, 0x0e, 0xa1, 0xcd, 0x72, 0xb5, 0xa9, 0x27, 0x6a, 0x53, 0x23, 0x13, 0x0e, 0xa2, 0xe5, 0x6a,
	0x53, 0x23, 0x34, 0x81, 0xb3, 0x3d, 0x1a, 0xb5, 0xcd, 0xaa, 0x86, 0xbd, 0x79, 0x11, 0x76, 0x35,
	0xb6, 0x0a, 0xb8, 0x9e, 0x61, 0x5d, 0xd9, 0xad, 0xcf, 0xa0, 0x74, 0x23, 0xec, 0x48, 0xf6, 0xbe,
	0xf4, 0xb5, 0xec, 0x2b, 0x38, 0x5a, 0xb6, 0xe6, 0x70, 0x1c, 0x29, 0x29, 0x1f, 0x41, 0x97, 0x80,
	0xbc, 0xb5, 0x72, 0x3c, 0x2e, 0x1c, 0xc2, 0x85, 0x62, 0xc1, 0x92, 0x2e, 0x98, 0x9e, 0x07, 0x03,
	0x37, 0xd3, 0xc8, 0x24, 0x0d, 0xa0, 0xcf, 0xa1, 0x22, 0xd8, 0x46, 0x91, 0x95, 0xe7, 0xeb, 0x31,
	0x30, 0xf0, 0x69, 0x64, 0x5b, 0x9e, 0xdf, 0xea, 0xe9, 0x67, 0xcc, 0x9e, 0x78, 0xc6, 0x27, 0x50,
	0xde, 0x9b, 0xb3, 0x93, 0x30, 0x9a, 0xad, 0x16, 0x86, 0x72, 0x2c, 0xa6, 0x88, 0x1a, 0xba, 0x54,
	0x10, 0x45, 0x1d, 0xbd, 0xaf, 0x8e, 0x4f, 0x23, 0x7b, 0x46, 0x9d, 0x57, 0xd6, 0xb7, 0x65, 0x0e,
	0x9f, 0xd9, 0x58, 0x3c, 0xe4, 0x60, 0x47, 0xfb, 0x07, 0x4b, 0x99, 0xa3, 0x0f, 0xc8, 0x7c, 0x13,
	0x33, 0x67, 0xc3, 0x4c, 0x73, 0x8a, 0xd9, 0xe6, 0xa4, 0x09, 0xa3, 0xe7, 0x12, 0xfe, 0x2e, 0xc2,
	0x89, 0x96, 0xc1, 0x6b, 0x5f, 0xe1, 0x25, 0x20, 0x2e, 0x16, 0xde, 0xfd, 0x7e, 0x7a, 0xfc, 0x32,
	0x9b, 0x69, 0x64, 0x97, 0xfe, 0x2d, 0x20, 0x77, 0x40, 0x24, 0x0b, 0x42, 0xbe, 0x60, 0x84, 0xda,
	0x76, 0xc0, 0xa4, 0xd4, 0x17, 0x97, 0x81, 0xcf, 0xdd, 0xc1, 0x34, 0x0e, 0x5c, 0xc5, 0xfe, 0xeb,
	0x06, 0xd4, 0xb2, 0x62, 0xbe, 0xf8, 0xf7, 0x08, 0xca, 0xbf, 0x79, 0x2e, 0x5f, 0x3c, 0x20, 0x04,
	0xc7, 0xf3, 0xdd, 0x6d, 0xac, 0xd7, 0xaf, 0xbe, 0x8e, 0xbf, 0x81, 0xa6, 0x0c, 0x56, 0x84, 0x89,
	0x05, 0xf5, 0xe5, 0x3a, 0x8e, 0xe8, 0xd2, 0x2b, 0xf8, 0x5c, 0x06, 0xab, 0x9b, 0xac, 0x1f, 0x7d,
	0x0d, 0x67, 0xd2, 0x0f, 0xe8, 0x03, 0x99, 0xb3, 0x15, 0x0d, 0xb9, 0xb7, 0x0e, 0x74, 0xd9, 0x15,
	0xdc, 0xd0, 0xee, 0xeb, 0xd4, 0x8b, 0x2c, 0xa8, 0x4b, 0xe6, 0xdc, 0x33, 0xa1, 0x88, 0xcb, 0xa5,
	0x92, 0xe6, 0x49, 0xbb, 0xd4, 0xa9, 0xf6, 0xdf, 0xe6, 0xe6, 0x2f, 0x3e, 0x49, 0x77, 0x1a, 0x27,
	0xdf, 0x72, 0xa9, 0x70, 0x4d, 0xee, 0x0c, 0xd9, 0xba, 0x82, 0x6a, 0x26, 0x88, 0x3e, 0x85, 0xf2,
	0x9f, 0x8c, 0x3b, 0x2b, 0x95, 0xbc, 0xba, 0xc4, 0x42, 0x2d, 0xa8, 0x24, 0xdb, 0xa4, 0x79, 0xd4,
	0x2e, 0x75, 0x0c, 0xbc, 0xb5, 0x2f, 0xfe, 0x29, 0x41, 0x65, 0xaa, 0x18, 0x0b, 0xb8, 0x70, 0xa2,
	0x9e, 0x09, 0x7a, 0xcf, 0xf4, 0xfd, 0x6e, 0x60, 0xbd, 0x46, 0x5f, 0x41, 0xd5, 0xd7, 0x75, 0x90,
	0x5d, 0x3b, 0xad, 0x02, 0x86, 0xd8, 0x79, 0x1d, 0xb5, 0xf5, 0x2d, 0xd4, 0x92, 0x14, 0x2e, 0x6c,
	0xb6, 0x89, 0xfb, 0x69, 0x15, 0x70, 0xb2, 0x71, 0x12, 0x39, 0xd1, 0x18, 0xc0, 0xed, 0x13, 0x15,
	0xd0, 0xe5, 0x92, 0x2f, 0x92, 0xff, 0xb4, 0xfc, 0x91, 0xd3, 0x52, 0xba, 0xb7, 0xfd, 0x59, 0x9c,
	0x6a, 0x15, 0xb1, 0xe1, 0xa6, 0x86, 0xa6, 0x0c, 0xb6, 0x94, 0xe3, 0xff, 0xa5, 0x0c, 0xb2, 0x94,
	0xd4, 0x68, 0xf5, 0xc1, 0xd8, 0xf2, 0xd1, 0x3b, 0x68, 0x6c, 0x75, 0x49, 0xf4, 0xf1, 0x63, 0xc9,
	0xd4, 0xb7, 0xde, 0x5f, 0xe9, 0x3d, 0x6b, 0xcd, 0xc1, 0xd8, 0xd2, 0x9e, 0x13, 0x52, 0xf1, 0x39,
	0x21, 0xbd, 0x83, 0x86, 0x1f, 0xb0, 0x25, 0xdf, 0x6c, 0x15, 0x1d, 0xcf, 0x4b, 0x3d, 0xf6, 0xa6,
	0x72, 0xae, 0x41, 0xd2, 0x56, 0x12, 0xb0, 0xe5, 0xb5, 0x01, 0xa7, 0xc9, 0x41, 0x2f, 0x7e, 0x02,
	0x98, 0xe2, 0x70, 0xf4, 0xb3, 0xeb, 0xcd, 0xa9, 0x8b, 0xbe, 0x83, 0x8f, 0xb5, 0x24, 0x89, 0xf4,
	0xd6, 0x41, 0x66, 0x4a, 0xe2, 0xba, 0x91, 0x8e, 0x4d, 0x75, 0x28, 0x05, 0xff, 0xf0, 0xc7, 0xf7,
	0x8e, 0x97, 0xb6, 0x89, 0xeb, 0x6f, 0x9b, 0x4b, 0xea, 0x30, 0xa1, 0x7a, 0x61, 0xbf, 0xa7, 0x3f,
	0x6d, 0x7a, 0x8f, 0xbe, 0x7a, 0x7e, 0x0c, 0x7d, 0x9f, 0x44, 0x8b, 0x79, 0x59, 0xc7, 0x07, 0xff,
	0x05, 0x00, 0x00, 0xff, 0xff, 0x63, 0x40, 0xe7, 0x66, 0x18, 0x09, 0x00, 0x00,
}
