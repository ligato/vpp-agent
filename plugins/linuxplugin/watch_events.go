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
	"golang.org/x/net/context"
)

// WatchEvents goroutine is used to watch for changes in the northbound configuration
func (plugin *Plugin) watchEvents(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	for {
		select {
		case resyncEv := <-plugin.resyncChan:
			req := resyncParseEvent(resyncEv, plugin.Log)
			err := plugin.resyncPropageRequest(req)

			resyncEv.Done(err)

		case dataChng := <-plugin.changeChan:
			err := plugin.changePropagateRequest(dataChng)

			dataChng.Done(err)

		case linuxIdxEv := <-plugin.ifIndexesWatchChan:
			if linuxIdxEv.IsDelete() {
				plugin.routeConfigurator.ResolveDeletedInterface(linuxIdxEv.Name, linuxIdxEv.Idx)
			} else {
				plugin.routeConfigurator.ResolveCreatedInterface(linuxIdxEv.Name, linuxIdxEv.Idx)
			}
			linuxIdxEv.Done()

		case <-ctx.Done():
			plugin.Log.Debug("Stop watching events")
			return
		}
	}
}
