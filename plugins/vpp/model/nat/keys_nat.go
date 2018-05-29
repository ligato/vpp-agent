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

const (
	// natPrefix
	prefix = "vpp/config/v1/nat"
	// globalConfigPrefix is relative prefix for global config
	globalConfig = prefix + "/global/"
	// sNatPrefix is relative prefix for SNAT setup
	sNatPrefix = prefix + "/snat/"
	// dNatPrefix is relative prefix for DNAT setup
	dNatPrefix = prefix + "/dnat/"
)

// GlobalConfigPrefix returns the prefix used in ETCD to store NAT global config
func GlobalConfigPrefix() string {
	return globalConfig
}

// GlobalConfigKey returns the key used in ETCD to store NAT global config. Global config can be stored only once,
// so the prefix == key
func GlobalConfigKey() string {
	return globalConfig
}

// SNatPrefix returns the prefix used in ETCD to store SNAT config
func SNatPrefix() string {
	return sNatPrefix
}

// SNatKey returns the key used in ETCD to store SNAT config
func SNatKey(label string) string {
	return sNatPrefix + label
}

// DNatPrefix returns the prefix used in ETCD to store NAT DNAT config
func DNatPrefix() string {
	return dNatPrefix
}

// DNatKey returns the key used in ETCD to store DNAT config
func DNatKey(label string) string {
	return dNatPrefix + label
}
