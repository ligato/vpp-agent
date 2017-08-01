// Package acl represents the VPP binary API of the 'acl' VPP module.
// DO NOT EDIT. Generated from './bin_api/acl.api.json'
package acl

import "git.fd.io/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0x3cd02d84

// ACLRule represents the VPP binary API data type 'acl_rule'.
// Generated from './bin_api/acl.api.json', line 3:
//
//        ["acl_rule",
//            ["u8", "is_permit"],
//            ["u8", "is_ipv6"],
//            ["u8", "src_ip_addr", 16],
//            ["u8", "src_ip_prefix_len"],
//            ["u8", "dst_ip_addr", 16],
//            ["u8", "dst_ip_prefix_len"],
//            ["u8", "proto"],
//            ["u16", "srcport_or_icmptype_first"],
//            ["u16", "srcport_or_icmptype_last"],
//            ["u16", "dstport_or_icmpcode_first"],
//            ["u16", "dstport_or_icmpcode_last"],
//            ["u8", "tcp_flags_mask"],
//            ["u8", "tcp_flags_value"],
//            {"crc" : "0x2715e1c0"}
//        ],
//
type ACLRule struct {
	IsPermit               uint8
	IsIpv6                 uint8
	SrcIPAddr              []byte `struc:"[16]byte"`
	SrcIPPrefixLen         uint8
	DstIPAddr              []byte `struc:"[16]byte"`
	DstIPPrefixLen         uint8
	Proto                  uint8
	SrcportOrIcmptypeFirst uint16
	SrcportOrIcmptypeLast  uint16
	DstportOrIcmpcodeFirst uint16
	DstportOrIcmpcodeLast  uint16
	TCPFlagsMask           uint8
	TCPFlagsValue          uint8
}

func (*ACLRule) GetTypeName() string {
	return "acl_rule"
}
func (*ACLRule) GetCrcString() string {
	return "2715e1c0"
}

// MacipACLRule represents the VPP binary API data type 'macip_acl_rule'.
// Generated from './bin_api/acl.api.json', line 19:
//
//        ["macip_acl_rule",
//            ["u8", "is_permit"],
//            ["u8", "is_ipv6"],
//            ["u8", "src_mac", 6],
//            ["u8", "src_mac_mask", 6],
//            ["u8", "src_ip_addr", 16],
//            ["u8", "src_ip_prefix_len"],
//            {"crc" : "0x6723f13e"}
//        ]
//
type MacipACLRule struct {
	IsPermit       uint8
	IsIpv6         uint8
	SrcMac         []byte `struc:"[6]byte"`
	SrcMacMask     []byte `struc:"[6]byte"`
	SrcIPAddr      []byte `struc:"[16]byte"`
	SrcIPPrefixLen uint8
}

func (*MacipACLRule) GetTypeName() string {
	return "macip_acl_rule"
}
func (*MacipACLRule) GetCrcString() string {
	return "6723f13e"
}

// ACLPluginGetVersion represents the VPP binary API message 'acl_plugin_get_version'.
// Generated from './bin_api/acl.api.json', line 30:
//
//        ["acl_plugin_get_version",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xd7c07748"}
//        ],
//
type ACLPluginGetVersion struct {
}

func (*ACLPluginGetVersion) GetMessageName() string {
	return "acl_plugin_get_version"
}
func (*ACLPluginGetVersion) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ACLPluginGetVersion) GetCrcString() string {
	return "d7c07748"
}
func NewACLPluginGetVersion() api.Message {
	return &ACLPluginGetVersion{}
}

// ACLPluginGetVersionReply represents the VPP binary API message 'acl_plugin_get_version_reply'.
// Generated from './bin_api/acl.api.json', line 36:
//
//        ["acl_plugin_get_version_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "major"],
//            ["u32", "minor"],
//            {"crc" : "0x43eb59a5"}
//        ],
//
type ACLPluginGetVersionReply struct {
	Major uint32
	Minor uint32
}

