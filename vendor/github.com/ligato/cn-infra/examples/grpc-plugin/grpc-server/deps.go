package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/cn-infra/rpc/grpc"
)

// Deps - dependencies for ExamplePlugin
type Deps struct {
	local.PluginLogDeps
	GRPC grpc.Server
}

// ExampleFlavor is a composition of ExamplePlugin with Local flavor.
type ExampleFlavor struct {
	rpc.FlavorRPC
	ExamplePlugin
}

// Plugins combines all plugins from the flavor into a slice.
func (f *ExampleFlavor) Plugins() []*core.NamedPlugin {
	if f.FlavorRPC.Inject() {
		f.ExamplePlugin.Deps.PluginLogDeps = *f.LogDeps("example")
		f.ExamplePlugin.Deps.GRPC = &f.GRPC
	}

	return core.ListPluginsInFlavor(f)
}
