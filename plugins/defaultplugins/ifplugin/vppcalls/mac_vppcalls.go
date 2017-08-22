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
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/interfaces"
)

// SetInterfaceMac calls SwInterfaceSetMacAddress bin API
func SetInterfaceMac(ifIdx uint32, macAddress string, vppChan *govppapi.Channel) error {
	mac, macErr := net.ParseMAC(macAddress)
	if macErr != nil {
		return macErr
	}

	req := &interfaces.SwInterfaceSetMacAddress{}
	req.SwIfIndex = ifIdx
	req.MacAddress = mac

	reply := &interfaces.SwInterfaceSetMacAddressReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("Adding MAC address returned %d", reply.Retval)
	}
	log.DefaultLogger().WithFields(log.Fields{"MAC address": mac.String(), "ifIdx": ifIdx}).Debug("MAC address added")
	return nil
}
