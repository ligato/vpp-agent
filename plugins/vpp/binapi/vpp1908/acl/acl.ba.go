// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/plugins/acl.api.json

/*
Package acl is a generated VPP binary API for 'acl' module.

It consists of:
	  2 types
	 38 messages
	 19 services
*/
package acl

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
	ModuleName = "acl"
	// APIVersion is the API version of this module.
	APIVersion = "1.0.1"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0xedb7b898
)

// ACLRule represents VPP binary API type 'acl_rule'.
type ACLRule struct {
	IsPermit               uint8
	IsIPv6                 uint8
	SrcIPAddr              []byte `struc:"[16]byte"`
	SrcIPPrefixLen         uint8
	DstIPAddr              []byte `struc:"[16]byte"`
	DstIPPrefixLen         uint8
	Proto                  uint8
	SrcportOrIcmptypeFirst uint16
	SrcportOrIcmptypeLast  uint16
	DstportOrIcmpcodeFirst uint16
	DstportOrIcmpcodeLast  uint16
	TCPFlagsMask           uint8
	TCPFlagsValue          uint8
}

func (*ACLRule) GetTypeName() string { return "acl_rule" }

// MacipACLRule represents VPP binary API type 'macip_acl_rule'.
type MacipACLRule struct {
	IsPermit       uint8
	IsIPv6         uint8
	SrcMac         []byte `struc:"[6]byte"`
	SrcMacMask     []byte `struc:"[6]byte"`
	SrcIPAddr      []byte `struc:"[16]byte"`
	SrcIPPrefixLen uint8
}

func (*MacipACLRule) GetTypeName() string { return "macip_acl_rule" }

// ACLAddReplace represents VPP binary API message 'acl_add_replace'.
type ACLAddReplace struct {
	ACLIndex uint32
	Tag      []byte `struc:"[64]byte"`
	Count    uint32 `struc:"sizeof=R"`
	R        []ACLRule
}

func (m *ACLAddReplace) Reset()                        { *m = ACLAddReplace{} }
func (*ACLAddReplace) GetMessageName() string          { return "acl_add_replace" }
func (*ACLAddReplace) GetCrcString() string            { return "13bc8539" }
func (*ACLAddReplace) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLAddReplaceReply represents VPP binary API message 'acl_add_replace_reply'.
type ACLAddReplaceReply struct {
	ACLIndex uint32
	Retval   int32
}

