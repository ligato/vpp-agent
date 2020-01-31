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
	"net"

	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/vppcalls"

	vpp_stn "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/stn"
	stn "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/stn"
)

// DumpSTNRules implements STN handler, it returns all STN rules present on the VPP
func (h *StnVppHandler) DumpSTNRules() ([]*vppcalls.StnDetails, error) {
	var stnDetails []*vppcalls.StnDetails

	req := &vpp_stn.StnRulesDump{}
	reqCtx := h.callsChannel.SendMultiRequest(req)
	for {
		msg := &vpp_stn.StnRulesDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return nil, errors.Errorf("error reading STN rules from the VPP: %v", err)
		}
		ifName, _, found := h.ifIndexes.LookupBySwIfIndex(uint32(msg.SwIfIndex))
		if !found {
			h.log.Warnf("STN dump: interface name not found for index %d", msg.SwIfIndex)
		}

		var stnIP net.IP
		if msg.IPAddress.Af == ip_types.ADDRESS_IP4 {
			stnAddr := msg.IPAddress.Un.GetIP4()
			stnIP = net.IP(stnAddr[:])
		} else {
			stnAddr := msg.IPAddress.Un.GetIP6()
			stnIP = net.IP(stnAddr[:])
		}

		stnRule := &stn.Rule{
			IpAddress: stnIP.String(),
			Interface: ifName,
		}
		stnMeta := &vppcalls.StnMeta{
			IfIdx: uint32(msg.SwIfIndex),
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
