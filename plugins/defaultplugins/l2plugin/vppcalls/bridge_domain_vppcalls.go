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
	"time"

	"github.com/ligato/cn-infra/logging/measure"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
)

// VppAddBridgeDomain adds new bridge domain.
func VppAddBridgeDomain(bdIdx uint32, bd *l2.BridgeDomains_BridgeDomain, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.BridgeDomainAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &l2ba.BridgeDomainAddDel{
		IsAdd:   1,
		BdID:    bdIdx,
		Learn:   boolToUint(bd.Learn),
		ArpTerm: boolToUint(bd.ArpTermination),
		Flood:   boolToUint(bd.Flood),
		UuFlood: boolToUint(bd.UnknownUnicastFlood),
		Forward: boolToUint(bd.Forward),
		MacAge:  uint8(bd.MacAge),
		BdTag:   []byte(bd.Name),
	}

	reply := &l2ba.BridgeDomainAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// VppDeleteBridgeDomain removes existing bridge domain.
func VppDeleteBridgeDomain(bdIdx uint32, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.BridgeDomainAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &l2ba.BridgeDomainAddDel{
		IsAdd: 0,
		BdID:  bdIdx,
	}

	reply := &l2ba.BridgeDomainAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

/*
// VppUpdateBridgeDomain updates bridge domain parameters.
// TODO: removed, unused
func VppUpdateBridgeDomain(oldBdIdx uint32, newBdIdx uint32, newBridgeDomain *l2.BridgeDomains_BridgeDomain, log logging.Logger,
	vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	log.Debug("Updating VPP bridge domain parameters ", newBridgeDomain.Name)

	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.BridgeDomainAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &l2ba.BridgeDomainAddDel{
		IsAdd:   1,
		BdID:    newBdIdx,
		Learn:   boolToUint(newBridgeDomain.Learn),
		ArpTerm: boolToUint(newBridgeDomain.ArpTermination),
		Flood:   boolToUint(newBridgeDomain.Flood),
		UuFlood: boolToUint(newBridgeDomain.UnknownUnicastFlood),
		Forward: boolToUint(newBridgeDomain.Forward),
		MacAge:  uint8(newBridgeDomain.MacAge),
	}

	reply := &l2ba.BridgeDomainAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return fmt.Errorf("updating bridge domain failed with error %v", err)
	}
	if 0 != reply.Retval {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	log.WithFields(logging.Fields{"Name": newBridgeDomain.Name, "Index": newBdIdx}).Debug("Bridge domain Updated.")
	return nil
}*/
