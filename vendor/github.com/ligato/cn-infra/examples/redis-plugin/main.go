package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	log "github.com/ligato/cn-infra/logging/logroot"
)

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugin (Status check, and Log)
// and example plugin which demonstrates use of Redis flavor.
func main() {
	// Init close channel used to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combination of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{closeChan: &exampleFinished}
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(flavor.Plugins())...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// ExamplePlugin to depict the use of Redis flavor
type ExamplePlugin struct {
	Deps // plugin dependencies are injected

	closeChannel *chan struct{}
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	return nil
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	*plugin.closeChannel <- struct{}{}
	return nil
}
