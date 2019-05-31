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

package supervisor

import pm "github.com/ligato/cn-infra/exec/processmanager"

// DefaultPlugin is a default instance of the supervisor plugin
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new supervisor plugin with the provided options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.PluginName = "supervisor"
	p.PM = &pm.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	p.PluginDeps.Setup()

	return p
}

// Option is a function that can be used in the NewPlugin to customize plugin options
type Option func(*Plugin)

// UseDeps returns an option that can inject custom dependencies
func UseDeps(cb func(*Deps)) Option {
	return func(p *Plugin) {
		cb(&p.Deps)
	}
}

// UseConf returns an option which injects a particular configuration
func UseConf(conf Config) Option {
	return func(p *Plugin) {
		p.config = &conf
	}
}
