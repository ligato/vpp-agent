// Copyright (c) 2018 Cisco and/or its affiliates.
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

	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/dhcp"
)

func handleInterfaceDHCP(ifIdx uint32, hostName string, isAdd bool, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(dhcp.DhcpClientConfig{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &dhcp.DhcpClientConfig{
		SwIfIndex:     ifIdx,
		Hostname:      []byte(hostName),
		WantDhcpEvent: 1,
	}
	if isAdd {
		req.IsAdd = 1
	}

	reply := &dhcp.DhcpClientConfigReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// SetInterfaceAsDHCPClient sets provided interface as a DHCP client
func SetInterfaceAsDHCPClient(ifIdx uint32, hostName string, vppChan VPPChannel, stopwatch *measure.Stopwatch) (err error) {
	return handleInterfaceDHCP(ifIdx, hostName, true, vppChan, stopwatch)
}

// UnsetInterfaceAsDHCPClient un-sets interface as DHCP client
func UnsetInterfaceAsDHCPClient(ifIdx uint32, hostName string, vppChan VPPChannel, stopwatch *measure.Stopwatch) (err error) {
	return handleInterfaceDHCP(ifIdx, hostName, false, vppChan, stopwatch)
}
