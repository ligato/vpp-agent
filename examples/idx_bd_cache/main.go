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
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/flavors/local"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/flavors/vpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/testing"
)

// Start Agent plugins selected for this example
func main() {
	// Init close channel to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combinatioplugin.GoVppmux, n of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{IdxBdCacheExample: ExamplePlugin{closeChannel: &exampleFinished}}
	agent := core.NewAgent(log.DefaultLogger(), 15*time.Second, append(flavor.Plugins())...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

/**********
 * Flavor *
 **********/

// ExampleFlavor is a set of plugins required for the datasync example.
type ExampleFlavor struct {
	// Local flavor to access to Infra (logger, service label, status check)
	*vpp.Flavor
	// Example plugin
	IdxBdCacheExample ExamplePlugin
	// Mark flavor as injected after Inject()
	injected bool
}

// Inject sets object references
func (ef *ExampleFlavor) Inject() (allReadyInjected bool) {
	// Every flavor should be injected only once
	if ef.injected {
		return false
	}
	ef.injected = true

	// Init local flavor
	if ef.Flavor == nil {
		ef.Flavor = &vpp.Flavor{}
	}
	ef.Flavor.Inject()

	// Inject infra + transport (publisher, watcher) to example plugin
	ef.IdxBdCacheExample.PluginInfraDeps = *ef.Flavor.InfraDeps("idx-bd-cache-example")
	ef.IdxBdCacheExample.Publisher = &ef.ETCDDataSync
	ef.IdxBdCacheExample.Agent1 = ef.Flavor.ETCDDataSync.OfDifferentAgent("agent1", ef)
	ef.IdxBdCacheExample.Agent2 = ef.Flavor.ETCDDataSync.OfDifferentAgent("agent2", ef)

	return true
}

// Plugins combines all Plugins in flavor to the list
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}

/******************
 * Example plugin *
 ******************/

// ExamplePlugin is used for demonstration of Bridge Domain Indexes - see Init()
type ExamplePlugin struct {
	Deps

	bdIdxLocal  bdidx.BDIndex
	bdIdxAgent1 bdidx.BDIndex
	bdIdxAgent2 bdidx.BDIndex

	// Fields below are used to properly finish the example
	closeChannel *chan struct{}
}

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	Publisher                 datasync.KeyProtoValWriter // injected
	Agent1                    *kvdbsync.Plugin           // injected
	Agent2                    *kvdbsync.Plugin           // injected
	local.PluginInfraDeps                            // injected
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

	// Cache other agent's bridge domain index mapping using injected plugin and local plugin name
	// /vnf-agent/agent1/vpp/config/v1/bd/
	plugin.bdIdxAgent1 = bdidx.Cache(plugin.Agent1, plugin.PluginName)
	// /vnf-agent/agent2/vpp/config/v1/bd/
	plugin.bdIdxAgent2 = bdidx.Cache(plugin.Agent2, plugin.PluginName)

	// Run consumer
	go plugin.consume()

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
	br0 := testing.SimpleBridgeDomain1XIfaceBuilder("bd0", "iface0", true)
	err = plugin.Publisher.Put(l2.BridgeDomainKey(br0.Name), &br0)
	if err != nil {
		return err
	}
	// Create bridge domain in agent1
	br1 := testing.SimpleBridgeDomain1XIfaceBuilder("bd1", "iface1", true)
	err = plugin.Agent1.Put(l2.BridgeDomainKey(br1.Name), &br1)
	if err != nil {
		return err
	}
	// Create bridge domain in agent2
	br2 := testing.SimpleBridgeDomain1XIfaceBuilder("bd2", "iface2", true)
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
			plugin.Log.Infof("Event received: bridge domain %v", bdIdxEvent.Name)
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
