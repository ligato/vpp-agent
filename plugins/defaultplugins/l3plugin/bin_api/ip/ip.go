// Package ip represents the VPP binary API of the 'ip' VPP module.
// DO NOT EDIT. Generated from '/usr/share/vpp/api/ip.api.json'
package ip

import "git.fd.io/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0x3e4696de

// FibPath represents the VPP binary API data type 'fib_path'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 3:
//
//        ["fib_path",
//            ["u32", "sw_if_index"],
//            ["u8", "weight"],
//            ["u8", "preference"],
//            ["u8", "is_local"],
//            ["u8", "is_drop"],
//            ["u8", "is_unreach"],
//            ["u8", "is_prohibit"],
//            ["u8", "afi"],
//            ["u8", "next_hop", 16],
//            {"crc" : "0x39f4808f"}
//        ]
//
type FibPath struct {
	SwIfIndex  uint32
	Weight     uint8
	Preference uint8
	IsLocal    uint8
	IsDrop     uint8
	IsUnreach  uint8
	IsProhibit uint8
	Afi        uint8
	NextHop    []byte `struc:"[16]byte"`
}

func (*FibPath) GetTypeName() string {
	return "fib_path"
}
func (*FibPath) GetCrcString() string {
	return "39f4808f"
}

// IPTableAddDel represents the VPP binary API message 'ip_table_add_del'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 17:
//
//        ["ip_table_add_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "table_id"],
//            ["u8", "is_ipv6"],
//            ["u8", "is_add"],
//            {"crc" : "0xff1fffe6"}
//        ],
//
type IPTableAddDel struct {
	TableID uint32
	IsIpv6  uint8
	IsAdd   uint8
}

func (*IPTableAddDel) GetMessageName() string {
	return "ip_table_add_del"
}
func (*IPTableAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPTableAddDel) GetCrcString() string {
	return "ff1fffe6"
}
func NewIPTableAddDel() api.Message {
	return &IPTableAddDel{}
}

// IPTableAddDelReply represents the VPP binary API message 'ip_table_add_del_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 26:
//
//        ["ip_table_add_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x7da725be"}
//        ],
//
type IPTableAddDelReply struct {
	Retval int32
}

func (*IPTableAddDelReply) GetMessageName() string {
	return "ip_table_add_del_reply"
}
func (*IPTableAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IPTableAddDelReply) GetCrcString() string {
	return "7da725be"
}
func NewIPTableAddDelReply() api.Message {
	return &IPTableAddDelReply{}
}

// IPFibDump represents the VPP binary API message 'ip_fib_dump'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 32:
//
//        ["ip_fib_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x5fe56ca3"}
//        ],
//
type IPFibDump struct {
}

func (*IPFibDump) GetMessageName() string {
	return "ip_fib_dump"
}
func (*IPFibDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPFibDump) GetCrcString() string {
	return "5fe56ca3"
}
func NewIPFibDump() api.Message {
	return &IPFibDump{}
}

// IPFibDetails represents the VPP binary API message 'ip_fib_details'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 38:
//
//        ["ip_fib_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "table_id"],
//            ["u8", "address_length"],
//            ["u8", "address", 4],
//            ["u32", "count"],
//            ["vl_api_fib_path_t", "path", 0, "count"],
//            {"crc" : "0xfd8c6584"}
//        ],
//
type IPFibDetails struct {
	TableID       uint32
	AddressLength uint8
	Address       []byte `struc:"[4]byte"`
	Count         uint32 `struc:"sizeof=Path"`
	Path          []FibPath
}

func (*IPFibDetails) GetMessageName() string {
	return "ip_fib_details"
}
func (*IPFibDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IPFibDetails) GetCrcString() string {
	return "fd8c6584"
}
func NewIPFibDetails() api.Message {
	return &IPFibDetails{}
}

// IP6FibDump represents the VPP binary API message 'ip6_fib_dump'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 48:
//
//        ["ip6_fib_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x25c89676"}
//        ],
//
type IP6FibDump struct {
}

