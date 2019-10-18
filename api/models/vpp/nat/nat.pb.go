// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/models/vpp/nat/nat.proto

package vpp_nat

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

type DNat44_Protocol int32

const (
	DNat44_TCP  DNat44_Protocol = 0
	DNat44_UDP  DNat44_Protocol = 1
	DNat44_ICMP DNat44_Protocol = 2
)

var DNat44_Protocol_name = map[int32]string{
	0: "TCP",
	1: "UDP",
	2: "ICMP",
}

var DNat44_Protocol_value = map[string]int32{
	"TCP":  0,
	"UDP":  1,
	"ICMP": 2,
}

func (x DNat44_Protocol) String() string {
	return proto.EnumName(DNat44_Protocol_name, int32(x))
}

func (DNat44_Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{1, 0}
}

type DNat44_StaticMapping_TwiceNatMode int32

const (
	DNat44_StaticMapping_DISABLED DNat44_StaticMapping_TwiceNatMode = 0
	DNat44_StaticMapping_ENABLED  DNat44_StaticMapping_TwiceNatMode = 1
	DNat44_StaticMapping_SELF     DNat44_StaticMapping_TwiceNatMode = 2
)

var DNat44_StaticMapping_TwiceNatMode_name = map[int32]string{
	0: "DISABLED",
	1: "ENABLED",
	2: "SELF",
}

var DNat44_StaticMapping_TwiceNatMode_value = map[string]int32{
	"DISABLED": 0,
	"ENABLED":  1,
	"SELF":     2,
}

func (x DNat44_StaticMapping_TwiceNatMode) String() string {
	return proto.EnumName(DNat44_StaticMapping_TwiceNatMode_name, int32(x))
}

func (DNat44_StaticMapping_TwiceNatMode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{1, 0, 0}
}

type Nat44Global struct {
	Forwarding           bool                     `protobuf:"varint,1,opt,name=forwarding,proto3" json:"forwarding,omitempty"`
	NatInterfaces        []*Nat44Global_Interface `protobuf:"bytes,2,rep,name=nat_interfaces,json=natInterfaces,proto3" json:"nat_interfaces,omitempty"`
	AddressPool          []*Nat44Global_Address   `protobuf:"bytes,3,rep,name=address_pool,json=addressPool,proto3" json:"address_pool,omitempty"`
	VirtualReassembly    *VirtualReassembly       `protobuf:"bytes,4,opt,name=virtual_reassembly,json=virtualReassembly,proto3" json:"virtual_reassembly,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *Nat44Global) Reset()         { *m = Nat44Global{} }
func (m *Nat44Global) String() string { return proto.CompactTextString(m) }
func (*Nat44Global) ProtoMessage()    {}
func (*Nat44Global) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{0}
}

func (m *Nat44Global) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Nat44Global.Unmarshal(m, b)
}
func (m *Nat44Global) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Nat44Global.Marshal(b, m, deterministic)
}
func (m *Nat44Global) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Nat44Global.Merge(m, src)
}
func (m *Nat44Global) XXX_Size() int {
	return xxx_messageInfo_Nat44Global.Size(m)
}
func (m *Nat44Global) XXX_DiscardUnknown() {
	xxx_messageInfo_Nat44Global.DiscardUnknown(m)
}

var xxx_messageInfo_Nat44Global proto.InternalMessageInfo

func (m *Nat44Global) GetForwarding() bool {
	if m != nil {
		return m.Forwarding
	}
	return false
}

func (m *Nat44Global) GetNatInterfaces() []*Nat44Global_Interface {
	if m != nil {
		return m.NatInterfaces
	}
	return nil
}

func (m *Nat44Global) GetAddressPool() []*Nat44Global_Address {
	if m != nil {
		return m.AddressPool
	}
	return nil
}

func (m *Nat44Global) GetVirtualReassembly() *VirtualReassembly {
	if m != nil {
		return m.VirtualReassembly
	}
	return nil
}

type Nat44Global_Interface struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	IsInside             bool     `protobuf:"varint,2,opt,name=is_inside,json=isInside,proto3" json:"is_inside,omitempty"`
	OutputFeature        bool     `protobuf:"varint,3,opt,name=output_feature,json=outputFeature,proto3" json:"output_feature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Nat44Global_Interface) Reset()         { *m = Nat44Global_Interface{} }
func (m *Nat44Global_Interface) String() string { return proto.CompactTextString(m) }
func (*Nat44Global_Interface) ProtoMessage()    {}
func (*Nat44Global_Interface) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{0, 0}
}

