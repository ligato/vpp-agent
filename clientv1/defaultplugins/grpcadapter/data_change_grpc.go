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

package grpcadapter

import (
	"strconv"

	"github.com/ligato/vpp-agent/clientv1/defaultplugins"
	"github.com/ligato/vpp-agent/flavors/rpc/model/vppsvc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l4"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/stn"
	"golang.org/x/net/context"
)

// NewDataChangeDSL is a constructor
func NewDataChangeDSL(client vppsvc.ChangeConfigServiceClient) *DataChangeDSL {
	return &DataChangeDSL{client,
		map[string] /*name*/ *interfaces.Interfaces_Interface{},
		map[string] /*name*/ *bfd.SingleHopBFD_Session{},
		map[string] /*id*/ *bfd.SingleHopBFD_Key{},
		map[string] /*name*/ *bfd.SingleHopBFD_EchoFunction{},
		map[string] /*name*/ *l2.BridgeDomains_BridgeDomain{},
		map[string] /*key*/ *l2.FibTableEntries_FibTableEntry{},
		map[string] /*name*/ *l2.XConnectPairs_XConnectPair{},
		map[string] /*key*/ *l3.StaticRoutes_Route{},
		map[string] /*name*/ *acl.AccessLists_Acl{},
		map[string] /*value*/ *l4.L4Features{},
		map[string] /*id*/ *l4.AppNamespaces_AppNamespace{},
		map[string] /*name*/ *l3.ArpTable_ArpTableEntry{},
		map[string] /*name*/ *stn.StnRule{},
		map[string] /*label*/ *nat.Nat44Global{},
		map[string] /*value*/ *nat.Nat44DNat_DNatConfig{},

		map[string] /*name*/ *struct{}{},
		map[string] /*name*/ *struct{}{},
		map[string] /*id*/ *struct{}{},
		map[string] /*name*/ *struct{}{},
		map[string] /*name*/ *struct{}{},
		map[string] /*key*/ *struct{}{},
		map[string] /*name*/ *struct{}{},
		map[string] /*key*/ *vppsvc.DelStaticRoutesRequest_DelStaticRoute{},
		map[string] /*name*/ *struct{}{},
		map[string] /*id*/ *l4.L4Features{},
		map[string] /*value*/ *l4.AppNamespaces_AppNamespace{},
		map[string] /*key*/ *l3.ArpTable_ArpTableEntry{},
		map[string] /*name*/ *stn.StnRule{},
		map[string] /*label*/ *nat.Nat44Global{},
		map[string] /*value*/ *nat.Nat44DNat_DNatConfig{},
	}
}

// DataChangeDSL is used to conveniently assign all the data that are needed for the DataChange.
// This is an implementation of Domain Specific Language (DSL) for a change of the VPP configuration.
type DataChangeDSL struct {
	client            vppsvc.ChangeConfigServiceClient
	txnPutIntf        map[string] /*name*/ *interfaces.Interfaces_Interface
	txnPutBfdSession  map[string] /*name*/ *bfd.SingleHopBFD_Session
	txnPutBfdAuthKey  map[string] /*id*/ *bfd.SingleHopBFD_Key
	txnPutBfdEcho     map[string] /*name*/ *bfd.SingleHopBFD_EchoFunction
	txnPutBD          map[string] /*name*/ *l2.BridgeDomains_BridgeDomain
	txnPutBDFIB       map[string] /*key*/ *l2.FibTableEntries_FibTableEntry
	txnPutXCon        map[string] /*name*/ *l2.XConnectPairs_XConnectPair
	txnPutStaticRoute map[string] /*key*/ *l3.StaticRoutes_Route
	txnPutACL         map[string] /*name*/ *acl.AccessLists_Acl
	txnPutL4Features  map[string] /*value*/ *l4.L4Features
	txnPutAppNs       map[string] /*id*/ *l4.AppNamespaces_AppNamespace
	txnPutArp         map[string] /*key*/ *l3.ArpTable_ArpTableEntry
	txnPutStn         map[string] /*value*/ *stn.StnRule
	txnPutNatGlobal   map[string] /*id*/ *nat.Nat44Global
	txnPutDNat        map[string] /*key*/ *nat.Nat44DNat_DNatConfig

	txnDelIntf        map[string] /*name*/ *struct{}
	txnDelBfdSession  map[string] /*name*/ *struct{}
	txnDelBfdAuthKey  map[string] /*id*/ *struct{}
	txnDelBfdEcho     map[string] /*name*/ *struct{}
	txnDelBD          map[string] /*name*/ *struct{}
	txnDelBDFIB       map[string] /*key*/ *struct{}
	txnDelXCon        map[string] /*name*/ *struct{}
	txnDelStaticRoute map[string] /*key*/ *vppsvc.DelStaticRoutesRequest_DelStaticRoute
	txnDelACL         map[string] /*name*/ *struct{}
	txnDelL4Features  map[string] /*value*/ *l4.L4Features
	txnDelAppNs       map[string] /*id*/ *l4.AppNamespaces_AppNamespace
	txnDelArp         map[string] /*value*/ *l3.ArpTable_ArpTableEntry
	txnDelStn         map[string] /*value*/ *stn.StnRule
	txnDelNatGlobal   map[string] /*id*/ *nat.Nat44Global
	txnDelDNat        map[string] /*key*/ *nat.Nat44DNat_DNatConfig
}

