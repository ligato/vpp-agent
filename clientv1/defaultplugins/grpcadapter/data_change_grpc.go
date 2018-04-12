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

	"github.com/gogo/protobuf/proto"
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
	return &DataChangeDSL{client, make(map[string]proto.Message, 0), make(map[string]proto.Message)}
}

// DataChangeDSL is used to conveniently assign all the data that are needed for the DataChange.
// This is an implementation of Domain Specific Language (DSL) for a change of the VPP configuration.
type DataChangeDSL struct {
	client rpc.ChangeConfigServiceClient
	put    map[string]proto.Message
	del    map[string]proto.Message
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
	dsl.parent.put[val.Name] = val
	return dsl
}

// BfdSession creates or updates the bidirectional forwarding detection session.
func (dsl *PutDSL) BfdSession(val *bfd.SingleHopBFD_Session) defaultplugins.PutDSL {
	dsl.parent.put[val.Interface] = val
	return dsl
}

// BfdAuthKeys creates or updates the bidirectional forwarding detection key.
func (dsl *PutDSL) BfdAuthKeys(val *bfd.SingleHopBFD_Key) defaultplugins.PutDSL {
	dsl.parent.put[strconv.Itoa(int(val.Id))] = val
	return dsl
}

// BfdEchoFunction creates or updates the bidirectional forwarding detection echo function.
func (dsl *PutDSL) BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) defaultplugins.PutDSL {
	dsl.parent.put[val.EchoSourceInterface] = val
	return dsl
}

// BD creates or updates the Bridge Domain.
func (dsl *PutDSL) BD(val *l2.BridgeDomains_BridgeDomain) defaultplugins.PutDSL {
	dsl.parent.put[val.Name] = val

	return dsl
}

// BDFIB deletes request for the L2 Forwarding Information Base.
func (dsl *PutDSL) BDFIB(val *l2.FibTableEntries_FibTableEntry) defaultplugins.PutDSL {
	dsl.parent.put[l2.FibKey(val.BridgeDomain, val.PhysAddress)] = val

	return dsl
}

// XConnect creates or updates the Cross Connect.
func (dsl *PutDSL) XConnect(val *l2.XConnectPairs_XConnectPair) defaultplugins.PutDSL {
	dsl.parent.put[val.ReceiveInterface] = val

	return dsl
}

// StaticRoute creates or updates the L3 Static Route.
func (dsl *PutDSL) StaticRoute(val *l3.StaticRoutes_Route) defaultplugins.PutDSL {
	dsl.parent.put[l3.RouteKey(val.VrfId, val.DstIpAddr, val.NextHopAddr)] = val

	return dsl
}

// ACL creates or updates request for the Access Control List.
func (dsl *PutDSL) ACL(val *acl.AccessLists_Acl) defaultplugins.PutDSL {
	dsl.parent.put[val.AclName] = val

	return dsl
}

// L4Features creates or updates the request for the L4Features.
func (dsl *PutDSL) L4Features(val *l4.L4Features) defaultplugins.PutDSL {
	dsl.parent.put[strconv.FormatBool(val.Enabled)] = val

	return dsl
}

// AppNamespace creates or updates the request for the Application Namespaces List.
func (dsl *PutDSL) AppNamespace(val *l4.AppNamespaces_AppNamespace) defaultplugins.PutDSL {
	dsl.parent.put[val.NamespaceId] = val

	return dsl
}

// Arp adds a request to create or update VPP L3 ARP entry.
func (dsl *PutDSL) Arp(val *l3.ArpTable_ArpTableEntry) defaultplugins.PutDSL {
	dsl.parent.put[l3.ArpEntryKey(val.Interface, val.IpAddress)] = val
	return dsl
}

// ProxyArpInterfaces adds a request to create or update VPP L3 proxy ARP interfaces.
func (dsl *PutDSL) ProxyArpInterfaces(val *l3.ProxyArpInterfaces_InterfaceList) defaultplugins.PutDSL {
	dsl.parent.put[val.Label] = val
	return dsl
}

// ProxyArpRanges adds a request to create or update VPP L3 proxy ARP ranges
func (dsl *PutDSL) ProxyArpRanges(val *l3.ProxyArpRanges_RangeList) defaultplugins.PutDSL {
	dsl.parent.put[val.Label] = val
	return dsl
}

