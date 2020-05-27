package ifplugin

import (
	"github.com/google/wire"
	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/servicelabel"

	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

var Wire = wire.NewSet(
	Provider,
	DepsProvider,
	//wire.Struct(new(Deps), "ServiceLabel", "AddrAlloc", "VppIfPlugin", "NsPlugin", "KVScheduler"),
	wire.Bind(new(API), new(*IfPlugin)),
)

func DepsProvider(
	scheduler kvs.KVScheduler,
	addrallocPlugin netalloc.AddressAllocator,
	nsPlugin nsplugin.API,
	ifPlugin ifplugin.API,
	label servicelabel.ReaderAPI,
) Deps {
	return Deps{
		ServiceLabel: label,
		KVScheduler:  scheduler,
		AddrAlloc:    addrallocPlugin,
		NsPlugin:     nsPlugin,
		VppIfPlugin:  ifPlugin,
	}
}

func Provider(deps Deps) (*IfPlugin, func(), error) {
	p := &IfPlugin{Deps: deps}
	p.SetName("linux-if-plugin")
	p.Setup()
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
func NewPlugin(opts ...Option) *IfPlugin {
	p := &IfPlugin{}

	p.PluginName = "linux-ifplugin"
	p.KVScheduler = &kvscheduler.DefaultPlugin
	p.NsPlugin = &nsplugin.DefaultPlugin
	p.AddrAlloc = &netalloc.DefaultPlugin
	p.ServiceLabel = &servicelabel.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	if p.Log == nil {
		p.Log = logging.ForPlugin(p.String())
	}
	if p.Cfg == nil {
		p.Cfg = config.ForPlugin(p.String(),
			config.WithCustomizedFlag(config.FlagName(p.String()), "linux-ifplugin.conf"),
		)
	}

	return p
}

// Option is a function that can be used in NewPlugin to customize Plugin.
type Option func(*IfPlugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(f func(*Deps)) Option {
	return func(p *IfPlugin) {
		f(&p.Deps)
	}
}
