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

package vpp1904

import (
	"net"

	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
)

func (h *DHCPProxyHandler) DumpDHCPProxy() ([]*vppcalls.DHCPProxyDetails, error) {
	var entry []*vppcalls.DHCPProxyDetails
	reqCtx := h.callsChannel.SendMultiRequest(&dhcp.DHCPProxyDump{IsIP6: 0})
	for {
		dhcpProxyDetails := &dhcp.DHCPProxyDetails{}
		stop, err := reqCtx.ReceiveReply(dhcpProxyDetails)
		if stop {
			break
		}
		if err != nil {
			h.log.Error(err)
			return nil, err
		}
		proxy := &vpp_l3.DHCPProxy{
			RxVrfId:         dhcpProxyDetails.RxVrfID,
			SourceIpAddress: net.IP(dhcpProxyDetails.DHCPSrcAddress[:4]).To4().String(),
		}

		for _, server := range dhcpProxyDetails.Servers {
			proxyServer := &vpp_l3.DHCPProxy_DHCPServer{
				IpAddress: net.IP(server.DHCPServer[:4]).To4().String(),
				VrfId:     server.ServerVrfID,
			}
			proxy.Servers = append(proxy.Servers, proxyServer)
		}

		entry = append(entry, &vppcalls.DHCPProxyDetails{
			DHCPProxy: proxy,
		})
	}

	reqCtx = h.callsChannel.SendMultiRequest(&dhcp.DHCPProxyDump{IsIP6: 1})
	for {
		dhcpProxyDetails := &dhcp.DHCPProxyDetails{}
		stop, err := reqCtx.ReceiveReply(dhcpProxyDetails)
		if stop {
			break
		}
		if err != nil {
			h.log.Error(err)
			return nil, err
		}
		proxy := &vpp_l3.DHCPProxy{
			RxVrfId:         dhcpProxyDetails.RxVrfID,
			SourceIpAddress: net.IP(dhcpProxyDetails.DHCPSrcAddress).To16().String(),
		}
		for _, server := range dhcpProxyDetails.Servers {
			proxyServer := &vpp_l3.DHCPProxy_DHCPServer{
				IpAddress: net.IP(server.DHCPServer).To16().String(),
				VrfId:     server.ServerVrfID,
			}
			proxy.Servers = append(proxy.Servers, proxyServer)
		}

		entry = append(entry, &vppcalls.DHCPProxyDetails{
			DHCPProxy: proxy,
		})
	}
	return entry, nil

}
