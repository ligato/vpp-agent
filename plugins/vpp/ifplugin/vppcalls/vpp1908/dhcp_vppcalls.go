//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package vpp1908

import (
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/dhcp"
)

func (h *InterfaceVppHandler) handleInterfaceDHCP(ifIdx uint32, hostName string, isAdd bool) error {
	req := &dhcp.DHCPClientConfig{
		IsAdd: boolToUint(isAdd),
		Client: dhcp.DHCPClient{
			SwIfIndex:     ifIdx,
			Hostname:      []byte(hostName),
			WantDHCPEvent: 1,
		},
	}
	reply := &dhcp.DHCPClientConfigReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// SetInterfaceAsDHCPClient implements interface handler.
func (h *InterfaceVppHandler) SetInterfaceAsDHCPClient(ifIdx uint32, hostName string) error {
	return h.handleInterfaceDHCP(ifIdx, hostName, true)
}

// UnsetInterfaceAsDHCPClient implements interface handler.
func (h *InterfaceVppHandler) UnsetInterfaceAsDHCPClient(ifIdx uint32, hostName string) error {
	return h.handleInterfaceDHCP(ifIdx, hostName, false)
}
