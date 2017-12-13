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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/stn"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/model/l4"
)

// DataChangeDSL defines Domain Specific Language (DSL) for data change.
// of the VPP configuration.
// Use this interface to make your implementation independent of the local
// and any remote client.
// Every DSL statement (apart from Send) returns the receiver (possibly wrapped
// to change the scope of DSL), allowing the calls to be chained together
// conveniently in a single statement.
type DataChangeDSL interface {
	// Put initiates a chained sequence of data change DSL statements, declaring
	// new or changing existing configurable objects, e.g.:
	//     Put().Interface(&memif).XConnect(&xconnect).BD(&BD) ... Send()
	// The set of available objects to be created or changed is defined by PutDSL.
	Put() PutDSL

	// Delete initiates a chained sequence of data change DSL statements,
	// removing existing configurable objects (by name), e.g.:
	//     Delete().Interface(memifName).XConnect(xconnectName).BD(BDName) ... Send()
	// The set of available objects to be removed is defined by DeleteDSL.
	Delete() DeleteDSL

	// Send propagates requested changes to the plugins.
	Send() Reply
}

// PutDSL is a subset of data change DSL statements, used to declare new
// VPP configuration or to change an existing one.
type PutDSL interface {
	// Interface adds a request to create or update VPP network interface.
	Interface(val *interfaces.Interfaces_Interface) PutDSL
	// BfdSession adds a request to create or update bidirectional forwarding
	// detection session.
	BfdSession(val *bfd.SingleHopBFD_Session) PutDSL
	// BfdAuthKeys adds a request to create or update bidirectional forwarding
	// detection key.
	BfdAuthKeys(val *bfd.SingleHopBFD_Key) PutDSL
	// BfdEchoFunction adds a request to create or update bidirectional
	// forwarding detection echo function.
	BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) PutDSL
	// BD adds a request to create or update VPP Bridge Domain.
	BD(val *l2.BridgeDomains_BridgeDomain) PutDSL
	// BDFIB adds a request to create or update VPP L2 Forwarding Information Base.
	BDFIB(fib *l2.FibTableEntries_FibTableEntry) PutDSL
	// XConnect adds a request to create or update VPP Cross Connect.
	XConnect(val *l2.XConnectPairs_XConnectPair) PutDSL
	// StaticRoute adds a request to create or update VPP L3 Static Route.
	StaticRoute(val *l3.StaticRoutes_Route) PutDSL
	// ACL adds a request to create or update VPP Access Control List.
	ACL(acl *acl.AccessLists_Acl) PutDSL
	// Arp adds a request to create or update VPP L3 ARP.
	Arp(arp *l3.ArpTable_ArpTableEntry) PutDSL
	// L4Features adds a request to enable or disable L4 features
	L4Features(val *l4.L4Features) PutDSL
	// AppNamespace adds a request to create or update VPP Application namespace
	AppNamespace(appNs *l4.AppNamespaces_AppNamespace) PutDSL
	// StnRule adds a request to create or update Stn rule.
	StnRule(stn *stn.StnRule) PutDSL

	// Delete changes the DSL mode to allow removal of an existing configuration.
	// See documentation for DataChangeDSL.Delete().
	Delete() DeleteDSL

	// Send propagates requested changes to the plugins.
	Send() Reply
}

// DeleteDSL is a subset of data change DSL statements, used to remove
// an existing VPP configuration.
type DeleteDSL interface {
	// Interface adds a request to delete an existing VPP network interface.
	Interface(ifaceName string) DeleteDSL
	// BfdSession adds a request to delete an existing bidirectional forwarding
	// detection session.
	BfdSession(bfdSessionIfaceName string) DeleteDSL
	// BfdAuthKeys adds a request to delete an existing bidirectional forwarding
	// detection key.
	BfdAuthKeys(bfdKey uint32) DeleteDSL
	// BfdEchoFunction adds a request to delete an existing bidirectional
	// forwarding detection echo function.
	BfdEchoFunction(bfdEchoName string) DeleteDSL
	// BD adds a request to delete an existing VPP Bridge Domain.
	BD(bdName string) DeleteDSL
	// BDFIB adds a request to delete an existing VPP L2 Forwarding Information
	// Base.
	BDFIB(bdName string, mac string) DeleteDSL
	// XConnect adds a request to delete an existing VPP Cross Connect.
	XConnect(rxIfaceName string) DeleteDSL
	// StaticRoute adds a request to delete an existing VPP L3 Static Route.
	StaticRoute(vrf uint32, dstAddr string, nextHopAddr string) DeleteDSL
	// ACL adds a request to delete an existing VPP Access Control List.
	ACL(aclName string) DeleteDSL
	// L4Features adds a request to enable or disable L4 features
	L4Features() DeleteDSL
	// AppNamespace adds a request to delete VPP Application namespace
	// Note: current version does not support application namespace deletion
	AppNamespace(id string) DeleteDSL
	// Arp adds a request to delete an existing VPP L3 ARP.
	Arp(ifaceName string, ipAddr string) DeleteDSL
	// StnRule adds a request to delete an existing Stn rule.
	StnRule(ruleName string) DeleteDSL

	// Put changes the DSL mode to allow configuration editing.
	// See documentation for DataChangeDSL.Put().
	Put() PutDSL

	// Send propagates requested changes to the plugins.
	Send() Reply
}

// Reply interface allows to wait for a reply to previously called Send() and
// extract the result from it (success/error).
type Reply interface {
	// ReceiveReply waits for a reply to previously called Send() and returns
	// the result (error or nil).
	ReceiveReply() error
}
