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

package mock_l2

import (
	"strings"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

// ModuleName is the module name used for models of the mock l2plugin.
const ModuleName = "mock"

var (
	ModelBridgeDomain = models.Register(&BridgeDomain{}, models.Spec{
		Module:  ModuleName,
		Type:    "bridge-domain",
		Version: "v1",
	})

	ModelFIBEntry = models.Register(&FIBEntry{}, models.Spec{
		Module:  ModuleName,
		Type:    "fib",
		Version: "v1",
	}, models.WithNameTemplate("{{.BridgeDomain}}/mac/{{.PhysAddress}}"))
)

// BridgeDomainKey returns the key used in NB DB to store the configuration of the
// given mock bridge domain.
func BridgeDomainKey(bdName string) string {
	return models.Key(&BridgeDomain{
		Name: bdName,
	})
}

// FIBKey returns the key used in NB DB to store the configuration of the
// given mock L2 FIB entry.
func FIBKey(bdName string, fibMac string) string {
	return models.Key(&FIBEntry{
		BridgeDomain: bdName,
		PhysAddress:  fibMac,
	})
}

/* Note: We do not yet provide tools to build models for derived values, therefore
   the key template and the key building/parsing methods have to be defined
   from the scratch for the time being.
*/

/* BD <-> interface binding (derived value) */
const (
	// bdInterfaceKeyTemplate is a template for (derived) key representing binding
	// between interface and a bridge domain.
	bdInterfaceKeyTemplate = "mock/bd/{bd}/interface/{iface}"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

// BDInterfaceKey returns the key used to represent binding between the given interface
// and the bridge domain.
func BDInterfaceKey(bdName string, iface string) string {
	if bdName == "" {
		bdName = InvalidKeyPart
	}
	if iface == "" {
		iface = InvalidKeyPart
	}
	key := strings.Replace(bdInterfaceKeyTemplate, "{bd}", bdName, 1)
	key = strings.Replace(key, "{iface}", iface, 1)
	return key
}

// ParseBDInterfaceKey parses key representing binding between interface and a bridge
// domain.
func ParseBDInterfaceKey(key string) (bdName string, iface string, isBDIfaceKey bool) {
	keyComps := strings.Split(key, "/")
	if len(keyComps) >= 5 && keyComps[0] == "mock" && keyComps[1] == "bd" && keyComps[3] == "interface" {
		iface = strings.Join(keyComps[4:], "/")
		return keyComps[2], iface, true
	}
	return "", "", false
}
