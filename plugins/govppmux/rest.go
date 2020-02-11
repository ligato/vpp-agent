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

package govppmux

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/rpc"

	"github.com/unrolled/render"
	"go.ligato.io/cn-infra/v2/rpc/rest"
)

// registerHandlers registers all supported REST APIs.
func (p *Plugin) registerHandlers(http rest.HTTPHandlers) {
	if http == nil {
		p.Log.Debug("No http handler provided, skipping registration of REST handlers")
		return
	}
	http.RegisterHTTPHandler("/govppmux/stats", p.statsHandler, "GET")
	http.RegisterHTTPHandler(rpc.DefaultRPCPath, p.proxyHandler, "CONNECT")
	http.RegisterHTTPHandler("/vpp/command", p.cliCommandHandler, "POST")
}

func (p *Plugin) statsHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := formatter.JSON(w, http.StatusOK, GetStats()); err != nil {
			p.Log.Warnf("stats handler errored: %v", err)
		}
	}
}

func (p *Plugin) proxyHandler(_ *render.Render) http.HandlerFunc {
	if !p.config.ProxyEnabled {
		return func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "VPP proxy not enabled", http.StatusServiceUnavailable)
		}
	}
	return func(w http.ResponseWriter, req *http.Request) {
		p.proxy.ServeHTTP(w, req)
	}
}

func (p *Plugin) cliCommandHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			errMsg := fmt.Sprintf("400 Bad request: failed to parse request body: %v", err)
			_ = formatter.JSON(w, http.StatusBadRequest, errMsg)
			return
		}
		var reqParam map[string]string

		if err = json.Unmarshal(body, &reqParam); err != nil {
			errMsg := fmt.Sprintf("400 Bad request: failed to unmarshall request body: %v\n", err)
			_ = formatter.JSON(w, http.StatusBadRequest, errMsg)
			return
		}
		command, ok := reqParam["vppclicommand"]
		if !ok || command == "" {
			errMsg := fmt.Sprintf("400 Bad request: vppclicommand parameter missing or empty\n")
			_ = formatter.JSON(w, http.StatusBadRequest, errMsg)
			return
		}

		p.Log.Debugf("VPPCLI command: %v", command)
		reply, err := p.vpeHandler.RunCli(req.Context(), command)
		if err != nil {
			errMsg := fmt.Sprintf("500 Internal server error: sending request failed: %v\n", err)
			_ = formatter.JSON(w, http.StatusInternalServerError, errMsg)
			return
		}

		p.Log.Debugf("VPPCLI response: %s", reply)
		_ = formatter.JSON(w, http.StatusOK, reply)
	}
}
