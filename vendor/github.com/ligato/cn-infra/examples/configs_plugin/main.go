package main

import (
	"time"

	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/namsral/flag"
)

// PluginName is used below in dependency injection & in following constant
// to achieve consistent string values.
const PluginName = "example"

// ExampleConfFlag used as flag name (see implementation in declareFlags())
// It is used to load configuration of Example plugin.
// This flag name is calculated from the name of the plugin.
const ExampleConfFlag = PluginName + config.FlagSuffix

// ExampleConfDefault is default (flag value) - filename for the configuration.
const ExampleConfDefault = PluginName + ".conf"

// ExampleConfUsage used as flag usage (see implementation in declareFlags())
const ExampleConfUsage = "Location of the example configuration file; also set via 'EXAMPLE_CONFIG' env variable."

// *************************************************************************
// This file contains a Plugin Config show case:
// - plugin binds it's configuration to a example specific Conf structure
//   (see code how default is handled & how it can be overridden by flags)
// - cn-infra helps by locating config file (flags)
//
// ************************************************************************/

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized.
func main() {
	// Init close channel to stop the example after everything was logged
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combination of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{ExamplePlugin: ExamplePlugin{exampleFinished: exampleFinished}}
	agent := core.NewAgent(logroot.StandardLogger(), 15*time.Second, flavor.Plugins()...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// ExampleFlavor is composition of ExamplePlugin and existing flavor
type ExampleFlavor struct {
	local.FlavorLocal
	ExamplePlugin
}

// Plugins combines all Plugins in flavor to the list
func (f *ExampleFlavor) Plugins() []*core.NamedPlugin {
	if f.FlavorLocal.Inject() {
		f.ExamplePlugin.PluginInfraDeps = *f.InfraDeps(PluginName)
		flag.String(ExampleConfFlag, ExampleConfDefault, ExampleConfUsage)
	}

	return core.ListPluginsInFlavor(f)
}

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct {
	local.PluginInfraDeps // this field is usually injected in flavor
	*Conf                 // it is possible to set config value programatically (can be overriden)
	exampleFinished chan struct{}
}

// Conf - example config binding
type Conf struct {
	Field1 string
	Sleep  time.Duration
	// even nested fields are possible
}

func (conf *Conf) String() string {
	return "{Field1:" + conf.Field1 + ", Sleep:" + conf.Sleep.String() + "}"
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	plugin.Log.Info("Loading plugin config ", plugin.PluginConfig.GetConfigName())

	if plugin.Conf == nil {
		plugin.Conf = &Conf{Field1: "some default value"}
	}

	found, err := plugin.PluginConfig.GetValue(plugin.Conf)
	if err != nil {
		plugin.Log.Error("Error loading config", err)
	} else if found {
		plugin.Log.Info("Loaded plugin config - found external configuration ", plugin.PluginConfig.GetConfigName())
	} else {
		plugin.Log.Info("Loaded plugin config - default")
	}
	plugin.Log.Info("Plugin Config ", plugin.Conf)
	time.Sleep(plugin.Conf.Sleep)
	plugin.exampleFinished <- struct{}{}

	return nil
}
