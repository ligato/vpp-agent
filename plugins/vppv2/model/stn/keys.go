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

package stn

import (
	"strings"
)

const (
	// Prefix is STN key prefix
	Prefix = "vpp/config/v2/stn/rule/"
)

const (
	// keyTemplate is a template for key representing configuration for a STN rule.
	keyTemplate = Prefix + "{iface}/ip/{ip}"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

// Key returns the prefix used in the ETCD to store a VPP STN config
// of a particular STN rule in selected VPP instance.
func Key(ifName, ipAddress string) string {
	if ifName == "" {
		ifName = InvalidKeyPart
	}
	if ipAddress == "" || len(strings.Split(ipAddress, "/")) != 1 {
		ipAddress = InvalidKeyPart
	}
	key := strings.Replace(keyTemplate, "{iface}", ifName, 1)
	key = strings.Replace(key, "{ip}", ipAddress, 1)
	return key
}

// ParseKey parses interface name and IP address from a STN key.
func ParseKey(key string) (ifName, ipAddress string, isSTNKey bool) {
	if strings.HasPrefix(key, Prefix) {
		stnSuffix := strings.TrimPrefix(key, Prefix)
		stnComps := strings.Split(stnSuffix, "/")
		if len(stnComps) == 3 && stnComps[1] == "ip" {
			return stnComps[0], stnComps[2], true
		}
	}
	return "", "", false
}