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
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/ipsec"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l4"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/rpc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/stn"
	"golang.org/x/net/context"
)

// NewDataChangeDSL is a constructor
func NewDataChangeDSL(client rpc.ChangeConfigServiceClient) *DataChangeDSL {
	return &DataChangeDSL{client, make([]proto.Message, 0), make([]proto.Message, 0)}
}

// DataChangeDSL is used to conveniently assign all the data that are needed for the DataChange.
// This is an implementation of Domain Specific Language (DSL) for a change of the VPP configuration.
type DataChangeDSL struct {
	client rpc.ChangeConfigServiceClient
	put    []proto.Message
	del    []proto.Message
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
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// BfdSession creates or updates the bidirectional forwarding detection session.
func (dsl *PutDSL) BfdSession(val *bfd.SingleHopBFD_Session) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// BfdAuthKeys creates or updates the bidirectional forwarding detection key.
func (dsl *PutDSL) BfdAuthKeys(val *bfd.SingleHopBFD_Key) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// BfdEchoFunction creates or updates the bidirectional forwarding detection echo function.
func (dsl *PutDSL) BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// BD creates or updates the Bridge Domain.
func (dsl *PutDSL) BD(val *l2.BridgeDomains_BridgeDomain) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// BDFIB deletes request for the L2 Forwarding Information Base.
func (dsl *PutDSL) BDFIB(val *l2.FibTable_FibEntry) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// XConnect creates or updates the Cross Connect.
func (dsl *PutDSL) XConnect(val *l2.XConnectPairs_XConnectPair) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// StaticRoute creates or updates the L3 Static Route.
func (dsl *PutDSL) StaticRoute(val *l3.StaticRoutes_Route) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// ACL creates or updates request for the Access Control List.
func (dsl *PutDSL) ACL(val *acl.AccessLists_Acl) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// L4Features creates or updates the request for the L4Features.
func (dsl *PutDSL) L4Features(val *l4.L4Features) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// AppNamespace creates or updates the request for the Application Namespaces List.
func (dsl *PutDSL) AppNamespace(val *l4.AppNamespaces_AppNamespace) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// Arp adds a request to create or update VPP L3 ARP entry.
func (dsl *PutDSL) Arp(val *l3.ArpTable_ArpEntry) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// ProxyArpInterfaces adds a request to create or update VPP L3 proxy ARP interfaces.
func (dsl *PutDSL) ProxyArpInterfaces(val *l3.ProxyArpInterfaces_InterfaceList) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// ProxyArpRanges adds a request to create or update VPP L3 proxy ARP ranges
func (dsl *PutDSL) ProxyArpRanges(val *l3.ProxyArpRanges_RangeList) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// StnRule adds a request to create or update STN rule.
func (dsl *PutDSL) StnRule(val *stn.STN_Rule) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// NAT44Global adds a request to set global configuration for NAT44
func (dsl *PutDSL) NAT44Global(val *nat.Nat44Global) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// NAT44DNat adds a request to create a new DNAT configuration
func (dsl *PutDSL) NAT44DNat(val *nat.Nat44DNat_DNatConfig) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// IPSecSA adds request to create a new Security Association
func (dsl *PutDSL) IPSecSA(val *ipsec.SecurityAssociations_SA) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
	return dsl
}

// IPSecSPD adds request to create a new Security Policy Database
func (dsl *PutDSL) IPSecSPD(val *ipsec.SecurityPolicyDatabases_SPD) defaultplugins.PutDSL {
	dsl.parent.put = append(dsl.parent.put, val)
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
	dsl.parent.del = append(dsl.parent.del, &interfaces.Interfaces_Interface{
		Name: interfaceName,
	})
	return dsl
}

// BfdSession adds a request to delete an existing bidirectional forwarding
// detection session.
func (dsl *DeleteDSL) BfdSession(ifName string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &bfd.SingleHopBFD_Session{
		Interface: ifName,
	})
	return dsl
}

