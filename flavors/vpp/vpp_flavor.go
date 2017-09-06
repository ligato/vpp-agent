package vpp

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/cn-infra/flavors/local"
	vpplocal "github.com/ligato/vpp-agent/flavors/local"
	"github.com/ligato/cn-infra/flavors/connectors"
)

// Flavor glues together multiple plugins to translate ETCD configuration into VPP.
type Flavor struct {
	*local.FlavorLocal
	*vpplocal.FlavorVppLocal
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

	f.injectEmbedded()

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
	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	if f.FlavorVppLocal == nil {
		f.FlavorVppLocal = &vpplocal.FlavorVppLocal{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorVppLocal.Inject()
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
