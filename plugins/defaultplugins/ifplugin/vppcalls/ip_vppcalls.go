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
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/interfaces"
	"github.com/ligato/cn-infra/logging"
)

// AddInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=1
func AddInterfaceIP(ifIdx uint32, addr *net.IPNet, log logging.Logger, vppChan *govppapi.Channel) error {
	// prepare the message
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

// DelInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=00
func DelInterfaceIP(ifIdx uint32, addr *net.IPNet, log logging.Logger, vppChan *govppapi.Channel) error {
	// prepare the message
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

	// send the message
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
