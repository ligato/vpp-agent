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

package vpp_acl

import (
	"strings"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

// ModuleName is the module name used for models.
const ModuleName = "vpp.acls"

var (
	ModelACL = models.Register(&ACL{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "acl",
	})
)

// Key returns the prefix used in ETCD to store vpp ACL config
// of a particular ACL in selected vpp instance.
func Key(aclName string) string {
	return models.Key(&ACL{
		Name: aclName,
	})
}

const (
	aclToInterfaceTemplate = "vpp/acl/{acl}/interface/{flow}/{iface}"

	// IngressFlow represents ingress packet flow
	IngressFlow = "ingress"
	// EgressFlow represents egress packet flow
	EgressFlow = "egress"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

// ToInterfaceKey returns key for ACL to interface
func ToInterfaceKey(acl, iface, flow string) string {
	if acl == "" {
		acl = InvalidKeyPart
	}
	if iface == "" {
		iface = InvalidKeyPart
	}
	if flow != IngressFlow && flow != EgressFlow {
		flow = InvalidKeyPart
	}
	key := aclToInterfaceTemplate
	key = strings.Replace(key, "{acl}", acl, 1)
	key = strings.Replace(key, "{flow}", flow, 1)
	key = strings.Replace(key, "{iface}", iface, 1)
	return key
}

// ParseACLToInterfaceKey parses ACL to interface key
func ParseACLToInterfaceKey(key string) (acl, iface, flow string, isACLToInterface bool) {
	parts := strings.Split(key, "/")

	interfaceIndex := -1
	for index, part := range parts {
		if part == "interface" {
			interfaceIndex = index
			break
		}
	}

	if len(parts) > interfaceIndex+1 {
		flow = parts[interfaceIndex+1]
	}

	if len(parts) >= 5 &&
		parts[0] == "vpp" && parts[1] == "acl" && interfaceIndex >= 3 &&
		(flow == IngressFlow || flow == EgressFlow || flow == InvalidKeyPart) {
		acl = strings.Join(parts[2:interfaceIndex], "/")
		iface = strings.Join(parts[interfaceIndex+2:], "/")
		if iface != "" && acl != "" {
			return acl, iface, flow, true
		}
	}
	return "", "", "", false
}
