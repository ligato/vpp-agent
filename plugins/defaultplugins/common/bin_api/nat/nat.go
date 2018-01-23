// Package nat represents the VPP binary API of the 'nat' VPP module.
// DO NOT EDIT. Generated from '/usr/share/vpp/api/nat.api.json'
package nat

import "git.fd.io/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0xae1c8462

// Nat44LbAddrPort represents the VPP binary API data type 'nat44_lb_addr_port'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 3:
//
//        ["nat44_lb_addr_port",
//            ["u8", "addr", 4],
//            ["u16", "port"],
//            ["u8", "probability"],
//            {"crc" : "0x69a407b1"}
//        ]
//
type Nat44LbAddrPort struct {
	Addr        []byte `struc:"[4]byte"`
	Port        uint16
	Probability uint8
}

func (*Nat44LbAddrPort) GetTypeName() string {
	return "nat44_lb_addr_port"
}
func (*Nat44LbAddrPort) GetCrcString() string {
	return "69a407b1"
}

// NatControlPing represents the VPP binary API message 'nat_control_ping'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 11:
//
//        ["nat_control_ping",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x96e6a834"}
//        ],
//
type NatControlPing struct {
}

func (*NatControlPing) GetMessageName() string {
	return "nat_control_ping"
}
func (*NatControlPing) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatControlPing) GetCrcString() string {
	return "96e6a834"
}
func NewNatControlPing() api.Message {
	return &NatControlPing{}
}

// NatControlPingReply represents the VPP binary API message 'nat_control_ping_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 17:
//
//        ["nat_control_ping_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u32", "client_index"],
//            ["u32", "vpe_pid"],
//            {"crc" : "0x2d86a59b"}
//        ],
//
type NatControlPingReply struct {
	Retval      int32
	ClientIndex uint32
	VpePid      uint32
}

func (*NatControlPingReply) GetMessageName() string {
	return "nat_control_ping_reply"
}
func (*NatControlPingReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatControlPingReply) GetCrcString() string {
	return "2d86a59b"
}
func NewNatControlPingReply() api.Message {
	return &NatControlPingReply{}
}

// NatShowConfig represents the VPP binary API message 'nat_show_config'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 25:
//
//        ["nat_show_config",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xf1e6587b"}
//        ],
//
type NatShowConfig struct {
}

func (*NatShowConfig) GetMessageName() string {
	return "nat_show_config"
}
func (*NatShowConfig) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatShowConfig) GetCrcString() string {
	return "f1e6587b"
}
func NewNatShowConfig() api.Message {
	return &NatShowConfig{}
}

// NatShowConfigReply represents the VPP binary API message 'nat_show_config_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 31:
//
//        ["nat_show_config_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u8", "static_mapping_only"],
//            ["u8", "static_mapping_connection_tracking"],
//            ["u8", "deterministic"],
//            ["u32", "translation_buckets"],
//            ["u32", "translation_memory_size"],
//            ["u32", "user_buckets"],
//            ["u32", "user_memory_size"],
//            ["u32", "max_translations_per_user"],
//            ["u32", "outside_vrf_id"],
//            ["u32", "inside_vrf_id"],
//            {"crc" : "0x4a456e3a"}
//        ],
//
type NatShowConfigReply struct {
	Retval                          int32
	StaticMappingOnly               uint8
	StaticMappingConnectionTracking uint8
	Deterministic                   uint8
	TranslationBuckets              uint32
	TranslationMemorySize           uint32
	UserBuckets                     uint32
	UserMemorySize                  uint32
	MaxTranslationsPerUser          uint32
	OutsideVrfID                    uint32
	InsideVrfID                     uint32
}

func (*NatShowConfigReply) GetMessageName() string {
	return "nat_show_config_reply"
}
func (*NatShowConfigReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatShowConfigReply) GetCrcString() string {
	return "4a456e3a"
}
func NewNatShowConfigReply() api.Message {
	return &NatShowConfigReply{}
}

// NatSetWorkers represents the VPP binary API message 'nat_set_workers'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 47:
//
//        ["nat_set_workers",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u64", "worker_mask"],
//            {"crc" : "0xf7b85189"}
//        ],
//
type NatSetWorkers struct {
	WorkerMask uint64
}

func (*NatSetWorkers) GetMessageName() string {
	return "nat_set_workers"
}
func (*NatSetWorkers) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatSetWorkers) GetCrcString() string {
	return "f7b85189"
}
func NewNatSetWorkers() api.Message {
	return &NatSetWorkers{}
}

// NatSetWorkersReply represents the VPP binary API message 'nat_set_workers_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 54:
//
//        ["nat_set_workers_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x9a7d70ae"}
//        ],
//
type NatSetWorkersReply struct {
	Retval int32
}

func (*NatSetWorkersReply) GetMessageName() string {
	return "nat_set_workers_reply"
}
func (*NatSetWorkersReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatSetWorkersReply) GetCrcString() string {
	return "9a7d70ae"
}
func NewNatSetWorkersReply() api.Message {
	return &NatSetWorkersReply{}
}

// NatWorkerDump represents the VPP binary API message 'nat_worker_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 60:
//
//        ["nat_worker_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x6adf1d97"}
//        ],
//
type NatWorkerDump struct {
}

func (*NatWorkerDump) GetMessageName() string {
	return "nat_worker_dump"
}
func (*NatWorkerDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatWorkerDump) GetCrcString() string {
	return "6adf1d97"
}
func NewNatWorkerDump() api.Message {
	return &NatWorkerDump{}
}

// NatWorkerDetails represents the VPP binary API message 'nat_worker_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 66:
//
//        ["nat_worker_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "worker_index"],
//            ["u32", "lcore_id"],
//            ["u8", "name", 64],
//            {"crc" : "0xd001e0c7"}
//        ],
//
type NatWorkerDetails struct {
	WorkerIndex uint32
	LcoreID     uint32
	Name        []byte `struc:"[64]byte"`
}

func (*NatWorkerDetails) GetMessageName() string {
	return "nat_worker_details"
}
func (*NatWorkerDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatWorkerDetails) GetCrcString() string {
	return "d001e0c7"
}
func NewNatWorkerDetails() api.Message {
	return &NatWorkerDetails{}
}

// NatIpfixEnableDisable represents the VPP binary API message 'nat_ipfix_enable_disable'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 74:
//
//        ["nat_ipfix_enable_disable",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "domain_id"],
//            ["u16", "src_port"],
//            ["u8", "enable"],
//            {"crc" : "0xc1b2fbba"}
//        ],
//
type NatIpfixEnableDisable struct {
	DomainID uint32
	SrcPort  uint16
	Enable   uint8
}

func (*NatIpfixEnableDisable) GetMessageName() string {
	return "nat_ipfix_enable_disable"
}
func (*NatIpfixEnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatIpfixEnableDisable) GetCrcString() string {
	return "c1b2fbba"
}
func NewNatIpfixEnableDisable() api.Message {
	return &NatIpfixEnableDisable{}
}

// NatIpfixEnableDisableReply represents the VPP binary API message 'nat_ipfix_enable_disable_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 83:
//
//        ["nat_ipfix_enable_disable_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x3bb820c4"}
//        ],
//
type NatIpfixEnableDisableReply struct {
	Retval int32
}

func (*NatIpfixEnableDisableReply) GetMessageName() string {
	return "nat_ipfix_enable_disable_reply"
}
func (*NatIpfixEnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatIpfixEnableDisableReply) GetCrcString() string {
	return "3bb820c4"
}
func NewNatIpfixEnableDisableReply() api.Message {
	return &NatIpfixEnableDisableReply{}
}

// NatSetReass represents the VPP binary API message 'nat_set_reass'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 89:
//
//        ["nat_set_reass",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "timeout"],
//            ["u16", "max_reass"],
//            ["u8", "max_frag"],
//            ["u8", "drop_frag"],
//            ["u8", "is_ip6"],
//            {"crc" : "0xd1a40860"}
//        ],
//
type NatSetReass struct {
	Timeout  uint32
	MaxReass uint16
	MaxFrag  uint8
	DropFrag uint8
	IsIP6    uint8
}

func (*NatSetReass) GetMessageName() string {
	return "nat_set_reass"
}
func (*NatSetReass) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatSetReass) GetCrcString() string {
	return "d1a40860"
}
func NewNatSetReass() api.Message {
	return &NatSetReass{}
}

// NatSetReassReply represents the VPP binary API message 'nat_set_reass_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 100:
//
//        ["nat_set_reass_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x9098cdf4"}
//        ],
//
type NatSetReassReply struct {
	Retval int32
}

func (*NatSetReassReply) GetMessageName() string {
	return "nat_set_reass_reply"
}
func (*NatSetReassReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatSetReassReply) GetCrcString() string {
	return "9098cdf4"
}
func NewNatSetReassReply() api.Message {
	return &NatSetReassReply{}
}

// NatGetReass represents the VPP binary API message 'nat_get_reass'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 106:
//
//        ["nat_get_reass",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x49b28ff4"}
//        ],
//
type NatGetReass struct {
}

func (*NatGetReass) GetMessageName() string {
	return "nat_get_reass"
}
func (*NatGetReass) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatGetReass) GetCrcString() string {
	return "49b28ff4"
}
func NewNatGetReass() api.Message {
	return &NatGetReass{}
}