// StnRule adds a request to create or update STN rule.
func (dsl *PutDSL) StnRule(val *stn.StnRule) defaultplugins.PutDSL {
	dsl.parent.put[val.RuleName] = val
	return dsl
}

// NAT44Global adds a request to set global configuration for NAT44
func (dsl *PutDSL) NAT44Global(val *nat.Nat44Global) defaultplugins.PutDSL {
	dsl.parent.put[nat.GlobalConfigKey()] = val
	return dsl
}

// NAT44DNat adds a request to create a new DNAT configuration
func (dsl *PutDSL) NAT44DNat(val *nat.Nat44DNat_DNatConfig) defaultplugins.PutDSL {
	dsl.parent.put[nat.DNatKey(val.Label)] = val
	return dsl
}

// IPSecSA adds request to create a new Security Association
func (dsl *PutDSL) IPSecSA(val *ipsec.SecurityAssociations_SA) defaultplugins.PutDSL {
	dsl.parent.put[ipsec.SAKey(val.Name)] = val
	return dsl
}

// IPSecSPD adds request to create a new Security Policy Database
func (dsl *PutDSL) IPSecSPD(val *ipsec.SecurityPolicyDatabases_SPD) defaultplugins.PutDSL {
	dsl.parent.put[ipsec.SPDKey(val.Name)] = val
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
	dsl.parent.del[interfaceName] = &interfaces.Interfaces_Interface{}

	return dsl
}

// BfdSession adds a request to delete an existing bidirectional forwarding
// detection session.
func (dsl *DeleteDSL) BfdSession(bfdSessionIfaceName string) defaultplugins.DeleteDSL {
	dsl.parent.del[bfdSessionIfaceName] = &bfd.SingleHopBFD_Session{}
	return dsl
}

// BfdAuthKeys adds a request to delete an existing bidirectional forwarding
// detection key.
func (dsl *DeleteDSL) BfdAuthKeys(bfdKeyID string) defaultplugins.DeleteDSL {
	dsl.parent.del[bfdKeyID] = &bfd.SingleHopBFD_Key{}
	return dsl
}

// BfdEchoFunction adds a request to delete an existing bidirectional forwarding
// detection echo function.
func (dsl *DeleteDSL) BfdEchoFunction(bfdEchoName string) defaultplugins.DeleteDSL {
	dsl.parent.del[bfdEchoName] = &bfd.SingleHopBFD_EchoFunction{}
	return dsl
}

// BD deletes request for the Bridge Domain.
func (dsl *DeleteDSL) BD(bdName string) defaultplugins.DeleteDSL {
	dsl.parent.del[bdName] = &l2.BridgeDomains_BridgeDomain{}
	return dsl
}

// BDFIB deletes request for the L2 Forwarding Information Base.
func (dsl *DeleteDSL) BDFIB(bdName string, mac string) defaultplugins.DeleteDSL {
	dsl.parent.del[l2.FibKey(bdName, mac)] = &l2.FibTableEntries_FibTableEntry{}
	return dsl
}

// XConnect deletes the Cross Connect.
func (dsl *DeleteDSL) XConnect(rxIfName string) defaultplugins.DeleteDSL {
	dsl.parent.del[rxIfName] = &l2.XConnectPairs_XConnectPair{}
	return dsl
}

// StaticRoute deletes the L3 Static Route.
func (dsl *DeleteDSL) StaticRoute(vrf uint32, dstAddr string, nextHopAddr string) defaultplugins.DeleteDSL {
	dsl.parent.del[l3.RouteKey(vrf, dstAddr, nextHopAddr)] = &rpc.DelStaticRouteRequest{
		VRF: vrf, DstAddr: dstAddr, NextHopAddr: nextHopAddr,
	}
	return dsl
}

// ACL deletes request for Access Control List.
func (dsl *DeleteDSL) ACL(aclName string) defaultplugins.DeleteDSL {
	dsl.parent.del[aclName] = &acl.AccessLists_Acl{}

	return dsl
}

// L4Features deletes request for the L4Features.
func (dsl *DeleteDSL) L4Features() defaultplugins.DeleteDSL {
	dsl.parent.del["l4"] = &l4.L4Features{}

	return dsl
}

// AppNamespace delets request for the Application Namespaces List.
func (dsl *DeleteDSL) AppNamespace(id string) defaultplugins.DeleteDSL {
	dsl.parent.del[id] = &l4.AppNamespaces_AppNamespace{}

	return dsl
}

