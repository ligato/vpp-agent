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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
)

// Num protocol representation
const (
	// ICMP proto
	ICMP uint8 = 1
	// TCP proto
	TCP uint8 = 6
	// UDP proto
	UDP uint8 = 17
)

// StaticMappingContext groups common fields required for static mapping
type StaticMappingContext struct {
	AddressOnly   bool
	LocalIP       []byte
	LocalPort     uint16
	ExternalIP    []byte
	ExternalPort  uint16
	ExternalIfIdx uint32
	Protocol      uint8
}

// StaticMappingLbContext groups common fields required for static mapping with load balancer
type StaticMappingLbContext struct {
	LocalIPs     []*LocalLbAddress
	ExternalIP   []byte
	ExternalPort uint16
	Protocol     uint8
}

// LocalLbAddress represents one local IP and address entry
type LocalLbAddress struct {
	LocalIP     []byte
	LocalPort   uint16
	Probability uint8
}

// SetNat44Forwarding configures global forwarding setup for NAT44
func SetNat44Forwarding(fwd bool, log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// Nat44ForwardingEnableDisable time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &nat.Nat44ForwardingEnableDisable{}
	req.Enable = func(value bool) uint8 {
		if value {
			return 1
		}
		return 0
	}(fwd)

	reply := &nat.Nat44ForwardingEnableDisableReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("setting up NAT forwarding returned %d", reply.Retval)
	}
	if fwd {
		log.Debugf("NAT forwarding enabled")
	} else {
		log.Debugf("NAT forwarding disabled")
	}

	return nil
}

// EnableNat44Interface enables NAT feature for provided interface
func EnableNat44Interface(ifName string, ifIdx uint32, isInside bool, log logging.Logger, vppChan *govppapi.Channel,
	timeLog measure.StopWatchEntry) error {
	// Nat44InterfaceAddDelFeature time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44Interface(ifName, ifIdx, isInside, true, vppChan); err != nil {
		return err
	}

	log.Debugf("NAT feature enabled for interface %v", ifName)

	return nil
}

// DisableNat44Interface enables NAT feature for provided interface
func DisableNat44Interface(ifName string, ifIdx uint32, isInside bool, log logging.Logger, vppChan *govppapi.Channel,
	timeLog measure.StopWatchEntry) error {
	// Nat44InterfaceAddDelFeature time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44Interface(ifName, ifIdx, isInside, false, vppChan); err != nil {
		return err
	}

	log.Debugf("NAT feature disabled for interface %v", ifName)

	return nil
}

// EnableNat44InterfaceOutput enables NAT output feature for provided interface
func EnableNat44InterfaceOutput(ifName string, ifIdx uint32, isInside bool, log logging.Logger, vppChan *govppapi.Channel,
	timeLog measure.StopWatchEntry) error {
	// Nat44InterfaceAddDelOutputFeature time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44InterfaceOutputFeature(ifName, ifIdx, isInside, true, vppChan); err != nil {
		return err
	}

	log.Debugf("NAT output feature enabled for interface %v", ifName)

	return nil
}

// DisableNat44InterfaceOutput disables NAT output feature for provided interface
func DisableNat44InterfaceOutput(ifName string, ifIdx uint32, isInside bool, log logging.Logger, vppChan *govppapi.Channel,
	timeLog measure.StopWatchEntry) error {
	// Nat44InterfaceAddDelOutputFeature time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44InterfaceOutputFeature(ifName, ifIdx, isInside, false, vppChan); err != nil {
		return err
	}

	log.Debugf("NAT output feature disabled for interface %v", ifName)

	return nil
}

// AddNat44AddressPool sets new NAT address pool
func AddNat44AddressPool(first, last []byte, vrf uint32, twiceNat bool, log logging.Logger,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// Nat44AddDelAddressRange time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44AddressPool(first, last, vrf, twiceNat, true, vppChan); err != nil {
		return nil
	}

	log.Debugf("Address pool set to %v - %v", first, last)

	return nil
}

// DelNat44AddressPool removes existing NAT address pool
func DelNat44AddressPool(first, last []byte, vrf uint32, twiceNat bool, log logging.Logger,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// Nat44AddDelAddressRange time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44AddressPool(first, last, vrf, twiceNat, false, vppChan); err != nil {
		return nil
	}

	log.Debugf("Address pool %v - %v removed", first, last)

	return nil
}