func (m *Nat44Global_Interface) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Nat44Global_Interface.Unmarshal(m, b)
}
func (m *Nat44Global_Interface) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Nat44Global_Interface.Marshal(b, m, deterministic)
}
func (m *Nat44Global_Interface) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Nat44Global_Interface.Merge(m, src)
}
func (m *Nat44Global_Interface) XXX_Size() int {
	return xxx_messageInfo_Nat44Global_Interface.Size(m)
}
func (m *Nat44Global_Interface) XXX_DiscardUnknown() {
	xxx_messageInfo_Nat44Global_Interface.DiscardUnknown(m)
}

var xxx_messageInfo_Nat44Global_Interface proto.InternalMessageInfo

func (m *Nat44Global_Interface) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Nat44Global_Interface) GetIsInside() bool {
	if m != nil {
		return m.IsInside
	}
	return false
}

func (m *Nat44Global_Interface) GetOutputFeature() bool {
	if m != nil {
		return m.OutputFeature
	}
	return false
}

type Nat44Global_Address struct {
	Address              string   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	VrfId                uint32   `protobuf:"varint,2,opt,name=vrf_id,json=vrfId,proto3" json:"vrf_id,omitempty"`
	TwiceNat             bool     `protobuf:"varint,3,opt,name=twice_nat,json=twiceNat,proto3" json:"twice_nat,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Nat44Global_Address) Reset()         { *m = Nat44Global_Address{} }
func (m *Nat44Global_Address) String() string { return proto.CompactTextString(m) }
func (*Nat44Global_Address) ProtoMessage()    {}
func (*Nat44Global_Address) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{0, 1}
}

func (m *Nat44Global_Address) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Nat44Global_Address.Unmarshal(m, b)
}
func (m *Nat44Global_Address) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Nat44Global_Address.Marshal(b, m, deterministic)
}
func (m *Nat44Global_Address) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Nat44Global_Address.Merge(m, src)
}
func (m *Nat44Global_Address) XXX_Size() int {
	return xxx_messageInfo_Nat44Global_Address.Size(m)
}
func (m *Nat44Global_Address) XXX_DiscardUnknown() {
	xxx_messageInfo_Nat44Global_Address.DiscardUnknown(m)
}

var xxx_messageInfo_Nat44Global_Address proto.InternalMessageInfo

func (m *Nat44Global_Address) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Nat44Global_Address) GetVrfId() uint32 {
	if m != nil {
		return m.VrfId
	}
	return 0
}

func (m *Nat44Global_Address) GetTwiceNat() bool {
	if m != nil {
		return m.TwiceNat
	}
	return false
}

type DNat44 struct {
	Label                string                    `protobuf:"bytes,1,opt,name=label,proto3" json:"label,omitempty"`
	StMappings           []*DNat44_StaticMapping   `protobuf:"bytes,2,rep,name=st_mappings,json=stMappings,proto3" json:"st_mappings,omitempty"`
	IdMappings           []*DNat44_IdentityMapping `protobuf:"bytes,3,rep,name=id_mappings,json=idMappings,proto3" json:"id_mappings,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *DNat44) Reset()         { *m = DNat44{} }
func (m *DNat44) String() string { return proto.CompactTextString(m) }
func (*DNat44) ProtoMessage()    {}
func (*DNat44) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{1}
}

