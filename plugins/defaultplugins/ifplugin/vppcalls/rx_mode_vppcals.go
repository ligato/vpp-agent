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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/interfaces"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"time"
)

// SetRxMode calls SwInterfaceSetRxMode bin
func SetRxMode(ifIdx uint32, rxModeSettings intf.Interfaces_Interface_RxModeSettings,
	log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// SwInterfaceSetRxMode time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	//prepare the message
	req := &interfaces.SwInterfaceSetRxMode{}
	req.SwIfIndex = ifIdx
	req.Mode = uint8(rxModeSettings.RxMode)
	req.QueueID = rxModeSettings.QueueID
	req.QueueIDValid = uint8(rxModeSettings.QueueIDValid)

	log.Debug("set rxModeSettings: ", rxModeSettings)

	reply := &interfaces.SwInterfaceSetRxModeReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("setting rxModeSettings returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"RxModeType": rxModeSettings}).Debug("RxModeType ", rxModeSettings, "for interface ", ifIdx, " was set.")

	return nil

}
