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
	isEnabled, err := nat44IsForwardingEnabled(log, vppChan, stopwatch)
	if err != nil {
		return nil, err
	}
	natInterfaces, err := nat44InterfaceDump(swIfIndices, log, vppChan, stopwatch)
	if err != nil {
		return nil, err
	}
	natOutputFeature, err := nat44InterfaceOutputFeatureDump(swIfIndices, log, vppChan, stopwatch)
	if err != nil {
		return nil, err
	}
	natAddressPools, err := nat44AddressDump(log, vppChan, stopwatch)
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

// NAT44NatDump
func NAT44DNatDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (*nat.Nat44DNat, error) {
	// Dump all necessary data to reconstruct DNAT configuration
	natStMappings, err := nat44StaticMappingDump(swIfIndices, log, vppChan, stopwatch)
	if err != nil {
		return nil, err
	}
	natStLbMapping, err := nat44StaticMappingLbDump(log, vppChan, stopwatch)
	if err != nil {
		return nil, err
	}
	natIdMappings, err := nat44IdentityMappingDump(swIfIndices, log, vppChan, stopwatch)
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
func nat44AddressDump(log logging.Logger, vppChan *govppapi.Channel,
	stopwatch *measure.Stopwatch) (addresses []*nat.Nat44Global_AddressPools, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bin_api.Nat44AddressDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44AddressDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44AddressDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 Address pool: %v", err)
		}
		if stop {
			break
		}

		ipAddress := net.IP(msg.IPAddress)

		addresses = append(addresses, &nat.Nat44Global_AddressPools{
			FirstSrcAddress: ipAddress.To4().String(),
			VrfId:           msg.VrfID,
			TwiceNat:        uintToBool(msg.TwiceNat),
		})
	}

	log.Debugf("NAT44 address pool dump complete, found %d entries", len(addresses))

	return
}

// nat44StaticMappingDump returns a list of static mapping entries
func nat44StaticMappingDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel,
	stopwatch *measure.Stopwatch) (entries []*nat.Nat44DNat_DNatConfig_StaticMappigs, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bin_api.Nat44StaticMappingDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44StaticMappingDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44StaticMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 static mapping: %v", err)
		}
		if stop {
			break
		}
		var locals []*nat.Nat44DNat_DNatConfig_StaticMappigs_LocalIPs
		lcIPAddress := net.IP(msg.LocalIPAddress)
		exIPAddress := net.IP(msg.ExternalIPAddress)

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
			TwiceNat: uintToBool(msg.TwiceNat),
		})
	}

	log.Debugf("NAT44 static mapping dump complete, found %d entries", len(entries))

	return entries, nil
}

// nat44StaticMappingLbDump returns a list of static mapping entries with load balancer
func nat44StaticMappingLbDump(log logging.Logger, vppChan *govppapi.Channel,
	stopwatch *measure.Stopwatch) (entries []*nat.Nat44DNat_DNatConfig_StaticMappigs, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bin_api.Nat44LbStaticMappingDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44LbStaticMappingDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44LbStaticMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 lb-static mapping: %v", err)
		}
		if stop {
			break
		}

		// Prepare localIPs
		var locals []*nat.Nat44DNat_DNatConfig_StaticMappigs_LocalIPs
		for _, localIPVal := range msg.Locals {
			localIP := net.IP(localIPVal.Addr)
			locals = append(locals, &nat.Nat44DNat_DNatConfig_StaticMappigs_LocalIPs{
				LocalIP:     localIP.To4().String(),
				LocalPort:   uint32(localIPVal.Port),
				Probability: uint32(localIPVal.Probability),
			})
		}

		exIPAddress := net.IP(msg.ExternalAddr)
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

	return entries, nil
}

// nat44IdentityMappingDump returns a list of static mapping entries with load balancer
func nat44IdentityMappingDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel,
	stopwatch *measure.Stopwatch) (entries []*nat.Nat44DNat_DNatConfig_IdentityMappings, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bin_api.Nat44IdentityMappingDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44IdentityMappingDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44IdentityMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 identity mapping: %v", err)
		}
		if stop {
			break
		}

		ipAddress := net.IP(msg.IPAddress)

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

	return entries, nil
}

// nat44InterfaceDump returns a list of interfaces enabled for NAT44
func nat44InterfaceDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel,
	stopwatch *measure.Stopwatch) (interfaces []*nat.Nat44Global_NatInterfaces, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bin_api.Nat44InterfaceDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44InterfaceDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44InterfaceDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface: %v", err)
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
			Name:     ifName,
			IsInside: uintToBool(msg.IsInside),
		})
	}

	log.Debugf("NAT44 interface dump complete, found %d entries", len(interfaces))

	return
}

// nat44InterfaceOutputFeatureDump returns a list of interfaces with output feature set
func nat44InterfaceOutputFeatureDump(swIfIndices ifaceidx.SwIfIndex, log logging.Logger,
	vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (ifaces []*nat.Nat44Global_NatInterfaces, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bin_api.Nat44InterfaceOutputFeatureDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44InterfaceOutputFeatureDump{}
	reqContext := vppChan.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44InterfaceOutputFeatureDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface: %v", err)
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

		ifaces = append(ifaces, &nat.Nat44Global_NatInterfaces{
			Name:          ifName,
			IsInside:      uintToBool(msg.IsInside),
			OutputFeature: true,
		})
	}

	log.Debugf("NAT44 interface with output feature dump complete, found %d entries", len(ifaces))

	return ifaces, nil
}

// Nat44IsForwardingEnabled returns a list of interfaces enabled for NAT44
func nat44IsForwardingEnabled(log logging.Logger, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (isEnabled bool, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bin_api.Nat44ForwardingIsEnabled{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44ForwardingIsEnabled{}

	reply := &bin_api.Nat44ForwardingIsEnabledReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return false, fmt.Errorf("failed to dump forwarding: %v", err)
	}

	isEnabled = uintToBool(reply.Enabled)
	log.Debugf("NAT44 forwarding dump complete, is enabled: %v", isEnabled)

	return isEnabled, nil
}

// returns NAT numeric representation of provided protocol value
func getNatProtocol(protocol uint8, log logging.Logger) (proto nat.Protocol) {
	switch protocol {
	case vppcalls.TCP:
		return nat.Protocol_TCP
	case vppcalls.UDP:
		return nat.Protocol_UDP
	case vppcalls.ICMP:
		return nat.Protocol_ICMP
	default:
		log.Warnf("Unknown protocol %v", protocol)
		return 0
	}
}

func uintToBool(value uint8) bool {
	if value == 0 {
		return false
	}
	return true
}