func (*IP6FibDump) GetMessageName() string {
	return "ip6_fib_dump"
}
func (*IP6FibDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IP6FibDump) GetCrcString() string {
	return "25c89676"
}
func NewIP6FibDump() api.Message {
	return &IP6FibDump{}
}

// IP6FibDetails represents the VPP binary API message 'ip6_fib_details'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 54:
//
//        ["ip6_fib_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "table_id"],
//            ["u8", "address_length"],
//            ["u8", "address", 16],
//            ["u32", "count"],
//            ["vl_api_fib_path_t", "path", 0, "count"],
//            {"crc" : "0xe0825cb5"}
//        ],
//
type IP6FibDetails struct {
	TableID       uint32
	AddressLength uint8
	Address       []byte `struc:"[16]byte"`
	Count         uint32 `struc:"sizeof=Path"`
	Path          []FibPath
}

func (*IP6FibDetails) GetMessageName() string {
	return "ip6_fib_details"
}
func (*IP6FibDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IP6FibDetails) GetCrcString() string {
	return "e0825cb5"
}
func NewIP6FibDetails() api.Message {
	return &IP6FibDetails{}
}

// IPNeighborDump represents the VPP binary API message 'ip_neighbor_dump'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 64:
//
//        ["ip_neighbor_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_ipv6"],
//            {"crc" : "0x3289e160"}
//        ],
//
type IPNeighborDump struct {
	SwIfIndex uint32
	IsIpv6    uint8
}

func (*IPNeighborDump) GetMessageName() string {
	return "ip_neighbor_dump"
}
func (*IPNeighborDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPNeighborDump) GetCrcString() string {
	return "3289e160"
}
func NewIPNeighborDump() api.Message {
	return &IPNeighborDump{}
}

// IPNeighborDetails represents the VPP binary API message 'ip_neighbor_details'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 72:
//
//        ["ip_neighbor_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u8", "is_static"],
//            ["u8", "is_ipv6"],
//            ["u8", "mac_address", 6],
//            ["u8", "ip_address", 16],
//            {"crc" : "0x3a00e32a"}
//        ],
//
type IPNeighborDetails struct {
	IsStatic   uint8
	IsIpv6     uint8
	MacAddress []byte `struc:"[6]byte"`
	IPAddress  []byte `struc:"[16]byte"`
}

func (*IPNeighborDetails) GetMessageName() string {
	return "ip_neighbor_details"
}
func (*IPNeighborDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IPNeighborDetails) GetCrcString() string {
	return "3a00e32a"
}
func NewIPNeighborDetails() api.Message {
	return &IPNeighborDetails{}
}

// IPNeighborAddDel represents the VPP binary API message 'ip_neighbor_add_del'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 81:
//
//        ["ip_neighbor_add_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_add"],
//            ["u8", "is_ipv6"],
//            ["u8", "is_static"],
//            ["u8", "is_no_adj_fib"],
//            ["u8", "mac_address", 6],
//            ["u8", "dst_address", 16],
//            {"crc" : "0x5a0d070b"}
//        ],
//
type IPNeighborAddDel struct {
	SwIfIndex  uint32
	IsAdd      uint8
	IsIpv6     uint8
	IsStatic   uint8
	IsNoAdjFib uint8
	MacAddress []byte `struc:"[6]byte"`
	DstAddress []byte `struc:"[16]byte"`
}

func (*IPNeighborAddDel) GetMessageName() string {
	return "ip_neighbor_add_del"
}
func (*IPNeighborAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPNeighborAddDel) GetCrcString() string {
	return "5a0d070b"
}
func NewIPNeighborAddDel() api.Message {
	return &IPNeighborAddDel{}
}

// IPNeighborAddDelReply represents the VPP binary API message 'ip_neighbor_add_del_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 94:
//
//        ["ip_neighbor_add_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xe5b0f318"}
//        ],
//
type IPNeighborAddDelReply struct {
	Retval int32
}

func (*IPNeighborAddDelReply) GetMessageName() string {
	return "ip_neighbor_add_del_reply"
}
func (*IPNeighborAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IPNeighborAddDelReply) GetCrcString() string {
	return "e5b0f318"
}
func NewIPNeighborAddDelReply() api.Message {
	return &IPNeighborAddDelReply{}
}

