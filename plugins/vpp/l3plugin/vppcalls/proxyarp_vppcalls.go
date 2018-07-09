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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
)

// ProxyArpMessages is list of used VPP messages for compatibility check
var ProxyArpMessages = []govppapi.Message{
	&ip.ProxyArpIntfcEnableDisable{},
	&ip.ProxyArpIntfcEnableDisableReply{},
	&ip.ProxyArpAddDel{},
	&ip.ProxyArpAddDelReply{},
}

// EnableProxyArpInterface enables interface for proxy ARP
func EnableProxyArpInterface(swIfIdx uint32, vppChan govppapi.Channel, log logging.Logger, stopwatch *measure.Stopwatch) error {
	return vppAddDelProxyArpInterface(swIfIdx, vppChan, true, log, stopwatch)
}

// DisableProxyArpInterface disables interface for proxy ARP
func DisableProxyArpInterface(swIfIdx uint32, vppChan govppapi.Channel, log logging.Logger, stopwatch *measure.Stopwatch) error {
	return vppAddDelProxyArpInterface(swIfIdx, vppChan, false, log, stopwatch)
}

// AddProxyArpRange adds new IP range for proxy ARP
func AddProxyArpRange(firstIP, lastIP []byte, vppChan govppapi.Channel, log logging.Logger, stopwatch *measure.Stopwatch) error {
	return vppAddDelProxyArpRange(firstIP, lastIP, vppChan, true, log, stopwatch)
}

// DeleteProxyArpRange removes proxy ARP IP range
func DeleteProxyArpRange(firstIP, lastIP []byte, vppChan govppapi.Channel, log logging.Logger, stopwatch *measure.Stopwatch) error {
	return vppAddDelProxyArpRange(firstIP, lastIP, vppChan, false, log, stopwatch)
}

// vppAddDelProxyArpInterface adds or removes proxy ARP interface entry according to provided input
func vppAddDelProxyArpInterface(swIfIdx uint32, vppChan govppapi.Channel, enable bool, log logging.Logger, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ip.ProxyArpIntfcEnableDisable{}).LogTimeEntry(time.Since(t))
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
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	log.Debugf("interface %v enabled for proxy arp: %v", req.SwIfIndex, enable)

	return nil
}

// vppAddDelProxyArpRange adds or removes proxy ARP range according to provided input
func vppAddDelProxyArpRange(firstIP, lastIP []byte, vppChan govppapi.Channel, isAdd bool, log logging.Logger, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ip.ProxyArpAddDel{}).LogTimeEntry(time.Since(t))
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
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	log.Debugf("proxy arp range: %v - %v added: %v", req.Proxy.LowAddress, req.Proxy.HiAddress, isAdd)

	return nil
}