// NatGetReassReply represents the VPP binary API message 'nat_get_reass_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 112:
//
//        ["nat_get_reass_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u32", "ip4_timeout"],
//            ["u16", "ip4_max_reass"],
//            ["u8", "ip4_max_frag"],
//            ["u8", "ip4_drop_frag"],
//            ["u32", "ip6_timeout"],
//            ["u16", "ip6_max_reass"],
//            ["u8", "ip6_max_frag"],
//            ["u8", "ip6_drop_frag"],
//            {"crc" : "0xaddd1031"}
//        ],
//
type NatGetReassReply struct {
	Retval      int32
	IP4Timeout  uint32
	IP4MaxReass uint16
	IP4MaxFrag  uint8
	IP4DropFrag uint8
	IP6Timeout  uint32
	IP6MaxReass uint16
	IP6MaxFrag  uint8
	IP6DropFrag uint8
}

func (*NatGetReassReply) GetMessageName() string {
	return "nat_get_reass_reply"
}
func (*NatGetReassReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatGetReassReply) GetCrcString() string {
	return "addd1031"
}
func NewNatGetReassReply() api.Message {
	return &NatGetReassReply{}
}

// NatReassDump represents the VPP binary API message 'nat_reass_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 126:
//
//        ["nat_reass_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x91f7b28d"}
//        ],
//
type NatReassDump struct {
}

func (*NatReassDump) GetMessageName() string {
	return "nat_reass_dump"
}
func (*NatReassDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatReassDump) GetCrcString() string {
	return "91f7b28d"
}
func NewNatReassDump() api.Message {
	return &NatReassDump{}
}

// NatReassDetails represents the VPP binary API message 'nat_reass_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 132:
//
//        ["nat_reass_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "is_ip4"],
//            ["u8", "src_addr", 16],
//            ["u8", "dst_addr", 16],
//            ["u32", "frag_id"],
//            ["u8", "proto"],
//            ["u8", "frag_n"],
//            {"crc" : "0x0f884b01"}
//        ],
//
type NatReassDetails struct {
	IsIP4   uint8
	SrcAddr []byte `struc:"[16]byte"`
	DstAddr []byte `struc:"[16]byte"`
	FragID  uint32
	Proto   uint8
	FragN   uint8
}

func (*NatReassDetails) GetMessageName() string {
	return "nat_reass_details"
}
func (*NatReassDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatReassDetails) GetCrcString() string {
	return "0f884b01"
}
func NewNatReassDetails() api.Message {
	return &NatReassDetails{}
}

// Nat44AddDelAddressRange represents the VPP binary API message 'nat44_add_del_address_range'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 143:
//
//        ["nat44_add_del_address_range",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "first_ip_address", 4],
//            ["u8", "last_ip_address", 4],
//            ["u32", "vrf_id"],
//            ["u8", "twice_nat"],
//            ["u8", "is_add"],
//            {"crc" : "0xee594dd1"}
//        ],
//
type Nat44AddDelAddressRange struct {
	FirstIPAddress []byte `struc:"[4]byte"`
	LastIPAddress  []byte `struc:"[4]byte"`
	VrfID          uint32
	TwiceNat       uint8
	IsAdd          uint8
}

func (*Nat44AddDelAddressRange) GetMessageName() string {
	return "nat44_add_del_address_range"
}
func (*Nat44AddDelAddressRange) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44AddDelAddressRange) GetCrcString() string {
	return "ee594dd1"
}
func NewNat44AddDelAddressRange() api.Message {
	return &Nat44AddDelAddressRange{}
}

// Nat44AddDelAddressRangeReply represents the VPP binary API message 'nat44_add_del_address_range_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 154:
//
//        ["nat44_add_del_address_range_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x4e9b159d"}
//        ],
//
type Nat44AddDelAddressRangeReply struct {
	Retval int32
}

func (*Nat44AddDelAddressRangeReply) GetMessageName() string {
	return "nat44_add_del_address_range_reply"
}
func (*Nat44AddDelAddressRangeReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44AddDelAddressRangeReply) GetCrcString() string {
	return "4e9b159d"
}
func NewNat44AddDelAddressRangeReply() api.Message {
	return &Nat44AddDelAddressRangeReply{}
}

// Nat44AddressDump represents the VPP binary API message 'nat44_address_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 160:
//
//        ["nat44_address_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xe75a129a"}
//        ],
//
type Nat44AddressDump struct {
}

func (*Nat44AddressDump) GetMessageName() string {
	return "nat44_address_dump"
}
func (*Nat44AddressDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44AddressDump) GetCrcString() string {
	return "e75a129a"
}
func NewNat44AddressDump() api.Message {
	return &Nat44AddressDump{}
}

// Nat44AddressDetails represents the VPP binary API message 'nat44_address_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 166:
//
//        ["nat44_address_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "ip_address", 4],
//            ["u8", "twice_nat"],
//            ["u32", "vrf_id"],
//            {"crc" : "0x1a61557a"}
//        ],
//
type Nat44AddressDetails struct {
	IPAddress []byte `struc:"[4]byte"`
	TwiceNat  uint8
	VrfID     uint32
}

func (*Nat44AddressDetails) GetMessageName() string {
	return "nat44_address_details"
}
func (*Nat44AddressDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44AddressDetails) GetCrcString() string {
	return "1a61557a"
}
func NewNat44AddressDetails() api.Message {
	return &Nat44AddressDetails{}
}

// Nat44InterfaceAddDelFeature represents the VPP binary API message 'nat44_interface_add_del_feature'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 174:
//
//        ["nat44_interface_add_del_feature",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "is_inside"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0xc5aa4c6e"}
//        ],
//
type Nat44InterfaceAddDelFeature struct {
	IsAdd     uint8
	IsInside  uint8
	SwIfIndex uint32
}

func (*Nat44InterfaceAddDelFeature) GetMessageName() string {
	return "nat44_interface_add_del_feature"
}
func (*Nat44InterfaceAddDelFeature) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44InterfaceAddDelFeature) GetCrcString() string {
	return "c5aa4c6e"
}
func NewNat44InterfaceAddDelFeature() api.Message {
	return &Nat44InterfaceAddDelFeature{}
}

// Nat44InterfaceAddDelFeatureReply represents the VPP binary API message 'nat44_interface_add_del_feature_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 183:
//
//        ["nat44_interface_add_del_feature_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xfefe38af"}
//        ],
//
type Nat44InterfaceAddDelFeatureReply struct {
	Retval int32
}

func (*Nat44InterfaceAddDelFeatureReply) GetMessageName() string {
	return "nat44_interface_add_del_feature_reply"
}
func (*Nat44InterfaceAddDelFeatureReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44InterfaceAddDelFeatureReply) GetCrcString() string {
	return "fefe38af"
}
func NewNat44InterfaceAddDelFeatureReply() api.Message {
	return &Nat44InterfaceAddDelFeatureReply{}
}

// Nat44InterfaceDump represents the VPP binary API message 'nat44_interface_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 189:
//
//        ["nat44_interface_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x4eae339a"}
//        ],
//
type Nat44InterfaceDump struct {
}

func (*Nat44InterfaceDump) GetMessageName() string {
	return "nat44_interface_dump"
}
func (*Nat44InterfaceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44InterfaceDump) GetCrcString() string {
	return "4eae339a"
}
func NewNat44InterfaceDump() api.Message {
	return &Nat44InterfaceDump{}
}

// Nat44InterfaceDetails represents the VPP binary API message 'nat44_interface_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 195:
//
//        ["nat44_interface_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "is_inside"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0xc3f78d04"}
//        ],
//
type Nat44InterfaceDetails struct {
	IsInside  uint8
	SwIfIndex uint32
}

func (*Nat44InterfaceDetails) GetMessageName() string {
	return "nat44_interface_details"
}
func (*Nat44InterfaceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44InterfaceDetails) GetCrcString() string {
	return "c3f78d04"
}
func NewNat44InterfaceDetails() api.Message {
	return &Nat44InterfaceDetails{}
}

// Nat44InterfaceAddDelOutputFeature represents the VPP binary API message 'nat44_interface_add_del_output_feature'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 202:
//
//        ["nat44_interface_add_del_output_feature",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "is_inside"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0xb819e4dd"}
//        ],
//
type Nat44InterfaceAddDelOutputFeature struct {
	IsAdd     uint8
	IsInside  uint8
	SwIfIndex uint32
}

func (*Nat44InterfaceAddDelOutputFeature) GetMessageName() string {
	return "nat44_interface_add_del_output_feature"
}
func (*Nat44InterfaceAddDelOutputFeature) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44InterfaceAddDelOutputFeature) GetCrcString() string {
	return "b819e4dd"
}
func NewNat44InterfaceAddDelOutputFeature() api.Message {
	return &Nat44InterfaceAddDelOutputFeature{}
}

// Nat44InterfaceAddDelOutputFeatureReply represents the VPP binary API message 'nat44_interface_add_del_output_feature_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 211:
//
//        ["nat44_interface_add_del_output_feature_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xafb362f5"}
//        ],
//
type Nat44InterfaceAddDelOutputFeatureReply struct {
	Retval int32
}

func (*Nat44InterfaceAddDelOutputFeatureReply) GetMessageName() string {
	return "nat44_interface_add_del_output_feature_reply"
}
func (*Nat44InterfaceAddDelOutputFeatureReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44InterfaceAddDelOutputFeatureReply) GetCrcString() string {
	return "afb362f5"
}
func NewNat44InterfaceAddDelOutputFeatureReply() api.Message {
	return &Nat44InterfaceAddDelOutputFeatureReply{}
}

