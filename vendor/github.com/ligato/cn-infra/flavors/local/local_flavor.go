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

package local

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"
)

// FlavorLocal glues together very minimal subset of cn-infra plugins
// that can be embeddable inside different project without running
// any agent specific server.
type FlavorLocal struct {
	logRegistry  logging.Registry
	ServiceLabel servicelabel.Plugin
	StatusCheck  statuscheck.Plugin

	injected bool
}

// Inject does nothing (it is here for potential later extensibility)
// Composite flavors embedding local flavor are supposed to call this
// method.
func (f *FlavorLocal) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	f.StatusCheck.Deps.Log = f.LoggerFor("status-check")
	f.StatusCheck.Deps.PluginName = core.PluginName("status-check")

	return true
}

// Plugins combines all Plugins in flavor to the list
func (f *FlavorLocal) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}

// LogRegistry for getting Logging Registry instance
// (not thread safe)
func (f *FlavorLocal) LogRegistry() logging.Registry {
	if f.logRegistry == nil {
		f.logRegistry = logrus.NewLogRegistry()
	}

	return f.logRegistry
}

// LoggerFor for getting PlugginLogger instance:
// - logger name is pre-initialized (see logging.ForPlugin)
// This method is just convenient shortcut for Flavor.Inject()
func (f *FlavorLocal) LoggerFor(pluginName string) logging.PluginLogger {
	return logging.ForPlugin(pluginName, f.LogRegistry())
}

// LogDeps for getting PlugginLofDeps instance.
// - pluginName argument value is assigned to Plugin
// - logger name is pre-initialized (see logging.ForPlugin)
// This method is just convenient shortcut for Flavor.Inject()
func (f *FlavorLocal) LogDeps(pluginName string) *PluginLogDeps {
	return &PluginLogDeps{
		logging.ForPlugin(pluginName, f.LogRegistry()),
		core.PluginName(pluginName)}

}

// InfraDeps for getting PlugginInfraDeps instance:
// - config file is preinitialized by pluginName (see config.ForPlugin method)
// This method is just convenient shortcut for Flavor.Inject()
func (f *FlavorLocal) InfraDeps(pluginName string) *PluginInfraDeps {
	return &PluginInfraDeps{
		*f.LogDeps(pluginName),
		config.ForPlugin(pluginName),
		&f.StatusCheck,
		&f.ServiceLabel}
}
