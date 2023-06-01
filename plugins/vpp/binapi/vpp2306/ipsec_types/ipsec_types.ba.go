// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

// Package ipsec_types contains generated bindings for API file ipsec_types.api.
//
// Contents:
// -  5 enums
// -  6 structs
package ipsec_types

import (
	"strconv"

	api "go.fd.io/govpp/api"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2306/interface_types"
	ip_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2306/ip_types"
	tunnel_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2306/tunnel_types"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "ipsec_types"
	APIVersion = "3.0.1"
	VersionCrc = 0x7892423b
)

// IpsecCryptoAlg defines enum 'ipsec_crypto_alg'.
type IpsecCryptoAlg uint32

const (
	IPSEC_API_CRYPTO_ALG_NONE              IpsecCryptoAlg = 0
	IPSEC_API_CRYPTO_ALG_AES_CBC_128       IpsecCryptoAlg = 1
	IPSEC_API_CRYPTO_ALG_AES_CBC_192       IpsecCryptoAlg = 2
	IPSEC_API_CRYPTO_ALG_AES_CBC_256       IpsecCryptoAlg = 3
	IPSEC_API_CRYPTO_ALG_AES_CTR_128       IpsecCryptoAlg = 4
	IPSEC_API_CRYPTO_ALG_AES_CTR_192       IpsecCryptoAlg = 5
	IPSEC_API_CRYPTO_ALG_AES_CTR_256       IpsecCryptoAlg = 6
	IPSEC_API_CRYPTO_ALG_AES_GCM_128       IpsecCryptoAlg = 7
	IPSEC_API_CRYPTO_ALG_AES_GCM_192       IpsecCryptoAlg = 8
	IPSEC_API_CRYPTO_ALG_AES_GCM_256       IpsecCryptoAlg = 9
	IPSEC_API_CRYPTO_ALG_DES_CBC           IpsecCryptoAlg = 10
	IPSEC_API_CRYPTO_ALG_3DES_CBC          IpsecCryptoAlg = 11
	IPSEC_API_CRYPTO_ALG_CHACHA20_POLY1305 IpsecCryptoAlg = 12
)

var (
	IpsecCryptoAlg_name = map[uint32]string{
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
		12: "IPSEC_API_CRYPTO_ALG_CHACHA20_POLY1305",
	}
	IpsecCryptoAlg_value = map[string]uint32{
		"IPSEC_API_CRYPTO_ALG_NONE":              0,
		"IPSEC_API_CRYPTO_ALG_AES_CBC_128":       1,
		"IPSEC_API_CRYPTO_ALG_AES_CBC_192":       2,
		"IPSEC_API_CRYPTO_ALG_AES_CBC_256":       3,
		"IPSEC_API_CRYPTO_ALG_AES_CTR_128":       4,
		"IPSEC_API_CRYPTO_ALG_AES_CTR_192":       5,
		"IPSEC_API_CRYPTO_ALG_AES_CTR_256":       6,
		"IPSEC_API_CRYPTO_ALG_AES_GCM_128":       7,
		"IPSEC_API_CRYPTO_ALG_AES_GCM_192":       8,
		"IPSEC_API_CRYPTO_ALG_AES_GCM_256":       9,
		"IPSEC_API_CRYPTO_ALG_DES_CBC":           10,
		"IPSEC_API_CRYPTO_ALG_3DES_CBC":          11,
		"IPSEC_API_CRYPTO_ALG_CHACHA20_POLY1305": 12,
	}
)

func (x IpsecCryptoAlg) String() string {
	s, ok := IpsecCryptoAlg_name[uint32(x)]
	if ok {
		return s
	}
	return "IpsecCryptoAlg(" + strconv.Itoa(int(x)) + ")"
}

// IpsecIntegAlg defines enum 'ipsec_integ_alg'.
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

