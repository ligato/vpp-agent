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
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/interfaces"
)

// SetInterfaceMtu calls SwInterfaceSetMtu bin API with desired MTU value
func SetInterfaceMtu(ifIdx uint32, mtu uint32, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &interfaces.SwInterfaceSetMtu{}
	req.SwIfIndex = ifIdx
	req.Mtu = uint16(mtu)

	reply := &interfaces.SwInterfaceSetMtuReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("setting up interface MTU returned %d", reply.Retval)
	}
	logroot.StandardLogger().Debugf("MTU %v set to interface %v.", mtu, ifIdx)

	return nil
}