func (m *DNat44) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DNat44.Unmarshal(m, b)
}
func (m *DNat44) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DNat44.Marshal(b, m, deterministic)
}
func (m *DNat44) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DNat44.Merge(m, src)
}
func (m *DNat44) XXX_Size() int {
	return xxx_messageInfo_DNat44.Size(m)
}
func (m *DNat44) XXX_DiscardUnknown() {
	xxx_messageInfo_DNat44.DiscardUnknown(m)
}

var xxx_messageInfo_DNat44 proto.InternalMessageInfo

func (m *DNat44) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *DNat44) GetStMappings() []*DNat44_StaticMapping {
	if m != nil {
		return m.StMappings
	}
	return nil
}

func (m *DNat44) GetIdMappings() []*DNat44_IdentityMapping {
	if m != nil {
		return m.IdMappings
	}
	return nil
}

type DNat44_StaticMapping struct {
	ExternalInterface    string                            `protobuf:"bytes,1,opt,name=external_interface,json=externalInterface,proto3" json:"external_interface,omitempty"`
	ExternalIp           string                            `protobuf:"bytes,2,opt,name=external_ip,json=externalIp,proto3" json:"external_ip,omitempty"`
	ExternalPort         uint32                            `protobuf:"varint,3,opt,name=external_port,json=externalPort,proto3" json:"external_port,omitempty"`
	LocalIps             []*DNat44_StaticMapping_LocalIP   `protobuf:"bytes,4,rep,name=local_ips,json=localIps,proto3" json:"local_ips,omitempty"`
	Protocol             DNat44_Protocol                   `protobuf:"varint,5,opt,name=protocol,proto3,enum=vpp.nat.DNat44_Protocol" json:"protocol,omitempty"`
	TwiceNat             DNat44_StaticMapping_TwiceNatMode `protobuf:"varint,6,opt,name=twice_nat,json=twiceNat,proto3,enum=vpp.nat.DNat44_StaticMapping_TwiceNatMode" json:"twice_nat,omitempty"`
	SessionAffinity      uint32                            `protobuf:"varint,7,opt,name=session_affinity,json=sessionAffinity,proto3" json:"session_affinity,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                          `json:"-"`
	XXX_unrecognized     []byte                            `json:"-"`
	XXX_sizecache        int32                             `json:"-"`
}

func (m *DNat44_StaticMapping) Reset()         { *m = DNat44_StaticMapping{} }
func (m *DNat44_StaticMapping) String() string { return proto.CompactTextString(m) }
func (*DNat44_StaticMapping) ProtoMessage()    {}
func (*DNat44_StaticMapping) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{1, 0}
}

func (m *DNat44_StaticMapping) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DNat44_StaticMapping.Unmarshal(m, b)
}
func (m *DNat44_StaticMapping) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DNat44_StaticMapping.Marshal(b, m, deterministic)
}
func (m *DNat44_StaticMapping) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DNat44_StaticMapping.Merge(m, src)
}
func (m *DNat44_StaticMapping) XXX_Size() int {
	return xxx_messageInfo_DNat44_StaticMapping.Size(m)
}
func (m *DNat44_StaticMapping) XXX_DiscardUnknown() {
	xxx_messageInfo_DNat44_StaticMapping.DiscardUnknown(m)
}

var xxx_messageInfo_DNat44_StaticMapping proto.InternalMessageInfo

func (m *DNat44_StaticMapping) GetExternalInterface() string {
	if m != nil {
		return m.ExternalInterface
	}
	return ""
}

func (m *DNat44_StaticMapping) GetExternalIp() string {
	if m != nil {
		return m.ExternalIp
	}
	return ""
}

func (m *DNat44_StaticMapping) GetExternalPort() uint32 {
	if m != nil {
		return m.ExternalPort
	}
	return 0
}

func (m *DNat44_StaticMapping) GetLocalIps() []*DNat44_StaticMapping_LocalIP {
	if m != nil {
		return m.LocalIps
	}
	return nil
}

func (m *DNat44_StaticMapping) GetProtocol() DNat44_Protocol {
	if m != nil {
		return m.Protocol
	}
	return DNat44_TCP
}

func (m *DNat44_StaticMapping) GetTwiceNat() DNat44_StaticMapping_TwiceNatMode {
	if m != nil {
		return m.TwiceNat
	}
	return DNat44_StaticMapping_DISABLED
}

func (m *DNat44_StaticMapping) GetSessionAffinity() uint32 {
	if m != nil {
		return m.SessionAffinity
	}
	return 0
}

type DNat44_StaticMapping_LocalIP struct {
	VrfId                uint32   `protobuf:"varint,1,opt,name=vrf_id,json=vrfId,proto3" json:"vrf_id,omitempty"`
	LocalIp              string   `protobuf:"bytes,2,opt,name=local_ip,json=localIp,proto3" json:"local_ip,omitempty"`
	LocalPort            uint32   `protobuf:"varint,3,opt,name=local_port,json=localPort,proto3" json:"local_port,omitempty"`
	Probability          uint32   `protobuf:"varint,4,opt,name=probability,proto3" json:"probability,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DNat44_StaticMapping_LocalIP) Reset()         { *m = DNat44_StaticMapping_LocalIP{} }
func (m *DNat44_StaticMapping_LocalIP) String() string { return proto.CompactTextString(m) }
func (*DNat44_StaticMapping_LocalIP) ProtoMessage()    {}
func (*DNat44_StaticMapping_LocalIP) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{1, 0, 0}
}

