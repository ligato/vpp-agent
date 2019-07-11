// Copyright (c) 2018 Cisco and/or its affiliates.
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

package linuxclient

import (
	linux_interfaces "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	linux_l3 "github.com/ligato/vpp-agent/api/models/linux/l3"
	vpp_abf "github.com/ligato/vpp-agent/api/models/vpp/abf"
	vpp_acl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	vpp_interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	vpp_l2 "github.com/ligato/vpp-agent/api/models/vpp/l2"
	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	nat "github.com/ligato/vpp-agent/api/models/vpp/nat"
	punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	vpp_stn "github.com/ligato/vpp-agent/api/models/vpp/stn"
	vpp_clientv2 "github.com/ligato/vpp-agent/clientv2/vpp"
)

// DataChangeDSL defines the Domain Specific Language (DSL) for data change
// of both Linux and VPP configuration.
// Use this interface to make your implementation independent of the local
// and any remote client.
// Every DSL statement (apart from Send) returns the receiver (possibly wrapped
// to change the scope of DSL), allowing the calls to be chained together
// conveniently in a single statement.
type DataChangeDSL interface {
	// Put initiates a chained sequence of data change DSL statements, declaring
	// new configurable objects or changing existing ones, e.g.:
	//     Put().LinuxInterface(&veth).VppInterface(&afpacket).BD(&BD) ... Send()
	// The set of available objects to be created or changed is defined by PutDSL.
	Put() PutDSL

	// Delete initiates a chained sequence of data change DSL statements,
	// removing existing configurable objects (by name), e.g:
	//     Delete().LinuxInterface(vethName).VppInterface(afpacketName).BD(BDName) ... Send()
	// The set of available objects to be removed is defined by DeleteDSL.
	Delete() DeleteDSL

	// Send propagates requested changes to the plugins.
	Send() vpp_clientv2.Reply
}

// PutDSL is a subset of data change DSL statements, used to declare new
// Linux or VPP configuration or change existing one.
type PutDSL interface {
	// LinuxInterface adds a request to create or update Linux network interface.
	LinuxInterface(val *linux_interfaces.Interface) PutDSL
	// LinuxArpEntry adds a request to crete or update Linux ARP entry
	LinuxArpEntry(val *linux_l3.ARPEntry) PutDSL
	// LinuxRoute adds a request to crete or update Linux route
	LinuxRoute(val *linux_l3.Route) PutDSL

	// VppInterface adds a request to create or update VPP network interface.
	VppInterface(val *vpp_interfaces.Interface) PutDSL
	// ACL adds a request to create or update VPP Access Control List.
	ACL(acl *vpp_acl.ACL) PutDSL
	// ABF adds a request to create or update VPP ACL-based forwarding.
	ABF(abf *vpp_abf.ABF) PutDSL
	/*// BfdSession adds a request to create or update VPP bidirectional
	// forwarding detection session.
	BfdSession(val *vpp_bfd.SingleHopBFD_Session) PutDSL
	// BfdAuthKeys adds a request to create or update VPP bidirectional
	// forwarding detection key.
	BfdAuthKeys(val *vpp_bfd.SingleHopBFD_Key) PutDSL
	// BfdEchoFunction adds a request to create or update VPP bidirectional
	// forwarding detection echo function.
	BfdEchoFunction(val *vpp_bfd.SingleHopBFD_EchoFunction) PutDSL*/
	// BD adds a request to create or update VPP Bridge Domain.
	BD(val *vpp_l2.BridgeDomain) PutDSL
	// BDFIB adds a request to create or update VPP L2 Forwarding Information Base.
	BDFIB(fib *vpp_l2.FIBEntry) PutDSL
	// VrfTable adds a request to create or update VPP VRF table.
	VrfTable(val *vpp_l3.VrfTable) PutDSL
	// XConnect adds a request to create or update VPP Cross Connect.
	XConnect(val *vpp_l2.XConnectPair) PutDSL
	// StaticRoute adds a request to create or update VPP L3 Static Route.
	StaticRoute(val *vpp_l3.Route) PutDSL
	// Arp adds a request to create or update VPP L3 ARP.
	Arp(arp *vpp_l3.ARPEntry) PutDSL
	// ProxyArp adds a request to create or update VPP L3 proxy ARP.
	ProxyArp(proxyArp *vpp_l3.ProxyARP) PutDSL
	// IPScanNeighbor adds L3 IP Scan Neighbor to the RESYNC request.
	IPScanNeighbor(ipScanNeigh *vpp_l3.IPScanNeighbor) PutDSL
	/*// L4Features adds a request to enable or disable L4 features
	L4Features(val *vpp_l4.L4Features) PutDSL
	// AppNamespace adds a request to create or update VPP Application namespace
	AppNamespace(appNs *vpp_l4.AppNamespaces_AppNamespace) PutDSL*/
	// StnRule adds a request to create or update VPP Stn rule.
	StnRule(stn *vpp_stn.Rule) PutDSL
	// NAT44Global adds a request to set global configuration for NAT44
	NAT44Global(nat *nat.Nat44Global) PutDSL
	// DNAT44 adds a request to create or update DNAT44 configuration
	DNAT44(dnat *nat.DNat44) PutDSL
	// IPSecSA adds request to create a new Security Association
	IPSecSA(sa *ipsec.SecurityAssociation) PutDSL
	// IPSecSPD adds request to create a new Security Policy Database
	IPSecSPD(spd *ipsec.SecurityPolicyDatabase) PutDSL
	// PuntIPRedirect adds request to create or update rule to punt L3 traffic via interface.
	PuntIPRedirect(val *punt.IPRedirect) PutDSL
	// PuntToHost adds request to create or update rule to punt L4 traffic to a host.
	PuntToHost(val *punt.ToHost) PutDSL
	// PuntException adds request to create or update exception to punt specific packets.
	PuntException(val *punt.Exception) PutDSL

	// Delete changes the DSL mode to allow removing an existing configuration.
	// See documentation for DataChangeDSL.Delete().
	Delete() DeleteDSL

	// Send propagates requested changes to the plugins.
	Send() vpp_clientv2.Reply
}

