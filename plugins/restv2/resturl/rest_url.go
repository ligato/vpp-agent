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

package resturl

// ACL
const (
	// REST ACL IP prefix
	ACLIP = "/vpp/dump/v2/acl/ip"
	// REST ACL MACIP prefix
	ACLMACIP = "/vpp/dump/v2/acl/macip"
)

// VPP interface
const (
	// Interface is rest interface path
	Interface = "/vpp/dump/v2/interfaces"
	// Loopback is path for loopback interface
	Loopback = "/vpp/dump/v2/interfaces/loopback"
	// Ethernet is path for physical interface
	Ethernet = "/vpp/dump/v2/interfaces/ethernet"
	// Memif is path for memif interface
	Memif = "/vpp/dump/v2/interfaces/memif"
	// Tap is path for tap interface
	Tap = "/vpp/dump/v2/interfaces/tap"
	// AfPacket is path for af-packet interface
	AfPacket = "/vpp/dump/v2/interfaces/afpacket"
	// VxLan is path for vxlan interface
	VxLan = "/vpp/dump/v2/interfaces/vxlan"
)

// Linux interface REST urls
const (
	// LinuxInterface is a linux interface rest path
	LinuxInterface = "/linux/dump/v2/interfaces"
	// LinuxInterfaceStats is a linux interface statistics rest path
	LinuxInterfaceStats = "/linux/dump/v2/interfaces/statistics"
)

// NAT REST urls
const (
	// NatURL is a REST path of a NAT
	NatURL = "/vpp/dump/v2/nat"
	// NatGlobal is a REST path of a global NAT config
	NatGlobal = "/vpp/dump/v2/nat/global"
	// NatDNat is a REST path of a DNAT configurations
	NatDNat = "/vpp/dump/v2/nat/dnat"
)

// L2 plugin
const (
	// restBd is rest bridge domain path
	Bd = "/vpp/dump/v2/bd"
	// restBdId is rest bridge domain ID path
	BdID = "/vpp/dump/v2/bdid"
	// restFib is rest FIB path
	Fib = "/vpp/dump/v2/fib"
	// restXc is rest cross-connect path
	Xc = "/vpp/dump/v2/xc"
)

// VPP L3 plugin
const (
	// Routes is rest static route path
	Routes = "/vpp/dump/v2/routes"
	// Arps is rest ARPs path
	Arps = "/vpp/dump/v2/arps"
	// PArpIfs is rest proxy ARP interfaces path
	PArpIfs = "/vpp/dump/v2/proxyarp/interfaces"
	// PArpRngs is rest proxy ARP ranges path
	PArpRngs = "/vpp/dump/v2/proxyarp/ranges"
)

// Linux L3 plugin
const (
	// LinuxRoutes is the rest linux route path
	LinuxRoutes = "/linux/dump/v2/routes"
	// LinuxArps is the rest linux ARPs path
	LinuxArps = "/linux/dump/v2/arps"
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

// Tracer
const (
	// Traced binary API calls
	Tracer = "/vpp/binapitrace"
)

// Index
const (
	// Index can be used to get the full index page
	Index = "/"
)
