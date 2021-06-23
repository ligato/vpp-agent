// Copyright (c) 2020 Pantheon.tech
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

package vpp2106

import (
	"fmt"
	"net"

	"github.com/go-errors/errors"
	vpp_dns "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/dns"
)

// EnableDNS make act VPP as DNS cache server
func (h *DNSVppHandler) EnableDNS() error {
	h.log.Debug("Enabling DNS functionality of VPP")
	if err := h.enableDisableDNS(true); err != nil {
		return err
	}
	h.log.Debug("DNS functionality of VPP enabled.")
	return nil
}

// DisableDNS disables functionality that makes VPP act as DNS cache server
func (h *DNSVppHandler) DisableDNS() error {
	h.log.Debug("Disabling DNS functionality of VPP")
	if err := h.enableDisableDNS(false); err != nil {
		return err
	}
	h.log.Debug("DNS functionality of VPP disabled.")
	return nil
}

func (h *DNSVppHandler) enableDisableDNS(enable bool) error {
	req := &vpp_dns.DNSEnableDisable{}
	if enable {
		req.Enable = 1
	}
	reply := &vpp_dns.DNSEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}
	return nil
}

// AddUpstreamDNSServer adds new upstream DNS Server to the upstream DNS server list
func (h *DNSVppHandler) AddUpstreamDNSServer(serverIPAddress net.IP) error {
	h.log.Debug("Adding upstream DNS server with IP %s", serverIPAddress)
	if err := h.addRemoveUpstreamDNSServer(true, serverIPAddress); err != nil {
		return err
	}
	h.log.Debug("Upstream DNS server with IP %s was added", serverIPAddress)
	return nil
}

// DeleteUpstreamDNSServer removes upstream DNS Server from the upstream DNS server list
func (h *DNSVppHandler) DeleteUpstreamDNSServer(serverIPAddress net.IP) error {
	h.log.Debug("Removing upstream DNS server with IP %s", serverIPAddress)
	if err := h.addRemoveUpstreamDNSServer(false, serverIPAddress); err != nil {
		return err
	}
	h.log.Debug("Upstream DNS server with IP %s was removed", serverIPAddress)
	return nil
}

func (h *DNSVppHandler) addRemoveUpstreamDNSServer(addition bool, serverIPAddress net.IP) error {
	if serverIPAddress == nil {
		return errors.New("upstream DNS server IP address can't be nil")
	}
	req := &vpp_dns.DNSNameServerAddDel{
		IsAdd: boolToUint(addition),
	}
	if serverIPAddress.To4() == nil { // IPv6
		req.IsIP6 = 1
		req.ServerAddress = serverIPAddress.To16()
	} else {
		req.ServerAddress = serverIPAddress.To4()
	}
	reply := &vpp_dns.DNSNameServerAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}
	return nil
}

func boolToUint(input bool) uint8 {
	if input {
		return uint8(1)
	}
	return uint8(0)
}
