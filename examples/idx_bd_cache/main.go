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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/tests/go/itest/l2tst"
)

// Start Agent plugins selected for this example
func main() {
	// Init close channel to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with VPP Flavor and ExampleFlavor
	vppFlavor := vpp.Flavor{}
	exampleFlavor := ExampleFlavor{
		IdxBdCacheExample: ExamplePlugin{closeChannel: &exampleFinished},
		Flavor:            &vppFlavor, // inject VPP flavor
	}
	agent := core.NewAgent(core.Inject(&vppFlavor, &exampleFlavor))

	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// ExamplePlugin is used for demonstration of Bridge Domain Indexes - see Init()
type ExamplePlugin struct {
	Deps

	bdIdxLocal  bdidx.BDIndex
	bdIdxAgent1 bdidx.BDIndex
	bdIdxAgent2 bdidx.BDIndex

	// Fields below are used to properly finish the example
	closeChannel *chan struct{}
}

// Init transport & bdIndexes then watch, publish & lookup
func (plugin *ExamplePlugin) Init() (err error) {
	// manually initialize 'other' agents (for example purpose only)
	err = plugin.Agent1.Init()
	if err != nil {
		return err
	}
	err = plugin.Agent2.Init()
	if err != nil {
		return err
	}

	// get access to local bridge domain indexes
	plugin.bdIdxLocal = defaultplugins.GetBDIndexes()

	// Run consumer
	go plugin.consume()

	// Cache other agent's bridge domain index mapping using injected plugin and local plugin name
	// /vnf-agent/agent1/vpp/config/v1/bd/
	plugin.bdIdxAgent1 = bdidx.Cache(plugin.Agent1, plugin.PluginName)
	// /vnf-agent/agent2/vpp/config/v1/bd/
	plugin.bdIdxAgent2 = bdidx.Cache(plugin.Agent2, plugin.PluginName)

	return nil
}

// AfterInit - call Cache()
func (plugin *ExamplePlugin) AfterInit() error {
	// manually run AfterInit() on 'other' agents (for example purpose only)
	err := plugin.Agent1.AfterInit()
	if err != nil {
		return err
	}
	err = plugin.Agent2.AfterInit()
	if err != nil {
		return err
	}

	// Publish test data
	plugin.publish()

	return nil
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	var wasErr error
	_, wasErr = safeclose.CloseAll(plugin.Agent1, plugin.Agent2, plugin.Publisher, plugin.Agent1, plugin.Agent2,
		plugin.bdIdxLocal, plugin.bdIdxAgent1, plugin.bdIdxAgent2, plugin.closeChannel)
	return wasErr
}

// Test data are published to different agents (including local)
func (plugin *ExamplePlugin) publish() (err error) {
	// Create bridge domain in local agent
	br0 := l2tst.SimpleBridgeDomain1XIfaceBuilder("bd0", "iface0", true)
	err = plugin.Publisher.Put(l2.BridgeDomainKey(br0.Name), &br0)
	if err != nil {
		return err
	}
	// Create bridge domain in agent1
	br1 := l2tst.SimpleBridgeDomain1XIfaceBuilder("bd1", "iface1", true)
	err = plugin.Agent1.Put(l2.BridgeDomainKey(br1.Name), &br1)
	if err != nil {
		return err
	}
	// Create bridge domain in agent2
	br2 := l2tst.SimpleBridgeDomain1XIfaceBuilder("bd2", "iface2", true)
	err = plugin.Agent2.Put(l2.BridgeDomainKey(br2.Name), &br2)
	return err
}

// uses the NameToIndexMapping to watch changes
func (plugin *ExamplePlugin) consume() {
	plugin.Log.Info("Watching started")
	bdIdxChan := make(chan bdidx.ChangeDto)
	// Subscribe local bd-idx-mapping and both of cache mapping
	plugin.bdIdxLocal.WatchNameToIdx(plugin.PluginName, bdIdxChan)
	plugin.bdIdxAgent1.WatchNameToIdx(plugin.PluginName, bdIdxChan)
	plugin.bdIdxAgent2.WatchNameToIdx(plugin.PluginName, bdIdxChan)

	counter := 0

	watching := true
	for watching {
		select {
		case bdIdxEvent := <-bdIdxChan:
			plugin.Log.Info("Event received: bridge domain ", bdIdxEvent.Name, " of ", bdIdxEvent.RegistryTitle)
			counter++
		}
		// Example is expecting 3 events
		if counter == 3 {
			watching = false
		}
	}

	// Do a lookup whether all mappings were registered
	plugin.lookup()
}

// use the NameToIndexMapping to lookup local mapping + external cached mappings
func (plugin *ExamplePlugin) lookup() {
	plugin.Log.Info("Lookup in progress")

	if index, _, found := plugin.bdIdxLocal.LookupIdx("bd0"); found {
		plugin.Log.Infof("Bridge domain bd0 (index %v) found in local mapping", index)
	}

	if index, _, found := plugin.bdIdxAgent1.LookupIdx("bd1"); found {
		plugin.Log.Infof("Bridge domain bd1 (index %v) found in local mapping", index)
	}

	if index, _, found := plugin.bdIdxAgent2.LookupIdx("bd2"); found {
		plugin.Log.Infof("Bridge domain bd2 (index %v) found in local mapping", index)
	}

	// End the example
	plugin.Log.Infof("idx-bd-cache example finished, sending shutdown ...")
	*plugin.closeChannel <- struct{}{}
}
