// Copyright (c) 2017 Cisco and/or its affiliates.
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

package vppcalls

import (
	"fmt"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
)

// ProxyArpMessages is list of used VPP messages for compatibility check
var ProxyArpMessages = []govppapi.Message{
	&ip.ProxyArpIntfcEnableDisable{},
	&ip.ProxyArpIntfcEnableDisableReply{},
	&ip.ProxyArpAddDel{},
	&ip.ProxyArpAddDelReply{},
}

func (handler *proxyArpVppHandler) EnableProxyArpInterface(swIfIdx uint32) error {
	return handler.vppAddDelProxyArpInterface(swIfIdx, true)
}

func (handler *proxyArpVppHandler) DisableProxyArpInterface(swIfIdx uint32) error {
	return handler.vppAddDelProxyArpInterface(swIfIdx, false)
}

func (handler *proxyArpVppHandler) AddProxyArpRange(firstIP, lastIP []byte) error {
	return handler.vppAddDelProxyArpRange(firstIP, lastIP, true)
}

func (handler *proxyArpVppHandler) DeleteProxyArpRange(firstIP, lastIP []byte) error {
	return handler.vppAddDelProxyArpRange(firstIP, lastIP, false)
}

// vppAddDelProxyArpInterface adds or removes proxy ARP interface entry according to provided input
func (handler *proxyArpVppHandler) vppAddDelProxyArpInterface(swIfIdx uint32, enable bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(ip.ProxyArpIntfcEnableDisable{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ip.ProxyArpIntfcEnableDisable{}
	if enable {
		req.EnableDisable = 1
	} else {
		req.EnableDisable = 0
	}
	req.SwIfIndex = swIfIdx

	// Send message
	reply := &ip.ProxyArpIntfcEnableDisableReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	handler.log.Debugf("interface %v enabled for proxy arp: %v", req.SwIfIndex, enable)

	return nil
}

// vppAddDelProxyArpRange adds or removes proxy ARP range according to provided input
func (handler *proxyArpVppHandler) vppAddDelProxyArpRange(firstIP, lastIP []byte, isAdd bool) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(ip.ProxyArpAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ip.ProxyArpAddDel{}
	if isAdd {
		req.IsAdd = 1
	} else {
		req.IsAdd = 0
	}
	req.Proxy = ip.ProxyArp{
		LowAddress: firstIP,
		HiAddress:  lastIP,
	}

	// Send message
	reply := &ip.ProxyArpAddDelReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	handler.log.Debugf("proxy arp range: %v - %v added: %v", req.Proxy.LowAddress, req.Proxy.HiAddress, isAdd)

	return nil
}
