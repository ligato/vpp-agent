package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/flavors/connectors"
	"github.com/ligato/cn-infra/flavors/local"
)

// Deps lists dependencies of ExamplePlugin.
type Deps struct {
	local.PluginInfraDeps                 // injected
	Publisher datasync.KeyProtoValWriter  // injected - To write ETCD data
	Watcher   datasync.KeyValProtoWatcher // injected - To watch ETCD data
}

// ExampleFlavor is a set of plugins required for the datasync example.
type ExampleFlavor struct {
	// Local flavor to access the Infra (logger, service label, status check)
	*local.FlavorLocal
	// Resync orchestrator
	ResyncOrch resync.Plugin
	// Etcd plugin
	ETCD etcdv3.Plugin
	// Etcd sync which manages and injects connection
	ETCDDataSync kvdbsync.Plugin
	// Example plugin
	DatasyncExample ExamplePlugin
	// For example purposes, use channel when the example is finished
	closeChan *chan struct{}
}

// Inject sets inter-plugin references.
func (ef *ExampleFlavor) Inject() (allReadyInjected bool) {
	// Init local flavor
	if ef.FlavorLocal == nil {
		ef.FlavorLocal = &local.FlavorLocal{}
	}
	ef.FlavorLocal.Inject()

	// Init Resync, ETCD + ETCD sync
	ef.ResyncOrch.Deps.PluginLogDeps = *ef.FlavorLocal.LogDeps("resync-orch")
	ef.ETCD.Deps.PluginInfraDeps = *ef.InfraDeps("etcdv3",
		local.WithConf())
	connectors.InjectKVDBSync(&ef.ETCDDataSync, &ef.ETCD, ef.ETCD.PluginName, ef.FlavorLocal, &ef.ResyncOrch)

	// Inject infra + transport (publisher, watcher) to example plugin
	ef.DatasyncExample.PluginInfraDeps = *ef.FlavorLocal.InfraDeps("datasync-example")
	ef.DatasyncExample.Publisher = &ef.ETCDDataSync
	ef.DatasyncExample.Watcher = &ef.ETCDDataSync
	ef.DatasyncExample.closeChannel = ef.closeChan

	return true
}

// Plugins combines all plugins in the flavor into a slice.
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}
