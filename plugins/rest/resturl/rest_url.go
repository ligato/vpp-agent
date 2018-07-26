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

// Access list REST keys
const (
	// REST Acl IP prefix
	AclIP = "/vpp/v1/acl/ip"
	// REST Acl IP example prefix
	AclIPExample = "/vpp/v1/acl/ip/example"
	// REST Acl MACIP prefix
	AclMACIP = "/vpp/v1/acl/macip"
	// REST Acl MACIP example prefix
	AclMACIPExample = "/vpp/v1/acl/macip/example"
)

// BFD REST keys
const (
	// restBfdKey is a REST path of a bfd
	BfdKey = "/vpp/v1/bfd"
	// restBfdSessionKey is a REST path of a bfd sessions
	BfdSessionKey = "/vpp/v1/bfd/sessions"
	// restBfdAuthKey is a REST path of a bfd authentication keys
	BfdAuthKey = "/vpp/v1/bfd/authkeys"
)

// Interface REST keys
const (
	// restInterface is rest interface path
	Interface = "/vpp/v1/interfaces"
	// restLoopback is path for loopback interface
	Loopback = "/vpp/v1/interfaces/loopback"
	// restLoopback is path for physical interface
	Ethernet = "/vpp/v1/interfaces/ethernet"
	// restLoopback is path for memif interface
	Memif = "/vpp/v1/interfaces/memif"
	// restLoopback is path for tap interface
	Tap = "/vpp/v1/interfaces/tap"
	// restAfPacket is path for af-packet interface
	AfPacket = "/vpp/v1/interfaces/afpacket"
	// restLoopback is path for vxlan interface
	VxLan = "/vpp/v1/interfaces/vxlan"
)

// L2 plugin
const (
	// restBd is rest bridge domain path
	Bd = "/vpp/v1/bd"
	// restBdId is rest bridge domain ID path
	BdId = "/vpp/v1/bdid"
	// restFib is rest FIB path
	Fib = "/vpp/v1/fib"
	// restXc is rest cross-connect path
	Xc = "/vpp/v1/xc"
)
