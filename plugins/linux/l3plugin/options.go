package l3plugin

import (
	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
)

// DefaultPlugin is a default instance of IfPlugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *L3Plugin {
	p := &L3Plugin{}

	p.PluginName = "linux-l3plugin"
	p.KVScheduler = &kvscheduler.DefaultPlugin
	p.NsPlugin = &nsplugin.DefaultPlugin
	p.AddrAlloc = &netalloc.DefaultPlugin
	p.IfPlugin = &ifplugin.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	if p.Log == nil {
		p.Log = logging.ForPlugin(p.String())
	}
	if p.Cfg == nil {
		p.Cfg = config.ForPlugin(p.String(),
			config.WithCustomizedFlag(config.FlagName(p.String()), "linux-l3plugin.conf"),
		)
	}

	return p
}

// Option is a function that can be used in NewPlugin to customize Plugin.
type Option func(*L3Plugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(f func(*Deps)) Option {
	return func(p *L3Plugin) {
		f(&p.Deps)
	}
}
