package main

import (
	"strings"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
	"golang.org/x/net/context"
)


// ETCD flag to load config
func init() {
	flag.String("etcdv3-config", "etcd.conf",
		"Location of the Etcd configuration file")
}

// ExampleFlavor is a set of plugins required for the datasync example.
type ExampleFlavor struct {
	// Local flavor to access to Infra (logger, service label, status check)
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

// Inject sets object references
func (ef *ExampleFlavor) Inject() (allReadyInjected bool) {
	// Init local flavor
	if ef.FlavorLocal == nil {
		ef.FlavorLocal = &local.FlavorLocal{}
	}
	ef.FlavorLocal.Inject()
	// Init Resync, ETCD + ETCD sync
	ef.ResyncOrch.Deps.PluginLogDeps = *ef.FlavorLocal.LogDeps("resync-orch")
	ef.ETCD.Deps.PluginInfraDeps = *ef.FlavorLocal.InfraDeps("etcdv3")
	ef.ETCDDataSync.Deps.PluginLogDeps = *ef.FlavorLocal.LogDeps("etcdv3-datasync")
	ef.ETCDDataSync.KvPlugin = &ef.ETCD
	ef.ETCDDataSync.ResyncOrch = &ef.ResyncOrch
	ef.ETCDDataSync.ServiceLabel = &ef.FlavorLocal.ServiceLabel
	// Inject infra + transport (publisher, watcher) to example plugin
	ef.DatasyncExample.PluginInfraDeps = *ef.FlavorLocal.InfraDeps("datasync-example")
	ef.DatasyncExample.Publisher = &ef.ETCDDataSync
	ef.DatasyncExample.Watcher = &ef.ETCDDataSync
	ef.DatasyncExample.closeChannel = ef.closeChan

	return true
}

// Plugins combines all Plugins in flavor to the list
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}

