//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp2106

import (
	"fmt"
	"net"

	"github.com/pkg/errors"

	vpp_arp "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/arp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
)

// EnableProxyArpInterface implements proxy arp handler.
func (h *ProxyArpVppHandler) EnableProxyArpInterface(ifName string) error {
	return h.vppAddDelProxyArpInterface(ifName, true)
}

// DisableProxyArpInterface implements proxy arp handler.
func (h *ProxyArpVppHandler) DisableProxyArpInterface(ifName string) error {
	return h.vppAddDelProxyArpInterface(ifName, false)
}

// AddProxyArpRange implements proxy arp handler.
func (h *ProxyArpVppHandler) AddProxyArpRange(firstIP, lastIP []byte, vrfID uint32) error {
	return h.vppAddDelProxyArpRange(firstIP, lastIP, vrfID, true)
}

// DeleteProxyArpRange implements proxy arp handler.
func (h *ProxyArpVppHandler) DeleteProxyArpRange(firstIP, lastIP []byte, vrfID uint32) error {
	return h.vppAddDelProxyArpRange(firstIP, lastIP, vrfID, false)
}

// vppAddDelProxyArpInterface adds or removes proxy ARP interface entry according to provided input
func (h *ProxyArpVppHandler) vppAddDelProxyArpInterface(ifName string, enable bool) error {
	meta, found := h.ifIndexes.LookupByName(ifName)
	if !found {
		return errors.Errorf("interface %s not found", ifName)
	}

	req := &vpp_arp.ProxyArpIntfcEnableDisable{
		Enable:    enable,
		SwIfIndex: interface_types.InterfaceIndex(meta.SwIfIndex),
	}

	reply := &vpp_arp.ProxyArpIntfcEnableDisableReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	h.log.Debugf("interface %v enabled for proxy arp: %v", req.SwIfIndex, enable)

	return nil
}

// vppAddDelProxyArpRange adds or removes proxy ARP range according to provided input.
func (h *ProxyArpVppHandler) vppAddDelProxyArpRange(firstIP, lastIP []byte, vrfID uint32, isAdd bool) error {
	validateIPBytes := func(b []byte) error {
		ip := net.IP(b)
		if ip.To4() == nil {
			return fmt.Errorf("IP bytes %v are not valid IPv4 address", b)
		}
		return nil
	}
	if err := validateIPBytes(firstIP); err != nil {
		return fmt.Errorf("bad first IP: %v", err)
	}
	if err := validateIPBytes(lastIP); err != nil {
		return fmt.Errorf("bad last IP: %v", err)
	}

	proxy := vpp_arp.ProxyArp{
		TableID: vrfID,
	}
	copy(proxy.Low[:], firstIP)
	copy(proxy.Hi[:], lastIP)

	req := &vpp_arp.ProxyArpAddDel{
		IsAdd: isAdd,
		Proxy: proxy,
	}

	reply := &vpp_arp.ProxyArpAddDelReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	h.log.Debugf("proxy arp range: %v - %v (vrf: %d) added: %v", proxy.Low, proxy.Hi, proxy.TableID, isAdd)

	return nil
}
