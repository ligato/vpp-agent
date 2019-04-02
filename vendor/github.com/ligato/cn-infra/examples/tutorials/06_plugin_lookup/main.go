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
	"github.com/ligato/cn-infra/agent"
	"log"
)

func main() {
	// Create new agent with our plugin instance.
	a := agent.NewAgent(agent.AllPlugins(New()))

	// Run starts the agent with plugins, wait until shutdown
	// and then stops the agent and its plugins.
	if err := a.Run(); err != nil {
		log.Fatalln(err)
	}
}

// Agent is a top-level plugin used as starter for other plugins
type Agent struct {
	Hw *HelloWorld
	Hu *HelloUniverse
}

// Init is an implementation of the plugin interface
func (p *Agent) Init() error {
	return nil
}

// Close is an implementation of the plugin interface
func (p *Agent) Close() error {
	return nil
}

// String is an implementation of the plugin interface
func (p *Agent) String() string {
	return "AgentPlugin"
}

// New returns top-level plugin object with defined plugins and their dependencies
func New() *Agent {
	hw := &HelloWorld{}
	hu := &HelloUniverse{}

	hw.Universe = hu
	hu.World = hw

	return &Agent{
		Hu: hu,
		Hw: hw,
	}
}
