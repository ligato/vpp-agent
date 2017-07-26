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
	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/core"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/safeclose"
	bin_api "github.com/ligato/vpp-agent/defaultplugins/l2plugin/bin_api/l2"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/flavours/vpp"
	"github.com/ligato/vpp-agent/govppmux"
	"time"
)

// *************************************************************************
// This file contains examples of GOVPP operations, conversion of a proto
// data to a binary api message and demonstration of how to send the message
// to the VPP with:
//
// requestContext = goVppChannel.SendRequest(requestMessage)
// requestContext.ReceiveReply(replyMessage)
//
// Note: this example shows how to work with VPP, so a real proto message
// structure is used (bridge domains).
// ************************************************************************/

/********
 * Main *
 ********/

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log), GOVPP, resync plugin and example plugin which demonstrates GOVPP call functionality.
func main() {
	// Init close channel to stop the example
	closeChannel := make(chan struct{}, 1)

	f := vpp.Flavour{}

	// Example plugin (GOVPP call)
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{}}

	// Create new agent
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(f.Plugins(), examplePlugin)...)

	// End when the GOVPP example is finished
	go closeExample("GOVPP call example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(10 * time.Second)
	log.Info(message)
	closeChannel <- struct{}{}
}

/**********************
 * Example plugin API *
 **********************/

// PluginID of the custom govpp_call plugin
const PluginID core.PluginName = "example-plugin"

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct {
	exampleConfigurator *ExampleConfigurator // Plugin configurator
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() error {
	// Initialize configurator
	plugin.exampleConfigurator = &ExampleConfigurator{
		exampleIDSeq: 1, // Example ID is plugin-specific number used as a data index
	}

	// Now initialize the plugin configurator
	err := plugin.exampleConfigurator.Init()

	log.Info("Initialization of the custom plugin for the GOVPP call example is completed")

	return err
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	plugin.exampleConfigurator.Close()
	return nil
}

/*************************
 * Example plugin config *
 *************************/

// ExampleConfigurator usually initializes configuration-specific fields or other tasks (e.g. defines GOVPP channels
// if they are used, checks VPP message compatibility etc.)
type ExampleConfigurator struct {
	exampleIDSeq uint32       // Plugin-specific ID initialization
	vppChannel   *api.Channel // Vpp channel to communicate with VPP
}

// Init members of configurator
func (configurator *ExampleConfigurator) Init() (err error) {
	// NewAPIChannel returns a new API channel for communication with VPP via govpp core. It uses default buffer
	// sizes for the request and reply Go channels
	configurator.vppChannel, err = govppmux.NewAPIChannel()

	log.Info("Default plugin configurator ready")

	// Make VPP call
	go configurator.VppCall()
	// Make VPP call

	return err
}

// Close function for example plugin
func (configurator *ExampleConfigurator) Close() {
	safeclose.Close(configurator.vppChannel)
}

/***********
 * VPPCall *
 ***********/

// VppCall uses created data to convert it to the binary api call. In the example, a bridge domain data are built and
// transformed to the BridgeDomainAddDel binary api call which is then sent to the VPP
func (configurator *ExampleConfigurator) VppCall() {
	time.Sleep(3 * time.Second)

	// Prepare a simple data
	log.Info("Preparing data ...")
	bds1 := buildData("br1")
	bds2 := buildData("br2")
	bds3 := buildData("br3")

	// Prepare binary api message from the data
	req1 := buildBinapiMessage(bds1, configurator.exampleIDSeq)
	configurator.exampleIDSeq++ // Change (raise) index to ensure every message uses unique ID
	req2 := buildBinapiMessage(bds2, configurator.exampleIDSeq)
	configurator.exampleIDSeq++
	req3 := buildBinapiMessage(bds3, configurator.exampleIDSeq)
	configurator.exampleIDSeq++

	// Generic bin api reply (request: BridgeDomainAddDel)
	reply := &bin_api.BridgeDomainAddDelReply{}

	log.Info("Sending data to VPP ...")

	// 1. Send the request and receive a reply directly (in one line)
	configurator.vppChannel.SendRequest(req1).ReceiveReply(reply)

	// 2. Send multiple different requests. Every request returns it's own request context
	reqCtx2 := configurator.vppChannel.SendRequest(req2)
	reqCtx3 := configurator.vppChannel.SendRequest(req3)
	// The context can be used later to get reply
	reqCtx2.ReceiveReply(reply)
	reqCtx3.ReceiveReply(reply)

	log.Info("Data sent to VPP")
}

// Auxiliary function to build bridge domain data
func buildData(name string) *l2.BridgeDomains {
	return &l2.BridgeDomains{
		BridgeDomains: []*l2.BridgeDomains_BridgeDomain{
			{
				Name:                name,
				Flood:               false,
				UnknownUnicastFlood: true,
				Forward:             true,
				Learn:               true,
				ArpTermination:      true,
				MacAge:              0,
				Interfaces: []*l2.BridgeDomains_BridgeDomain_Interfaces{
					{
						Name: "memif1",
					},
				},
			},
		},
	}
}

// Auxiliary method to transform agent model data to binary api format
func buildBinapiMessage(data *l2.BridgeDomains, id uint32) *bin_api.BridgeDomainAddDel {
	req := &bin_api.BridgeDomainAddDel{}
	req.IsAdd = 1
	req.BdID = id
	req.Flood = boolToInt(data.BridgeDomains[0].Flood)
	req.UuFlood = boolToInt(data.BridgeDomains[0].UnknownUnicastFlood)
	req.Forward = boolToInt(data.BridgeDomains[0].Forward)
	req.Learn = boolToInt(data.BridgeDomains[0].Learn)
	req.ArpTerm = boolToInt(data.BridgeDomains[0].ArpTermination)
	req.MacAge = uint8(data.BridgeDomains[0].MacAge)

	return req
}

func boolToInt(input bool) uint8 {
	if input {
		return uint8(1)
	}
	return uint8(0)
}
