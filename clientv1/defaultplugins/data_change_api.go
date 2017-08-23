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

package defaultplugins

import (
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"net"
)

// DataChangeDSL is used to conveniently assign all the data that are needed for the DataChange
// Use this interface to make you implementation independent of local or particular remote client.
// This interface defines the Domain Specific Language (DSL) for change of the VPP configuration.
type DataChangeDSL interface {
	// Put gives you the ability to create Interface/BD...
	Put() PutDSL

	// Delete gives you the ability to delete Interface/BD...
	Delete() DeleteDSL

	// Send will propagate changes to the channels
	Send() Reply
}

// PutDSL is here to put here most recent and previous value with revisions
type PutDSL interface {
	// Interface create or update request for the network interface
	Interface(val *interfaces.Interfaces_Interface) PutDSL
	// BD create or update request for the Bridge Domain
	BD(val *l2.BridgeDomains_BridgeDomain) PutDSL
	// FIB create or update request for the L2 Forwarding Information Base
	BDFIB(fib *l2.FibTableEntries_FibTableEntry) PutDSL
	// XConnect create or update request for the Cross Connect
	XConnect(val *l2.XConnectPairs_XConnectPair) PutDSL
	// StaticRoute create or update request for the L3 Static Route
	StaticRoute(val *l3.StaticRoutes_Route) PutDSL
	// ACL create or update request for the Access Control List
	ACL(acl *acl.AccessLists_Acl) PutDSL

	// Delete gives you the ability to delete Interface/BD...
	Delete() DeleteDSL
	// Send will propagate changes to the channels
	Send() Reply
}

// DeleteDSL is here to put here most recent and previous value with revisions
type DeleteDSL interface {
	// Interface delete request for the network interface
	Interface(ifaceName string) DeleteDSL
	// BD delete request for the Bridge Domain
	BD(bdName string) DeleteDSL
	// FIB delete request for the L2 Forwarding Information Base
	BDFIB(bdName string, mac string) DeleteDSL
	// XConnect delete request for the Cross Connect
	XConnect(rxIfaceName string) DeleteDSL
	// StaticRoute delete request for the L3 Static Route
	StaticRoute(vrf uint32, dstAddr *net.IPNet, nextHopAddr net.IP) DeleteDSL
	// ACL delete request for Access Control List
	ACL(aclName string) DeleteDSL

	// Put gives you the ability to create Interface/BD...
	Put() PutDSL
	// Send will propagate changes to the channels
	Send() Reply
}

// Reply is here to gives you the ability to wait for the reply and get result (success/error)
type Reply interface {
	// ReceiveReply TODO
	ReceiveReply() error
}