var (
	IpsecIntegAlg_name = map[uint32]string{
		0: "IPSEC_API_INTEG_ALG_NONE",
		1: "IPSEC_API_INTEG_ALG_MD5_96",
		2: "IPSEC_API_INTEG_ALG_SHA1_96",
		3: "IPSEC_API_INTEG_ALG_SHA_256_96",
		4: "IPSEC_API_INTEG_ALG_SHA_256_128",
		5: "IPSEC_API_INTEG_ALG_SHA_384_192",
		6: "IPSEC_API_INTEG_ALG_SHA_512_256",
	}
	IpsecIntegAlg_value = map[string]uint32{
		"IPSEC_API_INTEG_ALG_NONE":        0,
		"IPSEC_API_INTEG_ALG_MD5_96":      1,
		"IPSEC_API_INTEG_ALG_SHA1_96":     2,
		"IPSEC_API_INTEG_ALG_SHA_256_96":  3,
		"IPSEC_API_INTEG_ALG_SHA_256_128": 4,
		"IPSEC_API_INTEG_ALG_SHA_384_192": 5,
		"IPSEC_API_INTEG_ALG_SHA_512_256": 6,
	}
)

func (x IpsecIntegAlg) String() string {
	s, ok := IpsecIntegAlg_name[uint32(x)]
	if ok {
		return s
	}
	return "IpsecIntegAlg(" + strconv.Itoa(int(x)) + ")"
}

// IpsecProto defines enum 'ipsec_proto'.
type IpsecProto uint32

const (
	IPSEC_API_PROTO_ESP IpsecProto = 50
	IPSEC_API_PROTO_AH  IpsecProto = 51
)

var (
	IpsecProto_name = map[uint32]string{
		50: "IPSEC_API_PROTO_ESP",
		51: "IPSEC_API_PROTO_AH",
	}
	IpsecProto_value = map[string]uint32{
		"IPSEC_API_PROTO_ESP": 50,
		"IPSEC_API_PROTO_AH":  51,
	}
)

func (x IpsecProto) String() string {
	s, ok := IpsecProto_name[uint32(x)]
	if ok {
		return s
	}
	return "IpsecProto(" + strconv.Itoa(int(x)) + ")"
}

// IpsecSadFlags defines enum 'ipsec_sad_flags'.
type IpsecSadFlags uint32

const (
	IPSEC_API_SAD_FLAG_NONE            IpsecSadFlags = 0
	IPSEC_API_SAD_FLAG_USE_ESN         IpsecSadFlags = 1
	IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY IpsecSadFlags = 2
	IPSEC_API_SAD_FLAG_IS_TUNNEL       IpsecSadFlags = 4
	IPSEC_API_SAD_FLAG_IS_TUNNEL_V6    IpsecSadFlags = 8
	IPSEC_API_SAD_FLAG_UDP_ENCAP       IpsecSadFlags = 16
	IPSEC_API_SAD_FLAG_IS_INBOUND      IpsecSadFlags = 64
	IPSEC_API_SAD_FLAG_ASYNC           IpsecSadFlags = 128
)

var (
	IpsecSadFlags_name = map[uint32]string{
		0:   "IPSEC_API_SAD_FLAG_NONE",
		1:   "IPSEC_API_SAD_FLAG_USE_ESN",
		2:   "IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY",
		4:   "IPSEC_API_SAD_FLAG_IS_TUNNEL",
		8:   "IPSEC_API_SAD_FLAG_IS_TUNNEL_V6",
		16:  "IPSEC_API_SAD_FLAG_UDP_ENCAP",
		64:  "IPSEC_API_SAD_FLAG_IS_INBOUND",
		128: "IPSEC_API_SAD_FLAG_ASYNC",
	}
	IpsecSadFlags_value = map[string]uint32{
		"IPSEC_API_SAD_FLAG_NONE":            0,
		"IPSEC_API_SAD_FLAG_USE_ESN":         1,
		"IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY": 2,
		"IPSEC_API_SAD_FLAG_IS_TUNNEL":       4,
		"IPSEC_API_SAD_FLAG_IS_TUNNEL_V6":    8,
		"IPSEC_API_SAD_FLAG_UDP_ENCAP":       16,
		"IPSEC_API_SAD_FLAG_IS_INBOUND":      64,
		"IPSEC_API_SAD_FLAG_ASYNC":           128,
	}
)

