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

//go:generate descriptor-adapter --descriptor-name NAT44Global --value-type *vpp_nat.Nat44Global --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name NAT44GlobalInterface --value-type *vpp_nat.Nat44Global_Interface --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name NAT44GlobalAddress --value-type *vpp_nat.Nat44Global_Address --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name DNAT44 --value-type *vpp_nat.DNat44 --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name NAT44Interface --value-type *vpp_nat.Nat44Interface --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name NAT44AddressPool --value-type *vpp_nat.Nat44AddressPool --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat" --output-dir "descriptor"

package natplugin

import (
	"github.com/pkg/errors"

	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/infra"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls/vpp2001"
)

// NATPlugin configures VPP NAT.
type NATPlugin struct {
	Deps

	// handlers
	natHandler vppcalls.NatVppAPI
}

// Deps lists dependencies of the NAT plugin.
type Deps struct {
	infra.PluginDeps
	KVScheduler kvs.KVScheduler
	VPP         govppmux.API
	IfPlugin    ifplugin.API
	StatusCheck statuscheck.PluginStatusWriter // optional
}

// Init registers NAT-related descriptors.
func (p *NATPlugin) Init() (err error) {
	if !p.VPP.IsPluginLoaded("nat") {
		p.Log.Warnf("VPP plugin NAT was disabled by VPP")
		return nil
	}

	// init handlers
	p.natHandler = vppcalls.CompatibleNatVppHandler(p.VPP, p.IfPlugin.GetInterfaceIndex(), p.IfPlugin.GetDHCPIndex(), p.Log)
	if p.natHandler == nil {
		return errors.New("natHandler is not available")
	}

	// init and register descriptors
	nat44GlobalCtx, nat44GlobalDescriptor := descriptor.NewNAT44GlobalDescriptor(p.natHandler, p.Log)
	nat44GlobalIfaceDescriptor := descriptor.NewNAT44GlobalInterfaceDescriptor(p.natHandler, p.Log)
	nat44GlobalAddrDescriptor := descriptor.NewNAT44GlobalAddressDescriptor(p.natHandler, p.Log)
	dnat44Descriptor := descriptor.NewDNAT44Descriptor(p.natHandler, p.Log)
	nat44IfaceDescriptor := descriptor.NewNAT44InterfaceDescriptor(nat44GlobalCtx, p.natHandler, p.Log)
	nat44AddrPoolDescriptor := descriptor.NewNAT44AddressPoolDescriptor(nat44GlobalCtx, p.natHandler, p.Log)

	err = p.KVScheduler.RegisterKVDescriptor(
		nat44GlobalDescriptor,
		nat44GlobalIfaceDescriptor, // deprecated, kept for backward compatibility
		nat44GlobalAddrDescriptor,  // deprecated, kept for backward compatibility
		dnat44Descriptor,
		nat44IfaceDescriptor,
		nat44AddrPoolDescriptor,
	)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *NATPlugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}
