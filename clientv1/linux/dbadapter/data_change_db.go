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
	"github.com/ligato/vpp-agent/clientv1/linux"

	vpp_clientv1 "github.com/ligato/vpp-agent/clientv1/defaultplugins"
	vpp_dbadapter "github.com/ligato/vpp-agent/clientv1/defaultplugins/dbadapter"
	vpp_acl "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	vpp_bfd "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	vpp_intf "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	vpp_l2 "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	vpp_l3 "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	vpp_l4 "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l4"
	vpp_stn "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/stn"

	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
)

// NewDataChangeDSL returns a new instance of DataChangeDSL which implements
// the data change DSL for both Linux and VPP config (inherits dbadapter
// from defaultplugins).
// Transaction <txn> is used to propagate changes to plugins.
func NewDataChangeDSL(txn keyval.ProtoTxn) *DataChangeDSL {
	vppDbAdapter := vpp_dbadapter.NewDataChangeDSL(txn)
	return &DataChangeDSL{txn: txn, vppDataChange: vppDbAdapter}
}

// DataChangeDSL is an implementation of Domain Specific Language (DSL)
// for changes of both Linux and VPP configuration.
type DataChangeDSL struct {
	txn           keyval.ProtoTxn
	vppDataChange vpp_clientv1.DataChangeDSL
}

// PutDSL implements put operations of data change DSL.
type PutDSL struct {
	parent *DataChangeDSL
	vppPut vpp_clientv1.PutDSL
}

// DeleteDSL implements delete operations of data change DSL.
type DeleteDSL struct {
	parent    *DataChangeDSL
	vppDelete vpp_clientv1.DeleteDSL
}

// Put initiates a chained sequence of data change DSL statements and declares
// new configurable objects or changes existing ones.
func (dsl *DataChangeDSL) Put() linux.PutDSL {
	return &PutDSL{dsl, dsl.vppDataChange.Put()}
}

// Delete initiates a chained sequence of data change DSL statements
// removing existing configurable objects.
func (dsl *DataChangeDSL) Delete() linux.DeleteDSL {
	return &DeleteDSL{dsl, dsl.vppDataChange.Delete()}
}

// Send propagates requested changes to the plugins.
func (dsl *DataChangeDSL) Send() vpp_clientv1.Reply {
	return dsl.vppDataChange.Send()
}

// LinuxInterface adds a request to create or update Linux network interface.
func (dsl *PutDSL) LinuxInterface(val *interfaces.LinuxInterfaces_Interface) linux.PutDSL {
	dsl.parent.txn.Put(interfaces.InterfaceKey(val.Name), val)
	return dsl
}

// LinuxArpEntry adds a request to create or update Linux ARP entry.
func (dsl *PutDSL) LinuxArpEntry(val *l3.LinuxStaticArpEntries_ArpEntry) linux.PutDSL {
	dsl.parent.txn.Put(l3.StaticArpKey(val.Name), val)
	return dsl
}

// LinuxRoute adds a request to create or update Linux route.
func (dsl *PutDSL) LinuxRoute(val *l3.LinuxStaticRoutes_Route) linux.PutDSL {
	dsl.parent.txn.Put(l3.StaticRouteKey(val.Name), val)
	return dsl
}

// VppInterface adds a request to create or update VPP network interface.
func (dsl *PutDSL) VppInterface(val *vpp_intf.Interfaces_Interface) linux.PutDSL {
	dsl.vppPut.Interface(val)
	return dsl
}

// BfdSession adds a request to create or update VPP bidirectional forwarding
// detection session.
func (dsl *PutDSL) BfdSession(val *vpp_bfd.SingleHopBFD_Session) linux.PutDSL {
	dsl.vppPut.BfdSession(val)
	return dsl
}

// BfdAuthKeys adds a request to create or update VPP bidirectional forwarding
// detection key.
func (dsl *PutDSL) BfdAuthKeys(val *vpp_bfd.SingleHopBFD_Key) linux.PutDSL {
	dsl.vppPut.BfdAuthKeys(val)
	return dsl
}

// BfdEchoFunction adds a request to create or update VPP bidirectional forwarding
// detection echo function.
func (dsl *PutDSL) BfdEchoFunction(val *vpp_bfd.SingleHopBFD_EchoFunction) linux.PutDSL {
	dsl.vppPut.BfdEchoFunction(val)
	return dsl
}

// BD adds a request to create or update VPP Bridge Domain.
func (dsl *PutDSL) BD(val *vpp_l2.BridgeDomains_BridgeDomain) linux.PutDSL {
	dsl.vppPut.BD(val)
	return dsl
}

// BDFIB adds a request to create or update VPP L2 Forwarding Information Base.
func (dsl *PutDSL) BDFIB(fib *vpp_l2.FibTableEntries_FibTableEntry) linux.PutDSL {
	dsl.vppPut.BDFIB(fib)
	return dsl
}

// XConnect adds a request to create or update VPP Cross Connect.
func (dsl *PutDSL) XConnect(val *vpp_l2.XConnectPairs_XConnectPair) linux.PutDSL {
	dsl.vppPut.XConnect(val)
	return dsl
}

// StaticRoute adds a request to create or update VPP L3 Static Route.
func (dsl *PutDSL) StaticRoute(val *vpp_l3.StaticRoutes_Route) linux.PutDSL {
	dsl.vppPut.StaticRoute(val)
	return dsl
}

// ACL adds a request to create or update VPP Access Control List.
func (dsl *PutDSL) ACL(acl *vpp_acl.AccessLists_Acl) linux.PutDSL {
	dsl.vppPut.ACL(acl)
	return dsl
}

