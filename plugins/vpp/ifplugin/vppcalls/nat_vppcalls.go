// Copyright (c) 2018 Cisco and/or its affiliates.
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

package vppcalls

import (
	"fmt"
	"time"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/nat"
	nat2 "github.com/ligato/vpp-agent/plugins/vpp/model/nat"
)

// Num protocol representation
const (
	ICMP uint8 = 1
	TCP  uint8 = 6
	UDP  uint8 = 17
)

// NoInterface is sw-if-idx which means 'no interface'
const NoInterface uint32 = 0xffffffff

// StaticMappingContext groups common fields required for static mapping
type StaticMappingContext struct {
	Tag           string
	AddressOnly   bool
	LocalIP       []byte
	LocalPort     uint16
	ExternalIP    []byte
	ExternalPort  uint16
	ExternalIfIdx uint32
	Protocol      uint8
	Vrf           uint32
	TwiceNat      bool
	SelfTwiceNat  bool
}

// StaticMappingLbContext groups common fields required for static mapping with load balancer
type StaticMappingLbContext struct {
	Tag          string
	LocalIPs     []*LocalLbAddress
	ExternalIP   []byte
	ExternalPort uint16
	Protocol     uint8
	TwiceNat     bool
	SelfTwiceNat bool
}

// IdentityMappingContext groups common fields required for identity mapping
type IdentityMappingContext struct {
	Tag       string
	IPAddress []byte
	Protocol  uint8
	Port      uint16
	IfIdx     uint32
	Vrf       uint32
}

// LocalLbAddress represents one local IP and address entry
type LocalLbAddress struct {
	Vrf         uint32
	Tag         string
	LocalIP     []byte
	LocalPort   uint16
	Probability uint8
}

