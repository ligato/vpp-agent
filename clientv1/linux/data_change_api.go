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

package linux

import (
	"github.com/ligato/vpp-agent/plugins/linuxplugin/model/interfaces"

	vpp_clientv1 "github.com/ligato/vpp-agent/clientv1/defaultplugins"
	vpp_acl "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	vpp_intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	vpp_l2 "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	vpp_l3 "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"net"
)

// DataChangeDSL is used to conveniently assign all the data that are needed for the DataChange
// Use this interface to make you implementation independent of local or particular remote client.
// This interface defines the Domain Specific Language (DSL) for change of the Linux configuration.
type DataChangeDSL interface {
	// Put gives you the ability to create configurable object
	Put() PutDSL

	// Delete gives you the ability to delete configurable objects
	Delete() DeleteDSL

	// Send will propagate changes to the channels
	Send() vpp_clientv1.Reply
}

// PutDSL is here to put here most recent and previous value with revisions
type PutDSL interface {
	// Interface adds a request to create or update Linux network interface
	LinuxInterface(val *interfaces.LinuxInterfaces_Interface) PutDSL

	// VppInterface adds a request to create or update VPP network interface
	VppInterface(val *vpp_intf.Interfaces_Interface) PutDSL
	// BD adds a request to create or update VPP Bridge Domain
	BD(val *vpp_l2.BridgeDomains_BridgeDomain) PutDSL
	// FIB adds a request to create or update VPP L2 Forwarding Information Base
	BDFIB(fib *vpp_l2.FibTableEntries_FibTableEntry) PutDSL
	// XConnect adds a request to create or update VPP Cross Connect
	XConnect(val *vpp_l2.XConnectPairs_XConnectPair) PutDSL
	// StaticRoute adds a request to create or update VPP L3 Static Route
	StaticRoute(val *vpp_l3.StaticRoutes_Route) PutDSL
	// ACL adds a request to create or update VPP Access Control List
	ACL(acl *vpp_acl.AccessLists_Acl) PutDSL

	// Delete gives you the ability to delete configurable objects
	Delete() DeleteDSL
	// Send will propagate changes to the channels
	Send() vpp_clientv1.Reply
}

// DeleteDSL is here to put here most recent and previous value with revisions
type DeleteDSL interface {
	// Interface adds a request to delete an existing Linux network interface
	LinuxInterface(ifaceName string) DeleteDSL

	// VppInterface adds a request to delete an existing VPP network interface
	VppInterface(ifaceName string) DeleteDSL
	// BD adds a request to delete an existing VPP Bridge Domain
	BD(bdName string) DeleteDSL
	// FIB adds a request to delete an existing VPP L2 Forwarding Information Base
	BDFIB(bdName string, mac string) DeleteDSL
	// XConnect adds a request to delete an existing VPP Cross Connect
	XConnect(rxIfaceName string) DeleteDSL
	// StaticRoute adds a request to delete an existing VPP L3 Static Route
	StaticRoute(vrf uint32, dstAddr *net.IPNet, nextHopAddr net.IP) DeleteDSL
	// ACL adds a request to delete an existing VPP Access Control List
	ACL(aclName string) DeleteDSL

	// Put gives you the ability to create configurable objects
	Put() PutDSL
	// Send will propagate changes to the channels
	Send() vpp_clientv1.Reply
}
