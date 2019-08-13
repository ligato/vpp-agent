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
	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/vpp-agent/pkg/idxvpp"
	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/l2"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
)

func init() {
	vppcalls.Versions["vpp1904"] = vppcalls.HandlerVersion{
		Msgs: l2ba.Messages,
		New: func(ch govppapi.Channel,
			ifIdx ifaceidx.IfaceMetadataIndex, bdIdx idxvpp.NameToIndex, log logging.Logger,
		) vppcalls.L2VppAPI {
			return NewL2VppHandler(ch, ifIdx, bdIdx, log)
		},
	}
}

type L2VppHandler struct {
	*BridgeDomainVppHandler
	*FIBVppHandler
	*XConnectVppHandler
}

func NewL2VppHandler(ch govppapi.Channel,
	ifIdx ifaceidx.IfaceMetadataIndex, bdIdx idxvpp.NameToIndex, log logging.Logger,
) *L2VppHandler {
	return &L2VppHandler{
		BridgeDomainVppHandler: newBridgeDomainVppHandler(ch, ifIdx, log),
		FIBVppHandler:          newFIBVppHandler(ch, ifIdx, bdIdx, log),
		XConnectVppHandler:     newXConnectVppHandler(ch, ifIdx, log),
	}
}

// BridgeDomainVppHandler is accessor for bridge domain-related vppcalls methods.
type BridgeDomainVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// FIBVppHandler is accessor for FIB-related vppcalls methods.
type FIBVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	bdIndexes    idxvpp.NameToIndex
	log          logging.Logger
}

// XConnectVppHandler is accessor for cross-connect-related vppcalls methods.
type XConnectVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// NewBridgeDomainVppHandler creates new instance of bridge domain vppcalls handler.
func newBridgeDomainVppHandler(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) *BridgeDomainVppHandler {
	return &BridgeDomainVppHandler{
		callsChannel: ch,
		ifIndexes:    ifIdx,
		log:          log,
	}
}

// NewFIBVppHandler creates new instance of FIB vppcalls handler.
func newFIBVppHandler(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, bdIndexes idxvpp.NameToIndex, log logging.Logger) *FIBVppHandler {
	return &FIBVppHandler{
		callsChannel: ch,
		ifIndexes:    ifIdx,
		bdIndexes:    bdIndexes,
		log:          log,
	}
}

// NewXConnectVppHandler creates new instance of cross connect vppcalls handler.
func newXConnectVppHandler(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) *XConnectVppHandler {
	return &XConnectVppHandler{
		callsChannel: ch,
		ifIndexes:    ifIdx,
		log:          log,
	}
}

func ipToAddress(ipstr string) (addr l2ba.Address, err error) {
	netIP := net.ParseIP(ipstr)
	if netIP == nil {
		return l2ba.Address{}, fmt.Errorf("invalid IP: %q", ipstr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = l2ba.ADDRESS_IP6
		var ip6addr l2ba.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = l2ba.ADDRESS_IP4
		var ip4addr l2ba.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}
