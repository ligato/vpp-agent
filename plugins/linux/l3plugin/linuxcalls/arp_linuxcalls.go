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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/vishvananda/netlink"
	"time"
)

// AddArpEntry creates a new static ARP entry
func AddArpEntry(name string, arpEntry *netlink.Neigh, log logging.Logger, timeLog measure.StopWatchEntry) error {
	log.Debugf("Creating new ARP entry %v", name)
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return netlink.NeighAdd(arpEntry)
}

// ModifyArpEntry updates existing arp entry
func ModifyArpEntry(name string, arpEntry *netlink.Neigh, log logging.Logger, timeLog measure.StopWatchEntry) error {
	log.Debugf("Deleting an ARP entry %v", name)
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return netlink.NeighSet(arpEntry)
}

// DeleteArpEntry removes an static ARP entry
func DeleteArpEntry(name string, arpEntry *netlink.Neigh, log logging.Logger, timeLog measure.StopWatchEntry) error {
	log.Debugf("Deleting an ARP entry %v", name)
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return netlink.NeighDel(arpEntry)
}

// ReadArpEntries reads all configured static ARP entries for given interface
// <interfaceIdx> and <family> parameters works as filters, if they are set to zero, all arp entries are returned
func ReadArpEntries(interfaceIdx int, family int, log logging.Logger, timeLog measure.StopWatchEntry) ([]netlink.Neigh, error) {
	log.Debugf("Reading ARP entries")
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return netlink.NeighList(interfaceIdx, family)
}
