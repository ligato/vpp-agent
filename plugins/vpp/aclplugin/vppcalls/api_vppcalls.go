// Copyright (c) 2018 Cisco and/or its affiliates.
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
	"github.com/ligato/cn-infra/logging/measure"
	aclapi "github.com/ligato/vpp-agent/plugins/vpp/binapi/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
)

// AclVppAPI provides read/write methods required to handle VPP access lists
type AclVppAPI interface {
	AclVppWrite
	AclVppRead
}

// AclVppWrite provides write methods for ACL plugin
type AclVppWrite interface {
	// AddIPAcl create new L3/4 ACL. Input index == 0xffffffff, VPP provides index in reply.
	AddIPAcl(rules []*acl.AccessLists_Acl_Rule, aclName string) (uint32, error)
	// AddMacIPAcl creates new L2 MAC IP ACL. VPP provides index in reply.
	AddMacIPAcl(rules []*acl.AccessLists_Acl_Rule, aclName string) (uint32, error)
	// ModifyIPAcl uses index (provided by VPP) to identify ACL which is modified.
	ModifyIPAcl(aclIndex uint32, rules []*acl.AccessLists_Acl_Rule, aclName string) error
	// ModifyMACIPAcl uses index (provided by VPP) to identify ACL which is modified.
	ModifyMACIPAcl(aclIndex uint32, rules []*acl.AccessLists_Acl_Rule, aclName string) error
	// DeleteIPAcl removes L3/L4 ACL.
	DeleteIPAcl(aclIndex uint32) error
	// DeleteMacIPAcl removes L2 ACL.
	DeleteMacIPAcl(aclIndex uint32) error
	// SetACLToInterfacesAsIngress sets ACL to all provided interfaces as ingress
	SetACLToInterfacesAsIngress(ACLIndex uint32, ifIndices []uint32) error
	// RemoveIPIngressACLFromInterfaces removes ACL from interfaces
	RemoveIPIngressACLFromInterfaces(ACLIndex uint32, ifIndices []uint32) error
	// SetACLToInterfacesAsEgress sets ACL to all provided interfaces as egress
	SetACLToInterfacesAsEgress(ACLIndex uint32, ifIndices []uint32) error
	// RemoveIPEgressACLFromInterfaces removes ACL from interfaces
	RemoveIPEgressACLFromInterfaces(ACLIndex uint32, ifIndices []uint32) error
	// SetMacIPAclToInterface adds L2 ACL to interface.
	SetMacIPAclToInterface(aclIndex uint32, ifIndices []uint32) error
	// RemoveMacIPIngressACLFromInterfaces removes L2 ACL from interfaces.
	RemoveMacIPIngressACLFromInterfaces(removedACLIndex uint32, ifIndices []uint32) error
}

// AclVppRead provides read methods for ACL plugin
type AclVppRead interface {
	// GetAclPluginVersion returns version of the VPP ACL plugin
	GetAclPluginVersion() (string, error)
	// DumpIPACL returns all IP-type ACLs
	DumpIPACL(swIfIndices ifaceidx.SwIfIndex) ([]*ACLEntry, error)
	// DumpIPACL returns all MACIP-type ACLs
	DumpMACIPACL(swIfIndices ifaceidx.SwIfIndex) ([]*ACLEntry, error)
	// DumpACLInterfaces returns a map of IP ACL indices with interfaces
	DumpIPACLInterfaces(indices []uint32, swIfIndices ifaceidx.SwIfIndex) (map[uint32]*acl.AccessLists_Acl_Interfaces, error)
	// DumpMACIPACLInterfaces returns a map of MACIP ACL indices with interfaces
	DumpMACIPACLInterfaces(indices []uint32, swIfIndices ifaceidx.SwIfIndex) (map[uint32]*acl.AccessLists_Acl_Interfaces, error)
	// DumpIPAcls returns a list of all configured ACLs with IP-type ruleData.
	DumpIPAcls() (map[ACLIdentifier][]aclapi.ACLRule, error)
	// DumpMacIPAcls returns a list of all configured ACL with IPMAC-type ruleData.
	DumpMacIPAcls() (map[ACLIdentifier][]aclapi.MacipACLRule, error)
	// DumpInterfaceAcls finds interface in VPP and returns its ACL configuration
	DumpInterfaceIPAcls(swIndex uint32) (acl.AccessLists, error)
	// DumpInterfaceMACIPAcls finds interface in VPP and returns its MACIP ACL configuration
	DumpInterfaceMACIPAcls(swIndex uint32) (acl.AccessLists, error)
	// DumpInterfaceIPACLs finds interface in VPP and returns its IP ACL configuration.
	DumpInterfaceIPACLs(swIndex uint32) (*aclapi.ACLInterfaceListDetails, error)
	// DumpInterfaceMACIPACLs finds interface in VPP and returns its MACIP ACL configuration.
	DumpInterfaceMACIPACLs(swIndex uint32) (*aclapi.MacipACLInterfaceListDetails, error)
	// DumpInterfaces finds  all interfaces in VPP and returns their ACL configurations
	DumpInterfaces() ([]*aclapi.ACLInterfaceListDetails, []*aclapi.MacipACLInterfaceListDetails, error)
}

// aclVppHandler is accessor for acl-related vppcalls methods
type aclVppHandler struct {
	stopwatch    *measure.Stopwatch
	callsChannel govppapi.Channel
	dumpChannel  govppapi.Channel
}

// NewAclVppHandler creates new instance of acl vppcalls handler
func NewAclVppHandler(callsChan, dumpChan govppapi.Channel, stopwatch *measure.Stopwatch) (*aclVppHandler, error) {
	handler := &aclVppHandler{
		callsChannel: callsChan,
		dumpChannel:  dumpChan,
		stopwatch:    stopwatch,
	}
	if err := handler.callsChannel.CheckMessageCompatibility(AclMessages...); err != nil {
		return nil, err
	}
	return handler, nil
}
