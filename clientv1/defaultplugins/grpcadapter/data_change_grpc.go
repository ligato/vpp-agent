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
	"github.com/ligato/vpp-agent/clientv1/defaultplugins"
	"github.com/ligato/vpp-agent/flavors/rpc/model/vppsvc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"golang.org/x/net/context"
	"net"
)

// NewDataChangeDSL is a constructor
func NewDataChangeDSL(client vppsvc.ChangeConfigServiceClient) *DataChangeDSL {
	return &DataChangeDSL{client,
		map[string] /*name*/ *interfaces.Interfaces_Interface{},
		map[string] /*name*/ *bfd.SingleHopBFD_Session{},
		map[uint32] /*id*/ *bfd.SingleHopBFD_Key{},
		map[string] /*name*/ *bfd.SingleHopBFD_EchoFunction{},
		map[string] /*name*/ *l2.BridgeDomains_BridgeDomain{},
		map[string] /*key*/ *l2.FibTableEntries_FibTableEntry{},
		map[string] /*name*/ *l2.XConnectPairs_XConnectPair{},
		map[string] /*key*/ *l3.StaticRoutes_Route{},
		map[string] /*name*/ *acl.AccessLists_Acl{},

		map[string] /*name*/ *struct{}{},
		map[string] /*name*/ *struct{}{},
		map[uint32] /*id*/ *struct{}{},
		map[string] /*name*/ *struct{}{},
		map[string] /*name*/ *struct{}{},
		map[string] /*key*/ *struct{}{},
		map[string] /*name*/ *struct{}{},
		map[string] /*key*/ *vppsvc.DelStaticRoutesRequest_DelStaticRoute{},
		map[string] /*name*/ *struct{}{},
	}
}

// DataChangeDSL is used to conveniently assign all the data that are needed for the DataChange.
// This is implementation of Domain Specific Language (DSL) for change of the VPP configuration.
type DataChangeDSL struct {
	client            vppsvc.ChangeConfigServiceClient
	txnPutIntf        map[string] /*name*/ *interfaces.Interfaces_Interface
	txnPutBfdSession  map[string] /*name*/ *bfd.SingleHopBFD_Session
	txnPutBfdAuthKey  map[uint32] /*id*/ *bfd.SingleHopBFD_Key
	txnPutBfdEcho     map[string] /*name*/ *bfd.SingleHopBFD_EchoFunction
	txnPutBD          map[string] /*name*/ *l2.BridgeDomains_BridgeDomain
	txnPutBDFIB       map[string] /*key*/ *l2.FibTableEntries_FibTableEntry
	txnPutXCon        map[string] /*name*/ *l2.XConnectPairs_XConnectPair
	txnPutStaticRoute map[string] /*key*/ *l3.StaticRoutes_Route
	txnPutACL         map[string] /*name*/ *acl.AccessLists_Acl

	txnDelIntf        map[string] /*name*/ *struct{}
	txnDelBfdSession  map[string] /*name*/ *struct{}
	txnDelBfdAuthKey  map[uint32] /*id*/ *struct{}
	txnDelBfdEcho     map[string] /*name*/ *struct{}
	txnDelBD          map[string] /*name*/ *struct{}
	txnDelBDFIB       map[string] /*key*/ *struct{}
	txnDelXCon        map[string] /*name*/ *struct{}
	txnDelStaticRoute map[string] /*key*/ *vppsvc.DelStaticRoutesRequest_DelStaticRoute
	txnDelACL         map[string] /*name*/ *struct{}
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
	dsl.parent.txnPutIntf[val.Name] = val
	return dsl
}

// BfdSession create or update the bidirectional forwarding detection session
func (dsl *PutDSL) BfdSession(val *bfd.SingleHopBFD_Session) defaultplugins.PutDSL {
	dsl.parent.txnPutBfdSession[val.Interface] = val
	return dsl
}

// BfdAuthKeys create or update the bidirectional forwarding detection key
func (dsl *PutDSL) BfdAuthKeys(val *bfd.SingleHopBFD_Key) defaultplugins.PutDSL {
	dsl.parent.txnPutBfdAuthKey[val.Id] = val
	return dsl
}

// BfdEchoFunction create or update the bidirectional forwarding detection echod function
func (dsl *PutDSL) BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) defaultplugins.PutDSL {
	dsl.parent.txnPutBfdEcho[val.EchoSourceInterface] = val
	return dsl
}

// BD create or update the Bridge Domain
func (dsl *PutDSL) BD(val *l2.BridgeDomains_BridgeDomain) defaultplugins.PutDSL {
	dsl.parent.txnPutBD[val.Name] = val

	return dsl
}

// BDFIB delete request for the L2 Forwarding Information Base
func (dsl *PutDSL) BDFIB(val *l2.FibTableEntries_FibTableEntry) defaultplugins.PutDSL {
	dsl.parent.txnPutBDFIB[l2.FibKey(val.BridgeDomain, val.PhysAddress)] = val

	return dsl
}

// XConnect create or update the Cross Connect
func (dsl *PutDSL) XConnect(val *l2.XConnectPairs_XConnectPair) defaultplugins.PutDSL {
	dsl.parent.txnPutXCon[val.ReceiveInterface] = val

	return dsl
}

// StaticRoute create or update the L3 Static Route
func (dsl *PutDSL) StaticRoute(val *l3.StaticRoutes_Route) defaultplugins.PutDSL {
	_, dstAddr, _ := net.ParseCIDR(val.DstIpAddr)
	dsl.parent.txnPutStaticRoute[l3.RouteKey(val.VrfId, dstAddr, val.NextHopAddr)] = val

	return dsl
}

// ACL create or update request for the Access Control List
func (dsl *PutDSL) ACL(val *acl.AccessLists_Acl) defaultplugins.PutDSL {
	dsl.parent.txnPutACL[val.AclName] = val

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
	dsl.parent.txnDelIntf[interfaceName] = nil

	return dsl
}

// BD create or update the Bridge Domain
func (dsl *DeleteDSL) BD(bdName string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelBD[bdName] = nil

	return dsl
}

// BDFIB delete request for the L2 Forwarding Information Base
func (dsl *DeleteDSL) BDFIB(bdName string, mac string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelBDFIB[l2.FibKey(bdName, mac)] = nil

	return dsl
}

// XConnect create or update the Cross Connect
func (dsl *DeleteDSL) XConnect(rxIfName string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelXCon[rxIfName] = nil

	return dsl
}

// StaticRoute create or update the L3 Static Route
func (dsl *DeleteDSL) StaticRoute(vrf uint32, dstAddrInput *net.IPNet, nextHopAddr net.IP) defaultplugins.DeleteDSL {
	//_, dstAddr, _ := net.ParseCIDR(dstAddrInput)
	dsl.parent.txnDelStaticRoute[l3.RouteKey(vrf, dstAddrInput, nextHopAddr.String())] =
		&vppsvc.DelStaticRoutesRequest_DelStaticRoute{vrf, dstAddrInput.String(), nextHopAddr.String()}

	return dsl
}

// ACL delete request for Access Control List
func (dsl *DeleteDSL) ACL(aclName string) defaultplugins.DeleteDSL {
	dsl.parent.txnDelACL[aclName] = nil

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

// Reply is here to gives you the ability to wait for the reply and get result (success/error)
type Reply struct {
	err error
}

// ReceiveReply return error or nil
func (dsl Reply) ReceiveReply() error {
	return dsl.err
}
