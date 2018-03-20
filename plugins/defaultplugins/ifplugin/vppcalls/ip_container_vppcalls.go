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
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ip"
)

const (
	addContainerIP    uint8 = 1
	removeContainerIP uint8 = 0
)

func sendAndLogMessageForVpp(ifIdx uint32, addr string, isAdd uint8, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ip.IPContainerProxyAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ip.IPContainerProxyAddDel{
		SwIfIndex: ifIdx,
		IsAdd:     isAdd,
	}

	IPaddr, isIPv6, err := addrs.ParseIPWithPrefix(addr)
	if err != nil {
		return err
	}

	prefix, _ := IPaddr.Mask.Size()
	req.Plen = byte(prefix)
	if isIPv6 {
		req.IP = []byte(IPaddr.IP.To16())
		req.IsIP4 = 0
	} else {
		req.IP = []byte(IPaddr.IP.To4())
		req.IsIP4 = 1
	}

	reply := &ip.IPContainerProxyAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// AddContainerIP calls IPContainerProxyAddDel VPP API with IsAdd=1
func AddContainerIP(ifIdx uint32, addr string, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return sendAndLogMessageForVpp(ifIdx, addr, addContainerIP, vppChan, stopwatch)
}

// DelContainerIP calls IPContainerProxyAddDel VPP API with IsAdd=0
func DelContainerIP(ifIdx uint32, addr string, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return sendAndLogMessageForVpp(ifIdx, addr, removeContainerIP, vppChan, stopwatch)
}
