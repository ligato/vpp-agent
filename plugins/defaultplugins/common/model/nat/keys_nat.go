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
	globalConfigPrefix = prefix + "/global/"
	// sNatPrefix is relative prefix for SNAT setup
	sNatPrefix = prefix + "/snat/"
	// dNatPrefix is relative prefix for DNAT setup
	dNatPrefix = prefix + "/dnat/"
	// globalConfigKeyTemplate is prefix for global config with vrf placeholder
	globalConfigKeyTemplate = prefix + "/global/" + vrfPlaceholder
	// sNatKeyTemplate is prefix for SNAT setup with vrf placeholder
	sNatKeyTemplate = prefix + "/snat/" + vrfPlaceholder
	// dNatKeyTemplate is prefix for DNAT setup with vrf placeholder
	dNatKeyTemplate = prefix + "/dnat/" + vrfPlaceholder
)

// GlobalConfigPrefix returns the prefix used in ETCD to store NAT global config
func GlobalConfigPrefix() string {
	return globalConfigPrefix
}

// GlobalConfigKey returns the key used in ETCD to store NAT global config
func GlobalConfigKey(vrf string) string {
	return strings.Replace(globalConfigKeyTemplate, vrfPlaceholder, vrf, 1)
}

// SNatPrefix returns the prefix used in ETCD to store SNAT config
func SNatPrefix() string {
	return sNatPrefix
}

// SNatKey returns the key used in ETCD to store SNAT config
func SNatKey(vrf string, label string) string {
	return strings.Replace(sNatKeyTemplate, vrfPlaceholder, vrf, 1) + label
}

// DNatPrefix returns the prefix used in ETCD to store NAT DNAT config
func DNatPrefix() string {
	return dNatPrefix
}

// DNatKey returns the key used in ETCD to store DNAT config
func DNatKey(vrf string, label string) string {
	return strings.Replace(dNatKeyTemplate, vrfPlaceholder, vrf, 1) + label
}
