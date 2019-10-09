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

	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	vpp_dhcp "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
)

func (h *DHCPProxyHandler) DumpDHCPProxy() ([]*vppcalls.DHCPProxyDetails, error) {
	var entry []*vppcalls.DHCPProxyDetails
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_dhcp.DHCPProxyDump{})
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

	reqCtx = h.callsChannel.SendMultiRequest(&vpp_dhcp.DHCPProxyDump{IsIP6: true})
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
	return entry, nil
}

func addressToString(address vpp_dhcp.Address) string {
	ipByte := make([]byte, 16)
	copy(ipByte[:], address.Un.XXX_UnionData[:])
	if address.Af == vpp_dhcp.ADDRESS_IP6 {
		return net.IP(ipByte).To16().String()
	}
	return net.IP(ipByte).To4().String()
}
