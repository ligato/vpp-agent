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
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/flavors/linuxlocal"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
	linux_if "github.com/ligato/vpp-agent/plugins/linuxplugin/ifaceidx"
	linux_intf "github.com/ligato/vpp-agent/plugins/linuxplugin/model/interfaces"
)

// *************************************************************************
// This file contains examples of linux plugin name-to-index cache operations
//
// Two more transport adapters for different agents are registered using
// OfDifferentAgent() and their interface name-to-idx mapping is cached
// with linux_if.Cache() as a new map
//
// These new maps are watched and all new events are logged.
//
// VETH interfaces are then created on agents using both of the transports and
// data presence in cached name-to-idx map is verified
// ************************************************************************/

// Init sets the default logging level
func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logging.InfoLevel)
}

/********
 * Main *
 ********/

// Start Agent plugins selected for this example
func main() {
	// Init close channel to stop the example
	closeChannel := make(chan struct{}, 1)

	flavor := linuxlocal.Flavor{}
	// Example plugin and dependencies
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{
		Linux: &flavor.Linux,
	}}

	// Create new agent
	agentVar := core.NewAgent(log.StandardLogger(), 15*time.Second, append(flavor.Plugins(), examplePlugin)...)


	// End when the idx_veth_cache example is finished
	go closeExample("idx_veth_cache example finished", closeChannel)

	core.EventLoopWithInterrupt(agentVar, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(25 * time.Second)
	log.Info(message)
	closeChannel <- struct{}{}
}

/******************
 * Example plugin *
 ******************/

// PluginID of example plugin
const PluginID core.PluginName = "example-plugin"

// ExamplePlugin demonstrates the use of the name-to-idx cache in linux plugin
type ExamplePlugin struct {
	// Linux plugin dependency
	Linux			 *linuxplugin.Plugin

	// Other agents transport
	agent1           datasync.TransportAdapter
	agent2           datasync.TransportAdapter

	linuxIfIdxLocal  linux_if.LinuxIfIndex
	linuxIfIdxAgent1 linux_if.LinuxIfIndex
	linuxIfIdxAgent2 linux_if.LinuxIfIndex
	wg               sync.WaitGroup
	cancel           context.CancelFunc
}

// Init initializes example plugin
func (plugin *ExamplePlugin) Init() error {
	// Access DB of agent1 and agent2
	plugin.agent1 = datasync.OfDifferentAgent("agent1")
	plugin.agent2 = datasync.OfDifferentAgent("agent2")
	// Receive linux interfaces mapping
	if plugin.Linux != nil {
		plugin.linuxIfIdxLocal = plugin.Linux.GetLinuxIfIndexes()
	} else {
		return fmt.Errorf("Linux plugin not initialized")
	}
	// Cache the agent1/agent2 name-to-idx mapping to separate mapping within plugin example
	plugin.linuxIfIdxAgent1 = linux_if.Cache(plugin.agent1, PluginID)
	plugin.linuxIfIdxAgent2 = linux_if.Cache(plugin.agent2, PluginID)
	// Init chan to sent watch updates
	linuxIfIdxChan := make(chan linux_if.LinuxIfIndexDto)
	// Register all agents (incl. local) to watch name-to-idx mapping changes
	plugin.linuxIfIdxLocal.WatchNameToIdx(PluginID, linuxIfIdxChan)
	plugin.linuxIfIdxAgent1.WatchNameToIdx(PluginID, linuxIfIdxChan)
	plugin.linuxIfIdxAgent2.WatchNameToIdx(PluginID, linuxIfIdxChan)

	log.Info("Initialization of the example plugin has completed")

	var err error
	err = plugin.publish()
	if err != nil {
		return err
	}
	go func() {
		err = plugin.consume(linuxIfIdxChan)
	}()

	return err
}

// Close cleans up the resources
func (plugin *ExamplePlugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	log.Info("Closed example plugin")
	return nil
}

// publish propagates example configuration to ETCD
func (plugin *ExamplePlugin) publish() error {
	log.Infof("Putting interfaces to ETCD")

	// VETH pair in default namespace
	vethDef := &veth11DefaultNs
	vethDefPeer := &veth12DefaultNs

	// Push VETH pair to agent1
	err := plugin.agent1.PublishData(linux_intf.InterfaceKey(vethDef.Name), vethDef)
	err = plugin.agent1.PublishData(linux_intf.InterfaceKey(vethDefPeer.Name), vethDefPeer)

	// VETH pair in custom namespace
	vethNs1 := &veth21Ns1
	vethNs2Peer := &veth22Ns2

	// Publish VETH pair to agent2
	err = plugin.agent2.PublishData(linux_intf.InterfaceKey(vethNs1.Name), vethDef)
	err = plugin.agent2.PublishData(linux_intf.InterfaceKey(vethNs2Peer.Name), vethNs2Peer)

	if err != nil {
		log.Errorf("Failed to apply initial Linux&VPP configuration: %v", err)
		return err
	}
	log.Info("Successfully applied initial Linux&VPP configuration")

	return err
}

