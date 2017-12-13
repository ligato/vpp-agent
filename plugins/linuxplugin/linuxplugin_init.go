// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package linuxplugin implements the Linux plugin that handles management
// of Linux VETH interfaces.
package linuxplugin

import (
	"context"
	"sync"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/utils/safeclose"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/l3idx"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "linuxplugin"

// Plugin implements Plugin interface, therefore it can be loaded with other plugins.
type Plugin struct {
	Deps

	// interfaces
	ifIndexes          ifaceidx.LinuxIfIndexRW
	ifConfigurator     *ifplugin.LinuxInterfaceConfigurator
	ifIndexesWatchChan chan ifaceidx.LinuxIfIndexDto

	// ARPs
	arpIndexes      l3idx.LinuxARPIndexRW
	arpConfigurator *l3plugin.LinuxArpConfigurator

	// static routes
	rtIndexes         l3idx.LinuxRouteIndexRW
	rtCachedIndexes   l3idx.LinuxRouteIndexRW
	routeConfigurator *l3plugin.LinuxRouteConfigurator

	resyncChan chan datasync.ResyncEvent
	changeChan chan datasync.ChangeEvent // TODO dedicated type abstracted from ETCD
	msChan     chan *ifplugin.MicroserviceCtx

	watchDataReg datasync.WatchRegistration

	enableStopwatch bool

	cancel context.CancelFunc // Cancel can be used to cancel all goroutines and their jobs inside of the plugin.
	wg     sync.WaitGroup     // Wait group allows to wait until all goroutines of the plugin have finished.
}

// Deps groups injected dependencies of plugin
// so that they do not mix with other plugin fields.
type Deps struct {
	local.PluginInfraDeps                             // injected
	Watcher               datasync.KeyValProtoWatcher // injected
}

// LinuxConfig holds the linuxplugin configuration.
type LinuxConfig struct {
	Stopwatch bool `json:"Stopwatch"`
}

// GetLinuxIfIndexes gives access to mapping of logical names (used in ETCD configuration)
// interface indexes.
func (plugin *Plugin) GetLinuxIfIndexes() ifaceidx.LinuxIfIndex {
	return plugin.ifIndexes
}

// GetLinuxARPIndexes gives access to mapping of logical names (used in ETCD configuration) to corresponding Linux
// ARP entry indexes.
func (plugin *Plugin) GetLinuxARPIndexes() l3idx.LinuxARPIndex {
	return plugin.arpIndexes
}

// GetLinuxRouteIndexes gives access to mapping of logical names (used in ETCD configuration) to corresponding Linux
// route indexes.
func (plugin *Plugin) GetLinuxRouteIndexes() l3idx.LinuxRouteIndex {
	return plugin.rtIndexes
}

// Init gets handlers for ETCD and Kafka and delegates them to ifConfigurator.
func (plugin *Plugin) Init() error {
	plugin.Log.Debug("Initializing Linux plugins")

	config, err := plugin.retrieveLinuxConfig()
	if err != nil {
		return err
	}
	if config != nil {
		plugin.enableStopwatch = config.Stopwatch
		if plugin.enableStopwatch {
			plugin.Log.Infof("stopwatch enabled for %v", plugin.PluginName)
		} else {
			plugin.Log.Infof("stopwatch disabled for %v", plugin.PluginName)
		}
	} else {
		plugin.Log.Infof("stopwatch disabled for %v", plugin.PluginName)
	}

	plugin.resyncChan = make(chan datasync.ResyncEvent)
	plugin.changeChan = make(chan datasync.ChangeEvent)
	plugin.msChan = make(chan *ifplugin.MicroserviceCtx)
	plugin.ifIndexesWatchChan = make(chan ifaceidx.LinuxIfIndexDto, 100)

	// Create plugin context and save cancel function into the plugin handle.
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())

	// Run event handler go routines
	go plugin.watchEvents(ctx)

	err = plugin.initIF()
	if err != nil {
		return err
	}

	err = plugin.initL3()
	if err != nil {
		return err
	}

	return plugin.subscribeWatcher()
}

// Initialize linux interface plugin
func (plugin *Plugin) initIF() error {
	plugin.Log.Infof("Init Linux interface plugin")
	// Interface indexes
	plugin.ifIndexes = ifaceidx.NewLinuxIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), PluginID,
		"linux_if_indexes", nil))

	// Linux interface configurator
	linuxLogger := plugin.Log.NewLogger("-if-conf")
	var stopwatch *measure.Stopwatch
	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("LinuxInterfaceConfigurator", linuxLogger)
	}
	plugin.ifConfigurator = &ifplugin.LinuxInterfaceConfigurator{Log: linuxLogger, Stopwatch: stopwatch}
	return plugin.ifConfigurator.Init(plugin.ifIndexes, plugin.msChan)
}

// Initialize linux L3 plugin
func (plugin *Plugin) initL3() error {
	plugin.Log.Infof("Init Linux L3 plugin")
	// ARP indexes
	plugin.arpIndexes = l3idx.NewLinuxARPIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), PluginID,
		"linux_arp_indexes", nil))

	// Linux ARP configurator
	linuxARPLogger := plugin.Log.NewLogger("-arp-conf")
	var stopwatch *measure.Stopwatch
	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("LinuxARPConfigurator", linuxARPLogger)
	}
	plugin.arpConfigurator = &l3plugin.LinuxArpConfigurator{
		Log:        linuxARPLogger,
		LinuxIfIdx: plugin.ifIndexes,
		ArpIdxSeq:  1,
		Stopwatch:  stopwatch}
	if err := plugin.arpConfigurator.Init(plugin.arpIndexes); err != nil {
		return err
	}

	// Route indexes
	plugin.rtIndexes = l3idx.NewLinuxRouteIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), PluginID,
		"linux_route_indexes", nil))
	plugin.rtCachedIndexes = l3idx.NewLinuxRouteIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), PluginID,
		"linux_cached_route_indexes", nil))

	// Linux Route configurator
	linuxRouteLogger := plugin.Log.NewLogger("-route-conf")
	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("LinuxRouteConfigurator", linuxRouteLogger)
	}
	plugin.routeConfigurator = &l3plugin.LinuxRouteConfigurator{
		Log:         linuxRouteLogger,
		LinuxIfIdx:  plugin.ifIndexes,
		RouteIdxSeq: 1,
		Stopwatch:   stopwatch}
	return plugin.routeConfigurator.Init(plugin.rtIndexes, plugin.rtCachedIndexes)
}

// AfterInit runs subscribeWatcher
func (plugin *Plugin) AfterInit() error {
	return nil
}

// Close cleans up the resources.
func (plugin *Plugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	_, err := safeclose.CloseAll(plugin.watchDataReg, plugin.changeChan, plugin.resyncChan,
		plugin.ifConfigurator)

	return err
}

func (plugin *Plugin) retrieveLinuxConfig() (*LinuxConfig, error) {
	config := &LinuxConfig{}
	found, err := plugin.PluginInfraDeps.GetValue(config)
	if !found {
		plugin.Log.Debug("Linuxplugin config not found")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	plugin.Log.Debug("Linuxplugin config found")
	return config, err
}
