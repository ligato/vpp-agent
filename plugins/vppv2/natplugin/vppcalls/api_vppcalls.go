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
	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"net"

	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/nat"
)

// NatVppAPI provides methods for managing VPP NAT configuration.
type NatVppAPI interface {
	NatVppWrite
	NatVppRead
}

// NatVppWrite provides write methods for VPP NAT configuration.
type NatVppWrite interface {
	// SetNat44Forwarding configures global forwarding setup for NAT44
	SetNat44Forwarding(enableFwd bool) error
	// EnableNat44Interface enables NAT feature for provided interface
	EnableNat44Interface(iface string, isInside, isOutput bool) error
	// DisableNat44Interface disables NAT feature for provided interface
	DisableNat44Interface(iface string, isInside, isOutput bool) error
	// AddNat44Address adds new NAT address into the pool.
	AddNat44Address(address net.IP, vrf uint32, twiceNat bool) error
	// DelNat44Address removes existing NAT address from the pool.
	DelNat44Address(address net.IP, vrf uint32, twiceNat bool) error
	// SetVirtualReassemblyIPv4 configures NAT virtual reassembly for IPv4 packets.
	SetVirtualReassemblyIPv4(vrCfg *nat.Nat44Global_VirtualReassembly) error
	// SetVirtualReassemblyIPv6 configures NAT virtual reassembly for IPv6 packets.
	SetVirtualReassemblyIPv6(vrCfg *nat.Nat44Global_VirtualReassembly) error
	// AddNat44IdentityMapping adds new NAT44 identity mapping
	AddNat44IdentityMapping(mapping *nat.Nat44DNat_IdentityMapping, dnatLabel string) error
	// DelNat44IdentityMapping removes NAT44 identity mapping
	DelNat44IdentityMapping(mapping *nat.Nat44DNat_IdentityMapping, dnatLabel string) error
	// AddNat44StaticMapping creates new static mapping entry.
	AddNat44StaticMapping(mapping *nat.Nat44DNat_StaticMapping, dnatLabel string) error
	// DelNat44StaticMapping removes existing static mapping entry.
	DelNat44StaticMapping(mapping *nat.Nat44DNat_StaticMapping, dnatLabel string) error
}

// NatVppRead provides read methods for VPP NAT configuration.
type NatVppRead interface {
	// Nat44GlobalConfigDump dumps global NAT config in NB format.
	Nat44GlobalConfigDump() (*nat.Nat44Global, error)
	// NAT44NatDump dumps all configured DNAT configurations ordered by label.
	Nat44DNatDump() ([]*nat.Nat44DNat, error)
}

// NatVppHandler is accessor for NAT-related vppcalls methods.
type NatVppHandler struct {
	callsChannel api.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// NewNatVppHandler creates new instance of NAT vppcalls handler.
func NewNatVppHandler(callsChan api.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger) *NatVppHandler {
	return &NatVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}
