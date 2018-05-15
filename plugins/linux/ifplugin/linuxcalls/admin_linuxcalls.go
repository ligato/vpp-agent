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

	"github.com/ligato/cn-infra/logging/measure"
	"github.com/vishvananda/netlink"
)

// InterfaceAdminDown calls Netlink API LinkSetDown.
func InterfaceAdminDown(ifName string, timeLog measure.StopWatchEntry) error {
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
	return netlink.LinkSetDown(link)
}

// InterfaceAdminUp calls Netlink API LinkSetUp.
func InterfaceAdminUp(ifName string, timeLog measure.StopWatchEntry) error {
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
	return netlink.LinkSetUp(link)
}
