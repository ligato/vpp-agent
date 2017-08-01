// Package bfd represents the VPP binary API of the 'bfd' VPP module.
// DO NOT EDIT. Generated from '/usr/share/vpp/api/bfd.api.json'
package bfd

import "git.fd.io/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0xdd31bf39

// BfdUDPSetEchoSource represents the VPP binary API message 'bfd_udp_set_echo_source'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 6:
//
//        ["bfd_udp_set_echo_source",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x928d862a"}
//        ],
//
type BfdUDPSetEchoSource struct {
	SwIfIndex uint32
}

func (*BfdUDPSetEchoSource) GetMessageName() string {
	return "bfd_udp_set_echo_source"
}
func (*BfdUDPSetEchoSource) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdUDPSetEchoSource) GetCrcString() string {
	return "928d862a"
}
func NewBfdUDPSetEchoSource() api.Message {
	return &BfdUDPSetEchoSource{}
}

// BfdUDPSetEchoSourceReply represents the VPP binary API message 'bfd_udp_set_echo_source_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 13:
//
//        ["bfd_udp_set_echo_source_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xc7700775"}
//        ],
//
type BfdUDPSetEchoSourceReply struct {
	Retval int32
}

func (*BfdUDPSetEchoSourceReply) GetMessageName() string {
	return "bfd_udp_set_echo_source_reply"
}
func (*BfdUDPSetEchoSourceReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdUDPSetEchoSourceReply) GetCrcString() string {
	return "c7700775"
}
func NewBfdUDPSetEchoSourceReply() api.Message {
	return &BfdUDPSetEchoSourceReply{}
}

// BfdUDPDelEchoSource represents the VPP binary API message 'bfd_udp_del_echo_source'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 19:
//
//        ["bfd_udp_del_echo_source",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x2757531c"}
//        ],
//
type BfdUDPDelEchoSource struct {
}

func (*BfdUDPDelEchoSource) GetMessageName() string {
	return "bfd_udp_del_echo_source"
}
func (*BfdUDPDelEchoSource) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdUDPDelEchoSource) GetCrcString() string {
	return "2757531c"
}
func NewBfdUDPDelEchoSource() api.Message {
	return &BfdUDPDelEchoSource{}
}

// BfdUDPDelEchoSourceReply represents the VPP binary API message 'bfd_udp_del_echo_source_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 25:
//
//        ["bfd_udp_del_echo_source_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x63ae82c7"}
//        ],
//
type BfdUDPDelEchoSourceReply struct {
	Retval int32
}

func (*BfdUDPDelEchoSourceReply) GetMessageName() string {
	return "bfd_udp_del_echo_source_reply"
}
func (*BfdUDPDelEchoSourceReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdUDPDelEchoSourceReply) GetCrcString() string {
	return "63ae82c7"
}
func NewBfdUDPDelEchoSourceReply() api.Message {
	return &BfdUDPDelEchoSourceReply{}
}

// BfdUDPAdd represents the VPP binary API message 'bfd_udp_add'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 31:
//
//        ["bfd_udp_add",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u32", "desired_min_tx"],
//            ["u32", "required_min_rx"],
//            ["u8", "local_addr", 16],
//            ["u8", "peer_addr", 16],
//            ["u8", "is_ipv6"],
//            ["u8", "detect_mult"],
//            ["u8", "is_authenticated"],
//            ["u8", "bfd_key_id"],
//            ["u32", "conf_key_id"],
//            {"crc" : "0x5fe67640"}
//        ],
//
type BfdUDPAdd struct {
	SwIfIndex       uint32
	DesiredMinTx    uint32
	RequiredMinRx   uint32
	LocalAddr       []byte `struc:"[16]byte"`
	PeerAddr        []byte `struc:"[16]byte"`
	IsIpv6          uint8
	DetectMult      uint8
	IsAuthenticated uint8
	BfdKeyID        uint8
	ConfKeyID       uint32
}

