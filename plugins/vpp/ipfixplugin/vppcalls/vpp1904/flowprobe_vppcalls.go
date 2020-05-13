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
	"github.com/pkg/errors"

	vpp_flowprobe "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/flowprobe"
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

// SetFPParams sends message with configuration for Flowprobe.
func (h *IpfixVppHandler) SetFPParams(conf *ipfix.FlowProbeParams) error {
	var rL2, rL3, rL4 uint8

	if conf.GetRecordL2() {
		rL2 = 1
	}
	if conf.GetRecordL3() {
		rL3 = 1
	}
	if conf.GetRecordL4() {
		rL4 = 1
	}

	if rL2 != 1 && rL3 != 1 && rL4 != 1 {
		err := errors.New("one of the record fields (l2, l3 or l4) must be enabled")
		return err
	}

	req := &vpp_flowprobe.FlowprobeParams{
		RecordL2:     rL2,
		RecordL3:     rL3,
		RecordL4:     rL4,
		ActiveTimer:  conf.GetActiveTimer(),
		PassiveTimer: conf.GetPassiveTimer(),
	}
	reply := &vpp_flowprobe.FlowprobeParamsReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *IpfixVppHandler) sendFPFeature(isAdd uint8, conf *ipfix.FlowProbeFeature) error {
	meta, found := h.ifIndexes.LookupByName(conf.Interface)
	if !found {
		return errors.Errorf("interface %s not found", conf.Interface)
	}

	var flags uint8

	if conf.GetL2() {
		flags |= 0b001
	}
	if conf.GetIp4() {
		flags |= 0b010
	}
	if conf.GetIp6() {
		flags |= 0b100
	}

	req := &vpp_flowprobe.FlowprobeTxInterfaceAddDel{
		IsAdd:     isAdd,
		Which:     flags,
		SwIfIndex: meta.SwIfIndex,
	}
	reply := &vpp_flowprobe.FlowprobeTxInterfaceAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// AddFPFeature sends message to enable Flowprobe on interface.
func (h *IpfixVppHandler) AddFPFeature(conf *ipfix.FlowProbeFeature) error {
	return h.sendFPFeature(1, conf)
}

// DelFPFeature sends message to disable Flowprobe on interface.
func (h *IpfixVppHandler) DelFPFeature(conf *ipfix.FlowProbeFeature) error {
	return h.sendFPFeature(0, conf)
}
