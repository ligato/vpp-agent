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

// Info
const (
	// Version is a path for retrieving information about version of Agent.
	Version = "/info/version"

	// JSONSchema is a path for retrieving JSON Schema for VPP-Agent configuration (dynamically created
	// container of all registered configuration models).
	JSONSchema = "/info/configuration/jsonschema"
)

// Configuration
const (
	// Configuration is a path for handling(GET,PUT) all VPP-Agent NB configuration
	Configuration = "/configuration"

	// Validate is a path for validating NB yaml configuration for VPP-Agent (the same all-in-one dynamically
	// created yaml configuration as used in agentctl configuration get/update)
	Validate = "/configuration/validate"
)

// Linux Dumps
const (
	// Interfaces

	// LinuxInterface is a linux interface rest path
	LinuxInterface = "/dump/linux/v2/interfaces"

	// L3

	// LinuxRoutes is the rest linux route path
	LinuxRoutes = "/dump/linux/v2/routes"
	// LinuxArps is the rest linux ARPs path
	LinuxArps = "/dump/linux/v2/arps"
)

// VPP ABF
const (
	// REST ABF
	ABF = "/dump/vpp/v2/abf"
)

// VPP ACL
const (
	// REST ACL IP prefix
	ACLIP = "/dump/vpp/v2/acl/ip"
	// REST ACL MACIP prefix
	ACLMACIP = "/dump/vpp/v2/acl/macip"
)

// VPP Interfaces
const (
	// Interface is rest interface path
	Interface = "/dump/vpp/v2/interfaces"

	// Loopback is path for loopback interface
	Loopback = "/dump/vpp/v2/interfaces/loopback"
	// Ethernet is path for physical interface
	Ethernet = "/dump/vpp/v2/interfaces/ethernet"
	// Memif is path for memif interface
	Memif = "/dump/vpp/v2/interfaces/memif"
	// Tap is path for tap interface
	Tap = "/dump/vpp/v2/interfaces/tap"
	// AfPacket is path for af-packet interface
	AfPacket = "/dump/vpp/v2/interfaces/afpacket"
	// VxLan is path for vxlan interface
	VxLan = "/dump/vpp/v2/interfaces/vxlan"
)

// VPP NAT
const (
	// NatGlobal is a REST path of a global NAT config
	NatGlobal = "/dump/vpp/v2/nat/global"
	// NatDNat is a REST path of a DNAT configurations
	NatDNat = "/dump/vpp/v2/nat/dnat"
	// NatInterfaces is a REST path of NAT interfaces config
	NatInterfaces = "/dump/vpp/v2/nat/interfaces"
	// NatAddressPools is a REST path of NAT address pools config
	NatAddressPools = "/dump/vpp/v2/nat/pools"
)

// L2 plugin
const (
	// restBd is rest bridge domain path
	Bd = "/dump/vpp/v2/bd"
	// restFib is rest FIB path
	Fib = "/dump/vpp/v2/fib"
	// restXc is rest cross-connect path
	Xc = "/dump/vpp/v2/xc"
)

// VPP L3 plugin
const (
	// Routes is rest static route path
	Routes = "/dump/vpp/v2/routes"
	// Arps is rest ARPs path
	Arps = "/dump/vpp/v2/arps"
	// PArpIfs is rest proxy ARP interfaces path
	PArpIfs = "/dump/vpp/v2/proxyarp/interfaces"
	// PArpRngs is rest proxy ARP ranges path
	PArpRngs = "/dump/vpp/v2/proxyarp/ranges"
	// IPScanNeigh is rest IP scan neighbor setup path
	IPScanNeigh = "/dump/vpp/v2/ipscanneigh"
	// Vrrps is rest vrrp entries path
	Vrrps = "/dump/vpp/v2/vrrps"
)

// VPP IPSec plugin
const (
	// SPDs is rest IPSec security policy database path
	SPDs = "/dump/vpp/v2/ipsec/spds"
	// SPs is rest IPSec security policy path
	SPs = "/dump/vpp/v2/ipsec/sps"
	// SAs is rest IPSec security association path
	SAs = "/dump/vpp/v2/ipsec/sas"
)

const (
	// PuntSocket is rest punt registered socket path
	PuntSocket = "/dump/vpp/v2/punt/sockets"
)

// VPP Wireguard plugin
const (
	Peers = "/dump/vpp/v2/wireguard/peers"
)

// Telemetry
const (
	// Telemetry reads various types of metrics data from the VPP
	Telemetry  = "/vpp/telemetry"
	TMemory    = "/vpp/telemetry/memory"
	TRuntime   = "/vpp/telemetry/runtime"
	TNodeCount = "/vpp/telemetry/nodecount"
)

// Stats
const (
	// Configurator stats
	ConfiguratorStats = "/stats/configurator"
	// Linux interface stats
	LinuxInterfaceStats = "/stats/linux/interfaces"
)
