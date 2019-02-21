package main

import (
	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
)

func main() {
	// Create an instance of our plugin using its constructor.
	p := NewHelloWorld()

	// Create new agent with our plugin instance.
	a := agent.NewAgent(agent.Plugins(p))

	// Run starts the agent with plugins, wait until shutdown
	// and then stops the agent and its plugins.
	if err := a.Run(); err != nil {
		logging.Error(err)
	}
}

// HelloWorld represents our plugin.
type HelloWorld struct {
	// This embeds essential plugin deps into our plugin.
	infra.PluginDeps
}

// NewHelloWorld is a constructor for our HelloWorld plugin.
func NewHelloWorld() *HelloWorld {
	// Create new instance.
	p := new(HelloWorld)
	// Set the plugin name.
	p.SetName("helloworld")
	// Initialize essential plugin deps: logger and config.
	p.Setup()
	return p
}

type Config struct {
	Greeting string `json:"greeting"`
}

func defaultConfig() *Config {
	return &Config{
		Greeting: "Hello World!",
	}
}

// Init is executed on agent initialization.
func (p *HelloWorld) Init() error {
	p.Log.Debug("Loading config..")
	// Load config file.
	cfg := defaultConfig()
	found, err := p.Cfg.LoadValue(cfg)
	if err != nil {
		return err
	} else if !found {
		p.Log.Warnf("config not found")
	}

	p.Log.Info(cfg.Greeting)
	return nil
}

// Close is executed on agent shutdown.
func (p *HelloWorld) Close() error {
	p.Log.Info("Goodbye World!")
	return nil
}
