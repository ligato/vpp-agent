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

package stnplugin

import (
	"github.com/google/wire"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

var Wire = wire.NewSet(
	Provider,
	DepsProvider,
)

func DepsProvider(
	scheduler kvs.KVScheduler,
	govppmuxPlugin govppmux.API,
	ifPlugin ifplugin.API,
	statuscheck statuscheck.PluginStatusWriter,
) Deps {
	return Deps{
		StatusCheck: statuscheck,
		KVScheduler: scheduler,
		VPP:         govppmuxPlugin,
		IfPlugin:    ifPlugin,
	}
}

func Provider(deps Deps) (*STNPlugin, error) {
	p := &STNPlugin{Deps: deps}
	p.SetName("vpp-stnplugin")
	p.Log = logging.ForPlugin("vpp-stnplugin")
	return p, p.Init()
}

// DefaultPlugin is a default instance of STN plugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provided Options.
func NewPlugin(opts ...Option) *STNPlugin {
	p := &STNPlugin{}

	p.PluginName = "vpp-stn-plugin"
	p.KVScheduler = &kvscheduler.DefaultPlugin
	p.VPP = &govppmux.DefaultPlugin
	p.IfPlugin = &ifplugin.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	if p.Log == nil {
		p.Log = logging.ForPlugin(p.String())
	}

	return p
}

// Option is a function that can be used in NewPlugin to customize Plugin.
type Option func(*STNPlugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(f func(*Deps)) Option {
	return func(p *STNPlugin) {
		f(&p.Deps)
	}
}
