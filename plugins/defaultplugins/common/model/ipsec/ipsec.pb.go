// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ipsec.proto

/*
Package ipsec is a generated protocol buffer package.

It is generated from these files:
	ipsec.proto

It has these top-level messages:
	TunnelInterfaces
	SecurityPolicyDatabases
	SecurityAssociations
	ResyncRequest
*/
package ipsec

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// Cryptographic algorithm for encryption
type CryptoAlgorithm int32

const (
	CryptoAlgorithm_NONE_CRYPTO CryptoAlgorithm = 0
	CryptoAlgorithm_AES_CBC_128 CryptoAlgorithm = 1
	CryptoAlgorithm_AES_CBC_192 CryptoAlgorithm = 2
	CryptoAlgorithm_AES_CBC_256 CryptoAlgorithm = 3
)

var CryptoAlgorithm_name = map[int32]string{
	0: "NONE_CRYPTO",
	1: "AES_CBC_128",
	2: "AES_CBC_192",
	3: "AES_CBC_256",
}
var CryptoAlgorithm_value = map[string]int32{
	"NONE_CRYPTO": 0,
	"AES_CBC_128": 1,
	"AES_CBC_192": 2,
	"AES_CBC_256": 3,
}

func (x CryptoAlgorithm) String() string {
	return proto.EnumName(CryptoAlgorithm_name, int32(x))
}
func (CryptoAlgorithm) EnumDescriptor() ([]byte, []int) { return fileDescriptorIpsec, []int{0} }

// Cryptographic algorithm for authentication
type IntegAlgorithm int32

const (
	IntegAlgorithm_NONE_INTEG  IntegAlgorithm = 0
	IntegAlgorithm_MD5_96      IntegAlgorithm = 1
	IntegAlgorithm_SHA1_96     IntegAlgorithm = 2
	IntegAlgorithm_SHA_256_96  IntegAlgorithm = 3
	IntegAlgorithm_SHA_256_128 IntegAlgorithm = 4
	IntegAlgorithm_SHA_384_192 IntegAlgorithm = 5
	IntegAlgorithm_SHA_512_256 IntegAlgorithm = 6
)

var IntegAlgorithm_name = map[int32]string{
	0: "NONE_INTEG",
	1: "MD5_96",
	2: "SHA1_96",
	3: "SHA_256_96",
	4: "SHA_256_128",
	5: "SHA_384_192",
	6: "SHA_512_256",
}
var IntegAlgorithm_value = map[string]int32{
	"NONE_INTEG":  0,
	"MD5_96":      1,
	"SHA1_96":     2,
	"SHA_256_96":  3,
	"SHA_256_128": 4,
	"SHA_384_192": 5,
	"SHA_512_256": 6,
}

func (x IntegAlgorithm) String() string {
	return proto.EnumName(IntegAlgorithm_name, int32(x))
}
func (IntegAlgorithm) EnumDescriptor() ([]byte, []int) { return fileDescriptorIpsec, []int{1} }

// Policy action
type SecurityPolicyDatabases_SPD_PolicyEntry_Action int32

const (
	SecurityPolicyDatabases_SPD_PolicyEntry_BYPASS  SecurityPolicyDatabases_SPD_PolicyEntry_Action = 0
	SecurityPolicyDatabases_SPD_PolicyEntry_DISCARD SecurityPolicyDatabases_SPD_PolicyEntry_Action = 1
	// RESOLVE = 2; // unused in VPP
	SecurityPolicyDatabases_SPD_PolicyEntry_PROTECT SecurityPolicyDatabases_SPD_PolicyEntry_Action = 3
)

var SecurityPolicyDatabases_SPD_PolicyEntry_Action_name = map[int32]string{
	0: "BYPASS",
	1: "DISCARD",
	3: "PROTECT",
}
var SecurityPolicyDatabases_SPD_PolicyEntry_Action_value = map[string]int32{
	"BYPASS":  0,
	"DISCARD": 1,
	"PROTECT": 3,
}

func (x SecurityPolicyDatabases_SPD_PolicyEntry_Action) String() string {
	return proto.EnumName(SecurityPolicyDatabases_SPD_PolicyEntry_Action_name, int32(x))
}
func (SecurityPolicyDatabases_SPD_PolicyEntry_Action) EnumDescriptor() ([]byte, []int) {
	return fileDescriptorIpsec, []int{1, 0, 1, 0}
}

// IPSec protocol
type SecurityAssociations_SA_IPSecProtocol int32

const (
	SecurityAssociations_SA_AH  SecurityAssociations_SA_IPSecProtocol = 0
	SecurityAssociations_SA_ESP SecurityAssociations_SA_IPSecProtocol = 1
)

var SecurityAssociations_SA_IPSecProtocol_name = map[int32]string{
	0: "AH",
	1: "ESP",
}
var SecurityAssociations_SA_IPSecProtocol_value = map[string]int32{
	"AH":  0,
	"ESP": 1,
}

func (x SecurityAssociations_SA_IPSecProtocol) String() string {
	return proto.EnumName(SecurityAssociations_SA_IPSecProtocol_name, int32(x))
}
func (SecurityAssociations_SA_IPSecProtocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptorIpsec, []int{2, 0, 0}
}

