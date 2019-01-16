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
	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/api/models/vpp/acl"
	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	"github.com/ligato/vpp-agent/api/models/vpp/l2"
	"github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/api/models/vpp/nat"
	"github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-agent/api/models/vpp/stn"
)

type (
	// ACL
	ACL = vpp_acl.Acl

	// Interfaces
	Interface = vpp_interfaces.Interface

	// L2
	BridgeDomain = vpp_l2.BridgeDomain
	L2FIB        = vpp_l2.FIBEntry
	XConnect     = vpp_l2.XConnectPair

	// L3
	Route       = vpp_l3.StaticRoute
	ARPEntry    = vpp_l3.ARPEntry
	IPScanNeigh = vpp_l3.IPScanNeighbor
	ProxyARP    = vpp_l3.ProxyARP

	// IPSec
	IPSecSPD = vpp_ipsec.SecurityPolicyDatabase
	IPSecSA  = vpp_ipsec.SecurityAssociation

	// NAT
	NAT44Global = vpp_nat.Nat44Global
	DNAT44      = vpp_nat.DNat44

	// STN
	STNRule = vpp_stn.Rule

	// Punt
	PuntIP     = vpp_punt.IpRedirect
	PuntToHost = vpp_punt.ToHost
)

var (
	ACLModel = models.Model(&ACL{})

	InterfaceModel = models.Model(&Interface{})

	BridgeDomainModel = models.Model(&BridgeDomain{})
	FIBEntryModel     = models.Model(&L2FIB{})
	XConnectPairModel = models.Model(&XConnect{})

	RouteModel       = models.Model(&Route{})
	ARPEntryModel    = models.Model(&ARPEntry{})
	IPScanNeighModel = models.Model(&IPScanNeigh{})
	ProxyARPModel    = models.Model(&ProxyARP{})

	IPSecSPDModel = models.Model(&IPSecSPD{})
	IPSecSAModel  = models.Model(&IPSecSA{})

	NAT44GlobalModel = models.Model(&NAT44Global{})
	DNAT44Model      = models.Model(&DNAT44{})

	STNRuleModel = models.Model(&STNRule{})

	// Punt
	PuntIPModel     = models.Model(&PuntIP{})
	PuntToHostModel = models.Model(&PuntToHost{})
)
