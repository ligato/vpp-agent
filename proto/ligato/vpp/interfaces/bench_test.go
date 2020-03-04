//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package vpp_interfaces

import (
	"testing"
)

var (
	riface         string
	rfromIface     string
	rvrf           int
	ripv4          bool
	ripv6          bool
	risIfaceVrfKey bool
)

func BenchmarkParseInterfaceVrfKey(b *testing.B) {
	var key = "vpp/interface/memif0/vrf/0/ip-version/v4"
	var (
		iface         string
		vrf           int
		ipv4          bool
		ipv6          bool
		isIfaceVrfKey bool
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iface, vrf, ipv4, ipv6, isIfaceVrfKey = ParseInterfaceVrfKey(key)
	}
	riface, rvrf, ripv4, ripv6, risIfaceVrfKey = iface, vrf, ipv4, ipv6, isIfaceVrfKey
}

func BenchmarkParseInterfaceVrfKeyByte(b *testing.B) {
	var key = []byte("vpp/interface/memif0/vrf/0/ip-version/v4")
	var (
		iface         string // []byte
		vrf           int
		ipv4          bool
		ipv6          bool
		isIfaceVrfKey bool
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iface, vrf, ipv4, ipv6, isIfaceVrfKey = ParseInterfaceVrfKeyByte(key)
	}
	_, rvrf, ripv4, ripv6, risIfaceVrfKey = string(iface), vrf, ipv4, ipv6, isIfaceVrfKey
}

func BenchmarkParseInterfaceInheritedVrfKey(b *testing.B) {
	var key = "vpp/interface/memif0/vrf/from-interface/memif1"
	var (
		iface         string
		fromIface     string
		isIfaceVrfKey bool
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iface, fromIface, isIfaceVrfKey = ParseInterfaceInheritedVrfKey(key)
	}
	riface, rfromIface, risIfaceVrfKey = iface, fromIface, isIfaceVrfKey
}

func BenchmarkParseInterfaceInheritedVrfKeyByte(b *testing.B) {
	var key = []byte("vpp/interface/memif0/vrf/from-interface/memif1")
	var (
		iface         string
		fromIface     string
		isIfaceVrfKey bool
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iface, fromIface, isIfaceVrfKey = ParseInterfaceInheritedVrfKeyByte(key)
	}
	riface, rfromIface, risIfaceVrfKey = iface, fromIface, isIfaceVrfKey
}
