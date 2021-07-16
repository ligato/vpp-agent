//  Copyright (c) 2020 Doc.ai and/or its affiliates.
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

//go:generate descriptor-adapter --descriptor-name Peer --value-type *vpp_wg.Peer --meta-type *wgidx.WgMetadata --import "go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/wgidx" --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/wireguard" --output-dir "descriptor"

package wireguardplugin

import (
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/vppcalls"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/vppcalls/vpp2009"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/vppcalls/vpp2101"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/vppcalls/vpp2106"
)

type WgPlugin struct {
	Deps
	// handler
	WgHandler vppcalls.WgVppAPI

	peerDescriptor *descriptor.WgPeerDescriptor
}

type Deps struct {
	infra.PluginDeps
	KVScheduler kvs.KVScheduler
	VPP         govppmux.API
	IfPlugin    ifplugin.API
	StatusCheck statuscheck.PluginStatusWriter // optional
}

func (p *WgPlugin) Init() (err error) {
	if !p.VPP.IsPluginLoaded("wireguard") {
		p.Log.Warnf("VPP plugin wireguard was disabled by VPP")
		return nil
	}

	// init Wg handler
	p.WgHandler = vppcalls.CompatibleWgVppHandler(p.VPP, p.IfPlugin.GetInterfaceIndex(), p.Log)
	if p.WgHandler == nil {
		return errors.New("Wireguard handler is not available")
	}

	p.peerDescriptor = descriptor.NewWgPeerDescriptor(p.WgHandler, p.Log)
	peerDescriptor := adapter.NewPeerDescriptor(p.peerDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(peerDescriptor)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *WgPlugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}