func (m *DNat44_StaticMapping_LocalIP) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DNat44_StaticMapping_LocalIP.Unmarshal(m, b)
}
func (m *DNat44_StaticMapping_LocalIP) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DNat44_StaticMapping_LocalIP.Marshal(b, m, deterministic)
}
func (m *DNat44_StaticMapping_LocalIP) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DNat44_StaticMapping_LocalIP.Merge(m, src)
}
func (m *DNat44_StaticMapping_LocalIP) XXX_Size() int {
	return xxx_messageInfo_DNat44_StaticMapping_LocalIP.Size(m)
}
func (m *DNat44_StaticMapping_LocalIP) XXX_DiscardUnknown() {
	xxx_messageInfo_DNat44_StaticMapping_LocalIP.DiscardUnknown(m)
}

var xxx_messageInfo_DNat44_StaticMapping_LocalIP proto.InternalMessageInfo

func (m *DNat44_StaticMapping_LocalIP) GetVrfId() uint32 {
	if m != nil {
		return m.VrfId
	}
	return 0
}

func (m *DNat44_StaticMapping_LocalIP) GetLocalIp() string {
	if m != nil {
		return m.LocalIp
	}
	return ""
}

func (m *DNat44_StaticMapping_LocalIP) GetLocalPort() uint32 {
	if m != nil {
		return m.LocalPort
	}
	return 0
}

func (m *DNat44_StaticMapping_LocalIP) GetProbability() uint32 {
	if m != nil {
		return m.Probability
	}
	return 0
}

