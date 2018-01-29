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

package vppdump

import (
	"fmt"
	"net"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	bin_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
)

// Nat44AddressPool is a representation of single address used in NAT
type Nat44AddressPool struct {
	IPAddress string
	TwiceNat  bool
	VrfID     uint32
}

// Nat44StaticMappingEntry represents single NAT44 mapping entry
type Nat44StaticMappingEntry struct {
	AddressOnly   bool
	LocalIPs      []*LocalIPEntry // more than one if load balanced
	ExternalIP    string
	ExternalPort  uint32
	ExternalIfIdx uint32
	Protocol      nat.Nat44DNat_DNatConfig_Mapping_Protocol
	OutToInOnly   bool // rule match only out2in direction
	TwiceNat      bool
	VrfID         uint32
}

// Nat44Interface is an interface with flag telling about whether it is inside or outside interface
type Nat44Interface struct {
	IfIdx    uint32
	IsInside bool
}

// LocalIPEntry is a single Local IP/Port value with probability
type LocalIPEntry struct {
	LocalIP     string
	LocalPort   uint32
	Probability uint32
}

// Nat44AddressDump returns a list of NAT44 address pools configured in the VPP
func Nat44AddressDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (addresses []*Nat44AddressPool, err error) {
	// Nat44AddressDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	var ipAddress net.IP

	req := &bin_api.Nat44AddressDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44AddressDetails{}
		stop, replyErr := reqContext.ReceiveReply(msg)
		if replyErr != nil {
			err = fmt.Errorf("failed to dump NAT44 Address pool: %v", replyErr)
			return
		}
		if stop {
			break
		}

		ipAddress = msg.IPAddress

		addresses = append(addresses, &Nat44AddressPool{
			IPAddress: ipAddress.To4().String(),
			TwiceNat: func(twiceNat uint8) bool {
				if twiceNat == 1 {
					return true
				}
				return false
			}(msg.TwiceNat),
			VrfID: msg.VrfID,
		})
	}

	log.Debugf("NAT44 address pool dump complete, found %d entries", len(addresses))

	return
}

// Nat44StaticMappingDump returns a list of static mapping entries
func Nat44StaticMappingDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (entries []*Nat44StaticMappingEntry, err error) {
	// Nat44StaticMappingDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	var lcIPAddress net.IP
	var exIPAddress net.IP

	req := &bin_api.Nat44StaticMappingDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44StaticMappingDetails{}
		stop, replyErr := reqContext.ReceiveReply(msg)
		if replyErr != nil {
			err = fmt.Errorf("failed to dump NAT44 static mapping: %v", replyErr)
			return
		}
		if stop {
			break
		}
		var locals []*LocalIPEntry
		lcIPAddress = msg.LocalIPAddress
		exIPAddress = msg.ExternalIPAddress

		entries = append(entries, &Nat44StaticMappingEntry{
			AddressOnly: func(addrOnly uint8) bool {
				if addrOnly == 1 {
					return true
				}
				return false
			}(msg.AddrOnly),
			LocalIPs: append(locals, &LocalIPEntry{ // single-value
				LocalIP:   lcIPAddress.To4().String(),
				LocalPort: uint32(msg.LocalPort),
			}),
			ExternalIP:    exIPAddress.To4().String(),
			ExternalPort:  uint32(msg.ExternalPort),
			ExternalIfIdx: msg.ExternalSwIfIndex,
			Protocol: func(protocol uint8) nat.Nat44DNat_DNatConfig_Mapping_Protocol {
				if protocol == vppcalls.TCP {
					return nat.Nat44DNat_DNatConfig_Mapping_TCP
				} else if protocol == vppcalls.UDP {
					return nat.Nat44DNat_DNatConfig_Mapping_UDP
				} else if protocol == vppcalls.ICMP {
					return nat.Nat44DNat_DNatConfig_Mapping_ICMP
				}
				log.Warnf("Static mapping dump returned unknown protocol %d", protocol)
				return 0
			}(msg.Protocol),
			TwiceNat: func(twiceNat uint8) bool {
				if twiceNat == 1 {
					return true
				}
				return false
			}(msg.TwiceNat),
			VrfID: msg.VrfID,
		})
	}

	log.Debugf("NAT44 static mapping dump complete, found %d entries", len(entries))

	return
}