func (m *ACLAddReplaceReply) Reset()                        { *m = ACLAddReplaceReply{} }
func (*ACLAddReplaceReply) GetMessageName() string          { return "acl_add_replace_reply" }
func (*ACLAddReplaceReply) GetCrcString() string            { return "ac407b0c" }
func (*ACLAddReplaceReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLDel represents VPP binary API message 'acl_del'.
type ACLDel struct {
	ACLIndex uint32
}

func (m *ACLDel) Reset()                        { *m = ACLDel{} }
func (*ACLDel) GetMessageName() string          { return "acl_del" }
func (*ACLDel) GetCrcString() string            { return "ef34fea4" }
func (*ACLDel) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLDelReply represents VPP binary API message 'acl_del_reply'.
type ACLDelReply struct {
	Retval int32
}

func (m *ACLDelReply) Reset()                        { *m = ACLDelReply{} }
func (*ACLDelReply) GetMessageName() string          { return "acl_del_reply" }
func (*ACLDelReply) GetCrcString() string            { return "e8d4e804" }
func (*ACLDelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLDetails represents VPP binary API message 'acl_details'.
type ACLDetails struct {
	ACLIndex uint32
	Tag      []byte `struc:"[64]byte"`
	Count    uint32 `struc:"sizeof=R"`
	R        []ACLRule
}

func (m *ACLDetails) Reset()                        { *m = ACLDetails{} }
func (*ACLDetails) GetMessageName() string          { return "acl_details" }
func (*ACLDetails) GetCrcString() string            { return "f89d7a88" }
func (*ACLDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLDump represents VPP binary API message 'acl_dump'.
type ACLDump struct {
	ACLIndex uint32
}

func (m *ACLDump) Reset()                        { *m = ACLDump{} }
func (*ACLDump) GetMessageName() string          { return "acl_dump" }
func (*ACLDump) GetCrcString() string            { return "ef34fea4" }
func (*ACLDump) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLInterfaceAddDel represents VPP binary API message 'acl_interface_add_del'.
type ACLInterfaceAddDel struct {
	IsAdd     uint8
	IsInput   uint8
	SwIfIndex uint32
	ACLIndex  uint32
}

func (m *ACLInterfaceAddDel) Reset()                        { *m = ACLInterfaceAddDel{} }
func (*ACLInterfaceAddDel) GetMessageName() string          { return "acl_interface_add_del" }
func (*ACLInterfaceAddDel) GetCrcString() string            { return "0b2aedd1" }
func (*ACLInterfaceAddDel) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLInterfaceAddDelReply represents VPP binary API message 'acl_interface_add_del_reply'.
type ACLInterfaceAddDelReply struct {
	Retval int32
}

func (m *ACLInterfaceAddDelReply) Reset()                        { *m = ACLInterfaceAddDelReply{} }
func (*ACLInterfaceAddDelReply) GetMessageName() string          { return "acl_interface_add_del_reply" }
func (*ACLInterfaceAddDelReply) GetCrcString() string            { return "e8d4e804" }
func (*ACLInterfaceAddDelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLInterfaceEtypeWhitelistDetails represents VPP binary API message 'acl_interface_etype_whitelist_details'.
type ACLInterfaceEtypeWhitelistDetails struct {
	SwIfIndex uint32
	Count     uint8 `struc:"sizeof=Whitelist"`
	NInput    uint8
	Whitelist []uint16
}

func (m *ACLInterfaceEtypeWhitelistDetails) Reset() { *m = ACLInterfaceEtypeWhitelistDetails{} }
func (*ACLInterfaceEtypeWhitelistDetails) GetMessageName() string {
	return "acl_interface_etype_whitelist_details"
}
func (*ACLInterfaceEtypeWhitelistDetails) GetCrcString() string            { return "6a5d4e81" }
func (*ACLInterfaceEtypeWhitelistDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLInterfaceEtypeWhitelistDump represents VPP binary API message 'acl_interface_etype_whitelist_dump'.
type ACLInterfaceEtypeWhitelistDump struct {
	SwIfIndex uint32
}

func (m *ACLInterfaceEtypeWhitelistDump) Reset() { *m = ACLInterfaceEtypeWhitelistDump{} }
func (*ACLInterfaceEtypeWhitelistDump) GetMessageName() string {
	return "acl_interface_etype_whitelist_dump"
}
func (*ACLInterfaceEtypeWhitelistDump) GetCrcString() string            { return "529cb13f" }
func (*ACLInterfaceEtypeWhitelistDump) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLInterfaceListDetails represents VPP binary API message 'acl_interface_list_details'.
type ACLInterfaceListDetails struct {
	SwIfIndex uint32
	Count     uint8 `struc:"sizeof=Acls"`
	NInput    uint8
	Acls      []uint32
}

func (m *ACLInterfaceListDetails) Reset()                        { *m = ACLInterfaceListDetails{} }
func (*ACLInterfaceListDetails) GetMessageName() string          { return "acl_interface_list_details" }
func (*ACLInterfaceListDetails) GetCrcString() string            { return "d5e80809" }
func (*ACLInterfaceListDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLInterfaceListDump represents VPP binary API message 'acl_interface_list_dump'.
type ACLInterfaceListDump struct {
	SwIfIndex uint32
}

func (m *ACLInterfaceListDump) Reset()                        { *m = ACLInterfaceListDump{} }
func (*ACLInterfaceListDump) GetMessageName() string          { return "acl_interface_list_dump" }
func (*ACLInterfaceListDump) GetCrcString() string            { return "529cb13f" }
func (*ACLInterfaceListDump) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLInterfaceSetACLList represents VPP binary API message 'acl_interface_set_acl_list'.
type ACLInterfaceSetACLList struct {
	SwIfIndex uint32
	Count     uint8 `struc:"sizeof=Acls"`
	NInput    uint8
	Acls      []uint32
}

func (m *ACLInterfaceSetACLList) Reset()                        { *m = ACLInterfaceSetACLList{} }
func (*ACLInterfaceSetACLList) GetMessageName() string          { return "acl_interface_set_acl_list" }
func (*ACLInterfaceSetACLList) GetCrcString() string            { return "8baece38" }
func (*ACLInterfaceSetACLList) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLInterfaceSetACLListReply represents VPP binary API message 'acl_interface_set_acl_list_reply'.
type ACLInterfaceSetACLListReply struct {
	Retval int32
}

func (m *ACLInterfaceSetACLListReply) Reset() { *m = ACLInterfaceSetACLListReply{} }
func (*ACLInterfaceSetACLListReply) GetMessageName() string {
	return "acl_interface_set_acl_list_reply"
}
func (*ACLInterfaceSetACLListReply) GetCrcString() string            { return "e8d4e804" }
func (*ACLInterfaceSetACLListReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLInterfaceSetEtypeWhitelist represents VPP binary API message 'acl_interface_set_etype_whitelist'.
type ACLInterfaceSetEtypeWhitelist struct {
	SwIfIndex uint32
	Count     uint8 `struc:"sizeof=Whitelist"`
	NInput    uint8
	Whitelist []uint16
}

func (m *ACLInterfaceSetEtypeWhitelist) Reset() { *m = ACLInterfaceSetEtypeWhitelist{} }
func (*ACLInterfaceSetEtypeWhitelist) GetMessageName() string {
	return "acl_interface_set_etype_whitelist"
}
func (*ACLInterfaceSetEtypeWhitelist) GetCrcString() string            { return "f515efc5" }
func (*ACLInterfaceSetEtypeWhitelist) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLInterfaceSetEtypeWhitelistReply represents VPP binary API message 'acl_interface_set_etype_whitelist_reply'.
type ACLInterfaceSetEtypeWhitelistReply struct {
	Retval int32
}

func (m *ACLInterfaceSetEtypeWhitelistReply) Reset() { *m = ACLInterfaceSetEtypeWhitelistReply{} }
func (*ACLInterfaceSetEtypeWhitelistReply) GetMessageName() string {
	return "acl_interface_set_etype_whitelist_reply"
}
func (*ACLInterfaceSetEtypeWhitelistReply) GetCrcString() string            { return "e8d4e804" }
func (*ACLInterfaceSetEtypeWhitelistReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLPluginControlPing represents VPP binary API message 'acl_plugin_control_ping'.
type ACLPluginControlPing struct{}

func (m *ACLPluginControlPing) Reset()                        { *m = ACLPluginControlPing{} }
func (*ACLPluginControlPing) GetMessageName() string          { return "acl_plugin_control_ping" }
func (*ACLPluginControlPing) GetCrcString() string            { return "51077d14" }
func (*ACLPluginControlPing) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLPluginControlPingReply represents VPP binary API message 'acl_plugin_control_ping_reply'.
type ACLPluginControlPingReply struct {
	Retval      int32
	ClientIndex uint32
	VpePID      uint32
}

func (m *ACLPluginControlPingReply) Reset()                        { *m = ACLPluginControlPingReply{} }
func (*ACLPluginControlPingReply) GetMessageName() string          { return "acl_plugin_control_ping_reply" }
func (*ACLPluginControlPingReply) GetCrcString() string            { return "f6b0b8ca" }
func (*ACLPluginControlPingReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLPluginGetConnTableMaxEntries represents VPP binary API message 'acl_plugin_get_conn_table_max_entries'.
type ACLPluginGetConnTableMaxEntries struct{}

func (m *ACLPluginGetConnTableMaxEntries) Reset() { *m = ACLPluginGetConnTableMaxEntries{} }
func (*ACLPluginGetConnTableMaxEntries) GetMessageName() string {
	return "acl_plugin_get_conn_table_max_entries"
}
func (*ACLPluginGetConnTableMaxEntries) GetCrcString() string            { return "51077d14" }
func (*ACLPluginGetConnTableMaxEntries) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLPluginGetConnTableMaxEntriesReply represents VPP binary API message 'acl_plugin_get_conn_table_max_entries_reply'.
type ACLPluginGetConnTableMaxEntriesReply struct {
	ConnTableMaxEntries uint64
}

func (m *ACLPluginGetConnTableMaxEntriesReply) Reset() { *m = ACLPluginGetConnTableMaxEntriesReply{} }
func (*ACLPluginGetConnTableMaxEntriesReply) GetMessageName() string {
	return "acl_plugin_get_conn_table_max_entries_reply"
}
func (*ACLPluginGetConnTableMaxEntriesReply) GetCrcString() string { return "7a096d3d" }
func (*ACLPluginGetConnTableMaxEntriesReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// ACLPluginGetVersion represents VPP binary API message 'acl_plugin_get_version'.
type ACLPluginGetVersion struct{}

func (m *ACLPluginGetVersion) Reset()                        { *m = ACLPluginGetVersion{} }
func (*ACLPluginGetVersion) GetMessageName() string          { return "acl_plugin_get_version" }
func (*ACLPluginGetVersion) GetCrcString() string            { return "51077d14" }
func (*ACLPluginGetVersion) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLPluginGetVersionReply represents VPP binary API message 'acl_plugin_get_version_reply'.
type ACLPluginGetVersionReply struct {
	Major uint32
	Minor uint32
}

func (m *ACLPluginGetVersionReply) Reset()                        { *m = ACLPluginGetVersionReply{} }
func (*ACLPluginGetVersionReply) GetMessageName() string          { return "acl_plugin_get_version_reply" }
func (*ACLPluginGetVersionReply) GetCrcString() string            { return "9b32cf86" }
func (*ACLPluginGetVersionReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// ACLStatsIntfCountersEnable represents VPP binary API message 'acl_stats_intf_counters_enable'.
type ACLStatsIntfCountersEnable struct {
	Enable bool
}

func (m *ACLStatsIntfCountersEnable) Reset()                        { *m = ACLStatsIntfCountersEnable{} }
func (*ACLStatsIntfCountersEnable) GetMessageName() string          { return "acl_stats_intf_counters_enable" }
func (*ACLStatsIntfCountersEnable) GetCrcString() string            { return "b3e225d2" }
func (*ACLStatsIntfCountersEnable) GetMessageType() api.MessageType { return api.RequestMessage }

// ACLStatsIntfCountersEnableReply represents VPP binary API message 'acl_stats_intf_counters_enable_reply'.
type ACLStatsIntfCountersEnableReply struct {
	Retval int32
}

func (m *ACLStatsIntfCountersEnableReply) Reset() { *m = ACLStatsIntfCountersEnableReply{} }
func (*ACLStatsIntfCountersEnableReply) GetMessageName() string {
	return "acl_stats_intf_counters_enable_reply"
}
func (*ACLStatsIntfCountersEnableReply) GetCrcString() string            { return "e8d4e804" }
func (*ACLStatsIntfCountersEnableReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// MacipACLAdd represents VPP binary API message 'macip_acl_add'.
type MacipACLAdd struct {
	Tag   []byte `struc:"[64]byte"`
	Count uint32 `struc:"sizeof=R"`
	R     []MacipACLRule
}

func (m *MacipACLAdd) Reset()                        { *m = MacipACLAdd{} }
func (*MacipACLAdd) GetMessageName() string          { return "macip_acl_add" }
func (*MacipACLAdd) GetCrcString() string            { return "0c680ca5" }
func (*MacipACLAdd) GetMessageType() api.MessageType { return api.RequestMessage }

// MacipACLAddReplace represents VPP binary API message 'macip_acl_add_replace'.
type MacipACLAddReplace struct {
	ACLIndex uint32
	Tag      []byte `struc:"[64]byte"`
	Count    uint32 `struc:"sizeof=R"`
	R        []MacipACLRule
}

func (m *MacipACLAddReplace) Reset()                        { *m = MacipACLAddReplace{} }
func (*MacipACLAddReplace) GetMessageName() string          { return "macip_acl_add_replace" }
func (*MacipACLAddReplace) GetCrcString() string            { return "d3d313e7" }
func (*MacipACLAddReplace) GetMessageType() api.MessageType { return api.RequestMessage }

// MacipACLAddReplaceReply represents VPP binary API message 'macip_acl_add_replace_reply'.
type MacipACLAddReplaceReply struct {
	ACLIndex uint32
	Retval   int32
}

func (m *MacipACLAddReplaceReply) Reset()                        { *m = MacipACLAddReplaceReply{} }
func (*MacipACLAddReplaceReply) GetMessageName() string          { return "macip_acl_add_replace_reply" }
func (*MacipACLAddReplaceReply) GetCrcString() string            { return "ac407b0c" }
func (*MacipACLAddReplaceReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// MacipACLAddReply represents VPP binary API message 'macip_acl_add_reply'.
type MacipACLAddReply struct {
	ACLIndex uint32
	Retval   int32
}

func (m *MacipACLAddReply) Reset()                        { *m = MacipACLAddReply{} }
func (*MacipACLAddReply) GetMessageName() string          { return "macip_acl_add_reply" }
func (*MacipACLAddReply) GetCrcString() string            { return "ac407b0c" }
func (*MacipACLAddReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// MacipACLDel represents VPP binary API message 'macip_acl_del'.
type MacipACLDel struct {
	ACLIndex uint32
}

func (m *MacipACLDel) Reset()                        { *m = MacipACLDel{} }
func (*MacipACLDel) GetMessageName() string          { return "macip_acl_del" }
func (*MacipACLDel) GetCrcString() string            { return "ef34fea4" }
func (*MacipACLDel) GetMessageType() api.MessageType { return api.RequestMessage }

// MacipACLDelReply represents VPP binary API message 'macip_acl_del_reply'.
type MacipACLDelReply struct {
	Retval int32
}

func (m *MacipACLDelReply) Reset()                        { *m = MacipACLDelReply{} }
func (*MacipACLDelReply) GetMessageName() string          { return "macip_acl_del_reply" }
func (*MacipACLDelReply) GetCrcString() string            { return "e8d4e804" }
func (*MacipACLDelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// MacipACLDetails represents VPP binary API message 'macip_acl_details'.
type MacipACLDetails struct {
	ACLIndex uint32
	Tag      []byte `struc:"[64]byte"`
	Count    uint32 `struc:"sizeof=R"`
	R        []MacipACLRule
}

func (m *MacipACLDetails) Reset()                        { *m = MacipACLDetails{} }
func (*MacipACLDetails) GetMessageName() string          { return "macip_acl_details" }
func (*MacipACLDetails) GetCrcString() string            { return "e164e69a" }
func (*MacipACLDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// MacipACLDump represents VPP binary API message 'macip_acl_dump'.
type MacipACLDump struct {
	ACLIndex uint32
}

func (m *MacipACLDump) Reset()                        { *m = MacipACLDump{} }
func (*MacipACLDump) GetMessageName() string          { return "macip_acl_dump" }
func (*MacipACLDump) GetCrcString() string            { return "ef34fea4" }
func (*MacipACLDump) GetMessageType() api.MessageType { return api.RequestMessage }

// MacipACLInterfaceAddDel represents VPP binary API message 'macip_acl_interface_add_del'.
type MacipACLInterfaceAddDel struct {
	IsAdd     uint8
	SwIfIndex uint32
	ACLIndex  uint32
}

func (m *MacipACLInterfaceAddDel) Reset()                        { *m = MacipACLInterfaceAddDel{} }
func (*MacipACLInterfaceAddDel) GetMessageName() string          { return "macip_acl_interface_add_del" }
func (*MacipACLInterfaceAddDel) GetCrcString() string            { return "6a6be97c" }
func (*MacipACLInterfaceAddDel) GetMessageType() api.MessageType { return api.RequestMessage }

// MacipACLInterfaceAddDelReply represents VPP binary API message 'macip_acl_interface_add_del_reply'.
type MacipACLInterfaceAddDelReply struct {
	Retval int32
}

func (m *MacipACLInterfaceAddDelReply) Reset() { *m = MacipACLInterfaceAddDelReply{} }
func (*MacipACLInterfaceAddDelReply) GetMessageName() string {
	return "macip_acl_interface_add_del_reply"
}
func (*MacipACLInterfaceAddDelReply) GetCrcString() string            { return "e8d4e804" }
func (*MacipACLInterfaceAddDelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// MacipACLInterfaceGet represents VPP binary API message 'macip_acl_interface_get'.
type MacipACLInterfaceGet struct{}

func (m *MacipACLInterfaceGet) Reset()                        { *m = MacipACLInterfaceGet{} }
func (*MacipACLInterfaceGet) GetMessageName() string          { return "macip_acl_interface_get" }
func (*MacipACLInterfaceGet) GetCrcString() string            { return "51077d14" }
func (*MacipACLInterfaceGet) GetMessageType() api.MessageType { return api.RequestMessage }

// MacipACLInterfaceGetReply represents VPP binary API message 'macip_acl_interface_get_reply'.
type MacipACLInterfaceGetReply struct {
	Count uint32 `struc:"sizeof=Acls"`
	Acls  []uint32
}

func (m *MacipACLInterfaceGetReply) Reset()                        { *m = MacipACLInterfaceGetReply{} }
func (*MacipACLInterfaceGetReply) GetMessageName() string          { return "macip_acl_interface_get_reply" }
func (*MacipACLInterfaceGetReply) GetCrcString() string            { return "accf9b05" }
func (*MacipACLInterfaceGetReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// MacipACLInterfaceListDetails represents VPP binary API message 'macip_acl_interface_list_details'.
type MacipACLInterfaceListDetails struct {
	SwIfIndex uint32
	Count     uint8 `struc:"sizeof=Acls"`
	Acls      []uint32
}

func (m *MacipACLInterfaceListDetails) Reset() { *m = MacipACLInterfaceListDetails{} }
func (*MacipACLInterfaceListDetails) GetMessageName() string {
	return "macip_acl_interface_list_details"
}
func (*MacipACLInterfaceListDetails) GetCrcString() string            { return "29783fa0" }
func (*MacipACLInterfaceListDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// MacipACLInterfaceListDump represents VPP binary API message 'macip_acl_interface_list_dump'.
type MacipACLInterfaceListDump struct {
	SwIfIndex uint32
}

func (m *MacipACLInterfaceListDump) Reset()                        { *m = MacipACLInterfaceListDump{} }
func (*MacipACLInterfaceListDump) GetMessageName() string          { return "macip_acl_interface_list_dump" }
func (*MacipACLInterfaceListDump) GetCrcString() string            { return "529cb13f" }
func (*MacipACLInterfaceListDump) GetMessageType() api.MessageType { return api.RequestMessage }

func init() {
	api.RegisterMessage((*ACLAddReplace)(nil), "acl.ACLAddReplace")
	api.RegisterMessage((*ACLAddReplaceReply)(nil), "acl.ACLAddReplaceReply")
	api.RegisterMessage((*ACLDel)(nil), "acl.ACLDel")
	api.RegisterMessage((*ACLDelReply)(nil), "acl.ACLDelReply")
	api.RegisterMessage((*ACLDetails)(nil), "acl.ACLDetails")
	api.RegisterMessage((*ACLDump)(nil), "acl.ACLDump")
	api.RegisterMessage((*ACLInterfaceAddDel)(nil), "acl.ACLInterfaceAddDel")
	api.RegisterMessage((*ACLInterfaceAddDelReply)(nil), "acl.ACLInterfaceAddDelReply")
	api.RegisterMessage((*ACLInterfaceEtypeWhitelistDetails)(nil), "acl.ACLInterfaceEtypeWhitelistDetails")
	api.RegisterMessage((*ACLInterfaceEtypeWhitelistDump)(nil), "acl.ACLInterfaceEtypeWhitelistDump")
	api.RegisterMessage((*ACLInterfaceListDetails)(nil), "acl.ACLInterfaceListDetails")
	api.RegisterMessage((*ACLInterfaceListDump)(nil), "acl.ACLInterfaceListDump")
	api.RegisterMessage((*ACLInterfaceSetACLList)(nil), "acl.ACLInterfaceSetACLList")
	api.RegisterMessage((*ACLInterfaceSetACLListReply)(nil), "acl.ACLInterfaceSetACLListReply")
	api.RegisterMessage((*ACLInterfaceSetEtypeWhitelist)(nil), "acl.ACLInterfaceSetEtypeWhitelist")
	api.RegisterMessage((*ACLInterfaceSetEtypeWhitelistReply)(nil), "acl.ACLInterfaceSetEtypeWhitelistReply")
	api.RegisterMessage((*ACLPluginControlPing)(nil), "acl.ACLPluginControlPing")
	api.RegisterMessage((*ACLPluginControlPingReply)(nil), "acl.ACLPluginControlPingReply")
	api.RegisterMessage((*ACLPluginGetConnTableMaxEntries)(nil), "acl.ACLPluginGetConnTableMaxEntries")
	api.RegisterMessage((*ACLPluginGetConnTableMaxEntriesReply)(nil), "acl.ACLPluginGetConnTableMaxEntriesReply")
	api.RegisterMessage((*ACLPluginGetVersion)(nil), "acl.ACLPluginGetVersion")
	api.RegisterMessage((*ACLPluginGetVersionReply)(nil), "acl.ACLPluginGetVersionReply")
	api.RegisterMessage((*ACLStatsIntfCountersEnable)(nil), "acl.ACLStatsIntfCountersEnable")
	api.RegisterMessage((*ACLStatsIntfCountersEnableReply)(nil), "acl.ACLStatsIntfCountersEnableReply")
	api.RegisterMessage((*MacipACLAdd)(nil), "acl.MacipACLAdd")
	api.RegisterMessage((*MacipACLAddReplace)(nil), "acl.MacipACLAddReplace")
	api.RegisterMessage((*MacipACLAddReplaceReply)(nil), "acl.MacipACLAddReplaceReply")
	api.RegisterMessage((*MacipACLAddReply)(nil), "acl.MacipACLAddReply")
	api.RegisterMessage((*MacipACLDel)(nil), "acl.MacipACLDel")
	api.RegisterMessage((*MacipACLDelReply)(nil), "acl.MacipACLDelReply")
	api.RegisterMessage((*MacipACLDetails)(nil), "acl.MacipACLDetails")
	api.RegisterMessage((*MacipACLDump)(nil), "acl.MacipACLDump")
	api.RegisterMessage((*MacipACLInterfaceAddDel)(nil), "acl.MacipACLInterfaceAddDel")
	api.RegisterMessage((*MacipACLInterfaceAddDelReply)(nil), "acl.MacipACLInterfaceAddDelReply")
	api.RegisterMessage((*MacipACLInterfaceGet)(nil), "acl.MacipACLInterfaceGet")
	api.RegisterMessage((*MacipACLInterfaceGetReply)(nil), "acl.MacipACLInterfaceGetReply")
	api.RegisterMessage((*MacipACLInterfaceListDetails)(nil), "acl.MacipACLInterfaceListDetails")
	api.RegisterMessage((*MacipACLInterfaceListDump)(nil), "acl.MacipACLInterfaceListDump")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*ACLAddReplace)(nil),
		(*ACLAddReplaceReply)(nil),
		(*ACLDel)(nil),
		(*ACLDelReply)(nil),
		(*ACLDetails)(nil),
		(*ACLDump)(nil),
		(*ACLInterfaceAddDel)(nil),
		(*ACLInterfaceAddDelReply)(nil),
		(*ACLInterfaceEtypeWhitelistDetails)(nil),
		(*ACLInterfaceEtypeWhitelistDump)(nil),
		(*ACLInterfaceListDetails)(nil),
		(*ACLInterfaceListDump)(nil),
		(*ACLInterfaceSetACLList)(nil),
		(*ACLInterfaceSetACLListReply)(nil),
		(*ACLInterfaceSetEtypeWhitelist)(nil),
		(*ACLInterfaceSetEtypeWhitelistReply)(nil),
		(*ACLPluginControlPing)(nil),
		(*ACLPluginControlPingReply)(nil),
		(*ACLPluginGetConnTableMaxEntries)(nil),
		(*ACLPluginGetConnTableMaxEntriesReply)(nil),
		(*ACLPluginGetVersion)(nil),
		(*ACLPluginGetVersionReply)(nil),
		(*ACLStatsIntfCountersEnable)(nil),
		(*ACLStatsIntfCountersEnableReply)(nil),
		(*MacipACLAdd)(nil),
		(*MacipACLAddReplace)(nil),
		(*MacipACLAddReplaceReply)(nil),
		(*MacipACLAddReply)(nil),
		(*MacipACLDel)(nil),
		(*MacipACLDelReply)(nil),
		(*MacipACLDetails)(nil),
		(*MacipACLDump)(nil),
		(*MacipACLInterfaceAddDel)(nil),
		(*MacipACLInterfaceAddDelReply)(nil),
		(*MacipACLInterfaceGet)(nil),
		(*MacipACLInterfaceGetReply)(nil),
		(*MacipACLInterfaceListDetails)(nil),
		(*MacipACLInterfaceListDump)(nil),
	}
}

// RPCService represents RPC service API for acl module.
type RPCService interface {
	DumpACL(ctx context.Context, in *ACLDump) (RPCService_DumpACLClient, error)
	DumpACLInterfaceEtypeWhitelist(ctx context.Context, in *ACLInterfaceEtypeWhitelistDump) (RPCService_DumpACLInterfaceEtypeWhitelistClient, error)
	DumpACLInterfaceList(ctx context.Context, in *ACLInterfaceListDump) (RPCService_DumpACLInterfaceListClient, error)
	DumpMacipACL(ctx context.Context, in *MacipACLDump) (RPCService_DumpMacipACLClient, error)
	DumpMacipACLInterfaceList(ctx context.Context, in *MacipACLInterfaceListDump) (RPCService_DumpMacipACLInterfaceListClient, error)
	ACLAddReplace(ctx context.Context, in *ACLAddReplace) (*ACLAddReplaceReply, error)
	ACLDel(ctx context.Context, in *ACLDel) (*ACLDelReply, error)
	ACLInterfaceAddDel(ctx context.Context, in *ACLInterfaceAddDel) (*ACLInterfaceAddDelReply, error)
	ACLInterfaceSetACLList(ctx context.Context, in *ACLInterfaceSetACLList) (*ACLInterfaceSetACLListReply, error)
	ACLInterfaceSetEtypeWhitelist(ctx context.Context, in *ACLInterfaceSetEtypeWhitelist) (*ACLInterfaceSetEtypeWhitelistReply, error)
	ACLPluginControlPing(ctx context.Context, in *ACLPluginControlPing) (*ACLPluginControlPingReply, error)
	ACLPluginGetConnTableMaxEntries(ctx context.Context, in *ACLPluginGetConnTableMaxEntries) (*ACLPluginGetConnTableMaxEntriesReply, error)
	ACLPluginGetVersion(ctx context.Context, in *ACLPluginGetVersion) (*ACLPluginGetVersionReply, error)
	ACLStatsIntfCountersEnable(ctx context.Context, in *ACLStatsIntfCountersEnable) (*ACLStatsIntfCountersEnableReply, error)
	MacipACLAdd(ctx context.Context, in *MacipACLAdd) (*MacipACLAddReply, error)
	MacipACLAddReplace(ctx context.Context, in *MacipACLAddReplace) (*MacipACLAddReplaceReply, error)
	MacipACLDel(ctx context.Context, in *MacipACLDel) (*MacipACLDelReply, error)
	MacipACLInterfaceAddDel(ctx context.Context, in *MacipACLInterfaceAddDel) (*MacipACLInterfaceAddDelReply, error)
	MacipACLInterfaceGet(ctx context.Context, in *MacipACLInterfaceGet) (*MacipACLInterfaceGetReply, error)
}

type serviceClient struct {
	ch api.Channel
}

func NewServiceClient(ch api.Channel) RPCService {
	return &serviceClient{ch}
}

func (c *serviceClient) DumpACL(ctx context.Context, in *ACLDump) (RPCService_DumpACLClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpACLClient{stream}
	return x, nil
}

type RPCService_DumpACLClient interface {
	Recv() (*ACLDetails, error)
}

type serviceClient_DumpACLClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpACLClient) Recv() (*ACLDetails, error) {
	m := new(ACLDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpACLInterfaceEtypeWhitelist(ctx context.Context, in *ACLInterfaceEtypeWhitelistDump) (RPCService_DumpACLInterfaceEtypeWhitelistClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpACLInterfaceEtypeWhitelistClient{stream}
	return x, nil
}

type RPCService_DumpACLInterfaceEtypeWhitelistClient interface {
	Recv() (*ACLInterfaceEtypeWhitelistDetails, error)
}

type serviceClient_DumpACLInterfaceEtypeWhitelistClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpACLInterfaceEtypeWhitelistClient) Recv() (*ACLInterfaceEtypeWhitelistDetails, error) {
	m := new(ACLInterfaceEtypeWhitelistDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpACLInterfaceList(ctx context.Context, in *ACLInterfaceListDump) (RPCService_DumpACLInterfaceListClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpACLInterfaceListClient{stream}
	return x, nil
}

type RPCService_DumpACLInterfaceListClient interface {
	Recv() (*ACLInterfaceListDetails, error)
}

type serviceClient_DumpACLInterfaceListClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpACLInterfaceListClient) Recv() (*ACLInterfaceListDetails, error) {
	m := new(ACLInterfaceListDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpMacipACL(ctx context.Context, in *MacipACLDump) (RPCService_DumpMacipACLClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpMacipACLClient{stream}
	return x, nil
}

type RPCService_DumpMacipACLClient interface {
	Recv() (*MacipACLDetails, error)
}

type serviceClient_DumpMacipACLClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpMacipACLClient) Recv() (*MacipACLDetails, error) {
	m := new(MacipACLDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpMacipACLInterfaceList(ctx context.Context, in *MacipACLInterfaceListDump) (RPCService_DumpMacipACLInterfaceListClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpMacipACLInterfaceListClient{stream}
	return x, nil
}

type RPCService_DumpMacipACLInterfaceListClient interface {
	Recv() (*MacipACLInterfaceListDetails, error)
}

type serviceClient_DumpMacipACLInterfaceListClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpMacipACLInterfaceListClient) Recv() (*MacipACLInterfaceListDetails, error) {
	m := new(MacipACLInterfaceListDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) ACLAddReplace(ctx context.Context, in *ACLAddReplace) (*ACLAddReplaceReply, error) {
	out := new(ACLAddReplaceReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ACLDel(ctx context.Context, in *ACLDel) (*ACLDelReply, error) {
	out := new(ACLDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ACLInterfaceAddDel(ctx context.Context, in *ACLInterfaceAddDel) (*ACLInterfaceAddDelReply, error) {
	out := new(ACLInterfaceAddDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ACLInterfaceSetACLList(ctx context.Context, in *ACLInterfaceSetACLList) (*ACLInterfaceSetACLListReply, error) {
	out := new(ACLInterfaceSetACLListReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ACLInterfaceSetEtypeWhitelist(ctx context.Context, in *ACLInterfaceSetEtypeWhitelist) (*ACLInterfaceSetEtypeWhitelistReply, error) {
	out := new(ACLInterfaceSetEtypeWhitelistReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ACLPluginControlPing(ctx context.Context, in *ACLPluginControlPing) (*ACLPluginControlPingReply, error) {
	out := new(ACLPluginControlPingReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ACLPluginGetConnTableMaxEntries(ctx context.Context, in *ACLPluginGetConnTableMaxEntries) (*ACLPluginGetConnTableMaxEntriesReply, error) {
	out := new(ACLPluginGetConnTableMaxEntriesReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ACLPluginGetVersion(ctx context.Context, in *ACLPluginGetVersion) (*ACLPluginGetVersionReply, error) {
	out := new(ACLPluginGetVersionReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ACLStatsIntfCountersEnable(ctx context.Context, in *ACLStatsIntfCountersEnable) (*ACLStatsIntfCountersEnableReply, error) {
	out := new(ACLStatsIntfCountersEnableReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) MacipACLAdd(ctx context.Context, in *MacipACLAdd) (*MacipACLAddReply, error) {
	out := new(MacipACLAddReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) MacipACLAddReplace(ctx context.Context, in *MacipACLAddReplace) (*MacipACLAddReplaceReply, error) {
	out := new(MacipACLAddReplaceReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) MacipACLDel(ctx context.Context, in *MacipACLDel) (*MacipACLDelReply, error) {
	out := new(MacipACLDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) MacipACLInterfaceAddDel(ctx context.Context, in *MacipACLInterfaceAddDel) (*MacipACLInterfaceAddDelReply, error) {
	out := new(MacipACLInterfaceAddDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) MacipACLInterfaceGet(ctx context.Context, in *MacipACLInterfaceGet) (*MacipACLInterfaceGetReply, error) {
	out := new(MacipACLInterfaceGetReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
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
