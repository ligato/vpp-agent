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
	"io"
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

	// Used mainly for testing purposes
	listenAndServe ListenAndServe

	server    io.Closer
	mx        *mux.Router
	formatter *render.Render
}

// Deps lists the dependencies of the Rest plugin.
type Deps struct {
	infra.Deps

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
	if err := PluginConfig(plugin.Deps.PluginConfig, plugin.Config, plugin.Deps.PluginName); err != nil {
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

	if plugin.listenAndServe != nil {
		plugin.server, err = plugin.listenAndServe(cfgCopy, plugin.mx)
	} else {
		if cfgCopy.UseHTTPS() {
			plugin.Log.Info("Listening on https://", cfgCopy.Endpoint)
		} else {
			plugin.Log.Info("Listening on http://", cfgCopy.Endpoint)
		}

		plugin.server, err = ListenAndServeHTTP(cfgCopy, plugin.mx)
	}

	return err
}

// RegisterHTTPHandler registers HTTP <handler> at the given <path>.
func (plugin *Plugin) RegisterHTTPHandler(path string,
	handler func(formatter *render.Render) http.HandlerFunc,
	methods ...string) *mux.Route {
	plugin.Log.Debug("Registering handler: ", path)

	if plugin.Authenticator != nil {
		return plugin.mx.HandleFunc(path, auth(handler(plugin.formatter), plugin.Authenticator)).Methods(methods...)
	}
	return plugin.mx.HandleFunc(path, handler(plugin.formatter)).Methods(methods...)

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
