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

//go:generate descriptor-adapter --descriptor-name ABF --value-type *vpp_abf.ABF --meta-type *abfidx.ABFMetadata --import "abfidx" --import "github.com/ligato/vpp-agent/api/models/vpp/abf" --output-dir "descriptor"

package abfplugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/go-errors/errors"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/abfidx"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin"

	_ "github.com/ligato/vpp-agent/plugins/vpp/abfplugin/vppcalls/vpp1901"
	_ "github.com/ligato/vpp-agent/plugins/vpp/abfplugin/vppcalls/vpp1904"
)

// ABFPlugin is a plugin that manages ACL-based forwarding.
type ABFPlugin struct {
	Deps

	// GoVPP channels
	vppCh     govppapi.Channel
	dumpVppCh govppapi.Channel

	abfHandler             vppcalls.ABFVppAPI
	abfDescriptor          *descriptor.ABFDescriptor
	abfInterfaceDescriptor *descriptor.ABFToInterfaceDescriptor

	// index maps
	abfIndex abfidx.ABFMetadataIndex
}

// Deps represents dependencies for the plugin.
type Deps struct {
	infra.PluginDeps
	Scheduler   kvs.KVScheduler
	GoVppmux    govppmux.API
	ACLPlugin   aclplugin.API
	IfPlugin    ifplugin.API
	StatusCheck statuscheck.PluginStatusWriter // optional
}

// Init initializes ABF plugin.
func (p *ABFPlugin) Init() error {
	var err error

	// GoVPP channels
	if p.vppCh, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create GoVPP API channel: %v", err)
	}
	if p.dumpVppCh, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create GoVPP API dump channel: %v", err)
	}

	// init handler
	p.abfHandler = vppcalls.CompatibleABFVppHandler(p.vppCh, p.dumpVppCh, p.ACLPlugin.GetACLIndex(), p.IfPlugin.GetInterfaceIndex(), p.Log)
	if p.abfHandler == nil {
		return errors.New("abfHandler is not available")
	}

	// init & register descriptor
	abfDescriptor := descriptor.NewABFDescriptor(p.abfHandler, p.ACLPlugin.GetACLIndex(), p.Log)
	if err := p.Deps.Scheduler.RegisterKVDescriptor(abfDescriptor); err != nil {
		return err
	}

	// obtain read-only reference to index map
	var withIndex bool
	metadataMap := p.Scheduler.GetMetadataMap(abfDescriptor.Name)
	p.abfIndex, withIndex = metadataMap.(abfidx.ABFMetadataIndex)
	if !withIndex {
		return errors.New("missing index with ABF metadata")
	}

	// init & register derived value descriptor
	abfInterfaceDescriptor := descriptor.NewABFToInterfaceDescriptor(p.abfIndex, p.abfHandler, p.IfPlugin, p.Log)
	if err := p.Deps.Scheduler.RegisterKVDescriptor(abfInterfaceDescriptor); err != nil {
		return err
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *ABFPlugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}
