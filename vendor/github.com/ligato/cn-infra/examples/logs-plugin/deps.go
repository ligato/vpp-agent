package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
)

// Deps - dependencies for ExamplePlugin
type Deps struct {
	local.PluginLogDeps
}

// ExampleFlavor is a composition of ExamplePlugin with Local flavor.
type ExampleFlavor struct {
	local.FlavorLocal
	ExamplePlugin
}

// Plugins combines all plugins from the flavor into a slice.
func (f *ExampleFlavor) Plugins() []*core.NamedPlugin {
	if f.FlavorLocal.Inject() {
		f.ExamplePlugin.PluginLogDeps = *f.LogDeps("logs-example")
	}

	return core.ListPluginsInFlavor(f)
}
