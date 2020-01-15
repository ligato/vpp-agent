package netalloc

import (
	"github.com/ligato/cn-infra/logging"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
)

// DefaultPlugin is a default instance of netalloc plugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.PluginName = "netalloc"
	p.KVScheduler = &kvscheduler.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	if p.Log == nil {
		p.Log = logging.ForPlugin(p.String())
	}

	return p
}

// Option is a function that can be used in NewPlugin to customize Plugin.
type Option func(plugin *Plugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(f func(*Deps)) Option {
	return func(p *Plugin) {
		f(&p.Deps)
	}
}
