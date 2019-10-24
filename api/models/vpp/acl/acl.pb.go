// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/api/models/vpp/acl/acl.proto

package vpp_acl

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

type ACL_Rule_Action int32

const (
	ACL_Rule_DENY    ACL_Rule_Action = 0
	ACL_Rule_PERMIT  ACL_Rule_Action = 1
	ACL_Rule_REFLECT ACL_Rule_Action = 2
)

var ACL_Rule_Action_name = map[int32]string{
	0: "DENY",
	1: "PERMIT",
	2: "REFLECT",
}

var ACL_Rule_Action_value = map[string]int32{
	"DENY":    0,
	"PERMIT":  1,
	"REFLECT": 2,
}

func (x ACL_Rule_Action) String() string {
	return proto.EnumName(ACL_Rule_Action_name, int32(x))
}

func (ACL_Rule_Action) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0, 0}
}

// ACL defines Access Control List.
type ACL struct {
	// The name of an access list. A device MAY restrict the length
	// and value of this name, possibly spaces and special
	// characters are not allowed.
	Name                 string          `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Rules                []*ACL_Rule     `protobuf:"bytes,2,rep,name=rules,proto3" json:"rules,omitempty"`
	Interfaces           *ACL_Interfaces `protobuf:"bytes,3,opt,name=interfaces,proto3" json:"interfaces,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *ACL) Reset()         { *m = ACL{} }
func (m *ACL) String() string { return proto.CompactTextString(m) }
func (*ACL) ProtoMessage()    {}
func (*ACL) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0}
}

func (m *ACL) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL.Unmarshal(m, b)
}
func (m *ACL) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL.Marshal(b, m, deterministic)
}
func (m *ACL) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL.Merge(m, src)
}
func (m *ACL) XXX_Size() int {
	return xxx_messageInfo_ACL.Size(m)
}
func (m *ACL) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL.DiscardUnknown(m)
}

var xxx_messageInfo_ACL proto.InternalMessageInfo

func (m *ACL) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ACL) GetRules() []*ACL_Rule {
	if m != nil {
		return m.Rules
	}
	return nil
}

func (m *ACL) GetInterfaces() *ACL_Interfaces {
	if m != nil {
		return m.Interfaces
	}
	return nil
}

