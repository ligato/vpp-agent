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

package configurator

import (
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/orchestrator"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin"
)

// DefaultPlugin is default instance of Plugin
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.PluginName = "configurator"
	p.GRPCServer = &grpc.DefaultPlugin
	p.Dispatch = &orchestrator.DefaultPlugin
	p.GoVppmux = &govppmux.DefaultPlugin
	p.VPPACLPlugin = &aclplugin.DefaultPlugin
	p.VPPIfPlugin = &ifplugin.DefaultPlugin
	p.VPPL2Plugin = &l2plugin.DefaultPlugin
	p.VPPL3Plugin = &l3plugin.DefaultPlugin

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
