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

package l2plugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/go-errors/errors"

	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"

	"github.com/ligato/vpp-agent/plugins/govppmux"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	"github.com/ligato/vpp-agent/idxvpp2"
)


// L2Plugin configures VPP bridge domains, L2 FIBs and xConnects using GoVPP.
type L2Plugin struct {
	Deps

	// GoVPP
	vppCh govppapi.Channel

	// handlers
	bdHandler  vppcalls.BridgeDomainVppAPI
	fibHandler vppcalls.FIBVppAPI
	xCHandler  vppcalls.XConnectVppAPI

	// descriptors
	// TODO

	// index maps
	bdIndex idxvpp2.NameToIndex
}

// Deps lists dependencies of the interface L2 plugin.
type Deps struct {
	infra.PluginDeps
	Scheduler   scheduler.KVScheduler
	GoVppmux    govppmux.API
	IfPlugin    ifplugin.API
	StatusCheck statuscheck.PluginStatusWriter /* optional */
}

// Config holds the vpp-plugin configuration.
type Config struct {
	Mtu              uint32   `json:"mtu"`
	StatusPublishers []string `json:"status-publishers"`
}

// Init registers L2-related descriptors.
func (p *L2Plugin) Init() error {
	var err error

	// VPP channel
	if p.vppCh, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create GoVPP API channel: %v", err)
	}

	// init BD handler
	p.bdHandler = vppcalls.NewBridgeDomainVppHandler(p.vppCh, p.IfPlugin.GetInterfaceIndex(), p.Log)

	// TODO: register BD, BDInterface descriptors, get BD indexes

	// init FIB and xConnect handlers
	p.fibHandler = vppcalls.NewFIBVppHandler(p.vppCh, p.IfPlugin.GetInterfaceIndex(), p.bdIndex, p.Log)
	p.xCHandler = vppcalls.NewXConnectVppHandler(p.vppCh, p.IfPlugin.GetInterfaceIndex(), p.Log)

	// TODO: register FIB, xConnect descriptors

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *L2Plugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}