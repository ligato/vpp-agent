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

	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp"
	aclvppcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/l2idx"
	l2vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
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

	// Indexes
	ifIndexes ifaceidx.SwIfIndex
	bdIndexes l2idx.BDIndex

	// Handlers
	aclHandler aclvppcalls.AclVppRead
	ifHandler  ifvppcalls.IfVppRead
	bdHandler  l2vppcalls.BridgeDomainVppRead
	fibHandler l2vppcalls.FibVppRead
	xcHandler  l2vppcalls.XConnectVppRead
}

// Deps represents dependencies of Rest Plugin
type Deps struct {
	local.PluginInfraDeps
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
	// VPP channels
	if plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}
	if plugin.dumpChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}
	// Indexes
	if plugin.VPP != nil {
		plugin.ifIndexes = plugin.VPP.GetSwIfIndexes()
		plugin.bdIndexes = plugin.VPP.GetBDIndexes()
	}

	// Initialize handlers
	if plugin.aclHandler, err = aclvppcalls.NewAclVppHandler(plugin.vppChan, plugin.dumpChan, nil); err != nil {
		return err
	}
	if plugin.ifHandler, err = ifvppcalls.NewIfVppHandler(plugin.vppChan, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.ifIndexes != nil {
		if plugin.bdHandler, err = l2vppcalls.NewBridgeDomainVppHandler(plugin.vppChan, plugin.ifIndexes, plugin.Log, nil); err != nil {
			return err
		}
	}
	if plugin.ifIndexes != nil && plugin.bdIndexes != nil {
		if plugin.fibHandler, err = l2vppcalls.NewFibVppHandler(plugin.vppChan, plugin.dumpChan, make(chan *l2vppcalls.FibLogicalReq),
			plugin.ifIndexes, plugin.bdIndexes, plugin.Log, nil); err != nil {
			return err
		}
	}
	if plugin.ifIndexes != nil {
		if plugin.xcHandler, err = l2vppcalls.NewXConnectVppHandler(plugin.vppChan, plugin.ifIndexes, plugin.Log, nil); err != nil {
			return err
		}
	}

	plugin.indexItems = []indexItem{
		{Name: "ACL IP", Path: acl.RestIPKey()},
		{Name: "ACL MACIP", Path: acl.RestMACIPKey()},
		{Name: "Interfaces", Path: interfaces.RestInterfaceKey()},
		{Name: "Loopback interfaces", Path: interfaces.RestLoopbackKey()},
		{Name: "Ethernet interfaces", Path: interfaces.RestEthernetKey()},
		{Name: "Memif interfaces", Path: interfaces.RestMemifKey()},
		{Name: "Tap interfaces", Path: interfaces.RestTapKey()},
		{Name: "VxLAN interfaces", Path: interfaces.RestVxLanKey()},
		{Name: "Af-packet nterfaces", Path: interfaces.RestAfPAcketKey()},
		{Name: "Bridge domains", Path: l2.RestBridgeDomainKey()},
		{Name: "Bridge domain IDs", Path: l2.RestBridgeDomainIDKey()},
		{Name: "L2Fibs", Path: l2.RestFibKey()},
		{Name: "XConnectorPairs", Path: l2.RestXConnectKey()},

		{Name: "ARPs", Path: "/arps"},
		{Name: "Static routes", Path: "/staticroutes"},

		{Name: "Telemetry", Path: "/telemetry"},
	}
	return nil
}

// AfterInit is used to register HTTP handlers
func (plugin *Plugin) AfterInit() (err error) {
	plugin.Log.Debug("REST API Plugin is up and running")

	if err := plugin.registerAccessListHandlers(); err != nil {
		return err
	}
	if err := plugin.registerInterfaceHandlers(); err != nil {
		return err
	}
	if plugin.bdHandler != nil {
		if err := plugin.registerL2Handlers(); err != nil {
			return err
		}
	}

	plugin.HTTPHandlers.RegisterHTTPHandler("/arps", plugin.arpGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/staticroutes", plugin.staticRoutesGetHandler, "GET")
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
	return nil
}