// PutDSL allows to add or edit the configuration of delault plugins based on grpc requests.
type PutDSL struct {
	parent *DataChangeDSL
}

// DeleteDSL allows to remove the configuration of delault plugins based on grpc requests.
type DeleteDSL struct {
	parent *DataChangeDSL
}

// Interface creates or updates the network interface.
func (dsl *PutDSL) Interface(val *interfaces.Interfaces_Interface) defaultplugins.PutDSL {
	dsl.parent.txnPutIntf[val.Name] = val
	return dsl
}

// BfdSession creates or updates the bidirectional forwarding detection session.
func (dsl *PutDSL) BfdSession(val *bfd.SingleHopBFD_Session) defaultplugins.PutDSL {
	dsl.parent.txnPutBfdSession[val.Interface] = val
	return dsl
}

// BfdAuthKeys creates or updates the bidirectional forwarding detection key.
func (dsl *PutDSL) BfdAuthKeys(val *bfd.SingleHopBFD_Key) defaultplugins.PutDSL {
	dsl.parent.txnPutBfdAuthKey[strconv.Itoa(int(val.Id))] = val
	return dsl
}

// BfdEchoFunction creates or updates the bidirectional forwarding detection echo function.
func (dsl *PutDSL) BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) defaultplugins.PutDSL {
	dsl.parent.txnPutBfdEcho[val.EchoSourceInterface] = val
	return dsl
}

// BD creates or updates the Bridge Domain.
func (dsl *PutDSL) BD(val *l2.BridgeDomains_BridgeDomain) defaultplugins.PutDSL {
	dsl.parent.txnPutBD[val.Name] = val

	return dsl
}

// BDFIB deletes request for the L2 Forwarding Information Base.
func (dsl *PutDSL) BDFIB(val *l2.FibTableEntries_FibTableEntry) defaultplugins.PutDSL {
	dsl.parent.txnPutBDFIB[l2.FibKey(val.BridgeDomain, val.PhysAddress)] = val

	return dsl
}

// XConnect creates or updates the Cross Connect.
func (dsl *PutDSL) XConnect(val *l2.XConnectPairs_XConnectPair) defaultplugins.PutDSL {
	dsl.parent.txnPutXCon[val.ReceiveInterface] = val

	return dsl
}

// StaticRoute creates or updates the L3 Static Route.
func (dsl *PutDSL) StaticRoute(val *l3.StaticRoutes_Route) defaultplugins.PutDSL {
	dsl.parent.txnPutStaticRoute[l3.RouteKey(val.VrfID, val.DstIPAddr, val.NextHopAddr)] = val

	return dsl
}

// ACL creates or updates request for the Access Control List.
func (dsl *PutDSL) ACL(val *acl.AccessLists_Acl) defaultplugins.PutDSL {
	dsl.parent.txnPutACL[val.AclName] = val

	return dsl
}

// L4Features creates or updates the request for the L4Features.
func (dsl *PutDSL) L4Features(val *l4.L4Features) defaultplugins.PutDSL {
	dsl.parent.txnPutL4Features[strconv.FormatBool(val.Enabled)] = val

	return dsl
}

// AppNamespace creates or updates the request for the Application Namespaces List.
func (dsl *PutDSL) AppNamespace(val *l4.AppNamespaces_AppNamespace) defaultplugins.PutDSL {
	dsl.parent.txnPutAppNs[val.NamespaceId] = val

	return dsl
}

// Arp adds a request to create or update VPP L3 ARP entry.
func (dsl *PutDSL) Arp(arp *l3.ArpTable_ArpTableEntry) defaultplugins.PutDSL {
	dsl.parent.txnPutArp[l3.ArpEntryKey(arp.Interface, arp.IpAddress)] = arp
	return dsl
}

