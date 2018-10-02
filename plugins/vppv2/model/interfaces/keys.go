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
	"net"
	"strings"
)

const (
	/* Interface Config */

	// Prefix is a key prefix used in ETCD to store configuration for VPP interfaces.
	Prefix = "vpp/config/v1/interface/"

	/* Interface State */

	// StatePrefix is a key prefix used in ETCD to store interface states.
	StatePrefix = "vpp/status/v1/interface/"

	/* Interface Error */

	// ErrorPrefix is a key prefix used in ETCD to store interface errors.
	ErrorPrefix = "vpp/status/v1/interface/error/"

	/* Interface Address (derived) */

	// AddressKeyPrefix is used as a common prefix for keys derived from
	// interfaces to represent assigned IP addresses.
	AddressKeyPrefix = "vpp/interface/address/"

	// addressKeyTemplate is a template for (derived) key representing IP address
	// (incl. mask) assigned to a VPP interface.
	addressKeyTemplate = AddressKeyPrefix + "{ifName}/{addr}/{mask}"

	/* TAP host interface name (derived) */

	// TapHostNameKeyPrefix is used as a common prefix for keys derived from
	// TAP interfaces to represent the TAP host interface name.
	TapHostNameKeyPrefix = "vpp/interface/tap/host-if-name/"

	/* Unnumbered interface (derived) */

	// UnnumberedKeyPrefix is used as a common prefix for keys derived from
	// interfaces to represent unnumbered interfaces.
	UnnumberedKeyPrefix = "vpp/interface/unnumbered/"

)

/* Interface Config */

// ParseNameFromKey returns suffix of the key.
func ParseNameFromKey(key string) (name string, err error) {
	if strings.HasPrefix(key, Prefix) {
		name = strings.TrimPrefix(key, Prefix)
		return
	}
	return key, fmt.Errorf("wrong format of the key %s", key)
}

// InterfaceKey returns the key used in ETCD to store the configuration of the
// given vpp interface.
func InterfaceKey(ifaceLabel string) string {
	return Prefix + ifaceLabel
}

/* Interface Error */

// InterfaceErrorKey returns the key used in ETCD to store the interface errors.
func InterfaceErrorKey(ifaceLabel string) string {
	return ErrorPrefix + ifaceLabel
}

/* Interface State */

// InterfaceStateKey returns the key used in ETCD to store the state data of the
// given vpp interface.
func InterfaceStateKey(ifaceLabel string) string {
	return StatePrefix + ifaceLabel
}

/* Interface Address (derived) */

// InterfaceAddressKey returns key representing IP address assigned to VPP interface.
func InterfaceAddressKey(ifName string, address string) string {
	var mask string
	addrComps := strings.Split(address, "/")
	addr := addrComps[0]
	if len(addrComps) > 0 {
		mask = addrComps[1]
	}
	key := strings.Replace(addressKeyTemplate, "{ifName}", ifName, 1)
	key = strings.Replace(key, "{addr}", addr, 1)
	key = strings.Replace(key, "{mask}", mask, 1)
	return key
}

// ParseInterfaceAddressKey parses interface address from key derived
// from interface by InterfaceAddressKey().
func ParseInterfaceAddressKey(key string) (ifName string, ifAddr *net.IPNet, err error) {
	errPrefix := "invalid VPP interface address key: "
	if strings.HasPrefix(key, AddressKeyPrefix) {
		keySuffix := strings.TrimPrefix(key, AddressKeyPrefix)
		keyComps := strings.Split(keySuffix, "/")
		// beware: interface name may contain forward slashes (e.g. ETHERNET_CSMACD)
		if len(keyComps) < 3 {
			return "", nil, fmt.Errorf(errPrefix + "invalid suffix")
		}
		lastIdx := len(keyComps)-1
		_, ifAddr, err = net.ParseCIDR(keyComps[lastIdx-1] + "/" + keyComps[lastIdx])
		if err != nil {
			return "", nil, fmt.Errorf(errPrefix + "invalid address")
		}
		ifName = strings.Join(keyComps[:lastIdx-1], "/")
		return
	}
	return "", nil, fmt.Errorf(errPrefix + "invalid prefix")
}

/* TAP host interface name (derived) */

// TAPHostNameKey returns key representing TAP interface host name.
func TAPHostNameKey(hostIfName string) string {
	return TapHostNameKeyPrefix + hostIfName
}

/* Unnumbered interface (derived) */

// UnnumberedKey returns key representing unnumbered interface.
func UnnumberedKey(ifName string) string {
	return UnnumberedKeyPrefix + ifName
}