// Nat44InterfaceOutputFeatureDump represents the VPP binary API message 'nat44_interface_output_feature_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 217:
//
//        ["nat44_interface_output_feature_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xe7245137"}
//        ],
//
type Nat44InterfaceOutputFeatureDump struct {
}

func (*Nat44InterfaceOutputFeatureDump) GetMessageName() string {
	return "nat44_interface_output_feature_dump"
}
func (*Nat44InterfaceOutputFeatureDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44InterfaceOutputFeatureDump) GetCrcString() string {
	return "e7245137"
}
func NewNat44InterfaceOutputFeatureDump() api.Message {
	return &Nat44InterfaceOutputFeatureDump{}
}

// Nat44InterfaceOutputFeatureDetails represents the VPP binary API message 'nat44_interface_output_feature_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 223:
//
//        ["nat44_interface_output_feature_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "is_inside"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x119c7053"}
//        ],
//
type Nat44InterfaceOutputFeatureDetails struct {
	IsInside  uint8
	SwIfIndex uint32
}

func (*Nat44InterfaceOutputFeatureDetails) GetMessageName() string {
	return "nat44_interface_output_feature_details"
}
func (*Nat44InterfaceOutputFeatureDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44InterfaceOutputFeatureDetails) GetCrcString() string {
	return "119c7053"
}
func NewNat44InterfaceOutputFeatureDetails() api.Message {
	return &Nat44InterfaceOutputFeatureDetails{}
}

// Nat44AddDelStaticMapping represents the VPP binary API message 'nat44_add_del_static_mapping'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 230:
//
//        ["nat44_add_del_static_mapping",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "addr_only"],
//            ["u8", "local_ip_address", 4],
//            ["u8", "external_ip_address", 4],
//            ["u8", "protocol"],
//            ["u16", "local_port"],
//            ["u16", "external_port"],
//            ["u32", "external_sw_if_index"],
//            ["u32", "vrf_id"],
//            ["u8", "twice_nat"],
//            {"crc" : "0x90ce56b0"}
//        ],
//
type Nat44AddDelStaticMapping struct {
	IsAdd             uint8
	AddrOnly          uint8
	LocalIPAddress    []byte `struc:"[4]byte"`
	ExternalIPAddress []byte `struc:"[4]byte"`
	Protocol          uint8
	LocalPort         uint16
	ExternalPort      uint16
	ExternalSwIfIndex uint32
	VrfID             uint32
	TwiceNat          uint8
}

func (*Nat44AddDelStaticMapping) GetMessageName() string {
	return "nat44_add_del_static_mapping"
}
func (*Nat44AddDelStaticMapping) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44AddDelStaticMapping) GetCrcString() string {
	return "90ce56b0"
}
func NewNat44AddDelStaticMapping() api.Message {
	return &Nat44AddDelStaticMapping{}
}

// Nat44AddDelStaticMappingReply represents the VPP binary API message 'nat44_add_del_static_mapping_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 246:
//
//        ["nat44_add_del_static_mapping_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xa30f3c34"}
//        ],
//
type Nat44AddDelStaticMappingReply struct {
	Retval int32
}

func (*Nat44AddDelStaticMappingReply) GetMessageName() string {
	return "nat44_add_del_static_mapping_reply"
}
func (*Nat44AddDelStaticMappingReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44AddDelStaticMappingReply) GetCrcString() string {
	return "a30f3c34"
}
func NewNat44AddDelStaticMappingReply() api.Message {
	return &Nat44AddDelStaticMappingReply{}
}

// Nat44StaticMappingDump represents the VPP binary API message 'nat44_static_mapping_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 252:
//
//        ["nat44_static_mapping_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xdf3b078e"}
//        ],
//
type Nat44StaticMappingDump struct {
}

func (*Nat44StaticMappingDump) GetMessageName() string {
	return "nat44_static_mapping_dump"
}
func (*Nat44StaticMappingDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44StaticMappingDump) GetCrcString() string {
	return "df3b078e"
}
func NewNat44StaticMappingDump() api.Message {
	return &Nat44StaticMappingDump{}
}

// Nat44StaticMappingDetails represents the VPP binary API message 'nat44_static_mapping_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 258:
//
//        ["nat44_static_mapping_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "addr_only"],
//            ["u8", "local_ip_address", 4],
//            ["u8", "external_ip_address", 4],
//            ["u8", "protocol"],
//            ["u16", "local_port"],
//            ["u16", "external_port"],
//            ["u32", "external_sw_if_index"],
//            ["u32", "vrf_id"],
//            ["u8", "twice_nat"],
//            {"crc" : "0x6f606451"}
//        ],
//
type Nat44StaticMappingDetails struct {
	AddrOnly          uint8
	LocalIPAddress    []byte `struc:"[4]byte"`
	ExternalIPAddress []byte `struc:"[4]byte"`
	Protocol          uint8
	LocalPort         uint16
	ExternalPort      uint16
	ExternalSwIfIndex uint32
	VrfID             uint32
	TwiceNat          uint8
}

func (*Nat44StaticMappingDetails) GetMessageName() string {
	return "nat44_static_mapping_details"
}
func (*Nat44StaticMappingDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44StaticMappingDetails) GetCrcString() string {
	return "6f606451"
}
func NewNat44StaticMappingDetails() api.Message {
	return &Nat44StaticMappingDetails{}
}

// Nat44AddDelIdentityMapping represents the VPP binary API message 'nat44_add_del_identity_mapping'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 272:
//
//        ["nat44_add_del_identity_mapping",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "addr_only"],
//            ["u8", "ip_address", 4],
//            ["u8", "protocol"],
//            ["u16", "port"],
//            ["u32", "sw_if_index"],
//            ["u32", "vrf_id"],
//            {"crc" : "0x156e1ae7"}
//        ],
//
type Nat44AddDelIdentityMapping struct {
	IsAdd     uint8
	AddrOnly  uint8
	IPAddress []byte `struc:"[4]byte"`
	Protocol  uint8
	Port      uint16
	SwIfIndex uint32
	VrfID     uint32
}

func (*Nat44AddDelIdentityMapping) GetMessageName() string {
	return "nat44_add_del_identity_mapping"
}
func (*Nat44AddDelIdentityMapping) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44AddDelIdentityMapping) GetCrcString() string {
	return "156e1ae7"
}
func NewNat44AddDelIdentityMapping() api.Message {
	return &Nat44AddDelIdentityMapping{}
}

// Nat44AddDelIdentityMappingReply represents the VPP binary API message 'nat44_add_del_identity_mapping_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 285:
//
//        ["nat44_add_del_identity_mapping_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x01671195"}
//        ],
//
type Nat44AddDelIdentityMappingReply struct {
	Retval int32
}

func (*Nat44AddDelIdentityMappingReply) GetMessageName() string {
	return "nat44_add_del_identity_mapping_reply"
}
func (*Nat44AddDelIdentityMappingReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44AddDelIdentityMappingReply) GetCrcString() string {
	return "01671195"
}
func NewNat44AddDelIdentityMappingReply() api.Message {
	return &Nat44AddDelIdentityMappingReply{}
}

// Nat44IdentityMappingDump represents the VPP binary API message 'nat44_identity_mapping_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 291:
//
//        ["nat44_identity_mapping_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x12c99b5d"}
//        ],
//
type Nat44IdentityMappingDump struct {
}

func (*Nat44IdentityMappingDump) GetMessageName() string {
	return "nat44_identity_mapping_dump"
}
func (*Nat44IdentityMappingDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44IdentityMappingDump) GetCrcString() string {
	return "12c99b5d"
}
func NewNat44IdentityMappingDump() api.Message {
	return &Nat44IdentityMappingDump{}
}

// Nat44IdentityMappingDetails represents the VPP binary API message 'nat44_identity_mapping_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 297:
//
//        ["nat44_identity_mapping_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "addr_only"],
//            ["u8", "ip_address", 4],
//            ["u8", "protocol"],
//            ["u16", "port"],
//            ["u32", "sw_if_index"],
//            ["u32", "vrf_id"],
//            {"crc" : "0x0d279fbe"}
//        ],
//
type Nat44IdentityMappingDetails struct {
	AddrOnly  uint8
	IPAddress []byte `struc:"[4]byte"`
	Protocol  uint8
	Port      uint16
	SwIfIndex uint32
	VrfID     uint32
}

func (*Nat44IdentityMappingDetails) GetMessageName() string {
	return "nat44_identity_mapping_details"
}
func (*Nat44IdentityMappingDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44IdentityMappingDetails) GetCrcString() string {
	return "0d279fbe"
}
func NewNat44IdentityMappingDetails() api.Message {
	return &Nat44IdentityMappingDetails{}
}

// Nat44AddDelInterfaceAddr represents the VPP binary API message 'nat44_add_del_interface_addr'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 308:
//
//        ["nat44_add_del_interface_addr",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "twice_nat"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x2704ceea"}
//        ],
//
type Nat44AddDelInterfaceAddr struct {
	IsAdd     uint8
	TwiceNat  uint8
	SwIfIndex uint32
}

func (*Nat44AddDelInterfaceAddr) GetMessageName() string {
	return "nat44_add_del_interface_addr"
}
func (*Nat44AddDelInterfaceAddr) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44AddDelInterfaceAddr) GetCrcString() string {
	return "2704ceea"
}
func NewNat44AddDelInterfaceAddr() api.Message {
	return &Nat44AddDelInterfaceAddr{}
}

