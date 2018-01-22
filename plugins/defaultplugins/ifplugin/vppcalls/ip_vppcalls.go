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

	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
)

// AddInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=1.
func AddInterfaceIP(ifIdx uint32, addr *net.IPNet, log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// SwInterfaceAddDelAddress time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message.
	req := &interfaces.SwInterfaceAddDelAddress{}
	req.SwIfIndex = ifIdx
	req.IsAdd = 1

	prefix, _ := addr.Mask.Size()
	req.AddressLength = byte(prefix)

	v6, err := addrs.IsIPv6(addr.IP.String())
	if err != nil {
		return err
	}
	if v6 {
		req.Address = []byte(addr.IP.To16())
		req.IsIpv6 = 1
	} else {
		req.Address = []byte(addr.IP.To4())
		req.IsIpv6 = 0
	}

	log.Debug("add req: IsIpv6: ", req.IsIpv6, " len(req.Address)=", len(req.Address))

	reply := &interfaces.SwInterfaceAddDelAddressReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("adding IP address returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"IPAddress": addr.IP, "mask": addr.Mask, "ifIdx": ifIdx}).Debug("IP address added.")

	return nil

}

// DelInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=00.
func DelInterfaceIP(ifIdx uint32, addr *net.IPNet, log logging.Logger, vppChan *govppapi.Channel, timeLog *measure.TimeLog) error {
	// SwInterfaceAddDelAddressReply time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message.
	req := &interfaces.SwInterfaceAddDelAddress{}
	req.SwIfIndex = ifIdx
	req.IsAdd = 0

	prefix, _ := addr.Mask.Size()
	req.AddressLength = byte(prefix)

	v6, err := addrs.IsIPv6(addr.IP.String())
	if err != nil {
		return err
	}
	if v6 {
		req.Address = []byte(addr.IP.To16())
		req.IsIpv6 = 1
	} else {
		req.Address = []byte(addr.IP.To4())
		req.IsIpv6 = 0
	}

	log.Debug("del req: IsIpv6: ", req.IsIpv6, " len(req.Address)=", len(req.Address))

	// Send the message.
	reply := &interfaces.SwInterfaceAddDelAddressReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("removing IP address returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"IPAddress": addr.IP, "mask": addr.Mask, "ifIdx": ifIdx}).Debug("IP address removed.")

	return nil
}

// SetUnnumberedIP sets interface as un-numbered, linking IP address of the another interface (ifIdxWithIP)
func SetUnnumberedIP(uIfIdx uint32, ifIdxWithIP uint32, log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// SwInterfaceAddDelAddress time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message.
	req := &interfaces.SwInterfaceSetUnnumbered{}
	req.SwIfIndex = ifIdxWithIP
	req.UnnumberedSwIfIndex = uIfIdx
	req.IsAdd = 1

	log.Debugf("set interface %v as un-numbered, with IP address from interface %v", uIfIdx, ifIdxWithIP)

	reply := &interfaces.SwInterfaceSetUnnumberedReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("setting un-numbered interfaces returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"un-numberedIface": uIfIdx, "IfaceWithIP": ifIdxWithIP}).Debug("Interface set as un-numbered")

	return nil
}

// UnsetUnnumberedIP unset provided interface as un-numbered. IP address of the linked interface is removed
func UnsetUnnumberedIP(uIfIdx uint32, log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// SwInterfaceAddDelAddress time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message.
	req := &interfaces.SwInterfaceSetUnnumbered{}
	req.UnnumberedSwIfIndex = uIfIdx
	req.IsAdd = 0

	log.Debug("unset un-numbered interface %v ", uIfIdx)

	reply := &interfaces.SwInterfaceSetUnnumberedReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("unsetting un-numbered interfaces returned %d", reply.Retval)
	}
	log.WithFields(logging.Fields{"un-numberedIface": uIfIdx}).Debug("Un-numbered interface unset")

	return nil
}
