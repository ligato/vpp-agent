// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/vpp/acl/acl.proto

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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0, 0}
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
	return fileDescriptor_0255db9126aac1d4, []int{0}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0, 0}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0, 0, 0}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0, 0, 1}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0, 0, 1, 0}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0, 0, 2}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0, 0, 3}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0, 0, 4}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 0, 1}
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
	return fileDescriptor_0255db9126aac1d4, []int{0, 1}
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

func init() { proto.RegisterFile("ligato/vpp-agent/vpp/acl/acl.proto", fileDescriptor_0255db9126aac1d4) }

var fileDescriptor_0255db9126aac1d4 = []byte{
	// 747 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xd4, 0x55, 0x59, 0x6f, 0xd3, 0x4a,
	0x14, 0xbe, 0xb6, 0xd3, 0xa4, 0x39, 0x69, 0xda, 0x74, 0xda, 0xdb, 0x46, 0xbe, 0xb7, 0x52, 0x95,
	0xbb, 0x45, 0xea, 0xc5, 0xa1, 0x41, 0x80, 0x04, 0x12, 0x22, 0x0d, 0xa9, 0x14, 0xd1, 0x54, 0xd5,
	0x28, 0x45, 0x80, 0x90, 0xac, 0x61, 0x3c, 0x89, 0xac, 0x3a, 0xf6, 0xc8, 0x76, 0x52, 0xfa, 0x33,
	0xf8, 0x25, 0xbc, 0xf1, 0x27, 0xf8, 0x29, 0xbc, 0xf2, 0x8a, 0x84, 0x66, 0x71, 0x16, 0xda, 0xb0,
	0x3c, 0xf2, 0x10, 0x65, 0xe6, 0x7c, 0xcb, 0x9c, 0xe3, 0x33, 0x0b, 0xd4, 0x02, 0x7f, 0x48, 0xd2,
	0xa8, 0x31, 0xe1, 0xfc, 0x16, 0x19, 0xb2, 0x30, 0x15, 0xa3, 0x06, 0xa1, 0x81, 0xf8, 0x39, 0x3c,
	0x8e, 0xd2, 0x08, 0x15, 0x26, 0x9c, 0x3b, 0x84, 0x06, 0xb5, 0x4f, 0x6b, 0x60, 0xb5, 0xda, 0x27,
	0x08, 0x41, 0x2e, 0x24, 0x23, 0x56, 0x35, 0xf6, 0x8d, 0x7a, 0x11, 0xcb, 0x31, 0xfa, 0x0f, 0x56,
	0xe2, 0x71, 0xc0, 0x92, 0xaa, 0xb9, 0x6f, 0xd5, 0x4b, 0xcd, 0x4d, 0x47, 0x8b, 0x9c, 0x56, 0xfb,
	0xc4, 0xc1, 0xe3, 0x80, 0x61, 0x85, 0xa3, 0xfb, 0x00, 0x7e, 0x98, 0xb2, 0x78, 0x40, 0x28, 0x4b,
	0xaa, 0xd6, 0xbe, 0x51, 0x2f, 0x35, 0x77, 0x17, 0xd8, 0xdd, 0x29, 0x8c, 0xe7, 0xa8, 0xf6, 0xfb,
	0x12, 0xe4, 0x84, 0x11, 0xba, 0x0d, 0x79, 0x42, 0x53, 0x3f, 0x0a, 0x65, 0x02, 0xeb, 0xcd, 0xea,
	0xb5, 0xb5, 0x9c, 0x96, 0xc4, 0xb1, 0xe6, 0xa1, 0x43, 0x28, 0xf8, 0xdc, 0x15, 0xeb, 0x57, 0x4d,
	0xb9, 0xe0, 0x0d, 0x92, 0x2e, 0x97, 0x59, 0xe6, 0x7d, 0xf9, 0x8f, 0x1e, 0x00, 0x8c, 0x08, 0xcd,
	0x54, 0x2a, 0xcd, 0x3f, 0xae, 0xab, 0x7a, 0x84, 0x6a, 0x61, 0x51, 0xd2, 0xc5, 0xd0, 0x7e, 0xbb,
	0x0a, 0x79, 0x15, 0x45, 0x07, 0x60, 0xfa, 0x5c, 0xe6, 0x79, 0xa3, 0x5c, 0xb1, 0xc4, 0x9f, 0xe9,
	0x73, 0x74, 0x08, 0x39, 0x9f, 0x8e, 0xb8, 0xce, 0x71, 0x6f, 0x39, 0x9d, 0x8e, 0x38, 0x96, 0x54,
	0xe4, 0x80, 0x95, 0x52, 0xae, 0xf3, 0xfb, 0x73, 0xa9, 0xa2, 0x4f, 0x39, 0x16, 0x44, 0xc1, 0x1f,
	0x7b, 0xbc, 0x9a, 0xfb, 0x0e, 0xff, 0xdc, 0xe3, 0x58, 0x10, 0xed, 0x57, 0x60, 0x76, 0x39, 0x6a,
	0xc0, 0x96, 0xc7, 0x92, 0xd4, 0x0f, 0x89, 0xf8, 0x9c, 0x6e, 0xc8, 0xd2, 0xcb, 0x28, 0xbe, 0xd0,
	0xfd, 0x47, 0x73, 0xd0, 0xa9, 0x42, 0xd0, 0x3f, 0xb0, 0x9e, 0x44, 0xe3, 0x98, 0xb2, 0x29, 0xd7,
	0x94, 0xdc, 0xb2, 0x8a, 0x6a, 0x9a, 0xfd, 0xd1, 0x80, 0x9c, 0x28, 0x06, 0xed, 0x40, 0x5e, 0x94,
	0x33, 0xb9, 0x27, 0x3d, 0x57, 0xb1, 0x9e, 0xa1, 0xa7, 0xb0, 0x21, 0x46, 0x2e, 0x8d, 0x3c, 0xe6,
	0xc6, 0x24, 0x1c, 0x66, 0x0d, 0xfc, 0xeb, 0x9b, 0x1f, 0xc7, 0xc1, 0x82, 0x8a, 0xcb, 0x42, 0xdb,
	0x8e, 0x3c, 0x26, 0xa7, 0x53, 0xb3, 0xf4, 0x8a, 0x67, 0x66, 0xd6, 0x4f, 0x9a, 0xf5, 0xaf, 0xb8,
	0x32, 0xb3, 0x0f, 0x61, 0x45, 0xb9, 0x6e, 0xc3, 0xca, 0xc0, 0x8f, 0x93, 0x54, 0x66, 0x5e, 0xc6,
	0x6a, 0x22, 0x8e, 0x48, 0x40, 0x92, 0x54, 0x66, 0x5b, 0xc6, 0x72, 0x6c, 0x77, 0xa1, 0x78, 0x16,
	0xc5, 0xa9, 0x92, 0xed, 0x01, 0x04, 0xd1, 0x25, 0x8b, 0x5d, 0x1e, 0xc5, 0x99, 0xb6, 0x28, 0x23,
	0x82, 0x23, 0xe0, 0x31, 0xe7, 0x19, 0xac, 0x5c, 0x8a, 0x32, 0x22, 0x60, 0xfb, 0xb3, 0x01, 0x56,
	0x9f, 0x72, 0xf4, 0x1c, 0x76, 0xe6, 0x1b, 0x23, 0xc8, 0xba, 0x32, 0xb5, 0xe5, 0x6a, 0x4b, 0x2b,
	0x9b, 0x66, 0x82, 0xb7, 0xe7, 0x1c, 0x66, 0xf9, 0x9d, 0xc2, 0xa6, 0xee, 0xe0, 0x9c, 0xa9, 0xf9,
	0xc3, 0xa6, 0x1b, 0x4a, 0x3c, 0xf3, 0xfb, 0x1b, 0xd6, 0x53, 0xca, 0xdd, 0x41, 0x40, 0x86, 0x89,
	0x3b, 0x22, 0xc9, 0x85, 0xfc, 0xf6, 0x65, 0xbc, 0x96, 0x52, 0x7e, 0x2c, 0x82, 0x3d, 0x92, 0x5c,
	0xa0, 0x7f, 0x61, 0x63, 0xc6, 0x9a, 0x90, 0x60, 0xcc, 0xe4, 0x56, 0x2d, 0xe3, 0x72, 0x46, 0x7b,
	0x26, 0x82, 0xf6, 0x3b, 0x03, 0xac, 0x73, 0xef, 0x17, 0xaa, 0xdf, 0xfe, 0x60, 0x40, 0x71, 0x7a,
	0x59, 0xcc, 0x9d, 0x0f, 0xe2, 0x79, 0x31, 0x4b, 0x12, 0x7d, 0x96, 0xf4, 0xf9, 0x68, 0xa9, 0x20,
	0x6a, 0xc2, 0xef, 0x8b, 0x34, 0x97, 0xc7, 0x6c, 0xe0, 0xbf, 0xd1, 0x1b, 0x62, 0x6b, 0x81, 0x7d,
	0x26, 0x21, 0xf4, 0x3f, 0x20, 0xad, 0x19, 0x11, 0x3a, 0xb5, 0xb7, 0xa4, 0x7d, 0x45, 0x21, 0x3d,
	0x42, 0xb3, 0x15, 0xee, 0xc2, 0xee, 0x75, 0xb6, 0xea, 0x4f, 0x4e, 0x4a, 0xb6, 0xbf, 0x96, 0x88,
	0x3e, 0xd5, 0x0e, 0x20, 0xaf, 0xae, 0x58, 0xb4, 0x0a, 0xb9, 0x27, 0x9d, 0xd3, 0x17, 0x95, 0xdf,
	0x10, 0x40, 0xfe, 0xac, 0x83, 0x7b, 0xdd, 0x7e, 0xc5, 0x40, 0x25, 0x28, 0xe0, 0xce, 0xf1, 0x49,
	0xa7, 0xdd, 0xaf, 0x98, 0xf6, 0x23, 0x80, 0xd9, 0x95, 0x2e, 0x8e, 0x3a, 0x1b, 0xea, 0x92, 0xad,
	0x7a, 0x11, 0xeb, 0x19, 0xaa, 0x42, 0xc1, 0x0f, 0x15, 0x60, 0x4a, 0x20, 0x9b, 0x1e, 0x1d, 0xbd,
	0x7c, 0x3c, 0x8c, 0x1c, 0xf5, 0x50, 0x39, 0xfe, 0xc2, 0x5b, 0xd5, 0x6c, 0xc8, 0x27, 0xaa, 0xb1,
	0xec, 0x15, 0x7b, 0x38, 0xe1, 0xdc, 0x25, 0x34, 0x78, 0x9d, 0x97, 0xbc, 0x3b, 0x5f, 0x02, 0x00,
	0x00, 0xff, 0xff, 0xea, 0x7e, 0x22, 0x27, 0xf0, 0x06, 0x00, 0x00,
}
