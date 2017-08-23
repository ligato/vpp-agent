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
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/vpe"
	"net"
	"strconv"
	"strings"
)

// VppAddArpTerminationTableEntry adds ARP termination entry
func VppAddArpTerminationTableEntry(bridgeDomainID uint32, mac string, ip string, vppChan *govppapi.Channel) error {
	log.DefaultLogger().Println("Adding arp termination entry")

	parsedMac, errMac := net.ParseMAC(mac)
	if errMac != nil {
		return fmt.Errorf("Error while parsing MAC address %v", mac)
	}

	// Convert ipv4 string to []byte
	ipv4Octets := strings.Split(ip, ".")
	var parsedIP []byte
	for _, ipv4Octet := range ipv4Octets {
		ipv4IntPart, err := strconv.ParseInt(ipv4Octet, 0, 32)
		if err != nil {
			return fmt.Errorf("Unable to parse ip address %s", ip)
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
		return fmt.Errorf("Adding arp entry returned %d", reply.Retval)
	}
	log.DefaultLogger().WithFields(log.Fields{"Bridge domain": bridgeDomainID, "Mac": parsedMac, "Ip Address": ip}).Debug("Arp termination entry added.")

	return nil
}

// VppRemoveArpTerminationTableEntry removes ARP termination entry
func VppRemoveArpTerminationTableEntry(bdID uint32, mac string, ip string, vppChan *govppapi.Channel) error {
	log.DefaultLogger().Println("'Deleting' arp entry")

	parsedMac, errMac := net.ParseMAC(mac)
	if errMac != nil {
		return fmt.Errorf("Error while parsing MAC address %v", mac)
	}

	// Convert ipv4 string to []byte
	ipv4Octets := strings.Split(ip, ".")
	var parsedIP []byte
	for _, ipv4Octet := range ipv4Octets {
		ipv4IntPart, err := strconv.ParseInt(ipv4Octet, 0, 32)
		if err != nil {
			return fmt.Errorf("Unable to parse ip address %s", ip)
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
		return fmt.Errorf("Deleting arp entry returned %d", reply.Retval)
	}
	log.DefaultLogger().WithFields(log.Fields{"bdID": bdID, "Mac": parsedMac, "Ip Address": ip}).Debug("Arp termination entry removed.")

	return nil
}
