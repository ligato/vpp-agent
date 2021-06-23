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
	"fmt"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vrrp"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	centiMilliRatio uint32 = 10
)

func (h *VrrpVppHandler) vppAddDelVrrp(entry *l3.VRRPEntry, isAdd uint8) error {
	var addrs []ip_types.Address
	var isIpv6 bool
	for idx, addr := range entry.IpAddresses {
		ip, err := ip_types.ParseAddress(addr)
		if err != nil {
			return err
		}

		if idx == 0 && ip.Af == ip_types.ADDRESS_IP6 {
			isIpv6 = true
		}

		addrs = append(addrs, ip)
	}

	md, exist := h.ifIndexes.LookupByName(entry.Interface)
	if !exist {
		return fmt.Errorf("interface does not exist: %v", entry.Interface)
	}

	var flags uint32
	if entry.Preempt {
		flags |= uint32(vrrp.VRRP_API_VR_PREEMPT)
	}
	if entry.Accept {
		flags |= uint32(vrrp.VRRP_API_VR_ACCEPT)
	}
	if entry.Unicast {
		flags |= uint32(vrrp.VRRP_API_VR_UNICAST)
	}
	if isIpv6 {
		flags |= uint32(vrrp.VRRP_API_VR_IPV6)
	}

	req := &vrrp.VrrpVrAddDel{
		IsAdd:     isAdd,
		SwIfIndex: interface_types.InterfaceIndex(md.SwIfIndex),
		VrID:      uint8(entry.GetVrId()),
		Priority:  uint8(entry.GetPriority()),
		Interval:  uint16(entry.GetInterval() / centiMilliRatio),
		Flags:     vrrp.VrrpVrFlags(flags),
		NAddrs:    uint8(len(addrs)),
		Addrs:     addrs,
	}

	reply := &vrrp.VrrpVrAddDelReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// VppAddVrrp implements VRRP handler.
func (h *VrrpVppHandler) VppAddVrrp(entry *l3.VRRPEntry) error {
	return h.vppAddDelVrrp(entry, 1)
}

// VppDelVrrp implements VRRP handler.
func (h *VrrpVppHandler) VppDelVrrp(entry *l3.VRRPEntry) error {
	return h.vppAddDelVrrp(entry, 0)
}

func (h *VrrpVppHandler) vppStartStopVrrp(entry *l3.VRRPEntry, isStart uint8) error {

	md, exist := h.ifIndexes.LookupByName(entry.Interface)
	if !exist {
		return fmt.Errorf("interface does not exist: %v", entry.Interface)
	}

	var isIpv6 bool
	for idx, addr := range entry.IpAddresses {
		ip, err := ipToAddress(addr)
		if err != nil {
			return err
		}

		if idx == 0 && ip.Af == ip_types.ADDRESS_IP6 {
			isIpv6 = true
		}
	}

	req := &vrrp.VrrpVrStartStop{
		SwIfIndex: interface_types.InterfaceIndex(md.SwIfIndex),
		VrID:      uint8(entry.VrId),
		IsIPv6:    boolToUint(isIpv6),
		IsStart:   isStart,
	}

	reply := &vrrp.VrrpVrStartStopReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// VppStartVrrp implements VRRP handler.
func (h *VrrpVppHandler) VppStartVrrp(entry *l3.VRRPEntry) error {
	return h.vppStartStopVrrp(entry, 1)
}

// VppStopVrrp implements VRRP handler.
func (h *VrrpVppHandler) VppStopVrrp(entry *l3.VRRPEntry) error {
	return h.vppStartStopVrrp(entry, 0)
}
