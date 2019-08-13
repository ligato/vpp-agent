// Copyright (c) 2018 Cisco and/or its affiliates.
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

package vpp1901

import (
	"strings"

	"github.com/pkg/errors"

	punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	ba_ip "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/ip"
	ba_punt "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/punt"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

const PuntSocketHeaderVersion = 1

// AddPunt configures new punt entry
func (h *PuntVppHandler) AddPunt(p *punt.ToHost) error {
	return h.handlePuntToHost(p, true)
}

// DeletePunt removes punt entry
func (h *PuntVppHandler) DeletePunt(p *punt.ToHost) error {
	return h.handlePuntToHost(p, false)
}

func (h *PuntVppHandler) handlePuntToHost(toHost *punt.ToHost, isAdd bool) error {
	req := &ba_punt.SetPunt{
		IsAdd: boolToUint(isAdd),
		Punt: ba_punt.Punt{
			IPv:        resolveL3Proto(toHost.L3Protocol),
			L4Protocol: resolveL4Proto(toHost.L4Protocol),
			L4Port:     uint16(toHost.Port),
		},
	}
	reply := &ba_punt.SetPuntReply{}

	h.log.Debugf("Setting punt: %+v", req.Punt)
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// RegisterPuntSocket registers new punt to socket
func (h *PuntVppHandler) RegisterPuntSocket(toHost *punt.ToHost) (string, error) {
	req := &ba_punt.PuntSocketRegister{
		HeaderVersion: PuntSocketHeaderVersion,
		Punt: ba_punt.Punt{
			IPv:        resolveL3Proto(toHost.L3Protocol),
			L4Protocol: resolveL4Proto(toHost.L4Protocol),
			L4Port:     uint16(toHost.Port),
		},
		Pathname: []byte(toHost.SocketPath),
	}
	reply := &ba_punt.PuntSocketRegisterReply{}

	h.log.Debugf("Registering punt socket: %+v (pathname: %s)", req.Punt, req.Pathname)
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return "", err
	}
	h.log.Debugf("Punt socket registered with %s", reply.Pathname)

	p := *toHost
	p.SocketPath = strings.SplitN(string(reply.Pathname), "\x00", 2)[0]
	socketPathMap[toHost.Port] = &p

	return p.SocketPath, nil
}

// DeregisterPuntSocket removes existing punt to socket sogistration
func (h *PuntVppHandler) DeregisterPuntSocket(toHost *punt.ToHost) error {
	req := &ba_punt.PuntSocketDeregister{
		Punt: ba_punt.Punt{
			IPv:        resolveL3Proto(toHost.L3Protocol),
			L4Protocol: resolveL4Proto(toHost.L4Protocol),
			L4Port:     uint16(toHost.Port),
		},
	}
	reply := &ba_punt.PuntSocketDeregisterReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	delete(socketPathMap, toHost.Port)

	return nil
}

// AddPuntRedirect adds new redirect entry
func (h *PuntVppHandler) AddPuntRedirect(puntCfg *punt.IPRedirect) error {
	if puntCfg.L3Protocol == punt.L3Protocol_IPv4 || puntCfg.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handlePuntRedirectIPv4(puntCfg, true); err != nil {
			return err
		}
	}
	if puntCfg.L3Protocol == punt.L3Protocol_IPv6 || puntCfg.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handlePuntRedirectIPv6(puntCfg, true); err != nil {
			return err
		}
	}
	return nil
}

// DeletePuntRedirect removes existing redirect entry
func (h *PuntVppHandler) DeletePuntRedirect(puntCfg *punt.IPRedirect) error {
	if puntCfg.L3Protocol == punt.L3Protocol_IPv4 || puntCfg.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handlePuntRedirectIPv4(puntCfg, false); err != nil {
			return err
		}
	}
	if puntCfg.L3Protocol == punt.L3Protocol_IPv6 || puntCfg.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handlePuntRedirectIPv6(puntCfg, false); err != nil {
			return err
		}
	}
	return nil
}

func (h *PuntVppHandler) handlePuntRedirectIPv4(punt *punt.IPRedirect, isAdd bool) error {
	return h.handlePuntRedirect(punt, true, isAdd)
}

func (h *PuntVppHandler) handlePuntRedirectIPv6(punt *punt.IPRedirect, isAdd bool) error {
	return h.handlePuntRedirect(punt, false, isAdd)
}

func (h *PuntVppHandler) handlePuntRedirect(punt *punt.IPRedirect, isIPv4, isAdd bool) error {
	// rx interface
	var rxIfIdx uint32
	if punt.RxInterface == "" {
		rxIfIdx = ^uint32(0)
	} else {
		rxMetadata, exists := h.ifIndexes.LookupByName(punt.RxInterface)
		if !exists {
			return errors.Errorf("index not found for interface %s", punt.RxInterface)
		}
		rxIfIdx = rxMetadata.SwIfIndex
	}

	// tx interface
	txMetadata, exists := h.ifIndexes.LookupByName(punt.TxInterface)
	if !exists {
		return errors.Errorf("index not found for interface %s", punt.TxInterface)
	}

	// next hop address
	//  - remove mask from IP address if necessary
	nextHopStr := punt.NextHop
	ipParts := strings.Split(punt.NextHop, "/")
	if len(ipParts) > 1 {
		h.log.Debugf("IP punt redirect next hop IP address %s is defined with mask, removing it")
		nextHopStr = ipParts[0]
	}
	nextHop, err := ipToAddress(nextHopStr)
	if err != nil {
		return err
	}

	req := &ba_ip.IPPuntRedirect{
		IsAdd: boolToUint(isAdd),
		Punt: ba_ip.PuntRedirect{
			RxSwIfIndex: rxIfIdx,
			TxSwIfIndex: txMetadata.SwIfIndex,
			Nh:          nextHop,
		},
	}
	reply := &ba_ip.IPPuntRedirectReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *PuntVppHandler) AddPuntException(punt *punt.Exception) (string, error) {
	return "", vppcalls.ErrUnsupported
}

func (h *PuntVppHandler) DeletePuntException(punt *punt.Exception) error {
	return vppcalls.ErrUnsupported
}

func parseL3Proto(p uint8) punt.L3Protocol {
	switch p {
	case uint8(punt.L3Protocol_IPv4), uint8(punt.L3Protocol_IPv6):
		return punt.L3Protocol(p)
	case ^uint8(0):
		return punt.L3Protocol_ALL
	}
	return punt.L3Protocol_UNDEFINED_L3
}

func parseL4Proto(p uint8) punt.L4Protocol {
	switch p {
	case uint8(punt.L4Protocol_TCP):
		return punt.L4Protocol_TCP
	case uint8(punt.L4Protocol_UDP):
		return punt.L4Protocol_UDP
	}
	return punt.L4Protocol_UNDEFINED_L4
}

func resolveL3Proto(protocol punt.L3Protocol) uint8 {
	switch protocol {
	case punt.L3Protocol_IPv4:
		return uint8(punt.L3Protocol_IPv4)
	case punt.L3Protocol_IPv6:
		return uint8(punt.L3Protocol_IPv6)
	case punt.L3Protocol_ALL:
		return ^uint8(0) // binary API representation for both protocols
	}
	return uint8(punt.L3Protocol_UNDEFINED_L3)
}

func resolveL4Proto(protocol punt.L4Protocol) uint8 {
	switch protocol {
	case punt.L4Protocol_TCP:
		return uint8(punt.L4Protocol_TCP)
	case punt.L4Protocol_UDP:
		return uint8(punt.L4Protocol_UDP)
	}
	return uint8(punt.L4Protocol_UNDEFINED_L4)
}

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}
