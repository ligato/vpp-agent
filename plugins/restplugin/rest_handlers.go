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

package restplugin

import (
	"github.com/gorilla/mux"
	aclplugin "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppcalls"
	ifplugin "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
	l2plugin "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppdump"
	l3plugin "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppdump"
	"github.com/unrolled/render"
	"net/http"
	"strconv"
)

//interfaceGetHandler - used to get list of all interfaces
func (plugin *RESTAPIPlugin) interfacesGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all interfaces")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
		} else {
			res, err := ifplugin.DumpInterfaces(plugin.Deps.Log, ch, nil)
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
}

//bridgeDomainGetHandler - used to get list of all bridge domains
func (plugin *RESTAPIPlugin) bridgeDomainIdsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all bridge domain ids")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
		} else {
			res, err := l2plugin.DumpBridgeDomainIDs(plugin.Deps.Log, ch, nil)
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
}

//bridgeDomainGetHandler - used to get list of all bridge domains
func (plugin *RESTAPIPlugin) bridgeDomainsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all bridge domains")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
		} else {
			res, err := l2plugin.DumpBridgeDomains(plugin.Deps.Log, ch, nil)
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
}

//fibTableEntriesGetHandler - used to get list of all fib entries
func (plugin *RESTAPIPlugin) fibTableEntriesGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all fibs")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
		} else {
			res, err := l2plugin.DumpFIBTableEntries(plugin.Deps.Log, ch, nil)
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
}

//xconnectPairsGetHandler - used to get list of all connect pairs (transmit and receive interfaces)
func (plugin *RESTAPIPlugin) xconnectPairsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all xconnect pairs")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
		} else {
			res, err := l2plugin.DumpXConnectPairs(plugin.Deps.Log, ch, nil)
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
}

//staticRoutesGetHandler - used to get list of all static routes
func (plugin *RESTAPIPlugin) staticRoutesGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all static routes")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
		} else {
			res, err := l3plugin.DumpStaticRoutes(plugin.Deps.Log, ch, nil)
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
}

//interfaceAclPostHandler - used to get acl configuration for a particular interface
func (plugin *RESTAPIPlugin) interfaceAclPostHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all static routes")

		params := mux.Vars(req)
		if params != nil && len(params) > 0 {
			swIndexStr := params["swIndex"]
			if swIndexStr != "" {
				swIndexuInt64, err := strconv.ParseUint(swIndexStr, 10, 32)
				swIndex := uint32(swIndexuInt64)
				if err != nil {
					// create an API channel
					ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
					if err != nil {
						plugin.Deps.Log.Errorf("Error: %v", err)
						formatter.JSON(w, http.StatusInternalServerError, nil)
					} else {
						res, err := aclplugin.DumpInterface(swIndex, ch, nil)
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
			} else {
				formatter.JSON(w, http.StatusBadRequest, "swIndex parameter not found")
			}
		}
	}
}