func (*ACLPluginGetVersionReply) GetMessageName() string {
	return "acl_plugin_get_version_reply"
}
func (*ACLPluginGetVersionReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ACLPluginGetVersionReply) GetCrcString() string {
	return "43eb59a5"
}
func NewACLPluginGetVersionReply() api.Message {
	return &ACLPluginGetVersionReply{}
}

// ACLAddReplace represents the VPP binary API message 'acl_add_replace'.
// Generated from './bin_api/acl.api.json', line 43:
//
//        ["acl_add_replace",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "acl_index"],
//            ["u8", "tag", 64],
//            ["u32", "count"],
//            ["vl_api_acl_rule_t", "r", 0, "count"],
//            {"crc" : "0x3c317936"}
//        ],
//
type ACLAddReplace struct {
	ACLIndex uint32
	Tag      []byte `struc:"[64]byte"`
	Count    uint32 `struc:"sizeof=R"`
	R        []ACLRule
}

func (*ACLAddReplace) GetMessageName() string {
	return "acl_add_replace"
}
func (*ACLAddReplace) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ACLAddReplace) GetCrcString() string {
	return "3c317936"
}
func NewACLAddReplace() api.Message {
	return &ACLAddReplace{}
}

// ACLAddReplaceReply represents the VPP binary API message 'acl_add_replace_reply'.
// Generated from './bin_api/acl.api.json', line 53:
//
//        ["acl_add_replace_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "acl_index"],
//            ["i32", "retval"],
//            {"crc" : "0xa5e6d0cf"}
//        ],
//
type ACLAddReplaceReply struct {
	ACLIndex uint32
	Retval   int32
}

func (*ACLAddReplaceReply) GetMessageName() string {
	return "acl_add_replace_reply"
}
func (*ACLAddReplaceReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ACLAddReplaceReply) GetCrcString() string {
	return "a5e6d0cf"
}
func NewACLAddReplaceReply() api.Message {
	return &ACLAddReplaceReply{}
}

// ACLDel represents the VPP binary API message 'acl_del'.
// Generated from './bin_api/acl.api.json', line 60:
//
//        ["acl_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "acl_index"],
//            {"crc" : "0x82cc30ed"}
//        ],
//
type ACLDel struct {
	ACLIndex uint32
}

func (*ACLDel) GetMessageName() string {
	return "acl_del"
}
func (*ACLDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ACLDel) GetCrcString() string {
	return "82cc30ed"
}
func NewACLDel() api.Message {
	return &ACLDel{}
}

// ACLDelReply represents the VPP binary API message 'acl_del_reply'.
// Generated from './bin_api/acl.api.json', line 67:
//
//        ["acl_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xbbb83d84"}
//        ],
//
type ACLDelReply struct {
	Retval int32
}

func (*ACLDelReply) GetMessageName() string {
	return "acl_del_reply"
}
func (*ACLDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ACLDelReply) GetCrcString() string {
	return "bbb83d84"
}
func NewACLDelReply() api.Message {
	return &ACLDelReply{}
}

// ACLInterfaceAddDel represents the VPP binary API message 'acl_interface_add_del'.
// Generated from './bin_api/acl.api.json', line 73:
//
//        ["acl_interface_add_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "is_input"],
//            ["u32", "sw_if_index"],
//            ["u32", "acl_index"],
//            {"crc" : "0x98b53725"}
//        ],
//
type ACLInterfaceAddDel struct {
	IsAdd     uint8
	IsInput   uint8
	SwIfIndex uint32
	ACLIndex  uint32
}

func (*ACLInterfaceAddDel) GetMessageName() string {
	return "acl_interface_add_del"
}
func (*ACLInterfaceAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ACLInterfaceAddDel) GetCrcString() string {
	return "98b53725"
}
func NewACLInterfaceAddDel() api.Message {
	return &ACLInterfaceAddDel{}
}

// ACLInterfaceAddDelReply represents the VPP binary API message 'acl_interface_add_del_reply'.
// Generated from './bin_api/acl.api.json', line 83:
//
//        ["acl_interface_add_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xc1b3c077"}
//        ],
//
type ACLInterfaceAddDelReply struct {
	Retval int32
}

