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
	"fmt"

	"github.com/ligato/cn-infra/logging"
	"github.com/vishvananda/netlink"
	"github.com/ligato/cn-infra/logging/logrus"
	"time"
	"github.com/ligato/cn-infra/logging/timer"
)

// AddVethInterface calls LinkAdd Netlink API for the Netlink.Veth interface type.
func AddVethInterface(ifName, peerIfName string, log logging.Logger, stopwatch *timer.Stopwatch) error {
	log.WithFields(logging.Fields{"ifName": ifName, "peerIfName": peerIfName}).Debug("Creating new Linux VETH pair")
	start := time.Now()
	defer func() {
		if stopwatch != nil {
			stopwatch.LogTimeEntry("add_veth_iface", time.Since(start))
		}
	}()

	// Veth pair params
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:   ifName,
			TxQLen: 0,
		},
		PeerName: peerIfName,
	}

	// Create the veth pair
	err := netlink.LinkAdd(veth)
	return err
}

// DelVethInterface calls LinkDel Netlink API for the Netlink.Veth interface type.
func DelVethInterface(ifName, peerIfName string, log logging.Logger, stopwatch *timer.Stopwatch) error {
	log.WithFields(logging.Fields{"ifName": ifName, "peerIfName": peerIfName}).Debug("Deleting Linux VETH pair")
	start := time.Now()
	defer func() {
		if stopwatch != nil {
			stopwatch.LogTimeEntry("del_veth_iface", time.Since(start))
		}
	}()

	// Veth pair params
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:   ifName,
			TxQLen: 0,
		},
		PeerName: peerIfName,
	}

	// Create the veth pair
	err := netlink.LinkDel(veth)
	return err
}

// GetVethPeerName return the peer name for a given VETH interface.
func GetVethPeerName(ifName string, stopwatch *timer.Stopwatch) (string, error) {
	start := time.Now()
	defer func() {
		if stopwatch != nil {
			stopwatch.LogTimeEntry("get_veth_peer", time.Since(start))
		}
	}()

	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return "", err
	}
	veth, isVeth := link.(*netlink.Veth)
	if !isVeth {
		return "", fmt.Errorf("interface '%s' is not VETH", ifName)
	}
	return veth.PeerName, nil
}