// SetIPFlowHash represents the VPP binary API message 'set_ip_flow_hash'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 100:
//
//        ["set_ip_flow_hash",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "vrf_id"],
//            ["u8", "is_ipv6"],
//            ["u8", "src"],
//            ["u8", "dst"],
//            ["u8", "sport"],
//            ["u8", "dport"],
//            ["u8", "proto"],
//            ["u8", "reverse"],
//            {"crc" : "0x92ad3798"}
//        ],
//
type SetIPFlowHash struct {
	VrfID   uint32
	IsIpv6  uint8
	Src     uint8
	Dst     uint8
	Sport   uint8
	Dport   uint8
	Proto   uint8
	Reverse uint8
}

func (*SetIPFlowHash) GetMessageName() string {
	return "set_ip_flow_hash"
}
func (*SetIPFlowHash) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SetIPFlowHash) GetCrcString() string {
	return "92ad3798"
}
func NewSetIPFlowHash() api.Message {
	return &SetIPFlowHash{}
}

// SetIPFlowHashReply represents the VPP binary API message 'set_ip_flow_hash_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 114:
//
//        ["set_ip_flow_hash_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x35a9e5eb"}
//        ],
//
type SetIPFlowHashReply struct {
	Retval int32
}

func (*SetIPFlowHashReply) GetMessageName() string {
	return "set_ip_flow_hash_reply"
}
func (*SetIPFlowHashReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SetIPFlowHashReply) GetCrcString() string {
	return "35a9e5eb"
}
func NewSetIPFlowHashReply() api.Message {
	return &SetIPFlowHashReply{}
}

// SwInterfaceIP6ndRaConfig represents the VPP binary API message 'sw_interface_ip6nd_ra_config'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 120:
//
//        ["sw_interface_ip6nd_ra_config",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "suppress"],
//            ["u8", "managed"],
//            ["u8", "other"],
//            ["u8", "ll_option"],
//            ["u8", "send_unicast"],
//            ["u8", "cease"],
//            ["u8", "is_no"],
//            ["u8", "default_router"],
//            ["u32", "max_interval"],
//            ["u32", "min_interval"],
//            ["u32", "lifetime"],
//            ["u32", "initial_count"],
//            ["u32", "initial_interval"],
//            {"crc" : "0xec4a29f6"}
//        ],
//
type SwInterfaceIP6ndRaConfig struct {
	SwIfIndex       uint32
	Suppress        uint8
	Managed         uint8
	Other           uint8
	LlOption        uint8
	SendUnicast     uint8
	Cease           uint8
	IsNo            uint8
	DefaultRouter   uint8
	MaxInterval     uint32
	MinInterval     uint32
	Lifetime        uint32
	InitialCount    uint32
	InitialInterval uint32
}

func (*SwInterfaceIP6ndRaConfig) GetMessageName() string {
	return "sw_interface_ip6nd_ra_config"
}
func (*SwInterfaceIP6ndRaConfig) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceIP6ndRaConfig) GetCrcString() string {
	return "ec4a29f6"
}
func NewSwInterfaceIP6ndRaConfig() api.Message {
	return &SwInterfaceIP6ndRaConfig{}
}

// SwInterfaceIP6ndRaConfigReply represents the VPP binary API message 'sw_interface_ip6nd_ra_config_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 140:
//
//        ["sw_interface_ip6nd_ra_config_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x16e25c5b"}
//        ],
//
type SwInterfaceIP6ndRaConfigReply struct {
	Retval int32
}

func (*SwInterfaceIP6ndRaConfigReply) GetMessageName() string {
	return "sw_interface_ip6nd_ra_config_reply"
}
func (*SwInterfaceIP6ndRaConfigReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceIP6ndRaConfigReply) GetCrcString() string {
	return "16e25c5b"
}
func NewSwInterfaceIP6ndRaConfigReply() api.Message {
	return &SwInterfaceIP6ndRaConfigReply{}
}

