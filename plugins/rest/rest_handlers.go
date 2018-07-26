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

	aclcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	l3plugin "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/model/bfd"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
)

// Registers access list REST handlers
func (plugin *Plugin) registerAccessListHandlers() error {
	// GET IP ACLs
	if err := plugin.registerHTTPHandler(acl.RestIPKey(), GET, func() (interface{}, error) {
		return plugin.aclHandler.DumpIPACL(nil)
	}); err != nil {
		return err
	}
	// GET MACIP ACLs
	if err := plugin.registerHTTPHandler(acl.RestMACIPKey(), GET, func() (interface{}, error) {
		return plugin.aclHandler.DumpMacIPAcls()
	}); err != nil {
		return err
	}
	// GET IP ACL example
	plugin.HTTPHandlers.RegisterHTTPHandler(acl.RestIPExampleKey(), plugin.exampleIpACLGetHandler, GET)
	// GET MACIP ACL example
	plugin.HTTPHandlers.RegisterHTTPHandler(acl.RestMACIPExampleKey(), plugin.exampleMacIpACLGetHandler, GET)

	return nil
}

// Registers interface REST handlers
func (plugin *Plugin) registerInterfaceHandlers() error {
	// GET all interfaces
	if err := plugin.registerHTTPHandler(interfaces.RestInterfaceKey(), GET, func() (interface{}, error) {
		return plugin.ifHandler.DumpInterfaces()
	}); err != nil {
		return err
	}
	// GET loopback interfaces
	if err := plugin.registerHTTPHandler(interfaces.RestLoopbackKey(), GET, func() (interface{}, error) {
		ifs, err := plugin.ifHandler.DumpInterfaces()
		for ifKey, ifConfig := range ifs {
			if ifConfig.Interface.Type != interfaces.InterfaceType_SOFTWARE_LOOPBACK {
				delete(ifs, ifKey)
			}
		}
		return ifs, err
	}); err != nil {
		return err
	}
	// GET ethernet interfaces
	if err := plugin.registerHTTPHandler(interfaces.RestEthernetKey(), GET, func() (interface{}, error) {
		ifs, err := plugin.ifHandler.DumpInterfaces()
		for ifKey, ifConfig := range ifs {
			if ifConfig.Interface.Type != interfaces.InterfaceType_ETHERNET_CSMACD {
				delete(ifs, ifKey)
			}
		}
		return ifs, err
	}); err != nil {
		return err
	}
	// GET memif interfaces
	if err := plugin.registerHTTPHandler(interfaces.RestMemifKey(), GET, func() (interface{}, error) {
		ifs, err := plugin.ifHandler.DumpInterfaces()
		for ifKey, ifConfig := range ifs {
			if ifConfig.Interface.Type != interfaces.InterfaceType_MEMORY_INTERFACE {
				delete(ifs, ifKey)
			}
		}
		return ifs, err
	}); err != nil {
		return err
	}
	// GET tap interfaces
	if err := plugin.registerHTTPHandler(interfaces.RestTapKey(), GET, func() (interface{}, error) {
		ifs, err := plugin.ifHandler.DumpInterfaces()
		for ifKey, ifConfig := range ifs {
			if ifConfig.Interface.Type != interfaces.InterfaceType_TAP_INTERFACE {
				delete(ifs, ifKey)
			}
		}
		return ifs, err
	}); err != nil {
		return err
	}
	// GET af-packet interfaces
	if err := plugin.registerHTTPHandler(interfaces.RestAfPAcketKey(), GET, func() (interface{}, error) {
		ifs, err := plugin.ifHandler.DumpInterfaces()
		for ifKey, ifConfig := range ifs {
			if ifConfig.Interface.Type != interfaces.InterfaceType_AF_PACKET_INTERFACE {
				delete(ifs, ifKey)
			}
		}
		return ifs, err
	}); err != nil {
		return err
	}
	// GET VxLAN interfaces
	if err := plugin.registerHTTPHandler(interfaces.RestVxLanKey(), GET, func() (interface{}, error) {
		ifs, err := plugin.ifHandler.DumpInterfaces()
		for ifKey, ifConfig := range ifs {
			if ifConfig.Interface.Type != interfaces.InterfaceType_VXLAN_TUNNEL {
				delete(ifs, ifKey)
			}
		}
		return ifs, err
	}); err != nil {
		return err
	}

	return nil
}

func (plugin *Plugin) registerBfdHandlers() error {
	// GET BFD configuration
	if err := plugin.registerHTTPHandler(bfd.RestBfdKey(), GET, func() (interface{}, error) {
		return plugin.bfdHandler.DumpBfdSingleHop()
	}); err != nil {
		return err
	}
	// GET BFD sessions
	if err := plugin.registerHTTPHandler(bfd.RestSessionKey(), GET, func() (interface{}, error) {
		return plugin.bfdHandler.DumpBfdSessions()
	}); err != nil {
		return err
	}
	// GET BFD authentication keys
	if err := plugin.registerHTTPHandler(bfd.RestAuthKeysKey(), GET, func() (interface{}, error) {
		return plugin.bfdHandler.DumpBfdAuthKeys()
	}); err != nil {
		return err
	}

	return nil
}

