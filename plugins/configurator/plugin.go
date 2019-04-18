//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package configurator

import (
	"git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/api/models/vpp"
	"github.com/ligato/vpp-agent/plugins/orchestrator"
	ipsecvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls"
	puntvppcalls "github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"

	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/rpc/grpc"

	rpc "github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	iflinuxcalls "github.com/ligato/vpp-agent/plugins/linux/ifplugin/linuxcalls"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linux/l3plugin/linuxcalls"
	aclvppcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin"
	l2vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	l3vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	natvppcalls "github.com/ligato/vpp-agent/plugins/vpp/natplugin/vppcalls"
)

// Plugin registers VPP GRPC services in *grpc.Server.
type Plugin struct {
	Deps

	configurator configuratorServer

	// Channels
	vppChan  api.Channel
	dumpChan api.Channel
}

// Deps - dependencies of Plugin
type Deps struct {
	infra.PluginDeps
	GRPCServer  grpc.Server
	Dispatch    orchestrator.Dispatcher
	GoVppmux    govppmux.StatsAPI
	VPPIfPlugin ifplugin.API
	VPPL2Plugin *l2plugin.L2Plugin
}

// Init sets plugin child loggers
func (p *Plugin) Init() error {
	p.configurator.log = p.Log.NewLogger("configurator")
	p.configurator.dumpService.log = p.Log.NewLogger("dump")
	p.configurator.notifyService.log = p.Log.NewLogger("notify")
	p.configurator.dispatch = p.Dispatch

	if err := p.initHandlers(); err != nil {
		return err
	}

	grpcServer := p.GRPCServer.GetServer()
	if grpcServer != nil {
		rpc.RegisterConfiguratorServer(grpcServer, &p.configurator)
	}

	if p.VPPIfPlugin != nil {
		p.VPPIfPlugin.SetNotifyService(func(vppNotification *vpp.Notification) {
			p.configurator.notifyService.pushNotification(&rpc.Notification{
				Notification: &rpc.Notification_VppNotification{
					VppNotification: vppNotification,
				},
			})
		})
	}

	return nil
}

// Close does nothing.
func (p *Plugin) Close() error {
	return nil
}

// helper method initializes all VPP/Linux plugin handlers
func (p *Plugin) initHandlers() (err error) {
	// VPP channels
	if p.vppChan, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}
	if p.dumpChan, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}

	// VPP Indexes
	ifIndexes := p.VPPIfPlugin.GetInterfaceIndex()
	bdIndexes := p.VPPL2Plugin.GetBDIndex()
	dhcpIndexes := p.VPPIfPlugin.GetDHCPIndex()

	// Initialize VPP handlers
	p.configurator.aclHandler = aclvppcalls.CompatibleACLVppHandler(p.vppChan, p.dumpChan, ifIndexes, p.Log)
	p.configurator.ifHandler = ifvppcalls.CompatibleInterfaceVppHandler(p.vppChan, p.Log)
	p.configurator.natHandler = natvppcalls.CompatibleNatVppHandler(p.vppChan, ifIndexes, dhcpIndexes, p.Log)
	p.configurator.l2Handler = l2vppcalls.CompatibleL2VppHandler(p.vppChan, ifIndexes, bdIndexes, p.Log)
	p.configurator.l3Handler = l3vppcalls.CompatibleL3VppHandler(p.vppChan, ifIndexes, p.Log)
	p.configurator.ipsecHandler = ipsecvppcalls.CompatibleIPSecVppHandler(p.vppChan, ifIndexes, p.Log)
	p.configurator.puntHandler = puntvppcalls.CompatiblePuntVppHandler(p.vppChan, ifIndexes, p.Log)

	// Linux indexes and handlers
	p.configurator.linuxIfHandler = iflinuxcalls.NewNetLinkHandler()
	p.configurator.linuxL3Handler = l3linuxcalls.NewNetLinkHandler()

	return nil
}
