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
	"github.com/ligato/vpp-agent/clientv2/vpp"
	"github.com/ligato/vpp-agent/plugins/vpp/model/bfd"
	"github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l4"
	"github.com/ligato/vpp-agent/plugins/vpp/model/stn"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/acl"
	intf "github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/l2"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/l3"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/nat"
)

// NewDataChangeDSL returns a new instance of DataChangeDSL which implements
// the data change DSL for VPP config.
// Transaction <txn> is used to propagate changes to plugins.
func NewDataChangeDSL(txn keyval.ProtoTxn) *DataChangeDSL {
	return &DataChangeDSL{txn: txn}
}

// DataChangeDSL is an implementation of Domain Specific Language (DSL)
// for changes of the VPP configuration.
type DataChangeDSL struct {
	txn keyval.ProtoTxn
}

// PutDSL implements put operations of data change DSL.
type PutDSL struct {
	parent *DataChangeDSL
}

// DeleteDSL implements delete operations of data change DSL.
type DeleteDSL struct {
	parent *DataChangeDSL
}

// Put initiates a chained sequence of data change DSL statements declaring
// new configurable objects or changing existing ones.
func (dsl *DataChangeDSL) Put() vppclient.PutDSL {
	return &PutDSL{dsl}
}

// Delete initiates a chained sequence of data change DSL statements
// removing existing configurable objects.
func (dsl *DataChangeDSL) Delete() vppclient.DeleteDSL {
	return &DeleteDSL{dsl}
}

// Send propagates requested changes to the plugins.
func (dsl *DataChangeDSL) Send() vppclient.Reply {
	err := dsl.txn.Commit()
	return &Reply{err}
}

// Interface adds a request to create or update VPP network interface.
func (dsl *PutDSL) Interface(val *intf.Interface) vppclient.PutDSL {
	dsl.parent.txn.Put(intf.InterfaceKey(val.Name), val)
	return dsl
}

// ACL adds a request to create or update VPP Access Control List.
func (dsl *PutDSL) ACL(val *acl.Acl) vppclient.PutDSL {
	dsl.parent.txn.Put(acl.Key(val.Name), val)
	return dsl
}

// BfdSession adds a request to create or update bidirectional forwarding
// detection session.
func (dsl *PutDSL) BfdSession(val *bfd.SingleHopBFD_Session) vppclient.PutDSL {
	dsl.parent.txn.Put(bfd.SessionKey(val.Interface), val)
	return dsl
}

// BfdAuthKeys adds a request to create or update bidirectional forwarding
// detection key.
func (dsl *PutDSL) BfdAuthKeys(val *bfd.SingleHopBFD_Key) vppclient.PutDSL {
	dsl.parent.txn.Put(bfd.AuthKeysKey(string(val.Id)), val)
	return dsl
}

// BfdEchoFunction adds a request to create or update bidirectional forwarding
// detection echo function.
func (dsl *PutDSL) BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) vppclient.PutDSL {
	dsl.parent.txn.Put(bfd.EchoFunctionKey(val.EchoSourceInterface), val)
	return dsl
}

// BD adds a request to create or update VPP Bridge Domain.
func (dsl *PutDSL) BD(val *l2.BridgeDomain) vppclient.PutDSL {
	dsl.parent.txn.Put(l2.BridgeDomainKey(val.Name), val)
	return dsl
}

// BDFIB adds a request to create or update VPP L2 Forwarding Information Base.
func (dsl *PutDSL) BDFIB(val *l2.FIBEntry) vppclient.PutDSL {
	dsl.parent.txn.Put(l2.FIBKey(val.BridgeDomain, val.PhysAddress), val)
	return dsl
}

// XConnect adds a request to create or update VPP Cross Connect.
func (dsl *PutDSL) XConnect(val *l2.XConnectPair) vppclient.PutDSL {
	dsl.parent.txn.Put(l2.XConnectKey(val.ReceiveInterface), val)
	return dsl
}

