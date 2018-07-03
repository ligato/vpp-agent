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
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppdump"
	acl_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
)

// ACLInterfaceLogicalReq groups multiple fields to not enumerate all of them in one function call
type ACLInterfaceLogicalReq struct {
	aclIndex  uint32
	ifIndices []uint32
	ingress   bool
}

// ACLInterfacesVppCalls aggregates vpp calls related to the IP ACL interfaces
type ACLInterfacesVppCalls struct {
	log             logging.Logger
	vppChan         govppapi.Channel
	swIfIndexes     ifaceidx.SwIfIndex
	stopwatch       *measure.Stopwatch
	setACLStopwatch measure.StopWatchEntry
}

// NewACLInterfacesVppCalls constructs IP ACL interfaces vpp calls object
func NewACLInterfacesVppCalls(log logging.Logger, vppChan govppapi.Channel, swIfIndexes ifaceidx.SwIfIndex, stopwatch *measure.Stopwatch) *ACLInterfacesVppCalls {
	return &ACLInterfacesVppCalls{
		log:             log,
		vppChan:         vppChan,
		swIfIndexes:     swIfIndexes,
		setACLStopwatch: measure.GetTimeLog(acl_api.ACLInterfaceSetACLList{}, stopwatch),
	}
}

// SetACLToInterfacesAsIngress sets ACL to all provided interfaces as ingress
func (acl *ACLInterfacesVppCalls) SetACLToInterfacesAsIngress(ACLIndex uint32, ifIndices []uint32) error {
	acl.log.Debugf("Setting up IP ingress ACL from interfaces: %v ", ifIndices)

	return acl.requestSetACLToInterfaces(&ACLInterfaceLogicalReq{
		aclIndex:  ACLIndex,
		ifIndices: ifIndices,
		ingress:   true,
	})
}

// RemoveIPIngressACLFromInterfaces removes ACL from interfaces
func (acl *ACLInterfacesVppCalls) RemoveIPIngressACLFromInterfaces(ACLIndex uint32, ifIndices []uint32) error {
	acl.log.Debugf("Removing IP ingress ACL from interfaces: %v ", ifIndices)

	return acl.requestRemoveInterfacesFromACL(&ACLInterfaceLogicalReq{
		aclIndex:  ACLIndex,
		ifIndices: ifIndices,
		ingress:   true,
	})
}

// SetACLToInterfacesAsEgress sets ACL to all provided interfaces as egress
func (acl *ACLInterfacesVppCalls) SetACLToInterfacesAsEgress(ACLIndex uint32, ifIndices []uint32) error {
	acl.log.Debugf("Setting up IP egress ACL from interfaces: %v ", ifIndices)

	return acl.requestSetACLToInterfaces(&ACLInterfaceLogicalReq{
		aclIndex:  ACLIndex,
		ifIndices: ifIndices,
		ingress:   false,
	})
}

// RemoveIPEgressACLFromInterfaces removes ACL from interfaces
func (acl *ACLInterfacesVppCalls) RemoveIPEgressACLFromInterfaces(ACLIndex uint32, ifIndices []uint32) error {
	acl.log.Debugf("Removing IP egress ACL from interfaces: %v ", ifIndices)

	return acl.requestRemoveInterfacesFromACL(&ACLInterfaceLogicalReq{
		aclIndex:  ACLIndex,
		ifIndices: ifIndices,
		ingress:   false,
	})
}

