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

	"net"

	vpp_clientv1 "github.com/ligato/vpp-agent/clientv1/defaultplugins"
	vpp_acl "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	vpp_intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	vpp_l2 "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	vpp_l3 "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	vpp_bfd "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
)

// DataChangeDSL defines the Domain Specific Language (DSL) for data change
// of both Linux and VPP configuration.
// Use this interface to make your implementation independent of the local
// and any remote client.
// Every DSL statement (apart from Send) returns the receiver (possibly wrapped
// to change the scope of DSL), allowing the calls to be chained together
// conveniently in a single statement.
type DataChangeDSL interface {
	// Put initiates a chained sequence of data change DSL statements declaring
	// new or changing existing configurable objects, e.g:
	//     Put().LinuxInterface(&veth).VppInterface(&afpacket).BD(&BD) ... Send()
	// The set of available objects to create or change is defined by PutDSL
	// interface which the returned object implements.
	Put() PutDSL

	// Delete initiates a chained sequence of data change DSL statements
	// removing existing configurable objects (by name), e.g:
	//     Delete().LinuxInterface(vethName).VppInterface(afpacketName).BD(BDName) ... Send()
	// The set of available objects to remove is defined by DeleteDSL
	// interface which the returned object implements.
	Delete() DeleteDSL

	// Send propagates requested changes to the plugins.
	Send() vpp_clientv1.Reply
}

// PutDSL is a subset of data change DSL statements used to declare new
// or change existing Linux or VPP configuration.
type PutDSL interface {
	// LinuxInterface adds a request to create or update Linux network interface.
	LinuxInterface(val *interfaces.LinuxInterfaces_Interface) PutDSL

	// VppInterface adds a request to create or update VPP network interface.
	VppInterface(val *vpp_intf.Interfaces_Interface) PutDSL
	// BfdSession adds a request to create or update VPP bidirectional
	// forwarding detection session.
	BfdSession(val *vpp_bfd.SingleHopBFD_Session) PutDSL
	// BfdAuthKeys adds a request to create or update VPP bidirectional
	// forwarding detection key.
	BfdAuthKeys(val *vpp_bfd.SingleHopBFD_Key) PutDSL
	// BfdEchoFunction adds a request to create or update VPP bidirectional
	// forwarding detection echo function.
	BfdEchoFunction(val *vpp_bfd.SingleHopBFD_EchoFunction) PutDSL
	// BD adds a request to create or update VPP Bridge Domain.
	BD(val *vpp_l2.BridgeDomains_BridgeDomain) PutDSL
	// BDFIB adds a request to create or update VPP L2 Forwarding Information Base.
	BDFIB(fib *vpp_l2.FibTableEntries_FibTableEntry) PutDSL
	// XConnect adds a request to create or update VPP Cross Connect.
	XConnect(val *vpp_l2.XConnectPairs_XConnectPair) PutDSL
	// StaticRoute adds a request to create or update VPP L3 Static Route.
	StaticRoute(val *vpp_l3.StaticRoutes_Route) PutDSL
	// ACL adds a request to create or update VPP Access Control List.
	ACL(acl *vpp_acl.AccessLists_Acl) PutDSL

	// Delete changes the DSL mode to allow removal of an existing configuration.
	// See documentation for DataChangeDSL.Delete().
	Delete() DeleteDSL

	// Send propagates requested changes to the plugins.
	Send() vpp_clientv1.Reply
}

// DeleteDSL is a subset of data change DSL statements used to remove
// existing Linux or VPP configuration.
type DeleteDSL interface {
	// LinuxInterface adds a request to delete an existing Linux network
	// interface.
	LinuxInterface(ifaceName string) DeleteDSL

	// VppInterface adds a request to delete an existing VPP network interface.
	VppInterface(ifaceName string) DeleteDSL
	// BfdSession adds a request to delete an existing VPP bidirectional
	// forwarding/ detection session.
	BfdSession(bfdSessionIfaceName string) DeleteDSL
	// BfdAuthKeys adds a request to delete an existing VPP bidirectional
	// forwarding detection key.
	BfdAuthKeys(bfdKeyName string) DeleteDSL
	// BfdEchoFunction adds a request to delete an existing VPP bidirectional
	// forwarding detection echo function.
	BfdEchoFunction(bfdEchoName string) DeleteDSL
	// BD adds a request to delete an existing VPP Bridge Domain.
	BD(bdName string) DeleteDSL
	// FIB adds a request to delete an existing VPP L2 Forwarding Information
	// Base.
	BDFIB(bdName string, mac string) DeleteDSL
	// XConnect adds a request to delete an existing VPP Cross Connect.
	XConnect(rxIfaceName string) DeleteDSL
	// StaticRoute adds a request to delete an existing VPP L3 Static Route.
	StaticRoute(vrf uint32, dstAddr *net.IPNet, nextHopAddr net.IP) DeleteDSL
	// ACL adds a request to delete an existing VPP Access Control List.
	ACL(aclName string) DeleteDSL

	// Put changes the DSL mode to allow configuration editing.
	// See documentation for DataChangeDSL.Put().
	Put() PutDSL

	// Send propagates requested changes to the plugins.
	Send() vpp_clientv1.Reply
}
