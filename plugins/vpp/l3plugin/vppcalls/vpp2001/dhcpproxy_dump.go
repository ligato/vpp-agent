// Copyright (c) 2019 PANTHEON.tech
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

package vpp2001

import (
	"net"

	vpp_dhcp "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

func (h *DHCPProxyHandler) DumpDHCPProxy() (entry []*vppcalls.DHCPProxyDetails, err error) {
	ipv4Entries, err := h.dumpDHCPProxyForIPVersion(false)
	if err != nil {
		return nil, err
	}
	entry = append(entry, ipv4Entries...)
	ipv6Entries, err := h.dumpDHCPProxyForIPVersion(true)
	if err != nil {
		return nil, err
	}

	entry = append(entry, ipv6Entries...)
	return entry, nil
}

func (h *DHCPProxyHandler) dumpDHCPProxyForIPVersion(isIPv6 bool) (entry []*vppcalls.DHCPProxyDetails, err error) {
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_dhcp.DHCPProxyDump{IsIP6: isIPv6})
	for {
		dhcpProxyDetails := &vpp_dhcp.DHCPProxyDetails{}
		stop, err := reqCtx.ReceiveReply(dhcpProxyDetails)
		if stop {
			break
		}
		if err != nil {
			h.log.Error(err)
			return nil, err
		}
		proxy := &l3.DHCPProxy{
			RxVrfId:         dhcpProxyDetails.RxVrfID,
			SourceIpAddress: addressToString(dhcpProxyDetails.DHCPSrcAddress),
		}

		for _, server := range dhcpProxyDetails.Servers {
			proxyServer := &l3.DHCPProxy_DHCPServer{
				IpAddress: addressToString(server.DHCPServer),
				VrfId:     server.ServerVrfID,
			}
			proxy.Servers = append(proxy.Servers, proxyServer)
		}

		entry = append(entry, &vppcalls.DHCPProxyDetails{
			DHCPProxy: proxy,
		})
	}
	return
}

func addressToString(address vpp_dhcp.Address) string {
	if address.Af == ip_types.ADDRESS_IP6 {
		ipAddr := address.Un.GetIP6()
		return net.IP(ipAddr[:]).To16().String()
	}
	ipAddr := address.Un.GetIP4()
	return net.IP(ipAddr[:]).To4().String()
}
