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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/bin_api/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
)

// ACLInterfaceLogicalReq groups multiple fields to not enumerate all of them in one function call
type ACLInterfaceLogicalReq struct {
	aclIndex   uint32
	interfaces []string
	ingress    bool
}

// ACLInterfacesVppCalls aggregates vpp calls related to the IP ACL interfaces
type ACLInterfacesVppCalls struct {
	vppChan          *govppapi.Channel
	swIfIndexes      ifaceidx.SwIfIndex
	dumpACLStopwatch measure.StopWatchEntry
	setACLStopwatch  measure.StopWatchEntry
}

// NewACLInterfacesVppCalls constructs IP ACL interfaces vpp calls object
func NewACLInterfacesVppCalls(vppChan *govppapi.Channel, swIfIndexes ifaceidx.SwIfIndex, stopwatch *measure.Stopwatch) *ACLInterfacesVppCalls {
	return &ACLInterfacesVppCalls{
		vppChan:          vppChan,
		swIfIndexes:      swIfIndexes,
		dumpACLStopwatch: measure.GetTimeLog(acl_api.ACLInterfaceListDump{}, stopwatch),
		setACLStopwatch:  measure.GetTimeLog(acl_api.ACLInterfaceSetACLList{}, stopwatch),
	}
}

// SetACLToInterfacesAsIngress sets ACL to all provided interfaces as ingress
func (acl *ACLInterfacesVppCalls) SetACLToInterfacesAsIngress(ACLIndex uint32, interfaces []string, log logging.Logger) error {
	log.Debugf("Setting up IP ingress ACL from interfaces: %v ", interfaces)

	return acl.requestSetACLToInterfaces(&ACLInterfaceLogicalReq{
		aclIndex:   ACLIndex,
		interfaces: interfaces,
		ingress:    true,
	}, log)
}

// RemoveIPIngressACLFromInterfaces removes ACL from interfaces
func (acl *ACLInterfacesVppCalls) RemoveIPIngressACLFromInterfaces(ACLIndex uint32, interfaces []string, log logging.Logger) error {
	log.Debugf("Removing IP ingress ACL from interfaces: %v ", interfaces)

	return acl.requestRemoveInterfacesFromACL(&ACLInterfaceLogicalReq{
		aclIndex:   ACLIndex,
		interfaces: interfaces,
		ingress:    true,
	}, log)
}

// SetACLToInterfacesAsEgress sets ACL to all provided interfaces as egress
func (acl *ACLInterfacesVppCalls) SetACLToInterfacesAsEgress(ACLIndex uint32, interfaces []string, log logging.Logger) error {
	log.Debugf("Setting up IP egress ACL from interfaces: %v ", interfaces)

	return acl.requestSetACLToInterfaces(&ACLInterfaceLogicalReq{
		aclIndex:   ACLIndex,
		interfaces: interfaces,
		ingress:    false,
	}, log)
}

// RemoveIPEgressACLFromInterfaces removes ACL from interfaces
func (acl *ACLInterfacesVppCalls) RemoveIPEgressACLFromInterfaces(ACLIndex uint32, interfaces []string, log logging.Logger) error {
	log.Debugf("Removing IP egress ACL from interfaces: %v ", interfaces)

	return acl.requestRemoveInterfacesFromACL(&ACLInterfaceLogicalReq{
		aclIndex:   ACLIndex,
		interfaces: interfaces,
		ingress:    false,
	}, log)
}

