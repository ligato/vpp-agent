package vpp

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/vpp-agent/flavors/local"
	"github.com/ligato/cn-infra/flavors/connectors"
)

// Flavor glues together multiple plugins to translate ETCD configuration into VPP.
type Flavor struct {
	*local.FlavorVppLocal
	*rpc.FlavorRPC
	*connectors.AllConnectorsFlavor

	injected bool
}

// Inject sets object references
func (f *Flavor) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	f.VPP.Deps.Publish = &f.AllConnectorsFlavor.ETCDDataSync
	f.VPP.Deps.PublishStatistics = &datasync.CompositeKVProtoWriter{Adapters: []datasync.KeyProtoValWriter{
			&f.AllConnectorsFlavor.ETCDDataSync, &f.AllConnectorsFlavor.RedisDataSync},
	}
	f.VPP.Deps.Watch = &f.AllConnectorsFlavor.ETCDDataSync
	f.VPP.Deps.Messaging = &f.AllConnectorsFlavor.Kafka
	f.Linux.Deps.Watcher = &f.AllConnectorsFlavor.ETCDDataSync

	return true
}

func (f *Flavor) injectEmbedded() {
	if f.FlavorVppLocal == nil {
		f.FlavorVppLocal = &local.FlavorVppLocal{}
	}
	f.FlavorVppLocal.Inject()

	if f.AllConnectorsFlavor == nil {
		f.AllConnectorsFlavor = &connectors.AllConnectorsFlavor{FlavorLocal: f.FlavorVppLocal.FlavorLocal}
	}
	f.FlavorRPC.Inject()

	if f.FlavorRPC == nil {
		f.FlavorRPC = &rpc.FlavorRPC{FlavorLocal: f.FlavorVppLocal.FlavorLocal}
	}
	f.FlavorRPC.Inject()
}

// Plugins combines Generic Plugins and Standard VPP Plugins + (their ETCD Connector/Adapter with RESYNC)
func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
