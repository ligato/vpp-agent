// Code generated by GoVPP binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/sr.api.json

/*
Package sr is a generated from VPP binary API module 'sr'.

 The sr module consists of:
	  3 types
	 18 messages
	  9 services
*/
package sr

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
	ModuleName = "sr"
	// APIVersion is the API version of this module.
	APIVersion = "1.2.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0xaa3993c3
)

/* Types */

// SrIP6Address represents VPP binary API type 'sr_ip6_address':
type SrIP6Address struct {
	Data []byte `struc:"[16]byte"`
}

func (*SrIP6Address) GetTypeName() string {
	return "sr_ip6_address"
}

// Srv6Sid represents VPP binary API type 'srv6_sid':
type Srv6Sid struct {
	Addr []byte `struc:"[16]byte"`
}

func (*Srv6Sid) GetTypeName() string {
	return "srv6_sid"
}

// Srv6SidList represents VPP binary API type 'srv6_sid_list':
type Srv6SidList struct {
	NumSids uint8 `struc:"sizeof=Sids"`
	Weight  uint32
	Sids    []Srv6Sid
}

func (*Srv6SidList) GetTypeName() string {
	return "srv6_sid_list"
}

/* Messages */

// SrLocalsidAddDel represents VPP binary API message 'sr_localsid_add_del':
type SrLocalsidAddDel struct {
	IsDel     uint8
	Localsid  Srv6Sid
	EndPsp    uint8
	Behavior  uint8
	SwIfIndex uint32
	VlanIndex uint32
	FibTable  uint32
	NhAddr6   []byte `struc:"[16]byte"`
	NhAddr4   []byte `struc:"[4]byte"`
}

