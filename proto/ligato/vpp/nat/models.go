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

package vpp_nat

import (
	"fmt"
	"strconv"
	"strings"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

// ModuleName is the module name used for models.
const ModuleName = "vpp.nat"

var (
	ModelNat44Global = models.Register(&Nat44Global{}, models.Spec{
		Module:  ModuleName,
		Type:    "nat44-global",
		Version: "v2",
	})

	ModelDNat44 = models.Register(&DNat44{}, models.Spec{
		Module:  ModuleName,
		Type:    "dnat44",
		Version: "v2",
	}, models.WithNameTemplate("{{.Label}}"))

	ModelNat44Interface = models.Register(&Nat44Interface{}, models.Spec{
		Module:  ModuleName,
		Type:    "nat44-interface",
		Version: "v2",
	}, models.WithNameTemplate("{{.Name}}"))

	ModelNat44AddressPool = models.Register(&Nat44AddressPool{}, models.Spec{
		Module:  ModuleName,
		Type:    "nat44-pool",
		Version: "v2",
	}, models.WithNameTemplate(
		"{{if .Name}}"+
			"{{.Name}}"+
			"{{else}}"+
			"vrf/{{.VrfId}}"+
			"/address/{{.FirstIp}}"+
			"{{if and .LastIp (ne .FirstIp .LastIp)}}-{{.LastIp}}{{end}}"+
			"{{end}}",
	))
)

// GlobalNAT44Key returns key for Nat44Global.
func GlobalNAT44Key() string {
	return models.Key(&Nat44Global{})
}

// DNAT44Key returns the key used in NB DB to store the configuration of the
// given DNAT-44 configuration.
func DNAT44Key(label string) string {
	return models.Key(&DNat44{
		Label: label,
	})
}

// Nat44InterfaceKey returns the key used in NB DB to store the configuration of the
// given NAT44 interface.
func Nat44InterfaceKey(name string) string {
	return models.Key(&Nat44Interface{
		Name: name,
	})
}

/* NAT44 mode (derived) */

const (
	// key derived when NAT44 is configured in the endpoint-dependent mode
	Nat44EndpointDepKey = "vpp/nat44/endpoint-dependent"
)

/* NAT44 interface (derived) */

const (
	// interfaceNAT44KeyPrefix is a common prefix for (derived) keys each representing
	// NAT44 configuration for a single interface.
	interfaceNAT44KeyPrefix = "vpp/nat44/interface/"

	// interfaceNAT44KeyTemplate is a template for (derived) key representing
	// NAT44 configuration for a single interface.
	interfaceNAT44KeyTemplate = interfaceNAT44KeyPrefix + "{iface}/feature/{feature}"

	// NAT interface features
	inFeature  = "in"
	outFeature = "out"
)

/* NAT44 address (derived) */

const (
	// addressNAT44KeyPrefix is a common prefix for (derived) keys each representing
	// single address from the NAT44 address pool.
	addressNAT44KeyPrefix = "vpp/nat44/address/"

	// addressNAT44KeyTemplate is a template for (derived) key representing
	// single address from the NAT44 address pool.
	addressNAT44KeyTemplate = addressNAT44KeyPrefix + "{address}/twice-nat/{twice-nat}"

	// twice-NAT switch
	twiceNatOn  = "on"
	twiceNatOff = "off"

	// TwiceNATDerivedKeyPrefix is common prefix for (derived) keys each representing twiceNAT address pool
	TwiceNATDerivedKeyPrefix = "vpp/nat44/twiceNAT-pool/"

	// twiceNATKeyTemplate is a template for (derived) key
	// representing twiceNAT address pool with single IP range.
	twiceNATKeyTemplate = TwiceNATDerivedKeyPrefix + "vrf/{vrfID}/addresses/{firstIP}"

	// twiceNATWithMultipeAddressesKeyTemplate is a template for (derived) key
	// representing twiceNAT address pool with multiple IP range.
	twiceNATWithMultipeAddressesKeyTemplate = TwiceNATDerivedKeyPrefix + "vrf/{vrfID}/addresses/{firstIP}-{lastIP}"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

/* NAT44 interface (derived) */

// DerivedInterfaceNAT44Key returns (derived) key representing NAT44 configuration
// for a given interface.
func DerivedInterfaceNAT44Key(iface string, isInside bool) string {
	if iface == "" {
		iface = InvalidKeyPart
	}
	key := strings.Replace(interfaceNAT44KeyTemplate, "{iface}", iface, 1)
	feature := inFeature
	if !isInside {
		feature = outFeature
	}
	key = strings.Replace(key, "{feature}", feature, 1)
	return key
}

// ParseDerivedInterfaceNAT44Key parses interface name and the assigned NAT44 feature
// from Interface-NAT44 key.
func ParseDerivedInterfaceNAT44Key(key string) (iface string, isInside bool, isInterfaceNAT44Key bool) {
	trim := strings.TrimPrefix(key, interfaceNAT44KeyPrefix)
	if trim != key && trim != "" {
		fibComps := strings.Split(trim, "/")
		if len(fibComps) >= 3 && fibComps[len(fibComps)-2] == "feature" {
			isInside := true
			if fibComps[len(fibComps)-1] == outFeature {
				isInside = false
			}
			iface := strings.Join(fibComps[:len(fibComps)-2], "/")
			return iface, isInside, true
		}
	}
	return "", false, false
}

/* NAT44 address (derived) */

// DerivedAddressNAT44Key returns (derived) key representing NAT44 configuration
// for a single IP address from the NAT44 address pool.
// Address is inserted into the key without validation!
func DerivedAddressNAT44Key(address string, twiceNat bool) string {
	key := strings.Replace(addressNAT44KeyTemplate, "{address}", address, 1)
	twiceNatFlag := twiceNatOff
	if twiceNat {
		twiceNatFlag = twiceNatOn
	}
	key = strings.Replace(key, "{twice-nat}", twiceNatFlag, 1)
	return key
}

// ParseDerivedAddressNAT44Key parses configuration of a single NAT44 address from a key
// returned by DerivedAddressNAT44Key().
func ParseDerivedAddressNAT44Key(key string) (address string, twiceNat bool, isAddressNAT44Key bool) {
	trim := strings.TrimPrefix(key, addressNAT44KeyPrefix)
	if trim != key && trim != "" {
		fibComps := strings.Split(trim, "/")
		if len(fibComps) >= 3 && fibComps[len(fibComps)-2] == "twice-nat" {
			if fibComps[len(fibComps)-1] == twiceNatOn {
				twiceNat = true
			}
			address = strings.Join(fibComps[:len(fibComps)-2], "/")
			isAddressNAT44Key = true
			return
		}
	}
	return
}

// DerivedTwiceNATAddressPoolKey returns (derived) key representing TwiceNAT address pool configuration.
func DerivedTwiceNATAddressPoolKey(firstIP, lastIP string, vrfID uint32) (key string) {
	if lastIP == "" {
		key = strings.Replace(twiceNATKeyTemplate, "{vrfID}", fmt.Sprint(vrfID), 1)
		key = strings.Replace(key, "{firstIP}", firstIP, 1)
	} else {
		key = strings.Replace(twiceNATWithMultipeAddressesKeyTemplate, "{vrfID}", fmt.Sprint(vrfID), 1)
		key = strings.Replace(key, "{firstIP}", firstIP, 1)
		key = strings.Replace(key, "{lastIP}", lastIP, 1)
	}
	return key
}

// ParseDerivedTwiceNATAddressPoolKey parses configuration of a twiceNAT address pool from a key
// returned by DerivedTwiceNATAddressPoolKey().
func ParseDerivedTwiceNATAddressPoolKey(key string) (firstIP, lastIP string, vrfID uint32, isTwiceNatKey bool) {
	trim := strings.TrimPrefix(key, TwiceNATDerivedKeyPrefix)
	if trim != key && trim != "" {
		comps := strings.Split(trim, "/")
		if len(comps) == 4 && comps[0] == "vrf" && comps[2] == "addresses" {
			vrfID64, err := strconv.ParseUint(comps[1], 10, 32)
			if err != nil {
				return
			}
			vrfID = uint32(vrfID64)
			addrComps := strings.Split(comps[3], "-")
			firstIP = addrComps[0]
			if len(addrComps) > 1 {
				lastIP = addrComps[1]
			}
			isTwiceNatKey = true
			return
		}
	}
	return
}