// SwInterfaceIP6ndRaPrefix represents the VPP binary API message 'sw_interface_ip6nd_ra_prefix'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 146:
//
//        ["sw_interface_ip6nd_ra_prefix",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "address", 16],
//            ["u8", "address_length"],
//            ["u8", "use_default"],
//            ["u8", "no_advertise"],
//            ["u8", "off_link"],
//            ["u8", "no_autoconfig"],
//            ["u8", "no_onlink"],
//            ["u8", "is_no"],
//            ["u32", "val_lifetime"],
//            ["u32", "pref_lifetime"],
//            {"crc" : "0x5db6555c"}
//        ],
//
type SwInterfaceIP6ndRaPrefix struct {
	SwIfIndex     uint32
	Address       []byte `struc:"[16]byte"`
	AddressLength uint8
	UseDefault    uint8
	NoAdvertise   uint8
	OffLink       uint8
	NoAutoconfig  uint8
	NoOnlink      uint8
	IsNo          uint8
	ValLifetime   uint32
	PrefLifetime  uint32
}

func (*SwInterfaceIP6ndRaPrefix) GetMessageName() string {
	return "sw_interface_ip6nd_ra_prefix"
}
func (*SwInterfaceIP6ndRaPrefix) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceIP6ndRaPrefix) GetCrcString() string {
	return "5db6555c"
}
func NewSwInterfaceIP6ndRaPrefix() api.Message {
	return &SwInterfaceIP6ndRaPrefix{}
}

// SwInterfaceIP6ndRaPrefixReply represents the VPP binary API message 'sw_interface_ip6nd_ra_prefix_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 163:
//
//        ["sw_interface_ip6nd_ra_prefix_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x8050adb3"}
//        ],
//
type SwInterfaceIP6ndRaPrefixReply struct {
	Retval int32
}

func (*SwInterfaceIP6ndRaPrefixReply) GetMessageName() string {
	return "sw_interface_ip6nd_ra_prefix_reply"
}
func (*SwInterfaceIP6ndRaPrefixReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceIP6ndRaPrefixReply) GetCrcString() string {
	return "8050adb3"
}
func NewSwInterfaceIP6ndRaPrefixReply() api.Message {
	return &SwInterfaceIP6ndRaPrefixReply{}
}

// IP6ndProxyAddDel represents the VPP binary API message 'ip6nd_proxy_add_del'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 169:
//
//        ["ip6nd_proxy_add_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_del"],
//            ["u8", "address", 16],
//            {"crc" : "0xc56f802d"}
//        ],
//
type IP6ndProxyAddDel struct {
	SwIfIndex uint32
	IsDel     uint8
	Address   []byte `struc:"[16]byte"`
}

func (*IP6ndProxyAddDel) GetMessageName() string {
	return "ip6nd_proxy_add_del"
}
func (*IP6ndProxyAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IP6ndProxyAddDel) GetCrcString() string {
	return "c56f802d"
}
func NewIP6ndProxyAddDel() api.Message {
	return &IP6ndProxyAddDel{}
}

// IP6ndProxyAddDelReply represents the VPP binary API message 'ip6nd_proxy_add_del_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 178:
//
//        ["ip6nd_proxy_add_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x00ddc2d5"}
//        ],
//
type IP6ndProxyAddDelReply struct {
	Retval int32
}

func (*IP6ndProxyAddDelReply) GetMessageName() string {
	return "ip6nd_proxy_add_del_reply"
}
func (*IP6ndProxyAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IP6ndProxyAddDelReply) GetCrcString() string {
	return "00ddc2d5"
}
func NewIP6ndProxyAddDelReply() api.Message {
	return &IP6ndProxyAddDelReply{}
}

// IP6ndProxyDetails represents the VPP binary API message 'ip6nd_proxy_details'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 184:
//
//        ["ip6nd_proxy_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "address", 16],
//            {"crc" : "0xf805ccc1"}
//        ],
//
type IP6ndProxyDetails struct {
	SwIfIndex uint32
	Address   []byte `struc:"[16]byte"`
}

