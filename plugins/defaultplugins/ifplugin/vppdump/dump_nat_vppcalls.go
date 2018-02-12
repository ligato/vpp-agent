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

// Nat44GlobalConfigDump returns global config in NB format
func Nat44GlobalConfigDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (*nat.Nat44Global, error) {
	// Dump all necessary data to reconstruct global NAT configuration
	isEnabled, err := nat44IsForwardingEnabled(log, vppChan, measure.GetTimeLog(&bin_api.Nat44ForwardingIsEnabled{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natInterfaces, err := nat44InterfaceDump(swIfIndices, log, vppChan, measure.GetTimeLog(&bin_api.Nat44InterfaceDump{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natOutputFeature, err := nat44InterfaceOutputFeatureDump(swIfIndices, log, vppChan, measure.GetTimeLog(&bin_api.Nat44InterfaceDump{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natAddressPools, err := nat44AddressDump(log, vppChan, measure.GetTimeLog(&bin_api.Nat44AddressDump{}, stopwatch))
	if err != nil {
		return nil, err
	}

	// Combine interfaces with output feature with the rest of them
	var nat44GlobalInterfaces []*nat.Nat44Global_NatInterfaces
	for _, natInterface := range natInterfaces {
		nat44GlobalInterfaces = append(nat44GlobalInterfaces, &nat.Nat44Global_NatInterfaces{
			Name:     natInterface.Name,
			IsInside: natInterface.IsInside,
			OutputFeature: func(ofIfs []*nat.Nat44Global_NatInterfaces, ifName string) bool {
				for _, ofIf := range ofIfs {
					if ofIf.Name == ifName {
						return true
					}
				}
				return false
			}(natOutputFeature, natInterface.Name),
		})
	}

	// Set fields
	return &nat.Nat44Global{
		Forwarding:    isEnabled,
		NatInterfaces: nat44GlobalInterfaces,
		AddressPools:  natAddressPools,
	}, nil
}

func NAT44DNatDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (*nat.Nat44DNat, error) {
	// Dump all necessary data to reconstruct DNAT configuration
	natStMappings, err := nat44StaticMappingDump(swIfIndices, log, vppChan, measure.GetTimeLog(&bin_api.Nat44StaticMappingDump{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natStLbMapping, err := nat44StaticMappingLbDump(log, vppChan, measure.GetTimeLog(&bin_api.Nat44LbStaticMappingDump{}, stopwatch))
	if err != nil {
		return nil, err
	}
	natIdMappings, err := nat44IdentityMappingDump(swIfIndices, log, vppChan, measure.GetTimeLog(&bin_api.Nat44IdentityMappingDump{}, stopwatch))
	if err != nil {
		return nil, err
	}

	// Append static mappings
	natStMappings = append(natStMappings, natStLbMapping...)

	// Reconstruct DNat config object
	var dNatConfigs []*nat.Nat44DNat_DNatConfig
	dNatConfig := &nat.Nat44DNat_DNatConfig{
		StMappings: natStMappings,
		IdMappings: natIdMappings,
	}
	// Currently it is not possible to distinguish which mapping belongs to which DNAT configuration, so everything
	// is returned as a one config
	dNatConfigs = append(dNatConfigs, dNatConfig)

	return &nat.Nat44DNat{
		DnatConfig: dNatConfigs,
	}, nil
}

// nat44AddressDump returns a list of NAT44 address pools configured in the VPP
func nat44AddressDump(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (addresses []*nat.Nat44Global_AddressPools, err error) {
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

		addresses = append(addresses, &nat.Nat44Global_AddressPools{
			FirstSrcAddress: ipAddress.To4().String(),
			VrfId:           msg.VrfID,
			TwiceNat: func(twiceNat uint8) bool {
				if twiceNat == 1 {
					return true
				}
				return false
			}(msg.TwiceNat),
		})
	}

	log.Debugf("NAT44 address pool dump complete, found %d entries", len(addresses))

	return
}

// nat44StaticMappingDump returns a list of static mapping entries
func nat44StaticMappingDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel,
	timeLog measure.StopWatchEntry) (entries []*nat.Nat44DNat_DNatConfig_StaticMappigs, err error) {
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
		var locals []*nat.Nat44DNat_DNatConfig_StaticMappigs_LocalIPs
		lcIPAddress = msg.LocalIPAddress
		exIPAddress = msg.ExternalIPAddress

		entries = append(entries, &nat.Nat44DNat_DNatConfig_StaticMappigs{
			VrfId: msg.VrfID,
			ExternalInterface: func(ifIdx uint32) string {
				ifName, _, found := swIfIndices.LookupName(ifIdx)
				if !found && ifIdx != 0xffffffff {
					log.Warnf("Interface with index %v not found in the mapping", ifIdx)
				}
				return ifName
			}(msg.ExternalSwIfIndex),
			ExternalIP:   exIPAddress.To4().String(),
			ExternalPort: uint32(msg.ExternalPort),
			LocalIps: append(locals, &nat.Nat44DNat_DNatConfig_StaticMappigs_LocalIPs{ // single-value
				LocalIP:   lcIPAddress.To4().String(),
				LocalPort: uint32(msg.LocalPort),
			}),
			Protocol: getNatProtocol(msg.Protocol, log),
			TwiceNat: func(twiceNat uint8) bool {
				if twiceNat == 1 {
					return true
				}
				return false
			}(msg.TwiceNat),
		})
	}

	log.Debugf("NAT44 static mapping dump complete, found %d entries", len(entries))

	return
}

// nat44StaticMappingLbDump returns a list of static mapping entries with load balancer
func nat44StaticMappingLbDump(log logging.Logger, vppChan *govppapi.Channel,
	timeLog measure.StopWatchEntry) (entries []*nat.Nat44DNat_DNatConfig_StaticMappigs, err error) {
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
		var locals []*nat.Nat44DNat_DNatConfig_StaticMappigs_LocalIPs
		var localIP net.IP
		for _, localIPVal := range msg.Locals {
			localIP = localIPVal.Addr
			locals = append(locals, &nat.Nat44DNat_DNatConfig_StaticMappigs_LocalIPs{
				LocalIP:     localIP.To4().String(),
				LocalPort:   uint32(localIPVal.Port),
				Probability: uint32(localIPVal.Probability),
			})
		}
		exIPAddress = msg.ExternalAddr

		entries = append(entries, &nat.Nat44DNat_DNatConfig_StaticMappigs{
			VrfId:        msg.VrfID,
			ExternalIP:   exIPAddress.To4().String(),
			ExternalPort: uint32(msg.ExternalPort),
			LocalIps:     locals,
			Protocol:     getNatProtocol(msg.Protocol, log),
			TwiceNat: func(twiceNat uint8) bool {
				if twiceNat == 1 {
					return true
				}
				return false
			}(msg.TwiceNat),
		})
	}

	log.Debugf("NAT44 lb-static mapping dump complete, found %d entries", len(entries))

	return
}

// nat44IdentityMappingDump returns a list of static mapping entries with load balancer
func nat44IdentityMappingDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel,
	timeLog measure.StopWatchEntry) (entries []*nat.Nat44DNat_DNatConfig_IdentityMappings, err error) {
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

		entries = append(entries, &nat.Nat44DNat_DNatConfig_IdentityMappings{
			VrfId: msg.VrfID,
			AddressedInterface: func(ifIdx uint32) string {
				ifName, _, found := swIfIndices.LookupName(ifIdx)
				if !found && ifIdx != 0xffffffff {
					log.Warnf("Interface with index %v not found in the mapping", ifIdx)
				}
				return ifName
			}(msg.SwIfIndex),
			IpAddress: ipAddress.To4().String(),
			Port:      uint32(msg.Port),
			Protocol:  getNatProtocol(msg.Protocol, log),
		})
	}

	log.Debugf("NAT44 identity mapping dump complete, found %d entries", len(entries))

	return
}

// nat44InterfaceDump returns a list of interfaces enabled for NAT44
func nat44InterfaceDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel,
	timeLog measure.StopWatchEntry) (interfaces []*nat.Nat44Global_NatInterfaces, err error) {
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

		// Find interface name
		ifName, _, found := swIfIndices.LookupName(msg.SwIfIndex)
		if !found {
			log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}

		interfaces = append(interfaces, &nat.Nat44Global_NatInterfaces{
			Name: ifName,
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
func nat44InterfaceOutputFeatureDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel,
	timeLog measure.StopWatchEntry) (interfaces []*nat.Nat44Global_NatInterfaces, err error) {
	// Nat44InterfaceOutputFeatureDump time measurement
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

		// Find interface name
		ifName, _, found := swIfIndices.LookupName(msg.SwIfIndex)
		if !found {
			log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}

		interfaces = append(interfaces, &nat.Nat44Global_NatInterfaces{
			Name: ifName,
			IsInside: func(isInside uint8) bool {
				if isInside == 1 {
					return true
				}
				return false
			}(msg.IsInside),
			OutputFeature: true,
		})
	}

	log.Debugf("NAT44 interface with output feature dump complete, found %d entries", len(interfaces))

	return
}

// Nat44IsForwardingEnabled returns a list of interfaces enabled for NAT44
func nat44IsForwardingEnabled(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (isEnabled bool, err error) {
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

// returns NAT numeric representation of provided protocol value
func getNatProtocol(protocol uint8, log logging.Logger) (proto nat.Protocol) {
	if protocol == vppcalls.TCP {
		return nat.Protocol_TCP
	} else if protocol == vppcalls.UDP {
		return nat.Protocol_UDP
	} else if protocol == vppcalls.ICMP {
		return nat.Protocol_ICMP
	}
	log.Warnf("Identity mapping dump returned unknown protocol %d", protocol)
	return 0
}
