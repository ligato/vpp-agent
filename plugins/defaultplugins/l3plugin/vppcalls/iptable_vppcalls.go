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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/ip"
)

// IPTable (VRF table)
type IPTable struct {
	TableID uint32
	IsIPv6  bool
	Name    []byte
}

func vppAddDelIPTable(table *IPTable, vppChan *govppapi.Channel, delete bool) error {
	req := &ip.IPTableAddDel{
		TableID: table.TableID,
	}
	if delete {
		req.IsAdd = 0
	} else {
		req.IsAdd = 1
	}
	if table.IsIPv6 {
		req.IsIpv6 = 1
	} else {
		req.IsIpv6 = 0
	}

	// Send message
	reply := new(ip.IPTableAddDelReply)
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("IPTableAddDel returned %d", reply.Retval)
	}

	return nil
}

// VppAddIPTable adds new IP table according to provided input
func VppAddIPTable(table *IPTable, vppChan *govppapi.Channel) error {
	return vppAddDelIPTable(table, vppChan, false)
}

// VppDelIPTable removes old IP table according to provided input
func VppDelIPTable(table *IPTable, vppChan *govppapi.Channel) error {
	return vppAddDelIPTable(table, vppChan, true)
}
