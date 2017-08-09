package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/generic"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	"time"
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

	flavor := generic.Flavor{}

	// Example plugin (Logger)
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
	log.Info(message)
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

	// Basic logger options
	log.Print("----------- Log examples -----------")
	log.Printf("Print with format specifier. String: %s, Digit: %d, Value: %v", exampleString, exampleNum, plugin)

	// Format also available for all 6 levels of log levels
	log.Debug("Debug log example: Debugging information")
	log.Info("Info log example: Something informative")
	log.Warn("Warn log example: Something unexpected, warning")
	log.Error("Error log example: Failure without exit")

	//log.Panic("Panic log") calls panic() after logging
	//log.Fatal("Bye") calls os.Exit(1) after logging

	// Log with field - automatically adds timestamp
	log.WithField("Ex. string: ", exampleString).Info("Info log with field example")

	// For multiple fields
	log.WithFields(log.Fields{"Ex. string": exampleString, "Ex. num": exampleNum}).Info("Info log with field example string and num")

	// Log with error
	log.WithError(err).Error("Example log with error")

	// Set log level which logs only entries with current severity or above
	log.SetLevel(logging.DebugLevel) // everything
	log.SetLevel(logging.InfoLevel)  // info, warn, error, panic, fatal - default log level
	log.SetLevel(logging.WarnLevel)  // warn, error, panic, fatal
	// etc

	// Custom logger with generated name
	generatedLogger := log.New()
	// Custom logger with name
	namedLogger, err := log.NewNamed("myLogger")
	if err != nil {
		return err
	}

	// Usage of custom loggers
	generatedLogger.Infof("Log using logger with generated name: %v ", generatedLogger.GetName())
	namedLogger.Infof("Log using named logger with name: %v", namedLogger.GetName())

	// End the example
	log.SetLevel(logging.InfoLevel)

	return nil
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	return nil
}