type DNat44_IdentityMapping struct {
	VrfId                uint32          `protobuf:"varint,1,opt,name=vrf_id,json=vrfId,proto3" json:"vrf_id,omitempty"`
	Interface            string          `protobuf:"bytes,2,opt,name=interface,proto3" json:"interface,omitempty"`
	IpAddress            string          `protobuf:"bytes,3,opt,name=ip_address,json=ipAddress,proto3" json:"ip_address,omitempty"`
	Port                 uint32          `protobuf:"varint,4,opt,name=port,proto3" json:"port,omitempty"`
	Protocol             DNat44_Protocol `protobuf:"varint,5,opt,name=protocol,proto3,enum=vpp.nat.DNat44_Protocol" json:"protocol,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *DNat44_IdentityMapping) Reset()         { *m = DNat44_IdentityMapping{} }
func (m *DNat44_IdentityMapping) String() string { return proto.CompactTextString(m) }
func (*DNat44_IdentityMapping) ProtoMessage()    {}
func (*DNat44_IdentityMapping) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{1, 1}
}

func (m *DNat44_IdentityMapping) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DNat44_IdentityMapping.Unmarshal(m, b)
}
func (m *DNat44_IdentityMapping) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DNat44_IdentityMapping.Marshal(b, m, deterministic)
}
func (m *DNat44_IdentityMapping) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DNat44_IdentityMapping.Merge(m, src)
}
func (m *DNat44_IdentityMapping) XXX_Size() int {
	return xxx_messageInfo_DNat44_IdentityMapping.Size(m)
}
func (m *DNat44_IdentityMapping) XXX_DiscardUnknown() {
	xxx_messageInfo_DNat44_IdentityMapping.DiscardUnknown(m)
}

var xxx_messageInfo_DNat44_IdentityMapping proto.InternalMessageInfo

func (m *DNat44_IdentityMapping) GetVrfId() uint32 {
	if m != nil {
		return m.VrfId
	}
	return 0
}

func (m *DNat44_IdentityMapping) GetInterface() string {
	if m != nil {
		return m.Interface
	}
	return ""
}

func (m *DNat44_IdentityMapping) GetIpAddress() string {
	if m != nil {
		return m.IpAddress
	}
	return ""
}

func (m *DNat44_IdentityMapping) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *DNat44_IdentityMapping) GetProtocol() DNat44_Protocol {
	if m != nil {
		return m.Protocol
	}
	return DNat44_TCP
}

type VirtualReassembly struct {
	Timeout              uint32   `protobuf:"varint,1,opt,name=timeout,proto3" json:"timeout,omitempty"`
	MaxReassemblies      uint32   `protobuf:"varint,2,opt,name=max_reassemblies,json=maxReassemblies,proto3" json:"max_reassemblies,omitempty"`
	MaxFragments         uint32   `protobuf:"varint,3,opt,name=max_fragments,json=maxFragments,proto3" json:"max_fragments,omitempty"`
	DropFragments        bool     `protobuf:"varint,4,opt,name=drop_fragments,json=dropFragments,proto3" json:"drop_fragments,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *VirtualReassembly) Reset()         { *m = VirtualReassembly{} }
func (m *VirtualReassembly) String() string { return proto.CompactTextString(m) }
func (*VirtualReassembly) ProtoMessage()    {}
func (*VirtualReassembly) Descriptor() ([]byte, []int) {
	return fileDescriptor_2e774b1f7a9dab01, []int{2}
}

func (m *VirtualReassembly) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VirtualReassembly.Unmarshal(m, b)
}
func (m *VirtualReassembly) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VirtualReassembly.Marshal(b, m, deterministic)
}
func (m *VirtualReassembly) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VirtualReassembly.Merge(m, src)
}
func (m *VirtualReassembly) XXX_Size() int {
	return xxx_messageInfo_VirtualReassembly.Size(m)
}
func (m *VirtualReassembly) XXX_DiscardUnknown() {
	xxx_messageInfo_VirtualReassembly.DiscardUnknown(m)
}

var xxx_messageInfo_VirtualReassembly proto.InternalMessageInfo

func (m *VirtualReassembly) GetTimeout() uint32 {
	if m != nil {
		return m.Timeout
	}
	return 0
}

func (m *VirtualReassembly) GetMaxReassemblies() uint32 {
	if m != nil {
		return m.MaxReassemblies
	}
	return 0
}

func (m *VirtualReassembly) GetMaxFragments() uint32 {
	if m != nil {
		return m.MaxFragments
	}
	return 0
}

func (m *VirtualReassembly) GetDropFragments() bool {
	if m != nil {
		return m.DropFragments
	}
	return false
}

