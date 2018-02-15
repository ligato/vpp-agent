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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/dhcp"
)

// SetInterfaceAsDHCPClient sets provided interface as a DHCP client
func SetInterfaceAsDHCPClient(ifIdx uint32, hostName string, log logging.Logger, vppChan *govppapi.Channel, timeLog *measure.TimeLog) (err error) {
	// DhcpClientConfig time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err = handleInterfaceDHCP(ifIdx, hostName, log, vppChan, true); err != nil {
		return err
	}

	log.Debugf("Interface %v set as DHCP client", hostName)

	return err
}

// UnsetInterfaceAsDHCPClient un-sets interface as DHCP client
func UnsetInterfaceAsDHCPClient(ifIdx uint32, hostName string, log logging.Logger, vppChan *govppapi.Channel, timeLog *measure.TimeLog) (err error) {
	// DhcpClientConfig time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if err = handleInterfaceDHCP(ifIdx, hostName, log, vppChan, false); err != nil {
		return err
	}

	log.Debugf("Interface %v is no longer a DHCP client", hostName)

	return err
}

// SubscribeDHCPNotifications registers provided event channel to receive DHCP events
func SubscribeDHCPNotifications(eventChan chan govppapi.Message, vppChan *govppapi.Channel) (*govppapi.NotifSubscription, error) {
	if eventChan != nil {
		return vppChan.SubscribeNotification(eventChan, dhcp.NewDhcpComplEvent)
	} else {
		return nil, fmt.Errorf("provided channel is nil")
	}
}

func handleInterfaceDHCP(ifIdx uint32, hostName string, log logging.Logger, vppChan *govppapi.Channel, isAdd bool) error {
	req := &dhcp.DhcpClientConfig{
		SwIfIndex: ifIdx,
		Hostname:  []byte(hostName),
		IsAdd: func(isAdd bool) uint8 {
			if isAdd {
				return 1
			}
			return 0
		}(isAdd),
		WantDhcpEvent: 1,
	}

	reply := &dhcp.DhcpClientConfigReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("setting up interface as DHCP client returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"hostName": hostName, "ifIdx": ifIdx}).Debug("Interface set as DHCP client")

	return nil
}
