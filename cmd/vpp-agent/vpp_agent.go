package main

import (
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/datasync/msgsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linux"
	"github.com/ligato/vpp-agent/plugins/telemetry"
	"github.com/ligato/vpp-agent/plugins/vpp"
	"github.com/ligato/vpp-agent/plugins/vpp/rpc"
)

type VPPAgent struct {
	GoVPP *govppmux.Plugin
	Linux *linux.Plugin
	VPP   *vpp.Plugin

	IfStatePub      *msgsync.Plugin
	GRPCSvcPlugin   *rpc.GRPCSvcPlugin
	RESTAPIPlugin   *rest.Plugin
	TelemetryPlugin *telemetry.Plugin
}

func NewVppAgent() *VPPAgent {
	etcdDataSync := kvdbsync.NewPlugin(
		kvdbsync.UseDeps(func(deps *kvdbsync.Deps) {
			deps.KvPlugin = &etcd.DefaultPlugin
			deps.ResyncOrch = &resync.DefaultPlugin
		}),
	)
	watcher := &datasync.CompositeKVProtoWatcher{
		Adapters: []datasync.KeyValProtoWatcher{
			local.Get(),
			etcdDataSync,
		}}

	/*govpp := govppmux.NewPlugin(
		govppmux.UseDeps(govppmux.Deps{
			Resync: resync.DefaultPlugin,
		}),
	)*/
	govppPlugin := &govppmux.DefaultPlugin

	var linuxAPI vpp.LinuxpluginAPI
	vppPlugin := vpp.NewPlugin(
		vpp.UseDeps(func(deps *vpp.Deps) {
			deps.Linux = linuxAPI
			deps.GoVppmux = govppPlugin
			deps.Publish = etcdDataSync
			deps.Watch = watcher
			deps.DataSyncs = map[string]datasync.KeyProtoValWriter{
				"etcd": etcdDataSync,
			}
		}),
	)

	linuxPlugin := linux.NewPlugin(
		linux.UseDeps(func(deps *linux.Deps) {
			deps.VPP = vppPlugin
			deps.Watcher = watcher
		}),
	)
	linuxAPI = linuxPlugin

	return &VPPAgent{
		GoVPP: govppPlugin,
		Linux: linuxPlugin,
		VPP:   vppPlugin,
	}
}

func (VPPAgent) String() string {
	return "vpp-agent"
}

func (VPPAgent) Init() error {
	return nil
}

func (VPPAgent) AfterInit() error {
	// Manually run resync at the very end
	resync.DefaultPlugin.DoResync()
	return nil
}

func (VPPAgent) Close() error {
	return nil
}
