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
	aclvpp "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppcalls"
	acldump "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppdump"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	ifplugin "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
	l2plugin "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppdump"
	l3plugin "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppdump"
	"github.com/unrolled/render"
)

//interfacesGetHandler - used to get list of all interfaces
func (plugin *RESTAPIPlugin) interfacesGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all interfaces")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()

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
	}
}

//bridgeDomainIdsGetHandler - used to get list of all bridge domain ids
func (plugin *RESTAPIPlugin) bridgeDomainIdsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all bridge domain ids")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()

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
	}
}

//bridgeDomainsGetHandler - used to get list of all bridge domains
func (plugin *RESTAPIPlugin) bridgeDomainsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all bridge domains")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()

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
	}
}

//fibTableEntriesGetHandler - used to get list of all fib entries
func (plugin *RESTAPIPlugin) fibTableEntriesGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all fibs")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()

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
	}
}

//xconnectPairsGetHandler - used to get list of all connect pairs (transmit and receive interfaces)
func (plugin *RESTAPIPlugin) xconnectPairsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all xconnect pairs")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()

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
	}
}

//staticRoutesGetHandler - used to get list of all static routes
func (plugin *RESTAPIPlugin) staticRoutesGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting list of all static routes")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()

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
	}
}

//interfaceACLGetHandler - used to get acl configuration for a particular interface
func (plugin *RESTAPIPlugin) interfaceACLGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting acl configuration of interface")

		vars := mux.Vars(req)
		if vars == nil {
			plugin.Deps.Log.Error("Interface software index not specified.")
			formatter.JSON(w, http.StatusNotFound, struct{}{})
			return
		}

		plugin.Deps.Log.Infof("Received request for swIndex :: %v ", vars[swIndexVarName])

		swIndexuInt64, err := strconv.ParseUint(vars[swIndexVarName], 10, 32)
		if err != nil {
			plugin.Deps.Log.Error("Failed to unmarshal request body.")
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		swIndex := uint32(swIndexuInt64)
		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		res, _, err := acldump.DumpInterfaceAcls(plugin.Deps.Log, swIndex, ch, nil)
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		plugin.Deps.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

//ipACLGetHandler - used to get configuration of IP ACLs
func (plugin *RESTAPIPlugin) ipACLGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting acls")

		// create an API channel
		ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
		defer ch.Close()
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		res, err := acldump.DumpACLs(plugin.Deps.Log, nil, ch, nil)
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		plugin.Deps.Log.Debug(res)
		formatter.JSON(w, http.StatusOK, res)
	}
}

// exampleACLGetHandler - used to get an example ACL configuration
func (plugin *RESTAPIPlugin) exampleACLGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		plugin.Deps.Log.Info("Getting example acl")

		ipRule := acl.AccessLists_Acl_Rule_Matches_IpRule{
			Ip: &acl.AccessLists_Acl_Rule_Matches_IpRule_Ip{
				DestinationNetwork: "1.2.3.4/24",
				SourceNetwork:      "5.6.7.8/24"},
			Tcp: &acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp{
				DestinationPortRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_DestinationPortRange{
					LowerPort: 80,
					UpperPort: 8080,
				},
				SourcePortRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_SourcePortRange{
					LowerPort: 10,
					UpperPort: 1010,
				},
				TcpFlagsMask:  0xFF,
				TcpFlagsValue: 9,
			},
		}
		matches := acl.AccessLists_Acl_Rule_Matches{
			IpRule: &ipRule,
		}

		actions := acl.AccessLists_Acl_Rule_Actions{
			AclAction: acl.AclAction_PERMIT,
		}

		rule := acl.AccessLists_Acl_Rule{
			Matches: &matches,
			Actions: &actions,
		}
		rules := []*acl.AccessLists_Acl_Rule{&rule}

		aclRes := acl.AccessLists_Acl{
			AclName: "example",
			Rules:   rules,
		}

		plugin.Deps.Log.Debug(aclRes)
		formatter.JSON(w, http.StatusOK, aclRes)
	}
}

//ipACLPostHandler - used to get acl configuration for a particular interface
func (plugin *RESTAPIPlugin) ipACLPostHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		type ACLIndex struct {
			Idx uint32 `json:"acl_index"`
		}

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

		aclIndex := ACLIndex{}
		aclIndex.Idx, err = aclvpp.AddIPAcl(aclParam.Rules, aclParam.AclName, plugin.Deps.Log, ch, nil)
		if err != nil {
			plugin.Deps.Log.Errorf("Error: %v", err)
			formatter.JSON(w, http.StatusInternalServerError, aclIndex)
			return
		}

		plugin.Deps.Log.Debug(aclIndex)
		formatter.JSON(w, http.StatusOK, aclIndex)
	}
}

//showCommandHandler - used to execute VPP CLI commands
func (plugin *RESTAPIPlugin) showCommandHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var reqParam map[string]string
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			plugin.Deps.Log.Error("Failed to parse request body.")
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, &reqParam)
		if err != nil {
			plugin.Deps.Log.Error("Failed to unmarshal request body.")
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		command, ok := reqParam["vppclicommand"]

		if !ok {
			plugin.Deps.Log.Error("command parameter not included.")
			formatter.JSON(w, http.StatusInternalServerError, err)
			return
		}

		if command != "" {

			plugin.Deps.Log.WithField("VPPCLI command", command).Info("VPPCLI command")

			ch, err := plugin.Deps.GoVppmux.NewAPIChannel()
			defer ch.Close()

			if err != nil {
				plugin.Deps.Log.Errorf("Error creating channel: %v", err)
				formatter.JSON(w, http.StatusInternalServerError, err)
			} else {
				req := &vpe.CliInband{}
				req.Length = uint32(len(command))
				req.Cmd = []byte(command)

				reply := &vpe.CliInbandReply{}
				err = ch.SendRequest(req).ReceiveReply(reply)
				if err != nil {
					plugin.Deps.Log.Errorf("Error processing request: %v", err)
					formatter.JSON(w, http.StatusInternalServerError, err)
				}

				if reply.Retval > 0 {
					plugin.Deps.Log.Errorf("Command returned code: %v", reply.Retval)
				}

				plugin.Deps.Log.WithField("VPPCLI response", string(reply.Reply)).Info("VPPCLI response")

				formatter.Text(w, http.StatusOK, string(reply.Reply))
			}
		} else {
			formatter.JSON(w, http.StatusBadRequest, "showCommand parameter is empty")
		}
	}
}
