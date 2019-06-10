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

	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/cn-infra/utils/safeclose"

	rpc "github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/api/models/vpp"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	iflinuxcalls "github.com/ligato/vpp-agent/plugins/linux/ifplugin/linuxcalls"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linux/l3plugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/orchestrator"
	abfvppcalls "github.com/ligato/vpp-agent/plugins/vpp/abfplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin"
	aclvppcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	ipsecvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin"
	l2vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	l3vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	natvppcalls "github.com/ligato/vpp-agent/plugins/vpp/natplugin/vppcalls"
	puntvppcalls "github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

// Plugin registers VPP GRPC services in *grpc.Server.
type Plugin struct {
	Deps

	configurator configuratorServer

	// Channels
	vppChan api.Channel
}

// Deps - dependencies of Plugin
type Deps struct {
	infra.PluginDeps
	GRPCServer   grpc.Server
	Dispatch     orchestrator.Dispatcher
	GoVppmux     govppmux.StatsAPI
	VPPACLPlugin aclplugin.API
	VPPIfPlugin  ifplugin.API
	VPPL2Plugin  *l2plugin.L2Plugin
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
	return safeclose.Close(p.vppChan)
}

// helper method initializes all VPP/Linux plugin handlers
func (p *Plugin) initHandlers() (err error) {
	// VPP channels
	if p.vppChan, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}

	// VPP Indexes
	ifIndexes := p.VPPIfPlugin.GetInterfaceIndex()
	dhcpIndexes := p.VPPIfPlugin.GetDHCPIndex()
	bdIndexes := p.VPPL2Plugin.GetBDIndex()
	aclIndexes := p.VPPACLPlugin.GetACLIndex() // TODO: make ACL optional

	// VPP handlers

	// core
	p.configurator.ifHandler = ifvppcalls.CompatibleInterfaceVppHandler(p.vppChan, p.Log)
	if p.configurator.ifHandler == nil {
		p.Log.Info("VPP Interface handler is not available, it will be skipped")
	}
	p.configurator.l2Handler = l2vppcalls.CompatibleL2VppHandler(p.vppChan, ifIndexes, bdIndexes, p.Log)
	if p.configurator.l2Handler == nil {
		p.Log.Info("VPP L2 handler is not available, it will be skipped")
	}
	p.configurator.l3Handler = l3vppcalls.CompatibleL3VppHandler(p.vppChan, ifIndexes, p.Log)
	if p.configurator.l3Handler == nil {
		p.Log.Info("VPP L3 handler is not available, it will be skipped")
	}
	p.configurator.ipsecHandler = ipsecvppcalls.CompatibleIPSecVppHandler(p.vppChan, ifIndexes, p.Log)
	if p.configurator.ipsecHandler == nil {
		p.Log.Info("VPP IPSec handler is not available, it will be skipped")
	}

	// plugins
	p.configurator.abfHandler = abfvppcalls.CompatibleABFVppHandler(p.vppChan, aclIndexes, ifIndexes, p.Log)
	if p.configurator.abfHandler == nil {
		p.Log.Info("VPP ABF handler is not available, it will be skipped")
	}
	p.configurator.aclHandler = aclvppcalls.CompatibleACLVppHandler(p.vppChan, ifIndexes, p.Log)
	if p.configurator.aclHandler == nil {
		p.Log.Info("VPP ACL handler is not available, it will be skipped")
	}
	p.configurator.natHandler = natvppcalls.CompatibleNatVppHandler(p.vppChan, ifIndexes, dhcpIndexes, p.Log)
	if p.configurator.natHandler == nil {
		p.Log.Info("VPP NAT handler is not available, it will be skipped")
	}
	p.configurator.puntHandler = puntvppcalls.CompatiblePuntVppHandler(p.vppChan, ifIndexes, p.Log)
	if p.configurator.puntHandler == nil {
		p.Log.Info("VPP Punt handler is not available, it will be skipped")
	}

	// Linux indexes and handlers
	p.configurator.linuxIfHandler = iflinuxcalls.NewNetLinkHandler()
	p.configurator.linuxL3Handler = l3linuxcalls.NewNetLinkHandler()

	return nil
}
