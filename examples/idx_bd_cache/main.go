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
	"fmt"
	"github.com/ligato/cn-infra/core"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/flavours/vpp"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/vpp-agent/defaultplugins"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/bdidx"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/testing"
	"time"
	"github.com/ligato/cn-infra/logging"
)

// Start Agent plugins selected for this example
func main() {
	// Init close channel to stop the example
	closeChannel := make(chan struct{}, 1)

	f := vpp.Flavour{}
	// Example plugin will show index mapping
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &examplePlugin{}}

	// Create new agent
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(f.Plugins(), examplePlugin)...)

	// End when the idx_bd_cache example is finished
	go closeExample("idx_bd_cache example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(15 * time.Second)
	log.Info(message)
	closeChannel <- struct{}{}
}

// PluginID of example plugin
const PluginID core.PluginName = "example-plugin"

// used for demonstration of Bridge Domain Indexes - see Init()
type examplePlugin struct {
	agent1      datasync.TransportAdapter
	agent2      datasync.TransportAdapter
	bdIdxLocal  bdidx.BDIndex
	bdIdxAgent1 bdidx.BDIndex
	bdIdxAgent2 bdidx.BDIndex
}

// initialize transport & SwIfIndexes then watch, publish & lookup
func (plugin *examplePlugin) Init() (err error) {
	plugin.agent1 = datasync.OfDifferentAgent("agent1" /*TODO "br1", "br2"*/)
	plugin.agent2 = datasync.OfDifferentAgent("agent2")
	// /vnf-agent/agent0/vpp/config/v1/interface/
	plugin.bdIdxLocal = defaultplugins.GetBDIndexes()
	// /vnf-agent/agent1/vpp/config/v1/bd/
	plugin.bdIdxAgent1 = bdidx.Cache(plugin.agent1, PluginID)
	// /vnf-agent/agent2/vpp/config/v1/bd/
	plugin.bdIdxAgent2 = bdidx.Cache(plugin.agent2, PluginID)

	if err := plugin.publish(); err != nil {
		return err
	}

	return plugin.consume()
}

// prepares test data for different agents
func (plugin *examplePlugin) publish() (err error) {
	//br1 :=
	//	&testing.BDAfPacketVeth1VxlanVni5
	//err = plugin.agent1.PublishData(l2.BridgeDomainKey(br1.Name), br1)
	if err != nil {
		return err
	}
	br2 := &testing.BDMemif100011ToMemif100012
	err = plugin.agent2.PublishData(l2.BridgeDomainKey(br2.Name), br2)
	return err
}

// uses the NameToIndexMapping to watch changes
func (plugin *examplePlugin) consume() (err error) {
	bdIdxChan := make(chan bdidx.ChangeDto)
	plugin.bdIdxLocal.WatchNameToIdx(PluginID, bdIdxChan)
	plugin.bdIdxAgent1.WatchNameToIdx(PluginID, bdIdxChan)
	plugin.bdIdxAgent2.WatchNameToIdx(PluginID, bdIdxChan)

	go func() {
		var watching = true
		for watching {
			select {
			case bdIdxEvent := <-bdIdxChan:
				log.WithFields(logging.Fields{"RegistryTitle": bdIdxEvent.RegistryTitle,
												"Name": bdIdxEvent.Name, //br1, br2
												"Del": bdIdxEvent.Del,
												"IFaces": bdIdxEvent.Metadata.Interfaces}).
												Info("xxx event received")
			case <-time.After(10 * time.Second):
				watching = false
			}
		}

		plugin.lookup()
	}()

	return nil
}

// use the NameToIndexMapping to lookup
func (plugin *examplePlugin) lookup() (err error) {
	// /vnf-agent/agent0/vpp/config/v1/interface/egresXY
	if _, iface0, found0 := plugin.bdIdxLocal.LookupIdx("local0"); found0 {
		log.Println("local0 IPs:", iface0.Interfaces)
	}

	for i := 0; i < 10; i++ {
		// /vnf-agent/agent1/vpp/config/v1/bd/ingresXY
		if _, bd1, found1 := plugin.bdIdxAgent1.LookupIdx(testing.BDMemif100011ToMemif100012.Name); found1 {
			fmt.Println("found ", testing.BDAfPacketVeth1VxlanVni5.Name, " IFaces:", bd1.Interfaces)
			break
		} else {
			time.Sleep(100 * time.Millisecond) //to be sure that the cache is updated
		}
	}
	for i := 0; i < 10; i++ {
		// /vnf-agent/agent2/vpp/config/v1/bd/ingresXY
		if _, bd2, found2 := plugin.bdIdxAgent2.LookupIdx(testing.BDMemif100011ToMemif100012.Name); found2 {
			fmt.Println("found ", testing.BDMemif100011ToMemif100012.Name, " IFaces:", bd2.Interfaces)
			break
		} else {
			time.Sleep(100 * time.Millisecond) //to be sure that the cache is updated
		}
	}

	return err
}

func (plugin *examplePlugin) Close() error {
	return nil
}
