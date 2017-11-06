// Code generated by govpp binapi-generator DO NOT EDIT.
// Package stn represents the VPP binary API of the 'stn' VPP module.
// Generated from 'stn.api.json'
package stn
import "git.fd.io/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion =  0xe5fdd9f7


// StnAddDelRule represents the VPP binary API message 'stn_add_del_rule'.
// Generated from 'stn.api.json', line 6:
//
//        ["stn_add_del_rule",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_ip4"],
//            ["u8", "ip_address", 16],
//            ["u32", "sw_if_index"],
//            ["u8", "is_add"],
//            {"crc" : "0x4a761a12"}
//        ],
//
type StnAddDelRule struct {
	IsIP4 uint8
	IPAddress []byte	`struc:"[16]byte"`
	SwIfIndex uint32
	IsAdd uint8
}
func (*StnAddDelRule) GetMessageName() string {
	return "stn_add_del_rule"
}
func (*StnAddDelRule) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*StnAddDelRule) GetCrcString() string {
	return "4a761a12"
}
func NewStnAddDelRule() api.Message {
	return &StnAddDelRule{}
}

// StnAddDelRuleReply represents the VPP binary API message 'stn_add_del_rule_reply'.
// Generated from 'stn.api.json', line 16:
//
//        ["stn_add_del_rule_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x7af79b37"}
//        ],
//
type StnAddDelRuleReply struct {
	Retval int32
}
func (*StnAddDelRuleReply) GetMessageName() string {
	return "stn_add_del_rule_reply"
}
func (*StnAddDelRuleReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*StnAddDelRuleReply) GetCrcString() string {
	return "7af79b37"
}
func NewStnAddDelRuleReply() api.Message {
	return &StnAddDelRuleReply{}
}

// StnRulesDump represents the VPP binary API message 'stn_rules_dump'.
// Generated from 'stn.api.json', line 22:
//
//        ["stn_rules_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xe3b863a5"}
//        ],
//
type StnRulesDump struct {
}
func (*StnRulesDump) GetMessageName() string {
	return "stn_rules_dump"
}
func (*StnRulesDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*StnRulesDump) GetCrcString() string {
	return "e3b863a5"
}
func NewStnRulesDump() api.Message {
	return &StnRulesDump{}
}

// StnRuleDetails represents the VPP binary API message 'stn_rule_details'.
// Generated from 'stn.api.json', line 28:
//
//        ["stn_rule_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "is_ip4"],
//            ["u8", "ip_address", 16],
//            ["u32", "sw_if_index"],
//            {"crc" : "0xaf614822"}
//        ]
//
type StnRuleDetails struct {
	IsIP4 uint8
	IPAddress []byte	`struc:"[16]byte"`
	SwIfIndex uint32
}
func (*StnRuleDetails) GetMessageName() string {
	return "stn_rule_details"
}
func (*StnRuleDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*StnRuleDetails) GetCrcString() string {
	return "af614822"
}
func NewStnRuleDetails() api.Message {
	return &StnRuleDetails{}
}