// DeleteDSL is a subset of data change DSL statements, used to remove
// existing Linux or VPP configuration.
type DeleteDSL interface {
	// LinuxInterface adds a request to delete an existing Linux network
	// interface.
	LinuxInterface(ifaceName string) DeleteDSL
	// LinuxArpEntry adds a request to delete Linux ARP entry
	LinuxArpEntry(ifaceName string, ipAddr string) DeleteDSL
	// LinuxRoute adds a request to delete Linux route
	LinuxRoute(dstAddr, outIfaceName string) DeleteDSL

	// VppInterface adds a request to delete an existing VPP network interface.
	VppInterface(ifaceName string) DeleteDSL
	// ACL adds a request to delete an existing VPP Access Control List.
	ACL(aclName string) DeleteDSL
	// ABF adds a request to delete an existing VPP ACL-based forwarding.
	ABF(abfIndex uint32) DeleteDSL
	/*// BfdSession adds a request to delete an existing VPP bidirectional
	// forwarding detection session.
	BfdSession(bfdSessionIfaceName string) DeleteDSL
	// BfdAuthKeys adds a request to delete an existing VPP bidirectional
	// forwarding detection key.
	BfdAuthKeys(bfdKey string) DeleteDSL
	// BfdEchoFunction adds a request to delete an existing VPP bidirectional
	// forwarding detection echo function.
	BfdEchoFunction(bfdEchoName string) DeleteDSL*/
	// BD adds a request to delete an existing VPP Bridge Domain.
	BD(bdName string) DeleteDSL
	// FIB adds a request to delete an existing VPP L2 Forwarding Information
	// Base.
	BDFIB(bdName string, mac string) DeleteDSL
	// VrfTable adds a request to delete existing VPP VRF table.
	VrfTable(id uint32, proto vpp_l3.VrfTable_Protocol) DeleteDSL
	// XConnect adds a request to delete an existing VPP Cross Connect.
	XConnect(rxIfaceName string) DeleteDSL
	// StaticRoute adds a request to delete an existing VPP L3 Static Route.
	StaticRoute(iface string, vrf uint32, dstAddr string, nextHopAddr string) DeleteDSL
	/*// L4Features adds a request to enable or disable L4 features
	L4Features() DeleteDSL
	// AppNamespace adds a request to delete VPP Application namespace
	// Note: current version does not support application namespace deletion
	AppNamespace(id string) DeleteDSL*/
	// Arp adds a request to delete an existing VPP L3 ARP.
	Arp(ifaceName string, ipAddr string) DeleteDSL
	// ProxyArp adds a request to delete an existing VPP L3 proxy ARP
	ProxyArp() DeleteDSL
	// IPScanNeighbor adds a request to delete an existing VPP L3 IP Scan Neighbor.
	IPScanNeighbor() DeleteDSL
	// StnRule adds a request to delete an existing VPP Stn rule.
	StnRule(iface, addr string) DeleteDSL
	// NAT44Global adds a request to remove global configuration for NAT44
	NAT44Global() DeleteDSL
	// DNAT44 adds a request to delete an existing DNAT-44 configuration
	DNAT44(label string) DeleteDSL
	// IPSecSA adds request to delete a Security Association
	IPSecSA(saIndex string) DeleteDSL
	// IPSecSPD adds request to delete a Security Policy Database
	IPSecSPD(spdIndex string) DeleteDSL
	// PuntIPRedirect adds request to delete a rule used to punt L3 traffic via interface.
	PuntIPRedirect(l3Proto punt.L3Protocol, txInterface string) DeleteDSL
	// PuntToHost adds request to delete a rule used to punt L4 traffic to a host.
	PuntToHost(l3Proto punt.L3Protocol, l4Proto punt.L4Protocol, port uint32) DeleteDSL
	// PuntException adds request to delete exception to punt specific packets.
	PuntException(reason string) DeleteDSL

	// Put changes the DSL mode to allow configuration editing.
	// See documentation for DataChangeDSL.Put().
	Put() PutDSL

	// Send propagates requested changes to the plugins.
	Send() vpp_clientv2.Reply
}
