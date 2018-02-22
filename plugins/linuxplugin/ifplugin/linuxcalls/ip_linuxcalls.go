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
	"bytes"
	"net"

	"time"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/vishvananda/netlink"
)

// AddInterfaceIP calls AddrAdd Netlink API.
func AddInterfaceIP(log logging.Logger, ifName string, addr *net.IPNet, timeLog measure.StopWatchEntry) error {
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}

	exAddrList, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return err
	}

	// The check is basically because of link local addresses which sometimes cannot be reassigned
	for ipIdx, exAddr := range exAddrList {
		if bytes.Compare(exAddr.IP, addr.IP) == 0 {
			log.Debugf("Cannot assign %v to interface %v, IP already exists", addr.IP.String(), ifName)
			// Remove the address from the pool
			exAddrList = append(exAddrList[:ipIdx], exAddrList[ipIdx+1:]...)
			continue
		}
	}

	address := &netlink.Addr{IPNet: addr}
	return netlink.AddrAdd(link, address)
}

// DelInterfaceIP calls AddrDel Netlink API.
func DelInterfaceIP(ifName string, addr *net.IPNet, timeLog measure.StopWatchEntry) error {
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	address := &netlink.Addr{IPNet: addr}
	return netlink.AddrDel(link, address)
}

// SetInterfaceMTU calls LinkSetMTU Netlink API.
func SetInterfaceMTU(ifName string, mtu int, timeLog measure.StopWatchEntry) error {
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	return netlink.LinkSetMTU(link, mtu)
}
