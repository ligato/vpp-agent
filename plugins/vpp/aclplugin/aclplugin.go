//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

//go:generate descriptor-adapter --descriptor-name ACL --value-type *vpp_acl.ACL --meta-type *aclidx.ACLMetadata --import "go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/aclidx" --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl" --output-dir "descriptor"

package aclplugin

import (
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"
	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/aclidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls/vpp2001"
)

// ACLPlugin is a plugin that manages ACLs.
type ACLPlugin struct {
	Deps

	aclHandler             vppcalls.ACLVppAPI
	aclDescriptor          *descriptor.ACLDescriptor
	aclInterfaceDescriptor *descriptor.ACLToInterfaceDescriptor

	// index maps
	aclIndex aclidx.ACLMetadataIndex
}

// Deps represents dependencies for the plugin.
type Deps struct {
	infra.PluginDeps
	Scheduler   kvs.KVScheduler
	VPP         govppmux.API
	IfPlugin    ifplugin.API
	StatusCheck statuscheck.PluginStatusWriter // optional
}

// Init initializes ACL plugin.
func (p *ACLPlugin) Init() (err error) {
	if !p.VPP.IsPluginLoaded("acl") {
		p.Log.Warnf("VPP plugin ACL was disabled by VPP")
		return nil
	}

	// init handlers
	p.aclHandler = vppcalls.CompatibleACLHandler(p.VPP, p.IfPlugin.GetInterfaceIndex())
	if p.aclHandler == nil {
		return errors.New("aclHandler is not available")
	}

	// init & register descriptors
	p.aclDescriptor = descriptor.NewACLDescriptor(p.aclHandler, p.IfPlugin, p.Log)
	aclDescriptor := adapter.NewACLDescriptor(p.aclDescriptor.GetDescriptor())
	err = p.Scheduler.RegisterKVDescriptor(aclDescriptor)
	if err != nil {
		return err
	}

	// obtain read-only references to index maps
	var withIndex bool
	metadataMap := p.Scheduler.GetMetadataMap(aclDescriptor.Name)
	p.aclIndex, withIndex = metadataMap.(aclidx.ACLMetadataIndex)
	if !withIndex {
		return errors.New("missing index with ACL metadata")
	}

	p.aclInterfaceDescriptor = descriptor.NewACLToInterfaceDescriptor(p.aclIndex, p.aclHandler, p.Log)
	aclInterfaceDescriptor := p.aclInterfaceDescriptor.GetDescriptor()
	err = p.Scheduler.RegisterKVDescriptor(aclInterfaceDescriptor)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *ACLPlugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}

// GetInterfaceIndex gives read-only access to map with metadata of all configured
// VPP interfaces.
func (p *ACLPlugin) GetACLIndex() aclidx.ACLMetadataIndex {
	return p.aclIndex
}
