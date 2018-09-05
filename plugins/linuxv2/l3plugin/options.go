package l3plugin

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/kvscheduler"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin"
	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
)

// DefaultPlugin is a default instance of IfPlugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *L3PLugin {
	p := &L3PLugin{}

	p.PluginName = "linux-l3plugin"
	p.Scheduler = &kvscheduler.DefaultPlugin
	p.NsPlugin = &nsplugin.DefaultPlugin
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
type Option func(*L3PLugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(f func(*Deps)) Option {
	return func(p *L3PLugin) {
		f(&p.Deps)
	}
}
