// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/memclnt.api.json

/*
Package memclnt is a generated VPP binary API for 'memclnt' module.

It consists of:
	  2 types
	 22 messages
	 13 services
*/
package memclnt

import (
	"bytes"
	"context"
	"io"
	"strconv"

	api "git.fd.io/govpp.git/api"
	struc "github.com/lunixbochs/struc"
)

const (
	// ModuleName is the name of this module.
	ModuleName = "memclnt"
	// APIVersion is the API version of this module.
	APIVersion = "2.1.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0x8d3dd881
)

// MessageTableEntry represents VPP binary API type 'message_table_entry'.
type MessageTableEntry struct {
	Index uint16
	Name  string `struc:"[64]byte"`
}

func (*MessageTableEntry) GetTypeName() string { return "message_table_entry" }

// ModuleVersion represents VPP binary API type 'module_version'.
type ModuleVersion struct {
	Major uint32
	Minor uint32
	Patch uint32
	Name  string `struc:"[64]byte"`
}

func (*ModuleVersion) GetTypeName() string { return "module_version" }

// APIVersions represents VPP binary API message 'api_versions'.
type APIVersions struct{}

func (m *APIVersions) Reset()                        { *m = APIVersions{} }
func (*APIVersions) GetMessageName() string          { return "api_versions" }
func (*APIVersions) GetCrcString() string            { return "51077d14" }
func (*APIVersions) GetMessageType() api.MessageType { return api.RequestMessage }

// APIVersionsReply represents VPP binary API message 'api_versions_reply'.
type APIVersionsReply struct {
	Retval      int32
	Count       uint32 `struc:"sizeof=APIVersions"`
	APIVersions []ModuleVersion
}

