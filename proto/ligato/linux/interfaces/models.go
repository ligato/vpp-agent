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

package linux_interfaces

import (
	"strings"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
)

// ModuleName is the module name used for models.
const ModuleName = "linux.interfaces"

var (
	ModelInterface = models.Register(&Interface{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "interface",
	})
)

// InterfaceKey returns the key used in ETCD to store configuration of a particular Linux interface.
func InterfaceKey(name string) string {
	return models.Key(&Interface{
		Name: name,
	})
}

const (
	/* Interface host-name (default ns only, notifications) */

	// InterfaceHostNameKeyPrefix is the common prefix of all keys representing
	// existing Linux interfaces in the default namespace (referenced by host names).
	InterfaceHostNameKeyPrefix = "linux/interface/host-name/"

	interfaceHostNameWithAddrKeyTmpl = InterfaceHostNameKeyPrefix + "{host-name}/address/{address}"
	interfaceHostNameWithVrfKeyTmpl  = InterfaceHostNameKeyPrefix + "{host-name}/vrf-host-name/{vrf-host-name}"

	/* Interface State (derived) */

	// InterfaceStateKeyPrefix is used as a common prefix for keys derived from
	// interfaces to represent the interface admin state (up/down).
	InterfaceStateKeyPrefix = "linux/interface/state/"

	// interfaceStateKeyTemplate is a template for (derived) key representing interface
	// admin state (up/down).
	interfaceStateKeyTemplate = InterfaceStateKeyPrefix + "{ifName}/{ifState}"

	// interface admin state as printed in derived keys.
	interfaceUpState   = "UP"
	interfaceDownState = "DOWN"

	/* Interface Address (derived) */

	// interfaceAddressKeyPrefix is used as a common prefix for keys derived from
	// interfaces to represent assigned IP addresses.
	interfaceAddrKeyPrefix = "linux/interface/{iface}/address/"

	// interfaceAddrKeyTmpl is a template for (derived) key representing IP address
	// (incl. mask) assigned to a Linux interface (referenced by the logical name).
	interfaceAddrKeyTmpl = interfaceAddrKeyPrefix + "{address-source}/{address}"

	/* Interface VRF (derived) */

	// interfaceVrfKeyTmpl is a template for (derived) key representing assignment
	// of a Linux interface into a VRF.
	interfaceVrfKeyTmpl = "linux/interface/{iface}/vrf/{vrf}"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

/* Interface host-name (default ns only, notifications) */

// InterfaceHostNameKey returns key representing Linux interface host name.
func InterfaceHostNameKey(hostName string) string {
	return InterfaceHostNameKeyPrefix + hostName
}

// InterfaceHostNameWithAddrKey returns key representing assignment of an IP address
// to a Linux interface referenced by its host name.
// If address is empty, the function returns key prefix matching any IP address.
func InterfaceHostNameWithAddrKey(hostName, address string) string {
	if hostName == "" {
		hostName = InvalidKeyPart
	}
	tmpl := interfaceHostNameWithAddrKeyTmpl
	key := strings.Replace(tmpl, "{host-name}", hostName, 1)
	key = strings.Replace(key, "{address}", address, 1)
	return key
}

// InterfaceHostNameWithVrfKey returns key representing association between Linux
// interface and Linux VRF, both referenced by host names.
// If vrf is empty, the function returns key prefix matching any VRF.
func InterfaceHostNameWithVrfKey(hostName, vrf string) string {
	if hostName == "" {
		hostName = InvalidKeyPart
	}
	tmpl := interfaceHostNameWithVrfKeyTmpl
	key := strings.Replace(tmpl, "{host-name}", hostName, 1)
	key = strings.Replace(key, "{vrf-host-name}", vrf, 1)
	return key
}

/* Interface State (derived) */

// InterfaceStateKey returns key representing admin state of a Linux interface.
func InterfaceStateKey(ifName string, ifIsUp bool) string {
	ifState := interfaceDownState
	if ifIsUp {
		ifState = interfaceUpState
	}
	key := strings.Replace(interfaceStateKeyTemplate, "{ifName}", ifName, 1)
	key = strings.Replace(key, "{ifState}", ifState, 1)
	return key
}

// ParseInterfaceStateKey parses interface name and state from key derived
// from interface by InterfaceStateKey().
func ParseInterfaceStateKey(key string) (ifName string, ifIsUp bool, isStateKey bool) {
	if strings.HasPrefix(key, InterfaceStateKeyPrefix) {
		keySuffix := strings.TrimPrefix(key, InterfaceStateKeyPrefix)
		keyComps := strings.Split(keySuffix, "/")
		if len(keyComps) != 2 {
			return "", false, false
		}
		ifName = keyComps[0]
		isStateKey = true
		if keyComps[1] == interfaceUpState {
			ifIsUp = true
		}
		return
	}
	return "", false, false
}

/* Interface Address (derived) */

// InterfaceAddressPrefix returns longest-common prefix of keys representing
// assigned IP addresses to a specific Linux interface.
func InterfaceAddressPrefix(iface string) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	return strings.Replace(interfaceAddrKeyPrefix, "{iface}", iface, 1)
}

