// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

// Package gre contains generated bindings for API file gre.api.
//
// Contents:
//   1 enum
//   1 struct
//   4 messages
//
package gre

import (
	"strconv"

	api "git.fd.io/govpp.git/api"
	codec "git.fd.io/govpp.git/codec"
	interface_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/interface_types"
	ip_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ip_types"
	tunnel_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/tunnel_types"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "gre"
	APIVersion = "2.1.0"
	VersionCrc = 0xb9021569
)

// GreTunnelType defines enum 'gre_tunnel_type'.
type GreTunnelType uint8

const (
	GRE_API_TUNNEL_TYPE_L3     GreTunnelType = 0
	GRE_API_TUNNEL_TYPE_TEB    GreTunnelType = 1
	GRE_API_TUNNEL_TYPE_ERSPAN GreTunnelType = 2
)

var (
	GreTunnelType_name = map[uint8]string{
		0: "GRE_API_TUNNEL_TYPE_L3",
		1: "GRE_API_TUNNEL_TYPE_TEB",
		2: "GRE_API_TUNNEL_TYPE_ERSPAN",
	}
	GreTunnelType_value = map[string]uint8{
		"GRE_API_TUNNEL_TYPE_L3":     0,
		"GRE_API_TUNNEL_TYPE_TEB":    1,
		"GRE_API_TUNNEL_TYPE_ERSPAN": 2,
	}
)

func (x GreTunnelType) String() string {
	s, ok := GreTunnelType_name[uint8(x)]
	if ok {
		return s
	}
	return "GreTunnelType(" + strconv.Itoa(int(x)) + ")"
}