func (*ACLInterfaceAddDelReply) GetMessageName() string {
	return "acl_interface_add_del_reply"
}
func (*ACLInterfaceAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ACLInterfaceAddDelReply) GetCrcString() string {
	return "c1b3c077"
}
func NewACLInterfaceAddDelReply() api.Message {
	return &ACLInterfaceAddDelReply{}
}

// ACLInterfaceSetACLList represents the VPP binary API message 'acl_interface_set_acl_list'.
// Generated from './bin_api/acl.api.json', line 89:
//
//        ["acl_interface_set_acl_list",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "count"],
//            ["u8", "n_input"],
//            ["u32", "acls", 0, "count"],
//            {"crc" : "0x7562419c"}
//        ],
//
type ACLInterfaceSetACLList struct {
	SwIfIndex uint32
	Count     uint8 `struc:"sizeof=Acls"`
	NInput    uint8
	Acls      []uint32
}

func (*ACLInterfaceSetACLList) GetMessageName() string {
	return "acl_interface_set_acl_list"
}
func (*ACLInterfaceSetACLList) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ACLInterfaceSetACLList) GetCrcString() string {
	return "7562419c"
}
func NewACLInterfaceSetACLList() api.Message {
	return &ACLInterfaceSetACLList{}
}

// ACLInterfaceSetACLListReply represents the VPP binary API message 'acl_interface_set_acl_list_reply'.
// Generated from './bin_api/acl.api.json', line 99:
//
//        ["acl_interface_set_acl_list_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x435ddc2b"}
//        ],
//
type ACLInterfaceSetACLListReply struct {
	Retval int32
}

func (*ACLInterfaceSetACLListReply) GetMessageName() string {
	return "acl_interface_set_acl_list_reply"
}
func (*ACLInterfaceSetACLListReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ACLInterfaceSetACLListReply) GetCrcString() string {
	return "435ddc2b"
}
func NewACLInterfaceSetACLListReply() api.Message {
	return &ACLInterfaceSetACLListReply{}
}

// ACLDump represents the VPP binary API message 'acl_dump'.
// Generated from './bin_api/acl.api.json', line 105:
//
//        ["acl_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "acl_index"],
//            {"crc" : "0xc188156d"}
//        ],
//
type ACLDump struct {
	ACLIndex uint32
}

func (*ACLDump) GetMessageName() string {
	return "acl_dump"
}
func (*ACLDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ACLDump) GetCrcString() string {
	return "c188156d"
}
func NewACLDump() api.Message {
	return &ACLDump{}
}

// ACLDetails represents the VPP binary API message 'acl_details'.
// Generated from './bin_api/acl.api.json', line 112:
//
//        ["acl_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "acl_index"],
//            ["u8", "tag", 64],
//            ["u32", "count"],
//            ["vl_api_acl_rule_t", "r", 0, "count"],
//            {"crc" : "0x1c8916b7"}
//        ],
//
type ACLDetails struct {
	ACLIndex uint32
	Tag      []byte `struc:"[64]byte"`
	Count    uint32 `struc:"sizeof=R"`
	R        []ACLRule
}

func (*ACLDetails) GetMessageName() string {
	return "acl_details"
}
func (*ACLDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ACLDetails) GetCrcString() string {
	return "1c8916b7"
}
func NewACLDetails() api.Message {
	return &ACLDetails{}
}

// ACLInterfaceListDump represents the VPP binary API message 'acl_interface_list_dump'.
// Generated from './bin_api/acl.api.json', line 121:
//
//        ["acl_interface_list_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0xadfe84b8"}
//        ],
//
type ACLInterfaceListDump struct {
	SwIfIndex uint32
}

func (*ACLInterfaceListDump) GetMessageName() string {
	return "acl_interface_list_dump"
}
func (*ACLInterfaceListDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ACLInterfaceListDump) GetCrcString() string {
	return "adfe84b8"
}
func NewACLInterfaceListDump() api.Message {
	return &ACLInterfaceListDump{}
}

