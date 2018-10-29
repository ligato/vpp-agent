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

package nat

import (
	"strings"
	"net"
)

const (
	/* NAT44 */

	// PrefixNAT44 is a key prefix used in NB DB to store configuration for NAT44.
	PrefixNAT44 = "vpp/config/v2/nat44/"

	// GlobalNAT44Key is the key used in NB DB to store global NAT44 configuration.
	GlobalNAT44Key = PrefixNAT44 + "global"

	// DNAT44Prefix is a key prefix used in NB DB to store DNAT-44 configuration.
	DNAT44Prefix = PrefixNAT44 + "dnat/"

	/* NAT44 interface */

	// interfaceNAT44KeyPrefix is a common prefix for (derived) keys each representing
	// NAT44 configuration for a single interface.
	interfaceNAT44KeyPrefix = "vpp/nat44/interface/"

	// interfaceNAT44KeyTemplate is a template for (derived) key representing
	// NAT44 configuration for a single interface.
	interfaceNAT44KeyTemplate = interfaceNAT44KeyPrefix + "{iface}/feature/{feature}"

	// NAT interface features
	inFeature = "in"
	outFeature = "out"

	/* NAT44 address pool */

	// addressNAT44KeyPrefix is a common prefix for (derived) keys representing
	// addresses from NAT44 address pool.
	addressNAT44KeyPrefix = "vpp/nat44/address/"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

/* NAT44 */

// DNAT44Key returns the key used in NB DB to store the configuration of the
// given DNAT-44 configuration.
func DNAT44Key(label string) string {
	if label == "" {
		label = InvalidKeyPart
	}
	return DNAT44Prefix + label
}

/* NAT44 interface */

// InterfaceNAT44Key returns (derived) key representing NAT44 configuration
// for a given interface.
func InterfaceNAT44Key(iface string, isInside bool) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	key := strings.Replace(interfaceNAT44KeyTemplate, "{iface}", iface, 1)
	feature := inFeature
	if !isInside {
		feature = outFeature
	}
	key = strings.Replace(key, "{feature}", feature, 1)
	return key
}

// ParseInterfaceNAT44Key parses interface name and the assigned NAT44 feature
// from Interface-NAT44 key.
func ParseInterfaceNAT44Key(key string) (iface string, isInside bool, isInterfaceNAT44Key bool) {
	if strings.HasPrefix(key, interfaceNAT44KeyPrefix) {
		keySuffix := strings.TrimPrefix(key, interfaceNAT44KeyPrefix)
		fibComps := strings.Split(keySuffix, "/")
		if len(fibComps) >= 3 && fibComps[len(fibComps)-2] == "feature" {
			isInside := true
			if fibComps[len(fibComps)-1] == outFeature {
				isInside = false
			}
			iface := strings.Join(fibComps[:len(fibComps)-2], "/")
			return iface, isInside, true
		}
	}
	return "", false, false
}

/* NAT44 address pool */

// AddressNAT44Key returns (derived) key representing address from NAT44 address
// pool.
func AddressNAT44Key(address string) string {
	ipAddr := net.ParseIP(address)
	if ipAddr == nil {
		address = InvalidKeyPart
	} else {
		address = ipAddr.String()
	}
	return addressNAT44KeyPrefix + address
}