// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vppcalls

import (
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

// NatVppAPI provides methods for managing VPP NAT configuration.
type NatVppAPI interface {
	NatVppRead

	// Enable NAT44 plugin and apply the given set of options.
	EnableNAT44Plugin(opts Nat44InitOpts) error
	// DisableNAT44Plugin disables NAT44 plugin.
	DisableNAT44Plugin() error
	// SetNat44Forwarding configures NAT44 forwarding.
	SetNat44Forwarding(enableFwd bool) error
	// EnableNat44Interface enables NAT44 feature for provided interface
	EnableNat44Interface(iface string, isInside, isOutput bool) error
	// DisableNat44Interface disables NAT feature for provided interface
	DisableNat44Interface(iface string, isInside, isOutput bool) error
	// AddNat44AddressPool adds new IPV4 address pool into the NAT pools.
	AddNat44AddressPool(vrf uint32, firstIP, lastIP string, twiceNat bool) error
	// DelNat44AddressPool removes existing IPv4 address pool from the NAT pools.
	DelNat44AddressPool(vrf uint32, firstIP, lastIP string, twiceNat bool) error
	// SetVirtualReassemblyIPv4 configures NAT virtual reassembly for IPv4 packets.
	SetVirtualReassemblyIPv4(vrCfg *nat.VirtualReassembly) error
	// SetVirtualReassemblyIPv6 configures NAT virtual reassembly for IPv6 packets.
	SetVirtualReassemblyIPv6(vrCfg *nat.VirtualReassembly) error
	// AddNat44IdentityMapping adds new NAT44 identity mapping
	AddNat44IdentityMapping(mapping *nat.DNat44_IdentityMapping, dnatLabel string) error
	// DelNat44IdentityMapping removes NAT44 identity mapping
	DelNat44IdentityMapping(mapping *nat.DNat44_IdentityMapping, dnatLabel string) error
	// AddNat44StaticMapping creates new NAT44 static mapping entry.
	AddNat44StaticMapping(mapping *nat.DNat44_StaticMapping, dnatLabel string) error
	// DelNat44StaticMapping removes existing NAT44 static mapping entry.
	DelNat44StaticMapping(mapping *nat.DNat44_StaticMapping, dnatLabel string) error
}

// NatVppRead provides read methods for VPP NAT configuration.
type NatVppRead interface {
	// WithLegacyStartupConf returns true if the loaded VPP NAT plugin is still using
	// the legacy startup NAT configuration (this is the case for VPP <= 20.09).
	WithLegacyStartupConf() bool
	// DefaultNat44GlobalConfig returns default global configuration.
	DefaultNat44GlobalConfig() *nat.Nat44Global
	// Nat44GlobalConfigDump dumps global NAT44 config in NB format.
	// If dumpDeprecated is true, dumps deprecated NAT44 global config as well.
	Nat44GlobalConfigDump(dumpDeprecated bool) (*nat.Nat44Global, error)
	// DNat44Dump dumps all configured DNAT-44 configurations ordered by label.
	DNat44Dump() ([]*nat.DNat44, error)
	// Nat44InterfacesDump dumps NAT44 config of all NAT44-enabled interfaces.
	Nat44InterfacesDump() ([]*nat.Nat44Interface, error)
	// Nat44AddressPoolsDump dumps all configured NAT44 address pools.
	Nat44AddressPoolsDump() ([]*nat.Nat44AddressPool, error)
}

// Previously these options were configured for NAT44 plugin via the startup configuration file.
// As of VPP 21.01 it is possible to configure/change them in run-time (by disabling and then
// re-enabling the plugin with changed options).
// These are just some of the supported options. For full list of what VPP allows to configure,
// see nat44_plugin_enable_disable binary API.
type Nat44InitOpts struct {
	// Endpoint dependent mode uses 6-tuple: (source IP address, source port, target IP address,
	// target port, protocol, FIB table index) as session hash table key, whereas
	// in the endpoint independent mode only 4-tuple (source IP address, source port, protocol, FIB table index)
	// is used.
	EndpointDependent bool
	// Track connection (e.g. TCP states, timeout).
	// In the dynamic mode the connection tracking is essential and performed by default.
	// With StaticMappingOnly=true it is disabled and has to be turned on explicitly if needed.
	ConnectionTracking bool
	// If enabled only static translations are performed (i.e. no dynamic session entries).
	StaticMappingOnly bool
	// Policy-based packet processing and address translation.
	// Not supported in the endpoint-dependent mode.
	OutToInDPO bool
}

var handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "nat",
	HandlerAPI: (*NatVppAPI)(nil),
})

func AddNatHandlerVersion(version vpp.Version, msgs []govppapi.Message,
	h func(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, dhcpIdx idxmap.NamedMapping, log logging.Logger) NatVppAPI,
) {
	handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return ch.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			return h(c, a[0].(ifaceidx.IfaceMetadataIndex), a[1].(idxmap.NamedMapping), a[2].(logging.Logger))
		},
	})
}

func CompatibleNatVppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, dhcpIdx idxmap.NamedMapping, log logging.Logger) NatVppAPI {
	if v := handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx, dhcpIdx, log).(NatVppAPI)
	}
	return nil
}
