// Package session represents the VPP binary API of the 'session' VPP module.
// DO NOT EDIT. Generated from '/usr/share/vpp/api/session.api.json'
package session

import "git.fd.io/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0x15333a72

// ApplicationAttach represents the VPP binary API message 'application_attach'.
// Generated from '/usr/share/vpp/api/session.api.json', line 6:
//
//        ["application_attach",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "initial_segment_size"],
//            ["u64", "options", 16],
//            ["u8", "namespace_id_len"],
//            ["u8", "namespace_id", 64],
//            {"crc" : "0xf27e0a47"}
//        ],
//
type ApplicationAttach struct {
	InitialSegmentSize uint32
	Options            []uint64 `struc:"[16]uint64"`
	NamespaceIDLen     uint8
	NamespaceID        []byte `struc:"[64]byte"`
}

func (*ApplicationAttach) GetMessageName() string {
	return "application_attach"
}
func (*ApplicationAttach) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ApplicationAttach) GetCrcString() string {
	return "f27e0a47"
}
func NewApplicationAttach() api.Message {
	return &ApplicationAttach{}
}

// ApplicationAttachReply represents the VPP binary API message 'application_attach_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 16:
//
//        ["application_attach_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u64", "app_event_queue_address"],
//            ["u32", "segment_size"],
//            ["u8", "segment_name_length"],
//            ["u8", "segment_name", 128],
//            {"crc" : "0x0df5c138"}
//        ],
//
type ApplicationAttachReply struct {
	Retval               int32
	AppEventQueueAddress uint64
	SegmentSize          uint32
	SegmentNameLength    uint8
	SegmentName          []byte `struc:"[128]byte"`
}

func (*ApplicationAttachReply) GetMessageName() string {
	return "application_attach_reply"
}
func (*ApplicationAttachReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ApplicationAttachReply) GetCrcString() string {
	return "0df5c138"
}
func NewApplicationAttachReply() api.Message {
	return &ApplicationAttachReply{}
}

// ApplicationDetach represents the VPP binary API message 'application_detach'.
// Generated from '/usr/share/vpp/api/session.api.json', line 26:
//
//        ["application_detach",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xbf7e4352"}
//        ],
//
type ApplicationDetach struct {
}

func (*ApplicationDetach) GetMessageName() string {
	return "application_detach"
}
func (*ApplicationDetach) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ApplicationDetach) GetCrcString() string {
	return "bf7e4352"
}
func NewApplicationDetach() api.Message {
	return &ApplicationDetach{}
}

// ApplicationDetachReply represents the VPP binary API message 'application_detach_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 32:
//
//        ["application_detach_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xfb879289"}
//        ],
//
type ApplicationDetachReply struct {
	Retval int32
}

func (*ApplicationDetachReply) GetMessageName() string {
	return "application_detach_reply"
}
func (*ApplicationDetachReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ApplicationDetachReply) GetCrcString() string {
	return "fb879289"
}
func NewApplicationDetachReply() api.Message {
	return &ApplicationDetachReply{}
}

// MapAnotherSegment represents the VPP binary API message 'map_another_segment'.
// Generated from '/usr/share/vpp/api/session.api.json', line 38:
//
//        ["map_another_segment",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "segment_size"],
//            ["u8", "segment_name", 128],
//            {"crc" : "0x28ca2003"}
//        ],
//
type MapAnotherSegment struct {
	SegmentSize uint32
	SegmentName []byte `struc:"[128]byte"`
}

func (*MapAnotherSegment) GetMessageName() string {
	return "map_another_segment"
}
func (*MapAnotherSegment) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MapAnotherSegment) GetCrcString() string {
	return "28ca2003"
}
func NewMapAnotherSegment() api.Message {
	return &MapAnotherSegment{}
}

// MapAnotherSegmentReply represents the VPP binary API message 'map_another_segment_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 46:
//
//        ["map_another_segment_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x76d11a9d"}
//        ],
//
type MapAnotherSegmentReply struct {
	Retval int32
}

func (*MapAnotherSegmentReply) GetMessageName() string {
	return "map_another_segment_reply"
}
func (*MapAnotherSegmentReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*MapAnotherSegmentReply) GetCrcString() string {
	return "76d11a9d"
}
func NewMapAnotherSegmentReply() api.Message {
	return &MapAnotherSegmentReply{}
}

