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

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
)

// GetInterfaceVRF assigns VRF table to interface
func GetInterfaceVRF(ifIdx uint32, log logging.Logger, vppChan VPPChannel) (vrfID uint32, err error) {
	log.Debugf("Getting VRF for interface %v", ifIdx)

	req := &interfaces.SwInterfaceGetTable{
		SwIfIndex: ifIdx,
	}
	/*if table.IsIPv6 {
		req.IsIpv6 = 1
	} else {
		req.IsIpv6 = 0
	}*/

	// Send message
	reply := &interfaces.SwInterfaceGetTableReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return reply.VrfID, nil
}

// SetInterfaceVRF retrieves VRF table from interface
func SetInterfaceVRF(ifaceIndex, vrfID uint32, log logging.Logger, vppChan VPPChannel) error {
	if err := CreateVrfIfNeeded(vrfID, vppChan); err != nil {
		log.Warnf("creating VRF failed: %v", err)
		return err
	}

	log.Debugf("Setting interface %v to VRF %v", ifaceIndex, vrfID)

	req := &interfaces.SwInterfaceSetTable{
		VrfID:     vrfID,
		SwIfIndex: ifaceIndex,
	}
	/*if table.IsIPv6 {
		req.IsIpv6 = 1
	} else {
		req.IsIpv6 = 0
	}*/

	// Send message
	reply := new(interfaces.SwInterfaceSetTableReply)
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// TODO: manage VRF tables globally in separate configurator

// CreateVrfIfNeeded checks if VRF exists and creates it if not
func CreateVrfIfNeeded(vrfID uint32, vppChan VPPChannel) error {
	if vrfID == 0 {
		return nil
	}

	tables, err := dumpVrfTables(vppChan)
	if err != nil {
		logrus.DefaultLogger().Warnf("dumping VRF tables failed: %v", err)
		return err
	}
	if _, ok := tables[vrfID]; !ok {
		logrus.DefaultLogger().Warnf("VRF table %v does not exists, creating it", vrfID)
		return vppAddIPTable(vrfID, vppChan)
	}

	return nil
}

func dumpVrfTables(vppChan VPPChannel) (map[uint32][]*ip.IPFibDetails, error) {
	fibs := map[uint32][]*ip.IPFibDetails{}

	reqCtx := vppChan.SendMultiRequest(&ip.IPFibDump{})
	for {
		fibDetails := &ip.IPFibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			return nil, err
		}

		tableID := fibDetails.TableID
		fibs[tableID] = append(fibs[tableID], fibDetails)
	}

	return fibs, nil
}

func vppAddIPTable(tableID uint32, vppChan VPPChannel) error {
	req := &ip.IPTableAddDel{
		TableID: tableID,
		IsAdd:   1,
	}

	// Send message
	reply := &ip.IPTableAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}
