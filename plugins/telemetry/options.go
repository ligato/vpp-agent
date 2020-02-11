package telemetry

import (
	"go.ligato.io/cn-infra/v2/rpc/grpc"
	"go.ligato.io/cn-infra/v2/rpc/prometheus"
	"go.ligato.io/cn-infra/v2/rpc/rest"
	"go.ligato.io/cn-infra/v2/servicelabel"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

// DefaultPlugin is default instance of Plugin
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.PluginName = "telemetry"
	p.ServiceLabel = &servicelabel.DefaultPlugin
	p.VPP = &govppmux.DefaultPlugin
	p.Prometheus = &prometheus.DefaultPlugin
	p.GRPC = &grpc.DefaultPlugin
	p.HTTPHandlers = &rest.DefaultPlugin
	p.IfPlugin = &ifplugin.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	p.PluginDeps.Setup()

	return p
}

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Plugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(cb func(*Deps)) Option {
	return func(p *Plugin) {
		cb(&p.Deps)
	}
}