// Nat44AddDelInterfaceAddrReply represents the VPP binary API message 'nat44_add_del_interface_addr_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 317:
//
//        ["nat44_add_del_interface_addr_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x74d308ac"}
//        ],
//
type Nat44AddDelInterfaceAddrReply struct {
	Retval int32
}

func (*Nat44AddDelInterfaceAddrReply) GetMessageName() string {
	return "nat44_add_del_interface_addr_reply"
}
func (*Nat44AddDelInterfaceAddrReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44AddDelInterfaceAddrReply) GetCrcString() string {
	return "74d308ac"
}
func NewNat44AddDelInterfaceAddrReply() api.Message {
	return &Nat44AddDelInterfaceAddrReply{}
}

// Nat44InterfaceAddrDump represents the VPP binary API message 'nat44_interface_addr_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 323:
//
//        ["nat44_interface_addr_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x65ac3823"}
//        ],
//
type Nat44InterfaceAddrDump struct {
}

func (*Nat44InterfaceAddrDump) GetMessageName() string {
	return "nat44_interface_addr_dump"
}
func (*Nat44InterfaceAddrDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44InterfaceAddrDump) GetCrcString() string {
	return "65ac3823"
}
func NewNat44InterfaceAddrDump() api.Message {
	return &Nat44InterfaceAddrDump{}
}

// Nat44InterfaceAddrDetails represents the VPP binary API message 'nat44_interface_addr_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 329:
//
//        ["nat44_interface_addr_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "twice_nat"],
//            {"crc" : "0xebe0397b"}
//        ],
//
type Nat44InterfaceAddrDetails struct {
	SwIfIndex uint32
	TwiceNat  uint8
}

func (*Nat44InterfaceAddrDetails) GetMessageName() string {
	return "nat44_interface_addr_details"
}
func (*Nat44InterfaceAddrDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44InterfaceAddrDetails) GetCrcString() string {
	return "ebe0397b"
}
func NewNat44InterfaceAddrDetails() api.Message {
	return &Nat44InterfaceAddrDetails{}
}

// Nat44UserDump represents the VPP binary API message 'nat44_user_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 336:
//
//        ["nat44_user_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xbf92f8eb"}
//        ],
//
type Nat44UserDump struct {
}

func (*Nat44UserDump) GetMessageName() string {
	return "nat44_user_dump"
}
func (*Nat44UserDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44UserDump) GetCrcString() string {
	return "bf92f8eb"
}
func NewNat44UserDump() api.Message {
	return &Nat44UserDump{}
}

// Nat44UserDetails represents the VPP binary API message 'nat44_user_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 342:
//
//        ["nat44_user_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "vrf_id"],
//            ["u8", "ip_address", 4],
//            ["u32", "nsessions"],
//            ["u32", "nstaticsessions"],
//            {"crc" : "0x77ce783b"}
//        ],
//
type Nat44UserDetails struct {
	VrfID           uint32
	IPAddress       []byte `struc:"[4]byte"`
	Nsessions       uint32
	Nstaticsessions uint32
}

func (*Nat44UserDetails) GetMessageName() string {
	return "nat44_user_details"
}
func (*Nat44UserDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44UserDetails) GetCrcString() string {
	return "77ce783b"
}
func NewNat44UserDetails() api.Message {
	return &Nat44UserDetails{}
}

// Nat44UserSessionDump represents the VPP binary API message 'nat44_user_session_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 351:
//
//        ["nat44_user_session_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "ip_address", 4],
//            ["u32", "vrf_id"],
//            {"crc" : "0x597cec3f"}
//        ],
//
type Nat44UserSessionDump struct {
	IPAddress []byte `struc:"[4]byte"`
	VrfID     uint32
}

func (*Nat44UserSessionDump) GetMessageName() string {
	return "nat44_user_session_dump"
}
func (*Nat44UserSessionDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44UserSessionDump) GetCrcString() string {
	return "597cec3f"
}
func NewNat44UserSessionDump() api.Message {
	return &Nat44UserSessionDump{}
}

// Nat44UserSessionDetails represents the VPP binary API message 'nat44_user_session_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 359:
//
//        ["nat44_user_session_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "outside_ip_address", 4],
//            ["u16", "outside_port"],
//            ["u8", "inside_ip_address", 4],
//            ["u16", "inside_port"],
//            ["u16", "protocol"],
//            ["u8", "is_static"],
//            ["u64", "last_heard"],
//            ["u64", "total_bytes"],
//            ["u32", "total_pkts"],
//            {"crc" : "0x9abeddd4"}
//        ],
//
type Nat44UserSessionDetails struct {
	OutsideIPAddress []byte `struc:"[4]byte"`
	OutsidePort      uint16
	InsideIPAddress  []byte `struc:"[4]byte"`
	InsidePort       uint16
	Protocol         uint16
	IsStatic         uint8
	LastHeard        uint64
	TotalBytes       uint64
	TotalPkts        uint32
}

func (*Nat44UserSessionDetails) GetMessageName() string {
	return "nat44_user_session_details"
}
func (*Nat44UserSessionDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44UserSessionDetails) GetCrcString() string {
	return "9abeddd4"
}
func NewNat44UserSessionDetails() api.Message {
	return &Nat44UserSessionDetails{}
}

// Nat44AddDelLbStaticMapping represents the VPP binary API message 'nat44_add_del_lb_static_mapping'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 373:
//
//        ["nat44_add_del_lb_static_mapping",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "external_addr", 4],
//            ["u16", "external_port"],
//            ["u8", "protocol"],
//            ["u32", "vrf_id"],
//            ["u8", "twice_nat"],
//            ["u8", "local_num"],
//            ["vl_api_nat44_lb_addr_port_t", "locals", 0, "local_num"],
//            {"crc" : "0xe74eb092"}
//        ],
//
type Nat44AddDelLbStaticMapping struct {
	IsAdd        uint8
	ExternalAddr []byte `struc:"[4]byte"`
	ExternalPort uint16
	Protocol     uint8
	VrfID        uint32
	TwiceNat     uint8
	LocalNum     uint8 `struc:"sizeof=Locals"`
	Locals       []Nat44LbAddrPort
}

func (*Nat44AddDelLbStaticMapping) GetMessageName() string {
	return "nat44_add_del_lb_static_mapping"
}
func (*Nat44AddDelLbStaticMapping) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44AddDelLbStaticMapping) GetCrcString() string {
	return "e74eb092"
}
func NewNat44AddDelLbStaticMapping() api.Message {
	return &Nat44AddDelLbStaticMapping{}
}

// Nat44AddDelLbStaticMappingReply represents the VPP binary API message 'nat44_add_del_lb_static_mapping_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 387:
//
//        ["nat44_add_del_lb_static_mapping_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xc24bab18"}
//        ],
//
type Nat44AddDelLbStaticMappingReply struct {
	Retval int32
}

func (*Nat44AddDelLbStaticMappingReply) GetMessageName() string {
	return "nat44_add_del_lb_static_mapping_reply"
}
func (*Nat44AddDelLbStaticMappingReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44AddDelLbStaticMappingReply) GetCrcString() string {
	return "c24bab18"
}
func NewNat44AddDelLbStaticMappingReply() api.Message {
	return &Nat44AddDelLbStaticMappingReply{}
}

// Nat44LbStaticMappingDump represents the VPP binary API message 'nat44_lb_static_mapping_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 393:
//
//        ["nat44_lb_static_mapping_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x1ddda4c8"}
//        ],
//
type Nat44LbStaticMappingDump struct {
}

func (*Nat44LbStaticMappingDump) GetMessageName() string {
	return "nat44_lb_static_mapping_dump"
}
func (*Nat44LbStaticMappingDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44LbStaticMappingDump) GetCrcString() string {
	return "1ddda4c8"
}
func NewNat44LbStaticMappingDump() api.Message {
	return &Nat44LbStaticMappingDump{}
}

// Nat44LbStaticMappingDetails represents the VPP binary API message 'nat44_lb_static_mapping_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 399:
//
//        ["nat44_lb_static_mapping_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "external_addr", 4],
//            ["u16", "external_port"],
//            ["u8", "protocol"],
//            ["u32", "vrf_id"],
//            ["u8", "twice_nat"],
//            ["u8", "local_num"],
//            ["vl_api_nat44_lb_addr_port_t", "locals", 0, "local_num"],
//            {"crc" : "0xfbcd1d1d"}
//        ],
//
type Nat44LbStaticMappingDetails struct {
	ExternalAddr []byte `struc:"[4]byte"`
	ExternalPort uint16
	Protocol     uint8
	VrfID        uint32
	TwiceNat     uint8
	LocalNum     uint8 `struc:"sizeof=Locals"`
	Locals       []Nat44LbAddrPort
}

func (*Nat44LbStaticMappingDetails) GetMessageName() string {
	return "nat44_lb_static_mapping_details"
}
func (*Nat44LbStaticMappingDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44LbStaticMappingDetails) GetCrcString() string {
	return "fbcd1d1d"
}
func NewNat44LbStaticMappingDetails() api.Message {
	return &Nat44LbStaticMappingDetails{}
}

// Nat44DelSession represents the VPP binary API message 'nat44_del_session'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 411:
//
//        ["nat44_del_session",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_in"],
//            ["u8", "address", 4],
//            ["u8", "protocol"],
//            ["u16", "port"],
//            ["u32", "vrf_id"],
//            {"crc" : "0x63e3de7c"}
//        ],
//
type Nat44DelSession struct {
	IsIn     uint8
	Address  []byte `struc:"[4]byte"`
	Protocol uint8
	Port     uint16
	VrfID    uint32
}