// List of access list entries (Rules). Each Access Control Rule has
// a list of match criteria and a list of actions.
// Access List entry that can define:
// - IPv4/IPv6 src ip prefix
// - src MAC address mask
// - src MAC address value
// - can be used only for static ACLs.
type ACL_Rule struct {
	Action               ACL_Rule_Action     `protobuf:"varint,1,opt,name=action,proto3,enum=vpp.acl.ACL_Rule_Action" json:"action,omitempty"`
	IpRule               *ACL_Rule_IpRule    `protobuf:"bytes,2,opt,name=ip_rule,json=ipRule,proto3" json:"ip_rule,omitempty"`
	MacipRule            *ACL_Rule_MacIpRule `protobuf:"bytes,3,opt,name=macip_rule,json=macipRule,proto3" json:"macip_rule,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *ACL_Rule) Reset()         { *m = ACL_Rule{} }
func (m *ACL_Rule) String() string { return proto.CompactTextString(m) }
func (*ACL_Rule) ProtoMessage()    {}
func (*ACL_Rule) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0}
}

func (m *ACL_Rule) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Rule.Unmarshal(m, b)
}
func (m *ACL_Rule) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Rule.Marshal(b, m, deterministic)
}
func (m *ACL_Rule) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Rule.Merge(m, src)
}
func (m *ACL_Rule) XXX_Size() int {
	return xxx_messageInfo_ACL_Rule.Size(m)
}
func (m *ACL_Rule) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Rule.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Rule proto.InternalMessageInfo

func (m *ACL_Rule) GetAction() ACL_Rule_Action {
	if m != nil {
		return m.Action
	}
	return ACL_Rule_DENY
}

func (m *ACL_Rule) GetIpRule() *ACL_Rule_IpRule {
	if m != nil {
		return m.IpRule
	}
	return nil
}

func (m *ACL_Rule) GetMacipRule() *ACL_Rule_MacIpRule {
	if m != nil {
		return m.MacipRule
	}
	return nil
}

type ACL_Rule_IpRule struct {
	Ip                   *ACL_Rule_IpRule_Ip   `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Icmp                 *ACL_Rule_IpRule_Icmp `protobuf:"bytes,2,opt,name=icmp,proto3" json:"icmp,omitempty"`
	Tcp                  *ACL_Rule_IpRule_Tcp  `protobuf:"bytes,3,opt,name=tcp,proto3" json:"tcp,omitempty"`
	Udp                  *ACL_Rule_IpRule_Udp  `protobuf:"bytes,4,opt,name=udp,proto3" json:"udp,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *ACL_Rule_IpRule) Reset()         { *m = ACL_Rule_IpRule{} }
func (m *ACL_Rule_IpRule) String() string { return proto.CompactTextString(m) }
func (*ACL_Rule_IpRule) ProtoMessage()    {}
func (*ACL_Rule_IpRule) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0, 0}
}

func (m *ACL_Rule_IpRule) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Rule_IpRule.Unmarshal(m, b)
}
func (m *ACL_Rule_IpRule) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Rule_IpRule.Marshal(b, m, deterministic)
}
func (m *ACL_Rule_IpRule) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Rule_IpRule.Merge(m, src)
}
func (m *ACL_Rule_IpRule) XXX_Size() int {
	return xxx_messageInfo_ACL_Rule_IpRule.Size(m)
}
func (m *ACL_Rule_IpRule) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Rule_IpRule.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Rule_IpRule proto.InternalMessageInfo

func (m *ACL_Rule_IpRule) GetIp() *ACL_Rule_IpRule_Ip {
	if m != nil {
		return m.Ip
	}
	return nil
}

func (m *ACL_Rule_IpRule) GetIcmp() *ACL_Rule_IpRule_Icmp {
	if m != nil {
		return m.Icmp
	}
	return nil
}

func (m *ACL_Rule_IpRule) GetTcp() *ACL_Rule_IpRule_Tcp {
	if m != nil {
		return m.Tcp
	}
	return nil
}

func (m *ACL_Rule_IpRule) GetUdp() *ACL_Rule_IpRule_Udp {
	if m != nil {
		return m.Udp
	}
	return nil
}

// IP  used in this Access List Entry.
type ACL_Rule_IpRule_Ip struct {
	// Destination IPv4/IPv6 network address (<ip>/<network>)
	DestinationNetwork string `protobuf:"bytes,1,opt,name=destination_network,json=destinationNetwork,proto3" json:"destination_network,omitempty"`
	// Destination IPv4/IPv6 network address (<ip>/<network>)
	SourceNetwork        string   `protobuf:"bytes,2,opt,name=source_network,json=sourceNetwork,proto3" json:"source_network,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ACL_Rule_IpRule_Ip) Reset()         { *m = ACL_Rule_IpRule_Ip{} }
func (m *ACL_Rule_IpRule_Ip) String() string { return proto.CompactTextString(m) }
func (*ACL_Rule_IpRule_Ip) ProtoMessage()    {}
func (*ACL_Rule_IpRule_Ip) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0, 0, 0}
}

func (m *ACL_Rule_IpRule_Ip) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Rule_IpRule_Ip.Unmarshal(m, b)
}
func (m *ACL_Rule_IpRule_Ip) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Rule_IpRule_Ip.Marshal(b, m, deterministic)
}
func (m *ACL_Rule_IpRule_Ip) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Rule_IpRule_Ip.Merge(m, src)
}
func (m *ACL_Rule_IpRule_Ip) XXX_Size() int {
	return xxx_messageInfo_ACL_Rule_IpRule_Ip.Size(m)
}
func (m *ACL_Rule_IpRule_Ip) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Rule_IpRule_Ip.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Rule_IpRule_Ip proto.InternalMessageInfo

func (m *ACL_Rule_IpRule_Ip) GetDestinationNetwork() string {
	if m != nil {
		return m.DestinationNetwork
	}
	return ""
}

func (m *ACL_Rule_IpRule_Ip) GetSourceNetwork() string {
	if m != nil {
		return m.SourceNetwork
	}
	return ""
}

