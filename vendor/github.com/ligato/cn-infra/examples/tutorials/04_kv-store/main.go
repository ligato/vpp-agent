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

package main

import (
	"fmt"
	"github.com/ligato/cn-infra/datasync"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/examples/tutorials/04_kv-store/model"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
)

//go:generate protoc --proto_path=model --gogo_out=model ./model/model.proto

func main() {
	// Create an instance of our plugin using its constructor.
	p := NewMyPlugin()

	// Create new agent with our plugin instance.
	a := agent.NewAgent(agent.AllPlugins(p))

	// Run starts the agent with plugins, wait until shutdown
	// and then stops the agent and its plugins.
	if err := a.Run(); err != nil {
		logging.DefaultLogger.Error(err)
	}
}

const keyPrefix = "/myplugin/"

// MyPlugin represents our plugin.
type MyPlugin struct {
	infra.PluginDeps
	KVStore     keyval.KvProtoPlugin
	watchCloser chan string
}

// NewMyPlugin is a constructor for our MyPlugin plugin.
func NewMyPlugin() *MyPlugin {
	p := &MyPlugin{
		watchCloser: make(chan string),
	}
	p.SetName("myplugin")
	p.Setup()
	// Initialize key-value store
	p.KVStore = &etcd.DefaultPlugin
	return p
}

// Init is executed on agent initialization.
func (p *MyPlugin) Init() error {
	if p.KVStore.Disabled() {
		return fmt.Errorf("KV store is disabled")
	}

	watcher := p.KVStore.NewWatcher(keyPrefix)

	// Start watching for changes
	err := watcher.Watch(p.onChange, p.watchCloser, "greetings/")
	if err != nil {
		return err
	}

	return nil
}

// AfterInit is executed after agent initialization.
func (p *MyPlugin) AfterInit() error {
	go p.updater()
	return nil
}

func (p *MyPlugin) onChange(resp datasync.ProtoWatchResp) {
	value := new(model.Greetings)
	// Deserialize data
	if err := resp.GetValue(value); err != nil {
		p.Log.Errorf("GetValue for change failed: %v", err)
		return
	}
	p.Log.Infof("%v change, Key: %q Value: %+v", resp.GetChangeType(), resp.GetKey(), value)
}

func (p *MyPlugin) updater() {
	broker := p.KVStore.NewBroker(keyPrefix)

	// Retrieve value from KV store
	value := new(model.Greetings)
	found, _, err := broker.GetValue("greetings/hello", value)
	if err != nil {
		p.Log.Errorf("GetValue failed: %v", err)
	} else if !found {
		p.Log.Info("No greetings found..")
	} else {
		p.Log.Infof("Found some greetings: %+v", value)
	}

	// Wait few seconds
	time.Sleep(time.Second * 2)

	p.Log.Infof("updating..")

	// Prepare data
	value = &model.Greetings{
		Greeting: "Hello",
	}

	// Update value in KV store
	if err := broker.Put("greetings/hello", value); err != nil {
		p.Log.Errorf("Put failed: %v", err)
	}
}
