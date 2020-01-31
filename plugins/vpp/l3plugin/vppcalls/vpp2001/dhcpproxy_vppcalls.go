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
	"fmt"
	"net"

	"github.com/pkg/errors"
	vpp_dhcp "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

func (h *DHCPProxyHandler) createDeleteDHCPProxy(entry *l3.DHCPProxy, delete bool) (err error) {
	config := &vpp_dhcp.DHCPProxyConfig{
		RxVrfID: entry.RxVrfId,
		IsAdd:   !delete,
	}
	if config.DHCPSrcAddress, err = ipToDHCPAddress(entry.SourceIpAddress); err != nil {
		return errors.Errorf("Invalid source IP address: %q", entry.SourceIpAddress)
	}
	for _, server := range entry.Servers {
		config.ServerVrfID = server.VrfId
		if config.DHCPServer, err = ipToDHCPAddress(server.IpAddress); err != nil {
			return errors.Errorf("Invalid server IP address: %q", server.IpAddress)
		}
		reply := &vpp_dhcp.DHCPProxyConfigReply{}
		if err := h.callsChannel.SendRequest(config).ReceiveReply(reply); err != nil {
			return err
		}
	}

	return nil
}

func (h *DHCPProxyHandler) CreateDHCPProxy(entry *l3.DHCPProxy) error {
	return h.createDeleteDHCPProxy(entry, false)
}

func (h *DHCPProxyHandler) DeleteDHCPProxy(entry *l3.DHCPProxy) error {
	return h.createDeleteDHCPProxy(entry, true)
}

func ipToDHCPAddress(address string) (dhcpAddr vpp_dhcp.Address, err error) {
	netIP := net.ParseIP(address)
	if netIP == nil {
		return vpp_dhcp.Address{}, fmt.Errorf("invalid IP: %q", address)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		dhcpAddr.Af = ip_types.ADDRESS_IP6
		var ip6addr vpp_dhcp.IP6Address
		copy(ip6addr[:], netIP.To16())
		dhcpAddr.Un.SetIP6(ip6addr)
	} else {
		dhcpAddr.Af = ip_types.ADDRESS_IP4
		var ip4addr vpp_dhcp.IP4Address
		copy(ip4addr[:], ip4)
		dhcpAddr.Un.SetIP4(ip4addr)
	}
	return
}
