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

package rpc

import (
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/rpc/rest"
)

// RESTSvcPlugin - registers VPP REST Plugin
type RESTSvcPlugin struct {
	Deps RESTSvcPluginDeps
}

// RESTSvcPluginDeps - dependencies of RESTSvcPluginDeps
type RESTSvcPluginDeps struct {
	local.PluginInfraDeps
	HTTPHandlers rest.HTTPHandlers
}

// Init - initializes the RESTSvcPlugin
func (plugin *RESTSvcPlugin) Init() error {
	return nil
}

// AfterInit - used to register HTTP handlers
func (plugin *RESTSvcPlugin) AfterInit() error {
	plugin.Deps.Log.Info("VPP REST API Plugin is up and running !!")
	plugin.Deps.HTTPHandlers.RegisterHTTPHandler("/interfaces", plugin.interfaceGetHandler, "GET")
	return nil
}

// Close - used to clean up resources used by RESTSvcPlugin
func (plugin *RESTSvcPlugin) Close() error {
	return nil
}
