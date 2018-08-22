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

package resturl

// Access list REST urls
const (
	// REST ACL IP prefix
	ACLIP = "/vpp/dump/v1/acl/ip"
	// REST ACL MACIP prefix
	ACLMACIP = "/vpp/dump/v1/acl/macip"
)

// BFD REST urls
const (
	// BfdURL is a REST path of a bfd
	BfdURL = "/vpp/dump/v1/bfd"
	// BfdSession is a REST path of a bfd sessions
	BfdSession = "/vpp/dump/v1/bfd/sessions"
	// BfdAuthKey is a REST path of a bfd authentication keys
	BfdAuthKey = "/vpp/dump/v1/bfd/authkeys"
)

// Interface REST urls
const (
	// restInterface is rest interface path
	Interface = "/vpp/dump/v1/interfaces"
	// restLoopback is path for loopback interface
	Loopback = "/vpp/dump/v1/interfaces/loopback"
	// restLoopback is path for physical interface
	Ethernet = "/vpp/dump/v1/interfaces/ethernet"
	// restLoopback is path for memif interface
	Memif = "/vpp/dump/v1/interfaces/memif"
	// restLoopback is path for tap interface
	Tap = "/vpp/dump/v1/interfaces/tap"
	// restAfPacket is path for af-packet interface
	AfPacket = "/vpp/dump/v1/interfaces/afpacket"
	// restLoopback is path for vxlan interface
	VxLan = "/vpp/dump/v1/interfaces/vxlan"
)

// NAT REST urls
const (
	// NatURL is a REST path of a NAT
	NatURL = "/vpp/dump/v1/nat"
	// NatGlobal is a REST path of a global NAT config
	NatGlobal = "/vpp/dump/v1/nat/global"
	// NatDNat is a REST path of a DNAT configurations
	NatDNat = "/vpp/dump/v1/nat/dnat"
)

// STN REST url
const (
	// StnURL is a REST path of a STN
	StnURL = "/vpp/dump/v1/stn"
)

// IPSec REST url
const (
	// IPSecSpd is a REST path of a IPSec SPD
	IPSecSpd = "/vpp/dump/v1/ipsec/spd"
	// IPSecSa is a REST path of a IPSec SA
	IPSecSa = "/vpp/dump/v1/ipsec/sa"
	// IPSecTnIf is a REST path of a IPSec Tunnels
	IPSecTnIf = "/vpp/dump/v1/ipsec/tunnel"
)

// L2 plugin
const (
	// restBd is rest bridge domain path
	Bd = "/vpp/dump/v1/bd"
	// restBdId is rest bridge domain ID path
	BdID = "/vpp/dump/v1/bdid"
	// restFib is rest FIB path
	Fib = "/vpp/dump/v1/fib"
	// restXc is rest cross-connect path
	Xc = "/vpp/dump/v1/xc"
)

// L3 plugin
const (
	// Routes is rest static route path
	Routes = "/vpp/dump/v1/routes"
	// Arps is rest ARPs path
	Arps = "/vpp/dump/v1/arps"
	// PArpIfs is rest proxy ARP interfaces path
	PArpIfs = "/vpp/dump/v1/proxyarp/interfaces"
	// PArpRngs is rest proxy ARP ranges path
	PArpRngs = "/vpp/dump/v1/proxyarp/ranges"
)

// L4 plugin
const (
	// Sessions is rest L4 configuration path
	Sessions = "/vpp/dump/v1/sessions"
)

// Command
const (
	// Command allows to put CLI command to the rest
	Command = "/vpp/command"
)

// Telemetry
const (
	// Telemetry reads various types of metrics data from the VPP
	Telemetry  = "/vpp/telemetry"
	TMemory    = "/vpp/telemetry/memory"
	TRuntime   = "/vpp/telemetry/runtime"
	TNodeCount = "/vpp/telemetry/nodecount"
)

// Index
const (
	// Index can be used to get the full index page
	Index = "/"
)
