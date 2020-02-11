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

//go:generate descriptor-adapter --descriptor-name SPD  --value-type *vpp_ipsec.SecurityPolicyDatabase --meta-type *idxvpp.OnlyIndex --import "go.ligato.io/vpp-agent/v3/pkg/idxvpp" --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name SPDInterface --value-type *vpp_ipsec.SecurityPolicyDatabase_Interface --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name SPDPolicy --value-type *vpp_ipsec.SecurityPolicyDatabase_PolicyEntry --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name SA  --value-type *vpp_ipsec.SecurityAssociation --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec" --output-dir "descriptor"

package ipsecplugin

import (
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/infra"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls/vpp2001"
)

// IPSecPlugin configures VPP security policy databases and security associations using GoVPP.
type IPSecPlugin struct {
	Deps

	// handler
	ipSecHandler vppcalls.IPSecVppAPI

	// descriptors
	spdDescriptor       *descriptor.IPSecSPDDescriptor
	saDescriptor        *descriptor.IPSecSADescriptor
	spdIfDescriptor     *descriptor.SPDInterfaceDescriptor
	spdPolicyDescriptor *descriptor.SPDPolicyDescriptor
}

// Deps lists dependencies of the IPSec plugin.
type Deps struct {
	infra.PluginDeps
	KVScheduler kvs.KVScheduler
	VPP         govppmux.API
	IfPlugin    ifplugin.API
	StatusCheck statuscheck.PluginStatusWriter // optional
}

// Init registers IPSec-related descriptors.
func (p *IPSecPlugin) Init() (err error) {
	// init IPSec handler
	p.ipSecHandler = vppcalls.CompatibleIPSecVppHandler(p.VPP, p.IfPlugin.GetInterfaceIndex(), p.Log)
	if p.ipSecHandler == nil {
		return errors.New("ipsecHandler is not available")
	}

	// init and register security policy database descriptor
	p.spdDescriptor = descriptor.NewIPSecSPDDescriptor(p.ipSecHandler, p.Log)
	spdDescriptor := adapter.NewSPDDescriptor(p.spdDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(spdDescriptor)
	if err != nil {
		return err
	}

	// init and register security association descriptor
	p.saDescriptor = descriptor.NewIPSecSADescriptor(p.ipSecHandler, p.Log)
	saDescriptor := adapter.NewSADescriptor(p.saDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(saDescriptor)
	if err != nil {
		return err
	}

	// init & register other descriptors for derived types
	p.spdIfDescriptor = descriptor.NewSPDInterfaceDescriptor(p.ipSecHandler, p.Log)
	spdIfDescriptor := adapter.NewSPDInterfaceDescriptor(p.spdIfDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(spdIfDescriptor)
	if err != nil {
		return err
	}

	p.spdPolicyDescriptor = descriptor.NewSPDPolicyDescriptor(p.ipSecHandler, p.Log)
	spdPolicyDescriptor := adapter.NewSPDPolicyDescriptor(p.spdPolicyDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(spdPolicyDescriptor)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *IPSecPlugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}