// Tunnel Interfaces
type TunnelInterfaces struct {
	Tunnels []*TunnelInterfaces_Tunnel `protobuf:"bytes,1,rep,name=tunnels" json:"tunnels,omitempty"`
}

func (m *TunnelInterfaces) Reset()                    { *m = TunnelInterfaces{} }
func (m *TunnelInterfaces) String() string            { return proto.CompactTextString(m) }
func (*TunnelInterfaces) ProtoMessage()               {}
func (*TunnelInterfaces) Descriptor() ([]byte, []int) { return fileDescriptorIpsec, []int{0} }

func (m *TunnelInterfaces) GetTunnels() []*TunnelInterfaces_Tunnel {
	if m != nil {
		return m.Tunnels
	}
	return nil
}

type TunnelInterfaces_Tunnel struct {
	Name            string          `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Esn             bool            `protobuf:"varint,2,opt,name=esn,proto3" json:"esn,omitempty"`
	AntiReplay      bool            `protobuf:"varint,3,opt,name=anti_replay,json=antiReplay,proto3" json:"anti_replay,omitempty"`
	LocalIp         string          `protobuf:"bytes,4,opt,name=local_ip,json=localIp,proto3" json:"local_ip,omitempty"`
	RemoteIp        string          `protobuf:"bytes,5,opt,name=remote_ip,json=remoteIp,proto3" json:"remote_ip,omitempty"`
	LocalSpi        uint32          `protobuf:"varint,6,opt,name=local_spi,json=localSpi,proto3" json:"local_spi,omitempty"`
	RemoteSpi       uint32          `protobuf:"varint,7,opt,name=remote_spi,json=remoteSpi,proto3" json:"remote_spi,omitempty"`
	CryptoAlg       CryptoAlgorithm `protobuf:"varint,8,opt,name=crypto_alg,json=cryptoAlg,proto3,enum=ipsec.CryptoAlgorithm" json:"crypto_alg,omitempty"`
	LocalCryptoKey  string          `protobuf:"bytes,9,opt,name=local_crypto_key,json=localCryptoKey,proto3" json:"local_crypto_key,omitempty"`
	RemoteCryptoKey string          `protobuf:"bytes,10,opt,name=remote_crypto_key,json=remoteCryptoKey,proto3" json:"remote_crypto_key,omitempty"`
	IntegAlg        IntegAlgorithm  `protobuf:"varint,11,opt,name=integ_alg,json=integAlg,proto3,enum=ipsec.IntegAlgorithm" json:"integ_alg,omitempty"`
	LocalIntegKey   string          `protobuf:"bytes,12,opt,name=local_integ_key,json=localIntegKey,proto3" json:"local_integ_key,omitempty"`
	RemoteIntegKey  string          `protobuf:"bytes,13,opt,name=remote_integ_key,json=remoteIntegKey,proto3" json:"remote_integ_key,omitempty"`
	// Extra fields related to interface
	Enabled     bool     `protobuf:"varint,100,opt,name=enabled,proto3" json:"enabled,omitempty"`
	IpAddresses []string `protobuf:"bytes,101,rep,name=ip_addresses,json=ipAddresses" json:"ip_addresses,omitempty"`
	Vrf         uint32   `protobuf:"varint,102,opt,name=vrf,proto3" json:"vrf,omitempty"`
}

func (m *TunnelInterfaces_Tunnel) Reset()                    { *m = TunnelInterfaces_Tunnel{} }
func (m *TunnelInterfaces_Tunnel) String() string            { return proto.CompactTextString(m) }
func (*TunnelInterfaces_Tunnel) ProtoMessage()               {}
func (*TunnelInterfaces_Tunnel) Descriptor() ([]byte, []int) { return fileDescriptorIpsec, []int{0, 0} }

func (m *TunnelInterfaces_Tunnel) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *TunnelInterfaces_Tunnel) GetEsn() bool {
	if m != nil {
		return m.Esn
	}
	return false
}

func (m *TunnelInterfaces_Tunnel) GetAntiReplay() bool {
	if m != nil {
		return m.AntiReplay
	}
	return false
}

func (m *TunnelInterfaces_Tunnel) GetLocalIp() string {
	if m != nil {
		return m.LocalIp
	}
	return ""
}

func (m *TunnelInterfaces_Tunnel) GetRemoteIp() string {
	if m != nil {
		return m.RemoteIp
	}
	return ""
}

func (m *TunnelInterfaces_Tunnel) GetLocalSpi() uint32 {
	if m != nil {
		return m.LocalSpi
	}
	return 0
}

func (m *TunnelInterfaces_Tunnel) GetRemoteSpi() uint32 {
	if m != nil {
		return m.RemoteSpi
	}
	return 0
}

func (m *TunnelInterfaces_Tunnel) GetCryptoAlg() CryptoAlgorithm {
	if m != nil {
		return m.CryptoAlg
	}
	return CryptoAlgorithm_NONE_CRYPTO
}

func (m *TunnelInterfaces_Tunnel) GetLocalCryptoKey() string {
	if m != nil {
		return m.LocalCryptoKey
	}
	return ""
}

func (m *TunnelInterfaces_Tunnel) GetRemoteCryptoKey() string {
	if m != nil {
		return m.RemoteCryptoKey
	}
	return ""
}

func (m *TunnelInterfaces_Tunnel) GetIntegAlg() IntegAlgorithm {
	if m != nil {
		return m.IntegAlg
	}
	return IntegAlgorithm_NONE_INTEG
}

func (m *TunnelInterfaces_Tunnel) GetLocalIntegKey() string {
	if m != nil {
		return m.LocalIntegKey
	}
	return ""
}

func (m *TunnelInterfaces_Tunnel) GetRemoteIntegKey() string {
	if m != nil {
		return m.RemoteIntegKey
	}
	return ""
}

func (m *TunnelInterfaces_Tunnel) GetEnabled() bool {
	if m != nil {
		return m.Enabled
	}
	return false
}

func (m *TunnelInterfaces_Tunnel) GetIpAddresses() []string {
	if m != nil {
		return m.IpAddresses
	}
	return nil
}

func (m *TunnelInterfaces_Tunnel) GetVrf() uint32 {
	if m != nil {
		return m.Vrf
	}
	return 0
}

// Security Policy Database (SPD)
type SecurityPolicyDatabases struct {
	Spds []*SecurityPolicyDatabases_SPD `protobuf:"bytes,1,rep,name=spds" json:"spds,omitempty"`
}

func (m *SecurityPolicyDatabases) Reset()                    { *m = SecurityPolicyDatabases{} }
func (m *SecurityPolicyDatabases) String() string            { return proto.CompactTextString(m) }
func (*SecurityPolicyDatabases) ProtoMessage()               {}
func (*SecurityPolicyDatabases) Descriptor() ([]byte, []int) { return fileDescriptorIpsec, []int{1} }

func (m *SecurityPolicyDatabases) GetSpds() []*SecurityPolicyDatabases_SPD {
	if m != nil {
		return m.Spds
	}
	return nil
}

type SecurityPolicyDatabases_SPD struct {
	Name          string                                     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Interfaces    []*SecurityPolicyDatabases_SPD_Interface   `protobuf:"bytes,2,rep,name=interfaces" json:"interfaces,omitempty"`
	PolicyEntries []*SecurityPolicyDatabases_SPD_PolicyEntry `protobuf:"bytes,3,rep,name=policy_entries,json=policyEntries" json:"policy_entries,omitempty"`
}

func (m *SecurityPolicyDatabases_SPD) Reset()         { *m = SecurityPolicyDatabases_SPD{} }
func (m *SecurityPolicyDatabases_SPD) String() string { return proto.CompactTextString(m) }
func (*SecurityPolicyDatabases_SPD) ProtoMessage()    {}
func (*SecurityPolicyDatabases_SPD) Descriptor() ([]byte, []int) {
	return fileDescriptorIpsec, []int{1, 0}
}

func (m *SecurityPolicyDatabases_SPD) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *SecurityPolicyDatabases_SPD) GetInterfaces() []*SecurityPolicyDatabases_SPD_Interface {
	if m != nil {
		return m.Interfaces
	}
	return nil
}

func (m *SecurityPolicyDatabases_SPD) GetPolicyEntries() []*SecurityPolicyDatabases_SPD_PolicyEntry {
	if m != nil {
		return m.PolicyEntries
	}
	return nil
}

// Interface
type SecurityPolicyDatabases_SPD_Interface struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *SecurityPolicyDatabases_SPD_Interface) Reset()         { *m = SecurityPolicyDatabases_SPD_Interface{} }
func (m *SecurityPolicyDatabases_SPD_Interface) String() string { return proto.CompactTextString(m) }
func (*SecurityPolicyDatabases_SPD_Interface) ProtoMessage()    {}
func (*SecurityPolicyDatabases_SPD_Interface) Descriptor() ([]byte, []int) {
	return fileDescriptorIpsec, []int{1, 0, 0}
}

func (m *SecurityPolicyDatabases_SPD_Interface) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

// Policy Entry
type SecurityPolicyDatabases_SPD_PolicyEntry struct {
	Sa              string                                         `protobuf:"bytes,1,opt,name=sa,proto3" json:"sa,omitempty"`
	Priority        int32                                          `protobuf:"varint,2,opt,name=priority,proto3" json:"priority,omitempty"`
	IsOutbound      bool                                           `protobuf:"varint,3,opt,name=is_outbound,json=isOutbound,proto3" json:"is_outbound,omitempty"`
	RemoteAddrStart string                                         `protobuf:"bytes,4,opt,name=remote_addr_start,json=remoteAddrStart,proto3" json:"remote_addr_start,omitempty"`
	RemoteAddrStop  string                                         `protobuf:"bytes,5,opt,name=remote_addr_stop,json=remoteAddrStop,proto3" json:"remote_addr_stop,omitempty"`
	LocalAddrStart  string                                         `protobuf:"bytes,6,opt,name=local_addr_start,json=localAddrStart,proto3" json:"local_addr_start,omitempty"`
	LocalAddrStop   string                                         `protobuf:"bytes,7,opt,name=local_addr_stop,json=localAddrStop,proto3" json:"local_addr_stop,omitempty"`
	Protocol        uint32                                         `protobuf:"varint,8,opt,name=protocol,proto3" json:"protocol,omitempty"`
	RemotePortStart uint32                                         `protobuf:"varint,9,opt,name=remote_port_start,json=remotePortStart,proto3" json:"remote_port_start,omitempty"`
	RemotePortStop  uint32                                         `protobuf:"varint,10,opt,name=remote_port_stop,json=remotePortStop,proto3" json:"remote_port_stop,omitempty"`
	LocalPortStart  uint32                                         `protobuf:"varint,11,opt,name=local_port_start,json=localPortStart,proto3" json:"local_port_start,omitempty"`
	LocalPortStop   uint32                                         `protobuf:"varint,12,opt,name=local_port_stop,json=localPortStop,proto3" json:"local_port_stop,omitempty"`
	Action          SecurityPolicyDatabases_SPD_PolicyEntry_Action `protobuf:"varint,13,opt,name=action,proto3,enum=ipsec.SecurityPolicyDatabases_SPD_PolicyEntry_Action" json:"action,omitempty"`
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) Reset() {
	*m = SecurityPolicyDatabases_SPD_PolicyEntry{}
}
func (m *SecurityPolicyDatabases_SPD_PolicyEntry) String() string { return proto.CompactTextString(m) }
func (*SecurityPolicyDatabases_SPD_PolicyEntry) ProtoMessage()    {}
func (*SecurityPolicyDatabases_SPD_PolicyEntry) Descriptor() ([]byte, []int) {
	return fileDescriptorIpsec, []int{1, 0, 1}
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetSa() string {
	if m != nil {
		return m.Sa
	}
	return ""
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetPriority() int32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetIsOutbound() bool {
	if m != nil {
		return m.IsOutbound
	}
	return false
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetRemoteAddrStart() string {
	if m != nil {
		return m.RemoteAddrStart
	}
	return ""
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetRemoteAddrStop() string {
	if m != nil {
		return m.RemoteAddrStop
	}
	return ""
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetLocalAddrStart() string {
	if m != nil {
		return m.LocalAddrStart
	}
	return ""
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetLocalAddrStop() string {
	if m != nil {
		return m.LocalAddrStop
	}
	return ""
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetProtocol() uint32 {
	if m != nil {
		return m.Protocol
	}
	return 0
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetRemotePortStart() uint32 {
	if m != nil {
		return m.RemotePortStart
	}
	return 0
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetRemotePortStop() uint32 {
	if m != nil {
		return m.RemotePortStop
	}
	return 0
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetLocalPortStart() uint32 {
	if m != nil {
		return m.LocalPortStart
	}
	return 0
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetLocalPortStop() uint32 {
	if m != nil {
		return m.LocalPortStop
	}
	return 0
}

func (m *SecurityPolicyDatabases_SPD_PolicyEntry) GetAction() SecurityPolicyDatabases_SPD_PolicyEntry_Action {
	if m != nil {
		return m.Action
	}
	return SecurityPolicyDatabases_SPD_PolicyEntry_BYPASS
}

// Security Association (SA)
type SecurityAssociations struct {
	Sas []*SecurityAssociations_SA `protobuf:"bytes,1,rep,name=sas" json:"sas,omitempty"`
}

func (m *SecurityAssociations) Reset()                    { *m = SecurityAssociations{} }
func (m *SecurityAssociations) String() string            { return proto.CompactTextString(m) }
func (*SecurityAssociations) ProtoMessage()               {}
func (*SecurityAssociations) Descriptor() ([]byte, []int) { return fileDescriptorIpsec, []int{2} }

func (m *SecurityAssociations) GetSas() []*SecurityAssociations_SA {
	if m != nil {
		return m.Sas
	}
	return nil
}

type SecurityAssociations_SA struct {
	Name          string                                `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Spi           uint32                                `protobuf:"varint,2,opt,name=spi,proto3" json:"spi,omitempty"`
	Protocol      SecurityAssociations_SA_IPSecProtocol `protobuf:"varint,3,opt,name=protocol,proto3,enum=ipsec.SecurityAssociations_SA_IPSecProtocol" json:"protocol,omitempty"`
	CryptoAlg     CryptoAlgorithm                       `protobuf:"varint,4,opt,name=crypto_alg,json=cryptoAlg,proto3,enum=ipsec.CryptoAlgorithm" json:"crypto_alg,omitempty"`
	CryptoKey     string                                `protobuf:"bytes,5,opt,name=crypto_key,json=cryptoKey,proto3" json:"crypto_key,omitempty"`
	IntegAlg      IntegAlgorithm                        `protobuf:"varint,6,opt,name=integ_alg,json=integAlg,proto3,enum=ipsec.IntegAlgorithm" json:"integ_alg,omitempty"`
	IntegKey      string                                `protobuf:"bytes,7,opt,name=integ_key,json=integKey,proto3" json:"integ_key,omitempty"`
	UseEsn        bool                                  `protobuf:"varint,8,opt,name=use_esn,json=useEsn,proto3" json:"use_esn,omitempty"`
	UseAntiReplay bool                                  `protobuf:"varint,9,opt,name=use_anti_replay,json=useAntiReplay,proto3" json:"use_anti_replay,omitempty"`
	TunnelSrcAddr string                                `protobuf:"bytes,10,opt,name=tunnel_src_addr,json=tunnelSrcAddr,proto3" json:"tunnel_src_addr,omitempty"`
	TunnelDstAddr string                                `protobuf:"bytes,11,opt,name=tunnel_dst_addr,json=tunnelDstAddr,proto3" json:"tunnel_dst_addr,omitempty"`
}