// AddNat44IdentityMapping sets new NAT address pool
func AddNat44IdentityMapping(ip []byte, protocol uint8, port uint16, ifIdx, vrf uint32, log logging.Logger,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// Nat44AddDelAddressRange time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44IdentityMapping(ip, protocol, port, ifIdx, vrf, true, vppChan); err != nil {
		return nil
	}

	log.Debug("Identity mapping added")

	return nil
}

// DelNat44IdentityMapping sets new NAT address pool
func DelNat44IdentityMapping(ip []byte, protocol uint8, port uint16, ifIdx, vrf uint32, log logging.Logger,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// Nat44AddDelAddressRange time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44IdentityMapping(ip, protocol, port, ifIdx, vrf, false, vppChan); err != nil {
		return nil
	}

	log.Debug("Identity mapping removed")

	return nil
}

// AddNat44StaticMapping creates new static mapping entry (considering address only or both, address and port
// depending on the context)
func AddNat44StaticMapping(ctx *StaticMappingContext, vrfID uint32, twiceNat bool, log logging.Logger,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// Nat44AddDelStaticMapping time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if ctx.AddressOnly {
		if err := handleNat44StaticMappingAddrOnly(ctx, vrfID, twiceNat, true, log, vppChan); err != nil {
			return nil
		}
	} else {
		if err := handleNat44StaticMapping(ctx, vrfID, twiceNat, true, log, vppChan); err != nil {
			return nil
		}
	}

	log.Debug("Static mapping added")

	return nil
}

// DelNat44StaticMapping removes existing static mapping entry
func DelNat44StaticMapping(ctx *StaticMappingContext, vrfID uint32, twiceNat bool, log logging.Logger,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// Nat44AddDelStaticMapping time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if ctx.AddressOnly {
		if err := handleNat44StaticMappingAddrOnly(ctx, vrfID, twiceNat, false, log, vppChan); err != nil {
			return nil
		}
	} else {
		if err := handleNat44StaticMapping(ctx, vrfID, twiceNat, false, log, vppChan); err != nil {
			return nil
		}
	}

	log.Debug("Static mapping removed")

	return nil
}

// AddNat44StaticMappingLb creates new static mapping entry with load balancer
func AddNat44StaticMappingLb(ctx *StaticMappingLbContext, vrfID uint32, twiceNat bool, log logging.Logger,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// Nat44AddDelLbStaticMapping time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44StaticMappingLb(ctx, vrfID, twiceNat, true, vppChan); err != nil {
		return nil
	}

	log.Debug("Static mapping with load balancer added")

	return nil
}

// DelNat44StaticMappingLb removes existing static mapping entry with load balancer
func DelNat44StaticMappingLb(ctx *StaticMappingLbContext, vrfID uint32, twiceNat bool, log logging.Logger,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// Nat44AddDelLbStaticMapping time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err := handleNat44StaticMappingLb(ctx, vrfID, twiceNat, false, vppChan); err != nil {
		return nil
	}

	log.Debug("Static mapping with load balancer removed")

	return nil
}

// Calls VPP binary API to set/unset interface as NAT
func handleNat44Interface(ifName string, ifIdx uint32, isInside, isAdd bool, vppChan *govppapi.Channel) error {
	req := &nat.Nat44InterfaceAddDelFeature{
		SwIfIndex: ifIdx,
		IsInside: func(isInside bool) uint8 {
			if isInside {
				return 1
			}
			return 0
		}(isInside),
		IsAdd: func(isAdd bool) uint8 {
			if isAdd {
				return 1
			}
			return 0
		}(isAdd),
	}

	reply := &nat.Nat44InterfaceAddDelFeatureReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("enabling NAT for interface %v returned %d", ifName, reply.Retval)
	}

	return nil
}

