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

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
)

// SetInterfacesToBridgeDomain attempts to set all provided interfaces to bridge domain. It returns a list of interfaces
// which were successfully set.
func SetInterfacesToBridgeDomain(bdName string, bdIdx uint32, bdIfs []*l2.BridgeDomains_BridgeDomain_Interfaces,
	swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) (ifs []string, wasErr error) {

	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.SwInterfaceSetL2Bridge{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	if len(bdIfs) == 0 {
		log.Debugf("Bridge domain %v has no new interface to set", bdName)
		return nil, nil
	}

	for _, bdIf := range bdIfs {
		// Verify that interface exists, otherwise skip it.
		ifIdx, _, found := swIfIndices.LookupIdx(bdIf.Name)
		if !found {
			log.Debugf("Required bridge domain %v interface %v not found", bdName, bdIf.Name)
			continue
		}
		if err := addDelInterfaceToBridgeDomain(bdName, bdIdx, bdIf, ifIdx, log, vppChan, true); err != nil {
			wasErr = err
			log.Error(wasErr)
		} else {
			log.WithFields(logging.Fields{"Interface": bdIf.Name, "BD": bdName}).Debug("Interface set to bridge domain")
			ifs = append(ifs, bdIf.Name)
		}
	}

	return ifs, wasErr
}

// UnsetInterfacesFromBridgeDomain removes all interfaces from bridge domain. It returns a list of interfaces
// which were successfully unset.
func UnsetInterfacesFromBridgeDomain(bdName string, bdIdx uint32, bdIfs []*l2.BridgeDomains_BridgeDomain_Interfaces,
	swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) (ifs []string, wasErr error) {

	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.SwInterfaceSetL2Bridge{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	if len(bdIfs) == 0 {
		log.Debugf("Bridge domain %v has no obsolete interface to unset", bdName)
		return nil, nil
	}

	for _, bdIf := range bdIfs {
		// Verify that interface exists, otherwise skip it.
		ifIdx, _, found := swIfIndices.LookupIdx(bdIf.Name)
		if !found {
			log.Debugf("Required bridge domain %v interface %v not found", bdName, bdIf.Name)
			continue
		}
		if err := addDelInterfaceToBridgeDomain(bdName, bdIdx, bdIf, ifIdx, log, vppChan, false); err != nil {
			wasErr = err
			log.Error(wasErr)
		} else {
			log.WithFields(logging.Fields{"Interface": bdIf.Name, "BD": bdName}).Debug("Interface unset from bridge domain")
			ifs = append(ifs, bdIf.Name)
		}
	}

	return ifs, wasErr
}

func addDelInterfaceToBridgeDomain(bdName string, bdIdx uint32, bdIf *l2.BridgeDomains_BridgeDomain_Interfaces,
	ifIdx uint32, log logging.Logger, vppChan VPPChannel, add bool) error {
	req := &l2ba.SwInterfaceSetL2Bridge{
		BdID:        bdIdx,
		RxSwIfIndex: ifIdx,
		Shg:         uint8(bdIf.SplitHorizonGroup),
	}
	// Enable
	if add {
		req.Enable = 1
	}
	// Set as BVI.
	if bdIf.BridgedVirtualInterface {
		req.Bvi = 1
		log.Debugf("Interface %v set as BVI", bdIf.Name)
	}

	reply := &l2ba.SwInterfaceSetL2BridgeReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return fmt.Errorf("error while assigning/removing interface %v to bd %v: %v", bdIf.Name, bdName, err)
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d while assigning/removing interface %v (idx %v) to bd %v",
			reply.GetMessageName(), reply.Retval, bdIf.Name, ifIdx, bdName)
	}

	return nil
}