func (m *SecurityAssociations_SA) Reset()                    { *m = SecurityAssociations_SA{} }
func (m *SecurityAssociations_SA) String() string            { return proto.CompactTextString(m) }
func (*SecurityAssociations_SA) ProtoMessage()               {}
func (*SecurityAssociations_SA) Descriptor() ([]byte, []int) { return fileDescriptorIpsec, []int{2, 0} }

func (m *SecurityAssociations_SA) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *SecurityAssociations_SA) GetSpi() uint32 {
	if m != nil {
		return m.Spi
	}
	return 0
}

func (m *SecurityAssociations_SA) GetProtocol() SecurityAssociations_SA_IPSecProtocol {
	if m != nil {
		return m.Protocol
	}
	return SecurityAssociations_SA_AH
}

func (m *SecurityAssociations_SA) GetCryptoAlg() CryptoAlgorithm {
	if m != nil {
		return m.CryptoAlg
	}
	return CryptoAlgorithm_NONE_CRYPTO
}

func (m *SecurityAssociations_SA) GetCryptoKey() string {
	if m != nil {
		return m.CryptoKey
	}
	return ""
}

func (m *SecurityAssociations_SA) GetIntegAlg() IntegAlgorithm {
	if m != nil {
		return m.IntegAlg
	}
	return IntegAlgorithm_NONE_INTEG
}

