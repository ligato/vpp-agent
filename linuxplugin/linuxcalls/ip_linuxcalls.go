// +build !windows,!darwin

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

package linuxcalls

import (
	"net"

	"github.com/vishvananda/netlink"
)

// AddInterfaceIP calls AddrAdd Netlink API
func AddInterfaceIP(ifName string, addr *net.IPNet) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	address := &netlink.Addr{IPNet: addr}
	return netlink.AddrAdd(link, address)
}

// DelInterfaceIP calls AddrDel Netlink API
func DelInterfaceIP(ifName string, addr *net.IPNet) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	address := &netlink.Addr{IPNet: addr}
	return netlink.AddrDel(link, address)
}

// SetInterfaceMTU calls LinkSetMTU Netlink API
func SetInterfaceMTU(ifName string, mtu int) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	return netlink.LinkSetMTU(link, mtu)
}
