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

//go:generate go-bindata-assetfs -pkg rest -o bindata.go ./templates/...

package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	"github.com/gorilla/mux"
	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
	"github.com/unrolled/render"

	"github.com/ligato/vpp-agent/plugins/rest/resturl"
	aclcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	l3plugin "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
)

// Registers access list REST handlers
func (plugin *Plugin) registerAccessListHandlers() {
	// GET IP ACLs
	plugin.registerHTTPHandler(resturl.AclIP, GET, func() (interface{}, error) {
		return plugin.aclHandler.DumpIPACL(nil)
	})
	// GET MACIP ACLs
	plugin.registerHTTPHandler(resturl.AclMACIP, GET, func() (interface{}, error) {
		return plugin.aclHandler.DumpMACIPACL(nil)
	})
}

// Registers interface REST handlers
func (plugin *Plugin) registerInterfaceHandlers() {
	// GET all interfaces
	plugin.registerHTTPHandler(resturl.Interface, GET, func() (interface{}, error) {
		return plugin.ifHandler.DumpInterfaces()
	})
	// GET loopback interfaces
	plugin.registerHTTPHandler(resturl.Loopback, GET, func() (interface{}, error) {
		return plugin.ifHandler.DumpInterfacesByType(interfaces.InterfaceType_SOFTWARE_LOOPBACK)
	})
	// GET ethernet interfaces
	plugin.registerHTTPHandler(resturl.Ethernet, GET, func() (interface{}, error) {
		return plugin.ifHandler.DumpInterfacesByType(interfaces.InterfaceType_ETHERNET_CSMACD)
	})
	// GET memif interfaces
	plugin.registerHTTPHandler(resturl.Memif, GET, func() (interface{}, error) {
		return plugin.ifHandler.DumpInterfacesByType(interfaces.InterfaceType_MEMORY_INTERFACE)
	})
	// GET tap interfaces
	plugin.registerHTTPHandler(resturl.Tap, GET, func() (interface{}, error) {
		return plugin.ifHandler.DumpInterfacesByType(interfaces.InterfaceType_TAP_INTERFACE)
	})
	// GET af-packet interfaces
	plugin.registerHTTPHandler(resturl.AfPacket, GET, func() (interface{}, error) {
		return plugin.ifHandler.DumpInterfacesByType(interfaces.InterfaceType_AF_PACKET_INTERFACE)
	})
	// GET VxLAN interfaces
	plugin.registerHTTPHandler(resturl.VxLan, GET, func() (interface{}, error) {
		return plugin.ifHandler.DumpInterfacesByType(interfaces.InterfaceType_VXLAN_TUNNEL)
	})
}

func (plugin *Plugin) registerBfdHandlers() {
	// GET BFD configuration
	plugin.registerHTTPHandler(resturl.BfdUrl, GET, func() (interface{}, error) {
		return plugin.bfdHandler.DumpBfdSingleHop()
	})
	// GET BFD sessions
	plugin.registerHTTPHandler(resturl.BfdSession, GET, func() (interface{}, error) {
		return plugin.bfdHandler.DumpBfdSessions()
	})
	// GET BFD authentication keys
	plugin.registerHTTPHandler(resturl.BfdAuthKey, GET, func() (interface{}, error) {
		return plugin.bfdHandler.DumpBfdAuthKeys()
	})
}

// Registers L2 plugin REST handlers
func (plugin *Plugin) registerL2Handlers() {
	// GET bridge domain IDs
	plugin.registerHTTPHandler(resturl.BdId, GET, func() (interface{}, error) {
		return plugin.bdHandler.DumpBridgeDomainIDs()
	})
	// GET bridge domains
	plugin.registerHTTPHandler(resturl.Bd, GET, func() (interface{}, error) {
		return plugin.bdHandler.DumpBridgeDomains()
	})
	// GET FIB entries
	plugin.registerHTTPHandler(resturl.Fib, GET, func() (interface{}, error) {
		return plugin.fibHandler.DumpFIBTableEntries()
	})
	// GET cross connects
	plugin.registerHTTPHandler(resturl.Xc, GET, func() (interface{}, error) {
		return plugin.xcHandler.DumpXConnectPairs()
	})
}