func (m *SecurityAssociations_SA) GetIntegKey() string {
	if m != nil {
		return m.IntegKey
	}
	return ""
}

func (m *SecurityAssociations_SA) GetUseEsn() bool {
	if m != nil {
		return m.UseEsn
	}
	return false
}

func (m *SecurityAssociations_SA) GetUseAntiReplay() bool {
	if m != nil {
		return m.UseAntiReplay
	}
	return false
}

func (m *SecurityAssociations_SA) GetTunnelSrcAddr() string {
	if m != nil {
		return m.TunnelSrcAddr
	}
	return ""
}

func (m *SecurityAssociations_SA) GetTunnelDstAddr() string {
	if m != nil {
		return m.TunnelDstAddr
	}
	return ""
}

type ResyncRequest struct {
	Tunnels []*TunnelInterfaces_Tunnel     `protobuf:"bytes,1,rep,name=tunnels" json:"tunnels,omitempty"`
	Spds    []*SecurityPolicyDatabases_SPD `protobuf:"bytes,2,rep,name=spds" json:"spds,omitempty"`
	Sas     []*SecurityAssociations_SA     `protobuf:"bytes,3,rep,name=sas" json:"sas,omitempty"`
}

func (m *ResyncRequest) Reset()                    { *m = ResyncRequest{} }
func (m *ResyncRequest) String() string            { return proto.CompactTextString(m) }
func (*ResyncRequest) ProtoMessage()               {}
func (*ResyncRequest) Descriptor() ([]byte, []int) { return fileDescriptorIpsec, []int{3} }

