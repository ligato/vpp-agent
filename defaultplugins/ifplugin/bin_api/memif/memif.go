// Package memif represents the VPP binary API of the 'memif' VPP module.
// DO NOT EDIT. Generated from '/usr/share/vpp/api/memif.api.json'
package memif

import "git.fd.io/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0xed3def5d

// MemifCreate represents the VPP binary API message 'memif_create'.
// Generated from '/usr/share/vpp/api/memif.api.json', line 6:
//
//        ["memif_create",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "role"],
//            ["u8", "mode"],
//            ["u8", "rx_queues"],
//            ["u8", "tx_queues"],
//            ["u32", "id"],
//            ["u8", "socket_filename", 128],
//            ["u8", "secret", 24],
//            ["u32", "ring_size"],
//            ["u16", "buffer_size"],
//            ["u8", "hw_addr", 6],
//            {"crc" : "0xc809c235"}
//        ],
//
type MemifCreate struct {
	Role           uint8
	Mode           uint8
	RxQueues       uint8
	TxQueues       uint8
	ID             uint32
	SocketFilename []byte `struc:"[128]byte"`
	Secret         []byte `struc:"[24]byte"`
	RingSize       uint32
	BufferSize     uint16
	HwAddr         []byte `struc:"[6]byte"`
}

func (*MemifCreate) GetMessageName() string {
	return "memif_create"
}
func (*MemifCreate) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MemifCreate) GetCrcString() string {
	return "c809c235"
}
func NewMemifCreate() api.Message {
	return &MemifCreate{}
}

// MemifCreateReply represents the VPP binary API message 'memif_create_reply'.
// Generated from '/usr/share/vpp/api/memif.api.json', line 22:
//
//        ["memif_create_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x93d7498b"}
//        ],
//
type MemifCreateReply struct {
	Retval    int32
	SwIfIndex uint32
}

func (*MemifCreateReply) GetMessageName() string {
	return "memif_create_reply"
}
func (*MemifCreateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*MemifCreateReply) GetCrcString() string {
	return "93d7498b"
}
func NewMemifCreateReply() api.Message {
	return &MemifCreateReply{}
}

// MemifDelete represents the VPP binary API message 'memif_delete'.
// Generated from '/usr/share/vpp/api/memif.api.json', line 29:
//
//        ["memif_delete",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x12814e3d"}
//        ],
//
type MemifDelete struct {
	SwIfIndex uint32
}

func (*MemifDelete) GetMessageName() string {
	return "memif_delete"
}
func (*MemifDelete) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MemifDelete) GetCrcString() string {
	return "12814e3d"
}
func NewMemifDelete() api.Message {
	return &MemifDelete{}
}

// MemifDeleteReply represents the VPP binary API message 'memif_delete_reply'.
// Generated from '/usr/share/vpp/api/memif.api.json', line 36:
//
//        ["memif_delete_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x72c9fa3c"}
//        ],
//
type MemifDeleteReply struct {
	Retval int32
}

func (*MemifDeleteReply) GetMessageName() string {
	return "memif_delete_reply"
}
func (*MemifDeleteReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*MemifDeleteReply) GetCrcString() string {
	return "72c9fa3c"
}
func NewMemifDeleteReply() api.Message {
	return &MemifDeleteReply{}
}

// MemifDetails represents the VPP binary API message 'memif_details'.
// Generated from '/usr/share/vpp/api/memif.api.json', line 42:
//
//        ["memif_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "if_name", 64],
//            ["u8", "hw_addr", 6],
//            ["u32", "id"],
//            ["u8", "role"],
//            ["u8", "mode"],
//            ["u8", "socket_filename", 128],
//            ["u32", "ring_size"],
//            ["u16", "buffer_size"],
//            ["u8", "admin_up_down"],
//            ["u8", "link_up_down"],
//            {"crc" : "0x4d74fd96"}
//        ],
//
type MemifDetails struct {
	SwIfIndex      uint32
	IfName         []byte `struc:"[64]byte"`
	HwAddr         []byte `struc:"[6]byte"`
	ID             uint32
	Role           uint8
	Mode           uint8
	SocketFilename []byte `struc:"[128]byte"`
	RingSize       uint32
	BufferSize     uint16
	AdminUpDown    uint8
	LinkUpDown     uint8
}

func (*MemifDetails) GetMessageName() string {
	return "memif_details"
}
func (*MemifDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*MemifDetails) GetCrcString() string {
	return "4d74fd96"
}
func NewMemifDetails() api.Message {
	return &MemifDetails{}
}

// MemifDump represents the VPP binary API message 'memif_dump'.
// Generated from '/usr/share/vpp/api/memif.api.json', line 58:
//
//        ["memif_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x68d39e95"}
//        ]
//
type MemifDump struct {
}

func (*MemifDump) GetMessageName() string {
	return "memif_dump"
}
func (*MemifDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MemifDump) GetCrcString() string {
	return "68d39e95"
}
func NewMemifDump() api.Message {
	return &MemifDump{}
}
