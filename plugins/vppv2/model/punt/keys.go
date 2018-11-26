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

package punt

import (
	"strconv"
	"strings"
)

/* Punt */
const (
	// PrefixNAT44 is a key prefix used in NB DB to store configuration for NAT44.
	Prefix = "vpp/config/v2/punt/"
)

/* Punt to Host */
const (
	// PrefixToHost is the key used in NB DB to store punt to host or unix domain socket configuration.
	PrefixToHost = Prefix + "tohost/"

	// toHostTemplate is the relative key prefix for punt to host.
	toHostTemplate = PrefixToHost + "l3/{l3}/l4/{l4}/port/{port}"
)

/* IP redirect */
const (
	// PrefixIPRedirect is a key prefix used in NB DB to store punt IP redirect configuration.
	PrefixIPRedirect = Prefix + "ipredirect/"

	// ipRedirectTemplate is the relative key prefix for IP redirect.
	ipRedirectTemplate = PrefixIPRedirect + "l3/{l3}/tx/{tx}"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

/* Punt to host */

// ToHostKey returns key representing punt to host/socket configuration.
func ToHostKey(l3Proto L3Protocol, l4Proto ToHost_L4Protocol, port uint32) string {
	strL3, strL4, strPort := strconv.Itoa(int(l3Proto)), strconv.Itoa(int(l4Proto)), strconv.Itoa(int(port))
	if port == 0 {
		strPort = InvalidKeyPart
	}
	key := strings.Replace(toHostTemplate, "{l3}", strL3, 1)
	key = strings.Replace(key, "{l4}", strL4, 1)
	key = strings.Replace(key, "{port}", strPort, 1)

	return key
}

// ParsePuntToHostKey parses L3 and L4 protocol and port from Punt-to-Host key.
func ParsePuntToHostKey(key string) (l3ProtoIndex, l4ProtoIndex, port string, isPuntToHostKey bool) {
	if strings.HasPrefix(key, PrefixToHost) {
		keySuffix := strings.TrimPrefix(key, PrefixToHost)
		puntComps := strings.Split(keySuffix, "/")
		if len(puntComps) == 6 {
			if _, err := strconv.Atoi(puntComps[1]); err != nil {
				puntComps[1] = InvalidKeyPart
			}
			if _, err := strconv.Atoi(puntComps[3]); err != nil {
				puntComps[3] = InvalidKeyPart
			}
			if _, err := strconv.Atoi(puntComps[5]); err != nil {
				puntComps[5] = InvalidKeyPart
			}
			return puntComps[1], puntComps[3], puntComps[5], true
		}
	}
	return "", "", "", false
}

/* IP punt redirect */

// IPRedirectKey returns key representing IP punt redirect configuration.
func IPRedirectKey(l3Proto L3Protocol, txIf string) string {
	strL3 := strconv.Itoa(int(l3Proto))
	if txIf == "" {
		txIf = InvalidKeyPart
	}
	key := strings.Replace(ipRedirectTemplate, "{l3}", strL3, 1)
	key = strings.Replace(key, "{tx}", txIf, 1)

	return key
}

// ParseIPRedirectKey parses L3 and L4 protocol and port from Punt-to-Host key.
func ParseIPRedirectKey(key string) (l3Proto, txIf string, isIPRedirectKey bool) {
	if strings.HasPrefix(key, PrefixIPRedirect) {
		keySuffix := strings.TrimPrefix(key, PrefixIPRedirect)
		puntComps := strings.Split(keySuffix, "/")
		if len(puntComps) == 4 {
			return puntComps[1], puntComps[3], true
		}
	}
	return "", "", false
}
