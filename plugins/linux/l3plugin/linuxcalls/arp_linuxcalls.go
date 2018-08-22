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
	"time"

	"github.com/vishvananda/netlink"
)

// AddArpEntry creates a new static ARP entry
func (handler *NetLinkHandler) AddArpEntry(name string, arpEntry *netlink.Neigh) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("add-arp-entry").LogTimeEntry(time.Since(t))
	}(time.Now())

	return netlink.NeighAdd(arpEntry)
}

// SetArpEntry updates existing arp entry
func (handler *NetLinkHandler) SetArpEntry(name string, arpEntry *netlink.Neigh) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("set-arp-entry").LogTimeEntry(time.Since(t))
	}(time.Now())

	return netlink.NeighSet(arpEntry)
}

// DelArpEntry removes an static ARP entry
func (handler *NetLinkHandler) DelArpEntry(name string, arpEntry *netlink.Neigh) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("del-arp-entry").LogTimeEntry(time.Since(t))
	}(time.Now())

	return netlink.NeighDel(arpEntry)
}

// GetArpEntries reads all configured static ARP entries for given interface
// <interfaceIdx> and <family> parameters works as filters, if they are set to zero, all arp entries are returned
func (handler *NetLinkHandler) GetArpEntries(interfaceIdx int, family int) ([]netlink.Neigh, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("get-arp-entries").LogTimeEntry(time.Since(t))
	}(time.Now())

	return netlink.NeighList(interfaceIdx, family)
}
