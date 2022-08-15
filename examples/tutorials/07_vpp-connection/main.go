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
	"log"
	"time"

	"go.fd.io/govpp/api"
	"go.fd.io/govpp/binapi/ethernet_types"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.ligato.io/cn-infra/v2/agent"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
)

func main() {
	// Create an instance of our plugin.
	p := new(HelloWorld)
	p.GoVPPMux = &govppmux.DefaultPlugin

	// Create new agent with our plugin instance.
	a := agent.NewAgent(agent.AllPlugins(p))

	// Run starts the agent with plugins, wait until shutdown
	// and then stops the agent and its plugins.
	if err := a.Start(); err != nil {
		log.Fatalln(err)
	}

	p.syncVppCall()
	p.asyncVppCall()
	p.multiRequestCall()

	if err := a.Stop(); err != nil {
		log.Fatalln(err)
	}
}

// HelloWorld represents our plugin.
type HelloWorld struct {
	vppChan api.Channel

	GoVPPMux govppmux.API
}

// String is used to identify the plugin by giving it name.
func (p *HelloWorld) String() string {
	return "HelloWorld"
}

// Init is executed on agent initialization.
func (p *HelloWorld) Init() (err error) {
	log.Println("Hello World!")

	if p.vppChan, err = p.GoVPPMux.NewAPIChannel(); err != nil {
		panic(err)
	}

	return nil
}

func (p *HelloWorld) syncVppCall() {
	// prepare request
	mac, err := ethernet_types.ParseMacAddress("00:00:00:00:00:01")
	if err != nil {
		log.Panicf("ParseMacAddress failed: %v", err)
	}
	request := &interfaces.CreateLoopback{
		MacAddress: mac,
	}
	// prepare reply
	reply := &interfaces.CreateLoopbackReply{}
	// send request and obtain reply
	err = p.vppChan.SendRequest(request).ReceiveReply(reply)
	if err != nil {
		panic(err)
	}
	// check return value
	if reply.Retval != 0 {
		log.Panicf("Sync call loopback create returned %d", reply.Retval)
	}

	log.Printf("Sync call created loopback with index %d", reply.SwIfIndex)
}

func (p *HelloWorld) asyncVppCall() {
	// prepare requests
	mac1, err := ethernet_types.ParseMacAddress("00:00:00:00:00:02")
	if err != nil {
		log.Panicf("ParseMacAddress failed: %v", err)
	}
	mac2, err := ethernet_types.ParseMacAddress("00:00:00:00:00:03")
	if err != nil {
		log.Panicf("ParseMacAddress failed: %v", err)
	}
	request1 := &interfaces.CreateLoopback{
		MacAddress: mac1,
	}
	request2 := &interfaces.CreateLoopback{
		MacAddress: mac2,
	}

	// obtain contexts
	reqCtx1 := p.vppChan.SendRequest(request1)
	reqCtx2 := p.vppChan.SendRequest(request2)

	// wait a bit
	time.Sleep(2 * time.Second)

	// prepare replies
	reply1 := &interfaces.CreateLoopbackReply{}
	reply2 := &interfaces.CreateLoopbackReply{}

	// receive replies
	if err := reqCtx1.ReceiveReply(reply1); err != nil {
		panic(err)
	}
	if err := reqCtx2.ReceiveReply(reply2); err != nil {
		panic(err)
	}

	log.Printf("Async call created loopbacks with indexes %d and %d",
		reply1.SwIfIndex, reply2.SwIfIndex)
}

func (p *HelloWorld) multiRequestCall() {
	// prepare request
	request := &interfaces.SwInterfaceDump{}
	multiReqCtx := p.vppChan.SendMultiRequest(request)

	// read replies in the loop
	for {
		reply := &interfaces.SwInterfaceDetails{}
		last, err := multiReqCtx.ReceiveReply(reply)
		if err != nil {
			panic(err)
		}
		if last {
			break
		}
		log.Printf("received VPP interface with index %d", reply.SwIfIndex)
	}
}

// Close is executed on agent shutdown.
func (p *HelloWorld) Close() error {
	p.vppChan.Close()
	log.Println("Goodbye World!")
	return nil
}
