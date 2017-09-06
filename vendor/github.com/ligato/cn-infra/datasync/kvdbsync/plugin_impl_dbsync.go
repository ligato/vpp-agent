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

package kvdbsync

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/flavors/localdeps"
	"github.com/ligato/cn-infra/servicelabel"
)

// Plugin dbsync implements Plugin interface
type Plugin struct {
	Deps    // inject
	adapter *watcher
}

type infraDeps interface {
	// InfraDeps for getting PlugginInfraDeps instance (logger, config, plugin name, statuscheck)
	InfraDeps(pluginName string) *localdeps.PluginInfraDeps
}

// OfDifferentAgent allows access DB of different agent (with a particular microservice label).
// This method is a shortcut to simplify creating new instance of plugin
// that is supposed to watch different agent DB.
// Method intentionally copies instance of plugin (assuming it has set all dependencies)
// and sets microservice label.
func (plugin /*intentionally without pointer receiver*/ Plugin) OfDifferentAgent(
	microserviceLabel string, infraDeps infraDeps) *Plugin {

	// plugin name suffixed by micorservice label
	plugin.Deps.PluginInfraDeps = *infraDeps.InfraDeps(string(
		plugin.Deps.PluginInfraDeps.PluginName) + "-" + microserviceLabel)

	// this is important - here comes microservice label of different agent
	plugin.Deps.PluginInfraDeps.ServiceLabel = servicelabel.OfDifferentAgent(microserviceLabel)
	return &plugin // copy (no pointer receiver)
}


// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	localdeps.PluginInfraDeps                      // inject
	ResyncOrch                resync.Subscriber    // inject
	KvPlugin                  keyval.KvProtoPlugin // inject
}

// Init does nothing
func (plugin *Plugin) Init() error {
	return nil
}

// AfterInit uses provided connection to build new transport watcher
func (plugin *Plugin) AfterInit() error {
	if plugin.KvPlugin != nil && !plugin.KvPlugin.Disabled() {
		db := plugin.KvPlugin.NewBroker(plugin.ServiceLabel.GetAgentPrefix())
		dbW := plugin.KvPlugin.NewWatcher(plugin.ServiceLabel.GetAgentPrefix())

		plugin.adapter = &watcher{db, dbW, syncbase.NewWatcher()}
	}

	return nil
}

// Watch using ETCD or any other Key Val data store.
func (plugin *Plugin) Watch(resyncName string, changeChan chan datasync.ChangeEvent,
	resyncChan chan datasync.ResyncEvent, keyPrefixes ...string) (datasync.WatchRegistration, error) {

	if plugin.KvPlugin.Disabled() {
		return nil /*TODO*/, nil
	}

	if plugin.adapter == nil {
		return nil, errors.New("Transport adapter is not ready yet")
	}

	reg, err := plugin.adapter.base.Watch(resyncName, changeChan, resyncChan, keyPrefixes...)
	if err != nil {
		return nil, err
	}

	if plugin.ResyncOrch != nil {
		resyncReg := plugin.ResyncOrch.Register(resyncName)
		_, err = watchAndResyncBrokerKeys(resyncReg, changeChan, resyncChan, plugin.adapter, keyPrefixes...)
		if err != nil {
			return nil, err
		}
	}

	return reg, err
}

// Put to ETCD or any other data transport (from other Agent Plugins)
func (plugin *Plugin) Put(key string, data proto.Message, opts ...datasync.PutOption) error {
	if plugin.KvPlugin.Disabled() {
		return nil
	}

	if plugin.adapter != nil {
		return plugin.adapter.db.Put(key, data, opts...)
	}

	return errors.New("Transport adapter is not ready yet")
}

// Close resources
func (plugin *Plugin) Close() error {
	return nil
}

// String returns if set Deps.PluginName or "kvdbsync" otherwise
func (plugin *Plugin) String() string {
	if len(plugin.PluginName) == 0 {
		return "kvdbsync"
	}
	return string(plugin.PluginName)
}
