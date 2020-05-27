//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vppagent

import (
	"context"
	"fmt"
	"log"

	"github.com/google/wire"
	cninfra "go.ligato.io/cn-infra/v2"
	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync/local"
	"go.ligato.io/cn-infra/v2/health/probe"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logmanager"
	"go.ligato.io/cn-infra/v2/rpc/prometheus"

	"go.ligato.io/vpp-agent/v3/plugins/configurator"
	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	linux_ifplugin "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	linux_iptablesplugin "go.ligato.io/vpp-agent/v3/plugins/linux/iptablesplugin"
	linux_l3plugin "go.ligato.io/vpp-agent/v3/plugins/linux/l3plugin"
	linux_nsplugin "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	"go.ligato.io/vpp-agent/v3/plugins/restapi"
	"go.ligato.io/vpp-agent/v3/plugins/telemetry"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/abfplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/puntplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin"
)

type Agent struct {
	cninfra.Base
	cninfra.Server

	LogManager *logmanager.Plugin

	Probe      *probe.Plugin
	Prometheus *prometheus.Plugin

	KVScheduler  *kvscheduler.Scheduler
	Orchestrator *orchestrator.Plugin

	// Dataplane drivers
	Netalloc *netalloc.Plugin
	VPP
	VPPClient *govppmux.Plugin
	Linux

	Configurator *configurator.Plugin
	RestAPI      *restapi.Plugin
	Telemetry    *telemetry.Plugin

	ctx    context.Context `wire:"-"`
	cancel func()          `wire:"-"`
}

