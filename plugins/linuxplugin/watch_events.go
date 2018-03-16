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
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/ifaceidx"
	"golang.org/x/net/context"
)

// WatchEvents goroutine is used to watch for changes in the northbound configuration.
func (plugin *Plugin) watchEvents(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	for {
		select {
		case e := <-plugin.resyncChan:
			plugin.WatchEventsMutex.Lock()
			plugin.onResyncEvent(e)
			plugin.WatchEventsMutex.Unlock()

		case e := <-plugin.changeChan:
			plugin.WatchEventsMutex.Lock()
			plugin.onChangeEvent(e)
			plugin.WatchEventsMutex.Unlock()

		case ms := <-plugin.msChan:
			plugin.WatchEventsMutex.Lock()
			plugin.nsHandler.HandleMicroservices(ms)
			plugin.WatchEventsMutex.Unlock()

		case e := <-plugin.ifIndexesWatchChan:
			plugin.WatchEventsMutex.Lock()
			plugin.onLinuxIfaceEvent(e)
			plugin.WatchEventsMutex.Unlock()

		case <-ctx.Done():
			plugin.Log.Debug("Stop watching events")
			return
		}
	}
}

func (plugin *Plugin) onResyncEvent(e datasync.ResyncEvent) {
	req := resyncParseEvent(e, plugin.Log)
	err := plugin.resyncPropageRequest(req)
	e.Done(err)
}

func (plugin *Plugin) onChangeEvent(e datasync.ChangeEvent) {
	err := plugin.changePropagateRequest(e)
	e.Done(err)
}

func (plugin *Plugin) onLinuxIfaceEvent(e ifaceidx.LinuxIfIndexDto) {
	if e.IsDelete() {
		plugin.arpConfigurator.ResolveDeletedInterface(e.Name, e.Idx)
		plugin.routeConfigurator.ResolveDeletedInterface(e.Name, e.Idx)
	} else {
		plugin.arpConfigurator.ResolveCreatedInterface(e.Name, e.Idx)
		plugin.routeConfigurator.ResolveCreatedInterface(e.Name, e.Idx)
	}
	e.Done()
}
