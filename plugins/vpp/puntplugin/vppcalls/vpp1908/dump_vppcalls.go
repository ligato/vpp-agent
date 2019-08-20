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
	"bytes"
	"net"

	vpp_punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/punt"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

// DumpPuntRedirect dumps ip redirect punts
func (h *PuntVppHandler) DumpPuntRedirect() (punts []*vpp_punt.IPRedirect, err error) {
	punt4, err := h.dumpPuntRedirect(false)
	if err != nil {
		return nil, err
	}
	punts = append(punts, punt4...)

	punt6, err := h.dumpPuntRedirect(true)
	if err != nil {
		return nil, err
	}
	punts = append(punts, punt6...)

	return punts, nil
}

func (h *PuntVppHandler) dumpPuntRedirect(ipv6 bool) (punts []*vpp_punt.IPRedirect, err error) {
	req := h.callsChannel.SendMultiRequest(&ip.IPPuntRedirectDump{
		SwIfIndex: ^uint32(0),
		IsIPv6:    boolToUint(ipv6),
	})
	for {
		d := &ip.IPPuntRedirectDetails{}
		stop, err := req.ReceiveReply(d)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		rxIface, _, exists := h.ifIndexes.LookupBySwIfIndex(d.Punt.RxSwIfIndex)
		if !exists {
			h.log.Warnf("RX interface (%v) not found", d.Punt.RxSwIfIndex)
			continue
		}
		txIface, _, exists := h.ifIndexes.LookupBySwIfIndex(d.Punt.TxSwIfIndex)
		if !exists {
			h.log.Warnf("TX interface (%v) not found", d.Punt.TxSwIfIndex)
			continue
		}

		var l3proto vpp_punt.L3Protocol
		var nextHop string

		if d.Punt.Nh.Af == ip.ADDRESS_IP4 {
			l3proto = vpp_punt.L3Protocol_IPv4
			addr := d.Punt.Nh.Un.GetIP4()
			nextHop = net.IP(addr[:]).To4().String()
		} else if d.Punt.Nh.Af == ip.ADDRESS_IP6 {
			l3proto = vpp_punt.L3Protocol_IPv6
			addr := d.Punt.Nh.Un.GetIP6()
			nextHop = net.IP(addr[:]).To16().String()
		} else {
			h.log.Warnf("invalid address family (%v)", d.Punt.Nh.Af)
			continue
		}

		punts = append(punts, &vpp_punt.IPRedirect{
			L3Protocol:  l3proto,
			RxInterface: rxIface,
			TxInterface: txIface,
			NextHop:     nextHop,
		})
	}

	return punts, nil
}

// DumpExceptions returns dump of registered punt exceptions.
func (h *PuntVppHandler) DumpExceptions() (punts []*vppcalls.ExceptionDetails, err error) {
	reasons, err := h.dumpPuntReasons()
	if err != nil {
		return nil, err
	}
	reasonMap := make(map[uint32]string, len(reasons))
	for _, r := range reasons {
		reasonMap[r.ID] = r.Reason.Name
	}

	if punts, err = h.dumpPuntExceptions(reasonMap); err != nil {
		return nil, err
	}

	return punts, nil
}

func (h *PuntVppHandler) dumpPuntExceptions(reasons map[uint32]string) (punts []*vppcalls.ExceptionDetails, err error) {
	req := h.callsChannel.SendMultiRequest(&punt.PuntSocketDump{
		Type: punt.PUNT_API_TYPE_EXCEPTION,
	})
	for {
		d := &punt.PuntSocketDetails{}
		stop, err := req.ReceiveReply(d)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		if d.Punt.Type != punt.PUNT_API_TYPE_EXCEPTION {
			h.log.Warnf("VPP returned invalid punt type in exception punt dump: %v", d.Punt.Type)
			continue
		}

		puntData := d.Punt.Punt.GetException()
		reason := reasons[puntData.ID]
		socketPath := string(bytes.Trim(d.Pathname, "\x00"))

		punts = append(punts, &vppcalls.ExceptionDetails{
			Exception: &vpp_punt.Exception{
				Reason:     reason,
				SocketPath: vppConfigSocketPath,
			},
			SocketPath: socketPath,
		})
	}

	return punts, nil
}

// DumpRegisteredPuntSockets returns punt to host via registered socket entries
func (h *PuntVppHandler) DumpRegisteredPuntSockets() (punts []*vppcalls.PuntDetails, err error) {
	if punts, err = h.dumpPuntL4(); err != nil {
		return nil, err
	}

	return punts, nil
}

func (h *PuntVppHandler) dumpPuntL4() (punts []*vppcalls.PuntDetails, err error) {
	req := h.callsChannel.SendMultiRequest(&punt.PuntSocketDump{
		Type: punt.PUNT_API_TYPE_L4,
	})
	for {
		d := &punt.PuntSocketDetails{}
		stop, err := req.ReceiveReply(d)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		if d.Punt.Type != punt.PUNT_API_TYPE_L4 {
			h.log.Warnf("VPP returned invalid punt type in L4 punt dump: %v", d.Punt.Type)
			continue
		}

		puntData := d.Punt.Punt.GetL4()
		socketPath := string(bytes.Trim(d.Pathname, "\x00"))

		punts = append(punts, &vppcalls.PuntDetails{
			PuntData: &vpp_punt.ToHost{
				Port:       uint32(puntData.Port),
				L3Protocol: parseL3Proto(puntData.Af),
				L4Protocol: parseL4Proto(puntData.Protocol),
				SocketPath: vppConfigSocketPath,
			},
			SocketPath: socketPath,
		})
	}

	return punts, nil
}

// DumpPuntReasons returns all known punt reasons from VPP
func (h *PuntVppHandler) DumpPuntReasons() (reasons []*vppcalls.ReasonDetails, err error) {
	if reasons, err = h.dumpPuntReasons(); err != nil {
		return nil, err
	}

	return reasons, nil
}

func (h *PuntVppHandler) dumpPuntReasons() (reasons []*vppcalls.ReasonDetails, err error) {
	req := h.callsChannel.SendMultiRequest(&punt.PuntReasonDump{})
	for {
		d := &punt.PuntReasonDetails{}
		stop, err := req.ReceiveReply(d)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		reasons = append(reasons, &vppcalls.ReasonDetails{
			Reason: &vpp_punt.Reason{
				Name: d.Reason.Name,
			},
			ID: d.Reason.ID,
		})
	}

	return reasons, nil
}

func parseL3Proto(p punt.AddressFamily) vpp_punt.L3Protocol {
	switch p {
	case punt.ADDRESS_IP4:
		return vpp_punt.L3Protocol_IPv4
	case punt.ADDRESS_IP6:
		return vpp_punt.L3Protocol_IPv6
	}
	return vpp_punt.L3Protocol_UNDEFINED_L3
}

func parseL4Proto(p punt.IPProto) vpp_punt.L4Protocol {
	switch p {
	case punt.IP_API_PROTO_TCP:
		return vpp_punt.L4Protocol_TCP
	case punt.IP_API_PROTO_UDP:
		return vpp_punt.L4Protocol_UDP
	}
	return vpp_punt.L4Protocol_UNDEFINED_L4
}