// BindURI represents the VPP binary API message 'bind_uri'.
// Generated from '/usr/share/vpp/api/session.api.json', line 52:
//
//        ["bind_uri",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "accept_cookie"],
//            ["u8", "uri", 128],
//            {"crc" : "0xceafed7f"}
//        ],
//
type BindURI struct {
	AcceptCookie uint32
	URI          []byte `struc:"[128]byte"`
}

func (*BindURI) GetMessageName() string {
	return "bind_uri"
}
func (*BindURI) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BindURI) GetCrcString() string {
	return "ceafed7f"
}
func NewBindURI() api.Message {
	return &BindURI{}
}

// BindURIReply represents the VPP binary API message 'bind_uri_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 60:
//
//        ["bind_uri_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x75918978"}
//        ],
//
type BindURIReply struct {
	Retval int32
}

func (*BindURIReply) GetMessageName() string {
	return "bind_uri_reply"
}
func (*BindURIReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BindURIReply) GetCrcString() string {
	return "75918978"
}
func NewBindURIReply() api.Message {
	return &BindURIReply{}
}

// UnbindURI represents the VPP binary API message 'unbind_uri'.
// Generated from '/usr/share/vpp/api/session.api.json', line 66:
//
//        ["unbind_uri",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "uri", 128],
//            {"crc" : "0x46569743"}
//        ],
//
type UnbindURI struct {
	URI []byte `struc:"[128]byte"`
}

func (*UnbindURI) GetMessageName() string {
	return "unbind_uri"
}
func (*UnbindURI) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*UnbindURI) GetCrcString() string {
	return "46569743"
}
func NewUnbindURI() api.Message {
	return &UnbindURI{}
}

// UnbindURIReply represents the VPP binary API message 'unbind_uri_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 73:
//
//        ["unbind_uri_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x310db78f"}
//        ],
//
type UnbindURIReply struct {
	Retval int32
}

func (*UnbindURIReply) GetMessageName() string {
	return "unbind_uri_reply"
}
func (*UnbindURIReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*UnbindURIReply) GetCrcString() string {
	return "310db78f"
}
func NewUnbindURIReply() api.Message {
	return &UnbindURIReply{}
}

// ConnectURI represents the VPP binary API message 'connect_uri'.
// Generated from '/usr/share/vpp/api/session.api.json', line 79:
//
//        ["connect_uri",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u64", "client_queue_address"],
//            ["u64", "options", 16],
//            ["u8", "uri", 128],
//            {"crc" : "0x72035237"}
//        ],
//
type ConnectURI struct {
	ClientQueueAddress uint64
	Options            []uint64 `struc:"[16]uint64"`
	URI                []byte   `struc:"[128]byte"`
}

func (*ConnectURI) GetMessageName() string {
	return "connect_uri"
}
func (*ConnectURI) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ConnectURI) GetCrcString() string {
	return "72035237"
}
func NewConnectURI() api.Message {
	return &ConnectURI{}
}

// ConnectURIReply represents the VPP binary API message 'connect_uri_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 88:
//
//        ["connect_uri_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x54ec1256"}
//        ],
//
type ConnectURIReply struct {
	Retval int32
}

func (*ConnectURIReply) GetMessageName() string {
	return "connect_uri_reply"
}
func (*ConnectURIReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ConnectURIReply) GetCrcString() string {
	return "54ec1256"
}
func NewConnectURIReply() api.Message {
	return &ConnectURIReply{}
}

// AcceptSession represents the VPP binary API message 'accept_session'.
// Generated from '/usr/share/vpp/api/session.api.json', line 94:
//
//        ["accept_session",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u64", "listener_handle"],
//            ["u64", "handle"],
//            ["u64", "server_rx_fifo"],
//            ["u64", "server_tx_fifo"],
//            ["u64", "vpp_event_queue_address"],
//            ["u16", "port"],
//            ["u8", "is_ip4"],
//            ["u8", "ip", 16],
//            {"crc" : "0x8e2a127e"}
//        ],
//
type AcceptSession struct {
	ListenerHandle       uint64
	Handle               uint64
	ServerRxFifo         uint64
	ServerTxFifo         uint64
	VppEventQueueAddress uint64
	Port                 uint16
	IsIP4                uint8
	IP                   []byte `struc:"[16]byte"`
}