// ACLInterfaceListDetails represents the VPP binary API message 'acl_interface_list_details'.
// Generated from './bin_api/acl.api.json', line 128:
//
//        ["acl_interface_list_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "count"],
//            ["u8", "n_input"],
//            ["u32", "acls", 0, "count"],
//            {"crc" : "0xc8150656"}
//        ],
//
type ACLInterfaceListDetails struct {
	SwIfIndex uint32
	Count     uint8 `struc:"sizeof=Acls"`
	NInput    uint8
	Acls      []uint32
}

func (*ACLInterfaceListDetails) GetMessageName() string {
	return "acl_interface_list_details"
}
func (*ACLInterfaceListDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ACLInterfaceListDetails) GetCrcString() string {
	return "c8150656"
}
func NewACLInterfaceListDetails() api.Message {
	return &ACLInterfaceListDetails{}
}

// MacipACLAdd represents the VPP binary API message 'macip_acl_add'.
// Generated from './bin_api/acl.api.json', line 137:
//
//        ["macip_acl_add",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "tag", 64],
//            ["u32", "count"],
//            ["vl_api_macip_acl_rule_t", "r", 0, "count"],
//            {"crc" : "0x33356284"}
//        ],
//
type MacipACLAdd struct {
	Tag   []byte `struc:"[64]byte"`
	Count uint32 `struc:"sizeof=R"`
	R     []MacipACLRule
}

func (*MacipACLAdd) GetMessageName() string {
	return "macip_acl_add"
}
func (*MacipACLAdd) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MacipACLAdd) GetCrcString() string {
	return "33356284"
}
func NewMacipACLAdd() api.Message {
	return &MacipACLAdd{}
}

// MacipACLAddReply represents the VPP binary API message 'macip_acl_add_reply'.
// Generated from './bin_api/acl.api.json', line 146:
//
//        ["macip_acl_add_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "acl_index"],
//            ["i32", "retval"],
//            {"crc" : "0x472edb4c"}
//        ],
//
type MacipACLAddReply struct {
	ACLIndex uint32
	Retval   int32
}

func (*MacipACLAddReply) GetMessageName() string {
	return "macip_acl_add_reply"
}
func (*MacipACLAddReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*MacipACLAddReply) GetCrcString() string {
	return "472edb4c"
}
func NewMacipACLAddReply() api.Message {
	return &MacipACLAddReply{}
}

// MacipACLDel represents the VPP binary API message 'macip_acl_del'.
// Generated from './bin_api/acl.api.json', line 153:
//
//        ["macip_acl_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "acl_index"],
//            {"crc" : "0xdde1141f"}
//        ],
//
type MacipACLDel struct {
	ACLIndex uint32
}

func (*MacipACLDel) GetMessageName() string {
	return "macip_acl_del"
}
func (*MacipACLDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MacipACLDel) GetCrcString() string {
	return "dde1141f"
}
func NewMacipACLDel() api.Message {
	return &MacipACLDel{}
}

// MacipACLDelReply represents the VPP binary API message 'macip_acl_del_reply'.
// Generated from './bin_api/acl.api.json', line 160:
//
//        ["macip_acl_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xeeb60e0f"}
//        ],
//
type MacipACLDelReply struct {
	Retval int32
}

func (*MacipACLDelReply) GetMessageName() string {
	return "macip_acl_del_reply"
}
func (*MacipACLDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*MacipACLDelReply) GetCrcString() string {
	return "eeb60e0f"
}
func NewMacipACLDelReply() api.Message {
	return &MacipACLDelReply{}
}

// MacipACLInterfaceAddDel represents the VPP binary API message 'macip_acl_interface_add_del'.
// Generated from './bin_api/acl.api.json', line 166:
//
//        ["macip_acl_interface_add_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u32", "sw_if_index"],
//            ["u32", "acl_index"],
//            {"crc" : "0x03a4fab2"}
//        ],
//
type MacipACLInterfaceAddDel struct {
	IsAdd     uint8
	SwIfIndex uint32
	ACLIndex  uint32
}