func init() {
	proto.RegisterEnum("vpp.nat.DNat44_Protocol", DNat44_Protocol_name, DNat44_Protocol_value)
	proto.RegisterEnum("vpp.nat.DNat44_StaticMapping_TwiceNatMode", DNat44_StaticMapping_TwiceNatMode_name, DNat44_StaticMapping_TwiceNatMode_value)
	proto.RegisterType((*Nat44Global)(nil), "vpp.nat.Nat44Global")
	proto.RegisterType((*Nat44Global_Interface)(nil), "vpp.nat.Nat44Global.Interface")
	proto.RegisterType((*Nat44Global_Address)(nil), "vpp.nat.Nat44Global.Address")
	proto.RegisterType((*DNat44)(nil), "vpp.nat.DNat44")
	proto.RegisterType((*DNat44_StaticMapping)(nil), "vpp.nat.DNat44.StaticMapping")
	proto.RegisterType((*DNat44_StaticMapping_LocalIP)(nil), "vpp.nat.DNat44.StaticMapping.LocalIP")
	proto.RegisterType((*DNat44_IdentityMapping)(nil), "vpp.nat.DNat44.IdentityMapping")
	proto.RegisterType((*VirtualReassembly)(nil), "vpp.nat.VirtualReassembly")
}

func init() { proto.RegisterFile("api/models/vpp/nat/nat.proto", fileDescriptor_2e774b1f7a9dab01) }

