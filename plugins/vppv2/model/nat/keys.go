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

	// natInterfaceKeyTemplate is a template for (derived) key representing
	// NAT configuration for a single interface.
	natInterfaceKeyTemplate = "vpp/nat/interface/{iface}/{feature}"

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

// NATInterfaceKey returns (derived) key representing NAT configuration of a given
// interface.
func NATInterfaceKey(iface string, isInside bool) string {
	key := strings.Replace(natInterfaceKeyTemplate, "{iface}", iface, 1)
	feature := inFeature
	if !isInside {
		feature = outFeature
	}
	key = strings.Replace(key, "{feature}", feature, 1)
	return key
}
