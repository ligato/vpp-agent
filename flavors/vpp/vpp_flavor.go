package vpp

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/flavors/etcdkafka"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
	"github.com/ligato/cn-infra/flavors/redis"
	"github.com/ligato/cn-infra/flavors/rpc"
)

// Flavor glues together multiple plugins to translate ETCD configuration into VPP.
type Flavor struct {
	*rpc.FlavorRPC
	*etcdkafka.FlavorEtcdKafka
	*redis.FlavorRedis
	Resync    resync.Plugin
	GoVPP     govppmux.GOVPPPlugin
	Linux     linuxplugin.Plugin
	VPP       defaultplugins.Plugin

	injected bool
}

// Inject sets object references
func (f *Flavor) Inject() error {
	if f.injected {
		return nil
	}
	f.injected = true

	if f.FlavorRPC == nil {
		f.FlavorRPC = &rpc.FlavorRPC{}
	}
	f.FlavorRPC.Inject()

	if f.FlavorEtcdKafka == nil {
		f.FlavorEtcdKafka = &etcdkafka.FlavorEtcdKafka{FlavorLocal: f.FlavorRPC.FlavorLocal}
	}
	f.FlavorEtcdKafka.Inject()

	if f.FlavorRedis == nil {
		f.FlavorRedis = &redis.FlavorRedis{}
	}
	f.FlavorRedis.Inject()



	// Aggregated transport
	compositePublisher := datasync.CompositeKVProtoWriter{
		Adapters: []datasync.KeyProtoValWriter{&f.FlavorEtcdKafka.ETCDDataSync, &f.FlavorRedis.RedisDataSync},
	}

	f.GoVPP.Deps.PluginInfraDeps = *f.FlavorEtcdKafka.FlavorLocal.InfraDeps("govpp")
	f.VPP.Deps.PluginInfraDeps = *f.FlavorEtcdKafka.FlavorLocal.InfraDeps("default-plugins")
	f.VPP.Deps.Publish = &f.FlavorEtcdKafka.ETCDDataSync
	f.VPP.Deps.PublishStatistics = compositePublisher
	f.VPP.Deps.Watch = &f.FlavorEtcdKafka.ETCDDataSync
	f.VPP.Deps.Kafka = &f.FlavorEtcdKafka.Kafka
	f.VPP.Deps.GoVppmux = &f.GoVPP
	f.VPP.Deps.Linux = &f.Linux
	f.VPP.Linux.Deps.Watcher = &f.FlavorEtcdKafka.ETCDDataSync

	return nil
}

// Plugins combines Generic Plugins and Standard VPP Plugins + (their ETCD Connector/Adapter with RESYNC)
func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
