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
	"strconv"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logroot"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
)

// *************************************************************************
// This file contains example of how to watch on changes done in name-to-index
// mapping registry.
// The procedure requires a subscriber channel used in the watcher to listen on
// created, modified or removed items in the registry.
// ************************************************************************/

const expectedEvents = 5

/********
 * Main *
 ********/

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log), and example plugin which demonstrates index mapping watcher functionality.
func main() {
	// Init close channel to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combinatioplugin.GoVppmux, n of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{IdxWatchExample: ExamplePlugin{closeChannel: &exampleFinished}}
	agent := core.NewAgent(log.DefaultLogger(), 15*time.Second, append(flavor.Plugins())...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

/**********
 * Flavor *
 **********/

// ExampleFlavor is a set of plugins required for the datasync example.
type ExampleFlavor struct {
	*local.FlavorLocal
	IdxWatchExample ExamplePlugin
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

	ef.IdxWatchExample.PluginLogDeps = *ef.LogDeps("idx-watch-lookup")

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

	exampleIdx        idxvpp.NameToIdxRW         // Name-to-index mapping
	exampleIDSeq      uint32                     // Unique ID
	exIdxWatchChannel chan idxvpp.NameToIdxDto   // Channel to watch changes in mapping
	watchDataReg      datasync.WatchRegistration // To subscribe to mapping change events
	// Fields below are used to properly finish the example
	eventCounter uint8
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

	// Mapping channel is used to notify about changes in the mapping registry
	plugin.exIdxWatchChannel = make(chan idxvpp.NameToIdxDto, 100)

	plugin.Log.Info("Initialization of the custom plugin for the idx-mapping watcher example is completed")

	// Start watcher before plugin init
	go plugin.watchEvents()

	go func() {
		// This function registers several name to index items to registry owned by the plugin
		for i := 1; i <= 5; i++ {
			plugin.RegisterTestData(i)
		}
	}()

	// Subscribe name-to-index watcher
	plugin.exampleIdx.Watch(plugin.PluginName, nametoidx.ToChan(plugin.exIdxWatchChannel))

	return err
}

/************
 * Register *
 ************/

// RegisterTestData registers item to the name to index registry
func (plugin *ExamplePlugin) RegisterTestData(index int) {
	// Generate name used in registration. In the example, an index is added to the name to made it unique
	name := "example-entity-" + strconv.Itoa(index)
	// Register name to index mapping with name and index. In this example, no metadata is used so the last
	// is nil. Metadata are optional.
	plugin.exampleIdx.RegisterName(name, plugin.exampleIDSeq, nil)
	plugin.exampleIDSeq++
	plugin.Log.Infof("Name %v registered", name)
}

/***********
 * Watcher *
 ***********/

// Watch on name to index mapping changes created in plugin
func (plugin *ExamplePlugin) watchEvents() {
	plugin.Log.Info("Watcher started")
	for {
		select {
		case exIdx := <-plugin.exIdxWatchChannel:
			// Just for example purpose
			plugin.eventCounter++

			plugin.Log.Infof("Index event arrived to watcher, key %v", exIdx.Idx)
			if exIdx.IsDelete() {
				// IsDelete flag recognizes what kind of event arrived (put or delete)
			}
			// Done is used to signal to the event producer that the event consumer has processed the event.
			// User of the API is supposed to clear event with Done()
			exIdx.Done()

			// End the example when it is done (5 events are expected)
			if plugin.eventCounter == expectedEvents {
				plugin.Log.Infof("idx-watch-lookup example finished, sending shutdown ...")
				*plugin.closeChannel <- struct{}{}
			}
		}
	}
}
