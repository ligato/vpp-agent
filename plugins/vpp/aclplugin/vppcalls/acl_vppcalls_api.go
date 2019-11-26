//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"

	"go.ligato.io/vpp-agent/v2/plugins/vpp"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/ifaceidx"
	acl "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/acl"
)

// Protocol types that can occur in ACLs
const (
	ICMPv4Proto = 1
	TCPProto    = 6
	UDPProto    = 17
	ICMPv6Proto = 58
)

// ACLDetails is combination of proto-modelled ACL data and VPP provided metadata
type ACLDetails struct {
	ACL  *acl.ACL `json:"acl"`
	Meta *ACLMeta `json:"acl_meta"`
}

// ACLMeta holds VPP-specific metadata
type ACLMeta struct {
	Index uint32 `json:"acl_index"`
	Tag   string `json:"acl_tag"`
}

// ACLToInterface is definition of interface and all ACLs which are bound to
// the interface either as ingress or egress
type ACLToInterface struct {
	SwIfIdx    uint32
	IngressACL []uint32
	EgressACL  []uint32
}

// ACLVppAPI provides read/write methods required to handle VPP access lists
type ACLVppAPI interface {
	ACLVppRead

	// AddACL create new ACL (L3/L4). Returns ACL index provided by VPP.
	AddACL(rules []*acl.ACL_Rule, aclName string) (aclIdx uint32, err error)
	// AddMACIPACL creates new MACIP ACL (L2). Returns ACL index provided by VPP.
	AddMACIPACL(rules []*acl.ACL_Rule, aclName string) (aclIdx uint32, err error)
	// ModifyACL modifies ACL (L3/L4) by updating its rules. It uses ACL index to identify ACL.
	ModifyACL(aclIdx uint32, rules []*acl.ACL_Rule, aclName string) error
	// ModifyACL modifies MACIP ACL (L2) by updating its rules. It uses ACL index to identify ACL.
	ModifyMACIPACL(aclIdx uint32, rules []*acl.ACL_Rule, aclName string) error
	// DeleteACL removes ACL (L3/L4).
	DeleteACL(aclIdx uint32) error
	// DeleteMACIPACL removes MACIP ACL (L2).
	DeleteMACIPACL(aclIdx uint32) error
	// SetACLToInterfacesAsIngress sets ACL to interfaces as ingress.
	SetACLToInterfacesAsIngress(ACLIndex uint32, ifIndices []uint32) error
	// RemoveACLFromInterfacesAsIngress removes ACL from interfaces as ingress.
	RemoveACLFromInterfacesAsIngress(ACLIndex uint32, ifIndices []uint32) error
	// SetACLToInterfacesAsEgress sets ACL to interfaces as egress.
	SetACLToInterfacesAsEgress(ACLIndex uint32, ifIndices []uint32) error
	// RemoveACLFromInterfacesAsEgress removes ACL from interfaces as egress.
	RemoveACLFromInterfacesAsEgress(ACLIndex uint32, ifIndices []uint32) error
	// SetMACIPACLToInterfaces sets MACIP ACL to interfaces.
	SetMACIPACLToInterfaces(aclIndex uint32, ifIndices []uint32) error
	// RemoveMACIPACLFromInterfaces removes MACIP ACL from interfaces.
	RemoveMACIPACLFromInterfaces(removedACLIndex uint32, ifIndices []uint32) error
	// AddACLToInterfaceAsIngress adds ACL (L3/L4) to single interface as ingress.
	AddACLToInterfaceAsIngress(aclIndex uint32, ifName string) error
	// AddACLToInterfaceAsEgress adds ACL (L3/L4) to single interface as egress.
	AddACLToInterfaceAsEgress(aclIndex uint32, ifName string) error
	// DeleteACLFromInterfaceAsIngress deletes ACL (L3/L4) from single interface as ingress.
	DeleteACLFromInterfaceAsIngress(aclIndex uint32, ifName string) error
	// DeleteACLFromInterfaceAsEgress deletes ACL (L3/L4) from single interface as egress.
	DeleteACLFromInterfaceAsEgress(aclIndex uint32, ifName string) error
	// AddMACIPACLToInterface adds MACIP ACL (L2) to single interface.
	AddMACIPACLToInterface(aclIndex uint32, ifName string) error
	// DeleteMACIPACLFromInterface deletes MACIP ACL (L2) from single interface.
	DeleteMACIPACLFromInterface(aclIndex uint32, ifName string) error
}

// ACLVppRead provides read methods for ACL plugin
type ACLVppRead interface {
	// DumpACL dumps all ACLs (L3/L4).
	DumpACL() ([]*ACLDetails, error)
	// DumpMACIPACL dumps all MACIP ACLs (L2).
	DumpMACIPACL() ([]*ACLDetails, error)
	// DumpACLInterfaces dumps all ACLs (L3/L4) for given ACL indexes. Returns map of ACL indexes with assigned interfaces.
	DumpACLInterfaces(indices []uint32) (map[uint32]*acl.ACL_Interfaces, error)
	// DumpMACIPACLInterfaces dumps all ACLs (L2) for given ACL indexes. Returns map of MACIP ACL indexes with assigned interfaces.
	DumpMACIPACLInterfaces(indices []uint32) (map[uint32]*acl.ACL_Interfaces, error)
	// DumpInterfaceAcls finds interface in VPP and returns its ACL (L3/L4) configuration.
	DumpInterfaceACLs(ifIdx uint32) ([]*acl.ACL, error)
	// DumpInterfaceMACIPACLs finds interface in VPP and returns its MACIP ACL (L2) configuration.
	DumpInterfaceMACIPACLs(ifIdx uint32) ([]*acl.ACL, error)
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "acl",
	HandlerAPI: (*ACLVppAPI)(nil),
})

type NewHandlerFunc func(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex) ACLVppAPI

func AddHandlerVersion(version string, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			return c.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			return h(c, a[0].(ifaceidx.IfaceMetadataIndex))
		},
	})
}

func CompatibleACLHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex) ACLVppAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx).(ACLVppAPI)
	}
	return nil
}
