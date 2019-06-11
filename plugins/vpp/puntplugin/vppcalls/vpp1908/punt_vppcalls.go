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
	"strings"

	"github.com/pkg/errors"

	punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	ba_ip "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/ip"
	ba_punt "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/punt"
)

const PuntSocketHeaderVersion = 1

// AddPunt configures new punt entry
func (h *PuntVppHandler) AddPunt(p *punt.ToHost) error {
	return errors.Errorf("passive punt add is currently now available")

	// return h.addDelPunt(p, true)
}

// DeletePunt removes punt entry
func (h *PuntVppHandler) DeletePunt(p *punt.ToHost) error {
	return errors.Errorf("passive punt del is currently now available")

	// return h.addDelPunt(p, false)
}

func (h *PuntVppHandler) addDelPunt(p *punt.ToHost, isAdd bool) error {
	ipProto := resolveL4Proto(p.L4Protocol)
	if p.L3Protocol == punt.L3Protocol_IPv4 || p.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handlePuntToHost(ba_punt.ADDRESS_IP4, ipProto, uint16(p.Port), isAdd); err != nil {
			return err
		}
	}
	if p.L3Protocol == punt.L3Protocol_IPv6 || p.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handlePuntToHost(ba_punt.ADDRESS_IP6, ipProto, uint16(p.Port), isAdd); err != nil {
			return err
		}
	}
	return nil
}

func (h *PuntVppHandler) handlePuntToHost(ipv ba_punt.AddressFamily, ipProto ba_punt.IPProto, port uint16, isAdd bool) error {
	req := &ba_punt.SetPunt{
		IsAdd: boolToUint(isAdd),
		Punt:  getPuntConfig(ipv, ipProto, port),
	}
	reply := &ba_punt.SetPuntReply{}

	h.log.Debugf("Setting punt: %+v", req.Punt)
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// RegisterPuntSocket registers new punt to unix domain socket entry
func (h *PuntVppHandler) RegisterPuntSocket(p *punt.ToHost) (pathName string, err error) {
	ipProto := resolveL4Proto(p.L4Protocol)
	if p.L3Protocol == punt.L3Protocol_IPv4 || p.L3Protocol == punt.L3Protocol_ALL {
		if pathName, err = h.handleRegisterPuntSocket(ba_punt.ADDRESS_IP4, ipProto, uint16(p.Port), p.SocketPath); err != nil {
			return "", err
		}
	}
	if p.L3Protocol == punt.L3Protocol_IPv6 || p.L3Protocol == punt.L3Protocol_ALL {
		if pathName, err = h.handleRegisterPuntSocket(ba_punt.ADDRESS_IP6, ipProto, uint16(p.Port), p.SocketPath); err != nil {
			return "", err
		}
	}

	return
}

// DeregisterPuntSocket removes existing punt to socket registration
func (h *PuntVppHandler) DeregisterPuntSocket(p *punt.ToHost) error {
	ipProto := resolveL4Proto(p.L4Protocol)
	if p.L3Protocol == punt.L3Protocol_IPv4 || p.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handleDeregisterPuntSocket(ba_punt.ADDRESS_IP4, ipProto, uint16(p.Port)); err != nil {
			return err
		}
	}
	if p.L3Protocol == punt.L3Protocol_IPv6 || p.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handleDeregisterPuntSocket(ba_punt.ADDRESS_IP6, ipProto, uint16(p.Port)); err != nil {
			return err
		}
	}

	return nil
}

func (h *PuntVppHandler) handleRegisterPuntSocket(ipv ba_punt.AddressFamily, ipProto ba_punt.IPProto, port uint16, path string) (string, error) {
	req := &ba_punt.PuntSocketRegister{
		HeaderVersion: PuntSocketHeaderVersion,
		Punt:          getPuntConfig(ipv, ipProto, port),
		Pathname:      []byte(path),
	}
	reply := &ba_punt.PuntSocketRegisterReply{}

	h.log.Debugf("Registering punt socket: %+v (pathname: %s)", req.Punt, req.Pathname)
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return "", err
	}

	return strings.SplitN(string(reply.Pathname), "\x00", 2)[0], nil
}

// DeregisterPuntSocket removes existing punt to socket registration
func (h *PuntVppHandler) handleDeregisterPuntSocket(ipv ba_punt.AddressFamily, ipProto ba_punt.IPProto, port uint16) error {
	req := &ba_punt.PuntSocketDeregister{
		Punt: getPuntConfig(ipv, ipProto, port),
	}
	reply := &ba_punt.PuntSocketDeregisterReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func getPuntConfig(ipv ba_punt.AddressFamily, ipProto ba_punt.IPProto, port uint16) ba_punt.Punt {
	puntL4 := ba_punt.PuntL4{
		Af:       ipv,
		Protocol: ipProto,
		Port:     port,
	}

	puntD := ba_punt.Punt{
		Type: ba_punt.PUNT_API_TYPE_L4,
	}
	puntD.Punt.SetL4(puntL4)

	return puntD

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

func resolveL4Proto(protocol punt.L4Protocol) ba_punt.IPProto {
	if protocol == punt.L4Protocol_UDP {
		return ba_punt.IP_API_PROTO_UDP
	}
	return ba_punt.IP_API_PROTO_TCP
}

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}
