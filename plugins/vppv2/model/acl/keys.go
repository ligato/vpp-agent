//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package acl

import (
	"fmt"
	"strings"
)

const (
	// Prefix is ACL key prefix
	Prefix = "vpp/config/v2/acl/"

	aclToInterfaceTemplate = "vpp/acl/{acl}/interface/{flow}/{iface}"

	// IngressFlow represents ingress packet flow
	IngressFlow = "ingress"
	// EgressFlow represents egress packet flow
	EgressFlow = "egress"
)

// Key returns the prefix used in ETCD to store vpp ACL config
// of a particular ACL in selected vpp instance.
func Key(aclName string) string {
	return Prefix + aclName
}

// ParseNameFromKey returns suffix of the key.
func ParseNameFromKey(key string) (name string, err error) {
	if name = strings.TrimPrefix(key, Prefix); name == key {
		return name, fmt.Errorf("missing ACL prefix in key: %s", key)
	}
	return name, nil
}

// ACLToInterfaceKey returns key for ACL to interface
func ACLToInterfaceKey(acl, iface, flow string) string {
	key := aclToInterfaceTemplate
	key = strings.Replace(key, "{acl}", acl, 1)
	key = strings.Replace(key, "{flow}", flow, 1)
	key = strings.Replace(key, "{iface}", iface, 1)
	return key
}

// ParseACLToInterfaceKey parses ACL to interface key
func ParseACLToInterfaceKey(key string) (acl, iface, flow string, isACLToInterface bool) {
	parts := strings.Split(key, "/")
	if len(parts) == 6 &&
		parts[0] == "vpp" && parts[1] == "acl" && parts[3] == "interface" &&
		(parts[4] == IngressFlow || parts[4] == EgressFlow) {
		return parts[2], parts[5], parts[4], true
	}
	return "", "", "", false
}
