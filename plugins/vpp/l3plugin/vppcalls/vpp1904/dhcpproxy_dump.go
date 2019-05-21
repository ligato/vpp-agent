package vpp1904

import (
	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"net"
)

func (h *DHCPProxyHandler) DumpDHCPProxy() (*vppcalls.DHCPProxyDetails, error) {
	var entry *vppcalls.DHCPProxyDetails
	reqCtx := h.callsChannel.SendMultiRequest(&dhcp.DHCPProxyDump{})
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

		var srcIP string
		if dhcpProxyDetails.IsIPv6 == 1 {
			srcIP = net.IP(dhcpProxyDetails.DHCPSrcAddress).To16().String()
		} else {
			srcIP = net.IP(dhcpProxyDetails.DHCPSrcAddress).To4().String()
		}

		proxy := &vpp_l3.DHCPProxy{
			RxVrfId: dhcpProxyDetails.RxVrfID,
			SourceIpAddress: srcIP,
		}

		for _, server := range dhcpProxyDetails.Servers {
			var ip string
			if dhcpProxyDetails.IsIPv6 == 1 {
				ip = net.IP(server.DHCPServer).To16().String()
			} else {
				ip = net.IP(server.DHCPServer).To4().String()

			}
			proxyServer := &vpp_l3.DHCPProxy_DHCPServer{
				IpAddress: ip,
				VrfId: server.ServerVrfID,
			}
			proxy.Servers = append(proxy.Servers, proxyServer)
		}

		entry = &vppcalls.DHCPProxyDetails{
			DHCPProxy:  proxy,
		}
	}
	return entry, nil

}
