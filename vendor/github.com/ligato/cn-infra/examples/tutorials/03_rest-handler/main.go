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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/unrolled/render"
)

func main() {
	// Create an instance of our plugin using its constructor.
	p := NewMyPlugin()

	// Create new agent with our plugin instance.
	a := agent.NewAgent(agent.AllPlugins(p))

	// Run starts the agent with plugins, wait until shutdown
	// and then stops the agent and its plugins.
	if err := a.Run(); err != nil {
		logging.DefaultLogger.Error(err)
	}
}

// MyPlugin represents our plugin.
type MyPlugin struct {
	infra.PluginDeps
	REST rest.HTTPHandlers
}

// NewMyPlugin is a constructor for our MyPlugin plugin.
func NewMyPlugin() *MyPlugin {
	p := new(MyPlugin)
	p.SetName("myplugin")
	p.Setup()
	// This sets the REST to default instance of rest plugin.
	p.REST = &rest.DefaultPlugin
	return p
}

// Init is executed on agent initialization.
func (p *MyPlugin) Init() error {
	p.Log.Debug("Registering handler..")

	p.REST.RegisterHTTPHandler("/greeting", p.fooHandler, "POST")

	return nil
}

func (p *MyPlugin) fooHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			p.Log.Errorf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		formatter.Text(w, http.StatusOK, fmt.Sprintf("Hello %s", body))
	}
}
