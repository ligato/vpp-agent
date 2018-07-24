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

package interfaces

import (
	"fmt"
	"strings"

	"github.com/ligato/vpp-agent/plugins/vpp/model"
)

const (
	// interfacePrefix is interface prefix
	interfacePrefix = "vpp/config" + model.ProtoApiVersion + "interface/"
	// ifStatePrefix is interface state prefix
	ifStatePrefix = "vpp/status" + model.ProtoApiVersion + "interface/"
	// ifErrorPrefix is interface error prefix
	ifErrorPrefix = "vpp/status" + model.ProtoApiVersion + "interface/error/"
	// restInterface is rest interface path
	restInterface = model.ProtoApiVersion + "interface"
	// restLoopback is path for loopback interface
	restLoopback = model.ProtoApiVersion + "interface/loopback"
	// restLoopback is path for physical interface
	restEthernet = model.ProtoApiVersion + "interface/ethernet"
	// restLoopback is path for memif interface
	restMemif = model.ProtoApiVersion + "interface/memif"
	// restLoopback is path for tap interface
	restTap = model.ProtoApiVersion + "interface/tap"
	// restAfPacket is path for af-packet interface
	restAfPacket = model.ProtoApiVersion + "interface/afpacket"
	// restLoopback is path for vxlan interface
	restVxLan = model.ProtoApiVersion + "interface/vxlan"
)

// InterfaceKeyPrefix returns the prefix used in ETCD to store vpp interfaces config.
func InterfaceKeyPrefix() string {
	return interfacePrefix
}

// ParseNameFromKey returns suffix of the key.
func ParseNameFromKey(key string) (name string, err error) {
	lastSlashPos := strings.LastIndex(key, "/")
	if lastSlashPos > 0 && lastSlashPos < len(key)-1 {
		return key[lastSlashPos+1:], nil
	}

	return key, fmt.Errorf("wrong format of the key %s", key)
}

// InterfaceKey returns the prefix used in ETCD to store the vpp interface config
// of a particular interface in selected vpp instance.
func InterfaceKey(ifaceLabel string) string {
	return interfacePrefix + ifaceLabel
}

// InterfaceErrorPrefix returns the prefix used in ETCD to store the interface errors.
func InterfaceErrorPrefix() string {
	return ifErrorPrefix
}

// InterfaceErrorKey returns the key used in ETCD to store the interface errors.
func InterfaceErrorKey(ifaceLabel string) string {
	return ifErrorPrefix + ifaceLabel
}

// InterfaceStateKeyPrefix returns the prefix used in ETCD to store the vpp interfaces state data.
func InterfaceStateKeyPrefix() string {
	return ifStatePrefix
}

// InterfaceStateKey returns the prefix used in ETCD to store the vpp interface state data
// of particular interface in selected vpp instance.
func InterfaceStateKey(ifaceLabel string) string {
	return ifStatePrefix + ifaceLabel
}

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
