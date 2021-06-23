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

package vpp2106

import (
	"net"

	"github.com/go-errors/errors"

	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
)

func (h *InterfaceVppHandler) SetInterfaceMac(ifIdx uint32, macAddress string) error {
	mac, err := ParseMAC(macAddress)
	if err != nil {
		return err
	}
	req := &vpp_ifs.SwInterfaceSetMacAddress{
		SwIfIndex:  interface_types.InterfaceIndex(ifIdx),
		MacAddress: mac,
	}
	reply := &vpp_ifs.SwInterfaceSetMacAddressReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// parse mac string to uint8 array with len=6
func ParseMAC(mac string) (parsedMac [6]uint8, err error) {
	var hw net.HardwareAddr
	if hw, err = net.ParseMAC(mac); err != nil {
		err = errors.Errorf("invalid mac address %s: %v", mac, err)
		return
	}
	copy(parsedMac[:], hw[:])
	return
}
