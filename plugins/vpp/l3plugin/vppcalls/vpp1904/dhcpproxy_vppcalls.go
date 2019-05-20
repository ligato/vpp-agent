package vpp1904

import (
	"github.com/ligato/cn-infra/utils/addrs"
	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/dhcp"
	"github.com/pkg/errors"
	"net"
)

func (h *DHCPProxyHandler) createDeleteDHCPProxy(entry *vpp_l3.DHCPProxy, delete bool) error {

	for _, server := range entry.Servers {
		_, err := ipToAddress(server.ServerIpAddress)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	_, err := ipToAddress(entry.SourceIpAddress)
	if err != nil {
		return errors.WithStack(err)
	}

	isIPv6, err := addrs.IsIPv6(entry.SourceIpAddress)
	if err != nil {
		return errors.WithStack(err)
	}

	config := &dhcp.DHCPProxyConfig{
		RxVrfID:        entry.RxFibId,
		IsIPv6:         boolToUint(isIPv6),
		IsAdd: 			boolToUint(!delete),
		DHCPSrcAddress: net.ParseIP(entry.SourceIpAddress),
	}

	for _, server := range entry.Servers {
		config.ServerVrfID = server.ServerFibId
		config.DHCPServer = net.ParseIP(server.ServerIpAddress)
		reply := &dhcp.DHCPProxyConfigReply{}
		if err = h.callsChannel.SendRequest(config).ReceiveReply(reply); err != nil {
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

