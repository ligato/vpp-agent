//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package app

import (
	"sync"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/datasync/msgsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/consul"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/db/keyval/redis"
	"github.com/ligato/cn-infra/logging/logmanager"
	"github.com/ligato/cn-infra/messaging/kafka"
	"github.com/ligato/vpp-agent/plugins/linux"
	"github.com/ligato/vpp-agent/plugins/rest"
	"github.com/ligato/vpp-agent/plugins/telemetry"
	"github.com/ligato/vpp-agent/plugins/vpp"
	"github.com/ligato/vpp-agent/plugins/vpp/rpc"
)

type VPPAgent struct {
	LogManager *logmanager.Plugin

	ETCDDataSync   *kvdbsync.Plugin
	ConsulDataSync *kvdbsync.Plugin
	RedisDataSync  *kvdbsync.Plugin

	Linux *linux.Plugin
	VPP   *vpp.Plugin

	GRPCService *rpc.Plugin
	RESTAPI     *rest.Plugin
	Telemetry   *telemetry.Plugin
}

func New() *VPPAgent {
	var useKV = func(kv keyval.KvProtoPlugin) kvdbsync.Option {
		return kvdbsync.UseDeps(func(deps *kvdbsync.Deps) {
			deps.KvPlugin = kv
			deps.ResyncOrch = &resync.DefaultPlugin
		})
	}
	etcdDataSync := kvdbsync.NewPlugin(useKV(&etcd.DefaultPlugin))
	consulDataSync := kvdbsync.NewPlugin(useKV(&consul.DefaultPlugin))
	redisDataSync := kvdbsync.NewPlugin(useKV(&redis.DefaultPlugin))

	watcher := &datasync.CompositeKVProtoWatcher{Adapters: []datasync.KeyValProtoWatcher{
		local.Get(),
		etcdDataSync,
		consulDataSync,
	}}
	publisher := &datasync.CompositeKVProtoWriter{Adapters: []datasync.KeyProtoValWriter{
		etcdDataSync,
		consulDataSync,
	}}

	ifStatePub := msgsync.NewPlugin(
		msgsync.UseDeps(func(deps *msgsync.Deps) {
			deps.Messaging = &kafka.DefaultPlugin
		}),
		msgsync.UseConf(msgsync.Config{
			Topic: "if_state",
		}),
	)

	vppPlugin := vpp.NewPlugin(vpp.UseDeps(func(deps *vpp.Deps) {
		deps.Publish = publisher
		deps.Watcher = watcher
		deps.IfStatePub = ifStatePub
		deps.DataSyncs = map[string]datasync.KeyProtoValWriter{
			"etcd":  etcdDataSync,
			"redis": redisDataSync,
		}
	}))
	linuxPlugin := linux.NewPlugin(linux.UseDeps(func(deps *linux.Deps) {
		deps.VPP = vppPlugin
		deps.Watcher = watcher
	}))

	vppPlugin.Deps.Linux = linuxPlugin

	var watchEventsMutex sync.Mutex
	vppPlugin.Deps.WatchEventsMutex = &watchEventsMutex
	linuxPlugin.Deps.WatchEventsMutex = &watchEventsMutex

	restPlugin := rest.NewPlugin(rest.UseDeps(func(deps *rest.Deps) {
		deps.VPP = vppPlugin
	}))

	return &VPPAgent{
		LogManager:     &logmanager.DefaultPlugin,
		ETCDDataSync:   etcdDataSync,
		ConsulDataSync: consulDataSync,
		RedisDataSync:  redisDataSync,
		VPP:            vppPlugin,
		Linux:          linuxPlugin,
		GRPCService:    &rpc.DefaultPlugin,
		RESTAPI:        restPlugin,
		Telemetry:      &telemetry.DefaultPlugin,
	}
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

func (VPPAgent) String() string {
	return "VPPAgent"
}
