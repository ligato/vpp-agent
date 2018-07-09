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
	"net"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/l2"
)

func callBdIPMacAddDel(isAdd bool, bdID uint32, mac string, ip string, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.BdIPMacAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &l2ba.BdIPMacAddDel{
		BdID: bdID,
	}

	macAddr, err := net.ParseMAC(mac)
	if err != nil {
		return err
	}
	req.MacAddress = macAddr

	isIpv6, err := addrs.IsIPv6(ip)
	if err != nil {
		return err
	}
	ipAddr := net.ParseIP(ip)
	if isIpv6 {
		req.IsIpv6 = 1
		req.IPAddress = []byte(ipAddr.To16())
	} else {
		req.IsIpv6 = 0
		req.IPAddress = []byte(ipAddr.To4())
	}

	if isAdd {
		req.IsAdd = 1
	} else {
		req.IsAdd = 0
	}

	reply := &l2ba.BdIPMacAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// VppAddArpTerminationTableEntry creates ARP termination entry for bridge domain.
func VppAddArpTerminationTableEntry(bdID uint32, mac string, ip string, log logging.Logger, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	log.Info("Adding ARP termination entry")

	err := callBdIPMacAddDel(true, bdID, mac, ip, vppChan, stopwatch)
	if err != nil {
		return err
	}

	log.WithFields(logging.Fields{"bdID": bdID, "MAC": mac, "IP": ip}).
		Debug("ARP termination entry added")

	return nil
}

// VppRemoveArpTerminationTableEntry removes ARP termination entry from bridge domain
func VppRemoveArpTerminationTableEntry(bdID uint32, mac string, ip string, log logging.Logger, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	log.Info("Removing ARP termination entry")

	err := callBdIPMacAddDel(false, bdID, mac, ip, vppChan, stopwatch)
	if err != nil {
		return err
	}

	log.WithFields(logging.Fields{"bdID": bdID, "MAC": mac, "IP": ip}).
		Debug("ARP termination entry removed")

	return nil
}