func (*Nat44DelSession) GetMessageName() string {
	return "nat44_del_session"
}
func (*Nat44DelSession) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44DelSession) GetCrcString() string {
	return "63e3de7c"
}
func NewNat44DelSession() api.Message {
	return &Nat44DelSession{}
}

// Nat44DelSessionReply represents the VPP binary API message 'nat44_del_session_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 422:
//
//        ["nat44_del_session_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x47b425e3"}
//        ],
//
type Nat44DelSessionReply struct {
	Retval int32
}

func (*Nat44DelSessionReply) GetMessageName() string {
	return "nat44_del_session_reply"
}
func (*Nat44DelSessionReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44DelSessionReply) GetCrcString() string {
	return "47b425e3"
}
func NewNat44DelSessionReply() api.Message {
	return &Nat44DelSessionReply{}
}

// Nat44ForwardingEnableDisable represents the VPP binary API message 'nat44_forwarding_enable_disable'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 428:
//
//        ["nat44_forwarding_enable_disable",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "enable"],
//            {"crc" : "0x206be3d2"}
//        ],
//
type Nat44ForwardingEnableDisable struct {
	Enable uint8
}

func (*Nat44ForwardingEnableDisable) GetMessageName() string {
	return "nat44_forwarding_enable_disable"
}
func (*Nat44ForwardingEnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44ForwardingEnableDisable) GetCrcString() string {
	return "206be3d2"
}
func NewNat44ForwardingEnableDisable() api.Message {
	return &Nat44ForwardingEnableDisable{}
}

// Nat44ForwardingEnableDisableReply represents the VPP binary API message 'nat44_forwarding_enable_disable_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 435:
//
//        ["nat44_forwarding_enable_disable_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xf2741f48"}
//        ],
//
type Nat44ForwardingEnableDisableReply struct {
	Retval int32
}

func (*Nat44ForwardingEnableDisableReply) GetMessageName() string {
	return "nat44_forwarding_enable_disable_reply"
}
func (*Nat44ForwardingEnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44ForwardingEnableDisableReply) GetCrcString() string {
	return "f2741f48"
}
func NewNat44ForwardingEnableDisableReply() api.Message {
	return &Nat44ForwardingEnableDisableReply{}
}

// Nat44ForwardingIsEnabled represents the VPP binary API message 'nat44_forwarding_is_enabled'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 441:
//
//        ["nat44_forwarding_is_enabled",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x7be486df"}
//        ],
//
type Nat44ForwardingIsEnabled struct {
}

func (*Nat44ForwardingIsEnabled) GetMessageName() string {
	return "nat44_forwarding_is_enabled"
}
func (*Nat44ForwardingIsEnabled) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat44ForwardingIsEnabled) GetCrcString() string {
	return "7be486df"
}
func NewNat44ForwardingIsEnabled() api.Message {
	return &Nat44ForwardingIsEnabled{}
}

// Nat44ForwardingIsEnabledReply represents the VPP binary API message 'nat44_forwarding_is_enabled_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 447:
//
//        ["nat44_forwarding_is_enabled_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "enabled"],
//            {"crc" : "0x4e43b2c3"}
//        ],
//
type Nat44ForwardingIsEnabledReply struct {
	Enabled uint8
}

func (*Nat44ForwardingIsEnabledReply) GetMessageName() string {
	return "nat44_forwarding_is_enabled_reply"
}
func (*Nat44ForwardingIsEnabledReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat44ForwardingIsEnabledReply) GetCrcString() string {
	return "4e43b2c3"
}
func NewNat44ForwardingIsEnabledReply() api.Message {
	return &Nat44ForwardingIsEnabledReply{}
}

// NatDetAddDelMap represents the VPP binary API message 'nat_det_add_del_map'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 453:
//
//        ["nat_det_add_del_map",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "is_nat44"],
//            ["u8", "addr_only"],
//            ["u8", "in_addr", 16],
//            ["u8", "in_plen"],
//            ["u8", "out_addr", 4],
//            ["u8", "out_plen"],
//            {"crc" : "0x477b07ed"}
//        ],
//
type NatDetAddDelMap struct {
	IsAdd    uint8
	IsNat44  uint8
	AddrOnly uint8
	InAddr   []byte `struc:"[16]byte"`
	InPlen   uint8
	OutAddr  []byte `struc:"[4]byte"`
	OutPlen  uint8
}

func (*NatDetAddDelMap) GetMessageName() string {
	return "nat_det_add_del_map"
}
func (*NatDetAddDelMap) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetAddDelMap) GetCrcString() string {
	return "477b07ed"
}
func NewNatDetAddDelMap() api.Message {
	return &NatDetAddDelMap{}
}

// NatDetAddDelMapReply represents the VPP binary API message 'nat_det_add_del_map_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 466:
//
//        ["nat_det_add_del_map_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x2fa49765"}
//        ],
//
type NatDetAddDelMapReply struct {
	Retval int32
}

func (*NatDetAddDelMapReply) GetMessageName() string {
	return "nat_det_add_del_map_reply"
}
func (*NatDetAddDelMapReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatDetAddDelMapReply) GetCrcString() string {
	return "2fa49765"
}
func NewNatDetAddDelMapReply() api.Message {
	return &NatDetAddDelMapReply{}
}

// NatDetForward represents the VPP binary API message 'nat_det_forward'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 472:
//
//        ["nat_det_forward",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_nat44"],
//            ["u8", "in_addr", 16],
//            {"crc" : "0x85b74d31"}
//        ],
//
type NatDetForward struct {
	IsNat44 uint8
	InAddr  []byte `struc:"[16]byte"`
}

func (*NatDetForward) GetMessageName() string {
	return "nat_det_forward"
}
func (*NatDetForward) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetForward) GetCrcString() string {
	return "85b74d31"
}
func NewNatDetForward() api.Message {
	return &NatDetForward{}
}

// NatDetForwardReply represents the VPP binary API message 'nat_det_forward_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 480:
//
//        ["nat_det_forward_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u16", "out_port_lo"],
//            ["u16", "out_port_hi"],
//            ["u8", "out_addr", 4],
//            {"crc" : "0x037762b9"}
//        ],
//
type NatDetForwardReply struct {
	Retval    int32
	OutPortLo uint16
	OutPortHi uint16
	OutAddr   []byte `struc:"[4]byte"`
}

func (*NatDetForwardReply) GetMessageName() string {
	return "nat_det_forward_reply"
}
func (*NatDetForwardReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatDetForwardReply) GetCrcString() string {
	return "037762b9"
}
func NewNatDetForwardReply() api.Message {
	return &NatDetForwardReply{}
}

// NatDetReverse represents the VPP binary API message 'nat_det_reverse'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 489:
//
//        ["nat_det_reverse",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u16", "out_port"],
//            ["u8", "out_addr", 4],
//            {"crc" : "0xbb55c1d1"}
//        ],
//
type NatDetReverse struct {
	OutPort uint16
	OutAddr []byte `struc:"[4]byte"`
}

func (*NatDetReverse) GetMessageName() string {
	return "nat_det_reverse"
}
func (*NatDetReverse) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetReverse) GetCrcString() string {
	return "bb55c1d1"
}
func NewNatDetReverse() api.Message {
	return &NatDetReverse{}
}

// NatDetReverseReply represents the VPP binary API message 'nat_det_reverse_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 497:
//
//        ["nat_det_reverse_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u8", "is_nat44"],
//            ["u8", "in_addr", 16],
//            {"crc" : "0x61420be4"}
//        ],
//
type NatDetReverseReply struct {
	Retval  int32
	IsNat44 uint8
	InAddr  []byte `struc:"[16]byte"`
}

func (*NatDetReverseReply) GetMessageName() string {
	return "nat_det_reverse_reply"
}
func (*NatDetReverseReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatDetReverseReply) GetCrcString() string {
	return "61420be4"
}
func NewNatDetReverseReply() api.Message {
	return &NatDetReverseReply{}
}

// NatDetMapDump represents the VPP binary API message 'nat_det_map_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 505:
//
//        ["nat_det_map_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x717c3593"}
//        ],
//
type NatDetMapDump struct {
}

func (*NatDetMapDump) GetMessageName() string {
	return "nat_det_map_dump"
}
func (*NatDetMapDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetMapDump) GetCrcString() string {
	return "717c3593"
}
func NewNatDetMapDump() api.Message {
	return &NatDetMapDump{}
}

// NatDetMapDetails represents the VPP binary API message 'nat_det_map_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 511:
//
//        ["nat_det_map_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "is_nat44"],
//            ["u8", "in_addr", 16],
//            ["u8", "in_plen"],
//            ["u8", "out_addr", 4],
//            ["u8", "out_plen"],
//            ["u32", "sharing_ratio"],
//            ["u16", "ports_per_host"],
//            ["u32", "ses_num"],
//            {"crc" : "0xc403648b"}
//        ],
//
type NatDetMapDetails struct {
	IsNat44      uint8
	InAddr       []byte `struc:"[16]byte"`
	InPlen       uint8
	OutAddr      []byte `struc:"[4]byte"`
	OutPlen      uint8
	SharingRatio uint32
	PortsPerHost uint16
	SesNum       uint32
}

func (*NatDetMapDetails) GetMessageName() string {
	return "nat_det_map_details"
}
func (*NatDetMapDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatDetMapDetails) GetCrcString() string {
	return "c403648b"
}
func NewNatDetMapDetails() api.Message {
	return &NatDetMapDetails{}
}

