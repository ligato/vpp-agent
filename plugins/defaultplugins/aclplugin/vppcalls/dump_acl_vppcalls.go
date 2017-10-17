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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/bin_api/acl"
	"time"
)

// DumpInterface finds interface in VPP and returns its ACL configuration
func DumpInterface(swIndex uint32, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) (*acl.ACLInterfaceListDetails, error) {
	// ACLInterfaceListDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &acl.ACLInterfaceListDump{}
	req.SwIfIndex = swIndex

	msg := &acl.ACLInterfaceListDetails{}

	err := vppChannel.SendRequest(req).ReceiveReply(msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// DumpIPAcl test function
func DumpIPAcl(log logging.Logger, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// ACLDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &acl.ACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &acl.ACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return err
		}
		if stop {
			break
		}
		log.Infof("ACL index: %v, rule count: %v, tag: %v", msg.ACLIndex, msg.Count, string(msg.Tag[:]))

	}
	return nil
}

// DumpMacIPAcl test function
func DumpMacIPAcl(log logging.Logger, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// MacipACLDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &acl.MacipACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &acl.MacipACLDump{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return err
		}
		if stop {
			break
		}
		log.Info(msg.ACLIndex)
	}
	return nil
}

// DumpInterfaces test function
func DumpInterfaces(swIndexes idxvpp.NameToIdxRW, log logging.Logger, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// ACLInterfaceListDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &acl.ACLInterfaceListDump{}
	req.SwIfIndex = 0xffffffff
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &acl.ACLInterfaceListDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return err
		}
		if stop {
			break
		}
		name, _, found := swIndexes.LookupName(msg.SwIfIndex)
		if !found {
			continue
		}
		log.Infof("Interface %v is in %v acl in direction %v and applied in %v",
			name, msg.Count, msg.NInput, msg.Acls)
	}
	return nil
}