type ACL_Rule_IpRule_Icmp struct {
	// ICMPv6 flag, if false ICMPv4 will be used
	Icmpv6 bool `protobuf:"varint,1,opt,name=icmpv6,proto3" json:"icmpv6,omitempty"`
	// Inclusive range representing icmp codes to be used.
	IcmpCodeRange        *ACL_Rule_IpRule_Icmp_Range `protobuf:"bytes,2,opt,name=icmp_code_range,json=icmpCodeRange,proto3" json:"icmp_code_range,omitempty"`
	IcmpTypeRange        *ACL_Rule_IpRule_Icmp_Range `protobuf:"bytes,3,opt,name=icmp_type_range,json=icmpTypeRange,proto3" json:"icmp_type_range,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                    `json:"-"`
	XXX_unrecognized     []byte                      `json:"-"`
	XXX_sizecache        int32                       `json:"-"`
}

func (m *ACL_Rule_IpRule_Icmp) Reset()         { *m = ACL_Rule_IpRule_Icmp{} }
func (m *ACL_Rule_IpRule_Icmp) String() string { return proto.CompactTextString(m) }
func (*ACL_Rule_IpRule_Icmp) ProtoMessage()    {}
func (*ACL_Rule_IpRule_Icmp) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0, 0, 1}
}

func (m *ACL_Rule_IpRule_Icmp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Rule_IpRule_Icmp.Unmarshal(m, b)
}
func (m *ACL_Rule_IpRule_Icmp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Rule_IpRule_Icmp.Marshal(b, m, deterministic)
}
func (m *ACL_Rule_IpRule_Icmp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Rule_IpRule_Icmp.Merge(m, src)
}
func (m *ACL_Rule_IpRule_Icmp) XXX_Size() int {
	return xxx_messageInfo_ACL_Rule_IpRule_Icmp.Size(m)
}
func (m *ACL_Rule_IpRule_Icmp) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Rule_IpRule_Icmp.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Rule_IpRule_Icmp proto.InternalMessageInfo

func (m *ACL_Rule_IpRule_Icmp) GetIcmpv6() bool {
	if m != nil {
		return m.Icmpv6
	}
	return false
}

func (m *ACL_Rule_IpRule_Icmp) GetIcmpCodeRange() *ACL_Rule_IpRule_Icmp_Range {
	if m != nil {
		return m.IcmpCodeRange
	}
	return nil
}

func (m *ACL_Rule_IpRule_Icmp) GetIcmpTypeRange() *ACL_Rule_IpRule_Icmp_Range {
	if m != nil {
		return m.IcmpTypeRange
	}
	return nil
}

type ACL_Rule_IpRule_Icmp_Range struct {
	First                uint32   `protobuf:"varint,1,opt,name=first,proto3" json:"first,omitempty"`
	Last                 uint32   `protobuf:"varint,2,opt,name=last,proto3" json:"last,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ACL_Rule_IpRule_Icmp_Range) Reset()         { *m = ACL_Rule_IpRule_Icmp_Range{} }
func (m *ACL_Rule_IpRule_Icmp_Range) String() string { return proto.CompactTextString(m) }
func (*ACL_Rule_IpRule_Icmp_Range) ProtoMessage()    {}
func (*ACL_Rule_IpRule_Icmp_Range) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0, 0, 1, 0}
}

func (m *ACL_Rule_IpRule_Icmp_Range) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Rule_IpRule_Icmp_Range.Unmarshal(m, b)
}
func (m *ACL_Rule_IpRule_Icmp_Range) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Rule_IpRule_Icmp_Range.Marshal(b, m, deterministic)
}
func (m *ACL_Rule_IpRule_Icmp_Range) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Rule_IpRule_Icmp_Range.Merge(m, src)
}
func (m *ACL_Rule_IpRule_Icmp_Range) XXX_Size() int {
	return xxx_messageInfo_ACL_Rule_IpRule_Icmp_Range.Size(m)
}
func (m *ACL_Rule_IpRule_Icmp_Range) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Rule_IpRule_Icmp_Range.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Rule_IpRule_Icmp_Range proto.InternalMessageInfo

func (m *ACL_Rule_IpRule_Icmp_Range) GetFirst() uint32 {
	if m != nil {
		return m.First
	}
	return 0
}

func (m *ACL_Rule_IpRule_Icmp_Range) GetLast() uint32 {
	if m != nil {
		return m.Last
	}
	return 0
}

