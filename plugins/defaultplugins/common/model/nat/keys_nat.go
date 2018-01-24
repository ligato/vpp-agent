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
)

const (
	// VRF placeholder
	vrfPlaceholder = "vrf/{vrf}"
	// natPrefix
	prefix = "vpp/config/v1/nat/"
	// globalConfigPrefix is relative prefix for global config
	globalConfigPrefix = prefix + vrfPlaceholder + "/global/"
	// sNatPrefix is relative prefix for SNAT setup
	sNatPrefix = prefix + vrfPlaceholder + "/snat/"
	// dNatPrefix is relative prefix for DNAT setup
	dNatPrefix = prefix + vrfPlaceholder + "/dnat/"
)

// Prefix returns the common prefix for NAT configuration
func Prefix() string {
	return prefix
}

// GlobalConfigPrefix returns the prefix used in ETCD to store NAT global config
func GlobalConfigPrefix() string {
	return globalConfigPrefix
}

// GlobalConfigKey returns the key used in ETCD to store NAT global config
func GlobalConfigKey() string {
	return globalConfigPrefix + "config"
}

// SNatPrefix returns the prefix used in ETCD to store SNAT config
func SNatPrefix() string {
	return sNatPrefix
}

// SNatKey returns the key used in ETCD to store SNAT config
func SNatKey(vrf string, label string) string {
	return strings.Replace(sNatPrefix, vrfPlaceholder, vrf, 1) + label
}

// DNatPrefix returns the prefix used in ETCD to store NAT DNAT config
func DNatPrefix() string {
	return dNatPrefix
}

// DNatKey returns the key used in ETCD to store DNAT config
func DNatKey(vrf string, label string) string {
	return strings.Replace(dNatPrefix, vrfPlaceholder, vrf, 1) + label
}

// DeriveNATConfigType resolves NAT configuration type using provided key
func DeriveNATConfigType(key string) (global, snat, dnat bool) {
	if strings.Contains(key, "/global/") {
		global = true
	} else if strings.Contains(key, "/snat/") {
		snat = true
	} else if strings.Contains(key, "/dnat/") {
		dnat = true
	}
	return
}
