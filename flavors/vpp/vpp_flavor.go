package vpp

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/flavors/etcdkafka"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/flavors/redis"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
)

// Flavor glues together multiple plugins to translate ETCD configuration into VPP.
type Flavor struct {
	*local.FlavorLocal
	*rpc.FlavorRPC
	*etcdkafka.FlavorEtcdKafka
	*redis.FlavorRedis
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

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()

	if f.FlavorEtcdKafka == nil {
		f.FlavorEtcdKafka = &etcdkafka.FlavorEtcdKafka{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorEtcdKafka.Inject()
	
	if f.FlavorRPC == nil {
		f.FlavorRPC = &rpc.FlavorRPC{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorRPC.Inject()

	if f.FlavorRedis == nil {
		f.FlavorRedis = &redis.FlavorRedis{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorRedis.Inject()


	f.GoVPP.Deps.PluginInfraDeps = *f.FlavorEtcdKafka.FlavorLocal.InfraDeps("govpp")
	f.VPP.Deps.PluginInfraDeps = *f.FlavorEtcdKafka.FlavorLocal.InfraDeps("default-plugins")
	f.VPP.Deps.Publish = &f.FlavorEtcdKafka.ETCDDataSync
	f.VPP.Deps.PublishStatistics = datasync.CompositeKVProtoWriter{
		Adapters: []datasync.KeyProtoValWriter{&f.FlavorEtcdKafka.ETCDDataSync, &f.FlavorRedis.RedisDataSync},
	}
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
