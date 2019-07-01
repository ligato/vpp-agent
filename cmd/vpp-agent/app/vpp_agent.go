//  Copyright (c) 2019 Cisco and/or its affiliates.
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
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/datasync/msgsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/consul"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/db/keyval/redis"
	"github.com/ligato/cn-infra/health/probe"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/logging/logmanager"
	"github.com/ligato/cn-infra/messaging/kafka"

	"github.com/ligato/vpp-agent/plugins/configurator"
	linux_ifplugin "github.com/ligato/vpp-agent/plugins/linux/ifplugin"
	linux_iptablesplugin "github.com/ligato/vpp-agent/plugins/linux/iptablesplugin"
	linux_l3plugin "github.com/ligato/vpp-agent/plugins/linux/l3plugin"
	linux_nsplugin "github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	"github.com/ligato/vpp-agent/plugins/orchestrator"
	"github.com/ligato/vpp-agent/plugins/restapi"
	"github.com/ligato/vpp-agent/plugins/telemetry"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/stnplugin"
)

// VPPAgent defines plugins which will be loaded and their order.
// Note: the plugin itself is loaded after all its dependencies. It means that the VPP plugin is first in the list
// despite it needs to be loaded after the linux plugin.
type VPPAgent struct {
	LogManager *logmanager.Plugin

	// VPP & Linux are first to ensure that
	// all their descriptors are regitered to KVScheduler
	// before orchestrator that starts watch for their NB key prefixes.
	VPP
	Linux

	Orchestrator *orchestrator.Plugin

	ETCDDataSync   *kvdbsync.Plugin
	ConsulDataSync *kvdbsync.Plugin
	RedisDataSync  *kvdbsync.Plugin

	Configurator *configurator.Plugin
	RESTAPI      *restapi.Plugin
	Probe        *probe.Plugin
	Telemetry    *telemetry.Plugin
}

// New creates new VPPAgent instance.
func New() *VPPAgent {
	etcdDataSync := kvdbsync.NewPlugin(kvdbsync.UseKV(&etcd.DefaultPlugin))
	consulDataSync := kvdbsync.NewPlugin(kvdbsync.UseKV(&consul.DefaultPlugin))
	redisDataSync := kvdbsync.NewPlugin(kvdbsync.UseKV(&redis.DefaultPlugin))

	writers := datasync.KVProtoWriters{
		etcdDataSync,
		consulDataSync,
		redisDataSync,
	}
	statuscheck.DefaultPlugin.Transport = writers

	ifStatePub := msgsync.NewPlugin(
		msgsync.UseMessaging(&kafka.DefaultPlugin),
		msgsync.UseConf(msgsync.Config{
			Topic: "if_state",
		}),
	)

	// Set watcher for KVScheduler.
	watchers := datasync.KVProtoWatchers{
		local.DefaultRegistry,
		etcdDataSync,
		consulDataSync,
		redisDataSync,
	}
	orchestrator.DefaultPlugin.Watcher = watchers
	orchestrator.DefaultPlugin.StatusPublisher = writers

	ifplugin.DefaultPlugin.Watcher = watchers
	ifplugin.DefaultPlugin.NotifyStates = ifStatePub
	puntplugin.DefaultPlugin.PublishState = writers

	// No stats publishers by default, use `vpp-ifplugin.conf` config
	// ifplugin.DefaultPlugin.PublishStatistics = writers
	ifplugin.DefaultPlugin.DataSyncs = map[string]datasync.KeyProtoValWriter{
		"etcd":   etcdDataSync,
		"redis":  redisDataSync,
		"consul": consulDataSync,
	}

	// connect IfPlugins for Linux & VPP
	linux_ifplugin.DefaultPlugin.VppIfPlugin = &ifplugin.DefaultPlugin
	ifplugin.DefaultPlugin.LinuxIfPlugin = &linux_ifplugin.DefaultPlugin
	ifplugin.DefaultPlugin.NsPlugin = &linux_nsplugin.DefaultPlugin

	vpp := DefaultVPP()
	linux := DefaultLinux()

	return &VPPAgent{
		LogManager:     &logmanager.DefaultPlugin,
		Orchestrator:   &orchestrator.DefaultPlugin,
		ETCDDataSync:   etcdDataSync,
		ConsulDataSync: consulDataSync,
		RedisDataSync:  redisDataSync,
		VPP:            vpp,
		Linux:          linux,
		Configurator:   &configurator.DefaultPlugin,
		RESTAPI:        &restapi.DefaultPlugin,
		Probe:          &probe.DefaultPlugin,
		Telemetry:      &telemetry.DefaultPlugin,
	}
}

// Init initializes main plugin.
func (VPPAgent) Init() error {
	return nil
}

// AfterInit executes resync.
func (VPPAgent) AfterInit() error {
	// manually start resync after all plugins started
	resync.DefaultPlugin.DoResync()
	//orchestrator.DefaultPlugin.InitialSync()
	return nil
}

// Close could close used resources.
func (VPPAgent) Close() error {
	return nil
}

// String returns name of the plugin.
func (VPPAgent) String() string {
	return "VPPAgent"
}

// VPP contains all VPP plugins.
type VPP struct {
	ABFPlugin   *abfplugin.ABFPlugin
	ACLPlugin   *aclplugin.ACLPlugin
	IfPlugin    *ifplugin.IfPlugin
	IPSecPlugin *ipsecplugin.IPSecPlugin
	L2Plugin    *l2plugin.L2Plugin
	L3Plugin    *l3plugin.L3Plugin
	NATPlugin   *natplugin.NATPlugin
	PuntPlugin  *puntplugin.PuntPlugin
	STNPlugin   *stnplugin.STNPlugin
	SRPlugin    *srplugin.SRPlugin
}

func DefaultVPP() VPP {
	return VPP{
		ABFPlugin:   &abfplugin.DefaultPlugin,
		ACLPlugin:   &aclplugin.DefaultPlugin,
		IfPlugin:    &ifplugin.DefaultPlugin,
		IPSecPlugin: &ipsecplugin.DefaultPlugin,
		L2Plugin:    &l2plugin.DefaultPlugin,
		L3Plugin:    &l3plugin.DefaultPlugin,
		NATPlugin:   &natplugin.DefaultPlugin,
		PuntPlugin:  &puntplugin.DefaultPlugin,
		STNPlugin:   &stnplugin.DefaultPlugin,
		SRPlugin:    &srplugin.DefaultPlugin,
	}
}

// Linux contains all Linux plugins.
type Linux struct {
	IfPlugin       *linux_ifplugin.IfPlugin
	L3Plugin       *linux_l3plugin.L3Plugin
	NSPlugin       *linux_nsplugin.NsPlugin
	IPTablesPlugin *linux_iptablesplugin.IPTablesPlugin
}

func DefaultLinux() Linux {
	return Linux{
		IfPlugin:       &linux_ifplugin.DefaultPlugin,
		L3Plugin:       &linux_l3plugin.DefaultPlugin,
		NSPlugin:       &linux_nsplugin.DefaultPlugin,
		IPTablesPlugin: &linux_iptablesplugin.DefaultPlugin,
	}
}
