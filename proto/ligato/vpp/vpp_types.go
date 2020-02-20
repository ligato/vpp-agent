//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp

import (
	vpp_abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
	vpp_acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
	vpp_l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	vpp_nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
	vpp_punt "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/punt"
	vpp_srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"
)

type (
	// Interface
	Interface = vpp_interfaces.Interface
	Span = vpp_interfaces.Span

	// ACL & ABF
	ACL = vpp_acl.ACL
	ABF = vpp_abf.ABF

	// L2
	BridgeDomain = vpp_l2.BridgeDomain
	L2FIB        = vpp_l2.FIBEntry
	XConnect     = vpp_l2.XConnectPair

	// L3
	Route       = vpp_l3.Route
	ARPEntry    = vpp_l3.ARPEntry
	ProxyARP    = vpp_l3.ProxyARP
	IPScanNeigh = vpp_l3.IPScanNeighbor
	VRFTable    = vpp_l3.VrfTable
	L3XConnect    = vpp_l3.L3XConnect
	DHCPProxy    = vpp_l3.DHCPProxy

	// NAT
	NAT44Global = vpp_nat.Nat44Global
	DNAT44      = vpp_nat.DNat44
	Nat44AddressPool      = vpp_nat.Nat44AddressPool
	Nat44Interface      = vpp_nat.Nat44Interface

	// IPSec
	IPSecSPD = vpp_ipsec.SecurityPolicyDatabase
	IPSecSA  = vpp_ipsec.SecurityAssociation

	// Punt
	PuntIPRedirect = vpp_punt.IPRedirect
	PuntToHost     = vpp_punt.ToHost
	PuntException     = vpp_punt.Exception

	// SRv6
	SRv6Global = vpp_srv6.SRv6Global
	SRv6LocalSID = vpp_srv6.LocalSID
	SRv6Policy = vpp_srv6.Policy
	SRv6Steering = vpp_srv6.Steering
)