// Arp adds a request to create or update VPP L3 ARP.
func (dsl *PutDSL) Arp(arp *vpp_l3.ArpTable_ArpTableEntry) linux.PutDSL {
	dsl.vppPut.Arp(arp)
	return dsl
}

// L4Features adds a request to enable or disable L4 features
func (dsl *PutDSL) L4Features(val *vpp_l4.L4Features) linux.PutDSL {
	dsl.vppPut.L4Features(val)
	return dsl
}

// AppNamespace adds a request to create or update VPP Application namespace
func (dsl *PutDSL) AppNamespace(appNs *vpp_l4.AppNamespaces_AppNamespace) linux.PutDSL {
	dsl.vppPut.AppNamespace(appNs)
	return dsl
}

// StnRule adds a request to create or update VPP Stn rule.
func (dsl *PutDSL) StnRule(stn *vpp_stn.StnRule) linux.PutDSL {
	dsl.vppPut.StnRule(stn)
	return dsl
}

// Delete changes the DSL mode to allow removal of an existing configuration.
func (dsl *PutDSL) Delete() linux.DeleteDSL {
	return &DeleteDSL{dsl.parent, dsl.vppPut.Delete()}
}

// Send propagates requested changes to the plugins.
func (dsl *PutDSL) Send() vpp_clientv1.Reply {
	return dsl.parent.Send()
}

// LinuxInterface adds a request to delete an existing Linux network
// interface.
func (dsl *DeleteDSL) LinuxInterface(interfaceName string) linux.DeleteDSL {
	dsl.parent.txn.Delete(interfaces.InterfaceKey(interfaceName))
	return dsl
}

// LinuxArpEntry adds a request to delete Linux ARP entry.
func (dsl *DeleteDSL) LinuxArpEntry(entryName string) linux.DeleteDSL {
	dsl.parent.txn.Delete(l3.StaticArpKey(entryName))
	return dsl
}

// LinuxRoute adds a request to delete Linux route.
func (dsl *DeleteDSL) LinuxRoute(routeName string) linux.DeleteDSL {
	dsl.parent.txn.Delete(l3.StaticRouteKey(routeName))
	return dsl
}

// VppInterface adds a request to delete an existing VPP network interface.
func (dsl *DeleteDSL) VppInterface(ifaceName string) linux.DeleteDSL {
	dsl.vppDelete.Interface(ifaceName)
	return dsl
}

// BfdSession adds a request to delete an existing VPP bidirectional forwarding
// detection session.
func (dsl *DeleteDSL) BfdSession(bfdSessionIfaceName string) linux.DeleteDSL {
	dsl.vppDelete.BfdSession(bfdSessionIfaceName)
	return dsl
}

// BfdAuthKeys adds a request to delete an existing VPP bidirectional forwarding
// detection key.
func (dsl *DeleteDSL) BfdAuthKeys(bfdKey uint32) linux.DeleteDSL {
	dsl.vppDelete.BfdAuthKeys(bfdKey)
	return dsl
}

// BfdEchoFunction adds a request to delete an existing VPP bidirectional
// forwarding detection echo function.
func (dsl *DeleteDSL) BfdEchoFunction(bfdEchoName string) linux.DeleteDSL {
	dsl.vppDelete.BfdEchoFunction(bfdEchoName)
	return dsl
}

// BD adds a request to delete an existing VPP Bridge Domain.
func (dsl *DeleteDSL) BD(bdName string) linux.DeleteDSL {
	dsl.vppDelete.BD(bdName)
	return dsl
}

// BDFIB adds a request to delete an existing VPP L2 Forwarding Information Base.
func (dsl *DeleteDSL) BDFIB(bdName string, mac string) linux.DeleteDSL {
	dsl.vppDelete.BDFIB(bdName, mac)
	return dsl
}

// XConnect adds a request to delete an existing VPP Cross Connect.
func (dsl *DeleteDSL) XConnect(rxIfaceName string) linux.DeleteDSL {
	dsl.vppDelete.XConnect(rxIfaceName)
	return dsl
}

// StaticRoute adds a request to delete an existing VPP L3 Static Route.
func (dsl *DeleteDSL) StaticRoute(vrf uint32, dstAddr string, nextHopAddr string) linux.DeleteDSL {
	dsl.vppDelete.StaticRoute(vrf, dstAddr, nextHopAddr)
	return dsl
}

// ACL adds a request to delete an existing VPP Access Control List.
func (dsl *DeleteDSL) ACL(aclName string) linux.DeleteDSL {
	dsl.vppDelete.ACL(aclName)
	return dsl
}

// L4Features adds a request to enable or disable L4 features
func (dsl *DeleteDSL) L4Features() linux.DeleteDSL {
	dsl.vppDelete.L4Features()
	return dsl
}

// AppNamespace adds a request to delete VPP Application namespace
// Note: current version does not support application namespace deletion
func (dsl *DeleteDSL) AppNamespace(id string) linux.DeleteDSL {
	dsl.vppDelete.AppNamespace(id)
	return dsl
}

// Arp adds a request to delete an existing VPP L3 ARP.
func (dsl *DeleteDSL) Arp(ifaceName string, ipAddr string) linux.DeleteDSL {
	dsl.vppDelete.Arp(ifaceName, ipAddr)
	return dsl
}

// StnRule adds a request to delete an existing VPP Stn rule.
func (dsl *DeleteDSL) StnRule(ruleName string) linux.DeleteDSL {
	dsl.vppDelete.StnRule(ruleName)
	return dsl
}

// Put changes the DSL mode to allow configuration editing.
func (dsl *DeleteDSL) Put() linux.PutDSL {
	return &PutDSL{dsl.parent, dsl.vppDelete.Put()}
}

// Send propagates requested changes to the plugins.
func (dsl *DeleteDSL) Send() vpp_clientv1.Reply {
	return dsl.parent.Send()
}
