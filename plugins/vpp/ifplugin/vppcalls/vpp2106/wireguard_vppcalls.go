//  Copyright (c) 2020 Doc.ai and/or its affiliates.
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

package vpp2106

import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/wireguard"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// AddWireguardTunnel adds a new wireguard tunnel interface.
func (h *InterfaceVppHandler) AddWireguardTunnel(ifName string, wireguardLink *interfaces.WireguardLink) (uint32, error) {
	invalidIdx := ^uint32(0)
	if h.wireguard == nil {
		return invalidIdx, errors.WithMessage(vpp.ErrPluginDisabled, "wireguard")
	}

	wgItf := wireguard.WireguardInterface{
		UserInstance: ^uint32(0),
		Port:         uint16(wireguardLink.Port),
	}

	genKey := false
	if len(wireguardLink.PrivateKey) > 0 {
		publicKeyBin, err := base64.StdEncoding.DecodeString(wireguardLink.PrivateKey)
		if err != nil {
			return invalidIdx, err
		}
		wgItf.PrivateKey = publicKeyBin
	} else {
		genKey = true
	}

	srcAddr, err := IPToAddress(wireguardLink.SrcAddr)
	if err != nil {
		return invalidIdx, err
	}
	wgItf.SrcIP = srcAddr

	req := &wireguard.WireguardInterfaceCreate{
		Interface:   wgItf,
		GenerateKey: genKey,
	}

	// prepare reply
	reply := &wireguard.WireguardInterfaceCreateReply{}
	// send request and obtain reply
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return ^uint32(0), err
	}
	retSwIfIndex := uint32(reply.SwIfIndex)
	return retSwIfIndex, h.SetInterfaceTag(ifName, retSwIfIndex)
}

// DeleteWireguardTunnel removes wireguard tunnel interface.
func (h *InterfaceVppHandler) DeleteWireguardTunnel(ifName string, ifIdx uint32) error {
	if h.wireguard == nil {
		return errors.WithMessage(vpp.ErrPluginDisabled, "wireguard")
	}

	req := &wireguard.WireguardInterfaceDelete{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
	}
	// prepare reply
	reply := &wireguard.WireguardInterfaceDeleteReply{}
	// send request and obtain reply

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	return h.RemoveInterfaceTag(ifName, ifIdx);
}

// dumpWireguardDetails dumps wireguard interface details from VPP.
func (h *InterfaceVppHandler) dumpWireguardDetails(ifc map[uint32]*vppcalls.InterfaceDetails) error {
	if h.wireguard == nil {
		return nil
	}

	reqCtx := h.callsChannel.SendMultiRequest(&wireguard.WireguardInterfaceDump{})

	for {
		wgDetails := &wireguard.WireguardInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(wgDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump wireguard interface details: %v", err)
		}
		_, ifIdxExists := ifc[uint32(wgDetails.Interface.SwIfIndex)]
		if !ifIdxExists {
			h.log.Warnf("Wireguard interface dump: interface name for index %d not found", wgDetails.Interface.SwIfIndex)
			continue
		}

		wgLink := &interfaces.WireguardLink{
			Port:         uint32(wgDetails.Interface.Port),
		}
		wgLink.PrivateKey = base64.StdEncoding.EncodeToString(wgDetails.Interface.PrivateKey)

		srcAddr := wgDetails.Interface.SrcIP.ToIP()
		if !srcAddr.IsUnspecified() {
			wgLink.SrcAddr = srcAddr.String()
		}

		ifc[uint32(wgDetails.Interface.SwIfIndex)].Interface.Link = &interfaces.Interface_Wireguard { Wireguard: wgLink }
		ifc[uint32(wgDetails.Interface.SwIfIndex)].Interface.Type = interfaces.Interface_WIREGUARD_TUNNEL
	}
	return nil
}