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
	"strings"
)

const (
	// InterfacePrefix vpp/config/v1/interface/
	InterfacePrefix = "vpp/config/v1/interface/"
	// IfStatePrefix vpp/status/v1/interface/
	IfStatePrefix = "vpp/status/v1/interface/"
	// IfErrorPrefix vpp/status/v1/interface/error
	IfErrorPrefix = "vpp/status/v1/interface/error/"
)

// InterfaceKeyPrefix returns the prefix used in ETCD to store vpp interfaces config.
func InterfaceKeyPrefix() string {
	return InterfacePrefix
}

// ParseNameFromKey returns suffix of the key.
func ParseNameFromKey(key string) (name string, err error) {
	lastSlashPos := strings.LastIndex(key, "/")
	if lastSlashPos > 0 && lastSlashPos < len(key)-1 {
		return key[lastSlashPos+1:], nil
	}

	return key, fmt.Errorf("wrong format of the key %s", key)
}

// InterfaceKey returns the prefix used in ETCD to store the vpp interface config
// of a particular interface in selected vpp instance.
func InterfaceKey(ifaceLabel string) string {
	return InterfacePrefix + ifaceLabel
}

// InterfaceErrorPrefix returns the prefix used in ETCD to store the interface errors.
func InterfaceErrorPrefix() string {
	return IfErrorPrefix
}

// InterfaceErrorKey returns the key used in ETCD to store the interface errors.
func InterfaceErrorKey(ifaceLabel string) string {
	return IfErrorPrefix + ifaceLabel
}

// InterfaceStateKeyPrefix returns the prefix used in ETCD to store the vpp interfaces state data.
func InterfaceStateKeyPrefix() string {
	return IfStatePrefix
}

// InterfaceStateKey returns the prefix used in ETCD to store the vpp interface state data
// of particular interface in selected vpp instance.
func InterfaceStateKey(ifaceLabel string) string {
	return IfStatePrefix + ifaceLabel
}
