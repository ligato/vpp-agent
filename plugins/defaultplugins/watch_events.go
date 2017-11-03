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

package defaultplugins

import (
	"strings"

	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"golang.org/x/net/context"
)

// WatchEvents goroutine is used to watch for changes in the northbound configuration & NameToIdxMapping notifications
func (plugin *Plugin) watchEvents(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	for {
		select {
		case resyncConfigEv := <-plugin.resyncConfigChan:
			req := plugin.resyncParseEvent(resyncConfigEv)
			var err error
			if plugin.resyncStrategy == skipResync {
				// skip resync
				plugin.Log.Info("skip VPP resync strategy chosen, VPP resync is omitted")
			} else if plugin.resyncStrategy == optimizeColdStart {
				// optimize resync
				err = plugin.resyncConfigPropageOptimizedRequest(req)
			} else {
				// full resync
				err = plugin.resyncConfigPropageFullRequest(req)
			}
			resyncConfigEv.Done(err)

		case resyncStatusEv := <-plugin.resyncStatusChan:
			var wasError error
			for key, vals := range resyncStatusEv.GetValues() {
				plugin.Log.Debugf("trying to delete obsolete status for key %v begin ", key)
				if strings.HasPrefix(key, interfaces.IfStatePrefix) {
					keys := []string{}
					for {
						x, stop := vals.GetNext()
						if stop {
							break
						}
						keys = append(keys, x.GetKey())
					}
					if len(keys) > 0 {
						err := plugin.resyncIfStateEvents(keys)
						if err != nil {
							wasError = err
						}
					}
				} else if strings.HasPrefix(key, l2.BdStatePrefix) {
					keys := []string{}
					for {
						x, stop := vals.GetNext()
						if stop {
							break
						}
						keys = append(keys, x.GetKey())
					}
					if len(keys) > 0 {
						err := plugin.resyncBdStateEvents(keys)
						if err != nil {
							wasError = err
						}
					}
				}
			}
			resyncStatusEv.Done(wasError)

		case dataChng := <-plugin.changeChan:
			// For FIBs only: if changePropagateRequest ends up without errors, the dataChng.Done is called in l2fib_vppcalls,
			// otherwise the dataChng.Done is called here
			callbackCalled, err := plugin.changePropagateRequest(dataChng, dataChng.Done)
			// When the request propagation is completed, send the error context (even if the error is nil)
			plugin.errorChannel <- ErrCtx{dataChng, err}
			if !callbackCalled {
				dataChng.Done(err)
			}

		case ifIdxEv := <-plugin.ifIdxWatchCh:
			if !ifIdxEv.IsDelete() {
				// Keep order
				plugin.bdConfigurator.ResolveCreatedInterface(ifIdxEv.Name, ifIdxEv.Idx)
				plugin.fibConfigurator.ResolveCreatedInterface(ifIdxEv.Name, ifIdxEv.Idx, func(err error) {
					if err != nil {
						plugin.Log.Error(err)
					}
				})
				plugin.xcConfigurator.ResolveCreatedInterface(ifIdxEv.Name, ifIdxEv.Idx)
				// TODO propagate error
			} else {
				plugin.bdConfigurator.ResolveDeletedInterface(ifIdxEv.Name) //TODO ifIdxEv.Idx to not process data events
				plugin.fibConfigurator.ResolveDeletedInterface(ifIdxEv.Name, ifIdxEv.Idx, func(err error) {
					if err != nil {
						plugin.Log.Error(err)
					}
				})
				plugin.xcConfigurator.ResolveDeletedInterface(ifIdxEv.Name)
				// TODO propagate error
			}
			ifIdxEv.Done()

		case linuxIfIdxEv := <-plugin.linuxIfIdxWatchCh:
			var name string
			if linuxIfIdxEv.Metadata != nil && linuxIfIdxEv.Metadata.HostIfName != "" {
				name = linuxIfIdxEv.Metadata.HostIfName
			} else {
				name = linuxIfIdxEv.Name
			}
			if !linuxIfIdxEv.IsDelete() {
				plugin.ifConfigurator.ResolveCreatedLinuxInterface(name, linuxIfIdxEv.Idx)
				// TODO propagate error
			} else {
				plugin.ifConfigurator.ResolveDeletedLinuxInterface(name)
				// TODO propagate error
			}
			linuxIfIdxEv.Done()

		case bdIdxEv := <-plugin.bdIdxWatchCh:
			if !bdIdxEv.IsDelete() {
				plugin.fibConfigurator.ResolveCreatedBridgeDomain(bdIdxEv.Name, bdIdxEv.Idx, func(err error) {
					if err != nil {
						plugin.Log.Error(err)
					}
				})
				// TODO propagate error
			} else {
				plugin.fibConfigurator.ResolveDeletedBridgeDomain(bdIdxEv.Name, bdIdxEv.Idx, func(err error) {
					if err != nil {
						plugin.Log.Error(err)
					}
				})
				// TODO propagate error
			}
			bdIdxEv.Done()

		case <-ctx.Done():
			plugin.Log.Debug("Stop watching events")
			return
		}
	}
}
