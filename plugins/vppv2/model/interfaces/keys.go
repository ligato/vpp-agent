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
	"net"
	"strings"
)

const (
	/* Interface Config */

	// Prefix is a key prefix used in NB DB to store configuration for VPP interfaces.
	Prefix = "vpp/config/v2/interface/"

	/* Interface State */

	// StatePrefix is a key prefix used in NB DB to store interface states.
	StatePrefix = "vpp/status/v2/interface/"

	/* Interface Error */

	// ErrorPrefix is a key prefix used in NB DB to store interface errors.
	ErrorPrefix = "vpp/status/v2/interface/error/"

	/* Interface Address (derived) */

	// AddressKeyPrefix is used as a common prefix for keys derived from
	// interfaces to represent assigned IP addresses.
	AddressKeyPrefix = "vpp/interface/address/"

	// addressKeyTemplate is a template for (derived) key representing IP address
	// (incl. mask) assigned to a VPP interface.
	addressKeyTemplate = AddressKeyPrefix + "{iface}/{addr}/{mask}"

	/* Unnumbered interface (derived) */

	// UnnumberedKeyPrefix is used as a common prefix for keys derived from
	// interfaces to represent unnumbered interfaces.
	UnnumberedKeyPrefix = "vpp/interface/unnumbered/"

	/* DHCP (client - derived, lease - notification) */

	// DHCPClientKeyPrefix is used as a common prefix for keys derived from
	// interfaces to represent enabled DHCP clients.
	DHCPClientKeyPrefix = "vpp/interface/dhcp-client/"

	// DHCPLeaseKeyPrefix is used as a common prefix for keys representing
	// notifications with DHCP leases.
	DHCPLeaseKeyPrefix = "vpp/interface/dhcp-lease/"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

/* Interface Config */

// InterfaceKey returns the key used in NB DB to store the configuration of the
// given vpp interface.
func InterfaceKey(iface string) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	return Prefix + iface
}

// ParseNameFromKey returns suffix of the key.
func ParseNameFromKey(key string) (name string, isInterfaceKey bool) {
	if strings.HasPrefix(key, Prefix) {
		name = strings.TrimPrefix(key, Prefix)
		if name == "" {
			return "", false
		}
		return name, true
	}
	return "", false
}

/* Interface Error */

// InterfaceErrorKey returns the key used in NB DB to store the interface errors.
func InterfaceErrorKey(iface string) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	return ErrorPrefix + iface
}

/* Interface State */

// InterfaceStateKey returns the key used in NB DB to store the state data of the
// given vpp interface.
func InterfaceStateKey(iface string) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	return StatePrefix + iface
}

/* Interface Address (derived) */

// InterfaceAddressKey returns key representing IP address assigned to VPP interface.
func InterfaceAddressKey(iface string, address string) string {
	if iface == "" {
		iface = InvalidKeyPart
	}

	// parse address
	ipAddr, addrNet, err := net.ParseCIDR(address)
	if err != nil {
		address = InvalidKeyPart + "/" + InvalidKeyPart
	} else {
		addrNet.IP = ipAddr
		address = addrNet.String()
	}

	key := strings.Replace(addressKeyTemplate, "{iface}", iface, 1)
	key = strings.Replace(key, "{addr}/{mask}", address, 1)
	return key
}

// ParseInterfaceAddressKey parses interface address from key derived
// from interface by InterfaceAddressKey().
func ParseInterfaceAddressKey(key string) (iface string, ipAddr net.IP, ipAddrNet *net.IPNet, isAddrKey bool) {
	var err error
	if strings.HasPrefix(key, AddressKeyPrefix) {
		keySuffix := strings.TrimPrefix(key, AddressKeyPrefix)
		keyComps := strings.Split(keySuffix, "/")
		// beware: interface name may contain forward slashes (e.g. ETHERNET_CSMACD)
		if len(keyComps) < 3 {
			return "", nil, nil, false
		}
		// parse IP address
		lastIdx := len(keyComps) - 1
		ipAddr, ipAddrNet, err = net.ParseCIDR(keyComps[lastIdx-1] + "/" + keyComps[lastIdx])
		if err != nil {
			return "", nil, nil, false
		}
		// parse interface name
		iface = strings.Join(keyComps[:lastIdx-1], "/")
		if iface == "" {
			return "", nil, nil, false
		}
		return iface, ipAddr, ipAddrNet, true
	}
	return "", nil, nil, false
}

/* Unnumbered interface (derived) */

// UnnumberedKey returns key representing unnumbered interface.
func UnnumberedKey(iface string) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	return UnnumberedKeyPrefix + iface
}

/* DHCP (client - derived, lease - notification) */

// DHCPClientKey returns a (derived) key used to represent enabled DHCP lease.
func DHCPClientKey(iface string) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	return DHCPClientKeyPrefix + iface
}

// ParseNameFromDHCPClientKey returns suffix of the key.
func ParseNameFromDHCPClientKey(key string) (iface string, isDHCPClientKey bool) {
	if strings.HasPrefix(key, DHCPClientKeyPrefix) {
		iface = strings.TrimPrefix(key, DHCPClientKeyPrefix)
		if iface == "" {
			return "", false
		}
		return iface, true
	}
	return "", false
}

// DHCPLeaseKey returns a key used to represent DHCP lease for the given interface.
func DHCPLeaseKey(iface string) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	return DHCPLeaseKeyPrefix + iface
}
