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
	// Prefix is a key prefix used in NB DB to store configuration for punt.
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
func ToHostKey(l3Proto L3Protocol, l4Proto L4Protocol, port uint32) string {
	strL3, ok := L3Protocol_name[int32(l3Proto)]
	if !ok {
		strL3 = InvalidKeyPart
	}
	strL4, ok := L4Protocol_name[int32(l4Proto)]
	if !ok {
		strL3 = InvalidKeyPart
	}
	strPort := strconv.Itoa(int(port))
	if port == 0 {
		strPort = InvalidKeyPart
	}
	key := strings.Replace(toHostTemplate, "{l3}", strL3, 1)
	key = strings.Replace(key, "{l4}", strL4, 1)
	key = strings.Replace(key, "{port}", strPort, 1)

	return key
}

// ParsePuntToHostKey parses L3 and L4 protocol and port from Punt-to-Host key.
func ParsePuntToHostKey(key string) (l3Proto L3Protocol, l4Proto L4Protocol, port uint32, isPuntToHostKey bool) {
	if strings.HasPrefix(key, PrefixToHost) {
		keySuffix := strings.TrimPrefix(key, PrefixToHost)
		puntComps := strings.Split(keySuffix, "/")
		if len(puntComps) == 6 {
			l3Proto := L3Protocol_value[puntComps[1]]
			l4Proto := L4Protocol_value[puntComps[3]]
			keyPort, err := strconv.Atoi(puntComps[5])
			if err != nil {
				// Keep port at zero value
			}
			return L3Protocol(l3Proto), L4Protocol(l4Proto), uint32(keyPort), true
		}
	}
	return L3Protocol_UNDEFINED_L3, L4Protocol_UNDEFINED_L4, 0, false
}

/* IP punt redirect */

// IPRedirectKey returns key representing IP punt redirect configuration.
func IPRedirectKey(l3Proto L3Protocol, txIf string) string {
	strL3, ok := L3Protocol_name[int32(l3Proto)]
	if !ok {
		strL3 = InvalidKeyPart
	}
	if txIf == "" {
		txIf = InvalidKeyPart
	}
	key := strings.Replace(ipRedirectTemplate, "{l3}", strL3, 1)
	key = strings.Replace(key, "{tx}", txIf, 1)

	return key
}

// ParseIPRedirectKey parses L3 and L4 protocol and port from Punt-to-Host key.
func ParseIPRedirectKey(key string) (l3Proto L3Protocol, txIf string, isIPRedirectKey bool) {
	if strings.HasPrefix(key, PrefixIPRedirect) {
		keySuffix := strings.TrimPrefix(key, PrefixIPRedirect)
		puntComps := strings.Split(keySuffix, "/")
		if len(puntComps) == 4 {
			l3Proto := L3Protocol_value[puntComps[1]]
			return L3Protocol(l3Proto), puntComps[3], true
		}
	}
	return L3Protocol_UNDEFINED_L3, "", false
}
