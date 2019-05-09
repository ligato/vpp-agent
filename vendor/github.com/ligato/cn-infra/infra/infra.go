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

// Package infra provides Plugin interface and related utilities.
package infra

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/logging"
)

// Plugin interface defines plugin's basic life-cycle methods.
type Plugin interface {
	// Init is called in the agent`s startup phase.
	Init() error
	// Close is called in the agent`s cleanup phase.
	Close() error
	// String returns unique name of the plugin.
	String() string
}

// PostInit interface defines an optional method for plugins with additional initialization.
type PostInit interface {
	// AfterInit is called once Init() of all plugins have returned without error.
	AfterInit() error
}

// PluginName is a part of the plugin's API.
// It's used by embedding it into Plugin to
// provide unique name of the plugin.
type PluginName string

// String returns the PluginName.
func (name PluginName) String() string {
	return string(name)
}

// SetName sets plugin name.
func (name *PluginName) SetName(n string) {
	*name = PluginName(n)
}

// PluginDeps defines common dependencies for use with plugins.
// It can easily be embedded in Deps for Plugin:
//
// type Deps struct {
//     infra.PluginDeps
//     // other dependencies
// }
type PluginDeps struct {
	PluginName
	Log logging.PluginLogger
	Cfg config.PluginConfig
}

// SetupLog sets up default instance for plugin log dep.
func (d *PluginDeps) SetupLog() {
	if d.Log == nil {
		d.Log = logging.ForPlugin(d.String())
	}
}

// Setup sets up default instances for plugin deps.
func (d *PluginDeps) Setup() {
	d.SetupLog()
	if d.Cfg == nil {
		d.Cfg = config.ForPlugin(d.String())
	}
}

// Close is an empty implementation used to avoid need for
// implementing it by plugins that do not need it.
func (d *PluginDeps) Close() error {
	return nil
}
