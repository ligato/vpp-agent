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

package main

import (
	"log"

	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync/local"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/db/keyval/etcd"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/infra"

	"go.ligato.io/vpp-agent/v3/cmd/vpp-agent/app"
	"go.ligato.io/vpp-agent/v3/examples/extend/custom_vpp_plugin/syslog"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
)

//go:generate protoc --proto_path=. --go_out=paths=source_relative:. proto/custom/vpp/syslog/syslog.proto

func main() {
	ep := &Example{
		VPP:          app.DefaultVPP(),
		Syslog:       syslog.NewSyslogPlugin(),
		Orchestrator: &orchestrator.DefaultPlugin,
	}
	etcdDataSync := kvdbsync.NewPlugin(kvdbsync.UseKV(&etcd.DefaultPlugin))
	statuscheck.DefaultPlugin.Transport = etcdDataSync

	watchers := datasync.KVProtoWatchers{
		local.DefaultRegistry,
		etcdDataSync,
	}
	ep.Orchestrator.Watcher = watchers
	ep.Orchestrator.StatusPublisher = etcdDataSync
	ep.VPP.IfPlugin.DataSyncs = map[string]datasync.KeyProtoValWriter{
		"etcd": etcdDataSync,
	}

	ep.SetName("custom-vpp-plugin-example")
	ep.SetupLog()

	a := agent.NewAgent(
		agent.AllPlugins(ep),
	)
	if err := a.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("example agent ready!")

	if err := a.Wait(); err != nil {
		log.Fatal(err)
	}
}

type Example struct {
	infra.PluginDeps
	app.VPP
	Syslog       *syslog.SyslogPlugin
	Orchestrator *orchestrator.Plugin
}

func (p *Example) Init() error {
	return nil
}

func (p *Example) AfterInit() error {
	resync.DefaultPlugin.DoResync()
	return nil
}