func New(conf config.Config) Agent {
	ctx := context.Background()

	core, cancelCore, err := cninfra.InjectDefaultBase(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	server, cancelServer, err := cninfra.InjectDefaultServer(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	agent, cancelAgent, err := InjectAgent(ctx, conf, core, server)
	if err != nil {
		log.Fatal(err)
	}

	agent.cancel = func() {
		cancelAgent()
		cancelServer()
		cancelCore()
	}
	agent.ctx = ctx

	return agent
}

func (agent Agent) Start() error {
	if err := agent.Base.Start(agent.ctx); err != nil {
		return err
	}
	if err := agent.Server.Start(agent.ctx); err != nil {
		return err
	}

	logging.Debugf("Agent Run()")

	// TODO: replace runAfterInit hack that calls AfterInit() with a call to
	//  a properly named method that describes the actual action.

	// for example: this call should be replaced with call to a method(s)
	// that represent what the AfterInit does, e.g.
	// 	agent.VPPClient.StartConnectionHandler()
	runAfterInit(agent.VPPClient)

	// TODO: replace these hack calls with more generic approach of starting
	//  dataplane components in a way that new ones can be added and
	//  any existing ones can be removed or replaced with custom ones.
	runAfterInit(agent.Netalloc)
	forEachField(agent.Linux, runAfterInit)
	forEachField(agent.VPP, runAfterInit)

	if err := agent.Orchestrator.StartWatching(); err != nil {
		return err
	}

	agent.Prometheus.RegisterHandlers()
	agent.Telemetry.StartPeriodicUpdates()
	if err := agent.Probe.RegisterProbes(); err != nil {
		return err
	}

	logging.Info("--- Agent is RUNNING! ---")

	return nil
}

// Wait for a signal or cancellation
func (agent Agent) Wait() {
	select {
	case s := <-waitForSignal():
		logging.Info("received signal:", s)
	case <-agent.ctx.Done():
		logging.Info("agent context canceled")
	}
}

func (agent Agent) Stop() error {
	if agent.cancel == nil {
		return fmt.Errorf("agent not started")
	}

	agent.cancel()
	agent.cancel = nil

	return nil
}

// Run helper method to starts the agent, waits and stops the agent.
func (agent Agent) Run() {

	if err := agent.Start(); err != nil {
		logging.Fatal("agent failed to start:", err)
	}

	agent.Wait()

	if err := agent.Stop(); err != nil {
		logging.Fatal("agent failed to stop:", err)
	}
}

type VPP struct {
	ABF       *abfplugin.ABFPlugin
	ACL       *aclplugin.ACLPlugin
	Interface *ifplugin.IfPlugin
	IPSec     *ipsecplugin.IPSecPlugin
	L2        *l2plugin.L2Plugin
	L3        *l3plugin.L3Plugin
	NAT       *natplugin.NATPlugin
	Punt      *puntplugin.PuntPlugin
	STN       *stnplugin.STNPlugin
	SR        *srplugin.SRPlugin
}

type Linux struct {
	Interface *linux_ifplugin.IfPlugin
	L3        *linux_l3plugin.L3Plugin
	NS        *linux_nsplugin.NsPlugin
	IPTables  *linux_iptablesplugin.IPTablesPlugin
}

var WireDefaultNetAlloc = wire.NewSet(
	netalloc.Wire,
)

var WireDefaultVPP = wire.NewSet(
	abfplugin.Wire,
	aclplugin.Wire,
	ifplugin.Wire,
	l2plugin.Wire,
	l3plugin.Wire,
	ipsecplugin.Wire,
	natplugin.Wire,
	puntplugin.Wire,
	srplugin.Wire,
	stnplugin.Wire,
	wire.NewSet(wire.Struct(new(VPP), "*")),
)

var WireDefaultLinux = wire.NewSet(
	linux_ifplugin.Wire,
	linux_nsplugin.Wire,
	linux_l3plugin.Wire,
	linux_iptablesplugin.Wire,
	wire.NewSet(wire.Struct(new(Linux), "*")),
)

func TelemetryProvider(deps telemetry.Deps) (*telemetry.Plugin, func(), error) {
	p := &telemetry.Plugin{Deps: deps}
	p.SetName("telemetry")
	p.Log = logging.ForPlugin("telemetry")
	p.Cfg = config.ForPlugin("telemetry")
	cancel := func() {
		if err := p.Close(); err != nil {
			p.Log.Error(err)
		}
	}
	return p, cancel, p.Init()
}

var WireTelemetry = wire.NewSet(
	TelemetryProvider,
	wire.Struct(new(telemetry.Deps), "*"),
)

func RestapiProvider(deps restapi.Deps) (*restapi.Plugin, error) {
	p := &restapi.Plugin{Deps: deps}
	p.SetName("restapi")
	p.Log = logging.ForPlugin("restapi")
	return p, p.Init()
}

var WireRestAPI = wire.NewSet(
	RestapiProvider,
	wire.Struct(new(restapi.Deps), "*"),
)

func ConfiguratorProvider(deps configurator.Deps) (*configurator.Plugin, error) {
	p := &configurator.Plugin{Deps: deps}
	p.SetName("configurator")
	p.Log = logging.ForPlugin("configurator")
	return p, p.Init()
}

var WireConfigurator = wire.NewSet(
	ConfiguratorProvider,
	wire.Struct(new(configurator.Deps), "*"),
)

func OrchestratorProvider(deps orchestrator.Deps) (*orchestrator.Plugin, error) {
	p := &orchestrator.Plugin{Deps: deps}
	p.SetName("orchestrator")
	p.Log = logging.ForPlugin("orchestrator")
	return p, p.Init()
}

var WireOrchestrator = wire.NewSet(
	OrchestratorProvider,
	wire.Struct(new(orchestrator.Deps), "GRPC", "KVScheduler", "Watcher", "StatusPublisher"),
	wire.Bind(new(orchestrator.Dispatcher), new(*orchestrator.Plugin)),
	wire.InterfaceValue(new(datasync.KeyValProtoWatcher), local.DefaultRegistry),
	wire.InterfaceValue(new(datasync.KeyProtoValWriter), (datasync.KeyProtoValWriter)(nil)),
)

func KVSchedulerProvider(deps kvscheduler.Deps) (*kvscheduler.Scheduler, func(), error) {
	p := &kvscheduler.Scheduler{Deps: deps}
	p.SetName("kvscheduler")
	p.Log = logging.ForPlugin("kvscheduler")
	p.Cfg = config.ForPlugin("kvscheduler")
	cancel := func() {
		if err := p.Close(); err != nil {
			p.Log.Error(err)
		}
	}
	return p, cancel, p.Init()
}

var WireKVScheduler = wire.NewSet(
	KVSchedulerProvider,
	wire.Bind(new(api.KVScheduler), new(*kvscheduler.Scheduler)),
	wire.Struct(new(kvscheduler.Deps), "HTTPHandlers"),
)

func GovppProvider(deps govppmux.Deps) (*govppmux.Plugin, func(), error) {
	p := &govppmux.Plugin{Deps: deps}
	p.SetName("govpp")
	p.Log = logging.ForPlugin("govpp")
	p.Cfg = config.ForPlugin("govpp")
	cancel := func() {
		if err := p.Close(); err != nil {
			p.Log.Error(err)
		}
	}
	return p, cancel, p.Init()
}

var WireGoVppMux = wire.NewSet(
	GovppProvider,
	wire.Bind(new(govppmux.API), new(*govppmux.Plugin)),
	wire.NewSet(wire.Struct(new(govppmux.Deps), "HTTPHandlers", "StatusCheck", "Resync")),
)
