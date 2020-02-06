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
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interface_types"
	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	vpp_punt "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/punt"
	punt "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/punt"
)

const PuntSocketHeaderVersion = 1

// Socket path from the VPP startup config file, returned when a punt socket
// is retrieved. Limited to single entry as supported in the VPP.
var vppConfigSocketPath string

// AddPunt configures new punt entry
func (h *PuntVppHandler) AddPunt(p *punt.ToHost) error {
	return errors.Errorf("passive punt add is currently not available")
}

// DeletePunt removes punt entry
func (h *PuntVppHandler) DeletePunt(p *punt.ToHost) error {
	return errors.Errorf("passive punt del is currently not available")
}

// AddPuntException adds new punt exception entry
func (h *PuntVppHandler) AddPuntException(p *punt.Exception) (string, error) {
	return h.addDelPuntException(p, true)
}

// DeletePuntException removes punt exception entry
func (h *PuntVppHandler) DeletePuntException(p *punt.Exception) error {
	_, err := h.addDelPuntException(p, false)
	return err
}

func (h *PuntVppHandler) addDelPuntException(p *punt.Exception, isAdd bool) (pathName string, err error) {
	reasons, err := h.dumpPuntReasons()
	if err != nil {
		return "", fmt.Errorf("dumping punt reasons failed: %v", err)
	}

	h.log.Debugf("dumped %d punt reasons: %+v", len(reasons), reasons)

	var reasonID *uint32
	for _, r := range reasons {
		if r.Reason.Name == p.Reason {
			id := r.ID
			reasonID = &id
			break
		}
	}
	if reasonID == nil {
		return "", fmt.Errorf("punt reason %q not found", p.Reason)
	}

	baPunt := getPuntExceptionConfig(*reasonID)

	if isAdd {
		h.log.Debugf("adding punt exception: %+v", p)
		pathName, err = h.handleRegisterPuntSocket(baPunt, p.SocketPath)
		if err != nil {
			return "", err
		}
	} else {
		err = h.handleDeregisterPuntSocket(baPunt)
		if err != nil {
			return "", err
		}
	}

	return pathName, nil
}

// RegisterPuntSocket registers new punt to unix domain socket entry
func (h *PuntVppHandler) RegisterPuntSocket(p *punt.ToHost) (pathName string, err error) {
	ipProto := resolveL4Proto(p.L4Protocol)

	if p.L3Protocol == punt.L3Protocol_IPV4 || p.L3Protocol == punt.L3Protocol_ALL {
		baPunt := getPuntL4Config(ip_types.ADDRESS_IP4, ipProto, uint16(p.Port))
		if pathName, err = h.handleRegisterPuntSocket(baPunt, p.SocketPath); err != nil {
			return "", err
		}
	}
	if p.L3Protocol == punt.L3Protocol_IPV6 || p.L3Protocol == punt.L3Protocol_ALL {
		baPunt := getPuntL4Config(ip_types.ADDRESS_IP6, ipProto, uint16(p.Port))
		if pathName, err = h.handleRegisterPuntSocket(baPunt, p.SocketPath); err != nil {
			return "", err
		}
	}

	return pathName, nil
}

// DeregisterPuntSocket removes existing punt to socket registration
func (h *PuntVppHandler) DeregisterPuntSocket(p *punt.ToHost) error {
	ipProto := resolveL4Proto(p.L4Protocol)

	if p.L3Protocol == punt.L3Protocol_IPV4 || p.L3Protocol == punt.L3Protocol_ALL {
		baPunt := getPuntL4Config(ip_types.ADDRESS_IP4, ipProto, uint16(p.Port))
		if err := h.handleDeregisterPuntSocket(baPunt); err != nil {
			return err
		}
	}
	if p.L3Protocol == punt.L3Protocol_IPV6 || p.L3Protocol == punt.L3Protocol_ALL {
		baPunt := getPuntL4Config(ip_types.ADDRESS_IP6, ipProto, uint16(p.Port))
		if err := h.handleDeregisterPuntSocket(baPunt); err != nil {
			return err
		}
	}

	return nil
}

