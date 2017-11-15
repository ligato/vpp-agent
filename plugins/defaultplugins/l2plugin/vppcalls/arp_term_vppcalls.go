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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/l2"
)

// VppAddArpTerminationTableEntry creates ARP termination entry for bridge domain.
func VppAddArpTerminationTableEntry(bdID uint32, mac string, ip string,
	log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	log.Info("Adding ARP termination entry")

	err := callBdIPMacAddDel(true, bdID, mac, ip, vppChan, timeLog)
	if err != nil {
		return err
	}

	log.WithFields(logging.Fields{"bdID": bdID, "MAC": mac, "IP": ip}).
		Debug("ARP termination entry added")

	return nil
}

// VppRemoveArpTerminationTableEntry removes ARP termination entry from bridge domain
func VppRemoveArpTerminationTableEntry(bdID uint32, mac string, ip string, log logging.Logger,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	log.Info("Removing ARP termination entry")

	err := callBdIPMacAddDel(false, bdID, mac, ip, vppChan, timeLog)
	if err != nil {
		return err
	}

	log.WithFields(logging.Fields{"bdID": bdID, "MAC": mac, "IP": ip}).
		Debug("ARP termination entry removed")

	return nil
}

func callBdIPMacAddDel(isAdd bool, bdID uint32, mac string, ip string,
	vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// BdIPMacAddDel time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	macAddr, err := net.ParseMAC(mac)
	if err != nil {
		return err
	}
	ipAddr := []byte(net.ParseIP(ip).To4())
	if ipAddr == nil {
		return fmt.Errorf("invalid IP address: %q", ipAddr)
	}

	req := &l2.BdIPMacAddDel{
		BdID:       bdID,
		IPAddress:  ipAddr,
		MacAddress: macAddr,
		IsIpv6:     0,
	}
	if isAdd {
		req.IsAdd = 1
	} else {
		req.IsAdd = 0
	}

	reply := &l2.BdIPMacAddDelReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}
