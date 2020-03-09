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

package vpp_ipsec

import (
	"strconv"
	"strings"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

// ModuleName is the module name used for models.
const ModuleName = "vpp.ipsec"

var (
	ModelSecurityPolicyDatabase = models.Register(&SecurityPolicyDatabase{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "spd",
	}, models.WithNameTemplate("{{.Index}}"))

	ModelSecurityAssociation = models.Register(&SecurityAssociation{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "sa",
	}, models.WithNameTemplate("{{.Index}}"))

	ModelTunnelProtection = models.Register(&TunnelProtection{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "tun-protect",
	}, models.WithNameTemplate(
		`{{.Interface}}`,
	))
)

// SPDKey returns the key used in NB DB to store the configuration of the
// given security policy database configuration.
func SPDKey(index uint32) string {
	return models.Key(&SecurityPolicyDatabase{
		Index: index,
	})
}

// SAKey returns the key used in NB DB to store the configuration of the
// given security association configuration.
func SAKey(index uint32) string {
	return models.Key(&SecurityAssociation{
		Index: index,
	})
}

/* SPD <-> interface binding (derived) */
const (
	// spdInterfaceKeyTemplate is a template for (derived) key representing binding
	// between interface and a security policy database.
	spdInterfaceKeyTemplate = "vpp/spd/{spd}/interface/{iface}"
)

/* SPD <-> policy binding (derived) */
const (
	// spdPolicyKeyTemplate is a template for (derived) key representing binding
	// between policy (security association) and a security policy database.
	spdPolicyKeyTemplate = "vpp/spd/{spd}/sa/{sa}"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

/* SPD <-> interface binding (derived) */

// SPDInterfaceKey returns the key used to represent binding between the given interface
// and the security policy database.
func SPDInterfaceKey(spdIndex uint32, ifName string) string {
	if ifName == "" {
		ifName = InvalidKeyPart
	}
	key := strings.Replace(spdInterfaceKeyTemplate, "{spd}", strconv.FormatUint(uint64(spdIndex), 10), 1)
	key = strings.Replace(key, "{iface}", ifName, 1)
	return key
}

// ParseSPDInterfaceKey parses key representing binding between interface and a security
// policy database
func ParseSPDInterfaceKey(key string) (spdIndex string, iface string, isSPDIfaceKey bool) {
	keyComps := strings.Split(key, "/")
	if len(keyComps) >= 5 && keyComps[0] == "vpp" && keyComps[1] == "spd" && keyComps[3] == "interface" {
		iface = strings.Join(keyComps[4:], "/")
		return keyComps[2], iface, true
	}
	return "", "", false
}

/* SPD <-> policy binding (derived) */

// SPDPolicyKey returns the key used to represent binding between the given policy
// (security association) and the security policy database.
func SPDPolicyKey(spdIndex uint32, saIndex uint32) string {
	key := strings.Replace(spdPolicyKeyTemplate, "{spd}", strconv.FormatUint(uint64(spdIndex), 10), 1)
	key = strings.Replace(key, "{sa}", strconv.FormatUint(uint64(saIndex), 10), 1)
	return key
}

// ParseSPDPolicyKey parses key representing binding between policy (security
// association) and a security policy database
func ParseSPDPolicyKey(key string) (spdIndex string, saIndex string, isSPDIfaceKey bool) {
	keyComps := strings.Split(key, "/")
	if len(keyComps) >= 5 && keyComps[0] == "vpp" && keyComps[1] == "spd" && keyComps[3] == "sa" {
		saIndex = strings.Join(keyComps[4:], "/")
		return keyComps[2], saIndex, true
	}
	return "", "", false
}
