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

package statuscheck

import "github.com/ligato/cn-infra/core"

// PluginID uniquely identifies the plugin.
const PluginID core.PluginName = "StatusCheck"

// PluginState is a data type used to describe the current operational state of a plugin.
type PluginState string

// PluginStateProbe defines parameters of a function used for plugin state probing.
type PluginStateProbe func() (PluginState, error)

const (
	// Init state means that the initialization of the plugin is in progress.
	Init PluginState = "init"
	// OK state means that the plugin is healthy.
	OK PluginState = "ok"
	// Error state means that some error has occurred in the plugin.
	Error PluginState = "error"
)

// Register a plugin for status change reporting.
// If probe is not nil, statuscheck will periodically probe the plugin state, otherwise it is expected that
// the plugin itself will push state updates through ReportStateChange API.
func Register(pluginName core.PluginName, probe PluginStateProbe) {
	p := plugin()
	if p != nil { //This plugin is optional
		p.registerPlugin(pluginName, probe)
	}
}

// ReportStateChange can be used to report a change in the status of a previously registered plugin.
func ReportStateChange(pluginName core.PluginName, state PluginState, lastError error) {
	p := plugin()
	if p != nil { //This plugin is optional
		p.reportStateChange(pluginName, state, lastError)
	}
}