// Registers L3 plugin REST handlers
func (plugin *Plugin) registerL3Handlers() {
	// GET static routes
	plugin.registerHTTPHandler(resturl.Routes, GET, func() (interface{}, error) {
		return plugin.rtHandler.DumpStaticRoutes()
	})
}

// registerHTTPHandler is common register method for all handlers
func (plugin *Plugin) registerHTTPHandler(key, method string, f func() (interface{}, error)) {
	handlerFunc := func(formatter *render.Render) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			plugin.Lock()
			defer plugin.Unlock()

			res, err := f()
			if err != nil {
				plugin.Deps.Log.Errorf("Error: %v", err)
				errStr := fmt.Sprintf("500 Internal server error: %s\n", err.Error())
				formatter.Text(w, http.StatusInternalServerError, errStr)
				return
			}
			plugin.Deps.Log.Debugf("Rest uri: %s, data: %v", key, res)
			formatter.JSON(w, http.StatusOK, res)
		}
	}
	plugin.HTTPHandlers.RegisterHTTPHandler(key, handlerFunc, method)
}

// staticRoutesGetHandler - used to get list of all static routes
func (plugin *Plugin) arpGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting list of all ARPs")

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		l3Handler, err := l3plugin.NewArpVppHandler(ch, plugin.Log, nil)
		if err != nil {
			plugin.Log.Errorf("Error creating VPP handler: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		res, err := l3Handler.DumpArpEntries()
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
			return
		}

		plugin.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

// interfaceACLGetHandler - used to get acl configuration for a particular interface
func (plugin *Plugin) interfaceACLGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting acl configuration of interface")

		vars := mux.Vars(req)
		if vars == nil {
			plugin.Log.Error("Interface software index not specified.")
			formatter.JSON(w, http.StatusNotFound, "Interface software index not specified.")
			return
		}

		plugin.Log.Infof("Received request for swIndex: %v", vars[swIndexVarName])

		swIndexuInt64, err := strconv.ParseUint(vars[swIndexVarName], 10, 32)
		if err != nil {
			plugin.Log.Error("Failed to unmarshal request body.")
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		swIndex := uint32(swIndexuInt64)
		if err != nil {
			plugin.Log.Errorf("Error creating VPP handler: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		res, err := plugin.aclHandler.DumpInterfaceIPAcls(swIndex)
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		res, err = plugin.aclHandler.DumpInterfaceMACIPAcls(swIndex)
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		plugin.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

// ipACLPostHandler - used to get acl configuration for a particular interface
func (plugin *Plugin) ipACLPostHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			plugin.Deps.Log.Error("Failed to parse request body.")
			formatter.JSON(w, http.StatusBadRequest, err)
			return
		}
		aclParam := acl.AccessLists_Acl{}
		err = json.Unmarshal(body, &aclParam)
		if err != nil {
			plugin.Deps.Log.Error("Failed to unmarshal request body.")
			formatter.JSON(w, http.StatusBadRequest, err)
			return
		}

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		var aclIndex struct {
			Idx uint32 `json:"acl_index"`
		}
		aclHandler, err := aclcalls.NewAclVppHandler(ch, ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error creating VPP handler: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		aclIndex.Idx, err = aclHandler.AddIPAcl(aclParam.Rules, aclParam.AclName)
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, aclIndex)
			return
		}

		plugin.Deps.Log.Debug(aclIndex)
		formatter.JSON(w, http.StatusOK, aclIndex)
	}
}

func (plugin *Plugin) macipACLPostHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			plugin.Log.Error("Failed to parse request body.")
			formatter.JSON(w, http.StatusBadRequest, err)
			return
		}
		aclParam := acl.AccessLists_Acl{}
		err = json.Unmarshal(body, &aclParam)
		if err != nil {
			plugin.Log.Error("Failed to unmarshal request body.")
			formatter.JSON(w, http.StatusBadRequest, err)
			return
		}

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		var aclIndex struct {
			Idx uint32 `json:"acl_index"`
		}
		aclHandler, err := aclcalls.NewAclVppHandler(ch, ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error creating VPP handler: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		aclIndex.Idx, err = aclHandler.AddMacIPAcl(aclParam.Rules, aclParam.AclName)
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, aclIndex)
			return
		}

		plugin.Log.Debug(aclIndex)
		formatter.JSON(w, http.StatusOK, aclIndex)
	}
}

