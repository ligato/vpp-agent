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
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
)

// ArpMessages is list of used VPP messages for compatibility check
var ArpMessages = []govppapi.Message{
	&ip.IPNeighborAddDel{},
	&ip.IPNeighborAddDelReply{},
}

// ArpEntry represents ARP entry for interface
type ArpEntry struct {
	Interface  uint32
	IPAddress  net.IP
	MacAddress net.HardwareAddr
	Static     bool
}

// vppAddDelArp adds or removes ARP entry according to provided input
func vppAddDelArp(entry *ArpEntry, vppChan govppapi.Channel, delete bool, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ip.IPNeighborAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ip.IPNeighborAddDel{}
	if delete {
		req.IsAdd = 0
	} else {
		req.IsAdd = 1
	}

	isIpv6, err := addrs.IsIPv6(entry.IPAddress.String())
	if err != nil {
		return err
	}
	if isIpv6 {
		req.IsIpv6 = 1
		req.DstAddress = []byte(entry.IPAddress.To16())
	} else {
		req.IsIpv6 = 0
		req.DstAddress = []byte(entry.IPAddress.To4())
	}
	if entry.Static {
		req.IsStatic = 1
	} else {
		req.IsStatic = 0
	}
	req.MacAddress = []byte(entry.MacAddress)
	req.IsNoAdjFib = 1
	req.SwIfIndex = entry.Interface

	// Send message
	reply := &ip.IPNeighborAddDelReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// VppAddArp adds ARP entry according to provided input
func VppAddArp(entry *ArpEntry, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return vppAddDelArp(entry, vppChan, false, stopwatch)
}

// VppDelArp removes old ARP entry according to provided input
func VppDelArp(entry *ArpEntry, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return vppAddDelArp(entry, vppChan, true, stopwatch)
}