func (*SrLocalsidAddDel) GetMessageName() string {
	return "sr_localsid_add_del"
}
func (*SrLocalsidAddDel) GetCrcString() string {
	return "b30489eb"
}
func (*SrLocalsidAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SrLocalsidAddDelReply represents VPP binary API message 'sr_localsid_add_del_reply':
type SrLocalsidAddDelReply struct {
	Retval int32
}

func (*SrLocalsidAddDelReply) GetMessageName() string {
	return "sr_localsid_add_del_reply"
}
func (*SrLocalsidAddDelReply) GetCrcString() string {
	return "e8d4e804"
}
func (*SrLocalsidAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SrLocalsidsDetails represents VPP binary API message 'sr_localsids_details':
type SrLocalsidsDetails struct {
	Addr                    Srv6Sid
	EndPsp                  uint8
	Behavior                uint16
	FibTable                uint32
	VlanIndex               uint32
	XconnectNhAddr6         []byte `struc:"[16]byte"`
	XconnectNhAddr4         []byte `struc:"[4]byte"`
	XconnectIfaceOrVrfTable uint32
}

func (*SrLocalsidsDetails) GetMessageName() string {
	return "sr_localsids_details"
}
func (*SrLocalsidsDetails) GetCrcString() string {
	return "0791babc"
}
func (*SrLocalsidsDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SrLocalsidsDump represents VPP binary API message 'sr_localsids_dump':
type SrLocalsidsDump struct{}

func (*SrLocalsidsDump) GetMessageName() string {
	return "sr_localsids_dump"
}
func (*SrLocalsidsDump) GetCrcString() string {
	return "51077d14"
}
func (*SrLocalsidsDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SrPoliciesDetails represents VPP binary API message 'sr_policies_details':
type SrPoliciesDetails struct {
	Bsid        Srv6Sid
	Type        uint8
	IsEncap     uint8
	FibTable    uint32
	NumSidLists uint8 `struc:"sizeof=SidLists"`
	SidLists    []Srv6SidList
}

func (*SrPoliciesDetails) GetMessageName() string {
	return "sr_policies_details"
}
func (*SrPoliciesDetails) GetCrcString() string {
	return "5087f460"
}
func (*SrPoliciesDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SrPoliciesDump represents VPP binary API message 'sr_policies_dump':
type SrPoliciesDump struct{}

func (*SrPoliciesDump) GetMessageName() string {
	return "sr_policies_dump"
}
func (*SrPoliciesDump) GetCrcString() string {
	return "51077d14"
}
func (*SrPoliciesDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SrPolicyAdd represents VPP binary API message 'sr_policy_add':
type SrPolicyAdd struct {
	BsidAddr []byte `struc:"[16]byte"`
	Weight   uint32
	IsEncap  uint8
	Type     uint8
	FibTable uint32
	Sids     Srv6SidList
}

func (*SrPolicyAdd) GetMessageName() string {
	return "sr_policy_add"
}
func (*SrPolicyAdd) GetCrcString() string {
	return "4b6e2484"
}
func (*SrPolicyAdd) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SrPolicyAddReply represents VPP binary API message 'sr_policy_add_reply':
type SrPolicyAddReply struct {
	Retval int32
}

func (*SrPolicyAddReply) GetMessageName() string {
	return "sr_policy_add_reply"
}
func (*SrPolicyAddReply) GetCrcString() string {
	return "e8d4e804"
}
func (*SrPolicyAddReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SrPolicyDel represents VPP binary API message 'sr_policy_del':
type SrPolicyDel struct {
	BsidAddr      Srv6Sid
	SrPolicyIndex uint32
}

func (*SrPolicyDel) GetMessageName() string {
	return "sr_policy_del"
}
func (*SrPolicyDel) GetCrcString() string {
	return "e4133171"
}
func (*SrPolicyDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SrPolicyDelReply represents VPP binary API message 'sr_policy_del_reply':
type SrPolicyDelReply struct {
	Retval int32
}

func (*SrPolicyDelReply) GetMessageName() string {
	return "sr_policy_del_reply"
}
func (*SrPolicyDelReply) GetCrcString() string {
	return "e8d4e804"
}
func (*SrPolicyDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SrPolicyMod represents VPP binary API message 'sr_policy_mod':
type SrPolicyMod struct {
	BsidAddr      []byte `struc:"[16]byte"`
	SrPolicyIndex uint32
	FibTable      uint32
	Operation     uint8
	SlIndex       uint32
	Weight        uint32
	Sids          Srv6SidList
}

func (*SrPolicyMod) GetMessageName() string {
	return "sr_policy_mod"
}
func (*SrPolicyMod) GetCrcString() string {
	return "c1dfaee0"
}
func (*SrPolicyMod) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SrPolicyModReply represents VPP binary API message 'sr_policy_mod_reply':
type SrPolicyModReply struct {
	Retval int32
}

func (*SrPolicyModReply) GetMessageName() string {
	return "sr_policy_mod_reply"
}
func (*SrPolicyModReply) GetCrcString() string {
	return "e8d4e804"
}
func (*SrPolicyModReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SrSetEncapSource represents VPP binary API message 'sr_set_encap_source':
type SrSetEncapSource struct {
	EncapsSource []byte `struc:"[16]byte"`
}

func (*SrSetEncapSource) GetMessageName() string {
	return "sr_set_encap_source"
}
func (*SrSetEncapSource) GetCrcString() string {
	return "d05bb4de"
}
func (*SrSetEncapSource) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SrSetEncapSourceReply represents VPP binary API message 'sr_set_encap_source_reply':
type SrSetEncapSourceReply struct {
	Retval int32
}

func (*SrSetEncapSourceReply) GetMessageName() string {
	return "sr_set_encap_source_reply"
}
func (*SrSetEncapSourceReply) GetCrcString() string {
	return "e8d4e804"
}
func (*SrSetEncapSourceReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SrSteeringAddDel represents VPP binary API message 'sr_steering_add_del':
type SrSteeringAddDel struct {
	IsDel         uint8
	BsidAddr      []byte `struc:"[16]byte"`
	SrPolicyIndex uint32
	TableID       uint32
	PrefixAddr    []byte `struc:"[16]byte"`
	MaskWidth     uint32
	SwIfIndex     uint32
	TrafficType   uint8
}

func (*SrSteeringAddDel) GetMessageName() string {
	return "sr_steering_add_del"
}
func (*SrSteeringAddDel) GetCrcString() string {
	return "28b5dcab"
}
func (*SrSteeringAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// SrSteeringAddDelReply represents VPP binary API message 'sr_steering_add_del_reply':
type SrSteeringAddDelReply struct {
	Retval int32
}

func (*SrSteeringAddDelReply) GetMessageName() string {
	return "sr_steering_add_del_reply"
}
func (*SrSteeringAddDelReply) GetCrcString() string {
	return "e8d4e804"
}
func (*SrSteeringAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SrSteeringPolDetails represents VPP binary API message 'sr_steering_pol_details':
type SrSteeringPolDetails struct {
	TrafficType uint8
	FibTable    uint32
	PrefixAddr  []byte `struc:"[16]byte"`
	MaskWidth   uint32
	SwIfIndex   uint32
	Bsid        Srv6Sid
}

func (*SrSteeringPolDetails) GetMessageName() string {
	return "sr_steering_pol_details"
}
func (*SrSteeringPolDetails) GetCrcString() string {
	return "5627d483"
}
func (*SrSteeringPolDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// SrSteeringPolDump represents VPP binary API message 'sr_steering_pol_dump':
type SrSteeringPolDump struct{}

func (*SrSteeringPolDump) GetMessageName() string {
	return "sr_steering_pol_dump"
}
func (*SrSteeringPolDump) GetCrcString() string {
	return "51077d14"
}
func (*SrSteeringPolDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func init() {
	api.RegisterMessage((*SrLocalsidAddDel)(nil), "sr.SrLocalsidAddDel")
	api.RegisterMessage((*SrLocalsidAddDelReply)(nil), "sr.SrLocalsidAddDelReply")
	api.RegisterMessage((*SrLocalsidsDetails)(nil), "sr.SrLocalsidsDetails")
	api.RegisterMessage((*SrLocalsidsDump)(nil), "sr.SrLocalsidsDump")
	api.RegisterMessage((*SrPoliciesDetails)(nil), "sr.SrPoliciesDetails")
	api.RegisterMessage((*SrPoliciesDump)(nil), "sr.SrPoliciesDump")
	api.RegisterMessage((*SrPolicyAdd)(nil), "sr.SrPolicyAdd")
	api.RegisterMessage((*SrPolicyAddReply)(nil), "sr.SrPolicyAddReply")
	api.RegisterMessage((*SrPolicyDel)(nil), "sr.SrPolicyDel")
	api.RegisterMessage((*SrPolicyDelReply)(nil), "sr.SrPolicyDelReply")
	api.RegisterMessage((*SrPolicyMod)(nil), "sr.SrPolicyMod")
	api.RegisterMessage((*SrPolicyModReply)(nil), "sr.SrPolicyModReply")
	api.RegisterMessage((*SrSetEncapSource)(nil), "sr.SrSetEncapSource")
	api.RegisterMessage((*SrSetEncapSourceReply)(nil), "sr.SrSetEncapSourceReply")
	api.RegisterMessage((*SrSteeringAddDel)(nil), "sr.SrSteeringAddDel")
	api.RegisterMessage((*SrSteeringAddDelReply)(nil), "sr.SrSteeringAddDelReply")
	api.RegisterMessage((*SrSteeringPolDetails)(nil), "sr.SrSteeringPolDetails")
	api.RegisterMessage((*SrSteeringPolDump)(nil), "sr.SrSteeringPolDump")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*SrLocalsidAddDel)(nil),
		(*SrLocalsidAddDelReply)(nil),
		(*SrLocalsidsDetails)(nil),
		(*SrLocalsidsDump)(nil),
		(*SrPoliciesDetails)(nil),
		(*SrPoliciesDump)(nil),
		(*SrPolicyAdd)(nil),
		(*SrPolicyAddReply)(nil),
		(*SrPolicyDel)(nil),
		(*SrPolicyDelReply)(nil),
		(*SrPolicyMod)(nil),
		(*SrPolicyModReply)(nil),
		(*SrSetEncapSource)(nil),
		(*SrSetEncapSourceReply)(nil),
		(*SrSteeringAddDel)(nil),
		(*SrSteeringAddDelReply)(nil),
		(*SrSteeringPolDetails)(nil),
		(*SrSteeringPolDump)(nil),
	}
}
