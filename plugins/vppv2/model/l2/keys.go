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

package l2

import (
	"strings"
	"net"
)

// Prefixes
const (
	/* Bridge Domain */

	// BDPrefix is a key prefix used in NB DB to store configuration for bridge domains.
	BDPrefix = "vpp/config/v2/bd/"

	/* BD <-> interface binding (derived) */

	// bdInterfaceKeyTemplate is a template for (derived) key representing binding
	// between interface and a bridge domain.
	bdInterfaceKeyTemplate = "vpp/bd/{bd}/interface/{iface}"

	/* L2 FIB */

	// fibKeyTemplate is a template for key representing configuration for a L2 FIB.
	fibKeyTemplate = BDPrefix + "{bd}/fib/{hwAddr}"

	/* xConnect */

	// XConnectPrefix is a key prefix used in NB DB to store configuration for
	// xConnects.
	XConnectPrefix = "vpp/config/v2/xconnect/"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

/* Bridge Domain */

// BridgeDomainKey returns the key used in NB DB to store the configuration of the
// given bridge domain.
func BridgeDomainKey(bdName string) string {
	if bdName == "" {
		bdName = InvalidKeyPart
	}
	return BDPrefix + bdName
}

// ParseBDNameFromKey returns BD name from the key.
func ParseBDNameFromKey(key string) (name string, isBDKey bool) {
	if strings.HasPrefix(key, BDPrefix) {
		suffix := strings.TrimPrefix(key, BDPrefix)
		if strings.ContainsAny(suffix, "/") {
			return "", false
		}
		return suffix, true
	}
	return "", false
}

/* BD <-> interface binding (derived) */

// BDInterfaceKey returns the key used to represent binding between the given interface
// and the bridge domain.
func BDInterfaceKey(bdName string, iface string) string {
	if bdName == "" {
		bdName = InvalidKeyPart
	}
	if iface == "" {
		iface = InvalidKeyPart
	}
	key := strings.Replace(bdInterfaceKeyTemplate, "{bd}", bdName, 1)
	key = strings.Replace(key, "{iface}", iface, 1)
	return key
}

// ParseBDInterfaceKey parses key representing binding between interface and a bridge
// domain.
func ParseBDInterfaceKey(key string) (bdName string, iface string, isBDIfaceKey bool) {
	keyComps := strings.Split(key, "/")
	if len(keyComps) >= 5 && keyComps[0] == "vpp" && keyComps[1] == "bd" && keyComps[3] == "interface" {
		iface = strings.Join(keyComps[4:], "/")
		return keyComps[2], iface, true
	}
	return "", "", false
}

/* L2 FIB */

// FIBKey returns the key used in NB DB to store the configuration of the
// given L2 FIB entry.
func FIBKey(bdName string, fibMac string) string {
	if bdName == "" {
		bdName = InvalidKeyPart
	}
	if _, err := net.ParseMAC(fibMac); err != nil {
		fibMac = InvalidKeyPart
	}
	key := strings.Replace(fibKeyTemplate, "{bd}", bdName, 1)
	key = strings.Replace(key, "{hwAddr}", fibMac, 1)
	return key
}

// ParseFIBKey parses bridge domain label and FIB MAC address from a FIB key.
func ParseFIBKey(key string) (bdName string, fibMac string, isFIBKey bool) {
	if strings.HasPrefix(key, BDPrefix) {
		bdSuffix := strings.TrimPrefix(key, BDPrefix)
		fibComps := strings.Split(bdSuffix, "/")
		if len(fibComps) == 3 && fibComps[1] == "fib" {
			return fibComps[0], fibComps[2], true
		}
	}
	return "", "", false
}

/* xConnect */

// XConnectKey returns the key used in NB DB to store the configuration of the
// given xConnect (identified by RX interface).
func XConnectKey(rxIface string) string {
	if rxIface == "" {
		rxIface = InvalidKeyPart
	}
	return XConnectPrefix + rxIface
}
