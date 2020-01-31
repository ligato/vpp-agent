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

//go:generate descriptor-adapter --descriptor-name STN --value-type *vpp_stn.Rule --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/stn" --output-dir "descriptor"

package stnplugin

import (
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/vppcalls"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/vppcalls/vpp2001"
)

// STNPlugin configures VPP STN rules using GoVPP.
type STNPlugin struct {
	Deps

	// handlers
	stnHandler vppcalls.StnVppAPI

	// descriptors
	stnDescriptor *descriptor.STNDescriptor
}

// Deps lists dependencies of the STN plugin.
type Deps struct {
	infra.PluginDeps
	KVScheduler kvs.KVScheduler
	VPP         govppmux.API
	IfPlugin    ifplugin.API
	StatusCheck statuscheck.PluginStatusWriter // optional
}

// Init registers STN-related descriptors.
func (p *STNPlugin) Init() (err error) {
	if !p.VPP.IsPluginLoaded("stn") {
		p.Log.Warnf("VPP plugin STN was disabled by VPP")
		return nil
	}

	// init handlers
	p.stnHandler = vppcalls.CompatibleStnVppHandler(p.VPP, p.IfPlugin.GetInterfaceIndex(), p.Log)

	// init and register STN descriptor
	p.stnDescriptor = descriptor.NewSTNDescriptor(p.stnHandler, p.Log)
	stnDescriptor := adapter.NewSTNDescriptor(p.stnDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(stnDescriptor)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *STNPlugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}
