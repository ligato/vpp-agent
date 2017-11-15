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

// Note: acl_interface_add_del call is deprecated, replace by acl_interface_set_acl_list

import (
	"fmt"
	"time"

	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/bin_api/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
)

// SetACLToInterfacesAsIngress sets ACL to all provided interfaces as ingress.
func SetACLToInterfacesAsIngress(aclIndex uint32, interfaces []string, swIfIndexes ifaceidx.SwIfIndex, log logging.Logger,
	vppChannel *api.Channel, stopwatch *measure.Stopwatch) error {
	var timeLog measure.StopWatchEntry
	if stopwatch != nil {
		timeLog = measure.GetTimeLog(acl_api.ACLInterfaceSetACLList{}, stopwatch)
	}
	for _, ingressInterface := range interfaces {
		// Create acl list with new entry.
		var ACLs []uint32
		index, _, found := swIfIndexes.LookupIdx(ingressInterface)
		if !found {
			log.Debugf("Set interface to ACL: Interface %v not found ", ingressInterface)
			continue
		}
		// All previously assigned ACLs have to be dumped and added to acl list.
		aclInterface, err := DumpInterface(index, vppChannel, measure.GetTimeLog(acl_api.ACLInterfaceListDump{}, stopwatch))
		if err != nil {
			return err
		}
		// Construct ACL list. ACLs within NInput are defined as ingress, so provided new aclIndex has to be
		// added to the beginning of the list. todo it would be nicer to add new acl index to newNInput index
		if aclInterface != nil {
			ACLs = append(ACLs, aclIndex)
			for _, aclIndex := range aclInterface.Acls {
				ACLs = append(ACLs, aclIndex)
			}
		}
		newNInput := aclInterface.NInput + 1 // Rise NInput

		// Measure ACLInterfaceSetACLList time.
		start := time.Now()

		req := &acl_api.ACLInterfaceSetACLList{}
		req.Acls = ACLs
		req.Count = uint8(len(ACLs))
		req.SwIfIndex = index
		req.NInput = newNInput

		reply := &acl_api.ACLInterfaceSetACLListReply{}

		err = vppChannel.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to set interface %v to ACL %v", ingressInterface, aclIndex)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("set interface %v to ACL %v returned %v", ingressInterface, aclIndex, reply.Retval)
		}
		log.Debugf("Interface %v set to ACL %v as ingress", ingressInterface, aclIndex)

		// Log ACLInterfaceSetACLList time measurement results.
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

// SetACLToInterfacesAsEgress sets ACL to all provided interfaces as egress.
func SetACLToInterfacesAsEgress(aclIndex uint32, interfaces []string, swIfIndexes ifaceidx.SwIfIndex, log logging.Logger,
	vppChannel *api.Channel, stopwatch *measure.Stopwatch) error {
	var timeLog measure.StopWatchEntry
	if stopwatch != nil {
		timeLog = measure.GetTimeLog(acl_api.ACLInterfaceSetACLList{}, stopwatch)
	}
	for _, egressInterfaces := range interfaces {
		// Create empty ACL list.
		var ACLs []uint32
		index, _, found := swIfIndexes.LookupIdx(egressInterfaces)
		if !found {
			log.Debugf("Set interface to ACL: Interface %v not found ", egressInterfaces)
			continue
		}
		// All previously assigned ACLs have to be dumped and added to acl list.
		aclInterface, err := DumpInterface(index, vppChannel, measure.GetTimeLog(acl_api.ACLInterfaceListDump{}, stopwatch))
		if err != nil {
			return err
		}
		// Construct ACL list. ACLs outside of NInput are defined as egress,
		// so provided new aclIndex has to be added to the end of the list.
		if aclInterface != nil {
			for _, aclIndex := range aclInterface.Acls {
				ACLs = append(ACLs, aclIndex)
			}
			ACLs = append(ACLs, aclIndex)
		}

		// Measure ACLInterfaceSetACLList time.
		start := time.Now()

		req := &acl_api.ACLInterfaceSetACLList{}
		req.Acls = ACLs
		req.Count = uint8(len(ACLs))
		req.SwIfIndex = index
		req.NInput = aclInterface.NInput // NInput remains the same.

		reply := &acl_api.ACLInterfaceSetACLListReply{}

		err = vppChannel.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to set interface %v to ACL %v", egressInterfaces, aclIndex)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("set interface %v to ACL %v returned %v", egressInterfaces, aclIndex, reply.Retval)
		}
		log.Debugf("Interface %v set to ACL %v as egress", egressInterfaces, aclIndex)

		// Log ACLInterfaceSetACLList time measurement results.
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

// SetMacIPAclToInterface adds L2 ACL to interface.
func SetMacIPAclToInterface(aclIndex uint32, interfaces []string, swIfIndexes ifaceidx.SwIfIndex, log logging.Logger,
	vppChannel *api.Channel, timeLog measure.StopWatchEntry) error {
	for _, ingressInterface := range interfaces {
		// Measure MacipACLInterfaceAddDel time
		start := time.Now()

		ifIndex, _, found := swIfIndexes.LookupIdx(ingressInterface)
		if !found {
			log.Debugf("Set interface to ACL: Interface %v not found ", ingressInterface)
			continue
		}
		req := &acl_api.MacipACLInterfaceAddDel{}
		req.ACLIndex = aclIndex
		req.IsAdd = 1
		req.SwIfIndex = ifIndex

		reply := &acl_api.MacipACLInterfaceAddDelReply{}

		err := vppChannel.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to set interface %v to L2 ACL %v", ingressInterface, aclIndex)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("set interface %v to L2 ACL %v returned %v", ingressInterface, aclIndex, reply.Retval)
		}
		log.Debugf("Interface %v set to L2 ACL %v as ingress", ingressInterface, aclIndex)

		// Log MacipACLInterfaceAddDel time measurement results.
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

// RemoveIPIngressACLFromInterfaces removes ACL from interfaces.
func RemoveIPIngressACLFromInterfaces(removedACLIndex uint32, interfaces []string, swIfIndexes ifaceidx.SwIfIndex, log logging.Logger,
	vppChannel *api.Channel, stopwatch *measure.Stopwatch) error {
	var timeLog measure.StopWatchEntry
	if stopwatch != nil {
		timeLog = measure.GetTimeLog(acl_api.ACLInterfaceSetACLList{}, stopwatch)
	}
	for _, ingressInterface := range interfaces {
		// Create empty ACL list
		var ACLs []uint32
		index, _, found := swIfIndexes.LookupIdx(ingressInterface)
		if !found {
			log.Debugf("Remove interface from ACL: Interface %v not found ", ingressInterface)
			continue
		}
		// All assigned ACLs have to be dumped.
		aclInterface, err := DumpInterface(index, vppChannel, measure.GetTimeLog(acl_api.ACLInterfaceListDump{}, stopwatch))
		if err != nil {
			return err
		}
		// Reconstruct ACL list without removed ACL.
		if aclInterface != nil {
			for _, aclIndex := range aclInterface.Acls {
				if aclIndex != removedACLIndex {
					ACLs = append(ACLs, aclIndex)
				}
			}
		}
		newNInput := aclInterface.NInput - 1 // Decrease NInput.

		// Measure ACLInterfaceSetACLList time.
		start := time.Now()

		req := &acl_api.ACLInterfaceSetACLList{}
		req.Acls = ACLs
		req.Count = uint8(len(ACLs))
		req.SwIfIndex = index
		req.NInput = newNInput

		reply := &acl_api.ACLInterfaceSetACLListReply{}

		err = vppChannel.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to remove ACL %v from interface %v", removedACLIndex, aclInterface)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("remove ACL %v from interface %v returned error %v", removedACLIndex,
				removedACLIndex, reply.Retval)
		}
		log.Debugf("ACL %v removed from interface %v (ingress)", removedACLIndex, ingressInterface)

		// Log ACLInterfaceSetACLList time measurement results.
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}
	return nil
}

// RemoveIPEgressACLFromInterfaces removes ACL from interfaces.
func RemoveIPEgressACLFromInterfaces(removedACLIndex uint32, interfaces []string, swIfIndexes ifaceidx.SwIfIndex, log logging.Logger,
	vppChannel *api.Channel, stopwatch *measure.Stopwatch) error {
	var timeLog measure.StopWatchEntry
	if stopwatch != nil {
		timeLog = measure.GetTimeLog(acl_api.ACLInterfaceSetACLList{}, stopwatch)
	}
	for _, egressInterface := range interfaces {
		// Create empty ACL list.
		var ACLs []uint32
		index, _, found := swIfIndexes.LookupIdx(egressInterface)
		if !found {
			log.Debugf("Remove interface from ACL: Interface %v not found ", egressInterface)
			continue
		}
		// All assigned ACLs have to be dumped.
		aclInterface, err := DumpInterface(index, vppChannel, measure.GetTimeLog(acl_api.ACLInterfaceListDump{}, stopwatch))
		if err != nil {
			return err
		}
		// Reconstruct ACL list without removed ACL.
		if aclInterface != nil {
			for _, aclIndex := range aclInterface.Acls {
				if aclIndex != removedACLIndex {
					ACLs = append(ACLs, aclIndex)
				}
			}
		}

		// Measure ACLInterfaceSetACLList time.
		start := time.Now()

		req := &acl_api.ACLInterfaceSetACLList{}
		req.Acls = ACLs
		req.Count = uint8(len(ACLs))
		req.SwIfIndex = index
		req.NInput = aclInterface.NInput //NInput remains the same.

		reply := &acl_api.ACLInterfaceSetACLListReply{}

		err = vppChannel.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to remove ACL %v from interface %v", removedACLIndex, aclInterface)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("remove ACL %v from interface %v returned error %v", removedACLIndex,
				removedACLIndex, reply.Retval)
		}
		log.Debugf("ACL %v removed from interface %v (egress)", removedACLIndex, egressInterface)

		// Log ACLInterfaceSetACLList time measurement results.
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}
	return nil
}