func (*AcceptSession) GetMessageName() string {
	return "accept_session"
}
func (*AcceptSession) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*AcceptSession) GetCrcString() string {
	return "8e2a127e"
}
func NewAcceptSession() api.Message {
	return &AcceptSession{}
}

// AcceptSessionReply represents the VPP binary API message 'accept_session_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 108:
//
//        ["accept_session_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u64", "handle"],
//            {"crc" : "0x67d8c22a"}
//        ],
//
type AcceptSessionReply struct {
	Retval int32
	Handle uint64
}

func (*AcceptSessionReply) GetMessageName() string {
	return "accept_session_reply"
}
func (*AcceptSessionReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*AcceptSessionReply) GetCrcString() string {
	return "67d8c22a"
}
func NewAcceptSessionReply() api.Message {
	return &AcceptSessionReply{}
}

// DisconnectSession represents the VPP binary API message 'disconnect_session'.
// Generated from '/usr/share/vpp/api/session.api.json', line 115:
//
//        ["disconnect_session",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u64", "handle"],
//            {"crc" : "0x18addf61"}
//        ],
//
type DisconnectSession struct {
	Handle uint64
}

func (*DisconnectSession) GetMessageName() string {
	return "disconnect_session"
}
func (*DisconnectSession) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*DisconnectSession) GetCrcString() string {
	return "18addf61"
}
func NewDisconnectSession() api.Message {
	return &DisconnectSession{}
}

// DisconnectSessionReply represents the VPP binary API message 'disconnect_session_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 122:
//
//        ["disconnect_session_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u64", "handle"],
//            {"crc" : "0x6fb16b8f"}
//        ],
//
type DisconnectSessionReply struct {
	Retval int32
	Handle uint64
}

func (*DisconnectSessionReply) GetMessageName() string {
	return "disconnect_session_reply"
}
func (*DisconnectSessionReply) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*DisconnectSessionReply) GetCrcString() string {
	return "6fb16b8f"
}
func NewDisconnectSessionReply() api.Message {
	return &DisconnectSessionReply{}
}

// ResetSession represents the VPP binary API message 'reset_session'.
// Generated from '/usr/share/vpp/api/session.api.json', line 130:
//
//        ["reset_session",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u64", "handle"],
//            {"crc" : "0x601fefd7"}
//        ],
//
type ResetSession struct {
	Handle uint64
}

func (*ResetSession) GetMessageName() string {
	return "reset_session"
}
func (*ResetSession) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ResetSession) GetCrcString() string {
	return "601fefd7"
}
func NewResetSession() api.Message {
	return &ResetSession{}
}

// ResetSessionReply represents the VPP binary API message 'reset_session_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 137:
//
//        ["reset_session_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u64", "handle"],
//            {"crc" : "0x80f6c14f"}
//        ],
//
type ResetSessionReply struct {
	Retval int32
	Handle uint64
}

func (*ResetSessionReply) GetMessageName() string {
	return "reset_session_reply"
}
func (*ResetSessionReply) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ResetSessionReply) GetCrcString() string {
	return "80f6c14f"
}
func NewResetSessionReply() api.Message {
	return &ResetSessionReply{}
}

// BindSock represents the VPP binary API message 'bind_sock'.
// Generated from '/usr/share/vpp/api/session.api.json', line 145:
//
//        ["bind_sock",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "vrf"],
//            ["u8", "is_ip4"],
//            ["u8", "ip", 16],
//            ["u16", "port"],
//            ["u8", "proto"],
//            ["u64", "options", 16],
//            {"crc" : "0x3f898291"}
//        ],
//
type BindSock struct {
	Vrf     uint32
	IsIP4   uint8
	IP      []byte `struc:"[16]byte"`
	Port    uint16
	Proto   uint8
	Options []uint64 `struc:"[16]uint64"`
}

func (*BindSock) GetMessageName() string {
	return "bind_sock"
}
func (*BindSock) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BindSock) GetCrcString() string {
	return "3f898291"
}
func NewBindSock() api.Message {
	return &BindSock{}
}

