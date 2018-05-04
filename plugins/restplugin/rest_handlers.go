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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"git.fd.io/govpp.git/core/bin_api/vpe"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"

	aclvpp "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppcalls"
	acldump "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppdump"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	ifplugin "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
	l2plugin "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppdump"
	l3plugin "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppdump"
)

// interfacesGetHandler - used to get list of all interfaces
func (plugin *RESTAPIPlugin) interfacesGetHandler(formatter *render.Render) http.HandlerFunc {
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

		res, err := ifplugin.DumpInterfaces(plugin.Log, ch, nil)
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
func (plugin *RESTAPIPlugin) bridgeDomainIdsGetHandler(formatter *render.Render) http.HandlerFunc {
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
func (plugin *RESTAPIPlugin) bridgeDomainsGetHandler(formatter *render.Render) http.HandlerFunc {
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
func (plugin *RESTAPIPlugin) fibTableEntriesGetHandler(formatter *render.Render) http.HandlerFunc {
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
func (plugin *RESTAPIPlugin) xconnectPairsGetHandler(formatter *render.Render) http.HandlerFunc {
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
func (plugin *RESTAPIPlugin) staticRoutesGetHandler(formatter *render.Render) http.HandlerFunc {
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
func (plugin *RESTAPIPlugin) interfaceACLGetHandler(formatter *render.Render) http.HandlerFunc {
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
		res, _, err := acldump.DumpInterfaceAcls(plugin.Log, swIndex, ch, nil)
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
func (plugin *RESTAPIPlugin) ipACLGetHandler(formatter *render.Render) http.HandlerFunc {
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

		res, err := acldump.DumpACLs(plugin.Log, nil, ch, nil)
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
func (plugin *RESTAPIPlugin) exampleACLGetHandler(formatter *render.Render) http.HandlerFunc {
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

// ipACLPostHandler - used to get acl configuration for a particular interface
func (plugin *RESTAPIPlugin) ipACLPostHandler(formatter *render.Render) http.HandlerFunc {
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
		aclIndex.Idx, err = aclvpp.AddIPAcl(aclParam.Rules, aclParam.AclName, plugin.Log, ch, nil)
		if err != nil {
			plugin.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, aclIndex)
			return
		}

		plugin.Log.Debug(aclIndex)
		formatter.JSON(w, http.StatusOK, aclIndex)
	}
}

// showCommandHandler - used to execute VPP CLI commands
func (plugin *RESTAPIPlugin) showCommandHandler(formatter *render.Render) http.HandlerFunc {
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
			plugin.Log.Errorf("Error processing request: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		} else if reply.Retval > 0 {
			plugin.Log.Errorf("Command returned code: %v", reply.Retval)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		plugin.Log.Debugf("VPPCLI response: %s", reply.Reply)
		formatter.Text(w, http.StatusOK, string(reply.Reply))
	}
}
