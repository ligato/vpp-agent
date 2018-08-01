// Copyright (c) 2017 Cisco and/or its affiliates.
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

package rest

import (
	"fmt"
	"sync"

	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/rest/resturl"
	"github.com/ligato/vpp-agent/plugins/vpp"
	aclvppcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	l2vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	l3vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
)

const (
	swIndexVarName = "swindex"
)

// REST api methods
const (
	GET = "GET"
)

// Plugin registers Rest Plugin
type Plugin struct {
	Deps

	indexItems []indexItem

	// Channels
	vppChan  api.Channel
	dumpChan api.Channel

	// Handlers
	aclHandler aclvppcalls.AclVppRead
	ifHandler  ifvppcalls.IfVppRead
	bfdHandler ifvppcalls.BfdVppRead
	bdHandler  l2vppcalls.BridgeDomainVppRead
	fibHandler l2vppcalls.FibVppRead
	xcHandler  l2vppcalls.XConnectVppRead
	rtHandler  l3vppcalls.RouteVppRead

	sync.Mutex
}

// Deps represents dependencies of Rest Plugin
type Deps struct {
	infra.PluginDeps
	HTTPHandlers rest.HTTPHandlers
	GoVppmux     govppmux.API
	VPP          vpp.API
}

type indexItem struct {
	Name string
	Path string
}

// Init initializes the Rest Plugin
func (plugin *Plugin) Init() (err error) {
	// Check VPP dependency
	if plugin.VPP == nil {
		return fmt.Errorf("REST plugin requires VPP plugin API")
	}
	// VPP channels
	if plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}
	if plugin.dumpChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}
	// Indexes
	ifIndexes := plugin.VPP.GetSwIfIndexes()
	bdIndexes := plugin.VPP.GetBDIndexes()

	// Initialize handlers
	if plugin.aclHandler, err = aclvppcalls.NewAclVppHandler(plugin.vppChan, plugin.dumpChan, nil); err != nil {
		return err
	}
	if plugin.ifHandler, err = ifvppcalls.NewIfVppHandler(plugin.vppChan, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.bfdHandler, err = ifvppcalls.NewBfdVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.bdHandler, err = l2vppcalls.NewBridgeDomainVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.fibHandler, err = l2vppcalls.NewFibVppHandler(plugin.vppChan, plugin.dumpChan, make(chan *l2vppcalls.FibLogicalReq),
		ifIndexes, bdIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.xcHandler, err = l2vppcalls.NewXConnectVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.rtHandler, err = l3vppcalls.NewRouteVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}

	plugin.indexItems = []indexItem{
		{Name: "ACL IP", Path: resturl.AclIP},
		{Name: "ACL MACIP", Path: resturl.AclMACIP},
		{Name: "Interfaces", Path: resturl.Interface},
		{Name: "Loopback interfaces", Path: resturl.Loopback},
		{Name: "Ethernet interfaces", Path: resturl.Ethernet},
		{Name: "Memif interfaces", Path: resturl.Memif},
		{Name: "Tap interfaces", Path: resturl.Tap},
		{Name: "VxLAN interfaces", Path: resturl.VxLan},
		{Name: "Af-packet nterfaces", Path: resturl.AfPacket},
		{Name: "Bridge domains", Path: resturl.Bd},
		{Name: "Bridge domain IDs", Path: resturl.BdId},
		{Name: "L2Fibs", Path: resturl.Fib},
		{Name: "XConnectorPairs", Path: resturl.Xc},
		{Name: "Static routes", Path: resturl.Routes},

		{Name: "ARPs", Path: "/arps"},
		{Name: "Telemetry", Path: "/telemetry"},
	}
	return nil
}

// AfterInit is used to register HTTP handlers
func (plugin *Plugin) AfterInit() (err error) {
	plugin.Log.Debug("REST API Plugin is up and running")

	plugin.registerAccessListHandlers()
	plugin.registerInterfaceHandlers()
	plugin.registerBfdHandlers()
	plugin.registerL2Handlers()
	plugin.registerL3Handlers()

	plugin.HTTPHandlers.RegisterHTTPHandler("/arps", plugin.arpGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler(fmt.Sprintf("/acl/interface/{%s:[0-9]+}", swIndexVarName),
		plugin.interfaceACLGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/command", plugin.commandHandler, "POST")
	plugin.HTTPHandlers.RegisterHTTPHandler("/telemetry", plugin.telemetryHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/telemetry/memory", plugin.telemetryMemoryHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/telemetry/runtime", plugin.telemetryRuntimeHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/telemetry/nodecount", plugin.telemetryNodeCountHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/", plugin.indexHandler, "GET")

	return nil
}

// Close is used to clean up resources used by Plugin
func (plugin *Plugin) Close() (err error) {
	return safeclose.Close(plugin.vppChan, plugin.dumpChan)
}
