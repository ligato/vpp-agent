// Copyright (c) 2019 Cisco and/or its affiliates.
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

package iptablesplugin

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/kvscheduler"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
)

// DefaultPlugin is a default instance of IPTablesPlugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options.
func NewPlugin(opts ...Option) *IPTablesPlugin {
	p := &IPTablesPlugin{}

	p.PluginName = "linux-iptablesplugin"
	p.KVScheduler = &kvscheduler.DefaultPlugin
	p.NsPlugin = &nsplugin.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	if p.Log == nil {
		p.Log = logging.ForPlugin(p.String())
	}
	if p.Cfg == nil {
		p.Cfg = config.ForPlugin(p.String(),
			config.WithCustomizedFlag(config.FlagName(p.String()), "linux-iptablesplugin.conf"),
		)
	}

	return p
}

// Option is a function that can be used in NewPlugin to customize Plugin.
type Option func(*IPTablesPlugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(f func(*Deps)) Option {
	return func(p *IPTablesPlugin) {
		f(&p.Deps)
	}
}