// InterfaceAddressKey returns key representing IP address assigned to Linux interface.
// With undefined vrf the returned key can be also used as a key prefix, matching derived
// interface address key regardless of the VRF to which it belongs.
func InterfaceAddressKey(iface, address string, source netalloc.IPAddressSource) string {
	if iface == "" {
		iface = InvalidKeyPart
	}

	src := source.String()
	if src == "" {
		src = InvalidKeyPart
	}
	if strings.HasPrefix(address, netalloc.AllocRefPrefix) {
		src = netalloc.IPAddressSource_ALLOC_REF.String()
	}
	src = strings.ToLower(src)

	// construct key without validating the IP address
	tmpl := interfaceAddrKeyTmpl
	key := strings.Replace(tmpl, "{iface}", iface, 1)
	key = strings.Replace(key, "{address-source}", src, 1)
	key = strings.Replace(key, "{address}", address, 1)
	return key
}

// ParseInterfaceAddressKey parses interface address from key derived
// from interface by InterfaceAddressKey().
func ParseInterfaceAddressKey(key string) (iface, address string, source netalloc.IPAddressSource,
	invalidKey, isAddrKey bool) {
	parts := strings.Split(key, "/")
	if len(parts) < 4 || parts[0] != "linux" || parts[1] != "interface" {
		return
	}
	if parts[2] == "state" || parts[2] == "host-name" {
		return
	}

	addrIdx := -1
	for idx, part := range parts {
		switch part {
		case "vrf":
			// avoid collision with InterfaceVrfKey
			return
		case "address":
			addrIdx = idx
		}
	}
	if addrIdx == -1 {
		return
	}
	isAddrKey = true

	// parse interface name
	iface = strings.Join(parts[2:addrIdx], "/")
	if iface == "" {
		iface = InvalidKeyPart
		invalidKey = true
	}

	// parse address source
	src := strings.ToUpper(parts[addrIdx+1])
	srcInt, validSrc := netalloc.IPAddressSource_value[src]
	if !validSrc {
		invalidKey = true
		return
	}
	source = netalloc.IPAddressSource(srcInt)

	// return address as is (not parsed - this is done by the netalloc plugin)
	if addrIdx == len(parts)-1 {
		invalidKey = true
		return
	}
	address = strings.Join(parts[addrIdx+2:], "/")
	if address == "" {
		invalidKey = true
	}
	return
}

// InterfaceVrfKey returns key representing assignment of a Linux interface into a VRF.
func InterfaceVrfKey(iface string, vrf string) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	if vrf == "" {
		vrf = InvalidKeyPart
	}

	tmpl := interfaceVrfKeyTmpl
	key := strings.Replace(tmpl, "{iface}", iface, 1)
	key = strings.Replace(key, "{vrf}", vrf, 1)
	return key
}

// ParseInterfaceVrfKey parses interface VRF from key derived
// from interface by InterfaceVrfKey().
func ParseInterfaceVrfKey(key string) (iface string, vrf string, invalidKey, isVrfKey bool) {
	parts := strings.Split(key, "/")
	if len(parts) < 4 || parts[0] != "linux" || parts[1] != "interface" {
		return
	}
	if parts[2] == "state" || parts[2] == "host-name" {
		return
	}

	vrfIdx := -1
	for idx, part := range parts {
		switch part {
		case "address":
			// avoid collision with InterfaceAddressKey
			return
		case "vrf":
			vrfIdx = idx
		}
	}
	if vrfIdx == -1 {
		return
	}
	isVrfKey = true

	// parse interface name
	iface = strings.Join(parts[2:vrfIdx], "/")
	if iface == "" {
		iface = InvalidKeyPart
		invalidKey = true
	}

	// parse VRF
	if vrfIdx == len(parts)-1 {
		invalidKey = true
		vrf = InvalidKeyPart
		return
	}
	vrf = parts[vrfIdx+1]
	return
}
