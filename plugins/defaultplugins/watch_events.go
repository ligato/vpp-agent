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

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"golang.org/x/net/context"
)

// IfMappingEventHandler defines functions for name-to-index mappings for interfaces
type IfMappingEventHandler interface {
	// RegisteredInterface handles registered (not necessarily created) interface for particular configurator
	RegisteredInterface(ifName string, ifIdx uint32) error
	// UnregisteredInterface handles unregistered (not necessarily removed) interface for particular configurator
	UnregisteredInterface(ifName string, ifIdx uint32) error
}

// BDMappingEventHandler defines functions for name-to-index mappings for bridge domains
type BDMappingEventHandler interface {
	// RegisteredBridgeDomain handles registered (not necessarily created) bridge domain for particular configurator
	RegisteredBridgeDomain(ifName string, ifIdx uint32, callback func(err error)) error
	// UnregisteredBridgeDomain handles unregistered (not necessarily removed) bridge domain for particular configurator
	UnregisteredBridgeDomain(ifName string, ifIdx uint32, callback func(err error)) error
}

// LinuxIfMappingEventHandler defines functions for name-to-index mappings for linux interfaces
type LinuxIfMappingEventHandler interface {
	// RegisteredInterface handles registered (not necessarily created) linux interface for particular configurator
	RegisteredLinuxInterface(ifName, hostName string, ifIdx uint32) error
	// UnregisteredInterface handles unregistered (not necessarily removed) linux interface for particular configurator
	UnregisteredLinuxInterface(ifName, hostName string, ifIdx uint32) error
}

// WatchEvents goroutine is used to watch for changes in the northbound configuration & NameToIdxMapping notifications.
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
					var keys []string
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
					var keys []string
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
			// For asynchronous calls only: if changePropagateRequest ends up without errors,
			// the dataChng.Done is called in particular vppcall, otherwise the dataChng.Done is called here.
			callbackCalled, err := plugin.changePropagateRequest(dataChng, dataChng.Done)
			// When the request propagation is complete, send the error context (even if the error is nil).
			plugin.errorChannel <- ErrCtx{dataChng, err}
			if !callbackCalled {
				dataChng.Done(err)
			}

		case ifIdxEv := <-plugin.ifIdxWatchCh:
			// Keep order.
			configurators := []IfMappingEventHandler{plugin.aclConfigurator, plugin.arpConfigurator, plugin.bdConfigurator,
				plugin.xcConfigurator, plugin.l4Configurator, plugin.stnConfigurator, plugin.routeConfigurator}
			for _, configurator := range configurators {
				if ifIdxEv.IsDelete() {
					if err := configurator.UnregisteredInterface(ifIdxEv.Name, ifIdxEv.Idx); err != nil {
						plugin.Log.Error(err)
					}
				} else {
					if err := configurator.RegisteredInterface(ifIdxEv.Name, ifIdxEv.Idx); err != nil {
						plugin.Log.Error(err)
					}
				}
			}

			if !ifIdxEv.IsDelete() {
				plugin.fibConfigurator.ResolveRegisteredInterface(ifIdxEv.Name, ifIdxEv.Idx, func(err error) {
					if err != nil {
						plugin.Log.Error(err)
					}
				})
			} else {
				plugin.fibConfigurator.ResolveUnregisteredInterface(ifIdxEv.Name, ifIdxEv.Idx, func(err error) {
					if err != nil {
						plugin.Log.Error(err)
					}
				})
			}
			ifIdxEv.Done()

		case linuxIfIdxEv := <-plugin.linuxIfIdxWatchCh:
			ifName := linuxIfIdxEv.Name
			var hostIfName string
			if linuxIfIdxEv.Metadata != nil && linuxIfIdxEv.Metadata.Data != nil && linuxIfIdxEv.Metadata.Data.HostIfName != "" {
				hostIfName = linuxIfIdxEv.Metadata.Data.HostIfName
			}
			// Keep order.
			configurators := []LinuxIfMappingEventHandler{plugin.ifConfigurator}
			for _, configurator := range configurators {
				if linuxIfIdxEv.IsDelete() {
					if err := configurator.UnregisteredLinuxInterface(ifName, hostIfName, linuxIfIdxEv.Idx); err != nil {
						plugin.Log.Error(err)
					}
				} else {
					if err := configurator.RegisteredLinuxInterface(ifName, hostIfName, linuxIfIdxEv.Idx); err != nil {
						plugin.Log.Error(err)
					}
				}
			}
			linuxIfIdxEv.Done()

		case bdIdxEv := <-plugin.bdIdxWatchCh:
			// Keep order.
			configurators := []BDMappingEventHandler{plugin.fibConfigurator}
			for _, configurator := range configurators {
				if !bdIdxEv.IsDelete() {
					configurator.RegisteredBridgeDomain(bdIdxEv.Name, bdIdxEv.Idx, func(err error) {
						if err != nil {
							plugin.Log.Error(err)
						}
					})
				} else {
					configurator.UnregisteredBridgeDomain(bdIdxEv.Name, bdIdxEv.Idx, func(err error) {
						if err != nil {
							plugin.Log.Error(err)
						}
					})
				}
			}
			bdIdxEv.Done()

		case <-ctx.Done():
			plugin.Log.Debug("Stop watching events")
			return
		}
	}
}