func (*IP6ndProxyDetails) GetMessageName() string {
	return "ip6nd_proxy_details"
}
func (*IP6ndProxyDetails) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IP6ndProxyDetails) GetCrcString() string {
	return "f805ccc1"
}
func NewIP6ndProxyDetails() api.Message {
	return &IP6ndProxyDetails{}
}

// IP6ndProxyDump represents the VPP binary API message 'ip6nd_proxy_dump'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 192:
//
//        ["ip6nd_proxy_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x21597d88"}
//        ],
//
type IP6ndProxyDump struct {
}

func (*IP6ndProxyDump) GetMessageName() string {
	return "ip6nd_proxy_dump"
}
func (*IP6ndProxyDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IP6ndProxyDump) GetCrcString() string {
	return "21597d88"
}
func NewIP6ndProxyDump() api.Message {
	return &IP6ndProxyDump{}
}

// SwInterfaceIP6EnableDisable represents the VPP binary API message 'sw_interface_ip6_enable_disable'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 198:
//
//        ["sw_interface_ip6_enable_disable",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "enable"],
//            {"crc" : "0x4a4e5405"}
//        ],
//
type SwInterfaceIP6EnableDisable struct {
	SwIfIndex uint32
	Enable    uint8
}

func (*SwInterfaceIP6EnableDisable) GetMessageName() string {
	return "sw_interface_ip6_enable_disable"
}
func (*SwInterfaceIP6EnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceIP6EnableDisable) GetCrcString() string {
	return "4a4e5405"
}
func NewSwInterfaceIP6EnableDisable() api.Message {
	return &SwInterfaceIP6EnableDisable{}
}

// SwInterfaceIP6EnableDisableReply represents the VPP binary API message 'sw_interface_ip6_enable_disable_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 206:
//
//        ["sw_interface_ip6_enable_disable_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xeb8b4a40"}
//        ],
//
type SwInterfaceIP6EnableDisableReply struct {
	Retval int32
}

func (*SwInterfaceIP6EnableDisableReply) GetMessageName() string {
	return "sw_interface_ip6_enable_disable_reply"
}
func (*SwInterfaceIP6EnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceIP6EnableDisableReply) GetCrcString() string {
	return "eb8b4a40"
}
func NewSwInterfaceIP6EnableDisableReply() api.Message {
	return &SwInterfaceIP6EnableDisableReply{}
}

// SwInterfaceIP6SetLinkLocalAddress represents the VPP binary API message 'sw_interface_ip6_set_link_local_address'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 212:
//
//        ["sw_interface_ip6_set_link_local_address",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "address", 16],
//            {"crc" : "0x3db6d52b"}
//        ],
//
type SwInterfaceIP6SetLinkLocalAddress struct {
	SwIfIndex uint32
	Address   []byte `struc:"[16]byte"`
}

func (*SwInterfaceIP6SetLinkLocalAddress) GetMessageName() string {
	return "sw_interface_ip6_set_link_local_address"
}
func (*SwInterfaceIP6SetLinkLocalAddress) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceIP6SetLinkLocalAddress) GetCrcString() string {
	return "3db6d52b"
}
func NewSwInterfaceIP6SetLinkLocalAddress() api.Message {
	return &SwInterfaceIP6SetLinkLocalAddress{}
}

// SwInterfaceIP6SetLinkLocalAddressReply represents the VPP binary API message 'sw_interface_ip6_set_link_local_address_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 220:
//
//        ["sw_interface_ip6_set_link_local_address_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x0a781e17"}
//        ],
//
type SwInterfaceIP6SetLinkLocalAddressReply struct {
	Retval int32
}

func (*SwInterfaceIP6SetLinkLocalAddressReply) GetMessageName() string {
	return "sw_interface_ip6_set_link_local_address_reply"
}
func (*SwInterfaceIP6SetLinkLocalAddressReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceIP6SetLinkLocalAddressReply) GetCrcString() string {
	return "0a781e17"
}
func NewSwInterfaceIP6SetLinkLocalAddressReply() api.Message {
	return &SwInterfaceIP6SetLinkLocalAddressReply{}
}

