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

package vpp1901

import (
	"errors"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/tap"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/tapv2"
)

// AddTapInterface implements interface handler.
func (h *InterfaceVppHandler) AddTapInterface(ifName string, tapIf *interfaces.TapLink) (swIfIdx uint32, err error) {
	if tapIf == nil || tapIf.HostIfName == "" {
		return 0, errors.New("host interface name was not provided for the TAP interface")
	}

	if tapIf.Version == 2 {
		if tapIf.EnableGso {
			h.log.Warnf("GSO feature for TAP interface is not supported in VPP 19.01")
		}

		// Configure fast virtio-based TAP interface
		req := &tapv2.TapCreateV2{
			ID:            ^uint32(0),
			HostIfName:    []byte(tapIf.HostIfName),
			HostIfNameSet: 1,
			UseRandomMac:  1,
			RxRingSz:      uint16(tapIf.RxRingSize),
			TxRingSz:      uint16(tapIf.TxRingSize),
		}

		reply := &tapv2.TapCreateV2Reply{}
		err = h.callsChannel.SendRequest(req).ReceiveReply(reply)
		swIfIdx = reply.SwIfIndex
	} else {
		// Configure the original TAP interface
		req := &tap.TapConnect{
			TapName:      []byte(tapIf.HostIfName),
			UseRandomMac: 1,
		}

		reply := &tap.TapConnectReply{}
		err = h.callsChannel.SendRequest(req).ReceiveReply(reply)
		swIfIdx = reply.SwIfIndex
	}
	if err != nil {
		return 0, err
	}

	return swIfIdx, h.SetInterfaceTag(ifName, swIfIdx)
}

// DeleteTapInterface implements interface handler.
func (h *InterfaceVppHandler) DeleteTapInterface(ifName string, idx uint32, version uint32) error {
	var err error

	if version == 2 {
		req := &tapv2.TapDeleteV2{
			SwIfIndex: idx,
		}

		reply := &tapv2.TapDeleteV2Reply{}
		err = h.callsChannel.SendRequest(req).ReceiveReply(reply)
	} else {
		req := &tap.TapDelete{
			SwIfIndex: idx,
		}

		reply := &tap.TapDeleteReply{}
		err = h.callsChannel.SendRequest(req).ReceiveReply(reply)
	}
	if err != nil {
		return err
	}

	return h.RemoveInterfaceTag(ifName, idx)
}
