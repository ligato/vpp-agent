package vpp1901

import (
	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/dhcp"
	"github.com/pkg/errors"
	"net"
)

func (h *DHCPProxyHandler) createDeleteDHCPProxy(entry *vpp_l3.DHCPProxy, delete bool) error {
	config := &dhcp.DHCPProxyConfig{
		RxVrfID:        entry.RxVrfId,
		IsAdd:          boolToUint(!delete),
	}
	ipAddr := net.ParseIP(entry.SourceIpAddress)
	if ipAddr == nil {
		return errors.Errorf("invalid IP address: %q", entry.SourceIpAddress)
	}

	if ipAddr.To4() == nil {
		config.IsIPv6 = 1
		config.DHCPSrcAddress = []byte(ipAddr.To16())
	} else {
		config.IsIPv6 = 1
		config.DHCPSrcAddress = []byte(ipAddr.To4())
	}

	for _, server := range entry.Servers {
		config.ServerVrfID = server.VrfId
		config.DHCPServer = []byte(net.ParseIP(server.IpAddress).To4())
		ipAddr := net.ParseIP(server.IpAddress)
		if ipAddr == nil {
			return errors.Errorf("invalid IP address: %q", server.IpAddress)
		}

		if ipAddr.To4() == nil {
			config.DHCPServer = []byte(ipAddr.To16())
		} else {
			config.DHCPServer = []byte(ipAddr.To4())
		}
		reply := &dhcp.DHCPProxyConfigReply{}
		if err := h.callsChannel.SendRequest(config).ReceiveReply(reply); err != nil {
			return err
		}
	}

	return nil
}

func (h *DHCPProxyHandler) CreateDHCPProxy(entry *vpp_l3.DHCPProxy) error {
	return h.createDeleteDHCPProxy(entry , false)
}

func (h *DHCPProxyHandler) DeleteDHCPProxy(entry *vpp_l3.DHCPProxy) error {
	return h.createDeleteDHCPProxy(entry , true)
}

