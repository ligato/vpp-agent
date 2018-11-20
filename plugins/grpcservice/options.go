package grpcservice

import "github.com/ligato/cn-infra/rpc/grpc"

// DefaultPlugin is default instance of Plugin
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.PluginName = "grpc-sync-service"
	p.GRPC = &grpc.DefaultPlugin

	for _, o := range opts {
		o(p)
	}
	p.PluginDeps.Setup()

	return p
}

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Plugin)
