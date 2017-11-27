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
	"container/list"
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/bin_api/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"time"
)

// ACLInterfaceLogicalReq groups multiple fields to not enumerate all of them in one function call
type ACLInterfaceLogicalReq struct {
	aclIndex   uint32
	interfaces []string
	ingress    bool
	callback   func(error)
}

// ACLInterfacesVppCalls aggregates vpp calls related to the IP ACL interfaces
type ACLInterfacesVppCalls struct {
	vppChan         *govppapi.Channel
	asyncVppChan    *govppapi.Channel
	swIfIndexes     ifaceidx.SwIfIndex
	dumpIfaces      measure.StopWatchEntry
	ifaceSetACLList measure.StopWatchEntry
	waitingForReply *list.List
}

// NewACLInterfacesVppCalls constructs IP ACL interfaces vpp calls object
func NewACLInterfacesVppCalls(asyncVppChan *govppapi.Channel, vppChan *govppapi.Channel, swIfIndexes ifaceidx.SwIfIndex, stopwatch *measure.Stopwatch) *ACLInterfacesVppCalls {
	return &ACLInterfacesVppCalls{
		vppChan:         vppChan,
		asyncVppChan:    asyncVppChan,
		swIfIndexes:     swIfIndexes,
		dumpIfaces:      measure.GetTimeLog(acl_api.ACLInterfaceListDump{}, stopwatch),
		ifaceSetACLList: measure.GetTimeLog(acl_api.ACLInterfaceSetACLList{}, stopwatch),
		waitingForReply: list.New(),
	}
}

// SetACLToInterfacesAsIngress sets ACL to all provided interfaces as ingress
func (acl *ACLInterfacesVppCalls) SetACLToInterfacesAsIngress(ACLIndex uint32, interfaces []string, callback func(error), log logging.Logger) error {
	log.Debugf("Setting up IP ingress ACL from interfaces: %v ", interfaces)

	return acl.requestAsyncSetACLToInterfaces(&ACLInterfaceLogicalReq{
		aclIndex:   ACLIndex,
		interfaces: interfaces,
		ingress:    true,
		callback:   callback,
	}, log)
}

// RemoveIPIngressACLFromInterfaces removes ACL from interfaces
func (acl *ACLInterfacesVppCalls) RemoveIPIngressACLFromInterfaces(ACLIndex uint32, interfaces []string, callback func(error), log logging.Logger) error {
	log.Debugf("Removing IP ingress ACL from interfaces: %v ", interfaces)

	return acl.requestSyncRemoveInterfacesFromACL(&ACLInterfaceLogicalReq{
		aclIndex:   ACLIndex,
		interfaces: interfaces,
		ingress:    true,
		callback:   callback,
	}, log)
}

// SetACLToInterfacesAsEgress sets ACL to all provided interfaces as egress
func (acl *ACLInterfacesVppCalls) SetACLToInterfacesAsEgress(ACLIndex uint32, interfaces []string, callback func(error), log logging.Logger) error {
	log.Debugf("Setting up IP egress ACL from interfaces: %v ", interfaces)

	return acl.requestAsyncSetACLToInterfaces(&ACLInterfaceLogicalReq{
		aclIndex:   ACLIndex,
		interfaces: interfaces,
		ingress:    false,
		callback:   callback,
	}, log)
}

// RemoveIPEgressACLFromInterfaces removes ACL from interfaces
func (acl *ACLInterfacesVppCalls) RemoveIPEgressACLFromInterfaces(ACLIndex uint32, interfaces []string, callback func(error), log logging.Logger) error {
	log.Debugf("Removing IP egress ACL from interfaces: %v ", interfaces)

	return acl.requestSyncRemoveInterfacesFromACL(&ACLInterfaceLogicalReq{
		aclIndex:   ACLIndex,
		interfaces: interfaces,
		ingress:    false,
		callback:   callback,
	}, log)
}

func (acl *ACLInterfacesVppCalls) requestAsyncSetACLToInterfaces(logicalReq *ACLInterfaceLogicalReq, log logging.Logger) error {
	for _, aclInterface := range logicalReq.interfaces {
		// Create acl list with new entry
		var ACLs []uint32
		index, _, found := acl.swIfIndexes.LookupIdx(aclInterface)
		if !found {
			log.Debugf("Set interface to ACL: Interface %v not found ", aclInterface)
			continue
		}
		// All previously assigned ACLs have to be dumped and added to acl list
		aclInterfaceDetails, err := DumpInterface(index, acl.vppChan, acl.dumpIfaces)
		if err != nil {
			return err
		}

		nInput := aclInterfaceDetails.NInput
		if logicalReq.ingress {
			// Construct ACL list. ACLs within NInput are defined as ingress, so provided new aclIndex has to be
			// added to the beginning of the list todo it would be nicer to add new acl index to newNInput index
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

		acl.waitingForReply.PushFront(logicalReq)
		acl.asyncVppChan.ReqChan <- &govppapi.VppRequest{
			Message: msg,
		}

		log.WithFields(logging.Fields{"SwIdx index": msg.SwIfIndex, "AclIdx": logicalReq.aclIndex}).Debug("Interface set to ACL")

		// Log ACLInterfaceSetACLList time measurement results
		if acl.ifaceSetACLList != nil {
			acl.ifaceSetACLList.LogTimeEntry(time.Since(start))
		}
	}

	return nil
}

func (acl *ACLInterfacesVppCalls) requestSyncRemoveInterfacesFromACL(logicalReq *ACLInterfaceLogicalReq, log logging.Logger) error {
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
		aclInterfaceDetails, err := DumpInterface(index, acl.vppChan, acl.dumpIfaces)
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

		// Do not use asynchronous handling for removing ACLs from interface
		err = acl.vppChan.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			wasErr = err
		}
		if reply.Retval != 0 {
			log.Error("Setting up interface ACL list returned %v", reply.Retval)
		}

		log.WithFields(logging.Fields{"SwIdx index": msg.SwIfIndex, "AclIdx": logicalReq.aclIndex}).Debug("Interface removed from ACL")

		// Log ACLInterfaceSetACLList time measurement results
		if acl.ifaceSetACLList != nil {
			acl.ifaceSetACLList.LogTimeEntry(time.Since(start))
		}
	}

	return wasErr
}

// WatchACLInterfacesReplies is meant to be used in go routine
func (acl *ACLInterfacesVppCalls) WatchACLInterfacesReplies(log logging.Logger) {
	for {
		vppReply := <-acl.asyncVppChan.ReplyChan
		log.Debug("VPP ACL Reply ", vppReply)

		if vppReply.LastReplyReceived {
			log.Debug("Ping received")
			continue
		}

		if acl.waitingForReply.Len() == 0 {
			log.WithField("MessageID", vppReply.MessageID).Error("Unexpected message ", vppReply.Error)
			continue
		}

		logicalReq := acl.waitingForReply.Remove(acl.waitingForReply.Front()).(*ACLInterfaceLogicalReq)

		if vppReply.Error != nil {
			logicalReq.callback(vppReply.Error)
		} else {
			reply := &acl_api.ACLInterfaceSetACLListReply{}
			err := acl.asyncVppChan.MsgDecoder.DecodeMsg(vppReply.Data, reply)
			if err != nil {
				err = fmt.Errorf("set ACL list for interface returned %d", reply.Retval)
				logicalReq.callback(err)
			} else {
				logicalReq.callback(nil)
			}
		}
	}
}
