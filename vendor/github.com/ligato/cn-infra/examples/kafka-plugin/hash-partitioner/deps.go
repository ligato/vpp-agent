package main

import (
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/messaging"
)

// Deps lists dependencies of ExamplePlugin.
type Deps struct {
	Kafka messaging.Mux // injected
	local.PluginLogDeps // injected
}
