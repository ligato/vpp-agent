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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/af_packet"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
)

// AddAfPacketInterface calls AfPacketCreate VPP binary API.
func AddAfPacketInterface(afPacketIntf *intf.Interfaces_Interface_Afpacket, vppChan *govppapi.Channel) (swIndex uint32, err error) {
	// prepare the message
	req := &af_packet.AfPacketCreate{}

	req.HostIfName = []byte(afPacketIntf.HostIfName)
	req.UseRandomHwAddr = 1

	reply := &af_packet.AfPacketCreateReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return 0, err
	}

	if 0 != reply.Retval {
		return 0, fmt.Errorf("add af_packet interface returned %d", reply.Retval)
	}
	return reply.SwIfIndex, nil
}

// DeleteAfPacketInterface calls AfPacketDelete VPP binary API.
func DeleteAfPacketInterface(afPacketIntf *intf.Interfaces_Interface_Afpacket, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &af_packet.AfPacketDelete{}
	req.HostIfName = []byte(afPacketIntf.HostIfName)

	reply := &af_packet.AfPacketDeleteReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("deleting of af_packet interface returned %d", reply.Retval)
	}
	return nil
}