func (m *ResyncRequest) GetTunnels() []*TunnelInterfaces_Tunnel {
	if m != nil {
		return m.Tunnels
	}
	return nil
}

func (m *ResyncRequest) GetSpds() []*SecurityPolicyDatabases_SPD {
	if m != nil {
		return m.Spds
	}
	return nil
}

func (m *ResyncRequest) GetSas() []*SecurityAssociations_SA {
	if m != nil {
		return m.Sas
	}
	return nil
}

func init() {
	proto.RegisterType((*TunnelInterfaces)(nil), "ipsec.TunnelInterfaces")
	proto.RegisterType((*TunnelInterfaces_Tunnel)(nil), "ipsec.TunnelInterfaces.Tunnel")
	proto.RegisterType((*SecurityPolicyDatabases)(nil), "ipsec.SecurityPolicyDatabases")
	proto.RegisterType((*SecurityPolicyDatabases_SPD)(nil), "ipsec.SecurityPolicyDatabases.SPD")
	proto.RegisterType((*SecurityPolicyDatabases_SPD_Interface)(nil), "ipsec.SecurityPolicyDatabases.SPD.Interface")
	proto.RegisterType((*SecurityPolicyDatabases_SPD_PolicyEntry)(nil), "ipsec.SecurityPolicyDatabases.SPD.PolicyEntry")
	proto.RegisterType((*SecurityAssociations)(nil), "ipsec.SecurityAssociations")
	proto.RegisterType((*SecurityAssociations_SA)(nil), "ipsec.SecurityAssociations.SA")
	proto.RegisterType((*ResyncRequest)(nil), "ipsec.ResyncRequest")
	proto.RegisterEnum("ipsec.CryptoAlgorithm", CryptoAlgorithm_name, CryptoAlgorithm_value)
	proto.RegisterEnum("ipsec.IntegAlgorithm", IntegAlgorithm_name, IntegAlgorithm_value)
	proto.RegisterEnum("ipsec.SecurityPolicyDatabases_SPD_PolicyEntry_Action", SecurityPolicyDatabases_SPD_PolicyEntry_Action_name, SecurityPolicyDatabases_SPD_PolicyEntry_Action_value)
	proto.RegisterEnum("ipsec.SecurityAssociations_SA_IPSecProtocol", SecurityAssociations_SA_IPSecProtocol_name, SecurityAssociations_SA_IPSecProtocol_value)
}