// Uses the NameToIndexMapping to watch changes
func (plugin *ExamplePlugin) consume(linuxIfIdxChan chan linux_if.LinuxIfIndexDto) (err error) {
	var watching = true
	for watching {
		select {
		case swIfIdxEvent, done := <-linuxIfIdxChan:
			if !done {
				log.WithFields(logging.Fields{"RegistryTitle": swIfIdxEvent.RegistryTitle, //agent1, agent2
					"Name": swIfIdxEvent.Name,
					"Del":  swIfIdxEvent.Del,
				}).Info("Event received")
			}
		case <-time.After(10 * time.Second):
			watching = false
			log.Info("Watching stopped")
		}
	}

	success := plugin.lookup()
	if !success {
		return fmt.Errorf("idx_veth_cache example failed; one or more VETH interfaces are missing")
	}

	return nil
}

// Uses the NameToIndexMapping to lookup changes
func (plugin *ExamplePlugin) lookup() bool {
	var (
		loopback bool
		veth11   bool
		veth12   bool
		veth21   bool
		veth22   bool
	)

	// Look for loopback interface
	if _, _, loopback = plugin.linuxIfIdxLocal.LookupIdx("lo"); loopback {
		log.Info("Interface found: loopback")
	} else {
		log.Warn("Interface not found: loopback") // todo remove
	}
	// Look for VETH 11 default namespace interface on agent1
	for i := 0; i <= 3; i++ {
		if _, _, veth11 = plugin.linuxIfIdxAgent1.LookupIdx(veth11DefaultNs.Name); veth11 {
			log.Info("Interface found on agent1: veth11Def")
			break
		} else if i == 3 {
			log.Warn("Interface not found on agent1: veth11Def")
		} else {
			// Try several times in case cache is not updated yet
			time.Sleep(1 * time.Second)
			continue
		}
	}
	// Look for VETH 12 default namespace interface on agent1
	for i := 0; i <= 3; i++ {
		if _, _, veth12 = plugin.linuxIfIdxAgent1.LookupIdx(veth12DefaultNs.Name); veth12 {
			log.Info("Interface found on agent1: veth12Def")
			break
		} else if i == 3 {
			log.Warn("Interface not found on agent1: veth12Def")
		} else {
			// Try several times in case cache is not updated yet
			time.Sleep(1 * time.Second)
			continue
		}
	}
	// Look for VETH 21 ns1 namespace interface on agent2
	for i := 0; i <= 3; i++ {
		if _, _, veth21 = plugin.linuxIfIdxAgent2.LookupIdx(veth21Ns1.Name); veth21 {
			log.Info("Interface found on agent2: veth21ns1")
			break
		} else if i == 3 {
			log.Warn("Interface not found on agent2 : veth21ns1")
		} else {
			// Try several times in case cache is not updated yet
			time.Sleep(1 * time.Second)
			continue
		}
	}
	// Look for VETH 22 ns2 namespace interface on agent2
	for i := 0; i <= 3; i++ {
		if _, _, veth22 = plugin.linuxIfIdxAgent2.LookupIdx(veth22Ns2.Name); veth22 {
			log.Info("Interface found on agent2: veth22ns2")
			break
		} else if i == 3 {
			log.Warn("Interface not found on agent2: veth22ns2")
		} else {
			// Try several times in case cache is not updated yet
			time.Sleep(1 * time.Second)
			continue
		}
	}

	if loopback && veth11 && veth12 && veth21 && veth22 {
		return true
	}
	return false
}

// Interface data
var (
	// veth11DefaultNs is one end of the veth11-veth12DefaultNs VETH pair, put into the default namespace
	veth11DefaultNs = linux_intf.LinuxInterfaces_Interface{
		Name:    "veth1",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth12DefaultNs",
		},
		IpAddresses: []string{"10.0.0.1/24"},
	}
	// veth12DefaultNs is one end of the veth11-veth12DefaultNs VETH pair, put into the default namespace
	veth12DefaultNs = linux_intf.LinuxInterfaces_Interface{
		Name:    "veth12DefaultNs",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth11",
		},
	}
	// veth11DefaultNs is one end of the veth21-veth22 VETH pair, put into the ns1
	veth21Ns1 = linux_intf.LinuxInterfaces_Interface{
		Name:    "veth11",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth12DefaultNs",
		},
		IpAddresses: []string{"10.0.0.1/24"},
		Namespace: &linux_intf.LinuxInterfaces_Interface_Namespace{
			Type: linux_intf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
			Name: "ns1",
		},
	}
	// veth22Ns2 is one end of the veth21-veth22 VETH pair, put into the namespace "ns2"
	veth22Ns2 = linux_intf.LinuxInterfaces_Interface{
		Name:    "veth21",
		Type:    linux_intf.LinuxInterfaces_VETH,
		Enabled: true,
		Veth: &linux_intf.LinuxInterfaces_Interface_Veth{
			PeerIfName: "veth22",
		},
		IpAddresses: []string{"10.0.0.2/24"},
		Namespace: &linux_intf.LinuxInterfaces_Interface_Namespace{
			Type: linux_intf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
			Name: "ns2",
		},
	}
)
