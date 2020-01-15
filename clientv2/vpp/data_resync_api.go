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

package vppclient

import (
	abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
	acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
	punt "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/punt"
	stn "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/stn"
)

// DataResyncDSL defines the Domain Specific Language (DSL) for data RESYNC
// of the VPP configuration.
// Use this interface to make your implementation independent of the local
// and any remote client.
// Each method (apart from Send) returns the receiver, allowing the calls
// to be chained together conveniently in a single statement.
type DataResyncDSL interface {
	// Interface adds interface to the RESYNC request.
	Interface(intf *interfaces.Interface) DataResyncDSL
	// Span adds span to the RESYNC request.
	Span(intf *interfaces.Span) DataResyncDSL
	// ACL adds Access Control List to the RESYNC request.
	ACL(acl *acl.ACL) DataResyncDSL
	// ABF adds ACL-based forwarding to the RESYNC request.
	ABF(abf *abf.ABF) DataResyncDSL
	// BD adds Bridge Domain to the RESYNC request.
	BD(bd *l2.BridgeDomain) DataResyncDSL
	// BDFIB adds L2 Forwarding Information Base.
	BDFIB(fib *l2.FIBEntry) DataResyncDSL
	// XConnect adds Cross Connect to the RESYNC request.
	XConnect(xcon *l2.XConnectPair) DataResyncDSL
	// VrfTable adds VRF table to the RESYNC request.
	VrfTable(val *l3.VrfTable) DataResyncDSL
	// StaticRoute adds L3 Static Route to the RESYNC request.
	StaticRoute(staticRoute *l3.Route) DataResyncDSL
	// Arp adds VPP L3 ARP to the RESYNC request.
	Arp(arp *l3.ARPEntry) DataResyncDSL
	// ProxyArp adds L3 proxy ARP interfaces to the RESYNC request.
	ProxyArp(proxyArp *l3.ProxyARP) DataResyncDSL
	// IPScanNeighbor adds L3 IP Scan Neighbor to the RESYNC request.
	IPScanNeighbor(ipScanNeigh *l3.IPScanNeighbor) DataResyncDSL
	// StnRule adds Stn rule to the RESYNC request.
	StnRule(stn *stn.Rule) DataResyncDSL
	// NAT44Global adds global NAT44 configuration to the RESYNC request.
	NAT44Global(nat *nat.Nat44Global) DataResyncDSL
	// DNAT44 adds DNAT44 configuration to the RESYNC request
	DNAT44(dnat *nat.DNat44) DataResyncDSL
	// NAT44Interface adds NAT44 interface configuration to the RESYNC request.
	NAT44Interface(natIf *nat.Nat44Interface) DataResyncDSL
	// NAT44AddressPool adds NAT44 address pool configuration to the RESYNC request.
	NAT44AddressPool(pool *nat.Nat44AddressPool) DataResyncDSL
	// IPSecSA adds request to RESYNC a new Security Association
	IPSecSA(sa *ipsec.SecurityAssociation) DataResyncDSL
	// IPSecSPD adds request to RESYNC a new Security Policy Database
	IPSecSPD(spd *ipsec.SecurityPolicyDatabase) DataResyncDSL
	// PuntIPRedirect adds request to RESYNC a rule used to punt L3 traffic via interface.
	PuntIPRedirect(val *punt.IPRedirect) DataResyncDSL
	// PuntToHost adds request to RESYNC a rule used to punt L4 traffic to a host.
	PuntToHost(val *punt.ToHost) DataResyncDSL
	// PuntException adds request to create or update exception to punt specific packets.
	PuntException(val *punt.Exception) DataResyncDSL

	// Send propagates the RESYNC request to the plugins.
	Send() Reply
}