// Nat44StaticMappingLbDump returns a list of static mapping entries with load balancer
func Nat44StaticMappingLbDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (entries []*Nat44StaticMappingEntry, err error) {
	// Nat44LbStaticMappingDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	var exIPAddress net.IP

	req := &bin_api.Nat44LbStaticMappingDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44LbStaticMappingDetails{}
		stop, replyErr := reqContext.ReceiveReply(msg)
		if replyErr != nil {
			err = fmt.Errorf("failed to dump NAT44 lb-static mapping: %v", replyErr)
			return
		}
		if stop {
			break
		}

		// Prepare localIPs
		var locals []*LocalIPEntry
		var localIP net.IP
		for _, localIPVal := range msg.Locals {
			localIP = localIPVal.Addr
			locals = append(locals, &LocalIPEntry{
				LocalIP:     localIP.To4().String(),
				LocalPort:   uint32(localIPVal.Port),
				Probability: uint32(localIPVal.Probability),
			})
		}
		exIPAddress = msg.ExternalAddr

		entries = append(entries, &Nat44StaticMappingEntry{
			AddressOnly:  false,
			LocalIPs:     locals,
			ExternalIP:   exIPAddress.To4().String(),
			ExternalPort: uint32(msg.ExternalPort),
			Protocol: func(protocol uint8) nat.Nat44DNat_DNatConfig_Mapping_Protocol {
				if protocol == vppcalls.TCP {
					return nat.Nat44DNat_DNatConfig_Mapping_TCP
				} else if protocol == vppcalls.UDP {
					return nat.Nat44DNat_DNatConfig_Mapping_UDP
				} else if protocol == vppcalls.ICMP {
					return nat.Nat44DNat_DNatConfig_Mapping_ICMP
				}
				log.Warnf("Static mapping dump returned unknown protocol %d", protocol)
				return 0
			}(msg.Protocol),
			TwiceNat: func(twiceNat uint8) bool {
				if twiceNat == 1 {
					return true
				}
				return false
			}(msg.TwiceNat),
			OutToInOnly: func(out2in uint8) bool {
				if out2in == 1 {
					return true
				}
				return false
			}(msg.Out2inOnly),
			VrfID: msg.VrfID,
		})
	}

	log.Debugf("NAT44 lb-static mapping dump complete, found %d entries", len(entries))

	return
}

// Nat44InterfaceDump returns a list of interfaces enabled for NAT44
func Nat44InterfaceDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (interfaces []*Nat44Interface, err error) {
	// Nat44InterfaceDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &bin_api.Nat44InterfaceDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44InterfaceDetails{}
		stop, replyErr := reqContext.ReceiveReply(msg)
		if replyErr != nil {
			err = fmt.Errorf("failed to dump NAT44 interface: %v", replyErr)
			return
		}
		if stop {
			break
		}

		interfaces = append(interfaces, &Nat44Interface{
			IfIdx: msg.SwIfIndex,
			IsInside: func(isInside uint8) bool {
				if isInside == 1 {
					return true
				}
				return false
			}(msg.IsInside),
		})
	}

	log.Debugf("NAT44 interface dump complete, found %d entries", len(interfaces))

	return
}

// Nat44IsForwardingEnabled returns a list of interfaces enabled for NAT44
func Nat44IsForwardingEnabled(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (isEnabled bool, err error) {
	// Nat44ForwardingIsEnabled time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &bin_api.Nat44ForwardingIsEnabled{}

	msg := &bin_api.Nat44ForwardingIsEnabledReply{}
	replyErr := vppChan.SendRequest(req).ReceiveReply(msg)
	if replyErr != nil {
		err = fmt.Errorf("failed to dump forwarding: %v", replyErr)
		return
	}

	isEnabled = func(enabled uint8) bool {
		if enabled == 1 {
			return true
		}
		return false
	}(msg.Enabled)

	log.Debugf("NAT44 forwarding dump complete, is enabled: %v", isEnabled)

	return
}