// StaticRoute adds a request to create or update VPP L3 Static Route.
func (dsl *PutDSL) StaticRoute(val *l3.StaticRoute) vppclient.PutDSL {
	dsl.parent.txn.Put(l3.RouteKey(val.VrfId, val.DstNetwork, val.NextHopAddr), val)
	return dsl
}

// Arp adds a request to create or update VPP L3 ARP entry.
func (dsl *PutDSL) Arp(arp *l3.ARPEntry) vppclient.PutDSL {
	dsl.parent.txn.Put(l3.ArpEntryKey(arp.Interface, arp.IpAddress), arp)
	return dsl
}

// ProxyArp adds a request to create or update VPP L3 proxy ARP.
func (dsl *PutDSL) ProxyArp(proxyArp *l3.ProxyARP) vppclient.PutDSL {
	dsl.parent.txn.Put(l3.ProxyARPKey, proxyArp)
	return dsl
}

// IPScanNeighbor adds L3 IP Scan Neighbor to the RESYNC request.
func (dsl *PutDSL) IPScanNeighbor(ipScanNeigh *l3.IPScanNeighbor) vppclient.PutDSL {
	dsl.parent.txn.Put(l3.IPScanNeighborKey, ipScanNeigh)
	return dsl
}

// L4Features create or update request for the L4Features
func (dsl *PutDSL) L4Features(val *l4.L4Features) vppclient.PutDSL {
	dsl.parent.txn.Put(l4.FeatureKey(), val)

	return dsl
}

// AppNamespace create or update request for the Application Namespaces List
func (dsl *PutDSL) AppNamespace(val *l4.AppNamespaces_AppNamespace) vppclient.PutDSL {
	dsl.parent.txn.Put(l4.AppNamespacesKey(val.NamespaceId), val)

	return dsl
}

// StnRule adds a request to create or update STN rule.
func (dsl *PutDSL) StnRule(val *stn.STN_Rule) vppclient.PutDSL {
	dsl.parent.txn.Put(stn.Key(val.RuleName), val)
	return dsl
}

// NAT44Global adds a request to set global configuration for NAT44
func (dsl *PutDSL) NAT44Global(nat44 *nat.Nat44Global) vppclient.PutDSL {
	dsl.parent.txn.Put(nat.GlobalNAT44Key, nat44)
	return dsl
}

// DNAT44 adds a request to create or update DNAT44 configuration
func (dsl *PutDSL) DNAT44(nat44 *nat.DNat44) vppclient.PutDSL {
	dsl.parent.txn.Put(nat.DNAT44Key(nat44.Label), nat44)
	return dsl
}

// IPSecSA adds request to create a new Security Association
func (dsl *PutDSL) IPSecSA(sa *ipsec.SecurityAssociations_SA) vppclient.PutDSL {
	dsl.parent.txn.Put(ipsec.SAKey(sa.Name), sa)
	return dsl
}

// IPSecSPD adds request to create a new Security Policy Database
func (dsl *PutDSL) IPSecSPD(spd *ipsec.SecurityPolicyDatabases_SPD) vppclient.PutDSL {
	dsl.parent.txn.Put(ipsec.SPDKey(spd.Name), spd)
	return dsl
}

// Delete changes the DSL mode to allow removal of an existing configuration.
func (dsl *PutDSL) Delete() vppclient.DeleteDSL {
	return &DeleteDSL{dsl.parent}
}

// Send propagates requested changes to the plugins.
func (dsl *PutDSL) Send() vppclient.Reply {
	return dsl.parent.Send()
}

// Interface adds a request to delete an existing VPP network interface.
func (dsl *DeleteDSL) Interface(interfaceName string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(intf.InterfaceKey(interfaceName))
	return dsl
}

// ACL adds a request to delete an existing VPP Access Control List.
func (dsl *DeleteDSL) ACL(aclName string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(acl.Key(aclName))
	return dsl
}

// BfdSession adds a request to delete an existing bidirectional forwarding
// detection session.
func (dsl *DeleteDSL) BfdSession(bfdSessionIfaceName string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(bfd.SessionKey(bfdSessionIfaceName))
	return dsl
}