// NatDetSetTimeouts represents the VPP binary API message 'nat_det_set_timeouts'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 524:
//
//        ["nat_det_set_timeouts",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "udp"],
//            ["u32", "tcp_established"],
//            ["u32", "tcp_transitory"],
//            ["u32", "icmp"],
//            {"crc" : "0xf957576e"}
//        ],
//
type NatDetSetTimeouts struct {
	UDP            uint32
	TCPEstablished uint32
	TCPTransitory  uint32
	ICMP           uint32
}

func (*NatDetSetTimeouts) GetMessageName() string {
	return "nat_det_set_timeouts"
}
func (*NatDetSetTimeouts) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetSetTimeouts) GetCrcString() string {
	return "f957576e"
}
func NewNatDetSetTimeouts() api.Message {
	return &NatDetSetTimeouts{}
}

// NatDetSetTimeoutsReply represents the VPP binary API message 'nat_det_set_timeouts_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 534:
//
//        ["nat_det_set_timeouts_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x17a80a73"}
//        ],
//
type NatDetSetTimeoutsReply struct {
	Retval int32
}

func (*NatDetSetTimeoutsReply) GetMessageName() string {
	return "nat_det_set_timeouts_reply"
}
func (*NatDetSetTimeoutsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatDetSetTimeoutsReply) GetCrcString() string {
	return "17a80a73"
}
func NewNatDetSetTimeoutsReply() api.Message {
	return &NatDetSetTimeoutsReply{}
}

// NatDetGetTimeouts represents the VPP binary API message 'nat_det_get_timeouts'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 540:
//
//        ["nat_det_get_timeouts",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x8dac1e2d"}
//        ],
//
type NatDetGetTimeouts struct {
}

func (*NatDetGetTimeouts) GetMessageName() string {
	return "nat_det_get_timeouts"
}
func (*NatDetGetTimeouts) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetGetTimeouts) GetCrcString() string {
	return "8dac1e2d"
}
func NewNatDetGetTimeouts() api.Message {
	return &NatDetGetTimeouts{}
}

// NatDetGetTimeoutsReply represents the VPP binary API message 'nat_det_get_timeouts_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 546:
//
//        ["nat_det_get_timeouts_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u32", "udp"],
//            ["u32", "tcp_established"],
//            ["u32", "tcp_transitory"],
//            ["u32", "icmp"],
//            {"crc" : "0x4a25be37"}
//        ],
//
type NatDetGetTimeoutsReply struct {
	Retval         int32
	UDP            uint32
	TCPEstablished uint32
	TCPTransitory  uint32
	ICMP           uint32
}

func (*NatDetGetTimeoutsReply) GetMessageName() string {
	return "nat_det_get_timeouts_reply"
}
func (*NatDetGetTimeoutsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatDetGetTimeoutsReply) GetCrcString() string {
	return "4a25be37"
}
func NewNatDetGetTimeoutsReply() api.Message {
	return &NatDetGetTimeoutsReply{}
}

// NatDetCloseSessionOut represents the VPP binary API message 'nat_det_close_session_out'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 556:
//
//        ["nat_det_close_session_out",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "out_addr", 4],
//            ["u16", "out_port"],
//            ["u8", "ext_addr", 4],
//            ["u16", "ext_port"],
//            {"crc" : "0xfa5c6cc6"}
//        ],
//
type NatDetCloseSessionOut struct {
	OutAddr []byte `struc:"[4]byte"`
	OutPort uint16
	ExtAddr []byte `struc:"[4]byte"`
	ExtPort uint16
}

func (*NatDetCloseSessionOut) GetMessageName() string {
	return "nat_det_close_session_out"
}
func (*NatDetCloseSessionOut) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetCloseSessionOut) GetCrcString() string {
	return "fa5c6cc6"
}
func NewNatDetCloseSessionOut() api.Message {
	return &NatDetCloseSessionOut{}
}

// NatDetCloseSessionOutReply represents the VPP binary API message 'nat_det_close_session_out_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 566:
//
//        ["nat_det_close_session_out_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xff3961f2"}
//        ],
//
type NatDetCloseSessionOutReply struct {
	Retval int32
}

func (*NatDetCloseSessionOutReply) GetMessageName() string {
	return "nat_det_close_session_out_reply"
}
func (*NatDetCloseSessionOutReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatDetCloseSessionOutReply) GetCrcString() string {
	return "ff3961f2"
}
func NewNatDetCloseSessionOutReply() api.Message {
	return &NatDetCloseSessionOutReply{}
}

// NatDetCloseSessionIn represents the VPP binary API message 'nat_det_close_session_in'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 572:
//
//        ["nat_det_close_session_in",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_nat44"],
//            ["u8", "in_addr", 16],
//            ["u16", "in_port"],
//            ["u8", "ext_addr", 16],
//            ["u16", "ext_port"],
//            {"crc" : "0xff933811"}
//        ],
//
type NatDetCloseSessionIn struct {
	IsNat44 uint8
	InAddr  []byte `struc:"[16]byte"`
	InPort  uint16
	ExtAddr []byte `struc:"[16]byte"`
	ExtPort uint16
}

func (*NatDetCloseSessionIn) GetMessageName() string {
	return "nat_det_close_session_in"
}
func (*NatDetCloseSessionIn) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetCloseSessionIn) GetCrcString() string {
	return "ff933811"
}
func NewNatDetCloseSessionIn() api.Message {
	return &NatDetCloseSessionIn{}
}

// NatDetCloseSessionInReply represents the VPP binary API message 'nat_det_close_session_in_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 583:
//
//        ["nat_det_close_session_in_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x8d1e060c"}
//        ],
//
type NatDetCloseSessionInReply struct {
	Retval int32
}

func (*NatDetCloseSessionInReply) GetMessageName() string {
	return "nat_det_close_session_in_reply"
}
func (*NatDetCloseSessionInReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*NatDetCloseSessionInReply) GetCrcString() string {
	return "8d1e060c"
}
func NewNatDetCloseSessionInReply() api.Message {
	return &NatDetCloseSessionInReply{}
}

// NatDetSessionDump represents the VPP binary API message 'nat_det_session_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 589:
//
//        ["nat_det_session_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_nat44"],
//            ["u8", "user_addr", 16],
//            {"crc" : "0xe265f99d"}
//        ],
//
type NatDetSessionDump struct {
	IsNat44  uint8
	UserAddr []byte `struc:"[16]byte"`
}

func (*NatDetSessionDump) GetMessageName() string {
	return "nat_det_session_dump"
}
func (*NatDetSessionDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetSessionDump) GetCrcString() string {
	return "e265f99d"
}
func NewNatDetSessionDump() api.Message {
	return &NatDetSessionDump{}
}

// NatDetSessionDetails represents the VPP binary API message 'nat_det_session_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 597:
//
//        ["nat_det_session_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u16", "in_port"],
//            ["u8", "ext_addr", 4],
//            ["u16", "ext_port"],
//            ["u16", "out_port"],
//            ["u8", "state"],
//            ["u32", "expire"],
//            {"crc" : "0x7e7399c3"}
//        ],
//
type NatDetSessionDetails struct {
	InPort  uint16
	ExtAddr []byte `struc:"[4]byte"`
	ExtPort uint16
	OutPort uint16
	State   uint8
	Expire  uint32
}

func (*NatDetSessionDetails) GetMessageName() string {
	return "nat_det_session_details"
}
func (*NatDetSessionDetails) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*NatDetSessionDetails) GetCrcString() string {
	return "7e7399c3"
}
func NewNatDetSessionDetails() api.Message {
	return &NatDetSessionDetails{}
}

// Nat64AddDelPoolAddrRange represents the VPP binary API message 'nat64_add_del_pool_addr_range'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 609:
//
//        ["nat64_add_del_pool_addr_range",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "start_addr", 4],
//            ["u8", "end_addr", 4],
//            ["u32", "vrf_id"],
//            ["u8", "is_add"],
//            {"crc" : "0x669df3f0"}
//        ],
//
type Nat64AddDelPoolAddrRange struct {
	StartAddr []byte `struc:"[4]byte"`
	EndAddr   []byte `struc:"[4]byte"`
	VrfID     uint32
	IsAdd     uint8
}

func (*Nat64AddDelPoolAddrRange) GetMessageName() string {
	return "nat64_add_del_pool_addr_range"
}
func (*Nat64AddDelPoolAddrRange) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64AddDelPoolAddrRange) GetCrcString() string {
	return "669df3f0"
}
func NewNat64AddDelPoolAddrRange() api.Message {
	return &Nat64AddDelPoolAddrRange{}
}

// Nat64AddDelPoolAddrRangeReply represents the VPP binary API message 'nat64_add_del_pool_addr_range_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 619:
//
//        ["nat64_add_del_pool_addr_range_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x9c968408"}
//        ],
//
type Nat64AddDelPoolAddrRangeReply struct {
	Retval int32
}

func (*Nat64AddDelPoolAddrRangeReply) GetMessageName() string {
	return "nat64_add_del_pool_addr_range_reply"
}
func (*Nat64AddDelPoolAddrRangeReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64AddDelPoolAddrRangeReply) GetCrcString() string {
	return "9c968408"
}
func NewNat64AddDelPoolAddrRangeReply() api.Message {
	return &Nat64AddDelPoolAddrRangeReply{}
}

// Nat64PoolAddrDump represents the VPP binary API message 'nat64_pool_addr_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 625:
//
//        ["nat64_pool_addr_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x4ab68b4e"}
//        ],
//
type Nat64PoolAddrDump struct {
}

