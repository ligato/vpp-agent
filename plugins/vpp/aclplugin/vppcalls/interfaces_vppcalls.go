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
	acl_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/acl"
)

// ACLInterfaceLogicalReq groups multiple fields to not enumerate all of them in one function call
type ACLInterfaceLogicalReq struct {
	aclIndex  uint32
	ifIndices []uint32
	ingress   bool
}

func (handler *aclVppHandler) SetACLToInterfacesAsIngress(ACLIndex uint32, ifIndices []uint32) error {
	return handler.requestSetACLToInterfaces(&ACLInterfaceLogicalReq{
		aclIndex:  ACLIndex,
		ifIndices: ifIndices,
		ingress:   true,
	})
}

func (handler *aclVppHandler) RemoveIPIngressACLFromInterfaces(ACLIndex uint32, ifIndices []uint32) error {
	return handler.requestRemoveInterfacesFromACL(&ACLInterfaceLogicalReq{
		aclIndex:  ACLIndex,
		ifIndices: ifIndices,
		ingress:   true,
	})
}

func (handler *aclVppHandler) SetACLToInterfacesAsEgress(ACLIndex uint32, ifIndices []uint32) error {
	return handler.requestSetACLToInterfaces(&ACLInterfaceLogicalReq{
		aclIndex:  ACLIndex,
		ifIndices: ifIndices,
		ingress:   false,
	})
}

func (handler *aclVppHandler) RemoveIPEgressACLFromInterfaces(ACLIndex uint32, ifIndices []uint32) error {
	return handler.requestRemoveInterfacesFromACL(&ACLInterfaceLogicalReq{
		aclIndex:  ACLIndex,
		ifIndices: ifIndices,
		ingress:   false,
	})
}

func (handler *aclVppHandler) SetMacIPAclToInterface(aclIndex uint32, ifIndices []uint32) error {
	setACLStopwatch := measure.GetTimeLog(acl_api.MacipACLInterfaceAddDel{}, handler.stopwatch)
	for _, ingressIfIdx := range ifIndices {
		// Measure MacipACLInterfaceAddDel time
		start := time.Now()

		req := &acl_api.MacipACLInterfaceAddDel{
			ACLIndex:  aclIndex,
			IsAdd:     1,
			SwIfIndex: ingressIfIdx,
		}

		reply := &acl_api.MacipACLInterfaceAddDelReply{}

		err := handler.callsChannel.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to set interface %d to L2 ACL %d: %v", ingressIfIdx, aclIndex, err)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("set interface %d to L2 ACL %d returned %d", ingressIfIdx, aclIndex, reply.Retval)
		}

		// Log MacipACLInterfaceAddDel time measurement results.
		if setACLStopwatch != nil {
			setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

func (handler *aclVppHandler) RemoveMacIPIngressACLFromInterfaces(removedACLIndex uint32, ifIndices []uint32) error {
	setACLStopwatch := measure.GetTimeLog(acl_api.MacipACLInterfaceAddDel{}, handler.stopwatch)
	for _, ifIdx := range ifIndices {
		// Measure MacipACLInterfaceAddDel time.
		start := time.Now()

		req := &acl_api.MacipACLInterfaceAddDel{
			ACLIndex:  removedACLIndex,
			SwIfIndex: ifIdx,
			IsAdd:     0,
		}

		reply := &acl_api.MacipACLInterfaceAddDelReply{}

		err := handler.callsChannel.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("failed to remove L2 ACL %d from interface %d: %v", removedACLIndex, ifIdx, err)
		}
		if reply.Retval != 0 {
			return fmt.Errorf("remove L2 ACL %d from interface %d returned error %d", removedACLIndex,
				removedACLIndex, reply.Retval)
		}

		// Log MacipACLInterfaceAddDel time measurement results.
		if setACLStopwatch != nil {
			setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}
	return nil
}

func (handler *aclVppHandler) requestSetACLToInterfaces(logicalReq *ACLInterfaceLogicalReq) error {
	setACLStopwatch := measure.GetTimeLog(acl_api.ACLInterfaceSetACLList{}, handler.stopwatch)
	for _, aclIfIdx := range logicalReq.ifIndices {
		// Create acl list with new entry
		var ACLs []uint32

		// All previously assigned ACLs have to be dumped and added to acl list
		aclInterfaceDetails, err := handler.DumpInterfaceIPACLs(aclIfIdx)
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

		msg := &acl_api.ACLInterfaceSetACLList{
			Acls:      ACLs,
			Count:     uint8(len(ACLs)),
			SwIfIndex: aclIfIdx,
			NInput:    nInput,
		}

		reply := &acl_api.ACLInterfaceSetACLListReply{}
		err = handler.callsChannel.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			return err
		}
		if reply.Retval != 0 {
			return fmt.Errorf("setting up interface ACL list returned %v", reply.Retval)
		}

		// Log ACLInterfaceSetACLList time measurement results
		if setACLStopwatch != nil {
			setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

func (handler *aclVppHandler) requestRemoveInterfacesFromACL(logicalReq *ACLInterfaceLogicalReq) error {
	setACLStopwatch := measure.GetTimeLog(acl_api.ACLInterfaceSetACLList{}, handler.stopwatch)
	var wasErr error
	for _, aclIfIdx := range logicalReq.ifIndices {
		// Create empty ACL list
		var ACLs []uint32

		// All assigned ACLs have to be dumped
		aclInterfaceDetails, err := handler.DumpInterfaceIPACLs(aclIfIdx)
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

		msg := &acl_api.ACLInterfaceSetACLList{
			Acls:      ACLs,
			Count:     uint8(len(ACLs)),
			SwIfIndex: aclIfIdx,
			NInput:    nInput,
		}

		reply := &acl_api.ACLInterfaceSetACLListReply{}
		err = handler.callsChannel.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			wasErr = err
		}
		if reply.Retval != 0 {
			wasErr = fmt.Errorf("setting up interface ACL list returned %v", reply.Retval)
		}

		// Log ACLInterfaceSetACLList time measurement results
		if setACLStopwatch != nil {
			setACLStopwatch.LogTimeEntry(time.Since(start))
		}
	}

	return wasErr
}
