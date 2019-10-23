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
	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_ifs "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001/interfaces"
)

func (h *InterfaceVppHandler) SetRxMode(ifIdx uint32, rxMode *ifs.Interface_RxMode) error {

	req := &vpp_ifs.SwInterfaceSetRxMode{
		SwIfIndex:    vpp_ifs.InterfaceIndex(ifIdx),
		Mode:         setRxMode(rxMode.Mode),
		QueueID:      rxMode.Queue,
		QueueIDValid: !rxMode.DefaultMode,
	}
	reply := &vpp_ifs.SwInterfaceSetRxModeReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func setRxMode(mode ifs.Interface_RxMode_Type) vpp_ifs.RxMode {
	switch mode {
	case ifs.Interface_RxMode_POLLING:
		return vpp_ifs.RX_MODE_API_POLLING
	case ifs.Interface_RxMode_INTERRUPT:
		return vpp_ifs.RX_MODE_API_INTERRUPT
	case ifs.Interface_RxMode_ADAPTIVE:
		return vpp_ifs.RX_MODE_API_ADAPTIVE
	case ifs.Interface_RxMode_DEFAULT:
		return vpp_ifs.RX_MODE_API_DEFAULT
	default:
		return vpp_ifs.RX_MODE_API_UNKNOWN
	}
}