// BfdAuthKeys adds a request to delete an existing bidirectional forwarding
// detection key.
func (dsl *DeleteDSL) BfdAuthKeys(bfdKey string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(bfd.AuthKeysKey(bfdKey))
	return dsl
}

// BfdEchoFunction adds a request to delete an existing bidirectional forwarding
// detection echo function.
func (dsl *DeleteDSL) BfdEchoFunction(bfdEchoName string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(bfd.EchoFunctionKey(bfdEchoName))
	return dsl
}

// BD adds a request to delete an existing VPP Bridge Domain.
func (dsl *DeleteDSL) BD(bdName string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(l2.BridgeDomainKey(bdName))
	return dsl
}

// BDFIB adds a request to delete an existing VPP L2 Forwarding Information
// Base.
func (dsl *DeleteDSL) BDFIB(bdName string, mac string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(l2.FIBKey(bdName, mac))
	return dsl
}

// XConnect adds a request to delete an existing VPP Cross Connect.
func (dsl *DeleteDSL) XConnect(rxIfName string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(l2.XConnectKey(rxIfName))
	return dsl
}

// StaticRoute adds a request to delete an existing VPP L3 Static Route.
func (dsl *DeleteDSL) StaticRoute(vrf uint32, dstAddr string, nextHopAddr string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(l3.RouteKey(vrf, dstAddr, nextHopAddr))
	return dsl
}

// Arp adds a request to delete an existing VPP L3 ARP entry.
func (dsl *DeleteDSL) Arp(ifaceName string, ipAddr string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(l3.ArpEntryKey(ifaceName, ipAddr))
	return dsl
}

// ProxyArp adds a request to delete an existing VPP L3 proxy ARP.
func (dsl *DeleteDSL) ProxyArp() vppclient.DeleteDSL {
	dsl.parent.txn.Delete(l3.ProxyARPKey)
	return dsl
}

// IPScanNeighbor adds a request to delete an existing VPP L3 IP Scan Neighbor.
func (dsl *DeleteDSL) IPScanNeighbor() vppclient.DeleteDSL {
	dsl.parent.txn.Delete(l3.IPScanNeighborKey)
	return dsl
}

// L4Features delete request for the L4Features
func (dsl *DeleteDSL) L4Features() vppclient.DeleteDSL {
	dsl.parent.txn.Delete(l4.FeatureKey())
	return dsl
}

// AppNamespace adds a request to delete an existing VPP Application Namespace.
func (dsl *DeleteDSL) AppNamespace(id string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(l4.AppNamespacesKey(id))
	return dsl
}

// StnRule adds request to delete Stn rule.
func (dsl *DeleteDSL) StnRule(ruleName string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(stn.Key(ruleName))
	return dsl
}

// NAT44Global adds a request to remove global configuration for NAT44
func (dsl *DeleteDSL) NAT44Global() vppclient.DeleteDSL {
	dsl.parent.txn.Delete(nat.GlobalNAT44Key)
	return dsl
}

// DNAT44 adds a request to delete an existing DNAT44 configuration
func (dsl *DeleteDSL) DNAT44(label string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(nat.DNAT44Key(label))
	return dsl
}

// IPSecSA adds request to create a new Security Association
func (dsl *DeleteDSL) IPSecSA(saName string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(ipsec.SAKey(saName))
	return dsl
}

// IPSecSPD adds request to create a new Security Policy Database
func (dsl *DeleteDSL) IPSecSPD(spdName string) vppclient.DeleteDSL {
	dsl.parent.txn.Delete(ipsec.SPDKey(spdName))
	return dsl
}

// Put changes the DSL mode to allow configuration editing.
func (dsl *DeleteDSL) Put() vppclient.PutDSL {
	return &PutDSL{dsl.parent}
}

// Send propagates requested changes to the plugins.
func (dsl *DeleteDSL) Send() vppclient.Reply {
	return dsl.parent.Send()
}

// Reply interface allows to wait for a reply to previously called Send() and
// extract the result from it (success/error).
type Reply struct {
	err error
}

// ReceiveReply waits for a reply to previously called Send() and returns
// the result (error or nil).
func (dsl Reply) ReceiveReply() error {
	return dsl.err
}
