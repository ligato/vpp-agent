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
	"github.com/ligato/vpp-agent/clientv1/linux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/model/interfaces"

	vpp_clientv1 "github.com/ligato/vpp-agent/clientv1/defaultplugins"
	vpp_dbadapter "github.com/ligato/vpp-agent/clientv1/defaultplugins/dbadapter"
	vpp_acl "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	vpp_intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	vpp_l2 "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	vpp_l3 "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"

	"github.com/ligato/cn-infra/db/keyval"
	"net"
)

// NewDataChangeDSL is a constructor
func NewDataChangeDSL(txn keyval.ProtoTxn) *DataChangeDSL {
	vppDbAdapter := vpp_dbadapter.NewDataChangeDSL(txn)
	return &DataChangeDSL{txn: txn, vppDataChange: vppDbAdapter}
}

// DataChangeDSL is used to conveniently assign all the data that are needed for the DataChange.
// This is implementation of Domain Specific Language (DSL) for change of both Linux and VPP configuration.
type DataChangeDSL struct {
	txn           keyval.ProtoTxn
	vppDataChange vpp_clientv1.DataChangeDSL
}

// PutDSL is here to put here most recent and previous value with revisions
type PutDSL struct {
	parent *DataChangeDSL
	vppPut vpp_clientv1.PutDSL
}

// DeleteDSL is here to put here most recent and previous value with revisions
type DeleteDSL struct {
	parent    *DataChangeDSL
	vppDelete vpp_clientv1.DeleteDSL
}

// Put gives you the ability to create configurable object
func (dsl *DataChangeDSL) Put() linux.PutDSL {
	return &PutDSL{dsl, dsl.vppDataChange.Put()}
}

// Delete gives you the ability to delete configurable object
func (dsl *DataChangeDSL) Delete() linux.DeleteDSL {
	return &DeleteDSL{dsl, dsl.vppDataChange.Delete()}
}

// Send will propagate changes to the channels
func (dsl *DataChangeDSL) Send() vpp_clientv1.Reply {
	return dsl.vppDataChange.Send()
}

// LinuxInterface create or update a Linux network interface
func (dsl *PutDSL) LinuxInterface(val *interfaces.LinuxInterfaces_Interface) linux.PutDSL {
	dsl.parent.txn.Put(interfaces.InterfaceKey(val.Name), val)
	return dsl
}

// VppInterface adds a request to create or update VPP network interface
func (dsl *PutDSL) VppInterface(val *vpp_intf.Interfaces_Interface) linux.PutDSL {
	dsl.vppPut.Interface(val)
	return dsl
}

// BD adds a request to create or update VPP Bridge Domain
func (dsl *PutDSL) BD(val *vpp_l2.BridgeDomains_BridgeDomain) linux.PutDSL {
	dsl.vppPut.BD(val)
	return dsl
}

// BDFIB adds a request to create or update VPP L2 Forwarding Information Base
func (dsl *PutDSL) BDFIB(fib *vpp_l2.FibTableEntries_FibTableEntry) linux.PutDSL {
	dsl.vppPut.BDFIB(fib)
	return dsl
}

// XConnect adds a request to create or update VPP Cross Connect
func (dsl *PutDSL) XConnect(val *vpp_l2.XConnectPairs_XConnectPair) linux.PutDSL {
	dsl.vppPut.XConnect(val)
	return dsl
}

// StaticRoute adds a request to create or update VPP L3 Static Route
func (dsl *PutDSL) StaticRoute(val *vpp_l3.StaticRoutes_Route) linux.PutDSL {
	dsl.vppPut.StaticRoute(val)
	return dsl
}

// ACL adds a request to create or update VPP Access Control List
func (dsl *PutDSL) ACL(acl *vpp_acl.AccessLists_Acl) linux.PutDSL {
	dsl.vppPut.ACL(acl)
	return dsl
}

// Delete gives you the ability to delete configurable object
func (dsl *PutDSL) Delete() linux.DeleteDSL {
	return &DeleteDSL{dsl.parent, dsl.vppPut.Delete()}
}

// Send will propagate changes to the channels
func (dsl *PutDSL) Send() vpp_clientv1.Reply {
	return dsl.parent.Send()
}

// LinuxInterface create or update the network interface
func (dsl *DeleteDSL) LinuxInterface(interfaceName string) linux.DeleteDSL {
	dsl.parent.txn.Delete(interfaces.InterfaceKey(interfaceName))
	return dsl
}

// VppInterface adds a request to delete an existing VPP network interface
func (dsl *DeleteDSL) VppInterface(ifaceName string) linux.DeleteDSL {
	dsl.vppDelete.Interface(ifaceName)
	return dsl
}

// BD adds a request to delete an existing VPP Bridge Domain
func (dsl *DeleteDSL) BD(bdName string) linux.DeleteDSL {
	dsl.vppDelete.BD(bdName)
	return dsl
}

// BDFIB adds a request to delete an existing VPP L2 Forwarding Information Base
func (dsl *DeleteDSL) BDFIB(bdName string, mac string) linux.DeleteDSL {
	dsl.vppDelete.BDFIB(bdName, mac)
	return dsl
}

// XConnect adds a request to delete an existing VPP Cross Connect
func (dsl *DeleteDSL) XConnect(rxIfaceName string) linux.DeleteDSL {
	dsl.vppDelete.XConnect(rxIfaceName)
	return dsl
}

// StaticRoute adds a request to delete an existing VPP L3 Static Route
func (dsl *DeleteDSL) StaticRoute(vrf uint32, dstAddrInput *net.IPNet, nextHopAddrInput net.IP) linux.DeleteDSL {
	dsl.vppDelete.StaticRoute(vrf, dstAddrInput, nextHopAddrInput)
	return dsl
}

// ACL adds a request to delete an existing VPP Access Control List
func (dsl *DeleteDSL) ACL(aclName string) linux.DeleteDSL {
	dsl.vppDelete.ACL(aclName)
	return dsl
}

// Put gives you the ability to create configurable object
func (dsl *DeleteDSL) Put() linux.PutDSL {
	return &PutDSL{dsl.parent, dsl.vppDelete.Put()}
}

// Send will propagate changes to the channels
func (dsl *DeleteDSL) Send() vpp_clientv1.Reply {
	return dsl.parent.Send()
}