func (handler *natVppHandler) SetNat44Forwarding(enableFwd bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(nat.Nat44ForwardingEnableDisable{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44ForwardingEnableDisable{
		Enable: boolToUint(enableFwd),
	}

	reply := &nat.Nat44ForwardingEnableDisableReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to set/unset interface as NAT
func (handler *natVppHandler) handleNat44Interface(ifIdx uint32, isInside, isAdd bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(nat.Nat44InterfaceAddDelFeature{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44InterfaceAddDelFeature{
		SwIfIndex: ifIdx,
		IsInside:  boolToUint(isInside),
		IsAdd:     boolToUint(isAdd),
	}

	reply := &nat.Nat44InterfaceAddDelFeatureReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to set/unset interface as NAT with output feature
func (handler *natVppHandler) handleNat44InterfaceOutputFeature(ifIdx uint32, isInside, isAdd bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(nat.Nat44InterfaceAddDelOutputFeature{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44InterfaceAddDelOutputFeature{
		SwIfIndex: ifIdx,
		IsInside:  boolToUint(isInside),
		IsAdd:     boolToUint(isAdd),
	}

	reply := &nat.Nat44InterfaceAddDelOutputFeatureReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove address pool
func (handler *natVppHandler) handleNat44AddressPool(first, last []byte, vrf uint32, twiceNat, isAdd bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(nat.Nat44AddDelAddressRange{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44AddDelAddressRange{
		FirstIPAddress: first,
		LastIPAddress:  last,
		VrfID:          vrf,
		TwiceNat:       boolToUint(twiceNat),
		IsAdd:          boolToUint(isAdd),
	}

	reply := &nat.Nat44AddDelAddressRangeReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to setup NAT virtual reassembly
func (handler *natVppHandler) handleNat44VirtualReassembly(timeout, maxReass, maxFrag uint32, dropFrag, isIpv6 bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(nat.NatSetReass{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.NatSetReass{
		Timeout:  timeout,
		MaxReass: uint16(maxReass),
		MaxFrag:  uint8(maxFrag),
		DropFrag: boolToUint(dropFrag),
		IsIP6:    boolToUint(isIpv6),
	}

	reply := &nat.NatSetReassReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove static mapping
func (handler *natVppHandler) handleNat44StaticMapping(ctx *StaticMappingContext, isAdd, addrOnly bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(nat.Nat44AddDelStaticMapping{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44AddDelStaticMapping{
		Tag:               []byte(ctx.Tag),
		LocalIPAddress:    ctx.LocalIP,
		LocalPort:         ctx.LocalPort,
		ExternalIPAddress: ctx.ExternalIP,
		ExternalPort:      ctx.ExternalPort,
		Protocol:          ctx.Protocol,
		ExternalSwIfIndex: ctx.ExternalIfIdx,
		VrfID:             ctx.Vrf,
		TwiceNat:          boolToUint(ctx.TwiceNat),
		SelfTwiceNat:      boolToUint(ctx.SelfTwiceNat),
		Out2inOnly:        1,
		IsAdd:             boolToUint(isAdd),
	}
	if addrOnly {
		req.AddrOnly = 1
	} else {
		req.LocalPort = ctx.LocalPort
		req.ExternalPort = ctx.ExternalPort
	}

	reply := &nat.Nat44AddDelStaticMappingReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove static mapping with load balancer
func (handler *natVppHandler) handleNat44StaticMappingLb(ctx *StaticMappingLbContext, isAdd bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(nat.Nat44AddDelLbStaticMapping{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Transform local IP/Ports
	var localAddrPorts []nat.Nat44LbAddrPort
	for _, ctxLocal := range ctx.LocalIPs {
		localAddrPort := nat.Nat44LbAddrPort{
			Addr:        ctxLocal.LocalIP,
			Port:        ctxLocal.LocalPort,
			Probability: ctxLocal.Probability,
			VrfID:       ctxLocal.Vrf,
		}
		localAddrPorts = append(localAddrPorts, localAddrPort)
	}

	req := &nat.Nat44AddDelLbStaticMapping{
		Tag:          []byte(ctx.Tag),
		Locals:       localAddrPorts,
		LocalNum:     uint8(len(localAddrPorts)),
		ExternalAddr: ctx.ExternalIP,
		ExternalPort: ctx.ExternalPort,
		Protocol:     ctx.Protocol,
		TwiceNat:     boolToUint(ctx.TwiceNat),
		SelfTwiceNat: boolToUint(ctx.SelfTwiceNat),
		Out2inOnly:   1,
		IsAdd:        boolToUint(isAdd),
	}

	reply := &nat.Nat44AddDelLbStaticMappingReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove identity mapping
func (handler *natVppHandler) handleNat44IdentityMapping(ctx *IdentityMappingContext, isAdd bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(nat.Nat44AddDelIdentityMapping{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44AddDelIdentityMapping{
		Tag: []byte(ctx.Tag),
		AddrOnly: func(port uint16, ip []byte) uint8 {
			// Set addr only if port is set to zero
			if port == 0 || ip == nil {
				return 1
			}
			return 0
		}(ctx.Port, ctx.IPAddress),
		IPAddress: ctx.IPAddress,
		Port:      ctx.Port,
		Protocol:  ctx.Protocol,
		SwIfIndex: func(ifIdx uint32) uint32 {
			if ifIdx == 0 {
				return NoInterface
			}
			return ifIdx
		}(ctx.IfIdx),
		VrfID: ctx.Vrf,
		IsAdd: boolToUint(isAdd),
	}

	reply := &nat.Nat44AddDelIdentityMappingReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

func (handler *natVppHandler) EnableNat44Interface(ifIdx uint32, isInside bool) error {
	return handler.handleNat44Interface(ifIdx, isInside, true)
}

func (handler *natVppHandler) DisableNat44Interface(ifIdx uint32, isInside bool) error {
	return handler.handleNat44Interface(ifIdx, isInside, false)
}

func (handler *natVppHandler) EnableNat44InterfaceOutput(ifIdx uint32, isInside bool) error {
	return handler.handleNat44InterfaceOutputFeature(ifIdx, isInside, true)
}

func (handler *natVppHandler) DisableNat44InterfaceOutput(ifIdx uint32, isInside bool) error {
	return handler.handleNat44InterfaceOutputFeature(ifIdx, isInside, false)
}

func (handler *natVppHandler) AddNat44AddressPool(first, last []byte, vrf uint32, twiceNat bool) error {
	return handler.handleNat44AddressPool(first, last, vrf, twiceNat, true)
}

func (handler *natVppHandler) DelNat44AddressPool(first, last []byte, vrf uint32, twiceNat bool) error {
	return handler.handleNat44AddressPool(first, last, vrf, twiceNat, false)
}

func (handler *natVppHandler) SetVirtualReassemblyIPv4(vrCfg *nat2.Nat44Global_VirtualReassembly) error {
	return handler.handleNat44VirtualReassembly(vrCfg.Timeout, vrCfg.MaxReass, vrCfg.MaxFrag, vrCfg.DropFrag, false)
}

func (handler *natVppHandler) SetVirtualReassemblyIPv6(vrCfg *nat2.Nat44Global_VirtualReassembly) error {
	return handler.handleNat44VirtualReassembly(vrCfg.Timeout, vrCfg.MaxReass, vrCfg.MaxFrag, vrCfg.DropFrag, true)
}

func (handler *natVppHandler) AddNat44IdentityMapping(ctx *IdentityMappingContext) error {
	return handler.handleNat44IdentityMapping(ctx, true)
}

func (handler *natVppHandler) DelNat44IdentityMapping(ctx *IdentityMappingContext) error {
	return handler.handleNat44IdentityMapping(ctx, false)
}

func (handler *natVppHandler) AddNat44StaticMapping(ctx *StaticMappingContext) error {
	if ctx.AddressOnly {
		return handler.handleNat44StaticMapping(ctx, true, true)
	}
	return handler.handleNat44StaticMapping(ctx, true, false)
}

func (handler *natVppHandler) DelNat44StaticMapping(ctx *StaticMappingContext) error {
	if ctx.AddressOnly {
		return handler.handleNat44StaticMapping(ctx, false, true)
	}
	return handler.handleNat44StaticMapping(ctx, false, false)
}

func (handler *natVppHandler) AddNat44StaticMappingLb(ctx *StaticMappingLbContext) error {
	return handler.handleNat44StaticMappingLb(ctx, true)
}

func (handler *natVppHandler) DelNat44StaticMappingLb(ctx *StaticMappingLbContext) error {
	return handler.handleNat44StaticMappingLb(ctx, false)
}
