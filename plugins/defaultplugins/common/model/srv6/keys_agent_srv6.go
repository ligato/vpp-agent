// Copyright (c) 2018 Bell Canada, Pantheon Technologies and/or its affiliates.
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

package srv6

import (
	"net"

	"fmt"
	"regexp"
	"strings"
)

// Keys and prefixes(to keys) used for SRv6 in ETCD key-value store
const (
	basePrefix     = "vpp/config/v1/srv6/"
	localSIDPrefix = basePrefix + "localsid/" // full key is in form .../localsid/{sid}
	policyPrefix   = basePrefix + "policy/"   // full key is in form .../policy/{bsid}
	steeringPrefix = basePrefix + "steering/" // full key is in form .../steering/{name}
)

var policySegmentPrefixRegExp = regexp.MustCompile(policyPrefix + "([^/]+)/segment/") // full key is in form .../policy/{bsid}/segment/{name}

// EtcdKeyPathDelimiter is delimiter used in ETCD keys and can be used to combine multiple etcd key parts together
// (without worry that key part has accidentally this delimiter because otherwise it would not be one key part)
const EtcdKeyPathDelimiter = "/"

// SID (in srv6 package) is SRv6's segment id. It is always represented as IPv6 address
type SID = net.IP

// BasePrefix returns the prefix used in ETCD to store vpp SRv6 config.
func BasePrefix() string {
	return basePrefix
}

// LocalSIDPrefix returns longest common prefix for all local SID keys
func LocalSIDPrefix() string {
	return localSIDPrefix
}

// PolicyPrefix returns longest common prefix for all policy keys
func PolicyPrefix() string {
	return policyPrefix
}

// IsPolicySegmentPrefix check whether key has policy segment prefix
func IsPolicySegmentPrefix(key string) bool {
	return policySegmentPrefixRegExp.MatchString(key)
}

// SteeringPrefix returns longest common prefix for all steering keys
func SteeringPrefix() string {
	return steeringPrefix
}

// ParseLocalSIDKey parses SID from a key.
func ParseLocalSIDKey(key string) (SID, error) {
	sidStr, err := parseOneValuedKey(key, LocalSIDPrefix(), "sid")
	if err != nil {
		return nil, err
	}
	return parseIPv6(sidStr)
}

// ParsePolicyKey parses BSID from a key.
func ParsePolicyKey(key string) (net.IP, error) {
	sidStr, err := parseOneValuedKey(key, PolicyPrefix(), "bsid")
	if err != nil {
		return nil, err
	}
	return parseIPv6(sidStr)
}

// ParsePolicySegmentKey parses BSID of policy where policy segment belongs to and name of policy segment.
func ParsePolicySegmentKey(key string) (net.IP, string, error) {
	if !policySegmentPrefixRegExp.MatchString(key) {
		return nil, "", fmt.Errorf("key %v is not policy segment key", key)
	}
	suffix := strings.TrimPrefix(key, policyPrefix)
	keyComponents := strings.Split(suffix, EtcdKeyPathDelimiter)
	if len(keyComponents) != 3 {
		return nil, "", fmt.Errorf("key \"%v\" should have policy BSID and policy segment name", key)
	}
	bsid, err := parseIPv6(keyComponents[0])
	if err != nil {
		return nil, "", fmt.Errorf("can't parse \"%v\" into SRv6 BSID (IPv6 address)", keyComponents[0])
	}
	return bsid, keyComponents[2], nil
}

func parseOneValuedKey(key string, prefix string, valueName string) (value string, err error) {
	if !strings.HasPrefix(key, prefix) {
		return "", fmt.Errorf("key \"%v\" should have prefix \"%v\"", key, prefix)
	}
	suffix := strings.TrimPrefix(key, prefix)
	keyComponents := strings.Split(suffix, EtcdKeyPathDelimiter)
	if len(keyComponents) != 1 {
		return "", fmt.Errorf("key \"%v\" should have %v (and only %v) after \"%v\"", key, valueName, valueName, prefix)
	}
	return keyComponents[0], nil
}

// parseIPv6 parses string <str> to IPv6 address (including IPv4 address converted to IPv6 address)
func parseIPv6(str string) (net.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return nil, fmt.Errorf("\"%v\" is not ip address", str)
	}
	ipv6 := ip.To16()
	if ipv6 == nil {
		return nil, fmt.Errorf("\"%v\" is not ipv6 address", str)
	}
	return ipv6, nil
}
