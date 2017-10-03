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
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/vpe"
	"net"
	"strconv"
	"strings"
)

// VppAddArpTerminationTableEntry adds ARP termination entry
func VppAddArpTerminationTableEntry(bridgeDomainID uint32, mac string, ip string, log logging.Logger, vppChan *govppapi.Channel) error {
	log.Info("Adding arp termination entry")

	parsedMac, errMac := net.ParseMAC(mac)
	if errMac != nil {
		return fmt.Errorf("error while parsing MAC address %v", mac)
	}

	// Convert ipv4 string to []byte
	ipv4Octets := strings.Split(ip, ".")
	var parsedIP []byte
	for _, ipv4Octet := range ipv4Octets {
		ipv4IntPart, err := strconv.ParseInt(ipv4Octet, 0, 32)
		if err != nil {
			return fmt.Errorf("unable to parse ip address %s", ip)
		}
		parsedIP = append(parsedIP, byte(ipv4IntPart))
	}

	req := &vpe.BdIPMacAddDel{}
	req.BdID = bridgeDomainID
	req.IPAddress = parsedIP
	req.MacAddress = parsedMac
	req.IsIpv6 = 0
	req.IsAdd = 1

	reply := &vpe.BdIPMacAddDelReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("adding arp entry returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"Bridge domain": bridgeDomainID, "Mac": parsedMac, "Ip Address": ip}).Debug("Arp termination entry added.")

	return nil
}

// VppRemoveArpTerminationTableEntry removes ARP termination entry
func VppRemoveArpTerminationTableEntry(bdID uint32, mac string, ip string, log logging.Logger, vppChan *govppapi.Channel) error {
	log.Info("'Deleting' arp entry")

	parsedMac, errMac := net.ParseMAC(mac)
	if errMac != nil {
		return fmt.Errorf("error while parsing MAC address %v", mac)
	}

	// Convert ipv4 string to []byte
	ipv4Octets := strings.Split(ip, ".")
	var parsedIP []byte
	for _, ipv4Octet := range ipv4Octets {
		ipv4IntPart, err := strconv.ParseInt(ipv4Octet, 0, 32)
		if err != nil {
			return fmt.Errorf("unable to parse ip address %s", ip)
		}
		parsedIP = append(parsedIP, byte(ipv4IntPart))
	}

	req := &vpe.BdIPMacAddDel{}
	req.BdID = bdID
	req.MacAddress = parsedMac
	req.IPAddress = parsedIP
	req.IsAdd = 0

	reply := &vpe.BdIPMacAddDelReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("deleting arp entry returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"bdID": bdID, "Mac": parsedMac, "Ip Address": ip}).Debug("Arp termination entry removed.")

	return nil
}