func (m *APIVersionsReply) Reset()                        { *m = APIVersionsReply{} }
func (*APIVersionsReply) GetMessageName() string          { return "api_versions_reply" }
func (*APIVersionsReply) GetCrcString() string            { return "5f0d99d6" }
func (*APIVersionsReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// GetFirstMsgID represents VPP binary API message 'get_first_msg_id'.
type GetFirstMsgID struct {
	Name string `struc:"[64]byte"`
}

func (m *GetFirstMsgID) Reset()                        { *m = GetFirstMsgID{} }
func (*GetFirstMsgID) GetMessageName() string          { return "get_first_msg_id" }
func (*GetFirstMsgID) GetCrcString() string            { return "ebf79a66" }
func (*GetFirstMsgID) GetMessageType() api.MessageType { return api.RequestMessage }

// GetFirstMsgIDReply represents VPP binary API message 'get_first_msg_id_reply'.
type GetFirstMsgIDReply struct {
	Retval     int32
	FirstMsgID uint16
}

func (m *GetFirstMsgIDReply) Reset()                        { *m = GetFirstMsgIDReply{} }
func (*GetFirstMsgIDReply) GetMessageName() string          { return "get_first_msg_id_reply" }
func (*GetFirstMsgIDReply) GetCrcString() string            { return "7d337472" }
func (*GetFirstMsgIDReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// MemclntCreate represents VPP binary API message 'memclnt_create'.
type MemclntCreate struct {
	CtxQuota    int32
	InputQueue  uint64
	Name        string   `struc:"[64]byte"`
	APIVersions []uint32 `struc:"[8]uint32"`
}

func (m *MemclntCreate) Reset()                        { *m = MemclntCreate{} }
func (*MemclntCreate) GetMessageName() string          { return "memclnt_create" }
func (*MemclntCreate) GetCrcString() string            { return "9c5e1c2f" }
func (*MemclntCreate) GetMessageType() api.MessageType { return api.ReplyMessage }

// MemclntCreateReply represents VPP binary API message 'memclnt_create_reply'.
type MemclntCreateReply struct {
	Response     int32
	Handle       uint64
	Index        uint32
	MessageTable uint64
}

func (m *MemclntCreateReply) Reset()                        { *m = MemclntCreateReply{} }
func (*MemclntCreateReply) GetMessageName() string          { return "memclnt_create_reply" }
func (*MemclntCreateReply) GetCrcString() string            { return "42ec4560" }
func (*MemclntCreateReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// MemclntDelete represents VPP binary API message 'memclnt_delete'.
type MemclntDelete struct {
	Index     uint32
	Handle    uint64
	DoCleanup bool
}

func (m *MemclntDelete) Reset()                        { *m = MemclntDelete{} }
func (*MemclntDelete) GetMessageName() string          { return "memclnt_delete" }
func (*MemclntDelete) GetCrcString() string            { return "7e1c04e3" }
func (*MemclntDelete) GetMessageType() api.MessageType { return api.OtherMessage }

// MemclntDeleteReply represents VPP binary API message 'memclnt_delete_reply'.
type MemclntDeleteReply struct {
	Response int32
	Handle   uint64
}

func (m *MemclntDeleteReply) Reset()                        { *m = MemclntDeleteReply{} }
func (*MemclntDeleteReply) GetMessageName() string          { return "memclnt_delete_reply" }
func (*MemclntDeleteReply) GetCrcString() string            { return "3d3b6312" }
func (*MemclntDeleteReply) GetMessageType() api.MessageType { return api.OtherMessage }

// MemclntKeepalive represents VPP binary API message 'memclnt_keepalive'.
type MemclntKeepalive struct{}

func (m *MemclntKeepalive) Reset()                        { *m = MemclntKeepalive{} }
func (*MemclntKeepalive) GetMessageName() string          { return "memclnt_keepalive" }
func (*MemclntKeepalive) GetCrcString() string            { return "51077d14" }
func (*MemclntKeepalive) GetMessageType() api.MessageType { return api.RequestMessage }

// MemclntKeepaliveReply represents VPP binary API message 'memclnt_keepalive_reply'.
type MemclntKeepaliveReply struct {
	Retval int32
}

func (m *MemclntKeepaliveReply) Reset()                        { *m = MemclntKeepaliveReply{} }
func (*MemclntKeepaliveReply) GetMessageName() string          { return "memclnt_keepalive_reply" }
func (*MemclntKeepaliveReply) GetCrcString() string            { return "e8d4e804" }
func (*MemclntKeepaliveReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// MemclntReadTimeout represents VPP binary API message 'memclnt_read_timeout'.
type MemclntReadTimeout struct {
	Dummy uint8
}

func (m *MemclntReadTimeout) Reset()                        { *m = MemclntReadTimeout{} }
func (*MemclntReadTimeout) GetMessageName() string          { return "memclnt_read_timeout" }
func (*MemclntReadTimeout) GetCrcString() string            { return "c3a3a452" }
func (*MemclntReadTimeout) GetMessageType() api.MessageType { return api.OtherMessage }

// MemclntRxThreadSuspend represents VPP binary API message 'memclnt_rx_thread_suspend'.
type MemclntRxThreadSuspend struct {
	Dummy uint8
}

func (m *MemclntRxThreadSuspend) Reset()                        { *m = MemclntRxThreadSuspend{} }
func (*MemclntRxThreadSuspend) GetMessageName() string          { return "memclnt_rx_thread_suspend" }
func (*MemclntRxThreadSuspend) GetCrcString() string            { return "c3a3a452" }
func (*MemclntRxThreadSuspend) GetMessageType() api.MessageType { return api.OtherMessage }

// RPCCall represents VPP binary API message 'rpc_call'.
type RPCCall struct {
	Function        uint64
	Multicast       uint8
	NeedBarrierSync uint8
	SendReply       uint8
	DataLen         uint32 `struc:"sizeof=Data"`
	Data            []byte
}

func (m *RPCCall) Reset()                        { *m = RPCCall{} }
func (*RPCCall) GetMessageName() string          { return "rpc_call" }
func (*RPCCall) GetCrcString() string            { return "7e8a2c95" }
func (*RPCCall) GetMessageType() api.MessageType { return api.RequestMessage }

// RPCCallReply represents VPP binary API message 'rpc_call_reply'.
type RPCCallReply struct {
	Retval int32
}

func (m *RPCCallReply) Reset()                        { *m = RPCCallReply{} }
func (*RPCCallReply) GetMessageName() string          { return "rpc_call_reply" }
func (*RPCCallReply) GetCrcString() string            { return "e8d4e804" }
func (*RPCCallReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// RxThreadExit represents VPP binary API message 'rx_thread_exit'.
type RxThreadExit struct {
	Dummy uint8
}

func (m *RxThreadExit) Reset()                        { *m = RxThreadExit{} }
func (*RxThreadExit) GetMessageName() string          { return "rx_thread_exit" }
func (*RxThreadExit) GetCrcString() string            { return "c3a3a452" }
func (*RxThreadExit) GetMessageType() api.MessageType { return api.OtherMessage }

// SockInitShm represents VPP binary API message 'sock_init_shm'.
type SockInitShm struct {
	RequestedSize uint32
	Nitems        uint8 `struc:"sizeof=Configs"`
	Configs       []uint64
}

func (m *SockInitShm) Reset()                        { *m = SockInitShm{} }
func (*SockInitShm) GetMessageName() string          { return "sock_init_shm" }
func (*SockInitShm) GetCrcString() string            { return "51646d92" }
func (*SockInitShm) GetMessageType() api.MessageType { return api.RequestMessage }

// SockInitShmReply represents VPP binary API message 'sock_init_shm_reply'.
type SockInitShmReply struct {
	Retval int32
}

func (m *SockInitShmReply) Reset()                        { *m = SockInitShmReply{} }
func (*SockInitShmReply) GetMessageName() string          { return "sock_init_shm_reply" }
func (*SockInitShmReply) GetCrcString() string            { return "e8d4e804" }
func (*SockInitShmReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// SockclntCreate represents VPP binary API message 'sockclnt_create'.
type SockclntCreate struct {
	Name string `struc:"[64]byte"`
}

func (m *SockclntCreate) Reset()                        { *m = SockclntCreate{} }
func (*SockclntCreate) GetMessageName() string          { return "sockclnt_create" }
func (*SockclntCreate) GetCrcString() string            { return "455fb9c4" }
func (*SockclntCreate) GetMessageType() api.MessageType { return api.ReplyMessage }

// SockclntCreateReply represents VPP binary API message 'sockclnt_create_reply'.
type SockclntCreateReply struct {
	Response     int32
	Index        uint32
	Count        uint16 `struc:"sizeof=MessageTable"`
	MessageTable []MessageTableEntry
}

func (m *SockclntCreateReply) Reset()                        { *m = SockclntCreateReply{} }
func (*SockclntCreateReply) GetMessageName() string          { return "sockclnt_create_reply" }
func (*SockclntCreateReply) GetCrcString() string            { return "35166268" }
func (*SockclntCreateReply) GetMessageType() api.MessageType { return api.RequestMessage }

// SockclntDelete represents VPP binary API message 'sockclnt_delete'.
type SockclntDelete struct {
	Index uint32
}

func (m *SockclntDelete) Reset()                        { *m = SockclntDelete{} }
func (*SockclntDelete) GetMessageName() string          { return "sockclnt_delete" }
func (*SockclntDelete) GetCrcString() string            { return "8ac76db6" }
func (*SockclntDelete) GetMessageType() api.MessageType { return api.RequestMessage }

// SockclntDeleteReply represents VPP binary API message 'sockclnt_delete_reply'.
type SockclntDeleteReply struct {
	Response int32
}

func (m *SockclntDeleteReply) Reset()                        { *m = SockclntDeleteReply{} }
func (*SockclntDeleteReply) GetMessageName() string          { return "sockclnt_delete_reply" }
func (*SockclntDeleteReply) GetCrcString() string            { return "8f38b1ee" }
func (*SockclntDeleteReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// TracePluginMsgIds represents VPP binary API message 'trace_plugin_msg_ids'.
type TracePluginMsgIds struct {
	PluginName string `struc:"[128]byte"`
	FirstMsgID uint16
	LastMsgID  uint16
}

func (m *TracePluginMsgIds) Reset()                        { *m = TracePluginMsgIds{} }
func (*TracePluginMsgIds) GetMessageName() string          { return "trace_plugin_msg_ids" }
func (*TracePluginMsgIds) GetCrcString() string            { return "f476d3ce" }
func (*TracePluginMsgIds) GetMessageType() api.MessageType { return api.RequestMessage }

func init() {
	api.RegisterMessage((*APIVersions)(nil), "memclnt.APIVersions")
	api.RegisterMessage((*APIVersionsReply)(nil), "memclnt.APIVersionsReply")
	api.RegisterMessage((*GetFirstMsgID)(nil), "memclnt.GetFirstMsgID")
	api.RegisterMessage((*GetFirstMsgIDReply)(nil), "memclnt.GetFirstMsgIDReply")
	api.RegisterMessage((*MemclntCreate)(nil), "memclnt.MemclntCreate")
	api.RegisterMessage((*MemclntCreateReply)(nil), "memclnt.MemclntCreateReply")
	api.RegisterMessage((*MemclntDelete)(nil), "memclnt.MemclntDelete")
	api.RegisterMessage((*MemclntDeleteReply)(nil), "memclnt.MemclntDeleteReply")
	api.RegisterMessage((*MemclntKeepalive)(nil), "memclnt.MemclntKeepalive")
	api.RegisterMessage((*MemclntKeepaliveReply)(nil), "memclnt.MemclntKeepaliveReply")
	api.RegisterMessage((*MemclntReadTimeout)(nil), "memclnt.MemclntReadTimeout")
	api.RegisterMessage((*MemclntRxThreadSuspend)(nil), "memclnt.MemclntRxThreadSuspend")
	api.RegisterMessage((*RPCCall)(nil), "memclnt.RPCCall")
	api.RegisterMessage((*RPCCallReply)(nil), "memclnt.RPCCallReply")
	api.RegisterMessage((*RxThreadExit)(nil), "memclnt.RxThreadExit")
	api.RegisterMessage((*SockInitShm)(nil), "memclnt.SockInitShm")
	api.RegisterMessage((*SockInitShmReply)(nil), "memclnt.SockInitShmReply")
	api.RegisterMessage((*SockclntCreate)(nil), "memclnt.SockclntCreate")
	api.RegisterMessage((*SockclntCreateReply)(nil), "memclnt.SockclntCreateReply")
	api.RegisterMessage((*SockclntDelete)(nil), "memclnt.SockclntDelete")
	api.RegisterMessage((*SockclntDeleteReply)(nil), "memclnt.SockclntDeleteReply")
	api.RegisterMessage((*TracePluginMsgIds)(nil), "memclnt.TracePluginMsgIds")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*APIVersions)(nil),
		(*APIVersionsReply)(nil),
		(*GetFirstMsgID)(nil),
		(*GetFirstMsgIDReply)(nil),
		(*MemclntCreate)(nil),
		(*MemclntCreateReply)(nil),
		(*MemclntDelete)(nil),
		(*MemclntDeleteReply)(nil),
		(*MemclntKeepalive)(nil),
		(*MemclntKeepaliveReply)(nil),
		(*MemclntReadTimeout)(nil),
		(*MemclntRxThreadSuspend)(nil),
		(*RPCCall)(nil),
		(*RPCCallReply)(nil),
		(*RxThreadExit)(nil),
		(*SockInitShm)(nil),
		(*SockInitShmReply)(nil),
		(*SockclntCreate)(nil),
		(*SockclntCreateReply)(nil),
		(*SockclntDelete)(nil),
		(*SockclntDeleteReply)(nil),
		(*TracePluginMsgIds)(nil),
	}
}

// RPCService represents RPC service API for memclnt module.
type RPCService interface {
	APIVersions(ctx context.Context, in *APIVersions) (*APIVersionsReply, error)
	GetFirstMsgID(ctx context.Context, in *GetFirstMsgID) (*GetFirstMsgIDReply, error)
	MemclntCreate(ctx context.Context, in *MemclntCreate) (*MemclntCreateReply, error)
	MemclntDelete(ctx context.Context, in *MemclntDelete) (*MemclntDeleteReply, error)
	MemclntKeepalive(ctx context.Context, in *MemclntKeepalive) (*MemclntKeepaliveReply, error)
	MemclntReadTimeout(ctx context.Context, in *MemclntReadTimeout) error
	MemclntRxThreadSuspend(ctx context.Context, in *MemclntRxThreadSuspend) error
	RPCCall(ctx context.Context, in *RPCCall) (*RPCCallReply, error)
	RxThreadExit(ctx context.Context, in *RxThreadExit) error
	SockInitShm(ctx context.Context, in *SockInitShm) (*SockInitShmReply, error)
	SockclntCreate(ctx context.Context, in *SockclntCreate) (*SockclntCreateReply, error)
	SockclntDelete(ctx context.Context, in *SockclntDelete) (*SockclntDeleteReply, error)
	TracePluginMsgIds(ctx context.Context, in *TracePluginMsgIds) error
}

type serviceClient struct {
	ch api.Channel
}

func NewServiceClient(ch api.Channel) RPCService {
	return &serviceClient{ch}
}

func (c *serviceClient) APIVersions(ctx context.Context, in *APIVersions) (*APIVersionsReply, error) {
	out := new(APIVersionsReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) GetFirstMsgID(ctx context.Context, in *GetFirstMsgID) (*GetFirstMsgIDReply, error) {
	out := new(GetFirstMsgIDReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) MemclntCreate(ctx context.Context, in *MemclntCreate) (*MemclntCreateReply, error) {
	out := new(MemclntCreateReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) MemclntDelete(ctx context.Context, in *MemclntDelete) (*MemclntDeleteReply, error) {
	out := new(MemclntDeleteReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) MemclntKeepalive(ctx context.Context, in *MemclntKeepalive) (*MemclntKeepaliveReply, error) {
	out := new(MemclntKeepaliveReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) MemclntReadTimeout(ctx context.Context, in *MemclntReadTimeout) error {
	c.ch.SendRequest(in)
	return nil
}

func (c *serviceClient) MemclntRxThreadSuspend(ctx context.Context, in *MemclntRxThreadSuspend) error {
	c.ch.SendRequest(in)
	return nil
}

func (c *serviceClient) RPCCall(ctx context.Context, in *RPCCall) (*RPCCallReply, error) {
	out := new(RPCCallReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) RxThreadExit(ctx context.Context, in *RxThreadExit) error {
	c.ch.SendRequest(in)
	return nil
}

func (c *serviceClient) SockInitShm(ctx context.Context, in *SockInitShm) (*SockInitShmReply, error) {
	out := new(SockInitShmReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SockclntCreate(ctx context.Context, in *SockclntCreate) (*SockclntCreateReply, error) {
	out := new(SockclntCreateReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SockclntDelete(ctx context.Context, in *SockclntDelete) (*SockclntDeleteReply, error) {
	out := new(SockclntDeleteReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) TracePluginMsgIds(ctx context.Context, in *TracePluginMsgIds) error {
	c.ch.SendRequest(in)
	return nil
}

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion1 // please upgrade the GoVPP api package

// Reference imports to suppress errors if they are not otherwise used.
var _ = api.RegisterMessage
var _ = bytes.NewBuffer
var _ = context.Background
var _ = io.Copy
var _ = strconv.Itoa
var _ = struc.Pack