// UnbindSock represents the VPP binary API message 'unbind_sock'.
// Generated from '/usr/share/vpp/api/session.api.json', line 157:
//
//        ["unbind_sock",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u64", "handle"],
//            {"crc" : "0x9007c8c9"}
//        ],
//
type UnbindSock struct {
	Handle uint64
}

func (*UnbindSock) GetMessageName() string {
	return "unbind_sock"
}
func (*UnbindSock) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*UnbindSock) GetCrcString() string {
	return "9007c8c9"
}
func NewUnbindSock() api.Message {
	return &UnbindSock{}
}

// UnbindSockReply represents the VPP binary API message 'unbind_sock_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 164:
//
//        ["unbind_sock_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x5d9c5da6"}
//        ],
//
type UnbindSockReply struct {
	Retval int32
}

func (*UnbindSockReply) GetMessageName() string {
	return "unbind_sock_reply"
}
func (*UnbindSockReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*UnbindSockReply) GetCrcString() string {
	return "5d9c5da6"
}
func NewUnbindSockReply() api.Message {
	return &UnbindSockReply{}
}

// ConnectSock represents the VPP binary API message 'connect_sock'.
// Generated from '/usr/share/vpp/api/session.api.json', line 170:
//
//        ["connect_sock",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u64", "client_queue_address"],
//            ["u64", "options", 16],
//            ["u32", "vrf"],
//            ["u8", "is_ip4"],
//            ["u8", "ip", 16],
//            ["u16", "port"],
//            ["u8", "proto"],
//            {"crc" : "0x0bf00e15"}
//        ],
//
type ConnectSock struct {
	ClientQueueAddress uint64
	Options            []uint64 `struc:"[16]uint64"`
	Vrf                uint32
	IsIP4              uint8
	IP                 []byte `struc:"[16]byte"`
	Port               uint16
	Proto              uint8
}

func (*ConnectSock) GetMessageName() string {
	return "connect_sock"
}
func (*ConnectSock) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ConnectSock) GetCrcString() string {
	return "0bf00e15"
}
func NewConnectSock() api.Message {
	return &ConnectSock{}
}

// ConnectSockReply represents the VPP binary API message 'connect_sock_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 183:
//
//        ["connect_sock_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xf6988664"}
//        ],
//
type ConnectSockReply struct {
	Retval int32
}

func (*ConnectSockReply) GetMessageName() string {
	return "connect_sock_reply"
}
func (*ConnectSockReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ConnectSockReply) GetCrcString() string {
	return "f6988664"
}
func NewConnectSockReply() api.Message {
	return &ConnectSockReply{}
}

// BindSockReply represents the VPP binary API message 'bind_sock_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 189:
//
//        ["bind_sock_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u64", "handle"],
//            ["i32", "retval"],
//            ["u64", "server_event_queue_address"],
//            ["u32", "segment_size"],
//            ["u8", "segment_name_length"],
//            ["u8", "segment_name", 128],
//            {"crc" : "0xeecef9cc"}
//        ],
//
type BindSockReply struct {
	Handle                  uint64
	Retval                  int32
	ServerEventQueueAddress uint64
	SegmentSize             uint32
	SegmentNameLength       uint8
	SegmentName             []byte `struc:"[128]byte"`
}

func (*BindSockReply) GetMessageName() string {
	return "bind_sock_reply"
}
func (*BindSockReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BindSockReply) GetCrcString() string {
	return "eecef9cc"
}
func NewBindSockReply() api.Message {
	return &BindSockReply{}
}

// ConnectSession represents the VPP binary API message 'connect_session'.
// Generated from '/usr/share/vpp/api/session.api.json', line 200:
//
//        ["connect_session",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x40ae01d1"}
//        ],
//
type ConnectSession struct {
}

func (*ConnectSession) GetMessageName() string {
	return "connect_session"
}
func (*ConnectSession) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*ConnectSession) GetCrcString() string {
	return "40ae01d1"
}
func NewConnectSession() api.Message {
	return &ConnectSession{}
}

