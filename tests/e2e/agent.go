package e2e

import (
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logmanager"
	logregistry "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/cn-infra/db/keyval"

	"github.com/ligato/vpp-agent/client"
	"github.com/ligato/vpp-agent/cmd/vpp-agent/app"
	"github.com/ligato/vpp-agent/plugins/configurator"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/netalloc"
	"github.com/ligato/vpp-agent/plugins/orchestrator"

	"github.com/ligato/vpp-agent/plugins/kvscheduler"
	linux_ifplugin "github.com/ligato/vpp-agent/plugins/linux/ifplugin"
	iptables "github.com/ligato/vpp-agent/plugins/linux/iptablesplugin"
	linux_l3plugin "github.com/ligato/vpp-agent/plugins/linux/l3plugin"
	linux_nsplugin "github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin"
	vpp_ifplugin "github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/stnplugin"
)

type vppAgent struct {
	LogManager *logmanager.Plugin
	GoVppMux   *govppmux.Plugin

	app.VPP
	app.Linux
	Netalloc *netalloc.Plugin

	KvScheduler  *kvscheduler.Scheduler
	Orchestrator *orchestrator.Plugin
	Configurator *configurator.Plugin

	inSync chan struct{}
}

func (a *vppAgent) Init() error {
	return nil
}

func (a *vppAgent) AfterInit() error {
	orchestrator.DefaultPlugin.InitialSync()
	close(a.inSync)
	return nil
}

func (a *vppAgent) Close() error {
	return nil
}

func (a *vppAgent) String() string {
	return "VPPAgent"
}

type txnFactory struct {
	registry *syncbase.Registry
}

func (p *txnFactory) NewTxn(resync bool) keyval.ProtoTxn {
	if resync {
		return local.NewProtoTxn(p.registry.PropagateResync)
	}
	return local.NewProtoTxn(p.registry.PropagateChanges)
}

// setupAgent creates all the agent plugins (that are included in the end-to-end
// testing) and injects dependencies.
func setupAgent(inSync chan struct{}) *vppAgent {
	resetDefaultPlugins()

	vpp := app.DefaultVPP()
	linux := app.DefaultLinux()

	// connect IfPlugins for Linux & VPP
	linux.IfPlugin.VppIfPlugin = vpp.IfPlugin
	vpp.IfPlugin.LinuxIfPlugin = linux.IfPlugin
	vpp.IfPlugin.NsPlugin = linux.NSPlugin

	return &vppAgent{
		LogManager:   &logmanager.DefaultPlugin,
		GoVppMux:     &govppmux.DefaultPlugin,
		VPP:          vpp,
		Linux:        linux,
		Netalloc:     &netalloc.DefaultPlugin,
		KvScheduler:  &kvscheduler.DefaultPlugin,
		Orchestrator: &orchestrator.DefaultPlugin,
		Configurator: &configurator.DefaultPlugin,
		inSync:       inSync,
	}
}

// resetDefaultPlugins re-creates all the default plugins to ensure that
// agent restarts into a clean state. Alternatively, we could avoid default
// instances and create all the plugins using NewPlugin() methods instead, but
// that would be more boilerplate...
func resetDefaultPlugins() {
	// common
	local.DefaultRegistry = syncbase.NewRegistry()
	client.LocalClient = client.NewClient(&txnFactory{local.DefaultRegistry})
	logging.DefaultRegistry = logregistry.NewLogRegistry()
	logmanager.DefaultPlugin = *logmanager.NewPlugin(func(p *logmanager.Plugin) {
		p.Deps.HTTP = nil
	})
	resync.DefaultPlugin = *resync.NewPlugin()
	grpc.DefaultPlugin = *grpc.NewPlugin()
	govppmux.DefaultPlugin = *govppmux.NewPlugin(func(p *govppmux.Plugin) {
		p.Deps.HTTPHandlers = nil
	})
	kvscheduler.DefaultPlugin = *kvscheduler.NewPlugin(func(p *kvscheduler.Scheduler) {
		p.Deps.HTTPHandlers = nil
	})
	orchestrator.DefaultPlugin = *orchestrator.NewPlugin()
	netalloc.DefaultPlugin = *netalloc.NewPlugin()

	// VPP
	vpp_ifplugin.DefaultPlugin = *vpp_ifplugin.NewPlugin()
	aclplugin.DefaultPlugin = *aclplugin.NewPlugin()
	abfplugin.DefaultPlugin = *abfplugin.NewPlugin()
	ipsecplugin.DefaultPlugin = *ipsecplugin.NewPlugin()
	l2plugin.DefaultPlugin = *l2plugin.NewPlugin()
	l3plugin.DefaultPlugin = *l3plugin.NewPlugin()
	natplugin.DefaultPlugin = *natplugin.NewPlugin()
	puntplugin.DefaultPlugin = *puntplugin.NewPlugin()
	stnplugin.DefaultPlugin = *stnplugin.NewPlugin()
	srplugin.DefaultPlugin = *srplugin.NewPlugin()

	// Linux
	linux_nsplugin.DefaultPlugin = *linux_nsplugin.NewPlugin()
	linux_ifplugin.DefaultPlugin = *linux_ifplugin.NewPlugin()
	linux_l3plugin.DefaultPlugin = *linux_l3plugin.NewPlugin()
	iptables.DefaultPlugin = *iptables.NewPlugin()

	// Configurator
	configurator.DefaultPlugin = *configurator.NewPlugin()
}
