// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ifplugin

import (
	"github.com/google/wire"
	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/servicelabel"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
)

var Wire = wire.NewSet(
	Provider,
	DepsProvider,
	//wire.Struct(new(Deps), "NsPlugin", "StatusCheck", "ServiceLabel", "KVScheduler", "VPP", "AddrAlloc" /*"LinuxIfPlugin"*/),
	//wire.InterfaceValue(new(API), &IfPlugin{}),
	wire.Bind(new(API), new(*IfPlugin)),
)

func DepsProvider(
	scheduler kvs.KVScheduler,
	govppmuxPlugin govppmux.API,
	addrallocPlugin netalloc.AddressAllocator,
	nsPlugin nsplugin.API,
	statuscheck statuscheck.PluginStatusWriter,
) Deps {
	return Deps{
		StatusCheck: statuscheck,
		KVScheduler: scheduler,
		VPP:         govppmuxPlugin,
		AddrAlloc:   addrallocPlugin,
		NsPlugin:    nsPlugin,
	}
}

func Provider(deps Deps) (*IfPlugin, func(), error) {
	p := &IfPlugin{Deps: deps}
	p.SetName("vpp-ifplugin")
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

	p.PluginName = "vpp-ifplugin"
	p.StatusCheck = &statuscheck.DefaultPlugin
	p.KVScheduler = &kvscheduler.DefaultPlugin
	p.VPP = &govppmux.DefaultPlugin
	p.ServiceLabel = &servicelabel.DefaultPlugin
	p.AddrAlloc = &netalloc.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	if p.Log == nil {
		p.Log = logging.ForPlugin(p.String())
	}
	if p.Cfg == nil {
		p.Cfg = config.ForPlugin(p.String(),
			config.WithCustomizedFlag(config.FlagName(p.String()), "vpp-ifplugin.conf"),
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
