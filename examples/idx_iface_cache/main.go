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

package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/flavors/vpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/tests/go/itest/iftst"
)

// Start Agent plugins selected for this example.
func main() {
	// Init close channel to stop the example.
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor.
	vppFlavor := vpp.Flavor{}
	exampleFlavor := ExampleFlavor{
		IdxIfaceCacheExample: ExamplePlugin{closeChannel: &exampleFinished},
		Flavor:               &vppFlavor, // inject VPP flavor
	}
	agent := core.NewAgent(core.Inject(&vppFlavor, &exampleFlavor))

	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// ExamplePlugin used for demonstration of SwIfIndexes - see Init()
type ExamplePlugin struct {
	Deps

	// Linux plugin dependency
	VPP defaultplugins.API
	
	swIfIdxLocal  ifaceidx.SwIfIndex
	swIfIdxAgent1 ifaceidx.SwIfIndex
	swIfIdxAgent2 ifaceidx.SwIfIndex

	// Fields below are used to properly finish the example.
	closeChannel *chan struct{}
}

// Init initializes transport & SwIfIndexes then watch, publish & lookup.
func (plugin *ExamplePlugin) Init() (err error) {
	// Manually initialize 'other' agents (for the example purpose only).
	err = plugin.Agent1.Init()
	if err != nil {
		return err
	}
	err = plugin.Agent2.Init()
	if err != nil {
		return err
	}

	// Get access to local interface indexes.
	plugin.swIfIdxLocal = plugin.VPP.GetSwIfIndexes()

	// Run consumer
	go plugin.consume()

	// Cache other agent's interface index mapping using injected plugin and local plugin name.
	// /vnf-agent/agent1/vpp/config/v1/interface/
	plugin.swIfIdxAgent1 = ifaceidx.Cache(plugin.Agent1, plugin.PluginName)
	// /vnf-agent/agent2/vpp/config/v1/interface/
	plugin.swIfIdxAgent2 = ifaceidx.Cache(plugin.Agent2, plugin.PluginName)

	return nil
}

// AfterInit - call Cache()
func (plugin *ExamplePlugin) AfterInit() error {
	// Manually run AfterInit() on 'other' agents (for example purpose only).
	err := plugin.Agent1.AfterInit()
	if err != nil {
		return err
	}
	err = plugin.Agent2.AfterInit()
	if err != nil {
		return err
	}

	// Publish test data.
	plugin.publish()

	return nil
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed
// to clean up resources that were allocated by the plugin during its lifetime.
func (plugin *ExamplePlugin) Close() error {
	var wasErr error
	_, wasErr = safeclose.CloseAll(plugin.Agent1, plugin.Agent2, plugin.Publisher, plugin.Agent1, plugin.Agent2,
		plugin.swIfIdxLocal, plugin.swIfIdxAgent1, plugin.swIfIdxAgent2, plugin.closeChannel)
	return wasErr
}

// Test data are published to different agents (including local).
func (plugin *ExamplePlugin) publish() (err error) {
	// Create interface in local agent.
	iface0 := iftst.TapInterfaceBuilder("iface0", "192.168.1.1")
	err = plugin.Publisher.Put(interfaces.InterfaceKey(iface0.Name), &iface0)
	if err != nil {
		return err
	}
	// Create interface in agent1.
	iface1 := iftst.TapInterfaceBuilder("iface1", "192.168.0.2")
	err = plugin.Agent1.Put(interfaces.InterfaceKey(iface1.Name), &iface1)
	if err != nil {
		return err
	}
	// Create interface in agent2.
	iface2 := iftst.TapInterfaceBuilder("iface2", "192.168.0.3")
	err = plugin.Agent2.Put(interfaces.InterfaceKey(iface2.Name), &iface2)
	return err
}

// Use the NameToIndexMapping to watch changes.
func (plugin *ExamplePlugin) consume() {
	plugin.Log.Info("Watching started")
	swIfIdxChan := make(chan ifaceidx.SwIfIdxDto)
	// Subscribe local iface-idx-mapping and both of cache mapping.
	plugin.swIfIdxLocal.WatchNameToIdx(plugin.PluginName, swIfIdxChan)
	plugin.swIfIdxAgent1.WatchNameToIdx(plugin.PluginName, swIfIdxChan)
	plugin.swIfIdxAgent2.WatchNameToIdx(plugin.PluginName, swIfIdxChan)

	counter := 0

	watching := true
	for watching {
		select {
		case ifaceIdxEvent := <-swIfIdxChan:
			plugin.Log.Info("Event received: interface ", ifaceIdxEvent.Name, " of ", ifaceIdxEvent.RegistryTitle)
			counter++
		}
		// Example is expecting 3 events
		if counter == 3 {
			watching = false
		}
	}

	// Do a lookup whether all mappings were registered.
	plugin.lookup()
}

// Use the NameToIndexMapping to lookup local mapping + external cached mappings.
func (plugin *ExamplePlugin) lookup() {
	plugin.Log.Info("Lookup in progress")

	if index, _, found := plugin.swIfIdxLocal.LookupIdx("iface0"); found {
		plugin.Log.Infof("interface iface0 (index %v) found in local mapping", index)
	}

	if index, _, found := plugin.swIfIdxAgent1.LookupIdx("iface1"); found {
		plugin.Log.Infof("interface iface1 (index %v) found in local mapping", index)
	}

	if index, _, found := plugin.swIfIdxAgent2.LookupIdx("iface2"); found {
		plugin.Log.Infof("interface iface2 (index %v) found in local mapping", index)
	}

	// End the example.
	plugin.Log.Infof("idx-iface-cache example finished, sending shutdown ...")
	*plugin.closeChannel <- struct{}{}
}
