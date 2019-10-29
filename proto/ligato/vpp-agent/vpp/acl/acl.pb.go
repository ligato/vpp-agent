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
	Action               ACL_Rule_Action     `protobuf:"varint,1,opt,name=action,proto3,enum=ligato.vpp_agent.vpp.acl.ACL_Rule_Action" json:"action,omitempty"`
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
	proto.RegisterEnum("ligato.vpp_agent.vpp.acl.ACL_Rule_Action", ACL_Rule_Action_name, ACL_Rule_Action_value)
	proto.RegisterType((*ACL)(nil), "ligato.vpp_agent.vpp.acl.ACL")
	proto.RegisterType((*ACL_Rule)(nil), "ligato.vpp_agent.vpp.acl.ACL.Rule")
	proto.RegisterType((*ACL_Rule_IpRule)(nil), "ligato.vpp_agent.vpp.acl.ACL.Rule.IpRule")
	proto.RegisterType((*ACL_Rule_IpRule_Ip)(nil), "ligato.vpp_agent.vpp.acl.ACL.Rule.IpRule.Ip")
	proto.RegisterType((*ACL_Rule_IpRule_Icmp)(nil), "ligato.vpp_agent.vpp.acl.ACL.Rule.IpRule.Icmp")
	proto.RegisterType((*ACL_Rule_IpRule_Icmp_Range)(nil), "ligato.vpp_agent.vpp.acl.ACL.Rule.IpRule.Icmp.Range")
	proto.RegisterType((*ACL_Rule_IpRule_PortRange)(nil), "ligato.vpp_agent.vpp.acl.ACL.Rule.IpRule.PortRange")
	proto.RegisterType((*ACL_Rule_IpRule_Tcp)(nil), "ligato.vpp_agent.vpp.acl.ACL.Rule.IpRule.Tcp")
	proto.RegisterType((*ACL_Rule_IpRule_Udp)(nil), "ligato.vpp_agent.vpp.acl.ACL.Rule.IpRule.Udp")
	proto.RegisterType((*ACL_Rule_MacIpRule)(nil), "ligato.vpp_agent.vpp.acl.ACL.Rule.MacIpRule")
	proto.RegisterType((*ACL_Interfaces)(nil), "ligato.vpp_agent.vpp.acl.ACL.Interfaces")
}

func init() { proto.RegisterFile("ligato/vpp-agent/vpp/acl/acl.proto", fileDescriptor_0255db9126aac1d4) }

