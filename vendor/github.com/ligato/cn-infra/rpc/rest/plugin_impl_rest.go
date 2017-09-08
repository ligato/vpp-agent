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

	"fmt"

	"github.com/gorilla/mux"
	"github.com/ligato/cn-infra/datasync/grpcsync"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
	"github.com/unrolled/render"
)

const (
	// DefaultHTTPPort is used during HTTP server startup unless different port was configured
	DefaultHTTPPort = "9191"
)

var (
	httpPort string
)

// init is here only for parsing program arguments
func init() {
	flag.StringVar(&httpPort, "http-port", DefaultHTTPPort,
		"Listen port for the Agent's HTTP server.")
}

// Plugin implements the Plugin interface.
type Plugin struct {
	Deps

	// Used mainly for testing purposes
	listenAndServe ListenAndServe

	server     io.Closer
	mx         *mux.Router
	formatter  *render.Render
	grpcServer *grpcsync.Adapter
}

// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	local.PluginLogDeps // inject

	// Used to simplify if not whole config needs to be configured
	HTTPport string //inject optionally
	// Config is a rich alternative comparing to HTTPport
	// TODO Config *Config
}

// Init is entry point called by Agent Core
// - It prepares Gorilla MUX HTTP Router
// - registers grpc transport
func (plugin *Plugin) Init() (err error) {
	if plugin.HTTPport == "" /*TODO && plugin.Config == nil*/ {
		plugin.HTTPport = httpPort
	}

	plugin.mx = mux.NewRouter()
	plugin.formatter = render.New(render.Options{
		IndentJSON: true,
	})

	//TODO separate plugin:
	//plugin.grpcServer = grpcsync.NewAdapter()
	//plugin.Debug("grpctransp: ", plugin.grpcServer)
	//err = datasync.RegisterTransport(&syncbase.Adapter{Watcher: plugin.grpcServer})

	return err
}

// RegisterHTTPHandler propagates to Gorilla mux
func (plugin *Plugin) RegisterHTTPHandler(path string,
	handler func(formatter *render.Render) http.HandlerFunc,
	methods ...string) *mux.Route {
	return plugin.mx.HandleFunc(path, handler(plugin.formatter)).Methods(methods...)
}

// AfterInit starts the HTTP server
func (plugin *Plugin) AfterInit() (err error) {
	var cfgCopy Config
	/*TODO if plugin.Config != nil {
		cfgCopy = *plugin.Config
	}*/

	if cfgCopy.Endpoint == "" {
		cfgCopy.Endpoint = fmt.Sprintf("0.0.0.0:%s", plugin.HTTPport)
	}

	if plugin.listenAndServe != nil {
		plugin.server, err = plugin.listenAndServe(cfgCopy, plugin.mx)
	} else {
		plugin.Log.Info("Listening on http://", cfgCopy.Endpoint)
		plugin.server, err = ListenAndServeHTTP(cfgCopy, plugin.mx)
	}

	return err
}

// Close cleans up the resources
func (plugin *Plugin) Close() error {
	_, err := safeclose.CloseAll(plugin.grpcServer, plugin.server)
	return err
}