func (acl *ACLInterfacesVppCalls) requestSetACLToInterfaces(logicalReq *ACLInterfaceLogicalReq) error {
	for _, aclIfIdx := range logicalReq.ifIndices {
		// Create acl list with new entry
		var ACLs []uint32

		// All previously assigned ACLs have to be dumped and added to acl list
		aclInterfaceDetails, err := vppdump.DumpInterfaceIPACLs(aclIfIdx, acl.vppChan, acl.stopwatch)
		if err != nil {
			return err
		}

		var nInput uint8
		if aclInterfaceDetails != nil {
			nInput = aclInterfaceDetails.NInput
			if logicalReq.ingress {
				// Construct ACL list. ACLs within NInput are defined as ingress, so provided new aclIndex has to be
				// added to the beginning of the list
				// TODO it would be nicer to add new acl index to newNInput index
				ACLs = append(ACLs, logicalReq.aclIndex)
				for _, aclIndex := range aclInterfaceDetails.Acls {
					ACLs = append(ACLs, aclIndex)
				}
				nInput++ // Rise NInput
			} else {
				// Construct ACL list. ACLs outside of NInput are defined as egress, so provided new aclIndex has to be
				// added to the end of the list
				for _, aclIndex := range aclInterfaceDetails.Acls {
					ACLs = append(ACLs, aclIndex)
				}
				ACLs = append(ACLs, logicalReq.aclIndex)
				// NInput remains the same
			}
		}

		// Measure ACLInterfaceSetACLList time
		start := time.Now()

		msg := &acl_api.ACLInterfaceSetACLList{}
		msg.Acls = ACLs
		msg.Count = uint8(len(ACLs))
		msg.SwIfIndex = aclIfIdx
		msg.NInput = nInput

		reply := &acl_api.ACLInterfaceSetACLListReply{}
		err = acl.vppChan.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			return err
		}
		if reply.Retval != 0 {
			return fmt.Errorf("setting up interface ACL list returned %v", reply.Retval)
		}

		acl.log.WithFields(logging.Fields{"SwIdx index": msg.SwIfIndex, "AclIdx": logicalReq.aclIndex}).Debug("Interface set to ACL")

		// Log ACLInterfaceSetACLList time measurement results
		if acl.setACLStopwatch != nil {
			acl.setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

func (acl *ACLInterfacesVppCalls) requestRemoveInterfacesFromACL(logicalReq *ACLInterfaceLogicalReq) error {
	var wasErr error
	for _, aclIfIdx := range logicalReq.ifIndices {
		// Create empty ACL list
		var ACLs []uint32

		// All assigned ACLs have to be dumped
		aclInterfaceDetails, err := vppdump.DumpInterfaceIPACLs(aclIfIdx, acl.vppChan, acl.stopwatch)
		if err != nil {
			return err
		}

		// Reconstruct ACL list without removed ACL
		var nInput uint8
		if aclInterfaceDetails != nil {
			nInput = aclInterfaceDetails.NInput
			for idx, aclIndex := range aclInterfaceDetails.Acls {
				if (aclIndex != logicalReq.aclIndex) ||
					(logicalReq.ingress && idx >= int(aclInterfaceDetails.NInput)) ||
					(!logicalReq.ingress && idx < int(aclInterfaceDetails.NInput)) {
					ACLs = append(ACLs, aclIndex)
				} else {
					// Decrease NInput if ingress, otherwise keep it the same
					if logicalReq.ingress {
						nInput--
					}
				}
			}
		}

		// Measure ACLInterfaceSetACLList time
		start := time.Now()

		msg := &acl_api.ACLInterfaceSetACLList{}
		msg.Acls = ACLs
		msg.Count = uint8(len(ACLs))
		msg.SwIfIndex = aclIfIdx
		msg.NInput = nInput

		reply := &acl_api.ACLInterfaceSetACLListReply{}
		err = acl.vppChan.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			wasErr = err
		}
		if reply.Retval != 0 {
			wasErr = fmt.Errorf("setting up interface ACL list returned %v", reply.Retval)
		}

		acl.log.WithFields(logging.Fields{"SwIdx index": msg.SwIfIndex, "AclIdx": logicalReq.aclIndex}).Debug("Interface removed from ACL")

		// Log ACLInterfaceSetACLList time measurement results
		if acl.setACLStopwatch != nil {
			acl.setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}

	return wasErr
}

// SetMacIPAclToInterface adds L2 ACL to interface.
func (acl *ACLInterfacesVppCalls) SetMacIPAclToInterface(aclIndex uint32, ifIndices []uint32) error {
	for _, ingressIfIdx := range ifIndices {
		// Measure MacipACLInterfaceAddDel time
		start := time.Now()

		req := &acl_api.MacipACLInterfaceAddDel{}
		req.ACLIndex = aclIndex
		req.IsAdd = 1
		req.SwIfIndex = ingressIfIdx

		reply := &acl_api.MacipACLInterfaceAddDelReply{}

		err := acl.vppChan.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to set interface %v to L2 ACL %v", ingressIfIdx, aclIndex)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("set interface %v to L2 ACL %v returned %v", ingressIfIdx, aclIndex, reply.Retval)
		}
		acl.log.Debugf("Interface %v set to L2 ACL %v as ingress", ingressIfIdx, aclIndex)

		// Log MacipACLInterfaceAddDel time measurement results.
		if acl.setACLStopwatch != nil {
			acl.setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

// RemoveMacIPIngressACLFromInterfaces removes L2 ACL from interfaces.
func (acl *ACLInterfacesVppCalls) RemoveMacIPIngressACLFromInterfaces(removedACLIndex uint32, ifIndices []uint32) error {
	for _, ifIdx := range ifIndices {
		// Measure MacipACLInterfaceAddDel time.
		start := time.Now()

		req := &acl_api.MacipACLInterfaceAddDel{}
		req.ACLIndex = removedACLIndex
		req.SwIfIndex = ifIdx
		req.IsAdd = 0

		reply := &acl_api.MacipACLInterfaceAddDelReply{}

		err := acl.vppChan.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to remove L2 ACL %v from interface %v", removedACLIndex, ifIdx)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("remove L2 ACL %v from interface %v returned error %v", removedACLIndex,
				removedACLIndex, reply.Retval)
		}
		acl.log.Debugf("L2 ACL %v removed from interface %v (ingress)", removedACLIndex, ifIdx)

		// Log MacipACLInterfaceAddDel time measurement results.
		if acl.setACLStopwatch != nil {
			acl.setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}
	return nil
}