func (*Nat64PoolAddrDump) GetMessageName() string {
	return "nat64_pool_addr_dump"
}
func (*Nat64PoolAddrDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64PoolAddrDump) GetCrcString() string {
	return "4ab68b4e"
}
func NewNat64PoolAddrDump() api.Message {
	return &Nat64PoolAddrDump{}
}

// Nat64PoolAddrDetails represents the VPP binary API message 'nat64_pool_addr_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 631:
//
//        ["nat64_pool_addr_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "address", 4],
//            ["u32", "vrf_id"],
//            {"crc" : "0x235db4a3"}
//        ],
//
type Nat64PoolAddrDetails struct {
	Address []byte `struc:"[4]byte"`
	VrfID   uint32
}

func (*Nat64PoolAddrDetails) GetMessageName() string {
	return "nat64_pool_addr_details"
}
func (*Nat64PoolAddrDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64PoolAddrDetails) GetCrcString() string {
	return "235db4a3"
}
func NewNat64PoolAddrDetails() api.Message {
	return &Nat64PoolAddrDetails{}
}

// Nat64AddDelInterface represents the VPP binary API message 'nat64_add_del_interface'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 638:
//
//        ["nat64_add_del_interface",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_inside"],
//            ["u8", "is_add"],
//            {"crc" : "0xdf24a879"}
//        ],
//
type Nat64AddDelInterface struct {
	SwIfIndex uint32
	IsInside  uint8
	IsAdd     uint8
}

func (*Nat64AddDelInterface) GetMessageName() string {
	return "nat64_add_del_interface"
}
func (*Nat64AddDelInterface) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64AddDelInterface) GetCrcString() string {
	return "df24a879"
}
func NewNat64AddDelInterface() api.Message {
	return &Nat64AddDelInterface{}
}

// Nat64AddDelInterfaceReply represents the VPP binary API message 'nat64_add_del_interface_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 647:
//
//        ["nat64_add_del_interface_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xcfbb3de6"}
//        ],
//
type Nat64AddDelInterfaceReply struct {
	Retval int32
}

func (*Nat64AddDelInterfaceReply) GetMessageName() string {
	return "nat64_add_del_interface_reply"
}
func (*Nat64AddDelInterfaceReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64AddDelInterfaceReply) GetCrcString() string {
	return "cfbb3de6"
}
func NewNat64AddDelInterfaceReply() api.Message {
	return &Nat64AddDelInterfaceReply{}
}

// Nat64InterfaceDump represents the VPP binary API message 'nat64_interface_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 653:
//
//        ["nat64_interface_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x039fac3f"}
//        ],
//
type Nat64InterfaceDump struct {
}

func (*Nat64InterfaceDump) GetMessageName() string {
	return "nat64_interface_dump"
}
func (*Nat64InterfaceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64InterfaceDump) GetCrcString() string {
	return "039fac3f"
}
func NewNat64InterfaceDump() api.Message {
	return &Nat64InterfaceDump{}
}

// Nat64InterfaceDetails represents the VPP binary API message 'nat64_interface_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 659:
//
//        ["nat64_interface_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "is_inside"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x3d95fdbc"}
//        ],
//
type Nat64InterfaceDetails struct {
	IsInside  uint8
	SwIfIndex uint32
}

func (*Nat64InterfaceDetails) GetMessageName() string {
	return "nat64_interface_details"
}
func (*Nat64InterfaceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64InterfaceDetails) GetCrcString() string {
	return "3d95fdbc"
}
func NewNat64InterfaceDetails() api.Message {
	return &Nat64InterfaceDetails{}
}

// Nat64AddDelStaticBib represents the VPP binary API message 'nat64_add_del_static_bib'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 666:
//
//        ["nat64_add_del_static_bib",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "i_addr", 16],
//            ["u8", "o_addr", 4],
//            ["u16", "i_port"],
//            ["u16", "o_port"],
//            ["u32", "vrf_id"],
//            ["u8", "proto"],
//            ["u8", "is_add"],
//            {"crc" : "0xcb5fb6e9"}
//        ],
//
type Nat64AddDelStaticBib struct {
	IAddr []byte `struc:"[16]byte"`
	OAddr []byte `struc:"[4]byte"`
	IPort uint16
	OPort uint16
	VrfID uint32
	Proto uint8
	IsAdd uint8
}

func (*Nat64AddDelStaticBib) GetMessageName() string {
	return "nat64_add_del_static_bib"
}
func (*Nat64AddDelStaticBib) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64AddDelStaticBib) GetCrcString() string {
	return "cb5fb6e9"
}
func NewNat64AddDelStaticBib() api.Message {
	return &Nat64AddDelStaticBib{}
}

// Nat64AddDelStaticBibReply represents the VPP binary API message 'nat64_add_del_static_bib_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 679:
//
//        ["nat64_add_del_static_bib_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x7045bd17"}
//        ],
//
type Nat64AddDelStaticBibReply struct {
	Retval int32
}

func (*Nat64AddDelStaticBibReply) GetMessageName() string {
	return "nat64_add_del_static_bib_reply"
}
func (*Nat64AddDelStaticBibReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64AddDelStaticBibReply) GetCrcString() string {
	return "7045bd17"
}
func NewNat64AddDelStaticBibReply() api.Message {
	return &Nat64AddDelStaticBibReply{}
}

// Nat64BibDump represents the VPP binary API message 'nat64_bib_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 685:
//
//        ["nat64_bib_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "proto"],
//            {"crc" : "0xd48143dd"}
//        ],
//
type Nat64BibDump struct {
	Proto uint8
}

func (*Nat64BibDump) GetMessageName() string {
	return "nat64_bib_dump"
}
func (*Nat64BibDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64BibDump) GetCrcString() string {
	return "d48143dd"
}
func NewNat64BibDump() api.Message {
	return &Nat64BibDump{}
}

// Nat64BibDetails represents the VPP binary API message 'nat64_bib_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 692:
//
//        ["nat64_bib_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "i_addr", 16],
//            ["u8", "o_addr", 4],
//            ["u16", "i_port"],
//            ["u16", "o_port"],
//            ["u32", "vrf_id"],
//            ["u8", "proto"],
//            ["u8", "is_static"],
//            ["u32", "ses_num"],
//            {"crc" : "0xdc57f1a9"}
//        ],
//
type Nat64BibDetails struct {
	IAddr    []byte `struc:"[16]byte"`
	OAddr    []byte `struc:"[4]byte"`
	IPort    uint16
	OPort    uint16
	VrfID    uint32
	Proto    uint8
	IsStatic uint8
	SesNum   uint32
}

func (*Nat64BibDetails) GetMessageName() string {
	return "nat64_bib_details"
}
func (*Nat64BibDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64BibDetails) GetCrcString() string {
	return "dc57f1a9"
}
func NewNat64BibDetails() api.Message {
	return &Nat64BibDetails{}
}

// Nat64SetTimeouts represents the VPP binary API message 'nat64_set_timeouts'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 705:
//
//        ["nat64_set_timeouts",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "udp"],
//            ["u32", "icmp"],
//            ["u32", "tcp_trans"],
//            ["u32", "tcp_est"],
//            ["u32", "tcp_incoming_syn"],
//            {"crc" : "0x1b9f767e"}
//        ],
//
type Nat64SetTimeouts struct {
	UDP            uint32
	ICMP           uint32
	TCPTrans       uint32
	TCPEst         uint32
	TCPIncomingSyn uint32
}

func (*Nat64SetTimeouts) GetMessageName() string {
	return "nat64_set_timeouts"
}
func (*Nat64SetTimeouts) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64SetTimeouts) GetCrcString() string {
	return "1b9f767e"
}
func NewNat64SetTimeouts() api.Message {
	return &Nat64SetTimeouts{}
}

// Nat64SetTimeoutsReply represents the VPP binary API message 'nat64_set_timeouts_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 716:
//
//        ["nat64_set_timeouts_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xf8de0c6d"}
//        ],
//
type Nat64SetTimeoutsReply struct {
	Retval int32
}

func (*Nat64SetTimeoutsReply) GetMessageName() string {
	return "nat64_set_timeouts_reply"
}
func (*Nat64SetTimeoutsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64SetTimeoutsReply) GetCrcString() string {
	return "f8de0c6d"
}
func NewNat64SetTimeoutsReply() api.Message {
	return &Nat64SetTimeoutsReply{}
}

// Nat64GetTimeouts represents the VPP binary API message 'nat64_get_timeouts'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 722:
//
//        ["nat64_get_timeouts",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x62da1833"}
//        ],
//
type Nat64GetTimeouts struct {
}

func (*Nat64GetTimeouts) GetMessageName() string {
	return "nat64_get_timeouts"
}
func (*Nat64GetTimeouts) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64GetTimeouts) GetCrcString() string {
	return "62da1833"
}
func NewNat64GetTimeouts() api.Message {
	return &Nat64GetTimeouts{}
}

// Nat64GetTimeoutsReply represents the VPP binary API message 'nat64_get_timeouts_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 728:
//
//        ["nat64_get_timeouts_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u32", "udp"],
//            ["u32", "icmp"],
//            ["u32", "tcp_trans"],
//            ["u32", "tcp_est"],
//            ["u32", "tcp_incoming_syn"],
//            {"crc" : "0x3829fa1a"}
//        ],
//
type Nat64GetTimeoutsReply struct {
	Retval         int32
	UDP            uint32
	ICMP           uint32
	TCPTrans       uint32
	TCPEst         uint32
	TCPIncomingSyn uint32
}

