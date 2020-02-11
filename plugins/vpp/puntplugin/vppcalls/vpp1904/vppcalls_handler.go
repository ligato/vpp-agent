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

package vpp1904

import (
	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904"
	ba_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/ip"
	ba_punt "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/punt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/puntplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, ba_ip.AllMessages()...)
	msgs = append(msgs, ba_punt.AllMessages()...)

	vppcalls.AddHandlerVersion(vpp1904.Version, msgs, NewPuntVppHandler)
}

// PuntVppHandler is accessor for punt-related vppcalls methods.
type PuntVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// NewPuntVppHandler creates new instance of punt vppcalls handler
func NewPuntVppHandler(
	callsChan govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger,
) vppcalls.PuntVppAPI {
	return &PuntVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}

func ipToAddress(ipstr string) (addr ba_ip.Address, err error) {
	netIP := net.ParseIP(ipstr)
	if netIP == nil {
		return ba_ip.Address{}, fmt.Errorf("invalid IP: %q", ipstr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = ba_ip.ADDRESS_IP6
		var ip6addr ba_ip.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ba_ip.ADDRESS_IP4
		var ip4addr ba_ip.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}
