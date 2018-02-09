// Package vpp defines the standard flavor used for full-featured VPP agents.
package vpp

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/msgsync"
	"github.com/ligato/cn-infra/flavors/connectors"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
	"github.com/ligato/vpp-agent/plugins/restplugin"
)

// kafkaIfStateTopic is the topic where interface state changes are published.
const kafkaIfStateTopic = "if_state"

// NewAgent returns a new instance of the Agent with plugins.
// It is an alias for core.NewAgent() to implicit use of the FlavorVppLocal
func NewAgent(opts ...core.Option) *core.Agent {
	return core.NewAgent(&Flavor{}, opts...)
}

// WithPlugins for adding custom plugins to SFC Controller.
// <listPlugins> is a callback that uses flavor input to
// inject dependencies for custom plugins that are in output.
//
// Example:
//
//    NewAgent(vppFlavor.WithPlugins(func(flavor) {
// 	       return []*core.NamedPlugin{{"my-plugin", &MyPlugin{DependencyXY: &flavor.FlavorXY}}}
//    }))
func WithPlugins(listPlugins func(local *Flavor) []*core.NamedPlugin) core.WithPluginsOpt {
	return &withPluginsOpt{listPlugins}
}

// Flavor glues together multiple plugins to build a full-featured VPP agent.
type Flavor struct {
	*local.FlavorLocal
	*connectors.AllConnectorsFlavor // connectors have to be started before vpp flavor
	*rpc.FlavorRPC

	// This can be reused later even for the Linux plugin,
	// it has its own configuration.
	IfStatePub msgsync.PubPlugin

	GoVPP govppmux.GOVPPPlugin
	Linux linuxplugin.Plugin
	VPP   defaultplugins.Plugin

	RESTAPIPlugin restplugin.RESTAPIPlugin

	injected bool
}

// Inject sets inter-plugin references.
func (f *Flavor) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	f.injectEmbedded()

	f.GoVPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("govpp", local.WithConf())
	f.VPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("default-plugins", local.WithConf())
	f.VPP.Deps.Linux = &f.Linux
	f.VPP.Deps.GoVppmux = &f.GoVPP

	f.VPP.Deps.Publish = &f.AllConnectorsFlavor.ETCDDataSync

	/* note: now configurable with `status-publishers` in defaultplugins
		f.VPP.Deps.PublishStatistics = &datasync.CompositeKVProtoWriter{Adapters: []datasync.KeyProtoValWriter{
		&f.AllConnectorsFlavor.ETCDDataSync, &f.AllConnectorsFlavor.RedisDataSync},
	}*/
	f.VPP.Deps.DataSyncs = map[string]datasync.KeyProtoValWriter{
		"etcd":  &f.AllConnectorsFlavor.ETCDDataSync,
		"redis": &f.AllConnectorsFlavor.RedisDataSync,
	}

	f.IfStatePub.Messaging = &f.Kafka
	f.IfStatePub.PluginInfraDeps = *f.InfraDeps("ifstate-pub")
	// If needed, provide configuration using ifstate-pub-config.
	// Set default configuration; it is overridable using ifstate-pub-config.
	// Intent of not putting this configuration into the vpp plugin is that
	// this way it is reusable even for the Linux plugin.
	f.IfStatePub.Cfg.Topic = kafkaIfStateTopic

	f.VPP.Deps.IfStatePub = &f.IfStatePub
	f.VPP.Deps.Watch = &f.AllConnectorsFlavor.ETCDDataSync

	f.Linux.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("linuxplugin", local.WithConf())
	f.Linux.Deps.Watcher = &f.AllConnectorsFlavor.ETCDDataSync

	f.RESTAPIPlugin.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("restapiplugin")
	f.RESTAPIPlugin.Deps.HTTPHandlers = &f.FlavorRPC.HTTP
	f.RESTAPIPlugin.Deps.GoVppmux = &f.GoVPP

	return true
}

func (f *Flavor) injectEmbedded() {
	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()
	if f.FlavorRPC == nil {
		f.FlavorRPC = &rpc.FlavorRPC{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorRPC.Inject()
	if f.AllConnectorsFlavor == nil {
		f.AllConnectorsFlavor = &connectors.AllConnectorsFlavor{FlavorLocal: f.FlavorLocal}
	}
	f.AllConnectorsFlavor.Inject()
}

// Plugins combine all Plugins in the flavor to a list.
func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}

// withPluginsOpt is return value of vppLocal.WithPlugins() utility
// to easily define new plugins for the agent based on Flavor.
type withPluginsOpt struct {
	callback func(local *Flavor) []*core.NamedPlugin
}

// OptionMarkerCore is just for marking implementation that it implements this interface
func (opt *withPluginsOpt) OptionMarkerCore() {}

// Plugins methods is here to implement core.WithPluginsOpt go interface
// <flavor> is a callback that uses flavor input for dependency injection
// for custom plugins (returned as NamedPlugin)
func (opt *withPluginsOpt) Plugins(flavors ...core.Flavor) []*core.NamedPlugin {
	for _, flavor := range flavors {
		if f, ok := flavor.(*Flavor); ok {
			return opt.callback(f)
		}
	}

	panic("wrong usage of vpp.WithPlugin() for other than Flavor")
}