// BfdAuthKeys adds a request to delete an existing bidirectional forwarding
// detection key.
func (dsl *DeleteDSL) BfdAuthKeys(bfdKeyID string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &bfd.SingleHopBFD_Key{
		Name: bfdKeyID,
	})
	return dsl
}

// BfdEchoFunction adds a request to delete an existing bidirectional forwarding
// detection echo function.
func (dsl *DeleteDSL) BfdEchoFunction(bfdEchoName string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &bfd.SingleHopBFD_EchoFunction{
		Name: bfdEchoName,
	})
	return dsl
}

// BD deletes request for the Bridge Domain.
func (dsl *DeleteDSL) BD(bdName string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &l2.BridgeDomains_BridgeDomain{
		Name: bdName,
	})
	return dsl
}

// BDFIB deletes request for the L2 Forwarding Information Base.
func (dsl *DeleteDSL) BDFIB(bdName string, mac string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &l2.FibTable_FibEntry{
		PhysAddress:  mac,
		BridgeDomain: bdName,
	})
	return dsl
}

// XConnect deletes the Cross Connect.
func (dsl *DeleteDSL) XConnect(rxIfName string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &l2.XConnectPairs_XConnectPair{
		ReceiveInterface: rxIfName,
	})
	return dsl
}

// StaticRoute deletes the L3 Static Route.
func (dsl *DeleteDSL) StaticRoute(vrf uint32, dstAddr string, nextHopAddr string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &l3.StaticRoutes_Route{
		VrfId:       vrf,
		DstIpAddr:   dstAddr,
		NextHopAddr: nextHopAddr,
	})
	return dsl
}

// ACL deletes request for Access Control List.
func (dsl *DeleteDSL) ACL(aclName string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &acl.AccessLists_Acl{
		AclName: aclName,
	})
	return dsl
}

// L4Features deletes request for the L4Features.
func (dsl *DeleteDSL) L4Features() defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &l4.L4Features{})
	return dsl
}

// AppNamespace delets request for the Application Namespaces List.
func (dsl *DeleteDSL) AppNamespace(id string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &l4.AppNamespaces_AppNamespace{
		NamespaceId: id,
	})
	return dsl
}

// Arp adds a request to delete an existing VPP L3 ARP entry.
func (dsl *DeleteDSL) Arp(ifaceName string, ipAddr string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &l3.ArpTable_ArpEntry{
		Interface: ifaceName,
		IpAddress: ipAddr,
	})
	return dsl
}

// ProxyArpInterfaces adds a request to delete an existing VPP L3 proxy ARP interfaces
func (dsl *DeleteDSL) ProxyArpInterfaces(label string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &l3.ProxyArpInterfaces_InterfaceList{
		Label: label,
	})
	return dsl
}

// ProxyArpRanges adds a request to delete an existing VPP L3 proxy ARP ranges
func (dsl *DeleteDSL) ProxyArpRanges(label string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &l3.ProxyArpRanges_RangeList{
		Label: label,
	})
	return dsl
}

// StnRule adds request to delete Stn rule.
func (dsl *DeleteDSL) StnRule(name string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &stn.STN_Rule{
		RuleName: name,
	})
	return dsl
}

// NAT44Global adds a request to remove global configuration for NAT44
func (dsl *DeleteDSL) NAT44Global() defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &nat.Nat44Global{})
	return dsl
}

// NAT44DNat adds a request to delete a DNAT configuration
func (dsl *DeleteDSL) NAT44DNat(label string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &nat.Nat44DNat_DNatConfig{
		Label: label,
	})
	return dsl
}

// IPSecSA adds request to delete a Security Association
func (dsl *DeleteDSL) IPSecSA(name string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &ipsec.SecurityAssociations_SA{
		Name: name,
	})
	return dsl
}