func (*BfdUDPAdd) GetMessageName() string {
	return "bfd_udp_add"
}
func (*BfdUDPAdd) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdUDPAdd) GetCrcString() string {
	return "5fe67640"
}
func NewBfdUDPAdd() api.Message {
	return &BfdUDPAdd{}
}

// BfdUDPAddReply represents the VPP binary API message 'bfd_udp_add_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 47:
//
//        ["bfd_udp_add_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x95013fb7"}
//        ],
//
type BfdUDPAddReply struct {
	Retval int32
}

func (*BfdUDPAddReply) GetMessageName() string {
	return "bfd_udp_add_reply"
}
func (*BfdUDPAddReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdUDPAddReply) GetCrcString() string {
	return "95013fb7"
}
func NewBfdUDPAddReply() api.Message {
	return &BfdUDPAddReply{}
}

// BfdUDPMod represents the VPP binary API message 'bfd_udp_mod'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 53:
//
//        ["bfd_udp_mod",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u32", "desired_min_tx"],
//            ["u32", "required_min_rx"],
//            ["u8", "local_addr", 16],
//            ["u8", "peer_addr", 16],
//            ["u8", "is_ipv6"],
//            ["u8", "detect_mult"],
//            {"crc" : "0xcee1341e"}
//        ],
//
type BfdUDPMod struct {
	SwIfIndex     uint32
	DesiredMinTx  uint32
	RequiredMinRx uint32
	LocalAddr     []byte `struc:"[16]byte"`
	PeerAddr      []byte `struc:"[16]byte"`
	IsIpv6        uint8
	DetectMult    uint8
}

func (*BfdUDPMod) GetMessageName() string {
	return "bfd_udp_mod"
}
func (*BfdUDPMod) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdUDPMod) GetCrcString() string {
	return "cee1341e"
}
func NewBfdUDPMod() api.Message {
	return &BfdUDPMod{}
}

// BfdUDPModReply represents the VPP binary API message 'bfd_udp_mod_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 66:
//
//        ["bfd_udp_mod_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x6f9b0cf4"}
//        ],
//
type BfdUDPModReply struct {
	Retval int32
}

func (*BfdUDPModReply) GetMessageName() string {
	return "bfd_udp_mod_reply"
}
func (*BfdUDPModReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdUDPModReply) GetCrcString() string {
	return "6f9b0cf4"
}
func NewBfdUDPModReply() api.Message {
	return &BfdUDPModReply{}
}

// BfdUDPDel represents the VPP binary API message 'bfd_udp_del'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 72:
//
//        ["bfd_udp_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "local_addr", 16],
//            ["u8", "peer_addr", 16],
//            ["u8", "is_ipv6"],
//            {"crc" : "0xe95cc3ee"}
//        ],
//
type BfdUDPDel struct {
	SwIfIndex uint32
	LocalAddr []byte `struc:"[16]byte"`
	PeerAddr  []byte `struc:"[16]byte"`
	IsIpv6    uint8
}

func (*BfdUDPDel) GetMessageName() string {
	return "bfd_udp_del"
}
func (*BfdUDPDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdUDPDel) GetCrcString() string {
	return "e95cc3ee"
}
func NewBfdUDPDel() api.Message {
	return &BfdUDPDel{}
}

// BfdUDPDelReply represents the VPP binary API message 'bfd_udp_del_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 82:
//
//        ["bfd_udp_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xb9b0b355"}
//        ],
//
type BfdUDPDelReply struct {
	Retval int32
}

func (*BfdUDPDelReply) GetMessageName() string {
	return "bfd_udp_del_reply"
}
func (*BfdUDPDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdUDPDelReply) GetCrcString() string {
	return "b9b0b355"
}
func NewBfdUDPDelReply() api.Message {
	return &BfdUDPDelReply{}
}

// BfdUDPSessionDump represents the VPP binary API message 'bfd_udp_session_dump'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 88:
//
//        ["bfd_udp_session_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xb5bd25a6"}
//        ],
//
type BfdUDPSessionDump struct {
}

