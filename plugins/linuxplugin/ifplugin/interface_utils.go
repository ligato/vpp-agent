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

package ifplugin

import (
	"net"
)

// GetLinuxInterfaceIndex returns the index of a Linux interface identified by its name.
// In Linux, interface index is a positive integer that starts at one, zero is never used.
// Function returns negative number in case of a failure, such as when the interface doesn't exist.
// TODO: move to the package with network utilities
func GetLinuxInterfaceIndex(ifName string) int {
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return -1
	}
	return iface.Index
}
