package main

import (
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/flavors/local"
)

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	local.PluginLogDeps                 // injected
	Watcher datasync.KeyValProtoWatcher // injected
	DB      keyval.KvProtoPlugin        // injected
}