func (x IpsecSadFlags) String() string {
	s, ok := IpsecSadFlags_name[uint32(x)]
	if ok {
		return s
	}
	str := func(n uint32) string {
		s, ok := IpsecSadFlags_name[uint32(n)]
		if ok {
			return s
		}
		return "IpsecSadFlags(" + strconv.Itoa(int(n)) + ")"
	}
	for i := uint32(0); i <= 32; i++ {
		val := uint32(x)
		if val&(1<<i) != 0 {
			if s != "" {
				s += "|"
			}
			s += str(1 << i)
		}
	}
	if s == "" {
		return str(uint32(x))
	}
	return s
}

// IpsecSpdAction defines enum 'ipsec_spd_action'.
type IpsecSpdAction uint32

const (
	IPSEC_API_SPD_ACTION_BYPASS  IpsecSpdAction = 0
	IPSEC_API_SPD_ACTION_DISCARD IpsecSpdAction = 1
	IPSEC_API_SPD_ACTION_RESOLVE IpsecSpdAction = 2
	IPSEC_API_SPD_ACTION_PROTECT IpsecSpdAction = 3
)

var (
	IpsecSpdAction_name = map[uint32]string{
		0: "IPSEC_API_SPD_ACTION_BYPASS",
		1: "IPSEC_API_SPD_ACTION_DISCARD",
		2: "IPSEC_API_SPD_ACTION_RESOLVE",
		3: "IPSEC_API_SPD_ACTION_PROTECT",
	}
	IpsecSpdAction_value = map[string]uint32{
		"IPSEC_API_SPD_ACTION_BYPASS":  0,
		"IPSEC_API_SPD_ACTION_DISCARD": 1,
		"IPSEC_API_SPD_ACTION_RESOLVE": 2,
		"IPSEC_API_SPD_ACTION_PROTECT": 3,
	}
)

func (x IpsecSpdAction) String() string {
	s, ok := IpsecSpdAction_name[uint32(x)]
	if ok {
		return s
	}
	return "IpsecSpdAction(" + strconv.Itoa(int(x)) + ")"
}

// IpsecSadEntry defines type 'ipsec_sad_entry'.
type IpsecSadEntry struct {
	SadID              uint32           `binapi:"u32,name=sad_id" json:"sad_id,omitempty"`
	Spi                uint32           `binapi:"u32,name=spi" json:"spi,omitempty"`
	Protocol           IpsecProto       `binapi:"ipsec_proto,name=protocol" json:"protocol,omitempty"`
	CryptoAlgorithm    IpsecCryptoAlg   `binapi:"ipsec_crypto_alg,name=crypto_algorithm" json:"crypto_algorithm,omitempty"`
	CryptoKey          Key              `binapi:"key,name=crypto_key" json:"crypto_key,omitempty"`
	IntegrityAlgorithm IpsecIntegAlg    `binapi:"ipsec_integ_alg,name=integrity_algorithm" json:"integrity_algorithm,omitempty"`
	IntegrityKey       Key              `binapi:"key,name=integrity_key" json:"integrity_key,omitempty"`
	Flags              IpsecSadFlags    `binapi:"ipsec_sad_flags,name=flags" json:"flags,omitempty"`
	TunnelSrc          ip_types.Address `binapi:"address,name=tunnel_src" json:"tunnel_src,omitempty"`
	TunnelDst          ip_types.Address `binapi:"address,name=tunnel_dst" json:"tunnel_dst,omitempty"`
	TxTableID          uint32           `binapi:"u32,name=tx_table_id" json:"tx_table_id,omitempty"`
	Salt               uint32           `binapi:"u32,name=salt" json:"salt,omitempty"`
	UDPSrcPort         uint16           `binapi:"u16,name=udp_src_port,default=4500" json:"udp_src_port,omitempty"`
	UDPDstPort         uint16           `binapi:"u16,name=udp_dst_port,default=4500" json:"udp_dst_port,omitempty"`
}

