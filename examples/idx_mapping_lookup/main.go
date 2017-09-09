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
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logroot"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
)

// *************************************************************************
// This file contains example of how the name-to-index mapping registry
// can be used to register items with unique names, indexes and a metadata
// and how these values can be read.
// ************************************************************************/

/********
 * Main *
 ********/

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log), and example plugin which demonstrates index mapping lookup functionality.
func main() {
	// Init close channel to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combinatioplugin.GoVppmux, n of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{IdxLookupExample: ExamplePlugin{closeChannel: &exampleFinished}}
	agent := core.NewAgent(log.DefaultLogger(), 15*time.Second, append(flavor.Plugins())...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

/**********
 * Flavor *
 **********/

// ExampleFlavor is a set of plugins required for the datasync example.
type ExampleFlavor struct {
	*local.FlavorLocal
	IdxLookupExample ExamplePlugin
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

	if ef.FlavorLocal == nil {
		ef.FlavorLocal = &local.FlavorLocal{}
	}
	ef.FlavorLocal.Inject()

	ef.IdxLookupExample.PluginLogDeps = *ef.LogDeps("idx-mapping-lookup")

	return true
}

// Plugins combines all Plugins in flavor to the list
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct {
	Deps

	exampleIdx   idxvpp.NameToIdxRW // Name to index mapping registry
	exampleIDSeq uint32             // Provides unique ID for every item stored in mapping
	// Fields below are used to properly finish the example
	closeChannel *chan struct{}
}

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	local.PluginLogDeps // injected
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	// Init new name-to-index mapping
	plugin.exampleIdx = nametoidx.NewNameToIdx(logroot.StandardLogger(), plugin.PluginName, "example_index", nil)

	// Set initial ID. After every registration this ID has to be incremented, so new mapping is registered
	// under unique number
	plugin.exampleIDSeq = 1

	plugin.Log.Info("Initialization of the custom plugin for the idx-mapping lookup example is completed")

	// Demonstrate mapping lookup functionality
	plugin.exampleMappingUsage()

	// End the example
	plugin.Log.Infof("idx-mapping-lookup example finished, sending shutdown ...")
	*plugin.closeChannel <- struct{}{}

	return err
}

// Meta structure. It can contain any number of fields of different types. Metadata is optional and can be nil
type Meta struct {
	ip     string
	prefix uint32
}

// Illustration of index-mapping lookup usage
func (plugin *ExamplePlugin) exampleMappingUsage() {
	// Random name used to registration. Every registered name should be unique
	name := "example-entity"

	// Register name, unique ID and metadata to example index map. Metadata are optional, can be nil. Name and ID have
	// to be unique, otherwise the mapping will be overridden
	plugin.exampleIdx.RegisterName(name, plugin.exampleIDSeq, &Meta{})
	plugin.Log.Infof("Name %v registered", name)

	// Find the registered mapping using lookup index (name has to be known). Function returns an index related to
	// provided name, a metadata (nil if there are no metadata or mapping was not found) and a bool flag whether
	// the mapping with provided name was found or not
	_, meta, found := plugin.exampleIdx.LookupIdx(name)
	if found && meta != nil {
		plugin.Log.Infof("Name %v stored in mapping", name)
	} else {
		plugin.Log.Errorf("Name %v not found", name)
	}

	// Find the registered mapping using lookup name (index has to be known). Function returns a name related to
	// provided index, a metadata (nil if there are no metadata or mapping was not found) and a bool flag whether
	// the mapping with provided index was found or not
	_, meta, found = plugin.exampleIdx.LookupName(plugin.exampleIDSeq)
	if found && meta != nil {
		plugin.Log.Infof("Index %v stored in mapping", plugin.exampleIDSeq)
	} else {
		plugin.Log.Errorf("Index %v not found", plugin.exampleIDSeq)
	}

	// This is how to remove mapping from registry. Other plugins can be notified about this change
	plugin.exampleIdx.UnregisterName(name)
	plugin.Log.Infof("Name %v unregistered", name)
}