func (acl *ACLInterfacesVppCalls) requestSetACLToInterfaces(logicalReq *ACLInterfaceLogicalReq, log logging.Logger) error {
	for _, aclInterface := range logicalReq.interfaces {
		// Create acl list with new entry
		var ACLs []uint32
		index, _, found := acl.swIfIndexes.LookupIdx(aclInterface)
		if !found {
			log.Debugf("Set interface to ACL: Interface %v not found ", aclInterface)
			continue
		}
		// All previously assigned ACLs have to be dumped and added to acl list
		aclInterfaceDetails, err := DumpInterface(index, acl.vppChan, acl.dumpACLStopwatch)
		if err != nil {
			return err
		}

		nInput := aclInterfaceDetails.NInput
		if logicalReq.ingress {
			// Construct ACL list. ACLs within NInput are defined as ingress, so provided new aclIndex has to be
			// added to the beginning of the list
			// TODO it would be nicer to add new acl index to newNInput index
			if aclInterfaceDetails != nil {
				ACLs = append(ACLs, logicalReq.aclIndex)
				for _, aclIndex := range aclInterfaceDetails.Acls {
					ACLs = append(ACLs, aclIndex)
				}
			}
			nInput++ // Rise NInput
		} else {
			// Construct ACL list. ACLs outside of NInput are defined as egress, so provided new aclIndex has to be
			// added to the end of the list
			if aclInterfaceDetails != nil {
				for _, aclIndex := range aclInterfaceDetails.Acls {
					ACLs = append(ACLs, aclIndex)
				}
				ACLs = append(ACLs, logicalReq.aclIndex)
			}
			// NInput remains the same
		}

		// Measure ACLInterfaceSetACLList time
		start := time.Now()

		msg := &acl_api.ACLInterfaceSetACLList{}
		msg.Acls = ACLs
		msg.Count = uint8(len(ACLs))
		msg.SwIfIndex = index
		msg.NInput = nInput

		reply := &acl_api.ACLInterfaceSetACLListReply{}
		err = acl.vppChan.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			return err
		}
		if reply.Retval != 0 {
			return fmt.Errorf("setting up interface ACL list returned %v", reply.Retval)
		}

		log.WithFields(logging.Fields{"SwIdx index": msg.SwIfIndex, "AclIdx": logicalReq.aclIndex}).Debug("Interface set to ACL")

		// Log ACLInterfaceSetACLList time measurement results
		if acl.setACLStopwatch != nil {
			acl.setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

func (acl *ACLInterfacesVppCalls) requestRemoveInterfacesFromACL(logicalReq *ACLInterfaceLogicalReq, log logging.Logger) error {
	var wasErr error
	for _, aclInterface := range logicalReq.interfaces {
		// Create empty ACL list
		var ACLs []uint32
		index, _, found := acl.swIfIndexes.LookupIdx(aclInterface)
		if !found {
			log.Debugf("Remove interface from ACL: Interface %v not found ", aclInterface)
			continue
		}
		// All assigned ACLs have to be dumped
		aclInterfaceDetails, err := DumpInterface(index, acl.vppChan, acl.dumpACLStopwatch)
		if err != nil {
			return err
		}
		// Reconstruct ACL list without removed ACL
		if aclInterfaceDetails != nil {
			for _, aclIndex := range aclInterfaceDetails.Acls {
				if aclIndex != logicalReq.aclIndex {
					ACLs = append(ACLs, aclIndex)
				}
			}
		}

		nInput := aclInterfaceDetails.NInput
		// Decrease NInput if ingress, otherwise keep it the same
		if logicalReq.ingress {
			nInput--
		}

		// Measure ACLInterfaceSetACLList time
		start := time.Now()

		msg := &acl_api.ACLInterfaceSetACLList{}
		msg.Acls = ACLs
		msg.Count = uint8(len(ACLs))
		msg.SwIfIndex = index
		msg.NInput = nInput

		reply := &acl_api.ACLInterfaceSetACLListReply{}
		err = acl.vppChan.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			wasErr = err
		}
		if reply.Retval != 0 {
			wasErr = fmt.Errorf("setting up interface ACL list returned %v", reply.Retval)
		}

		log.WithFields(logging.Fields{"SwIdx index": msg.SwIfIndex, "AclIdx": logicalReq.aclIndex}).Debug("Interface removed from ACL")

		// Log ACLInterfaceSetACLList time measurement results
		if acl.setACLStopwatch != nil {
			acl.setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}

	return wasErr
}

// SetMacIPAclToInterface adds L2 ACL to interface.
func (acl *ACLInterfacesVppCalls) SetMacIPAclToInterface(aclIndex uint32, interfaces []string, log logging.Logger) error {
	for _, ingressInterface := range interfaces {
		// Measure MacipACLInterfaceAddDel time
		start := time.Now()

		ifIndex, _, found := acl.swIfIndexes.LookupIdx(ingressInterface)
		if !found {
			log.Debugf("Set interface to ACL: Interface %v not found ", ingressInterface)
			continue
		}
		req := &acl_api.MacipACLInterfaceAddDel{}
		req.ACLIndex = aclIndex
		req.IsAdd = 1
		req.SwIfIndex = ifIndex

		reply := &acl_api.MacipACLInterfaceAddDelReply{}

		err := acl.vppChan.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to set interface %v to L2 ACL %v", ingressInterface, aclIndex)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("set interface %v to L2 ACL %v returned %v", ingressInterface, aclIndex, reply.Retval)
		}
		log.Debugf("Interface %v set to L2 ACL %v as ingress", ingressInterface, aclIndex)

		// Log MacipACLInterfaceAddDel time measurement results.
		if acl.setACLStopwatch != nil {
			acl.setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

// RemoveMacIPIngressACLFromInterfaces removes L2 ACL from interfaces.
func (acl *ACLInterfacesVppCalls) RemoveMacIPIngressACLFromInterfaces(removedACLIndex uint32, interfaces []string, log logging.Logger) error {
	for _, ingressInterface := range interfaces {
		// Measure MacipACLInterfaceAddDel time.
		start := time.Now()
		ifIndex, _, found := acl.swIfIndexes.LookupIdx(ingressInterface)
		if !found {
			log.Debugf("Remove interface from ACL: Interface %v not found ", ingressInterface)
			continue
		}
		req := &acl_api.MacipACLInterfaceAddDel{}
		req.ACLIndex = removedACLIndex
		req.SwIfIndex = ifIndex
		req.IsAdd = 0

		reply := &acl_api.MacipACLInterfaceAddDelReply{}

		err := acl.vppChan.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to remove L2 ACL %v from interface %v", removedACLIndex, ingressInterface)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("remove L2 ACL %v from interface %v returned error %v", removedACLIndex,
				removedACLIndex, reply.Retval)
		}
		log.Debugf("L2 ACL %v removed from interface %v (ingress)", removedACLIndex, ingressInterface)

		// Log MacipACLInterfaceAddDel time measurement results.
		if acl.setACLStopwatch != nil {
			acl.setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}
	return nil
}
