// Copyright (c) 2017 Cisco and/or its affiliates.
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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/vpe"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
)

// VppSetAllInterfacesToBridgeDomain does lookup all interfaces which belongs to bridge domain, and bvi interface
func VppSetAllInterfacesToBridgeDomain(bridgeDomain *l2.BridgeDomains_BridgeDomain, bridgeDomainIndex uint32,
	swIfIndexes ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel) ([]string, []string, string) {
	log.Debug("Interface lookup started for ", bridgeDomain.Name)

	var allBdInterfaces []string
	var configuredBdInterfaces []string
	var bviInterfaceName string

	// Find bridge domain interfaces
	if len(bridgeDomain.Interfaces) == 0 {
		log.Infof("Bridge domain %v has no interface to set", bridgeDomain.Name)
		return allBdInterfaces, configuredBdInterfaces, bviInterfaceName
	}

	bridgeDomainInterfaces := bridgeDomain.Interfaces
	for _, bdInterface := range bridgeDomainInterfaces {
		// Find which interface is bvi (if any)
		if bdInterface.BridgedVirtualInterface {
			bviInterfaceName = bdInterface.Name
		}
		// Find wheteher interface already exists
		interfaceIndex, _, found := swIfIndexes.LookupIdx(bdInterface.Name)
		if !found {
			log.Infof("Interface %v not found", bdInterface.Name)
			allBdInterfaces = append(allBdInterfaces, bdInterface.Name)
			continue
		}
		req := &vpe.SwInterfaceSetL2Bridge{}
		req.BdID = bridgeDomainIndex
		req.RxSwIfIndex = interfaceIndex
		req.Enable = 1
		if bdInterface.BridgedVirtualInterface {
			// Set up BVI interface
			req.Bvi = 1
			log.Debugf("Interface %v set as BVI", bdInterface.Name)
		}
		reply := &vpe.SwInterfaceSetL2BridgeReply{}
		err := vppChan.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			log.WithFields(logging.Fields{"Error": err, "Bridge Domain": bridgeDomain.Name}).Error("Error while assigning interface to bridge domain")
			continue
		}
		if 0 != reply.Retval {
			log.WithFields(logging.Fields{"Return value": reply.Retval}).Error("Unexpected return value")
			continue
		}
		log.WithFields(logging.Fields{"Interface": bdInterface.Name, "BD": bridgeDomain.Name}).Debug("Interface set to bridge domain.")
		allBdInterfaces = append(allBdInterfaces, bdInterface.Name)
		configuredBdInterfaces = append(configuredBdInterfaces, bdInterface.Name)
	}

	return allBdInterfaces, configuredBdInterfaces, bviInterfaceName
}

// VppUnsetAllInterfacesFromBridgeDomain removes all interfaces from bridge domain (set them as L3)
func VppUnsetAllInterfacesFromBridgeDomain(bridgeDomain *l2.BridgeDomains_BridgeDomain, bridgeDomainIndex uint32,
	swIfIndexes ifaceidx.SwIfIndex, log logging.Logger, vppChan *govppapi.Channel) []string {
	log.Debug("Interface lookup started for ", bridgeDomain.Name)

	// Store all interface names, will be used to unregister potential bridge domain to interface pairs
	var interfaces []string

	// Find all interfaces including BVI
	if len(bridgeDomain.Interfaces) == 0 {
		log.Infof("Bridge domain %v has no interfaces, nothin go unset", bridgeDomain.Name)
		return interfaces
	}

	bridgeDomainInterfaces := bridgeDomain.Interfaces
	for _, bdInterface := range bridgeDomainInterfaces {
		interfaces = append(interfaces, bdInterface.Name)
		// Find interface
		interfaceIndex, _, found := swIfIndexes.LookupIdx(bdInterface.Name)
		if !found {
			log.Debugf("Interface %v not found, no need to unset", bdInterface.Name)
			continue
		}
		req := &vpe.SwInterfaceSetL2Bridge{}
		req.BdID = bridgeDomainIndex
		req.RxSwIfIndex = interfaceIndex
		req.Enable = 0

		reply := &vpe.SwInterfaceSetL2BridgeReply{}
		err := vppChan.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			log.WithFields(logging.Fields{"Error": err, "Bridge Domain": bridgeDomain.Name}).Error("Error while setting up interface as L3")
			continue
		}
		if 0 != reply.Retval {
			log.WithFields(logging.Fields{"Return value": reply.Retval}).Error("Unexpected return value")
			continue
		}
		log.WithFields(logging.Fields{"Interface": bdInterface.Name, "BD": bridgeDomain.Name}).Debug("Interface removed from bridge domain.")
	}

	return interfaces
}

// VppSetInterfaceToBridgeDomain sets provided interface to bridge domain
func VppSetInterfaceToBridgeDomain(bridgeDomainIndex uint32, interfaceIndex uint32, bvi bool, log logging.Logger, vppChan *govppapi.Channel) {
	log.Debugf("Setting up interface %v to bridge domain %v ", interfaceIndex, bridgeDomainIndex)

	req := &vpe.SwInterfaceSetL2Bridge{}
	req.BdID = bridgeDomainIndex
	req.RxSwIfIndex = interfaceIndex
	req.Enable = 1
	if bvi {
		req.Bvi = 1
	} else {
		req.Bvi = 0
	}

	reply := &vpe.SwInterfaceSetL2BridgeReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		log.WithFields(logging.Fields{"Error": err, "Bridge Domain": bridgeDomainIndex}).Error("Error while assigning interface to bridge domain")
	}
	if 0 != reply.Retval {
		log.WithFields(logging.Fields{"Return value": reply.Retval}).Error("Unexpected return value")
	}
	log.WithFields(logging.Fields{"Interface": interfaceIndex, "BD": bridgeDomainIndex}).Debug("Interface set to bridge domain.")
}
