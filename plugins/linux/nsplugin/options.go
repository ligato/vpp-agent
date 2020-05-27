package nsplugin

import (
	"github.com/google/wire"
	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

var Wire = wire.NewSet(
	Provider,
	ConfigProvider,
	DepsProvider,
	wire.Bind(new(API), new(*NsPlugin)),
)

func DepsProvider(scheduler kvs.KVScheduler) Deps {
	return Deps{
		KVScheduler: scheduler,
	}
}

func ConfigProvider(conf config.Config) *Config {
	var cfg = DefaultConfig()
	if err := conf.UnmarshalKey("http", &cfg); err != nil {
		logging.Errorf("unmarshal key failed: %v", err)
	}
	return cfg
}

func Provider(deps Deps, conf *Config) (*NsPlugin, func(), error) {
	p := &NsPlugin{Deps: deps}
	p.conf = conf
	p.SetName("linux-nsplugin")
	p.Log = logging.ForPlugin("linux-nsplugin")
	cancel := func() {
		if err := p.Close(); err != nil {
			p.Log.Error(err)
		}
	}
	return p, cancel, p.Init()
}

// DefaultPlugin is a default instance of IfPlugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *NsPlugin {
	p := &NsPlugin{}

	p.PluginName = "linux-nsplugin"
	p.KVScheduler = &kvscheduler.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	if p.Log == nil {
		p.Log = logging.ForPlugin(p.String())
	}
	if p.Cfg == nil {
		p.Cfg = config.ForPlugin(p.String(),
			config.WithCustomizedFlag(config.FlagName(p.String()), "linux-nsplugin.conf"),
		)
	}

	return p
}

// Option is a function that can be used in NewPlugin to customize Plugin.
type Option func(*NsPlugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(f func(*Deps)) Option {
	return func(p *NsPlugin) {
		f(&p.Deps)
	}
}
