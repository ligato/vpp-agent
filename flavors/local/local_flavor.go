// Copyright (c) 2017 Cisco and/or its affiliates.
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

// Package local defines flavor used for VPP agents managed only locally.
package local

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"

	"github.com/ligato/cn-infra/datasync"
	local_sync "github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/vpp-agent/clientv1/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
)

// NewAgent returns a new instance of the Agent with plugins.
// It is an alias for core.NewAgent() to implicit use of the FlavorVppLocal
func NewAgent(opts ...core.Option) *core.Agent {
	return core.NewAgent(&FlavorVppLocal{}, opts...)
}

// WithPlugins for adding custom plugins to SFC Controller
// <listPlugins> is a callback that uses flavor input to
// inject dependencies for custom plugins that are in output
//
// Example:
//
//    NewAgent(vppFlavor.WithPlugins(func(flavor) {
// 	       return []*core.NamedPlugin{{"my-plugin", &MyPlugin{DependencyXY: &flavor.FlavorXY}}}
//    }))
func WithPlugins(listPlugins func(local *FlavorVppLocal) []*core.NamedPlugin) core.WithPluginsOpt {
	return &withPluginsOpt{listPlugins}
}

// FlavorVppLocal glues together multiple plugins to manage VPP and Linux
// configuration using the local client.
type FlavorVppLocal struct {
	*local.FlavorLocal
	LinuxLocalClient localclient.Plugin
	GoVPP            govppmux.GOVPPPlugin
	Linux            linuxplugin.Plugin
	VPP              defaultplugins.Plugin

	injected bool
}

// Inject sets inter-plugin references.
func (f *FlavorVppLocal) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()

	f.GoVPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("govpp")
	f.Linux.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("linuxplugin", local.WithConf())
	f.Linux.Watcher = &datasync.CompositeKVProtoWatcher{Adapters: []datasync.KeyValProtoWatcher{local_sync.Get()}}
	f.VPP.Watch = &datasync.CompositeKVProtoWatcher{Adapters: []datasync.KeyValProtoWatcher{local_sync.Get()}}
	f.VPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("default-plugins", local.WithConf())
	f.VPP.Deps.Linux = &f.Linux
	f.VPP.Deps.GoVppmux = &f.GoVPP

	return true
}

// Plugins combines all Plugins in the flavor to a list.
func (f *FlavorVppLocal) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}

// withPluginsOpt is return value of vppLocal.WithPlugins() utility
// to easily define new plugins for the agent based on FlavorVppLocal.
type withPluginsOpt struct {
	callback func(local *FlavorVppLocal) []*core.NamedPlugin
}

// OptionMarkerCore is just for marking implementation that it implements this interface
func (opt *withPluginsOpt) OptionMarkerCore() {}

// Plugins methods is here to implement core.WithPluginsOpt go interface
// <flavor> is a callback that uses flavor input for dependency injection
// for custom plugins (returned as NamedPlugin)
func (opt *withPluginsOpt) Plugins(flavors ...core.Flavor) []*core.NamedPlugin {
	for _, flavor := range flavors {
		if f, ok := flavor.(*FlavorVppLocal); ok {
			return opt.callback(f)
		}
	}

	panic("wrong usage of vppLocal.WithPlugin() for other than FlavorVppLocal")
}
