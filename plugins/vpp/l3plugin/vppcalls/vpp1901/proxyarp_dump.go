//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package vpp1901

import (
	"fmt"
	"net"

	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	l3binapi "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
)

// DumpProxyArpRanges implements proxy arp handler.
func (h *ProxyArpVppHandler) DumpProxyArpRanges() (pArpRngs []*vppcalls.ProxyArpRangesDetails, err error) {
	reqCtx := h.callsChannel.SendMultiRequest(&l3binapi.ProxyArpDump{})

	for {
		proxyArpDetails := &l3binapi.ProxyArpDetails{}
		stop, err := reqCtx.ReceiveReply(proxyArpDetails)
		if stop {
			break
		}
		if err != nil {
			h.log.Error(err)
			return nil, err
		}

		pArpRngs = append(pArpRngs, &vppcalls.ProxyArpRangesDetails{
			Range: &l3.ProxyARP_Range{
				FirstIpAddr: fmt.Sprintf("%s", net.IP(proxyArpDetails.Proxy.LowAddress[:4]).To4().String()),
				LastIpAddr:  fmt.Sprintf("%s", net.IP(proxyArpDetails.Proxy.HiAddress[:4]).To4().String()),
			},
		})
	}

	return pArpRngs, nil
}

// DumpProxyArpInterfaces implements proxy arp handler.
func (h *ProxyArpVppHandler) DumpProxyArpInterfaces() (pArpIfs []*vppcalls.ProxyArpInterfaceDetails, err error) {
	reqCtx := h.callsChannel.SendMultiRequest(&l3binapi.ProxyArpIntfcDump{})

	for {
		proxyArpDetails := &l3binapi.ProxyArpIntfcDetails{}
		stop, err := reqCtx.ReceiveReply(proxyArpDetails)
		if stop {
			break
		}
		if err != nil {
			h.log.Error(err)
			return nil, err
		}

		// Interface
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(proxyArpDetails.SwIfIndex)
		if !exists {
			h.log.Warnf("Proxy ARP interface dump: missing name for interface index %d", proxyArpDetails.SwIfIndex)
		}

		// Create entry
		pArpIfs = append(pArpIfs, &vppcalls.ProxyArpInterfaceDetails{
			Interface: &l3.ProxyARP_Interface{
				Name: ifName,
			},
			Meta: &vppcalls.ProxyArpInterfaceMeta{
				SwIfIndex: proxyArpDetails.SwIfIndex,
			},
		})

	}

	return pArpIfs, nil
}