func (h *PuntVppHandler) handleRegisterPuntSocket(punt vpp_punt.Punt, path string) (string, error) {
	req := &vpp_punt.PuntSocketRegister{
		HeaderVersion: PuntSocketHeaderVersion,
		Punt:          punt,
		Pathname:      path,
	}
	reply := &vpp_punt.PuntSocketRegisterReply{}

	h.log.Debugf("registering punt socket: %+v (pathname: %s)", req.Punt, req.Pathname)
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return "", err
	}

	// socket pathname from VPP config
	pathName := strings.SplitN(string(reply.Pathname), "\x00", 2)[0]

	// VPP startup config socket path name is always the same
	if vppConfigSocketPath != pathName {
		h.log.Debugf("setting vpp punt socket path to: %q (%s)", pathName, vppConfigSocketPath)
		vppConfigSocketPath = pathName
	}

	return pathName, nil
}

// DeregisterPuntSocket removes existing punt to socket registration
func (h *PuntVppHandler) handleDeregisterPuntSocket(punt vpp_punt.Punt) error {
	req := &vpp_punt.PuntSocketDeregister{
		Punt: punt,
	}
	reply := &vpp_punt.PuntSocketDeregisterReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func getPuntExceptionConfig(reasonID uint32) vpp_punt.Punt {
	p := vpp_punt.PuntException{
		ID: reasonID,
	}
	return vpp_punt.Punt{
		Type: vpp_punt.PUNT_API_TYPE_EXCEPTION,
		Punt: vpp_punt.PuntUnionException(p),
	}
}

func getPuntL4Config(ipv vpp_punt.AddressFamily, ipProto vpp_punt.IPProto, port uint16) vpp_punt.Punt {
	puntL4 := vpp_punt.PuntL4{
		Af:       ipv,
		Protocol: ipProto,
		Port:     port,
	}
	return vpp_punt.Punt{
		Type: vpp_punt.PUNT_API_TYPE_L4,
		Punt: vpp_punt.PuntUnionL4(puntL4),
	}
}

// AddPuntRedirect adds new redirect entry
func (h *PuntVppHandler) AddPuntRedirect(puntCfg *punt.IPRedirect) error {
	if puntCfg.L3Protocol == punt.L3Protocol_IPV4 || puntCfg.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handlePuntRedirectIPv4(puntCfg, true); err != nil {
			return err
		}
	}
	if puntCfg.L3Protocol == punt.L3Protocol_IPV6 || puntCfg.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handlePuntRedirectIPv6(puntCfg, true); err != nil {
			return err
		}
	}
	return nil
}

// DeletePuntRedirect removes existing redirect entry
func (h *PuntVppHandler) DeletePuntRedirect(puntCfg *punt.IPRedirect) error {
	if puntCfg.L3Protocol == punt.L3Protocol_IPV4 || puntCfg.L3Protocol == punt.L3Protocol_ALL {
		if err := h.handlePuntRedirectIPv4(puntCfg, false); err != nil {
			return err
		}
	}
	if puntCfg.L3Protocol == punt.L3Protocol_IPV6 || puntCfg.L3Protocol == punt.L3Protocol_ALL {
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

	req := &vpp_ip.IPPuntRedirect{
		IsAdd: isAdd,
		Punt: vpp_ip.PuntRedirect{
			RxSwIfIndex: interface_types.InterfaceIndex(rxIfIdx),
			TxSwIfIndex: interface_types.InterfaceIndex(txMetadata.SwIfIndex),
			Nh:          nextHop,
		},
	}
	reply := &vpp_ip.IPPuntRedirectReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func resolveL4Proto(protocol punt.L4Protocol) vpp_punt.IPProto {
	if protocol == punt.L4Protocol_UDP {
		return ip_types.IP_API_PROTO_UDP
	}
	return ip_types.IP_API_PROTO_TCP
}

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}
