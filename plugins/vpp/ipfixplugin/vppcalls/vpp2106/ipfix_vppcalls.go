//  Copyright (c) 2020 Cisco and/or its affiliates.
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
	"bytes"
	"errors"
	"fmt"
	"net"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	vpp_ipfix "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ipfix_export"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls"
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

// SetExporter configures IP Flow Information eXport (IPFIX).
func (h *IpfixVppHandler) SetExporter(conf *ipfix.IPFIX) error {
	collectorAddr, err := prepareAddress(conf.GetCollector().GetAddress())
	if err != nil {
		return fmt.Errorf("bad collector address: %v", err)
	}

	sourceAddr, err := prepareAddress(conf.GetSourceAddress())
	if err != nil {
		return fmt.Errorf("bad source address: %v", err)
	}

	collectorPort := uint16(conf.GetCollector().GetPort())
	if collectorPort == 0 {
		// Will be set by VPP to the default value: 4739.
		collectorPort = ^uint16(0)
	}

	mtu := conf.GetPathMtu()
	if mtu == 0 {
		// Will be set by VPP to the default value: 512 bytes.
		mtu = ^uint32(0)
	} else if mtu < vppcalls.MinPathMTU || mtu > vppcalls.MaxPathMTU {
		err := errors.New("path MTU is not in allowed range")
		return err
	}

	tmplInterval := conf.GetTemplateInterval()
	if tmplInterval == 0 {
		// Will be set by VPP to the default value: 20 sec.
		tmplInterval = ^uint32(0)
	}

	req := &vpp_ipfix.SetIpfixExporter{
		CollectorAddress: collectorAddr,
		CollectorPort:    collectorPort,
		SrcAddress:       sourceAddr,
		VrfID:            conf.GetVrfId(),
		PathMtu:          mtu,
		TemplateInterval: tmplInterval,
	}
	reply := &vpp_ipfix.SetIpfixExporterReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// prepareAddress validates and converts IP address, defined as a string,
// to the type which represents address in VPP binary API.
func prepareAddress(addrStr string) (ip_types.Address, error) {
	var a ip_types.Address

	addr := net.ParseIP(addrStr)
	if addr == nil {
		err := errors.New("can not parse address")
		return a, err
	}
	if addr.To4() == nil {
		err := errors.New("IPv6 is not supported")
		return a, err
	}

	var addrBytes [4]byte
	copy(addrBytes[:], addr.To4())
	if bytes.Equal(addrBytes[:], []byte{0, 0, 0, 0}) {
		err := errors.New("address must not be all zeros")
		return a, err
	}

	a = ip_types.Address{
		Af: ip_types.ADDRESS_IP4,
		Un: ip_types.AddressUnionIP4(addrBytes),
	}

	return a, nil
}
