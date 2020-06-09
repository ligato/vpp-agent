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

// Example Custom VPP plugin contains a working example of custom agent which
// adds support for a custom VPP plugin. This example can serve as a skeleton
// code for developing custom agents adding new VPP functionality that is not
// part of official VPP Agent.
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
	"go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/syslog"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
)

// This go generate directive will generate Go code for Proto definition.
//go:generate protoc --proto_path=. --go_out=paths=source_relative:. proto/custom/vpp/syslog/syslog.proto

func main() {
	example := NewExample()

	a := agent.NewAgent(
		agent.AllPlugins(example),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

type Example struct {
	infra.PluginDeps
	app.VPP
	Syslog       *syslog.SyslogPlugin
	Orchestrator *orchestrator.Plugin
}

func NewExample() *Example {
	example := &Example{
		VPP:          app.DefaultVPP(),
		Syslog:       syslog.NewSyslogPlugin(),
		Orchestrator: &orchestrator.DefaultPlugin,
	}
	example.SetName("custom-vpp-plugin-example")
	example.SetupLog()

	etcdDataSync := kvdbsync.NewPlugin(kvdbsync.UseKV(&etcd.DefaultPlugin))
	statuscheck.DefaultPlugin.Transport = etcdDataSync

	watchers := datasync.KVProtoWatchers{
		local.DefaultRegistry,
		etcdDataSync,
	}
	example.Orchestrator.Watcher = watchers
	example.Orchestrator.StatusPublisher = etcdDataSync
	example.VPP.IfPlugin.DataSyncs = map[string]datasync.KeyProtoValWriter{
		"etcd": etcdDataSync,
	}

	return example
}

func (p *Example) Init() error {
	return nil
}

func (p *Example) AfterInit() error {
	resync.DefaultPlugin.DoResync()
	return nil
}
