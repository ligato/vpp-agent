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
