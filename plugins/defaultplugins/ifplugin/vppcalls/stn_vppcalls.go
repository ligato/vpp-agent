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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/stn"
	"time"
)

type StnRule struct {
	IpAddress net.IPNet
	IfaceIdx  uint32
}

// AddStnRule calls StnAddDelRule bin API with IsAdd=1
func AddStnRule(ifIdx uint32, addr *net.IPNet, log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// StnAddDelRule time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// prepare the message
	req := &stn.StnAddDelRule{}
	req.SwIfIndex = ifIdx
	req.IsAdd = 1

	v6, err := addrs.IsIPv6(addr.IP.String())
	if err != nil {
		return err
	}
	if v6 {
		req.IPAddress = []byte(addr.IP.To16())
		req.IsIP4 = 0
	} else {
		req.IPAddress = []byte(addr.IP.To4())
		req.IsIP4 = 1
	}

	log.Debug("stn rule add req: IPAdress: ", req.IPAddress, "interface: ", req.SwIfIndex)

	reply := &stn.StnAddDelRuleReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("stn rule adding returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"IPAddress": addr.IP, "mask": addr.Mask, "ifIdx": ifIdx}).Debug("rule added.")

	return nil

}

// DelStnRule calls StnAddDelRule bin API with IsAdd=00
func DelStnRule(ifIdx uint32, addr *net.IPNet, log logging.Logger, vppChan *govppapi.Channel, timeLog *measure.TimeLog) error {
	// StnAddDelRuleReply time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// prepare the message
	req := &stn.StnAddDelRule{}
	req.SwIfIndex = ifIdx
	req.IsAdd = 0

	v6, err := addrs.IsIPv6(addr.IP.String())
	if err != nil {
		return err
	}
	if v6 {
		req.IPAddress = []byte(addr.IP.To16())
		req.IsIP4 = 0
	} else {
		req.IPAddress = []byte(addr.IP.To4())
		req.IsIP4 = 1
	}

	log.Debug("stn rule del req: IPAdress: ", req.IPAddress, "interface: ", req.SwIfIndex)

	// send the message
	reply := &stn.StnAddDelRuleReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("stn rule del returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"IPAddress": addr.IP, "mask": addr.Mask, "ifIdx": ifIdx}).Debug("rule removed.")

	return nil
}
