package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/rpc"
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
	closeChannel := make(chan struct{}, 1)

	flavor := rpc.FlavorRPC{}

	// Example plugin (flags)
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{}}
	// Create new agent
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(flavor.Plugins(), examplePlugin)...)

	// End when the flag example is finished
	go closeExample("Flags example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(8 * time.Second)
	log.StandardLogger().Info(message)
	closeChannel <- struct{}{}
}

/**********************
 * Example plugin API *
 **********************/

// PluginID of the custom flags plugin
const PluginID core.PluginName = "example-plugin"

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct{}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	// RegisterFlags contains examples of how register flags of various types. Has to be called from plugin Init()
	// function.
	registerFlags()

	log.StandardLogger().Info("Initialization of the custom plugin for the flags example is completed")

	go func() {
		// logFlags shows the runtime values of CLI flags registered in RegisterFlags()
		logFlags()
	}()

	return err
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime (just for reference, nothing needs to be cleaned up here)
func (plugin *ExamplePlugin) Close() error {
	return nil
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
func registerFlags() {
	log.StandardLogger().Info("Registering flags")
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
func logFlags() {
	time.Sleep(3 * time.Second)
	log.StandardLogger().Info("Logging flags")
	log.StandardLogger().Infof("testFlagString:'%s'", testFlagString)
	log.StandardLogger().Infof("testFlagInt:'%d'", testFlagInt)
	log.StandardLogger().Infof("testFlagInt64:'%d'", testFlagInt64)
	log.StandardLogger().Infof("testFlagUint:'%d'", testFlagUint)
	log.StandardLogger().Infof("testFlagUint64:'%d'", testFlagUint64)
	log.StandardLogger().Infof("testFlagBool:'%v'", testFlagBool)
	log.StandardLogger().Infof("testFlagDur:'%v'", testFlagDur)
}
