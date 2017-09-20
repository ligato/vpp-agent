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
	"github.com/namsral/flag"
)

const kafkaIfStateTopic = "if_state" // IfStatePub topic where interface state changes are published.

// DefaultPluginsConfFlag used as flag name (see implementation in declareFlags())
// It is used to load configuration of MTU for defaultplugins.
const DefaultPluginsConfFlag = "default-plugins-config"

// DefaultPluginsConf is default (flag value) - filename for the configuration.
const DefaultPluginsConf = "defaultplugins.conf"

// DefaultPluginsConfUsage used as flag usage (see implementation in declareFlags())
const DefaultPluginsConfUsage = "Location of the MTU configuration file; also set via 'MTU_CONFIG' env variable."

// IfStatePubConfFlag used as flag name (see implementation in declareFlags())
// It is used to load configuration of Cassandra client plugin.
// This flag name is calculated from the name of the plugin.
const IfStatePubConfFlag = "ifstate_pub-config"

// IfStatePubConf  is default (flag value) - filename for the configuration.
const IfStatePubConf = "ifstate-pub.conf"

// IfStatePubConfUsage used as flag usage (see implementation in declareFlags())
const IfStatePubConfUsage = "Location of the interface state publish configuration file; also set via 'IFSTATE_PUB_CONFIG' env variable."

// GoVPPConfFlag used as flag name (see implementation in declareFlags())
// It is used to load configuration of GoVPP client plugin.
// This flag name is calculated from the name of the plugin.
const GoVPPConfFlag = "govpp-config"

// GoVPPConf  is default (flag value) - filename for the configuration.
const GoVPPConf = "govpp.conf"

// GoVPPConfUsage used as flag usage (see implementation in declareFlags())
const GoVPPConfUsage = "Location of the GoVPP configuration file; also set via 'GOVPP_CONFIG' env variable."

// Flavor glues together multiple plugins to translate ETCD configuration into VPP.
type Flavor struct {
	*local.FlavorLocal
	*connectors.AllConnectorsFlavor // connectors have to be started before vpp flavor
	*rpc.FlavorRPC

	//this can be reused later even for Linux plugin
	//it has its own configuration
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

	declareFlags()
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
	f.IfStatePub.PluginInfraDeps = *f.InfraDeps("ifstate-pub")
	// If needed provide configuration using ifstate-pub-config.
	// Set default configuration, it is overridable using ifstate-pub-config
	// Intent not putting this configuration to vpp plugin is that
	// this way it is reusable even for Linux plugin.
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

func declareFlags() {
	flag.String(DefaultPluginsConfFlag, DefaultPluginsConf, DefaultPluginsConfUsage)
	flag.String(IfStatePubConfFlag, IfStatePubConf, IfStatePubConfUsage)
	flag.String(GoVPPConfFlag, GoVPPConf, GoVPPConfUsage)
}
