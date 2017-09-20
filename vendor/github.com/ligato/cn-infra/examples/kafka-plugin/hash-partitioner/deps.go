package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/messaging/kafka"
)

// Deps lists dependencies of ExamplePlugin.
type Deps struct {
	Kafka               messaging.Mux // injected
	local.PluginLogDeps               // injected
}

// ExampleFlavor is a set of plugins required for the example.
type ExampleFlavor struct {
	// Local flavor to access the Infra (logger, service label, status check)
	*local.FlavorLocal
	// Kafka plugin
	Kafka kafka.Plugin
	// Example plugin
	KafkaExample ExamplePlugin
	// For example purposes, use channel when the example is finished
	closeChan *chan struct{}
}

// Inject sets inter-plugin references.
func (ef *ExampleFlavor) Inject() (allReadyInjected bool) {
	// Init local flavor
	if ef.FlavorLocal == nil {
		ef.FlavorLocal = &local.FlavorLocal{}
	}
	ef.FlavorLocal.Inject()
	// Init kafka
	ef.Kafka.Deps.PluginInfraDeps = *ef.FlavorLocal.InfraDeps("kafka",
		local.WithConf())
	// Inject kafka to example plugin
	ef.KafkaExample.Deps.PluginLogDeps = *ef.FlavorLocal.LogDeps("kafka-example")
	ef.KafkaExample.Kafka = &ef.Kafka
	ef.KafkaExample.closeChannel = ef.closeChan

	return true
}

// Plugins combines all plugins in the flavor into a slice.
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}