// Inclusive range representing destination ports to be used. When
// only lower-port is present, it represents a single port.
type ACL_Rule_IpRule_PortRange struct {
	LowerPort uint32 `protobuf:"varint,1,opt,name=lower_port,json=lowerPort,proto3" json:"lower_port,omitempty"`
	// If upper port is set, it must
	// be greater or equal to lower port
	UpperPort            uint32   `protobuf:"varint,2,opt,name=upper_port,json=upperPort,proto3" json:"upper_port,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ACL_Rule_IpRule_PortRange) Reset()         { *m = ACL_Rule_IpRule_PortRange{} }
func (m *ACL_Rule_IpRule_PortRange) String() string { return proto.CompactTextString(m) }
func (*ACL_Rule_IpRule_PortRange) ProtoMessage()    {}
func (*ACL_Rule_IpRule_PortRange) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0, 0, 2}
}

func (m *ACL_Rule_IpRule_PortRange) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Rule_IpRule_PortRange.Unmarshal(m, b)
}
func (m *ACL_Rule_IpRule_PortRange) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Rule_IpRule_PortRange.Marshal(b, m, deterministic)
}
func (m *ACL_Rule_IpRule_PortRange) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Rule_IpRule_PortRange.Merge(m, src)
}
func (m *ACL_Rule_IpRule_PortRange) XXX_Size() int {
	return xxx_messageInfo_ACL_Rule_IpRule_PortRange.Size(m)
}
func (m *ACL_Rule_IpRule_PortRange) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Rule_IpRule_PortRange.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Rule_IpRule_PortRange proto.InternalMessageInfo

func (m *ACL_Rule_IpRule_PortRange) GetLowerPort() uint32 {
	if m != nil {
		return m.LowerPort
	}
	return 0
}

func (m *ACL_Rule_IpRule_PortRange) GetUpperPort() uint32 {
	if m != nil {
		return m.UpperPort
	}
	return 0
}

type ACL_Rule_IpRule_Tcp struct {
	DestinationPortRange *ACL_Rule_IpRule_PortRange `protobuf:"bytes,1,opt,name=destination_port_range,json=destinationPortRange,proto3" json:"destination_port_range,omitempty"`
	SourcePortRange      *ACL_Rule_IpRule_PortRange `protobuf:"bytes,2,opt,name=source_port_range,json=sourcePortRange,proto3" json:"source_port_range,omitempty"`
	// Binary mask for tcp flags to match. MSB order (FIN at position 0).
	// Applied as logical AND to tcp flags field of the packet being matched,
	// before it is compared with tcp-flags-value.
	TcpFlagsMask uint32 `protobuf:"varint,3,opt,name=tcp_flags_mask,json=tcpFlagsMask,proto3" json:"tcp_flags_mask,omitempty"`
	// Binary value for tcp flags to match. MSB order (FIN at position 0).
	// Before tcp-flags-value is compared with tcp flags field of the packet being matched,
	// tcp-flags-mask is applied to packet field value.
	TcpFlagsValue        uint32   `protobuf:"varint,4,opt,name=tcp_flags_value,json=tcpFlagsValue,proto3" json:"tcp_flags_value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ACL_Rule_IpRule_Tcp) Reset()         { *m = ACL_Rule_IpRule_Tcp{} }
func (m *ACL_Rule_IpRule_Tcp) String() string { return proto.CompactTextString(m) }
func (*ACL_Rule_IpRule_Tcp) ProtoMessage()    {}
func (*ACL_Rule_IpRule_Tcp) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0, 0, 3}
}

func (m *ACL_Rule_IpRule_Tcp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Rule_IpRule_Tcp.Unmarshal(m, b)
}
func (m *ACL_Rule_IpRule_Tcp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Rule_IpRule_Tcp.Marshal(b, m, deterministic)
}
func (m *ACL_Rule_IpRule_Tcp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Rule_IpRule_Tcp.Merge(m, src)
}
func (m *ACL_Rule_IpRule_Tcp) XXX_Size() int {
	return xxx_messageInfo_ACL_Rule_IpRule_Tcp.Size(m)
}
func (m *ACL_Rule_IpRule_Tcp) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Rule_IpRule_Tcp.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Rule_IpRule_Tcp proto.InternalMessageInfo

func (m *ACL_Rule_IpRule_Tcp) GetDestinationPortRange() *ACL_Rule_IpRule_PortRange {
	if m != nil {
		return m.DestinationPortRange
	}
	return nil
}

func (m *ACL_Rule_IpRule_Tcp) GetSourcePortRange() *ACL_Rule_IpRule_PortRange {
	if m != nil {
		return m.SourcePortRange
	}
	return nil
}

func (m *ACL_Rule_IpRule_Tcp) GetTcpFlagsMask() uint32 {
	if m != nil {
		return m.TcpFlagsMask
	}
	return 0
}

func (m *ACL_Rule_IpRule_Tcp) GetTcpFlagsValue() uint32 {
	if m != nil {
		return m.TcpFlagsValue
	}
	return 0
}

