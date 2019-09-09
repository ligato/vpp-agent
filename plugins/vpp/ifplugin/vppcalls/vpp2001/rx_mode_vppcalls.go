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

package vpp2001

import (
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	binapi_interface "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001/interfaces"
)

// SetRxMode implements interface handler.
func (h *InterfaceVppHandler) SetRxMode(ifIdx uint32, rxMode *interfaces.Interface_RxMode) error {

	req := &binapi_interface.SwInterfaceSetRxMode{
		SwIfIndex:    binapi_interface.InterfaceIndex(ifIdx),
		Mode:         setRxMode(rxMode.Mode),
		QueueID:      rxMode.Queue,
		QueueIDValid: !rxMode.DefaultMode,
	}
	reply := &binapi_interface.SwInterfaceSetRxModeReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func setRxMode(mode interfaces.Interface_RxMode_Type) binapi_interface.RxMode {
	switch mode {
	case interfaces.Interface_RxMode_POLLING:
		return binapi_interface.RX_MODE_API_POLLING
	case interfaces.Interface_RxMode_INTERRUPT:
		return binapi_interface.RX_MODE_API_INTERRUPT
	case interfaces.Interface_RxMode_ADAPTIVE:
		return binapi_interface.RX_MODE_API_ADAPTIVE
	case interfaces.Interface_RxMode_DEFAULT:
		return binapi_interface.RX_MODE_API_DEFAULT
	default:
		return binapi_interface.RX_MODE_API_UNKNOWN
	}
}