func (*BfdUDPSessionDump) GetMessageName() string {
	return "bfd_udp_session_dump"
}
func (*BfdUDPSessionDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdUDPSessionDump) GetCrcString() string {
	return "b5bd25a6"
}
func NewBfdUDPSessionDump() api.Message {
	return &BfdUDPSessionDump{}
}

// BfdUDPSessionDetails represents the VPP binary API message 'bfd_udp_session_details'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 94:
//
//        ["bfd_udp_session_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "local_addr", 16],
//            ["u8", "peer_addr", 16],
//            ["u8", "is_ipv6"],
//            ["u8", "state"],
//            ["u8", "is_authenticated"],
//            ["u8", "bfd_key_id"],
//            ["u32", "conf_key_id"],
//            ["u32", "required_min_rx"],
//            ["u32", "desired_min_tx"],
//            ["u8", "detect_mult"],
//            {"crc" : "0x1a431796"}
//        ],
//
type BfdUDPSessionDetails struct {
	SwIfIndex       uint32
	LocalAddr       []byte `struc:"[16]byte"`
	PeerAddr        []byte `struc:"[16]byte"`
	IsIpv6          uint8
	State           uint8
	IsAuthenticated uint8
	BfdKeyID        uint8
	ConfKeyID       uint32
	RequiredMinRx   uint32
	DesiredMinTx    uint32
	DetectMult      uint8
}

func (*BfdUDPSessionDetails) GetMessageName() string {
	return "bfd_udp_session_details"
}
func (*BfdUDPSessionDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdUDPSessionDetails) GetCrcString() string {
	return "1a431796"
}
func NewBfdUDPSessionDetails() api.Message {
	return &BfdUDPSessionDetails{}
}

// BfdUDPSessionSetFlags represents the VPP binary API message 'bfd_udp_session_set_flags'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 110:
//
//        ["bfd_udp_session_set_flags",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "local_addr", 16],
//            ["u8", "peer_addr", 16],
//            ["u8", "is_ipv6"],
//            ["u8", "admin_up_down"],
//            {"crc" : "0x7b8518ba"}
//        ],
//
type BfdUDPSessionSetFlags struct {
	SwIfIndex   uint32
	LocalAddr   []byte `struc:"[16]byte"`
	PeerAddr    []byte `struc:"[16]byte"`
	IsIpv6      uint8
	AdminUpDown uint8
}

func (*BfdUDPSessionSetFlags) GetMessageName() string {
	return "bfd_udp_session_set_flags"
}
func (*BfdUDPSessionSetFlags) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdUDPSessionSetFlags) GetCrcString() string {
	return "7b8518ba"
}
func NewBfdUDPSessionSetFlags() api.Message {
	return &BfdUDPSessionSetFlags{}
}

// BfdUDPSessionSetFlagsReply represents the VPP binary API message 'bfd_udp_session_set_flags_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 121:
//
//        ["bfd_udp_session_set_flags_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x1a8335c3"}
//        ],
//
type BfdUDPSessionSetFlagsReply struct {
	Retval int32
}

func (*BfdUDPSessionSetFlagsReply) GetMessageName() string {
	return "bfd_udp_session_set_flags_reply"
}
func (*BfdUDPSessionSetFlagsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdUDPSessionSetFlagsReply) GetCrcString() string {
	return "1a8335c3"
}
func NewBfdUDPSessionSetFlagsReply() api.Message {
	return &BfdUDPSessionSetFlagsReply{}
}

// WantBfdEvents represents the VPP binary API message 'want_bfd_events'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 127:
//
//        ["want_bfd_events",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "enable_disable"],
//            ["u32", "pid"],
//            {"crc" : "0xbc6547f0"}
//        ],
//
type WantBfdEvents struct {
	EnableDisable uint32
	Pid           uint32
}

func (*WantBfdEvents) GetMessageName() string {
	return "want_bfd_events"
}
func (*WantBfdEvents) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*WantBfdEvents) GetCrcString() string {
	return "bc6547f0"
}
func NewWantBfdEvents() api.Message {
	return &WantBfdEvents{}
}

