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

import "strings"

const (
	/* NAT */

	// Prefix is a key prefix used in NB DB to store configuration for NAT.
	Prefix = "vpp/config/v2/nat/"

	// GlobalKey is the key used in NB DB to store global NAT configuration.
	GlobalKey = Prefix + "global/"

	// DNatPrefix is a key prefix used in NB DB to store DNAT configuration.
	DNatPrefix = Prefix + "dnat/"

	/* NAT interface */

	// natInterfaceKeyPrefix is a common prefix for (derived) keys each representing
	// NAT configuration for a single interface.
	natInterfaceKeyPrefix = "vpp/nat/interface/"

	// natInterfaceKeyTemplate is a template for (derived) key representing
	// NAT configuration for a single interface.
	natInterfaceKeyTemplate = natInterfaceKeyPrefix + "{iface}/feature/{feature}"

	// NAT interface features
	inFeature = "in"
	outFeature = "out"
)

/* NAT */

// DNatKey returns the key used in NB DB to store the configuration of the
// given DNAT configuration.
func DNatKey(label string) string {
	return DNatPrefix + label
}

/* NAT interface */

// InterfaceKey returns (derived) key representing NAT configuration of a given
// interface.
func InterfaceKey(iface string, isInside bool) string {
	key := strings.Replace(natInterfaceKeyTemplate, "{iface}", iface, 1)
	feature := inFeature
	if !isInside {
		feature = outFeature
	}
	key = strings.Replace(key, "{feature}", feature, 1)
	return key
}

// ParseInterfaceKey parses interface name and the assigned feature from NAT interface key.
func ParseInterfaceKey(key string) (iface string, isInside bool, isNATInterfaceKey bool) {
	if strings.HasPrefix(key, natInterfaceKeyPrefix) {
		keySuffix := strings.TrimPrefix(key, natInterfaceKeyPrefix)
		fibComps := strings.Split(keySuffix, "/")
		if len(fibComps) == 3 && fibComps[1] == "feature" {
			isInside := true
			if fibComps[2] == outFeature {
				isInside = false
			}
			return fibComps[0], isInside, true
		}
	}
	return "", false, false
}