// Registers L2 plugin REST handlers
func (plugin *Plugin) registerL2Handlers() error {
	// GET bridge domain IDs
	if err := plugin.registerHTTPHandler(l2.RestBridgeDomainIDKey(), GET, func() (interface{}, error) {
		return plugin.bdHandler.DumpBridgeDomainIDs()
	}); err != nil {
		return err
	}
	// GET bridge domains
	if err := plugin.registerHTTPHandler(l2.RestBridgeDomainKey(), GET, func() (interface{}, error) {
		return plugin.bdHandler.DumpBridgeDomains()
	}); err != nil {
		return err
	}
	// GET FIB entries
	if err := plugin.registerHTTPHandler(l2.RestFibKey(), GET, func() (interface{}, error) {
		return plugin.fibHandler.DumpFIBTableEntries()
	}); err != nil {
		return err
	}
	// GET cross connects
	if err := plugin.registerHTTPHandler(l2.RestXConnectKey(), GET, func() (interface{}, error) {
		return plugin.xcHandler.DumpXConnectPairs()
	}); err != nil {
		return err
	}

	return nil
}

// registerHTTPHandler is common register method for all handlers
func (plugin *Plugin) registerHTTPHandler(key, method string, f func() (interface{}, error)) error {
	var err error
	handlerFunc := func(formatter *render.Render) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			res, err := f()
			if err != nil {
				plugin.Deps.Log.Errorf("Error: %v", err)
				err = formatter.JSON(w, http.StatusInternalServerError, err)
			}

			plugin.Deps.Log.Debug(res)
			formatter.JSON(w, http.StatusOK, res)
		}
	}
	if err != nil {
		return err
	}
	plugin.HTTPHandlers.RegisterHTTPHandler(key, handlerFunc, method)
	return nil
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

// staticRoutesGetHandler - used to get list of all static routes
func (plugin *Plugin) staticRoutesGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting list of all static routes")

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		l3Handler, err := l3plugin.NewRouteVppHandler(ch, plugin.Log, nil)
		if err != nil {
			plugin.Log.Errorf("Error creating VPP handler: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		res, err := l3Handler.DumpStaticRoutes()
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

// ipACLGetHandler - used to get configuration of IP ACLs
func (plugin *Plugin) ipACLGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting acls")

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()
		aclHandler, err := aclcalls.NewAclVppHandler(ch, ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error creating VPP handler: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		res, err := aclHandler.DumpIPACL(nil)
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		plugin.Deps.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

func (plugin *Plugin) macipACLGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		plugin.Deps.Log.Info("Getting macip acls")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		aclHandler, err := aclcalls.NewAclVppHandler(ch, ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error creating VPP handler: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		res, err := aclHandler.DumpMACIPACL(nil)
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		plugin.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

// exampleACLGetHandler - used to get an example ACL configuration
func (plugin *Plugin) exampleIpACLGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting example acl")

		ipRule := &acl.AccessLists_Acl_Rule_Match_IpRule{
			Ip: &acl.AccessLists_Acl_Rule_Match_IpRule_Ip{
				DestinationNetwork: "1.2.3.4/24",
				SourceNetwork:      "5.6.7.8/24",
			},
			Tcp: &acl.AccessLists_Acl_Rule_Match_IpRule_Tcp{
				DestinationPortRange: &acl.AccessLists_Acl_Rule_Match_IpRule_PortRange{
					LowerPort: 80,
					UpperPort: 8080,
				},
				SourcePortRange: &acl.AccessLists_Acl_Rule_Match_IpRule_PortRange{
					LowerPort: 10,
					UpperPort: 1010,
				},
				TcpFlagsMask:  0xFF,
				TcpFlagsValue: 9,
			},
		}

		rule := &acl.AccessLists_Acl_Rule{
			Match: &acl.AccessLists_Acl_Rule_Match{
				IpRule: ipRule,
			},
			AclAction: acl.AclAction_PERMIT,
		}

		aclRes := acl.AccessLists_Acl{
			AclName: "example",
			Rules:   []*acl.AccessLists_Acl_Rule{rule},
		}

		plugin.Log.Debug(aclRes)
		formatter.JSON(w, http.StatusOK, aclRes)
	}
}

func (plugin *Plugin) exampleMacIpACLGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		plugin.Deps.Log.Info("Getting example macip acl")

		macipRule := &acl.AccessLists_Acl_Rule_Match_MacIpRule{
			SourceAddress:        "192.168.0.1",
			SourceAddressPrefix:  uint32(16),
			SourceMacAddress:     "02:00:DE:AD:00:02",
			SourceMacAddressMask: "ff:ff:ff:ff:00:00",
		}

		rule := &acl.AccessLists_Acl_Rule{
			Match: &acl.AccessLists_Acl_Rule_Match{
				MacipRule: macipRule,
			},
			AclAction: acl.AclAction_PERMIT,
		}

		aclRes := acl.AccessLists_Acl{
			AclName: "example",
			Rules:   []*acl.AccessLists_Acl_Rule{rule},
		}

		plugin.Deps.Log.Debug(aclRes)
		formatter.JSON(w, http.StatusOK, aclRes)
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
