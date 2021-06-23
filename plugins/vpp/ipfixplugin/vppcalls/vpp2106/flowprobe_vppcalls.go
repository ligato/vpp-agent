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

package vpp2106

import (
	"github.com/pkg/errors"

	vpp_flowprobe "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/flowprobe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

// SetFPParams sends message with configuration for Flowprobe.
func (h *IpfixVppHandler) SetFPParams(conf *ipfix.FlowProbeParams) error {
	var flags vpp_flowprobe.FlowprobeRecordFlags

	if conf.GetRecordL2() {
		flags |= vpp_flowprobe.FLOWPROBE_RECORD_FLAG_L2
	}

	if conf.GetRecordL3() {
		flags |= vpp_flowprobe.FLOWPROBE_RECORD_FLAG_L3
	}

	if conf.GetRecordL4() {
		flags |= vpp_flowprobe.FLOWPROBE_RECORD_FLAG_L4
	}

	if flags == 0 {
		err := errors.New("one of the record fields (l2, l3 or l4) must be enabled")
		return err
	}

	req := &vpp_flowprobe.FlowprobeParams{
		RecordFlags:  flags,
		ActiveTimer:  conf.GetActiveTimer(),
		PassiveTimer: conf.GetPassiveTimer(),
	}
	reply := &vpp_flowprobe.FlowprobeParamsReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *IpfixVppHandler) sendFPFeature(isAdd bool, conf *ipfix.FlowProbeFeature) error {
	meta, found := h.ifIndexes.LookupByName(conf.Interface)
	if !found {
		return errors.Errorf("interface %s not found", conf.Interface)
	}

	var flags vpp_flowprobe.FlowprobeWhichFlags

	if conf.GetL2() {
		flags |= vpp_flowprobe.FLOWPROBE_WHICH_FLAG_L2
	}

	if conf.GetIp4() {
		flags |= vpp_flowprobe.FLOWPROBE_WHICH_FLAG_IP4
	}

	if conf.GetIp6() {
		flags |= vpp_flowprobe.FLOWPROBE_WHICH_FLAG_IP6
	}

	req := &vpp_flowprobe.FlowprobeTxInterfaceAddDel{
		IsAdd:     isAdd,
		Which:     flags,
		SwIfIndex: interface_types.InterfaceIndex(meta.SwIfIndex),
	}
	reply := &vpp_flowprobe.FlowprobeTxInterfaceAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// AddFPFeature sends message to enable Flowprobe on interface.
func (h *IpfixVppHandler) AddFPFeature(conf *ipfix.FlowProbeFeature) error {
	return h.sendFPFeature(true, conf)
}

// DelFPFeature sends message to disable Flowprobe on interface.
func (h *IpfixVppHandler) DelFPFeature(conf *ipfix.FlowProbeFeature) error {
	return h.sendFPFeature(false, conf)
}
