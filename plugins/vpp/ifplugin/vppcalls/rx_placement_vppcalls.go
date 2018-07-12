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

	"strconv"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
)

// SetRxPlacement
func SetRxPlacement(vppInternalName string, rxPlacement *intf.Interfaces_Interface_RxPlacementSettings,
	vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(interfaces.SwInterfaceSetRxMode{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	queue := strconv.Itoa(int(rxPlacement.Queue))
	worker := strconv.Itoa(int(rxPlacement.Worker))

	command := "set interface rx-placement " + vppInternalName + " queue " + queue + " worker " + worker

	logrus.DefaultLogger().Warnf("Setting rx-placement commnad %s", command)

	// todo: binary api call for rx-placement is not available
	req := &vpe.CliInband{
		Length: uint32(len(command)),
		Cmd:    []byte(command),
	}

	reply := &vpe.CliInbandReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}
	if reply.Length > 0 {
		return fmt.Errorf("rx-placement setup replied with %s", string(reply.Reply))
	}

	return nil
}
