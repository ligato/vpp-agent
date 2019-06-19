// Code generated by GoVPP binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/punt.api.json

/*
Package punt is a generated from VPP binary API module 'punt'.

 The punt module consists of:
	  1 type
	 10 messages
	  5 services
*/
package punt

import api "git.fd.io/govpp.git/api"
import bytes "bytes"
import context "context"
import strconv "strconv"
import struc "github.com/lunixbochs/struc"

// Reference imports to suppress errors if they are not otherwise used.
var _ = api.RegisterMessage
var _ = bytes.NewBuffer
var _ = context.Background
var _ = strconv.Itoa
var _ = struc.Pack

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion1 // please upgrade the GoVPP api package

const (
	// ModuleName is the name of this module.
	ModuleName = "punt"
	// APIVersion is the API version of this module.
	APIVersion = "2.0.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0x1c487f2b
)

/* Types */

// Punt represents VPP binary API type 'punt':
type Punt struct {
	IPv        uint8
	L4Protocol uint8
	L4Port     uint16
}

func (*Punt) GetTypeName() string {
	return "punt"
}
func (*Punt) GetCrcString() string {
	return "fe4f98ac"
}

/* Messages */

// PuntDetails represents VPP binary API message 'punt_details':
type PuntDetails struct {
	Punt Punt
}

func (*PuntDetails) GetMessageName() string {
	return "punt_details"
}
func (*PuntDetails) GetCrcString() string {
	return "e905318e"
}
func (*PuntDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// PuntDump represents VPP binary API message 'punt_dump':
type PuntDump struct {
	IsIPv6 uint8
}

func (*PuntDump) GetMessageName() string {
	return "punt_dump"
}
func (*PuntDump) GetCrcString() string {
	return "de883da4"
}
func (*PuntDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// PuntSocketDeregister represents VPP binary API message 'punt_socket_deregister':
type PuntSocketDeregister struct {
	Punt Punt
}

func (*PuntSocketDeregister) GetMessageName() string {
	return "punt_socket_deregister"
}
func (*PuntSocketDeregister) GetCrcString() string {
	return "0603ba46"
}
func (*PuntSocketDeregister) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// PuntSocketDeregisterReply represents VPP binary API message 'punt_socket_deregister_reply':
type PuntSocketDeregisterReply struct {
	Retval int32
}

func (*PuntSocketDeregisterReply) GetMessageName() string {
	return "punt_socket_deregister_reply"
}
func (*PuntSocketDeregisterReply) GetCrcString() string {
	return "e8d4e804"
}
func (*PuntSocketDeregisterReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// PuntSocketDetails represents VPP binary API message 'punt_socket_details':
type PuntSocketDetails struct {
	Punt     Punt
	Pathname []byte `struc:"[108]byte"`
}

func (*PuntSocketDetails) GetMessageName() string {
	return "punt_socket_details"
}
func (*PuntSocketDetails) GetCrcString() string {
	return "8911c6c5"
}
func (*PuntSocketDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// PuntSocketDump represents VPP binary API message 'punt_socket_dump':
type PuntSocketDump struct {
	IsIPv6 uint8
}

func (*PuntSocketDump) GetMessageName() string {
	return "punt_socket_dump"
}
func (*PuntSocketDump) GetCrcString() string {
	return "de883da4"
}
func (*PuntSocketDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// PuntSocketRegister represents VPP binary API message 'punt_socket_register':
type PuntSocketRegister struct {
	HeaderVersion uint32
	Punt          Punt
	Pathname      []byte `struc:"[108]byte"`
}

func (*PuntSocketRegister) GetMessageName() string {
	return "punt_socket_register"
}
func (*PuntSocketRegister) GetCrcString() string {
	return "9f3e2877"
}
func (*PuntSocketRegister) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// PuntSocketRegisterReply represents VPP binary API message 'punt_socket_register_reply':
type PuntSocketRegisterReply struct {
	Retval   int32
	Pathname []byte `struc:"[64]byte"`
}

func (*PuntSocketRegisterReply) GetMessageName() string {
	return "punt_socket_register_reply"
}
func (*PuntSocketRegisterReply) GetCrcString() string {
	return "42dc0ee6"
}
func (*PuntSocketRegisterReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SetPunt represents VPP binary API message 'set_punt':
type SetPunt struct {
	IsAdd uint8
	Punt  Punt
}

func (*SetPunt) GetMessageName() string {
	return "set_punt"
}
func (*SetPunt) GetCrcString() string {
	return "332d88dc"
}
func (*SetPunt) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SetPuntReply represents VPP binary API message 'set_punt_reply':
type SetPuntReply struct {
	Retval int32
}

func (*SetPuntReply) GetMessageName() string {
	return "set_punt_reply"
}
func (*SetPuntReply) GetCrcString() string {
	return "e8d4e804"
}
func (*SetPuntReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func init() {
	api.RegisterMessage((*PuntDetails)(nil), "punt.PuntDetails")
	api.RegisterMessage((*PuntDump)(nil), "punt.PuntDump")
	api.RegisterMessage((*PuntSocketDeregister)(nil), "punt.PuntSocketDeregister")
	api.RegisterMessage((*PuntSocketDeregisterReply)(nil), "punt.PuntSocketDeregisterReply")
	api.RegisterMessage((*PuntSocketDetails)(nil), "punt.PuntSocketDetails")
	api.RegisterMessage((*PuntSocketDump)(nil), "punt.PuntSocketDump")
	api.RegisterMessage((*PuntSocketRegister)(nil), "punt.PuntSocketRegister")
	api.RegisterMessage((*PuntSocketRegisterReply)(nil), "punt.PuntSocketRegisterReply")
	api.RegisterMessage((*SetPunt)(nil), "punt.SetPunt")
	api.RegisterMessage((*SetPuntReply)(nil), "punt.SetPuntReply")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*PuntDetails)(nil),
		(*PuntDump)(nil),
		(*PuntSocketDeregister)(nil),
		(*PuntSocketDeregisterReply)(nil),
		(*PuntSocketDetails)(nil),
		(*PuntSocketDump)(nil),
		(*PuntSocketRegister)(nil),
		(*PuntSocketRegisterReply)(nil),
		(*SetPunt)(nil),
		(*SetPuntReply)(nil),
	}
}
