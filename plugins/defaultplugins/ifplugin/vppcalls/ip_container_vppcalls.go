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
	"net"

	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/ip"
)

const (
	addContainerIP    uint8 = 1
	removeContainerIP uint8 = 0
)

// AddContainerIP calls IPContainerProxyAddDel VPP API with IsAdd=1
func AddContainerIP(ifIdx uint32, addr *net.IPNet, isIpv6 bool, log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// IPContainerProxyAddDelReply time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := prepareMessageForVpp(ifIdx, addr, isIpv6, addContainerIP)
	return sendAndLogMessageForVpp(ifIdx, req, "creat", log, vppChan)
}

// DelContainerIP calls IPContainerProxyAddDel VPP API with IsAdd=0
func DelContainerIP(ifIdx uint32, addr *net.IPNet, isIpv6 bool, log logging.Logger, vppChan *govppapi.Channel, timeLog *measure.TimeLog) error {
	// IPContainerProxyAddDelReply time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := prepareMessageForVpp(ifIdx, addr, isIpv6, removeContainerIP)
	return sendAndLogMessageForVpp(ifIdx, req, "delet", log, vppChan)
}

func prepareMessageForVpp(ifIdx uint32, addr *net.IPNet, isIpv6 bool, isAdd uint8) *ip.IPContainerProxyAddDel {
	req := &ip.IPContainerProxyAddDel{}
	req.SwIfIndex = ifIdx
	req.IsAdd = isAdd
	prefix, _ := addr.Mask.Size()
	req.Plen = byte(prefix)
	isIpv4 := !isIpv6
	if isIpv4 {
		req.IP = []byte(addr.IP.To4())
		req.IsIP4 = 1
	} else {
		req.IP = []byte(addr.IP.To16())
		req.IsIP4 = 0
	}
	return req
}

func sendAndLogMessageForVpp(ifIdx uint32, req *ip.IPContainerProxyAddDel, logActionType string, log logging.Logger, vppChan *govppapi.Channel) error {
	log.WithFields(logging.Fields{"isIpv4": req.IsIP4, "prefix": req.Plen, "address": req.IP, "if_index": ifIdx}).
		Debug("Container IP address ", logActionType, "ing...")

	// send the message
	reply := &ip.IPContainerProxyAddDelReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf(logActionType, "ing IP address returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"isIpv4": req.IsIP4, "prefix": req.Plen, "address": req.IP, "if_index": ifIdx}).
		Debug("Container IP address ", logActionType, "ed.")

	return nil
}
