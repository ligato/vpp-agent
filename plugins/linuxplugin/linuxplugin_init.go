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

package linuxplugin

import (
	"context"
	"sync"

	"github.com/ligato/cn-infra/core"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/safeclose"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "linuxplugin"

// Plugin implements Plugin interface, therefore it can be loaded with other plugins
type Plugin struct {
	transport datasync.TransportAdapter // data transport adapter

	ifIndexes      idxvpp.NameToIdxRW
	ifConfigurator *LinuxInterfaceConfigurator

	resyncChan chan datasync.ResyncEvent
	changeChan chan datasync.ChangeEvent // TODO dedicated type abstracted from ETCD

	watchDataReg datasync.WatchDataRegistration

	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// GetLinuxIfIndexes gives access to mapping of logical names (used in ETCD configuration) to corresponding Linux
// interface indexes. This mapping is especially helpful for plugins that need to watch for newly added or deleted
// Linux interfaces.
func (plugin *Plugin) GetLinuxIfIndexes() idxvpp.NameToIdxRW {
	return plugin.ifIndexes
}

// Init gets handlers for ETCD, Kafka and delegates them to ifConfigurator
func (plugin *Plugin) Init() error {
	var err error
	plugin.transport = datasync.GetTransport()

	log.Debug("Initializing Linux interface plugin")

	plugin.resyncChan = make(chan datasync.ResyncEvent)
	plugin.changeChan = make(chan datasync.ChangeEvent)

	// create plugin context, save cancel function into the plugin handle
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())

	// run event handler go routines
	go plugin.watchEvents(ctx)

	// Interface indexes
	plugin.ifIndexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "linux_if_indexes", nil)

	// Linux interface configurator
	plugin.ifConfigurator = &LinuxInterfaceConfigurator{}
	plugin.ifConfigurator.Init(plugin.ifIndexes)

	err = plugin.subscribeWatcher()
	if err != nil {
		return err
	}

	return nil
}

// Close cleans up the resources
func (plugin *Plugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	_, err := safeclose.CloseAll(plugin.watchDataReg, plugin.changeChan, plugin.resyncChan,
		plugin.ifConfigurator)

	return err
}