// Arp adds a request to delete an existing VPP L3 ARP entry.
func (dsl *DeleteDSL) Arp(ifaceName string, ipAddr string) defaultplugins.DeleteDSL {
	dsl.parent.del[l3.ArpEntryKey(ifaceName, ipAddr)] = &l3.ArpTable_ArpTableEntry{}
	return dsl
}

// ProxyArpInterfaces adds a request to delete an existing VPP L3 proxy ARP interfaces
func (dsl *DeleteDSL) ProxyArpInterfaces(label string) defaultplugins.DeleteDSL {
	dsl.parent.del[label] = &l3.ProxyArpInterfaces_InterfaceList{}
	return dsl
}

// ProxyArpRanges adds a request to delete an existing VPP L3 proxy ARP ranges
func (dsl *DeleteDSL) ProxyArpRanges(label string) defaultplugins.DeleteDSL {
	dsl.parent.del[label] = &l3.ProxyArpRanges_RangeList{}
	return dsl
}

// StnRule adds request to delete Stn rule.
func (dsl *DeleteDSL) StnRule(name string) defaultplugins.DeleteDSL {
	dsl.parent.del[name] = &stn.StnRule{}
	return dsl
}

// NAT44Global adds a request to remove global configuration for NAT44
func (dsl *DeleteDSL) NAT44Global() defaultplugins.DeleteDSL {
	dsl.parent.del[nat.GlobalConfigKey()] = &nat.Nat44Global{}
	return dsl
}

// NAT44DNat adds a request to delete a DNAT configuration
func (dsl *DeleteDSL) NAT44DNat(label string) defaultplugins.DeleteDSL {
	dsl.parent.del[nat.DNatKey(label)] = &nat.Nat44DNat_DNatConfig{}
	return dsl
}

// IPSecSA adds request to delete a Security Association
func (dsl *DeleteDSL) IPSecSA(name string) defaultplugins.DeleteDSL {
	dsl.parent.del[ipsec.SAKey(name)] = &ipsec.SecurityAssociations_SA{}
	return dsl
}

// IPSecSPD adds request to delete a Security Policy Database
func (dsl *DeleteDSL) IPSecSPD(name string) defaultplugins.DeleteDSL {
	dsl.parent.del[ipsec.SPDKey(name)] = &ipsec.SecurityPolicyDatabases_SPD{}
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

	for _, val := range dsl.put {
		switch typed := val.(type) {
		case *interfaces.Interfaces_Interface:
			_, err := dsl.client.PutInterface(context.Background(), typed)
			if err != nil {
				wasErr = err
			}
		case *l2.BridgeDomains_BridgeDomain:
			_, err := dsl.client.PutBD(context.Background(), typed)
			if err != nil {
				wasErr = err
			}
		case *l2.XConnectPairs_XConnectPair:
			_, err := dsl.client.PutXCon(context.Background(), typed)
			if err != nil {
				wasErr = err
			}
		case *l3.StaticRoutes_Route:
			_, err := dsl.client.PutStaticRoute(context.Background(), typed)
			if err != nil {
				wasErr = err
			}
		case *acl.AccessLists_Acl:
			_, err := dsl.client.PutACL(context.Background(), typed)
			if err != nil {
				wasErr = err
			}
		}
	}

	for key, val := range dsl.del {
		switch typed := val.(type) {
		case *interfaces.Interfaces_Interface:
			_, err := dsl.client.DelInterface(context.Background(), &rpc.DelNameRequest{Name: key})
			if err != nil {
				wasErr = err
			}
		case *l2.BridgeDomains_BridgeDomain:
			_, err := dsl.client.DelBD(context.Background(), &rpc.DelNameRequest{Name: key})
			if err != nil {
				wasErr = err
			}
		case *l2.XConnectPairs_XConnectPair:
			_, err := dsl.client.DelXCon(context.Background(), &rpc.DelNameRequest{Name: key})
			if err != nil {
				wasErr = err
			}
		case *rpc.DelStaticRouteRequest:
			_, err := dsl.client.DelStaticRoute(context.Background(), typed)
			if err != nil {
				wasErr = err
			}
		case *acl.AccessLists_Acl:
			_, err := dsl.client.DelACL(context.Background(), &rpc.DelNameRequest{Name: key})
			if err != nil {
				wasErr = err
			}
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
