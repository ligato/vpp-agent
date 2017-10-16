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

package restplugin

import (
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// RESTAPIPlugin - registers VPP REST API Plugin
type RESTAPIPlugin struct {
	Deps RESTAPIPluginDeps
}

// RESTAPIPluginDeps - dependencies of RESTSvcPluginDeps
type RESTAPIPluginDeps struct {
	local.PluginInfraDeps
	HTTPHandlers rest.HTTPHandlers
	GoVppmux          govppmux.API
}

// Init - initializes the RESTSvcPlugin
func (plugin *RESTAPIPlugin) Init() (err error) {
	return nil
}

// AfterInit - used to register HTTP handlers
func (plugin *RESTAPIPlugin) AfterInit() (err error) {
	plugin.Deps.Log.Info("VPP REST API Plugin is up and running !!")
	plugin.Deps.HTTPHandlers.RegisterHTTPHandler("/interfaces", plugin.interfaceGetHandler, "GET")
	return nil
}

// Close - used to clean up resources used by RESTSvcPlugin
func (plugin *RESTAPIPlugin) Close() (err error) {
	return nil
}