// commandHandler - used to execute VPP CLI commands
func (plugin *Plugin) commandHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			plugin.Log.Error("Failed to parse request body.")
			formatter.JSON(w, http.StatusBadRequest, err)
			return
		}

		var reqParam map[string]string
		err = json.Unmarshal(body, &reqParam)
		if err != nil {
			plugin.Log.Error("Failed to unmarshal request body.")
			formatter.JSON(w, http.StatusBadRequest, err)
			return
		}

		command, ok := reqParam["vppclicommand"]
		if !ok || command == "" {
			plugin.Log.Error("vppclicommand parameter missing or empty")
			formatter.JSON(w, http.StatusBadRequest, "vppclicommand parameter missing or empty")
			return
		}

		plugin.Log.Debugf("VPPCLI command: %v", command)

		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		r := &vpe.CliInband{
			Length: uint32(len(command)),
			Cmd:    []byte(command),
		}
		reply := &vpe.CliInbandReply{}
		err = ch.SendRequest(r).ReceiveReply(reply)
		if err != nil {
			err = fmt.Errorf("Sending request failed: %v", err)
			plugin.Log.Error(err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		} else if reply.Retval > 0 {
			err = fmt.Errorf("Request returned error code: %v", reply.Retval)
			plugin.Log.Error(err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		plugin.Log.Debugf("VPPCLI response: %s", reply.Reply)
		formatter.Text(w, http.StatusOK, string(reply.Reply))
	}
}

func (plugin *Plugin) sendCommand(ch govppapi.Channel, command string) ([]byte, error) {
	r := &vpe.CliInband{
		Length: uint32(len(command)),
		Cmd:    []byte(command),
	}

	reply := &vpe.CliInbandReply{}
	if err := ch.SendRequest(r).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("Sending request failed: %v", err)
	} else if reply.Retval > 0 {
		return nil, fmt.Errorf("Request returned error code: %v", reply.Retval)
	}

	return reply.Reply[:reply.Length], nil
}

// telemetryHandler - returns various telemetry data
func (plugin *Plugin) telemetryHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		type cmdOut struct {
			Command string
			Output  interface{}
		}
		var cmdOuts []cmdOut

		var runCmd = func(command string) {
			out, err := plugin.sendCommand(ch, command)
			if err != nil {
				plugin.Log.Errorf("Sending command failed: %v", err)
				formatter.JSON(w, http.StatusInternalServerError, err)
				return
			}
			cmdOuts = append(cmdOuts, cmdOut{
				Command: command,
				Output:  string(out),
			})
		}

		runCmd("show node counters")
		runCmd("show runtime")
		runCmd("show buffers")
		runCmd("show memory")
		runCmd("show ip fib")
		runCmd("show ip6 fib")

		formatter.JSON(w, http.StatusOK, cmdOuts)
	}
}

// telemetryMemoryHandler - returns various telemetry data
func (plugin *Plugin) telemetryMemoryHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		info, err := vppcalls.GetMemory(ch)
		if err != nil {
			plugin.Log.Errorf("Sending command failed: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		formatter.JSON(w, http.StatusOK, info)
	}
}

// telemetryHandler - returns various telemetry data
func (plugin *Plugin) telemetryRuntimeHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		runtimeInfo, err := vppcalls.GetRuntimeInfo(ch)
		if err != nil {
			plugin.Log.Errorf("Sending command failed: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		formatter.JSON(w, http.StatusOK, runtimeInfo)
	}
}

// telemetryHandler - returns various telemetry data
func (plugin *Plugin) telemetryNodeCountHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		nodeCounters, err := vppcalls.GetNodeCounters(ch)
		if err != nil {
			plugin.Log.Errorf("Sending command failed: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		formatter.JSON(w, http.StatusOK, nodeCounters)
	}
}

// indexHandler - used to get index page
func (plugin *Plugin) indexHandler(formatter *render.Render) http.HandlerFunc {
	r := render.New(render.Options{
		Directory:  "templates",
		Asset:      Asset,
		AssetNames: AssetNames,
	})
	return func(w http.ResponseWriter, req *http.Request) {
		plugin.Log.Debugf("%v - %s %q", req.RemoteAddr, req.Method, req.URL)

		r.HTML(w, http.StatusOK, "index", plugin.indexItems)
	}
}