// StnRule adds a request to create or update STN rule.
func (dsl *PutDSL) StnRule(val *stn.StnRule) defaultplugins.PutDSL {
	dsl.parent.txnPutStn[val.RuleName] = val
	return dsl
}

// NAT44Global adds a request to set global configuration for NAT44
func (dsl *PutDSL) NAT44Global(nat44 *nat.Nat44Global) defaultplugins.PutDSL {
	dsl.parent.txnPutNatGlobal["global"] = nat44
	return dsl
}

// NAT44DNat adds a request to create a new DNAT configuration
func (dsl *PutDSL) NAT44DNat(nat44 *nat.Nat44DNat_DNatConfig) defaultplugins.PutDSL {
	dsl.parent.txnPutDNat[nat.DNatKey(nat44.Label)] = nat44
	return dsl
}

// Put enables creating Interface/BD...
func (dsl *DataChangeDSL) Put() defaultplugins.PutDSL {
	return &PutDSL{dsl}
}

// Delete enables deleting Interface/BD...
func (dsl *DataChangeDSL) Delete() defaultplugins.DeleteDSL {
	return &DeleteDSL{dsl}
}

// Delete enables deleting Interface/BD...
func (dsl *PutDSL) Delete() defaultplugins.DeleteDSL {
	return &DeleteDSL{dsl.parent}
}

// Send propagates changes to the channels.
func (dsl *PutDSL) Send() defaultplugins.Reply {
	return dsl.parent.Send()
}

// Interface deletes request for the network interface.
func (dsl *DeleteDSL) Interface(interfaceName string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelIntf[interfaceName] = nil

	return dsl
}

// BfdSession adds a request to delete an existing bidirectional forwarding
// detection session.
func (dsl *DeleteDSL) BfdSession(bfdSessionIfaceName string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelBfdSession[bfdSessionIfaceName] = nil
	return dsl
}

// BfdAuthKeys adds a request to delete an existing bidirectional forwarding
// detection key.
func (dsl *DeleteDSL) BfdAuthKeys(bfdKeyID string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelBfdAuthKey[bfdKeyID] = nil
	return dsl
}

// BfdEchoFunction adds a request to delete an existing bidirectional forwarding
// detection echo function.
func (dsl *DeleteDSL) BfdEchoFunction(bfdEchoName string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelBfdEcho[bfdEchoName] = nil
	return dsl
}

// BD deletes request for the Bridge Domain.
func (dsl *DeleteDSL) BD(bdName string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelBD[bdName] = nil

	return dsl
}

// BDFIB deletes request for the L2 Forwarding Information Base.
func (dsl *DeleteDSL) BDFIB(bdName string, mac string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelBDFIB[l2.FibKey(bdName, mac)] = nil

	return dsl
}

// XConnect deletes the Cross Connect.
func (dsl *DeleteDSL) XConnect(rxIfName string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelXCon[rxIfName] = nil

	return dsl
}

// StaticRoute deletes the L3 Static Route.
func (dsl *DeleteDSL) StaticRoute(vrf uint32, dstAddr string, nextHopAddr string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelStaticRoute[l3.RouteKey(vrf, dstAddr, nextHopAddr)] =
		&vppsvc.DelStaticRoutesRequest_DelStaticRoute{vrf, dstAddr, nextHopAddr}

	return dsl
}

// ACL deletes request for Access Control List.
func (dsl *DeleteDSL) ACL(aclName string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelACL[aclName] = nil

	return dsl
}

// L4Features deletes request for the L4Features.
func (dsl *DeleteDSL) L4Features() defaultplugins.DeleteDSL {
	dsl.parent.txnDelL4Features["l4"] = nil

	return dsl
}

// AppNamespace delets request for the Application Namespaces List.
func (dsl *DeleteDSL) AppNamespace(id string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelAppNs[id] = nil

	return dsl
}

// Arp adds a request to delete an existing VPP L3 ARP entry.
func (dsl *DeleteDSL) Arp(ifaceName string, ipAddr string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelArp[l3.ArpEntryKey(ifaceName, ipAddr)] = nil
	return dsl
}

// StnRule adds request to delete Stn rule.
func (dsl *DeleteDSL) StnRule(name string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelStn[name] = nil
	return dsl
}

// NAT44Global adds a request to remove global configuration for NAT44
func (dsl *DeleteDSL) NAT44Global() defaultplugins.DeleteDSL {
	dsl.parent.txnPutNatGlobal["global"] = nil
	return dsl
}

// NAT44DNat adds a request to delete a new DNAT configuration
func (dsl *DeleteDSL) NAT44DNat(label string) defaultplugins.DeleteDSL {
	dsl.parent.txnPutDNat[nat.DNatKey(label)] = nil
	return dsl
}