// IPSecSPD adds request to delete a Security Policy Database
func (dsl *DeleteDSL) IPSecSPD(name string) defaultplugins.DeleteDSL {
	dsl.parent.del = append(dsl.parent.del, &ipsec.SecurityPolicyDatabases_SPD{
		Name: name,
	})
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

	var (
		ifsPut                         []*interfaces.Interfaces_Interface
		bdsPut                         []*l2.BridgeDomains_BridgeDomain
		xCsPut                         []*l2.XConnectPairs_XConnectPair
		rtsPut                         []*l3.StaticRoutes_Route
		aclPut                         []*acl.AccessLists_Acl
		ifsDel, bdsDel, xCsDel, aclDel []string
		rtsDel                         []*rpc.DelStaticRoutesRequest_DelStaticRoute
	)

	// 'PUT'
	for _, val := range dsl.put {
		switch typed := val.(type) {
		case *interfaces.Interfaces_Interface:
			ifsPut = append(ifsPut, typed)
		case *l2.BridgeDomains_BridgeDomain:
			bdsPut = append(bdsPut, typed)
		case *l2.XConnectPairs_XConnectPair:
			xCsPut = append(xCsPut, typed)
		case *l3.StaticRoutes_Route:
			rtsPut = append(rtsPut, typed)
		case *acl.AccessLists_Acl:
			aclPut = append(aclPut, typed)
		default:
			logrus.DefaultLogger().Error("Unsupported data type: %v", val)
		}
	}

	// 'DEL'
	for _, val := range dsl.put {
		switch typed := val.(type) {
		case *interfaces.Interfaces_Interface:
			ifsDel = append(ifsDel, typed.Name)
		case *l2.BridgeDomains_BridgeDomain:
			bdsDel = append(bdsDel, typed.Name)
		case *l2.XConnectPairs_XConnectPair:
			xCsDel = append(xCsDel, typed.ReceiveInterface)
		case *l3.StaticRoutes_Route:
			rtsDel = append(rtsDel, &rpc.DelStaticRoutesRequest_DelStaticRoute{
				VRF: typed.VrfId, DstAddr: typed.DstIpAddr, NextHopAddr: typed.NextHopAddr,
			})
		case *acl.AccessLists_Acl:
			aclDel = append(aclDel, typed.AclName)
		default:
			logrus.DefaultLogger().Error("Unsupported data type: %v", val)
		}
	}

	ctx := context.Background()

	// Call 'DEL'
	if _, err := dsl.client.DelInterfaces(ctx, &rpc.DelNamesRequest{Name: ifsDel}); err != nil {
		wasErr = err
	}
	if _, err := dsl.client.DelBDs(ctx, &rpc.DelNamesRequest{Name: bdsDel}); err != nil {
		wasErr = err
	}
	if _, err := dsl.client.DelXCons(ctx, &rpc.DelNamesRequest{Name: xCsDel}); err != nil {
		wasErr = err
	}
	if _, err := dsl.client.DelStaticRoutes(ctx, &rpc.DelStaticRoutesRequest{Route: rtsDel}); err != nil {
		wasErr = err
	}
	if _, err := dsl.client.DelACLs(ctx, &rpc.DelNamesRequest{Name: aclDel}); err != nil {
		wasErr = err
	}

	// Call 'PUT'
	if _, err := dsl.client.PutInterfaces(ctx, &interfaces.Interfaces{Interfaces: ifsPut}); err != nil {
		wasErr = err
	}
	if _, err := dsl.client.PutBDs(ctx, &l2.BridgeDomains{BridgeDomains: bdsPut}); err != nil {
		wasErr = err
	}
	if _, err := dsl.client.PutXCons(ctx, &l2.XConnectPairs{XConnectPairs: xCsPut}); err != nil {
		wasErr = err
	}
	if _, err := dsl.client.PutStaticRoutes(ctx, &l3.StaticRoutes{Routes: rtsPut}); err != nil {
		wasErr = err
	}
	if _, err := dsl.client.PutACLs(ctx, &acl.AccessLists{Acl: aclPut}); err != nil {
		wasErr = err
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
