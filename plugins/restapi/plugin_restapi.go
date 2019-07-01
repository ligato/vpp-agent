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

package restapi

import (
	"net/http"
	"sync"

	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/rpc/rest"
	access "github.com/ligato/cn-infra/rpc/rest/security/model/access-security"
	"github.com/ligato/cn-infra/utils/safeclose"

	"github.com/ligato/vpp-agent/plugins/govppmux"
	vpevppcalls "github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
	iflinuxcalls "github.com/ligato/vpp-agent/plugins/linux/ifplugin/linuxcalls"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linux/l3plugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/restapi/resturl"
	telemetryvppcalls "github.com/ligato/vpp-agent/plugins/telemetry/vppcalls"
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

// REST api methods
const (
	GET  = http.MethodGet
	POST = http.MethodPost
)

// Plugin registers Rest Plugin
type Plugin struct {
	Deps

	// Index page
	index *index

	// Channels
	vppChan api.Channel

	// Handlers
	vpeHandler  vpevppcalls.VpeVppAPI
	teleHandler telemetryvppcalls.TelemetryVppAPI
	// VPP Handlers
	abfHandler   abfvppcalls.ABFVppRead
	aclHandler   aclvppcalls.ACLVppRead
	ifHandler    ifvppcalls.InterfaceVppRead
	natHandler   natvppcalls.NatVppRead
	l2Handler    l2vppcalls.L2VppAPI
	l3Handler    l3vppcalls.L3VppAPI
	ipSecHandler ipsecvppcalls.IPSecVPPRead
	puntHandler  puntvppcalls.PuntVPPRead
	// Linux handlers
	linuxIfHandler iflinuxcalls.NetlinkAPIRead
	linuxL3Handler l3linuxcalls.NetlinkAPIRead

	govppmux sync.Mutex
}

// Deps represents dependencies of Rest Plugin
type Deps struct {
	infra.PluginDeps
	HTTPHandlers rest.HTTPHandlers
	GoVppmux     govppmux.StatsAPI
	VPPACLPlugin aclplugin.API
	VPPIfPlugin  ifplugin.API
	VPPL2Plugin  *l2plugin.L2Plugin
}

// index defines map of main index page entries
type index struct {
	ItemMap map[string][]indexItem
}

// indexItem is single index page entry
type indexItem struct {
	Name string
	Path string
}

// Init initializes the Rest Plugin
func (p *Plugin) Init() (err error) {
	// VPP channels
	if p.vppChan, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}

	// VPP Indexes
	ifIndexes := p.VPPIfPlugin.GetInterfaceIndex()
	bdIndexes := p.VPPL2Plugin.GetBDIndex()
	dhcpIndexes := p.VPPIfPlugin.GetDHCPIndex()
	aclIndexes := p.VPPACLPlugin.GetACLIndex() // TODO: make ACL optional

	// Initialize VPP handlers
	p.vpeHandler = vpevppcalls.CompatibleVpeHandler(p.vppChan)
	if p.vpeHandler == nil {
		p.Log.Info("VPP main handler is not available, it will be skipped")
	}
	p.teleHandler = telemetryvppcalls.CompatibleTelemetryHandler(p.vppChan, p.GoVppmux)
	if p.teleHandler == nil {
		p.Log.Info("VPP Telemetry handler is not available, it will be skipped")
	}

	// core
	p.ifHandler = ifvppcalls.CompatibleInterfaceVppHandler(p.vppChan, p.Log)
	if p.ifHandler == nil {
		p.Log.Info("VPP Interface handler is not available, it will be skipped")
	}
	p.l2Handler = l2vppcalls.CompatibleL2VppHandler(p.vppChan, ifIndexes, bdIndexes, p.Log)
	if p.l2Handler == nil {
		p.Log.Info("VPP L2 handler is not available, it will be skipped")
	}
	p.l3Handler = l3vppcalls.CompatibleL3VppHandler(p.vppChan, ifIndexes, p.Log)
	if p.l3Handler == nil {
		p.Log.Info("VPP L3 handler is not available, it will be skipped")
	}
	p.ipSecHandler = ipsecvppcalls.CompatibleIPSecVppHandler(p.vppChan, ifIndexes, p.Log)
	if p.ipSecHandler == nil {
		p.Log.Info("VPP IPSec handler is not available, it will be skipped")
	}

	// plugins (might not be available - disabled)
	p.abfHandler = abfvppcalls.CompatibleABFVppHandler(p.vppChan, aclIndexes, ifIndexes, p.Log)
	if p.abfHandler == nil {
		p.Log.Infof("ABF handler is not available, it will be skipped")
	}
	p.aclHandler = aclvppcalls.CompatibleACLVppHandler(p.vppChan, ifIndexes, p.Log)
	if p.aclHandler == nil {
		p.Log.Infof("ACL handler is not available, it will be skipped")
	}
	p.natHandler = natvppcalls.CompatibleNatVppHandler(p.vppChan, ifIndexes, dhcpIndexes, p.Log)
	if p.natHandler == nil {
		p.Log.Infof("NAT handler is not available, it will be skipped")
	}
	p.puntHandler = puntvppcalls.CompatiblePuntVppHandler(p.vppChan, ifIndexes, p.Log)
	if p.puntHandler == nil {
		p.Log.Infof("Punt handler is not available, it will be skipped")
	}

	// Linux handlers
	p.linuxIfHandler = iflinuxcalls.NewNetLinkHandler()
	p.linuxL3Handler = l3linuxcalls.NewNetLinkHandler()

	p.index = &index{
		ItemMap: getIndexPageItems(),
	}

	// Register permission groups, used if REST security is enabled
	p.HTTPHandlers.RegisterPermissionGroup(getPermissionsGroups()...)

	return nil
}

