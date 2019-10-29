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
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	vpp_ip "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001/ip"
	vpp_punt "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001/punt"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/puntplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_ip.AllMessages()...)
	msgs = append(msgs, vpp_punt.AllMessages()...)

	vppcalls.Versions["vpp2001"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(channel govppapi.Channel, index ifaceidx.IfaceMetadataIndex, logger logging.Logger) vppcalls.PuntVppAPI {
			return NewPuntVppHandler(channel, index, logger)
		},
	}
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
) *PuntVppHandler {
	return &PuntVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}

func ipToAddress(ipstr string) (addr vpp_ip.Address, err error) {
	netIP := net.ParseIP(ipstr)
	if netIP == nil {
		return vpp_ip.Address{}, fmt.Errorf("invalid IP: %q", ipstr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = vpp_ip.ADDRESS_IP6
		var ip6addr vpp_ip.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = vpp_ip.ADDRESS_IP4
		var ip4addr vpp_ip.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}
