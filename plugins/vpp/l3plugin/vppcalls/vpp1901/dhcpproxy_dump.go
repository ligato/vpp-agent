package vpp1901

import (
	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"net"
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
			RxVrfId: dhcpProxyDetails.RxVrfID,
			SourceIpAddress: net.IP(dhcpProxyDetails.DHCPSrcAddress[:4]).To4().String(),
		}

		for _, server := range dhcpProxyDetails.Servers {
			proxyServer := &vpp_l3.DHCPProxy_DHCPServer{
				IpAddress: net.IP(server.DHCPServer[:4]).To4().String(),
				VrfId: server.ServerVrfID,
			}
			proxy.Servers = append(proxy.Servers, proxyServer)
		}

		entry = append(entry, &vppcalls.DHCPProxyDetails{
			DHCPProxy:  proxy,
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
			RxVrfId: dhcpProxyDetails.RxVrfID,
			SourceIpAddress: net.IP(dhcpProxyDetails.DHCPSrcAddress).To16().String(),
		}

		for _, server := range dhcpProxyDetails.Servers {
			proxyServer := &vpp_l3.DHCPProxy_DHCPServer{
				IpAddress: net.IP(server.DHCPServer).To16().String(),
				VrfId: server.ServerVrfID,
			}
			proxy.Servers = append(proxy.Servers, proxyServer)
		}

		entry = append(entry, &vppcalls.DHCPProxyDetails{
			DHCPProxy:  proxy,
		})
	}
	return entry, nil

}
