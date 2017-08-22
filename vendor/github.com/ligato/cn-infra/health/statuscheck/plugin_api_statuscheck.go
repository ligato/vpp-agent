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

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/health/statuscheck/model/status"
)

//go:generate protoc --proto_path=model/status --gogo_out=model/status model/status/status.proto

// PluginState is a data type used to describe the current operational state of a plugin.
type PluginState string

// PluginStateProbe defines parameters of a function used for plugin state probing.
type PluginStateProbe func() (PluginState, error)

// PluginStatusWriter allows to register & write plugin status by other plugins
type PluginStatusWriter interface {
	// put registers a plugin for status change reporting.
	Register(pluginName core.PluginName, probe PluginStateProbe)

	// ReportStateChange can be used to report a change in the status of a previously registered plugin.
	ReportStateChange(pluginName core.PluginName, state PluginState, lastError error)
}

// AgentStatusReader allows to lookup agent status by other plugins
type AgentStatusReader interface {
	// GetAgentStatus return current global operational state of the agent.
	GetAgentStatus() status.AgentStatus
}