// IPAddDelRoute represents the VPP binary API message 'ip_add_del_route'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 226:
//
//        ["ip_add_del_route",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "next_hop_sw_if_index"],
//            ["u32", "table_id"],
//            ["u32", "classify_table_index"],
//            ["u32", "next_hop_table_id"],
//            ["u8", "create_vrf_if_needed"],
//            ["u8", "is_add"],
//            ["u8", "is_drop"],
//            ["u8", "is_unreach"],
//            ["u8", "is_prohibit"],
//            ["u8", "is_ipv6"],
//            ["u8", "is_local"],
//            ["u8", "is_classify"],
//            ["u8", "is_multipath"],
//            ["u8", "is_resolve_host"],
//            ["u8", "is_resolve_attached"],
//            ["u8", "not_last"],
//            ["u8", "next_hop_weight"],
//            ["u8", "next_hop_preference"],
//            ["u8", "dst_address_length"],
//            ["u8", "dst_address", 16],
//            ["u8", "next_hop_address", 16],
//            ["u8", "next_hop_n_out_labels"],
//            ["u32", "next_hop_via_label"],
//            ["u32", "next_hop_out_label_stack", 0, "next_hop_n_out_labels"],
//            {"crc" : "0x93e81ee6"}
//        ],
//
type IPAddDelRoute struct {
	NextHopSwIfIndex     uint32
	TableID              uint32
	ClassifyTableIndex   uint32
	NextHopTableID       uint32
	CreateVrfIfNeeded    uint8
	IsAdd                uint8
	IsDrop               uint8
	IsUnreach            uint8
	IsProhibit           uint8
	IsIpv6               uint8
	IsLocal              uint8
	IsClassify           uint8
	IsMultipath          uint8
	IsResolveHost        uint8
	IsResolveAttached    uint8
	NotLast              uint8
	NextHopWeight        uint8
	NextHopPreference    uint8
	DstAddressLength     uint8
	DstAddress           []byte `struc:"[16]byte"`
	NextHopAddress       []byte `struc:"[16]byte"`
	NextHopNOutLabels    uint8  `struc:"sizeof=NextHopOutLabelStack"`
	NextHopViaLabel      uint32
	NextHopOutLabelStack []uint32
}

func (*IPAddDelRoute) GetMessageName() string {
	return "ip_add_del_route"
}
func (*IPAddDelRoute) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPAddDelRoute) GetCrcString() string {
	return "93e81ee6"
}
func NewIPAddDelRoute() api.Message {
	return &IPAddDelRoute{}
}

// IPAddDelRouteReply represents the VPP binary API message 'ip_add_del_route_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 256:
//
//        ["ip_add_del_route_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xea57492b"}
//        ],
//
type IPAddDelRouteReply struct {
	Retval int32
}

func (*IPAddDelRouteReply) GetMessageName() string {
	return "ip_add_del_route_reply"
}
func (*IPAddDelRouteReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IPAddDelRouteReply) GetCrcString() string {
	return "ea57492b"
}
func NewIPAddDelRouteReply() api.Message {
	return &IPAddDelRouteReply{}
}

// IPMrouteAddDel represents the VPP binary API message 'ip_mroute_add_del'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 262:
//
//        ["ip_mroute_add_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "next_hop_sw_if_index"],
//            ["u32", "table_id"],
//            ["u32", "entry_flags"],
//            ["u32", "itf_flags"],
//            ["u32", "rpf_id"],
//            ["u16", "grp_address_length"],
//            ["u8", "create_vrf_if_needed"],
//            ["u8", "is_add"],
//            ["u8", "is_ipv6"],
//            ["u8", "is_local"],
//            ["u8", "grp_address", 16],
//            ["u8", "src_address", 16],
//            {"crc" : "0x8f5f21a8"}
//        ],
//
type IPMrouteAddDel struct {
	NextHopSwIfIndex  uint32
	TableID           uint32
	EntryFlags        uint32
	ItfFlags          uint32
	RpfID             uint32
	GrpAddressLength  uint16
	CreateVrfIfNeeded uint8
	IsAdd             uint8
	IsIpv6            uint8
	IsLocal           uint8
	GrpAddress        []byte `struc:"[16]byte"`
	SrcAddress        []byte `struc:"[16]byte"`
}