// WantBfdEventsReply represents the VPP binary API message 'want_bfd_events_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 135:
//
//        ["want_bfd_events_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xbe8b3ff3"}
//        ],
//
type WantBfdEventsReply struct {
	Retval int32
}

func (*WantBfdEventsReply) GetMessageName() string {
	return "want_bfd_events_reply"
}
func (*WantBfdEventsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*WantBfdEventsReply) GetCrcString() string {
	return "be8b3ff3"
}
func NewWantBfdEventsReply() api.Message {
	return &WantBfdEventsReply{}
}

// BfdAuthSetKey represents the VPP binary API message 'bfd_auth_set_key'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 141:
//
//        ["bfd_auth_set_key",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "conf_key_id"],
//            ["u8", "key_len"],
//            ["u8", "auth_type"],
//            ["u8", "key", 20],
//            {"crc" : "0xabbe70cc"}
//        ],
//
type BfdAuthSetKey struct {
	ConfKeyID uint32
	KeyLen    uint8
	AuthType  uint8
	Key       []byte `struc:"[20]byte"`
}

func (*BfdAuthSetKey) GetMessageName() string {
	return "bfd_auth_set_key"
}
func (*BfdAuthSetKey) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdAuthSetKey) GetCrcString() string {
	return "abbe70cc"
}
func NewBfdAuthSetKey() api.Message {
	return &BfdAuthSetKey{}
}

// BfdAuthSetKeyReply represents the VPP binary API message 'bfd_auth_set_key_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 151:
//
//        ["bfd_auth_set_key_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x68ed0d61"}
//        ],
//
type BfdAuthSetKeyReply struct {
	Retval int32
}

func (*BfdAuthSetKeyReply) GetMessageName() string {
	return "bfd_auth_set_key_reply"
}
func (*BfdAuthSetKeyReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdAuthSetKeyReply) GetCrcString() string {
	return "68ed0d61"
}
func NewBfdAuthSetKeyReply() api.Message {
	return &BfdAuthSetKeyReply{}
}

// BfdAuthDelKey represents the VPP binary API message 'bfd_auth_del_key'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 157:
//
//        ["bfd_auth_del_key",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "conf_key_id"],
//            {"crc" : "0x4e4d7318"}
//        ],
//
type BfdAuthDelKey struct {
	ConfKeyID uint32
}

func (*BfdAuthDelKey) GetMessageName() string {
	return "bfd_auth_del_key"
}
func (*BfdAuthDelKey) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdAuthDelKey) GetCrcString() string {
	return "4e4d7318"
}
func NewBfdAuthDelKey() api.Message {
	return &BfdAuthDelKey{}
}

// BfdAuthDelKeyReply represents the VPP binary API message 'bfd_auth_del_key_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 164:
//
//        ["bfd_auth_del_key_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xa0db385f"}
//        ],
//
type BfdAuthDelKeyReply struct {
	Retval int32
}

func (*BfdAuthDelKeyReply) GetMessageName() string {
	return "bfd_auth_del_key_reply"
}
func (*BfdAuthDelKeyReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdAuthDelKeyReply) GetCrcString() string {
	return "a0db385f"
}
func NewBfdAuthDelKeyReply() api.Message {
	return &BfdAuthDelKeyReply{}
}

// BfdAuthKeysDump represents the VPP binary API message 'bfd_auth_keys_dump'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 170:
//
//        ["bfd_auth_keys_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x336fa6ba"}
//        ],
//
type BfdAuthKeysDump struct {
}

func (*BfdAuthKeysDump) GetMessageName() string {
	return "bfd_auth_keys_dump"
}
func (*BfdAuthKeysDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdAuthKeysDump) GetCrcString() string {
	return "336fa6ba"
}
func NewBfdAuthKeysDump() api.Message {
	return &BfdAuthKeysDump{}
}

// BfdAuthKeysDetails represents the VPP binary API message 'bfd_auth_keys_details'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 176:
//
//        ["bfd_auth_keys_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "conf_key_id"],
//            ["u32", "use_count"],
//            ["u8", "auth_type"],
//            {"crc" : "0x377927eb"}
//        ],
//
type BfdAuthKeysDetails struct {
	ConfKeyID uint32
	UseCount  uint32
	AuthType  uint8
}