// GreTunnel defines type 'gre_tunnel'.
type GreTunnel struct {
	Type         GreTunnelType                      `binapi:"gre_tunnel_type,name=type" json:"type,omitempty"`
	Mode         tunnel_types.TunnelMode            `binapi:"tunnel_mode,name=mode" json:"mode,omitempty"`
	Flags        tunnel_types.TunnelEncapDecapFlags `binapi:"tunnel_encap_decap_flags,name=flags" json:"flags,omitempty"`
	SessionID    uint16                             `binapi:"u16,name=session_id" json:"session_id,omitempty"`
	Instance     uint32                             `binapi:"u32,name=instance" json:"instance,omitempty"`
	OuterTableID uint32                             `binapi:"u32,name=outer_table_id" json:"outer_table_id,omitempty"`
	SwIfIndex    interface_types.InterfaceIndex     `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	Src          ip_types.Address                   `binapi:"address,name=src" json:"src,omitempty"`
	Dst          ip_types.Address                   `binapi:"address,name=dst" json:"dst,omitempty"`
}

// GreTunnelAddDel defines message 'gre_tunnel_add_del'.
type GreTunnelAddDel struct {
	IsAdd  bool      `binapi:"bool,name=is_add" json:"is_add,omitempty"`
	Tunnel GreTunnel `binapi:"gre_tunnel,name=tunnel" json:"tunnel,omitempty"`
}

func (m *GreTunnelAddDel) Reset()               { *m = GreTunnelAddDel{} }
func (*GreTunnelAddDel) GetMessageName() string { return "gre_tunnel_add_del" }
func (*GreTunnelAddDel) GetCrcString() string   { return "6efc9c22" }
func (*GreTunnelAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *GreTunnelAddDel) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1      // m.IsAdd
	size += 1      // m.Tunnel.Type
	size += 1      // m.Tunnel.Mode
	size += 1      // m.Tunnel.Flags
	size += 2      // m.Tunnel.SessionID
	size += 4      // m.Tunnel.Instance
	size += 4      // m.Tunnel.OuterTableID
	size += 4      // m.Tunnel.SwIfIndex
	size += 1      // m.Tunnel.Src.Af
	size += 1 * 16 // m.Tunnel.Src.Un
	size += 1      // m.Tunnel.Dst.Af
	size += 1 * 16 // m.Tunnel.Dst.Un
	return size
}
func (m *GreTunnelAddDel) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsAdd)
	buf.EncodeUint8(uint8(m.Tunnel.Type))
	buf.EncodeUint8(uint8(m.Tunnel.Mode))
	buf.EncodeUint8(uint8(m.Tunnel.Flags))
	buf.EncodeUint16(m.Tunnel.SessionID)
	buf.EncodeUint32(m.Tunnel.Instance)
	buf.EncodeUint32(m.Tunnel.OuterTableID)
	buf.EncodeUint32(uint32(m.Tunnel.SwIfIndex))
	buf.EncodeUint8(uint8(m.Tunnel.Src.Af))
	buf.EncodeBytes(m.Tunnel.Src.Un.XXX_UnionData[:], 16)
	buf.EncodeUint8(uint8(m.Tunnel.Dst.Af))
	buf.EncodeBytes(m.Tunnel.Dst.Un.XXX_UnionData[:], 16)
	return buf.Bytes(), nil
}
func (m *GreTunnelAddDel) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsAdd = buf.DecodeBool()
	m.Tunnel.Type = GreTunnelType(buf.DecodeUint8())
	m.Tunnel.Mode = tunnel_types.TunnelMode(buf.DecodeUint8())
	m.Tunnel.Flags = tunnel_types.TunnelEncapDecapFlags(buf.DecodeUint8())
	m.Tunnel.SessionID = buf.DecodeUint16()
	m.Tunnel.Instance = buf.DecodeUint32()
	m.Tunnel.OuterTableID = buf.DecodeUint32()
	m.Tunnel.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.Tunnel.Src.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.Tunnel.Src.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.Tunnel.Dst.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.Tunnel.Dst.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	return nil
}

// GreTunnelAddDelReply defines message 'gre_tunnel_add_del_reply'.
type GreTunnelAddDelReply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *GreTunnelAddDelReply) Reset()               { *m = GreTunnelAddDelReply{} }
func (*GreTunnelAddDelReply) GetMessageName() string { return "gre_tunnel_add_del_reply" }
func (*GreTunnelAddDelReply) GetCrcString() string   { return "5383d31f" }
func (*GreTunnelAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *GreTunnelAddDelReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *GreTunnelAddDelReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *GreTunnelAddDelReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// GreTunnelDetails defines message 'gre_tunnel_details'.
type GreTunnelDetails struct {
	Tunnel GreTunnel `binapi:"gre_tunnel,name=tunnel" json:"tunnel,omitempty"`
}

func (m *GreTunnelDetails) Reset()               { *m = GreTunnelDetails{} }
func (*GreTunnelDetails) GetMessageName() string { return "gre_tunnel_details" }
func (*GreTunnelDetails) GetCrcString() string   { return "003bfbf1" }
func (*GreTunnelDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *GreTunnelDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1      // m.Tunnel.Type
	size += 1      // m.Tunnel.Mode
	size += 1      // m.Tunnel.Flags
	size += 2      // m.Tunnel.SessionID
	size += 4      // m.Tunnel.Instance
	size += 4      // m.Tunnel.OuterTableID
	size += 4      // m.Tunnel.SwIfIndex
	size += 1      // m.Tunnel.Src.Af
	size += 1 * 16 // m.Tunnel.Src.Un
	size += 1      // m.Tunnel.Dst.Af
	size += 1 * 16 // m.Tunnel.Dst.Un
	return size
}
func (m *GreTunnelDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint8(uint8(m.Tunnel.Type))
	buf.EncodeUint8(uint8(m.Tunnel.Mode))
	buf.EncodeUint8(uint8(m.Tunnel.Flags))
	buf.EncodeUint16(m.Tunnel.SessionID)
	buf.EncodeUint32(m.Tunnel.Instance)
	buf.EncodeUint32(m.Tunnel.OuterTableID)
	buf.EncodeUint32(uint32(m.Tunnel.SwIfIndex))
	buf.EncodeUint8(uint8(m.Tunnel.Src.Af))
	buf.EncodeBytes(m.Tunnel.Src.Un.XXX_UnionData[:], 16)
	buf.EncodeUint8(uint8(m.Tunnel.Dst.Af))
	buf.EncodeBytes(m.Tunnel.Dst.Un.XXX_UnionData[:], 16)
	return buf.Bytes(), nil
}
func (m *GreTunnelDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Tunnel.Type = GreTunnelType(buf.DecodeUint8())
	m.Tunnel.Mode = tunnel_types.TunnelMode(buf.DecodeUint8())
	m.Tunnel.Flags = tunnel_types.TunnelEncapDecapFlags(buf.DecodeUint8())
	m.Tunnel.SessionID = buf.DecodeUint16()
	m.Tunnel.Instance = buf.DecodeUint32()
	m.Tunnel.OuterTableID = buf.DecodeUint32()
	m.Tunnel.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.Tunnel.Src.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.Tunnel.Src.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.Tunnel.Dst.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.Tunnel.Dst.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	return nil
}

// GreTunnelDump defines message 'gre_tunnel_dump'.
type GreTunnelDump struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *GreTunnelDump) Reset()               { *m = GreTunnelDump{} }
func (*GreTunnelDump) GetMessageName() string { return "gre_tunnel_dump" }
func (*GreTunnelDump) GetCrcString() string   { return "f9e6675e" }
func (*GreTunnelDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *GreTunnelDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *GreTunnelDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *GreTunnelDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

func init() { file_gre_binapi_init() }
func file_gre_binapi_init() {
	api.RegisterMessage((*GreTunnelAddDel)(nil), "gre_tunnel_add_del_6efc9c22")
	api.RegisterMessage((*GreTunnelAddDelReply)(nil), "gre_tunnel_add_del_reply_5383d31f")
	api.RegisterMessage((*GreTunnelDetails)(nil), "gre_tunnel_details_003bfbf1")
	api.RegisterMessage((*GreTunnelDump)(nil), "gre_tunnel_dump_f9e6675e")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*GreTunnelAddDel)(nil),
		(*GreTunnelAddDelReply)(nil),
		(*GreTunnelDetails)(nil),
		(*GreTunnelDump)(nil),
	}
}
