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

	"github.com/ligato/cn-infra/agent"
)

func main() {
	// Create an instance of our plugin.
	p := new(HelloWorld)

	// Create new agent with our plugin instance.
	a := agent.NewAgent(agent.Plugins(p))

	// Run starts the agent with plugins, wait until shutdown
	// and then stops the agent and its plugins.
	if err := a.Run(); err != nil {
		log.Fatalln(err)
	}
}

// HelloWorld represents our plugin.
type HelloWorld struct{}

// String is used to identify the plugin by giving it name.
func (p *HelloWorld) String() string {
	return "HelloWorld"
}

// Init is executed on agent initialization.
func (p *HelloWorld) Init() error {
	log.Println("Hello World!")
	return nil
}

// AfterInit is executed after initialization of all plugins. It's optional
// and used for executing operations that require plugins to be initalized.
func (p *HelloWorld) AfterInit() error {
	log.Println("All systems go!")
	return nil
}

// Close is executed on agent shutdown.
func (p *HelloWorld) Close() error {
	log.Println("Goodbye World!")
	return nil
}
