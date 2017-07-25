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
	"github.com/ligato/cn-infra/logging/logroot"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/flavours/vpp"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"strconv"
	"time"
)

// *************************************************************************
// This file contains example of how to watch on changes done in name-to-index
// mapping registry.
// The procedure requires a subscriber channel used in the watcher to listen on
// created, modified or removed items in the registry.
// ************************************************************************/

/********
 * Main *
 ********/

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log), and example plugin which demonstrates index mapping watcher functionality.
func main() {
	// Init close channel to stop the example
	closeChannel := make(chan struct{}, 1)

	f := vpp.Flavour{}

	// Example plugin (Index mapping watcher)
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{}}

	// Create new agent
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(f.Plugins(), examplePlugin)...)

	// End when the idx_mapping_watcher example is finished
	go closeExample("idx_mapping_watcher example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(7 * time.Second)
	log.Info(message)
	closeChannel <- struct{}{}
}

/**********************
 * Example plugin API *
 **********************/

// PluginID of the custom index mapping watcher plugin
const PluginID core.PluginName = "example-plugin"

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct {
	exampleConfigurator *ExampleConfigurator           // Plugin configurator
	exampleIdx          idxvpp.NameToIdxRW             // Name-to-index mapping
	exIdxWatchChannel   chan idxvpp.NameToIdxDto       // Channel to watch changes in mapping
	watchDataReg        datasync.WatchDataRegistration // To subscribe to mapping change events
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	// Init new name-to-index mapping
	plugin.exampleIdx = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "example_index", nil)

	// Initialize configurator
	plugin.exampleConfigurator = &ExampleConfigurator{
		exampleIndex: plugin.exampleIdx, // Pass index mapping
		exampleIDSeq: 1,                 // Set initial ID
	}

	// Mapping channel is used to notify about changes in the mapping registry
	plugin.exIdxWatchChannel = make(chan idxvpp.NameToIdxDto, 100)

	// Start watcher before configurator init
	go plugin.watchEvents()

	// Init configurator
	err = plugin.exampleConfigurator.Init()

	// Subscribe name-to-index watcher
	plugin.exampleIdx.Watch(PluginID, nametoidx.ToChan(plugin.exIdxWatchChannel))

	log.Info("Initialization of the custom plugin for the idx-mapping watcher example is completed")

	return err
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	plugin.exampleConfigurator.Close()
	return nil
}

/*************************
 * Example plugin config *
 *************************/

// ExampleConfigurator usually initializes configuration-specific fields or other tasks (e.g. defines GOVPP channels
// if they are used, checks VPP message compatibility etc.)
type ExampleConfigurator struct {
	exampleIDSeq uint32             // Unique ID
	exampleIndex idxvpp.NameToIdxRW // Index mapping
}

// Init members of configurator (none in this example)
func (configurator *ExampleConfigurator) Init() (err error) {
	log.Info("Default plugin configurator ready")

	go func() {
		// This function registers several name to index items to registry owned by the configurator
		for i := 1; i <= 5; i++ {
			configurator.RegisterTestData(i)
		}
	}()

	return err
}

// Close function for example plugin (just for representation, there is nothing to close in the example)
func (configurator *ExampleConfigurator) Close() {}

/************
 * Register *
 ************/

// RegisterTestData registers item to the name to index registry
func (configurator *ExampleConfigurator) RegisterTestData(index int) {
	// Generate name used in registration. In the example, an index is added to the name to made it unique
	name := "example-entity-" + strconv.Itoa(index)
	// Register name to index mapping with name and index. In this example, no metadata is used so the last
	// is nil. Metadata are optional.
	configurator.exampleIndex.RegisterName(name, configurator.exampleIDSeq, nil)
	configurator.exampleIDSeq++
	log.Infof("Name %v registered", name)
}

/***********
 * Watcher *
 ***********/

// Watch on name to index mapping changes created in configurator
func (plugin *ExamplePlugin) watchEvents() {
	log.Info("Watcher started")
	for {
		select {
		case exIdx := <-plugin.exIdxWatchChannel:
			log.Infof("Index event arrived to watcher, key %v", exIdx.Idx)
			if exIdx.IsDelete() {
				// IsDelete flag recognizes what kind of event arrived (put or delete)
			}
			// Done is used to signal to the event producer that the event consumer has processed the event.
			// User of the API is supposed to clear event with Done()
			exIdx.Done()
		}
	}
}
