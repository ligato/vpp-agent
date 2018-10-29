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

//go:generate descriptor-adapter --descriptor-name ACL --value-type *acl.Acl --meta-type *aclidx.ACLMetadata --import "aclidx" --import "../model/acl" --output-dir "descriptor"

package aclplugin

import (
	"context"
	"sync"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/infra"
	"github.com/pkg/errors"

	"github.com/ligato/vpp-agent/plugins/govppmux"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/aclidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
)

// ACLPlugin is a plugin that manages ACLs.
type ACLPlugin struct {
	Deps

	// GoVPP channels
	vppCh     govppapi.Channel
	dumpVppCh govppapi.Channel

	aclHandler             vppcalls.ACLVppAPI
	aclDescriptor          *descriptor.ACLDescriptor
	aclInterfaceDescriptor *descriptor.ACLToInterfaceDescriptor

	aclIndex aclidx.ACLMetadataIndex

	// go routine management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Deps represents dependencies for the plugin.
type Deps struct {
	infra.PluginDeps
	Scheduler scheduler.KVScheduler
	GoVppmux  govppmux.API
	IfPlugin  ifplugin.API
}

// Init initializes ACL plugin.
func (p *ACLPlugin) Init() (err error) {
	// Create plugin context, save cancel function into the plugin handle.
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// GoVPP channels
	if p.vppCh, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create GoVPP API channel: %v", err)
	}
	if p.dumpVppCh, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create GoVPP API dump channel: %v", err)
	}

	// init handlers
	p.aclHandler = vppcalls.NewACLVppHandler(p.vppCh, p.dumpVppCh, p.IfPlugin.GetInterfaceIndex())

	// init descriptors
	p.aclDescriptor = descriptor.NewACLDescriptor(p.aclHandler, p.IfPlugin, p.Log)
	aclDescriptor := adapter.NewACLDescriptor(p.aclDescriptor.GetDescriptor())

	// register descriptors
	p.Scheduler.RegisterKVDescriptor(aclDescriptor)

	// obtain read-only references to index maps
	var withIndex bool
	metadataMap := p.Scheduler.GetMetadataMap(aclDescriptor.Name)
	p.aclIndex, withIndex = metadataMap.(aclidx.ACLMetadataIndex)
	if !withIndex {
		return errors.New("missing index with acl metadata")
	}

	p.aclInterfaceDescriptor = descriptor.NewACLToInterfaceDescriptor(p.aclIndex, p.aclHandler, p.Log)
	aclInterfaceDescriptor := p.aclInterfaceDescriptor.GetDescriptor()

	p.Scheduler.RegisterKVDescriptor(aclInterfaceDescriptor)

	return nil
}

// AfterInit initializes extra things.
func (p *ACLPlugin) AfterInit() error {
	// TODO: statucheck

	return nil
}

// Close stops all go routines and frees resources.
func (p *ACLPlugin) Close() error {
	// stop publishing of state data
	p.cancel()
	p.wg.Wait()

	// TODO: close all resources
	return nil
}
