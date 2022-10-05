// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

// Package acl_types contains generated bindings for API file acl_types.api.
//
// Contents:
//   1 enum
//   2 structs
//
package acl_types

import (
	"strconv"

	api "go.fd.io/govpp/api"
	ethernet_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ethernet_types"
	ip_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ip_types"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "acl_types"
	APIVersion = "1.0.0"
	VersionCrc = 0x878da4fa
)

// ACLAction defines enum 'acl_action'.
type ACLAction uint8

const (
	ACL_ACTION_API_DENY           ACLAction = 0
	ACL_ACTION_API_PERMIT         ACLAction = 1
	ACL_ACTION_API_PERMIT_REFLECT ACLAction = 2
)

var (
	ACLAction_name = map[uint8]string{
		0: "ACL_ACTION_API_DENY",
		1: "ACL_ACTION_API_PERMIT",
		2: "ACL_ACTION_API_PERMIT_REFLECT",
	}
	ACLAction_value = map[string]uint8{
		"ACL_ACTION_API_DENY":           0,
		"ACL_ACTION_API_PERMIT":         1,
		"ACL_ACTION_API_PERMIT_REFLECT": 2,
	}
)

func (x ACLAction) String() string {
	s, ok := ACLAction_name[uint8(x)]
	if ok {
		return s
	}
	return "ACLAction(" + strconv.Itoa(int(x)) + ")"
}

// ACLRule defines type 'acl_rule'.
type ACLRule struct {
	IsPermit               ACLAction        `binapi:"acl_action,name=is_permit" json:"is_permit,omitempty"`
	SrcPrefix              ip_types.Prefix  `binapi:"prefix,name=src_prefix" json:"src_prefix,omitempty"`
	DstPrefix              ip_types.Prefix  `binapi:"prefix,name=dst_prefix" json:"dst_prefix,omitempty"`
	Proto                  ip_types.IPProto `binapi:"ip_proto,name=proto" json:"proto,omitempty"`
	SrcportOrIcmptypeFirst uint16           `binapi:"u16,name=srcport_or_icmptype_first" json:"srcport_or_icmptype_first,omitempty"`
	SrcportOrIcmptypeLast  uint16           `binapi:"u16,name=srcport_or_icmptype_last" json:"srcport_or_icmptype_last,omitempty"`
	DstportOrIcmpcodeFirst uint16           `binapi:"u16,name=dstport_or_icmpcode_first" json:"dstport_or_icmpcode_first,omitempty"`
	DstportOrIcmpcodeLast  uint16           `binapi:"u16,name=dstport_or_icmpcode_last" json:"dstport_or_icmpcode_last,omitempty"`
	TCPFlagsMask           uint8            `binapi:"u8,name=tcp_flags_mask" json:"tcp_flags_mask,omitempty"`
	TCPFlagsValue          uint8            `binapi:"u8,name=tcp_flags_value" json:"tcp_flags_value,omitempty"`
}

// MacipACLRule defines type 'macip_acl_rule'.
type MacipACLRule struct {
	IsPermit   ACLAction                 `binapi:"acl_action,name=is_permit" json:"is_permit,omitempty"`
	SrcMac     ethernet_types.MacAddress `binapi:"mac_address,name=src_mac" json:"src_mac,omitempty"`
	SrcMacMask ethernet_types.MacAddress `binapi:"mac_address,name=src_mac_mask" json:"src_mac_mask,omitempty"`
	SrcPrefix  ip_types.Prefix           `binapi:"prefix,name=src_prefix" json:"src_prefix,omitempty"`
}