var fileDescriptor_0255db9126aac1d4 = []byte{
	// 757 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xdc, 0x55, 0xdd, 0x6e, 0xd3, 0x48,
	0x14, 0xde, 0xd8, 0x69, 0x5a, 0x9f, 0xae, 0xd3, 0xec, 0xb4, 0xdb, 0x8d, 0x2c, 0xad, 0x54, 0x45,
	0xbb, 0x28, 0x88, 0xd6, 0x11, 0x29, 0x20, 0x24, 0x10, 0x90, 0x84, 0x54, 0x44, 0x34, 0x55, 0x35,
	0x4a, 0x91, 0x40, 0x95, 0xac, 0x61, 0x3c, 0x89, 0xac, 0x3a, 0xf6, 0xc8, 0x76, 0x52, 0xfa, 0x10,
	0x5c, 0xf1, 0x00, 0x3c, 0x08, 0x8f, 0xc0, 0x15, 0xcf, 0xc2, 0x0b, 0xa0, 0xf9, 0x71, 0x92, 0x82,
	0x0a, 0x29, 0xdc, 0x71, 0x61, 0x79, 0xe6, 0x9c, 0xef, 0xfb, 0xe6, 0x7c, 0x3e, 0x33, 0x1e, 0xa8,
	0x85, 0xc1, 0x88, 0x64, 0x71, 0x63, 0xca, 0xf9, 0x1e, 0x19, 0xb1, 0x28, 0x13, 0xa3, 0x06, 0xa1,
	0xa1, 0x78, 0x5c, 0x9e, 0xc4, 0x59, 0x8c, 0xaa, 0x0a, 0xe3, 0x4e, 0x39, 0xf7, 0x24, 0x46, 0x8c,
	0x5c, 0x42, 0xc3, 0xda, 0xe7, 0x32, 0x98, 0xad, 0xce, 0x21, 0x42, 0x50, 0x8c, 0xc8, 0x98, 0x55,
	0x0b, 0x3b, 0x85, 0xba, 0x85, 0xe5, 0x18, 0xdd, 0x87, 0x95, 0x64, 0x12, 0xb2, 0xb4, 0x6a, 0xec,
	0x98, 0xf5, 0xf5, 0x66, 0xcd, 0xbd, 0x4a, 0xc5, 0x6d, 0x75, 0x0e, 0x5d, 0x3c, 0x09, 0x19, 0x56,
	0x04, 0xf4, 0x0c, 0x20, 0x88, 0x32, 0x96, 0x0c, 0x09, 0x65, 0x69, 0xd5, 0xdc, 0x29, 0xd4, 0xd7,
	0x9b, 0xf5, 0xef, 0xd3, 0x7b, 0x33, 0x3c, 0x5e, 0xe0, 0x3a, 0xef, 0x6c, 0x28, 0x0a, 0x65, 0xd4,
	0x82, 0x12, 0xa1, 0x59, 0x10, 0x47, 0xb2, 0xc4, 0x72, 0xf3, 0xe6, 0x8f, 0xab, 0x71, 0x5b, 0x92,
	0x80, 0x35, 0x11, 0xb5, 0x61, 0x35, 0xe0, 0x9e, 0xa8, 0xb0, 0x6a, 0xc8, 0x92, 0x96, 0xd1, 0xe8,
	0x71, 0x69, 0xac, 0x14, 0xc8, 0x37, 0x7a, 0x0e, 0x30, 0x26, 0x34, 0x97, 0x51, 0xce, 0x76, 0x97,
	0x90, 0xe9, 0x13, 0xaa, 0x95, 0x2c, 0xc9, 0x17, 0x43, 0xe7, 0x83, 0x05, 0x25, 0x15, 0x45, 0x0f,
	0xc1, 0x08, 0xb8, 0xb4, 0xb6, 0x9c, 0x9e, 0xa2, 0x89, 0x97, 0x11, 0x70, 0xd4, 0x86, 0x62, 0x40,
	0xc7, 0x5c, 0xdb, 0x72, 0xaf, 0xc1, 0xa7, 0x63, 0x8e, 0x25, 0x17, 0x3d, 0x06, 0x33, 0xa3, 0x5c,
	0x5b, 0xda, 0x5b, 0x5e, 0x62, 0x40, 0x39, 0x16, 0x4c, 0x21, 0x30, 0xf1, 0x79, 0xb5, 0x78, 0x5d,
	0x81, 0x13, 0x9f, 0x63, 0xc1, 0x74, 0x4e, 0xc1, 0xe8, 0x71, 0xd4, 0x80, 0x4d, 0x9f, 0xa5, 0x59,
	0x10, 0x11, 0xd1, 0x34, 0x2f, 0x62, 0xd9, 0x79, 0x9c, 0x9c, 0xe9, 0x8d, 0x89, 0x16, 0x52, 0x47,
	0x2a, 0x83, 0xfe, 0x87, 0x72, 0x1a, 0x4f, 0x12, 0xca, 0x66, 0x58, 0x43, 0x62, 0x6d, 0x15, 0xd5,
	0x30, 0xe7, 0xad, 0x01, 0x45, 0x61, 0x17, 0x6d, 0x43, 0x49, 0x18, 0x9e, 0xde, 0x93, 0x9a, 0x6b,
	0x58, 0xcf, 0xd0, 0x29, 0x6c, 0x88, 0x91, 0x47, 0x63, 0x9f, 0x79, 0x09, 0x89, 0x46, 0xf9, 0x36,
	0xb9, 0x73, 0xbd, 0xef, 0xe9, 0x62, 0xc1, 0xc5, 0xb6, 0x10, 0xeb, 0xc4, 0x3e, 0x93, 0xd3, 0x99,
	0x7a, 0x76, 0xc1, 0x73, 0x75, 0xf3, 0x57, 0xd5, 0x07, 0x17, 0x5c, 0xa9, 0x3b, 0xb7, 0x61, 0x45,
	0x2d, 0xb3, 0x05, 0x2b, 0xc3, 0x20, 0x49, 0x33, 0xe9, 0xcd, 0xc6, 0x6a, 0x22, 0x4e, 0x77, 0x48,
	0xd2, 0x4c, 0xfa, 0xb1, 0xb1, 0x1c, 0x3b, 0x3d, 0xb0, 0x8e, 0xe3, 0x24, 0x53, 0xb4, 0x7f, 0x01,
	0xc2, 0xf8, 0x9c, 0x25, 0x1e, 0x8f, 0x93, 0x9c, 0x6b, 0xc9, 0x88, 0xc0, 0x88, 0xf4, 0x84, 0xf3,
	0x3c, 0xad, 0x54, 0x2c, 0x19, 0x11, 0x69, 0xe7, 0xbd, 0x01, 0xe6, 0x80, 0x72, 0x14, 0xc0, 0xf6,
	0x62, 0xeb, 0x04, 0x58, 0x5b, 0x55, 0x1b, 0x7b, 0x7f, 0x79, 0xab, 0xb3, 0xd2, 0xf0, 0xd6, 0x82,
	0xe4, 0xbc, 0x60, 0x0f, 0xfe, 0xd2, 0x4d, 0x5f, 0x58, 0xc5, 0xf8, 0xf9, 0x55, 0x36, 0x94, 0xda,
	0x7c, 0x81, 0xff, 0xa0, 0x9c, 0x51, 0xee, 0x0d, 0x43, 0x32, 0x4a, 0xbd, 0x31, 0x49, 0xcf, 0x64,
	0xbb, 0x6c, 0xfc, 0x67, 0x46, 0xf9, 0x81, 0x08, 0xf6, 0x49, 0x7a, 0x86, 0x6e, 0xc0, 0xc6, 0x1c,
	0x35, 0x25, 0xe1, 0x84, 0xc9, 0xfd, 0x6f, 0x63, 0x3b, 0x87, 0xbd, 0x10, 0x41, 0xe7, 0x53, 0x01,
	0xcc, 0x13, 0xff, 0xb7, 0xfa, 0x42, 0xce, 0xc7, 0x02, 0x58, 0xb3, 0xdf, 0xda, 0xc2, 0x29, 0x24,
	0xbe, 0x9f, 0xb0, 0x34, 0xd5, 0x27, 0x56, 0x9f, 0xc2, 0x96, 0x0a, 0xa2, 0x26, 0xfc, 0x7d, 0x19,
	0xe6, 0xf1, 0x84, 0x0d, 0x83, 0x37, 0x7a, 0x53, 0x6d, 0x5e, 0x42, 0x1f, 0xcb, 0x14, 0xda, 0x05,
	0xa4, 0x39, 0x63, 0x42, 0x67, 0xf2, 0xa6, 0x94, 0xaf, 0xa8, 0x4c, 0x9f, 0xd0, 0x7c, 0x85, 0xbb,
	0xf0, 0xcf, 0xb7, 0x68, 0xd5, 0xc1, 0xa2, 0xa4, 0x6c, 0x7d, 0x4d, 0x11, 0x9d, 0xac, 0xdd, 0x82,
	0x92, 0xba, 0x2e, 0xd0, 0x1a, 0x14, 0x9f, 0x76, 0x8f, 0x5e, 0x56, 0xfe, 0x40, 0x00, 0xa5, 0xe3,
	0x2e, 0xee, 0xf7, 0x06, 0x95, 0x02, 0x5a, 0x87, 0x55, 0xdc, 0x3d, 0x38, 0xec, 0x76, 0x06, 0x15,
	0xc3, 0x79, 0x04, 0x30, 0xbf, 0xaf, 0xc4, 0x0f, 0x85, 0x8d, 0xb4, 0x65, 0xb3, 0x6e, 0x61, 0x3d,
	0x43, 0x55, 0x58, 0x0d, 0x22, 0x95, 0x30, 0x64, 0x22, 0x9f, 0xb6, 0xdb, 0xaf, 0x9e, 0x8c, 0xe2,
	0xbc, 0x09, 0xc1, 0xa5, 0xbb, 0xbb, 0xd9, 0x90, 0x57, 0x76, 0xe3, 0xaa, 0x5b, 0xfd, 0x81, 0xec,
	0x19, 0x0d, 0x5f, 0x97, 0x24, 0x6e, 0xff, 0x4b, 0x00, 0x00, 0x00, 0xff, 0xff, 0xd3, 0x5e, 0x8b,
	0xa4, 0x00, 0x08, 0x00, 0x00,
}
