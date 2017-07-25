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
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/testing"
	"github.com/ligato/vpp-agent/flavours/vpp"
	"time"
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

	// End when the idx_iface_cache example is finished
	go closeExample("idx_iface_cache example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(12 * time.Second)
	log.Info(message)
	closeChannel <- struct{}{}
}

// PluginID of example plugin
const PluginID core.PluginName = "example-plugin"

// used for demonstration of SwIfIndexes - see Init()
type examplePlugin struct {
	agent1        datasync.TransportAdapter
	agent2        datasync.TransportAdapter
	swIfIdxLocal  ifaceidx.SwIfIndex
	swIfIdxAgent1 ifaceidx.SwIfIndex
	swIfIdxAgent2 ifaceidx.SwIfIndex
}

// initialize transport & SwIfIndexes then watch, publish & lookup
func (plugin *examplePlugin) Init() (err error) {
	plugin.agent1 = datasync.OfDifferentAgent("agent1" /*TODO "rpd", "vschxy"*/)
	plugin.agent2 = datasync.OfDifferentAgent("agent2")
	// /vnf-agent/agent0/vpp/config/v1/interface/
	plugin.swIfIdxLocal = defaultplugins.GetSwIfIndexes()
	// /vnf-agent/agent1/vpp/config/v1/interface/
	plugin.swIfIdxAgent1 = ifaceidx.Cache(plugin.agent1, PluginID)
	// /vnf-agent/agent2/vpp/config/v1/interface/
	plugin.swIfIdxAgent2 = ifaceidx.Cache(plugin.agent2, PluginID)

	if err := plugin.publish(); err != nil {
		return err
	}

	return plugin.consume()
}

// prepares test data for different agents
func (plugin *examplePlugin) publish() (err error) {
	iface1 := &testing.Memif100012
	err = plugin.agent1.PublishData(interfaces.InterfaceKey(iface1.Name), iface1)
	if err != nil {
		return err
	}
	iface2 := &testing.Memif100013
	err = plugin.agent2.PublishData(interfaces.InterfaceKey(iface2.Name), iface2)
	return err
}

// uses the NameToIndexMapping to watch changes
func (plugin *examplePlugin) consume() (err error) {
	swIfIdxChan := make(chan ifaceidx.SwIfIdxDto)
	plugin.swIfIdxLocal.WatchNameToIdx(PluginID, swIfIdxChan)
	plugin.swIfIdxAgent1.WatchNameToIdx(PluginID, swIfIdxChan)
	plugin.swIfIdxAgent2.WatchNameToIdx(PluginID, swIfIdxChan)

	go func() {
		var watching = true
		for watching {
			select {
			case swIfIdxEvent, done := <-swIfIdxChan:
				if !done {
					log.WithFields(logging.Fields{"RegistryTitle": swIfIdxEvent.RegistryTitle, //agent1, agent2
						"Name": swIfIdxEvent.Name, //ingresXY, egresXY
						"Del":  swIfIdxEvent.Del,
						"IP":   swIfIdxEvent.Metadata.IpAddresses,
					}).Info("xxx event received")
				}
			case <-time.After(5 * time.Second):
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
	if _, iface0, found0 := plugin.swIfIdxLocal.LookupIdx("local0"); found0 {
		log.Println("local0 IPs:", iface0.IpAddresses)
	}

	for i := 0; i < 10; i++ {
		// /vnf-agent/agent1/vpp/config/v1/interface/ingresXY
		// Possible usage: for example we need to configure L3 routes to the VPP and we need to know
		// the next hop IP address
		if _, iface1, found1 := plugin.swIfIdxAgent1.LookupIdx(testing.Memif100012.Name); found1 {
			fmt.Println("found ", testing.Memif100012.Name, " IPs:", iface1.IpAddresses)
			break
		} else {
			time.Sleep(100 * time.Millisecond) //to be sure that the cache is updated
		}
	}
	for i := 0; i < 10; i++ {
		// /vnf-agent/agent2/vpp/config/v1/interface/ingresXY
		if _, iface2, found2 := plugin.swIfIdxAgent2.LookupIdx(testing.Memif100013.Name); found2 {
			fmt.Println("found ", testing.Memif100013.Name, " IPs:", iface2.IpAddresses)
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
