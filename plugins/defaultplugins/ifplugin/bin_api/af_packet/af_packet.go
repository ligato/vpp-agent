// Package af_packet represents the VPP binary API of the 'af_packet' VPP module.
// DO NOT EDIT. Generated from '/usr/share/vpp/api/af_packet.api.json'
package af_packet

import "git.fd.io/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0xd4ce9f85

// AfPacketCreate represents the VPP binary API message 'af_packet_create'.
// Generated from '/usr/share/vpp/api/af_packet.api.json', line 6:
//
//        ["af_packet_create",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "host_if_name", 64],
//            ["u8", "hw_addr", 6],
//            ["u8", "use_random_hw_addr"],
//            {"crc" : "0x92768640"}
//        ],
//
type AfPacketCreate struct {
	HostIfName      []byte `struc:"[64]byte"`
	HwAddr          []byte `struc:"[6]byte"`
	UseRandomHwAddr uint8
}

func (*AfPacketCreate) GetMessageName() string {
	return "af_packet_create"
}
func (*AfPacketCreate) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*AfPacketCreate) GetCrcString() string {
	return "92768640"
}
func NewAfPacketCreate() api.Message {
	return &AfPacketCreate{}
}

// AfPacketCreateReply represents the VPP binary API message 'af_packet_create_reply'.
// Generated from '/usr/share/vpp/api/af_packet.api.json', line 15:
//
//        ["af_packet_create_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x718bac92"}
//        ],
//
type AfPacketCreateReply struct {
	Retval    int32
	SwIfIndex uint32
}

func (*AfPacketCreateReply) GetMessageName() string {
	return "af_packet_create_reply"
}
func (*AfPacketCreateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*AfPacketCreateReply) GetCrcString() string {
	return "718bac92"
}
func NewAfPacketCreateReply() api.Message {
	return &AfPacketCreateReply{}
}

// AfPacketDelete represents the VPP binary API message 'af_packet_delete'.
// Generated from '/usr/share/vpp/api/af_packet.api.json', line 22:
//
//        ["af_packet_delete",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "host_if_name", 64],
//            {"crc" : "0xc063ce85"}
//        ],
//
type AfPacketDelete struct {
	HostIfName []byte `struc:"[64]byte"`
}

func (*AfPacketDelete) GetMessageName() string {
	return "af_packet_delete"
}
func (*AfPacketDelete) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*AfPacketDelete) GetCrcString() string {
	return "c063ce85"
}
func NewAfPacketDelete() api.Message {
	return &AfPacketDelete{}
}

// AfPacketDeleteReply represents the VPP binary API message 'af_packet_delete_reply'.
// Generated from '/usr/share/vpp/api/af_packet.api.json', line 29:
//
//        ["af_packet_delete_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x1a80431a"}
//        ]
//
type AfPacketDeleteReply struct {
	Retval int32
}

func (*AfPacketDeleteReply) GetMessageName() string {
	return "af_packet_delete_reply"
}
func (*AfPacketDeleteReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*AfPacketDeleteReply) GetCrcString() string {
	return "1a80431a"
}
func NewAfPacketDeleteReply() api.Message {
	return &AfPacketDeleteReply{}
}
