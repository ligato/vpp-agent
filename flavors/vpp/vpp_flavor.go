package vpp

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/flavors/etcdkafka"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
)

// Flavor glues together multiple plugins to translate ETCD configuration into VPP.
type Flavor struct {
	Base   etcdkafka.Flavor
	Resync resync.Plugin
	GoVPP  govppmux.GOVPPPlugin
	VPP    defaultplugins.Plugin
	Linux  linuxplugin.Plugin

	injected bool
}

// Inject sets object references
func (f *Flavor) Inject() error {
	if f.injected {
		return nil
	}

	f.Base.Inject()

	f.GoVPP.Deps.PluginInfraDeps = *f.Base.FlavorLocal.InfraDeps("GOVPP")
	f.VPP.Deps.PluginInfraDeps = *f.Base.FlavorLocal.InfraDeps("default-plugins")
	f.VPP.Transport = &f.Base.ETCDDataSync
	f.VPP.Watch = &f.Base.ETCDDataSync
	f.VPP.Kafka = &f.Base.Kafka
	f.VPP.GoVppmux = &f.GoVPP
	f.VPP.Linux = &f.Linux
	f.VPP.Linux.Watcher = &f.Base.ETCDDataSync

	f.injected = true
	return nil
}

// Plugins combines Generic Plugins and Standard VPP Plugins + (their ETCD Connector/Adapter with RESYNC)
func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
