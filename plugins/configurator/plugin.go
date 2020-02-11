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
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/rpc/grpc"
	"go.ligato.io/cn-infra/v2/servicelabel"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	iflinuxplugin "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	iflinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
	l3linuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/l3plugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	abfvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/abfplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin"
	aclvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	ifvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	ipsecvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin"
	l2vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
	l3vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	natvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	puntvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/puntplugin/vppcalls"
	rpc "go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
)

// Default Go routine count for linux configuration retrieval
const defaultGoRoutineCount = 10

// Plugin registers VPP GRPC services in *grpc.Server.
type Plugin struct {
	Deps

	configurator configuratorServer
}

// Deps - dependencies of Plugin
type Deps struct {
	infra.PluginDeps
	GRPCServer    grpc.Server
	Dispatch      orchestrator.Dispatcher
	VPP           govppmux.API
	ServiceLabel  servicelabel.ReaderAPI
	AddrAlloc     netalloc.AddressAllocator
	VPPACLPlugin  aclplugin.API
	VPPIfPlugin   ifplugin.API
	VPPL2Plugin   *l2plugin.L2Plugin
	VPPL3Plugin   l3plugin.API
	LinuxIfPlugin iflinuxplugin.API
	NsPlugin      nsplugin.API
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
		rpc.RegisterConfiguratorServiceServer(grpcServer, &p.configurator)
	}

	if p.VPPIfPlugin != nil {
		p.VPPIfPlugin.SetNotifyService(p.sendVppNotification)
	}

	return nil
}

func (p *Plugin) sendVppNotification(vppNotification *vpp.Notification) {
	p.configurator.notifyService.pushNotification(&rpc.Notification{
		Notification: &rpc.Notification_VppNotification{
			VppNotification: vppNotification,
		},
	})
}

// Close does nothing.
func (p *Plugin) Close() error {
	return nil
}

// helper method initializes all VPP/Linux plugin handlers
func (p *Plugin) initHandlers() (err error) {
	// VPP Indexes
	ifIndexes := p.VPPIfPlugin.GetInterfaceIndex()
	dhcpIndexes := p.VPPIfPlugin.GetDHCPIndex()
	bdIndexes := p.VPPL2Plugin.GetBDIndex()
	aclIndexes := p.VPPACLPlugin.GetACLIndex() // TODO: make ACL optional
	vrfIndexes := p.VPPL3Plugin.GetVRFIndex()

	// Linux Indexes
	linuxIfIndexes := p.LinuxIfPlugin.GetInterfaceIndex()

	// VPP handlers
	p.configurator.ifHandler = ifvppcalls.CompatibleInterfaceVppHandler(p.VPP, p.Log)
	if p.configurator.ifHandler == nil {
		p.Log.Info("VPP Interface handler is not available, it will be skipped")
	}
	p.configurator.l2Handler = l2vppcalls.CompatibleL2VppHandler(p.VPP, ifIndexes, bdIndexes, p.Log)
	if p.configurator.l2Handler == nil {
		p.Log.Info("VPP L2 handler is not available, it will be skipped")
	}
	p.configurator.l3Handler = l3vppcalls.CompatibleL3VppHandler(p.VPP, ifIndexes, vrfIndexes, p.AddrAlloc, p.Log)
	if p.configurator.l3Handler == nil {
		p.Log.Info("VPP L3 handler is not available, it will be skipped")
	}
	p.configurator.ipsecHandler = ipsecvppcalls.CompatibleIPSecVppHandler(p.VPP, ifIndexes, p.Log)
	if p.configurator.ipsecHandler == nil {
		p.Log.Info("VPP IPSec handler is not available, it will be skipped")
	}
	// plugins
	p.configurator.abfHandler = abfvppcalls.CompatibleABFHandler(p.VPP, aclIndexes, ifIndexes, p.Log)
	if p.configurator.abfHandler == nil {
		p.Log.Info("VPP ABF handler is not available, it will be skipped")
	}
	p.configurator.aclHandler = aclvppcalls.CompatibleACLHandler(p.VPP, ifIndexes)
	if p.configurator.aclHandler == nil {
		p.Log.Info("VPP ACL handler is not available, it will be skipped")
	}
	p.configurator.natHandler = natvppcalls.CompatibleNatVppHandler(p.VPP, ifIndexes, dhcpIndexes, p.Log)
	if p.configurator.natHandler == nil {
		p.Log.Info("VPP NAT handler is not available, it will be skipped")
	}
	p.configurator.puntHandler = puntvppcalls.CompatiblePuntVppHandler(p.VPP, ifIndexes, p.Log)
	if p.configurator.puntHandler == nil {
		p.Log.Info("VPP Punt handler is not available, it will be skipped")
	}

	// Linux handlers
	p.configurator.linuxIfHandler = iflinuxcalls.NewNetLinkHandler(p.NsPlugin, linuxIfIndexes,
		p.ServiceLabel.GetAgentPrefix(), defaultGoRoutineCount, p.Log)
	p.configurator.linuxL3Handler = l3linuxcalls.NewNetLinkHandler(p.NsPlugin, linuxIfIndexes, defaultGoRoutineCount, p.Log)

	return nil
}