// Calls VPP binary API to set/unset interface as NAT with output feature
func handleNat44InterfaceOutputFeature(ifName string, ifIdx uint32, isInside, isAdd bool, vppChan *govppapi.Channel) error {
	req := &nat.Nat44InterfaceAddDelOutputFeature{
		SwIfIndex: ifIdx,
		IsInside: func(isInside bool) uint8 {
			if isInside {
				return 1
			}
			return 0
		}(isInside),
		IsAdd: func(isAdd bool) uint8 {
			if isAdd {
				return 1
			}
			return 0
		}(isAdd),
	}

	reply := &nat.Nat44InterfaceAddDelOutputFeatureReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("enabling NAT output feature for interface %v returned %d", ifName, reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove address pool
func handleNat44AddressPool(first, last []byte, vrf uint32, twiceNat, isAdd bool, vppChan *govppapi.Channel) error {
	req := &nat.Nat44AddDelAddressRange{
		FirstIPAddress: first,
		LastIPAddress:  last,
		VrfID:          vrf,
		TwiceNat: func(twiceNat bool) uint8 {
			if twiceNat {
				return 1
			}
			return 0
		}(twiceNat),
		IsAdd: func(isAdd bool) uint8 {
			if isAdd {
				return 1
			}
			return 0
		}(isAdd),
	}

	reply := &nat.Nat44AddDelAddressRangeReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("adding NAT44 address pool returned %d", reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove static mapping
func handleNat44StaticMapping(ctx *StaticMappingContext, vrf uint32, twiceNat, isAdd bool,
	log logging.Logger, vppChan *govppapi.Channel) error {
	log.Debugf("NAT44 static mapping adding: %v", isAdd)
	req := &nat.Nat44AddDelStaticMapping{
		LocalIPAddress:    ctx.LocalIP,
		LocalPort:         ctx.LocalPort,
		ExternalIPAddress: ctx.ExternalIP,
		ExternalPort:      ctx.ExternalPort,
		Protocol:          ctx.Protocol,
		ExternalSwIfIndex: ctx.ExternalIfIdx,
		VrfID:             vrf,
		TwiceNat: func(twiceNat bool) uint8 {
			if twiceNat {
				return 1
			}
			return 0
		}(twiceNat),
		IsAdd: func(isAdd bool) uint8 {
			if isAdd {
				return 1
			}
			return 0
		}(isAdd),
	}

	reply := &nat.Nat44AddDelStaticMappingReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("adding NAT44 static mapping (address only) returned %d", reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove static mapping (address only)
func handleNat44StaticMappingAddrOnly(ctx *StaticMappingContext, vrf uint32, twiceNat, isAdd bool, log logging.Logger,
	vppChan *govppapi.Channel) error {
	log.Debugf("NAT44 static mapping (address only) adding: %v", isAdd)
	req := &nat.Nat44AddDelStaticMapping{
		AddrOnly:          1,
		LocalIPAddress:    ctx.LocalIP,
		ExternalIPAddress: ctx.ExternalIP,
		Protocol:          ctx.Protocol,
		ExternalSwIfIndex: ctx.ExternalIfIdx,
		VrfID:             vrf,
		TwiceNat: func(twiceNat bool) uint8 {
			if twiceNat {
				return 1
			}
			return 0
		}(twiceNat),
		IsAdd: func(isAdd bool) uint8 {
			if isAdd {
				return 1
			}
			return 0
		}(isAdd),
	}

	reply := &nat.Nat44AddDelStaticMappingReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("adding NAT44 static mapping (address only) returned %d", reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove static mapping with load balancer
func handleNat44StaticMappingLb(ctx *StaticMappingLbContext, vrf uint32, twiceNat, isAdd bool, vppChan *govppapi.Channel) error {
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
		Locals:       localAddrPorts,
		LocalNum:     uint8(len(localAddrPorts)),
		ExternalAddr: ctx.ExternalIP,
		ExternalPort: ctx.ExternalPort,
		Protocol:     ctx.Protocol,

		VrfID: vrf,
		TwiceNat: func(twiceNat bool) uint8 {
			if twiceNat {
				return 1
			}
			return 0
		}(twiceNat),
		IsAdd: func(isAdd bool) uint8 {
			if isAdd {
				return 1
			}
			return 0
		}(isAdd),
	}

	reply := &nat.Nat44AddDelLbStaticMappingReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("adding NAT44 static mapping with load ballancer returned %d", reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove identity mapping
func handleNat44IdentityMapping(ip []byte, protocol uint8, port uint16, ifIdx, vrf uint32, isAdd bool, vppChan *govppapi.Channel) error {
	req := &nat.Nat44AddDelIdentityMapping{
		AddrOnly: func(port uint16, ip []byte) uint8 {
			// Set addr only if port is set to zero
			if port == 0 || ip == nil {
				return 1
			}
			return 0
		}(port, ip),
		IPAddress: ip,
		Port:      port,
		Protocol:  protocol,
		SwIfIndex: func(ifIdx uint32) uint32 {
			if ifIdx == 0 {
				return 0xffffffff // means no interface
			}
			return ifIdx
		}(ifIdx),
		VrfID: vrf,
		IsAdd: func(isAdd bool) uint8 {
			if isAdd {
				return 1
			}
			return 0
		}(isAdd),
	}

	reply := &nat.Nat44AddDelIdentityMappingReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("adding NAT44 identity mapping returned %d", reply.Retval)
	}

	return nil
}
