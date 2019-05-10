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

package vpp1904

import (
	"errors"
	"fmt"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/tapv2"
)

// TapFlags definitions from https://github.com/FDio/vpp/blob/stable/1904/src/vnet/devices/tap/tap.h#L33
const (
	TapFlagGSO uint32 = 1 << iota
)

// AddTapInterface implements interface handler.
func (h *InterfaceVppHandler) AddTapInterface(ifName string, tapIf *interfaces.TapLink) (swIfIdx uint32, err error) {
	if tapIf == nil || tapIf.HostIfName == "" {
		return 0, errors.New("host interface name was not provided for the TAP interface")
	}

	if tapIf.Version == 1 {
		return 0, errors.New("tap version 1 has been deprecated")
	} else if tapIf.Version == 2 {
		var flags uint32
		if tapIf.EnableGso {
			flags |= TapFlagGSO
		}

		// Configure fast virtio-based TAP interface
		req := &tapv2.TapCreateV2{
			ID:            ^uint32(0),
			HostIfName:    []byte(tapIf.HostIfName),
			HostIfNameSet: 1,
			UseRandomMac:  1,
			RxRingSz:      uint16(tapIf.RxRingSize),
			TxRingSz:      uint16(tapIf.TxRingSize),
			TapFlags:      flags,
		}

		reply := &tapv2.TapCreateV2Reply{}
		if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
			return 0, err
		}
		swIfIdx = reply.SwIfIndex
	} else {
		return 0, fmt.Errorf("invalid tap version (%v)", tapIf.Version)
	}

	return swIfIdx, h.SetInterfaceTag(ifName, swIfIdx)
}

// DeleteTapInterface implements interface handler.
func (h *InterfaceVppHandler) DeleteTapInterface(ifName string, idx uint32, version uint32) error {
	if version == 1 {
		return errors.New("tap version 1 has been deprecated")
	} else if version == 2 {
		req := &tapv2.TapDeleteV2{
			SwIfIndex: idx,
		}

		reply := &tapv2.TapDeleteV2Reply{}
		if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid tap version (%v)", version)
	}

	return h.RemoveInterfaceTag(ifName, idx)
}
