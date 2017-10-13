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

package rpc

import (
	"git.fd.io/govpp.git"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
	"github.com/unrolled/render"
	"net/http"
)

//interfaceGetHandler - used to get list of all interfaces
func (plugin *RESTSvcPlugin) interfaceGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Debug("Getting list of all interfaces")
		// connect to VPP
		conn, err := govpp.Connect()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
		} else {
			// create an API channel
			ch, err := conn.NewAPIChannel()
			if err != nil {
				plugin.Deps.Log.Errorf("Error: %v", err)
				formatter.JSON(w, http.StatusInternalServerError, nil)
			} else {
				res, err := vppdump.DumpInterfaces(plugin.Deps.Log, ch, nil)
				if err != nil {
					plugin.Deps.Log.Errorf("Error: %v", err)
					formatter.JSON(w, http.StatusInternalServerError, nil)
				} else {
					plugin.Deps.Log.Debug(res)
					formatter.JSON(w, http.StatusOK, res)
				}
			}
			defer ch.Close()
		}
		defer conn.Disconnect()
	}
}
