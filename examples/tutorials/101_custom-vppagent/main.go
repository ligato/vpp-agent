package main

import (
	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync/local"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/db/keyval/etcd"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logmanager"

	"go.ligato.io/vpp-agent/v3/cmd/vpp-agent/app"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
)

func main() {
	// Create an instance of our custom agent.
	p := NewCustomAgent()

	// Create new agent with our plugin instance.
	a := agent.NewAgent(agent.AllPlugins(p))

	// Run starts the agent with plugins, waits until shutdown
	// and then stops the agent and plugins.
	if err := a.Run(); err != nil {
		logging.DefaultLogger.Fatalln(err)
	}
}

// CustomAgent represents our plugin.
type CustomAgent struct {
	LogManager *logmanager.Plugin

	app.VPP
	app.Linux

	Orchestrator *orchestrator.Plugin
	KVDBSync     *kvdbsync.Plugin
	Resync       *resync.Plugin
}

// NewCustomAgent returns new CustomAgent instance.
func NewCustomAgent() *CustomAgent {
	p := &CustomAgent{
		LogManager: &logmanager.DefaultPlugin,
	}

	etcdDataSync := kvdbsync.NewPlugin(kvdbsync.UseKV(&etcd.DefaultPlugin))
	p.KVDBSync = etcdDataSync
	p.Resync = &resync.DefaultPlugin

	p.VPP = app.DefaultVPP()
	p.Linux = app.DefaultLinux()

	// connect IfPlugins for Linux & VPP
	p.Linux.IfPlugin.VppIfPlugin = p.VPP.IfPlugin
	p.VPP.IfPlugin.LinuxIfPlugin = p.Linux.IfPlugin
	p.VPP.IfPlugin.NsPlugin = p.Linux.NSPlugin

	// Set watcher for KVScheduler.
	watchers := datasync.KVProtoWatchers{
		local.DefaultRegistry,
		etcdDataSync,
	}
	orch := &orchestrator.DefaultPlugin
	orch.Watcher = watchers
	p.Orchestrator = orch

	return p
}

// String is used to identify the plugin by giving it name.
func (p *CustomAgent) String() string {
	return "CustomAgent"
}

// Init is executed on agent initialization.
func (p *CustomAgent) Init() error {
	logging.DefaultLogger.Info("Initializing CustomAgent")
	return nil
}

// AfterInit is executed after initialization of all plugins. It's optional
// and used for executing operations that require plugins to be initalized.
func (p *CustomAgent) AfterInit() error {
	p.Resync.DoResync()
	logging.DefaultLogger.Info("CustomAgent is Ready")
	return nil
}

// Close is executed on agent shutdown.
func (p *CustomAgent) Close() error {
	logging.DefaultLogger.Info("Shutting down CustomAgent")
	return nil
}
