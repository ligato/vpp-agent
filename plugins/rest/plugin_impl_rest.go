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

	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

const (
	swIndexVarName = "swindex"
)

// Plugin registers Rest Plugin
type Plugin struct {
	Deps

	indexItems []indexItem
}

// Deps represents dependencies of Rest Plugin
type Deps struct {
	local.PluginInfraDeps
	HTTPHandlers rest.HTTPHandlers
	GoVppmux     govppmux.API
}

type indexItem struct {
	Name string
	Path string
}

// Init initializes the Rest Plugin
func (plugin *Plugin) Init() (err error) {
	plugin.indexItems = []indexItem{
		{Name: "Interfaces", Path: "/interfaces"},
		{Name: "Bridge domains", Path: "/bridgedomains"},
		{Name: "L2Fibs", Path: "/l2fibs"},
		{Name: "XConnectorPairs", Path: "/xconnectpairs"},
		{Name: "ARPs", Path: "/arps"},
		{Name: "Static routes", Path: "/staticroutes"},
		{Name: "ACL IP", Path: "/acl/ip"},
		{Name: "Telemetry", Path: "/telemetry"},
	}
	return nil
}

// AfterInit is used to register HTTP handlers
func (plugin *Plugin) AfterInit() (err error) {
	plugin.Log.Debug("REST API Plugin is up and running")

	plugin.HTTPHandlers.RegisterHTTPHandler("/interfaces", plugin.interfacesGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/bridgedomains", plugin.bridgeDomainsGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/bridgedomainids", plugin.bridgeDomainIdsGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/l2fibs", plugin.fibTableEntriesGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/xconnectpairs", plugin.xconnectPairsGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/arps", plugin.arpGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/staticroutes", plugin.staticRoutesGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler(fmt.Sprintf("/acl/interface/{%s:[0-9]+}", swIndexVarName),
		plugin.interfaceACLGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/acl/ip", plugin.ipACLPostHandler, "POST")
	plugin.HTTPHandlers.RegisterHTTPHandler("/acl/ip", plugin.ipACLGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/acl/ip/example", plugin.exampleIpACLGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/acl/macip", plugin.macipACLPostHandler, "POST")
	plugin.HTTPHandlers.RegisterHTTPHandler("/acl/macip", plugin.macipACLGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/acl/macip/example", plugin.exampleMacIpACLGetHandler, "GET")
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
