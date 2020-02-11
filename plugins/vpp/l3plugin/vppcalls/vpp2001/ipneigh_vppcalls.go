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

package vpp2001

import (
	"strconv"

	//vpp_ip_neighbor "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_neighbor"
	//"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

/*
	FIXME: IP neighbor configuraton is not implemented for 20.01, because
 	of breaking change in the API.
	New proto model must be defined to support configuring this properly.
	The current model does not allow separated config for IPv4 and IPv6.
*/

// DefaultIPScanNeighbor implements ip neigh handler.
func (h *IPNeighHandler) DefaultIPScanNeighbor() *l3.IPScanNeighbor {
	return nil

	/*return &l3.IPScanNeighbor{
		Mode:           l3.IPScanNeighbor_DISABLED,
		MaxProcTime:    0,
		MaxUpdate:      50000,
		ScanInterval:   0,
		ScanIntDelay:   0,
		StaleThreshold: 0,
	}*/
}

// SetIPScanNeighbor implements ip neigh handler.
func (h *IPNeighHandler) SetIPScanNeighbor(data *l3.IPScanNeighbor) (err error) {
	return vppcalls.ErrIPNeighborNotImplemented

	/*switch data.Mode {
	case l3.IPScanNeighbor_IPV4:
		return h.setIPScanNeighbor(ip_types.ADDRESS_IP4, data.MaxUpdate, data.MaxProcTime, recycle)
	case l3.IPScanNeighbor_IPV6:
		return h.setIPScanNeighbor(ip_types.ADDRESS_IP6, data.MaxUpdate, data.MaxProcTime, recycle)
	case l3.IPScanNeighbor_BOTH:
		err = h.setIPScanNeighbor(ip_types.ADDRESS_IP4, data.MaxUpdate, data.MaxProcTime, recycle)
		if err != nil {
			return err
		}
		err = h.setIPScanNeighbor(ip_types.ADDRESS_IP6, data.MaxUpdate, data.MaxProcTime, recycle)
		if err != nil {
			return err
		}
	case l3.IPScanNeighbor_DISABLED:
		err = h.setIPScanNeighbor(ip_types.ADDRESS_IP4, 0, 0, false)
		if err != nil {
			return err
		}
		err = h.setIPScanNeighbor(ip_types.ADDRESS_IP6, 0, 0, false)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown IP Scan Neighbor mode: %v", data.Mode)
	}
	return nil*/
}

/*func (h *IPNeighHandler) setIPScanNeighbor(af ip_types.AddressFamily, maxNum, maxAge uint32, recycle bool) error {
	req := &vpp_ip_neighbor.IPNeighborConfig{
		Af:        af,
		MaxNumber: maxNum,
		MaxAge:    maxAge,
		Recycle:   recycle,
	}
	reply := &vpp_ip_neighbor.IPNeighborConfigReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}*/

var (
/*
	Sample outputs for VPP CLI 'show ip neighbor-config'
	---
	ip4:
	  limit:50000, age:0, recycle:0
	ip6:
	  limit:50000, age:0, recycle:0
	---
*/
//cliIPScanNeighRe = regexp.MustCompile(`(ip4|ip6):\n\s+limit:([0-9]+),\s+age:([0-9]+),\s+recycle:([0-9]+)\s+`)
)

// GetIPScanNeighbor dumps current IP Scan Neighbor configuration.
func (h *IPNeighHandler) GetIPScanNeighbor() (*l3.IPScanNeighbor, error) {
	return nil, vppcalls.ErrIPNeighborNotImplemented

	/*data, err := h.RunCli(context.TODO(), "show ip neighbor-config")
	if err != nil {
		return nil, err
	}

	allMatches := cliIPScanNeighRe.FindAllStringSubmatch(data, 2)

	fmt.Printf("%d MATCHES:\n%q\n", len(allMatches), allMatches)

	if len(allMatches) != 2 || len(allMatches[0]) != 5 || len(allMatches[1]) != 5 {
		h.log.Warnf("invalid 'show ip neighbor-config' output: %q", data)
		return nil, errors.Errorf("invalid VPP CLI output for ip neighbor config")
	}

	ipScanNeigh := &l3.IPScanNeighbor{}

	for _, matches := range allMatches {
		switch matches[1] {
		case "ip4":
			ipScanNeigh.Mode = l3.IPScanNeighbor_IPV4
		case "ip6":
			ipScanNeigh.Mode = l3.IPScanNeighbor_IPV6
		}
		ipScanNeigh.MaxUpdate = h.strToUint32(matches[2])
		ipScanNeigh.MaxProcTime = h.strToUint32(matches[3])
		ipScanNeigh.ScanInterval = h.strToUint32(matches[4])
	}

	return ipScanNeigh, nil*/
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