var fileDescriptor_2e774b1f7a9dab01 = []byte{
	// 806 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x55, 0x5d, 0x6b, 0x23, 0x37,
	0x14, 0x5d, 0xc7, 0x4e, 0x66, 0x7c, 0x27, 0xce, 0x3a, 0xa2, 0x85, 0xa9, 0x9b, 0xdd, 0x35, 0x2e,
	0x5b, 0xd2, 0xc2, 0xda, 0x90, 0x0d, 0xa5, 0x50, 0x68, 0x9b, 0x6c, 0x92, 0x65, 0x20, 0x09, 0x66,
	0xb2, 0x6d, 0xa1, 0x2f, 0x83, 0xec, 0xd1, 0xb8, 0x82, 0x19, 0x49, 0x48, 0xb2, 0x37, 0x81, 0xfe,
	0x9a, 0xd2, 0x3e, 0xf7, 0x1f, 0xf4, 0xb7, 0x95, 0x91, 0x34, 0x1f, 0xf5, 0xb6, 0x79, 0xe8, 0x83,
	0x41, 0x3a, 0x3a, 0x3a, 0xba, 0xf7, 0xdc, 0x3b, 0xd7, 0x70, 0x84, 0x05, 0x9d, 0x15, 0x3c, 0x25,
	0xb9, 0x9a, 0x6d, 0x84, 0x98, 0x31, 0xac, 0xcb, 0xdf, 0x54, 0x48, 0xae, 0x39, 0xf2, 0x36, 0x42,
	0x4c, 0x19, 0xd6, 0x93, 0xbf, 0xba, 0x10, 0xdc, 0x62, 0x7d, 0x7a, 0xfa, 0x36, 0xe7, 0x0b, 0x9c,
	0xa3, 0xe7, 0x00, 0x19, 0x97, 0xef, 0xb1, 0x4c, 0x29, 0x5b, 0x85, 0x9d, 0x71, 0xe7, 0xd8, 0x8f,
	0x5b, 0x08, 0xba, 0x84, 0x03, 0x86, 0x75, 0x42, 0x99, 0x26, 0x32, 0xc3, 0x4b, 0xa2, 0xc2, 0x9d,
	0x71, 0xf7, 0x38, 0x38, 0x79, 0x3e, 0x75, 0x8a, 0xd3, 0x96, 0xda, 0x34, 0xaa, 0x68, 0xf1, 0x80,
	0x61, 0x5d, 0xef, 0x14, 0xfa, 0x0e, 0xf6, 0x71, 0x9a, 0x4a, 0xa2, 0x54, 0x22, 0x38, 0xcf, 0xc3,
	0xae, 0x11, 0x39, 0xfa, 0x57, 0x91, 0x33, 0x4b, 0x8c, 0x03, 0x77, 0x63, 0xce, 0x79, 0x8e, 0x22,
	0x40, 0x1b, 0x2a, 0xf5, 0x1a, 0xe7, 0x89, 0x24, 0x58, 0x29, 0x52, 0x2c, 0xf2, 0x87, 0xb0, 0x37,
	0xee, 0x1c, 0x07, 0x27, 0xa3, 0x5a, 0xe6, 0x47, 0x4b, 0x89, 0x6b, 0x46, 0x7c, 0xb8, 0xd9, 0x86,
	0x46, 0x4b, 0xe8, 0xd7, 0x91, 0x21, 0x04, 0x3d, 0x86, 0x0b, 0x62, 0x32, 0xef, 0xc7, 0x66, 0x8d,
	0x3e, 0x85, 0x3e, 0x55, 0x09, 0x65, 0x8a, 0xa6, 0x24, 0xdc, 0x31, 0x96, 0xf8, 0x54, 0x45, 0x66,
	0x8f, 0x5e, 0xc2, 0x01, 0x5f, 0x6b, 0xb1, 0xd6, 0x49, 0x46, 0xb0, 0x5e, 0x4b, 0x12, 0x76, 0x0d,
	0x63, 0x60, 0xd1, 0x2b, 0x0b, 0x8e, 0x7e, 0x02, 0xcf, 0xe5, 0x81, 0x42, 0xf0, 0x5c, 0x26, 0xee,
	0x95, 0x6a, 0x8b, 0x3e, 0x86, 0xbd, 0x8d, 0xcc, 0x12, 0x9a, 0x9a, 0x57, 0x06, 0xf1, 0xee, 0x46,
	0x66, 0x51, 0x5a, 0xbe, 0xaf, 0xdf, 0xd3, 0x25, 0x49, 0x18, 0xd6, 0x4e, 0xdd, 0x37, 0xc0, 0x2d,
	0xd6, 0x93, 0xdf, 0x3c, 0xd8, 0xbb, 0x30, 0x76, 0xa1, 0x8f, 0x60, 0x37, 0xc7, 0x0b, 0x92, 0x3b,
	0x59, 0xbb, 0x41, 0xdf, 0x42, 0xa0, 0x74, 0x52, 0x60, 0x21, 0x28, 0x5b, 0x55, 0xe5, 0x7a, 0x56,
	0x5b, 0x64, 0xef, 0x4e, 0xef, 0x34, 0xd6, 0x74, 0x79, 0x63, 0x59, 0x31, 0x28, 0xed, 0x96, 0x0a,
	0x7d, 0x0f, 0x01, 0x4d, 0x9b, 0xfb, 0xb6, 0x52, 0x2f, 0xb6, 0xef, 0x47, 0x29, 0x61, 0x9a, 0xea,
	0x87, 0x5a, 0x81, 0xa6, 0x95, 0xc2, 0xe8, 0x8f, 0x1e, 0x0c, 0xfe, 0xa1, 0x8f, 0x5e, 0x01, 0x22,
	0xf7, 0x9a, 0x48, 0x86, 0xf3, 0xa6, 0x95, 0x5c, 0xd8, 0x87, 0xd5, 0x49, 0x53, 0x94, 0x17, 0x10,
	0x34, 0x74, 0x61, 0xcc, 0xe9, 0xc7, 0x50, 0xf3, 0x04, 0xfa, 0x0c, 0x06, 0x35, 0x41, 0x70, 0x69,
	0x5d, 0x1a, 0xc4, 0xfb, 0x15, 0x38, 0xe7, 0x52, 0xa3, 0x73, 0xe8, 0xe7, 0x7c, 0x69, 0x24, 0x54,
	0xd8, 0x33, 0x69, 0xbc, 0x7c, 0xd4, 0x86, 0xe9, 0x75, 0x49, 0x8f, 0xe6, 0xb1, 0x6f, 0xee, 0x45,
	0x42, 0xa1, 0x53, 0xf0, 0xcd, 0x07, 0xb4, 0xe4, 0x79, 0xb8, 0x3b, 0xee, 0x1c, 0x1f, 0x9c, 0x84,
	0xdb, 0x12, 0x73, 0x77, 0x1e, 0xd7, 0x4c, 0xf4, 0xb6, 0x5d, 0xc0, 0x3d, 0x73, 0xed, 0xcb, 0xc7,
	0x5f, 0x7e, 0xe7, 0xca, 0x7b, 0xc3, 0x53, 0xd2, 0x14, 0x1b, 0x7d, 0x01, 0x43, 0x45, 0x94, 0xa2,
	0x9c, 0x25, 0x38, 0xcb, 0x28, 0xa3, 0xfa, 0x21, 0xf4, 0x4c, 0xaa, 0x4f, 0x1d, 0x7e, 0xe6, 0xe0,
	0xd1, 0xaf, 0xe0, 0xb9, 0xf0, 0x5b, 0x6d, 0xd5, 0x69, 0xb7, 0xd5, 0x27, 0xe0, 0x57, 0x7e, 0x38,
	0x4b, 0x3d, 0x97, 0x27, 0x7a, 0x06, 0x60, 0x8f, 0x5a, 0x66, 0x5a, 0xf3, 0x8c, 0x93, 0x63, 0x08,
	0x84, 0xe4, 0x0b, 0xbc, 0xa0, 0x79, 0x19, 0x41, 0xcf, 0x9c, 0xb7, 0xa1, 0xc9, 0x6b, 0xd8, 0x6f,
	0xa7, 0x80, 0xf6, 0xc1, 0xbf, 0x88, 0xee, 0xce, 0xce, 0xaf, 0x2f, 0x2f, 0x86, 0x4f, 0x50, 0x00,
	0xde, 0xe5, 0xad, 0xdd, 0x74, 0x90, 0x0f, 0xbd, 0xbb, 0xcb, 0xeb, 0xab, 0xe1, 0xce, 0xe8, 0xcf,
	0x0e, 0x3c, 0xdd, 0xea, 0xa3, 0xff, 0x8a, 0xfd, 0x08, 0xfa, 0x4d, 0xdf, 0xd8, 0xe0, 0x1b, 0xa0,
	0x0c, 0x9f, 0x8a, 0xa4, 0xfa, 0xc8, 0xba, 0xee, 0x58, 0x54, 0x1f, 0x20, 0x82, 0x9e, 0xc9, 0xcb,
	0xc6, 0x6d, 0xd6, 0xff, 0xaf, 0xb0, 0x93, 0xcf, 0xc1, 0xaf, 0x50, 0xe4, 0x41, 0xf7, 0xdd, 0x9b,
	0xf9, 0xf0, 0x49, 0xb9, 0xf8, 0xe1, 0x62, 0x6e, 0x33, 0x8b, 0xde, 0xdc, 0xcc, 0x87, 0x3b, 0x93,
	0xdf, 0x3b, 0x70, 0xf8, 0xc1, 0x2c, 0x2a, 0x07, 0x81, 0xa6, 0x05, 0xe1, 0x6b, 0xed, 0x92, 0xab,
	0xb6, 0x65, 0x9d, 0x0b, 0x7c, 0xdf, 0x4c, 0x36, 0x6a, 0xe6, 0xac, 0xa9, 0x73, 0x81, 0xef, 0xe3,
	0x16, 0x5c, 0xb6, 0x7e, 0x49, 0xcd, 0x24, 0x5e, 0x15, 0x84, 0x69, 0x55, 0xb5, 0x7e, 0x81, 0xef,
	0xaf, 0x2a, 0xac, 0x1c, 0x52, 0xa9, 0xe4, 0xa2, 0xc5, 0xea, 0xd9, 0x21, 0x55, 0xa2, 0x35, 0xed,
	0xfc, 0xeb, 0x9f, 0xbf, 0x5a, 0x51, 0xfd, 0xcb, 0x7a, 0x31, 0x5d, 0xf2, 0x62, 0x96, 0xd3, 0x15,
	0xd6, 0xbc, 0xfc, 0xf3, 0x78, 0x85, 0x57, 0x84, 0xe9, 0xd9, 0x87, 0xff, 0x28, 0xdf, 0x6c, 0x84,
	0x28, 0x9b, 0x7a, 0xb1, 0x67, 0x2c, 0x79, 0xfd, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x76, 0x16,
	0x30, 0xe6, 0x76, 0x06, 0x00, 0x00,
}
