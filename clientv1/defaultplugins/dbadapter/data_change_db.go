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

package dbadapter

import (
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"net"
)

// NewDataChangeDSL is a constructor
func NewDataChangeDSL(txn keyval.ProtoTxn) *DataChangeDSL {
	return &DataChangeDSL{txn: txn}
}

// DataChangeDSL is used to conveniently assign all the data that are needed for the DataChange.
// This is implementation of Domain Specific Language (DSL) for change of the VPP configuration.
type DataChangeDSL struct {
	txn keyval.ProtoTxn
}

// PutDSL is here to put here most recent and previous value with revisions
type PutDSL struct {
	parent *DataChangeDSL
}

// DeleteDSL is here to put here most recent and previous value with revisions
type DeleteDSL struct {
	parent *DataChangeDSL
}

// Interface create or update the network interface
func (dsl *PutDSL) Interface(val *interfaces.Interfaces_Interface) defaultplugins.PutDSL {
	dsl.parent.txn.Put(intf.InterfaceKey(val.Name), val)
	return dsl
}

// BfdSession create or update the bidirectional forwarding detection session
func (dsl *PutDSL) BfdSession(val *bfd.SingleHopBFD_Session) defaultplugins.PutDSL {
	dsl.parent.txn.Put(bfd.SessionKey(val.Interface), val)
	return dsl
}

// BfdAuthKeys create or update the bidirectional forwarding detection key
func (dsl *PutDSL) BfdAuthKeys(val *bfd.SingleHopBFD_Key) defaultplugins.PutDSL {
	dsl.parent.txn.Put(bfd.AuthKeysKey(string(val.Id)), val)
	return dsl
}

// BfdEchoFunction create or update the bidirectional forwarding detection echod function
func (dsl *PutDSL) BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) defaultplugins.PutDSL {
	dsl.parent.txn.Put(bfd.EchoFunctionKey(val.EchoSourceInterface), val)
	return dsl
}

// BD create or update the Bridge Domain
func (dsl *PutDSL) BD(val *l2.BridgeDomains_BridgeDomain) defaultplugins.PutDSL {
	dsl.parent.txn.Put(l2.BridgeDomainKey(val.Name), val)

	return dsl
}

// BDFIB delete request for the L2 Forwarding Information Base
func (dsl *PutDSL) BDFIB(val *l2.FibTableEntries_FibTableEntry) defaultplugins.PutDSL {
	dsl.parent.txn.Put(l2.FibKey(val.BridgeDomain, val.PhysAddress), val)

	return dsl
}

// XConnect create or update the Cross Connect
func (dsl *PutDSL) XConnect(val *l2.XConnectPairs_XConnectPair) defaultplugins.PutDSL {
	dsl.parent.txn.Put(l2.XConnectKey(val.ReceiveInterface), val)

	return dsl
}

// StaticRoute create or update the L3 Static Route
func (dsl *PutDSL) StaticRoute(val *l3.StaticRoutes_Route) defaultplugins.PutDSL {
	_, dstAddr, _ := net.ParseCIDR(val.DstIpAddr)
	dsl.parent.txn.Put(l3.RouteKey(val.VrfId, dstAddr, val.NextHopAddr), val)

	return dsl
}

// ACL create or update request for the Access Control List
func (dsl *PutDSL) ACL(val *acl.AccessLists_Acl) defaultplugins.PutDSL {
	dsl.parent.txn.Put(acl.Key(val.AclName), val)

	return dsl
}

// Put gives you the ability to create Interface/BD...
func (dsl *DataChangeDSL) Put() defaultplugins.PutDSL {
	return &PutDSL{dsl}
}

// Delete gives you the ability to delete Interface/BD...
func (dsl *DataChangeDSL) Delete() defaultplugins.DeleteDSL {
	return &DeleteDSL{dsl}
}

// Delete gives you the ability to delete Interface/BD...
func (dsl *PutDSL) Delete() defaultplugins.DeleteDSL {
	return &DeleteDSL{dsl.parent}
}

// Send will propagate changes to the channels
func (dsl *PutDSL) Send() defaultplugins.Reply {
	return dsl.parent.Send()
}

// Interface create or update the network interface
func (dsl *DeleteDSL) Interface(interfaceName string) defaultplugins.DeleteDSL {
	dsl.parent.txn.Delete(intf.InterfaceKey(interfaceName))

	return dsl
}

// BD create or update the Bridge Domain
func (dsl *DeleteDSL) BD(bdName string) defaultplugins.DeleteDSL {
	dsl.parent.txn.Delete(l2.BridgeDomainKey(bdName))

	return dsl
}

// BDFIB delete request for the L2 Forwarding Information Base
func (dsl *DeleteDSL) BDFIB(bdName string, mac string) defaultplugins.DeleteDSL {
	dsl.parent.txn.Delete(l2.FibKey(bdName, mac))

	return dsl
}

// XConnect create or update the Cross Connect
func (dsl *DeleteDSL) XConnect(rxIfName string) defaultplugins.DeleteDSL {
	dsl.parent.txn.Delete(l2.XConnectKey(rxIfName))

	return dsl
}

// StaticRoute create or update the L3 Static Route
func (dsl *DeleteDSL) StaticRoute(vrf uint32, dstAddrInput *net.IPNet, nextHopAddr net.IP) defaultplugins.DeleteDSL {
	//_, dstAddr, _ := net.ParseCIDR(dstAddrInput)
	dsl.parent.txn.Delete(l3.RouteKey(vrf, dstAddrInput, nextHopAddr.String()))

	return dsl
}

// ACL delete request for Access Control List
func (dsl *DeleteDSL) ACL(aclName string) defaultplugins.DeleteDSL {
	dsl.parent.txn.Delete(acl.Key(aclName))

	return dsl
}

// Put gives you the ability to create Interface/BD...
func (dsl *DeleteDSL) Put() defaultplugins.PutDSL {
	return &PutDSL{dsl.parent}
}

// Send will propagate changes to the channels
func (dsl *DeleteDSL) Send() defaultplugins.Reply {
	return dsl.parent.Send()
}

// Send will propagate changes to the channels
func (dsl *DataChangeDSL) Send() defaultplugins.Reply {
	err := dsl.txn.Commit()

	return &Reply{err}
}

// Reply is here to gives you the ability to wait for the reply and get result (success/error)
type Reply struct {
	err error
}

// ReceiveReply TODO
func (dsl Reply) ReceiveReply() error {
	return dsl.err
}