// Put enables creating Interface/BD...
func (dsl *DeleteDSL) Put() defaultplugins.PutDSL {
	return &PutDSL{dsl.parent}
}

// Send propagates changes to the channels.
func (dsl *DeleteDSL) Send() defaultplugins.Reply {
	return dsl.parent.Send()
}

// Send propagates changes to the channels.
func (dsl *DataChangeDSL) Send() defaultplugins.Reply {
	var wasErr error

	if len(dsl.txnPutIntf) > 0 {
		putIntfs := []*interfaces.Interfaces_Interface{}
		for _, intf := range dsl.txnPutIntf {
			putIntfs = append(putIntfs, intf)
		}
		_, err := dsl.client.PutInterfaces(context.Background(), &interfaces.Interfaces{Interface: putIntfs})
		if err != nil {
			wasErr = err
		}
	}
	if len(dsl.txnPutBD) > 0 {
		putBDs := []*l2.BridgeDomains_BridgeDomain{}
		for _, bd := range dsl.txnPutBD {
			putBDs = append(putBDs, bd)
		}
		_, err := dsl.client.PutBDs(context.Background(), &l2.BridgeDomains{BridgeDomains: putBDs})
		if err != nil {
			wasErr = err
		}
	}
	if len(dsl.txnPutXCon) > 0 {
		putXCons := []*l2.XConnectPairs_XConnectPair{}
		for _, xcon := range dsl.txnPutXCon {
			putXCons = append(putXCons, xcon)
		}
		_, err := dsl.client.PutXCons(context.Background(), &l2.XConnectPairs{XConnectPairs: putXCons})
		if err != nil {
			wasErr = err
		}
	}
	if len(dsl.txnPutStaticRoute) > 0 {
		putRoutes := []*l3.StaticRoutes_Route{}
		for _, route := range dsl.txnPutStaticRoute {
			putRoutes = append(putRoutes, route)
		}
		_, err := dsl.client.PutStaticRoutes(context.Background(), &l3.StaticRoutes{Route: putRoutes})
		if err != nil {
			wasErr = err
		}
	}
	if len(dsl.txnPutACL) > 0 {
		putACLs := []*acl.AccessLists_Acl{}
		for _, acl := range dsl.txnPutACL {
			putACLs = append(putACLs, acl)
		}
		_, err := dsl.client.PutACLs(context.Background(), &acl.AccessLists{Acl: putACLs})
		if err != nil {
			wasErr = err
		}
	}

	if len(dsl.txnDelIntf) > 0 {
		delIntfs := []string{}
		for intfName := range dsl.txnDelIntf {
			delIntfs = append(delIntfs, intfName)
		}
		_, err := dsl.client.DelInterfaces(context.Background(), &vppsvc.DelNamesRequest{delIntfs})
		if err != nil {
			wasErr = err
		}
	}
	if len(dsl.txnDelBD) > 0 {
		bdNames := []string{}
		for intfName := range dsl.txnDelBD {
			bdNames = append(bdNames, intfName)
		}
		_, err := dsl.client.DelBDs(context.Background(), &vppsvc.DelNamesRequest{bdNames})
		if err != nil {
			wasErr = err
		}
	}
	if len(dsl.txnDelXCon) > 0 {
		delXCons := []string{}
		for intfName := range dsl.txnDelXCon {
			delXCons = append(delXCons, intfName)
		}
		_, err := dsl.client.DelXCons(context.Background(), &vppsvc.DelNamesRequest{delXCons})
		if err != nil {
			wasErr = err
		}
	}
	if len(dsl.txnDelStaticRoute) > 0 {
		delRoutes := []*vppsvc.DelStaticRoutesRequest_DelStaticRoute{}
		for _, route := range dsl.txnDelStaticRoute {
			delRoutes = append(delRoutes, route)
		}
		_, err := dsl.client.DelStaticRoutes(context.Background(), &vppsvc.DelStaticRoutesRequest{delRoutes})
		if err != nil {
			wasErr = err
		}
	}
	if len(dsl.txnDelACL) > 0 {
		delACLs := []string{}
		for intfName := range dsl.txnDelACL {
			delACLs = append(delACLs, intfName)
		}
		_, err := dsl.client.DelACLs(context.Background(), &vppsvc.DelNamesRequest{delACLs})
		if err != nil {
			wasErr = err
		}
	}

	return &Reply{wasErr}
}

// Reply enables waiting for the reply and getting result (success/error).
type Reply struct {
	err error
}

// ReceiveReply returns error or nil.
func (dsl Reply) ReceiveReply() error {
	return dsl.err
}
