package vpp

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/vpp-agent/govppmux"

	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/http"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/messaging/kafka"
	"github.com/ligato/cn-infra/servicelabel"

	"github.com/ligato/vpp-agent/defaultplugins"
)

// Flavour glues together multiple plugins to translate ETCD configuration into VPP.
type Flavour struct {
	injected     bool
	Logrus       logrus.Plugin
	HTTP         http.Plugin
	ServiceLabel servicelabel.Plugin
	Etcd         etcdv3.Plugin
	Kafka        kafka.Plugin
	Resync       resync.Plugin
	GoVPP        govppmux.GOVPPPlugin
	VPP          defaultplugins.Plugin
}

// Inject sets object references
func (f *Flavour) Inject() error {
	if f.injected {
		return nil
	}
	f.injected = true
	f.HTTP.LogFactory = &f.Logrus
	f.Etcd.LogFactory = &f.Logrus
	f.Etcd.ServiceLabel = &f.ServiceLabel
	f.Kafka.LogFactory = &f.Logrus
	f.Kafka.ServiceLabel = &f.ServiceLabel
	f.VPP.ServiceLabel = &f.ServiceLabel

	return nil
}

// Plugins combines Generic Plugins and Standard VPP Plugins + (their ETCD Connector/Adapter with RESYNC)
func (f *Flavour) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
