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

// Package rpc defines flavor used for VPP agents managed using GPRC service.
package rpc

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"

	"github.com/ligato/cn-infra/datasync"
	local_sync "github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/vpp-agent/clientv1/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
)

// checkImplemensPlugin is used to let compiler check if
// a particular plugin implements go interface core.Plugin
// (see following Inject() method).
// This construct is used because the following method Plugins()
// uses reflection rather than enumerating all field again.
var checkImplemensPlugin core.Plugin

// NewAgent returns a new instance of the Agent with plugins.
// It is an alias for core.NewAgent() to implicit use of the FlavorVppLocal
func NewAgent(opts ...core.Option) *core.Agent {
	return core.NewAgent(&FlavorVppRPC{}, opts...)
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
func WithPlugins(listPlugins func(local *FlavorVppRPC) []*core.NamedPlugin) core.WithPluginsOpt {
	return &withPluginsOpt{listPlugins}
}

// FlavorVppRPC glues together multiple plugins to mange VPP and linux interfaces configuration using
// GRPC service.
type FlavorVppRPC struct {
	*local.FlavorLocal
	*rpc.FlavorRPC
	LinuxLocalClient localclient.Plugin
	GoVPP            govppmux.GOVPPPlugin
	Linux            linuxplugin.Plugin
	VPP              defaultplugins.Plugin

	GRPCSvcPlugin GRPCSvcPlugin

	injected bool
}

// Inject sets object references.
func (f *FlavorVppRPC) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()
	if f.FlavorRPC == nil {
		f.FlavorRPC = &rpc.FlavorRPC{}
	}
	f.FlavorRPC.Inject()

	f.GoVPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("govpp")
	f.Linux.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("linuxplugin", local.WithConf())
	f.Linux.Watcher = &datasync.CompositeKVProtoWatcher{Adapters: []datasync.KeyValProtoWatcher{local_sync.Get()}}
	f.VPP.Watch = &datasync.CompositeKVProtoWatcher{Adapters: []datasync.KeyValProtoWatcher{local_sync.Get()}}
	f.VPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("default-plugins", local.WithConf())
	f.VPP.Deps.Linux = &f.Linux
	f.VPP.Deps.GoVppmux = &f.GoVPP

	f.GRPCSvcPlugin.Deps.PluginLogDeps = *f.LogDeps("vpp-grpc-svc")
	f.GRPCSvcPlugin.Deps.GRPC = &f.FlavorRPC.GRPC
	checkImplemensPlugin = &f.GRPCSvcPlugin

	return true
}

// Plugins combine Generic Plugins and Standard VPP Plugins.
func (f *FlavorVppRPC) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}

// withPluginsOpt is return value of vppLocal.WithPlugins() utility
// to easily define new plugins for the agent based on FlavorVppRPC.
type withPluginsOpt struct {
	callback func(local *FlavorVppRPC) []*core.NamedPlugin
}

// OptionMarkerCore marks that the implementation implements this interface.
func (opt *withPluginsOpt) OptionMarkerCore() {}

// Plugins methods is here to implement core.WithPluginsOpt go interface
// <flavor> is a callback that uses flavor input for dependency injection
// for custom plugins (returned as NamedPlugin)
func (opt *withPluginsOpt) Plugins(flavors ...core.Flavor) []*core.NamedPlugin {
	for _, flavor := range flavors {
		if f, ok := flavor.(*FlavorVppRPC); ok {
			return opt.callback(f)
		}
	}

	panic("wrong usage of vppRpc.WithPlugin() for other than FlavorVppRPC")
}
