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

	if err := handleNat44AdressPool(first, last, vrf, twiceNat, true, vppChan); err != nil {
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

	if err := handleNat44AdressPool(first, last, vrf, twiceNat, false, vppChan); err != nil {
		return nil
	}

	log.Debugf("Address pool %v - %v removed", first, last)

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
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("Enabling NAT for interface %v returned %d", ifName, reply.Retval)
	}

	return nil
}

// Calls VPP binary API to add/remove address pool
func handleNat44AdressPool(first, last []byte, vrf uint32, twiceNat, isAdd bool, vppChan *govppapi.Channel) error {
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
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("Adding NAT44 address pool returned %d", reply.Retval)
	}

	return nil
}
