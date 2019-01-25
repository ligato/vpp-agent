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

package vppcalls

import (
	"github.com/ligato/vpp-agent/api/models/vpp"
	"github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-binapi/binapi/punt"
)

// PuntDetails includes proto-modelled punt object and its socket path
type PuntDetails struct {
	PuntData   *vpp_punt.ToHost
	SocketPath string
}

// FIXME: temporary solutions for providing data in dump
var socketPathMap = map[uint32]*vpp.PuntToHost{}

// DumpRegisteredPuntSockets returns punt to host via registered socket entries
// TODO since the binary API is not available, all data are read from local cache for now
func (h *PuntVppHandler) DumpRegisteredPuntSockets() (punts []*PuntDetails, err error) {
	// TODO: use binapi dumps
	if _, err := h.dumpPunts(false); err != nil {
		h.log.Errorf("punt dump failed: %v", err)
	}
	if _, err := h.dumpPunts(true); err != nil {
		h.log.Errorf("punt dump failed: %v", err)
	}
	if _, err := h.dumpPuntSockets(false); err != nil {
		h.log.Errorf("punt socket dump failed: %v", err)
	}
	if _, err := h.dumpPuntSockets(true); err != nil {
		h.log.Errorf("punt socket dump failed: %v", err)
	}

	for _, punt := range socketPathMap {
		punts = append(punts, &PuntDetails{
			PuntData:   punt,
			SocketPath: punt.SocketPath,
		})
	}

	if len(punts) > 0 {
		h.log.Warnf("Dump punt socket register: all entries were read from local cache")
	}

	return punts, nil
}

func (h *PuntVppHandler) dumpPuntSockets(ipv6 bool) (punts []*PuntDetails, err error) {
	req := h.callsChannel.SendMultiRequest(&punt.PuntSocketDump{
		IsIPv6: boolToUint(ipv6),
	})
	h.log.Debugf("=> dumping punt sockets (IPv6:%v)", ipv6)
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

		punts = append(punts, &PuntDetails{
			PuntData: &vpp_punt.ToHost{
				Port:       uint32(d.Punt.L4Port), // FIXME: this seems to return 0 when registering ALL
				L3Protocol: parseL3Proto(d.Punt.IPv),
				L4Protocol: parseL4Proto(d.Punt.L4Protocol),
			},
		})
	}

	return punts, nil
}

func (h *PuntVppHandler) dumpPunts(ipv6 bool) (punts []*PuntDetails, err error) {
	req := h.callsChannel.SendMultiRequest(&punt.PuntDump{
		IsIPv6: boolToUint(ipv6),
	})
	h.log.Debugf("=> dumping punts (IPv6:%v)", ipv6)
	for {
		d := &punt.PuntDetails{}
		stop, err := req.ReceiveReply(d)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		h.log.Debugf(" - dumped punt: %+v", d.Punt)

		punts = append(punts, &PuntDetails{
			PuntData: &vpp_punt.ToHost{
				Port:       uint32(d.Punt.L4Port),
				L3Protocol: parseL3Proto(d.Punt.IPv),
				L4Protocol: parseL4Proto(d.Punt.L4Protocol),
			},
		})
	}

	return punts, nil
}
