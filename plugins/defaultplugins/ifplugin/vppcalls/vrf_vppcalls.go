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
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/ip"
)

/*
// VppUnsetAllInterfacesFromVRF removes all interfaces from VRF (set them to default VRF 0)
func VppUnsetAllInterfacesFromVRF(vrfIndex uint32, log logging.Logger,
	vppChan *govppapi.Channel) error {
	log.Debugf("Unsetting all interfaces from VRF %v", vrfIndex)

	return nil
}*/

// GetInterfaceVRF assigns VRF table to interface
func GetInterfaceVRF(ifaceIndex uint32, log logging.Logger,
	vppChan *govppapi.Channel) (uint32, error) {
	log.Debugf("Getting VRF for interface %v", ifaceIndex)

	req := &interfaces.SwInterfaceGetTable{
		SwIfIndex: ifaceIndex,
	}
	/*if table.IsIPv6 {
		req.IsIpv6 = 1
	} else {
		req.IsIpv6 = 0
	}*/

	// Send message
	reply := new(interfaces.SwInterfaceGetTableReply)
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("SwInterfaceGetTableReply returned %d", reply.Retval)
	}

	return reply.VrfID, nil
}

// SetInterfaceVRF retrieves VRF table from interface
func SetInterfaceVRF(ifaceIndex, vrfIndex uint32, log logging.Logger,
	vppChan *govppapi.Channel) error {
	log.Debugf("Setting interface %v to VRF %v", ifaceIndex, vrfIndex)

	req := &interfaces.SwInterfaceSetTable{
		VrfID:     vrfIndex,
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
		return fmt.Errorf("SwInterfaceSetTableReply returned %d", reply.Retval)
	}

	return nil
}

func dumpVrfTables(vppChan *govppapi.Channel) (map[uint32][]*ip.IPFibDetails, error) {
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

func vppAddDelIPTable(tableID uint32, vppChan *govppapi.Channel, delete bool) error {
	req := &ip.IPTableAddDel{
		TableID: tableID,
	}
	if delete {
		req.IsAdd = 0
	} else {
		req.IsAdd = 1
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

func createVrfIfNeeded(vrf uint32, vppChan *govppapi.Channel) error {
	if vrf == 0 {
		return nil
	}
	tables, err := dumpVrfTables(vppChan)
	if err != nil {
		return err
	}
	if _, ok := tables[vrf]; !ok {
		logroot.StandardLogger().Warnf("VXLAN: VRF table %v does not exists, creating it", vrf)
		return vppAddDelIPTable(vrf, vppChan, false)
	}
	return nil
}
