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

// SetInterfacesToBridgeDomain sets all provided interfaces to bridge domain.
func SetInterfacesToBridgeDomain(bd *l2.BridgeDomains_BridgeDomain, bdIdx uint32, bdIfaces []*l2.BridgeDomains_BridgeDomain_Interfaces,
	swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) {

	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.SwInterfaceSetL2Bridge{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	if len(bdIfaces) == 0 {
		log.Debugf("Bridge domain %v has no new interface to set", bd.Name)
		return
	}

	for _, bdIface := range bdIfaces {
		// Verify that interface exists, otherwise skip it.
		ifIdx, _, found := swIfIndices.LookupIdx(bdIface.Name)
		if !found {
			log.Debugf("Required bridge domain %v interface %v not found", bd.Name, bdIface.Name)
			continue
		}
		req := &l2ba.SwInterfaceSetL2Bridge{
			Enable:      1,
			BdID:        bdIdx,
			RxSwIfIndex: ifIdx,
			Shg:         uint8(bdIface.SplitHorizonGroup),
		}
		// Set as BVI.
		if bdIface.BridgedVirtualInterface {
			req.Bvi = 1
			log.Debugf("Interface %v set as BVI", bdIface.Name)
		}

		reply := &l2ba.SwInterfaceSetL2BridgeReply{}
		if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
			log.Errorf("Error while assigning interface %v to bd %v: %v", bdIface.Name, bd.Name, err)
			continue
		}
		if reply.Retval != 0 {
			log.Errorf("%s returned %d while assigning interface %v (idx %v) to bd %v", reply.GetMessageName(), reply.Retval, bdIface.Name, ifIdx, bd.Name)
			continue
		}

		log.WithFields(logging.Fields{"Interface": bdIface.Name, "BD": bd.Name}).Debug("Interface set to bridge domain")
	}
}

// UnsetInterfacesFromBridgeDomain removes all interfaces from bridge domain.
func UnsetInterfacesFromBridgeDomain(bd *l2.BridgeDomains_BridgeDomain, bdIdx uint32, bdIfaces []*l2.BridgeDomains_BridgeDomain_Interfaces,
	swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) {

	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.SwInterfaceSetL2Bridge{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	if len(bdIfaces) == 0 {
		log.Debugf("Bridge domain %v has no obsolete interface to unset", bd.Name)
		return
	}

	for _, bdIface := range bdIfaces {
		// If interface is not found, it's not needed to unset it.
		ifIdx, _, found := swIfIndices.LookupIdx(bdIface.Name)
		if !found {
			continue
		}
		req := &l2ba.SwInterfaceSetL2Bridge{
			Enable:      0,
			BdID:        bdIdx,
			RxSwIfIndex: ifIdx,
			Shg:         uint8(bdIface.SplitHorizonGroup),
		}
		reply := &l2ba.SwInterfaceSetL2BridgeReply{}
		if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
			log.Errorf("Error while removing interface %v from bd %v: %v", bdIface.Name, bd.Name, err)
			continue
		}
		if reply.Retval != 0 {
			log.Errorf("%s returned %d while removing interface %v from bd %v", reply.GetMessageName(), reply.Retval, bdIface.Name, bd.Name)
			continue
		}

		log.WithFields(logging.Fields{"Interface": bdIface.Name, "BD": bd.Name}).Debug("Interface unset from bridge domain")
	}
}

// SetInterfaceToBridgeDomain sets single interface to bridge domain.
func SetInterfaceToBridgeDomain(bridgeDomainIndex uint32, interfaceIndex uint32, bvi bool, log logging.Logger,
	vppChan VPPChannel, stopwatch *measure.Stopwatch) error {

	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.SwInterfaceSetL2Bridge{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &l2ba.SwInterfaceSetL2Bridge{
		Enable:      1,
		BdID:        bridgeDomainIndex,
		RxSwIfIndex: interfaceIndex,
	}
	if bvi {
		req.Bvi = 1
	} else {
		req.Bvi = 0
	}

	reply := &l2ba.SwInterfaceSetL2BridgeReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	log.WithFields(logging.Fields{"Interface": interfaceIndex, "BD": bridgeDomainIndex}).
		Debug("Interface set to bridge domain")

	return nil
}
