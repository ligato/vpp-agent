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
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/ip"
)

type ARPEntry struct {
	Interface   string
	IpAddress   net.IP
	PhysAddress string
	Static      bool
}

// vppAddDelARP adds or removes ARP table according to provided input.
func vppAddDelARP(arpEntry *ARPEntry, vppChan *govppapi.Channel, delete bool, timeLog measure.StopWatchEntry) error {
	// Time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()
	req := &ip.IPNeighborAddDel{}
	if delete {
		req.IsAdd = 0
	} else {
		req.IsAdd = 1
	}
	isIpv6, err := addrs.IsIPv6(arpEntry.IpAddress.String())
	if err != nil {
		return err
	}
	if isIpv6 {
		req.IsIpv6 = 1
		req.DstAddress = []byte(arpEntry.IpAddress.To16())
	} else {
		req.IsIpv6 = 0
		req.DstAddress = []byte(arpEntry.IpAddress.To4())
	}
	req.MacAddress = []byte(arpEntry.PhysAddress)
	if arpEntry.Static {
		req.IsStatic = 1
	} else {
		req.IsStatic = 0
	}
	req.IsNoAdjFib = 1
	req.SwIfIndex

	// Send message
	reply := &ip.IPNeighborAddDelReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("IPNeighborAddDel returned %d", reply.Retval)
	}

	return nil
}

// VppAddARP adds ARP entry according to provided input.
func VppAddARP(arpEntry *ARPEntry, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	return vppAddDelARP(arpEntry, vppChan, false, timeLog)
}

// VppDelARP removes old ARP entry according to provided input.
func VppDelARP(arpEntry *ARPEntry, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	return vppAddDelARP(arpEntry, vppChan, true, timeLog)
}
