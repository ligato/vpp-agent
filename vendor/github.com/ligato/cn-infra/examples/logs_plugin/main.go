package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
)

// *************************************************************************
// This file contains a logger use cases. To define a custom logger, use
// PluginLogger.NewLogger(name). The logger is using 6 levels of logging:
// - Debug
// - Info (this one is default)
// - Warn
// - Error
// - Panic
// - Fatal
//
// Global log levels can be changed locally with the Logger.SetLevel()
// or remotely using REST (but different flavor must be used: rpc.RpcFlavor).
// ************************************************************************/

/********
 * Main *
 ********/

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log) and example plugin which demonstrates Logs functionality.
func main() {
	// Init close channel to stop the example after everything was logged
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combination of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{ExamplePlugin: ExamplePlugin{exampleFinished: exampleFinished}}
	agent := core.NewAgent(logroot.StandardLogger(), 15*time.Second, flavor.Plugins()...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

/**********
 * Flavor *
 **********/

// ExampleFlavor is composition of ExamplePlugin and existing flavor
type ExampleFlavor struct {
	local.FlavorLocal
	ExamplePlugin
}

// Plugins combines all Plugins in flavor to the list
func (f *ExampleFlavor) Plugins() []*core.NamedPlugin {
	if f.FlavorLocal.Inject() {
		f.ExamplePlugin.PluginLogDeps = *f.LogDeps("logs-example")
	}

	return core.ListPluginsInFlavor(f)
}

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct {
	local.PluginLogDeps // this field is usually injected in flavor
	exampleFinished         chan struct{}
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	exampleString := "example"
	exampleNum := 15

	// Set log level which logs only entries with current severity or above
	plugin.Log.SetLevel(logging.WarnLevel)  // warn, error, panic, fatal
	plugin.Log.SetLevel(logging.InfoLevel)  // info, warn, error, panic, fatal - default log level
	plugin.Log.SetLevel(logging.DebugLevel) // everything

	// Basic logger options
	plugin.Log.Print("----------- Log examples -----------")
	plugin.Log.Printf("Print with format specifier. String: %s, Digit: %d, Value: %v", exampleString, exampleNum, plugin)

	// Format also available for all 6 levels of log levels
	plugin.Log.Debug("Debug log example: Debugging information")
	plugin.Log.Info("Info log example: Something informative")
	plugin.Log.Warn("Warn log example: Something unexpected, warning")
	plugin.Log.Error("Error log example: Failure without exit")
	plugin.showPanicLog()
	//log.Fatal("Bye") calls os.Exit(1) after logging

	// Log with field - automatically adds timestamp
	plugin.Log.WithField("Ex. string: ", exampleString).Info("Info log with field example")
	// For multiple fields
	plugin.Log.WithFields(map[string]interface{}{"Ex. string": exampleString, "Ex. num": exampleNum}).Info("Info log with field example string and num")

	// Custom (child) logger with name
	childLogger := plugin.Log.NewLogger("childLogger")
	// Usage of custom loggers
	childLogger.Infof("Log using named logger with name: %v", childLogger.GetName())

	// End the example
	plugin.Log.Info("logs in plugin example finished, sending shutdown ...")
	plugin.exampleFinished <- struct{}{}

	return nil
}

// Demostrates panic log + recovering
func (plugin *ExamplePlugin) showPanicLog() {
	defer func() {
		if err := recover(); err != nil {
			plugin.Log.Info("Recovered from panic")
		}
	}()
	plugin.Log.Panic("Panic log: calls panic() after log, will be recovered") //calls panic() after logging
}
