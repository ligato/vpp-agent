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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
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
	Protocol      nat.Protocol
	OutToInOnly   bool // rule match only out2in direction
	TwiceNat      bool
	VrfID         uint32
}

// Nat44IdentityMappingEntry represents single NAT44 identity mapping entry
type Nat44IdentityMappingEntry struct {
	AddressOnly bool
	IPAddress   string
	Port        uint32
	IfIdx       uint32
	Protocol    nat.Protocol
	VrfID       uint32
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

// Nat44GlobalConfigDump returns global config in NB format
func Nat44GlobalConfigDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (*nat.Nat44Global, error) {
	// Dump all necessary data to reconstruct global NAT configuration
	isEnabled, err := Nat44IsForwardingEnabled(log, vppChan, measure.GetTimeLog(&bin_api.Nat44ForwardingIsEnabled{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natInterfaces, err := nat44InterfaceDump(log, vppChan, measure.GetTimeLog(&bin_api.Nat44InterfaceDump{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natOutputFeature, err := nat44InterfaceOutputFeatureDump(log, vppChan, measure.GetTimeLog(&bin_api.Nat44InterfaceDump{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natAddressPools, err := nat44AddressDump(log, vppChan, measure.GetTimeLog(&bin_api.Nat44AddressDump{}, stopwatch))
	if err != nil {
		return nil, err
	}

	var vrfID uint32

	// Interfaces
	var nat44GlobalInterfaces []*nat.Nat44Global_NatInterface
	for _, natInterface := range natInterfaces {
		nat44GlobalInterfaces = append(nat44GlobalInterfaces, &nat.Nat44Global_NatInterface{
			Name: func(ifIdx uint32) (ifName string) {
				var found bool
				ifName, _, found = swIfIndices.LookupName(ifIdx)
				if !found {
					log.Warnf("Interface with index %v not found in the mapping", ifIdx)
					return
				}
				return
			}(natInterface.IfIdx),
			IsInside: natInterface.IsInside,
			OutputFeature: func(ofIfaces []*Nat44Interface, ifIdx uint32) bool {
				for _, ofIface := range ofIfaces {
					if ofIface.IfIdx == ifIdx {
						return true
					}
				}
				return false
			}(natOutputFeature, natInterface.IfIdx),
		})
	}

	// NAT address pools
	var nat44AddressPools []*nat.Nat44Global_AddressPool
	for _, addressPool := range natAddressPools {
		nat44AddressPools = append(nat44AddressPools, &nat.Nat44Global_AddressPool{
			FirstSrcAddress: addressPool.IPAddress,
			TwiceNat:        addressPool.TwiceNat,
		})
		// VRF ID is the same for every entry
		vrfID = addressPool.VrfID
	}

	// Set fields
	return &nat.Nat44Global{
		VrfId:        vrfID,
		Forwarding:   isEnabled,
		NatInterface: nat44GlobalInterfaces,
		AddressPool:  nat44AddressPools,
	}, nil
}

func NAT44DNatDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (*nat.Nat44DNat, error) {
	// Dump all necessary data to reconstruct DNAT configuration
	natStaticMappings, err := nat44StaticMappingDump(log, vppChan, measure.GetTimeLog(&bin_api.Nat44StaticMappingDump{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natStaticLbMappings, err := nat44StaticMappingLbDump(log, vppChan, measure.GetTimeLog(&bin_api.Nat44LbStaticMappingDump{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natIdentityMappings, err := nat44IdentityMappingDump(log, vppChan, measure.GetTimeLog(&bin_api.Nat44IdentityMappingDump{}, stopwatch))
	if err != nil {
		return nil, err
	}

	// Common fields
	var vrfID uint32
	var sNat bool

	// Static mapping
	var nat44AllStaticMappings []*nat.Nat44DNat_DNatConfig_Mapping
	for _, staticMapping := range append(natStaticMappings, natStaticLbMappings...) {
		nat44AllStaticMappings = append(nat44AllStaticMappings, &nat.Nat44DNat_DNatConfig_Mapping{
			ExternalInterface: func(ifIdx uint32) (ifName string) {
				var found bool
				ifName, _, found = swIfIndices.LookupName(ifIdx)
				if !found {
					log.Warnf("Interface with index %v not found in the mapping", ifIdx)
					return
				}
				return
			}(staticMapping.ExternalIfIdx),
			ExternalIP:   staticMapping.ExternalIP,
			ExternalPort: staticMapping.ExternalPort,
			LocalIp: func(localIPPort []*LocalIPEntry) (locals []*nat.Nat44DNat_DNatConfig_Mapping_LocalIP) {
				for _, localIP := range localIPPort {
					locals = append(locals, &nat.Nat44DNat_DNatConfig_Mapping_LocalIP{
						LocalIP:     localIP.LocalIP,
						LocalPort:   localIP.LocalPort,
						Probability: localIP.Probability,
					})
				}
				return
			}(staticMapping.LocalIPs),
			Protocol: staticMapping.Protocol,
		})
		// Common fields are the same
		vrfID = staticMapping.VrfID
		sNat = staticMapping.TwiceNat
	}

	// Identity mapping
	var nat44IdentityMappings []*nat.Nat44DNat_DNatConfig_IdentityMapping
	for _, identityMapping := range natIdentityMappings {
		nat44IdentityMappings = append(nat44IdentityMappings, &nat.Nat44DNat_DNatConfig_IdentityMapping{
			AddressedInterface: func(ifIdx uint32) (ifName string) {
				var found bool
				ifName, _, found = swIfIndices.LookupName(ifIdx)
				if !found && ifIdx != 0xffffffff {
					log.Warnf("Interface with index %v not found in the mapping", ifIdx)
					return
				}
				return
			}(identityMapping.IfIdx),
			IpAddress: identityMapping.IPAddress,
			Port:      identityMapping.Port,
			Protocol:  identityMapping.Protocol,
		})
	}

	// Reconstruct DNat config object
	var dNatConfigs []*nat.Nat44DNat_DNatConfig
	dNatConfig := &nat.Nat44DNat_DNatConfig{
		VrfId:       vrfID,
		SNatEnabled: sNat,
		Mapping:     nat44AllStaticMappings,
		IdMapping:   nat44IdentityMappings,
	}
	// Currently it is not possible to distinguish which mapping belongs to which DNAT configuration, so everything
	// is returned as a one config
	dNatConfigs = append(dNatConfigs, dNatConfig)

	return &nat.Nat44DNat{
		DnatConfig: dNatConfigs,
	}, nil
}

// nat44AddressDump returns a list of NAT44 address pools configured in the VPP
func nat44AddressDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (addresses []*Nat44AddressPool, err error) {
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

// nat44StaticMappingDump returns a list of static mapping entries
func nat44StaticMappingDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (entries []*Nat44StaticMappingEntry, err error) {
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
			Protocol: func(protocol uint8) nat.Protocol {
				if protocol == vppcalls.TCP {
					return nat.Protocol_TCP
				} else if protocol == vppcalls.UDP {
					return nat.Protocol_UDP
				} else if protocol == vppcalls.ICMP {
					return nat.Protocol_ICMP
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

// nat44StaticMappingLbDump returns a list of static mapping entries with load balancer
func nat44StaticMappingLbDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (entries []*Nat44StaticMappingEntry, err error) {
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
			Protocol: func(protocol uint8) nat.Protocol {
				if protocol == vppcalls.TCP {
					return nat.Protocol_TCP
				} else if protocol == vppcalls.UDP {
					return nat.Protocol_UDP
				} else if protocol == vppcalls.ICMP {
					return nat.Protocol_ICMP
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

// nat44IdentityMappingDump returns a list of static mapping entries with load balancer
func nat44IdentityMappingDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (entries []*Nat44IdentityMappingEntry, err error) {
	// Nat44IdentityMappingDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	var ipAddress net.IP

	req := &bin_api.Nat44IdentityMappingDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44IdentityMappingDetails{}
		stop, replyErr := reqContext.ReceiveReply(msg)
		if replyErr != nil {
			err = fmt.Errorf("failed to dump NAT44 identity mapping: %v", replyErr)
			return
		}
		if stop {
			break
		}

		ipAddress = msg.IPAddress

		entries = append(entries, &Nat44IdentityMappingEntry{
			AddressOnly: func(addrOnly uint8) bool {
				if addrOnly == 1 {
					return true
				}
				return false
			}(msg.AddrOnly),
			IPAddress: ipAddress.To4().String(),
			Port:      uint32(msg.Port),
			IfIdx:     msg.SwIfIndex,
			Protocol: func(protocol uint8) nat.Protocol {
				if protocol == vppcalls.TCP {
					return nat.Protocol_TCP
				} else if protocol == vppcalls.UDP {
					return nat.Protocol_UDP
				} else if protocol == vppcalls.ICMP {
					return nat.Protocol_ICMP
				}
				log.Warnf("Identity mapping dump returned unknown protocol %d", protocol)
				return 0
			}(msg.Protocol),
			VrfID: msg.VrfID,
		})
	}

	log.Debugf("NAT44 identity mapping dump complete, found %d entries", len(entries))

	return
}

// nat44InterfaceDump returns a list of interfaces enabled for NAT44
func nat44InterfaceDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (interfaces []*Nat44Interface, err error) {
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

// nat44InterfaceOutputFeatureDump returns a list of interfaces with output feature set
func nat44InterfaceOutputFeatureDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (interfaces []*Nat44Interface, err error) {
	// Nat44InterfaceDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &bin_api.Nat44InterfaceOutputFeatureDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44InterfaceOutputFeatureDetails{}
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

	log.Debugf("NAT44 interface with output feature dump complete, found %d entries", len(interfaces))

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
