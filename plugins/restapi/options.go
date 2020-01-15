//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package restapi

import (
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/servicelabel"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	linuxifplugin "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
)

// DefaultPlugin is a default instance of Plugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.PluginName = "restpapi"
	p.HTTPHandlers = &rest.DefaultPlugin
	p.VPP = &govppmux.DefaultPlugin
	p.ServiceLabel = &servicelabel.DefaultPlugin
	p.AddrAlloc = &netalloc.DefaultPlugin
	p.VPPACLPlugin = &aclplugin.DefaultPlugin
	p.VPPIfPlugin = &ifplugin.DefaultPlugin
	p.VPPL2Plugin = &l2plugin.DefaultPlugin
	p.VPPL3Plugin = &l3plugin.DefaultPlugin
	p.LinuxIfPlugin = &linuxifplugin.DefaultPlugin
	p.NsPlugin = &nsplugin.DefaultPlugin

	for _, o := range opts {
		o(p)
	}

	p.PluginDeps.Setup()

	return p
}

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Plugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(cb func(*Deps)) Option {
	return func(p *Plugin) {
		cb(&p.Deps)
	}
}
