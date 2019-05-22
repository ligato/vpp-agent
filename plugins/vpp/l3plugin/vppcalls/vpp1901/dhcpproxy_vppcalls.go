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

package vpp1901

import (
	"net"

	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/dhcp"
	"github.com/pkg/errors"
)

func (h *DHCPProxyHandler) createDeleteDHCPProxy(entry *vpp_l3.DHCPProxy, delete bool) error {
	config := &dhcp.DHCPProxyConfig{
		RxVrfID: entry.RxVrfId,
		IsAdd:   boolToUint(!delete),
	}

	ipAddr := net.ParseIP(entry.SourceIpAddress)

	if ipAddr == nil {
		return errors.Errorf("Invalid source IP address: %q", entry.SourceIpAddress)
	}

	if ipAddr.To4() == nil {
		config.IsIPv6 = 1
		config.DHCPSrcAddress = []byte(ipAddr.To16())
	} else {
		config.IsIPv6 = 0
		config.DHCPSrcAddress = []byte(ipAddr.To4())
	}

	for _, server := range entry.Servers {
		config.ServerVrfID = server.VrfId
		ipAddr := net.ParseIP(server.IpAddress)
		if ipAddr == nil {
			return errors.Errorf("Invalid server IP address: %q", server.IpAddress)
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
	return h.createDeleteDHCPProxy(entry, false)
}

func (h *DHCPProxyHandler) DeleteDHCPProxy(entry *vpp_l3.DHCPProxy) error {
	return h.createDeleteDHCPProxy(entry, true)
}