func (*MacipACLInterfaceAddDel) GetMessageName() string {
	return "macip_acl_interface_add_del"
}
func (*MacipACLInterfaceAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MacipACLInterfaceAddDel) GetCrcString() string {
	return "03a4fab2"
}
func NewMacipACLInterfaceAddDel() api.Message {
	return &MacipACLInterfaceAddDel{}
}

// MacipACLInterfaceAddDelReply represents the VPP binary API message 'macip_acl_interface_add_del_reply'.
// Generated from './bin_api/acl.api.json', line 175:
//
//        ["macip_acl_interface_add_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x9e9ee485"}
//        ],
//
type MacipACLInterfaceAddDelReply struct {
	Retval int32
}

func (*MacipACLInterfaceAddDelReply) GetMessageName() string {
	return "macip_acl_interface_add_del_reply"
}
func (*MacipACLInterfaceAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*MacipACLInterfaceAddDelReply) GetCrcString() string {
	return "9e9ee485"
}
func NewMacipACLInterfaceAddDelReply() api.Message {
	return &MacipACLInterfaceAddDelReply{}
}

// MacipACLDump represents the VPP binary API message 'macip_acl_dump'.
// Generated from './bin_api/acl.api.json', line 181:
//
//        ["macip_acl_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "acl_index"],
//            {"crc" : "0xd38227cb"}
//        ],
//
type MacipACLDump struct {
	ACLIndex uint32
}

func (*MacipACLDump) GetMessageName() string {
	return "macip_acl_dump"
}
func (*MacipACLDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MacipACLDump) GetCrcString() string {
	return "d38227cb"
}
func NewMacipACLDump() api.Message {
	return &MacipACLDump{}
}

// MacipACLDetails represents the VPP binary API message 'macip_acl_details'.
// Generated from './bin_api/acl.api.json', line 188:
//
//        ["macip_acl_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "acl_index"],
//            ["u8", "tag", 64],
//            ["u32", "count"],
//            ["vl_api_macip_acl_rule_t", "r", 0, "count"],
//            {"crc" : "0xee1c50db"}
//        ],
//
type MacipACLDetails struct {
	ACLIndex uint32
	Tag      []byte `struc:"[64]byte"`
	Count    uint32 `struc:"sizeof=R"`
	R        []MacipACLRule
}

func (*MacipACLDetails) GetMessageName() string {
	return "macip_acl_details"
}
func (*MacipACLDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*MacipACLDetails) GetCrcString() string {
	return "ee1c50db"
}
func NewMacipACLDetails() api.Message {
	return &MacipACLDetails{}
}

// MacipACLInterfaceGet represents the VPP binary API message 'macip_acl_interface_get'.
// Generated from './bin_api/acl.api.json', line 197:
//
//        ["macip_acl_interface_get",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x317ce31c"}
//        ],
//
type MacipACLInterfaceGet struct {
}

func (*MacipACLInterfaceGet) GetMessageName() string {
	return "macip_acl_interface_get"
}
func (*MacipACLInterfaceGet) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MacipACLInterfaceGet) GetCrcString() string {
	return "317ce31c"
}
func NewMacipACLInterfaceGet() api.Message {
	return &MacipACLInterfaceGet{}
}

// MacipACLInterfaceGetReply represents the VPP binary API message 'macip_acl_interface_get_reply'.
// Generated from './bin_api/acl.api.json', line 203:
//
//        ["macip_acl_interface_get_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "count"],
//            ["u32", "acls", 0, "count"],
//            {"crc" : "0x6c86a56c"}
//        ]
//
type MacipACLInterfaceGetReply struct {
	Count uint32 `struc:"sizeof=Acls"`
	Acls  []uint32
}

func (*MacipACLInterfaceGetReply) GetMessageName() string {
	return "macip_acl_interface_get_reply"
}
func (*MacipACLInterfaceGetReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*MacipACLInterfaceGetReply) GetCrcString() string {
	return "6c86a56c"
}
func NewMacipACLInterfaceGetReply() api.Message {
	return &MacipACLInterfaceGetReply{}
}
