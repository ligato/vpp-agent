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

package vpp1904

import (
	"fmt"

	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// AddIpipTunnel adds new IPIP tunnel interface.
func (h *InterfaceVppHandler) AddIpipTunnel(ifName string, vrf uint32, ipipLink *interfaces.IPIPLink) (uint32, error) {
	return 0, fmt.Errorf("IPIP interface unsupported in VPP 1904")
}

// DelIpipTunnel removes IPIP tunnel interface.
func (h *InterfaceVppHandler) DelIpipTunnel(ifName string, ifIdx uint32) error {
	return fmt.Errorf("IPIP interface unsupported in VPP 1904")
}
