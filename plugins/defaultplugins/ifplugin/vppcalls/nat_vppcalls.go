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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
)

// Num protocol representation
const (
	ICMP uint8 = 1
	TCP  uint8 = 6
	UDP  uint8 = 17
)

const noInterface uint32 = 0xffffffff

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
}

// StaticMappingLbContext groups common fields required for static mapping with load balancer
type StaticMappingLbContext struct {
	Tag          string
	LocalIPs     []*LocalLbAddress
	ExternalIP   []byte
	ExternalPort uint16
	Protocol     uint8
	Vrf          uint32
	TwiceNat     bool
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
	Tag         string
	LocalIP     []byte
	LocalPort   uint16
	Probability uint8
}

// SetNat44Forwarding configures global forwarding setup for NAT44
func SetNat44Forwarding(enableFwd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(nat.Nat44ForwardingEnableDisable{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44ForwardingEnableDisable{
		Enable: boolToUint(enableFwd),
	}

	reply := &nat.Nat44ForwardingEnableDisableReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to set/unset interface as NAT
func handleNat44Interface(ifIdx uint32, isInside, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(nat.Nat44InterfaceAddDelFeature{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44InterfaceAddDelFeature{
		SwIfIndex: ifIdx,
		IsInside:  boolToUint(isInside),
		IsAdd:     boolToUint(isAdd),
	}

	reply := &nat.Nat44InterfaceAddDelFeatureReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to set/unset interface as NAT with output feature
func handleNat44InterfaceOutputFeature(ifIdx uint32, isInside, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(nat.Nat44InterfaceAddDelOutputFeature{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44InterfaceAddDelOutputFeature{
		SwIfIndex: ifIdx,
		IsInside:  boolToUint(isInside),
		IsAdd:     boolToUint(isAdd),
	}

	reply := &nat.Nat44InterfaceAddDelOutputFeatureReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove address pool
func handleNat44AddressPool(first, last []byte, vrf uint32, twiceNat, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(nat.Nat44AddDelAddressRange{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &nat.Nat44AddDelAddressRange{
		FirstIPAddress: first,
		LastIPAddress:  last,
		VrfID:          vrf,
		TwiceNat:       boolToUint(twiceNat),
		IsAdd:          boolToUint(isAdd),
	}

	reply := &nat.Nat44AddDelAddressRangeReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove static mapping
func handleNat44StaticMapping(ctx *StaticMappingContext, isAdd, addrOnly bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(nat.Nat44AddDelStaticMapping{}).LogTimeEntry(time.Since(t))
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
		IsAdd:             boolToUint(isAdd),
	}
	if addrOnly {
		req.AddrOnly = 1
	} else {
		req.LocalPort = ctx.LocalPort
		req.ExternalPort = ctx.ExternalPort
	}

	reply := &nat.Nat44AddDelStaticMappingReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove static mapping with load balancer
func handleNat44StaticMappingLb(ctx *StaticMappingLbContext, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(nat.Nat44AddDelLbStaticMapping{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Transform local IP/Ports
	var localAddrPorts []nat.Nat44LbAddrPort
	for _, ctxLocal := range ctx.LocalIPs {
		localAddrPort := nat.Nat44LbAddrPort{
			Addr:        ctxLocal.LocalIP,
			Port:        ctxLocal.LocalPort,
			Probability: ctxLocal.Probability,
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
		VrfID:        ctx.Vrf,
		TwiceNat:     boolToUint(ctx.TwiceNat),
		IsAdd:        boolToUint(isAdd),
	}

	reply := &nat.Nat44AddDelLbStaticMappingReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove identity mapping
func handleNat44IdentityMapping(ctx *IdentityMappingContext, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(nat.Nat44AddDelIdentityMapping{}).LogTimeEntry(time.Since(t))
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
				return 0xffffffff // means no interface
			}
			return ifIdx
		}(ctx.IfIdx),
		VrfID: ctx.IfIdx,
		IsAdd: boolToUint(isAdd),
	}

	reply := &nat.Nat44AddDelIdentityMappingReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// EnableNat44Interface enables NAT feature for provided interface
func EnableNat44Interface(ifIdx uint32, isInside bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44Interface(ifIdx, isInside, true, vppChan, stopwatch)
}

// DisableNat44Interface enables NAT feature for provided interface
func DisableNat44Interface(ifIdx uint32, isInside bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44Interface(ifIdx, isInside, false, vppChan, stopwatch)
}

// EnableNat44InterfaceOutput enables NAT output feature for provided interface
func EnableNat44InterfaceOutput(ifIdx uint32, isInside bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44InterfaceOutputFeature(ifIdx, isInside, true, vppChan, stopwatch)
}

// DisableNat44InterfaceOutput disables NAT output feature for provided interface
func DisableNat44InterfaceOutput(ifIdx uint32, isInside bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44InterfaceOutputFeature(ifIdx, isInside, false, vppChan, stopwatch)
}

// AddNat44AddressPool sets new NAT address pool
func AddNat44AddressPool(first, last []byte, vrf uint32, twiceNat bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44AddressPool(first, last, vrf, twiceNat, true, vppChan, stopwatch)
}

// DelNat44AddressPool removes existing NAT address pool
func DelNat44AddressPool(first, last []byte, vrf uint32, twiceNat bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44AddressPool(first, last, vrf, twiceNat, false, vppChan, stopwatch)
}

// AddNat44IdentityMapping sets new NAT address pool
func AddNat44IdentityMapping(ctx *IdentityMappingContext, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44IdentityMapping(ctx, true, vppChan, stopwatch)
}

// DelNat44IdentityMapping sets new NAT address pool
func DelNat44IdentityMapping(ctx *IdentityMappingContext, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44IdentityMapping(ctx, false, vppChan, stopwatch)
}

// AddNat44StaticMapping creates new static mapping entry
// (considering address only or both, address and port depending on the context)
func AddNat44StaticMapping(ctx *StaticMappingContext, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	if ctx.AddressOnly {
		return handleNat44StaticMapping(ctx, true, true, vppChan, stopwatch)
	}
	return handleNat44StaticMapping(ctx, true, false, vppChan, stopwatch)
}

// DelNat44StaticMapping removes existing static mapping entry
func DelNat44StaticMapping(ctx *StaticMappingContext, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	if ctx.AddressOnly {
		return handleNat44StaticMapping(ctx, false, true, vppChan, stopwatch)
	}
	return handleNat44StaticMapping(ctx, false, false, vppChan, stopwatch)
}

// AddNat44StaticMappingLb creates new static mapping entry with load balancer
func AddNat44StaticMappingLb(ctx *StaticMappingLbContext, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44StaticMappingLb(ctx, true, vppChan, stopwatch)
}

// DelNat44StaticMappingLb removes existing static mapping entry with load balancer
func DelNat44StaticMappingLb(ctx *StaticMappingLbContext, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return handleNat44StaticMappingLb(ctx, false, vppChan, stopwatch)
}
