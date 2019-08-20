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
	"net/http"

	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/unrolled/render"
)

// registerHandlers registers all supported REST APIs.
func (p *Plugin) registerHandlers(http rest.HTTPHandlers) {
	if http == nil {
		p.Log.Debug("No http handler provided, skipping registration of REST handlers")
		return
	}
	http.RegisterHTTPHandler("/govppmux/stats", p.statsHandler, "GET")
}

func (p *Plugin) statsHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := formatter.JSON(w, http.StatusOK, GetStats()); err != nil {
			p.Log.Warnf("stats handler errored: %v", err)
		}
	}
}