func (*IPMrouteAddDel) GetMessageName() string {
	return "ip_mroute_add_del"
}
func (*IPMrouteAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPMrouteAddDel) GetCrcString() string {
	return "8f5f21a8"
}
func NewIPMrouteAddDel() api.Message {
	return &IPMrouteAddDel{}
}

// IPMrouteAddDelReply represents the VPP binary API message 'ip_mroute_add_del_reply'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 280:
//
//        ["ip_mroute_add_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x8cabe02c"}
//        ],
//
type IPMrouteAddDelReply struct {
	Retval int32
}

func (*IPMrouteAddDelReply) GetMessageName() string {
	return "ip_mroute_add_del_reply"
}
func (*IPMrouteAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IPMrouteAddDelReply) GetCrcString() string {
	return "8cabe02c"
}
func NewIPMrouteAddDelReply() api.Message {
	return &IPMrouteAddDelReply{}
}

// IPMfibDump represents the VPP binary API message 'ip_mfib_dump'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 286:
//
//        ["ip_mfib_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xee61390e"}
//        ],
//
type IPMfibDump struct {
}

func (*IPMfibDump) GetMessageName() string {
	return "ip_mfib_dump"
}
func (*IPMfibDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPMfibDump) GetCrcString() string {
	return "ee61390e"
}
func NewIPMfibDump() api.Message {
	return &IPMfibDump{}
}

// IPMfibDetails represents the VPP binary API message 'ip_mfib_details'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 292:
//
//        ["ip_mfib_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "table_id"],
//            ["u32", "entry_flags"],
//            ["u32", "rpf_id"],
//            ["u8", "address_length"],
//            ["u8", "grp_address", 4],
//            ["u8", "src_address", 4],
//            ["u32", "count"],
//            ["vl_api_fib_path_t", "path", 0, "count"],
//            {"crc" : "0x395e5699"}
//        ],
//
type IPMfibDetails struct {
	TableID       uint32
	EntryFlags    uint32
	RpfID         uint32
	AddressLength uint8
	GrpAddress    []byte `struc:"[4]byte"`
	SrcAddress    []byte `struc:"[4]byte"`
	Count         uint32 `struc:"sizeof=Path"`
	Path          []FibPath
}

func (*IPMfibDetails) GetMessageName() string {
	return "ip_mfib_details"
}
func (*IPMfibDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IPMfibDetails) GetCrcString() string {
	return "395e5699"
}
func NewIPMfibDetails() api.Message {
	return &IPMfibDetails{}
}

// IP6MfibDump represents the VPP binary API message 'ip6_mfib_dump'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 305:
//
//        ["ip6_mfib_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0x0839e143"}
//        ],
//
type IP6MfibDump struct {
}

func (*IP6MfibDump) GetMessageName() string {
	return "ip6_mfib_dump"
}
func (*IP6MfibDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IP6MfibDump) GetCrcString() string {
	return "0839e143"
}
func NewIP6MfibDump() api.Message {
	return &IP6MfibDump{}
}

// IP6MfibDetails represents the VPP binary API message 'ip6_mfib_details'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 311:
//
//        ["ip6_mfib_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "table_id"],
//            ["u8", "address_length"],
//            ["u8", "grp_address", 16],
//            ["u8", "src_address", 16],
//            ["u32", "count"],
//            ["vl_api_fib_path_t", "path", 0, "count"],
//            {"crc" : "0x921b153f"}
//        ],
//
type IP6MfibDetails struct {
	TableID       uint32
	AddressLength uint8
	GrpAddress    []byte `struc:"[16]byte"`
	SrcAddress    []byte `struc:"[16]byte"`
	Count         uint32 `struc:"sizeof=Path"`
	Path          []FibPath
}

func (*IP6MfibDetails) GetMessageName() string {
	return "ip6_mfib_details"
}
func (*IP6MfibDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*IP6MfibDetails) GetCrcString() string {
	return "921b153f"
}
func NewIP6MfibDetails() api.Message {
	return &IP6MfibDetails{}
}

