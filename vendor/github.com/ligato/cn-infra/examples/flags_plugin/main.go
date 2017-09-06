package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	log "github.com/ligato/cn-infra/logging/logroot"
	"github.com/namsral/flag"
)

// *************************************************************************
// This file contains example of how to register CLI flags and how to show
// their runtime values
// ************************************************************************/

/********
 * Main *
 ********/

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log) and example plugin which demonstrates usage of flags
func main() {
	// Init close channel to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combination of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{closeChan: &exampleFinished}
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, flavor.Plugins()...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

/**********
 * Flavor *
 **********/

// ExampleFlavor is a set of plugins required for the datasync example.
type ExampleFlavor struct {
	// Local flavor to access to Infra (logger, service label, status check)
	Local local.FlavorLocal
	// Example plugin
	FlagsExample ExamplePlugin
	// For example purposes, use channel when the example is finished
	closeChan *chan struct{}
}

// Inject sets object references
func (ef *ExampleFlavor) Inject() (allReadyInjected bool) {
	// Init local flavor
	ef.Local.Inject()
	// Inject infra to example plugin
	ef.FlagsExample.PluginLogDeps = *ef.Local.LogDeps("flags-example")
	ef.FlagsExample.closeChannel = ef.closeChan

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

	// Fields below are used to properly finish the example
	done         bool
	closeChannel *chan struct{}
}

// Deps is here to group injected dependencies of plugin to not mix with other plugin fields
type Deps struct {
	local.PluginLogDeps // injected
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	// RegisterFlags contains examples of how register flags of various types. Has to be called from plugin Init()
	// function.
	plugin.registerFlags()

	plugin.Log.Info("Initialization of the custom plugin for the flags example is completed")

	// logFlags shows the runtime values of CLI flags registered in RegisterFlags()
	plugin.logFlags()

	go plugin.closeExample()

	return err
}

func (plugin *ExamplePlugin) closeExample() {
	for {
		if plugin.done {
			plugin.Log.Info("flags example finished, sending shutdown ...")
			*plugin.closeChannel <- struct{}{}
			break
		}
	}
}

/*********
 * Flags *
 *********/

// Flag variables
var (
	testFlagString string
	testFlagInt    int
	testFlagInt64  int64
	testFlagUint   uint
	testFlagUint64 uint64
	testFlagBool   bool
	testFlagDur    time.Duration
)

// RegisterFlags contains examples of how to register flags of various types
func (plugin *ExamplePlugin) registerFlags() {
	plugin.Log.Info("Registering flags")
	flag.StringVar(&testFlagString, "ep-string", "my-value",
		"Example of a string flag.")
	flag.IntVar(&testFlagInt, "ep-int", 1122,
		"Example of an int flag.")
	flag.Int64Var(&testFlagInt64, "ep-int64", -3344,
		"Example of an int64 flag.")
	flag.UintVar(&testFlagUint, "ep-uint", 5566,
		"Example of a uint flag.")
	flag.Uint64Var(&testFlagUint64, "ep-uint64", 7788,
		"Example of a uint64 flag.")
	flag.BoolVar(&testFlagBool, "ep-bool", true,
		"Example of a bool flag.")
	flag.DurationVar(&testFlagDur, "ep-duration", time.Second*5,
		"Example of a duration flag.")
}

// LogFlags shows the runtime values of CLI flags
func (plugin *ExamplePlugin) logFlags() {
	plugin.Log.Info("Logging flags")
	plugin.Log.Infof("testFlagString:'%s'", testFlagString)
	plugin.Log.Infof("testFlagInt:'%d'", testFlagInt)
	plugin.Log.Infof("testFlagInt64:'%d'", testFlagInt64)
	plugin.Log.Infof("testFlagUint:'%d'", testFlagUint)
	plugin.Log.Infof("testFlagUint64:'%d'", testFlagUint64)
	plugin.Log.Infof("testFlagBool:'%v'", testFlagBool)
	plugin.Log.Infof("testFlagDur:'%v'", testFlagDur)
	plugin.done = true
}