// AfterInit is used to register HTTP handlers
func (p *Plugin) AfterInit() (err error) {
	p.Log.Debug("REST API Plugin is up and running")

	// VPP handlers
	p.registerTelemetryHandlers()
	p.registerCommandHandler()

	// core
	p.registerInterfaceHandlers()
	p.registerL2Handlers()
	p.registerL3Handlers()
	p.registerIPSecHandlers()

	// plugins
	p.registerABFHandler()
	p.registerACLHandlers()
	p.registerNATHandlers()
	p.registerPuntHandlers()

	// Linux handlers
	p.registerLinuxInterfaceHandlers()
	p.registerLinuxL3Handlers()

	// Index and stats handlers
	p.registerIndexHandlers()
	p.registerStatsHandler()

	return nil
}

// Close is used to clean up resources used by Plugin
func (p *Plugin) Close() (err error) {
	return safeclose.Close(p.vppChan)
}

// Fill index item lists
func getIndexPageItems() map[string][]indexItem {
	idxMap := map[string][]indexItem{
		"ACL plugin": {
			{Name: "IP-type access lists", Path: resturl.ACLIP},
			{Name: "MACIP-type access lists", Path: resturl.ACLMACIP},
		},
		"Interface plugin": {
			{Name: "All interfaces", Path: resturl.Interface},
			{Name: "Loopbacks", Path: resturl.Loopback},
			{Name: "Ethernets", Path: resturl.Ethernet},
			{Name: "Memifs", Path: resturl.Memif},
			{Name: "Taps", Path: resturl.Tap},
			{Name: "VxLANs", Path: resturl.VxLan},
			{Name: "Af-packets", Path: resturl.AfPacket},
		},
		"L2 plugin": {
			{Name: "Bridge domains", Path: resturl.Bd},
			{Name: "L2Fibs", Path: resturl.Fib},
			{Name: "Cross connects", Path: resturl.Xc},
		},
		"L3 plugin": {
			{Name: "Routes", Path: resturl.Routes},
			{Name: "ARPs", Path: resturl.Arps},
			{Name: "Proxy ARP interfaces", Path: resturl.PArpIfs},
			{Name: "Proxy ARP ranges", Path: resturl.PArpRngs},
		},
		"Telemetry": {
			{Name: "All data", Path: resturl.Telemetry},
			{Name: "Memory", Path: resturl.TMemory},
			{Name: "Runtime", Path: resturl.TRuntime},
			{Name: "Node count", Path: resturl.TNodeCount},
		},
		"Stats": {
			{Name: "VPP Binary API", Path: resturl.Tracer},
			{Name: "Configurator Stats", Path: resturl.ConfiguratorStats},
		},
	}
	return idxMap
}

// Create permission groups (tracer, telemetry, dump - optionally add more in the future). Used only if
// REST security is enabled in plugin
func getPermissionsGroups() []*access.PermissionGroup {
	tracerPg := &access.PermissionGroup{
		Name: "stats",
		Permissions: []*access.PermissionGroup_Permissions{
			newPermission(resturl.Index, GET),
			newPermission(resturl.Tracer, GET),
			newPermission(resturl.ConfiguratorStats, GET),
		},
	}
	telemetryPg := &access.PermissionGroup{
		Name: "telemetry",
		Permissions: []*access.PermissionGroup_Permissions{
			newPermission(resturl.Index, GET),
			newPermission(resturl.Telemetry, GET),
			newPermission(resturl.TMemory, GET),
			newPermission(resturl.TRuntime, GET),
			newPermission(resturl.TNodeCount, GET),
		},
	}
	dumpPg := &access.PermissionGroup{
		Name: "dump",
		Permissions: []*access.PermissionGroup_Permissions{
			newPermission(resturl.Index, GET),
			newPermission(resturl.ABF, GET),
			newPermission(resturl.ACLIP, GET),
			newPermission(resturl.ACLMACIP, GET),
			newPermission(resturl.Interface, GET),
			newPermission(resturl.Loopback, GET),
			newPermission(resturl.Ethernet, GET),
			newPermission(resturl.Memif, GET),
			newPermission(resturl.Tap, GET),
			newPermission(resturl.VxLan, GET),
			newPermission(resturl.AfPacket, GET),
			newPermission(resturl.Bd, GET),
			newPermission(resturl.Fib, GET),
			newPermission(resturl.Xc, GET),
			newPermission(resturl.Arps, GET),
			newPermission(resturl.Routes, GET),
			newPermission(resturl.PArpIfs, GET),
			newPermission(resturl.PArpRngs, GET),
		},
	}

	return []*access.PermissionGroup{tracerPg, telemetryPg, dumpPg}
}

// Returns permission object with url and provided methods
func newPermission(url string, methods ...string) *access.PermissionGroup_Permissions {
	return &access.PermissionGroup_Permissions{
		Url:            url,
		AllowedMethods: methods,
	}
}
