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
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
)

// PluginName is injected as the plugin name.
// LocalFlavor.InfraDeps() will create and initialize a new flag used to make
// the plugin config file name configurable for the user (via dedicated CLI
// option and env. variable).
// The flag name is composed of the plugin name and the suffix config.FlagSuffix.
// The default (flag value) filename for the configuration file is the plugin
// name with the extension ".conf".
const PluginName = "example"

// *************************************************************************
// This file contains a PluginConfig show case:
// - plugin binds it's configuration to an example specific Conf structure
//   (see Init() to learn how the default configuration is set & how it can be
//    overridden via flags)
// - cn-infra helps by locating and parsing the configuration file
//
// ************************************************************************/

func main() {
	// Init close channel to stop the example after everything was logged.
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor
	// (combination of ExamplePlugin & Local flavor)
	flavor := ExampleFlavor{ExamplePlugin: ExamplePlugin{exampleFinished: exampleFinished}}
	plugins := flavor.Plugins()
	agent := core.NewAgent(flavor.LogRegistry().NewLogger("core"), 15*time.Second, plugins...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// ExamplePlugin demonstrates the use of injected Config plugin.
type ExamplePlugin struct {
	local.PluginInfraDeps // this field is usually injected in flavor
	*Conf                 // it is possible to set config value programmatically (can be overridden)
	exampleFinished       chan struct{}
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

// Init loads the configuration file assigned to ExamplePlugin (can be changed
// via the example-config flag).
// Loaded config is printed into the log file.
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
