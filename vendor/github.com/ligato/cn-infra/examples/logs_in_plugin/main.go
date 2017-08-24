package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/logging/logrus"
)

// *************************************************************************
// This file contains a logger use cases. To define a custom logger, use
// log.New() (or log.NewNamed(name) with a user-specified name). The logger is
// using 6 levels of logging:
// - Debug
// - Info (this one is default)
// - Warn
// - Error
// - Panic
// - Fatal
//
// Global log levels can be changed with the log.SetLevel()
// ************************************************************************/

/********
 * Main *
 ********/

// Channel used to close agent if example is finished
var closeChannel chan struct{}

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log) and example plugin which demonstrates Logs functionality.
func main() {
	// Init close channel to stop the example
	closeChannel = make(chan struct{}, 1)

	flavor := rpc.FlavorRPC{}

	// Example plugin (StandardLogger)
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{}}

	// Create new agent
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(flavor.Plugins(), examplePlugin)...)

	// End when the logs example is finished
	go closeExample("logs example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(6 * time.Second)
	log.StandardLogger().Info(message)
	closeChannel <- struct{}{}
}

/**********************
 * Example plugin API *
 **********************/

// PluginID of the custom ETCD plugin
const PluginID core.PluginName = "example-plugin"

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct{}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	exampleString := "example"
	exampleNum := 15

	//TODO FIXME Inject PluginLogger instead of global log

	// Basic logger options
	log.StandardLogger().Print("----------- Log examples -----------")
	log.StandardLogger().Printf("Print with format specifier. String: %s, Digit: %d, Value: %v", exampleString, exampleNum, plugin)

	// Format also available for all 6 levels of log levels
	log.StandardLogger().Debug("Debug log example: Debugging information")
	log.StandardLogger().Info("Info log example: Something informative")
	log.StandardLogger().Warn("Warn log example: Something unexpected, warning")
	log.StandardLogger().Error("Error log example: Failure without exit")

	//log.Panic("Panic log") calls panic() after logging
	//log.Fatal("Bye") calls os.Exit(1) after logging

	// Log with field - automatically adds timestamp
	log.StandardLogger().WithField("Ex. string: ", exampleString).Info("Info log with field example")

	// For multiple fields
	log.StandardLogger().WithFields(map[string]interface{}{"Ex. string": exampleString, "Ex. num": exampleNum}).Info("Info log with field example string and num")

	// Set log level which logs only entries with current severity or above
	log.StandardLogger().SetLevel(logging.DebugLevel) // everything
	log.StandardLogger().SetLevel(logging.InfoLevel)  // info, warn, error, panic, fatal - default log level
	log.StandardLogger().SetLevel(logging.WarnLevel)  // warn, error, panic, fatal
	// etc

	// Custom logger with name
	namedLogger := logrus.NewLogger("myLogger")

	// Usage of custom loggers
	namedLogger.Infof("Log using named logger with name: %v", namedLogger.GetName())

	// End the example

	return nil
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	return nil
}
