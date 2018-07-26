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

package url

// Access list REST keys
const (
	// REST Acl IP prefix
	restAclIP = "/v1/acl/ip"
	// REST Acl IP example prefix
	restAclIPExample = "/v1/acl/ip/example"
	// REST Acl MACIP prefix
	restAclMACIP = "/v1/acl/macip"
	// REST Acl MACIP example prefix
	restAclMACIPExample = "/v1/acl/macip/example"
)

// RestIPKey returns prefix used in REST to dump ACL IP config
func RestIPKey() string {
	return restAclIP
}

// RestIPExampleKey returns prefix used in REST to dump ACL IP example config
func RestIPExampleKey() string {
	return restAclIPExample
}

// RestMACIPKey returns prefix used in REST to dump ACL MACIP config
func RestMACIPKey() string {
	return restAclMACIP
}

// RestMACIPExampleKey returns prefix used in REST to dump ACL MACIP example config
func RestMACIPExampleKey() string {
	return restAclMACIPExample
}

// BFD REST keys
const (
	// restBfdKey is a REST path of a bfd
	restBfdKey = "/v1/bfd"
	// restBfdSessionKey is a REST path of a bfd sessions
	restBfdSessionKey = "/v1/bfd/sessions"
	// restBfdAuthKey is a REST path of a bfd authentication keys
	restBfdAuthKey = "/v1/bfd/authkeys"
)

// RestBfdKey returns prefix used in REST to dump bfd config
func RestBfdKey() string {
	return restBfdKey
}

// RestSessionKey returns prefix used in REST to dump bfd session config
func RestSessionKey() string {
	return restBfdSessionKey
}

// RestAuthKeysKey returns prefix used in REST to dump bfd authentication config
func RestAuthKeysKey() string {
	return restBfdAuthKey
}

// Interface REST keys
const (
	// restInterface is rest interface path
	restInterface = "/v1/interfaces"
	// restLoopback is path for loopback interface
	restLoopback = "/v1/interfaces/loopback"
	// restLoopback is path for physical interface
	restEthernet = "/v1/interfaces/ethernet"
	// restLoopback is path for memif interface
	restMemif = "/v1/interfaces/memif"
	// restLoopback is path for tap interface
	restTap = "/v1/interfaces/tap"
	// restAfPacket is path for af-packet interface
	restAfPacket = "/v1/interfaces/afpacket"
	// restLoopback is path for vxlan interface
	restVxLan = "/v1/interfaces/vxlan"
)

// RestInterfaceKey returns prefix used in REST to dump interface config
func RestInterfaceKey() string {
	return restInterface
}

// RestLoopbackKey returns prefix used in REST to dump loopback interface config
func RestLoopbackKey() string {
	return restLoopback
}

// RestEthernetKey returns prefix used in REST to dump ethernet interface config
func RestEthernetKey() string {
	return restEthernet
}

// RestMemifKey returns prefix used in REST to dump memif interface config
func RestMemifKey() string {
	return restMemif
}

// RestTapKey returns prefix used in REST to dump tap interface config
func RestTapKey() string {
	return restTap
}

// RestAfPAcketKey returns prefix used in REST to dump af-packet interface config
func RestAfPAcketKey() string {
	return restAfPacket
}

// RestVxLanKey returns prefix used in REST to dump VxLAN interface config
func RestVxLanKey() string {
	return restVxLan
}

// L2 plugin
const (
	// restBd is rest bridge domain path
	restBd = "/v1/bd"
	// restBdId is rest bridge domain ID path
	restBdId = "/v1/bdid"
	// restFib is rest FIB path
	restFib = "/v1/fib"
	// restXc is rest cross-connect path
	restXc = "/v1/xc"
)

// RestBridgeDomainKey returns the key used in REST to dump bridge domains.
func RestBridgeDomainKey() string {
	return restBd
}

// RestBridgeDomainIDKey returns the key used in REST to dump bridge domain IDs.
func RestBridgeDomainIDKey() string {
	return restBdId
}

// RestFibKey returns the prefix used in REST to dump vpp fib table entry config.
func RestFibKey() string {
	return restFib
}

// RestXConnectKey returns the prefix used in REST to dump vpp xConnect pair config.
func RestXConnectKey() string {
	return restXc
}