// RemoveMacIPIngressACLFromInterfaces removes L2 ACL from interfaces.
func RemoveMacIPIngressACLFromInterfaces(removedACLIndex uint32, interfaces []string, swIfIndexes ifaceidx.SwIfIndex, log logging.Logger,
	vppChannel *api.Channel, timeLog measure.StopWatchEntry) error {
	for _, ingressInterface := range interfaces {
		// Measure MacipACLInterfaceAddDel time.
		start := time.Now()
		ifIndex, _, found := swIfIndexes.LookupIdx(ingressInterface)
		if !found {
			log.Debugf("Remove interface from ACL: Interface %v not found ", ingressInterface)
			continue
		}
		req := &acl_api.MacipACLInterfaceAddDel{}
		req.ACLIndex = removedACLIndex
		req.SwIfIndex = ifIndex
		req.IsAdd = 0

		reply := &acl_api.MacipACLInterfaceAddDelReply{}

		err := vppChannel.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to remove L2 ACL %v from interface %v", removedACLIndex, ingressInterface)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("remove L2 ACL %v from interface %v returned error %v", removedACLIndex,
				removedACLIndex, reply.Retval)
		}
		log.Debugf("L2 ACL %v removed from interface %v (ingress)", removedACLIndex, ingressInterface)

		// Log MacipACLInterfaceAddDel time measurement results.
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}
	return nil
}