// IpsecSadEntryV2 defines type 'ipsec_sad_entry_v2'.
type IpsecSadEntryV2 struct {
	SadID              uint32                             `binapi:"u32,name=sad_id" json:"sad_id,omitempty"`
	Spi                uint32                             `binapi:"u32,name=spi" json:"spi,omitempty"`
	Protocol           IpsecProto                         `binapi:"ipsec_proto,name=protocol" json:"protocol,omitempty"`
	CryptoAlgorithm    IpsecCryptoAlg                     `binapi:"ipsec_crypto_alg,name=crypto_algorithm" json:"crypto_algorithm,omitempty"`
	CryptoKey          Key                                `binapi:"key,name=crypto_key" json:"crypto_key,omitempty"`
	IntegrityAlgorithm IpsecIntegAlg                      `binapi:"ipsec_integ_alg,name=integrity_algorithm" json:"integrity_algorithm,omitempty"`
	IntegrityKey       Key                                `binapi:"key,name=integrity_key" json:"integrity_key,omitempty"`
	Flags              IpsecSadFlags                      `binapi:"ipsec_sad_flags,name=flags" json:"flags,omitempty"`
	TunnelSrc          ip_types.Address                   `binapi:"address,name=tunnel_src" json:"tunnel_src,omitempty"`
	TunnelDst          ip_types.Address                   `binapi:"address,name=tunnel_dst" json:"tunnel_dst,omitempty"`
	TunnelFlags        tunnel_types.TunnelEncapDecapFlags `binapi:"tunnel_encap_decap_flags,name=tunnel_flags" json:"tunnel_flags,omitempty"`
	Dscp               ip_types.IPDscp                    `binapi:"ip_dscp,name=dscp" json:"dscp,omitempty"`
	TxTableID          uint32                             `binapi:"u32,name=tx_table_id" json:"tx_table_id,omitempty"`
	Salt               uint32                             `binapi:"u32,name=salt" json:"salt,omitempty"`
	UDPSrcPort         uint16                             `binapi:"u16,name=udp_src_port,default=4500" json:"udp_src_port,omitempty"`
	UDPDstPort         uint16                             `binapi:"u16,name=udp_dst_port,default=4500" json:"udp_dst_port,omitempty"`
}

// IpsecSadEntryV3 defines type 'ipsec_sad_entry_v3'.
type IpsecSadEntryV3 struct {
	SadID              uint32              `binapi:"u32,name=sad_id" json:"sad_id,omitempty"`
	Spi                uint32              `binapi:"u32,name=spi" json:"spi,omitempty"`
	Protocol           IpsecProto          `binapi:"ipsec_proto,name=protocol" json:"protocol,omitempty"`
	CryptoAlgorithm    IpsecCryptoAlg      `binapi:"ipsec_crypto_alg,name=crypto_algorithm" json:"crypto_algorithm,omitempty"`
	CryptoKey          Key                 `binapi:"key,name=crypto_key" json:"crypto_key,omitempty"`
	IntegrityAlgorithm IpsecIntegAlg       `binapi:"ipsec_integ_alg,name=integrity_algorithm" json:"integrity_algorithm,omitempty"`
	IntegrityKey       Key                 `binapi:"key,name=integrity_key" json:"integrity_key,omitempty"`
	Flags              IpsecSadFlags       `binapi:"ipsec_sad_flags,name=flags" json:"flags,omitempty"`
	Tunnel             tunnel_types.Tunnel `binapi:"tunnel,name=tunnel" json:"tunnel,omitempty"`
	Salt               uint32              `binapi:"u32,name=salt" json:"salt,omitempty"`
	UDPSrcPort         uint16              `binapi:"u16,name=udp_src_port,default=4500" json:"udp_src_port,omitempty"`
	UDPDstPort         uint16              `binapi:"u16,name=udp_dst_port,default=4500" json:"udp_dst_port,omitempty"`
}

