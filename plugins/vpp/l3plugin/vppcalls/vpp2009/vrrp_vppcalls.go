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

package vpp2009

import (
	"errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/vrrp"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	maxUint8  = 255
	maxUint16 = 65535
)

var (
	errInvalidAddrNum   = errors.New("addrs quantity should be > 0 && <= 255")
	errIvalidVrID       = errors.New("vr_id should be > 0 && <= 255")
	errIvalidPriority   = errors.New("priority should be > 0 && <= 255")
	errIvalidInterval   = errors.New("interval should be > 0 && <= 65535")
	errInvalidIPVersion = errors.New("ipv6_flag does not correspond to IP version of the provided address")
	errInvalidInterface = errors.New("interface does not exist")
)

func (h *VrrpVppHandler) vppAddDelVrrp(entry *l3.VRRPEntry, isAdd uint8) error {
	var flags uint32
	if entry.PreemtpFlag {
		flags |= uint32(vrrp.VRRP_API_VR_PREEMPT)
	}
	if entry.AcceptFlag {
		flags |= uint32(vrrp.VRRP_API_VR_ACCEPT)
	}
	if entry.UnicastFlag {
		flags |= uint32(vrrp.VRRP_API_VR_UNICAST)
	}
	if entry.Ipv6Flag {
		flags |= uint32(vrrp.VRRP_API_VR_IPV6)
	}

	var addrs []ip_types.Address
	for _, addr := range entry.Addrs {
		ip, err := ipToAddress(addr)
		if err != nil {
			return err
		}

		if entry.Ipv6Flag && ip.Af == ip_types.ADDRESS_IP4 ||
			!entry.Ipv6Flag && ip.Af == ip_types.ADDRESS_IP6 {
			return errInvalidIPVersion
		}

		addrs = append(addrs, ip)
	}

	addrsLen := len(addrs)
	if addrsLen > maxUint8 || addrsLen == 0 {
		return errInvalidAddrNum
	}

	if entry.GetVrId() > maxUint8 || entry.GetVrId() == 0 {
		return errIvalidVrID
	}

	if entry.GetPriority() > maxUint8 || entry.GetPriority() == 0 {
		return errIvalidPriority
	}

	if entry.GetInterval() > maxUint16 || entry.GetInterval() == 0 {
		return errIvalidInterval
	}

	md, exist := h.ifIndexes.LookupByName(entry.Interface)
	if !exist {
		return errInvalidInterface
	}

	req := &vrrp.VrrpVrAddDel{
		IsAdd:     isAdd,
		SwIfIndex: interface_types.InterfaceIndex(md.SwIfIndex),
		VrID:      uint8(entry.GetVrId()),
		Priority:  uint8(entry.GetPriority()),
		Interval:  uint16(entry.GetInterval()),
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
		return errInvalidInterface
	}

	req := &vrrp.VrrpVrStartStop{
		SwIfIndex: interface_types.InterfaceIndex(md.SwIfIndex),
		VrID:      uint8(entry.VrId),
		IsIPv6:    boolToUint(entry.GetIpv6Flag()),
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
