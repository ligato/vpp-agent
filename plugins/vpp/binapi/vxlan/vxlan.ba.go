// Code generated by GoVPP binapi-generator. DO NOT EDIT.
//  source: /usr/share/vpp/api/vxlan.api.json

/*
 Package vxlan is a generated from VPP binary API module 'vxlan'.

 It contains following objects:
	  4 services
	  8 messages
*/
package vxlan

import api "git.fd.io/govpp.git/api"
import struc "github.com/lunixbochs/struc"
import bytes "bytes"

// Reference imports to suppress errors if they are not otherwise used.
var _ = api.RegisterMessage
var _ = struc.Pack
var _ = bytes.NewBuffer

// Services represents VPP binary API services:
type Services interface {
	DumpVxlanTunnel(*VxlanTunnelDump) ([]*VxlanTunnelDetails, error)
	SwInterfaceSetVxlanBypass(*SwInterfaceSetVxlanBypass) (*SwInterfaceSetVxlanBypassReply, error)
	VxlanAddDelTunnel(*VxlanAddDelTunnel) (*VxlanAddDelTunnelReply, error)
	VxlanOffloadRx(*VxlanOffloadRx) (*VxlanOffloadRxReply, error)
}

/* Messages */

// SwInterfaceSetVxlanBypass represents VPP binary API message 'sw_interface_set_vxlan_bypass':
type SwInterfaceSetVxlanBypass struct {
	SwIfIndex uint32
	IsIPv6    uint8
	Enable    uint8
}

func (*SwInterfaceSetVxlanBypass) GetMessageName() string {
	return "sw_interface_set_vxlan_bypass"
}
func (*SwInterfaceSetVxlanBypass) GetCrcString() string {
	return "e74ca095"
}
func (*SwInterfaceSetVxlanBypass) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SwInterfaceSetVxlanBypassReply represents VPP binary API message 'sw_interface_set_vxlan_bypass_reply':
type SwInterfaceSetVxlanBypassReply struct {
	Retval int32
}

func (*SwInterfaceSetVxlanBypassReply) GetMessageName() string {
	return "sw_interface_set_vxlan_bypass_reply"
}
func (*SwInterfaceSetVxlanBypassReply) GetCrcString() string {
	return "e8d4e804"
}
func (*SwInterfaceSetVxlanBypassReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// VxlanAddDelTunnel represents VPP binary API message 'vxlan_add_del_tunnel':
type VxlanAddDelTunnel struct {
	IsAdd          uint8
	IsIPv6         uint8
	Instance       uint32
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
func (*VxlanAddDelTunnel) GetCrcString() string {
	return "00f4bdd0"
}
func (*VxlanAddDelTunnel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// VxlanAddDelTunnelReply represents VPP binary API message 'vxlan_add_del_tunnel_reply':
type VxlanAddDelTunnelReply struct {
	Retval    int32
	SwIfIndex uint32
}

func (*VxlanAddDelTunnelReply) GetMessageName() string {
	return "vxlan_add_del_tunnel_reply"
}
func (*VxlanAddDelTunnelReply) GetCrcString() string {
	return "fda5941f"
}
func (*VxlanAddDelTunnelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// VxlanOffloadRx represents VPP binary API message 'vxlan_offload_rx':
type VxlanOffloadRx struct {
	HwIfIndex uint32
	SwIfIndex uint32
	Enable    uint8
}

func (*VxlanOffloadRx) GetMessageName() string {
	return "vxlan_offload_rx"
}
func (*VxlanOffloadRx) GetCrcString() string {
	return "f0b08786"
}
func (*VxlanOffloadRx) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// VxlanOffloadRxReply represents VPP binary API message 'vxlan_offload_rx_reply':
type VxlanOffloadRxReply struct {
	Retval int32
}

func (*VxlanOffloadRxReply) GetMessageName() string {
	return "vxlan_offload_rx_reply"
}
func (*VxlanOffloadRxReply) GetCrcString() string {
	return "e8d4e804"
}
func (*VxlanOffloadRxReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// VxlanTunnelDetails represents VPP binary API message 'vxlan_tunnel_details':
type VxlanTunnelDetails struct {
	SwIfIndex      uint32
	Instance       uint32
	SrcAddress     []byte `struc:"[16]byte"`
	DstAddress     []byte `struc:"[16]byte"`
	McastSwIfIndex uint32
	EncapVrfID     uint32
	DecapNextIndex uint32
	Vni            uint32
	IsIPv6         uint8
}

func (*VxlanTunnelDetails) GetMessageName() string {
	return "vxlan_tunnel_details"
}
func (*VxlanTunnelDetails) GetCrcString() string {
	return "ce38e127"
}
func (*VxlanTunnelDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// VxlanTunnelDump represents VPP binary API message 'vxlan_tunnel_dump':
type VxlanTunnelDump struct {
	SwIfIndex uint32
}

func (*VxlanTunnelDump) GetMessageName() string {
	return "vxlan_tunnel_dump"
}
func (*VxlanTunnelDump) GetCrcString() string {
	return "529cb13f"
}
func (*VxlanTunnelDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func init() {
	api.RegisterMessage((*SwInterfaceSetVxlanBypass)(nil), "vxlan.SwInterfaceSetVxlanBypass")
	api.RegisterMessage((*SwInterfaceSetVxlanBypassReply)(nil), "vxlan.SwInterfaceSetVxlanBypassReply")
	api.RegisterMessage((*VxlanAddDelTunnel)(nil), "vxlan.VxlanAddDelTunnel")
	api.RegisterMessage((*VxlanAddDelTunnelReply)(nil), "vxlan.VxlanAddDelTunnelReply")
	api.RegisterMessage((*VxlanOffloadRx)(nil), "vxlan.VxlanOffloadRx")
	api.RegisterMessage((*VxlanOffloadRxReply)(nil), "vxlan.VxlanOffloadRxReply")
	api.RegisterMessage((*VxlanTunnelDetails)(nil), "vxlan.VxlanTunnelDetails")
	api.RegisterMessage((*VxlanTunnelDump)(nil), "vxlan.VxlanTunnelDump")
}

var Messages = []api.Message{
	(*SwInterfaceSetVxlanBypass)(nil),
	(*SwInterfaceSetVxlanBypassReply)(nil),
	(*VxlanAddDelTunnel)(nil),
	(*VxlanAddDelTunnelReply)(nil),
	(*VxlanOffloadRx)(nil),
	(*VxlanOffloadRxReply)(nil),
	(*VxlanTunnelDetails)(nil),
	(*VxlanTunnelDump)(nil),
}