type ACL_Rule_IpRule_Udp struct {
	DestinationPortRange *ACL_Rule_IpRule_PortRange `protobuf:"bytes,1,opt,name=destination_port_range,json=destinationPortRange,proto3" json:"destination_port_range,omitempty"`
	SourcePortRange      *ACL_Rule_IpRule_PortRange `protobuf:"bytes,2,opt,name=source_port_range,json=sourcePortRange,proto3" json:"source_port_range,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                   `json:"-"`
	XXX_unrecognized     []byte                     `json:"-"`
	XXX_sizecache        int32                      `json:"-"`
}

func (m *ACL_Rule_IpRule_Udp) Reset()         { *m = ACL_Rule_IpRule_Udp{} }
func (m *ACL_Rule_IpRule_Udp) String() string { return proto.CompactTextString(m) }
func (*ACL_Rule_IpRule_Udp) ProtoMessage()    {}
func (*ACL_Rule_IpRule_Udp) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0, 0, 4}
}

func (m *ACL_Rule_IpRule_Udp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Rule_IpRule_Udp.Unmarshal(m, b)
}
func (m *ACL_Rule_IpRule_Udp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Rule_IpRule_Udp.Marshal(b, m, deterministic)
}
func (m *ACL_Rule_IpRule_Udp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Rule_IpRule_Udp.Merge(m, src)
}
func (m *ACL_Rule_IpRule_Udp) XXX_Size() int {
	return xxx_messageInfo_ACL_Rule_IpRule_Udp.Size(m)
}
func (m *ACL_Rule_IpRule_Udp) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Rule_IpRule_Udp.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Rule_IpRule_Udp proto.InternalMessageInfo

func (m *ACL_Rule_IpRule_Udp) GetDestinationPortRange() *ACL_Rule_IpRule_PortRange {
	if m != nil {
		return m.DestinationPortRange
	}
	return nil
}

func (m *ACL_Rule_IpRule_Udp) GetSourcePortRange() *ACL_Rule_IpRule_PortRange {
	if m != nil {
		return m.SourcePortRange
	}
	return nil
}

type ACL_Rule_MacIpRule struct {
	SourceAddress       string `protobuf:"bytes,1,opt,name=source_address,json=sourceAddress,proto3" json:"source_address,omitempty"`
	SourceAddressPrefix uint32 `protobuf:"varint,2,opt,name=source_address_prefix,json=sourceAddressPrefix,proto3" json:"source_address_prefix,omitempty"`
	// Before source-mac-address is compared with source mac address field of the packet
	// being matched, source-mac-address-mask is applied to packet field value.
	SourceMacAddress string `protobuf:"bytes,3,opt,name=source_mac_address,json=sourceMacAddress,proto3" json:"source_mac_address,omitempty"`
	// Source MAC address mask.
	// Applied as logical AND with source mac address field of the packet being matched,
	// before it is compared with source-mac-address.
	SourceMacAddressMask string   `protobuf:"bytes,4,opt,name=source_mac_address_mask,json=sourceMacAddressMask,proto3" json:"source_mac_address_mask,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ACL_Rule_MacIpRule) Reset()         { *m = ACL_Rule_MacIpRule{} }
func (m *ACL_Rule_MacIpRule) String() string { return proto.CompactTextString(m) }
func (*ACL_Rule_MacIpRule) ProtoMessage()    {}
func (*ACL_Rule_MacIpRule) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 0, 1}
}

func (m *ACL_Rule_MacIpRule) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Rule_MacIpRule.Unmarshal(m, b)
}
func (m *ACL_Rule_MacIpRule) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Rule_MacIpRule.Marshal(b, m, deterministic)
}
func (m *ACL_Rule_MacIpRule) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Rule_MacIpRule.Merge(m, src)
}
func (m *ACL_Rule_MacIpRule) XXX_Size() int {
	return xxx_messageInfo_ACL_Rule_MacIpRule.Size(m)
}
func (m *ACL_Rule_MacIpRule) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Rule_MacIpRule.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Rule_MacIpRule proto.InternalMessageInfo

func (m *ACL_Rule_MacIpRule) GetSourceAddress() string {
	if m != nil {
		return m.SourceAddress
	}
	return ""
}

func (m *ACL_Rule_MacIpRule) GetSourceAddressPrefix() uint32 {
	if m != nil {
		return m.SourceAddressPrefix
	}
	return 0
}

func (m *ACL_Rule_MacIpRule) GetSourceMacAddress() string {
	if m != nil {
		return m.SourceMacAddress
	}
	return ""
}

func (m *ACL_Rule_MacIpRule) GetSourceMacAddressMask() string {
	if m != nil {
		return m.SourceMacAddressMask
	}
	return ""
}

