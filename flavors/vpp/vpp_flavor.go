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
	Linux  linuxplugin.Plugin
	VPP    defaultplugins.Plugin

	injected bool
}

// Inject sets object references
func (f *Flavor) Inject() error {
	if f.injected {
		return nil
	}

	f.Base.Inject()

	f.GoVPP.StatusCheck = &f.Base.Generic.StatusCheck
	f.GoVPP.LogFactory = &f.Base.Generic.Logrus
	f.VPP.ServiceLabel = &f.Base.Generic.ServiceLabel
	f.VPP.Kafka = &f.Base.Kafka
	f.VPP.GoVppmux = &f.GoVPP
	f.VPP.Linux = &f.Linux

	f.injected = true
	return nil
}

// Plugins combines Generic Plugins and Standard VPP Plugins + (their ETCD Connector/Adapter with RESYNC)
func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
