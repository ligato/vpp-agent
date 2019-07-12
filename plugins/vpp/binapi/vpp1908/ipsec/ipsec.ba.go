// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/ipsec.api.json

/*
Package ipsec is a generated VPP binary API for 'ipsec' module.

It consists of:
	  9 enums
	  3 aliases
	  9 types
	  1 union
	 30 messages
	 15 services
*/
package ipsec

import (
	bytes "bytes"
	context "context"
	api "git.fd.io/govpp.git/api"
	struc "github.com/lunixbochs/struc"
	io "io"
	strconv "strconv"
)

const (
	// ModuleName is the name of this module.
	ModuleName = "ipsec"
	// APIVersion is the API version of this module.
	APIVersion = "3.0.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0xd2452344
)

// AddressFamily represents VPP binary API enum 'address_family'.
type AddressFamily uint32

const (
	ADDRESS_IP4 AddressFamily = 0
	ADDRESS_IP6 AddressFamily = 1
)

var AddressFamily_name = map[uint32]string{
	0: "ADDRESS_IP4",
	1: "ADDRESS_IP6",
}

var AddressFamily_value = map[string]uint32{
	"ADDRESS_IP4": 0,
	"ADDRESS_IP6": 1,
}

func (x AddressFamily) String() string {
	s, ok := AddressFamily_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IPDscp represents VPP binary API enum 'ip_dscp'.
type IPDscp uint8

const (
	IP_API_DSCP_CS0  IPDscp = 0
	IP_API_DSCP_CS1  IPDscp = 8
	IP_API_DSCP_AF11 IPDscp = 10
	IP_API_DSCP_AF12 IPDscp = 12
	IP_API_DSCP_AF13 IPDscp = 14
	IP_API_DSCP_CS2  IPDscp = 16
	IP_API_DSCP_AF21 IPDscp = 18
	IP_API_DSCP_AF22 IPDscp = 20
	IP_API_DSCP_AF23 IPDscp = 22
	IP_API_DSCP_CS3  IPDscp = 24
	IP_API_DSCP_AF31 IPDscp = 26
	IP_API_DSCP_AF32 IPDscp = 28
	IP_API_DSCP_AF33 IPDscp = 30
	IP_API_DSCP_CS4  IPDscp = 32
	IP_API_DSCP_AF41 IPDscp = 34
	IP_API_DSCP_AF42 IPDscp = 36
	IP_API_DSCP_AF43 IPDscp = 38
	IP_API_DSCP_CS5  IPDscp = 40
	IP_API_DSCP_EF   IPDscp = 46
	IP_API_DSCP_CS6  IPDscp = 48
	IP_API_DSCP_CS7  IPDscp = 50
)

var IPDscp_name = map[uint8]string{
	0:  "IP_API_DSCP_CS0",
	8:  "IP_API_DSCP_CS1",
	10: "IP_API_DSCP_AF11",
	12: "IP_API_DSCP_AF12",
	14: "IP_API_DSCP_AF13",
	16: "IP_API_DSCP_CS2",
	18: "IP_API_DSCP_AF21",
	20: "IP_API_DSCP_AF22",
	22: "IP_API_DSCP_AF23",
	24: "IP_API_DSCP_CS3",
	26: "IP_API_DSCP_AF31",
	28: "IP_API_DSCP_AF32",
	30: "IP_API_DSCP_AF33",
	32: "IP_API_DSCP_CS4",
	34: "IP_API_DSCP_AF41",
	36: "IP_API_DSCP_AF42",
	38: "IP_API_DSCP_AF43",
	40: "IP_API_DSCP_CS5",
	46: "IP_API_DSCP_EF",
	48: "IP_API_DSCP_CS6",
	50: "IP_API_DSCP_CS7",
}

var IPDscp_value = map[string]uint8{
	"IP_API_DSCP_CS0":  0,
	"IP_API_DSCP_CS1":  8,
	"IP_API_DSCP_AF11": 10,
	"IP_API_DSCP_AF12": 12,
	"IP_API_DSCP_AF13": 14,
	"IP_API_DSCP_CS2":  16,
	"IP_API_DSCP_AF21": 18,
	"IP_API_DSCP_AF22": 20,
	"IP_API_DSCP_AF23": 22,
	"IP_API_DSCP_CS3":  24,
	"IP_API_DSCP_AF31": 26,
	"IP_API_DSCP_AF32": 28,
	"IP_API_DSCP_AF33": 30,
	"IP_API_DSCP_CS4":  32,
	"IP_API_DSCP_AF41": 34,
	"IP_API_DSCP_AF42": 36,
	"IP_API_DSCP_AF43": 38,
	"IP_API_DSCP_CS5":  40,
	"IP_API_DSCP_EF":   46,
	"IP_API_DSCP_CS6":  48,
	"IP_API_DSCP_CS7":  50,
}

func (x IPDscp) String() string {
	s, ok := IPDscp_name[uint8(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IPEcn represents VPP binary API enum 'ip_ecn'.
type IPEcn uint8

const (
	IP_API_ECN_NONE IPEcn = 0
	IP_API_ECN_ECT0 IPEcn = 1
	IP_API_ECN_ECT1 IPEcn = 2
	IP_API_ECN_CE   IPEcn = 3
)

var IPEcn_name = map[uint8]string{
	0: "IP_API_ECN_NONE",
	1: "IP_API_ECN_ECT0",
	2: "IP_API_ECN_ECT1",
	3: "IP_API_ECN_CE",
}

var IPEcn_value = map[string]uint8{
	"IP_API_ECN_NONE": 0,
	"IP_API_ECN_ECT0": 1,
	"IP_API_ECN_ECT1": 2,
	"IP_API_ECN_CE":   3,
}

func (x IPEcn) String() string {
	s, ok := IPEcn_name[uint8(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IPProto represents VPP binary API enum 'ip_proto'.
type IPProto uint32

const (
	IP_API_PROTO_HOPOPT   IPProto = 0
	IP_API_PROTO_ICMP     IPProto = 1
	IP_API_PROTO_IGMP     IPProto = 2
	IP_API_PROTO_TCP      IPProto = 6
	IP_API_PROTO_UDP      IPProto = 17
	IP_API_PROTO_GRE      IPProto = 47
	IP_API_PROTO_AH       IPProto = 50
	IP_API_PROTO_ESP      IPProto = 51
	IP_API_PROTO_EIGRP    IPProto = 88
	IP_API_PROTO_OSPF     IPProto = 89
	IP_API_PROTO_SCTP     IPProto = 132
	IP_API_PROTO_RESERVED IPProto = 255
)

var IPProto_name = map[uint32]string{
	0:   "IP_API_PROTO_HOPOPT",
	1:   "IP_API_PROTO_ICMP",
	2:   "IP_API_PROTO_IGMP",
	6:   "IP_API_PROTO_TCP",
	17:  "IP_API_PROTO_UDP",
	47:  "IP_API_PROTO_GRE",
	50:  "IP_API_PROTO_AH",
	51:  "IP_API_PROTO_ESP",
	88:  "IP_API_PROTO_EIGRP",
	89:  "IP_API_PROTO_OSPF",
	132: "IP_API_PROTO_SCTP",
	255: "IP_API_PROTO_RESERVED",
}

var IPProto_value = map[string]uint32{
	"IP_API_PROTO_HOPOPT":   0,
	"IP_API_PROTO_ICMP":     1,
	"IP_API_PROTO_IGMP":     2,
	"IP_API_PROTO_TCP":      6,
	"IP_API_PROTO_UDP":      17,
	"IP_API_PROTO_GRE":      47,
	"IP_API_PROTO_AH":       50,
	"IP_API_PROTO_ESP":      51,
	"IP_API_PROTO_EIGRP":    88,
	"IP_API_PROTO_OSPF":     89,
	"IP_API_PROTO_SCTP":     132,
	"IP_API_PROTO_RESERVED": 255,
}

func (x IPProto) String() string {
	s, ok := IPProto_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IpsecCryptoAlg represents VPP binary API enum 'ipsec_crypto_alg'.
type IpsecCryptoAlg uint32

const (
	IPSEC_API_CRYPTO_ALG_NONE        IpsecCryptoAlg = 0
	IPSEC_API_CRYPTO_ALG_AES_CBC_128 IpsecCryptoAlg = 1
	IPSEC_API_CRYPTO_ALG_AES_CBC_192 IpsecCryptoAlg = 2
	IPSEC_API_CRYPTO_ALG_AES_CBC_256 IpsecCryptoAlg = 3
	IPSEC_API_CRYPTO_ALG_AES_CTR_128 IpsecCryptoAlg = 4
	IPSEC_API_CRYPTO_ALG_AES_CTR_192 IpsecCryptoAlg = 5
	IPSEC_API_CRYPTO_ALG_AES_CTR_256 IpsecCryptoAlg = 6
	IPSEC_API_CRYPTO_ALG_AES_GCM_128 IpsecCryptoAlg = 7
	IPSEC_API_CRYPTO_ALG_AES_GCM_192 IpsecCryptoAlg = 8
	IPSEC_API_CRYPTO_ALG_AES_GCM_256 IpsecCryptoAlg = 9
	IPSEC_API_CRYPTO_ALG_DES_CBC     IpsecCryptoAlg = 10
	IPSEC_API_CRYPTO_ALG_3DES_CBC    IpsecCryptoAlg = 11
)

var IpsecCryptoAlg_name = map[uint32]string{
	0:  "IPSEC_API_CRYPTO_ALG_NONE",
	1:  "IPSEC_API_CRYPTO_ALG_AES_CBC_128",
	2:  "IPSEC_API_CRYPTO_ALG_AES_CBC_192",
	3:  "IPSEC_API_CRYPTO_ALG_AES_CBC_256",
	4:  "IPSEC_API_CRYPTO_ALG_AES_CTR_128",
	5:  "IPSEC_API_CRYPTO_ALG_AES_CTR_192",
	6:  "IPSEC_API_CRYPTO_ALG_AES_CTR_256",
	7:  "IPSEC_API_CRYPTO_ALG_AES_GCM_128",
	8:  "IPSEC_API_CRYPTO_ALG_AES_GCM_192",
	9:  "IPSEC_API_CRYPTO_ALG_AES_GCM_256",
	10: "IPSEC_API_CRYPTO_ALG_DES_CBC",
	11: "IPSEC_API_CRYPTO_ALG_3DES_CBC",
}

var IpsecCryptoAlg_value = map[string]uint32{
	"IPSEC_API_CRYPTO_ALG_NONE":        0,
	"IPSEC_API_CRYPTO_ALG_AES_CBC_128": 1,
	"IPSEC_API_CRYPTO_ALG_AES_CBC_192": 2,
	"IPSEC_API_CRYPTO_ALG_AES_CBC_256": 3,
	"IPSEC_API_CRYPTO_ALG_AES_CTR_128": 4,
	"IPSEC_API_CRYPTO_ALG_AES_CTR_192": 5,
	"IPSEC_API_CRYPTO_ALG_AES_CTR_256": 6,
	"IPSEC_API_CRYPTO_ALG_AES_GCM_128": 7,
	"IPSEC_API_CRYPTO_ALG_AES_GCM_192": 8,
	"IPSEC_API_CRYPTO_ALG_AES_GCM_256": 9,
	"IPSEC_API_CRYPTO_ALG_DES_CBC":     10,
	"IPSEC_API_CRYPTO_ALG_3DES_CBC":    11,
}

func (x IpsecCryptoAlg) String() string {
	s, ok := IpsecCryptoAlg_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IpsecIntegAlg represents VPP binary API enum 'ipsec_integ_alg'.
type IpsecIntegAlg uint32

const (
	IPSEC_API_INTEG_ALG_NONE        IpsecIntegAlg = 0
	IPSEC_API_INTEG_ALG_MD5_96      IpsecIntegAlg = 1
	IPSEC_API_INTEG_ALG_SHA1_96     IpsecIntegAlg = 2
	IPSEC_API_INTEG_ALG_SHA_256_96  IpsecIntegAlg = 3
	IPSEC_API_INTEG_ALG_SHA_256_128 IpsecIntegAlg = 4
	IPSEC_API_INTEG_ALG_SHA_384_192 IpsecIntegAlg = 5
	IPSEC_API_INTEG_ALG_SHA_512_256 IpsecIntegAlg = 6
)

var IpsecIntegAlg_name = map[uint32]string{
	0: "IPSEC_API_INTEG_ALG_NONE",
	1: "IPSEC_API_INTEG_ALG_MD5_96",
	2: "IPSEC_API_INTEG_ALG_SHA1_96",
	3: "IPSEC_API_INTEG_ALG_SHA_256_96",
	4: "IPSEC_API_INTEG_ALG_SHA_256_128",
	5: "IPSEC_API_INTEG_ALG_SHA_384_192",
	6: "IPSEC_API_INTEG_ALG_SHA_512_256",
}

var IpsecIntegAlg_value = map[string]uint32{
	"IPSEC_API_INTEG_ALG_NONE":        0,
	"IPSEC_API_INTEG_ALG_MD5_96":      1,
	"IPSEC_API_INTEG_ALG_SHA1_96":     2,
	"IPSEC_API_INTEG_ALG_SHA_256_96":  3,
	"IPSEC_API_INTEG_ALG_SHA_256_128": 4,
	"IPSEC_API_INTEG_ALG_SHA_384_192": 5,
	"IPSEC_API_INTEG_ALG_SHA_512_256": 6,
}

func (x IpsecIntegAlg) String() string {
	s, ok := IpsecIntegAlg_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IpsecProto represents VPP binary API enum 'ipsec_proto'.
type IpsecProto uint32

const (
	IPSEC_API_PROTO_ESP IpsecProto = 1
	IPSEC_API_PROTO_AH  IpsecProto = 2
)

var IpsecProto_name = map[uint32]string{
	1: "IPSEC_API_PROTO_ESP",
	2: "IPSEC_API_PROTO_AH",
}

var IpsecProto_value = map[string]uint32{
	"IPSEC_API_PROTO_ESP": 1,
	"IPSEC_API_PROTO_AH":  2,
}

func (x IpsecProto) String() string {
	s, ok := IpsecProto_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IpsecSadFlags represents VPP binary API enum 'ipsec_sad_flags'.
type IpsecSadFlags uint32

const (
	IPSEC_API_SAD_FLAG_NONE            IpsecSadFlags = 0
	IPSEC_API_SAD_FLAG_USE_ESN         IpsecSadFlags = 1
	IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY IpsecSadFlags = 2
	IPSEC_API_SAD_FLAG_IS_TUNNEL       IpsecSadFlags = 4
	IPSEC_API_SAD_FLAG_IS_TUNNEL_V6    IpsecSadFlags = 8
	IPSEC_API_SAD_FLAG_UDP_ENCAP       IpsecSadFlags = 16
)

var IpsecSadFlags_name = map[uint32]string{
	0:  "IPSEC_API_SAD_FLAG_NONE",
	1:  "IPSEC_API_SAD_FLAG_USE_ESN",
	2:  "IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY",
	4:  "IPSEC_API_SAD_FLAG_IS_TUNNEL",
	8:  "IPSEC_API_SAD_FLAG_IS_TUNNEL_V6",
	16: "IPSEC_API_SAD_FLAG_UDP_ENCAP",
}

var IpsecSadFlags_value = map[string]uint32{
	"IPSEC_API_SAD_FLAG_NONE":            0,
	"IPSEC_API_SAD_FLAG_USE_ESN":         1,
	"IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY": 2,
	"IPSEC_API_SAD_FLAG_IS_TUNNEL":       4,
	"IPSEC_API_SAD_FLAG_IS_TUNNEL_V6":    8,
	"IPSEC_API_SAD_FLAG_UDP_ENCAP":       16,
}

func (x IpsecSadFlags) String() string {
	s, ok := IpsecSadFlags_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// IpsecSpdAction represents VPP binary API enum 'ipsec_spd_action'.
type IpsecSpdAction uint32

const (
	IPSEC_API_SPD_ACTION_BYPASS  IpsecSpdAction = 0
	IPSEC_API_SPD_ACTION_DISCARD IpsecSpdAction = 1
	IPSEC_API_SPD_ACTION_RESOLVE IpsecSpdAction = 2
	IPSEC_API_SPD_ACTION_PROTECT IpsecSpdAction = 3
)

var IpsecSpdAction_name = map[uint32]string{
	0: "IPSEC_API_SPD_ACTION_BYPASS",
	1: "IPSEC_API_SPD_ACTION_DISCARD",
	2: "IPSEC_API_SPD_ACTION_RESOLVE",
	3: "IPSEC_API_SPD_ACTION_PROTECT",
}

var IpsecSpdAction_value = map[string]uint32{
	"IPSEC_API_SPD_ACTION_BYPASS":  0,
	"IPSEC_API_SPD_ACTION_DISCARD": 1,
	"IPSEC_API_SPD_ACTION_RESOLVE": 2,
	"IPSEC_API_SPD_ACTION_PROTECT": 3,
}

func (x IpsecSpdAction) String() string {
	s, ok := IpsecSpdAction_name[uint32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

// InterfaceIndex represents VPP binary API alias 'interface_index'.
type InterfaceIndex uint32

// IP4Address represents VPP binary API alias 'ip4_address'.
type IP4Address [4]uint8

// IP6Address represents VPP binary API alias 'ip6_address'.
type IP6Address [16]uint8

// Address represents VPP binary API type 'address'.
type Address struct {
	Af AddressFamily
	Un AddressUnion
}

func (*Address) GetTypeName() string {
	return "address"
}

// IP4Prefix represents VPP binary API type 'ip4_prefix'.
type IP4Prefix struct {
	Address IP4Address
	Len     uint8
}

func (*IP4Prefix) GetTypeName() string {
	return "ip4_prefix"
}

// IP6Prefix represents VPP binary API type 'ip6_prefix'.
type IP6Prefix struct {
	Address IP6Address
	Len     uint8
}

func (*IP6Prefix) GetTypeName() string {
	return "ip6_prefix"
}

// IpsecSadEntry represents VPP binary API type 'ipsec_sad_entry'.
type IpsecSadEntry struct {
	SadID              uint32
	Spi                uint32
	Protocol           IpsecProto
	CryptoAlgorithm    IpsecCryptoAlg
	CryptoKey          Key
	IntegrityAlgorithm IpsecIntegAlg
	IntegrityKey       Key
	Flags              IpsecSadFlags
	TunnelSrc          Address
	TunnelDst          Address
	TxTableID          uint32
	Salt               uint32
}

func (*IpsecSadEntry) GetTypeName() string {
	return "ipsec_sad_entry"
}

// IpsecSpdEntry represents VPP binary API type 'ipsec_spd_entry'.
type IpsecSpdEntry struct {
	SpdID              uint32
	Priority           int32
	IsOutbound         uint8
	SaID               uint32
	Policy             IpsecSpdAction
	Protocol           uint8
	RemoteAddressStart Address
	RemoteAddressStop  Address
	LocalAddressStart  Address
	LocalAddressStop   Address
	RemotePortStart    uint16
	RemotePortStop     uint16
	LocalPortStart     uint16
	LocalPortStop      uint16
}

func (*IpsecSpdEntry) GetTypeName() string {
	return "ipsec_spd_entry"
}

// IpsecTunnelProtect represents VPP binary API type 'ipsec_tunnel_protect'.
type IpsecTunnelProtect struct {
	SwIfIndex InterfaceIndex
	SaOut     uint32
	NSaIn     uint8 `struc:"sizeof=SaIn"`
	SaIn      []uint32
}

func (*IpsecTunnelProtect) GetTypeName() string {
	return "ipsec_tunnel_protect"
}

// Key represents VPP binary API type 'key'.
type Key struct {
	Length uint8
	Data   []byte `struc:"[128]byte"`
}

func (*Key) GetTypeName() string {
	return "key"
}

// Mprefix represents VPP binary API type 'mprefix'.
type Mprefix struct {
	Af               AddressFamily
	GrpAddressLength uint16
	GrpAddress       AddressUnion
	SrcAddress       AddressUnion
}

func (*Mprefix) GetTypeName() string {
	return "mprefix"
}

// Prefix represents VPP binary API type 'prefix'.
type Prefix struct {
	Address Address
	Len     uint8
}

func (*Prefix) GetTypeName() string {
	return "prefix"
}

// AddressUnion represents VPP binary API union 'address_union'.
type AddressUnion struct {
	XXX_UnionData [16]byte
}

func (*AddressUnion) GetTypeName() string {
	return "address_union"
}

func AddressUnionIP4(a IP4Address) (u AddressUnion) {
	u.SetIP4(a)
	return
}
func (u *AddressUnion) SetIP4(a IP4Address) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.XXX_UnionData[:], b.Bytes())
}
func (u *AddressUnion) GetIP4() (a IP4Address) {
	var b = bytes.NewReader(u.XXX_UnionData[:])
	struc.Unpack(b, &a)
	return
}

func AddressUnionIP6(a IP6Address) (u AddressUnion) {
	u.SetIP6(a)
	return
}
func (u *AddressUnion) SetIP6(a IP6Address) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.XXX_UnionData[:], b.Bytes())
}
func (u *AddressUnion) GetIP6() (a IP6Address) {
	var b = bytes.NewReader(u.XXX_UnionData[:])
	struc.Unpack(b, &a)
	return
}

// IpsecBackendDetails represents VPP binary API message 'ipsec_backend_details'.
type IpsecBackendDetails struct {
	Name     []byte `struc:"[128]byte"`
	Protocol IpsecProto
	Index    uint8
	Active   uint8
}

func (*IpsecBackendDetails) GetMessageName() string {
	return "ipsec_backend_details"
}
func (*IpsecBackendDetails) GetCrcString() string {
	return "7700751c"
}
func (*IpsecBackendDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecBackendDump represents VPP binary API message 'ipsec_backend_dump'.
type IpsecBackendDump struct{}

func (*IpsecBackendDump) GetMessageName() string {
	return "ipsec_backend_dump"
}
func (*IpsecBackendDump) GetCrcString() string {
	return "51077d14"
}
func (*IpsecBackendDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecInterfaceAddDelSpd represents VPP binary API message 'ipsec_interface_add_del_spd'.
type IpsecInterfaceAddDelSpd struct {
	IsAdd     uint8
	SwIfIndex uint32
	SpdID     uint32
}

func (*IpsecInterfaceAddDelSpd) GetMessageName() string {
	return "ipsec_interface_add_del_spd"
}
func (*IpsecInterfaceAddDelSpd) GetCrcString() string {
	return "1e3b8286"
}
func (*IpsecInterfaceAddDelSpd) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecInterfaceAddDelSpdReply represents VPP binary API message 'ipsec_interface_add_del_spd_reply'.
type IpsecInterfaceAddDelSpdReply struct {
	Retval int32
}

func (*IpsecInterfaceAddDelSpdReply) GetMessageName() string {
	return "ipsec_interface_add_del_spd_reply"
}
func (*IpsecInterfaceAddDelSpdReply) GetCrcString() string {
	return "e8d4e804"
}
func (*IpsecInterfaceAddDelSpdReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecSaDetails represents VPP binary API message 'ipsec_sa_details'.
type IpsecSaDetails struct {
	Entry          IpsecSadEntry
	SwIfIndex      uint32
	Salt           uint32
	SeqOutbound    uint64
	LastSeqInbound uint64
	ReplayWindow   uint64
	TotalDataSize  uint64
}

func (*IpsecSaDetails) GetMessageName() string {
	return "ipsec_sa_details"
}
func (*IpsecSaDetails) GetCrcString() string {
	return "9c8d829a"
}
func (*IpsecSaDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecSaDump represents VPP binary API message 'ipsec_sa_dump'.
type IpsecSaDump struct {
	SaID uint32
}

func (*IpsecSaDump) GetMessageName() string {
	return "ipsec_sa_dump"
}
func (*IpsecSaDump) GetCrcString() string {
	return "2076c2f4"
}
func (*IpsecSaDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecSadEntryAddDel represents VPP binary API message 'ipsec_sad_entry_add_del'.
type IpsecSadEntryAddDel struct {
	IsAdd uint8
	Entry IpsecSadEntry
}

func (*IpsecSadEntryAddDel) GetMessageName() string {
	return "ipsec_sad_entry_add_del"
}
func (*IpsecSadEntryAddDel) GetCrcString() string {
	return "a25ab61e"
}
func (*IpsecSadEntryAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecSadEntryAddDelReply represents VPP binary API message 'ipsec_sad_entry_add_del_reply'.
type IpsecSadEntryAddDelReply struct {
	Retval    int32
	StatIndex uint32
}

func (*IpsecSadEntryAddDelReply) GetMessageName() string {
	return "ipsec_sad_entry_add_del_reply"
}
func (*IpsecSadEntryAddDelReply) GetCrcString() string {
	return "9ffac24b"
}
func (*IpsecSadEntryAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecSelectBackend represents VPP binary API message 'ipsec_select_backend'.
type IpsecSelectBackend struct {
	Protocol IpsecProto
	Index    uint8
}

func (*IpsecSelectBackend) GetMessageName() string {
	return "ipsec_select_backend"
}
func (*IpsecSelectBackend) GetCrcString() string {
	return "4fd24836"
}
func (*IpsecSelectBackend) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecSelectBackendReply represents VPP binary API message 'ipsec_select_backend_reply'.
type IpsecSelectBackendReply struct {
	Retval int32
}

func (*IpsecSelectBackendReply) GetMessageName() string {
	return "ipsec_select_backend_reply"
}
func (*IpsecSelectBackendReply) GetCrcString() string {
	return "e8d4e804"
}
func (*IpsecSelectBackendReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecSpdAddDel represents VPP binary API message 'ipsec_spd_add_del'.
type IpsecSpdAddDel struct {
	IsAdd uint8
	SpdID uint32
}

func (*IpsecSpdAddDel) GetMessageName() string {
	return "ipsec_spd_add_del"
}
func (*IpsecSpdAddDel) GetCrcString() string {
	return "9ffdf5da"
}
func (*IpsecSpdAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecSpdAddDelReply represents VPP binary API message 'ipsec_spd_add_del_reply'.
type IpsecSpdAddDelReply struct {
	Retval int32
}

func (*IpsecSpdAddDelReply) GetMessageName() string {
	return "ipsec_spd_add_del_reply"
}
func (*IpsecSpdAddDelReply) GetCrcString() string {
	return "e8d4e804"
}
func (*IpsecSpdAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecSpdDetails represents VPP binary API message 'ipsec_spd_details'.
type IpsecSpdDetails struct {
	Entry IpsecSpdEntry
}

func (*IpsecSpdDetails) GetMessageName() string {
	return "ipsec_spd_details"
}
func (*IpsecSpdDetails) GetCrcString() string {
	return "06df7fb3"
}
func (*IpsecSpdDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecSpdDump represents VPP binary API message 'ipsec_spd_dump'.
type IpsecSpdDump struct {
	SpdID uint32
	SaID  uint32
}

func (*IpsecSpdDump) GetMessageName() string {
	return "ipsec_spd_dump"
}
func (*IpsecSpdDump) GetCrcString() string {
	return "afefbf7d"
}
func (*IpsecSpdDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecSpdEntryAddDel represents VPP binary API message 'ipsec_spd_entry_add_del'.
type IpsecSpdEntryAddDel struct {
	IsAdd uint8
	Entry IpsecSpdEntry
}

func (*IpsecSpdEntryAddDel) GetMessageName() string {
	return "ipsec_spd_entry_add_del"
}
func (*IpsecSpdEntryAddDel) GetCrcString() string {
	return "6bc6a3b5"
}
func (*IpsecSpdEntryAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecSpdEntryAddDelReply represents VPP binary API message 'ipsec_spd_entry_add_del_reply'.
type IpsecSpdEntryAddDelReply struct {
	Retval    int32
	StatIndex uint32
}

func (*IpsecSpdEntryAddDelReply) GetMessageName() string {
	return "ipsec_spd_entry_add_del_reply"
}
func (*IpsecSpdEntryAddDelReply) GetCrcString() string {
	return "9ffac24b"
}
func (*IpsecSpdEntryAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecSpdInterfaceDetails represents VPP binary API message 'ipsec_spd_interface_details'.
type IpsecSpdInterfaceDetails struct {
	SpdIndex  uint32
	SwIfIndex uint32
}

func (*IpsecSpdInterfaceDetails) GetMessageName() string {
	return "ipsec_spd_interface_details"
}
func (*IpsecSpdInterfaceDetails) GetCrcString() string {
	return "2c54296d"
}
func (*IpsecSpdInterfaceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecSpdInterfaceDump represents VPP binary API message 'ipsec_spd_interface_dump'.
type IpsecSpdInterfaceDump struct {
	SpdIndex      uint32
	SpdIndexValid uint8
}

func (*IpsecSpdInterfaceDump) GetMessageName() string {
	return "ipsec_spd_interface_dump"
}
func (*IpsecSpdInterfaceDump) GetCrcString() string {
	return "8971de19"
}
func (*IpsecSpdInterfaceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecSpdsDetails represents VPP binary API message 'ipsec_spds_details'.
type IpsecSpdsDetails struct {
	SpdID     uint32
	Npolicies uint32
}

func (*IpsecSpdsDetails) GetMessageName() string {
	return "ipsec_spds_details"
}
func (*IpsecSpdsDetails) GetCrcString() string {
	return "a04bb254"
}
func (*IpsecSpdsDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecSpdsDump represents VPP binary API message 'ipsec_spds_dump'.
type IpsecSpdsDump struct{}

func (*IpsecSpdsDump) GetMessageName() string {
	return "ipsec_spds_dump"
}
func (*IpsecSpdsDump) GetCrcString() string {
	return "51077d14"
}
func (*IpsecSpdsDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecTunnelIfAddDel represents VPP binary API message 'ipsec_tunnel_if_add_del'.
type IpsecTunnelIfAddDel struct {
	IsAdd              uint8
	Esn                uint8
	AntiReplay         uint8
	LocalIP            Address
	RemoteIP           Address
	LocalSpi           uint32
	RemoteSpi          uint32
	CryptoAlg          uint8
	LocalCryptoKeyLen  uint8
	LocalCryptoKey     []byte `struc:"[128]byte"`
	RemoteCryptoKeyLen uint8
	RemoteCryptoKey    []byte `struc:"[128]byte"`
	IntegAlg           uint8
	LocalIntegKeyLen   uint8
	LocalIntegKey      []byte `struc:"[128]byte"`
	RemoteIntegKeyLen  uint8
	RemoteIntegKey     []byte `struc:"[128]byte"`
	Renumber           uint8
	ShowInstance       uint32
	UDPEncap           uint8
	TxTableID          uint32
	Salt               uint32
}

func (*IpsecTunnelIfAddDel) GetMessageName() string {
	return "ipsec_tunnel_if_add_del"
}
func (*IpsecTunnelIfAddDel) GetCrcString() string {
	return "aa539b47"
}
func (*IpsecTunnelIfAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecTunnelIfAddDelReply represents VPP binary API message 'ipsec_tunnel_if_add_del_reply'.
type IpsecTunnelIfAddDelReply struct {
	Retval    int32
	SwIfIndex uint32
}

func (*IpsecTunnelIfAddDelReply) GetMessageName() string {
	return "ipsec_tunnel_if_add_del_reply"
}
func (*IpsecTunnelIfAddDelReply) GetCrcString() string {
	return "fda5941f"
}
func (*IpsecTunnelIfAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecTunnelIfSetSa represents VPP binary API message 'ipsec_tunnel_if_set_sa'.
type IpsecTunnelIfSetSa struct {
	SwIfIndex  uint32
	SaID       uint32
	IsOutbound uint8
}

func (*IpsecTunnelIfSetSa) GetMessageName() string {
	return "ipsec_tunnel_if_set_sa"
}
func (*IpsecTunnelIfSetSa) GetCrcString() string {
	return "6ab567f2"
}
func (*IpsecTunnelIfSetSa) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecTunnelIfSetSaReply represents VPP binary API message 'ipsec_tunnel_if_set_sa_reply'.
type IpsecTunnelIfSetSaReply struct {
	Retval int32
}

func (*IpsecTunnelIfSetSaReply) GetMessageName() string {
	return "ipsec_tunnel_if_set_sa_reply"
}
func (*IpsecTunnelIfSetSaReply) GetCrcString() string {
	return "e8d4e804"
}
func (*IpsecTunnelIfSetSaReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecTunnelProtectDel represents VPP binary API message 'ipsec_tunnel_protect_del'.
type IpsecTunnelProtectDel struct {
	SwIfIndex InterfaceIndex
}

func (*IpsecTunnelProtectDel) GetMessageName() string {
	return "ipsec_tunnel_protect_del"
}
func (*IpsecTunnelProtectDel) GetCrcString() string {
	return "d85aab0d"
}
func (*IpsecTunnelProtectDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecTunnelProtectDelReply represents VPP binary API message 'ipsec_tunnel_protect_del_reply'.
type IpsecTunnelProtectDelReply struct {
	Retval int32
}

func (*IpsecTunnelProtectDelReply) GetMessageName() string {
	return "ipsec_tunnel_protect_del_reply"
}
func (*IpsecTunnelProtectDelReply) GetCrcString() string {
	return "e8d4e804"
}
func (*IpsecTunnelProtectDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecTunnelProtectDetails represents VPP binary API message 'ipsec_tunnel_protect_details'.
type IpsecTunnelProtectDetails struct {
	Tun IpsecTunnelProtect
}

func (*IpsecTunnelProtectDetails) GetMessageName() string {
	return "ipsec_tunnel_protect_details"
}
func (*IpsecTunnelProtectDetails) GetCrcString() string {
	return "f724bc50"
}
func (*IpsecTunnelProtectDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

// IpsecTunnelProtectDump represents VPP binary API message 'ipsec_tunnel_protect_dump'.
type IpsecTunnelProtectDump struct {
	SwIfIndex InterfaceIndex
}

func (*IpsecTunnelProtectDump) GetMessageName() string {
	return "ipsec_tunnel_protect_dump"
}
func (*IpsecTunnelProtectDump) GetCrcString() string {
	return "d85aab0d"
}
func (*IpsecTunnelProtectDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecTunnelProtectUpdate represents VPP binary API message 'ipsec_tunnel_protect_update'.
type IpsecTunnelProtectUpdate struct {
	Tunnel IpsecTunnelProtect
}

func (*IpsecTunnelProtectUpdate) GetMessageName() string {
	return "ipsec_tunnel_protect_update"
}
func (*IpsecTunnelProtectUpdate) GetCrcString() string {
	return "316dab99"
}
func (*IpsecTunnelProtectUpdate) GetMessageType() api.MessageType {
	return api.RequestMessage
}

// IpsecTunnelProtectUpdateReply represents VPP binary API message 'ipsec_tunnel_protect_update_reply'.
type IpsecTunnelProtectUpdateReply struct {
	Retval int32
}

func (*IpsecTunnelProtectUpdateReply) GetMessageName() string {
	return "ipsec_tunnel_protect_update_reply"
}
func (*IpsecTunnelProtectUpdateReply) GetCrcString() string {
	return "e8d4e804"
}
func (*IpsecTunnelProtectUpdateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func init() {
	api.RegisterMessage((*IpsecBackendDetails)(nil), "ipsec.IpsecBackendDetails")
	api.RegisterMessage((*IpsecBackendDump)(nil), "ipsec.IpsecBackendDump")
	api.RegisterMessage((*IpsecInterfaceAddDelSpd)(nil), "ipsec.IpsecInterfaceAddDelSpd")
	api.RegisterMessage((*IpsecInterfaceAddDelSpdReply)(nil), "ipsec.IpsecInterfaceAddDelSpdReply")
	api.RegisterMessage((*IpsecSaDetails)(nil), "ipsec.IpsecSaDetails")
	api.RegisterMessage((*IpsecSaDump)(nil), "ipsec.IpsecSaDump")
	api.RegisterMessage((*IpsecSadEntryAddDel)(nil), "ipsec.IpsecSadEntryAddDel")
	api.RegisterMessage((*IpsecSadEntryAddDelReply)(nil), "ipsec.IpsecSadEntryAddDelReply")
	api.RegisterMessage((*IpsecSelectBackend)(nil), "ipsec.IpsecSelectBackend")
	api.RegisterMessage((*IpsecSelectBackendReply)(nil), "ipsec.IpsecSelectBackendReply")
	api.RegisterMessage((*IpsecSpdAddDel)(nil), "ipsec.IpsecSpdAddDel")
	api.RegisterMessage((*IpsecSpdAddDelReply)(nil), "ipsec.IpsecSpdAddDelReply")
	api.RegisterMessage((*IpsecSpdDetails)(nil), "ipsec.IpsecSpdDetails")
	api.RegisterMessage((*IpsecSpdDump)(nil), "ipsec.IpsecSpdDump")
	api.RegisterMessage((*IpsecSpdEntryAddDel)(nil), "ipsec.IpsecSpdEntryAddDel")
	api.RegisterMessage((*IpsecSpdEntryAddDelReply)(nil), "ipsec.IpsecSpdEntryAddDelReply")
	api.RegisterMessage((*IpsecSpdInterfaceDetails)(nil), "ipsec.IpsecSpdInterfaceDetails")
	api.RegisterMessage((*IpsecSpdInterfaceDump)(nil), "ipsec.IpsecSpdInterfaceDump")
	api.RegisterMessage((*IpsecSpdsDetails)(nil), "ipsec.IpsecSpdsDetails")
	api.RegisterMessage((*IpsecSpdsDump)(nil), "ipsec.IpsecSpdsDump")
	api.RegisterMessage((*IpsecTunnelIfAddDel)(nil), "ipsec.IpsecTunnelIfAddDel")
	api.RegisterMessage((*IpsecTunnelIfAddDelReply)(nil), "ipsec.IpsecTunnelIfAddDelReply")
	api.RegisterMessage((*IpsecTunnelIfSetSa)(nil), "ipsec.IpsecTunnelIfSetSa")
	api.RegisterMessage((*IpsecTunnelIfSetSaReply)(nil), "ipsec.IpsecTunnelIfSetSaReply")
	api.RegisterMessage((*IpsecTunnelProtectDel)(nil), "ipsec.IpsecTunnelProtectDel")
	api.RegisterMessage((*IpsecTunnelProtectDelReply)(nil), "ipsec.IpsecTunnelProtectDelReply")
	api.RegisterMessage((*IpsecTunnelProtectDetails)(nil), "ipsec.IpsecTunnelProtectDetails")
	api.RegisterMessage((*IpsecTunnelProtectDump)(nil), "ipsec.IpsecTunnelProtectDump")
	api.RegisterMessage((*IpsecTunnelProtectUpdate)(nil), "ipsec.IpsecTunnelProtectUpdate")
	api.RegisterMessage((*IpsecTunnelProtectUpdateReply)(nil), "ipsec.IpsecTunnelProtectUpdateReply")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*IpsecBackendDetails)(nil),
		(*IpsecBackendDump)(nil),
		(*IpsecInterfaceAddDelSpd)(nil),
		(*IpsecInterfaceAddDelSpdReply)(nil),
		(*IpsecSaDetails)(nil),
		(*IpsecSaDump)(nil),
		(*IpsecSadEntryAddDel)(nil),
		(*IpsecSadEntryAddDelReply)(nil),
		(*IpsecSelectBackend)(nil),
		(*IpsecSelectBackendReply)(nil),
		(*IpsecSpdAddDel)(nil),
		(*IpsecSpdAddDelReply)(nil),
		(*IpsecSpdDetails)(nil),
		(*IpsecSpdDump)(nil),
		(*IpsecSpdEntryAddDel)(nil),
		(*IpsecSpdEntryAddDelReply)(nil),
		(*IpsecSpdInterfaceDetails)(nil),
		(*IpsecSpdInterfaceDump)(nil),
		(*IpsecSpdsDetails)(nil),
		(*IpsecSpdsDump)(nil),
		(*IpsecTunnelIfAddDel)(nil),
		(*IpsecTunnelIfAddDelReply)(nil),
		(*IpsecTunnelIfSetSa)(nil),
		(*IpsecTunnelIfSetSaReply)(nil),
		(*IpsecTunnelProtectDel)(nil),
		(*IpsecTunnelProtectDelReply)(nil),
		(*IpsecTunnelProtectDetails)(nil),
		(*IpsecTunnelProtectDump)(nil),
		(*IpsecTunnelProtectUpdate)(nil),
		(*IpsecTunnelProtectUpdateReply)(nil),
	}
}

// RPCService represents RPC service API for ipsec module.
type RPCService interface {
	DumpIpsecBackend(ctx context.Context, in *IpsecBackendDump) (RPCService_DumpIpsecBackendClient, error)
	DumpIpsecSa(ctx context.Context, in *IpsecSaDump) (RPCService_DumpIpsecSaClient, error)
	DumpIpsecSpd(ctx context.Context, in *IpsecSpdDump) (RPCService_DumpIpsecSpdClient, error)
	DumpIpsecSpdInterface(ctx context.Context, in *IpsecSpdInterfaceDump) (RPCService_DumpIpsecSpdInterfaceClient, error)
	DumpIpsecSpds(ctx context.Context, in *IpsecSpdsDump) (RPCService_DumpIpsecSpdsClient, error)
	DumpIpsecTunnelProtect(ctx context.Context, in *IpsecTunnelProtectDump) (RPCService_DumpIpsecTunnelProtectClient, error)
	IpsecInterfaceAddDelSpd(ctx context.Context, in *IpsecInterfaceAddDelSpd) (*IpsecInterfaceAddDelSpdReply, error)
	IpsecSadEntryAddDel(ctx context.Context, in *IpsecSadEntryAddDel) (*IpsecSadEntryAddDelReply, error)
	IpsecSelectBackend(ctx context.Context, in *IpsecSelectBackend) (*IpsecSelectBackendReply, error)
	IpsecSpdAddDel(ctx context.Context, in *IpsecSpdAddDel) (*IpsecSpdAddDelReply, error)
	IpsecSpdEntryAddDel(ctx context.Context, in *IpsecSpdEntryAddDel) (*IpsecSpdEntryAddDelReply, error)
	IpsecTunnelIfAddDel(ctx context.Context, in *IpsecTunnelIfAddDel) (*IpsecTunnelIfAddDelReply, error)
	IpsecTunnelIfSetSa(ctx context.Context, in *IpsecTunnelIfSetSa) (*IpsecTunnelIfSetSaReply, error)
	IpsecTunnelProtectDel(ctx context.Context, in *IpsecTunnelProtectDel) (*IpsecTunnelProtectDelReply, error)
	IpsecTunnelProtectUpdate(ctx context.Context, in *IpsecTunnelProtectUpdate) (*IpsecTunnelProtectUpdateReply, error)
}

type serviceClient struct {
	ch api.Channel
}

func NewServiceClient(ch api.Channel) RPCService {
	return &serviceClient{ch}
}

func (c *serviceClient) DumpIpsecBackend(ctx context.Context, in *IpsecBackendDump) (RPCService_DumpIpsecBackendClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpsecBackendClient{stream}
	return x, nil
}

type RPCService_DumpIpsecBackendClient interface {
	Recv() (*IpsecBackendDetails, error)
}

type serviceClient_DumpIpsecBackendClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpsecBackendClient) Recv() (*IpsecBackendDetails, error) {
	m := new(IpsecBackendDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpIpsecSa(ctx context.Context, in *IpsecSaDump) (RPCService_DumpIpsecSaClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpsecSaClient{stream}
	return x, nil
}

type RPCService_DumpIpsecSaClient interface {
	Recv() (*IpsecSaDetails, error)
}

type serviceClient_DumpIpsecSaClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpsecSaClient) Recv() (*IpsecSaDetails, error) {
	m := new(IpsecSaDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpIpsecSpd(ctx context.Context, in *IpsecSpdDump) (RPCService_DumpIpsecSpdClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpsecSpdClient{stream}
	return x, nil
}

type RPCService_DumpIpsecSpdClient interface {
	Recv() (*IpsecSpdDetails, error)
}

type serviceClient_DumpIpsecSpdClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpsecSpdClient) Recv() (*IpsecSpdDetails, error) {
	m := new(IpsecSpdDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpIpsecSpdInterface(ctx context.Context, in *IpsecSpdInterfaceDump) (RPCService_DumpIpsecSpdInterfaceClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpsecSpdInterfaceClient{stream}
	return x, nil
}

type RPCService_DumpIpsecSpdInterfaceClient interface {
	Recv() (*IpsecSpdInterfaceDetails, error)
}

type serviceClient_DumpIpsecSpdInterfaceClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpsecSpdInterfaceClient) Recv() (*IpsecSpdInterfaceDetails, error) {
	m := new(IpsecSpdInterfaceDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpIpsecSpds(ctx context.Context, in *IpsecSpdsDump) (RPCService_DumpIpsecSpdsClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpsecSpdsClient{stream}
	return x, nil
}

type RPCService_DumpIpsecSpdsClient interface {
	Recv() (*IpsecSpdsDetails, error)
}

type serviceClient_DumpIpsecSpdsClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpsecSpdsClient) Recv() (*IpsecSpdsDetails, error) {
	m := new(IpsecSpdsDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpIpsecTunnelProtect(ctx context.Context, in *IpsecTunnelProtectDump) (RPCService_DumpIpsecTunnelProtectClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpsecTunnelProtectClient{stream}
	return x, nil
}

type RPCService_DumpIpsecTunnelProtectClient interface {
	Recv() (*IpsecTunnelProtectDetails, error)
}

type serviceClient_DumpIpsecTunnelProtectClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpsecTunnelProtectClient) Recv() (*IpsecTunnelProtectDetails, error) {
	m := new(IpsecTunnelProtectDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) IpsecInterfaceAddDelSpd(ctx context.Context, in *IpsecInterfaceAddDelSpd) (*IpsecInterfaceAddDelSpdReply, error) {
	out := new(IpsecInterfaceAddDelSpdReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpsecSadEntryAddDel(ctx context.Context, in *IpsecSadEntryAddDel) (*IpsecSadEntryAddDelReply, error) {
	out := new(IpsecSadEntryAddDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpsecSelectBackend(ctx context.Context, in *IpsecSelectBackend) (*IpsecSelectBackendReply, error) {
	out := new(IpsecSelectBackendReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpsecSpdAddDel(ctx context.Context, in *IpsecSpdAddDel) (*IpsecSpdAddDelReply, error) {
	out := new(IpsecSpdAddDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpsecSpdEntryAddDel(ctx context.Context, in *IpsecSpdEntryAddDel) (*IpsecSpdEntryAddDelReply, error) {
	out := new(IpsecSpdEntryAddDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpsecTunnelIfAddDel(ctx context.Context, in *IpsecTunnelIfAddDel) (*IpsecTunnelIfAddDelReply, error) {
	out := new(IpsecTunnelIfAddDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpsecTunnelIfSetSa(ctx context.Context, in *IpsecTunnelIfSetSa) (*IpsecTunnelIfSetSaReply, error) {
	out := new(IpsecTunnelIfSetSaReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpsecTunnelProtectDel(ctx context.Context, in *IpsecTunnelProtectDel) (*IpsecTunnelProtectDelReply, error) {
	out := new(IpsecTunnelProtectDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) IpsecTunnelProtectUpdate(ctx context.Context, in *IpsecTunnelProtectUpdate) (*IpsecTunnelProtectUpdateReply, error) {
	out := new(IpsecTunnelProtectUpdateReply)
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
