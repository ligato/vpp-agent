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
	// Prefix is a key prefix used in ETCD to store configuration for VPP interfaces.
	Prefix = "vpp/config/v1/interface/"
	// StatePrefix is a key prefix used in ETCD to store interface states.
	StatePrefix = "vpp/status/v1/interface/"
	// ErrorPrefix is a key prefix used in ETCD to store interface errors.
	ErrorPrefix = "vpp/status/v1/interface/error/"
)

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

// InterfaceErrorKey returns the key used in ETCD to store the interface errors.
func InterfaceErrorKey(ifaceLabel string) string {
	return ErrorPrefix + ifaceLabel
}

// InterfaceStateKey returns the key used in ETCD to store the state data of the
// given vpp interface.
func InterfaceStateKey(ifaceLabel string) string {
	return StatePrefix + ifaceLabel
}
