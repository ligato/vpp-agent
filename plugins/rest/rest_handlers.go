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
	ifplugin "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	l2plugin "github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppdump"
	l3plugin "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppdump"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
)

// interfacesGetHandler - used to get list of all interfaces
func (plugin *Plugin) interfacesGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting list of all interfaces")

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		ifHandler, err := ifplugin.NewIfVppHandler(ch, plugin.Log, nil)
		if err != nil {
			plugin.Log.Errorf("Error creating VPP handler: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		res, err := ifHandler.DumpInterfaces()
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		plugin.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

// bridgeDomainIdsGetHandler - used to get list of all bridge domain ids
func (plugin *Plugin) bridgeDomainIdsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting list of all bridge domain ids")

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		res, err := l2plugin.DumpBridgeDomainIDs(ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		plugin.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

// bridgeDomainsGetHandler - used to get list of all bridge domains
func (plugin *Plugin) bridgeDomainsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting list of all bridge domains")

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		res, err := l2plugin.DumpBridgeDomains(ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
			return
		}

		plugin.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

// fibTableEntriesGetHandler - used to get list of all fib entries
func (plugin *Plugin) fibTableEntriesGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting list of all fibs")

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		res, err := l2plugin.DumpFIBTableEntries(ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
			return
		}

		plugin.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

// xconnectPairsGetHandler - used to get list of all connect pairs (transmit and receive interfaces)
func (plugin *Plugin) xconnectPairsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Log.Debug("Getting list of all xconnect pairs")

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		res, err := l2plugin.DumpXConnectPairs(ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, nil)
		}

		plugin.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
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

		res, err := l3plugin.DumpArps(plugin.Log, ch, nil)
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

		res, err := l3plugin.DumpStaticRoutes(plugin.Log, ch, nil)
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

		// create an API channel
		ch, err := plugin.GoVppmux.NewAPIChannel()
		if err != nil {
			plugin.Log.Errorf("Error creating channel: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		defer ch.Close()

		swIndex := uint32(swIndexuInt64)
		aclHandler, err := aclcalls.NewAclVppHandler(ch, ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error creating VPP handler: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		res, err := aclHandler.DumpInterfaceIPAcls(swIndex)
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}
		res, err = aclHandler.DumpInterfaceMACIPAcls(swIndex)
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
