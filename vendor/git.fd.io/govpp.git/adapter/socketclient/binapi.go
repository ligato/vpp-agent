package socketclient

import (
	"git.fd.io/govpp.git/api"
)

// MessageTableEntry represents VPP binary API type 'message_table_entry'.
type MessageTableEntry struct {
	Index uint16
	Name  string `struc:"[64]byte"`
}

func (*MessageTableEntry) GetTypeName() string {
	return "message_table_entry"
}

// SockclntCreate represents VPP binary API message 'sockclnt_create'.
type SockclntCreate struct {
	Name string `struc:"[64]byte"`
}

func (*SockclntCreate) GetMessageName() string {
	return "sockclnt_create"
}
func (*SockclntCreate) GetCrcString() string {
	return "455fb9c4"
}
func (*SockclntCreate) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SockclntCreateReply represents VPP binary API message 'sockclnt_create_reply'.
type SockclntCreateReply struct {
	Response     int32
	Index        uint32
	Count        uint16 `struc:"sizeof=MessageTable"`
	MessageTable []MessageTableEntry
}

func (*SockclntCreateReply) GetMessageName() string {
	return "sockclnt_create_reply"
}
func (*SockclntCreateReply) GetCrcString() string {
	return "35166268"
}
func (*SockclntCreateReply) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SockclntDelete represents VPP binary API message 'sockclnt_delete'.
type SockclntDelete struct {
	Index uint32
}

func (*SockclntDelete) GetMessageName() string {
	return "sockclnt_delete"
}
func (*SockclntDelete) GetCrcString() string {
	return "8ac76db6"
}
func (*SockclntDelete) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SockclntDeleteReply represents VPP binary API message 'sockclnt_delete_reply'.
type SockclntDeleteReply struct {
	Response int32
}

func (*SockclntDeleteReply) GetMessageName() string {
	return "sockclnt_delete_reply"
}
func (*SockclntDeleteReply) GetCrcString() string {
	return "8f38b1ee"
}
func (*SockclntDeleteReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// ModuleVersion represents VPP binary API type 'module_version'.
type ModuleVersion struct {
	Major uint32
	Minor uint32
	Patch uint32
	Name  string `struc:"[64]byte"`
}

func (*ModuleVersion) GetTypeName() string {
	return "module_version"
}

// APIVersions represents VPP binary API message 'api_versions'.
type APIVersions struct{}

func (*APIVersions) GetMessageName() string {
	return "api_versions"
}
func (*APIVersions) GetCrcString() string {
	return "51077d14"
}
func (*APIVersions) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// APIVersionsReply represents VPP binary API message 'api_versions_reply'.
type APIVersionsReply struct {
	Retval      int32
	Count       uint32 `struc:"sizeof=APIVersions"`
	APIVersions []ModuleVersion
}

func (*APIVersionsReply) GetMessageName() string {
	return "api_versions_reply"
}
func (*APIVersionsReply) GetCrcString() string {
	return "5f0d99d6"
}
func (*APIVersionsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
