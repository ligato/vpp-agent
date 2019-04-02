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
	"context"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/examples/tutorials/05_kv-scheduler/model"
	"github.com/ligato/vpp-agent/plugins/kvscheduler"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

//go:generate protoc --proto_path=model --gogo_out=model ./model/model.proto
//go:generate descriptor-adapter --descriptor-name Interface --value-type *model.Interface --import "github.com/ligato/vpp-agent/examples/tutorials/05_kv-scheduler/model"
//go:generate descriptor-adapter --descriptor-name Route --value-type *model.Route --import "github.com/ligato/vpp-agent/examples/tutorials/05_kv-scheduler/model"

func main() {
	// Create an instance of our plugin using its constructor.
	p := NewHelloWorld()

	// Create new agent with our plugin instance.
	a := agent.NewAgent(agent.AllPlugins(p))

	// Run starts the agent with plugins, wait until shutdown
	// and then stops the agent and its plugins.
	if err := a.Run(); err != nil {
		logging.DefaultLogger.Error(err)
	}
}

// HelloWorld represents our plugin.
type HelloWorld struct {
	// This embeds essential plugin deps into our plugin.
	infra.PluginDeps
	// This embeds KV scheduler as a dependency for our plugin
	KVScheduler api.KVScheduler
}

// NewHelloWorld is a constructor for our HelloWorld plugin.
func NewHelloWorld() *HelloWorld {
	// Create new instance.
	p := new(HelloWorld)
	// Set the plugin name.
	p.SetName("helloworld")
	// Initialize essential plugin deps: logger and config.
	p.Setup()
	// Initialize the KV scheduler
	p.KVScheduler = &kvscheduler.DefaultPlugin
	return p
}

// Init is executed on agent initialization.
func (p *HelloWorld) Init() error {
	p.Log.Println("Hello World!")

	err := p.KVScheduler.RegisterKVDescriptor(NewIfDescriptor(p.Log))
	if err != nil {
		// handle error
	}

	err = p.KVScheduler.RegisterKVDescriptor(NewRouteDescriptor(p.Log))
	if err != nil {
		// handle error
	}

	return nil
}

// AfterInit is called after the Init() for all other plugins. In the example, the AfterInit
// is used for testing teh hello world plugin.
func (p *HelloWorld) AfterInit() error {
	p.Log.Println("Testing the KV scheduler ... ")

	go func() {
		p.Log.Info("CASE 1: the interface created before the route")

		// Case 1: the interface created before the route
		txn := p.KVScheduler.StartNBTransaction()
		txn.SetValue("/interface/if1", &model.Interface{Name: "if1"})
		txn.SetValue("/route/route1", &model.Route{Name: "route1", InterfaceName: "if1"})
		_, err := txn.Commit(context.Background())
		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)
		p.Log.Info("CASE 2: the route created before the interface")

		// Case 2: the route is created before the interface
		txn = p.KVScheduler.StartNBTransaction()
		txn.SetValue("/route/route2", &model.Route{Name: "route2", InterfaceName: "if2"})
		txn.SetValue("/interface/if2", &model.Interface{Name: "if2"})
		_, err = txn.Commit(context.Background())
		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)
		p.Log.Info("CASE 3: the route created before the interface in separate transaction")

		// Case 3: the route is created before the interface in separate transaction
		txn = p.KVScheduler.StartNBTransaction()
		txn.SetValue("/route/route3", &model.Route{Name: "route3", InterfaceName: "if3"})
		_, err = txn.Commit(context.Background())
		if err != nil {
			panic(err)
		}
		txn = p.KVScheduler.StartNBTransaction()
		txn.SetValue("/interface/if3", &model.Interface{Name: "if3"})
		_, err = txn.Commit(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

// Close is executed on agent shutdown.
func (p *HelloWorld) Close() error {
	p.Log.Info("Goodbye World!")
	return nil
}