// IpsecSpdEntry defines type 'ipsec_spd_entry'.
type IpsecSpdEntry struct {
	SpdID              uint32           `binapi:"u32,name=spd_id" json:"spd_id,omitempty"`
	Priority           int32            `binapi:"i32,name=priority" json:"priority,omitempty"`
	IsOutbound         bool             `binapi:"bool,name=is_outbound" json:"is_outbound,omitempty"`
	SaID               uint32           `binapi:"u32,name=sa_id" json:"sa_id,omitempty"`
	Policy             IpsecSpdAction   `binapi:"ipsec_spd_action,name=policy" json:"policy,omitempty"`
	Protocol           uint8            `binapi:"u8,name=protocol" json:"protocol,omitempty"`
	RemoteAddressStart ip_types.Address `binapi:"address,name=remote_address_start" json:"remote_address_start,omitempty"`
	RemoteAddressStop  ip_types.Address `binapi:"address,name=remote_address_stop" json:"remote_address_stop,omitempty"`
	LocalAddressStart  ip_types.Address `binapi:"address,name=local_address_start" json:"local_address_start,omitempty"`
	LocalAddressStop   ip_types.Address `binapi:"address,name=local_address_stop" json:"local_address_stop,omitempty"`
	RemotePortStart    uint16           `binapi:"u16,name=remote_port_start" json:"remote_port_start,omitempty"`
	RemotePortStop     uint16           `binapi:"u16,name=remote_port_stop" json:"remote_port_stop,omitempty"`
	LocalPortStart     uint16           `binapi:"u16,name=local_port_start" json:"local_port_start,omitempty"`
	LocalPortStop      uint16           `binapi:"u16,name=local_port_stop" json:"local_port_stop,omitempty"`
}

// IpsecSpdEntryV2 defines type 'ipsec_spd_entry_v2'.
type IpsecSpdEntryV2 struct {
	SpdID              uint32           `binapi:"u32,name=spd_id" json:"spd_id,omitempty"`
	Priority           int32            `binapi:"i32,name=priority" json:"priority,omitempty"`
	IsOutbound         bool             `binapi:"bool,name=is_outbound" json:"is_outbound,omitempty"`
	SaID               uint32           `binapi:"u32,name=sa_id" json:"sa_id,omitempty"`
	Policy             IpsecSpdAction   `binapi:"ipsec_spd_action,name=policy" json:"policy,omitempty"`
	Protocol           uint8            `binapi:"u8,name=protocol" json:"protocol,omitempty"`
	RemoteAddressStart ip_types.Address `binapi:"address,name=remote_address_start" json:"remote_address_start,omitempty"`
	RemoteAddressStop  ip_types.Address `binapi:"address,name=remote_address_stop" json:"remote_address_stop,omitempty"`
	LocalAddressStart  ip_types.Address `binapi:"address,name=local_address_start" json:"local_address_start,omitempty"`
	LocalAddressStop   ip_types.Address `binapi:"address,name=local_address_stop" json:"local_address_stop,omitempty"`
	RemotePortStart    uint16           `binapi:"u16,name=remote_port_start" json:"remote_port_start,omitempty"`
	RemotePortStop     uint16           `binapi:"u16,name=remote_port_stop" json:"remote_port_stop,omitempty"`
	LocalPortStart     uint16           `binapi:"u16,name=local_port_start" json:"local_port_start,omitempty"`
	LocalPortStop      uint16           `binapi:"u16,name=local_port_stop" json:"local_port_stop,omitempty"`
}

// Key defines type 'key'.
type Key struct {
	Length uint8  `binapi:"u8,name=length" json:"length,omitempty"`
	Data   []byte `binapi:"u8[128],name=data" json:"data,omitempty"`
}
