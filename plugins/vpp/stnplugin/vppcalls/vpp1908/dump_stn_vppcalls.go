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
	"net"

	"github.com/ligato/vpp-agent/plugins/vpp/stnplugin/vppcalls"
	"github.com/pkg/errors"

	stn "github.com/ligato/vpp-agent/api/models/vpp/stn"
	api "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/stn"
)

// DumpSTNRules implements STN handler, it returns all STN rules present on the VPP
func (h *StnVppHandler) DumpSTNRules() ([]*vppcalls.StnDetails, error) {
	var stnDetails []*vppcalls.StnDetails

	req := &api.StnRulesDump{}
	reqCtx := h.callsChannel.SendMultiRequest(req)
	for {
		msg := &api.StnRulesDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return nil, errors.Errorf("error reading STN rules from the VPP: %v", err)
		}
		ifName, _, found := h.ifIndexes.LookupBySwIfIndex(msg.SwIfIndex)
		if !found {
			h.log.Warnf("STN dump: interface name not found for index %d", msg.SwIfIndex)
		}

		var stnIP string
		if uintToBool(msg.IsIP4) {
			stnIP = net.IP(msg.IPAddress[:4]).To4().String()
		} else {
			stnIP = net.IP(msg.IPAddress).To16().String()
		}

		stnRule := &stn.Rule{
			IpAddress: stnIP,
			Interface: ifName,
		}
		stnMeta := &vppcalls.StnMeta{
			IfIdx: msg.SwIfIndex,
		}

		stnDetails = append(stnDetails, &vppcalls.StnDetails{
			Rule: stnRule,
			Meta: stnMeta,
		})
	}

	return stnDetails, nil
}

func uintToBool(value uint8) bool {
	if value == 0 {
		return false
	}
	return true
}
