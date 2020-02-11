//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp1908

import (
	"context"
	"regexp"
	"strconv"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ip"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// DefaultIPScanNeighbor implements ip neigh handler.
func (h *IPNeighHandler) DefaultIPScanNeighbor() *l3.IPScanNeighbor {
	return &l3.IPScanNeighbor{
		Mode:           l3.IPScanNeighbor_DISABLED,
		MaxProcTime:    20,
		MaxUpdate:      10,
		ScanInterval:   1,
		ScanIntDelay:   1,
		StaleThreshold: 4,
	}
}

// SetIPScanNeighbor implements ip neigh  handler.
func (h *IPNeighHandler) SetIPScanNeighbor(data *l3.IPScanNeighbor) error {
	req := &ip.IPScanNeighborEnableDisable{
		Mode:           uint8(data.Mode),
		ScanInterval:   uint8(data.ScanInterval),
		MaxProcTime:    uint8(data.MaxProcTime),
		MaxUpdate:      uint8(data.MaxUpdate),
		ScanIntDelay:   uint8(data.ScanIntDelay),
		StaleThreshold: uint8(data.StaleThreshold),
	}
	reply := &ip.IPScanNeighborEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

/*
	Sample outputs for VPP CLI 'show ip scan-neighbor'
	---
	IP neighbor scan disabled - current time is 5.5101 sec
	---
	IP neighbor scan enabled for IPv4 neighbors - current time is 95133.3063 sec
	   Full_scan_interval: 1 min  Stale_purge_threshod: 4 min
	   Max_process_time: 20 usec  Max_updates 10  Delay_to_resume_after_max_limit: 231 msec
	---
	IP neighbor scan enabled for IPv4 and IPv6 neighbors - current time is 95.6033 sec
	   Full_scan_interval: 1 min  Stale_purge_threshod: 4 min
	   Max_process_time: 20 usec  Max_updates 10  Delay_to_resume_after_max_limit: 1 msec
	---
*/
var (
	cliIPScanNeighRe = regexp.MustCompile(`IP neighbor scan (disabled|enabled)(?: for (IPv4|IPv6|IPv4 and IPv6) neighbors)? - current time is [0-9\.]+ sec(?:
\s+Full_scan_interval: ([0-9]+) min\s+Stale_purge_threshod: ([0-9]+) min
\s+Max_process_time: ([0-9]+) usec\s+Max_updates ([0-9]+)\s+Delay_to_resume_after_max_limit: ([0-9]+) msec)?`)
)

// GetIPScanNeighbor dumps current IP Scan Neighbor configuration.
func (h *IPNeighHandler) GetIPScanNeighbor() (*l3.IPScanNeighbor, error) {
	data, err := h.RunCli(context.TODO(), "show ip scan-neighbor")
	if err != nil {
		return nil, err
	}

	ipScanNeigh := &l3.IPScanNeighbor{}

	matches := cliIPScanNeighRe.FindStringSubmatch(data)

	if len(matches) != 8 {
		h.log.Warnf("invalid 'show ip scan-neighbor' output: %q", data)
		return nil, errors.Errorf("invalid 'show ip scan-neighbor' output")
	}

	if matches[1] == "enabled" {
		switch matches[2] {
		case "IPv4":
			ipScanNeigh.Mode = l3.IPScanNeighbor_IPV4
		case "IPv6":
			ipScanNeigh.Mode = l3.IPScanNeighbor_IPV6
		case "IPv4 and IPv6":
			ipScanNeigh.Mode = l3.IPScanNeighbor_BOTH
		}
	}
	ipScanNeigh.ScanInterval = h.strToUint32(matches[3])
	ipScanNeigh.StaleThreshold = h.strToUint32(matches[4])
	ipScanNeigh.MaxProcTime = h.strToUint32(matches[5])
	ipScanNeigh.MaxUpdate = h.strToUint32(matches[6])
	ipScanNeigh.ScanIntDelay = h.strToUint32(matches[7])

	return ipScanNeigh, nil
}

func (h *IPNeighHandler) strToUint32(s string) uint32 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		h.log.Error(err)
	}
	return uint32(n)
}
