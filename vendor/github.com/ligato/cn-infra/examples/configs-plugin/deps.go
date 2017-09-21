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
	"github.com/ligato/cn-infra/flavors/local"
)

// ExampleFlavor is composition of ExamplePlugin and Local flavor.
type ExampleFlavor struct {
	*local.FlavorLocal
	ExamplePlugin
}

// Plugins combines all Plugins in the flavor into a slice.
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	// Init local flavor
	if ef.FlavorLocal == nil {
		ef.FlavorLocal = &local.FlavorLocal{}
	}
	ef.FlavorLocal.Inject()

	// Inject plugins from the Local flavor into the Example plugin.
	// This function also initialize the flag used to make the plugin config
	// file name configurable for the user.
	ef.ExamplePlugin.PluginInfraDeps = *ef.InfraDeps(PluginName, local.WithConf())

	// Return plugins in a slice.
	return core.ListPluginsInFlavor(ef)
}
