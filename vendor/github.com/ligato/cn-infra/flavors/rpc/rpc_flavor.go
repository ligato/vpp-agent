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
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/health/probe"
	"github.com/ligato/cn-infra/logging/logmanager"
	"github.com/ligato/cn-infra/rpc/rest"
)

// FlavorRPC glues together multiple plugins that are useful for almost every micro-service
type FlavorRPC struct {
	*local.FlavorLocal

	HTTP rest.Plugin
	//TODO GRPC (& enable/disable using config)

	HealthRPC probe.Plugin
	LogMngRPC logmanager.Plugin

	injected bool
}

// Inject sets object references
func (f *FlavorRPC) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()

	f.HTTP.Deps.PluginLogDeps = *f.LogDeps("http")

	f.LogMngRPC.Deps.PluginLogDeps = *f.LogDeps("log-mng-rpc")
	f.LogMngRPC.LogRegistry = f.FlavorLocal.LogRegistry()
	f.LogMngRPC.HTTP = &f.HTTP

	f.HealthRPC.Deps.PluginLogDeps = *f.LogDeps("health-rpc")
	f.HealthRPC.Deps.HTTP = &f.HTTP
	f.HealthRPC.Deps.StatusCheck = &f.StatusCheck
	//TODO f.HealthRPC.Transport inject restsync

	return true
}

// Plugins combines all Plugins in flavor to the list
func (f *FlavorRPC) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
