package main

import (
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/rpc/grpc"
)

// Deps - dependencies for ExamplePlugin
type Deps struct {
	local.PluginLogDeps
	GRPC grpc.Server
}
