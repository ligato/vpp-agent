package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/logroot"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/flavours/vpp"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"time"
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
	closeChannel := make(chan struct{}, 1)
	f := vpp.Flavour{}
	// Example plugin (Index mapping lookup)
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{}}

	// Create new agent
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(f.Plugins(), examplePlugin)...)

	// End when the idx_mapping_lookup example is finished
	go closeExample("idx_mapping_lookup example finished", closeChannel)

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

// PluginID of the custom index mapping lookup plugin
const PluginID core.PluginName = "example-plugin"

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct {
	exampleIdx   idxvpp.NameToIdxRW // Name to index mapping registry
	exampleIDSeq uint32             // Provides unique ID for every item stored in mapping
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	// Init new name-to-index mapping
	plugin.exampleIdx = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "example_index", nil)

	// Set initial ID. After every registration this ID has to be incremented, so new mapping is registered
	// under unique number
	plugin.exampleIDSeq = 1

	log.Info("Initialization of the custom plugin for the idx-mapping lookup example is completed")

	// Demonstrate mapping lookup functionality
	go plugin.exampleMappingUsage()

	return err
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime (just for reference, nothing needs to be cleaned up here)
func (plugin *ExamplePlugin) Close() error {
	return nil
}

// Metadata structure. It can contain any number of fields of different types. Metadata is optional and can be nil
type Meta struct {
	ip     string
	prefix uint32
}

// Illustration of index-mapping lookup usage
func (plugin *ExamplePlugin) exampleMappingUsage() {
	time.Sleep(3 * time.Second)

	// Random name used to registration. Every registered name should be unique
	name := "example-entity"

	// Register name, unique ID and metadata to example index map. Metadata are optional, can be nil. Name and ID have
	// to be unique, otherwise the mapping will be overridden
	plugin.exampleIdx.RegisterName(name, plugin.exampleIDSeq, &Meta{})
	log.Infof("Name %v registered", name)

	// Find the registered mapping using lookup index (name has to be known). Function returns an index related to
	// provided name, a metadata (nil if there are no metadata or mapping was not found) and a bool flag whether
	// the mapping with provided name was found or not
	_, meta, found := plugin.exampleIdx.LookupIdx(name)
	if found && meta != nil {
		log.Infof("Name %v stored in mapping", name)
	} else {
		log.Errorf("Name %v not found", name)
	}

	// Find the registered mapping using lookup name (index has to be known). Function returns a name related to
	// provided index, a metadata (nil if there are no metadata or mapping was not found) and a bool flag whether
	// the mapping with provided index was found or not
	_, meta, found = plugin.exampleIdx.LookupName(plugin.exampleIDSeq)
	if found && meta != nil {
		log.Infof("Index %v stored in mapping", plugin.exampleIDSeq)
	} else {
		log.Errorf("Index %v not found", plugin.exampleIDSeq)
	}

	// This is how to remove mapping from registry. Other plugins can be notified about this change
	plugin.exampleIdx.UnregisterName(name)
	log.Infof("Name %v unregistered", name)
}
