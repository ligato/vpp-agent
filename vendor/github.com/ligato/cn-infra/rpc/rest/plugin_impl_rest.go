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

package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"

	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/utils/safeclose"
)

// Plugin struct holds all plugin-related data.
type Plugin struct {
	Deps

	*Config

	server    *http.Server
	mx        *mux.Router
	formatter *render.Render
}

// Deps lists the dependencies of the Rest plugin.
type Deps struct {
	infra.PluginDeps

	// Authenticator can be injected in a flavor inject method.
	// If there is no authenticator injected and config contains
	// user password, the default staticAuthenticator is instantiated.
	// By default the authenticator is disabled.
	Authenticator BasicHTTPAuthenticator //inject
}

// Init is the plugin entry point called by Agent Core
// - It prepares Gorilla MUX HTTP Router
func (plugin *Plugin) Init() (err error) {
	if plugin.Config == nil {
		plugin.Config = DefaultConfig()
	}
	if err := PluginConfig(plugin.Cfg, plugin.Config, plugin.PluginName); err != nil {
		return err
	}

	// if there is no injected authenticator and there are credentials defined in the config file
	// instantiate staticAuthenticator otherwise do not use basic Auth
	if plugin.Authenticator == nil && len(plugin.Config.ClientBasicAuth) > 0 {
		plugin.Authenticator, err = newStaticAuthenticator(plugin.Config.ClientBasicAuth)
		if err != nil {
			return err
		}
	}

	plugin.mx = mux.NewRouter()
	plugin.formatter = render.New(render.Options{
		IndentJSON: true,
	})

	return err
}

// AfterInit starts the HTTP server.
func (plugin *Plugin) AfterInit() (err error) {
	cfgCopy := *plugin.Config

	var handler http.Handler = plugin.mx
	if plugin.Authenticator != nil {
		handler = auth(handler, plugin.Authenticator)
	}

	plugin.server, err = ListenAndServe(cfgCopy, handler)
	if err != nil {
		return err
	}

	if cfgCopy.UseHTTPS() {
		plugin.Log.Info("Listening on https://", cfgCopy.Endpoint)
	} else {
		plugin.Log.Info("Listening on http://", cfgCopy.Endpoint)
	}

	return nil
}

// RegisterHTTPHandler registers HTTP <handler> at the given <path>.
func (plugin *Plugin) RegisterHTTPHandler(path string, provider HandlerProvider, methods ...string) *mux.Route {
	plugin.Log.Debug("Registering handler: ", path)

	return plugin.mx.Handle(path, provider(plugin.formatter)).Methods(methods...)
}

// GetPort returns plugin configuration port
func (plugin *Plugin) GetPort() int {
	if plugin.Config != nil {
		return plugin.Config.GetPort()
	}
	return 0
}

// Close stops the HTTP server.
func (plugin *Plugin) Close() error {
	return safeclose.Close(plugin.server)
}
