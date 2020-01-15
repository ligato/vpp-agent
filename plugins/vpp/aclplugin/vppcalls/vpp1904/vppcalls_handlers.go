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

package vpp1904

import (
	govppapi "git.fd.io/govpp.git/api"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/acl"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
)

func init() {
	msgs := acl.AllMessages()
	vppcalls.AddHandlerVersion(vpp1904.Version, msgs, NewACLVppHandler)
}

// ACLVppHandler is accessor for acl-related vppcalls methods
type ACLVppHandler struct {
	callsChannel govppapi.Channel
	// TODO: use only RPC service
	acl       acl.RPCService
	ifIndexes ifaceidx.IfaceMetadataIndex
}

func NewACLVppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex) vppcalls.ACLVppAPI {
	ch, err := c.NewAPIChannel()
	if err != nil {
		return nil
	}
	return &ACLVppHandler{
		callsChannel: ch,
		acl:          acl.NewServiceClient(ch),
		ifIndexes:    ifIdx,
	}
}
