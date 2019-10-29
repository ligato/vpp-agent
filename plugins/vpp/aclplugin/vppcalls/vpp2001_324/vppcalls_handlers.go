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

package vpp2001_324

import (
	govppapi "git.fd.io/govpp.git/api"

	"go.ligato.io/vpp-agent/v2/plugins/vpp/aclplugin/vppcalls"
	vpp_acl "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/acl"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/ifaceidx"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_acl.AllMessages()...)

	vppcalls.Versions["vpp2001_324"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex) vppcalls.ACLVppAPI {
			return NewACLVppHandler(ch, ifIdx)
		},
	}
}

// ACLVppHandler is accessor for acl-related vppcalls methods
type ACLVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
}

func NewACLVppHandler(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex) *ACLVppHandler {
	return &ACLVppHandler{
		callsChannel: ch,
		ifIndexes:    ifIdx,
	}
}