// ConnectSessionReply represents the VPP binary API message 'connect_session_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 206:
//
//        ["connect_session_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u64", "handle"],
//            ["u64", "server_rx_fifo"],
//            ["u64", "server_tx_fifo"],
//            ["u64", "vpp_event_queue_address"],
//            ["u32", "segment_size"],
//            ["u8", "segment_name_length"],
//            ["u8", "segment_name", 128],
//            ["u8", "lcl_ip", 16],
//            ["u8", "is_ip4"],
//            ["u16", "lcl_port"],
//            {"crc" : "0x0bf3fdc2"}
//        ],
//
type ConnectSessionReply struct {
	Retval               int32
	Handle               uint64
	ServerRxFifo         uint64
	ServerTxFifo         uint64
	VppEventQueueAddress uint64
	SegmentSize          uint32
	SegmentNameLength    uint8
	SegmentName          []byte `struc:"[128]byte"`
	LclIP                []byte `struc:"[16]byte"`
	IsIP4                uint8
	LclPort              uint16
}

func (*ConnectSessionReply) GetMessageName() string {
	return "connect_session_reply"
}
func (*ConnectSessionReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*ConnectSessionReply) GetCrcString() string {
	return "0bf3fdc2"
}
func NewConnectSessionReply() api.Message {
	return &ConnectSessionReply{}
}

// SessionEnableDisable represents the VPP binary API message 'session_enable_disable'.
// Generated from '/usr/share/vpp/api/session.api.json', line 222:
//
//        ["session_enable_disable",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_enable"],
//            {"crc" : "0xa4cfced4"}
//        ],
//
type SessionEnableDisable struct {
	IsEnable uint8
}

func (*SessionEnableDisable) GetMessageName() string {
	return "session_enable_disable"
}
func (*SessionEnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SessionEnableDisable) GetCrcString() string {
	return "a4cfced4"
}
func NewSessionEnableDisable() api.Message {
	return &SessionEnableDisable{}
}

// SessionEnableDisableReply represents the VPP binary API message 'session_enable_disable_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 229:
//
//        ["session_enable_disable_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xcfb0e390"}
//        ],
//
type SessionEnableDisableReply struct {
	Retval int32
}

func (*SessionEnableDisableReply) GetMessageName() string {
	return "session_enable_disable_reply"
}
func (*SessionEnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SessionEnableDisableReply) GetCrcString() string {
	return "cfb0e390"
}
func NewSessionEnableDisableReply() api.Message {
	return &SessionEnableDisableReply{}
}

// AppNamespaceAddDel represents the VPP binary API message 'app_namespace_add_del'.
// Generated from '/usr/share/vpp/api/session.api.json', line 235:
//
//        ["app_namespace_add_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u64", "secret"],
//            ["u32", "sw_if_index"],
//            ["u32", "ip4_fib_id"],
//            ["u32", "ip6_fib_id"],
//            ["u8", "namespace_id_len"],
//            ["u8", "namespace_id", 64],
//            {"crc" : "0x9733ef2a"}
//        ],
//
type AppNamespaceAddDel struct {
	Secret         uint64
	SwIfIndex      uint32
	IP4FibID       uint32
	IP6FibID       uint32
	NamespaceIDLen uint8
	NamespaceID    []byte `struc:"[64]byte"`
}

func (*AppNamespaceAddDel) GetMessageName() string {
	return "app_namespace_add_del"
}
func (*AppNamespaceAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*AppNamespaceAddDel) GetCrcString() string {
	return "9733ef2a"
}
func NewAppNamespaceAddDel() api.Message {
	return &AppNamespaceAddDel{}
}

// AppNamespaceAddDelReply represents the VPP binary API message 'app_namespace_add_del_reply'.
// Generated from '/usr/share/vpp/api/session.api.json', line 247:
//
//        ["app_namespace_add_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x7c9b7424"}
//        ]
//
type AppNamespaceAddDelReply struct {
	Retval int32
}

func (*AppNamespaceAddDelReply) GetMessageName() string {
	return "app_namespace_add_del_reply"
}
func (*AppNamespaceAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*AppNamespaceAddDelReply) GetCrcString() string {
	return "7c9b7424"
}
func NewAppNamespaceAddDelReply() api.Message {
	return &AppNamespaceAddDelReply{}
}
