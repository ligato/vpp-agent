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
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	vpp_stn "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/stn"
	stn "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/stn"
)

// AddSTNRule implements STN handler, adds a new STN rule to the VPP.
func (h *StnVppHandler) AddSTNRule(stnRule *stn.Rule) error {
	return h.addDelStnRule(stnRule, true)
}

// DeleteSTNRule implements STN handler, removes the provided STN rule from the VPP.
func (h *StnVppHandler) DeleteSTNRule(stnRule *stn.Rule) error {
	return h.addDelStnRule(stnRule, false)
}

func (h *StnVppHandler) addDelStnRule(stnRule *stn.Rule, isAdd bool) error {
	// get interface index
	ifaceMeta, found := h.ifIndexes.LookupByName(stnRule.Interface)
	if !found {
		return errors.New("failed to get interface metadata")
	}
	swIfIndex := ifaceMeta.GetIndex()

	// remove mask from IP address if necessary
	ipAddr := stnRule.IpAddress
	ipParts := strings.Split(ipAddr, "/")
	if len(ipParts) > 1 {
		h.log.Debugf("STN IP address %s is defined with mask, removing it")
		ipAddr = ipParts[0]
	}

	// parse IP address
	ip, err := ipToAddress(ipAddr)
	if err != nil {
		return err
	}

	// add STN rule
	req := &vpp_stn.StnAddDelRule{
		IPAddress: ip,
		SwIfIndex: interface_types.InterfaceIndex(swIfIndex),
		IsAdd:     isAdd,
	}
	reply := &vpp_stn.StnAddDelRuleReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil

}

func ipToAddress(address string) (dhcpAddr ip_types.Address, err error) {
	netIP := net.ParseIP(address)
	if netIP == nil {
		return ip_types.Address{}, fmt.Errorf("invalid IP: %q", address)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		dhcpAddr.Af = ip_types.ADDRESS_IP6
		var ip6addr ip_types.IP6Address
		copy(ip6addr[:], netIP.To16())
		dhcpAddr.Un.SetIP6(ip6addr)
	} else {
		dhcpAddr.Af = ip_types.ADDRESS_IP4
		var ip4addr ip_types.IP4Address
		copy(ip4addr[:], ip4)
		dhcpAddr.Un.SetIP4(ip4addr)
	}
	return
}