// The set of interfaces that has assigned this ACL on ingres or egress.
type ACL_Interfaces struct {
	Egress               []string `protobuf:"bytes,1,rep,name=egress,proto3" json:"egress,omitempty"`
	Ingress              []string `protobuf:"bytes,2,rep,name=ingress,proto3" json:"ingress,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ACL_Interfaces) Reset()         { *m = ACL_Interfaces{} }
func (m *ACL_Interfaces) String() string { return proto.CompactTextString(m) }
func (*ACL_Interfaces) ProtoMessage()    {}
func (*ACL_Interfaces) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c13e6c52cc05371, []int{0, 1}
}

func (m *ACL_Interfaces) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ACL_Interfaces.Unmarshal(m, b)
}
func (m *ACL_Interfaces) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ACL_Interfaces.Marshal(b, m, deterministic)
}
func (m *ACL_Interfaces) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ACL_Interfaces.Merge(m, src)
}
func (m *ACL_Interfaces) XXX_Size() int {
	return xxx_messageInfo_ACL_Interfaces.Size(m)
}
func (m *ACL_Interfaces) XXX_DiscardUnknown() {
	xxx_messageInfo_ACL_Interfaces.DiscardUnknown(m)
}

var xxx_messageInfo_ACL_Interfaces proto.InternalMessageInfo

func (m *ACL_Interfaces) GetEgress() []string {
	if m != nil {
		return m.Egress
	}
	return nil
}

func (m *ACL_Interfaces) GetIngress() []string {
	if m != nil {
		return m.Ingress
	}
	return nil
}

func init() {
	proto.RegisterEnum("vpp.acl.ACL_Rule_Action", ACL_Rule_Action_name, ACL_Rule_Action_value)
	proto.RegisterType((*ACL)(nil), "vpp.acl.ACL")
	proto.RegisterType((*ACL_Rule)(nil), "vpp.acl.ACL.Rule")
	proto.RegisterType((*ACL_Rule_IpRule)(nil), "vpp.acl.ACL.Rule.IpRule")
	proto.RegisterType((*ACL_Rule_IpRule_Ip)(nil), "vpp.acl.ACL.Rule.IpRule.Ip")
	proto.RegisterType((*ACL_Rule_IpRule_Icmp)(nil), "vpp.acl.ACL.Rule.IpRule.Icmp")
	proto.RegisterType((*ACL_Rule_IpRule_Icmp_Range)(nil), "vpp.acl.ACL.Rule.IpRule.Icmp.Range")
	proto.RegisterType((*ACL_Rule_IpRule_PortRange)(nil), "vpp.acl.ACL.Rule.IpRule.PortRange")
	proto.RegisterType((*ACL_Rule_IpRule_Tcp)(nil), "vpp.acl.ACL.Rule.IpRule.Tcp")
	proto.RegisterType((*ACL_Rule_IpRule_Udp)(nil), "vpp.acl.ACL.Rule.IpRule.Udp")
	proto.RegisterType((*ACL_Rule_MacIpRule)(nil), "vpp.acl.ACL.Rule.MacIpRule")
	proto.RegisterType((*ACL_Interfaces)(nil), "vpp.acl.ACL.Interfaces")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/api/models/vpp/acl/acl.proto", fileDescriptor_4c13e6c52cc05371)
}

var fileDescriptor_4c13e6c52cc05371 = []byte{
	// 753 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xd4, 0x55, 0xdd, 0x4e, 0xdb, 0x48,
	0x14, 0x5e, 0xdb, 0x21, 0x21, 0x27, 0x04, 0xc2, 0xc0, 0x42, 0xe4, 0x5d, 0x24, 0xc4, 0xfe, 0x45,
	0x62, 0x71, 0x96, 0xac, 0x96, 0xad, 0x5a, 0xa9, 0x52, 0x9a, 0x06, 0x29, 0x2a, 0x41, 0x68, 0x14,
	0xaa, 0xb6, 0xaa, 0x64, 0x0d, 0xe3, 0x49, 0x6a, 0xe1, 0x9f, 0x91, 0xed, 0x84, 0xf2, 0x18, 0x7d,
	0x92, 0xde, 0xf5, 0x25, 0xfa, 0x28, 0xbd, 0xed, 0x6d, 0xa5, 0x6a, 0x7e, 0x9c, 0x84, 0x02, 0x2d,
	0xbd, 0xec, 0x45, 0x94, 0x99, 0xf3, 0xfd, 0xcc, 0x39, 0x73, 0x66, 0xc6, 0xb0, 0x17, 0xf8, 0x23,
	0x92, 0xc5, 0xcd, 0x09, 0xe7, 0x7b, 0x64, 0xc4, 0xa2, 0xac, 0x49, 0xb8, 0xdf, 0x0c, 0x63, 0x8f,
	0x05, 0xa9, 0x08, 0x36, 0x09, 0x0d, 0xc4, 0xcf, 0xe1, 0x49, 0x9c, 0xc5, 0xa8, 0x34, 0xe1, 0xdc,
	0x21, 0x34, 0xd8, 0xf9, 0xb8, 0x04, 0x56, 0xbb, 0x73, 0x84, 0x10, 0x14, 0x22, 0x12, 0xb2, 0xba,
	0xb1, 0x6d, 0x34, 0xca, 0x58, 0x8e, 0xd1, 0x5f, 0xb0, 0x90, 0x8c, 0x03, 0x96, 0xd6, 0xcd, 0x6d,
	0xab, 0x51, 0x69, 0xad, 0x3a, 0x5a, 0xe4, 0xb4, 0x3b, 0x47, 0x0e, 0x1e, 0x07, 0x0c, 0x2b, 0x1c,
	0xfd, 0x0f, 0xe0, 0x47, 0x19, 0x4b, 0x86, 0x84, 0xb2, 0xb4, 0x6e, 0x6d, 0x1b, 0x8d, 0x4a, 0x6b,
	0xf3, 0x0a, 0xbb, 0x37, 0x85, 0xf1, 0x1c, 0xd5, 0x7e, 0x57, 0x81, 0x82, 0x30, 0x42, 0xff, 0x40,
	0x91, 0xd0, 0xcc, 0x8f, 0x23, 0x99, 0xc0, 0x72, 0xab, 0x7e, 0x6d, 0x2d, 0xa7, 0x2d, 0x71, 0xac,
	0x79, 0x68, 0x1f, 0x4a, 0x3e, 0x77, 0xc5, 0xfa, 0x75, 0x53, 0x2e, 0x78, 0x83, 0xa4, 0xc7, 0x65,
	0x96, 0x45, 0x5f, 0xfe, 0xa3, 0xfb, 0x00, 0x21, 0xa1, 0xb9, 0x4a, 0xa5, 0xf9, 0xcb, 0x75, 0x55,
	0x9f, 0x50, 0x2d, 0x2c, 0x4b, 0xba, 0x18, 0xda, 0x6f, 0x16, 0xa1, 0xa8, 0xa2, 0x68, 0x17, 0x4c,
	0x9f, 0xcb, 0x3c, 0x6f, 0x94, 0x2b, 0x96, 0xf8, 0x33, 0x7d, 0x8e, 0xf6, 0xa1, 0xe0, 0xd3, 0x90,
	0xeb, 0x1c, 0xb7, 0x6e, 0xa7, 0xd3, 0x90, 0x63, 0x49, 0x45, 0x0e, 0x58, 0x19, 0xe5, 0x3a, 0xbf,
	0x5f, 0x6f, 0x55, 0x0c, 0x28, 0xc7, 0x82, 0x28, 0xf8, 0x63, 0x8f, 0xd7, 0x0b, 0xdf, 0xe0, 0x9f,
	0x7a, 0x1c, 0x0b, 0xa2, 0xfd, 0x12, 0xcc, 0x1e, 0x47, 0x4d, 0x58, 0xf3, 0x58, 0x9a, 0xf9, 0x11,
	0x11, 0xdb, 0xe9, 0x46, 0x2c, 0xbb, 0x88, 0x93, 0x73, 0xdd, 0x7f, 0x34, 0x07, 0x1d, 0x2b, 0x04,
	0xfd, 0x01, 0xcb, 0x69, 0x3c, 0x4e, 0x28, 0x9b, 0x72, 0x4d, 0xc9, 0xad, 0xaa, 0xa8, 0xa6, 0xd9,
	0x1f, 0x0c, 0x28, 0x88, 0x62, 0xd0, 0x06, 0x14, 0x45, 0x39, 0x93, 0x03, 0xe9, 0xb9, 0x88, 0xf5,
	0x0c, 0x3d, 0x81, 0x15, 0x31, 0x72, 0x69, 0xec, 0x31, 0x37, 0x21, 0xd1, 0x28, 0x6f, 0xe0, 0x6f,
	0x5f, 0xdd, 0x1c, 0x07, 0x0b, 0x2a, 0xae, 0x0a, 0x6d, 0x27, 0xf6, 0x98, 0x9c, 0x4e, 0xcd, 0xb2,
	0x4b, 0x9e, 0x9b, 0x59, 0xdf, 0x69, 0x36, 0xb8, 0xe4, 0xca, 0xcc, 0xde, 0x87, 0x05, 0xe5, 0xba,
	0x0e, 0x0b, 0x43, 0x3f, 0x49, 0x33, 0x99, 0x79, 0x15, 0xab, 0x89, 0xb8, 0x22, 0x01, 0x49, 0x33,
	0x99, 0x6d, 0x15, 0xcb, 0xb1, 0xdd, 0x83, 0xf2, 0x49, 0x9c, 0x64, 0x4a, 0xb6, 0x05, 0x10, 0xc4,
	0x17, 0x2c, 0x71, 0x79, 0x9c, 0xe4, 0xda, 0xb2, 0x8c, 0x08, 0x8e, 0x80, 0xc7, 0x9c, 0xe7, 0xb0,
	0x72, 0x29, 0xcb, 0x88, 0x80, 0xed, 0x4f, 0x06, 0x58, 0x03, 0xca, 0xd1, 0x33, 0xd8, 0x98, 0x6f,
	0x8c, 0x20, 0xeb, 0xca, 0xd4, 0x91, 0xdb, 0xb9, 0xb5, 0xb2, 0x69, 0x26, 0x78, 0x7d, 0xce, 0x61,
	0x96, 0xdf, 0x31, 0xac, 0xea, 0x0e, 0xce, 0x99, 0x9a, 0x77, 0x36, 0x5d, 0x51, 0xe2, 0x99, 0xdf,
	0xef, 0xb0, 0x9c, 0x51, 0xee, 0x0e, 0x03, 0x32, 0x4a, 0xdd, 0x90, 0xa4, 0xe7, 0x72, 0xef, 0xab,
	0x78, 0x29, 0xa3, 0xfc, 0x50, 0x04, 0xfb, 0x24, 0x3d, 0x47, 0x7f, 0xc2, 0xca, 0x8c, 0x35, 0x21,
	0xc1, 0x98, 0xc9, 0xa3, 0x5a, 0xc5, 0xd5, 0x9c, 0xf6, 0x54, 0x04, 0xed, 0xb7, 0x06, 0x58, 0xa7,
	0xde, 0x0f, 0x54, 0xbf, 0xfd, 0xde, 0x80, 0xf2, 0xf4, 0xb1, 0x98, 0xbb, 0x1f, 0xc4, 0xf3, 0x12,
	0x96, 0xa6, 0xfa, 0x2e, 0xe9, 0xfb, 0xd1, 0x56, 0x41, 0xd4, 0x82, 0x9f, 0xaf, 0xd2, 0x5c, 0x9e,
	0xb0, 0xa1, 0xff, 0x5a, 0x1f, 0x88, 0xb5, 0x2b, 0xec, 0x13, 0x09, 0xa1, 0xbf, 0x01, 0x69, 0x4d,
	0x48, 0xe8, 0xd4, 0xde, 0x92, 0xf6, 0x35, 0x85, 0xf4, 0x09, 0xcd, 0x57, 0xf8, 0x0f, 0x36, 0xaf,
	0xb3, 0x55, 0x7f, 0x0a, 0x52, 0xb2, 0xfe, 0xa5, 0x44, 0xf4, 0x69, 0x67, 0x17, 0x8a, 0xea, 0x89,
	0x45, 0x8b, 0x50, 0x78, 0xdc, 0x3d, 0x7e, 0x5e, 0xfb, 0x09, 0x01, 0x14, 0x4f, 0xba, 0xb8, 0xdf,
	0x1b, 0xd4, 0x0c, 0x54, 0x81, 0x12, 0xee, 0x1e, 0x1e, 0x75, 0x3b, 0x83, 0x9a, 0x69, 0x3f, 0x04,
	0x98, 0x3d, 0xe9, 0xe2, 0xaa, 0xb3, 0x91, 0x2e, 0xd9, 0x6a, 0x94, 0xb1, 0x9e, 0xa1, 0x3a, 0x94,
	0xfc, 0x48, 0x01, 0xa6, 0x04, 0xf2, 0xe9, 0xa3, 0x7b, 0x2f, 0x0e, 0x46, 0x7e, 0xf6, 0x6a, 0x7c,
	0xe6, 0xd0, 0x38, 0x6c, 0xde, 0xe1, 0xdb, 0xf5, 0x60, 0xc2, 0xb9, 0x4b, 0x68, 0x70, 0x56, 0x94,
	0x1f, 0xb0, 0x7f, 0x3f, 0x07, 0x00, 0x00, 0xff, 0xff, 0xbc, 0x04, 0x8e, 0xaf, 0xf1, 0x06, 0x00,
	0x00,
}