func (*Nat64GetTimeoutsReply) GetMessageName() string {
	return "nat64_get_timeouts_reply"
}
func (*Nat64GetTimeoutsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64GetTimeoutsReply) GetCrcString() string {
	return "3829fa1a"
}
func NewNat64GetTimeoutsReply() api.Message {
	return &Nat64GetTimeoutsReply{}
}

// Nat64StDump represents the VPP binary API message 'nat64_st_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 739:
//
//        ["nat64_st_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "proto"],
//            {"crc" : "0xdf3cd00f"}
//        ],
//
type Nat64StDump struct {
	Proto uint8
}

func (*Nat64StDump) GetMessageName() string {
	return "nat64_st_dump"
}
func (*Nat64StDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64StDump) GetCrcString() string {
	return "df3cd00f"
}
func NewNat64StDump() api.Message {
	return &Nat64StDump{}
}

// Nat64StDetails represents the VPP binary API message 'nat64_st_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 746:
//
//        ["nat64_st_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "il_addr", 16],
//            ["u8", "ol_addr", 4],
//            ["u16", "il_port"],
//            ["u16", "ol_port"],
//            ["u8", "ir_addr", 16],
//            ["u8", "or_addr", 4],
//            ["u16", "r_port"],
//            ["u32", "vrf_id"],
//            ["u8", "proto"],
//            {"crc" : "0x5ac62548"}
//        ],
//
type Nat64StDetails struct {
	IlAddr []byte `struc:"[16]byte"`
	OlAddr []byte `struc:"[4]byte"`
	IlPort uint16
	OlPort uint16
	IrAddr []byte `struc:"[16]byte"`
	OrAddr []byte `struc:"[4]byte"`
	RPort  uint16
	VrfID  uint32
	Proto  uint8
}

func (*Nat64StDetails) GetMessageName() string {
	return "nat64_st_details"
}
func (*Nat64StDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64StDetails) GetCrcString() string {
	return "5ac62548"
}
func NewNat64StDetails() api.Message {
	return &Nat64StDetails{}
}

// Nat64AddDelPrefix represents the VPP binary API message 'nat64_add_del_prefix'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 760:
//
//        ["nat64_add_del_prefix",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "prefix", 16],
//            ["u8", "prefix_len"],
//            ["u32", "vrf_id"],
//            ["u8", "is_add"],
//            {"crc" : "0x1e126638"}
//        ],
//
type Nat64AddDelPrefix struct {
	Prefix    []byte `struc:"[16]byte"`
	PrefixLen uint8
	VrfID     uint32
	IsAdd     uint8
}

func (*Nat64AddDelPrefix) GetMessageName() string {
	return "nat64_add_del_prefix"
}
func (*Nat64AddDelPrefix) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64AddDelPrefix) GetCrcString() string {
	return "1e126638"
}
func NewNat64AddDelPrefix() api.Message {
	return &Nat64AddDelPrefix{}
}

// Nat64AddDelPrefixReply represents the VPP binary API message 'nat64_add_del_prefix_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 770:
//
//        ["nat64_add_del_prefix_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x098c2c7a"}
//        ],
//
type Nat64AddDelPrefixReply struct {
	Retval int32
}

func (*Nat64AddDelPrefixReply) GetMessageName() string {
	return "nat64_add_del_prefix_reply"
}
func (*Nat64AddDelPrefixReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64AddDelPrefixReply) GetCrcString() string {
	return "098c2c7a"
}
func NewNat64AddDelPrefixReply() api.Message {
	return &Nat64AddDelPrefixReply{}
}

// Nat64PrefixDump represents the VPP binary API message 'nat64_prefix_dump'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 776:
//
//        ["nat64_prefix_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x333a751d"}
//        ],
//
type Nat64PrefixDump struct {
}

func (*Nat64PrefixDump) GetMessageName() string {
	return "nat64_prefix_dump"
}
func (*Nat64PrefixDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64PrefixDump) GetCrcString() string {
	return "333a751d"
}
func NewNat64PrefixDump() api.Message {
	return &Nat64PrefixDump{}
}

// Nat64PrefixDetails represents the VPP binary API message 'nat64_prefix_details'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 782:
//
//        ["nat64_prefix_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "prefix", 16],
//            ["u8", "prefix_len"],
//            ["u32", "vrf_id"],
//            {"crc" : "0x521be2eb"}
//        ],
//
type Nat64PrefixDetails struct {
	Prefix    []byte `struc:"[16]byte"`
	PrefixLen uint8
	VrfID     uint32
}

func (*Nat64PrefixDetails) GetMessageName() string {
	return "nat64_prefix_details"
}
func (*Nat64PrefixDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64PrefixDetails) GetCrcString() string {
	return "521be2eb"
}
func NewNat64PrefixDetails() api.Message {
	return &Nat64PrefixDetails{}
}

// Nat64AddDelInterfaceAddr represents the VPP binary API message 'nat64_add_del_interface_addr'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 790:
//
//        ["nat64_add_del_interface_addr",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "is_inside"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0xd87095f4"}
//        ],
//
type Nat64AddDelInterfaceAddr struct {
	IsAdd     uint8
	IsInside  uint8
	SwIfIndex uint32
}

func (*Nat64AddDelInterfaceAddr) GetMessageName() string {
	return "nat64_add_del_interface_addr"
}
func (*Nat64AddDelInterfaceAddr) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*Nat64AddDelInterfaceAddr) GetCrcString() string {
	return "d87095f4"
}
func NewNat64AddDelInterfaceAddr() api.Message {
	return &Nat64AddDelInterfaceAddr{}
}

// Nat64AddDelInterfaceAddrReply represents the VPP binary API message 'nat64_add_del_interface_addr_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 799:
//
//        ["nat64_add_del_interface_addr_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x9fa4401a"}
//        ],
//
type Nat64AddDelInterfaceAddrReply struct {
	Retval int32
}

func (*Nat64AddDelInterfaceAddrReply) GetMessageName() string {
	return "nat64_add_del_interface_addr_reply"
}
func (*Nat64AddDelInterfaceAddrReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*Nat64AddDelInterfaceAddrReply) GetCrcString() string {
	return "9fa4401a"
}
func NewNat64AddDelInterfaceAddrReply() api.Message {
	return &Nat64AddDelInterfaceAddrReply{}
}

// DsliteAddDelPoolAddrRange represents the VPP binary API message 'dslite_add_del_pool_addr_range'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 805:
//
//        ["dslite_add_del_pool_addr_range",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "start_addr", 4],
//            ["u8", "end_addr", 4],
//            ["u8", "is_add"],
//            {"crc" : "0x179d6adf"}
//        ],
//
type DsliteAddDelPoolAddrRange struct {
	StartAddr []byte `struc:"[4]byte"`
	EndAddr   []byte `struc:"[4]byte"`
	IsAdd     uint8
}

func (*DsliteAddDelPoolAddrRange) GetMessageName() string {
	return "dslite_add_del_pool_addr_range"
}
func (*DsliteAddDelPoolAddrRange) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*DsliteAddDelPoolAddrRange) GetCrcString() string {
	return "179d6adf"
}
func NewDsliteAddDelPoolAddrRange() api.Message {
	return &DsliteAddDelPoolAddrRange{}
}

// DsliteAddDelPoolAddrRangeReply represents the VPP binary API message 'dslite_add_del_pool_addr_range_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 814:
//
//        ["dslite_add_del_pool_addr_range_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xb8de163d"}
//        ],
//
type DsliteAddDelPoolAddrRangeReply struct {
	Retval int32
}

func (*DsliteAddDelPoolAddrRangeReply) GetMessageName() string {
	return "dslite_add_del_pool_addr_range_reply"
}
func (*DsliteAddDelPoolAddrRangeReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*DsliteAddDelPoolAddrRangeReply) GetCrcString() string {
	return "b8de163d"
}
func NewDsliteAddDelPoolAddrRangeReply() api.Message {
	return &DsliteAddDelPoolAddrRangeReply{}
}

// DsliteSetAftrAddr represents the VPP binary API message 'dslite_set_aftr_addr'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 820:
//
//        ["dslite_set_aftr_addr",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "ip4_addr", 4],
//            ["u8", "ip6_addr", 16],
//            {"crc" : "0xe2e0c530"}
//        ],
//
type DsliteSetAftrAddr struct {
	IP4Addr []byte `struc:"[4]byte"`
	IP6Addr []byte `struc:"[16]byte"`
}

func (*DsliteSetAftrAddr) GetMessageName() string {
	return "dslite_set_aftr_addr"
}
func (*DsliteSetAftrAddr) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*DsliteSetAftrAddr) GetCrcString() string {
	return "e2e0c530"
}
func NewDsliteSetAftrAddr() api.Message {
	return &DsliteSetAftrAddr{}
}

// DsliteSetAftrAddrReply represents the VPP binary API message 'dslite_set_aftr_addr_reply'.
// Generated from '/usr/share/vpp/api/nat.api.json', line 828:
//
//        ["dslite_set_aftr_addr_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x87be299d"}
//        ]
//
type DsliteSetAftrAddrReply struct {
	Retval int32
}

func (*DsliteSetAftrAddrReply) GetMessageName() string {
	return "dslite_set_aftr_addr_reply"
}
func (*DsliteSetAftrAddrReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*DsliteSetAftrAddrReply) GetCrcString() string {
	return "87be299d"
}
func NewDsliteSetAftrAddrReply() api.Message {
	return &DsliteSetAftrAddrReply{}
}