func init() { proto.RegisterFile("ipsec.proto", fileDescriptorIpsec) }

var fileDescriptorIpsec = []byte{
	// 1038 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x56, 0xdd, 0x6e, 0xe3, 0x44,
	0x14, 0x6e, 0xe2, 0xd4, 0x49, 0x4e, 0xd6, 0xa9, 0x19, 0x01, 0x6b, 0xb2, 0x5a, 0x36, 0xe4, 0x62,
	0x15, 0x55, 0xab, 0x88, 0x66, 0x69, 0xd5, 0xbd, 0xf4, 0x26, 0x11, 0xad, 0x60, 0x5b, 0xcb, 0xee,
	0x5e, 0xec, 0x95, 0xe5, 0x3a, 0xd3, 0x32, 0x22, 0xf5, 0x18, 0xcf, 0x04, 0x29, 0x12, 0x2f, 0xc2,
	0x03, 0x20, 0x1e, 0x80, 0x67, 0xe0, 0x96, 0xa7, 0xe1, 0x8e, 0x1b, 0x34, 0x67, 0xc6, 0x8e, 0x5b,
	0xba, 0xd0, 0xe5, 0x2e, 0xe7, 0x3b, 0xdf, 0xf9, 0x66, 0xce, 0x9c, 0x1f, 0x07, 0x7a, 0x2c, 0x17,
	0x34, 0x9d, 0xe4, 0x05, 0x97, 0x9c, 0xec, 0xa2, 0x31, 0xfa, 0xb3, 0x05, 0xee, 0xc5, 0x3a, 0xcb,
	0xe8, 0xea, 0x34, 0x93, 0xb4, 0xb8, 0x4a, 0x52, 0x2a, 0xc8, 0x31, 0xb4, 0x25, 0x62, 0xc2, 0x6b,
	0x0c, 0xad, 0x71, 0x6f, 0xfa, 0xf9, 0x44, 0x87, 0xde, 0x65, 0x1a, 0x20, 0x2c, 0xe9, 0x83, 0x5f,
	0x5a, 0x60, 0x6b, 0x8c, 0x10, 0x68, 0x65, 0xc9, 0x0d, 0xf5, 0x1a, 0xc3, 0xc6, 0xb8, 0x1b, 0xe2,
	0x6f, 0xe2, 0x82, 0x45, 0x45, 0xe6, 0x35, 0x87, 0x8d, 0x71, 0x27, 0x54, 0x3f, 0xc9, 0x33, 0xe8,
	0x25, 0x99, 0x64, 0x71, 0x41, 0xf3, 0x55, 0xb2, 0xf1, 0x2c, 0xf4, 0x80, 0x82, 0x42, 0x44, 0xc8,
	0x67, 0xd0, 0x59, 0xf1, 0x34, 0x59, 0xc5, 0x2c, 0xf7, 0x5a, 0x28, 0xd5, 0x46, 0xfb, 0x34, 0x27,
	0x4f, 0xa0, 0x5b, 0xd0, 0x1b, 0x2e, 0xa9, 0xf2, 0xed, 0xa2, 0xaf, 0xa3, 0x01, 0xed, 0xd4, 0x71,
	0x22, 0x67, 0x9e, 0x3d, 0x6c, 0x8c, 0x9d, 0x50, 0x0b, 0x45, 0x39, 0x23, 0x4f, 0x01, 0x4c, 0xa4,
	0xf2, 0xb6, 0xd1, 0x6b, 0xb4, 0x94, 0xfb, 0x10, 0x20, 0x2d, 0x36, 0xb9, 0xe4, 0x71, 0xb2, 0xba,
	0xf6, 0x3a, 0xc3, 0xc6, 0xb8, 0x3f, 0xfd, 0xd4, 0x3c, 0xc1, 0x0c, 0x1d, 0xfe, 0xea, 0x9a, 0x17,
	0x4c, 0x7e, 0x77, 0x13, 0x76, 0xd3, 0x12, 0x20, 0x63, 0x70, 0xf5, 0x91, 0x26, 0xf8, 0x7b, 0xba,
	0xf1, 0xba, 0x78, 0xad, 0x3e, 0xe2, 0x3a, 0xf4, 0x1b, 0xba, 0x21, 0xfb, 0xf0, 0x91, 0x39, 0xbf,
	0x46, 0x05, 0xa4, 0xee, 0x69, 0xc7, 0x96, 0x3b, 0x85, 0x2e, 0xcb, 0x24, 0xbd, 0xc6, 0xbb, 0xf4,
	0xf0, 0x2e, 0x9f, 0x98, 0xbb, 0xa8, 0x42, 0x5c, 0x6f, 0xaf, 0xd2, 0x61, 0xc6, 0x26, 0xcf, 0x61,
	0xcf, 0x3c, 0x1a, 0x46, 0x2a, 0xf5, 0x47, 0xa8, 0xee, 0xe8, 0xb7, 0x53, 0xa8, 0xd2, 0x1e, 0x83,
	0x5b, 0xbe, 0x60, 0x45, 0x74, 0xf4, 0x8d, 0xcd, 0x43, 0x96, 0x4c, 0x0f, 0xda, 0x34, 0x4b, 0x2e,
	0x57, 0x74, 0xe9, 0x2d, 0xb1, 0x46, 0xa5, 0x49, 0xbe, 0x80, 0x47, 0x2c, 0x8f, 0x93, 0xe5, 0xb2,
	0xa0, 0x42, 0x50, 0xe1, 0xd1, 0xa1, 0x35, 0xee, 0x86, 0x3d, 0x96, 0xfb, 0x25, 0xa4, 0xca, 0xfe,
	0x63, 0x71, 0xe5, 0x5d, 0xe1, 0x3b, 0xab, 0x9f, 0xa3, 0xbf, 0x6c, 0x78, 0x1c, 0xd1, 0x74, 0x5d,
	0x30, 0xb9, 0x09, 0xf8, 0x8a, 0xa5, 0x9b, 0x79, 0x22, 0x93, 0xcb, 0x44, 0xb1, 0x8f, 0xa0, 0x25,
	0xf2, 0x65, 0xd9, 0x7a, 0x23, 0x93, 0xeb, 0x7b, 0xd8, 0x93, 0x28, 0x98, 0x87, 0xc8, 0x1f, 0xfc,
	0x6a, 0x83, 0x15, 0x05, 0xf3, 0x7b, 0x1b, 0xef, 0x5b, 0x00, 0x56, 0x75, 0xad, 0xd7, 0x44, 0xe5,
	0x17, 0xff, 0xad, 0x3c, 0xa9, 0x5a, 0x3d, 0xac, 0xc5, 0x93, 0xb7, 0xd0, 0xcf, 0x91, 0x1c, 0xd3,
	0x4c, 0x16, 0x8c, 0x0a, 0xcf, 0x42, 0xc5, 0xc9, 0x03, 0x14, 0x35, 0xb6, 0xc8, 0x64, 0xb1, 0x09,
	0x9d, 0xbc, 0x32, 0x18, 0x15, 0x83, 0x67, 0xd0, 0xad, 0xce, 0xbb, 0x2f, 0x8b, 0xc1, 0xef, 0x2d,
	0xe8, 0xd5, 0xe2, 0x49, 0x1f, 0x9a, 0x22, 0x31, 0x8c, 0xa6, 0x48, 0xc8, 0x00, 0x3a, 0x79, 0xc1,
	0x54, 0x37, 0x6c, 0x70, 0xc6, 0x76, 0xc3, 0xca, 0x56, 0x83, 0xc6, 0x44, 0xcc, 0xd7, 0xf2, 0x92,
	0xaf, 0xb3, 0x65, 0x39, 0x68, 0x4c, 0x9c, 0x1b, 0xa4, 0xd6, 0x93, 0xaa, 0x96, 0xb1, 0x90, 0x49,
	0x21, 0xcd, 0xc4, 0x99, 0x9e, 0x54, 0x05, 0x8d, 0x14, 0x5c, 0xeb, 0x1b, 0xc3, 0xe5, 0xe5, 0x00,
	0xf6, 0xeb, 0x54, 0x9e, 0x6f, 0x67, 0xa2, 0x26, 0x6a, 0xd7, 0x66, 0x62, 0xab, 0x59, 0xf5, 0xec,
	0x56, 0xb2, 0x5d, 0xeb, 0xd9, 0x4a, 0x11, 0x93, 0xe4, 0x92, 0xa7, 0x7c, 0x85, 0xa3, 0xe9, 0x84,
	0x95, 0x5d, 0xcb, 0x21, 0xe7, 0x85, 0x34, 0xc7, 0x75, 0x91, 0x64, 0x72, 0x08, 0x78, 0x21, 0xef,
	0xe6, 0x60, 0xb8, 0x3c, 0xc7, 0x11, 0x74, 0xca, 0x1c, 0x34, 0xb5, 0x9e, 0x43, 0x4d, 0xb4, 0xa7,
	0x99, 0x88, 0x6f, 0x35, 0xab, 0x1c, 0xb6, 0x92, 0x8f, 0x90, 0xe8, 0xd4, 0x88, 0x3c, 0x27, 0x6f,
	0xc0, 0x4e, 0x52, 0xc9, 0x78, 0x86, 0xd3, 0xd6, 0x9f, 0x1e, 0x7e, 0x58, 0xe3, 0x4c, 0x7c, 0x0c,
	0x0e, 0x8d, 0xc8, 0x68, 0x02, 0xb6, 0x46, 0x08, 0x80, 0xfd, 0xfa, 0x5d, 0xe0, 0x47, 0x91, 0xbb,
	0x43, 0x7a, 0xd0, 0x9e, 0x9f, 0x46, 0x33, 0x3f, 0x9c, 0xbb, 0x0d, 0x65, 0x04, 0xe1, 0xf9, 0xc5,
	0x62, 0x76, 0xe1, 0x5a, 0xa3, 0x9f, 0x5b, 0xf0, 0x71, 0x79, 0x94, 0x2f, 0x04, 0x4f, 0x59, 0xa2,
	0xa2, 0x05, 0xf9, 0x12, 0x2c, 0x91, 0xdc, 0x5d, 0xfa, 0xf7, 0x31, 0x27, 0x91, 0x1f, 0x2a, 0xea,
	0xe0, 0x0f, 0x0b, 0x9a, 0x91, 0xff, 0xbe, 0x65, 0xaf, 0xb6, 0x6b, 0x53, 0x4f, 0xbd, 0xc8, 0x19,
	0x39, 0xa9, 0x95, 0xce, 0xc2, 0xc4, 0x5f, 0xfc, 0xfb, 0x19, 0x93, 0xd3, 0x20, 0xa2, 0x69, 0x60,
	0x62, 0x6a, 0x85, 0xbe, 0xbd, 0xa1, 0x5b, 0x0f, 0xdd, 0xd0, 0x4f, 0xab, 0x30, 0xb5, 0xe9, 0x74,
	0xc7, 0x1a, 0xf7, 0x3f, 0x56, 0xad, 0xfd, 0xb0, 0x55, 0xfb, 0xa4, 0x8c, 0x51, 0x8a, 0xba, 0x61,
	0xb5, 0x53, 0x09, 0x3e, 0x86, 0xf6, 0x5a, 0xd0, 0x58, 0x7d, 0xf3, 0x3a, 0x38, 0x70, 0xf6, 0x5a,
	0xd0, 0x85, 0xc8, 0x54, 0xa3, 0x28, 0x47, 0xfd, 0xd3, 0xd7, 0x45, 0x82, 0xb3, 0x16, 0xd4, 0xdf,
	0x7e, 0xfd, 0x9e, 0xc3, 0x9e, 0xfe, 0xb4, 0xc6, 0xa2, 0x48, 0x71, 0x32, 0xcc, 0x67, 0xc2, 0xd1,
	0x70, 0x54, 0xa4, 0x6a, 0x30, 0x6a, 0xbc, 0xa5, 0x90, 0x9a, 0xd7, 0xab, 0xf3, 0xe6, 0x42, 0x2a,
	0xde, 0x68, 0x08, 0xce, 0xad, 0x27, 0x25, 0x36, 0x34, 0xfd, 0x13, 0x77, 0x87, 0xb4, 0xc1, 0x5a,
	0x44, 0x81, 0xdb, 0x18, 0xfd, 0xd6, 0x00, 0x27, 0xa4, 0x62, 0x93, 0xa5, 0x21, 0xfd, 0x61, 0x4d,
	0x85, 0xfc, 0xff, 0xff, 0x06, 0xaa, 0x4d, 0xde, 0xfc, 0xb0, 0x4d, 0x5e, 0xb6, 0xa1, 0xf5, 0xe0,
	0x36, 0xdc, 0x7f, 0x0b, 0x7b, 0x77, 0xca, 0x4e, 0xf6, 0xa0, 0x77, 0x76, 0x7e, 0xb6, 0x88, 0x67,
	0xe1, 0xbb, 0xe0, 0xe2, 0xdc, 0xdd, 0x51, 0x80, 0xbf, 0x88, 0xe2, 0xd9, 0xeb, 0x59, 0x7c, 0x30,
	0x3d, 0x76, 0x1b, 0xb7, 0x80, 0x57, 0x53, 0xb7, 0x59, 0x07, 0xa6, 0x87, 0x47, 0xae, 0xb5, 0xff,
	0x13, 0xf4, 0x6f, 0x17, 0x9e, 0xf4, 0x01, 0x50, 0xf5, 0xf4, 0xec, 0x62, 0xf1, 0xb5, 0xbb, 0xa3,
	0x06, 0xee, 0xcd, 0xfc, 0x30, 0x7e, 0x75, 0xa4, 0x67, 0x2c, 0x3a, 0xf1, 0x0f, 0x94, 0xd1, 0x54,
	0xc4, 0xe8, 0xc4, 0x57, 0x3a, 0xca, 0xb6, 0x94, 0x76, 0x69, 0xab, 0xd3, 0x5b, 0x25, 0xf0, 0xf2,
	0xf8, 0x2b, 0x3c, 0x7d, 0xb7, 0x04, 0x0e, 0x0f, 0xa6, 0x78, 0xba, 0x7d, 0x69, 0x63, 0xbb, 0xbf,
	0xfc, 0x3b, 0x00, 0x00, 0xff, 0xff, 0xb4, 0x10, 0xd6, 0x6b, 0xb8, 0x09, 0x00, 0x00,
}
