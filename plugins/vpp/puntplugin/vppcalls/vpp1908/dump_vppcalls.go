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

	vpp_punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/punt"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

// DumpRegisteredPuntSockets returns punt to host via registered socket entries
func (h *PuntVppHandler) DumpRegisteredPuntSockets() (punts []*vppcalls.PuntDetails, err error) {
	// TODO: use set_punt dumps from binapi
	if _, err := h.dumpPunts(); err != nil {
		h.log.Errorf("punt dump failed: %v", err)
	}
	if punts, err = h.dumpPuntSockets(); err != nil {
		h.log.Errorf("punt socket dump failed: %v", err)
	}

	return punts, nil
}

func (h *PuntVppHandler) dumpPuntSockets() (punts []*vppcalls.PuntDetails, err error) {
	h.log.Debug("=> dumping punt sockets")

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
		h.log.Debugf(" - dumped punt socket (%s): %+v", d.Pathname, d.Punt)

		puntL4Data := d.Punt.Punt.GetL4()
		punts = append(punts, &vppcalls.PuntDetails{
			PuntData: &vpp_punt.ToHost{
				Port: uint32(puntL4Data.Port),
				// FIXME: L3Protocol seems to return 0 when registering ALL
				L3Protocol: parseL3Proto(puntL4Data.Af),
				L4Protocol: parseL4Proto(puntL4Data.Protocol),
			},
			SocketPath: string(bytes.Trim(d.Pathname, "\x00")),
		})
	}

	return punts, nil
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

func (h *PuntVppHandler) dumpPunts() (punts []*vppcalls.PuntDetails, err error) {
	h.log.Debugf("=> dumping punts")

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
		h.log.Debugf(" - dumped punt: %+v", d.Reason)

		// TODO Re-enable with the Punt-To-Host
		//punts = append(punts, &vppcalls.PuntDetails{
		//	PuntData: &vpp_punt.ToHost{
		//		Port:       uint32(d.Punt.L4Port),
		//		L3Protocol: parseL3Proto(d.Punt.IPv),
		//		L4Protocol: parseL4Proto(d.Punt.L4Protocol),
		//	},
		//})
	}

	return punts, nil
}
