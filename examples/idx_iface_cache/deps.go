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

package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/vpp-agent/flavors/vpp"
)

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	Publisher             datasync.KeyProtoValWriter // injected
	Agent1                *kvdbsync.Plugin           // injected
	Agent2                *kvdbsync.Plugin           // injected
	local.PluginInfraDeps                            // injected
}

// ExampleFlavor is a set of plugins required for the datasync example.
type ExampleFlavor struct {
	// Local flavor to access to Infra (logger, service label, status check)
	*vpp.Flavor
	// Example plugin
	IdxIfaceCacheExample ExamplePlugin
	// Mark flavor as injected after Inject()
	injected bool
}

// Inject sets object references
func (ef *ExampleFlavor) Inject() (allReadyInjected bool) {
	// Every flavor should be injected only once
	if ef.injected {
		return false
	}
	ef.injected = true

	// Init local flavor
	if ef.Flavor == nil {
		ef.Flavor = &vpp.Flavor{}
	}
	ef.Flavor.Inject()

	// Inject infra + transport (publisher, watcher) to example plugin
	ef.IdxIfaceCacheExample.PluginInfraDeps = *ef.Flavor.InfraDeps("idx-iface-cache-example")
	ef.IdxIfaceCacheExample.Publisher = &ef.ETCDDataSync
	ef.IdxIfaceCacheExample.Agent1 = ef.Flavor.ETCDDataSync.OfDifferentAgent("agent1", ef)
	ef.IdxIfaceCacheExample.Agent2 = ef.Flavor.ETCDDataSync.OfDifferentAgent("agent2", ef)

	return true
}

// Plugins combines all Plugins in flavor to the list
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}
