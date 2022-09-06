//  Copyright (c) 2022 Cisco and/or its affiliates.
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

package vpp2202

import (
	"fmt"
	"net"

	govppapi "go.fd.io/govpp/api"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls"
	vpp2202 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/acl"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
)

func init() {
	msgs := acl.AllMessages()
	vppcalls.AddHandlerVersion(vpp2202.Version, msgs, NewACLVppHandler)
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
		acl:          acl.NewServiceClient(c),
		ifIndexes:    ifIdx,
	}
}

func prefixToString(address ip_types.Prefix) string {
	if address.Address.Af == ip_types.ADDRESS_IP6 {
		ip6 := address.Address.Un.GetIP6()
		return fmt.Sprintf("%s/%d", net.IP(ip6[:]).To16(), address.Len)
	} else {
		ip4 := address.Address.Un.GetIP4()
		return fmt.Sprintf("%s/%d", net.IP(ip4[:]).To4(), address.Len)
	}
}

func addressToIP(address ip_types.Address) net.IP {
	if address.Af == ip_types.ADDRESS_IP6 {
		ipAddr := address.Un.GetIP6()
		return net.IP(ipAddr[:]).To16()
	}
	ipAddr := address.Un.GetIP4()
	return net.IP(ipAddr[:]).To4()
}
