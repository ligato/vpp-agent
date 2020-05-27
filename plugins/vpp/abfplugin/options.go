//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package abfplugin

import (
	"github.com/google/wire"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

var Wire = wire.NewSet(
	Provider,
	DepsProvider,
	//wire.Struct(new(Deps), "StatusCheck", "Scheduler", "VPP", "ACLPlugin", "IfPlugin"),
)

func DepsProvider(
	scheduler kvs.KVScheduler,
	govppmuxPlugin govppmux.API,
	aclPlugin aclplugin.API,
	ifPlugin ifplugin.API,
	statuscheck statuscheck.PluginStatusWriter,
) Deps {
	return Deps{
		StatusCheck: statuscheck,
		Scheduler:   scheduler,
		VPP:         govppmuxPlugin,
		ACLPlugin:   aclPlugin,
		IfPlugin:    ifPlugin,
	}
}

func Provider(deps Deps) (*ABFPlugin, error) {
	p := &ABFPlugin{Deps: deps}
	p.SetName("vpp-abfplugin")
	p.Log = logging.ForPlugin("vpp-abf-plugin")
	return p, p.Init()
}

// DefaultPlugin is a default instance of ABFPlugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *ABFPlugin {
	p := &ABFPlugin{}

	p.PluginName = "vpp-abfplugin"
	p.StatusCheck = &statuscheck.DefaultPlugin
	p.Scheduler = &kvscheduler.DefaultPlugin
	p.VPP = &govppmux.DefaultPlugin
	p.ACLPlugin = &aclplugin.DefaultPlugin
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
type Option func(plugin *ABFPlugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(f func(*Deps)) Option {
	return func(p *ABFPlugin) {
		f(&p.Deps)
	}
}
