package vpp

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/msgsync"
	"github.com/ligato/cn-infra/flavors/connectors"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

const kafkaIfStateTopic = "if_state" // IfStatePub topic where interface state changes are published.

// Flavor glues together multiple plugins to translate ETCD configuration into VPP.
type Flavor struct {
	*local.FlavorLocal
	*connectors.AllConnectorsFlavor // connectors have to be started before vpp flavor
	*rpc.FlavorRPC
	IfStatePub msgsync.PubPlugin

	GoVPP govppmux.GOVPPPlugin
	Linux linuxplugin.Plugin
	VPP   defaultplugins.Plugin

	injected bool
}

// Inject sets object references
func (f *Flavor) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	f.injectEmbedded()

	f.GoVPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("govpp")
	f.VPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("default-plugins")
	f.VPP.Deps.Linux = &f.Linux
	f.VPP.Deps.GoVppmux = &f.GoVPP

	f.VPP.Deps.Publish = &f.AllConnectorsFlavor.ETCDDataSync
	f.VPP.Deps.PublishStatistics = &datasync.CompositeKVProtoWriter{Adapters: []datasync.KeyProtoValWriter{
		&f.AllConnectorsFlavor.ETCDDataSync, &f.AllConnectorsFlavor.RedisDataSync},
	}

	f.IfStatePub.Messaging = &f.Kafka
	f.IfStatePub.PluginInfraDeps = *f.InfraDeps("messaging-sync")
	f.IfStatePub.Cfg.Topic = kafkaIfStateTopic

	f.VPP.Deps.IfStatePub = &f.IfStatePub
	f.VPP.Deps.Watch = &f.AllConnectorsFlavor.ETCDDataSync

	f.Linux.Deps.Watcher = &f.AllConnectorsFlavor.ETCDDataSync

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

// Plugins combines Generic Plugins and Standard VPP Plugins + (their ETCD Connector/Adapter with RESYNC)
func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