func (*BfdAuthKeysDetails) GetMessageName() string {
	return "bfd_auth_keys_details"
}
func (*BfdAuthKeysDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdAuthKeysDetails) GetCrcString() string {
	return "377927eb"
}
func NewBfdAuthKeysDetails() api.Message {
	return &BfdAuthKeysDetails{}
}

// BfdUDPAuthActivate represents the VPP binary API message 'bfd_udp_auth_activate'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 184:
//
//        ["bfd_udp_auth_activate",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "local_addr", 16],
//            ["u8", "peer_addr", 16],
//            ["u8", "is_ipv6"],
//            ["u8", "is_delayed"],
//            ["u8", "bfd_key_id"],
//            ["u32", "conf_key_id"],
//            {"crc" : "0x87ac919e"}
//        ],
//
type BfdUDPAuthActivate struct {
	SwIfIndex uint32
	LocalAddr []byte `struc:"[16]byte"`
	PeerAddr  []byte `struc:"[16]byte"`
	IsIpv6    uint8
	IsDelayed uint8
	BfdKeyID  uint8
	ConfKeyID uint32
}

func (*BfdUDPAuthActivate) GetMessageName() string {
	return "bfd_udp_auth_activate"
}
func (*BfdUDPAuthActivate) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdUDPAuthActivate) GetCrcString() string {
	return "87ac919e"
}
func NewBfdUDPAuthActivate() api.Message {
	return &BfdUDPAuthActivate{}
}

// BfdUDPAuthActivateReply represents the VPP binary API message 'bfd_udp_auth_activate_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 197:
//
//        ["bfd_udp_auth_activate_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xba8f2610"}
//        ],
//
type BfdUDPAuthActivateReply struct {
	Retval int32
}

func (*BfdUDPAuthActivateReply) GetMessageName() string {
	return "bfd_udp_auth_activate_reply"
}
func (*BfdUDPAuthActivateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdUDPAuthActivateReply) GetCrcString() string {
	return "ba8f2610"
}
func NewBfdUDPAuthActivateReply() api.Message {
	return &BfdUDPAuthActivateReply{}
}

// BfdUDPAuthDeactivate represents the VPP binary API message 'bfd_udp_auth_deactivate'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 203:
//
//        ["bfd_udp_auth_deactivate",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "local_addr", 16],
//            ["u8", "peer_addr", 16],
//            ["u8", "is_ipv6"],
//            ["u8", "is_delayed"],
//            {"crc" : "0x75f4d9e3"}
//        ],
//
type BfdUDPAuthDeactivate struct {
	SwIfIndex uint32
	LocalAddr []byte `struc:"[16]byte"`
	PeerAddr  []byte `struc:"[16]byte"`
	IsIpv6    uint8
	IsDelayed uint8
}

func (*BfdUDPAuthDeactivate) GetMessageName() string {
	return "bfd_udp_auth_deactivate"
}
func (*BfdUDPAuthDeactivate) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*BfdUDPAuthDeactivate) GetCrcString() string {
	return "75f4d9e3"
}
func NewBfdUDPAuthDeactivate() api.Message {
	return &BfdUDPAuthDeactivate{}
}

// BfdUDPAuthDeactivateReply represents the VPP binary API message 'bfd_udp_auth_deactivate_reply'.
// Generated from '/usr/share/vpp/api/bfd.api.json', line 214:
//
//        ["bfd_udp_auth_deactivate_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x1885e013"}
//        ]
//
type BfdUDPAuthDeactivateReply struct {
	Retval int32
}

func (*BfdUDPAuthDeactivateReply) GetMessageName() string {
	return "bfd_udp_auth_deactivate_reply"
}
func (*BfdUDPAuthDeactivateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*BfdUDPAuthDeactivateReply) GetCrcString() string {
	return "1885e013"
}
func NewBfdUDPAuthDeactivateReply() api.Message {
	return &BfdUDPAuthDeactivateReply{}
}
