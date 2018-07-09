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
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/stn"
)

// StnRule represents stn rule entry
type StnRule struct {
	IPAddress net.IP
	IfaceIdx  uint32
}

func addDelStnRule(ifIdx uint32, addr *net.IP, isAdd bool, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(stn.StnAddDelRule{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// prepare the message
	req := &stn.StnAddDelRule{
		SwIfIndex: ifIdx,
		IsAdd:     boolToUint(isAdd),
	}

	isIPv6, err := addrs.IsIPv6(addr.String())
	if err != nil {
		return err
	}
	if isIPv6 {
		req.IPAddress = []byte(addr.To16())
		req.IsIP4 = 0
	} else {
		req.IPAddress = []byte(addr.To4())
		req.IsIP4 = 1
	}

	reply := &stn.StnAddDelRuleReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil

}

// AddStnRule calls StnAddDelRule bin API with IsAdd=1
func AddStnRule(ifIdx uint32, addr *net.IP, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return addDelStnRule(ifIdx, addr, true, vppChan, stopwatch)

}

// DelStnRule calls StnAddDelRule bin API with IsAdd=0
func DelStnRule(ifIdx uint32, addr *net.IP, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return addDelStnRule(ifIdx, addr, false, vppChan, stopwatch)
}