// IPAddressDetails represents the VPP binary API message 'ip_address_details'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 322:
//
//        ["ip_address_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "ip", 16],
//            ["u8", "prefix_length"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_ipv6"],
//            {"crc" : "0x190d4266"}
//        ],
//
type IPAddressDetails struct {
	IP           []byte `struc:"[16]byte"`
	PrefixLength uint8
	SwIfIndex    uint32
	IsIpv6       uint8
}

func (*IPAddressDetails) GetMessageName() string {
	return "ip_address_details"
}
func (*IPAddressDetails) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPAddressDetails) GetCrcString() string {
	return "190d4266"
}
func NewIPAddressDetails() api.Message {
	return &IPAddressDetails{}
}

// IPAddressDump represents the VPP binary API message 'ip_address_dump'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 332:
//
//        ["ip_address_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_ipv6"],
//            {"crc" : "0x632e859a"}
//        ],
//
type IPAddressDump struct {
	SwIfIndex uint32
	IsIpv6    uint8
}

func (*IPAddressDump) GetMessageName() string {
	return "ip_address_dump"
}
func (*IPAddressDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPAddressDump) GetCrcString() string {
	return "632e859a"
}
func NewIPAddressDump() api.Message {
	return &IPAddressDump{}
}

// IPDetails represents the VPP binary API message 'ip_details'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 340:
//
//        ["ip_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "sw_if_index"],
//            ["u32", "context"],
//            ["u8", "is_ipv6"],
//            {"crc" : "0x695c8227"}
//        ],
//
type IPDetails struct {
	SwIfIndex uint32
	Context   uint32
	IsIpv6    uint8
}

func (*IPDetails) GetMessageName() string {
	return "ip_details"
}
func (*IPDetails) GetMessageType() api.MessageType {
	return api.OtherMessage
}
func (*IPDetails) GetCrcString() string {
	return "695c8227"
}
func NewIPDetails() api.Message {
	return &IPDetails{}
}

// IPDump represents the VPP binary API message 'ip_dump'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 347:
//
//        ["ip_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_ipv6"],
//            {"crc" : "0x3c1e33e0"}
//        ],
//
type IPDump struct {
	IsIpv6 uint8
}

func (*IPDump) GetMessageName() string {
	return "ip_dump"
}
func (*IPDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*IPDump) GetCrcString() string {
	return "3c1e33e0"
}
func NewIPDump() api.Message {
	return &IPDump{}
}

// MfibSignalDump represents the VPP binary API message 'mfib_signal_dump'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 354:
//
//        ["mfib_signal_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xbbbbd40d"}
//        ],
//
type MfibSignalDump struct {
}

func (*MfibSignalDump) GetMessageName() string {
	return "mfib_signal_dump"
}
func (*MfibSignalDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MfibSignalDump) GetCrcString() string {
	return "bbbbd40d"
}
func NewMfibSignalDump() api.Message {
	return &MfibSignalDump{}
}

// MfibSignalDetails represents the VPP binary API message 'mfib_signal_details'.
// Generated from '/usr/share/vpp/api/ip.api.json', line 360:
//
//        ["mfib_signal_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u32", "table_id"],
//            ["u16", "grp_address_len"],
//            ["u8", "grp_address", 16],
//            ["u8", "src_address", 16],
//            ["u16", "ip_packet_len"],
//            ["u8", "ip_packet_data", 256],
//            {"crc" : "0x6ba92c72"}
//        ]
//
type MfibSignalDetails struct {
	SwIfIndex     uint32
	TableID       uint32
	GrpAddressLen uint16
	GrpAddress    []byte `struc:"[16]byte"`
	SrcAddress    []byte `struc:"[16]byte"`
	IPPacketLen   uint16
	IPPacketData  []byte `struc:"[256]byte"`
}

func (*MfibSignalDetails) GetMessageName() string {
	return "mfib_signal_details"
}
func (*MfibSignalDetails) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*MfibSignalDetails) GetCrcString() string {
	return "6ba92c72"
}
func NewMfibSignalDetails() api.Message {
	return &MfibSignalDetails{}
}
