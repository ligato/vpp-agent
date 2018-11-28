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
	"github.com/ligato/vpp-agent/api/models/vpp/stn"
)

var (
	// ACL
	ACL = models.MustSpec(&vpp_acl.Acl{})

	// Interfaces
	Interface = models.MustSpec(&vpp_interfaces.Interface{})

	// L2
	BridgeDomain = models.MustSpec(&vpp_l2.BridgeDomain{})
	L2FIB        = models.MustSpec(&vpp_l2.FIBEntry{})
	XConnect     = models.MustSpec(&vpp_l2.XConnectPair{})

	// L3
	L3Route     = models.MustSpec(&vpp_l3.StaticRoute{})
	L3ARP       = models.MustSpec(&vpp_l3.ARPEntry{})
	IPScanNeigh = models.MustSpec(&vpp_l3.IPScanNeighbor{})
	ProxyARP    = models.MustSpec(&vpp_l3.ProxyARP{})

	// IPSec
	IPSecSPD = models.MustSpec(&vpp_ipsec.SecurityPolicyDatabase{})
	IPSecSA  = models.MustSpec(&vpp_ipsec.SecurityAssociation{})

	// NAT
	NAT44Global = models.MustSpec(&vpp_nat.Nat44Global{})
	DNAT44      = models.MustSpec(&vpp_nat.DNat44{})

	// STN
	STNRule = models.MustSpec(&vpp_stn.Rule{})
)
