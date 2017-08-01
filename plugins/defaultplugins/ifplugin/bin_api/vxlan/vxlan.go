// Package vxlan represents the VPP binary API of the 'vxlan' VPP module.
// DO NOT EDIT. Generated from '/usr/share/vpp/api/vxlan.api.json'
package vxlan

import "git.fd.io/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0x1ca2f88d

// VxlanAddDelTunnel represents the VPP binary API message 'vxlan_add_del_tunnel'.
// Generated from '/usr/share/vpp/api/vxlan.api.json', line 6:
//
//        ["vxlan_add_del_tunnel",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u8", "is_ipv6"],
//            ["u8", "src_address", 16],
//            ["u8", "dst_address", 16],
//            ["u32", "mcast_sw_if_index"],
//            ["u32", "encap_vrf_id"],
//            ["u32", "decap_next_index"],
//            ["u32", "vni"],
//            {"crc" : "0x79be0753"}
//        ],
//
type VxlanAddDelTunnel struct {
	IsAdd          uint8
	IsIpv6         uint8
	SrcAddress     []byte `struc:"[16]byte"`
	DstAddress     []byte `struc:"[16]byte"`
	McastSwIfIndex uint32
	EncapVrfID     uint32
	DecapNextIndex uint32
	Vni            uint32
}

func (*VxlanAddDelTunnel) GetMessageName() string {
	return "vxlan_add_del_tunnel"
}
func (*VxlanAddDelTunnel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*VxlanAddDelTunnel) GetCrcString() string {
	return "79be0753"
}
func NewVxlanAddDelTunnel() api.Message {
	return &VxlanAddDelTunnel{}
}

// VxlanAddDelTunnelReply represents the VPP binary API message 'vxlan_add_del_tunnel_reply'.
// Generated from '/usr/share/vpp/api/vxlan.api.json', line 20:
//
//        ["vxlan_add_del_tunnel_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x3965e5df"}
//        ],
//
type VxlanAddDelTunnelReply struct {
	Retval    int32
	SwIfIndex uint32
}

func (*VxlanAddDelTunnelReply) GetMessageName() string {
	return "vxlan_add_del_tunnel_reply"
}
func (*VxlanAddDelTunnelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*VxlanAddDelTunnelReply) GetCrcString() string {
	return "3965e5df"
}
func NewVxlanAddDelTunnelReply() api.Message {
	return &VxlanAddDelTunnelReply{}
}

// VxlanTunnelDump represents the VPP binary API message 'vxlan_tunnel_dump'.
// Generated from '/usr/share/vpp/api/vxlan.api.json', line 27:
//
//        ["vxlan_tunnel_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x7d29e867"}
//        ],
//
type VxlanTunnelDump struct {
	SwIfIndex uint32
}

func (*VxlanTunnelDump) GetMessageName() string {
	return "vxlan_tunnel_dump"
}
func (*VxlanTunnelDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*VxlanTunnelDump) GetCrcString() string {
	return "7d29e867"
}
func NewVxlanTunnelDump() api.Message {
	return &VxlanTunnelDump{}
}

// VxlanTunnelDetails represents the VPP binary API message 'vxlan_tunnel_details'.
// Generated from '/usr/share/vpp/api/vxlan.api.json', line 34:
//
//        ["vxlan_tunnel_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "src_address", 16],
//            ["u8", "dst_address", 16],
//            ["u32", "mcast_sw_if_index"],
//            ["u32", "encap_vrf_id"],
//            ["u32", "decap_next_index"],
//            ["u32", "vni"],
//            ["u8", "is_ipv6"],
//            {"crc" : "0xfa28d42c"}
//        ],
//
type VxlanTunnelDetails struct {
	SwIfIndex      uint32
	SrcAddress     []byte `struc:"[16]byte"`
	DstAddress     []byte `struc:"[16]byte"`
	McastSwIfIndex uint32
	EncapVrfID     uint32
	DecapNextIndex uint32
	Vni            uint32
	IsIpv6         uint8
}

func (*VxlanTunnelDetails) GetMessageName() string {
	return "vxlan_tunnel_details"
}
func (*VxlanTunnelDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*VxlanTunnelDetails) GetCrcString() string {
	return "fa28d42c"
}
func NewVxlanTunnelDetails() api.Message {
	return &VxlanTunnelDetails{}
}

// SwInterfaceSetVxlanBypass represents the VPP binary API message 'sw_interface_set_vxlan_bypass'.
// Generated from '/usr/share/vpp/api/vxlan.api.json', line 47:
//
//        ["sw_interface_set_vxlan_bypass",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_ipv6"],
//            ["u8", "enable"],
//            {"crc" : "0xda63ecfd"}
//        ],
//
type SwInterfaceSetVxlanBypass struct {
	SwIfIndex uint32
	IsIpv6    uint8
	Enable    uint8
}

func (*SwInterfaceSetVxlanBypass) GetMessageName() string {
	return "sw_interface_set_vxlan_bypass"
}
func (*SwInterfaceSetVxlanBypass) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceSetVxlanBypass) GetCrcString() string {
	return "da63ecfd"
}
func NewSwInterfaceSetVxlanBypass() api.Message {
	return &SwInterfaceSetVxlanBypass{}
}

// SwInterfaceSetVxlanBypassReply represents the VPP binary API message 'sw_interface_set_vxlan_bypass_reply'.
// Generated from '/usr/share/vpp/api/vxlan.api.json', line 56:
//
//        ["sw_interface_set_vxlan_bypass_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xc4609ab5"}
//        ]
//
type SwInterfaceSetVxlanBypassReply struct {
	Retval int32
}

func (*SwInterfaceSetVxlanBypassReply) GetMessageName() string {
	return "sw_interface_set_vxlan_bypass_reply"
}
func (*SwInterfaceSetVxlanBypassReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceSetVxlanBypassReply) GetCrcString() string {
	return "c4609ab5"
}
func NewSwInterfaceSetVxlanBypassReply() api.Message {
	return &SwInterfaceSetVxlanBypassReply{}
}
