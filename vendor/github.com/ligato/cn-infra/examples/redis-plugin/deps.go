package main

import (
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/redis"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/connectors"
)

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	local.PluginLogDeps // injected
}

// ExampleFlavor is a set of plugins required for the redis example.
type ExampleFlavor struct {
	// Local flavor to access to Infra (logger, service label, status check)
	*local.FlavorLocal
	// Redis plugin
	Redis         redis.Plugin
	RedisDataSync kvdbsync.Plugin
	ResyncOrch resync.Plugin
	// Example plugin
	RedisExample ExamplePlugin
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
	ef.Redis.Deps.PluginInfraDeps = *ef.InfraDeps("redis", local.WithConf())
	ef.ResyncOrch.Deps.PluginLogDeps = *ef.LogDeps("redis-resync")
	connectors.InjectKVDBSync(&ef.RedisDataSync, &ef.Redis, ef.Redis.PluginName, ef.FlavorLocal, &ef.ResyncOrch)
	ef.RedisExample.Deps.PluginLogDeps = *ef.FlavorLocal.LogDeps("redis-example")
	ef.RedisExample.closeChannel = ef.closeChan

	return true
}

// Plugins combines all Plugins in flavor to the list
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}