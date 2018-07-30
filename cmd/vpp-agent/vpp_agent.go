package main

import (
	"sync"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/datasync/msgsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/logging/logmanager"
	"github.com/ligato/cn-infra/messaging/kafka"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linux"
	"github.com/ligato/vpp-agent/plugins/telemetry"
	"github.com/ligato/vpp-agent/plugins/vpp"
	"github.com/ligato/vpp-agent/plugins/vpp/rpc"
)

type VPPAgent struct {
	LogManager *logmanager.Plugin

	GoVPP *govppmux.Plugin
	Linux *linux.Plugin
	VPP   *vpp.Plugin

	IfStatePub *msgsync.Plugin

	GRPCSvcPlugin   *rpc.GRPCSvcPlugin
	RESTAPIPlugin   *rest.Plugin
	TelemetryPlugin *telemetry.Plugin
}

func NewVppAgent() *VPPAgent {
	a := &VPPAgent{
		LogManager: &logmanager.DefaultPlugin,
		GoVPP:      &govppmux.DefaultPlugin,
	}

	ifStatePub := msgsync.NewPlugin(
		msgsync.UseDeps(func(deps *msgsync.Deps) {
			deps.Messaging = &kafka.DefaultPlugin
		}), msgsync.UseConf(msgsync.Cfg{
			Topic: "if_state",
		}),
	)
	a.IfStatePub = ifStatePub

	dataSync := kvdbsync.NewPlugin(kvdbsync.UseDeps(func(deps *kvdbsync.Deps) {
		deps.KvPlugin = &etcd.DefaultPlugin
		deps.ResyncOrch = &resync.DefaultPlugin
	}),
	)
	watcher := &datasync.CompositeKVProtoWatcher{Adapters: []datasync.KeyValProtoWatcher{
		local.Get(),
		dataSync,
	}}

	var watchEventsMutex sync.Mutex
	a.VPP = vpp.NewPlugin(vpp.UseDeps(func(deps *vpp.Deps) {
		deps.GoVppmux = a.GoVPP
		deps.Publish = dataSync
		deps.Watch = watcher
		deps.DataSyncs = map[string]datasync.KeyProtoValWriter{
			"etcd": dataSync,
		}
		deps.WatchEventsMutex = &watchEventsMutex
		deps.IfStatePub = a.IfStatePub
	}))
	a.Linux = linux.NewPlugin(linux.UseDeps(func(deps *linux.Deps) {
		deps.VPP = a.VPP
		deps.Watcher = watcher
		deps.WatchEventsMutex = &watchEventsMutex
	}))
	a.VPP.Deps.Linux = a.Linux

	return a
}

func (VPPAgent) String() string {
	return "VPPAgent"
}

func (VPPAgent) Init() error {
	return nil
}

func (VPPAgent) AfterInit() error {
	// manually start resync after all plugins started
	resync.DefaultPlugin.DoResync()
	return nil
}

func (VPPAgent) Close() error {
	return nil
}
