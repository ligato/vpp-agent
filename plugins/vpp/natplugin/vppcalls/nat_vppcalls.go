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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"

	nat "github.com/ligato/vpp-agent/api/models/vpp/nat"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
)

// NatVppAPI provides methods for managing VPP NAT configuration.
type NatVppAPI interface {
	NatVppRead

	// SetNat44Forwarding configures NAT44 forwarding.
	SetNat44Forwarding(enableFwd bool) error
	// EnableNat44Interface enables NAT44 feature for provided interface
	EnableNat44Interface(iface string, isInside, isOutput bool) error
	// DisableNat44Interface disables NAT feature for provided interface
	DisableNat44Interface(iface string, isInside, isOutput bool) error
	// AddNat44Address adds new IPV4 address into the NAT pool.
	AddNat44Address(address string, vrf uint32, twiceNat bool) error
	// DelNat44Address removes existing IPv4 address from the NAT pool.
	DelNat44Address(address string, vrf uint32, twiceNat bool) error
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
	// Nat44GlobalConfigDump dumps global NAT44 config in NB format.
	Nat44GlobalConfigDump() (*nat.Nat44Global, error)
	// DNat44Dump dumps all configured DNAT-44 configurations ordered by label.
	DNat44Dump() ([]*nat.DNat44, error)
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel, ifaceidx.IfaceMetadataIndex, idxmap.NamedMapping, logging.Logger) NatVppAPI
}

func CompatibleNatVppHandler(
	ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, dhcpIdx idxmap.NamedMapping, log logging.Logger,
) NatVppAPI {
	if len(Versions) == 0 {
		// natplugin is not loaded
		return nil
	}
	for ver, h := range Versions {
		log.Debugf("checking compatibility with %s", ver)
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch, ifIdx, dhcpIdx, log)
	}
	panic("no compatible version available")
}
